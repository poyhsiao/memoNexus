// Package models provides data model definitions for MemoNexus Core.
package models

import "time"

// Tag represents a user-defined label for organizing content.
type Tag struct {
	ID        UUID  `db:"id" json:"id"`
	Name      string `db:"name" json:"name"`
	Color     string `db:"color" json:"color"`
	IsDeleted bool  `db:"is_deleted" json:"is_deleted"`
	CreatedAt int64 `db:"created_at" json:"created_at"`
	UpdatedAt int64 `db:"updated_at" json:"updated_at"`
}

// TableName returns the table name for Tag.
func (Tag) TableName() string {
	return "tags"
}

// CreatedAtTime returns the CreatedAt as time.Time.
func (t *Tag) CreatedAtTime() time.Time {
	return time.Unix(t.CreatedAt, 0)
}

// UpdatedAtTime returns the UpdatedAt as time.Time.
func (t *Tag) UpdatedAtTime() time.Time {
	return time.Unix(t.UpdatedAt, 0)
}

// Touch updates the UpdatedAt timestamp.
func (t *Tag) Touch() {
	t.UpdatedAt = time.Now().Unix()
}
