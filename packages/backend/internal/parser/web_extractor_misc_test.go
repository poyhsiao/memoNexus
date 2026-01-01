// Package parser tests for web extractor internal functions.
package parser

import (
	"testing"
)

// TestCleanText verifies text normalization (internal function).
func TestCleanText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple text",
			input:    "hello world",
			expected: "hello world",
		},
		{
			name:     "multiple spaces",
			input:    "hello    world",
			expected: "hello world",
		},
		{
			name:     "tabs and newlines",
			input:    "hello\t\tworld\n\ntest",
			expected: "hello world test",
		},
		{
			name:     "leading/trailing whitespace",
			input:    "  hello world  ",
			expected: "hello world",
		},
		{
			name:     "mixed whitespace",
			input: "  hello    \t\tworld\n\n  test  ",
			expected: "hello world test",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only whitespace",
			input: "   \t\n   ",
			expected: "",
		},
		{
			name:     "CJK text with spaces",
			input:    "你好    世界",
			expected: "你好 世界",
		},
		{
			name:     "single word",
			input:    "hello",
			expected: "hello",
		},
		{
			name:     "single word with spaces",
			input: "   hello   ",
			expected: "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanText(tt.input)
			if result != tt.expected {
				t.Errorf("cleanText(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

