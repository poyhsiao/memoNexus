// Package s3 provides unit tests for Cloudflare R2 provider.
// T153: Unit test for R2 provider configuration.
package s3

import (
	"testing"
)

// TestNewR2Client tests creating an R2 client.
func TestNewR2Client(t *testing.T) {
	config := &R2Config{
		AccountID:  "abc123def4567890abc123def4567890", // 32-char hex
		BucketName: "test-bucket",
		AccessKey:  "access_key_id",
		SecretKey:  "secret_access_key",
	}

	client := NewR2Client(config)

	if client == nil {
		t.Fatal("Expected non-nil client")
	}
}

// TestNewR2ClientPanicWithMissingAccountID tests that NewR2Client panics without account ID.
func TestNewR2ClientPanicWithMissingAccountID(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when AccountID is empty")
		}
	}()

	config := &R2Config{
		AccountID:  "", // Empty account ID
		BucketName: "test-bucket",
		AccessKey:  "access_key",
		SecretKey:  "secret_key",
	}

	_ = NewR2Client(config)
}

// TestR2EndpointForAccount tests generating R2 endpoints.
func TestR2EndpointForAccount(t *testing.T) {
	tests := []struct {
		name      string
		accountID string
		expected  string
	}{
		{
			name:      "standard account ID",
			accountID: "abc123def4567890abc123def4567890",
			expected:  "abc123def4567890abc123def4567890.r2.cloudflarestorage.com",
		},
		{
			name:      "uppercase account ID",
			accountID: "ABC123DEF4567890ABC123DEF4567890",
			expected:  "ABC123DEF4567890ABC123DEF4567890.r2.cloudflarestorage.com",
		},
		{
			name:      "mixed case account ID",
			accountID: "AbC123DeF4567890AbC123DeF4567890",
			expected:  "AbC123DeF4567890AbC123DeF4567890.r2.cloudflarestorage.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			endpoint := R2EndpointForAccount(tt.accountID)
			if endpoint != tt.expected {
				t.Errorf("Expected endpoint %s, got %s", tt.expected, endpoint)
			}
		})
	}
}

// TestIsValidR2AccountID tests R2 account ID validation.
func TestIsValidR2AccountID(t *testing.T) {
	tests := []struct {
		name      string
		accountID string
		valid     bool
	}{
		{"valid lowercase", "abc123def4567890abc123def4567890", true},
		{"valid uppercase", "ABC123DEF4567890ABC123DEF4567890", true},
		{"valid mixed case", "AbC123DeF4567890AbC123DeF4567890", true},
		{"too short", "abc123", false},
		{"too long", "abc123def4567890abc123def4567890abc123def4567890", false},
		{"invalid characters", "ghijklmnopqrstuvwxyz1234567890abcdef", false}, // Contains g-z (not hex)
		{"empty string", "", false},
		{"contains special chars", "abc123-def4567890abc123def4567890", false},
		{"contains spaces", "abc123def4567890abc 123def4567890", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidR2AccountID(tt.accountID)
			if result != tt.valid {
				t.Errorf("IsValidR2AccountID(%s) = %v, want %v", tt.accountID, result, tt.valid)
			}
		})
	}
}

// TestR2PublicURL tests generating public URLs.
func TestR2PublicURL(t *testing.T) {
	tests := []struct {
		name         string
		customDomain string
		key          string
		expected     string
	}{
		{
			name:         "basic URL",
			customDomain: "cdn.example.com",
			key:          "items/file.json",
			expected:     "https://cdn.example.com/items/file.json",
		},
		{
			name:         "nested path",
			customDomain: "files.example.com",
			key:          "items/2024/01/document.pdf",
			expected:     "https://files.example.com/items/2024/01/document.pdf",
		},
		{
			name:         "root level file",
			customDomain: "example.com",
			key:          "index.html",
			expected:     "https://example.com/index.html",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := R2PublicURL(tt.customDomain, tt.key)
			if url != tt.expected {
				t.Errorf("Expected URL %s, got %s", tt.expected, url)
			}
		})
	}
}

// TestR2S3URL tests generating S3 API URLs.
func TestR2S3URL(t *testing.T) {
	accountID := "abc123def4567890abc123def4567890"
	bucket := "my-bucket"
	key := "items/file.json"

	expected := "https://my-bucket.abc123def4567890abc123def4567890.r2.cloudflarestorage.com/items/file.json"
	url := R2S3URL(accountID, bucket, key)

	if url != expected {
		t.Errorf("Expected URL %s, got %s", expected, url)
	}
}
