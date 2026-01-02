// Package export provides data export/import capabilities with encryption.
package export

import (
	"archive/tar"
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	stderrors "errors"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/pbkdf2"

	"github.com/kimhsiao/memonexus/backend/internal/db"
	"github.com/kimhsiao/memonexus/backend/internal/errors"
	"github.com/kimhsiao/memonexus/backend/internal/logging"
	"github.com/kimhsiao/memonexus/backend/internal/models"
)

// ExportService provides export/import functionality.
type ExportService struct {
	repo db.ContentItemRepository
}

// NewExportService creates a new ExportService.
func NewExportService(repo db.ContentItemRepository) *ExportService {
	return &ExportService{repo: repo}
}

// ExportConfig holds export configuration.
type ExportConfig struct {
	OutputPath   string
	Password     string // For encryption
	IncludeMedia bool
}

// ImportConfig holds import configuration.
type ImportConfig struct {
	ArchivePath string
	Password    string // For decryption
}

// ExportManifest represents the export manifest metadata.
type ExportManifest struct {
	Version      string    `json:"version"`
	ExportedAt   time.Time `json:"exported_at"`
	ItemCount    int       `json:"item_count"`
	Checksum     string    `json:"checksum"`
	Encrypted    bool      `json:"encrypted"`
	IncludeMedia bool      `json:"include_media"`
}

// ExportResult represents the result of an export operation.
type ExportResult struct {
	FilePath  string
	SizeBytes int64
	ItemCount int
	Checksum  string
	Encrypted bool
	Duration  time.Duration
}

// ImportResult represents the result of an import operation.
type ImportResult struct {
	ImportedCount int
	SkippedCount  int
	Duration      time.Duration
}

// Export creates an encrypted export archive of all data.
// T211: Critical operations logging (start/complete/success/failure).
func (s *ExportService) Export(config *ExportConfig) (*ExportResult, error) {
	startTime := time.Now()
	correlationID := uuid.New().String()

	// Log operation start
	logging.Info("Export operation started",
		map[string]interface{}{
			"correlation_id": correlationID,
			"output_path":    config.OutputPath,
			"include_media":  config.IncludeMedia,
			"encryption":     config.Password != "",
		})

	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "memonexus-export-*")
	if err != nil {
		logging.ErrorWithCode("Export failed: temp directory creation",
			string(errors.ErrInternal), err,
			map[string]interface{}{
				"correlation_id": correlationID,
			})
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Create manifest
	manifest := ExportManifest{
		Version:      "1.0",
		ExportedAt:   startTime,
		Encrypted:    config.Password != "",
		IncludeMedia: config.IncludeMedia,
	}

	// Create data file
	dataFile := filepath.Join(tempDir, "data.json")
	itemCount, checksum, err := s.createDataFile(dataFile)
	if err != nil {
		logging.ErrorWithCode("Export failed: data file creation",
			string(errors.ErrInternal), err,
			map[string]interface{}{
				"correlation_id": correlationID,
			})
		return nil, fmt.Errorf("failed to create data file: %w", err)
	}

	manifest.ItemCount = itemCount
	manifest.Checksum = checksum

	// Create manifest file
	manifestFile := filepath.Join(tempDir, "manifest.json")
	if err := s.writeManifest(manifestFile, &manifest); err != nil {
		logging.ErrorWithCode("Export failed: manifest write",
			string(errors.ErrInternal), err,
			map[string]interface{}{
				"correlation_id": correlationID,
			})
		return nil, fmt.Errorf("failed to write manifest: %w", err)
	}

	// Determine output path
	archivePath := config.OutputPath
	if archivePath == "" {
		archivePath = fmt.Sprintf("exports/memonexus_%s.tar.gz",
			startTime.Format("20060102_150405"))
	}

	// Ensure exports directory exists
	if err := os.MkdirAll(filepath.Dir(archivePath), 0755); err != nil {
		logging.ErrorWithCode("Export failed: exports directory creation",
			string(errors.ErrInternal), err,
			map[string]interface{}{
				"correlation_id": correlationID,
				"archive_path":   archivePath,
			})
		return nil, fmt.Errorf("failed to create exports directory: %w", err)
	}

	// Create archive (tar.gz)
	plainPath := archivePath + ".plain.tmp"
	if err := writeTarGz(tempDir, plainPath); err != nil {
		logging.ErrorWithCode("Export failed: archive creation",
			string(errors.ErrInternal), err,
			map[string]interface{}{
				"correlation_id": correlationID,
			})
		return nil, fmt.Errorf("failed to create tar.gz: %w", err)
	}
	defer os.Remove(plainPath)

	// Optionally encrypt
	var finalPath string
	var sizeBytes int64

	if config.Password == "" {
		// No encryption, just rename
		finalPath = archivePath
		if err := os.Rename(plainPath, finalPath); err != nil {
			logging.ErrorWithCode("Export failed: archive move",
				string(errors.ErrInternal), err,
				map[string]interface{}{
					"correlation_id": correlationID,
				})
			return nil, fmt.Errorf("failed to move archive: %w", err)
		}
		info, err := os.Stat(finalPath)
		if err != nil {
			logging.ErrorWithCode("Export failed: archive stat",
				string(errors.ErrInternal), err,
				map[string]interface{}{
					"correlation_id": correlationID,
				})
			return nil, fmt.Errorf("failed to stat archive: %w", err)
		}
		sizeBytes = info.Size()
	} else {
		// Encrypt the archive
		finalPath = archivePath
		sizeBytes, err = encryptFile(plainPath, finalPath, config.Password)
		if err != nil {
			logging.ErrorWithCode("Export failed: encryption",
				string(errors.ErrCryptoFailed), err,
				map[string]interface{}{
					"correlation_id": correlationID,
				})
			return nil, fmt.Errorf("failed to encrypt archive: %w", err)
		}
	}

	result := &ExportResult{
		FilePath:  finalPath,
		SizeBytes: sizeBytes,
		ItemCount: itemCount,
		Checksum:  checksum,
		Encrypted: config.Password != "",
		Duration:  time.Since(startTime),
	}

	// Log successful completion
	logging.Info("Export operation completed successfully",
		map[string]interface{}{
			"correlation_id": correlationID,
			"file_path":      result.FilePath,
			"size_bytes":     result.SizeBytes,
			"item_count":     result.ItemCount,
			"checksum":       result.Checksum,
			"encrypted":      result.Encrypted,
			"duration_ms":    result.Duration.Milliseconds(),
		})

	return result, nil
}

// Import imports data from an encrypted export archive.
// T211: Critical operations logging (start/complete/success/failure).
func (s *ExportService) Import(config *ImportConfig) (*ImportResult, error) {
	startTime := time.Now()
	correlationID := uuid.New().String()

	// Log operation start
	logging.Info("Import operation started",
		map[string]interface{}{
			"correlation_id": correlationID,
			"archive_path":   config.ArchivePath,
			"encrypted":      config.Password != "",
		})

	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "memonexus-import-*")
	if err != nil {
		logging.ErrorWithCode("Import failed: temp directory creation",
			string(errors.ErrInternal), err,
			map[string]interface{}{
				"correlation_id": correlationID,
			})
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Decrypt if needed
	plainArchivePath := filepath.Join(tempDir, "archive.tar.gz")
	if config.Password != "" {
		if err := decryptFile(config.ArchivePath, plainArchivePath, config.Password); err != nil {
			logging.ErrorWithCode("Import failed: decryption",
				string(errors.ErrCryptoFailed), err,
				map[string]interface{}{
					"correlation_id": correlationID,
				})
			return nil, fmt.Errorf("failed to decrypt archive: %w", err)
		}
		defer os.Remove(plainArchivePath)
	} else {
		// Just copy the file
		if err := copyFile(config.ArchivePath, plainArchivePath); err != nil {
			logging.ErrorWithCode("Import failed: archive copy",
				string(errors.ErrInternal), err,
				map[string]interface{}{
					"correlation_id": correlationID,
				})
			return nil, fmt.Errorf("failed to copy archive: %w", err)
		}
		defer os.Remove(plainArchivePath)
	}

	// Extract tar.gz
	if err := extractTarGz(plainArchivePath, tempDir); err != nil {
		logging.ErrorWithCode("Import failed: archive extraction",
			string(errors.ErrInternal), err,
			map[string]interface{}{
				"correlation_id": correlationID,
			})
		return nil, fmt.Errorf("failed to extract archive: %w", err)
	}

	// Read manifest
	manifest, err := s.readManifest(filepath.Join(tempDir, "manifest.json"))
	if err != nil {
		logging.ErrorWithCode("Import failed: manifest read",
			string(errors.ErrInternal), err,
			map[string]interface{}{
				"correlation_id": correlationID,
			})
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	// Verify checksum
	dataFilePath := filepath.Join(tempDir, "data.json")
	if err := verifyChecksum(dataFilePath, manifest.Checksum); err != nil {
		logging.ErrorWithCode("Import failed: checksum verification",
			string(errors.ErrValidation), err,
			map[string]interface{}{
				"correlation_id":    correlationID,
				"expected_checksum": manifest.Checksum,
			})
		return nil, fmt.Errorf("checksum verification failed: %w", err)
	}

	// Import data
	importedCount, skippedCount, err := s.importDataFile(dataFilePath)
	if err != nil {
		logging.ErrorWithCode("Import failed: data import",
			string(errors.ErrInternal), err,
			map[string]interface{}{
				"correlation_id": correlationID,
			})
		return nil, fmt.Errorf("failed to import data: %w", err)
	}

	result := &ImportResult{
		ImportedCount: importedCount,
		SkippedCount:  skippedCount,
		Duration:      time.Since(startTime),
	}

	// Log successful completion
	logging.Info("Import operation completed successfully",
		map[string]interface{}{
			"correlation_id": correlationID,
			"imported_count": result.ImportedCount,
			"skipped_count":  result.SkippedCount,
			"duration_ms":    result.Duration.Milliseconds(),
		})

	return result, nil
}

// createDataFile creates the data JSON file with all items.
func (s *ExportService) createDataFile(path string) (int, string, error) {
	// Get all items
	items, err := s.repo.ListContentItems(100000, 0, "")
	if err != nil {
		return 0, "", err
	}

	// Serialize to JSON
	data, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return 0, "", err
	}

	// Calculate checksum
	hash := sha256.Sum256(data)
	checksum := hex.EncodeToString(hash[:])

	// Write file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return 0, "", err
	}

	return len(items), checksum, nil
}

// writeManifest writes the export manifest.
func (s *ExportService) writeManifest(path string, manifest *ExportManifest) error {
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// readManifest reads and parses the export manifest.
func (s *ExportService) readManifest(path string) (*ExportManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var manifest ExportManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, err
	}

	return &manifest, nil
}

// importDataFile imports items from the data JSON file.
func (s *ExportService) importDataFile(path string) (int, int, error) {
	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, 0, err
	}

	// Parse JSON
	var items []*models.ContentItem
	if err := json.Unmarshal(data, &items); err != nil {
		return 0, 0, err
	}

	importedCount := 0
	skippedCount := 0

	for _, item := range items {
		exists, err := s.itemExists(item.ID)
		if err != nil {
			// Log and continue on error
			skippedCount++
			continue
		}
		if exists {
			skippedCount++
			continue
		}

		if err := s.createItem(item); err != nil {
			// Log and continue on error
			skippedCount++
			continue
		}

		importedCount++
	}

	return importedCount, skippedCount, nil
}

// itemExists checks if an item already exists.
func (s *ExportService) itemExists(id models.UUID) (bool, error) {
	_, err := s.repo.GetContentItem(string(id))
	if err == nil {
		return true, nil
	}
	// Check if it's a "not found" error (sql.ErrNoRows)
	// All other errors should be propagated
	if stderrors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return false, err
}

// createItem creates a new item.
func (s *ExportService) createItem(item *models.ContentItem) error {
	return s.repo.CreateContentItem(item)
}

// verifyChecksum verifies the SHA-256 checksum of a file.
func verifyChecksum(path, expectedChecksum string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file for checksum verification: %w", err)
	}

	hash := sha256.Sum256(data)
	actualChecksum := hex.EncodeToString(hash[:])

	if actualChecksum != expectedChecksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum)
	}

	return nil
}

// writeTarGz creates a tar.gz archive from a directory.
func writeTarGz(sourceDir, targetPath string) error {
	outFile, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	gzw := gzip.NewWriter(outFile)
	defer gzw.Close()

	tw := tar.NewWriter(gzw)
	defer tw.Close()

	return filepath.Walk(sourceDir, func(file string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if fi.IsDir() {
			return nil
		}

		header, err := tar.FileInfoHeader(fi, fi.Name())
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(sourceDir, file)
		if err != nil {
			return err
		}
		header.Name = relPath

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		data, err := os.ReadFile(file)
		if err != nil {
			return err
		}

		if _, err := tw.Write(data); err != nil {
			return err
		}

		return nil
	})
}

// extractTarGz extracts a tar.gz archive to a directory.
func extractTarGz(archivePath, targetDir string) error {
	inFile, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer inFile.Close()

	gzr, err := gzip.NewReader(inFile)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		targetPath := filepath.Join(targetDir, header.Name)

		if header.Typeflag == tar.TypeDir {
			if err := os.MkdirAll(targetPath, 0755); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return err
		}

		outFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
		if err != nil {
			return err
		}

		if _, err := io.Copy(outFile, tr); err != nil {
			outFile.Close()
			return err
		}
		outFile.Close()
	}

	return nil
}

// encryptFile encrypts a file using AES-256-GCM with PBKDF2 key derivation.
func encryptFile(srcPath, dstPath, password string) (int64, error) {
	// Read plaintext
	plaintext, err := os.ReadFile(srcPath)
	if err != nil {
		return 0, fmt.Errorf("failed to read source file: %w", err)
	}

	// Generate salt
	salt := make([]byte, aes.BlockSize)
	if _, err := rand.Read(salt); err != nil {
		return 0, fmt.Errorf("failed to generate salt: %w", err)
	}

	// Derive key using PBKDF2
	key := deriveKey(password, salt)

	// Create cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return 0, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return 0, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return 0, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt
	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	// Write output: salt + nonce + ciphertext
	outFile, err := os.Create(dstPath)
	if err != nil {
		return 0, fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	if _, err := outFile.Write(salt); err != nil {
		return 0, fmt.Errorf("failed to write salt: %w", err)
	}
	if _, err := outFile.Write(nonce); err != nil {
		return 0, fmt.Errorf("failed to write nonce: %w", err)
	}
	if _, err := outFile.Write(ciphertext); err != nil {
		return 0, fmt.Errorf("failed to write ciphertext: %w", err)
	}

	info, err := outFile.Stat()
	if err != nil {
		return 0, fmt.Errorf("failed to stat output file: %w", err)
	}

	return info.Size(), nil
}

// decryptFile decrypts a file that was encrypted with encryptFile.
func decryptFile(srcPath, dstPath, password string) error {
	inFile, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open encrypted file: %w", err)
	}
	defer inFile.Close()

	// Read salt
	salt := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(inFile, salt); err != nil {
		return fmt.Errorf("failed to read salt: %w", err)
	}

	// Read nonce
	nonce := make([]byte, 12) // Read to determine size
	if _, err := io.ReadFull(inFile, nonce[:8]); err != nil {
		return fmt.Errorf("failed to read nonce (part 1): %w", err)
	}
	if _, err := io.ReadFull(inFile, nonce[8:]); err != nil {
		return fmt.Errorf("failed to read nonce (part 2): %w", err)
	}

	// Read remaining ciphertext
	ciphertext, err := io.ReadAll(inFile)
	if err != nil {
		return fmt.Errorf("failed to read ciphertext: %w", err)
	}

	// Derive key
	key := deriveKey(password, salt)

	// Create cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("failed to create GCM: %w", err)
	}

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return fmt.Errorf("decryption failed: %w", err)
	}

	// Write output
	if err := os.WriteFile(dstPath, plaintext, 0644); err != nil {
		return fmt.Errorf("failed to write decrypted file: %w", err)
	}

	return nil
}

// deriveKey derives an encryption key from password and salt using PBKDF2.
func deriveKey(password string, salt []byte) []byte {
	// PBKDF2 with SHA-256, 100,000 iterations, 32-byte key
	return pbkdf2.Key([]byte(password), salt, 100000, 32, sha256.New)
}

// copyFile copies a file from src to dst.
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

// ArchiveInfo represents metadata about an export archive.
type ArchiveInfo struct {
	ID        string    `json:"id"`
	FilePath  string    `json:"file_path"`
	Checksum  string    `json:"checksum"`
	SizeBytes int64     `json:"size_bytes"`
	ItemCount int       `json:"item_count"`
	CreatedAt time.Time `json:"created_at"`
	Encrypted bool      `json:"encrypted"`
}

// ListArchives returns all export archives in the exports directory.
func (s *ExportService) ListArchives(exportDir string) ([]*ArchiveInfo, error) {
	var archives []*ArchiveInfo

	// Ensure directory exists
	if _, err := os.Stat(exportDir); os.IsNotExist(err) {
		return archives, nil
	}

	// Walk directory
	err := filepath.Walk(exportDir, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-tar.gz files
		if fi.IsDir() || filepath.Ext(path) != ".gz" {
			return nil
		}

		// Try to read manifest from archive for metadata
		info := &ArchiveInfo{
			ID:        filepath.Base(path),
			FilePath:  path,
			SizeBytes: fi.Size(),
			CreatedAt: fi.ModTime(),
		}

		// Attempt to read checksum from filename or manifest
		// For now, we'll use the filename as a basic identifier
		archives = append(archives, info)
		return nil
	})

	return archives, err
}

// DeleteArchive removes an export archive file.
func (s *ExportService) DeleteArchive(archivePath string) error {
	// Verify file exists
	if _, err := os.Stat(archivePath); os.IsNotExist(err) {
		return fmt.Errorf("archive not found: %s", archivePath)
	}

	// Delete the file
	if err := os.Remove(archivePath); err != nil {
		return fmt.Errorf("failed to delete archive: %w", err)
	}

	return nil
}

// ApplyRetentionPolicy removes old archives according to retention policy.
// Keeps the most recent `retentionCount` archives and deletes older ones.
func (s *ExportService) ApplyRetentionPolicy(exportDir string, retentionCount int) (int, error) {
	if retentionCount <= 0 {
		return 0, nil // No retention limit
	}

	// List all archives
	archives, err := s.ListArchives(exportDir)
	if err != nil {
		return 0, fmt.Errorf("failed to list archives: %w", err)
	}

	// Sort by creation time (oldest first)
	sort.Slice(archives, func(i, j int) bool {
		return archives[i].CreatedAt.Before(archives[j].CreatedAt)
	})

	// Determine how many to delete
	deleteCount := len(archives) - retentionCount
	if deleteCount <= 0 {
		return 0, nil // No archives to delete
	}

	// Delete oldest archives
	deleted := 0
	for i := 0; i < deleteCount; i++ {
		if err := s.DeleteArchive(archives[i].FilePath); err != nil {
			// Log error but continue with other deletions
			continue
		}
		deleted++
	}

	return deleted, nil
}

// ErrNotFound is returned when an item is not found.
var ErrNotFound = stderrors.New("item not found")
