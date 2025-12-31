// Package export provides unit tests for encryption functionality.
// T176: Unit test for archive encryption (AES-256-GCM).
package export

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestEncryptFile verifies file encryption with AES-256-GCM.
func TestEncryptFile(t *testing.T) {
	// Create temporary directory
	tempDir := t.TempDir()

	// Create a test file
	srcFile := filepath.Join(tempDir, "test.txt")
	plaintext := []byte("Hello, World! This is a test file for encryption.")
	if err := os.WriteFile(srcFile, plaintext, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	dstFile := filepath.Join(tempDir, "encrypted.bin")
	password := "test-password-123"

	// Encrypt the file
	size, err := encryptFile(srcFile, dstFile, password)
	if err != nil {
		t.Fatalf("encryptFile failed: %v", err)
	}

	// Verify output file exists
	if _, err := os.Stat(dstFile); os.IsNotExist(err) {
		t.Fatal("Encrypted file was not created")
	}

	// Verify size is reasonable (should be larger than plaintext due to GCM overhead)
	if size <= int64(len(plaintext)) {
		t.Errorf("Encrypted file size (%d) should be larger than plaintext (%d)", size, len(plaintext))
	}

	// Verify file content is different from plaintext
	encrypted, err := os.ReadFile(dstFile)
	if err != nil {
		t.Fatalf("Failed to read encrypted file: %v", err)
	}

	if strings.Contains(string(encrypted), string(plaintext)) {
		t.Error("Encrypted file contains plaintext data")
	}
}

// TestEncryptDecrypt verifies encryption and decryption round-trip.
func TestEncryptDecrypt(t *testing.T) {
	tempDir := t.TempDir()

	// Create test file
	srcFile := filepath.Join(tempDir, "test.txt")
	plaintext := []byte("Sensitive data that needs encryption. " +
		"Lorem ipsum dolor sit amet, consectetur adipiscing elit. " +
		"Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.")
	if err := os.WriteFile(srcFile, plaintext, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	encryptedFile := filepath.Join(tempDir, "encrypted.bin")
	decryptedFile := filepath.Join(tempDir, "decrypted.txt")
	password := "secure-password-456"

	// Encrypt
	_, err := encryptFile(srcFile, encryptedFile, password)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	// Decrypt
	err = decryptFile(encryptedFile, decryptedFile, password)
	if err != nil {
		t.Fatalf("Decryption failed: %v", err)
	}

	// Verify decrypted content matches original
	decrypted, err := os.ReadFile(decryptedFile)
	if err != nil {
		t.Fatalf("Failed to read decrypted file: %v", err)
	}

	if string(decrypted) != string(plaintext) {
		t.Errorf("Decrypted content does not match original.\nGot: %s\nWant: %s", string(decrypted), string(plaintext))
	}
}

// TestDecryptWithWrongPassword verifies decryption fails with wrong password.
func TestDecryptWithWrongPassword(t *testing.T) {
	tempDir := t.TempDir()

	// Create and encrypt file
	srcFile := filepath.Join(tempDir, "test.txt")
	plaintext := []byte("Secret message")
	if err := os.WriteFile(srcFile, plaintext, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	encryptedFile := filepath.Join(tempDir, "encrypted.bin")
	decryptedFile := filepath.Join(tempDir, "decrypted.txt")
	correctPassword := "correct-password"
	wrongPassword := "wrong-password"

	_, err := encryptFile(srcFile, encryptedFile, correctPassword)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	// Try to decrypt with wrong password
	err = decryptFile(encryptedFile, decryptedFile, wrongPassword)
	if err == nil {
		t.Error("Decryption with wrong password should fail")
	}

	// Verify error message indicates authentication failure
	if !strings.Contains(err.Error(), "decryption") && !strings.Contains(err.Error(), "authentication") {
		t.Logf("Error: %v", err)
	}
}

// TestEncryptEmptyFile verifies encryption of empty files works.
func TestEncryptEmptyFile(t *testing.T) {
	tempDir := t.TempDir()

	// Create empty file
	srcFile := filepath.Join(tempDir, "empty.txt")
	if err := os.WriteFile(srcFile, []byte{}, 0644); err != nil {
		t.Fatalf("Failed to create empty file: %v", err)
	}

	dstFile := filepath.Join(tempDir, "encrypted.bin")
	password := "test-password"

	// Should succeed even with empty file
	_, err := encryptFile(srcFile, dstFile, password)
	if err != nil {
		t.Errorf("Failed to encrypt empty file: %v", err)
	}
}

// TestEncryptLargeFile verifies encryption of larger files.
func TestEncryptLargeFile(t *testing.T) {
	tempDir := t.TempDir()

	// Create a 1MB file
	srcFile := filepath.Join(tempDir, "large.bin")
	largeData := make([]byte, 1024*1024)
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}
	if err := os.WriteFile(srcFile, largeData, 0644); err != nil {
		t.Fatalf("Failed to create large file: %v", err)
	}

	dstFile := filepath.Join(tempDir, "encrypted.bin")
	decryptedFile := filepath.Join(tempDir, "decrypted.bin")
	password := "large-file-password"

	// Encrypt
	size, err := encryptFile(srcFile, dstFile, password)
	if err != nil {
		t.Fatalf("Failed to encrypt large file: %v", err)
	}

	// Verify size is reasonable (plaintext + salt + nonce + GCM tag)
	if size < int64(len(largeData)) {
		t.Errorf("Encrypted size (%d) less than plaintext (%d)", size, len(largeData))
	}

	// Decrypt and verify
	err = decryptFile(dstFile, decryptedFile, password)
	if err != nil {
		t.Fatalf("Failed to decrypt large file: %v", err)
	}

	decrypted, err := os.ReadFile(decryptedFile)
	if err != nil {
		t.Fatalf("Failed to read decrypted file: %v", err)
	}

	if len(decrypted) != len(largeData) {
		t.Errorf("Decrypted size mismatch: got %d, want %d", len(decrypted), len(largeData))
	}

	// Verify first and last bytes match
	if decrypted[0] != largeData[0] || decrypted[len(decrypted)-1] != largeData[len(largeData)-1] {
		t.Error("Decrypted data does not match original")
	}
}

// TestDeriveKey verifies PBKDF2 key derivation.
func TestDeriveKey(t *testing.T) {
	password := "test-password"
	salt := make([]byte, 16)
	for i := range salt {
		salt[i] = byte(i)
	}

	// Derive key
	key1 := deriveKey(password, salt)

	// Key should be 32 bytes (256 bits)
	if len(key1) != 32 {
		t.Errorf("Derived key length is %d, expected 32", len(key1))
	}

	// Same password and salt should produce same key
	key2 := deriveKey(password, salt)
	if string(key1) != string(key2) {
		t.Error("Same inputs should produce same key")
	}

	// Different salt should produce different key
	differentSalt := make([]byte, 16)
	for i := range differentSalt {
		differentSalt[i] = byte(i + 1)
	}
	key3 := deriveKey(password, differentSalt)
	if string(key1) == string(key3) {
		t.Error("Different salt should produce different key")
	}

	// Different password should produce different key
	key4 := deriveKey("different-password", salt)
	if string(key1) == string(key4) {
		t.Error("Different password should produce different key")
	}
}
