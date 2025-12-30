// Package document provides Markdown content extraction.
package document

import (
	"fmt"
	"io"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"

	"github.com/kimhsiao/memonexus/backend/internal/parser"
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
func (e *MarkdownExtractor) Extract(r io.Reader, sourceURL string) (*parser.ParseResult, error) {
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

	return &parser.ParseResult{
		Title:       title,
		ContentText: contentText,
		MediaType:   parser.MediaTypeMarkdown,
		WordCount:   parser.CountWords(contentText),
		Language:    parser.DetectLanguage(contentText),
		SourceURL:   sourceURL,
	}, nil
}

// SupportedMediaTypes returns the media types this extractor handles.
func (e *MarkdownExtractor) SupportedMediaTypes() []parser.MediaType {
	return []parser.MediaType{parser.MediaTypeMarkdown}
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
				return parser.Truncate(title, 500)
			}
		} else if line != "" {
			// If first non-empty line is not a heading, use it as title
			return parser.Truncate(line, 100)
		}
	}

	// Fallback to first line
	if len(lines) > 0 && strings.TrimSpace(lines[0]) != "" {
		return parser.Truncate(strings.TrimSpace(lines[0]), 100)
	}

	return "Untitled"
}

// markdownToPlainText converts markdown to plain text.
func (e *MarkdownExtractor) markdownToPlainText(markdown string) string {
	// Use goldmark to parse markdown
	md := goldmark.New()
	node := md.Parser().Parse(text.NewReader([]byte(markdown)))

	var builder strings.Builder
	source := []byte(markdown)

	// Traverse AST and extract text
	ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		switch n.Kind() {
		case ast.KindText:
			textNode := n.(*ast.Text)
			segment := textNode.Segment
			builder.Write(segment.Value(source))
		case ast.KindParagraph:
			builder.WriteString("\n\n")
		case ast.KindHeading:
			// Add extra newline before headings
			builder.WriteString("\n")
		case ast.KindList:
			builder.WriteString("\n")
		case ast.KindListItem:
			builder.WriteString("• ")
		case ast.KindFencedCodeBlock:
			code := n.(*ast.FencedCodeBlock)
			builder.WriteString("\n```\n")
			for i := 0; i < code.Lines().Len(); i++ {
				line := code.Lines().At(i)
				builder.Write(line.Value(source))
			}
			builder.WriteString("\n```\n\n")
			return ast.WalkSkipChildren, nil
		}

		return ast.WalkContinue, nil
	})

	return strings.TrimSpace(builder.String())
}
