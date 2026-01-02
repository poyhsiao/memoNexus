// Package sync provides synchronization interfaces and implementations.
package sync

import (
	"context"
	"time"
)

// SyncEngineInterface defines the interface for sync engine operations.
// This interface allows for mocking in tests and alternative implementations.
type SyncEngineInterface interface {
	// Sync performs a full synchronization operation.
	// Returns the sync result with statistics or an error if sync fails.
	Sync(ctx context.Context) (*SyncResult, error)

	// SetEventHandler sets the event handler for sync notifications.
	// The handler receives events during sync operations.
	SetEventHandler(handler SyncEventHandler)

	// Status returns the current sync status.
	Status() SyncStatus

	// LastSync returns the timestamp of the last successful sync.
	LastSync() *time.Time

	// PendingChanges returns the number of pending changes to sync.
	PendingChanges() int

	// LastError returns the last error that occurred during sync.
	LastError() error
}
