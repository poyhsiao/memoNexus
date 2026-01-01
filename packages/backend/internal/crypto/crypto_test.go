// Package crypto tests for encryption and key derivation functionality.
package crypto

import (
	"strings"
	"testing"
)

// TestEncryptDecrypt_roundtrip verifies basic encryption and decryption.
func TestEncryptDecrypt_roundtrip(t *testing.T) {
	plaintext := []byte("Hello, World!")
	key := []byte("test-key-12345")

	ciphertext, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	// Verify ciphertext is valid base64
	if ciphertext == "" {
		t.Error("Encrypt() returned empty string")
	}

	// Verify we can decrypt it back
	decrypted, err := Decrypt(ciphertext, key)
	if err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}

	if string(decrypted) != string(plaintext) {
		t.Errorf("Decrypt() = %q, want %q", string(decrypted), string(plaintext))
	}
}

// TestEncrypt_differentKeys verifies different keys produce different ciphertexts.
func TestEncrypt_differentKeys(t *testing.T) {
	plaintext := []byte("Hello, World!")
	key1 := []byte("key-one")
	key2 := []byte("key-two")

	ciphertext1, err := Encrypt(plaintext, key1)
	if err != nil {
		t.Fatalf("Encrypt() with key1 error = %v", err)
	}

	ciphertext2, err := Encrypt(plaintext, key2)
	if err != nil {
		t.Fatalf("Encrypt() with key2 error = %v", err)
	}

	// Different keys should produce different ciphertexts (due to random nonce)
	if ciphertext1 == ciphertext2 {
		t.Error("Encrypt() with different keys produced same ciphertext")
	}
}

// TestEncrypt_sameKeyDifferentNonce verifies each encryption produces unique ciphertext.
func TestEncrypt_sameKeyDifferentNonce(t *testing.T) {
	plaintext := []byte("Hello, World!")
	key := []byte("test-key-12345")

	ciphertext1, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("Encrypt() first error = %v", err)
	}

	ciphertext2, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("Encrypt() second error = %v", err)
	}

	// Should be different due to random nonce
	if ciphertext1 == ciphertext2 {
		t.Error("Encrypt() twice with same key produced same ciphertext (nonce should be random)")
	}
}

// TestDecrypt_invalidBase64 verifies invalid base64 is rejected.
func TestDecrypt_invalidBase64(t *testing.T) {
	key := []byte("test-key-12345")

	tests := []struct {
		name  string
		input string
	}{
		{"not base64", "not-valid-base64!!!"},
		{"empty string", ""},
		{"special chars", "!@#$%^&*()"},
		{"incomplete base64", "YWJj"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Decrypt(tt.input, key)
			if err != ErrInvalidCiphertext {
				t.Errorf("Decrypt() error = %v, want ErrInvalidCiphertext", err)
			}
		})
	}
}

// TestDecrypt_wrongKey verifies wrong key fails decryption.
func TestDecrypt_wrongKey(t *testing.T) {
	plaintext := []byte("Hello, World!")
	key1 := []byte("key-one")
	key2 := []byte("key-two")

	ciphertext, err := Encrypt(plaintext, key1)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	_, err = Decrypt(ciphertext, key2)
	if err != ErrInvalidCiphertext {
		t.Errorf("Decrypt() with wrong key error = %v, want ErrInvalidCiphertext", err)
	}
}

// TestDecrypt_tamperedCiphertext verifies modified ciphertext is rejected.
func TestDecrypt_tamperedCiphertext(t *testing.T) {
	plaintext := []byte("Hello, World!")
	key := []byte("test-key-12345")

	ciphertext, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	// Tamper with the ciphertext by changing some characters
	tampered := strings.ToUpper(ciphertext[:10]) + ciphertext[10:]

	_, err = Decrypt(tampered, key)
	if err != ErrInvalidCiphertext {
		t.Errorf("Decrypt() with tampered ciphertext error = %v, want ErrInvalidCiphertext", err)
	}
}

// TestEncrypt_emptyPlaintext verifies empty plaintext works.
func TestEncrypt_emptyPlaintext(t *testing.T) {
	plaintext := []byte("")
	key := []byte("test-key-12345")

	ciphertext, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	decrypted, err := Decrypt(ciphertext, key)
	if err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}

	if string(decrypted) != string(plaintext) {
		t.Errorf("Decrypt() = %q, want %q", string(decrypted), string(plaintext))
	}
}

// TestEncrypt_largeData verifies encryption of larger data.
func TestEncrypt_largeData(t *testing.T) {
	// Create 1KB of data
	plaintext := make([]byte, 1024)
	for i := range plaintext {
		plaintext[i] = byte(i % 256)
	}
	key := []byte("test-key-12345")

	ciphertext, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	decrypted, err := Decrypt(ciphertext, key)
	if err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}

	if string(decrypted) != string(plaintext) {
		t.Error("Decrypt() of large data does not match original")
	}
}

// TestEncryptString_roundtrip verifies string wrapper functions.
func TestEncryptString_roundtrip(t *testing.T) {
	plaintext := "Hello, World!"
	key := "test-key-12345"

	ciphertext, err := EncryptString(plaintext, key)
	if err != nil {
		t.Fatalf("EncryptString() error = %v", err)
	}

	decrypted, err := DecryptString(ciphertext, key)
	if err != nil {
		t.Fatalf("DecryptString() error = %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("DecryptString() = %q, want %q", decrypted, plaintext)
	}
}

// TestEncryptString_emptyKey verifies empty key is rejected.
func TestEncryptString_emptyKey(t *testing.T) {
	_, err := EncryptString("plaintext", "")
	if err != ErrInvalidKey {
		t.Errorf("EncryptString() with empty key error = %v, want ErrInvalidKey", err)
	}
}

// TestDecryptString_emptyKey verifies empty key is rejected.
func TestDecryptString_emptyKey(t *testing.T) {
	_, err := DecryptString("ciphertext", "")
	if err != ErrInvalidKey {
		t.Errorf("DecryptString() with empty key error = %v, want ErrInvalidKey", err)
	}
}

// TestEncryptString_unicode verifies unicode strings work correctly.
func TestEncryptString_unicode(t *testing.T) {
	tests := []struct {
		name string
		text string
	}{
		{"chinese", "你好世界"},
		{"japanese", "こんにちは"},
		{"korean", "안녕하세요"},
		{"emoji", "👋🌍🎉"},
		{"mixed", "Hello 你好 123 🌍"},
	}

	key := "test-key-12345"

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ciphertext, err := EncryptString(tt.text, key)
			if err != nil {
				t.Fatalf("EncryptString() error = %v", err)
			}

			decrypted, err := DecryptString(ciphertext, key)
			if err != nil {
				t.Fatalf("DecryptString() error = %v", err)
			}

			if decrypted != tt.text {
				t.Errorf("DecryptString() = %q, want %q", decrypted, tt.text)
			}
		})
	}
}

// TestDeriveKey_consistency verifies same input produces same key.
func TestDeriveKey_consistency(t *testing.T) {
	machineID := "machine-123"

	key1 := DeriveKey(machineID)
	key2 := DeriveKey(machineID)

	// Should produce identical keys
	if string(key1) != string(key2) {
		t.Error("DeriveKey() produced different keys for same input")
	}

	// Key should be 32 bytes (SHA-256 output)
	if len(key1) != 32 {
		t.Errorf("DeriveKey() key length = %d, want 32", len(key1))
	}
}

// TestDeriveKey_differentInputs verifies different inputs produce different keys.
func TestDeriveKey_differentInputs(t *testing.T) {
	key1 := DeriveKey("machine-1")
	key2 := DeriveKey("machine-2")

	if string(key1) == string(key2) {
		t.Error("DeriveKey() produced same keys for different inputs")
	}
}

// TestGetMachineKey_withID verifies key derivation with machine ID.
func TestGetMachineKey_withID(t *testing.T) {
	machineID := "test-machine-123"

	key := GetMachineKey(machineID)

	if len(key) != 32 {
		t.Errorf("GetMachineKey() length = %d, want 32", len(key))
	}

	// Should be consistent with DeriveKey
	expected := DeriveKey(machineID)
	if string(key) != string(expected) {
		t.Error("GetMachineKey() does not match DeriveKey()")
	}
}

// TestGetMachineKey_emptyID verifies default key is used when ID is empty.
func TestGetMachineKey_emptyID(t *testing.T) {
	key1 := GetMachineKey("")
	key2 := GetMachineKey("")

	if string(key1) != string(key2) {
		t.Error("GetMachineKey() with empty ID produced different keys")
	}

	// Should also match explicit default
	key3 := GetMachineKey("memonexus-default-key")
	if string(key1) != string(key3) {
		t.Error("GetMachineKey() empty ID does not match explicit default key")
	}
}

// TestEncryptAPIKey_roundtrip verifies API key encryption and decryption.
func TestEncryptAPIKey_roundtrip(t *testing.T) {
	apiKey := "sk-1234567890abcdefghijklmnopqrstuvwxyz"
	machineID := "machine-123"

	encrypted, err := EncryptAPIKey(apiKey, machineID)
	if err != nil {
		t.Fatalf("EncryptAPIKey() error = %v", err)
	}

	decrypted, err := DecryptAPIKey(encrypted, machineID)
	if err != nil {
		t.Fatalf("DecryptAPIKey() error = %v", err)
	}

	if decrypted != apiKey {
		t.Errorf("DecryptAPIKey() = %q, want %q", decrypted, apiKey)
	}
}

// TestEncryptAPIKey_emptyKey verifies empty API key is rejected.
func TestEncryptAPIKey_emptyKey(t *testing.T) {
	_, err := EncryptAPIKey("", "machine-123")
	if err == nil {
		t.Error("EncryptAPIKey() with empty key should return error")
	}
	if !strings.Contains(err.Error(), "cannot be empty") {
		t.Errorf("EncryptAPIKey() error = %v, should mention empty", err)
	}
}

// TestDecryptAPIKey_emptyCiphertext verifies empty ciphertext returns empty.
func TestDecryptAPIKey_emptyCiphertext(t *testing.T) {
	result, err := DecryptAPIKey("", "machine-123")
	if err != nil {
		t.Fatalf("DecryptAPIKey() error = %v", err)
	}
	if result != "" {
		t.Errorf("DecryptAPIKey() with empty ciphertext = %q, want empty", result)
	}
}

// TestDecryptAPIKey_wrongMachineID verifies wrong machine ID fails decryption.
func TestDecryptAPIKey_wrongMachineID(t *testing.T) {
	apiKey := "sk-1234567890abcdefghijklmnopqrstuvwxyz"
	machineID1 := "machine-1"
	machineID2 := "machine-2"

	encrypted, err := EncryptAPIKey(apiKey, machineID1)
	if err != nil {
		t.Fatalf("EncryptAPIKey() error = %v", err)
	}

	_, err = DecryptAPIKey(encrypted, machineID2)
	if err != ErrInvalidCiphertext {
		t.Errorf("DecryptAPIKey() with wrong machine ID error = %v, want ErrInvalidCiphertext", err)
	}
}

// TestDecryptAPIKey_emptyMachineID verifies default key works for empty machine ID.
func TestDecryptAPIKey_emptyMachineID(t *testing.T) {
	apiKey := "sk-1234567890abcdefghijklmnopqrstuvwxyz"

	encrypted, err := EncryptAPIKey(apiKey, "")
	if err != nil {
		t.Fatalf("EncryptAPIKey() error = %v", err)
	}

	decrypted, err := DecryptAPIKey(encrypted, "")
	if err != nil {
		t.Fatalf("DecryptAPIKey() error = %v", err)
	}

	if decrypted != apiKey {
		t.Errorf("DecryptAPIKey() = %q, want %q", decrypted, apiKey)
	}
}

// TestErrInvalidCiphertext verifies error variable is exported correctly.
func TestErrInvalidCiphertext(t *testing.T) {
	if ErrInvalidCiphertext == nil {
		t.Error("ErrInvalidCiphertext should not be nil")
	}
	if ErrInvalidCiphertext.Error() == "" {
		t.Error("ErrInvalidCiphertext should have error message")
	}
}

// TestErrInvalidKey verifies error variable is exported correctly.
func TestErrInvalidKey(t *testing.T) {
	if ErrInvalidKey == nil {
		t.Error("ErrInvalidKey should not be nil")
	}
	if ErrInvalidKey.Error() == "" {
		t.Error("ErrInvalidKey should have error message")
	}
}
