// Package errors tests for error code definitions and error handling.
package errors

import (
	"errors"
	"strings"
	"testing"
)

// TestErrorCodeValues verifies all error codes have non-empty values.
func TestErrorCodeValues(t *testing.T) {
	tests := []struct {
		name string
		code ErrorCode
	}{
		// General errors
		{"internal", ErrInternal},
		{"invalid", ErrInvalid},
		{"not found", ErrNotFound},
		{"duplicate", ErrDuplicate},
		{"permission", ErrPermission},
		{"validation", ErrValidation},

		// Database errors
		{"database", ErrDatabase},
		{"migration", ErrMigration},
		{"constraint", ErrConstraint},

		// Content errors
		{"content not found", ErrContentNotFound},
		{"content invalid", ErrContentInvalid},
		{"content duplicate", ErrContentDuplicate},

		// Tag errors
		{"tag not found", ErrTagNotFound},
		{"tag invalid", ErrTagInvalid},

		// Sync errors
		{"sync not configured", ErrSyncNotConfigured},
		{"sync failed", ErrSyncFailed},
		{"sync conflict", ErrSyncConflict},
		{"sync auth failed", ErrSyncAuthFailed},
		{"sync quota exceeded", ErrSyncQuotaExceeded},
		{"sync timeout", ErrSyncTimeout},

		// AI errors
		{"AI not configured", ErrAINotConfigured},
		{"AI failed", ErrAIFailed},
		{"AI rate limit", ErrAIRateLimit},
		{"AI timeout", ErrAITimeout},
		{"AI invalid credentials", ErrAIInvalidCredentials},

		// Export errors
		{"export failed", ErrExportFailed},
		{"import failed", ErrImportFailed},
		{"invalid password", ErrInvalidPassword},
		{"corrupted archive", ErrCorruptedArchive},
		{"crypto failed", ErrCryptoFailed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.code == "" {
				t.Errorf("ErrorCode %q should not be empty", tt.name)
			}
		})
	}
}

// TestAppError_Error verifies error message formatting.
func TestAppError_Error(t *testing.T) {
	tests := []struct {
		name     string
		appError *AppError
		want     string
	}{
		{
			name:     "error without underlying error",
			appError: &AppError{Code: ErrInternal, Message: "something failed"},
			want:     "[INTERNAL_ERROR] something failed",
		},
		{
			name:     "error with underlying error",
			appError: &AppError{Code: ErrDatabase, Message: "query failed", Err: errors.New("connection lost")},
			want:     "[DATABASE_ERROR] query failed: connection lost",
		},
		{
			name:     "not found error",
			appError: &AppError{Code: ErrNotFound, Message: "item not found"},
			want:     "[NOT_FOUND] item not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.appError.Error()
			if got != tt.want {
				t.Errorf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestAppError_Unwrap verifies unwrapping of underlying error.
func TestAppError_Unwrap(t *testing.T) {
	underlyingErr := errors.New("underlying error")

	tests := []struct {
		name          string
		appError      *AppError
		wantUnwrapped error
	}{
		{
			name:          "with underlying error",
			appError:      &AppError{Code: ErrInternal, Message: "failed", Err: underlyingErr},
			wantUnwrapped: underlyingErr,
		},
		{
			name:          "without underlying error",
			appError:      &AppError{Code: ErrInternal, Message: "failed"},
			wantUnwrapped: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.appError.Unwrap()
			if got != tt.wantUnwrapped {
				t.Errorf("Unwrap() = %v, want %v", got, tt.wantUnwrapped)
			}
		})
	}
}

// TestNew verifies AppError creation.
func TestNew(t *testing.T) {
	err := New(ErrInternal, "test error")
	if err == nil {
		t.Fatal("New() returned nil")
	}
	if err.Code != ErrInternal {
		t.Errorf("New() code = %q, want %q", err.Code, ErrInternal)
	}
	if err.Message != "test error" {
		t.Errorf("New() message = %q, want 'test error'", err.Message)
	}
	if err.Err != nil {
		t.Error("New() should not wrap an error")
	}
}

// TestWrap verifies error wrapping.
func TestWrap(t *testing.T) {
	underlyingErr := errors.New("underlying")

	err := Wrap(ErrDatabase, "query failed", underlyingErr)
	if err == nil {
		t.Fatal("Wrap() returned nil")
	}
	if err.Code != ErrDatabase {
		t.Errorf("Wrap() code = %q, want %q", err.Code, ErrDatabase)
	}
	if err.Message != "query failed" {
		t.Errorf("Wrap() message = %q, want 'query failed'", err.Message)
	}
	if err.Err != underlyingErr {
		t.Errorf("Wrap() underlying error = %v, want %v", err.Err, underlyingErr)
	}

	// Verify error implements error interface
	var _ error = err
	if err.Error() == "" {
		t.Error("Wrap() error message should not be empty")
	}
}

// TestWrap_withNilError verifies wrapping nil error.
func TestWrap_withNilError(t *testing.T) {
	err := Wrap(ErrInternal, "test", nil)
	if err.Err != nil {
		t.Errorf("Wrap() with nil error should have nil Err, got %v", err.Err)
	}
}

// TestIs verifies error code checking.
func TestIs(t *testing.T) {
	tests := []struct {
		name  string
		err   error
		code  ErrorCode
		want  bool
	}{
		{
			name: "matching AppError",
			err:  &AppError{Code: ErrNotFound, Message: "not found"},
			code: ErrNotFound,
			want: true,
		},
		{
			name: "non-matching AppError",
			err:  &AppError{Code: ErrNotFound, Message: "not found"},
			code: ErrInternal,
			want: false,
		},
		{
			name: "non-AppError",
			err:  errors.New("standard error"),
			code: ErrInternal,
			want: false,
		},
		{
			name:  "nil error",
			err:  nil,
			code: ErrInternal,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Is(tt.err, tt.code)
			if got != tt.want {
				t.Errorf("Is() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestErrorInterface verifies AppError implements error interface.
func TestErrorInterface(t *testing.T) {
	err := New(ErrInternal, "test")

	// Should implement error interface
	var _ error = err

	// Error() should return non-empty string
	if err.Error() == "" {
		t.Error("Error() should return non-empty string")
	}
}

// TestErrorCodes_areUnique verifies all error codes are unique.
func TestErrorCodes_areUnique(t *testing.T) {
	codes := []ErrorCode{
		ErrInternal, ErrInvalid, ErrNotFound, ErrDuplicate, ErrPermission, ErrValidation,
		ErrDatabase, ErrMigration, ErrConstraint,
		ErrContentNotFound, ErrContentInvalid, ErrContentDuplicate,
		ErrTagNotFound, ErrTagInvalid,
		ErrSyncNotConfigured, ErrSyncFailed, ErrSyncConflict, ErrSyncAuthFailed, ErrSyncQuotaExceeded, ErrSyncTimeout,
		ErrAINotConfigured, ErrAIFailed, ErrAIRateLimit, ErrAITimeout, ErrAIInvalidCredentials,
		ErrExportFailed, ErrImportFailed, ErrInvalidPassword, ErrCorruptedArchive, ErrCryptoFailed,
	}

	seen := make(map[ErrorCode]bool)
	for _, code := range codes {
		if seen[code] {
			t.Errorf("ErrorCode %q is duplicated", code)
		}
		seen[code] = true
	}
}

// TestErrorCodes_areStringType verifies error codes are string-based.
func TestErrorCodes_areStringType(t *testing.T) {
	code := ErrInternal
	if string(code) != "INTERNAL_ERROR" {
		t.Errorf("ErrorCode should be string-based, got %q", string(code))
	}
}

// TestCommonErrorCodes verifies commonly used error codes.
func TestCommonErrorCodes(t *testing.T) {
	tests := []struct {
		code     ErrorCode
		expected string
	}{
		{ErrInternal, "INTERNAL_ERROR"},
		{ErrInvalid, "INVALID_INPUT"},
		{ErrNotFound, "NOT_FOUND"},
		{ErrDatabase, "DATABASE_ERROR"},
		{ErrSyncFailed, "SYNC_FAILED"},
		{ErrAIFailed, "AI_FAILED"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if string(tt.code) != tt.expected {
				t.Errorf("ErrorCode = %q, want %q", string(tt.code), tt.expected)
			}
		})
	}
}

// TestAppError_fields verifies AppError struct fields.
func TestAppError_fields(t *testing.T) {
	underlyingErr := errors.New("underlying")
	err := &AppError{
		Code:    ErrValidation,
		Message: "validation failed",
		Err:     underlyingErr,
	}

	if err.Code != ErrValidation {
		t.Errorf("Code = %q, want %q", err.Code, ErrValidation)
	}
	if err.Message != "validation failed" {
		t.Errorf("Message = %q, want 'validation failed'", err.Message)
	}
	if err.Err != underlyingErr {
		t.Errorf("Err = %v, want %v", err.Err, underlyingErr)
	}
}

// TestStandardErrorComparison verifies Is works with standard errors.
func TestStandardErrorComparison(t *testing.T) {
	appErr := New(ErrInternal, "app error")
	standardErr := errors.New("standard error")

	// AppError should match its own code
	if !Is(appErr, ErrInternal) {
		t.Error("Is() should return true for matching AppError")
	}

	// Standard error should not match
	if Is(standardErr, ErrInternal) {
		t.Error("Is() should return false for standard error")
	}
}

// TestErrorCode_prefix verifies error codes follow naming convention.
func TestErrorCode_prefix(t *testing.T) {
	codes := []ErrorCode{
		ErrInternal, ErrInvalid, ErrNotFound, ErrDuplicate, ErrPermission, ErrValidation,
		ErrDatabase, ErrMigration, ErrConstraint,
		ErrContentNotFound, ErrContentInvalid, ErrContentDuplicate,
		ErrTagNotFound, ErrTagInvalid,
		ErrSyncNotConfigured, ErrSyncFailed, ErrSyncConflict, ErrSyncAuthFailed, ErrSyncQuotaExceeded, ErrSyncTimeout,
		ErrAINotConfigured, ErrAIFailed, ErrAIRateLimit, ErrAITimeout, ErrAIInvalidCredentials,
		ErrExportFailed, ErrImportFailed, ErrInvalidPassword, ErrCorruptedArchive, ErrCryptoFailed,
	}

	for _, code := range codes {
		str := string(code)
		// Verify all caps with underscores
		if str != strings.ToUpper(str) {
			t.Errorf("ErrorCode %q should be uppercase", str)
		}
	}
}

// TestNew_withEmptyMessage verifies empty message is allowed.
func TestNew_withEmptyMessage(t *testing.T) {
	err := New(ErrInternal, "")
	if err.Message != "" {
		t.Errorf("New() with empty message should preserve it, got %q", err.Message)
	}
}

// TestError_formats verifies different error formats.
func TestError_formats(t *testing.T) {
	tests := []struct {
		name  string
		code  ErrorCode
		msg   string
		wrapped error
	}{
		{
			name:  "simple error",
			code:  ErrInternal,
			msg:   "Internal error occurred",
		},
		{
			name:  "validation error",
			code:  ErrValidation,
			msg:   "Invalid input parameter",
		},
		{
			name:    "wrapped error",
			code:    ErrDatabase,
			msg:     "Database query failed",
			wrapped: errors.New("connection timeout"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			if tt.wrapped != nil {
				err = Wrap(tt.code, tt.msg, tt.wrapped)
			} else {
				err = New(tt.code, tt.msg)
			}

			// Verify error string format
			errStr := err.Error()
			if errStr == "" {
				t.Error("Error() should return non-empty string")
			}

			// Verify code is in error string
			if !strings.Contains(errStr, string(tt.code)) {
				t.Errorf("Error() should contain code %q", tt.code)
			}

			// Verify message is in error string
			if !strings.Contains(errStr, tt.msg) {
				t.Errorf("Error() should contain message %q", tt.msg)
			}
		})
	}
}
