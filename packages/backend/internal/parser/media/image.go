// Package media provides image metadata extraction.
package media

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"

	"github.com/disintegration/imaging"
	_ "golang.org/x/image/webp"

	"github.com/kimhsiao/memonexus/backend/internal/parser"
)

// ImageExtractor implements Extractor for image files.
type ImageExtractor struct {
	// GenerateThumbnail indicates whether to generate thumbnails
	generateThumbnail bool
	// ThumbnailWidth is the width of generated thumbnails
	thumbnailWidth int
	// ThumbnailHeight is the height of generated thumbnails
	thumbnailHeight int
}

// NewImageExtractor creates a new ImageExtractor.
func NewImageExtractor() *ImageExtractor {
	return &ImageExtractor{
		generateThumbnail: false, // Don't generate by default
		thumbnailWidth:     200,
		thumbnailHeight:    200,
	}
}

// NewImageExtractorWithThumbnails creates an ImageExtractor with thumbnail generation.
func NewImageExtractorWithThumbnails(width, height int) *ImageExtractor {
	return &ImageExtractor{
		generateThumbnail: true,
		thumbnailWidth:     width,
		thumbnailHeight:    height,
	}
}

// Extract extracts metadata from an image file.
func (e *ImageExtractor) Extract(r io.Reader, sourceURL string) (*parser.ParseResult, error) {
	// Decode image
	img, format, err := image.Decode(r)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	bounds := img.Bounds()

	// Extract basic metadata
	result := &parser.ParseResult{
		Title:       e.extractTitle(sourceURL),
		ContentText: fmt.Sprintf("%s image (%dx%d)", format, bounds.Dx(), bounds.Dy()),
		MediaType:   parser.MediaTypeImage,
		WordCount:   0, // Images have no word count
		Language:    "", // Images don't have language
		SourceURL:   sourceURL,
	}

	// Extract EXIF data if available
	// TODO: Add EXIF extraction using exif package

	// Generate thumbnail if requested
	if e.generateThumbnail {
		// Use imaging.Resize to generate thumbnail directly from image.Image
		thumbnail := imaging.Resize(img, e.thumbnailWidth, e.thumbnailHeight, imaging.Lanczos)

		// In a real implementation, we would save the thumbnail
		// For now, we just note that thumbnail generation happened
		_ = thumbnail
	}

	return result, nil
}

// SupportedMediaTypes returns the media types this extractor handles.
func (e *ImageExtractor) SupportedMediaTypes() []parser.MediaType {
	return []parser.MediaType{parser.MediaTypeImage}
}

// extractTitle extracts title from source URL.
func (e *ImageExtractor) extractTitle(sourceURL string) string {
	if sourceURL == "" {
		return "Untitled Image"
	}

	// Get filename from URL
	return parser.BasenameFromURLFull(sourceURL)
}

// ImageMetadata represents extracted image metadata.
type ImageMetadata struct {
	Width       int
	Height      int
	Format      string
	HasAlpha    bool
	ColorSpace  string
	Orientation int
	// EXIF data
	CameraMake      string
	CameraModel     string
	DateTime        string
	OrientationEXIF int
	GPSLatitude     float64
	GPSLongitude    float64
}

// ExtractImageMetadata extracts detailed metadata from an image.
func ExtractImageMetadata(img image.Image, format string) (*ImageMetadata, error) {
	bounds := img.Bounds()

	meta := &ImageMetadata{
		Width:  bounds.Dx(),
		Height: bounds.Dy(),
		Format: format,
	}

	// Check for alpha channel
	if _, ok := img.(interface{ Opaque() bool }); ok {
		// Image has Opaque() method
	} else {
		meta.HasAlpha = true
	}

	return meta, nil
}
