// Package textrank implements TextRank algorithm for keyword extraction.
// T131: TextRank keyword extraction using graph-based ranking.
package textrank

import (
	"math"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

// TextRankExtractor extracts keywords using the TextRank algorithm.
// Based on Mihalcea & Tarau (2004) - TextRank: Bringing Order into Texts.
type TextRankExtractor struct {
	// Window size for co-occurrence
	windowSize int

	// Damping factor for PageRank (default 0.85)
	dampingFactor float64

	// Convergence threshold
	convergenceThreshold float64

	// Maximum iterations
	maxIterations int

	// Number of keywords to extract
	numKeywords int
}

// NewTextRankExtractor creates a new TextRankExtractor with sensible defaults.
func NewTextRankExtractor() *TextRankExtractor {
	return &TextRankExtractor{
		windowSize:           5,
		dampingFactor:        0.85,
		convergenceThreshold: 0.0001,
		maxIterations:        100,
		numKeywords:          10,
	}
}

// SetWindowSize sets the co-occurrence window size.
func (e *TextRankExtractor) SetWindowSize(size int) *TextRankExtractor {
	e.windowSize = size
	return e
}

// SetNumKeywords sets the number of keywords to extract.
func (e *TextRankExtractor) SetNumKeywords(num int) *TextRankExtractor {
	e.numKeywords = num
	return e
}

// Extract extracts keywords from text using TextRank algorithm.
func (e *TextRankExtractor) Extract(text string) ([]string, error) {
	if text == "" {
		return []string{}, nil
	}

	// 1. Tokenize and filter candidates
	candidates := e.extractCandidates(text)
	if len(candidates) == 0 {
		return []string{}, nil
	}

	// 2. Build co-occurrence graph
	graph := e.buildGraph(candidates, text)

	// 3. Run PageRank to get node scores
	scores := e.pageRank(graph)

	// 4. Sort by score and return top N keywords
	keywords := e.getTopKeywords(candidates, scores)

	return keywords, nil
}

// candidate represents a potential keyword.
type candidate struct {
	text  string
	start int
	end   int
}

// extractCandidates extracts candidate keywords from text.
// Candidates are sequences of alphanumeric characters (including CJK).
func (e *TextRankExtractor) extractCandidates(text string) []candidate {
	// Find all words (including CJK characters)
	re := regexp.MustCompile(`[\p{L}\p{N}]+`)
	matches := re.FindAllStringIndex(text, -1)

	candidates := make([]candidate, 0, len(matches))
	for _, m := range matches {
		word := text[m[0]:m[1]]
		// Filter out very short words and common stop words
		if len(word) >= 2 && !isStopWord(word) {
			candidates = append(candidates, candidate{
				text:  word,
				start: m[0],
				end:   m[1],
			})
		}
	}

	return candidates
}

// isStopWord checks if a word is a common stop word.
func isStopWord(word string) bool {
	lowerWord := strings.ToLower(word)
	stopWords := map[string]bool{
		// English stop words
		"a": true, "an": true, "and": true, "are": true, "as": true,
		"at": true, "be": true, "by": true, "for": true, "from": true,
		"has": true, "he": true, "in": true, "is": true, "it": true,
		"its": true, "of": true, "on": true, "that": true, "the": true,
		"to": true, "was": true, "will": true, "with": true,

		// Common function words
		"can": true, "could": true, "should": true, "would": true,
		"this": true, "these": true, "those": true, "they": true,

		// Numbers (single digits)
		"0": true, "1": true, "2": true, "3": true, "4": true,
		"5": true, "6": true, "7": true, "8": true, "9": true,
	}
	return stopWords[lowerWord]
}

// graph represents a co-occurrence graph as an adjacency matrix.
type graph map[string]map[string]float64

// buildGraph builds a co-occurrence graph from candidates.
// Two candidates are connected if they appear within the window size.
func (e *TextRankExtractor) buildGraph(candidates []candidate, text string) graph {
	g := make(graph)

	// Initialize graph
	for _, c := range candidates {
		g[c.text] = make(map[string]float64)
	}

	// Find co-occurrences within window
	for i, c1 := range candidates {
		for j, c2 := range candidates {
			if i >= j {
				continue // Avoid duplicates and self-loops
			}

			// Check if candidates are within window distance
			distance := c2.start - c1.end
			if distance >= 0 && distance <= e.windowSize {
				// Add edge (undirected)
				g[c1.text][c2.text] = 1.0
				g[c2.text][c1.text] = 1.0
			}
		}
	}

	return g
}

// pageRank implements the PageRank algorithm on the co-occurrence graph.
func (e *TextRankExtractor) pageRank(g graph) map[string]float64 {
	if len(g) == 0 {
		return nil
	}

	// Initialize scores (equal for all nodes)
	nodes := make([]string, 0, len(g))
	for node := range g {
		nodes = append(nodes, node)
	}

	scores := make(map[string]float64)
	initialScore := 1.0 / float64(len(nodes))
	for _, node := range nodes {
		scores[node] = initialScore
	}

	// Iteratively update scores until convergence
	for iter := 0; iter < e.maxIterations; iter++ {
		newScores := make(map[string]float64)
		maxChange := 0.0

		for _, node := range nodes {
			// Calculate score from incoming edges
			sum := 0.0
			neighbors := g[node]
			degree := float64(len(neighbors))

			if degree > 0 {
				for neighbor := range neighbors {
					neighborDegree := float64(len(g[neighbor]))
					if neighborDegree > 0 {
						sum += scores[neighbor] / neighborDegree
					}
				}
			}

			// PageRank formula
			newScore := (1-e.dampingFactor)/float64(len(nodes)) + e.dampingFactor*sum
			newScores[node] = newScore

			// Track maximum change for convergence
			change := math.Abs(newScore - scores[node])
			if change > maxChange {
				maxChange = change
			}
		}

		scores = newScores

		// Check convergence
		if maxChange < e.convergenceThreshold {
			break
		}
	}

	return scores
}

// getTopKeywords returns the top N keywords by score.
func (e *TextRankExtractor) getTopKeywords(candidates []candidate, scores map[string]float64) []string {
	// Create list of (word, score) pairs
	type wordScore struct {
		word  string
		score float64
	}

	pairs := make([]wordScore, 0, len(candidates))
	for _, c := range candidates {
		if score, ok := scores[c.text]; ok {
			pairs = append(pairs, wordScore{
				word:  c.text,
				score: score,
			})
		}
	}

	// Sort by score (descending)
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].score > pairs[j].score
	})

	// Return top N unique keywords
	seen := make(map[string]bool)
	keywords := make([]string, 0, e.numKeywords)

	for _, pair := range pairs {
		if !seen[pair.word] {
			keywords = append(keywords, pair.word)
			seen[pair.word] = true
			if len(keywords) >= e.numKeywords {
				break
			}
		}
	}

	return keywords
}

// ExtractKeywords is a convenience method that creates an extractor
// and extracts keywords in one call.
func ExtractKeywords(text string, numKeywords int) ([]string, error) {
	extractor := NewTextRankExtractor()
	extractor.SetNumKeywords(numKeywords)
	return extractor.Extract(text)
}

// IsCJKCharacter checks if a rune is a CJK character.
func IsCJKCharacter(r rune) bool {
	// Chinese
	if (r >= 0x1100 && r <= 0x11FF) || (r >= 0x2E80 && r <= 0x9FFF) {
		return true
	}
	// Japanese Hiragana/Katakana
	if (r >= 0x3040 && r <= 0x30FF) {
		return true
	}
	// Korean Hangul
	if (r >= 0xAC00 && r <= 0xD7AF) {
		return true
	}
	return false
}

// HasCJKText checks if text contains CJK characters.
func HasCJKText(text string) bool {
	for _, r := range text {
		if IsCJKCharacter(r) {
			return true
		}
	}
	return false
}

// TokenizeForTextRank tokenizes text for TextRank processing.
// Handles both space-separated languages and CJK text.
func TokenizeForTextRank(text string) []string {
	var tokens []string
	var currentToken strings.Builder

	for _, r := range text {
		if IsCJKCharacter(r) {
			// CJK character - emit as individual token
			if currentToken.Len() > 0 {
				tokens = append(tokens, currentToken.String())
				currentToken.Reset()
			}
			tokens = append(tokens, string(r))
		} else if unicode.IsSpace(r) {
			// Word boundary
			if currentToken.Len() > 0 {
				tokens = append(tokens, currentToken.String())
				currentToken.Reset()
			}
		} else if unicode.IsLetter(r) || unicode.IsDigit(r) {
			// Part of a word
			currentToken.WriteRune(r)
		}
		// Ignore punctuation
	}

	// Don't forget the last token
	if currentToken.Len() > 0 {
		tokens = append(tokens, currentToken.String())
	}

	return tokens
}
