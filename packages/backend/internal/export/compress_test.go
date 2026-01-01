// Package export provides unit tests for compression functionality.
// T177: Unit test for archive compression (tar.gz).
package export

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestWriteTarGz verifies tar.gz archive creation.
func TestWriteTarGz(t *testing.T) {
	tempDir := t.TempDir()

	// Create source directory with test files
	sourceDir := filepath.Join(tempDir, "source")
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}

	// Create test files
	testFiles := map[string]string{
		"file1.txt":        "Content of file 1",
		"file2.txt":        "Content of file 2 with some more text",
		"subdir/file3.txt": "Content in subdirectory",
	}

	for name, content := range testFiles {
		path := filepath.Join(sourceDir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", name, err)
		}
	}

	targetPath := filepath.Join(tempDir, "archive.tar.gz")

	// Create tar.gz
	if err := writeTarGz(sourceDir, targetPath); err != nil {
		t.Fatalf("writeTarGz failed: %v", err)
	}

	// Verify archive exists
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		t.Fatal("Archive file was not created")
	}

	// Verify archive can be opened and read
	if err := verifyTarGzContents(targetPath, testFiles); err != nil {
		t.Fatalf("Archive verification failed: %v", err)
	}
}

// TestExtractTarGz verifies tar.gz extraction.
func TestExtractTarGz(t *testing.T) {
	tempDir := t.TempDir()

	// Create source directory with test files
	sourceDir := filepath.Join(tempDir, "source")
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}

	testFiles := map[string]string{
		"data.json":     `{"items": [{"id": "1", "title": "Test"}]}`,
		"manifest.json": `{"version": "1.0"}`,
	}

	for name, content := range testFiles {
		path := filepath.Join(sourceDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// Create archive
	archivePath := filepath.Join(tempDir, "archive.tar.gz")
	if err := writeTarGz(sourceDir, archivePath); err != nil {
		t.Fatalf("Failed to create archive: %v", err)
	}

	// Extract to different directory
	targetDir := filepath.Join(tempDir, "extracted")
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		t.Fatalf("Failed to create target directory: %v", err)
	}

	if err := extractTarGz(archivePath, targetDir); err != nil {
		t.Fatalf("extractTarGz failed: %v", err)
	}

	// Verify extracted files
	for name, expectedContent := range testFiles {
		path := filepath.Join(targetDir, name)
		content, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("Failed to read extracted file %s: %v", name, err)
			continue
		}

		if string(content) != expectedContent {
			t.Errorf("Content mismatch for %s.\nGot: %s\nWant: %s", name, string(content), expectedContent)
		}
	}
}

// TestCompressDecompressRoundTrip verifies round-trip compression/decompression.
func TestCompressDecompressRoundTrip(t *testing.T) {
	tempDir := t.TempDir()

	// Create source with complex directory structure
	sourceDir := filepath.Join(tempDir, "source")
	dirs := []string{
		"dir1",
		"dir1/subdir1",
		"dir1/subdir2",
		"dir2",
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(sourceDir, dir), 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
	}

	// Create files with various content
	testData := map[string]string{
		"root.txt":              "Root file content",
		"dir1/file1.txt":        "File in dir1",
		"dir1/subdir1/file.txt": "Nested file",
		"dir1/subdir2/data.txt": "Another nested file",
		"dir2/large.txt":        strings.Repeat("Large file content. ", 1000),
	}

	for name, content := range testData {
		path := filepath.Join(sourceDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", name, err)
		}
	}

	// Compress
	archivePath := filepath.Join(tempDir, "archive.tar.gz")
	if err := writeTarGz(sourceDir, archivePath); err != nil {
		t.Fatalf("Compression failed: %v", err)
	}

	// Decompress
	extractDir := filepath.Join(tempDir, "extracted")
	if err := os.MkdirAll(extractDir, 0755); err != nil {
		t.Fatalf("Failed to create extract directory: %v", err)
	}

	if err := extractTarGz(archivePath, extractDir); err != nil {
		t.Fatalf("Decompression failed: %v", err)
	}

	// Verify all files and content match
	for name, expectedContent := range testData {
		path := filepath.Join(extractDir, name)
		content, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("Failed to read extracted file %s: %v", name, err)
			continue
		}

		if string(content) != expectedContent {
			t.Errorf("Content mismatch for %s (got %d bytes, want %d bytes)", name, len(content), len(expectedContent))
		}
	}
}

// TestCompressEmptyDirectory verifies compression of empty directory.
func TestCompressEmptyDirectory(t *testing.T) {
	tempDir := t.TempDir()

	// Create empty directory
	sourceDir := filepath.Join(tempDir, "empty")
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	targetPath := filepath.Join(tempDir, "archive.tar.gz")

	// Should succeed even with empty directory
	if err := writeTarGz(sourceDir, targetPath); err != nil {
		t.Errorf("Failed to compress empty directory: %v", err)
	}

	// Verify archive exists
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		t.Error("Archive was not created for empty directory")
	}
}

// TestCompressLargeFile verifies compression of large files.
func TestCompressLargeFile(t *testing.T) {
	tempDir := t.TempDir()

	sourceDir := filepath.Join(tempDir, "source")
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	// Create a 10MB file
	largeContent := make([]byte, 10*1024*1024)
	for i := range largeContent {
		largeContent[i] = byte(i % 256)
	}

	largeFile := filepath.Join(sourceDir, "large.bin")
	if err := os.WriteFile(largeFile, largeContent, 0644); err != nil {
		t.Fatalf("Failed to create large file: %v", err)
	}

	archivePath := filepath.Join(tempDir, "archive.tar.gz")

	// Compress
	if err := writeTarGz(sourceDir, archivePath); err != nil {
		t.Fatalf("Failed to compress large file: %v", err)
	}

	// Verify archive size is reasonable (compression should reduce size somewhat)
	info, err := os.Stat(archivePath)
	if err != nil {
		t.Fatalf("Failed to stat archive: %v", err)
	}

	// Compressed size should be less than original (gzip is effective on repetitive data)
	if info.Size() >= int64(len(largeContent)) {
		t.Logf("Warning: Compressed size (%d) >= original (%d). Data may not be compressible.", info.Size(), len(largeContent))
	}

	// Verify extraction works
	extractDir := filepath.Join(tempDir, "extracted")
	if err := os.MkdirAll(extractDir, 0755); err != nil {
		t.Fatalf("Failed to create extract directory: %v", err)
	}

	if err := extractTarGz(archivePath, extractDir); err != nil {
		t.Fatalf("Failed to extract large archive: %v", err)
	}

	// Verify content
	extractedPath := filepath.Join(extractDir, "large.bin")
	extracted, err := os.ReadFile(extractedPath)
	if err != nil {
		t.Fatalf("Failed to read extracted file: %v", err)
	}

	if len(extracted) != len(largeContent) {
		t.Errorf("Size mismatch: got %d, want %d", len(extracted), len(largeContent))
	}
}

// TestExtractInvalidArchive verifies extraction fails gracefully.
func TestExtractInvalidArchive(t *testing.T) {
	tempDir := t.TempDir()

	// Create invalid "archive" file
	invalidArchive := filepath.Join(tempDir, "invalid.tar.gz")
	if err := os.WriteFile(invalidArchive, []byte("not a valid tar.gz"), 0644); err != nil {
		t.Fatalf("Failed to create invalid archive: %v", err)
	}

	targetDir := filepath.Join(tempDir, "extracted")

	// Should return error
	err := extractTarGz(invalidArchive, targetDir)
	if err == nil {
		t.Error("Extracting invalid archive should fail")
	}
}

// verifyTarGzContents reads and verifies the contents of a tar.gz archive.
func verifyTarGzContents(archivePath string, expectedFiles map[string]string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	foundFiles := make(map[string]bool)

	for {
		header, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				break // Normal end of archive
			}
			return err
		}

		// Skip directories
		if header.Typeflag == tar.TypeDir {
			continue
		}

		foundFiles[header.Name] = true

		// Verify content if expected
		if expectedContent, ok := expectedFiles[header.Name]; ok {
			content := make([]byte, header.Size)
			n, err := tr.Read(content)
			if err != nil && err != io.EOF {
				return err
			}

			if int64(n) != header.Size {
				return fmt.Errorf("size mismatch for %s: read %d, expected %d", header.Name, n, header.Size)
			}

			if string(content) != expectedContent {
				return fmt.Errorf("content mismatch for %s", header.Name)
			}
		}
	}

	// Verify all expected files were found
	for name := range expectedFiles {
		if !foundFiles[name] {
			return fmt.Errorf("expected file not found: %s", name)
		}
	}

	return nil
}
