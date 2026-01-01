// Package crypto tests for export archive encryption and decryption.
package crypto

import (
	"bytes"
	"strings"
	"testing"
)

// =====================================================
// ValidatePassword Tests
// =====================================================

// TestValidatePassword_success verifies valid password acceptance.
func TestValidatePassword_success(t *testing.T) {
	password := "valid-password-123"

	err := ValidatePassword(password)
	if err != nil {
		t.Errorf("ValidatePassword() error = %v, want nil", err)
	}
}

// TestValidatePassword_tooShort verifies password length validation.
func TestValidatePassword_tooShort(t *testing.T) {
	passwords := []string{
		"",                     // 0 chars
		"short",                // 5 chars
		"1234567",              // 7 chars
		"abc",                  // 3 chars
	}

	for _, pw := range passwords {
		t.Run(pw, func(t *testing.T) {
			err := ValidatePassword(pw)
			if err == nil {
				t.Errorf("ValidatePassword(%q) should return error", pw)
			}
			if !strings.Contains(err.Error(), "must be at least") {
				t.Errorf("Error should mention minimum length, got: %v", err)
			}
		})
	}
}

// TestValidatePassword_exactlyMinLength verifies boundary condition.
func TestValidatePassword_exactlyMinLength(t *testing.T) {
	password := strings.Repeat("a", PasswordMinLength)

	err := ValidatePassword(password)
	if err != nil {
		t.Errorf("ValidatePassword() with exact minimum length error = %v", err)
	}
}

// =====================================================
// GeneratePassword Tests
// =====================================================

// TestGeneratePassword_success verifies password generation.
func TestGeneratePassword_success(t *testing.T) {
	password, err := GeneratePassword(16)

	if err != nil {
		t.Fatalf("GeneratePassword() error = %v", err)
	}

	if len(password) != 16 {
		t.Errorf("GeneratePassword() length = %d, want 16", len(password))
	}

	// Verify it's valid
	if err := ValidatePassword(password); err != nil {
		t.Errorf("Generated password is invalid: %v", err)
	}
}

// TestGeneratePassword_tooShort verifies minimum length enforcement.
func TestGeneratePassword_tooShort(t *testing.T) {
	password, err := GeneratePassword(4)

	if err != nil {
		t.Fatalf("GeneratePassword() error = %v", err)
	}

	// Should use minimum length
	if len(password) < PasswordMinLength {
		t.Errorf("GeneratePassword() length = %d, want >= %d", len(password), PasswordMinLength)
	}
}

// TestGeneratePassword_uniqueness verifies passwords are unique.
func TestGeneratePassword_uniqueness(t *testing.T) {
	passwords := make(map[string]bool)

	for i := 0; i < 100; i++ {
		password, err := GeneratePassword(16)
		if err != nil {
			t.Fatalf("GeneratePassword() error = %v", err)
		}

		if passwords[password] {
			t.Error("GeneratePassword() generated duplicate password")
		}
		passwords[password] = true
	}
}

// =====================================================
// deriveKey Tests
// =====================================================

// TestDeriveKey_consistency verifies same inputs produce same key.
func TestDeriveKey_consistency(t *testing.T) {
	password := "test-password-123"
	salt := make([]byte, SaltLength)

	key1 := deriveKey(password, salt)
	key2 := deriveKey(password, salt)

	if !bytes.Equal(key1, key2) {
		t.Error("deriveKey() produced different keys for same inputs")
	}

	// Key should be 32 bytes (SHA-256 output)
	if len(key1) != 32 {
		t.Errorf("deriveKey() key length = %d, want 32", len(key1))
	}
}

// TestDeriveKey_differentPasswords verifies different passwords produce different keys.
func TestDeriveKey_differentPasswords(t *testing.T) {
	salt := make([]byte, SaltLength)

	key1 := deriveKey("password1", salt)
	key2 := deriveKey("password2", salt)

	if bytes.Equal(key1, key2) {
		t.Error("deriveKey() produced same keys for different passwords")
	}
}

// TestDeriveKey_differentSalts verifies different salts produce different keys.
func TestDeriveKey_differentSalts(t *testing.T) {
	password := "test-password"
	salt1 := make([]byte, SaltLength)
	salt2 := make([]byte, SaltLength)
	salt2[0] = 1 // Make it different

	key1 := deriveKey(password, salt1)
	key2 := deriveKey(password, salt2)

	if bytes.Equal(key1, key2) {
		t.Error("deriveKey() produced same keys for different salts")
	}
}

// =====================================================
// serializeHeader Tests
// =====================================================

// TestSerializeHeader_success verifies successful header serialization.
func TestSerializeHeader_success(t *testing.T) {
	header := ArchiveHeader{
		Version:    1,
		Algorithm:  "AES-256-GCM",
		Nonce:      make([]byte, 12),
		Salt:       make([]byte, SaltLength),
	}

	data, err := serializeHeader(header)
	if err != nil {
		t.Fatalf("serializeHeader() error = %v", err)
	}

	// Verify magic number
	if !bytes.HasPrefix(data, []byte(headerMagic)) {
		t.Errorf("serializeHeader() should start with %q", headerMagic)
	}

	// Verify data is not empty
	if len(data) == 0 {
		t.Error("serializeHeader() returned empty data")
	}

	// Should be able to parse it back
	parsed, remaining, err := parseHeader(data)
	if err != nil {
		t.Fatalf("parseHeader() of serialized data error = %v", err)
	}

	if parsed.Version != 1 {
		t.Errorf("Parsed Version = %d, want 1", parsed.Version)
	}

	if parsed.Algorithm != "AES-256-GCM" {
		t.Errorf("Parsed Algorithm = %q, want 'AES-256-GCM'", parsed.Algorithm)
	}

	// Remaining should be empty (no payload)
	if len(remaining) != 0 {
		t.Errorf("Remaining data = %d, want 0", len(remaining))
	}
}

// TestSerializeHeader_algorithmTooLong verifies error handling.
func TestSerializeHeader_algorithmTooLong(t *testing.T) {
	header := ArchiveHeader{
		Version:    1,
		Algorithm:  strings.Repeat("A", 256), // Too long
		Nonce:      make([]byte, 12),
		Salt:       make([]byte, SaltLength),
	}

	_, err := serializeHeader(header)
	if err == nil {
		t.Error("serializeHeader() with long algorithm should return error")
	}

	if !strings.Contains(err.Error(), "too long") {
		t.Errorf("Error should mention 'too long', got: %v", err)
	}
}

// TestSerializeHeader_nonceTooLong verifies error handling.
func TestSerializeHeader_nonceTooLong(t *testing.T) {
	header := ArchiveHeader{
		Version:    1,
		Algorithm:  "AES-256-GCM",
		Nonce:      make([]byte, 256), // Too long
		Salt:       make([]byte, SaltLength),
	}

	_, err := serializeHeader(header)
	if err == nil {
		t.Error("serializeHeader() with long nonce should return error")
	}
}

// TestSerializeHeader_saltTooLong verifies error handling.
func TestSerializeHeader_saltTooLong(t *testing.T) {
	header := ArchiveHeader{
		Version:    1,
		Algorithm:  "AES-256-GCM",
		Nonce:      make([]byte, 12),
		Salt:       make([]byte, 256), // Too long
	}

	_, err := serializeHeader(header)
	if err == nil {
		t.Error("serializeHeader() with long salt should return error")
	}
}

// =====================================================
// parseHeader Tests
// =====================================================

// TestParseHeader_success verifies successful header parsing.
func TestParseHeader_success(t *testing.T) {
	// Create a valid header
	header := ArchiveHeader{
		Version:    1,
		Algorithm:  "AES-256-GCM",
		Nonce:      make([]byte, 12),
		Salt:       make([]byte, SaltLength),
	}
	headerData, _ := serializeHeader(header)

	// Verify it can be parsed back
	parsed, remaining, err := parseHeader(headerData)
	if err != nil {
		t.Fatalf("parseHeader() error = %v", err)
	}

	if parsed.Version != 1 {
		t.Errorf("parseHeader() Version = %d, want 1", parsed.Version)
	}

	if parsed.Algorithm != "AES-256-GCM" {
		t.Errorf("parseHeader() Algorithm = %q, want 'AES-256-GCM'", parsed.Algorithm)
	}

	if len(parsed.Nonce) != 12 {
		t.Errorf("parseHeader() Nonce length = %d, want 12", len(parsed.Nonce))
	}

	if len(parsed.Salt) != SaltLength {
		t.Errorf("parseHeader() Salt length = %d, want %d", len(parsed.Salt), SaltLength)
	}

	// Remaining should be empty (no payload in this case)
	if len(remaining) != 0 {
		t.Errorf("parseHeader() remaining length = %d, want 0", len(remaining))
	}
}

// TestParseHeader_invalidMagic verifies error handling.
func TestParseHeader_invalidMagic(t *testing.T) {
	invalidData := []byte("INVALID" + strings.Repeat("\x00", 100))

	_, _, err := parseHeader(invalidData)
	if err == nil {
		t.Error("parseHeader() with invalid magic should return error")
	}

	if !strings.Contains(err.Error(), "invalid magic") {
		t.Errorf("Error should mention 'invalid magic', got: %v", err)
	}
}

// TestParseHeader_tooShort verifies error handling for truncated data.
func TestParseHeader_tooShort(t *testing.T) {
	shortData := []byte("MNEXARC") // Only magic, nothing else

	_, _, err := parseHeader(shortData)
	if err == nil {
		t.Error("parseHeader() with short data should return error")
	}
}

// TestParseHeader_emptyData verifies error handling for empty data.
func TestParseHeader_emptyData(t *testing.T) {
	_, _, err := parseHeader([]byte{})
	if err == nil {
		t.Error("parseHeader() with empty data should return error")
	}
}

// =====================================================
// EncryptArchive Tests
// =====================================================

// TestEncryptArchive_success verifies successful encryption.
func TestEncryptArchive_success(t *testing.T) {
	data := []byte("test data for encryption")
	password := "test-password-123"

	encrypted, err := EncryptArchive(data, password)
	if err != nil {
		t.Fatalf("EncryptArchive() error = %v", err)
	}

	if len(encrypted) == 0 {
		t.Error("EncryptArchive() returned empty data")
	}

	// Verify header is present
	if !bytes.HasPrefix(encrypted, []byte(headerMagic)) {
		t.Error("EncryptArchive() should start with magic number")
	}
}

// TestEncryptArchive_shortPassword verifies password validation.
func TestEncryptArchive_shortPassword(t *testing.T) {
	data := []byte("test data")
	password := "short" // Less than 8 characters

	_, err := EncryptArchive(data, password)
	if err == nil {
		t.Error("EncryptArchive() with short password should return error")
	}

	if !strings.Contains(err.Error(), "must be at least") {
		t.Errorf("Error should mention minimum length, got: %v", err)
	}
}

// TestEncryptArchive_emptyData verifies encryption of empty data.
func TestEncryptArchive_emptyData(t *testing.T) {
	data := []byte{}
	password := "test-password-123"

	encrypted, err := EncryptArchive(data, password)
	if err != nil {
		t.Fatalf("EncryptArchive() error = %v", err)
	}

	// Empty data should still produce encrypted output (with header)
	if len(encrypted) == 0 {
		t.Error("EncryptArchive() with empty data should still produce output")
	}
}

// TestEncryptArchive_largeData verifies encryption of larger data.
func TestEncryptArchive_largeData(t *testing.T) {
	// Create 10KB of data
	data := make([]byte, 10240)
	for i := range data {
		data[i] = byte(i % 256)
	}
	password := "test-password-123"

	encrypted, err := EncryptArchive(data, password)
	if err != nil {
		t.Fatalf("EncryptArchive() error = %v", err)
	}

	if len(encrypted) == 0 {
		t.Error("EncryptArchive() returned empty data")
	}

	// Encrypted data should be larger than original (header + GCM tag)
	if len(encrypted) <= len(data) {
		t.Errorf("Encrypted length = %d, should be > %d", len(encrypted), len(data))
	}
}

// TestEncryptArchive_uniqueness verifies each encryption produces unique output.
func TestEncryptArchive_uniqueness(t *testing.T) {
	data := []byte("test data")
	password := "test-password-123"

	encrypted1, err1 := EncryptArchive(data, password)
	encrypted2, err2 := EncryptArchive(data, password)

	if err1 != nil || err2 != nil {
		t.Fatalf("EncryptArchive() error = %v, %v", err1, err2)
	}

	// Should be different due to random salt/nonce
	if bytes.Equal(encrypted1, encrypted2) {
		t.Error("EncryptArchive() produced identical output (salt/nonce should be random)")
	}
}

// =====================================================
// DecryptArchive Tests
// =====================================================

// TestDecryptArchive_success verifies successful decryption.
func TestDecryptArchive_success(t *testing.T) {
	data := []byte("test data for encryption")
	password := "test-password-123"

	// First encrypt
	encrypted, err := EncryptArchive(data, password)
	if err != nil {
		t.Fatalf("EncryptArchive() error = %v", err)
	}

	// Then decrypt
	decrypted, err := DecryptArchive(encrypted, password)
	if err != nil {
		t.Fatalf("DecryptArchive() error = %v", err)
	}

	if !bytes.Equal(decrypted, data) {
		t.Errorf("DecryptArchive() = %q, want %q", string(decrypted), string(data))
	}
}

// TestDecryptArchive_wrongPassword verifies error handling.
func TestDecryptArchive_wrongPassword(t *testing.T) {
	data := []byte("test data")
	correctPassword := "correct-password-123"
	wrongPassword := "wrong-password-456"

	encrypted, err := EncryptArchive(data, correctPassword)
	if err != nil {
		t.Fatalf("EncryptArchive() error = %v", err)
	}

	_, err = DecryptArchive(encrypted, wrongPassword)
	if err == nil {
		t.Error("DecryptArchive() with wrong password should return error")
	}

	// Should be ErrInvalidPassword (wrapped)
	if !strings.Contains(err.Error(), "invalid password") {
		t.Errorf("Error should mention 'invalid password', got: %v", err)
	}
}

// TestDecryptArchive_invalidData verifies error handling.
func TestDecryptArchive_invalidData(t *testing.T) {
	password := "test-password-123"
	invalidData := [][]byte{
		[]byte("not valid data"),
		[]byte("MNEXARC"), // Only magic
		make([]byte, 100), // All zeros
	}

	for _, data := range invalidData {
		_, err := DecryptArchive(data, password)
		if err == nil {
			t.Error("DecryptArchive() with invalid data should return error")
		}
	}
}

// TestDecryptArchive_tamperedData verifies error handling for modified data.
func TestDecryptArchive_tamperedData(t *testing.T) {
	data := []byte("test data for encryption")
	password := "test-password-123"

	encrypted, err := EncryptArchive(data, password)
	if err != nil {
		t.Fatalf("EncryptArchive() error = %v", err)
	}

	// Tamper with the ciphertext (after header)
	// Header size = magic(7) + version(1) + algLen(1) + "AES-256-GCM"(11) + nonceLen(1) + nonce(12) + saltLen(1) + salt(32) = 66
	headerSize := 7 + 1 + 1 + 11 + 1 + 12 + 1 + 32
	if len(encrypted) > headerSize {
		encrypted[headerSize] ^= 0xFF // Flip bits in ciphertext
	}

	_, err = DecryptArchive(encrypted, password)
	if err == nil {
		t.Error("DecryptArchive() with tampered data should return error")
	}

	// Should be authentication failure
	if !strings.Contains(err.Error(), "invalid password") {
		t.Errorf("Error should mention 'invalid password', got: %v", err)
	}
}

// TestDecryptArchive_emptyData verifies error handling.
func TestDecryptArchive_emptyData(t *testing.T) {
	password := "test-password-123"

	_, err := DecryptArchive([]byte{}, password)
	if err == nil {
		t.Error("DecryptArchive() with empty data should return error")
	}
}

// TestDecryptArchive_unsupportedVersion verifies error handling.
func TestDecryptArchive_unsupportedVersion(t *testing.T) {
	// Manually create a header with unsupported version
	password := "test-password-123"

	data := []byte("test data")
	encrypted, _ := EncryptArchive(data, password)

	// Modify version byte to something unsupported
	// Version is at index 7 (after magic "MNEXARC" which is 7 bytes)
	if len(encrypted) > 7 {
		encrypted[7] = 99 // Version 99 doesn't exist
	}

	_, err := DecryptArchive(encrypted, password)
	if err == nil {
		t.Error("DecryptArchive() with unsupported version should return error")
	}

	if !strings.Contains(err.Error(), "unsupported archive version") {
		t.Errorf("Error should mention 'unsupported version', got: %v", err)
	}
}

// =====================================================
// Round-Trip Tests
// =====================================================

// TestEncryptDecrypt_roundTrip verifies complete encryption/decryption cycle.
func TestEncryptDecrypt_roundTrip(t *testing.T) {
	testCases := []struct {
		name string
		data []byte
	}{
		{"small", []byte("small")},
		{"medium", make([]byte, 1024)},
		{"large", make([]byte, 10240)},
		{"unicode", []byte("Hello 世界 🌍")},
	}

	password := "test-password-123"

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Fill large data with pattern
			if tc.name == "medium" || tc.name == "large" {
				for i := range tc.data {
					tc.data[i] = byte(i % 256)
				}
			}

			encrypted, err := EncryptArchive(tc.data, password)
			if err != nil {
				t.Fatalf("EncryptArchive() error = %v", err)
			}

			decrypted, err := DecryptArchive(encrypted, password)
			if err != nil {
				t.Fatalf("DecryptArchive() error = %v", err)
			}

			if !bytes.Equal(decrypted, tc.data) {
				t.Errorf("Round-trip failed: original length=%d, decrypted length=%d",
					len(tc.data), len(decrypted))
			}
		})
	}
}

// TestEncryptDecrypt_differentPasswords verifies encryption is password-specific.
func TestEncryptDecrypt_differentPasswords(t *testing.T) {
	data := []byte("sensitive data")
	password1 := "password-one-123"
	password2 := "password-two-456"

	encrypted, err := EncryptArchive(data, password1)
	if err != nil {
		t.Fatalf("EncryptArchive() error = %v", err)
	}

	// Try to decrypt with wrong password
	_, err = DecryptArchive(encrypted, password2)
	if err == nil {
		t.Error("DecryptArchive() with different password should fail")
	}
}

// =====================================================
// Error Variable Tests
// =====================================================

// TestErrInvalidPassword verifies error variable is exported.
func TestErrInvalidPassword(t *testing.T) {
	if ErrInvalidPassword == nil {
		t.Error("ErrInvalidPassword should not be nil")
	}
	if ErrInvalidPassword.Error() == "" {
		t.Error("ErrInvalidPassword should have error message")
	}
}

// TestErrInvalidArchive verifies error variable is exported.
func TestErrInvalidArchive(t *testing.T) {
	if ErrInvalidArchive == nil {
		t.Error("ErrInvalidArchive should not be nil")
	}
	if ErrInvalidArchive.Error() == "" {
		t.Error("ErrInvalidArchive should have error message")
	}
}

// =====================================================
// Constant Tests
// =====================================================

// TestPasswordMinLength verifies constant value.
func TestPasswordMinLength(t *testing.T) {
	if PasswordMinLength != 8 {
		t.Errorf("PasswordMinLength = %d, want 8", PasswordMinLength)
	}
}

// TestSaltLength verifies constant value.
func TestSaltLength(t *testing.T) {
	if SaltLength != 32 {
		t.Errorf("SaltLength = %d, want 32", SaltLength)
	}
}

// TestHeaderMagic verifies constant value.
func TestHeaderMagic(t *testing.T) {
	if headerMagic != "MNEXARC" {
		t.Errorf("headerMagic = %q, want 'MNEXARC'", headerMagic)
	}
}
