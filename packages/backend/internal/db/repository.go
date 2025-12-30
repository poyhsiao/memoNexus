// Package db provides CRUD repository operations for MemoNexus data models.
package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/kimhsiao/memonexus/backend/internal/models"
	"github.com/kimhsiao/memonexus/backend/internal/uuid"
)

// Repository provides CRUD operations for all models.
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new Repository instance.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// =====================================================
// ContentItem Operations
// =====================================================

// CreateContentItem creates a new content item.
func (r *Repository) CreateContentItem(item *models.ContentItem) error {
	now := time.Now().Unix()
	item.ID = models.UUID(uuid.New())
	item.CreatedAt = now
	item.UpdatedAt = now
	item.Version = 1

	query := `
	INSERT INTO content_items (id, title, content_text, source_url, media_type, tags, summary,
		is_deleted, created_at, updated_at, version, content_hash)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.Exec(query, item.ID, item.Title, item.ContentText, item.SourceURL,
		item.MediaType, item.Tags, item.Summary, item.IsDeleted,
		item.CreatedAt, item.UpdatedAt, item.Version, item.ContentHash)
	return err
}

// GetContentItem retrieves a content item by ID.
func (r *Repository) GetContentItem(id string) (*models.ContentItem, error) {
	query := `
	SELECT id, title, content_text, source_url, media_type, tags, summary,
		   is_deleted, created_at, updated_at, version, content_hash
	FROM content_items WHERE id = ? AND is_deleted = 0
	`
	var item models.ContentItem
	var sourceURL, summary, contentHash sql.NullString
	err := r.db.QueryRow(query, id).Scan(
		&item.ID, &item.Title, &item.ContentText, &sourceURL, &item.MediaType,
		&item.Tags, &summary, &item.IsDeleted, &item.CreatedAt, &item.UpdatedAt,
		&item.Version, &contentHash,
	)
	if err != nil {
		return nil, err
	}
	if sourceURL.Valid {
		item.SourceURL = sourceURL.String
	}
	if summary.Valid {
		item.Summary = summary.String
	}
	if contentHash.Valid {
		item.ContentHash = contentHash.String
	}
	return &item, nil
}

// ListContentItems returns content items with pagination and filters.
func (r *Repository) ListContentItems(limit, offset int, mediaType string) ([]*models.ContentItem, error) {
	query := `
	SELECT id, title, content_text, source_url, media_type, tags, summary,
		   is_deleted, created_at, updated_at, version, content_hash
	FROM content_items WHERE is_deleted = 0
	`
	args := []interface{}{}

	if mediaType != "" {
		query += " AND media_type = ?"
		args = append(args, mediaType)
	}

	query += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*models.ContentItem
	for rows.Next() {
		var item models.ContentItem
		var sourceURL, summary, contentHash sql.NullString
		err := rows.Scan(
			&item.ID, &item.Title, &item.ContentText, &sourceURL, &item.MediaType,
			&item.Tags, &summary, &item.IsDeleted, &item.CreatedAt, &item.UpdatedAt,
			&item.Version, &contentHash,
		)
		if err != nil {
			return nil, err
		}
		if sourceURL.Valid {
			item.SourceURL = sourceURL.String
		}
		if summary.Valid {
			item.Summary = summary.String
		}
		if contentHash.Valid {
			item.ContentHash = contentHash.String
		}
		items = append(items, &item)
	}
	// Check for errors that occurred during iteration
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

// UpdateContentItem updates an existing content item.
func (r *Repository) UpdateContentItem(item *models.ContentItem) error {
	item.Touch()
	query := `
	UPDATE content_items
	SET title = ?, content_text = ?, source_url = ?, media_type = ?, tags = ?,
		summary = ?, updated_at = ?, version = ?, content_hash = ?
	WHERE id = ? AND is_deleted = 0
	`
	result, err := r.db.Exec(query, item.Title, item.ContentText, item.SourceURL,
		item.MediaType, item.Tags, item.Summary, item.UpdatedAt, item.Version,
		item.ContentHash, item.ID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("content item not found: %s", item.ID)
	}
	return nil
}

// DeleteContentItem soft deletes a content item.
func (r *Repository) DeleteContentItem(id string) error {
	query := `UPDATE content_items SET is_deleted = 1, updated_at = ? WHERE id = ?`
	now := time.Now().Unix()
	result, err := r.db.Exec(query, now, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("content item not found: %s", id)
	}
	return nil
}

// =====================================================
// Tag Operations
// =====================================================

// CreateTag creates a new tag.
func (r *Repository) CreateTag(tag *models.Tag) error {
	now := time.Now().Unix()
	tag.ID = models.UUID(uuid.New())
	tag.CreatedAt = now
	tag.UpdatedAt = now

	query := `
	INSERT INTO tags (id, name, color, is_deleted, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.Exec(query, tag.ID, tag.Name, tag.Color, tag.IsDeleted,
		tag.CreatedAt, tag.UpdatedAt)
	return err
}

// GetTag retrieves a tag by ID.
func (r *Repository) GetTag(id string) (*models.Tag, error) {
	query := `SELECT id, name, color, is_deleted, created_at, updated_at FROM tags WHERE id = ?`
	var tag models.Tag
	err := r.db.QueryRow(query, id).Scan(&tag.ID, &tag.Name, &tag.Color,
		&tag.IsDeleted, &tag.CreatedAt, &tag.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

// ListTags returns all tags.
func (r *Repository) ListTags() ([]*models.Tag, error) {
	query := `SELECT id, name, color, is_deleted, created_at, updated_at FROM tags WHERE is_deleted = 0 ORDER BY name`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []*models.Tag
	for rows.Next() {
		var tag models.Tag
		err := rows.Scan(&tag.ID, &tag.Name, &tag.Color, &tag.IsDeleted,
			&tag.CreatedAt, &tag.UpdatedAt)
		if err != nil {
			return nil, err
		}
		tags = append(tags, &tag)
	}
	return tags, nil
}

// UpdateTag updates an existing tag.
func (r *Repository) UpdateTag(tag *models.Tag) error {
	tag.UpdatedAt = time.Now().Unix()
	query := `UPDATE tags SET name = ?, color = ?, updated_at = ? WHERE id = ? AND is_deleted = 0`
	result, err := r.db.Exec(query, tag.Name, tag.Color, tag.UpdatedAt, tag.ID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// DeleteTag soft deletes a tag.
func (r *Repository) DeleteTag(id string) error {
	query := `UPDATE tags SET is_deleted = 1, updated_at = ? WHERE id = ? AND is_deleted = 0`
	now := time.Now().Unix()
	result, err := r.db.Exec(query, now, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// =====================================================
// ChangeLog Operations
// =====================================================

// CreateChangeLog creates a new change log entry.
func (r *Repository) CreateChangeLog(log *models.ChangeLog) error {
	log.ID = models.UUID(uuid.New())
	log.Timestamp = time.Now().Unix()

	query := `
	INSERT INTO change_log (id, item_id, operation, version, timestamp)
	VALUES (?, ?, ?, ?, ?)
	`
	_, err := r.db.Exec(query, log.ID, log.ItemID, log.Operation, log.Version, log.Timestamp)
	return err
}

// =====================================================
// ConflictLog Operations
// =====================================================

// CreateConflictLog creates a new conflict log entry.
func (r *Repository) CreateConflictLog(log *models.ConflictLog) error {
	log.ID = models.UUID(uuid.New())
	log.DetectedAt = time.Now().Unix()

	query := `
	INSERT INTO conflict_log (id, item_id, local_timestamp, remote_timestamp, resolution, detected_at)
	VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.Exec(query, log.ID, log.ItemID, log.LocalTimestamp, log.RemoteTimestamp,
		log.Resolution, log.DetectedAt)
	return err
}

// =====================================================
// SyncQueue Operations
// =====================================================

// CreateSyncQueue creates a new sync queue entry.
func (r *Repository) CreateSyncQueue(entry *models.SyncQueue) error {
	entry.ID = models.UUID(uuid.New())
	now := time.Now().Unix()
	entry.CreatedAt = now
	entry.UpdatedAt = now

	query := `
	INSERT INTO sync_queue (id, operation, payload, retry_count, max_retries, next_retry_at, status, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.Exec(query, entry.ID, entry.Operation, entry.Payload, entry.RetryCount,
		entry.MaxRetries, entry.NextRetryAt, entry.Status, entry.CreatedAt, entry.UpdatedAt)
	return err
}
