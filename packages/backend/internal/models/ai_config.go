// Package models provides data model definitions for MemoNexus Core.
package models

import "time"

// AIConfig holds encrypted AI service configuration.
// APIKeyEncrypted is never exposed in JSON responses.
type AIConfig struct {
	ID              UUID   `db:"id" json:"id"`
	Provider        string `db:"provider" json:"provider"` // openai, claude, ollama
	APIEndpoint     string `db:"api_endpoint" json:"api_endpoint"`
	APIKeyEncrypted string `db:"api_key_encrypted" json:"-"` // Never expose
	ModelName       string `db:"model_name" json:"model_name"`
	MaxTokens       int    `db:"max_tokens" json:"max_tokens"`
	IsEnabled       bool   `db:"is_enabled" json:"is_enabled"`
	CreatedAt       int64  `db:"created_at" json:"created_at"`
	UpdatedAt       int64  `db:"updated_at" json:"updated_at"`
}

// TableName returns the table name for AIConfig.
func (AIConfig) TableName() string {
	return "ai_config"
}

// CreatedAtTime returns the CreatedAt as time.Time.
func (a *AIConfig) CreatedAtTime() time.Time {
	return time.Unix(a.CreatedAt, 0)
}

// UpdatedAtTime returns the UpdatedAt as time.Time.
func (a *AIConfig) UpdatedAtTime() time.Time {
	return time.Unix(a.UpdatedAt, 0)
}
