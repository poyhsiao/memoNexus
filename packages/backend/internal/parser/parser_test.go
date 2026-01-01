// Package parser tests for content parsing functionality.
package parser

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestNewParserService verifies parser service creation.
func TestNewParserService(t *testing.T) {
	service := NewParserService()

	if service == nil {
		t.Fatal("NewParserService() returned nil")
	}

	if service.httpClient == nil {
		t.Error("httpClient is nil")
	}

	if service.extractors == nil {
		t.Error("extractors map is nil")
	}

	if service.httpClient.Timeout != 30*1000000000 {
		t.Errorf("expected timeout 30s, got %v", service.httpClient.Timeout)
	}
}

// TestRegisterExtractor verifies extractor registration.
func TestRegisterExtractor(t *testing.T) {
	service := NewParserService()

	// Create a mock extractor
	mockExtractor := &mockExtractor{
		mediaTypes: []MediaType{MediaTypeImage, MediaTypeVideo},
	}

	service.RegisterExtractor(mockExtractor)

	// Verify both media types are registered
	if service.extractors[MediaTypeImage] == nil {
		t.Error("MediaTypeImage not registered")
	}

	if service.extractors[MediaTypeVideo] == nil {
		t.Error("MediaTypeVideo not registered")
	}
}

// TestDetectMediaTypeFromURL verifies media type detection from URL.
func TestDetectMediaTypeFromURL(t *testing.T) {
	service := NewParserService()

	tests := []struct {
		name     string
		url      string
		expected MediaType
	}{
		{"image jpg", "https://example.com/photo.jpg", MediaTypeImage},
		{"image jpeg", "https://example.com/photo.jpeg", MediaTypeImage},
		{"image png", "https://example.com/photo.png", MediaTypeImage},
		{"image gif", "https://example.com/photo.gif", MediaTypeImage},
		{"image webp", "https://example.com/photo.webp", MediaTypeImage},
		{"image svg", "https://example.com/photo.svg", MediaTypeImage},
		{"video mp4", "https://example.com/video.mp4", MediaTypeVideo},
		{"video webm", "https://example.com/video.webm", MediaTypeVideo},
		{"pdf", "https://example.com/doc.pdf", MediaTypePDF},
		{"markdown", "https://example.com/doc.md", MediaTypeMarkdown},
		{"markdown long", "https://example.com/doc.markdown", MediaTypeMarkdown},
		{"web html", "https://example.com/page", MediaTypeWeb},
		{"web with query", "https://example.com/photo.jpg?width=200", MediaTypeImage},
		{"web with fragment", "https://example.com/page#section", MediaTypeWeb},
		{"web both", "https://example.com/doc.pdf?x=1#y", MediaTypePDF},
		{"no extension", "https://example.com/page", MediaTypeWeb},
		{"dot in path", "https://example.com/v1.2/docs", MediaTypeWeb},
		{"extension after slash", "https://example.com/folder.jpg/file", MediaTypeWeb},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.detectMediaTypeFromURL(tt.url)
			if result != tt.expected {
				t.Errorf("detectMediaTypeFromURL(%q) = %v, want %v", tt.url, result, tt.expected)
			}
		})
	}
}

// TestDetectMediaTypeFromContentType verifies media type detection from Content-Type header.
func TestDetectMediaTypeFromContentType(t *testing.T) {
	service := NewParserService()

	tests := []struct {
		name     string
		ct       string
		expected MediaType
	}{
		{"image jpeg", "image/jpeg", MediaTypeImage},
		{"image png", "image/png", MediaTypeImage},
		{"video mp4", "video/mp4", MediaTypeVideo},
		{"pdf", "application/pdf", MediaTypePDF},
		{"html", "text/html", MediaTypeWeb},
		{"xhtml", "text/xhtml", MediaTypeWeb},
		{"plain", "text/plain", MediaTypeWeb},
		{"with charset", "text/html; charset=utf-8", MediaTypeWeb},
		{"with spaces", "  application/pdf  ", MediaTypePDF},
		{"unknown", "application/octet-stream", MediaTypeWeb},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.detectMediaTypeFromContentType(tt.ct)
			if result != tt.expected {
				t.Errorf("detectMediaTypeFromContentType(%q) = %v, want %v", tt.ct, result, tt.expected)
			}
		})
	}
}

// TestDetectMediaTypeFromPath verifies media type detection from file path.
func TestDetectMediaTypeFromPath(t *testing.T) {
	service := NewParserService()

	tests := []struct {
		name     string
		path     string
		expected MediaType
	}{
		{"image jpg", "/path/to/photo.jpg", MediaTypeImage},
		{"image png", "/path/to/photo.png", MediaTypeImage},
		{"video mp4", "/path/to/video.mp4", MediaTypeVideo},
		{"pdf", "/path/to/doc.pdf", MediaTypePDF},
		{"markdown", "/path/to/doc.md", MediaTypeMarkdown},
		{"unix path", "/home/user/docs/file.txt", MediaTypeWeb},
		{"windows path", `C:\Users\user\docs\file.txt`, MediaTypeWeb},
		{"no extension", "/path/to/file", MediaTypeWeb},
		{"dot in directory", "/path/v1.2/file", MediaTypeWeb},
		{"extension after dir", "/path/.git/config", MediaTypeWeb},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.detectMediaTypeFromPath(tt.path)
			if result != tt.expected {
				t.Errorf("detectMediaTypeFromPath(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

// TestBasenameFromURL verifies basename extraction from URL.
func TestBasenameFromURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{"simple file", "https://example.com/file.pdf", "file.pdf"},
		{"with query", "https://example.com/file.jpg?width=200", "file.jpg"},
		{"with fragment", "https://example.com/page#section", "page"},
		{"with both", "https://example.com/doc.pdf?a=1#x", "doc.pdf"},
		{"no path", "https://example.com", "example.com"},
		{"trailing slash", "https://example.com/docs/", "example.com"},
		{"empty segment", "https://example.com///", "example.com"},
		{"complex path", "https://example.com/docs/v1/file.pdf", "file.pdf"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := basenameFromURL(tt.url)
			if result != tt.expected {
				t.Errorf("basenameFromURL(%q) = %q, want %q", tt.url, result, tt.expected)
			}
		})
	}
}

// TestBasenameFromPath verifies basename extraction from file path.
func TestBasenameFromPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{"unix path", "/home/user/docs/file.pdf", "file.pdf"},
		{"windows path", `C:\Users\user\docs\file.txt`, "file.txt"},
		{"relative path", "docs/file.md", "file.md"},
		{"just filename", "file.txt", "file.txt"},
		{"trailing slash unix", "/home/user/docs/", ""},
		{"trailing slash windows", `C:\Users\user\docs\`, ""},
		{"nested path", "/a/b/c/d/file.ext", "file.ext"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := basenameFromPath(tt.path)
			if result != tt.expected {
				t.Errorf("basenameFromPath(%q) = %q, want %q", tt.path, result, tt.expected)
			}
		})
	}
}

// TestParseURL_invalidURL verifies error handling for invalid URLs.
func TestParseURL_invalidURL(t *testing.T) {
	service := NewParserService()

	tests := []struct {
		name        string
		url         string
		expectError bool
		errorMsg    string
	}{
		{"invalid url format", "://not-a-url", true, "invalid URL"},
		{"unsupported scheme", "ftp://example.com/file", true, "unsupported URL scheme"},
		{"no scheme", "example.com/file", true, "unsupported URL scheme"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.ParseURL(tt.url)
			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("error message should contain %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

// TestParseURL_HTTP_errors verifies HTTP error handling.
func TestParseURL_HTTP_errors(t *testing.T) {
	service := NewParserService()

	tests := []struct {
		name        string
		statusCode  int
		expectError bool
		errorMsg    string
	}{
		{"404 not found", 404, true, "client error"},
		{"403 forbidden", 403, true, "client error"},
		{"500 server error", 500, true, "server error"},
		{"502 bad gateway", 502, true, "server error"},
		{"200 OK", 200, false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			_, err := service.ParseURL(server.URL)

			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("error message should contain %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

// TestParseURL_noExtractor verifies behavior when no extractor is registered.
func TestParseURL_noExtractor(t *testing.T) {
	service := NewParserService()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	result, err := service.ParseURL(server.URL)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("result is nil")
	}

	if result.SourceURL != server.URL {
		t.Errorf("SourceURL = %q, want %q", result.SourceURL, server.URL)
	}

	if result.MediaType != MediaTypeWeb {
		t.Errorf("MediaType = %v, want %v", result.MediaType, MediaTypeWeb)
	}
}

// TestParseFile verifies file parsing.
func TestParseFile(t *testing.T) {
	service := NewParserService()

	tests := []struct {
		name           string
		path           string
		expectedTitle  string
		expectedType   MediaType
	}{
		{"pdf file", "/path/to/document.pdf", "document.pdf", MediaTypePDF},
		{"image file", "/path/to/photo.jpg", "photo.jpg", MediaTypeImage},
		{"markdown file", "/path/to/doc.md", "doc.md", MediaTypeMarkdown},
		{"unknown file", "/path/to/file.txt", "file.txt", MediaTypeWeb},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.ParseFile(tt.path)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if result == nil {
				t.Fatal("result is nil")
			}

			if result.Title != tt.expectedTitle {
				t.Errorf("Title = %q, want %q", result.Title, tt.expectedTitle)
			}

			if result.MediaType != tt.expectedType {
				t.Errorf("MediaType = %v, want %v", result.MediaType, tt.expectedType)
			}

			if result.SourceURL != tt.path {
				t.Errorf("SourceURL = %q, want %q", result.SourceURL, tt.path)
			}
		})
	}
}

// TestParseReader verifies reader parsing.
func TestParseReader(t *testing.T) {
	service := NewParserService()

	// Test without registered extractor
	content := "test content"
	result, err := service.ParseReader(strings.NewReader(content), MediaTypeWeb)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("result is nil")
	}

	if result.MediaType != MediaTypeWeb {
		t.Errorf("MediaType = %v, want %v", result.MediaType, MediaTypeWeb)
	}
}

// TestParseReader_withExtractor verifies reader parsing with registered extractor.
func TestParseReader_withExtractor(t *testing.T) {
	service := NewParserService()

	mockExtractor := &mockExtractor{
		mediaTypes: []MediaType{MediaTypeImage},
		result: &ParseResult{
			Title:       "Test Image",
			ContentText: "Image content",
			MediaType:   MediaTypeImage,
		},
	}

	service.RegisterExtractor(mockExtractor)

	content := "fake image data"
	result, err := service.ParseReader(strings.NewReader(content), MediaTypeImage)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if result.Title != "Test Image" {
		t.Errorf("Title = %q, want 'Test Image'", result.Title)
	}

	if result.ContentText != "Image content" {
		t.Errorf("ContentText = %q, want 'Image content'", result.ContentText)
	}
}

// mockExtractor is a test implementation of Extractor.
type mockExtractor struct {
	mediaTypes []MediaType
	result     *ParseResult
	errorToReturn error
}

func (m *mockExtractor) Extract(r io.Reader, sourceURL string) (*ParseResult, error) {
	if m.errorToReturn != nil {
		return nil, m.errorToReturn
	}
	if m.result != nil {
		return m.result, nil
	}
	return &ParseResult{
		Title:       "Mock",
		ContentText: "Mock content",
		MediaType:   MediaTypeWeb,
	}, nil
}

func (m *mockExtractor) SupportedMediaTypes() []MediaType {
	return m.mediaTypes
}

// TestParseURL_withExtractor verifies full URL parsing with extractor.
func TestParseURL_withExtractor(t *testing.T) {
	service := NewParserService()

	expectedResult := &ParseResult{
		Title:       "Test Page",
		ContentText: "Page content",
		MediaType:   MediaTypeWeb,
		Tags:        []string{"test", "example"},
	}

	mockExtractor := &mockExtractor{
		mediaTypes: []MediaType{MediaTypeWeb},
		result:     expectedResult,
	}

	service.RegisterExtractor(mockExtractor)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	result, err := service.ParseURL(server.URL)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("result is nil")
	}

	if result.Title != expectedResult.Title {
		t.Errorf("Title = %q, want %q", result.Title, expectedResult.Title)
	}

	if result.ContentText != expectedResult.ContentText {
		t.Errorf("ContentText = %q, want %q", result.ContentText, expectedResult.ContentText)
	}

	if result.SourceURL != server.URL {
		t.Errorf("SourceURL = %q, want %q", result.SourceURL, server.URL)
	}
}

// TestParseURL_extractorError verifies error handling when extractor fails.
func TestParseURL_extractorError(t *testing.T) {
	service := NewParserService()

	mockExtractor := &mockExtractor{
		mediaTypes:   []MediaType{MediaTypeWeb},
		errorToReturn: io.EOF,
	}

	service.RegisterExtractor(mockExtractor)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	_, err := service.ParseURL(server.URL)

	if err == nil {
		t.Error("expected error, got nil")
	} else if !strings.Contains(err.Error(), "extraction failed") {
		t.Errorf("error should mention 'extraction failed', got %q", err.Error())
	}
}

// TestRegisterExtractor_overwrite verifies extractor overwriting.
func TestRegisterExtractor_overwrite(t *testing.T) {
	service := NewParserService()

	extractor1 := &mockExtractor{
		mediaTypes: []MediaType{MediaTypeWeb},
		result: &ParseResult{Title: "Extractor 1"},
	}

	extractor2 := &mockExtractor{
		mediaTypes: []MediaType{MediaTypeWeb},
		result: &ParseResult{Title: "Extractor 2"},
	}

	service.RegisterExtractor(extractor1)
	service.RegisterExtractor(extractor2)

	// Second extractor should overwrite first
	registered := service.extractors[MediaTypeWeb]
	if registered != extractor2 {
		t.Error("second extractor should overwrite first")
	}
}
