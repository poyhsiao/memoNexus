// Package textrank implements unit tests for TextRank keyword extraction.
// T128: Unit test for TextRank algorithm.
package textrank

import (
	"reflect"
	"testing"
)

// TestNewTextRankExtractor verifies default configuration.
func TestNewTextRankExtractor(t *testing.T) {
	extractor := NewTextRankExtractor()

	if extractor == nil {
		t.Fatal("NewTextRankExtractor() returned nil")
	}

	if extractor.windowSize != 5 {
		t.Errorf("expected windowSize 5, got %d", extractor.windowSize)
	}

	if extractor.dampingFactor != 0.85 {
		t.Errorf("expected dampingFactor 0.85, got %f", extractor.dampingFactor)
	}

	if extractor.convergenceThreshold != 0.0001 {
		t.Errorf("expected convergenceThreshold 0.0001, got %f", extractor.convergenceThreshold)
	}

	if extractor.maxIterations != 100 {
		t.Errorf("expected maxIterations 100, got %d", extractor.maxIterations)
	}

	if extractor.numKeywords != 10 {
		t.Errorf("expected numKeywords 10, got %d", extractor.numKeywords)
	}
}

// TestExtractEmptyText verifies empty text handling.
func TestExtractEmptyText(t *testing.T) {
	extractor := NewTextRankExtractor()

	keywords, err := extractor.Extract("")

	if err != nil {
		t.Fatalf("Extract() returned error: %v", err)
	}

	if keywords == nil {
		t.Fatal("Extract() returned nil keywords")
	}

	if len(keywords) != 0 {
		t.Errorf("expected 0 keywords, got %d", len(keywords))
	}
}

// TestExtractEnglishText verifies English keyword extraction.
func TestExtractEnglishText(t *testing.T) {
	extractor := NewTextRankExtractor()

	text := "Machine learning is transforming artificial intelligence. " +
		"Machine learning algorithms enable computers to learn from data. " +
		"Deep learning is a subset of machine learning."

	keywords, err := extractor.Extract(text)

	if err != nil {
		t.Fatalf("Extract() returned error: %v", err)
	}

	if keywords == nil {
		t.Fatal("Extract() returned nil keywords")
	}

	if len(keywords) == 0 {
		t.Fatal("expected keywords to be extracted, got empty list")
	}

	// Verify "learning" and "machine" appear in top keywords (they co-occur frequently)
	hasMachine := false
	hasLearning := false

	for _, kw := range keywords {
		switch kw {
		case "machine":
			hasMachine = true
		case "learning":
			hasLearning = true
		}
	}

	if !hasMachine {
		t.Error("expected 'machine' in keywords (high co-occurrence)")
	}
	if !hasLearning {
		t.Error("expected 'learning' in keywords (high co-occurrence)")
	}

	// Verify stop words are filtered
	stopWords := []string{"is", "a", "of", "to"}
	for _, kw := range keywords {
		for _, sw := range stopWords {
			if sw == kw {
				t.Errorf("stop word '%s' found in keywords", sw)
			}
		}
	}
}

// TestExtractChineseText verifies Chinese keyword extraction.
func TestExtractChineseText(t *testing.T) {
	extractor := NewTextRankExtractor()

	text := "机器学习正在改变人工智能。" +
		"机器学习算法使计算机能够从数据中学习。" +
		"深度学习是机器学习的一个子集。"

	keywords, err := extractor.Extract(text)

	if err != nil {
		t.Fatalf("Extract() returned error: %v", err)
	}

	if keywords == nil {
		t.Fatal("Extract() returned nil keywords")
	}

	if len(keywords) == 0 {
		t.Fatal("expected keywords to be extracted, got empty list")
	}

	// Verify some meaningful Chinese keywords appear
	hasMachine := false
	hasLearning := false
	hasAI := false

	for _, kw := range keywords {
		if contains(kw, "机器") || contains(kw, "机") {
			hasMachine = true
		}
		if contains(kw, "学习") || contains(kw, "习") {
			hasLearning = true
		}
		if contains(kw, "人工智能") || contains(kw, "智能") {
			hasAI = true
		}
	}

	if !hasMachine {
		t.Error("expected '机器' or '机器学习' in keywords")
	}
	if !hasLearning {
		t.Error("expected '学习' in keywords")
	}
	if !hasAI {
		t.Error("expected '人工智能' or '智能' in keywords")
	}
}

// TestSetWindowSize verifies window size configuration.
func TestSetWindowSize(t *testing.T) {
	extractor := NewTextRankExtractor()
	extractor.SetWindowSize(3)

	if extractor.windowSize != 3 {
		t.Errorf("expected windowSize 3, got %d", extractor.windowSize)
	}
}

// TestSetNumKeywords verifies keyword count configuration.
func TestSetNumKeywords(t *testing.T) {
	extractor := NewTextRankExtractor()
	extractor.SetNumKeywords(5)

	if extractor.numKeywords != 5 {
		t.Errorf("expected numKeywords 5, got %d", extractor.numKeywords)
	}

	text := "Machine learning algorithms artificial intelligence data science."

	keywords, err := extractor.Extract(text)
	if err != nil {
		t.Fatalf("Extract() returned error: %v", err)
	}

	if len(keywords) > 5 {
		t.Errorf("expected at most 5 keywords, got %d", len(keywords))
	}
}

// TestExtractCandidates verifies candidate extraction.
func TestExtractCandidates(t *testing.T) {
	extractor := NewTextRankExtractor()

	text := "Machine learning is artificial intelligence."

	candidates := extractor.extractCandidates(text)

	if len(candidates) == 0 {
		t.Fatal("expected candidates to be extracted, got empty list")
	}

	// Verify all candidates are at least 2 characters
	for _, c := range candidates {
		if len(c.text) < 2 {
			t.Errorf("candidate too short: '%s' (length %d)", c.text, len(c.text))
		}
	}

	// Verify positions are valid
	for i, c := range candidates {
		if c.start < 0 || c.start >= len(text) {
			t.Errorf("candidate %d has invalid start: %d", i, c.start)
		}
		if c.end <= c.start || c.end > len(text) {
			t.Errorf("candidate %d has invalid end: %d", i, c.end)
		}
	}
}

// TestIsStopWord verifies stop word detection.
func TestIsStopWord(t *testing.T) {
	stopWords := []string{"the", "a", "an", "and", "is", "in", "on", "can", "could"}

	for _, word := range stopWords {
		if !isStopWord(word) {
			t.Errorf("expected '%s' to be a stop word", word)
		}
	}

	// Test case insensitivity
	if !isStopWord("The") {
		t.Error("expected 'The' to be a stop word (case insensitive)")
	}

	// Non-stop words
	nonStopWords := []string{"machine", "learning", "algorithm"}
	for _, word := range nonStopWords {
		if isStopWord(word) {
			t.Errorf("did not expect '%s' to be a stop word", word)
		}
	}
}

// TestBuildGraph verifies co-occurrence graph building.
func TestBuildGraph(t *testing.T) {
	extractor := NewTextRankExtractor()
	extractor.SetWindowSize(2) // Small window for testing

	text := "machine learning algorithm"

	candidates := extractor.extractCandidates(text)
	graph := extractor.buildGraph(candidates, text)

	if len(graph) == 0 {
		t.Fatal("expected non-empty graph")
	}

	// With window size 2, "machine" should be connected to "learning"
	// and "learning" should be connected to both "machine" and "algorithm"
	if graph["machine"]["learning"] != 1.0 {
		t.Error("expected edge between 'machine' and 'learning'")
	}

	if graph["learning"]["algorithm"] != 1.0 {
		t.Error("expected edge between 'learning' and 'algorithm'")
	}

	// Verify symmetry (undirected graph)
	for node, neighbors := range graph {
		for neighbor := range neighbors {
			if graph[neighbor][node] != 1.0 {
				t.Errorf("graph not symmetric: %s->%s = 1.0 but %s->%s = %f",
					node, neighbor, neighbor, node, graph[neighbor][node])
			}
		}
	}
}

// TestPageRank verifies PageRank algorithm.
func TestPageRank(t *testing.T) {
	extractor := NewTextRankExtractor()

	// Create a simple graph: A <-> B <-> C (chain)
	graph := graph{
		"A": map[string]float64{"B": 1.0},
		"B": map[string]float64{"A": 1.0, "C": 1.0},
		"C": map[string]float64{"B": 1.0},
	}

	scores := extractor.pageRank(graph)

	if len(scores) != 3 {
		t.Fatalf("expected 3 scores, got %d", len(scores))
	}

	// Verify all scores are positive
	for node, score := range scores {
		if score <= 0 {
			t.Errorf("node %s has non-positive score: %f", node, score)
		}
	}

	// In a chain, B should have the highest score (most connections)
	hasA := false
	hasB := false
	hasC := false
	var scoreA, scoreB, scoreC float64

	for node, score := range scores {
		switch node {
		case "A":
			hasA = true
			scoreA = score
		case "B":
			hasB = true
			scoreB = score
		case "C":
			hasC = true
			scoreC = score
		}
	}

	if !hasA || !hasB || !hasC {
		t.Fatal("not all nodes scored")
	}

	// B should have higher score than A and C (more connections)
	if scoreB <= scoreA {
		t.Error("expected B to have higher score than A")
	}
	if scoreB <= scoreC {
		t.Error("expected B to have higher score than C")
	}
}

// TestPageRankConvergence verifies convergence detection.
func TestPageRankConvergence(t *testing.T) {
	extractor := NewTextRankExtractor()
	extractor.SetNumKeywords(10)

	// Create a star graph: A connected to all others
	nodes := []string{"A", "B", "C", "D", "E"}
	graph := make(graph)

	for _, node := range nodes {
		graph[node] = make(map[string]float64)
	}

	// A connects to everyone
	for _, node := range nodes[1:] {
		graph["A"][node] = 1.0
		graph[node]["A"] = 1.0
	}

	scores := extractor.pageRank(graph)

	// A should have the highest score (center of star)
	scoreA := scores["A"]
	for node, score := range scores {
		if node != "A" && score >= scoreA {
			t.Errorf("expected A to have highest score, but %s has %f >= %f",
				node, score, scoreA)
		}
	}
}

// TestGetTopKeywords verifies top N keyword selection.
func TestGetTopKeywords(t *testing.T) {
	extractor := NewTextRankExtractor()
	extractor.SetNumKeywords(3)

	// Create mock candidates and scores
	candidates := []candidate{
		{text: "zebra"},
		{text: "apple"},
		{text: "banana"},
		{text: "cherry"},
		{text: "date"},
	}

	scores := map[string]float64{
		"zebra":  0.05,
		"apple":  0.50,
		"banana": 0.30,
		"cherry": 0.20,
		"date":   0.10,
	}

	keywords := extractor.getTopKeywords(candidates, scores)

	if len(keywords) != 3 {
		t.Fatalf("expected 3 keywords, got %d", len(keywords))
	}

	// Verify order (highest score first)
	expected := []string{"apple", "banana", "cherry"}
	if !reflect.DeepEqual(keywords, expected) {
		t.Errorf("expected %v, got %v", expected, keywords)
	}
}

// TestExtractKeywordsConvenience verifies convenience function.
func TestExtractKeywordsConvenience(t *testing.T) {
	text := "Machine learning algorithms enable computers to learn."

	keywords, err := ExtractKeywords(text, 5)

	if err != nil {
		t.Fatalf("ExtractKeywords() returned error: %v", err)
	}

	if keywords == nil {
		t.Fatal("ExtractKeywords() returned nil keywords")
	}

	if len(keywords) > 5 {
		t.Errorf("expected at most 5 keywords, got %d", len(keywords))
	}

	if len(keywords) == 0 {
		t.Error("expected keywords to be extracted")
	}
}

// TestIsCJKCharacter verifies CJK character detection.
func TestIsCJKCharacter(t *testing.T) {
	tests := []struct {
		r     rune
		valid bool
	}{
		// Chinese (Han)
		{'中', true},
		{'文', true},
		{0x4E00, true}, // CJK Unified Ideographs-4E00
		{0x9FFF, true}, // CJK Unified Ideographs-9FFF

		// Japanese Hiragana
		{'あ', true},
		{'か', true},
		{0x3040, true},
		{0x309F, true},

		// Japanese Katakana
		{'ア', true},
		{'カ', true},
		{0x30A0, true},
		{0x30FF, true},

		// Korean Hangul
		{'가', true},
		{'다', true},
		{0xAC00, true},
		{0xD7AF, true},

		// ASCII (not CJK)
		{'A', false},
		{'z', false},
		{'0', false},
		{'9', false},

		// Other ranges (not CJK)
		{' ', false},
		{'.', false},
		// Note: '。' (U+3002) is in CJK Symbols & Punctuation range (0x3000-0x303F)
		// which is included in the 0x2E80-0x9FFF check, so it's treated as CJK
	}

	for _, tt := range tests {
		result := IsCJKCharacter(tt.r)
		if result != tt.valid {
			t.Errorf("IsCJKCharacter('%c'): expected %v, got %v", tt.r, tt.valid, result)
		}
	}
}

// TestHasCJKText verifies CJK text detection.
func TestHasCJKText(t *testing.T) {
	tests := []struct {
		name string
		text string
		has  bool
	}{
		{
			name: "Chinese characters",
			text: "这是中文",
			has:  true,
		},
		{
			name: "Japanese Hiragana",
			text: "これはひらがな",
			has:  true,
		},
		{
			name: "Japanese Katakana",
			text: "これはカタカナ",
			has:  true,
		},
		{
			name: "Korean Hangul",
			text: "이것은한글",
			has:  true,
		},
		{
			name: "English only",
			text: "This is English text",
			has:  false,
		},
		{
			name: "Mixed with CJK",
			text: "This has 中文 mixed in",
			has:  true,
		},
		{
			name: "Numbers and punctuation",
			text: "123.!?@#",
			has:  false,
		},
		{
			name: "Empty string",
			text: "",
			has:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HasCJKText(tt.text)
			if result != tt.has {
				t.Errorf("HasCJKText(): expected %v, got %v", tt.has, result)
			}
		})
	}
}

// TestTokenizeForTextRank verifies TextRank tokenization.
func TestTokenizeForTextRank(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		minCount int
	}{
		{
			name:     "English text",
			text:     "Hello world this is a test",
			minCount: 5,
		},
		{
			name:     "Chinese text",
			text:     "机器学习算法",
			minCount: 3,
		},
		{
			name:     "Mixed CJK and English",
			text:     "Machine 机器 learning 学习",
			minCount: 4,
		},
		{
			name:     "Text with punctuation",
			text:     "Hello, world! How are you?",
			minCount: 4,
		},
		{
			name:     "Empty string",
			text:     "",
			minCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := TokenizeForTextRank(tt.text)
			if len(tokens) < tt.minCount {
				t.Errorf("expected at least %d tokens, got %d", tt.minCount, len(tokens))
			}

			// Verify no empty tokens
			for i, token := range tokens {
				if token == "" {
					t.Errorf("token %d is empty", i)
				}
			}
		})
	}
}

// TestCoOccurrenceWindowing verifies co-occurrence detection.
func TestCoOccurrenceWindowing(t *testing.T) {
	extractor := NewTextRankExtractor()
	extractor.SetWindowSize(2) // Very small window

	text := "one two three four five"

	candidates := extractor.extractCandidates(text)
	graph := extractor.buildGraph(candidates, text)

	// With window size 2, only adjacent words should be connected
	// "one" should connect only to "two"
	neighborsOfOne := graph["one"]
	if len(neighborsOfOne) != 1 {
		t.Errorf("expected 1 neighbor for 'one', got %d", len(neighborsOfOne))
	}
	if neighborsOfOne["two"] != 1.0 {
		t.Error("expected 'one' to connect to 'two'")
	}

	// "three" should connect to both "two" and "four"
	neighborsOfThree := graph["three"]
	expectedNeighbors := 2
	if len(neighborsOfThree) != expectedNeighbors {
		t.Errorf("expected %d neighbors for 'three', got %d", expectedNeighbors, len(neighborsOfThree))
	}
}

// TestKeywordUniqueness verifies unique keyword extraction.
func TestKeywordUniqueness(t *testing.T) {
	extractor := NewTextRankExtractor()

	// Text with repeated words
	text := "machine learning machine learning machine learning"

	keywords, err := extractor.Extract(text)

	if err != nil {
		t.Fatalf("Extract() returned error: %v", err)
	}

	// Verify no duplicates
	seen := make(map[string]bool)
	for _, kw := range keywords {
		if seen[kw] {
			t.Errorf("duplicate keyword found: '%s'", kw)
		}
		seen[kw] = true
	}
}

// BenchmarkExtract benchmarks keyword extraction.
func BenchmarkExtract(b *testing.B) {
	extractor := NewTextRankExtractor()

	text := "Machine learning is a subfield of artificial intelligence. " +
		"Machine learning algorithms build models based on sample data. " +
		"Artificial intelligence and machine learning are transforming technology. " +
		"Deep learning neural networks have revolutionized computer vision. " +
		"Natural language processing enables machines to understand text."

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = extractor.Extract(text)
	}
}

// BenchmarkExtractChinese benchmarks Chinese keyword extraction.
func BenchmarkExtractChinese(b *testing.B) {
	extractor := NewTextRankExtractor()

	text := "机器学习是人工智能的一个子领域。" +
		"机器学习算法基于样本数据构建模型。" +
		"人工智能和机器学习正在改变技术。" +
		"深度学习神经网络已经改变了计算机视觉。" +
		"自然语言处理使机器能够理解文本。"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = extractor.Extract(text)
	}
}

// Helper function for substring check
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		len(s) > len(substr) && indexOf(s, substr) >= 0)
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
