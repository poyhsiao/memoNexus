// Package sync tests for S3 client utility functions.
package sync

import (
	"net/http"
	"testing"

	"github.com/kimhsiao/memonexus/backend/internal/errors"
)

// TestTruncateString verifies string truncation.
func TestTruncateString(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		maxLen   int
		expected string
	}{
		{
			name:     "shorter than max",
			s:        "hello",
			maxLen:   10,
			expected: "hello",
		},
		{
			name:     "exact length",
			s:        "hello",
			maxLen:   5,
			expected: "hello",
		},
		{
			name:     "needs truncation",
			s:        "hello world test",
			maxLen:   10,
			expected: "hello worl...",
		},
		{
			name:     "empty string",
			s:        "",
			maxLen:   10,
			expected: "",
		},
		{
			name:     "max len zero",
			s:        "hello",
			maxLen:   0,
			expected: "...",
		},
		{
			name:     "unicode text",
			s:        "你好世界测试",
			maxLen:   8,
			expected: "你好\xe4\xb8...", // Byte boundary truncation (each Chinese char is 3 bytes)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateString(tt.s, tt.maxLen)
			if result != tt.expected {
				t.Errorf("truncateString(%q, %d) = %q, want %q", tt.s, tt.maxLen, result, tt.expected)
			}
		})
	}
}

// TestCategorizeHTTPError verifies HTTP error categorization.
func TestCategorizeHTTPError(t *testing.T) {
	client := &S3Client{}

	tests := []struct {
		name       string
		statusCode int
		body       string
		expected   errors.ErrorCode
	}{
		{
			name:       "401 unauthorized",
			statusCode: http.StatusUnauthorized,
			body:       "",
			expected:   errors.ErrSyncAuthFailed,
		},
		{
			name:       "403 forbidden - signature mismatch",
			statusCode: http.StatusForbidden,
			body:       "SignatureDoesNotMatch",
			expected:   errors.ErrSyncAuthFailed,
		},
		{
			name:       "403 forbidden - invalid key",
			statusCode: http.StatusForbidden,
			body:       "InvalidAccessKeyId",
			expected:   errors.ErrSyncAuthFailed,
		},
		{
			name:       "403 forbidden - access denied",
			statusCode: http.StatusForbidden,
			body:       "AccessDenied",
			expected:   errors.ErrSyncAuthFailed,
		},
		{
			name:       "403 forbidden - other",
			statusCode: http.StatusForbidden,
			body:       "SomeOtherError",
			expected:   errors.ErrSyncFailed,
		},
		{
			name:       "503 service unavailable - slow down",
			statusCode: 503,
			body:       "SlowDown",
			expected:   errors.ErrSyncQuotaExceeded,
		},
		{
			name:       "503 service unavailable - quota",
			statusCode: 503,
			body:       "Quota exceeded",
			expected:   errors.ErrSyncQuotaExceeded,
		},
		{
			name:       "503 service unavailable - other",
			statusCode: 503,
			body:       "Service temporarily unavailable",
			expected:   errors.ErrSyncFailed,
		},
		{
			name:       "500 internal server error",
			statusCode: 500,
			body:       "Internal server error",
			expected:   errors.ErrSyncFailed,
		},
		{
			name:       "404 not found",
			statusCode: 404,
			body:       "Not found",
			expected:   errors.ErrSyncFailed,
		},
		{
			name:       "200 OK",
			statusCode: 200,
			body:       "OK",
			expected:   errors.ErrSyncFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.categorizeHTTPError(tt.statusCode, tt.body)
			if result != tt.expected {
				t.Errorf("categorizeHTTPError(%d, %q) = %v, want %v", tt.statusCode, tt.body, result, tt.expected)
			}
		})
	}
}

// TestCategorizeHTTPError_caseSensitivity verifies body matching is case-sensitive.
func TestCategorizeHTTPError_caseSensitivity(t *testing.T) {
	client := &S3Client{}

	tests := []struct {
		name       string
		statusCode int
		body       string
		expected   errors.ErrorCode
	}{
		{
			name:       "lowercase signaturedoesnotmatch",
			statusCode: http.StatusForbidden,
			body:       "signaturedoesnotmatch",
			expected:   errors.ErrSyncFailed, // Case-sensitive, no match
		},
		{
			name:       "uppercase SIGNATUREDOESNOTMATCH",
			statusCode: http.StatusForbidden,
			body:       "SIGNATUREDOESNOTMATCH",
			expected:   errors.ErrSyncFailed, // Case-sensitive, no match
		},
		{
			name:       "lowercase slowdown",
			statusCode: 503,
			body:       "slowdown",
			expected:   errors.ErrSyncFailed, // Case-sensitive, no match
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.categorizeHTTPError(tt.statusCode, tt.body)
			if result != tt.expected {
				t.Errorf("categorizeHTTPError(%d, %q) = %v, want %v", tt.statusCode, tt.body, result, tt.expected)
			}
		})
	}
}
