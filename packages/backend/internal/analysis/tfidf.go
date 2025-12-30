// Package analysis provides content analysis capabilities.
package analysis

import (
	"regexp"
	"sort"
	"strings"
	"unicode"
)

// TFIDFAnalyzer provides TF-IDF based keyword extraction.
type TFIDFAnalyzer struct {
	// Stop words for filtering
	stopWords map[string]bool
}

// NewTFIDFAnalyzer creates a new TFIDFAnalyzer.
func NewTFIDFAnalyzer() *TFIDFAnalyzer {
	return &TFIDFAnalyzer{
		stopWords: buildStopWords(),
	}
}

// AnalysisResult represents the result of content analysis.
type AnalysisResult struct {
	Keywords []string // Top keywords
	Summary  string   // Generated summary
	Language string   // Detected language
}

// Analyze performs TF-IDF analysis on text and returns keywords.
func (a *TFIDFAnalyzer) Analyze(text string) (*AnalysisResult, error) {
	if text == "" {
		return &AnalysisResult{
			Keywords: []string{},
			Summary:  "",
			Language: "unknown",
		}, nil
	}

	// Detect language
	language := detectLanguage(text)

	// Tokenize
	tokens := a.tokenize(text, language)

	// Calculate term frequencies
	tf := a.calculateTermFrequency(tokens)

	// Filter stop words and low-frequency terms
	filtered := a.filterTerms(tf)

	// Extract top keywords
	keywords := a.extractTopKeywords(filtered, 10)

	// Generate summary (first N sentences)
	summary := a.generateSummary(text, language)

	return &AnalysisResult{
		Keywords: keywords,
		Summary:  summary,
		Language: language,
	}, nil
}

// tokenize splits text into terms based on language.
func (a *TFIDFAnalyzer) tokenize(text string, language string) []string {
	// Normalize whitespace
	text = strings.TrimSpace(text)
	re := regexp.MustCompile(`\s+`)
	text = re.ReplaceAllString(text, " ")

	if language == "zh" {
		return a.tokenizeChinese(text)
	}
	return a.tokenizeEnglish(text)
}

// tokenizeChinese tokenizes Chinese text (character-based n-grams).
func (a *TFIDFAnalyzer) tokenizeChinese(text string) []string {
	var tokens []string

	// Extract Chinese characters and bigrams
	chars := []rune(text)
	for i, r := range chars {
		// Skip punctuation and non-words
		if !isChineseChar(r) {
			continue
		}

		// Single character
		tokens = append(tokens, string(r))

		// Bigram (if next char is also Chinese)
		if i < len(chars)-1 && isChineseChar(chars[i+1]) {
			bigram := string(r) + string(chars[i+1])
			tokens = append(tokens, bigram)
		}
	}

	return tokens
}

// tokenizeEnglish tokenizes English text.
func (a *TFIDFAnalyzer) tokenizeEnglish(text string) []string {
	// Split on whitespace and punctuation
	words := strings.Fields(text)

	var tokens []string
	for _, word := range words {
		// Lowercase
		word = strings.ToLower(word)

		// Strip punctuation from start/end
		word = strings.TrimFunc(word, func(r rune) bool {
			return unicode.IsPunct(r) || unicode.IsSymbol(r)
		})

		// Filter empty and stop words
		if word != "" && !a.stopWords[word] && len(word) > 2 {
			tokens = append(tokens, word)
		}
	}

	return tokens
}

// calculateTermFrequency calculates term frequencies.
func (a *TFIDFAnalyzer) calculateTermFrequency(tokens []string) map[string]float64 {
	tf := make(map[string]float64)
	total := float64(len(tokens))

	for _, token := range tokens {
		tf[token]++
	}

	// Normalize
	for term := range tf {
		tf[term] = tf[term] / total
	}

	return tf
}

// filterTerms removes stop words and low-frequency terms.
func (a *TFIDFAnalyzer) filterTerms(tf map[string]float64) map[string]float64 {
	filtered := make(map[string]float64)

	minFreq := 0.01 // Minimum 1% frequency

	for term, freq := range tf {
		// Skip stop words
		if a.stopWords[term] {
			continue
		}

		// Skip very low frequency
		if freq < minFreq {
			continue
		}

		// Skip single characters for English
		if len(term) == 1 && isASCII(term[0]) {
			continue
		}

		filtered[term] = freq
	}

	return filtered
}

// extractTopKeywords extracts top N keywords by TF score.
func (a *TFIDFAnalyzer) extractTopKeywords(tf map[string]float64, n int) []string {
	// Sort by frequency
	type kv struct {
		Key   string
		Value float64
	}

	var pairs []kv
	for k, v := range tf {
		pairs = append(pairs, kv{k, v})
	}

	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].Value > pairs[j].Value
	})

	// Extract top N
	var keywords []string
	for i := 0; i < n && i < len(pairs); i++ {
		keywords = append(keywords, pairs[i].Key)
	}

	return keywords
}

// generateSummary generates a summary by extracting first sentences.
func (a *TFIDFAnalyzer) generateSummary(text string, language string) string {
	// Truncate to ~500 characters
	maxLen := 500

	if len(text) <= maxLen {
		return text
	}

	// Try to truncate at sentence boundary
	var sentenceEndRE *regexp.Regexp
	if language == "zh" {
		sentenceEndRE = regexp.MustCompile(`。[！？]`)
	} else {
		sentenceEndRE = regexp.MustCompile(`[.!?]\s`)
	}

	// Find last sentence end within maxLen
	matches := sentenceEndRE.FindAllStringIndex(text, -1)
	for i := len(matches) - 1; i >= 0; i-- {
		if matches[i][1] <= maxLen {
			return text[:matches[i][1]]
		}
	}

	// Fallback: truncate at word boundary
	summary := text[:maxLen]
	if i := strings.LastIndex(summary, " "); i > 0 {
		return summary[:i] + "..."
	}

	return summary + "..."
}

// detectLanguage detects if text is primarily Chinese or English.
func detectLanguage(text string) string {
	chineseChars := 0
	totalChars := 0

	for _, r := range text {
		if unicode.IsLetter(r) || unicode.Is(unicode.Han, r) {
			totalChars++
			if isChineseChar(r) {
				chineseChars++
			}
		}
	}

	if totalChars == 0 {
		return "unknown"
	}

	// If >30% Chinese characters, treat as Chinese
	if float64(chineseChars)/float64(totalChars) > 0.3 {
		return "zh"
	}

	return "en"
}

// isChineseChar checks if rune is a Chinese character.
func isChineseChar(r rune) bool {
	return unicode.Is(unicode.Han, r) ||
		unicode.Is(unicode.Hiragana, r) ||
		unicode.Is(unicode.Katakana, r) ||
		unicode.Is(unicode.Hangul, r)
}

// isASCII checks if byte is ASCII.
func isASCII(b byte) bool {
	return b < 128
}

// buildStopWords builds a map of stop words.
func buildStopWords() map[string]bool {
	// Common English stop words
	english := []string{
		"a", "an", "and", "are", "as", "at", "be", "by", "for",
		"from", "has", "he", "in", "is", "it", "its", "of", "on",
		"that", "the", "to", "was", "will", "with", "the", "this",
		"but", "they", "have", "had", "what", "when", "where", "who",
		"which", "why", "how", "all", "each", "every", "both", "few",
		"more", "most", "other", "some", "such", "no", "nor", "not",
		"only", "own", "same", "so", "than", "too", "very", "just",
		"can", "about", "into", "through", "during", "before", "after",
		"above", "below", "between", "under", "again", "further", "then",
		"once", "here", "there", "when", "where", "why", "how", "all",
		"any", "both", "each", "few", "more", "most", "other", "some",
		"such", "no", "nor", "not", "only", "own", "same", "so", "than",
		"too", "very", "get", "got", "getting", "got", "gotten",
	}

	// Common Chinese stop words (single characters and common particles)
	chinese := []string{
		"的", "了", "在", "是", "我", "有", "和", "就", "不", "人",
		"都", "一", "一個", "上", "也", "很", "到", "說", "要", "去",
		"你", "會", "著", "沒有", "看", "好", "自己", "這", "那",
		"裡", "就是", "嗎", "啊", "吧", "呢", "嘛", "哦", "呀",
	}

	stopWords := make(map[string]bool)
	for _, w := range english {
		stopWords[w] = true
	}
	for _, w := range chinese {
		stopWords[w] = true
	}

	return stopWords
}
