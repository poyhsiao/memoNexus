// Package uuid provides UUID v4 generation and validation utilities.
package uuid

import (
	"fmt"
	"regexp"

	"github.com/google/uuid"
)

// UUID v4 format: xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx
// where y is one of [8, 9, a, b] (variant bits)
var uuidV4Regex = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-4[0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`)

// New generates a new UUID v4.
func New() string {
	return uuid.New().String()
}

// NewFromString creates a UUID from a string.
// Returns an error if the string is not a valid UUID v4.
func NewFromString(s string) (uuid.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid UUID: %w", err)
	}
	if id.Version() != 4 {
		return uuid.Nil, fmt.Errorf("expected UUID v4, got v%d", id.Version())
	}
	return id, nil
}

// IsValid checks if a string is a valid UUID v4.
// Enforces strict format with dashes and correct variant bits.
func IsValid(s string) bool {
	return uuidV4Regex.MatchString(s)
}

// Validate returns an error if the string is not a valid UUID v4.
func Validate(s string) error {
	if !IsValid(s) {
		return fmt.Errorf("invalid UUID v4 format: %q", s)
	}
	return nil
}
