// Package media provides thumbnail generation with background queue processing.
// T221: Thumbnail generation background queue (non-blocking UI).
package media

import (
	"context"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/kimhsiao/memonexus/backend/internal/logging"
)

// ThumbnailJob represents a thumbnail generation job.
type ThumbnailJob struct {
	ID          string
	SourcePath  string
	ThumbnailPath string
	Width       int
	Height      int
	Priority    int // 0=low, 1=normal, 2=high
	CreatedAt   time.Time
	Callback    func(error)
}

// ThumbnailQueue manages background thumbnail generation.
type ThumbnailQueue struct {
	jobs       chan *ThumbnailJob
	workers    int
	wg         sync.WaitGroup
	stopCh     chan struct{}
	mu         sync.Mutex
	isRunning  bool
	stats      *ThumbnailStats
}

// ThumbnailStats holds thumbnail generation statistics.
type ThumbnailStats struct {
	TotalProcessed int
	SuccessCount   int
	FailureCount   int
	PendingCount   int
	AvgDurationMs  int64
}

// NewThumbnailQueue creates a new thumbnail generation queue.
// T221: Background queue for non-blocking thumbnail generation.
func NewThumbnailQueue(queueSize int, workers int) *ThumbnailQueue {
	return &ThumbnailQueue{
		jobs:    make(chan *ThumbnailJob, queueSize),
		workers: workers,
		stopCh:  make(chan struct{}),
		stats:   &ThumbnailStats{},
	}
}

// Start starts the thumbnail generation workers.
func (q *ThumbnailQueue) Start(ctx context.Context) {
	q.mu.Lock()
	if q.isRunning {
		q.mu.Unlock()
		return
	}
	q.isRunning = true
	q.mu.Unlock()

	logging.Info("Starting thumbnail generation queue",
		map[string]interface{}{
			"workers":   q.workers,
			"queue_size": cap(q.jobs),
		})

	// Start worker goroutines
	for i := 0; i < q.workers; i++ {
		q.wg.Add(1)
		go q.worker(ctx, i)
	}
}

// Stop stops the thumbnail generation workers gracefully.
func (q *ThumbnailQueue) Stop() {
	q.mu.Lock()
	if !q.isRunning {
		q.mu.Unlock()
		return
	}
	q.isRunning = false
	q.mu.Unlock()

	// Signal stop to all workers
	close(q.stopCh)

	// Wait for workers to finish
	q.wg.Wait()

	logging.Info("Thumbnail generation queue stopped",
		map[string]interface{}{
			"total_processed": q.stats.TotalProcessed,
			"success_count":    q.stats.SuccessCount,
			"failure_count":    q.stats.FailureCount,
		})
}

// Generate requests thumbnail generation without blocking.
// Returns immediately with a job ID. The thumbnail is generated in the background.
func (q *ThumbnailQueue) Generate(sourcePath, thumbnailPath string, width, height int, callback func(error)) (string, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if !q.isRunning {
		return "", fmt.Errorf("thumbnail queue is not running")
	}

	job := &ThumbnailJob{
		ID:            fmt.Sprintf("%d-%s", time.Now().UnixNano(), filepath.Base(sourcePath)),
		SourcePath:    sourcePath,
		ThumbnailPath: thumbnailPath,
		Width:         width,
		Height:        height,
		Priority:      1, // Normal priority by default
		CreatedAt:     time.Now(),
		Callback:      callback,
	}

	// Try to enqueue without blocking
	select {
	case q.jobs <- job:
		q.stats.PendingCount++
		logging.Info("Thumbnail job enqueued",
			map[string]interface{}{
				"job_id":       job.ID,
				"source_path":  sourcePath,
				"thumbnail_path": thumbnailPath,
			})
		return job.ID, nil
	default:
		return "", fmt.Errorf("thumbnail queue is full (capacity: %d)", cap(q.jobs))
	}
}

// GenerateSync generates a thumbnail synchronously (blocks until complete).
// Use this for critical thumbnails that must be generated immediately.
func (q *ThumbnailQueue) GenerateSync(ctx context.Context, sourcePath, thumbnailPath string, width, height int) error {
	startTime := time.Now()

	err := generateThumbnail(ctx, sourcePath, thumbnailPath, width, height)

	duration := time.Since(startTime).Milliseconds()

	// Update stats
	q.mu.Lock()
	if err != nil {
		q.stats.FailureCount++
	} else {
		q.stats.SuccessCount++
	}
	q.stats.TotalProcessed++

	// Update average duration
	if q.stats.TotalProcessed > 0 {
		totalDuration := q.stats.AvgDurationMs*(int64(q.stats.TotalProcessed-1)) + duration
		q.stats.AvgDurationMs = totalDuration / int64(q.stats.TotalProcessed)
	}
	q.mu.Unlock()

	return err
}

// worker processes thumbnail generation jobs from the queue.
func (q *ThumbnailQueue) worker(ctx context.Context, workerID int) {
	defer q.wg.Done()

	logging.Info("Thumbnail worker started",
		map[string]interface{}{
			"worker_id": workerID,
		})

	for {
		select {
		case <-ctx.Done():
			logging.Info("Thumbnail worker stopping (context done)",
				map[string]interface{}{
					"worker_id": workerID,
				})
			return
		case <-q.stopCh:
			logging.Info("Thumbnail worker stopping (stop signal)",
				map[string]interface{}{
					"worker_id": workerID,
				})
			return
		case job := <-q.jobs:
			q.processJob(ctx, job, workerID)
		}
	}
}

// processJob processes a single thumbnail generation job.
func (q *ThumbnailQueue) processJob(ctx context.Context, job *ThumbnailJob, workerID int) {
	startTime := time.Now()

	logging.Info("Processing thumbnail job",
		map[string]interface{}{
			"job_id":     job.ID,
			"worker_id":   workerID,
			"source":     job.SourcePath,
			"priority":   job.Priority,
		})

	err := generateThumbnail(ctx, job.SourcePath, job.ThumbnailPath, job.Width, job.Height)

	duration := time.Since(startTime).Milliseconds()

	// Update stats
	q.mu.Lock()
	q.stats.PendingCount--
	q.stats.TotalProcessed++
	if err != nil {
		q.stats.FailureCount++
	} else {
		q.stats.SuccessCount++
	}

	// Update average duration
	if q.stats.TotalProcessed > 0 {
		totalDuration := q.stats.AvgDurationMs*(int64(q.stats.TotalProcessed-1)) + duration
		q.stats.AvgDurationMs = totalDuration / int64(q.stats.TotalProcessed)
	}
	q.mu.Unlock()

	// Call callback if provided
	if job.Callback != nil {
		// Run callback in separate goroutine to avoid blocking worker
		go job.Callback(err)
	}

	if err != nil {
		logging.Error("Thumbnail generation failed", err,
			map[string]interface{}{
				"job_id":       job.ID,
				"worker_id":     workerID,
				"duration_ms":   duration,
				"source_path":   job.SourcePath,
			})
	} else {
		logging.Info("Thumbnail generated successfully",
			map[string]interface{}{
				"job_id":       job.ID,
				"worker_id":     workerID,
				"duration_ms":   duration,
				"thumbnail_path": job.ThumbnailPath,
			})
	}
}

// generateThumbnail generates a thumbnail from the source image.
func generateThumbnail(ctx context.Context, sourcePath, thumbnailPath string, width, height int) error {
	// Open source image file
	file, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to open source image: %w", err)
	}
	defer file.Close()

	// Decode image
	img, _, err := image.Decode(file)
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	// Calculate thumbnail dimensions maintaining aspect ratio
	srcBounds := img.Bounds()
	srcW := srcBounds.Dx()
	srcH := srcBounds.Dy()

	dstW, dstH := width, height
	if srcW > srcH {
		// Landscape: constrain by width
		dstH = srcH * width / srcW
	} else {
		// Portrait: constrain by height
		dstW = srcW * height / srcH
	}

	// Create thumbnail image
	thumbnail := image.NewRGBA(image.Rect(0, 0, dstW, dstH))

	// Resize using simple nearest-neighbor (for better quality, consider using imaging library)
	// This is a basic implementation - production code should use proper resampling
	for y := 0; y < dstH; y++ {
		for x := 0; x < dstW; x++ {
			srcX := x * srcW / dstW
			srcY := y * srcH / dstH
			thumbnail.Set(x, y, img.At(srcX, srcY))
		}
	}

	// Ensure thumbnail directory exists
 thumbnailDir := filepath.Dir(thumbnailPath)
	if err := os.MkdirAll(thumbnailDir, 0755); err != nil {
		return fmt.Errorf("failed to create thumbnail directory: %w", err)
	}

	// Create thumbnail file
	outFile, err := os.Create(thumbnailPath)
	if err != nil {
		return fmt.Errorf("failed to create thumbnail file: %w", err)
	}
	defer outFile.Close()

	// Encode thumbnail as JPEG
	if err := jpeg.Encode(outFile, thumbnail, &jpeg.Options{Quality: 85}); err != nil {
		return fmt.Errorf("failed to encode thumbnail: %w", err)
	}

	return nil
}

// GetStats returns current thumbnail generation statistics.
func (q *ThumbnailQueue) GetStats() *ThumbnailStats {
	q.mu.Lock()
	defer q.mu.Unlock()

	// Return a copy to avoid external modification
	return &ThumbnailStats{
		TotalProcessed: q.stats.TotalProcessed,
		SuccessCount:   q.stats.SuccessCount,
		FailureCount:   q.stats.FailureCount,
		PendingCount:   q.stats.PendingCount,
		AvgDurationMs:  q.stats.AvgDurationMs,
	}
}

// IsRunning returns whether the queue is running.
func (q *ThumbnailQueue) IsRunning() bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.isRunning
}

// GetPendingCount returns the number of pending jobs.
func (q *ThumbnailQueue) GetPendingCount() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.stats.PendingCount
}

// Clear clears all pending jobs from the queue.
func (q *ThumbnailQueue) Clear() int {
	q.mu.Lock()
	defer q.mu.Unlock()

	cleared := 0
	for {
		select {
		case <-q.jobs:
			cleared++
		default:
			q.stats.PendingCount = 0
			logging.Info("Cleared pending thumbnail jobs",
				map[string]interface{}{
					"cleared": cleared,
				})
			return cleared
		}
	}
}
