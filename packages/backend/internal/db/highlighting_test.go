// Package db tests for search result highlighting functionality.
package db

import (
	"testing"
)

// TestIsCJKCharacter verifies CJK character detection.
func TestIsCJKCharacter(t *testing.T) {
	tests := []struct {
		name     string
		r        rune
		expected bool
	}{
		// Chinese (Hanzi)
		{"chinese common", '你', true},
		{"chinese range 1", 0x1100, true},
		{"chinese range 2", 0x2E80, true},
		{"chinese range 3", 0x9FFF, true},

		// Japanese Hiragana
		{"japanese hiragana start", 0x3040, true},
		{"japanese hiragana", 'あ', true},
		{"japanese hiragana end", 0x309F, true},

		// Japanese Katakana
		{"japanese katakana start", 0x30A0, true},
		{"japanese katakana", 'ア', true},
		{"japanese katakana end", 0x30FF, true},

		// Korean Hangul
		{"korean hangul start", 0xAC00, true},
		{"korean hangul", '한', true},
		{"korean hangul end", 0xD7AF, true},

		// CJK Extensions
		{"cjk extension start", 0xF900, true},
		{"cjk extension end", 0xFAFF, true},
		{"cjk compatibility start", 0xFF00, true},
		{"cjk compatibility end", 0xFFEF, true},

		// Non-CJK
		{"ascii lowercase", 'a', false},
		{"ascii uppercase", 'A', false},
		{"ascii digit", '0', false},
		{"space", ' ', false},
		{"punctuation", ',', false},
		{"latin accented", 'é', false},
		{"cyrillic", 'д', false},
		{"arabic", 'ا', false},
		{"emoji", '👋', false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsCJKCharacter(tt.r)
			if result != tt.expected {
				t.Errorf("IsCJKCharacter(%U) = %v, want %v", tt.r, result, tt.expected)
			}
		})
	}
}

// TestHasCJKText verifies CJK text detection.
func TestHasCJKText(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected bool
	}{
		{"empty", "", false},
		{"ascii only", "hello world", false},
		{"with numbers", "test 123", false},
		{"chinese single", "你", true},
		{"chinese multiple", "你好世界", true},
		{"japanese hiragana", "こんにちは", true},
		{"japanese katakana", "コンニチハ", true},
		{"korean hangul", "안녕하세요", true},
		{"mixed ascii chinese", "Hello 你好", true},
		{"mixed ascii japanese", "Test こんにちは", true},
		{"mixed ascii korean", "Start 안녕하세요 End", true},
		{"cjk extension",  string(rune(0xF900)), true},
		{"punctuation only", "!@#$%", false},
		{"whitespace only", "   \t\n", false},
		{"latin accented", "café résumé", false},
		{"emoji only", "👋🌍🎉", false},
		{"mixed all", "Hello 你好 123 🌍", true},
		{"mixed cjk types", "你好こんにちは안녕하세요", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HasCJKText(tt.text)
			if result != tt.expected {
				t.Errorf("HasCJKText(%q) = %v, want %v", tt.text, result, tt.expected)
			}
		})
	}
}

// TestWordCountApproximate verifies word counting for CJK and space-separated text.
func TestWordCountApproximate(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected int
	}{
		{"empty", "", 0},
		{"whitespace only", "   \t\n", 0},

		// Space-separated languages
		{"single word", "hello", 1},
		{"two words", "hello world", 2},
		{"multiple words", "the quick brown fox", 4},
		{"with extra spaces", "hello   world", 2},
		{"with tabs", "hello\tworld", 2},
		{"with newlines", "hello\nworld", 2},
		{"mixed whitespace", "hello \t\n world", 2},

		// CJK text (each character is a word)
		{"chinese single", "你", 1},
		{"chinese multiple", "你好世界", 4},
		{"japanese hiragana", "こんにちは", 5},
		{"japanese katakana", "コンニチハ", 5},
		{"korean hangul", "안녕하세요", 5},

		// Mixed ASCII and CJK
		{"mixed ascii chinese", "Hello 你好", 3},     // "Hello" (1) + "你" (1) + "好" (1) = 3
		{"mixed prefix", "Test 測試", 3},             // "Test" (1) + "測" (1) + "試" (1) = 3
		{"mixed suffix", "測試 Test", 3},             // "測" (1) + "試" (1) + "Test" (1) = 3
		{"mixed middle", "Start 測試 End", 4},        // "Start" (1) + "測" (1) + "試" (1) + "End" (1) = 4
		{"mixed no spaces", "Hello你好", 3},          // "Hello" (1) + "你" (1) + "好" (1) = 3

		// Complex cases
		{"numbers", "123 456", 2},
		{"punctuation", "hello, world!", 2},
		{"mixed cjk types", "你好こんにちは안녕하세요", 12},     // 2 + 5 + 5 = 12 CJK chars
		{"cjk with spaces", "你 好 世 界", 4},
		{"ascii with punctuation attached", "hello,world", 1}, // comma not treated as separator

		// Edge cases
		{"single char", "a", 1},
		{"single cjk", "你", 1},
		{"spaces around word", "  hello  ", 1},
		{"multiple spaces between", "word1    word2", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WordCountApproximate(tt.text)
			if result != tt.expected {
				t.Errorf("WordCountApproximate(%q) = %d, want %d", tt.text, result, tt.expected)
			}
		})
	}
}

// TestWordCountApproximate_edgeCases tests edge cases.
func TestWordCountApproximate_edgeCases(t *testing.T) {
	t.Run("long ascii text", func(t *testing.T) {
		text := ""
		for i := 0; i < 100; i++ {
			text += "word "
		}
		result := WordCountApproximate(text)
		if result != 100 {
			t.Errorf("WordCountApproximate(long text) = %d, want 100", result)
		}
	})

	t.Run("long cjk text", func(t *testing.T) {
		text := ""
		for i := 0; i < 100; i++ {
			text += "你"
		}
		result := WordCountApproximate(text)
		if result != 100 {
			t.Errorf("WordCountApproximate(long cjk) = %d, want 100", result)
		}
	})

	t.Run("mixed long text", func(t *testing.T) {
		text := "Hello 你好 World 測試"
		result := WordCountApproximate(text)
		// "Hello" (1) + "你" (1) + "好" (1) + "World" (1) + "測" (1) + "試" (1) = 6
		if result != 6 {
			t.Errorf("WordCountApproximate(mixed) = %d, want 6", result)
		}
	})
}
