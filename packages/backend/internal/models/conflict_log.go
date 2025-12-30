// Package models provides data model definitions for MemoNexus Core.
package models

import "time"

// ConflictLog records resolved concurrent edits for user awareness.
type ConflictLog struct {
	ID             UUID   `db:"id" json:"id"`
	ItemID         UUID   `db:"item_id" json:"item_id"`
	LocalTimestamp int64  `db:"local_timestamp" json:"local_timestamp"`
	RemoteTimestamp int64 `db:"remote_timestamp" json:"remote_timestamp"`
	Resolution     string `db:"resolution" json:"resolution"` // last_write_wins, manual
	DetectedAt     int64  `db:"detected_at" json:"detected_at"`
}

// TableName returns the table name for ConflictLog.
func (ConflictLog) TableName() string {
	return "conflict_log"
}

// DetectedAtTime returns the DetectedAt as time.Time.
func (c *ConflictLog) DetectedAtTime() time.Time {
	return time.Unix(c.DetectedAt, 0)
}
