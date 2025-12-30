// Package models provides data model definitions for MemoNexus Core.
package models

import "time"

// SyncCredential holds encrypted S3 configuration.
// AccessKeyEncrypted and SecretKeyEncrypted are never exposed in JSON responses.
type SyncCredential struct {
	ID                UUID   `db:"id" json:"id"`
	Endpoint          string `db:"endpoint" json:"endpoint"`
	BucketName        string `db:"bucket_name" json:"bucket_name"`
	Region            string `db:"region" json:"region,omitempty"`
	AccessKeyEncrypted string `db:"access_key_encrypted" json:"-"` // Never expose
	SecretKeyEncrypted string `db:"secret_key_encrypted" json:"-"` // Never expose
	IsEnabled         bool   `db:"is_enabled" json:"is_enabled"`
	CreatedAt         int64  `db:"created_at" json:"created_at"`
	UpdatedAt         int64  `db:"updated_at" json:"updated_at"`
}

// TableName returns the table name for SyncCredential.
func (SyncCredential) TableName() string {
	return "sync_credentials"
}

// CreatedAtTime returns the CreatedAt as time.Time.
func (s *SyncCredential) CreatedAtTime() time.Time {
	return time.Unix(s.CreatedAt, 0)
}

// UpdatedAtTime returns the UpdatedAt as time.Time.
func (s *SyncCredential) UpdatedAtTime() time.Time {
	return time.Unix(s.UpdatedAt, 0)
}
