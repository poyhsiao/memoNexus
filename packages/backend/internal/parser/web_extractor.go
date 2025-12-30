// Package parser provides web content extraction using HTML parsing.
package parser

import (
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// Pre-compiled regex patterns for performance
var (
	whitespaceRegex = regexp.MustCompile(`\s+`)
	cjkRegex        = regexp.MustCompile(`[\p{Han}\p{Hiragana}\p{Katakana}\p{Hangul}]`)
)

// WebExtractor implements Extractor for web/HTML content.
type WebExtractor struct {
	// Minimum content length to consider valid
	minContentLength int
}

// NewWebExtractor creates a new WebExtractor.
func NewWebExtractor() *WebExtractor {
	return &WebExtractor{
		minContentLength: 50, // Minimum 50 characters
	}
}

// Extract extracts content from HTML.
func (e *WebExtractor) Extract(r io.Reader, sourceURL string) (*ParseResult, error) {
	// Parse HTML
	doc, err := html.Parse(r)
	if err != nil {
		return &ParseResult{Error: err}, fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Extract content
	result := &ParseResult{
		MediaType: MediaTypeWeb,
	}

	// Extract title
	result.Title = e.extractTitle(doc)

	// Extract main content
	result.ContentText = e.extractContent(doc)

	// Extract metadata
	result.Author = e.extractAuthor(doc)
	result.PublishDate = e.extractPublishDate(doc)

	// Calculate word count
	result.WordCount = countWords(result.ContentText)

	// Detect language
	result.Language = detectLanguage(result.ContentText)

	// Extract tags from meta keywords
	result.Tags = e.extractTags(doc)

	// Validate minimum content length
	if len(result.ContentText) < e.minContentLength {
		result.ContentText = ""
	}

	return result, nil
}

// SupportedMediaTypes returns the media types this extractor handles.
func (e *WebExtractor) SupportedMediaTypes() []MediaType {
	return []MediaType{MediaTypeWeb}
}

// extractTitle extracts the page title.
func (e *WebExtractor) extractTitle(doc *html.Node) string {
	// Try <title> tag first
	var title string
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "title" {
			if n.FirstChild != nil {
				title = strings.TrimSpace(n.FirstChild.Data)
				return
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
			if title != "" {
				return
			}
		}
	}
	f(doc)

	// Fallback to og:title
	if title == "" {
		title = e.extractMetaProperty(doc, "og:title")
	}

	// Fallback to h1
	if title == "" {
		title = e.extractTextByTag(doc, "h1")
	}

	// Clean title
	title = cleanText(title)

	if title == "" {
		title = "Untitled"
	}

	return truncate(title, 500)
}

// extractContent extracts the main content from HTML.
func (e *WebExtractor) extractContent(doc *html.Node) string {
	var content strings.Builder

	// Strategy: Look for common content containers
	// Priority: article > main > body

	var f func(*html.Node)
	f = func(n *html.Node) {
		// Skip script, style, nav, footer, aside
		if n.Type == html.ElementNode {
			switch n.Data {
			case "script", "style", "nav", "footer", "aside", "header", "noscript":
				return
			}
		}

		// Found main content container
		if n.Type == html.ElementNode && (n.Data == "article" || n.Data == "main") {
			content.WriteString(e.extractNodeText(n))
			return
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}

	// Try to find article/main
	f(doc)

	result := strings.TrimSpace(content.String())

	// Fallback: extract from body if no content found
	if result == "" {
		result = e.extractBodyText(doc)
	}

	return cleanText(result)
}

// extractNodeText extracts text content from a node and its children.
func (e *WebExtractor) extractNodeText(n *html.Node) string {
	var content strings.Builder

	var f func(*html.Node)
	f = func(node *html.Node) {
		if node.Type == html.TextNode {
			text := strings.TrimSpace(node.Data)
			if text != "" {
				content.WriteString(text)
				content.WriteString(" ")
			}
		}

		if node.Type == html.ElementNode {
			// Block elements get newlines
			switch node.Data {
			case "p", "div", "br", "h1", "h2", "h3", "h4", "h5", "h6", "li", "tr":
				content.WriteString("\n")
			}
		}

		for c := node.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}

	f(n)
	return strings.TrimSpace(content.String())
}

// extractBodyText extracts text from the entire body.
func (e *WebExtractor) extractBodyText(doc *html.Node) string {
	var body *html.Node

	var findBody func(*html.Node)
	findBody = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "body" {
			body = n
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findBody(c)
			if body != nil {
				return
			}
		}
	}
	findBody(doc)

	if body != nil {
		return e.extractNodeText(body)
	}

	return ""
}

// extractAuthor extracts the author from meta tags.
func (e *WebExtractor) extractAuthor(doc *html.Node) string {
	// Try meta name="author"
	if author := e.extractMetaName(doc, "author"); author != "" {
		return author
	}

	// Try meta property="article:author"
	if author := e.extractMetaProperty(doc, "article:author"); author != "" {
		return author
	}

	return ""
}

// extractPublishDate extracts the publish date from meta tags.
func (e *WebExtractor) extractPublishDate(doc *html.Node) *time.Time {
	// Try various meta tags
	dateStr := ""
	if dateStr = e.extractMetaProperty(doc, "article:published_time"); dateStr != "" {
	}
	if dateStr == "" {
		dateStr = e.extractMetaName(doc, "date")
	}
	if dateStr == "" {
		dateStr = e.extractMetaProperty(doc, "og:published_time")
	}

	if dateStr == "" {
		return nil
	}

	// Parse date (RFC3339, ISO8601 formats)
	layouts := []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05-07:00",
		"2006-01-02",
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, dateStr); err == nil {
			return &t
		}
	}

	return nil
}

// extractTags extracts tags from meta keywords.
func (e *WebExtractor) extractTags(doc *html.Node) []string {
	keywords := e.extractMetaName(doc, "keywords")
	if keywords == "" {
		return nil
	}

	// Split by comma and clean
	var tags []string
	for _, tag := range strings.Split(keywords, ",") {
		tag = strings.TrimSpace(tag)
		if tag != "" {
			tags = append(tags, tag)
		}
	}

	return tags
}

// extractMetaName extracts content from meta tag by name attribute.
func (e *WebExtractor) extractMetaName(doc *html.Node, name string) string {
	var content string

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "meta" {
			var metaName, metaContent string
			for _, attr := range n.Attr {
				if attr.Key == "name" && attr.Val == name {
					metaName = attr.Val
				}
				if attr.Key == "content" {
					metaContent = attr.Val
				}
			}
			if metaName == name && metaContent != "" {
				content = metaContent
				return
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
			if content != "" {
				return
			}
		}
	}
	f(doc)

	return content
}

// extractMetaProperty extracts content from meta tag by property attribute.
func (e *WebExtractor) extractMetaProperty(doc *html.Node, property string) string {
	var content string

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "meta" {
			var metaProp, metaContent string
			for _, attr := range n.Attr {
				if attr.Key == "property" && attr.Val == property {
					metaProp = attr.Val
				}
				if attr.Key == "content" {
					metaContent = attr.Val
				}
			}
			if metaProp == property && metaContent != "" {
				content = metaContent
				return
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
			if content != "" {
				return
			}
		}
	}
	f(doc)

	return content
}

// extractTextByTag extracts text content from the first occurrence of a tag.
func (e *WebExtractor) extractTextByTag(doc *html.Node, tag string) string {
	var text string

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == tag {
			if n.FirstChild != nil && n.FirstChild.Type == html.TextNode {
				text = strings.TrimSpace(n.FirstChild.Data)
				return
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
			if text != "" {
				return
			}
		}
	}
	f(doc)

	return text
}

// cleanText normalizes whitespace in text.
func cleanText(s string) string {
	// Replace multiple whitespace with single space (using pre-compiled regex)
	s = whitespaceRegex.ReplaceAllString(s, " ")

	return strings.TrimSpace(s)
}

// countWords counts words in text.
func countWords(s string) int {
	words := strings.Fields(s)
	return len(words)
}

// detectLanguage detects the language of text (heuristic).
func detectLanguage(s string) string {
	// Simple heuristic: check for CJK characters (using pre-compiled regex)
	hasCJK := cjkRegex.MatchString(s)

	if hasCJK {
		return "zh" // Default to Chinese for CJK (could be refined)
	}

	return "en" // Default to English
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
