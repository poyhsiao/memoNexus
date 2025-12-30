// Package db provides search result highlighting functionality.
package db

import (
	"database/sql"
	"fmt"
	"html"
	"regexp"
	"strings"
	"unicode"
)

// HighlightOptions controls how search terms are highlighted.
type HighlightOptions struct {
	// MaxResults is the maximum number of snippets to extract per field
	MaxResults int

	// MaxChars is the maximum length of each snippet in characters
	MaxChars int

	// TagOpen is the HTML tag to use for opening highlight (default: <mark>)
	TagOpen string

	// TagClose is the HTML tag to use for closing highlight (default: </mark>)
	TagClose string
}

// DefaultHighlightOptions returns sensible defaults for highlighting.
func DefaultHighlightOptions() *HighlightOptions {
	return &HighlightOptions{
		MaxResults: 3,
		MaxChars:   150,
		TagOpen:    "<mark>",
		TagClose:   "</mark>",
	}
}

// HighlightedText contains a text snippet with highlighted search terms.
type HighlightedText struct {
	Text    string
	Snippet string
}

// HighlightResult contains highlighted versions of content fields.
type HighlightResult struct {
	Title       *HighlightedText
	ContentText *HighlightedText
	Tags        *HighlightedText
}

// Highlight extracts matched terms from FTS5 results and wraps them in highlight tags.
// Implements T117: Search term highlighting using FTS5 snippet function.
func (r *Repository) Highlight(db *sql.DB, itemID string, query string, opts *HighlightOptions) (*HighlightResult, error) {
	if opts == nil {
		opts = DefaultHighlightOptions()
	}

	// Sanitize the query for FTS5 snippet syntax
	// Remove special FTS5 operators but keep the search terms
	cleanQuery := sanitizeQueryForHighlight(query)

	// Build snippet queries using FTS5 highlight function
	// The snippet() function extracts context around matches and marks them with <b> tags
	snippetQuery := `
		SELECT
			snippet(content_items_fts, 1, '<b>', '</b>', '...', ?) as title_snippet,
			snippet(content_items_fts, 2, '<b>', '</b>', '...', ?) as content_snippet,
			snippet(content_items_fts, 3, '<b>', '</b>', '...', ?) as tags_snippet
		FROM content_items_fts
		WHERE rowid = (SELECT rowid FROM content_items WHERE id = ?)
			AND content_items_fts MATCH ?
		LIMIT 1
	`

	var titleSnippet, contentSnippet, tagsSnippet sql.NullString
	err := db.QueryRow(snippetQuery,
		opts.MaxChars, opts.MaxChars, opts.MaxChars,
		itemID, cleanQuery,
	).Scan(&titleSnippet, &contentSnippet, &tagsSnippet)

	if err != nil {
		return nil, fmt.Errorf("failed to extract highlights: %w", err)
	}

	result := &HighlightResult{}

	if titleSnippet.Valid && titleSnippet.String != "" {
		snippet := strings.ReplaceAll(titleSnippet.String, "<b>", opts.TagOpen)
		snippet = strings.ReplaceAll(snippet, "</b>", opts.TagClose)
		result.Title = &HighlightedText{
			Text:    snippet,
			Snippet: snippet,
		}
	}

	if contentSnippet.Valid && contentSnippet.String != "" {
		snippet := strings.ReplaceAll(contentSnippet.String, "<b>", opts.TagOpen)
		snippet = strings.ReplaceAll(snippet, "</b>", opts.TagClose)
		result.ContentText = &HighlightedText{
			Text:    snippet,
			Snippet: snippet,
		}
	}

	if tagsSnippet.Valid && tagsSnippet.String != "" {
		snippet := strings.ReplaceAll(tagsSnippet.String, "<b>", opts.TagOpen)
		snippet = strings.ReplaceAll(snippet, "</b>", opts.TagClose)
		result.Tags = &HighlightedText{
			Text:    snippet,
			Snippet: snippet,
		}
	}

	return result, nil
}

// HighlightInText highlights search terms in a given text without database access.
// Uses regex-based matching for simple client-side highlighting.
func HighlightInText(text, query string, opts *HighlightOptions) (*HighlightedText, error) {
	if opts == nil {
		opts = DefaultHighlightOptions()
	}

	if text == "" || query == "" {
		return &HighlightedText{Text: text}, nil
	}

	// Extract search terms from query
	terms := extractSearchTerms(query)

	// Build regex pattern for all terms
	pattern, err := buildHighlightPattern(terms)
	if err != nil {
		return nil, fmt.Errorf("failed to build highlight pattern: %w", err)
	}

	// Find all matches and extract snippet
	snippet := extractSnippet(text, pattern, opts.MaxChars)

	// Apply highlighting
	highlighted := pattern.ReplaceAllStringFunc(snippet, func(match string) string {
		return opts.TagOpen + match + opts.TagClose
	})

	// HTML-escape the text except for our highlight tags
	highlighted = escapeHTMLPreserveTags(highlighted, opts.TagOpen, opts.TagClose)

	return &HighlightedText{
		Text:    highlighted,
		Snippet: snippet,
	}, nil
}

// sanitizeQueryForHighlight removes FTS5 operators that interfere with snippet().
// Keeps the core search terms.
func sanitizeQueryForHighlight(query string) string {
	// Remove FTS5 operators
	operators := []string{"AND", "OR", "NOT", "NEAR", "\"", "(", ")", "*", "^"}
	cleanQuery := query
	for _, op := range operators {
		cleanQuery = strings.ReplaceAll(cleanQuery, op, " ")
	}

	// Collapse multiple spaces
	cleanQuery = strings.Join(strings.Fields(cleanQuery), " ")
	return cleanQuery
}

// extractSearchTerms extracts individual search terms from a query string.
func extractSearchTerms(query string) []string {
	// Remove quotes and split on whitespace
	query = strings.ReplaceAll(query, "\"", "")
	parts := strings.Fields(query)

	var terms []string
	for _, part := range parts {
		// Skip FTS5 operators
		part = strings.ToUpper(part)
		if part == "AND" || part == "OR" || part == "NOT" {
			continue
		}
		// Remove trailing wildcards
		part = strings.TrimSuffix(part, "*")
		if len(part) > 0 {
			terms = append(terms, part)
		}
	}

	if len(terms) == 0 {
		return []string{query}
	}
	return terms
}

// buildHighlightPattern creates a regex pattern for matching search terms.
func buildHighlightPattern(terms []string) (*regexp.Regexp, error) {
	if len(terms) == 0 {
		return regexp.Compile(`(?i).`)
	}

	// Build pattern that matches any of the terms (case-insensitive)
	var patterns []string
	for _, term := range terms {
		// Escape special regex characters in the term
		escaped := regexp.QuoteMeta(term)
		patterns = append(patterns, escaped)
	}

	pattern := "(?i)(" + strings.Join(patterns, "|") + "))"
	return regexp.Compile(pattern)
}

// extractSnippet extracts a snippet of text containing the first match.
func extractSnippet(text string, pattern *regexp.Regexp, maxChars int) string {
	if len(text) <= maxChars {
		return text
	}

	// Find first match
	loc := pattern.FindStringIndex(text)
	if loc == nil {
		// No match found, return beginning of text
		if len(text) <= maxChars {
			return text
		}
		return text[:maxChars] + "..."
	}

	// Calculate snippet boundaries centered on the match
	matchStart := loc[0]
	matchEnd := loc[1]
	matchLength := matchEnd - matchStart

	// Allocate equal space before and after the match
	contextSize := (maxChars - matchLength) / 2
	start := matchStart - contextSize
	end := matchEnd + contextSize

	// Adjust boundaries
	if start < 0 {
		start = 0
		end = maxChars
	}
	if end > len(text) {
		end = len(text)
		start = end - maxChars
		if start < 0 {
			start = 0
		}
	}

	snippet := text[start:end]

	// Add ellipsis if truncated
	if start > 0 {
		snippet = "..." + snippet
	}
	if end < len(text) {
		snippet = snippet + "..."
	}

	return snippet
}

// escapeHTMLPreserveTags escapes HTML but preserves highlight tags.
func escapeHTMLPreserveTags(text, tagOpen, tagClose string) string {
	// Create a placeholder for highlight tags to protect them during escaping
	placeholderOpen := "\x00HLIGHT_OPEN\x00"
	placeholderClose := "\x00HLIGHT_CLOSE\x00"

	text = strings.ReplaceAll(text, tagOpen, placeholderOpen)
	text = strings.ReplaceAll(text, tagClose, placeholderClose)

	// Escape HTML
	text = html.EscapeString(text)

	// Restore highlight tags
	text = strings.ReplaceAll(text, placeholderOpen, tagOpen)
	text = strings.ReplaceAll(text, placeholderClose, tagClose)

	return text
}

// IsHighlightQuery checks if a query contains special syntax that requires FTS5 highlighting.
func IsHighlightQuery(query string) bool {
	return strings.ContainsAny(query, "\"()") ||
		strings.Contains(strings.ToUpper(query), "NEAR") ||
		strings.Contains(strings.ToUpper(query), "OR ")
}

// ExtractMatchedTerms identifies which terms from the query matched in the text.
func ExtractMatchedTerms(text, query string) []string {
	terms := extractSearchTerms(query)
	pattern, err := buildHighlightPattern(terms)
	if err != nil {
		return nil
	}

	var matched []string
	seen := make(map[string]bool)

	matches := pattern.FindAllString(text, -1)
	for _, match := range matches {
		upperMatch := strings.ToUpper(match)
		if !seen[upperMatch] {
			matched = append(matched, match)
			seen[upperMatch] = true
		}
	}

	return matched
}

// TruncateWords intelligently truncates text at word boundaries.
func TruncateWords(text string, maxChars int) string {
	if len(text) <= maxChars {
		return text
	}

	// Find the last word boundary before maxChars
	truncated := text[:maxChars]
	lastSpace := strings.LastIndexAny(truncated, " \t\n\r")

	if lastSpace > 0 {
		return text[:lastSpace] + "..."
	}

	// No word boundary found, truncate at maxChars
	return truncated + "..."
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
	// CJK Extensions and compatibility
	if (r >= 0xF900 && r <= 0xFAFF) || (r >= 0xFF00 && r <= 0xFFEF) {
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

// WordCountApproximate approximates the word count of text.
// Handles both space-separated languages and CJK text.
func WordCountApproximate(text string) int {
	count := 0
	inWord := false

	for _, r := range text {
		if IsCJKCharacter(r) {
			count++
			inWord = false
		} else if unicode.IsSpace(r) {
			inWord = false
		} else {
			if !inWord {
				count++
				inWord = true
			}
		}
	}

	return count
}
