// Package document provides Markdown content extraction.
package document

import (
	"fmt"
	"io"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// MarkdownExtractor implements Extractor for Markdown files.
type MarkdownExtractor struct {
	// IncludeFrontmatter indicates whether to include YAML frontmatter
	includeFrontmatter bool
}

// NewMarkdownExtractor creates a new MarkdownExtractor.
func NewMarkdownExtractor() *MarkdownExtractor {
	return &MarkdownExtractor{
		includeFrontmatter: false, // Exclude frontmatter by default
	}
}

// NewMarkdownExtractorWithFrontmatter creates a MarkdownExtractor that includes frontmatter.
func NewMarkdownExtractorWithFrontmatter() *MarkdownExtractor {
	return &MarkdownExtractor{
		includeFrontmatter: true,
	}
}

// Extract extracts content from a Markdown file.
func (e *MarkdownExtractor) Extract(r io.Reader, sourceURL string) (*ParseResult, error) {
	// Read markdown content
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read markdown: %w", err)
	}

	markdown := string(data)

	// Remove frontmatter if needed
	if !e.includeFrontmatter {
		markdown = e.removeFrontmatter(markdown)
	}

	// Extract title (first heading or frontmatter title)
	title := e.extractTitle(markdown)

	// Convert markdown to plain text for content
	contentText := e.markdownToPlainText(markdown)

	return &ParseResult{
		Title:       title,
		ContentText: contentText,
		MediaType:   MediaTypeMarkdown,
		WordCount:   countWords(contentText),
		Language:    detectLanguage(contentText),
		SourceURL:   sourceURL,
	}, nil
}

// SupportedMediaTypes returns the media types this extractor handles.
func (e *MarkdownExtractor) SupportedMediaTypes() []MediaType {
	return []MediaType{MediaTypeMarkdown}
}

// removeFrontmatter removes YAML frontmatter from markdown.
func (e *MarkdownExtractor) removeFrontmatter(markdown string) string {
	lines := strings.Split(markdown, "\n")
	if len(lines) < 2 {
		return markdown
	}

	// Check for YAML frontmatter
	if strings.TrimSpace(lines[0]) != "---" {
		return markdown
	}

	// Find end of frontmatter
	for i, line := range lines[1:] {
		if strings.TrimSpace(line) == "---" {
			// Return content after frontmatter
			return strings.Join(lines[i+2:], "\n")
		}
	}

	return markdown
}

// extractTitle extracts the title from markdown.
func (e *MarkdownExtractor) extractTitle(markdown string) string {
	// Remove frontmatter first
	content := e.removeFrontmatter(markdown)
	lines := strings.Split(content, "\n")

	// Look for first heading
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") {
			// Remove # symbols and trim
			title := strings.TrimLeft(line, "#")
			title = strings.TrimSpace(title)
			if title != "" {
				return truncate(title, 500)
			}
		} else if line != "" {
			// If first non-empty line is not a heading, use it as title
			return truncate(line, 100)
		}
	}

	// Fallback to first line
	if len(lines) > 0 && strings.TrimSpace(lines[0]) != "" {
		return truncate(strings.TrimSpace(lines[0]), 100)
	}

	return "Untitled"
}

// markdownToPlainText converts markdown to plain text.
func (e *MarkdownExtractor) markdownToPlainText(markdown string) string {
	// Use goldmark to parse markdown
	md := goldmark.New()
	node := md.Parser().Parse(text.NewReader([]byte(markdown)))

	var builder strings.Builder

	// Traverse AST and extract text
	ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		switch n.Kind() {
		case ast.KindText:
			text := n.(*ast.Text).Value
			builder.Write(text)
		case ast.KindSoftLineBreak:
			builder.WriteString("\n")
		case ast.KindHardLineBreak:
			builder.WriteString("\n")
		case ast.KindParagraph:
			builder.WriteString("\n\n")
		case ast.KindHeading:
			// Add extra newline before headings
			builder.WriteString("\n")
		case ast.KindList:
			builder.WriteString("\n")
		case ast.KindListItem:
			builder.WriteString("• ")
		case ast.KindCodeBlock:
			code := n.(*ast.FencedCodeBlock)
			builder.WriteString("\n```\n")
			for i := 0; i < code.Lines().Len(); i++ {
				line := code.Lines().At(i)
				builder.Write(line.Value(sourceURL))
			}
			builder.WriteString("\n```\n\n")
			return ast.WalkSkipChildren, nil
		}

		return ast.WalkContinue, nil
	})

	return strings.TrimSpace(builder.String())
}

// countWords counts words in text.
func countWords(s string) int {
	words := strings.Fields(s)
	return len(words)
}

// detectLanguage detects the language of text (heuristic).
func detectLanguage(s string) string {
	// Simple heuristic: check for non-ASCII characters
	hasNonASCII := false
	for _, r := range s {
		if r > 127 {
			hasNonASCII = true
			break
		}
	}

	if hasNonASCII {
		return "unknown"
	}

	return "en"
}

// truncate truncates string to max length.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}

	// Try to truncate at word boundary
	if i := strings.LastIndex(s[:maxLen], " "); i > 0 {
		return s[:i] + "..."
	}

	return s[:maxLen] + "..."
}
