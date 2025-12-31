// Package storage provides content-addressed storage for media files.
// T158: SHA-256 content addressing for media file deduplication.
package storage

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
)

// ContentAddressedStorage stores files by their content hash (SHA-256).
// This enables deduplication - identical files are stored only once.
type ContentAddressedStorage struct {
	baseDir string
}

// NewContentAddressedStorage creates a new ContentAddressedStorage.
func NewContentAddressedStorage(baseDir string) *ContentAddressedStorage {
	return &ContentAddressedStorage{
		baseDir: baseDir,
	}
}

// CalculateHash calculates SHA-256 hash of file content.
func CalculateHash(data []byte) string {
	h := sha256.New()
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

// CalculateHashFromReader calculates SHA-256 hash from an io.Reader.
func CalculateHashFromReader(r io.Reader) (string, error) {
	h := sha256.New()
	if _, err := io.Copy(h, r); err != nil {
		return "", fmt.Errorf("failed to calculate hash: %w", err)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// CalculateHashFromFile calculates SHA-256 hash of a file.
func CalculateHashFromFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	return CalculateHashFromReader(file)
}

// Store stores data and returns its content hash.
// The file is stored at baseDir/{hash}[0:2]/{hash}[2:4]/{hash}.
// This creates a two-level directory structure for better performance.
func (s *ContentAddressedStorage) Store(data []byte) (string, error) {
	hash := CalculateHash(data)

	// Create directory structure: baseDir/{hash[0:2]}/{hash[2:4]}/
	dir := filepath.Join(s.baseDir, hash[0:2], hash[2:4])
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	// Store file at baseDir/{hash[0:2]}/{hash[2:4]}/{hash}
	filePath := filepath.Join(dir, hash)

	// Check if file already exists (deduplication)
	if _, err := os.Stat(filePath); err == nil {
		// File exists, return hash without rewriting
		return hash, nil
	}

	// Write file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return hash, nil
}

// StoreFile stores a file from disk and returns its content hash.
// The source file is not modified.
func (s *ContentAddressedStorage) StoreFile(sourcePath string) (string, error) {
	// Calculate hash first
	hash, err := CalculateHashFromFile(sourcePath)
	if err != nil {
		return "", fmt.Errorf("failed to calculate hash: %w", err)
	}

	// Create directory structure
	dir := filepath.Join(s.baseDir, hash[0:2], hash[2:4])
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	// Destination path
	destPath := filepath.Join(dir, hash)

	// Check if file already exists
	if _, err := os.Stat(destPath); err == nil {
		// File exists (deduplication)
		return hash, nil
	}

	// Copy file
	srcFile, err := os.Open(sourcePath)
	if err != nil {
		return "", fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	destFile, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	// Copy content
	if _, err := io.Copy(destFile, srcFile); err != nil {
		return "", fmt.Errorf("failed to copy file: %w", err)
	}

	return hash, nil
}

// Retrieve retrieves data by content hash.
// Returns os.ErrNotExist if the hash is not found.
func (s *ContentAddressedStorage) Retrieve(hash string) ([]byte, error) {
	filePath := s.getPath(hash)

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("content not found: %w", err)
		}
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Verify hash matches
	calculatedHash := CalculateHash(data)
	if calculatedHash != hash {
		return nil, fmt.Errorf("hash mismatch: expected %s, got %s", hash, calculatedHash)
	}

	return data, nil
}

// RetrieveToFile retrieves content by hash and writes it to a file.
func (s *ContentAddressedStorage) RetrieveToFile(hash, destPath string) error {
	filePath := s.getPath(hash)

	// Open source file
	srcFile, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("content not found: %w", err)
		}
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	// Create destination file
	destFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	// Copy content
	if _, err := io.Copy(destFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	return nil
}

// Delete removes stored content by hash.
func (s *ContentAddressedStorage) Delete(hash string) error {
	filePath := s.getPath(hash)

	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			return nil // Already deleted
		}
		return fmt.Errorf("failed to delete file: %w", err)
	}

	// Try to remove empty directories
	dir := filepath.Dir(filePath)
	os.Remove(dir) // Ignore error

	parentDir := filepath.Dir(dir)
	os.Remove(parentDir) // Ignore error

	return nil
}

// Exists checks if content exists for a given hash.
func (s *ContentAddressedStorage) Exists(hash string) bool {
	filePath := s.getPath(hash)
	_, err := os.Stat(filePath)
	return err == nil
}

// GetPath returns the file system path for a given hash.
// This is useful for external access to the stored files.
func (s *ContentAddressedStorage) GetPath(hash string) string {
	return s.getPath(hash)
}

// getPath calculates the file path for a hash.
func (s *ContentAddressedStorage) getPath(hash string) string {
	return filepath.Join(s.baseDir, hash[0:2], hash[2:4], hash)
}

// Size returns the size of stored content in bytes.
func (s *ContentAddressedStorage) Size(hash string) (int64, error) {
	filePath := s.getPath(hash)

	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, fmt.Errorf("content not found: %w", err)
		}
		return 0, fmt.Errorf("failed to stat file: %w", err)
	}

	return info.Size(), nil
}

// ListAll lists all stored content hashes.
// This is a potentially expensive operation for large storage.
func (s *ContentAddressedStorage) ListAll() ([]string, error) {
	var hashes []string

	err := filepath.Walk(s.baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			// Don't recurse into the top-level directory
			if path == s.baseDir {
				return nil
			}
			return nil
		}

		// Extract hash from path
		relPath, err := filepath.Rel(s.baseDir, path)
		if err != nil {
			return err
		}

		// Path format: {hash[0:2]}/{hash[2:4]}/{hash}
		// Reconstruct hash
		parts := filepath.SplitList(relPath)
		if len(parts) == 3 {
			hash := parts[0] + parts[1] + parts[2]
			if len(hash) == 64 { // SHA-256 hex length
				hashes = append(hashes, hash)
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk storage: %w", err)
	}

	return hashes, nil
}

// VerifyAll verifies all stored content by recalculating hashes.
// Returns a list of corrupted hashes.
func (s *ContentAddressedStorage) VerifyAll() ([]string, error) {
	hashes, err := s.ListAll()
	if err != nil {
		return nil, err
	}

	var corrupted []string

	for _, hash := range hashes {
		data, err := s.Retrieve(hash)
		if err != nil {
			corrupted = append(corrupted, hash)
			continue
		}

		// Re-verify hash
		calculatedHash := CalculateHash(data)
		if calculatedHash != hash {
			corrupted = append(corrupted, hash)
		}
	}

	return corrupted, nil
}

// DuplicateFinder helps identify duplicate files.
type DuplicateFinder struct {
	hashes map[string][]string // hash -> list of original file paths
}

// NewDuplicateFinder creates a new DuplicateFinder.
func NewDuplicateFinder() *DuplicateFinder {
	return &DuplicateFinder{
		hashes: make(map[string][]string),
	}
}

// AddFile adds a file and calculates its hash.
func (d *DuplicateFinder) AddFile(filePath string) (string, error) {
	hash, err := CalculateHashFromFile(filePath)
	if err != nil {
		return "", err
	}

	d.hashes[hash] = append(d.hashes[hash], filePath)

	return hash, nil
}

// GetDuplicates returns all duplicate file groups.
// Each group contains files with identical content.
func (d *DuplicateFinder) GetDuplicates() map[string][]string {
	duplicates := make(map[string][]string)

	for hash, files := range d.hashes {
		if len(files) > 1 {
			duplicates[hash] = files
		}
	}

	return duplicates
}

// GetUniqueCount returns the number of unique files (deduplicated count).
func (d *DuplicateFinder) GetUniqueCount() int {
	return len(d.hashes)
}

// GetTotalCount returns the total number of files (including duplicates).
func (d *DuplicateFinder) GetTotalCount() int {
	count := 0
	for _, files := range d.hashes {
		count += len(files)
	}
	return count
}

// CalculateSavings calculates storage savings from deduplication.
// Returns (uniqueSize, totalSize, savedSize).
func (d *DuplicateFinder) CalculateSavings() (int64, int64, int64) {
	uniqueSizes := make(map[string]int64)

	// First pass: get unique file sizes
	for hash, files := range d.hashes {
		if len(files) > 0 {
			info, err := os.Stat(files[0])
			if err == nil {
				uniqueSizes[hash] = info.Size()
			}
		}
	}

	// Calculate unique size
	var uniqueSize int64
	for _, size := range uniqueSizes {
		uniqueSize += size
	}

	// Calculate total size (including duplicates)
	var totalSize int64
	for hash, files := range d.hashes {
		if size, ok := uniqueSizes[hash]; ok {
			totalSize += size * int64(len(files))
		}
	}

	savedSize := totalSize - uniqueSize

	return uniqueSize, totalSize, savedSize
}

// StreamingHash wraps a writer and calculates hash as data is written.
type StreamingHash struct {
	hash   hash.Hash
	writer io.Writer
}

// NewStreamingHash creates a new StreamingHash.
func NewStreamingHash(writer io.Writer) *StreamingHash {
	return &StreamingHash{
		hash:   sha256.New(),
		writer: writer,
	}
}

// Write writes data and updates hash.
func (s *StreamingHash) Write(p []byte) (int, error) {
	s.hash.Write(p)
	return s.writer.Write(p)
}

// Hash returns the calculated hash.
func (s *StreamingHash) Hash() string {
	return hex.EncodeToString(s.hash.Sum(nil))
}

// MultiHashReader wraps a reader and calculates hash as data is read.
type MultiHashReader struct {
	reader io.Reader
	hash   hash.Hash
}

// NewMultiHashReader creates a new MultiHashReader.
func NewMultiHashReader(reader io.Reader) *MultiHashReader {
	return &MultiHashReader{
		reader: reader,
		hash:   sha256.New(),
	}
}

// Read reads data and updates hash.
func (m *MultiHashReader) Read(p []byte) (int, error) {
	n, err := m.reader.Read(p)
	if n > 0 {
		m.hash.Write(p[:n])
	}
	return n, err
}

// Hash returns the calculated hash.
func (m *MultiHashReader) Hash() string {
	return hex.EncodeToString(m.hash.Sum(nil))
}
