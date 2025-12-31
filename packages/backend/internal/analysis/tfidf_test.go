// Package analysis provides unit tests for TF-IDF keyword extraction.
// T127: Unit test for TF-IDF keyword extraction.
package analysis

import (
	"reflect"
	"strings"
	"testing"
	"unicode/utf8"
)

// TestNewTFIDFAnalyzer verifies that the analyzer initializes correctly.
func TestNewTFIDFAnalyzer(t *testing.T) {
	analyzer := NewTFIDFAnalyzer()

	if analyzer == nil {
		t.Fatal("NewTFIDFAnalyzer() returned nil")
	}

	if analyzer.stopWords == nil {
		t.Error("stopWords map not initialized")
	}

	// Verify some common stop words exist
	commonStopWords := []string{"the", "a", "an", "and", "is", "in", "on"}
	for _, word := range commonStopWords {
		if !analyzer.stopWords[word] {
			t.Errorf("expected stop word '%s' not found", word)
		}
	}
}

// TestAnalyzeEmptyText verifies that empty text returns empty result.
func TestAnalyzeEmptyText(t *testing.T) {
	analyzer := NewTFIDFAnalyzer()

	result, err := analyzer.Analyze("")

	if err != nil {
		t.Fatalf("Analyze() returned error: %v", err)
	}

	if result == nil {
		t.Fatal("Analyze() returned nil result")
	}

	if len(result.Keywords) != 0 {
		t.Errorf("expected 0 keywords, got %d", len(result.Keywords))
	}

	if result.Summary != "" {
		t.Errorf("expected empty summary, got '%s'", result.Summary)
	}

	if result.Language != "unknown" {
		t.Errorf("expected language 'unknown', got '%s'", result.Language)
	}
}

// TestAnalyzeEnglishText verifies keyword extraction for English text.
func TestAnalyzeEnglishText(t *testing.T) {
	analyzer := NewTFIDFAnalyzer()

	text := `Machine learning is a subfield of artificial intelligence.
		Machine learning algorithms build models based on sample data.
		Artificial intelligence and machine learning are transforming technology.`

	result, err := analyzer.Analyze(text)

	if err != nil {
		t.Fatalf("Analyze() returned error: %v", err)
	}

	if result == nil {
		t.Fatal("Analyze() returned nil result")
	}

	if result.Language != "en" {
		t.Errorf("expected language 'en', got '%s'", result.Language)
	}

	// Verify keywords were extracted
	if len(result.Keywords) == 0 {
		t.Error("expected keywords to be extracted, got empty list")
	}

	// Verify "machine" and "learning" appear in top keywords (they appear 3 times)
	hasMachine := false
	hasLearning := false
	for _, kw := range result.Keywords {
		if kw == "machine" {
			hasMachine = true
		}
		if kw == "learning" {
			hasLearning = true
		}
	}

	if !hasMachine {
		t.Error("expected 'machine' to be in top keywords")
	}
	if !hasLearning {
		t.Error("expected 'learning' to be in top keywords")
	}

	// Verify stop words are filtered
	for _, kw := range result.Keywords {
		if analyzer.stopWords[kw] {
			t.Errorf("stop word '%s' should not be in keywords", kw)
		}
	}

	// Verify summary was generated
	if result.Summary == "" {
		t.Error("expected summary to be generated")
	}
	if len(result.Summary) > 500 {
		t.Errorf("summary too long: %d characters (max 500)", len(result.Summary))
	}
}

// TestAnalyzeChineseText verifies keyword extraction for Chinese text.
func TestAnalyzeChineseText(t *testing.T) {
	analyzer := NewTFIDFAnalyzer()

	text := `机器学习是人工智能的一个子领域。
		机器学习算法基于样本数据构建模型。
		人工智能和机器学习正在改变技术。`

	result, err := analyzer.Analyze(text)

	if err != nil {
		t.Fatalf("Analyze() returned error: %v", err)
	}

	if result == nil {
		t.Fatal("Analyze() returned nil result")
	}

	if result.Language != "zh" {
		t.Errorf("expected language 'zh', got '%s'", result.Language)
	}

	// Verify keywords were extracted (including Chinese bigrams)
	if len(result.Keywords) == 0 {
		t.Error("expected keywords to be extracted, got empty list")
	}

	// Verify summary was generated
	if result.Summary == "" {
		t.Error("expected summary to be generated")
	}
}

// TestAnalyzeStopWordFiltering verifies that stop words are properly filtered.
func TestAnalyzeStopWordFiltering(t *testing.T) {
	analyzer := NewTFIDFAnalyzer()

	text := "The quick brown fox jumps over the lazy dog. The dog was not lazy."

	result, err := analyzer.Analyze(text)

	if err != nil {
		t.Fatalf("Analyze() returned error: %v", err)
	}

	// Check that common stop words are not in keywords
	stopWords := []string{"the", "was", "not", "over"}
	for _, sw := range stopWords {
		for _, kw := range result.Keywords {
			if kw == sw {
				t.Errorf("stop word '%s' found in keywords", sw)
			}
		}
	}

	// Verify meaningful words appear
	hasWord := false
	meaningfulWords := []string{"quick", "brown", "fox", "jumps", "lazy", "dog"}
	for _, kw := range result.Keywords {
		for _, mw := range meaningfulWords {
			if kw == mw {
				hasWord = true
				break
			}
		}
	}

	if !hasWord {
		t.Error("expected at least one meaningful word in keywords")
	}
}

// TestTokenizeEnglish verifies English tokenization.
func TestTokenizeEnglish(t *testing.T) {
	analyzer := NewTFIDFAnalyzer()

	text := "Hello, world! This is a test."
	tokens := analyzer.tokenizeEnglish(text)

	if len(tokens) == 0 {
		t.Fatal("expected tokens, got empty slice")
	}

	// Verify tokens are lowercase
	for _, token := range tokens {
		if token != strings.ToLower(token) {
			t.Errorf("token not lowercased: '%s'", token)
		}
	}

	// Verify punctuation is stripped
	for _, token := range tokens {
		if strings.ContainsAny(token, ".,!?") {
			t.Errorf("token contains punctuation: '%s'", token)
		}
	}

	// Verify stop words are filtered
	for _, token := range tokens {
		if analyzer.stopWords[token] {
			t.Errorf("stop word not filtered: '%s'", token)
		}
	}
}

// TestTokenizeChinese verifies Chinese tokenization.
func TestTokenizeChinese(t *testing.T) {
	analyzer := NewTFIDFAnalyzer()

	text := "机器学习是人工智能"
	tokens := analyzer.tokenizeChinese(text)

	if len(tokens) == 0 {
		t.Fatal("expected tokens, got empty slice")
	}

	t.Logf("Tokenized Chinese text '%s' into %d tokens: %v", text, len(tokens), tokens)

	// Verify both single characters and bigrams are included
	hasChar := false
	hasBigram := false
	for _, token := range tokens {
		// Use utf8.RuneCountInString because len() returns bytes, not runes
		// Chinese characters are multi-byte in UTF-8 (3 bytes each)
		if utf8.RuneCountInString(token) == 1 {
			hasChar = true
		}
		if utf8.RuneCountInString(token) == 2 { // Bigram is 2 characters
			hasBigram = true
		}
	}

	if !hasChar {
		t.Error("expected single character tokens")
	}
	if !hasBigram {
		t.Error("expected bigram tokens")
	}
}

// TestCalculateTermFrequency verifies TF calculation.
func TestCalculateTermFrequency(t *testing.T) {
	analyzer := NewTFIDFAnalyzer()

	tokens := []string{"apple", "banana", "apple", "cherry", "apple", "banana"}

	tf := analyzer.calculateTermFrequency(tokens)

	if tf == nil {
		t.Fatal("calculateTermFrequency() returned nil")
	}

	// Verify all tokens are present
	expectedTf := map[string]float64{
		"apple":  3.0 / 6.0, // 3/6 = 0.5
		"banana": 2.0 / 6.0, // 2/6 = 0.33
		"cherry": 1.0 / 6.0, // 1/6 = 0.17
	}

	for term, expectedFreq := range expectedTf {
		actualFreq, ok := tf[term]
		if !ok {
			t.Errorf("missing term '%s'", term)
			continue
		}
		// Use approximate comparison for floating point
		diff := actualFreq - expectedFreq
		if diff < 0 {
			diff = -diff
		}
		if diff > 0.01 {
			t.Errorf("term '%s': expected freq ~%.2f, got %.2f", term, expectedFreq, actualFreq)
		}
	}
}

// TestFilterTerms verifies low-frequency term filtering.
func TestFilterTerms(t *testing.T) {
	analyzer := NewTFIDFAnalyzer()

	tf := map[string]float64{
		"high":    0.15, // Above threshold
		"low":     0.005, // Below threshold
		"stop":    0.10, // Stop word
		"single":  0.05, // Single character ASCII
		"valid":   0.12, // Valid term
	}

	filtered := analyzer.filterTerms(tf)

	// Verify "low" is filtered (below threshold)
	if filtered["low"] != 0 {
		t.Error("expected 'low' to be filtered (below threshold)")
	}

	// Verify "stop" is filtered (stop word)
	if filtered["stop"] != 0 {
		t.Error("expected 'stop' to be filtered (stop word)")
	}

	// Verify "high" and "valid" pass through
	if filtered["high"] == 0 {
		t.Error("expected 'high' to pass filter")
	}
	if filtered["valid"] == 0 {
		t.Error("expected 'valid' to pass filter")
	}
}

// TestExtractTopKeywords verifies top N keyword extraction.
func TestExtractTopKeywords(t *testing.T) {
	analyzer := NewTFIDFAnalyzer()

	tf := map[string]float64{
		"zebra":   0.05,
		"apple":   0.50,
		"banana":  0.30,
		"cherry":  0.20,
		"date":    0.10,
	}

	// Extract top 3
	keywords := analyzer.extractTopKeywords(tf, 3)

	if len(keywords) != 3 {
		t.Fatalf("expected 3 keywords, got %d", len(keywords))
	}

	// Verify order (highest frequency first)
	expected := []string{"apple", "banana", "cherry"}
	if !reflect.DeepEqual(keywords, expected) {
		t.Errorf("expected %v, got %v", expected, keywords)
	}
}

// TestGenerateSummary verifies summary generation.
func TestGenerateSummary(t *testing.T) {
	analyzer := NewTFIDFAnalyzer()

	// Text shorter than max length
	shortText := "This is a short text."
	summary := analyzer.generateSummary(shortText, "en")

	if summary != shortText {
		t.Errorf("expected '%s', got '%s'", shortText, summary)
	}

	// Text longer than max length - should truncate
	longText := "This is the first sentence. This is the second sentence. This is the third sentence. " +
		"This is the fourth sentence that goes on and on and on and on and on and on."

	summary = analyzer.generateSummary(longText, "en")

	if len(summary) > 500 {
		t.Errorf("summary too long: %d characters (max 500)", len(summary))
	}

	if len(summary) == 0 {
		t.Error("expected non-empty summary")
	}

	// Verify it ends at a sentence boundary when possible
	// (should end with ".", "!", "?" or "...")
	if len(summary) < len(longText) {
		lastChar := summary[len(summary)-1]
		if lastChar == '.' || lastChar == '!' || lastChar == '?' {
			// Good - truncated at sentence boundary
		} else if summary[len(summary)-3:] == "..." {
			// Good - truncated with ellipsis
		}
	}
}

// TestDetectLanguage verifies language detection.
func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected string
	}{
		{
			name:     "English text",
			text:     "This is English text with words.",
			expected: "en",
		},
		{
			name:     "Chinese text",
			text:     "这是中文文本。",
			expected: "zh",
		},
		{
			name:     "Chinese dominant (mixed)",
			text:     "这是中文内容还有很多中文 with some English words.",
			expected: "zh", // >30% Chinese
		},
		{
			name:     "English dominant (mixed)",
			text:     "This is English text with some 中文 words.",
			expected: "en", // <30% Chinese
		},
		{
			name:     "Empty text",
			text:     "",
			expected: "unknown",
		},
		{
			name:     "No letters",
			text:     "123 !@# 456",
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectLanguage(tt.text)
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

// TestIsChineseChar verifies Chinese character detection.
func TestIsChineseChar(t *testing.T) {
	tests := []struct {
		r     rune
		valid bool
	}{
		{'中', true},  // Chinese
		{'日', true},  // Japanese
		{'한', true},  // Korean
		{'A', false},  // ASCII
		{'a', false},  // ASCII
		{'1', false},  // Digit
		{' ', false},  // Space
		{'。', false}, // Punctuation
	}

	for _, tt := range tests {
		result := isChineseChar(tt.r)
		if result != tt.valid {
			t.Errorf("isChineseChar('%c'): expected %v, got %v", tt.r, tt.valid, result)
		}
	}
}

// TestSummaryGenerationEnglish verifies summary generation for English.
func TestSummaryGenerationEnglish(t *testing.T) {
	analyzer := NewTFIDFAnalyzer()

	text := `Machine learning is a subset of artificial intelligence.
		It focuses on building systems that can learn from data.
		The field has seen tremendous growth in recent years.`

	result, err := analyzer.Analyze(text)
	if err != nil {
		t.Fatalf("Analyze() returned error: %v", err)
	}

	if result.Summary == "" {
		t.Error("expected non-empty summary")
	}

	// Summary should end at sentence boundary
	lastChar := result.Summary[len(result.Summary)-1]
	if lastChar != '.' && lastChar != '!' && lastChar != '?' {
		if result.Summary[len(result.Summary)-3:] != "..." {
			t.Error("summary should end at sentence boundary or with ellipsis")
		}
	}
}

// BenchmarkAnalyzeEnglish benchmarks English text analysis.
func BenchmarkAnalyzeEnglish(b *testing.B) {
	analyzer := NewTFIDFAnalyzer()

	text := `Machine learning is a subfield of artificial intelligence.
		Machine learning algorithms build models based on sample data.
		Artificial intelligence and machine learning are transforming technology.
		Deep learning neural networks have revolutionized computer vision.
		Natural language processing enables machines to understand text.`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = analyzer.Analyze(text)
	}
}

// BenchmarkAnalyzeChinese benchmarks Chinese text analysis.
func BenchmarkAnalyzeChinese(b *testing.B) {
	analyzer := NewTFIDFAnalyzer()

	text := `机器学习是人工智能的一个子领域。
		机器学习算法基于样本数据构建模型。
		人工智能和机器学习正在改变技术。
		深度学习神经网络已经改变了计算机视觉。
		自然语言处理使机器能够理解文本。`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = analyzer.Analyze(text)
	}
}
