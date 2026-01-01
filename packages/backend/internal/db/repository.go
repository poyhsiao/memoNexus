// Package db provides CRUD repository operations for MemoNexus data models.
package db

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/kimhsiao/memonexus/backend/internal/models"
	"github.com/kimhsiao/memonexus/backend/internal/uuid"
)

// Repository provides CRUD operations for all models.
// T222: Added prepared statement cache for query optimization.
type Repository struct {
	db *sql.DB

	// T222: Prepared statement cache for frequently used queries
	// Statements are prepared on first use and cached for reuse
	stmtCache sync.Map // map[string]*sql.Stmt
}

// T222: PrepareStmt gets or creates a prepared statement from cache.
// Key is the query string, value is the prepared statement.
// Prepared statements are cached to avoid repeated SQL parsing overhead.
func (r *Repository) PrepareStmt(query string) (*sql.Stmt, error) {
	// Try to get from cache first
	if stmt, ok := r.stmtCache.Load(query); ok {
		return stmt.(*sql.Stmt), nil
	}

	// Prepare and cache
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}

	// Store in cache (if already stored by another goroutine, use existing)
	actual, loaded := r.stmtCache.LoadOrStore(query, stmt)
	if loaded {
		// Another goroutine already prepared this, close our duplicate
		stmt.Close()
		return actual.(*sql.Stmt), nil
	}

	return stmt, nil
}

// Close closes all cached prepared statements.
// Should be called when the Repository is no longer needed.
func (r *Repository) Close() error {
	var firstErr error
	r.stmtCache.Range(func(key, value interface{}) bool {
		stmt := value.(*sql.Stmt)
		if err := stmt.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
		return true
	})
	return firstErr
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
// T222: Uses prepared statement for repeated queries.
func (r *Repository) GetContentItem(id string) (*models.ContentItem, error) {
	query := `
	SELECT id, title, content_text, source_url, media_type, tags, summary,
		   is_deleted, created_at, updated_at, version, content_hash
	FROM content_items WHERE id = ? AND is_deleted = 0
	`
	// T222: Use prepared statement from cache
	stmt, err := r.PrepareStmt(query)
	if err != nil {
		return nil, err
	}

	var item models.ContentItem
	var sourceURL, summary, contentHash sql.NullString
	err = stmt.QueryRow(id).Scan(
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
// T222: Uses prepared statements for both query variants (with/without mediaType filter).
func (r *Repository) ListContentItems(limit, offset int, mediaType string) ([]*models.ContentItem, error) {
	// T222: Build query based on filters
	baseQuery := `
	SELECT id, title, content_text, source_url, media_type, tags, summary,
		   is_deleted, created_at, updated_at, version, content_hash
	FROM content_items WHERE is_deleted = 0
	`
	orderLimit := " ORDER BY created_at DESC LIMIT ? OFFSET ?"

	var query string
	var args []interface{}

	if mediaType != "" {
		query = baseQuery + " AND media_type = ?" + orderLimit
		args = []interface{}{mediaType, limit, offset}
	} else {
		query = baseQuery + orderLimit
		args = []interface{}{limit, offset}
	}

	// T222: Use prepared statement from cache
	stmt, err := r.PrepareStmt(query)
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(args...)
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

// =====================================================
// AIConfig Operations
// =====================================================

// GetAIConfig retrieves the current AI configuration (only one active config allowed).
func (r *Repository) GetAIConfig() (*models.AIConfig, error) {
	query := `
	SELECT id, provider, api_endpoint, api_key_encrypted, model_name, max_tokens, is_enabled, created_at, updated_at
	FROM ai_config
	WHERE is_enabled = 1
	ORDER BY updated_at DESC
	LIMIT 1
	`
	var config models.AIConfig
	err := r.db.QueryRow(query).Scan(
		&config.ID, &config.Provider, &config.APIEndpoint, &config.APIKeyEncrypted,
		&config.ModelName, &config.MaxTokens, &config.IsEnabled, &config.CreatedAt, &config.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// SaveAIConfig saves or updates AI configuration.
// If ID is empty, creates new config; otherwise updates existing.
func (r *Repository) SaveAIConfig(config *models.AIConfig) error {
	now := time.Now().Unix()

	if config.ID == "" {
		// Create new
		config.ID = models.UUID(uuid.New())
		config.CreatedAt = now
		config.UpdatedAt = now

		query := `
		INSERT INTO ai_config (id, provider, api_endpoint, api_key_encrypted, model_name, max_tokens, is_enabled, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`
		_, err := r.db.Exec(query, config.ID, config.Provider, config.APIEndpoint, config.APIKeyEncrypted,
			config.ModelName, config.MaxTokens, config.IsEnabled, config.CreatedAt, config.UpdatedAt)
		return err
	}

	// Update existing
	config.UpdatedAt = now
	query := `
	UPDATE ai_config
	SET provider = ?, api_endpoint = ?, api_key_encrypted = ?, model_name = ?, max_tokens = ?, is_enabled = ?, updated_at = ?
	WHERE id = ?
	`
	_, err := r.db.Exec(query, config.Provider, config.APIEndpoint, config.APIKeyEncrypted,
		config.ModelName, config.MaxTokens, config.IsEnabled, config.UpdatedAt, config.ID)
	return err
}

// DeleteAIConfig deletes (disables) the AI configuration with the given ID.
func (r *Repository) DeleteAIConfig(id string) error {
	query := `DELETE FROM ai_config WHERE id = ?`
	_, err := r.db.Exec(query, id)
	return err
}

// DisableAllAIConfig disables all AI configurations (used when setting a new one).
func (r *Repository) DisableAllAIConfig() error {
	query := `UPDATE ai_config SET is_enabled = 0 WHERE is_enabled = 1`
	_, err := r.db.Exec(query)
	return err
}

// =====================================================
// Sync Credential Methods (T159-T161)
// =====================================================

// GetSyncCredentials retrieves the currently enabled sync credentials.
func (r *Repository) GetSyncCredentials() (*models.SyncCredential, error) {
	query := `SELECT id, endpoint, bucket_name, region, access_key_encrypted, secret_key_encrypted, is_enabled, created_at, updated_at
			  FROM sync_credentials WHERE is_enabled = 1 LIMIT 1`

	var cred models.SyncCredential
	err := r.db.QueryRow(query).Scan(
		&cred.ID, &cred.Endpoint, &cred.BucketName, &cred.Region,
		&cred.AccessKeyEncrypted, &cred.SecretKeyEncrypted,
		&cred.IsEnabled, &cred.CreatedAt, &cred.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}
	return &cred, nil
}

// SaveSyncCredential saves a new sync credential configuration.
func (r *Repository) SaveSyncCredential(cred *models.SyncCredential) error {
	query := `INSERT INTO sync_credentials (id, endpoint, bucket_name, region, access_key_encrypted, secret_key_encrypted, is_enabled, created_at, updated_at)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	cred.ID = models.UUID(uuid.New())
	now := time.Now().Unix()
	cred.CreatedAt = now
	cred.UpdatedAt = now

	_, err := r.db.Exec(query,
		cred.ID, cred.Endpoint, cred.BucketName, cred.Region,
		cred.AccessKeyEncrypted, cred.SecretKeyEncrypted,
		cred.IsEnabled, cred.CreatedAt, cred.UpdatedAt,
	)

	return err
}

// DeleteSyncCredential deletes a sync credential by ID.
func (r *Repository) DeleteSyncCredential(id string) error {
	query := `DELETE FROM sync_credentials WHERE id = ?`
	_, err := r.db.Exec(query, id)
	return err
}

// DisableAllSyncCredentials disables all sync credentials (used when setting a new one).
func (r *Repository) DisableAllSyncCredentials() error {
	query := `UPDATE sync_credentials SET is_enabled = 0 WHERE is_enabled = 1`
	_, err := r.db.Exec(query)
	return err
}
