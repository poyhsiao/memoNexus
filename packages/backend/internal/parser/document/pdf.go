// Package document provides PDF content extraction.
package document

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/unidoc/unipdf/v3/extractor"
	"github.com/unidoc/unipdf/v3/model"
)

// PDFExtractor implements Extractor for PDF files.
type PDFExtractor struct {
	// Maximum pages to extract (0 = all)
	maxPages int
}

// NewPDFExtractor creates a new PDFExtractor.
func NewPDFExtractor() *PDFExtractor {
	return &PDFExtractor{
		maxPages: 0, // Extract all pages by default
	}
}

// NewPDFExtractorWithMaxPages creates a PDFExtractor with page limit.
func NewPDFExtractorWithMaxPages(maxPages int) *PDFExtractor {
	return &PDFExtractor{
		maxPages: maxPages,
	}
}

// Extract extracts text content from a PDF file.
func (e *PDFExtractor) Extract(r io.Reader, sourceURL string) (*ParseResult, error) {
	// Read PDF data
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read PDF: %w", err)
	}

	// Create PDF reader
	pdfReader, err := model.NewPdfReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to open PDF: %w", err)
	}

	// Get page count
	numPages, err := pdfReader.GetNumPages()
	if err != nil {
		return nil, fmt.Errorf("failed to get page count: %w", err)
	}

	// Extract pages
	var contentBuilder strings.Builder
	pageLimit := numPages
	if e.maxPages > 0 && e.maxPages < numPages {
		pageLimit = e.maxPages
	}

	for i := 1; i <= pageLimit; i++ {
		page, err := pdfReader.GetPage(i)
		if err != nil {
			return nil, fmt.Errorf("failed to get page %d: %w", i, err)
		}

		ex, err := extractor.New(page)
		if err != nil {
			return nil, fmt.Errorf("failed to create extractor for page %d: %w", i, err)
		}

		text, err := ex.ExtractText()
		if err != nil {
			return nil, fmt.Errorf("failed to extract text from page %d: %w", i, err)
		}

		contentBuilder.WriteString(text)
		contentBuilder.WriteString("\n\n")
	}

	// Extract metadata
	title, _ := pdfReader.GetPdfMetaData()

	// Build result
	contentText := strings.TrimSpace(contentBuilder.String())

	return &ParseResult{
		Title:       e.extractTitle(title, sourceURL),
		ContentText: contentText,
		MediaType:   MediaTypePDF,
		WordCount:   countWords(contentText),
		Language:    detectLanguage(contentText),
		SourceURL:   sourceURL,
	}, nil
}

// SupportedMediaTypes returns the media types this extractor handles.
func (e *PDFExtractor) SupportedMediaTypes() []MediaType {
	return []MediaType{MediaTypePDF}
}

// extractTitle extracts title from PDF metadata or source URL.
func (e *PDFExtractor) extractTitle(meta *model.PdfPdfMetaData, sourceURL string) string {
	// Try PDF metadata title
	if title := meta.Title; title != "" {
		return truncate(title, 500)
	}

	// Fallback to filename from URL
	if sourceURL != "" {
		return basenameFromURL(sourceURL)
	}

	return "Untitled PDF"
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
		// Could add more sophisticated detection
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

// basenameFromURL extracts a basename from URL for default title.
func basenameFromURL(sourceURL string) string {
	u := sourceURL
	if i := strings.Index(u, "?"); i > 0 {
		u = u[:i]
	}

	// Get last path segment
	if i := strings.LastIndex(u, "/"); i >= 0 {
		base := u[i+1:]
		if base != "" {
			return base
		}
	}

	return "Untitled"
}
