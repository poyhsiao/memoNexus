// Package storage provides file storage with content-addressed storage.
package storage

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// StorageManager handles file storage with SHA-256 content addressing.
type StorageManager struct {
	// Base directory for storing media files
	baseDir string
}

// NewStorageManager creates a new StorageManager.
func NewStorageManager(baseDir string) (*StorageManager, error) {
	// Create base directory if it doesn't exist
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	return &StorageManager{
		baseDir: baseDir,
	}, nil
}

// StoreFile stores a file with content addressing (SHA-256).
// Returns the content hash (used as file path) and any error.
func (s *StorageManager) StoreFile(r io.Reader) (string, int64, error) {
	// First pass: read data and calculate hash
	hasher := sha256.New()
	var size int64

	// Create a temporary file to buffer the data
	tmpFile, err := os.CreateTemp("", "storage-*")
	if err != nil {
		return "", 0, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Tee reader: writes to both temp file and hash
	multiWriter := io.MultiWriter(tmpFile, hasher)
	size, err = io.Copy(multiWriter, r)
	if err != nil {
		// Enhanced error handling for corrupted/unreadable files
		return "", 0, fmt.Errorf("failed to read file data: file may be corrupted or incomplete: %w", err)
	}

	// Validate file size (detect empty or suspiciously small files)
	if size == 0 {
		return "", 0, fmt.Errorf("invalid file: empty file (0 bytes)")
	}
	if size < 16 {
		// Very small files are likely corrupted or invalid
		return "", 0, fmt.Errorf("invalid file: too small (%d bytes), may be corrupted", size)
	}

	// Get SHA-256 hash
	contentHash := hex.EncodeToString(hasher.Sum(nil))

	// Create content-addressed path: baseDir/XX/XXXX...
	// First two characters as directory prefix
	prefix := contentHash[:2]
	dirPath := filepath.Join(s.baseDir, prefix)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return "", 0, fmt.Errorf("failed to create storage directory: %w", err)
	}

	// Final file path
	filePath := filepath.Join(dirPath, contentHash)

	// Check if file already exists (deduplication)
	if _, err := os.Stat(filePath); err == nil {
		// File already exists, return existing hash
		return contentHash, size, nil
	}

	// Move temp file to final location
	if err := os.Rename(tmpFile.Name(), filePath); err != nil {
		return "", 0, fmt.Errorf("failed to move file to storage: %w", err)
	}

	return contentHash, size, nil
}

// RetrieveFile retrieves a file by its content hash.
// Returns an io.ReadCloser that must be closed by the caller.
func (s *StorageManager) RetrieveFile(contentHash string) (io.ReadCloser, error) {
	if len(contentHash) != 64 {
		return nil, fmt.Errorf("invalid content hash length: %d", len(contentHash))
	}

	// Build file path
	prefix := contentHash[:2]
	filePath := filepath.Join(s.baseDir, prefix, contentHash)

	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	return file, nil
}

// FileExists checks if a file with the given content hash exists.
func (s *StorageManager) FileExists(contentHash string) (bool, error) {
	if len(contentHash) != 64 {
		return false, fmt.Errorf("invalid content hash length: %d", len(contentHash))
	}

	prefix := contentHash[:2]
	filePath := filepath.Join(s.baseDir, prefix, contentHash)

	_, err := os.Stat(filePath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// DeleteFile deletes a file by its content hash.
func (s *StorageManager) DeleteFile(contentHash string) error {
	if len(contentHash) != 64 {
		return fmt.Errorf("invalid content hash length: %d", len(contentHash))
	}

	prefix := contentHash[:2]
	filePath := filepath.Join(s.baseDir, prefix, contentHash)

	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	// Try to remove empty prefix directory
	dirPath := filepath.Join(s.baseDir, prefix)
	os.Remove(dirPath)

	return nil
}

// GetFilePath returns the file system path for a content hash.
// This is useful for serving files directly.
func (s *StorageManager) GetFilePath(contentHash string) (string, error) {
	if len(contentHash) != 64 {
		return "", fmt.Errorf("invalid content hash length: %d", len(contentHash))
	}

	prefix := contentHash[:2]
	filePath := filepath.Join(s.baseDir, prefix, contentHash)

	// Verify file exists
	if _, err := os.Stat(filePath); err != nil {
		return "", fmt.Errorf("file not found: %w", err)
	}

	return filePath, nil
}

// CalculateHash calculates SHA-256 hash of an io.Reader.
func CalculateHash(r io.Reader) (string, error) {
	hasher := sha256.New()

	if _, err := io.Copy(hasher, r); err != nil {
		return "", fmt.Errorf("failed to calculate hash: %w", err)
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// CalculateHashOfFile calculates SHA-256 hash of a file.
func CalculateHashOfFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	return CalculateHash(file)
}

// StorageStats provides statistics about storage usage.
type StorageStats struct {
	TotalFiles int
	TotalSize  int64
	ByPrefix   map[string]int // Files per prefix directory
}

// GetStats returns storage statistics.
func (s *StorageManager) GetStats() (*StorageStats, error) {
	stats := &StorageStats{
		ByPrefix: make(map[string]int),
	}

	// Walk through all prefix directories
	entries, err := os.ReadDir(s.baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read base directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		prefix := entry.Name()
		if len(prefix) != 2 {
			continue
		}

		// Count files in prefix directory
		dirPath := filepath.Join(s.baseDir, prefix)
		files, err := os.ReadDir(dirPath)
		if err != nil {
			continue
		}

		for _, file := range files {
			if file.IsDir() {
				continue
			}

			stats.TotalFiles++
			stats.ByPrefix[prefix]++

			// Add file size
			info, err := file.Info()
			if err == nil {
				stats.TotalSize += info.Size()
			}
		}
	}

	return stats, nil
}

// Cleanup removes files that are not referenced in the provided set.
// The safeHashes parameter contains all hashes that should be kept.
func (s *StorageManager) Cleanup(safeHashes map[string]bool) (removed int, freedBytes int64, err error) {
	// Build a set of safe hashes for fast lookup
	safeSet := make(map[string]bool)
	for hash := range safeHashes {
		safeSet[hash] = true
	}

	// Walk through all files and remove unreferenced ones
	entries, err := os.ReadDir(s.baseDir)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to read base directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		prefix := entry.Name()
		if len(prefix) != 2 {
			continue
		}

		dirPath := filepath.Join(s.baseDir, prefix)
		files, err := os.ReadDir(dirPath)
		if err != nil {
			continue
		}

		for _, file := range files {
			if file.IsDir() {
				continue
			}

			hash := file.Name()
			if len(hash) != 64 {
				continue
			}

			// Check if hash is in safe set
			if !safeSet[hash] {
				filePath := filepath.Join(dirPath, hash)

				// Get file size before removal
				info, err := file.Info()
				var size int64
				if err == nil {
					size = info.Size()
				}

				// Remove file
				if err := os.Remove(filePath); err == nil {
					removed++
					freedBytes += size
				}
			}
		}

		// Remove empty prefix directory
		files, _ = os.ReadDir(dirPath)
		if len(files) == 0 {
			os.Remove(dirPath)
		}
	}

	return removed, freedBytes, nil
}
