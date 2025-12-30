// Package models provides data model definitions for MemoNexus Core.
package models

import "time"

// ChangeLog tracks mutations for incremental sync and conflict detection.
type ChangeLog struct {
	ID        UUID   `db:"id" json:"id"`
	ItemID    UUID   `db:"item_id" json:"item_id"`
	Operation string `db:"operation" json:"operation"` // create, update, delete
	Version   int    `db:"version" json:"version"`
	Timestamp int64  `db:"timestamp" json:"timestamp"`
}

// TableName returns the table name for ChangeLog.
func (ChangeLog) TableName() string {
	return "change_log"
}

// Time returns the Timestamp as time.Time.
func (c *ChangeLog) Time() time.Time {
	return time.Unix(c.Timestamp, 0)
}
