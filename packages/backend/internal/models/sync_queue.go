// Package models provides data model definitions for MemoNexus Core.
package models

import "encoding/json"

// SyncQueue represents a pending sync operation.
type SyncQueue struct {
	ID          UUID          `db:"id" json:"id"`
	Operation   string        `db:"operation" json:"operation"` // upload, download, delete
	Payload     json.RawMessage `db:"payload" json:"payload"`
	RetryCount  int           `db:"retry_count" json:"retry_count"`
	MaxRetries  int           `db:"max_retries" json:"max_retries"`
	NextRetryAt int64         `db:"next_retry_at" json:"next_retry_at"`
	Status      string        `db:"status" json:"status"` // pending, in_progress, failed, completed
	CreatedAt   int64         `db:"created_at" json:"created_at"`
	UpdatedAt   int64         `db:"updated_at" json:"updated_at"`
}

// TableName returns the table name for SyncQueue.
func (SyncQueue) TableName() string {
	return "sync_queue"
}
