// Package media provides video metadata extraction.
//go:build !novidcodec
// +build !novidcodec

package media

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/giorgisio/goavcodec/avcodec"
	"github.com/giorgisio/goavcodec/avformat"
	"github.com/giorgisio/goavutil"
)

// VideoExtractor implements Extractor for video files.
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
		thumbnailTime:     5.0, // Default to 5 seconds
	}
}

// NewVideoExtractorWithThumbnails creates a VideoExtractor with thumbnail generation.
func NewVideoExtractorWithThumbnails(thumbnailTime float64) *VideoExtractor {
	return &VideoExtractor{
		generateThumbnail: true,
		thumbnailTime:     thumbnailTime,
	}
}

// Extract extracts metadata from a video file.
func (e *VideoExtractor) Extract(r io.Reader, sourceURL string) (*ParseResult, error) {
	// For video files, we need to seek, so we first save to a temp file
	// This is a limitation of ffmpeg/goavcodec - it requires file paths

	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "video-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Copy reader to temp file
	_, err = io.Copy(tmpFile, r)
	if err != nil {
		return nil, fmt.Errorf("failed to write video data: %w", err)
	}

	// Close file before reading with ffmpeg
	tmpFile.Close()

	// Open video file with ffmpeg
	formatCtx := avformat.AvformatAllocContext()
	if formatCtx == nil {
		return nil, fmt.Errorf("failed to allocate format context")
	}
	defer avformat.AvformatClose(formatCtx)

	// Open video file
	if avformat.AvformatOpenInput(formatCtx, tmpFile.Name(), nil, nil) != 0 {
		return nil, fmt.Errorf("failed to open video file")
	}
	defer avformat.AvformatCloseInput(formatCtx)

	// Retrieve stream information
	if formatCtx.AvformatFindStreamInfo(nil) < 0 {
		return nil, fmt.Errorf("failed to retrieve stream info")
	}

	// Extract metadata
	duration := formatCtx.Duration()
	width := 0
	height := 0
	videoCodec := ""

	// Find video stream
	for i := uint32(0); i < formatCtx.NbStreams(); i++ {
		stream := formatCtx.Streams()[i]
		if stream == nil {
			continue
		}

		codecParams := stream.CodecPar()
		if codecParams == nil {
			continue
		}

		// Check if this is a video stream
		if codecParams.AvCodecGetType() == avcodec.AVMEDIA_TYPE_VIDEO {
			width = int(codecParams.AvCodecGetWidth())
			height = int(codecParams.AvCodecGetHeight())

			// Get codec name
			codec := avcodec.AvcodecFindDecoder(codecParams.AvCodecGetId())
			if codec != nil && codec.Name() != nil {
				videoCodec = goavutil.StrFromC(codec.Name())
			}

			break
		}
	}

	// Calculate duration in seconds
	durationSec := float64(duration) / float64(avutil.AV_TIME_BASE)

	// Build result
	title := e.extractTitle(sourceURL)
	contentText := fmt.Sprintf("Video (%s, %dx%d, %.2fs)", videoCodec, width, height, durationSec)

	result := &ParseResult{
		Title:       title,
		ContentText: contentText,
		MediaType:   MediaTypeVideo,
		WordCount:   0,
		Language:    "",
		SourceURL:   sourceURL,
	}

	// Generate thumbnail if requested
	if e.generateThumbnail {
		// TODO: Implement thumbnail generation
		// This requires more complex ffmpeg integration
	}

	return result, nil
}

// SupportedMediaTypes returns the media types this extractor handles.
func (e *VideoExtractor) SupportedMediaTypes() []MediaType {
	return []MediaType{MediaTypeVideo}
}

// extractTitle extracts title from source URL.
func (e *VideoExtractor) extractTitle(sourceURL string) string {
	if sourceURL == "" {
		return "Untitled Video"
	}

	// Get filename from URL
	return basenameFromURL(sourceURL)
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

// VideoMetadata represents extracted video metadata.
type VideoMetadata struct {
	Width       int
	Height      int
	Duration    float64 // in seconds
	Format      string
	VideoCodec   string
	AudioCodec  string
	FrameRate   float64
	BitRate     int64
	HasAudio    bool
	HasSubtitle bool
}
