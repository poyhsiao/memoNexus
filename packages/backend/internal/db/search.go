// Package db provides FTS5 search functionality for content items.
package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/kimhsiao/memonexus/backend/internal/models"
	"github.com/kimhsiao/memonexus/backend/internal/uuid"
)

// SearchOptions contains parameters for search queries.
type SearchOptions struct {
	// Query is the FTS5 search query (required)
	Query string

	// Limit is the maximum number of results (default: 20, max: 100)
	Limit int

	// MediaType filters results by media type
	MediaType string

	// Tags filters results by tag names (comma-separated)
	Tags string

	// DateFrom filters results created after this Unix timestamp
	DateFrom int64

	// DateTo filters results created before this Unix timestamp
	DateTo int64
}

// SearchResult represents a single search result with relevance score.
type SearchResult struct {
	Item         *models.ContentItem
	Relevance    float64
	MatchedTerms []string
}

// SearchResponse contains search results and metadata.
type SearchResponse struct {
	Results []*SearchResult
	Total   int
	Query   string
}

// Search performs FTS5 full-text search on content items.
// Implements T115: FTS5 search service with BM25 ranking and Unicode support.
// Constitution requirement SC-002: <100ms for 10,000 items
func (r *Repository) Search(opts *SearchOptions) (*SearchResponse, error) {
	if opts == nil || opts.Query == "" {
		return nil, fmt.Errorf("search query is required")
	}

	// Apply defaults and limits
	if opts.Limit <= 0 {
		opts.Limit = 20
	}
	if opts.Limit > 100 {
		opts.Limit = 100
	}

	// Build the search query with filters
	baseQuery := `
		SELECT ci.id, ci.title, ci.content_text, ci.source_url, ci.media_type, ci.tags,
			   ci.summary, ci.is_deleted, ci.created_at, ci.updated_at, ci.version, ci.content_hash
		FROM content_items ci
		INNER JOIN content_items_fts fts ON ci.rowid = fts.rowid
		WHERE content_items_fts MATCH ? AND ci.is_deleted = 0
	`

	whereClauses := []string{}
	args := []interface{}{opts.Query}

	// Add media type filter
	if opts.MediaType != "" {
		whereClauses = append(whereClauses, "ci.media_type = ?")
		args = append(args, opts.MediaType)
	}

	// Add date range filters
	if opts.DateFrom > 0 {
		whereClauses = append(whereClauses, "ci.created_at >= ?")
		args = append(args, opts.DateFrom)
	}
	if opts.DateTo > 0 {
		whereClauses = append(whereClauses, "ci.created_at <= ?")
		args = append(args, opts.DateTo)
	}

	// Add tag filter (simple LIKE matching - can be enhanced later)
	if opts.Tags != "" {
		// Split tags and build OR conditions
		tagList := strings.Split(opts.Tags, ",")
		tagConditions := make([]string, 0, len(tagList))
		for _, tag := range tagList {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				tagConditions = append(tagConditions, "ci.tags LIKE ?")
				args = append(args, "%"+tag+"%")
			}
		}
		if len(tagConditions) > 0 {
			whereClauses = append(whereClauses, "("+strings.Join(tagConditions, " OR ")+")")
		}
	}

	// Combine WHERE clauses
	if len(whereClauses) > 0 {
		baseQuery += " AND " + strings.Join(whereClauses, " AND ")
	}

	// Add ordering and limit
	baseQuery += " ORDER BY rank LIMIT ?"
	args = append(args, opts.Limit)

	// Execute the search query
	rows, err := r.db.Query(baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("search query failed: %w", err)
	}
	defer rows.Close()

	var results []*SearchResult
	for rows.Next() {
		var item models.ContentItem
		var sourceURL, summary, contentHash sql.NullString
		err := rows.Scan(
			&item.ID, &item.Title, &item.ContentText, &sourceURL, &item.MediaType,
			&item.Tags, &summary, &item.IsDeleted, &item.CreatedAt, &item.UpdatedAt,
			&item.Version, &contentHash,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan search result: %w", err)
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

		result := &SearchResult{
			Item:      &item,
			Relevance: 0, // BM25 rank is implicit in ORDER BY, we can add rank column if needed
		}
		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating search results: %w", err)
	}

	// Get total count (without limit)
	countQuery := `
		SELECT COUNT(*)
		FROM content_items ci
		INNER JOIN content_items_fts fts ON ci.rowid = fts.rowid
		WHERE content_items_fts MATCH ? AND ci.is_deleted = 0
	`
	countArgs := []interface{}{opts.Query}
	if len(whereClauses) > 0 {
		countQuery += " AND " + strings.Join(whereClauses, " AND ")
		// Rebuild count args (excluding LIMIT)
		for i := 1; i < len(args)-1; i++ { // Skip first (query) and last (limit)
			countArgs = append(countArgs, args[i])
		}
	}

	var total int
	err = r.db.QueryRow(countQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count search results: %w", err)
	}

	return &SearchResponse{
		Results: results,
		Total:   total,
		Query:   opts.Query,
	}, nil
}

// SearchSimple performs a simple search query with just the search string.
// Convenience method for basic search without filters.
func (r *Repository) SearchSimple(query string, limit int) (*SearchResponse, error) {
	return r.Search(&SearchOptions{
		Query: query,
		Limit: limit,
	})
}

// =====================================================
// FTS Index Management (T223)
// =====================================================

// OptimizeFTSIndex optimizes the FTS5 full-text search index.
// T223: Compacts the index by merging smaller segments into larger ones.
// This should be called periodically or after large bulk imports.
// For 10K+ items, this can significantly improve query performance.
func (r *Repository) OptimizeFTSIndex() error {
	// FTS5 optimize command compacts the index
	query := `INSERT INTO content_fts(content_fts) VALUES('optimize')`
	_, err := r.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to optimize FTS index: %w", err)
	}
	return nil
}

// RebuildFTSIndex completely rebuilds the FTS index from content_items.
// T223: Use this if the index becomes corrupted or out of sync.
// This is a slow operation for large datasets and should be used sparingly.
func (r *Repository) RebuildFTSIndex() error {
	// Start transaction
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Drop existing FTS table
	_, err = tx.Exec(`DROP TABLE IF EXISTS content_fts`)
	if err != nil {
		return fmt.Errorf("failed to drop FTS table: %w", err)
	}

	// Recreate FTS table
	createFTS := `
	CREATE VIRTUAL TABLE content_fts USING fts5(
		title,
		content_text,
		tags,
		content=content_items,
		content_rowid=rowid,
		tokenize='unicode61 remove_diacritics 1'
	)`
	_, err = tx.Exec(createFTS)
	if err != nil {
		return fmt.Errorf("failed to recreate FTS table: %w", err)
	}

	// Populate FTS index with existing content
	populateFTS := `
	INSERT INTO content_fts(rowid, title, content_text, tags)
	SELECT rowid, title, content_text, tags FROM content_items
	`
	_, err = tx.Exec(populateFTS)
	if err != nil {
		return fmt.Errorf("failed to populate FTS index: %w", err)
	}

	// Recreate triggers
	triggers := []string{
		`CREATE TRIGGER content_items_ai AFTER INSERT ON content_items BEGIN
			INSERT INTO content_fts(rowid, title, content_text, tags)
			VALUES (new.rowid, new.title, new.content_text, new.tags);
		END`,
		`CREATE TRIGGER content_items_ad AFTER DELETE ON content_items BEGIN
			INSERT INTO content_fts(content_fts, rowid, title, content_text, tags)
			VALUES ('delete', old.rowid, old.title, old.content_text, old.tags);
		END`,
		`CREATE TRIGGER content_items_au AFTER UPDATE ON content_items BEGIN
			INSERT INTO content_fts(content_fts, rowid, title, content_text, tags)
			VALUES ('delete', old.rowid, old.title, old.content_text, old.tags);
			INSERT INTO content_fts(rowid, title, content_text, tags)
			VALUES (new.rowid, new.title, new.content_text, new.tags);
		END`,
	}

	for _, trigger := range triggers {
		_, err = tx.Exec(trigger)
		if err != nil {
			return fmt.Errorf("failed to recreate trigger: %w", err)
		}
	}

	return tx.Commit()
}

// FTSIntegrityCheck verifies the FTS index is consistent with source data.
// T223: Returns true if index is valid, false otherwise.
// For large datasets, this can take significant time.
func (r *Repository) FTSIntegrityCheck() (bool, error) {
	// FTS5 integrity check command
	query := `INSERT INTO content_fts(content_fts, rank) VALUES('integrity-check', 0)`
	_, err := r.db.Exec(query)
	if err != nil {
		// Integrity check failed - index is corrupted
		return false, nil
	}
	return true, nil
}

// FTSIndexSize returns the approximate size of the FTS index in bytes.
// T223: Useful for monitoring index growth and deciding when to optimize.
func (r *Repository) FTSIndexSize() (int64, error) {
	query := `
	SELECT SUM(pgsize) AS size
	FROM sqlite_dbpage
	WHERE pgno IN (
		SELECT rootpage FROM sqlite_master
		WHERE tbl_name = 'content_fts'
		UNION ALL
		SELECT rootpage FROM sqlite_master
		WHERE sql LIKE '%content_fts%'
	)
	`
	var size sql.NullInt64
	err := r.db.QueryRow(query).Scan(&size)
	if err != nil {
		return 0, fmt.Errorf("failed to get FTS index size: %w", err)
	}
	if size.Valid {
		return size.Int64, nil
	}
	return 0, nil
}

// BatchImportContent imports multiple content items efficiently.
// T223: Disables triggers during import, then rebuilds FTS index once.
// Much faster than inserting items one-by-one with triggers enabled.
// Returns the number of items imported.
func (r *Repository) BatchImportContent(items []*models.ContentItem) (int, error) {
	if len(items) == 0 {
		return 0, nil
	}

	// Start transaction
	tx, err := r.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Disable FTS triggers temporarily for bulk insert
	_, err = tx.Exec(`DROP TRIGGER IF EXISTS content_items_ai`)
	if err != nil {
		return 0, fmt.Errorf("failed to disable FTS insert trigger: %w", err)
	}
	_, err = tx.Exec(`DROP TRIGGER IF EXISTS content_items_au`)
	if err != nil {
		return 0, fmt.Errorf("failed to disable FTS update trigger: %w", err)
	}

	// Prepare insert statement
	stmt, err := tx.Prepare(`
	INSERT INTO content_items (id, title, content_text, source_url, media_type, tags, summary,
		is_deleted, created_at, updated_at, version, content_hash)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return 0, fmt.Errorf("failed to prepare insert statement: %w", err)
	}
	defer stmt.Close()

	// Insert all items
	now := time.Now().Unix()
	count := 0
	for _, item := range items {
		if item.ID == "" {
			item.ID = models.UUID(uuid.New())
			item.CreatedAt = now
			item.UpdatedAt = now
			item.Version = 1
		}

		_, err := stmt.Exec(
			item.ID, item.Title, item.ContentText, item.SourceURL,
			item.MediaType, item.Tags, item.Summary, item.IsDeleted,
			item.CreatedAt, item.UpdatedAt, item.Version, item.ContentHash,
		)
		if err != nil {
			return 0, fmt.Errorf("failed to insert item %s: %w", item.ID, err)
		}
		count++
	}

	// Recreate FTS triggers
	triggers := []string{
		`CREATE TRIGGER content_items_ai AFTER INSERT ON content_items BEGIN
			INSERT INTO content_fts(rowid, title, content_text, tags)
			VALUES (new.rowid, new.title, new.content_text, new.tags);
		END`,
		`CREATE TRIGGER content_items_au AFTER UPDATE ON content_items BEGIN
			INSERT INTO content_fts(content_fts, rowid, title, content_text, tags)
			VALUES ('delete', old.rowid, old.title, old.content_text, old.tags);
			INSERT INTO content_fts(rowid, title, content_text, tags)
			VALUES (new.rowid, new.title, new.content_text, new.tags);
		END`,
	}

	for _, trigger := range triggers {
		_, err = tx.Exec(trigger)
		if err != nil {
			return 0, fmt.Errorf("failed to recreate trigger: %w", err)
		}
	}

	// Populate FTS index in one batch operation
	_, err = tx.Exec(`
	INSERT INTO content_fts(rowid, title, content_text, tags)
	SELECT rowid, title, content_text, tags FROM content_items
	WHERE rowid > COALESCE((SELECT MAX(rowid) FROM content_fts), 0)
	`)
	if err != nil {
		return 0, fmt.Errorf("failed to populate FTS index: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return count, nil
}
