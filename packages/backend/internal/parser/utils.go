// Package parser provides utility functions for content parsing.
package parser

import (
	"strings"
)

// CountWords counts words in text.
func CountWords(s string) int {
	words := strings.Fields(s)
	return len(words)
}

// DetectLanguage detects the language of text (heuristic).
func DetectLanguage(s string) string {
	// Simple heuristic: check for non-ASCII characters
	hasNonASCII := false
	for _, r := range s {
		if r > 127 {
			hasNonASCII = true
			break
		}
	}

	if hasNonASCII {
		// Could add more sophisticated detection
		return "unknown"
	}

	return "en"
}

// Truncate truncates string to max length.
func Truncate(s string, maxLen int) string {
	// Handle invalid maxLen
	if maxLen <= 0 {
		return "..."
	}

	if len(s) <= maxLen {
		return s
	}

	// Try to truncate at word boundary
	if i := strings.LastIndex(s[:maxLen], " "); i > 0 {
		return s[:i] + "..."
	}

	return s[:maxLen] + "..."
}

// BasenameFromURLFull extracts a basename from URL for default title.
// This is the version with full URL parsing (with hostname fallback).
func BasenameFromURLFull(sourceURL string) string {
	u := sourceURL
	if i := strings.Index(u, "?"); i > 0 {
		u = u[:i]
	}
	if i := strings.Index(u, "#"); i > 0 {
		u = u[:i]
	}

	// Get last path segment
	if i := strings.LastIndex(u, "/"); i >= 0 {
		base := u[i+1:]
		if base != "" {
			return base
		}
	}

	return "Untitled"
}
