// Package errors provides error code definitions for Go-Dart boundary bridging.
package errors

import "fmt"

// ErrorCode represents a unique error code that can be bridged to Dart.
type ErrorCode string

const (
	// General errors
	ErrInternal    ErrorCode = "INTERNAL_ERROR"
	ErrInvalid     ErrorCode = "INVALID_INPUT"
	ErrNotFound    ErrorCode = "NOT_FOUND"
	ErrDuplicate   ErrorCode = "DUPLICATE"
	ErrPermission  ErrorCode = "PERMISSION_DENIED"
	ErrValidation  ErrorCode = "VALIDATION_ERROR"

	// Database errors
	ErrDatabase    ErrorCode = "DATABASE_ERROR"
	ErrMigration   ErrorCode = "MIGRATION_FAILED"
	ErrConstraint  ErrorCode = "CONSTRAINT_VIOLATION"

	// Content errors
	ErrContentNotFound    ErrorCode = "CONTENT_NOT_FOUND"
	ErrContentInvalid     ErrorCode = "CONTENT_INVALID"
	ErrContentDuplicate   ErrorCode = "CONTENT_DUPLICATE"

	// Tag errors
	ErrTagNotFound ErrorCode = "TAG_NOT_FOUND"
	ErrTagInvalid  ErrorCode = "TAG_INVALID"

	// Sync errors
	ErrSyncNotConfigured  ErrorCode = "SYNC_NOT_CONFIGURED"
	ErrSyncFailed         ErrorCode = "SYNC_FAILED"
	ErrSyncConflict       ErrorCode = "SYNC_CONFLICT"
	ErrSyncAuthFailed     ErrorCode = "SYNC_AUTH_FAILED"
	ErrSyncQuotaExceeded  ErrorCode = "SYNC_QUOTA_EXCEEDED"
	ErrSyncTimeout        ErrorCode = "SYNC_TIMEOUT"

	// AI errors
	ErrAINotConfigured     ErrorCode = "AI_NOT_CONFIGURED"
	ErrAIFailed            ErrorCode = "AI_FAILED"
	ErrAIRateLimit         ErrorCode = "AI_RATE_LIMIT"
	ErrAITimeout           ErrorCode = "AI_TIMEOUT"
	ErrAIInvalidCredentials ErrorCode = "AI_INVALID_CREDENTIALS"

	// Export errors
	ErrExportFailed      ErrorCode = "EXPORT_FAILED"
	ErrImportFailed      ErrorCode = "IMPORT_FAILED"
	ErrInvalidPassword   ErrorCode = "INVALID_PASSWORD"
	ErrCorruptedArchive  ErrorCode = "CORRUPTED_ARCHIVE"
	ErrCryptoFailed      ErrorCode = "CRYPTO_FAILED"
)

// AppError represents an application error with code and message.
type AppError struct {
	Code    ErrorCode
	Message string
	Err     error
}

// Error implements the error interface.
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the underlying error.
func (e *AppError) Unwrap() error {
	return e.Err
}

// New creates a new AppError.
func New(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

// Wrap wraps an existing error with an error code.
func Wrap(code ErrorCode, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// Is checks if an error is of a specific code.
func Is(err error, code ErrorCode) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code == code
	}
	return false
}
