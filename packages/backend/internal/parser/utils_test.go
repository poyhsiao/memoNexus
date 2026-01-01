// Package parser tests for utility functions.
package parser

import (
	"strings"
	"testing"
)

// TestCountWords verifies word counting functionality.
func TestCountWords(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected int
	}{
		{"empty", "", 0},
		{"single word", "hello", 1},
		{"multiple words", "hello world test", 3},
		{"with spaces", "  hello   world  ", 2},
		{"with newlines", "hello\nworld\ntest", 3},
		{"with tabs", "hello\tworld\ttest", 3},
		{"with mixed whitespace", "  hello\nworld\ttest  ", 3},
		{"punctuation", "hello, world! test.", 3},
		{"numbers", "123 456 789", 3},
		{"unicode", "hello 世界 test", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CountWords(tt.text)
			if result != tt.expected {
				t.Errorf("CountWords(%q) = %d, want %d", tt.text, result, tt.expected)
			}
		})
	}
}

// TestDetectLanguage verifies language detection.
func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected string
	}{
		{"empty", "", "en"},
		{"ascii only", "hello world", "en"},
		{"simple english", "The quick brown fox jumps over the lazy dog", "en"},
		{"with numbers", "test 123 abc", "en"},
		{"punctuation", "hello, world!", "en"},
		{"latin extended", "café résumé", "unknown"},
		{"chinese", "你好世界", "unknown"},
		{"japanese", "こんにちは", "unknown"},
		{"korean", "안녕하세요", "unknown"},
		{"cyrillic", "Привет мир", "unknown"},
		{"arabic", "مرحبا بالعالم", "unknown"},
		{"emoji", "Hello 👋 World", "unknown"},
		{"mixed", "Hello 世界", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectLanguage(tt.text)
			if result != tt.expected {
				t.Errorf("DetectLanguage(%q) = %q, want %q", tt.text, result, tt.expected)
			}
		})
	}
}

// TestTruncate verifies string truncation.
func TestTruncate(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		maxLen   int
		expected string
	}{
		{"shorter than max", "hello", 10, "hello"},
		{"exact length", "hello", 5, "hello"},
		{"needs truncation", "hello world", 8, "hello..."},
		{"truncate at word boundary", "hello world test", 11, "hello..."},
		{"no word boundary", "helloworld", 5, "hello..."},
		{"single word truncate", "supercalifragilistic", 10, "supercalif..."},
		{"empty string", "", 10, ""},
		{"max len zero", "hello", 0, "..."},
		{"multiple spaces", "hello  world test", 12, "hello ..."},
		{"trailing space", "hello ", 8, "hello "},
		{"leading space", " hello", 8, " hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Truncate(tt.text, tt.maxLen)
			if result != tt.expected {
				t.Errorf("Truncate(%q, %d) = %q, want %q", tt.text, tt.maxLen, result, tt.expected)
			}
		})
	}
}

// TestBasenameFromURLFull verifies basename extraction from full URL.
func TestBasenameFromURLFull(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{"simple file", "https://example.com/file.pdf", "file.pdf"},
		{"with query", "https://example.com/file.jpg?width=200", "file.jpg"},
		{"with fragment", "https://example.com/page#section", "page"},
		{"with both", "https://example.com/doc.pdf?a=1#x", "doc.pdf"},
		{"no path", "https://example.com", "example.com"},
		{"trailing slash", "https://example.com/docs/", "Untitled"},
		{"empty segment", "https://example.com///", "Untitled"},
		{"complex path", "https://example.com/docs/v1/file.pdf", "file.pdf"},
		{"just fragment", "https://example.com#only", "example.com"},
		{"query no file", "https://example.com?x=1", "example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BasenameFromURLFull(tt.url)
			if result != tt.expected {
				t.Errorf("BasenameFromURLFull(%q) = %q, want %q", tt.url, result, tt.expected)
			}
		})
	}
}

// TestCountWords_edgeCases tests edge cases for word counting.
func TestCountWords_edgeCases(t *testing.T) {
	t.Run("very long text", func(t *testing.T) {
		text := "word "
		for i := 0; i < 1000; i++ {
			text += "word "
		}
		expected := 1001
		result := CountWords(text)
		if result != expected {
			t.Errorf("CountWords(long text) = %d, want %d", result, expected)
		}
	})

	t.Run("single character", func(t *testing.T) {
		if CountWords("a") != 1 {
			t.Error("CountWords('a') should be 1")
		}
	})

	t.Run("only whitespace", func(t *testing.T) {
		if CountWords("   \t\n   ") != 0 {
			t.Error("CountWords('   \\t\\n   ') should be 0")
		}
	})
}

// TestTruncate_edgeCases tests edge cases for truncation.
func TestTruncate_edgeCases(t *testing.T) {
	t.Run("very long text", func(t *testing.T) {
		text := ""
		for i := 0; i < 1000; i++ {
			text += "word "
		}
		result := Truncate(text, 50)
		if len(result) > 53 { // 50 + "..."
			t.Errorf("Truncate result too long: %d", len(result))
		}
		if !strings.HasSuffix(result, "...") {
			t.Error("Truncate should end with '...'")
		}
	})

	t.Run("max len zero", func(t *testing.T) {
		result := Truncate("hello world", 0)
		if result != "..." {
			t.Errorf("Truncate with maxLen=0 should return '...', got %q", result)
		}
	})

	t.Run("chinese characters", func(t *testing.T) {
		// Each Chinese character is 3 bytes
		result := Truncate("你好世界", 5)
		// Should truncate at byte boundary, not character boundary
		if !strings.HasSuffix(result, "...") {
			t.Error("Truncate should end with '...'")
		}
	})
}

// TestDetectLanguage_edgeCases tests edge cases for language detection.
func TestDetectLanguage_edgeCases(t *testing.T) {
	t.Run("very long ascii", func(t *testing.T) {
		text := ""
		for i := 0; i < 10000; i++ {
			text += "a"
		}
		result := DetectLanguage(text)
		if result != "en" {
			t.Errorf("DetectLanguage(long ascii) = %q, want 'en'", result)
		}
	})

	t.Run("mixed script", func(t *testing.T) {
		result := DetectLanguage("Hello 世界 123")
		if result != "unknown" {
			t.Errorf("DetectLanguage(mixed) = %q, want 'unknown'", result)
		}
	})
}
