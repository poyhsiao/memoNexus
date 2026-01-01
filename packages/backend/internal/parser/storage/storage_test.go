// Package storage tests for file storage with content addressing.
package storage

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestNewStorageManager verifies storage manager creation.
func TestNewStorageManager(t *testing.T) {
	// Create temp directory for testing
	tempDir := t.TempDir()

	manager, err := NewStorageManager(tempDir)
	if err != nil {
		t.Fatalf("NewStorageManager() error = %v", err)
	}

	if manager == nil {
		t.Fatal("NewStorageManager() returned nil")
	}

	if manager.baseDir != tempDir {
		t.Errorf("baseDir = %q, want %q", manager.baseDir, tempDir)
	}

	// Verify directory was created
	if _, err := os.Stat(tempDir); err != nil {
		t.Errorf("Base directory was not created: %v", err)
	}
}

// TestNewStorageManager_createsDirectory verifies directory creation.
func TestNewStorageManager_createsDirectory(t *testing.T) {
	tempDir := t.TempDir()
	subDir := filepath.Join(tempDir, "storage", "files")

	manager, err := NewStorageManager(subDir)
	if err != nil {
		t.Fatalf("NewStorageManager() error = %v", err)
	}

	// Verify nested directory was created
	if _, err := os.Stat(subDir); err != nil {
		t.Errorf("Nested directory was not created: %v", err)
	}

	if manager.baseDir != subDir {
		t.Errorf("baseDir = %q, want %q", manager.baseDir, subDir)
	}
}

// TestNewStorageManager_mkdirError verifies error handling when directory creation fails.
func TestNewStorageManager_mkdirError(t *testing.T) {
	// Use an invalid path that should fail
	invalidPath := "/dev/null/invalid/path/that/cannot/be/created"

	_, err := NewStorageManager(invalidPath)
	if err == nil {
		t.Error("NewStorageManager() with invalid path should return error")
	}
}

// TestStoreFile_success verifies successful file storage.
func TestStoreFile_success(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewStorageManager(tempDir)
	if err != nil {
		t.Fatalf("NewStorageManager() error = %v", err)
	}

	content := []byte("test content for storage - at least 16 bytes")
	hash, size, err := manager.StoreFile(bytes.NewReader(content))

	if err != nil {
		t.Fatalf("StoreFile() error = %v", err)
	}

	if hash == "" {
		t.Error("StoreFile() returned empty hash")
	}

	if len(hash) != 64 { // SHA-256 hex is 64 characters
		t.Errorf("Hash length = %d, want 64", len(hash))
	}

	if size != int64(len(content)) {
		t.Errorf("Size = %d, want %d", size, len(content))
	}

	// Verify file was stored
	prefix := hash[:2]
	filePath := filepath.Join(tempDir, prefix, hash)
	if _, err := os.Stat(filePath); err != nil {
		t.Errorf("File was not stored at %s: %v", filePath, err)
	}
}

// TestStoreFile_empty verifies error handling for empty file.
func TestStoreFile_empty(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewStorageManager(tempDir)
	if err != nil {
		t.Fatalf("NewStorageManager() error = %v", err)
	}

	_, _, err = manager.StoreFile(bytes.NewReader([]byte{}))

	if err == nil {
		t.Error("StoreFile() with empty content should return error")
	}

	if !strings.Contains(err.Error(), "empty file") {
		t.Errorf("Error should mention 'empty file', got: %v", err)
	}
}

// TestStoreFile_tooSmall verifies error handling for very small files.
func TestStoreFile_tooSmall(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewStorageManager(tempDir)
	if err != nil {
		t.Fatalf("NewStorageManager() error = %v", err)
	}

	_, _, err = manager.StoreFile(bytes.NewReader([]byte("small"))) // 5 bytes

	if err == nil {
		t.Error("StoreFile() with very small content should return error")
	}

	if !strings.Contains(err.Error(), "too small") {
		t.Errorf("Error should mention 'too small', got: %v", err)
	}
}

// TestStoreFile_deduplication verifies content addressing deduplication.
func TestStoreFile_deduplication(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewStorageManager(tempDir)
	if err != nil {
		t.Fatalf("NewStorageManager() error = %v", err)
	}

	content := []byte("test content for deduplication")

	// Store the same content twice
	hash1, size1, err1 := manager.StoreFile(bytes.NewReader(content))
	hash2, size2, err2 := manager.StoreFile(bytes.NewReader(content))

	if err1 != nil {
		t.Fatalf("First StoreFile() error = %v", err1)
	}
	if err2 != nil {
		t.Fatalf("Second StoreFile() error = %v", err2)
	}

	// Should return same hash (content addressing)
	if hash1 != hash2 {
		t.Errorf("Hash mismatch: first = %q, second = %q", hash1, hash2)
	}

	if size1 != size2 {
		t.Errorf("Size mismatch: first = %d, second = %d", size1, size2)
	}

	// Verify only one file exists
	prefix := hash1[:2]
	entries, err := os.ReadDir(filepath.Join(tempDir, prefix))
	if err != nil {
		t.Fatalf("Failed to read directory: %v", err)
	}

	if len(entries) != 1 {
		t.Errorf("Expected 1 file, found %d", len(entries))
	}
}

// TestStoreFile_readError verifies error handling for read failures.
func TestStoreFile_readError(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewStorageManager(tempDir)
	if err != nil {
		t.Fatalf("NewStorageManager() error = %v", err)
	}

	// Create a reader that always fails
	errReader := &errorReader{err: io.ErrUnexpectedEOF}

	_, _, err = manager.StoreFile(errReader)

	if err == nil {
		t.Error("StoreFile() with failing reader should return error")
	}
}

// TestRetrieveFile_success verifies successful file retrieval.
func TestRetrieveFile_success(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewStorageManager(tempDir)
	if err != nil {
		t.Fatalf("NewStorageManager() error = %v", err)
	}

	content := []byte("test content for retrieval")
	hash, _, err := manager.StoreFile(bytes.NewReader(content))
	if err != nil {
		t.Fatalf("StoreFile() error = %v", err)
	}

	reader, err := manager.RetrieveFile(hash)
	if err != nil {
		t.Fatalf("RetrieveFile() error = %v", err)
	}
	defer reader.Close()

	retrieved, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("Failed to read retrieved file: %v", err)
	}

	if string(retrieved) != string(content) {
		t.Errorf("Retrieved content = %q, want %q", string(retrieved), string(content))
	}
}

// TestRetrieveFile_invalidHash verifies error handling for invalid hash.
func TestRetrieveFile_invalidHash(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewStorageManager(tempDir)
	if err != nil {
		t.Fatalf("NewStorageManager() error = %v", err)
	}

	_, err = manager.RetrieveFile("invalid-hash")

	if err == nil {
		t.Error("RetrieveFile() with invalid hash should return error")
	}

	if !strings.Contains(err.Error(), "invalid content hash length") {
		t.Errorf("Error should mention 'invalid content hash length', got: %v", err)
	}
}

// TestRetrieveFile_notFound verifies error handling for missing file.
func TestRetrieveFile_notFound(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewStorageManager(tempDir)
	if err != nil {
		t.Fatalf("NewStorageManager() error = %v", err)
	}

	// A valid SHA-256 hash (64 hex characters)
	validHash := strings.Repeat("a", 64)

	_, err = manager.RetrieveFile(validHash)

	if err == nil {
		t.Error("RetrieveFile() with non-existent hash should return error")
	}

	if !strings.Contains(err.Error(), "failed to open file") {
		t.Errorf("Error should mention 'failed to open file', got: %v", err)
	}
}

// TestFileExists_success verifies file existence check.
func TestFileExists_success(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewStorageManager(tempDir)
	if err != nil {
		t.Fatalf("NewStorageManager() error = %v", err)
	}

	content := []byte("test content - more than 16 bytes")
	hash, _, err := manager.StoreFile(bytes.NewReader(content))
	if err != nil {
		t.Fatalf("StoreFile() error = %v", err)
	}

	exists, err := manager.FileExists(hash)
	if err != nil {
		t.Fatalf("FileExists() error = %v", err)
	}

	if !exists {
		t.Error("FileExists() should return true for stored file")
	}
}

// TestFileExists_notFound verifies file existence check for missing file.
func TestFileExists_notFound(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewStorageManager(tempDir)
	if err != nil {
		t.Fatalf("NewStorageManager() error = %v", err)
	}

	validHash := strings.Repeat("b", 64)

	exists, err := manager.FileExists(validHash)
	if err != nil {
		t.Fatalf("FileExists() error = %v", err)
	}

	if exists {
		t.Error("FileExists() should return false for non-existent file")
	}
}

// TestFileExists_invalidHash verifies error handling for invalid hash.
func TestFileExists_invalidHash(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewStorageManager(tempDir)
	if err != nil {
		t.Fatalf("NewStorageManager() error = %v", err)
	}

	_, err = manager.FileExists("invalid")

	if err == nil {
		t.Error("FileExists() with invalid hash should return error")
	}
}

// TestDeleteFile_success verifies successful file deletion.
func TestDeleteFile_success(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewStorageManager(tempDir)
	if err != nil {
		t.Fatalf("NewStorageManager() error = %v", err)
	}

	content := []byte("test content - more than 16 bytes")
	hash, _, err := manager.StoreFile(bytes.NewReader(content))
	if err != nil {
		t.Fatalf("StoreFile() error = %v", err)
	}

	err = manager.DeleteFile(hash)
	if err != nil {
		t.Fatalf("DeleteFile() error = %v", err)
	}

	// Verify file was deleted
	exists, _ := manager.FileExists(hash)
	if exists {
		t.Error("File should be deleted")
	}
}

// TestDeleteFile_notFound verifies deleting non-existent file is OK.
func TestDeleteFile_notFound(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewStorageManager(tempDir)
	if err != nil {
		t.Fatalf("NewStorageManager() error = %v", err)
	}

	validHash := strings.Repeat("c", 64)

	// Deleting non-existent file should not error
	err = manager.DeleteFile(validHash)
	if err != nil {
		t.Errorf("DeleteFile() with non-existent file should not error, got: %v", err)
	}
}

// TestDeleteFile_invalidHash verifies error handling for invalid hash.
func TestDeleteFile_invalidHash(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewStorageManager(tempDir)
	if err != nil {
		t.Fatalf("NewStorageManager() error = %v", err)
	}

	err = manager.DeleteFile("invalid")

	if err == nil {
		t.Error("DeleteFile() with invalid hash should return error")
	}
}

// TestGetFilePath_success verifies getting file path.
func TestGetFilePath_success(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewStorageManager(tempDir)
	if err != nil {
		t.Fatalf("NewStorageManager() error = %v", err)
	}

	content := []byte("test content - more than 16 bytes")
	hash, _, err := manager.StoreFile(bytes.NewReader(content))
	if err != nil {
		t.Fatalf("StoreFile() error = %v", err)
	}

	filePath, err := manager.GetFilePath(hash)
	if err != nil {
		t.Fatalf("GetFilePath() error = %v", err)
	}

	if filePath == "" {
		t.Error("GetFilePath() returned empty path")
	}

	// Verify file exists at returned path
	if _, err := os.Stat(filePath); err != nil {
		t.Errorf("File doesn't exist at returned path %s: %v", filePath, err)
	}
}

// TestGetFilePath_notFound verifies error for missing file.
func TestGetFilePath_notFound(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewStorageManager(tempDir)
	if err != nil {
		t.Fatalf("NewStorageManager() error = %v", err)
	}

	validHash := strings.Repeat("d", 64)

	_, err = manager.GetFilePath(validHash)

	if err == nil {
		t.Error("GetFilePath() with non-existent file should return error")
	}

	if !strings.Contains(err.Error(), "file not found") {
		t.Errorf("Error should mention 'file not found', got: %v", err)
	}
}

// TestGetFilePath_invalidHash verifies error handling for invalid hash.
func TestGetFilePath_invalidHash(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewStorageManager(tempDir)
	if err != nil {
		t.Fatalf("NewStorageManager() error = %v", err)
	}

	_, err = manager.GetFilePath("invalid")

	if err == nil {
		t.Error("GetFilePath() with invalid hash should return error")
	}
}

// TestCalculateHash verifies hash calculation.
func TestCalculateHash(t *testing.T) {
	content := []byte("test content for hashing")

	hash, err := CalculateHash(bytes.NewReader(content))
	if err != nil {
		t.Fatalf("CalculateHash() error = %v", err)
	}

	if len(hash) != 64 {
		t.Errorf("Hash length = %d, want 64", len(hash))
	}

	// Verify hash is deterministic
	hash2, err := CalculateHash(bytes.NewReader(content))
	if err != nil {
		t.Fatalf("CalculateHash() second call error = %v", err)
	}

	if hash != hash2 {
		t.Errorf("Hash should be deterministic: first = %q, second = %q", hash, hash2)
	}
}

// TestCalculateHash_empty verifies empty content handling.
func TestCalculateHash_empty(t *testing.T) {
	hash, err := CalculateHash(bytes.NewReader([]byte{}))

	if err != nil {
		t.Fatalf("CalculateHash() with empty content error = %v", err)
	}

	// SHA-256 of empty string is known value
	expected := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	if hash != expected {
		t.Errorf("Hash of empty string = %q, want %q", hash, expected)
	}
}

// TestCalculateHash_readError verifies error handling for read failures.
func TestCalculateHash_readError(t *testing.T) {
	errReader := &errorReader{err: io.ErrUnexpectedEOF}

	_, err := CalculateHash(errReader)

	if err == nil {
		t.Error("CalculateHash() with failing reader should return error")
	}
}

// TestGetStats_emptyStorage verifies stats for empty storage.
func TestGetStats_emptyStorage(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewStorageManager(tempDir)
	if err != nil {
		t.Fatalf("NewStorageManager() error = %v", err)
	}

	stats, err := manager.GetStats()
	if err != nil {
		t.Fatalf("GetStats() error = %v", err)
	}

	if stats.TotalFiles != 0 {
		t.Errorf("TotalFiles = %d, want 0", stats.TotalFiles)
	}

	if stats.TotalSize != 0 {
		t.Errorf("TotalSize = %d, want 0", stats.TotalSize)
	}

	if stats.ByPrefix == nil {
		t.Error("ByPrefix should not be nil")
	}
}

// TestGetStats_withFiles verifies stats with stored files.
func TestGetStats_withFiles(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewStorageManager(tempDir)
	if err != nil {
		t.Fatalf("NewStorageManager() error = %v", err)
	}

	// Store multiple files
	content1 := []byte("content 1 - more than 16 bytes")
	content2 := []byte("content 2 - more than 16 bytes")
	content3 := []byte("content 3 - more than 16 bytes")

	hash1, _, _ := manager.StoreFile(bytes.NewReader(content1))
	hash2, _, _ := manager.StoreFile(bytes.NewReader(content2))
	hash3, _, _ := manager.StoreFile(bytes.NewReader(content3))

	stats, err := manager.GetStats()
	if err != nil {
		t.Fatalf("GetStats() error = %v", err)
	}

	if stats.TotalFiles != 3 {
		t.Errorf("TotalFiles = %d, want 3", stats.TotalFiles)
	}

	expectedSize := int64(len(content1) + len(content2) + len(content3))
	if stats.TotalSize != expectedSize {
		t.Errorf("TotalSize = %d, want %d", stats.TotalSize, expectedSize)
	}

	// Verify prefix counting
	prefixCount := 0
	for _, count := range stats.ByPrefix {
		prefixCount += count
	}
	if prefixCount != 3 {
		t.Errorf("ByPrefix total = %d, want 3", prefixCount)
	}

	// All files should have the same prefix if hashes start with same characters
	// or different prefixes if they don't
	if hash1[0:2] == hash2[0:2] || hash1[0:2] == hash3[0:2] || hash2[0:2] == hash3[0:2] {
		// At least two files share a prefix
	}
}

// TestCleanup_removesUnreferencedFiles verifies cleanup functionality.
func TestCleanup_removesUnreferencedFiles(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewStorageManager(tempDir)
	if err != nil {
		t.Fatalf("NewStorageManager() error = %v", err)
	}

	// Store files
	content1 := []byte("content 1 - more than 16 bytes")
	content2 := []byte("content 2 - more than 16 bytes")
	content3 := []byte("content 3 - more than 16 bytes")

	hash1, _, _ := manager.StoreFile(bytes.NewReader(content1))
	hash2, _, _ := manager.StoreFile(bytes.NewReader(content2))
	hash3, size3, _ := manager.StoreFile(bytes.NewReader(content3))

	// Mark only hash1 and hash2 as safe (hash3 should be removed)
	safeHashes := map[string]bool{
		hash1: true,
		hash2: true,
	}

	removed, freed, err := manager.Cleanup(safeHashes)
	if err != nil {
		t.Fatalf("Cleanup() error = %v", err)
	}

	if removed != 1 {
		t.Errorf("Removed = %d, want 1", removed)
	}

	if freed != size3 {
		t.Errorf("Freed bytes = %d, want %d", freed, size3)
	}

	// Verify hash3 is gone
	exists, _ := manager.FileExists(hash3)
	if exists {
		t.Error("Unreferenced file should be deleted")
	}

	// Verify hash1 and hash2 still exist
	for _, hash := range []string{hash1, hash2} {
		exists, _ := manager.FileExists(hash)
		if !exists {
			t.Errorf("Safe file %s should still exist", hash)
		}
	}
}

// TestCleanup_emptySafeHashes verifies cleanup with no safe hashes.
func TestCleanup_emptySafeHashes(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewStorageManager(tempDir)
	if err != nil {
		t.Fatalf("NewStorageManager() error = %v", err)
	}

	content := []byte("test content - more than 16 bytes")
	hash, _, _ := manager.StoreFile(bytes.NewReader(content))

	safeHashes := map[string]bool{} // Empty

	removed, freed, err := manager.Cleanup(safeHashes)
	if err != nil {
		t.Fatalf("Cleanup() error = %v", err)
	}

	if removed != 1 {
		t.Errorf("Removed = %d, want 1", removed)
	}

	if freed == 0 {
		t.Error("Freed bytes should be > 0")
	}

	// Verify file was removed
	exists, _ := manager.FileExists(hash)
	if exists {
		t.Error("All files should be removed with empty safe hashes")
	}
}

// TestCleanup_keepsReferencedFiles verifies cleanup keeps referenced files.
func TestCleanup_keepsReferencedFiles(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewStorageManager(tempDir)
	if err != nil {
		t.Fatalf("NewStorageManager() error = %v", err)
	}

	content := []byte("test content - more than 16 bytes")
	hash, _, _ := manager.StoreFile(bytes.NewReader(content))

	safeHashes := map[string]bool{
		hash: true,
	}

	removed, freed, err := manager.Cleanup(safeHashes)
	if err != nil {
		t.Fatalf("Cleanup() error = %v", err)
	}

	if removed != 0 {
		t.Errorf("Removed = %d, want 0 (all files are safe)", removed)
	}

	if freed != 0 {
		t.Errorf("Freed bytes = %d, want 0 (all files are safe)", freed)
	}

	// Verify file still exists
	exists, _ := manager.FileExists(hash)
	if !exists {
		t.Error("Safe file should still exist")
	}
}

// =====================================================
// Helper Types
// =====================================================

// errorReader is a test helper that always returns an error.
type errorReader struct {
	err error
}

func (r *errorReader) Read(p []byte) (n int, err error) {
	return 0, r.err
}
