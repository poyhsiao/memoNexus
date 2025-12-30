// Package document provides PDF content extraction.
package document

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/unidoc/unipdf/v3/extractor"
	"github.com/unidoc/unipdf/v3/model"

	"github.com/kimhsiao/memonexus/backend/internal/parser"
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
func (e *PDFExtractor) Extract(r io.Reader, sourceURL string) (*parser.ParseResult, error) {
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
	_, _ = pdfReader.GetCatalogMetadata()

	// Build result
	contentText := strings.TrimSpace(contentBuilder.String())

	// Try to get title from metadata - use PdfInfo if available
	var title string
	pdfInfo, err := pdfReader.GetPdfInfo()
	if err == nil && pdfInfo != nil && pdfInfo.Title != nil {
		title = pdfInfo.Title.Str()
	}

	return &parser.ParseResult{
		Title:       e.extractTitle(title, sourceURL),
		ContentText: contentText,
		MediaType:   parser.MediaTypePDF,
		WordCount:   parser.CountWords(contentText),
		Language:    parser.DetectLanguage(contentText),
		SourceURL:   sourceURL,
	}, nil
}

// SupportedMediaTypes returns the media types this extractor handles.
func (e *PDFExtractor) SupportedMediaTypes() []parser.MediaType {
	return []parser.MediaType{parser.MediaTypePDF}
}

// extractTitle extracts title from PDF metadata or source URL.
func (e *PDFExtractor) extractTitle(metaTitle string, sourceURL string) string {
	// Try PDF metadata title
	if metaTitle != "" {
		return parser.Truncate(metaTitle, 500)
	}

	// Fallback to filename from URL
	if sourceURL != "" {
		return parser.BasenameFromURLFull(sourceURL)
	}

	return "Untitled PDF"
}
