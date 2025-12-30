// Package uuid provides UUID v4 generation and validation utilities.
package uuid

import (
	"fmt"

	"github.com/google/uuid"
)

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
func IsValid(s string) bool {
	_, err := NewFromString(s)
	return err == nil
}

// Validate returns an error if the string is not a valid UUID v4.
func Validate(s string) error {
	_, err := NewFromString(s)
	return err
}
