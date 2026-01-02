// Package sync provides cloud synchronization capabilities.
package sync

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/kimhsiao/memonexus/backend/internal/db"
	"github.com/kimhsiao/memonexus/backend/internal/errors"
	"github.com/kimhsiao/memonexus/backend/internal/logging"
	"github.com/kimhsiao/memonexus/backend/internal/models"
)

// SyncDirection represents the direction of sync operation.
type SyncDirection string

const (
	SyncDirectionUpload   SyncDirection = "upload"
	SyncDirectionDownload SyncDirection = "download"
)

// SyncStatus represents the current sync status.
type SyncStatus string

const (
	SyncStatusIdle    SyncStatus = "idle"
	SyncStatusSyncing SyncStatus = "syncing"
	SyncStatusFailed  SyncStatus = "failed"
)

// SyncEventType represents the type of sync event.
type SyncEventType string

const (
	SyncEventStarted      SyncEventType = "started"
	SyncEventProgress     SyncEventType = "progress"
	SyncEventUploadItem   SyncEventType = "upload_item"
	SyncEventDownloadItem SyncEventType = "download_item"
	SyncEventConflict     SyncEventType = "conflict"
	SyncEventCompleted    SyncEventType = "completed"
	SyncEventFailed       SyncEventType = "failed"
	SyncEventWarning      SyncEventType = "warning"
)

// SyncEvent represents a sync event for graceful degradation notifications.
// T175: Non-blocking event notifications for sync failures (FR-057).
type SyncEvent struct {
	Type      SyncEventType
	Timestamp time.Time
	ItemID    string
	Message   string
	Error     error
	Data      map[string]interface{}
}

// SyncEventHandler defines the interface for handling sync events.
// T175: Event handler interface for non-blocking sync notifications.
// Implementations should avoid blocking operations to ensure sync continues.
type SyncEventHandler interface {
	// OnSyncEvent is called when a sync event occurs.
	// IMPORTANT: This callback should be non-blocking and return quickly.
	// For expensive operations, spawn a goroutine or use a queue.
	OnSyncEvent(event SyncEvent)
}

// SyncErrorEntry represents an error in the error history.
type SyncErrorEntry struct {
	Timestamp time.Time
	ItemID    string
	Operation string
	Error     string
}

const (
	// maxErrorHistory is the maximum number of errors to keep in history.
	maxErrorHistory = 100
)

// SyncOperation represents a single sync operation.
type SyncOperation struct {
	ID          string
	Direction   SyncDirection
	ItemID      string
	StatusCode  int
	Error       string
	CreatedAt   time.Time
	CompletedAt *time.Time
}

// SyncEngine provides synchronization capabilities.
type SyncEngine struct {
	repo         db.SyncRepository
	storage      ObjectStore
	status       SyncStatus
	lastSync     *time.Time
	pending      int
	lastErr      error
	eventHandler SyncEventHandler
	errorHistory []SyncErrorEntry
	mu           sync.RWMutex
}

// ObjectStore defines the interface for cloud storage operations.
type ObjectStore interface {
	// Upload uploads data to the store.
	Upload(ctx context.Context, key string, data []byte) error

	// Download downloads data from the store.
	Download(ctx context.Context, key string) ([]byte, error)

	// Delete deletes data from the store.
	Delete(ctx context.Context, key string) error

	// List lists all keys with a prefix.
	List(ctx context.Context, prefix string) ([]string, error)
}

// NewSyncEngine creates a new SyncEngine.
func NewSyncEngine(repo db.SyncRepository, storage ObjectStore) *SyncEngine {
	return &SyncEngine{
		repo:         repo,
		storage:      storage,
		status:       SyncStatusIdle,
		errorHistory: make([]SyncErrorEntry, 0, maxErrorHistory),
	}
}

// SetEventHandler sets the event handler for sync events.
// T175: Non-blocking event handler for graceful degradation.
func (e *SyncEngine) SetEventHandler(handler SyncEventHandler) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.eventHandler = handler
}

// GetErrorHistory returns the error history.
// T175: Error history tracking for graceful degradation.
func (e *SyncEngine) GetErrorHistory() []SyncErrorEntry {
	e.mu.RLock()
	defer e.mu.RUnlock()

	// Return a copy to prevent external modification
	history := make([]SyncErrorEntry, len(e.errorHistory))
	copy(history, e.errorHistory)
	return history
}

// ClearErrorHistory clears the error history.
func (e *SyncEngine) ClearErrorHistory() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.errorHistory = make([]SyncErrorEntry, 0, maxErrorHistory)
}

// emitEvent sends an event to the event handler in a non-blocking way.
// T175: Graceful degradation - events are sent via goroutine to avoid blocking sync.
func (e *SyncEngine) emitEvent(event SyncEvent) {
	e.mu.RLock()
	handler := e.eventHandler
	e.mu.RUnlock()

	if handler == nil {
		return
	}

	// Set timestamp if not already set
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Non-blocking: spawn goroutine to handle event
	// This ensures sync continues even if handler is slow
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logging.ErrorWithCode("Panic in event handler", string(errors.ErrInternal), fmt.Errorf("%v", r),
					map[string]interface{}{"panic": true})
			}
		}()
		handler.OnSyncEvent(event)
	}()
}

// recordError records an error in the error history.
func (e *SyncEngine) recordError(itemID, operation string, err error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	entry := SyncErrorEntry{
		Timestamp: time.Now(),
		ItemID:    itemID,
		Operation: operation,
		Error:     err.Error(),
	}

	e.errorHistory = append(e.errorHistory, entry)

	// Keep only the most recent errors
	if len(e.errorHistory) > maxErrorHistory {
		// Remove oldest entries
		e.errorHistory = e.errorHistory[len(e.errorHistory)-maxErrorHistory:]
	}
}

// Status returns the current sync status.
func (e *SyncEngine) Status() SyncStatus {
	return e.status
}

// LastSync returns the timestamp of the last successful sync.
func (e *SyncEngine) LastSync() *time.Time {
	return e.lastSync
}

// PendingChanges returns the number of pending changes to sync.
func (e *SyncEngine) PendingChanges() int {
	return e.pending
}

// LastError returns the last sync error.
func (e *SyncEngine) LastError() error {
	return e.lastErr
}

// Sync performs a full sync operation.
// It uploads local changes and downloads remote changes.
// T211: Critical operations logging (start/complete/success/failure).
func (e *SyncEngine) Sync(ctx context.Context) (*SyncResult, error) {
	if e.status == SyncStatusSyncing {
		return nil, fmt.Errorf("sync already in progress")
	}

	e.status = SyncStatusSyncing
	e.lastErr = nil

	result := &SyncResult{
		StartTime: time.Now(),
	}

	// Generate sync ID for correlation across all sync-related logs
	syncID := uuid.New().String()

	// Log sync started
	logging.Info("Sync operation started",
		map[string]interface{}{
			"start_time": result.StartTime.Format(time.RFC3339),
			"sync_id":    syncID,
		})

	// Emit sync started event
	e.emitEvent(SyncEvent{
		Type:    SyncEventStarted,
		Message: "Sync operation started",
	})

	defer func() {
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)

		if e.lastErr != nil {
			e.status = SyncStatusFailed
			result.Error = e.lastErr.Error()

			// Log sync failure
			logging.ErrorWithCode("Sync operation failed", string(errors.ErrSyncFailed), e.lastErr,
				map[string]interface{}{
					"sync_id":     syncID,
					"uploaded":    result.Uploaded,
					"downloaded":  result.Downloaded,
					"duration_ms": result.Duration.Milliseconds(),
				})

			// Emit sync failed event
			e.emitEvent(SyncEvent{
				Type:    SyncEventFailed,
				Message: "Sync operation failed",
				Error:   e.lastErr,
				Data: map[string]interface{}{
					"sync_id":    syncID,
					"uploaded":   result.Uploaded,
					"downloaded": result.Downloaded,
					"duration":   result.Duration.String(),
				},
			})
		} else {
			e.status = SyncStatusIdle
			e.lastSync = &result.EndTime
			e.pending = 0

			// Log sync success
			logging.Info("Sync operation completed successfully",
				map[string]interface{}{
					"sync_id":     syncID,
					"uploaded":    result.Uploaded,
					"downloaded":  result.Downloaded,
					"conflicts":   result.Conflicts,
					"duration_ms": result.Duration.Milliseconds(),
				})

			// Emit sync completed event
			e.emitEvent(SyncEvent{
				Type:    SyncEventCompleted,
				Message: fmt.Sprintf("Sync completed: %d uploaded, %d downloaded", result.Uploaded, result.Downloaded),
				Data: map[string]interface{}{
					"sync_id":    syncID,
					"uploaded":   result.Uploaded,
					"downloaded": result.Downloaded,
					"conflicts":  result.Conflicts,
					"duration":   result.Duration.String(),
				},
			})
		}
	}()

	// Step 1: Upload local changes
	uploaded, err := e.uploadChanges(ctx, syncID)
	if err != nil {
		e.lastErr = fmt.Errorf("upload failed: %w", err)
		result.Uploaded = uploaded
		return result, e.lastErr
	}
	result.Uploaded = uploaded

	// Step 2: Download remote changes
	downloaded, err := e.downloadChanges(ctx, syncID)
	if err != nil {
		e.lastErr = fmt.Errorf("download failed: %w", err)
		result.Downloaded = downloaded
		return result, e.lastErr
	}
	result.Downloaded = downloaded

	// Step 3: Resolve conflicts
	conflicts := e.resolveConflicts(ctx)
	result.Conflicts = len(conflicts)

	return result, nil
}

// SyncResult represents the result of a sync operation.
type SyncResult struct {
	StartTime  time.Time
	EndTime    time.Time
	Duration   time.Duration
	Uploaded   int
	Downloaded int
	Conflicts  int
	Error      string
}

// uploadChanges uploads local changes to the remote store.
func (e *SyncEngine) uploadChanges(ctx context.Context, syncID string) (int, error) {
	// Get items modified since last sync
	// For simplicity, we're syncing all non-deleted items
	items, err := e.repo.ListContentItems(1000, 0, "")
	if err != nil {
		return 0, err
	}

	uploaded := 0
	warnings := 0

	for _, item := range items {
		select {
		case <-ctx.Done():
			return uploaded, ctx.Err()
		default:
		}

		// Serialize item
		data := e.serializeItem(item)

		// Upload to storage
		key := fmt.Sprintf("items/%s.json", item.ID)
		if err := e.storage.Upload(ctx, key, data); err != nil {
			// T175: Graceful degradation - record error but continue
			e.recordError(string(item.ID), "upload", err)
			logging.Warn("Failed to upload item",
				map[string]interface{}{
					"sync_id": syncID,
					"item_id": item.ID,
					"error":   err.Error(),
				})

			// Emit warning event for non-blocking error
			e.emitEvent(SyncEvent{
				Type:    SyncEventWarning,
				ItemID:  string(item.ID),
				Message: fmt.Sprintf("Failed to upload item %s", item.ID),
				Error:   err,
				Data:    map[string]interface{}{"sync_id": syncID},
			})
			warnings++
			continue
		}

		uploaded++

		// Emit upload item event
		e.emitEvent(SyncEvent{
			Type:    SyncEventUploadItem,
			ItemID:  string(item.ID),
			Message: fmt.Sprintf("Uploaded item %s", item.ID),
			Data: map[string]interface{}{
				"sync_id": syncID,
				"version": item.Version,
			},
		})

		// Create change log
		changeLog := &models.ChangeLog{
			ItemID:    item.ID,
			Operation: "update",
			Version:   item.Version,
		}
		if err := e.repo.CreateChangeLog(changeLog); err != nil {
			logging.ErrorWithCode("Failed to create change log", string(errors.ErrDatabase), err,
				map[string]interface{}{
					"sync_id": syncID,
					"item_id": item.ID,
				})
		}
	}

	// Emit progress event
	e.emitEvent(SyncEvent{
		Type:    SyncEventProgress,
		Message: fmt.Sprintf("Upload phase completed: %d uploaded, %d warnings", uploaded, warnings),
		Data: map[string]interface{}{
			"sync_id":  syncID,
			"uploaded": uploaded,
			"warnings": warnings,
		},
	})

	return uploaded, nil
}

// downloadChanges downloads remote changes from the store.
func (e *SyncEngine) downloadChanges(ctx context.Context, syncID string) (int, error) {
	// List all items in storage
	keys, err := e.storage.List(ctx, "items/")
	if err != nil {
		return 0, err
	}

	downloaded := 0
	warnings := 0

	for _, key := range keys {
		select {
		case <-ctx.Done():
			return downloaded, ctx.Err()
		default:
		}

		// Download item data
		data, err := e.storage.Download(ctx, key)
		if err != nil {
			// T175: Graceful degradation - record error but continue
			e.recordError(key, "download", err)
			logging.Warn("Failed to download",
				map[string]interface{}{
					"sync_id": syncID,
					"key":     key,
					"error":   err.Error(),
				})

			// Emit warning event
			e.emitEvent(SyncEvent{
				Type:    SyncEventWarning,
				Message: fmt.Sprintf("Failed to download %s", key),
				Error:   err,
				Data:    map[string]interface{}{"sync_id": syncID},
			})
			warnings++
			continue
		}

		// Deserialize item
		item, err := e.deserializeItem(data)
		if err != nil {
			e.recordError(key, "deserialize", err)
			logging.Warn("Failed to deserialize",
				map[string]interface{}{
					"sync_id": syncID,
					"key":     key,
					"error":   err.Error(),
				})
			warnings++
			continue
		}

		// Check if local exists
		localItem, err := e.repo.GetContentItem(string(item.ID))
		if err == sql.ErrNoRows {
			// Item doesn't exist locally, create it
			if err := e.repo.CreateContentItem(item); err != nil {
				e.recordError(string(item.ID), "create", err)
				logging.ErrorWithCode("Failed to create item", string(errors.ErrDatabase), err,
					map[string]interface{}{
						"sync_id": syncID,
						"item_id": item.ID,
					})
				warnings++
				continue
			}
			downloaded++

			// Emit download item event
			e.emitEvent(SyncEvent{
				Type:    SyncEventDownloadItem,
				ItemID:  string(item.ID),
				Message: fmt.Sprintf("Downloaded new item %s", item.ID),
				Data: map[string]interface{}{
					"sync_id": syncID,
					"version": item.Version,
				},
			})
		} else if err != nil {
			e.recordError(string(item.ID), "fetch_local", err)
			logging.ErrorWithCode("Failed to get local item", string(errors.ErrDatabase), err,
				map[string]interface{}{
					"sync_id": syncID,
					"item_id": item.ID,
				})
			warnings++
			continue
		} else {
			// Item exists, check for conflict
			if localItem.Version < item.Version {
				// Remote is newer, update local
				if err := e.repo.UpdateContentItem(item); err != nil {
					e.recordError(string(item.ID), "update", err)
					logging.ErrorWithCode("Failed to update item", string(errors.ErrDatabase), err,
						map[string]interface{}{
							"sync_id": syncID,
							"item_id": item.ID,
						})
					warnings++
					continue
				}
				downloaded++

				// Emit download item event
				e.emitEvent(SyncEvent{
					Type:    SyncEventDownloadItem,
					ItemID:  string(item.ID),
					Message: fmt.Sprintf("Updated item %s to version %d", item.ID, item.Version),
					Data: map[string]interface{}{
						"sync_id": syncID,
						"version": item.Version,
					},
				})
			} else if localItem.Version > item.Version {
				// Local is newer, log conflict (will be resolved in next upload)
				conflictLog := &models.ConflictLog{
					ItemID:          item.ID,
					LocalTimestamp:  localItem.UpdatedAt,
					RemoteTimestamp: item.UpdatedAt,
					Resolution:      "last_write_wins",
				}
				if err := e.repo.CreateConflictLog(conflictLog); err != nil {
					logging.ErrorWithCode("Failed to create conflict log", string(errors.ErrDatabase), err,
						map[string]interface{}{
							"sync_id": syncID,
							"item_id": item.ID,
						})
				}

				// Emit conflict event
				e.emitEvent(SyncEvent{
					Type:    SyncEventConflict,
					ItemID:  string(item.ID),
					Message: fmt.Sprintf("Conflict detected for item %s", item.ID),
					Data: map[string]interface{}{
						"sync_id":        syncID,
						"local_version":  localItem.Version,
						"remote_version": item.Version,
						"resolution":     "last_write_wins",
					},
				})
			}
		}
	}

	// Emit progress event
	e.emitEvent(SyncEvent{
		Type:    SyncEventProgress,
		Message: fmt.Sprintf("Download phase completed: %d downloaded, %d warnings", downloaded, warnings),
		Data: map[string]interface{}{
			"sync_id":    syncID,
			"downloaded": downloaded,
			"warnings":   warnings,
		},
	})

	return downloaded, nil
}

// resolveConflicts resolves any detected conflicts.
// For now, we use "last write wins" strategy.
func (e *SyncEngine) resolveConflicts(ctx context.Context) []*models.ConflictLog {
	// Get all unresolved conflicts
	// For simplicity, this is a placeholder
	// In production, you'd query the conflict_log table
	return nil
}

// serializeItem serializes a content item to JSON.
func (e *SyncEngine) serializeItem(item *models.ContentItem) []byte {
	data, err := json.Marshal(item)
	if err != nil {
		// Fallback to empty JSON on error
		return []byte("{}")
	}
	return data
}

// deserializeItem deserializes a content item from JSON.
func (e *SyncEngine) deserializeItem(data []byte) (*models.ContentItem, error) {
	var item models.ContentItem
	if err := json.Unmarshal(data, &item); err != nil {
		return nil, fmt.Errorf("failed to deserialize item: %w", err)
	}
	return &item, nil
}

// StartPeriodicSync starts periodic background sync.
func (e *SyncEngine) StartPeriodicSync(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if e.status == SyncStatusIdle {
				go func() {
					syncCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
					defer cancel()
					_, err := e.Sync(syncCtx)
					if err != nil {
						logging.ErrorWithCode("Periodic sync failed", string(errors.ErrSyncFailed), err,
							map[string]interface{}{"interval_minutes": interval.Minutes()})
					}
				}()
			}
		}
	}
}
