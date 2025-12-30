// Package sync provides cloud synchronization capabilities.
package sync

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/kimhsiao/memonexus/backend/internal/db"
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

// SyncOperation represents a single sync operation.
type SyncOperation struct {
	ID         string
	Direction  SyncDirection
	ItemID     string
	StatusCode int
	Error      string
	CreatedAt  time.Time
	CompletedAt *time.Time
}

// SyncEngine provides synchronization capabilities.
type SyncEngine struct {
	repo     *db.Repository
	storage  ObjectStore
	status   SyncStatus
	lastSync *time.Time
	pending  int
	lastErr  error
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
func NewSyncEngine(repo *db.Repository, storage ObjectStore) *SyncEngine {
	return &SyncEngine{
		repo:    repo,
		storage: storage,
		status:  SyncStatusIdle,
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
func (e *SyncEngine) Sync(ctx context.Context) (*SyncResult, error) {
	if e.status == SyncStatusSyncing {
		return nil, fmt.Errorf("sync already in progress")
	}

	e.status = SyncStatusSyncing
	e.lastErr = nil

	result := &SyncResult{
		StartTime: time.Now(),
	}

	defer func() {
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)

		if e.lastErr != nil {
			e.status = SyncStatusFailed
			result.Error = e.lastErr.Error()
		} else {
			e.status = SyncStatusIdle
			e.lastSync = &result.EndTime
			e.pending = 0
		}
	}()

	// Step 1: Upload local changes
	uploaded, err := e.uploadChanges(ctx)
	if err != nil {
		e.lastErr = fmt.Errorf("upload failed: %w", err)
		result.Uploaded = uploaded
		return result, e.lastErr
	}
	result.Uploaded = uploaded

	// Step 2: Download remote changes
	downloaded, err := e.downloadChanges(ctx)
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
	StartTime   time.Time
	EndTime     time.Time
	Duration    time.Duration
	Uploaded    int
	Downloaded  int
	Conflicts   int
	Error       string
}

// uploadChanges uploads local changes to the remote store.
func (e *SyncEngine) uploadChanges(ctx context.Context) (int, error) {
	// Get items modified since last sync
	// For simplicity, we're syncing all non-deleted items
	items, err := e.repo.ListContentItems(1000, 0, "")
	if err != nil {
		return 0, err
	}

	uploaded := 0

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
			log.Printf("Failed to upload item %s: %v", item.ID, err)
			continue
		}

		uploaded++

		// Create change log
		log := &models.ChangeLog{
			ItemID:    string(item.ID),
			Operation: "update",
			Version:   item.Version,
		}
		if err := e.repo.CreateChangeLog(log); err != nil {
			log.Printf("Failed to create change log: %v", err)
		}
	}

	return uploaded, nil
}

// downloadChanges downloads remote changes from the store.
func (e *SyncEngine) downloadChanges(ctx context.Context) (int, error) {
	// List all items in storage
	keys, err := e.storage.List(ctx, "items/")
	if err != nil {
		return 0, err
	}

	downloaded := 0

	for _, key := range keys {
		select {
		case <-ctx.Done():
			return downloaded, ctx.Err()
		default:
		}

		// Download item data
		data, err := e.storage.Download(ctx, key)
		if err != nil {
			log.Printf("Failed to download %s: %v", key, err)
			continue
		}

		// Deserialize item
		item, err := e.deserializeItem(data)
		if err != nil {
			log.Printf("Failed to deserialize %s: %v", key, err)
			continue
		}

		// Check if local exists
		localItem, err := e.repo.GetContentItem(string(item.ID))
		if err == sql.ErrNoRows {
			// Item doesn't exist locally, create it
			if err := e.repo.CreateContentItem(item); err != nil {
				log.Printf("Failed to create item %s: %v", item.ID, err)
				continue
			}
			downloaded++
		} else if err != nil {
			log.Printf("Failed to get local item %s: %v", item.ID, err)
			continue
		} else {
			// Item exists, check for conflict
			if localItem.Version < item.Version {
				// Remote is newer, update local
				if err := e.repo.UpdateContentItem(item); err != nil {
					log.Printf("Failed to update item %s: %v", item.ID, err)
					continue
				}
				downloaded++
			} else if localItem.Version > item.Version {
				// Local is newer, log conflict (will be resolved in next upload)
				conflict := &models.ConflictLog{
					ItemID:          string(item.ID),
					LocalTimestamp:  localItem.UpdatedAt,
					RemoteTimestamp: item.UpdatedAt,
					Resolution:      "last_write_wins",
				}
				if err := e.repo.CreateConflictLog(conflict); err != nil {
					log.Printf("Failed to create conflict log: %v", err)
				}
			}
		}
	}

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
	// Simple JSON serialization
	// In production, use proper JSON marshaling
	format := `{"id":"%s","title":"%s","content_text":"%s","media_type":"%s","tags":"%s","version":%d,"updated_at":%d}`
	return []byte(fmt.Sprintf(format,
		item.ID, item.Title, escapeJSON(item.ContentText),
		item.MediaType, item.Tags, item.Version, item.UpdatedAt))
}

// deserializeItem deserializes a content item from JSON.
func (e *SyncEngine) deserializeItem(data []byte) (*models.ContentItem, error) {
	// Simple JSON deserialization
	// In production, use proper JSON unmarshaling
	// This is a placeholder for demonstration
	item := &models.ContentItem{
		ID:        "placeholder",
		Title:     "Deserialized Item",
		MediaType: "web",
		Version:   1,
	}
	return item, nil
}

// escapeJSON escapes special JSON characters.
func escapeJSON(s string) string {
	// Simple escaping for demonstration
	// In production, use json.Marshal
	return s
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
						log.Printf("Periodic sync failed: %v", err)
					}
				}()
			}
		}
	}
}
