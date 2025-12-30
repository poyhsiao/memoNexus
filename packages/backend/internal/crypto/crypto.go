// Package crypto provides simple encryption for sensitive data like API keys.
// Uses AES-256-GCM for authenticated encryption.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
)

var (
	// ErrInvalidCiphertext is returned when decryption fails.
	ErrInvalidCiphertext = errors.New("invalid ciphertext")
	// ErrInvalidKey is returned when the key is invalid.
	ErrInvalidKey = errors.New("invalid key")
)

// Encrypt encrypts plaintext using AES-256-GCM.
// The key is derived from the input using SHA-256.
func Encrypt(plaintext, key []byte) (string, error) {
	// Derive a 32-byte key from the input key
	derivedKey := sha256.Sum256(key)

	block, err := aes.NewCipher(derivedKey[:])
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Generate a random nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// Encrypt and authenticate
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	// Encode as base64 for storage
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts ciphertext that was encrypted with Encrypt.
func Decrypt(ciphertext string, key []byte) ([]byte, error) {
	// Derive the same key
	derivedKey := sha256.Sum256(key)

	// Decode base64
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return nil, ErrInvalidCiphertext
	}

	block, err := aes.NewCipher(derivedKey[:])
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Check minimum length
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, ErrInvalidCiphertext
	}

	// Extract nonce and ciphertext
	nonce, cipherData := data[:nonceSize], data[nonceSize:]

	// Decrypt and verify
	plaintext, err := gcm.Open(nil, nonce, cipherData, nil)
	if err != nil {
		return nil, ErrInvalidCiphertext
	}

	return plaintext, nil
}

// EncryptString encrypts a string to a base64-encoded string.
func EncryptString(plaintext, key string) (string, error) {
	if key == "" {
		return "", ErrInvalidKey
	}
	return Encrypt([]byte(plaintext), []byte(key))
}

// DecryptString decrypts a base64-encoded string to a string.
func DecryptString(ciphertext, key string) (string, error) {
	if key == "" {
		return "", ErrInvalidKey
	}
	plaintext, err := Decrypt(ciphertext, []byte(key))
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

// DeriveKey derives a consistent key from a machine-specific identifier.
// This is a simple implementation - in production, use platform-specific key stores.
func DeriveKey(machineID string) []byte {
	// Use SHA-256 to derive a key from the machine ID
	hash := sha256.Sum256([]byte("memonexus:" + machineID))
	return hash[:]
}

// GetMachineKey returns a key derived from a machine identifier.
// Falls back to a default key if no machine ID is provided.
func GetMachineKey(machineID string) []byte {
	if machineID == "" {
		machineID = "memonexus-default-key"
	}
	return DeriveKey(machineID)
}

// EncryptAPIKey encrypts an API key for storage.
func EncryptAPIKey(apiKey, machineID string) (string, error) {
	if apiKey == "" {
		return "", fmt.Errorf("API key cannot be empty")
	}
	key := GetMachineKey(machineID)
	return EncryptString(apiKey, string(key))
}

// DecryptAPIKey decrypts a stored API key.
func DecryptAPIKey(encryptedKey, machineID string) (string, error) {
	if encryptedKey == "" {
		return "", nil // Empty means no key set
	}
	key := GetMachineKey(machineID)
	return DecryptString(encryptedKey, string(key))
}
