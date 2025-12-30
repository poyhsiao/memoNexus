// Package db provides FTS5 search functionality for content items.
package db

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/kimhsiao/memonexus/backend/internal/models"
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
