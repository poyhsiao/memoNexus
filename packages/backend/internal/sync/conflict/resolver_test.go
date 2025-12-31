// Package conflict provides unit tests for conflict resolution.
// T149: Unit test for conflict resolution (last write wins).
package conflict

import (
	"testing"
	"time"

	"github.com/kimhsiao/memonexus/backend/internal/models"
)

// TestResolverLastWriteWins tests the last write wins resolution strategy.
func TestResolverLastWriteWins(t *testing.T) {
	resolver := NewResolver(ResolutionStrategyLastWriteWins)

	now := time.Now().Unix()

	// Create test items with different timestamps
	localItem := &models.ContentItem{
		ID:        "item-1",
		Title:     "Local Title",
		UpdatedAt: now + 100, // Local is newer
		Version:   2,
	}

	remoteItem := &models.ContentItem{
		ID:        "item-1",
		Title:     "Remote Title",
		UpdatedAt: now, // Remote is older
		Version:   1,
	}

	conflict := &Conflict{
		ItemID:          "item-1",
		LocalItem:       localItem,
		RemoteItem:      remoteItem,
		LocalTimestamp:  localItem.UpdatedAt,
		RemoteTimestamp: remoteItem.UpdatedAt,
		DetectedAt:      now,
	}

	result, err := resolver.Resolve(conflict)

	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}

	// Local should win (newer timestamp)
	if result.WinningItem.ID != localItem.ID {
		t.Errorf("Expected local item to win, got %s", result.WinningItem.ID)
	}

	if result.Strategy != ResolutionStrategyLastWriteWins {
		t.Errorf("Expected LastWriteWins strategy, got %s", result.Strategy)
	}

	if result.ConflictLog == nil {
		t.Error("Expected conflict log to be created")
	}

	// Verify conflict log resolution
	if result.ConflictLog.Resolution != "local_wins" {
		t.Errorf("Expected 'local_wins' resolution, got %s", result.ConflictLog.Resolution)
	}
}

// TestResolverLastWriteWinsRemoteNewer tests when remote item is newer.
func TestResolverLastWriteWinsRemoteNewer(t *testing.T) {
	resolver := NewResolver(ResolutionStrategyLastWriteWins)

	now := time.Now().Unix()

	localItem := &models.ContentItem{
		ID:        "item-1",
		Title:     "Local Title",
		UpdatedAt: now, // Local is older
		Version:   1,
	}

	remoteItem := &models.ContentItem{
		ID:        "item-1",
		Title:     "Remote Title",
		UpdatedAt: now + 100, // Remote is newer
		Version:   2,
	}

	conflict := &Conflict{
		ItemID:          "item-1",
		LocalItem:       localItem,
		RemoteItem:      remoteItem,
		LocalTimestamp:  localItem.UpdatedAt,
		RemoteTimestamp: remoteItem.UpdatedAt,
	}

	result, err := resolver.Resolve(conflict)

	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}

	// Remote should win (newer timestamp)
	if result.WinningItem.ID != remoteItem.ID {
		t.Errorf("Expected remote item to win, got %s", result.WinningItem.ID)
	}

	if result.ConflictLog.Resolution != "remote_wins" {
		t.Errorf("Expected 'remote_wins' resolution, got %s", result.ConflictLog.Resolution)
	}
}

// TestResolverSameTimestamp tests when timestamps are equal.
func TestResolverSameTimestamp(t *testing.T) {
	resolver := NewResolver(ResolutionStrategyLastWriteWins)

	now := time.Now().Unix()

	localItem := &models.ContentItem{
		ID:        "item-1",
		Title:     "Local Title",
		UpdatedAt: now,
		Version:   1,
	}

	remoteItem := &models.ContentItem{
		ID:        "item-1",
		Title:     "Remote Title",
		UpdatedAt: now, // Same timestamp
		Version:   1,
	}

	conflict := &Conflict{
		ItemID:          "item-1",
		LocalItem:       localItem,
		RemoteItem:      remoteItem,
		LocalTimestamp:  localItem.UpdatedAt,
		RemoteTimestamp: remoteItem.UpdatedAt,
	}

	result, err := resolver.Resolve(conflict)

	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}

	// Local should win when timestamps are equal (>= comparison)
	if result.WinningItem.ID != localItem.ID {
		t.Errorf("Expected local item to win on equal timestamps, got %s", result.WinningItem.ID)
	}
}

// TestResolverInvalidConflict tests error handling for invalid conflicts.
func TestResolverInvalidConflict(t *testing.T) {
	resolver := NewResolver(ResolutionStrategyLastWriteWins)

	tests := []struct {
		name     string
		conflict *Conflict
		wantErr  error
	}{
		{
			name:     "nil local item",
			conflict: &Conflict{LocalItem: nil, RemoteItem: &models.ContentItem{ID: "item-1"}},
			wantErr:  ErrInvalidConflict,
		},
		{
			name:     "nil remote item",
			conflict: &Conflict{LocalItem: &models.ContentItem{ID: "item-1"}, RemoteItem: nil},
			wantErr:  ErrInvalidConflict,
		},
		{
			name: "ID mismatch",
			conflict: &Conflict{
				LocalItem:  &models.ContentItem{ID: "item-1"},
				RemoteItem: &models.ContentItem{ID: "item-2"},
			},
			wantErr: ErrItemIDMismatch,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := resolver.Resolve(tt.conflict)
			if err != tt.wantErr {
				t.Errorf("Expected %v, got %v", tt.wantErr, err)
			}
		})
	}
}

// TestDetectConflict tests conflict detection.
func TestDetectConflict(t *testing.T) {
	resolver := NewResolver(ResolutionStrategyLastWriteWins)

	now := time.Now().Unix()

	t.Run("conflict detected", func(t *testing.T) {
		localItem := &models.ContentItem{
			ID:        "item-1",
			Title:     "Local Title",
			UpdatedAt: now,
			Version:   2,
		}

		remoteItem := &models.ContentItem{
			ID:        "item-1",
			Title:     "Remote Title",
			UpdatedAt: now + 100,
			Version:   1,
		}

		conflict, detected := resolver.DetectConflict(localItem, remoteItem)

		if !detected {
			t.Error("Expected conflict to be detected")
		}

		if conflict == nil {
			t.Fatal("Expected conflict object, got nil")
		}

		if conflict.ItemID != "item-1" {
			t.Errorf("Expected item-1, got %s", conflict.ItemID)
		}
	})

	t.Run("no conflict - nil local", func(t *testing.T) {
		_, detected := resolver.DetectConflict(nil, &models.ContentItem{ID: "item-1"})
		if detected {
			t.Error("Expected no conflict with nil local item")
		}
	})

	t.Run("no conflict - nil remote", func(t *testing.T) {
		_, detected := resolver.DetectConflict(&models.ContentItem{ID: "item-1"}, nil)
		if detected {
			t.Error("Expected no conflict with nil remote item")
		}
	})

	t.Run("no conflict - ID mismatch", func(t *testing.T) {
		localItem := &models.ContentItem{ID: "item-1", Version: 1}
		remoteItem := &models.ContentItem{ID: "item-2", Version: 1}

		_, detected := resolver.DetectConflict(localItem, remoteItem)
		if detected {
			t.Error("Expected no conflict with different IDs")
		}
	})

	t.Run("no conflict - same version", func(t *testing.T) {
		localItem := &models.ContentItem{
			ID:        "item-1",
			UpdatedAt: now,
			Version:   1,
		}

		remoteItem := &models.ContentItem{
			ID:        "item-1",
			UpdatedAt: now,
			Version:   1,
		}

		_, detected := resolver.DetectConflict(localItem, remoteItem)
		if detected {
			t.Error("Expected no conflict when versions are same")
		}
	})
}

// TestResolveMultiple tests batch conflict resolution.
func TestResolveMultiple(t *testing.T) {
	resolver := NewResolver(ResolutionStrategyLastWriteWins)

	now := time.Now().Unix()

	conflicts := []*Conflict{
		{
			ItemID:     "item-1",
			LocalItem:  &models.ContentItem{ID: "item-1", UpdatedAt: now + 100},
			RemoteItem: &models.ContentItem{ID: "item-1", UpdatedAt: now},
		},
		{
			ItemID:     "item-2",
			LocalItem:  &models.ContentItem{ID: "item-2", UpdatedAt: now},
			RemoteItem: &models.ContentItem{ID: "item-2", UpdatedAt: now + 100},
		},
	}

	results, err := resolver.ResolveMultiple(conflicts)

	if err != nil {
		t.Fatalf("ResolveMultiple failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	// Verify first conflict (local wins)
	if results[0].WinningItem.ID != "item-1" {
		t.Errorf("Expected item-1 to win first conflict, got %s", results[0].WinningItem.ID)
	}

	// Verify second conflict (remote wins)
	if results[1].WinningItem.ID != "item-2" {
		t.Errorf("Expected item-2 to win second conflict, got %s", results[1].WinningItem.ID)
	}
}

// TestShouldAutoResolve tests the auto-resolve decision logic.
func TestShouldAutoResolve(t *testing.T) {
	resolver := NewResolver(ResolutionStrategyLastWriteWins)

	now := time.Now().Unix()

	t.Run("should auto-resolve - significant time difference", func(t *testing.T) {
		conflict := &Conflict{
			LocalTimestamp:  now + 100,
			RemoteTimestamp: now,
		}

		if !resolver.ShouldAutoResolve(conflict) {
			t.Error("Expected auto-resolve with significant time difference")
		}
	})

	t.Run("should not auto-resolve - 1 second difference", func(t *testing.T) {
		conflict := &Conflict{
			LocalTimestamp:  now + 1,
			RemoteTimestamp: now,
		}

		if resolver.ShouldAutoResolve(conflict) {
			t.Error("Expected no auto-resolve with 1 second difference")
		}
	})

	t.Run("should not auto-resolve - same timestamp", func(t *testing.T) {
		conflict := &Conflict{
			LocalTimestamp:  now,
			RemoteTimestamp: now,
		}

		if resolver.ShouldAutoResolve(conflict) {
			t.Error("Expected no auto-resolve with same timestamp")
		}
	})

	t.Run("should not auto-resolve - nil conflict", func(t *testing.T) {
		if resolver.ShouldAutoResolve(nil) {
			t.Error("Expected no auto-resolve with nil conflict")
		}
	})
}

// TestManualResolution tests the manual resolution strategy.
func TestManualResolution(t *testing.T) {
	resolver := NewResolver(ResolutionStrategyManual)

	now := time.Now().Unix()

	conflict := &Conflict{
		ItemID:     "item-1",
		LocalItem:  &models.ContentItem{ID: "item-1", UpdatedAt: now + 100},
		RemoteItem: &models.ContentItem{ID: "item-1", UpdatedAt: now},
	}

	result, err := resolver.Resolve(conflict)

	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}

	if result.Strategy != ResolutionStrategyManual {
		t.Errorf("Expected Manual strategy, got %s", result.Strategy)
	}

	if result.ConflictLog.Resolution != "manual_review_required" {
		t.Errorf("Expected 'manual_review_required' resolution, got %s", result.ConflictLog.Resolution)
	}
}

// TestMergeItems tests the merge operation.
func TestMergeItems(t *testing.T) {
	resolver := NewResolver(ResolutionStrategyLastWriteWins)

	now := time.Now().Unix()

	localItem := &models.ContentItem{
		ID:        "item-1",
		Title:     "Local Title",
		UpdatedAt: now,
		Version:   1,
	}

	remoteItem := &models.ContentItem{
		ID:        "item-1",
		Title:     "Remote Title",
		UpdatedAt: now + 100,
		Version:   2,
	}

	_, err := resolver.MergeItems(localItem, remoteItem)

	if err != ErrMergeNotSupported {
		t.Errorf("Expected ErrMergeNotSupported, got %v", err)
	}
}

// TestConflictError tests the ConflictError type.
func TestConflictError(t *testing.T) {
	err := &ConflictError{Message: "test error"}

	if err.Error() != "test error" {
		t.Errorf("Expected 'test error', got %s", err.Error())
	}

	if !IsConflictError(err) {
		t.Error("Expected IsConflictError to return true")
	}

	if IsConflictError(&ConflictError{}) {
		// Test with different error type
		otherErr := &ConflictError{}
		if !IsConflictError(otherErr) {
			t.Error("Expected IsConflictError to return true for ConflictError")
		}
	}
}
