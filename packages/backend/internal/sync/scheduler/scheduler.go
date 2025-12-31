// Package scheduler provides background sync scheduling for offline operations.
// T174: Background sync scheduler with periodic sync and queue processing.
package scheduler

import (
	"context"
	"log"
	"sync"
	"time"

	syncpkg "github.com/kimhsiao/memonexus/backend/internal/sync"
	"github.com/kimhsiao/memonexus/backend/internal/sync/queue"
)

// Scheduler manages background sync operations.
type Scheduler struct {
	engine          *syncpkg.SyncEngine
	queue           *queue.SyncQueue
	syncInterval    time.Duration
	queueInterval   time.Duration
	stopCh          chan struct{}
	wg              sync.WaitGroup
	mu              sync.RWMutex
	isRunning       bool
	isOnline        bool
	lastSyncTime    time.Time
	syncInProgress  bool
	queueInProgress bool
}

// SchedulerConfig holds scheduler configuration.
type SchedulerConfig struct {
	SyncInterval  time.Duration // How often to sync when online (default: 15 minutes)
	QueueInterval time.Duration // How often to process queue when offline (default: 1 minute)
}

// DefaultSchedulerConfig returns default scheduler configuration.
func DefaultSchedulerConfig() *SchedulerConfig {
	return &SchedulerConfig{
		SyncInterval:  15 * time.Minute,
		QueueInterval: 1 * time.Minute,
	}
}

// NewScheduler creates a new Scheduler.
// T174: Background sync scheduler with periodic sync when online, queue processing when offline.
func NewScheduler(engine *syncpkg.SyncEngine, queue *queue.SyncQueue, config *SchedulerConfig) *Scheduler {
	if config == nil {
		config = DefaultSchedulerConfig()
	}

	return &Scheduler{
		engine:         engine,
		queue:          queue,
		syncInterval:   config.SyncInterval,
		queueInterval:  config.QueueInterval,
		stopCh:         make(chan struct{}),
		isOnline:       true, // Assume online initially
	}
}

// Start starts the background sync scheduler.
func (s *Scheduler) Start(ctx context.Context) {
	s.mu.Lock()
	if s.isRunning {
		s.mu.Unlock()
		return
	}
	s.isRunning = true
	s.mu.Unlock()

	s.wg.Add(2)

	// Start periodic sync goroutine
	go s.periodicSyncLoop(ctx)

	// Start queue processor goroutine
	go s.queueProcessorLoop(ctx)

	log.Println("[Scheduler] Background sync scheduler started")
}

// Stop stops the background sync scheduler gracefully.
func (s *Scheduler) Stop() {
	s.mu.Lock()
	if !s.isRunning {
		s.mu.Unlock()
		return
	}
	s.isRunning = false
	s.mu.Unlock()

	// Signal stop to all goroutines
	close(s.stopCh)

	// Wait for goroutines to finish
	s.wg.Wait()

	log.Println("[Scheduler] Background sync scheduler stopped")
}

// SetOnlineStatus changes the online status of the scheduler.
// When offline, only queue processing runs (no sync attempts).
// When online, both periodic sync and queue processing run.
func (s *Scheduler) SetOnlineStatus(isOnline bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	wasOnline := s.isOnline
	s.isOnline = isOnline

	if wasOnline != isOnline {
		log.Printf("[Scheduler] Online status changed: %v -> %v", wasOnline, isOnline)
	}
}

// periodicSyncLoop runs periodic sync when online.
func (s *Scheduler) periodicSyncLoop(ctx context.Context) {
	defer s.wg.Done()

	// Create ticker for periodic sync
	ticker := time.NewTicker(s.syncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopCh:
			return
		case <-ticker.C:
			if !s.isOnline {
				continue
			}

			// Check if already syncing
			s.mu.RLock()
			isSyncing := s.syncInProgress
			s.mu.RUnlock()

			if isSyncing {
				log.Println("[Scheduler] Sync already in progress, skipping")
				continue
			}

			// Start sync
			go s.runSync(ctx)
		}
	}
}

// queueProcessorLoop processes the sync queue when offline or online.
func (s *Scheduler) queueProcessorLoop(ctx context.Context) {
	defer s.wg.Done()

	// Create ticker for queue processing
	ticker := time.NewTicker(s.queueInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopCh:
			return
		case <-ticker.C:
			// Process queue regardless of online status
			// (queue items are processed when network becomes available)
			go s.processQueue(ctx)
		}
	}
}

// runSync executes a sync operation.
func (s *Scheduler) runSync(ctx context.Context) {
	s.mu.Lock()
	s.syncInProgress = true
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		s.syncInProgress = false
		s.mu.Unlock()
	}()

	log.Println("[Scheduler] Starting periodic sync...")

	// Create sync context with timeout
	syncCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	result, err := s.engine.Sync(syncCtx)

	if err != nil {
		log.Printf("[Scheduler] Periodic sync failed: %v", err)
		return
	}

	s.mu.Lock()
	s.lastSyncTime = time.Now()
	s.mu.Unlock()

	log.Printf("[Scheduler] Periodic sync completed: uploaded=%d, downloaded=%d, conflicts=%d",
		result.Uploaded, result.Downloaded, result.Conflicts)
}

// processQueue processes pending items in the sync queue.
func (s *Scheduler) processQueue(ctx context.Context) {
	// Get pending items from queue
	pending := s.queue.GetPending()

	if len(pending) == 0 {
		return
	}

	s.mu.Lock()
	if s.queueInProgress {
		s.mu.Unlock()
		return
	}
	s.queueInProgress = true
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		s.queueInProgress = false
		s.mu.Unlock()
	}()

	log.Printf("[Scheduler] Processing %d pending queue items...", len(pending))

	processed := 0
	for _, item := range pending {
		select {
		case <-ctx.Done():
			return
		case <-s.stopCh:
			return
		default:
			// Process queue item
			// TODO: Implement actual item processing based on operation type
			// For now, just mark as complete (simulation)
			if err := s.queue.Complete(item.ID); err == nil {
				processed++
			} else {
				log.Printf("[Scheduler] Failed to complete item %s: %v", item.ID, err)
			}
		}
	}

	log.Printf("[Scheduler] Queue processing completed: %d items processed", processed)
}

// TriggerSync triggers an immediate sync operation.
// Returns true if sync was started, false if sync is already in progress.
func (s *Scheduler) TriggerSync(ctx context.Context) bool {
	s.mu.RLock()
	isSyncing := s.syncInProgress
	s.mu.RUnlock()

	if isSyncing {
		return false
	}

	go s.runSync(ctx)
	return true
}

// GetStatus returns the current status of the scheduler.
type SchedulerStatus struct {
	IsRunning      bool
	IsOnline       bool
	LastSyncTime   *time.Time
	SyncInProgress bool
	QueueInProgress bool
	PendingItems   int
	QueueStats     map[string]int
}

func (s *Scheduler) GetStatus() SchedulerStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	status := SchedulerStatus{
		IsRunning:       s.isRunning,
		IsOnline:        s.isOnline,
		SyncInProgress:  s.syncInProgress,
		QueueInProgress: s.queueInProgress,
	}

	if !s.lastSyncTime.IsZero() {
		status.LastSyncTime = &s.lastSyncTime
	}

	status.PendingItems = len(s.queue.GetPending())
	status.QueueStats = s.queue.GetStats()

	return status
}

// SyncNow triggers an immediate sync and waits for completion.
// T174: Method to trigger immediate sync from API.
func (s *Scheduler) SyncNow(ctx context.Context) error {
	s.mu.Lock()
	s.syncInProgress = true
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		s.syncInProgress = false
		s.mu.Unlock()
	}()

	// Create sync context with timeout
	syncCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	result, err := s.engine.Sync(syncCtx)

	if err != nil {
		return err
	}

	s.mu.Lock()
	s.lastSyncTime = time.Now()
	s.mu.Unlock()

	log.Printf("[Scheduler] Manual sync completed: uploaded=%d, downloaded=%d, conflicts=%d",
		result.Uploaded, result.Downloaded, result.Conflicts)

	return nil
}

// IsOnline returns whether the scheduler is in online mode.
func (s *Scheduler) IsOnline() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isOnline
}

// IsRunning returns whether the scheduler is running.
func (s *Scheduler) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isRunning
}
