// Package models provides data model definitions for MemoNexus Core.
package models

import (
	"time"

	"github.com/kimhsiao/memonexus/backend/internal/crypto"
)

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

// =====================================================
// T226: S3 Credential Encryption/Decryption (AES-256-GCM)
// =====================================================

// SetAccessKey encrypts and sets the S3 access key using AES-256-GCM.
// T226: Encryption at rest for sensitive credentials.
func (s *SyncCredential) SetAccessKey(accessKey, machineID string) error {
	encrypted, err := crypto.EncryptAPIKey(accessKey, machineID)
	if err != nil {
		return err
	}
	s.AccessKeyEncrypted = encrypted
	return nil
}

// GetAccessKey decrypts and returns the S3 access key.
// T226: Decryption of encrypted credentials at rest.
func (s *SyncCredential) GetAccessKey(machineID string) (string, error) {
	if s.AccessKeyEncrypted == "" {
		return "", nil
	}
	return crypto.DecryptAPIKey(s.AccessKeyEncrypted, machineID)
}

// SetSecretKey encrypts and sets the S3 secret key using AES-256-GCM.
// T226: Encryption at rest for sensitive credentials.
func (s *SyncCredential) SetSecretKey(secretKey, machineID string) error {
	encrypted, err := crypto.EncryptAPIKey(secretKey, machineID)
	if err != nil {
		return err
	}
	s.SecretKeyEncrypted = encrypted
	return nil
}

// GetSecretKey decrypts and returns the S3 secret key.
// T226: Decryption of encrypted credentials at rest.
func (s *SyncCredential) GetSecretKey(machineID string) (string, error) {
	if s.SecretKeyEncrypted == "" {
		return "", nil
	}
	return crypto.DecryptAPIKey(s.SecretKeyEncrypted, machineID)
}

// HasCredentials returns true if both access key and secret key are stored.
func (s *SyncCredential) HasCredentials() bool {
	return s.AccessKeyEncrypted != "" && s.SecretKeyEncrypted != ""
}
