// Package sync tests for S3 client error categorization.
package sync

import (
	"fmt"
	"net/url"
	"testing"
	"time"

	apperrors "github.com/kimhsiao/memonexus/backend/internal/errors"
)

// TestCategorizeError verifies network error categorization.
func TestCategorizeError(t *testing.T) {
	client := &S3Client{}

	tests := []struct {
		name     string
		err      error
		expected apperrors.ErrorCode
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: apperrors.ErrSyncFailed,
		},
		{
			name:     "timeout error",
			err:      &testTimeoutError{},
			expected: apperrors.ErrSyncTimeout,
		},
		{
			name:     "URL error",
			err:      &url.Error{Op: "GET"},
			expected: apperrors.ErrSyncFailed,
		},
		{
			name:     "generic error",
			err:      fmt.Errorf("some error"),
			expected: apperrors.ErrSyncFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.categorizeError(tt.err)
			if result != tt.expected {
				t.Errorf("categorizeError() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// testTimeoutError implements Timeout() for testing.
type testTimeoutError struct{}

func (e *testTimeoutError) Error() string {
	return "timeout"
}

func (e *testTimeoutError) Timeout() bool {
	return true
}

// timeoutWrapper wraps an error and implements Timeout() interface.
type timeoutWrapper struct {
	error
}

func (e *timeoutWrapper) Timeout() bool {
	return true
}

// TestCategorizeError_actualTimeoutError tests with timeout interface.
func TestCategorizeError_actualTimeoutError(t *testing.T) {
	client := &S3Client{}

	// Verify that errors implementing Timeout() interface are detected
	err := &timeoutWrapper{error: fmt.Errorf("operation timed out")}

	result := client.categorizeError(err)
	if result != apperrors.ErrSyncTimeout {
		t.Errorf("categorizeError() with timeout interface = %v, want %v", result, apperrors.ErrSyncTimeout)
	}
}

// TestNewS3Client verifies S3 client creation.
func TestNewS3Client(t *testing.T) {
	config := &S3Config{
		Endpoint:       "https://s3.amazonaws.com",
		BucketName:     "test-bucket",
		AccessKey:      "test-access",
		SecretKey:      "test-secret",
		Region:         "us-east-1",
		ForcePathStyle: false,
	}

	client := NewS3Client(config)

	if client == nil {
		t.Fatal("NewS3Client() returned nil")
	}

	if client.config != config {
		t.Error("NewS3Client() config not set")
	}

	if client.httpClient == nil {
		t.Error("NewS3Client() httpClient is nil")
	}

	if client.httpClient.Timeout != 30*time.Second {
		t.Errorf("NewS3Client() timeout = %v, want 30s", client.httpClient.Timeout)
	}
}

// TestNewS3Client_defaultConfig verifies client with minimal config.
func TestNewS3Client_defaultConfig(t *testing.T) {
	config := &S3Config{
		BucketName: "test-bucket",
		// Other fields use zero values
	}

	client := NewS3Client(config)

	if client == nil {
		t.Fatal("NewS3Client() returned nil")
	}

	// Verify default timeout
	if client.httpClient.Timeout != 30*time.Second {
		t.Errorf("Default timeout = %v, want 30s", client.httpClient.Timeout)
	}
}

// TestS3Config verifies S3Config structure.
func TestS3Config(t *testing.T) {
	config := S3Config{
		Endpoint:       "https://s3.amazonaws.com",
		BucketName:     "my-bucket",
		AccessKey:      "AKIAIOSFODNN7EXAMPLE",
		SecretKey:      "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		Region:         "us-west-2",
		ForcePathStyle: true,
	}

	if config.Endpoint != "https://s3.amazonaws.com" {
		t.Errorf("Endpoint = %q", config.Endpoint)
	}
	if config.BucketName != "my-bucket" {
		t.Errorf("BucketName = %q", config.BucketName)
	}
	if config.AccessKey != "AKIAIOSFODNN7EXAMPLE" {
		t.Errorf("AccessKey = %q", config.AccessKey)
	}
	if config.SecretKey != "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY" {
		t.Errorf("SecretKey = %q", config.SecretKey)
	}
	if config.Region != "us-west-2" {
		t.Errorf("Region = %q", config.Region)
	}
	if !config.ForcePathStyle {
		t.Error("ForcePathStyle should be true")
	}
}
