// Package db provides repository interfaces for MemoNexus data models.
package db

import (
	"github.com/kimhsiao/memonexus/backend/internal/models"
)

// ContentItemRepository defines operations for content item persistence.
// This interface allows mocking for testing and follows the Interface Segregation Principle.
type ContentItemRepository interface {
	// CreateContentItem creates a new content item.
	CreateContentItem(item *models.ContentItem) error

	// GetContentItem retrieves a content item by ID.
	GetContentItem(id string) (*models.ContentItem, error)

	// ListContentItems returns content items with pagination and filters.
	ListContentItems(limit, offset int, mediaType string) ([]*models.ContentItem, error)

	// UpdateContentItem updates an existing content item.
	UpdateContentItem(item *models.ContentItem) error

	// DeleteContentItem soft deletes a content item.
	DeleteContentItem(id string) error
}

// ChangeLogRepository defines operations for change log persistence.
type ChangeLogRepository interface {
	// CreateChangeLog creates a new change log entry.
	CreateChangeLog(log *models.ChangeLog) error
}

// ConflictLogRepository defines operations for conflict log persistence.
type ConflictLogRepository interface {
	// CreateConflictLog creates a new conflict log entry.
	CreateConflictLog(log *models.ConflictLog) error
}

// SyncRepository combines repositories needed for sync operations.
// This is a marker interface that groups related repositories for convenience.
type SyncRepository interface {
	ContentItemRepository
	ChangeLogRepository
	ConflictLogRepository
}

// Ensure *Repository implements the interfaces at compile time.
var (
	_ ContentItemRepository = (*Repository)(nil)
	_ ChangeLogRepository   = (*Repository)(nil)
	_ ConflictLogRepository = (*Repository)(nil)
	_ SyncRepository        = (*Repository)(nil)
)
