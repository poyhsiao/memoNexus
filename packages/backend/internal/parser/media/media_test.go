//go:build novidcodec
// +build novidcodec

// Package media tests for image and video metadata extraction, and thumbnail generation.
package media

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/kimhsiao/memonexus/backend/internal/parser"
)

// =====================================================
// ImageExtractor Tests
// =====================================================

// TestNewImageExtractor verifies default constructor.
func TestNewImageExtractor(t *testing.T) {
	extractor := NewImageExtractor()

	if extractor == nil {
		t.Fatal("NewImageExtractor() returned nil")
	}

	if extractor.generateThumbnail {
		t.Error("generateThumbnail should be false by default")
	}

	if extractor.thumbnailWidth != 200 {
		t.Errorf("thumbnailWidth = %d, want 200", extractor.thumbnailWidth)
	}

	if extractor.thumbnailHeight != 200 {
		t.Errorf("thumbnailHeight = %d, want 200", extractor.thumbnailHeight)
	}
}

// TestNewImageExtractorWithThumbnails verifies constructor with thumbnail generation.
func TestNewImageExtractorWithThumbnails(t *testing.T) {
	extractor := NewImageExtractorWithThumbnails(150, 100)

	if extractor == nil {
		t.Fatal("NewImageExtractorWithThumbnails() returned nil")
	}

	if !extractor.generateThumbnail {
		t.Error("generateThumbnail should be true")
	}

	if extractor.thumbnailWidth != 150 {
		t.Errorf("thumbnailWidth = %d, want 150", extractor.thumbnailWidth)
	}

	if extractor.thumbnailHeight != 100 {
		t.Errorf("thumbnailHeight = %d, want 100", extractor.thumbnailHeight)
	}
}

// createTestImage creates a simple test image.
func createTestImage(t *testing.T, format string) []byte {
	t.Helper()

	// Create a simple 100x100 red image
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	red := color.RGBA{255, 0, 0, 255}
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, red)
		}
	}

	var buf bytes.Buffer
	switch format {
	case "jpeg":
		err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90})
		if err != nil {
			t.Fatalf("Failed to encode JPEG: %v", err)
		}
	case "png":
		err := png.Encode(&buf, img)
		if err != nil {
			t.Fatalf("Failed to encode PNG: %v", err)
		}
	}

	return buf.Bytes()
}

// TestImageExtractor_Extract_success verifies successful image extraction.
func TestImageExtractor_Extract_success(t *testing.T) {
	extractor := NewImageExtractor()
	imageData := createTestImage(t, "jpeg")

	result, err := extractor.Extract(bytes.NewReader(imageData), "photo/test.jpg")

	if err != nil {
		t.Fatalf("Extract() error = %v", err)
	}

	if result == nil {
		t.Fatal("Extract() returned nil result")
	}

	if result.Title != "test.jpg" {
		t.Errorf("Title = %q, want 'test.jpg'", result.Title)
	}

	if result.MediaType != parser.MediaTypeImage {
		t.Errorf("MediaType = %v, want %v", result.MediaType, parser.MediaTypeImage)
	}

	if !strings.Contains(result.ContentText, "100x100") {
		t.Errorf("ContentText should contain dimensions, got: %s", result.ContentText)
	}

	if result.WordCount != 0 {
		t.Errorf("WordCount = %d, want 0 for images", result.WordCount)
	}

	if result.Language != "" {
		t.Errorf("Language = %q, want empty string for images", result.Language)
	}
}

// TestImageExtractor_Extract_png verifies PNG extraction.
func TestImageExtractor_Extract_png(t *testing.T) {
	extractor := NewImageExtractor()
	imageData := createTestImage(t, "png")

	result, err := extractor.Extract(bytes.NewReader(imageData), "photo.png")

	if err != nil {
		t.Fatalf("Extract() error = %v", err)
	}

	if result == nil {
		t.Fatal("Extract() returned nil result")
	}

	if !strings.Contains(result.ContentText, "png") {
		t.Errorf("ContentText should contain format 'png', got: %s", result.ContentText)
	}
}

// TestImageExtractor_Extract_invalidImage verifies error handling for invalid data.
func TestImageExtractor_Extract_invalidImage(t *testing.T) {
	extractor := NewImageExtractor()
	invalidData := []byte("not an image")

	_, err := extractor.Extract(bytes.NewReader(invalidData), "test.jpg")

	if err == nil {
		t.Error("Extract() with invalid image should return error")
	}

	if !strings.Contains(err.Error(), "failed to decode image") {
		t.Errorf("Error should mention 'failed to decode image', got: %v", err)
	}
}

// TestImageExtractor_Extract_readError verifies error handling for read failures.
func TestImageExtractor_Extract_readError(t *testing.T) {
	extractor := NewImageExtractor()

	errReader := &errorReader{err: io.ErrUnexpectedEOF}

	_, err := extractor.Extract(errReader, "test.jpg")

	if err == nil {
		t.Error("Extract() with failing reader should return error")
	}
}

// TestImageExtractor_SupportedMediaTypes verifies media type list.
func TestImageExtractor_SupportedMediaTypes(t *testing.T) {
	extractor := NewImageExtractor()
	types := extractor.SupportedMediaTypes()

	if len(types) != 1 {
		t.Errorf("SupportedMediaTypes() length = %d, want 1", len(types))
	}

	if types[0] != parser.MediaTypeImage {
		t.Errorf("SupportedMediaTypes()[0] = %v, want %v", types[0], parser.MediaTypeImage)
	}
}

// TestImageExtractor_extractTitle_fromURL verifies title extraction from URL.
func TestImageExtractor_extractTitle_fromURL(t *testing.T) {
	extractor := NewImageExtractor()

	tests := []struct {
		name      string
		sourceURL string
		expected  string
	}{
		{"simple filename", "photo/test.jpg", "test.jpg"},
		{"with path", "/path/to/image.png", "image.png"},
		{"with URL", "https://example.com/photos/test.jpeg", "test.jpeg"},
		{"empty", "", "Untitled Image"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractor.extractTitle(tt.sourceURL)
			if result != tt.expected {
				t.Errorf("extractTitle(%q) = %q, want %q", tt.sourceURL, result, tt.expected)
			}
		})
	}
}

// TestExtractImageMetadata verifies metadata extraction.
func TestExtractImageMetadata(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 200, 150))
	format := "jpeg"

	meta, err := ExtractImageMetadata(img, format)

	if err != nil {
		t.Fatalf("ExtractImageMetadata() error = %v", err)
	}

	if meta.Width != 200 {
		t.Errorf("Width = %d, want 200", meta.Width)
	}

	if meta.Height != 150 {
		t.Errorf("Height = %d, want 150", meta.Height)
	}

	if meta.Format != format {
		t.Errorf("Format = %q, want %q", meta.Format, format)
	}
}

// TestExtractImageMetadata_nilImage verifies nil image handling.
func TestExtractImageMetadata_nilImage(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("ExtractImageMetadata() with nil image should panic")
		}
	}()

	ExtractImageMetadata(nil, "jpeg")
}

// =====================================================
// VideoExtractor Tests
// =====================================================

// TestNewVideoExtractor verifies default constructor.
func TestNewVideoExtractor(t *testing.T) {
	extractor := NewVideoExtractor()

	if extractor == nil {
		t.Fatal("NewVideoExtractor() returned nil")
	}

	if extractor.generateThumbnail {
		t.Error("generateThumbnail should be false by default")
	}

	if extractor.thumbnailTime != 5.0 {
		t.Errorf("thumbnailTime = %f, want 5.0", extractor.thumbnailTime)
	}
}

// TestVideoExtractor_Extract_success verifies stub extraction.
func TestVideoExtractor_Extract_success(t *testing.T) {
	extractor := NewVideoExtractor()
	data := []byte("fake video data")

	result, err := extractor.Extract(bytes.NewReader(data), "video/movie.mp4")

	if err != nil {
		t.Fatalf("Extract() error = %v", err)
	}

	if result == nil {
		t.Fatal("Extract() returned nil result")
	}

	if result.MediaType != parser.MediaTypeVideo {
		t.Errorf("MediaType = %v, want %v", result.MediaType, parser.MediaTypeVideo)
	}

	if result.Title != "movie.mp4" {
		t.Errorf("Title = %q, want 'movie.mp4'", result.Title)
	}

	if result.ContentText != "" {
		t.Errorf("ContentText should be empty for stub, got: %s", result.ContentText)
	}
}

// TestVideoExtractor_Extract_emptyURL verifies empty URL handling.
func TestVideoExtractor_Extract_emptyURL(t *testing.T) {
	extractor := NewVideoExtractor()
	data := []byte("fake video data")

	result, err := extractor.Extract(bytes.NewReader(data), "")

	if err != nil {
		t.Fatalf("Extract() error = %v", err)
	}

	if result.Title != "Untitled" {
		t.Errorf("Title = %q, want 'Untitled'", result.Title)
	}
}

// TestVideoExtractor_Extract_complexURL verifies complex URL handling.
func TestVideoExtractor_Extract_complexURL(t *testing.T) {
	extractor := NewVideoExtractor()

	tests := []struct {
		name      string
		sourceURL string
		expected  string
	}{
		{"with query", "path/video.mp4?quality=high", "video.mp4"},
		{"with fragment", "path/clip.mkv#t=10", "clip.mkv"},
		{"with both", "path/movie.avi?v=1#start", "movie.avi"},
		{"with path", "/path/to/video.mov", "video.mov"},
		{"full URL", "https://example.com/video.mp4", "video.mp4"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := extractor.Extract(bytes.NewReader([]byte("data")), tt.sourceURL)
			if err != nil {
				t.Fatalf("Extract() error = %v", err)
			}
			if result.Title != tt.expected {
				t.Errorf("Title = %q, want %q", result.Title, tt.expected)
			}
		})
	}
}

// TestVideoExtractor_SupportedMediaTypes verifies media type list.
func TestVideoExtractor_SupportedMediaTypes(t *testing.T) {
	extractor := NewVideoExtractor()
	types := extractor.SupportedMediaTypes()

	if len(types) != 1 {
		t.Errorf("SupportedMediaTypes() length = %d, want 1", len(types))
	}

	if types[0] != parser.MediaTypeVideo {
		t.Errorf("SupportedMediaTypes()[0] = %v, want %v", types[0], parser.MediaTypeVideo)
	}
}

// TestVideoExtractor_SetThumbnailGeneration verifies thumbnail generation toggle.
func TestVideoExtractor_SetThumbnailGeneration(t *testing.T) {
	extractor := NewVideoExtractor()

	if extractor.generateThumbnail {
		t.Error("generateThumbnail should be false initially")
	}

	extractor.SetThumbnailGeneration(true)

	if !extractor.generateThumbnail {
		t.Error("generateThumbnail should be true after SetThumbnailGeneration(true)")
	}

	extractor.SetThumbnailGeneration(false)

	if extractor.generateThumbnail {
		t.Error("generateThumbnail should be false after SetThumbnailGeneration(false)")
	}
}

// TestVideoExtractor_SetThumbnailTime verifies thumbnail time setting.
func TestVideoExtractor_SetThumbnailTime(t *testing.T) {
	extractor := NewVideoExtractor()

	extractor.SetThumbnailTime(10.5)

	if extractor.thumbnailTime != 10.5 {
		t.Errorf("thumbnailTime = %f, want 10.5", extractor.thumbnailTime)
	}
}

// TestBasenameFromURL verifies basename extraction logic.
func TestBasenameFromURL(t *testing.T) {
	tests := []struct {
		name      string
		sourceURL string
		expected  string
	}{
		{"simple", "path/video.mp4", "video.mp4"},
		{"with path", "/path/to/video.mp4", "video.mp4"},
		{"with query", "path/video.mp4?v=1", "video.mp4"},
		{"with fragment", "path/video.mp4#t=10", "video.mp4"},
		{"with both", "path/video.mp4?v=1#t=10", "video.mp4"},
		{"with path and query", "/path/video.mp4?quality=high", "video.mp4"},
		{"complex URL", "https://example.com/path/video.mov?download=1", "video.mov"},
		{"empty", "", "Untitled"},
		{"no filename", "https://example.com/", "Untitled"},
		{"trailing slash", "https://example.com/path/", "Untitled"},
		{"only query", "?v=1", "Untitled"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := basenameFromURL(tt.sourceURL)
			if result != tt.expected {
				t.Errorf("basenameFromURL(%q) = %q, want %q", tt.sourceURL, result, tt.expected)
			}
		})
	}
}

// =====================================================
// ThumbnailQueue Tests
// =====================================================

// TestNewThumbnailQueue verifies queue creation.
func TestNewThumbnailQueue(t *testing.T) {
	queue := NewThumbnailQueue(10, 2)

	if queue == nil {
		t.Fatal("NewThumbnailQueue() returned nil")
	}

	if cap(queue.jobs) != 10 {
		t.Errorf("jobs channel capacity = %d, want 10", cap(queue.jobs))
	}

	if queue.workers != 2 {
		t.Errorf("workers = %d, want 2", queue.workers)
	}

	if queue.isRunning {
		t.Error("isRunning should be false initially")
	}
}

// TestThumbnailQueue_Start verifies queue start.
func TestThumbnailThumbnailQueue_Start(t *testing.T) {
	queue := NewThumbnailQueue(5, 1)
	ctx := context.Background()

	queue.Start(ctx)

	if !queue.IsRunning() {
		t.Error("IsRunning() should return true after Start()")
	}

	// Starting again should be idempotent
	queue.Start(ctx)

	if !queue.IsRunning() {
		t.Error("IsRunning() should still be true after second Start()")
	}

	queue.Stop()
}

// TestThumbnailQueue_Stop verifies queue stop.
func TestThumbnailQueue_Stop(t *testing.T) {
	queue := NewThumbnailQueue(5, 1)
	ctx := context.Background()

	queue.Start(ctx)
	queue.Stop()

	if queue.IsRunning() {
		t.Error("IsRunning() should return false after Stop()")
	}

	// Stopping again should be idempotent
	queue.Stop()

	if queue.IsRunning() {
		t.Error("IsRunning() should still be false after second Stop()")
	}
}

// TestThumbnailQueue_Generate_success verifies job generation.
func TestThumbnailQueue_Generate_success(t *testing.T) {
	queue := NewThumbnailQueue(5, 1)
	ctx := context.Background()
	queue.Start(ctx)
	defer queue.Stop()

	tempDir := t.TempDir()
	sourcePath := filepath.Join(tempDir, "source.jpg")
	thumbnailPath := filepath.Join(tempDir, "thumb.jpg")

	// Create source image
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	file, err := os.Create(sourcePath)
	if err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}
	defer file.Close()

	if err := jpeg.Encode(file, img, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatalf("Failed to encode image: %v", err)
	}
	file.Close()

	jobID, err := queue.Generate(sourcePath, thumbnailPath, 50, 50, nil)

	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if jobID == "" {
		t.Error("Generate() returned empty job ID")
	}

	// Wait for job to be processed
	time.Sleep(200 * time.Millisecond)

	// Verify thumbnail was created
	if _, err := os.Stat(thumbnailPath); err != nil {
		t.Errorf("Thumbnail was not created: %v", err)
	}
}

// TestThumbnailQueue_Generate_notRunning verifies error when queue not running.
func TestThumbnailQueue_Generate_notRunning(t *testing.T) {
	queue := NewThumbnailQueue(5, 1)

	_, err := queue.Generate("source.jpg", "thumb.jpg", 50, 50, nil)

	if err == nil {
		t.Error("Generate() should return error when queue not running")
	}

	if !strings.Contains(err.Error(), "not running") {
		t.Errorf("Error should mention 'not running', got: %v", err)
	}
}

// TestThumbnailQueue_Generate_fullQueue verifies error when queue is full.
func TestThumbnailQueue_Generate_fullQueue(t *testing.T) {
	queue := NewThumbnailQueue(1, 0) // 0 workers so jobs won't process
	ctx := context.Background()
	queue.Start(ctx)
	defer queue.Stop()

	// Fill the queue
	tempDir := t.TempDir()
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))

	sourcePath := filepath.Join(tempDir, "source.jpg")
	file, _ := os.Create(sourcePath)
	jpeg.Encode(file, img, &jpeg.Options{Quality: 90})
	file.Close()

	// First job should succeed
	_, err := queue.Generate(sourcePath, "thumb1.jpg", 50, 50, nil)
	if err != nil {
		t.Fatalf("First Generate() error = %v", err)
	}

	// Second job should fail because queue is full
	_, err = queue.Generate(sourcePath, "thumb2.jpg", 50, 50, nil)
	if err == nil {
		t.Error("Second Generate() should return error when queue is full")
	}

	if !strings.Contains(err.Error(), "full") {
		t.Errorf("Error should mention 'full', got: %v", err)
	}
}

// TestThumbnailQueue_GenerateSync verifies synchronous generation.
func TestThumbnailQueue_GenerateSync(t *testing.T) {
	queue := NewThumbnailQueue(5, 1)
	ctx := context.Background()
	queue.Start(ctx)
	defer queue.Stop()

	tempDir := t.TempDir()
	sourcePath := filepath.Join(tempDir, "source.jpg")
	thumbnailPath := filepath.Join(tempDir, "thumb.jpg")

	// Create source image
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	file, err := os.Create(sourcePath)
	if err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}
	defer file.Close()

	if err := jpeg.Encode(file, img, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatalf("Failed to encode image: %v", err)
	}
	file.Close()

	err = queue.GenerateSync(ctx, sourcePath, thumbnailPath, 50, 50)

	if err != nil {
		t.Fatalf("GenerateSync() error = %v", err)
	}

	// Verify thumbnail was created
	if _, err := os.Stat(thumbnailPath); err != nil {
		t.Errorf("Thumbnail was not created: %v", err)
	}

	stats := queue.GetStats()
	if stats.TotalProcessed != 1 {
		t.Errorf("TotalProcessed = %d, want 1", stats.TotalProcessed)
	}
	if stats.SuccessCount != 1 {
		t.Errorf("SuccessCount = %d, want 1", stats.SuccessCount)
	}
}

// TestThumbnailQueue_GenerateSync_invalidSource verifies error handling.
func TestThumbnailQueue_GenerateSync_invalidSource(t *testing.T) {
	queue := NewThumbnailQueue(5, 1)
	ctx := context.Background()
	queue.Start(ctx)
	defer queue.Stop()

	err := queue.GenerateSync(ctx, "nonexistent.jpg", "thumb.jpg", 50, 50)

	if err == nil {
		t.Error("GenerateSync() with invalid source should return error")
	}

	stats := queue.GetStats()
	if stats.FailureCount != 1 {
		t.Errorf("FailureCount = %d, want 1", stats.FailureCount)
	}
}

// TestThumbnailQueue_GetStats verifies statistics retrieval.
func TestThumbnailQueue_GetStats(t *testing.T) {
	queue := NewThumbnailQueue(5, 1)
	ctx := context.Background()
	queue.Start(ctx)
	defer queue.Stop()

	stats := queue.GetStats()

	if stats == nil {
		t.Fatal("GetStats() returned nil")
	}

	// Verify we got a copy, not the original
	stats.TotalProcessed = 999
	newStats := queue.GetStats()
	if newStats.TotalProcessed == 999 {
		t.Error("GetStats() should return a copy, not the original")
	}
}

// TestThumbnailQueue_GetPendingCount verifies pending count.
func TestThumbnailQueue_GetPendingCount(t *testing.T) {
	queue := NewThumbnailQueue(5, 0) // 0 workers so jobs won't process
	ctx := context.Background()
	queue.Start(ctx)
	defer queue.Stop()

	// Initially no pending jobs
	if count := queue.GetPendingCount(); count != 0 {
		t.Errorf("GetPendingCount() = %d, want 0 initially", count)
	}

	// Add a job
	tempDir := t.TempDir()
	sourcePath := filepath.Join(tempDir, "source.jpg")
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	file, _ := os.Create(sourcePath)
	jpeg.Encode(file, img, &jpeg.Options{Quality: 90})
	file.Close()

	_, err := queue.Generate(sourcePath, "thumb.jpg", 50, 50, nil)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Should have 1 pending job
	if count := queue.GetPendingCount(); count != 1 {
		t.Errorf("GetPendingCount() = %d, want 1 after adding job", count)
	}
}

// TestThumbnailQueue_Clear verifies clearing the queue.
func TestThumbnailQueue_Clear(t *testing.T) {
	queue := NewThumbnailQueue(5, 0) // 0 workers so jobs won't process
	ctx := context.Background()
	queue.Start(ctx)
	defer queue.Stop()

	// Add some jobs
	tempDir := t.TempDir()
	sourcePath := filepath.Join(tempDir, "source.jpg")
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	file, _ := os.Create(sourcePath)
	jpeg.Encode(file, img, &jpeg.Options{Quality: 90})
	file.Close()

	for i := 0; i < 3; i++ {
		_, err := queue.Generate(sourcePath, "thumb.jpg", 50, 50, nil)
		if err != nil {
			t.Fatalf("Generate() error = %v", err)
		}
	}

	// Clear the queue
	cleared := queue.Clear()

	if cleared != 3 {
		t.Errorf("Clear() returned %d, want 3", cleared)
	}

	// No more pending jobs
	if count := queue.GetPendingCount(); count != 0 {
		t.Errorf("GetPendingCount() = %d, want 0 after Clear()", count)
	}
}

// =====================================================
// Helper Types
// =====================================================

// errorReader is a test helper that always returns an error.
type errorReader struct {
	err error
}

func (r *errorReader) Read(p []byte) (n int, err error) {
	return 0, r.err
}
