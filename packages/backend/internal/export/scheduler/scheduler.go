// Package scheduler provides automatic export scheduling functionality.
package scheduler

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/kimhsiao/memonexus/backend/internal/export"
)

// ExportInterval defines the scheduling frequency.
type ExportInterval string

const (
	IntervalManual ExportInterval = "manual"
	IntervalDaily  ExportInterval = "daily"
	IntervalWeekly ExportInterval = "weekly"
	IntervalMonthly ExportInterval = "monthly"
)

// SchedulerConfig holds the scheduler configuration.
type SchedulerConfig struct {
	Interval       ExportInterval // How often to export
	RetentionCount int             // Number of archives to keep (0 = unlimited)
	IncludeMedia   bool            // Whether to include media files in exports
	ExportDir      string          // Directory to store exports (default: "exports")
	Password       string          // Password for encryption (empty = no encryption)
}

// Scheduler manages automatic export scheduling.
type Scheduler struct {
	service *export.ExportService
	config  *SchedulerConfig
	ticker  *time.Ticker
	stopCh  chan struct{}
	logger  *slog.Logger
}

// NewScheduler creates a new export scheduler.
func NewScheduler(service *export.ExportService, config *SchedulerConfig) *Scheduler {
	if config.ExportDir == "" {
		config.ExportDir = "exports"
	}
	if config.RetentionCount < 0 {
		config.RetentionCount = 0
	}

	return &Scheduler{
		service: service,
		config:  config,
		stopCh:  make(chan struct{}),
		logger:  slog.Default(),
	}
}

// Start begins the automatic export scheduler.
// It will perform an initial export if configured interval is not manual.
func (s *Scheduler) Start(ctx context.Context) error {
	if s.config.Interval == IntervalManual {
		s.logger.Info("scheduler in manual mode, automatic exports disabled")
		return nil
	}

	// Calculate ticker duration
	dur, err := s.intervalDuration()
	if err != nil {
		return fmt.Errorf("invalid interval: %w", err)
	}

	s.ticker = time.NewTicker(dur)
	s.logger.Info("scheduler started",
		"interval", s.config.Interval,
		"retention_count", s.config.RetentionCount,
		"include_media", s.config.IncludeMedia)

	// Perform initial export
	go func() {
		if err := s.runExport(ctx); err != nil {
			s.logger.Error("initial export failed", "error", err)
		}
	}()

	// Start periodic exports
	go func() {
		for {
			select {
			case <-s.ticker.C:
				if err := s.runExport(ctx); err != nil {
					s.logger.Error("scheduled export failed", "error", err)
				}
			case <-s.stopCh:
				s.logger.Info("scheduler stopped")
				return
			case <-ctx.Done():
				s.logger.Info("scheduler context cancelled")
				return
			}
		}
	}()

	return nil
}

// Stop gracefully shuts down the scheduler.
func (s *Scheduler) Stop() {
	close(s.stopCh)
	if s.ticker != nil {
		s.ticker.Stop()
	}
}

// runExport performs a single export operation with retention management.
func (s *Scheduler) runExport(ctx context.Context) error {
	s.logger.Info("starting scheduled export")

	// Generate output path with timestamp
	timestamp := time.Now().Format("20060102_150405")
	outputPath := filepath.Join(s.config.ExportDir, fmt.Sprintf("memonexus_%s.tar.gz", timestamp))

	// Create export config
	config := &export.ExportConfig{
		OutputPath:   outputPath,
		Password:     s.config.Password,
		IncludeMedia: s.config.IncludeMedia,
	}

	// Perform export
	result, err := s.service.Export(config)
	if err != nil {
		return fmt.Errorf("export failed: %w", err)
	}

	s.logger.Info("export completed",
		"file", result.FilePath,
		"size_bytes", result.SizeBytes,
		"item_count", result.ItemCount,
		"duration", result.Duration)

	// Apply retention policy
	if s.config.RetentionCount > 0 {
		if err := s.applyRetentionPolicy(ctx); err != nil {
			s.logger.Error("retention policy failed", "error", err)
			// Don't fail the export if retention fails
		}
	}

	return nil
}

// intervalDuration converts the interval to a time.Duration.
func (s *Scheduler) intervalDuration() (time.Duration, error) {
	switch s.config.Interval {
	case IntervalDaily:
		return 24 * time.Hour, nil
	case IntervalWeekly:
		return 7 * 24 * time.Hour, nil
	case IntervalMonthly:
		// Approximate as 30 days
		return 30 * 24 * time.Hour, nil
	case IntervalManual:
		return 0, fmt.Errorf("manual interval has no duration")
	default:
		return 0, fmt.Errorf("unknown interval: %s", s.config.Interval)
	}
}

// applyRetentionPolicy removes old exports according to retention count.
func (s *Scheduler) applyRetentionPolicy(ctx context.Context) error {
	// List all archives in the export directory
	archives, err := listArchives(s.config.ExportDir)
	if err != nil {
		return fmt.Errorf("failed to list archives: %w", err)
	}

	// Sort by creation time (oldest first)
	sort.Slice(archives, func(i, j int) bool {
		return archives[i].CreatedAt.Before(archives[j].CreatedAt)
	})

	// Delete excess archives
	if len(archives) > s.config.RetentionCount {
		toDelete := archives[:len(archives)-s.config.RetentionCount]
		for _, archive := range toDelete {
			if err := os.Remove(archive.Path); err != nil {
				s.logger.Error("failed to delete old archive",
					"path", archive.Path,
					"error", err)
				continue
			}
			s.logger.Info("deleted old archive", "path", archive.Path)
		}
	}

	return nil
}

// ArchiveInfo represents metadata about an export archive.
type ArchiveInfo struct {
	Path       string
	SizeBytes  int64
	CreatedAt  time.Time
	Checksum   string
	ItemCount  int
	Encrypted  bool
}

// listArchives returns all archives in the export directory.
func listArchives(exportDir string) ([]*ArchiveInfo, error) {
	var archives []*ArchiveInfo

	// Ensure directory exists
	if _, err := os.Stat(exportDir); os.IsNotExist(err) {
		return archives, nil
	}

	// Walk directory
	err := filepath.Walk(exportDir, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-tar.gz files
		if fi.IsDir() || filepath.Ext(path) != ".gz" {
			return nil
		}

		// Create archive info
		info := &ArchiveInfo{
			Path:      path,
			SizeBytes: fi.Size(),
			CreatedAt: fi.ModTime(),
		}

		archives = append(archives, info)
		return nil
	})

	return archives, err
}

// UpdateConfig updates the scheduler configuration.
func (s *Scheduler) UpdateConfig(config *SchedulerConfig) error {
	s.config = config

	// Restart scheduler if running
	if s.ticker != nil {
		s.Stop()
		s.stopCh = make(chan struct{})
		// Note: caller needs to call Start again with new context
	}

	return nil
}

// GetConfig returns the current scheduler configuration.
func (s *Scheduler) GetConfig() *SchedulerConfig {
	return s.config
}
