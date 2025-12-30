// Package models provides data model definitions for MemoNexus Core.
package models

import (
	"database/sql/driver"
	"time"
)

// UUID is a wrapper around string for UUID v4 type safety.
type UUID string

// Value implements driver.Valuer for UUID.
func (u UUID) Value() (driver.Value, error) {
	return string(u), nil
}

// Scan implements sql.Scanner for UUID.
func (u *UUID) Scan(value interface{}) error {
	if value == nil {
		*u = ""
		return nil
	}
	*u = UUID(value.([]byte))
	return nil
}

// String returns the string representation of the UUID.
func (u UUID) String() string {
	return string(u)
}

// ContentItem represents a captured content item.
type ContentItem struct {
	ID          UUID    `db:"id" json:"id"`
	Title       string  `db:"title" json:"title"`
	ContentText string  `db:"content_text" json:"content_text"`
	SourceURL   string  `db:"source_url" json:"source_url,omitempty"`
	MediaType   string  `db:"media_type" json:"media_type"`
	Tags        string  `db:"tags" json:"tags"` // Comma-separated
	Summary     string  `db:"summary" json:"summary,omitempty"`
	IsDeleted   bool    `db:"is_deleted" json:"is_deleted"`
	CreatedAt   int64   `db:"created_at" json:"created_at"`
	UpdatedAt   int64   `db:"updated_at" json:"updated_at"`
	Version     int     `db:"version" json:"version"`
	ContentHash string  `db:"content_hash" json:"content_hash,omitempty"`
}

// TableName returns the table name for ContentItem.
func (ContentItem) TableName() string {
	return "content_items"
}

// CreatedAtTime returns the CreatedAt as time.Time.
func (c *ContentItem) CreatedAtTime() time.Time {
	return time.Unix(c.CreatedAt, 0)
}

// UpdatedAtTime returns the UpdatedAt as time.Time.
func (c *ContentItem) UpdatedAtTime() time.Time {
	return time.Unix(c.UpdatedAt, 0)
}

// Touch updates the UpdatedAt timestamp.
func (c *ContentItem) Touch() {
	c.UpdatedAt = time.Now().Unix()
	c.Version++
}
