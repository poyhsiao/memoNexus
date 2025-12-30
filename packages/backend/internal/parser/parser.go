// Package parser provides content extraction and parsing capabilities.
package parser

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// MediaType represents the type of content being parsed.
type MediaType string

const (
	MediaTypeWeb      MediaType = "web"
	MediaTypeImage    MediaType = "image"
	MediaTypeVideo    MediaType = "video"
	MediaTypePDF      MediaType = "pdf"
	MediaTypeMarkdown MediaType = "markdown"
)

// ParseResult represents the result of content parsing.
type ParseResult struct {
	// Content metadata
	Title       string
	ContentText string
	MediaType   MediaType
	Tags        []string

	// Source information
	SourceURL    string
	CanonicalURL string

	// Extraction metadata
	Author      string
	PublishDate *time.Time
	WordCount   int
	Language    string

	// Error if parsing failed
	Error error
}

// ContentParser defines the interface for content parsing operations.
type ContentParser interface {
	// ParseURL fetches and parses content from a URL.
	ParseURL(sourceURL string) (*ParseResult, error)

	// ParseFile parses content from a local file.
	ParseFile(filePath string) (*ParseResult, error)

	// ParseReader parses content from an io.Reader.
	ParseReader(r io.Reader, mediaType MediaType) (*ParseResult, error)
}

// ParserService implements ContentParser with pluggable extractors.
type ParserService struct {
	httpClient *http.Client
	extractors map[MediaType]Extractor
}

// Extractor defines the interface for type-specific content extraction.
type Extractor interface {
	// Extract extracts content from the reader.
	Extract(r io.Reader, sourceURL string) (*ParseResult, error)

	// SupportedMediaTypes returns the media types this extractor handles.
	SupportedMediaTypes() []MediaType
}

// NewParserService creates a new ParserService.
func NewParserService() *ParserService {
	return &ParserService{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        10,
				IdleConnTimeout:     30 * time.Second,
				DisableCompression:  false,
				DisableKeepAlives:   false,
				MaxIdleConnsPerHost: 10,
			},
		},
		extractors: make(map[MediaType]Extractor),
	}
}

// RegisterExtractor registers a content extractor for a media type.
func (p *ParserService) RegisterExtractor(extractor Extractor) {
	for _, mt := range extractor.SupportedMediaTypes() {
		p.extractors[mt] = extractor
	}
}

// ParseURL fetches and parses content from a URL.
func (p *ParserService) ParseURL(sourceURL string) (*ParseResult, error) {
	// Validate URL
	parsedURL, err := url.Parse(sourceURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return nil, fmt.Errorf("unsupported URL scheme: %s", parsedURL.Scheme)
	}

	// Fetch content
	resp, err := p.httpClient.Get(sourceURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	// Detect media type from URL
	mediaType := p.detectMediaTypeFromURL(sourceURL)
	if mediaType == MediaTypeWeb {
		// Also check Content-Type header
		if ct := resp.Header.Get("Content-Type"); ct != "" {
			mediaType = p.detectMediaTypeFromContentType(ct)
		}
	}

	// Get extractor
	extractor, ok := p.extractors[mediaType]
	if !ok {
		// Return basic result if no extractor
		return &ParseResult{
			SourceURL:  sourceURL,
			MediaType:  mediaType,
			ContentText: "",
			Title:       basenameFromURL(sourceURL),
		}, nil
	}

	// Extract content
	result, err := extractor.Extract(resp.Body, sourceURL)
	if err != nil {
		return nil, fmt.Errorf("extraction failed: %w", err)
	}

	result.SourceURL = sourceURL
	return result, nil
}

// ParseFile parses content from a local file.
// Note: This is a placeholder for future file-based parsing.
func (p *ParserService) ParseFile(filePath string) (*ParseResult, error) {
	// TODO: Implement file parsing
	// For now, return a basic result with media type detected from extension
	mediaType := p.detectMediaTypeFromPath(filePath)

	return &ParseResult{
		SourceURL:   filePath,
		MediaType:   mediaType,
		ContentText: "",
		Title:       basenameFromPath(filePath),
	}, nil
}

// ParseReader parses content from an io.Reader.
func (p *ParserService) ParseReader(r io.Reader, mediaType MediaType) (*ParseResult, error) {
	extractor, ok := p.extractors[mediaType]
	if !ok {
		return &ParseResult{
			MediaType:   mediaType,
			ContentText: "",
		}, nil
	}

	return extractor.Extract(r, "")
}

// detectMediaTypeFromURL detects media type from URL extension.
func (p *ParserService) detectMediaTypeFromURL(sourceURL string) MediaType {
	u := sourceURL
	if i := strings.Index(u, "?"); i > 0 {
		u = u[:i]
	}

	ext := strings.ToLower(u[strings.LastIndex(u, "."):])
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg", ".ico":
		return MediaTypeImage
	case ".mp4", ".webm", ".ogg", ".mov", ".avi":
		return MediaTypeVideo
	case ".pdf":
		return MediaTypePDF
	case ".md", ".markdown":
		return MediaTypeMarkdown
	default:
		return MediaTypeWeb
	}
}

// detectMediaTypeFromContentType detects media type from Content-Type header.
func (p *ParserService) detectMediaTypeFromContentType(ct string) MediaType {
	ct = strings.ToLower(ct)

	// Parse media type (skip charset and other parameters)
	if i := strings.Index(ct, ";"); i > 0 {
		ct = ct[:i]
	}

	ct = strings.TrimSpace(ct)

	switch {
	case strings.HasPrefix(ct, "image/"):
		return MediaTypeImage
	case strings.HasPrefix(ct, "video/"):
		return MediaTypeVideo
	case ct == "application/pdf":
		return MediaTypePDF
	case strings.HasPrefix(ct, "text/html"),
		strings.HasPrefix(ct, "text/xhtml"),
		strings.HasPrefix(ct, "text/plain"):
		return MediaTypeWeb
	default:
		return MediaTypeWeb
	}
}

// detectMediaTypeFromPath detects media type from file path extension.
func (p *ParserService) detectMediaTypeFromPath(path string) MediaType {
	ext := strings.ToLower(path[strings.LastIndex(path, "."):])

	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg":
		return MediaTypeImage
	case ".mp4", ".webm", ".mov", ".avi":
		return MediaTypeVideo
	case ".pdf":
		return MediaTypePDF
	case ".md", ".markdown":
		return MediaTypeMarkdown
	default:
		return MediaTypeWeb
	}
}

// basenameFromURL extracts a basename from URL for default title.
func basenameFromURL(sourceURL string) string {
	u := sourceURL
	if i := strings.Index(u, "?"); i > 0 {
		u = u[:i]
	}
	if i := strings.Index(u, "#"); i > 0 {
		u = u[:i]
	}

	// Get last path segment
	if i := strings.LastIndex(u, "/"); i >= 0 {
		base := u[i+1:]
		if base != "" {
			return base
		}
	}

	// Fallback to hostname
	if parsedURL, err := url.Parse(sourceURL); err == nil {
		return parsedURL.Hostname()
	}

	return "Untitled"
}

// basenameFromPath extracts basename from file path.
func basenameFromPath(path string) string {
	if i := strings.LastIndex(path, "/"); i >= 0 {
		return path[i+1:]
	}
	if i := strings.LastIndex(path, "\\"); i >= 0 {
		return path[i+1:]
	}
	return path
}
