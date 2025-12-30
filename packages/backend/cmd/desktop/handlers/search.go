// Package handlers provides REST API handlers for search.
package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
)

// SearchHandler handles search operations using FTS5.
type SearchHandler struct {
	db *sql.DB
}

// NewSearchHandler creates a new SearchHandler.
func NewSearchHandler(db *sql.DB) *SearchHandler {
	return &SearchHandler{db: db}
}

// Search handles GET /search
func (h *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Search query 'q' is required", http.StatusBadRequest)
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	mediaType := r.URL.Query().Get("media_type")
	tags := r.URL.Query().Get("tags")
	dateFrom := r.URL.Query().Get("date_from")
	dateTo := r.URL.Query().Get("date_to")

	// Build FTS5 query with BM25 ranking
	sqlQuery := `
		SELECT
			ci.id, ci.title, ci.content_text, ci.source_url, ci.media_type,
			ci.tags, ci.summary, ci.created_at, ci.updated_at, ci.version,
			bm25(content_fts) as relevance
		FROM content_items ci
		INNER JOIN content_fts ON ci.rowid = content_fts.rowid
		WHERE content_fts MATCH ? AND ci.is_deleted = 0
	`

	args := []interface{}{query}

	// Add optional filters
	if mediaType != "" {
		sqlQuery += " AND ci.media_type = ?"
		args = append(args, mediaType)
	}
	if tags != "" {
		sqlQuery += " AND ci.tags LIKE ?"
		args = append(args, "%"+tags+"%")
	}
	if dateFrom != "" {
		sqlQuery += " AND ci.created_at >= ?"
		args = append(args, dateFrom)
	}
	if dateTo != "" {
		sqlQuery += " AND ci.created_at <= ?"
		args = append(args, dateTo)
	}

	sqlQuery += " ORDER BY bm25(content_fts) LIMIT ?"
	args = append(args, limit)

	rows, err := h.db.Query(sqlQuery, args...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var id, title, contentText, mediaType, tags string
		var sourceURL, summary sql.NullString
		var createdAt, updatedAt, version int64
		var relevance float64

		err := rows.Scan(
			&id, &title, &contentText, &sourceURL, &mediaType,
			&tags, &summary, &createdAt, &updatedAt, &version,
			&relevance,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		item := map[string]interface{}{
			"id":          id,
			"title":       title,
			"content_text": contentText,
			"media_type":  mediaType,
			"tags":        tags,
			"created_at":  createdAt,
			"updated_at":  updatedAt,
			"version":     version,
		}

		if sourceURL.Valid {
			item["source_url"] = sourceURL.String
		}
		if summary.Valid {
			item["summary"] = summary.String
		}

		// Extract matched terms (simplified - parse FTS5 matchinfo in production)
		matchedTerms := []string{query} // TODO: Parse actual matched terms from FTS5

		result := map[string]interface{}{
			"item":         item,
			"relevance":    relevance,
			"matched_terms": matchedTerms,
		}

		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"results": results,
		"total":   len(results),
		"query":   query,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
