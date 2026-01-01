// Package conflict provides conflict resolution for multi-device synchronization.
// T156: Conflict resolver using "last write wins" strategy.
package conflict

import (
	"time"

	"github.com/kimhsiao/memonexus/backend/internal/logging"
	"github.com/kimhsiao/memonexus/backend/internal/models"
)

// ResolutionStrategy defines how conflicts are resolved.
type ResolutionStrategy string

const (
	ResolutionStrategyLastWriteWins ResolutionStrategy = "last_write_wins"
	ResolutionStrategyManual        ResolutionStrategy = "manual"
)

// Resolver handles conflict resolution during synchronization.
type Resolver struct {
	strategy ResolutionStrategy
}

// NewResolver creates a new Resolver with the specified strategy.
func NewResolver(strategy ResolutionStrategy) *Resolver {
	return &Resolver{
		strategy: strategy,
	}
}

// Conflict represents a detected conflict between local and remote changes.
type Conflict struct {
	ItemID          string
	LocalItem       *models.ContentItem
	RemoteItem      *models.ContentItem
	LocalTimestamp  int64
	RemoteTimestamp int64
	DetectedAt      int64
}

// ResolveResult represents the outcome of conflict resolution.
type ResolveResult struct {
	WinningItem *models.ContentItem // The item that should be kept
	LosingItem  *models.ContentItem // The item that was overwritten
	Strategy    ResolutionStrategy
	ConflictLog *models.ConflictLog // Log entry for awareness
}

// Resolve resolves a conflict using the configured strategy.
// T212: Concurrent edit conflict logging with item UUID and both timestamps.
func (r *Resolver) Resolve(conflict *Conflict) (*ResolveResult, error) {
	if conflict.LocalItem == nil || conflict.RemoteItem == nil {
		return nil, ErrInvalidConflict
	}

	// Ensure IDs match
	if conflict.LocalItem.ID != conflict.RemoteItem.ID {
		return nil, ErrItemIDMismatch
	}

	logging.Info("Resolving conflict",
		map[string]interface{}{
			"item_id":          conflict.LocalItem.ID,
			"local_timestamp":  conflict.LocalItem.UpdatedAt,
			"remote_timestamp": conflict.RemoteItem.UpdatedAt,
			"strategy":         r.strategy,
		})

	switch r.strategy {
	case ResolutionStrategyLastWriteWins:
		return r.resolveLastWriteWins(conflict)
	case ResolutionStrategyManual:
		return r.resolveManual(conflict)
	default:
		return r.resolveLastWriteWins(conflict)
	}
}

// resolveLastWriteWins implements "last write wins" strategy.
// The item with the newer UpdatedAt timestamp wins.
// T212: Conflict resolution logging.
func (r *Resolver) resolveLastWriteWins(conflict *Conflict) (*ResolveResult, error) {
	var winningItem, losingItem *models.ContentItem
	var resolution string

	if conflict.LocalItem.UpdatedAt >= conflict.RemoteItem.UpdatedAt {
		// Local wins (local is newer or same timestamp)
		winningItem = conflict.LocalItem
		losingItem = conflict.RemoteItem
		resolution = "local_wins"
	} else {
		// Remote wins (remote is newer)
		winningItem = conflict.RemoteItem
		losingItem = conflict.LocalItem
		resolution = "remote_wins"
	}

	// Create conflict log for user awareness
	conflictLog := &models.ConflictLog{
		ItemID:          models.UUID(winningItem.ID),
		LocalTimestamp:  conflict.LocalItem.UpdatedAt,
		RemoteTimestamp: conflict.RemoteItem.UpdatedAt,
		Resolution:      resolution,
		DetectedAt:      time.Now().Unix(),
	}

	// Determine winner side for structured logging
	winnerSide := ""
	switch resolution {
	case "local_wins":
		winnerSide = "local"
	case "remote_wins":
		winnerSide = "remote"
	default:
		winnerSide = "unknown"
	}

	logging.Info("Conflict resolved using last-write-wins",
		map[string]interface{}{
			"item_id":          winningItem.ID,
			"winner_id":        winningItem.ID,
			"winner_side":      winnerSide,
			"local_timestamp":  conflict.LocalItem.UpdatedAt,
			"remote_timestamp": conflict.RemoteItem.UpdatedAt,
			"resolution":       resolution,
		})

	return &ResolveResult{
		WinningItem: winningItem,
		LosingItem:  losingItem,
		Strategy:    ResolutionStrategyLastWriteWins,
		ConflictLog: conflictLog,
	}, nil
}

// resolveManual returns both items for manual resolution.
// In a real implementation, this would queue the conflict for user review.
// T212: Manual conflict resolution logging.
func (r *Resolver) resolveManual(conflict *Conflict) (*ResolveResult, error) {
	// For manual resolution, we keep the local version and mark for review
	conflictLog := &models.ConflictLog{
		ItemID:          models.UUID(conflict.LocalItem.ID),
		LocalTimestamp:  conflict.LocalItem.UpdatedAt,
		RemoteTimestamp: conflict.RemoteItem.UpdatedAt,
		Resolution:      "manual_review_required",
		DetectedAt:      time.Now().Unix(),
	}

	logging.Warn("Conflict queued for manual review",
		map[string]interface{}{
			"item_id":          conflict.LocalItem.ID,
			"local_timestamp":  conflict.LocalItem.UpdatedAt,
			"remote_timestamp": conflict.RemoteItem.UpdatedAt,
			"resolution":       "manual_review_required",
		})

	return &ResolveResult{
		WinningItem: conflict.LocalItem,
		LosingItem:  conflict.RemoteItem,
		Strategy:    ResolutionStrategyManual,
		ConflictLog: conflictLog,
	}, nil
}

// DetectConflict detects if there's a conflict between local and remote items.
// A conflict exists when both items have been modified since last sync.
// T212: Conflict detection logging.
func (r *Resolver) DetectConflict(localItem, remoteItem *models.ContentItem) (*Conflict, bool) {
	// No conflict if one item doesn't exist
	if localItem == nil || remoteItem == nil {
		return nil, false
	}

	// No conflict if IDs don't match (different items)
	if localItem.ID != remoteItem.ID {
		return nil, false
	}

	// Check if both have been modified (version differs)
	if localItem.Version == remoteItem.Version {
		return nil, false
	}

	// Conflict detected
	conflict := &Conflict{
		ItemID:          string(localItem.ID),
		LocalItem:       localItem,
		RemoteItem:      remoteItem,
		LocalTimestamp:  localItem.UpdatedAt,
		RemoteTimestamp: remoteItem.UpdatedAt,
		DetectedAt:      time.Now().Unix(),
	}

	logging.Warn("Concurrent edit conflict detected",
		map[string]interface{}{
			"item_id":          localItem.ID,
			"local_timestamp":  localItem.UpdatedAt,
			"remote_timestamp": remoteItem.UpdatedAt,
			"local_version":    localItem.Version,
			"remote_version":   remoteItem.Version,
		})

	return conflict, true
}

// ResolveMultiple resolves multiple conflicts in batch.
func (r *Resolver) ResolveMultiple(conflicts []*Conflict) ([]*ResolveResult, error) {
	results := make([]*ResolveResult, 0, len(conflicts))

	for _, conflict := range conflicts {
		result, err := r.Resolve(conflict)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	return results, nil
}

// ShouldAutoResolve determines if a conflict can be auto-resolved.
// Returns true if the timestamp difference is significant enough.
func (r *Resolver) ShouldAutoResolve(conflict *Conflict) bool {
	if conflict == nil {
		return false
	}

	// Auto-resolve if timestamps differ by more than 1 second
	diff := conflict.LocalTimestamp - conflict.RemoteTimestamp
	if diff < 0 {
		diff = -diff
	}

	return diff > 1
}

// MergeItems attempts to merge local and remote changes.
// This is a placeholder for future merge strategies.
func (r *Resolver) MergeItems(localItem, remoteItem *models.ContentItem) (*models.ContentItem, error) {
	// For now, we don't support automatic merging
	// This could be implemented in the future for specific field-level merges
	return nil, ErrMergeNotSupported
}

// Errors
var (
	ErrInvalidConflict    = &ConflictError{Message: "invalid conflict: both items must be non-nil"}
	ErrItemIDMismatch     = &ConflictError{Message: "item ID mismatch"}
	ErrMergeNotSupported  = &ConflictError{Message: "merge not supported"}
	ErrConflictUnresolved = &ConflictError{Message: "conflict could not be resolved"}
)

// ConflictError represents a conflict resolution error.
type ConflictError struct {
	Message string
}

func (e *ConflictError) Error() string {
	return e.Message
}

// IsConflictError checks if an error is a ConflictError.
func IsConflictError(err error) bool {
	_, ok := err.(*ConflictError)
	return ok
}
