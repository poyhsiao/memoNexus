// Package media provides video metadata extraction (stub implementation).
//go:build novidcodec
// +build novidcodec

package media

import (
	"io"
	"strings"

	"github.com/kimhsiao/memonexus/backend/internal/parser"
)

// VideoExtractor implements Extractor for video files (stub without ffmpeg).
type VideoExtractor struct {
	// GenerateThumbnail indicates whether to generate thumbnails
	generateThumbnail bool
	// ThumbnailTime is the timestamp (in seconds) for thumbnail capture
	thumbnailTime float64
}

// NewVideoExtractor creates a new VideoExtractor.
func NewVideoExtractor() *VideoExtractor {
	return &VideoExtractor{
		generateThumbnail: false,
		thumbnailTime:     5.0,
	}
}

// Extract extracts metadata from video (stub implementation).
func (e *VideoExtractor) Extract(r io.Reader, sourceURL string) (*parser.ParseResult, error) {
	// Return basic result without ffmpeg
	return &parser.ParseResult{
		MediaType:   parser.MediaTypeVideo,
		ContentText: "",
		Title:       basenameFromURL(sourceURL),
		SourceURL:   sourceURL,
	}, nil
}

// SupportedMediaTypes returns the media types handled by this extractor.
func (e *VideoExtractor) SupportedMediaTypes() []parser.MediaType {
	return []parser.MediaType{parser.MediaTypeVideo}
}

// SetThumbnailGeneration enables/disables thumbnail generation.
func (e *VideoExtractor) SetThumbnailGeneration(enable bool) {
	e.generateThumbnail = enable
}

// SetThumbnailTime sets the timestamp for thumbnail capture.
func (e *VideoExtractor) SetThumbnailTime(seconds float64) {
	e.thumbnailTime = seconds
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

	return "Untitled"
}
