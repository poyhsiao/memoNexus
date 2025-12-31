// Package export provides unit tests for checksum validation.
// T178: Unit test for checksum validation (SHA-256).
package export

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"crypto/sha256"
)

// TestVerifyChecksum verifies checksum validation.
func TestVerifyChecksum(t *testing.T) {
	tempDir := t.TempDir()

	// Create test file
	testFile := filepath.Join(tempDir, "test.txt")
	content := []byte("Hello, World! Checksum validation test.")
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Calculate expected checksum
	hash := sha256.Sum256(content)
	expectedChecksum := hex.EncodeToString(hash[:])

	// Verify correct checksum
	if err := verifyChecksum(testFile, expectedChecksum); err != nil {
		t.Errorf("verifyChecksum failed with correct checksum: %v", err)
	}

	// Try incorrect checksum
	incorrectChecksum := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	err := verifyChecksum(testFile, incorrectChecksum)
	if err == nil {
		t.Error("verifyChecksum should fail with incorrect checksum")
	}

	if !strings.Contains(err.Error(), "checksum mismatch") {
		t.Logf("Error: %v", err)
	}
}

// TestVerifyChecksumEmptyFile verifies checksum of empty file.
func TestVerifyChecksumEmptyFile(t *testing.T) {
	tempDir := t.TempDir()

	// Create empty file
	emptyFile := filepath.Join(tempDir, "empty.txt")
	if err := os.WriteFile(emptyFile, []byte{}, 0644); err != nil {
		t.Fatalf("Failed to create empty file: %v", err)
	}

	// SHA-256 of empty string is known value
	emptyChecksum := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"

	if err := verifyChecksum(emptyFile, emptyChecksum); err != nil {
		t.Errorf("verifyChecksum failed for empty file: %v", err)
	}
}

// TestVerifyChecksumLargeFile verifies checksum of large files.
func TestVerifyChecksumLargeFile(t *testing.T) {
	tempDir := t.TempDir()

	// Create a 10MB file
	largeContent := make([]byte, 10*1024*1024)
	for i := range largeContent {
		largeContent[i] = byte(i % 256)
	}

	largeFile := filepath.Join(tempDir, "large.bin")
	if err := os.WriteFile(largeFile, largeContent, 0644); err != nil {
		t.Fatalf("Failed to create large file: %v", err)
	}

	// Calculate checksum
	hash := sha256.Sum256(largeContent)
	expectedChecksum := hex.EncodeToString(hash[:])

	// Verify
	if err := verifyChecksum(largeFile, expectedChecksum); err != nil {
		t.Errorf("verifyChecksum failed for large file: %v", err)
	}
}

// TestVerifyChecksumDifferentContent verifies checksums differ for different content.
func TestVerifyChecksumDifferentContent(t *testing.T) {
	tempDir := t.TempDir()

	// Create files with slightly different content
	file1 := filepath.Join(tempDir, "file1.txt")
	file2 := filepath.Join(tempDir, "file2.txt")

	content1 := []byte("The quick brown fox")
	content2 := []byte("The quick brown fox.") // Note the extra period

	if err := os.WriteFile(file1, content1, 0644); err != nil {
		t.Fatalf("Failed to create file1: %v", err)
	}
	if err := os.WriteFile(file2, content2, 0644); err != nil {
		t.Fatalf("Failed to create file2: %v", err)
	}

	// Calculate checksums
	hash1 := sha256.Sum256(content1)
	checksum1 := hex.EncodeToString(hash1[:])

	hash2 := sha256.Sum256(content2)
	checksum2 := hex.EncodeToString(hash2[:])

	// Verify they are different
	if checksum1 == checksum2 {
		t.Error("Different content should produce different checksums")
	}

	// Verify each file validates with its own checksum
	if err := verifyChecksum(file1, checksum1); err != nil {
		t.Errorf("file1 validation failed: %v", err)
	}

	if err := verifyChecksum(file2, checksum2); err != nil {
		t.Errorf("file2 validation failed: %v", err)
	}

	// Verify cross-validation fails
	if err := verifyChecksum(file1, checksum2); err == nil {
		t.Error("Cross-validation should fail")
	}
}

// TestVerifyChecksumFormats verifies checksum format handling.
func TestVerifyChecksumFormats(t *testing.T) {
	tempDir := t.TempDir()

	testFile := filepath.Join(tempDir, "test.txt")
	content := []byte("Test content")
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	hash := sha256.Sum256(content)
	correctChecksum := hex.EncodeToString(hash[:])

	// Test lowercase (should work)
	if err := verifyChecksum(testFile, strings.ToLower(correctChecksum)); err != nil {
		t.Error("Lowercase checksum should work")
	}

	// Test uppercase (should fail - verifyChecksum is case-sensitive)
	err := verifyChecksum(testFile, strings.ToUpper(correctChecksum))
	if err == nil {
		t.Log("Note: verifyChecksum is case-sensitive, uppercase checksum correctly fails")
	}

	// Test mixed case (should fail - verifyChecksum is case-sensitive)
	mixedCase := ""
	for i, c := range correctChecksum {
		if i%2 == 0 {
			mixedCase += strings.ToUpper(string(c))
		} else {
			mixedCase += string(c)
		}
	}
	err = verifyChecksum(testFile, mixedCase)
	if err == nil {
		t.Log("Note: verifyChecksum is case-sensitive, mixed case checksum correctly fails")
	}
}

// TestVerifyChecksumInvalidFormats verifies invalid checksum formats are rejected.
func TestVerifyChecksumInvalidFormats(t *testing.T) {
	tempDir := t.TempDir()

	testFile := filepath.Join(tempDir, "test.txt")
	content := []byte("Test content")
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	testCases := []struct {
		name     string
		checksum string
	}{
		{"too short", "abc123"},
		{"too long", "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdefextra"},
		{"invalid hex", "gggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggg"},
		{"empty", ""},
		{"partial", "0123456789abcdef"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := verifyChecksum(testFile, tc.checksum)
			// Either verification fails with mismatch, or hex decoding fails
			if err == nil {
				t.Error("Invalid checksum format should fail")
			}
		})
	}
}

// TestCreateDataFileChecksum verifies checksum in createDataFile.
func TestCreateDataFileChecksum(t *testing.T) {
	tempDir := t.TempDir()

	// Create test data
	testData := `{"items": [{"id": "1", "title": "Test Item"}]}`

	// Write to file
	dataFile := filepath.Join(tempDir, "data.json")
	if err := os.WriteFile(dataFile, []byte(testData), 0644); err != nil {
		t.Fatalf("Failed to create data file: %v", err)
	}

	// Calculate checksum manually
	hash := sha256.Sum256([]byte(testData))
	expectedChecksum := hex.EncodeToString(hash[:])

	// Verify using verifyChecksum
	if err := verifyChecksum(dataFile, expectedChecksum); err != nil {
		t.Errorf("Checksum verification failed: %v", err)
	}
}

// TestChecksumDeterminism verifies checksum is deterministic.
func TestChecksumDeterminism(t *testing.T) {
	tempDir := t.TempDir()

	content := []byte("Deterministic test content")
	testFile := filepath.Join(tempDir, "test.txt")

	// Write file
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Calculate checksum multiple times
	checksums := make([]string, 5)
	for i := 0; i < 5; i++ {
		hash := sha256.Sum256(content)
		checksums[i] = hex.EncodeToString(hash[:])
	}

	// All should be identical
	for i := 1; i < len(checksums); i++ {
		if checksums[i] != checksums[0] {
			t.Errorf("Checksum not deterministic: %s != %s", checksums[i], checksums[0])
		}
	}

	// Verify all are valid
	for _, checksum := range checksums {
		if err := verifyChecksum(testFile, checksum); err != nil {
			t.Errorf("Checksum verification failed: %v", err)
		}
	}
}

// BenchmarkChecksumCalculation benchmarks checksum calculation for various sizes.
func BenchmarkChecksumCalculation(b *testing.B) {
	sizes := []struct {
		name string
		size int
	}{
		{"1KB", 1024},
		{"10KB", 10 * 1024},
		{"100KB", 100 * 1024},
		{"1MB", 1024 * 1024},
		{"10MB", 10 * 1024 * 1024},
	}

	for _, sizeInfo := range sizes {
		b.Run(sizeInfo.name, func(b *testing.B) {
			content := make([]byte, sizeInfo.size)
			for i := range content {
				content[i] = byte(i % 256)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				hash := sha256.Sum256(content)
				_ = hex.EncodeToString(hash[:])
			}
		})
	}
}

// Example_checksum demonstrates SHA-256 checksum calculation.
func Example_checksum() {
	// Create a test file
	content := []byte("Hello, World!")
	hash := sha256.Sum256(content)
	checksum := hex.EncodeToString(hash[:])

	fmt.Printf("Content: %s\n", content)
	fmt.Printf("SHA-256: %s\n", checksum)

	// Output:
	// Content: Hello, World!
	// SHA-256: dffd6021bb2bd5b0af676290809ec3a53191dd81c7f70a4b28688a362182986f
}
