// Package models provides data model definitions for MemoNexus Core.
package models

import "time"

// ExportArchive holds metadata for exported archives.
type ExportArchive struct {
	ID          UUID    `db:"id" json:"id"`
	FilePath    string  `db:"file_path" json:"file_path"`
	Checksum    string  `db:"checksum" json:"checksum"` // SHA-256
	SizeBytes   int64   `db:"size_bytes" json:"size_bytes"`
	ItemCount   int     `db:"item_count" json:"item_count"`
	IsEncrypted bool    `db:"is_encrypted" json:"is_encrypted"`
	CreatedAt   int64   `db:"created_at" json:"created_at"`
}

// TableName returns the table name for ExportArchive.
func (ExportArchive) TableName() string {
	return "export_archives"
}

// CreatedAtTime returns the CreatedAt as time.Time.
func (e *ExportArchive) CreatedAtTime() time.Time {
	return time.Unix(e.CreatedAt, 0)
}
