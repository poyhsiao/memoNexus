// Package main provides the FFI bridge for mobile platforms.
// Build as shared library: libmemonexus.so (Android) / memonexus.framework (iOS)
// +build !linux

package main

/*
#cgo CFLAGS: -Wall -Wextra
#cgo LDFLAGS: -shared
#include <stdlib.h>
#include <string.h>
*/
import "C"
import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"unsafe"

	"github.com/kimhsiao/memonexus/backend/internal/analysis"
	"github.com/kimhsiao/memonexus/backend/internal/db"
	"github.com/kimhsiao/memonexus/backend/internal/models"
)

var (
	once     sync.Once
	repo     *db.Repository
	database *db.DB
	lastErr  string
	lastMu   sync.RWMutex
)

//export Init
// Init initializes the MemoNexus Core.
func Init() {
	once.Do(func() {
		// Open database at default location
		var err error
		database, err = db.Open("./data")
		if err != nil {
			setLastError(fmt.Sprintf("Failed to open database: %v", err))
			return
		}

		// Run migrations
		migrator := db.NewMigrator(database.DB, "./internal/db/migrations")
		if err := migrator.Initialize(); err != nil {
			setLastError(fmt.Sprintf("Failed to initialize migrator: %v", err))
			return
		}

		if err := migrator.Up(); err != nil {
			setLastError(fmt.Sprintf("Failed to apply migrations: %v", err))
			return
		}

		// Create repository
		repo = db.NewRepository(database.DB)
	})
}

//export Cleanup
// Cleanup cleans up resources.
func Cleanup() {
	if database != nil {
		database.Close()
	}
}

//export GetLastError
// GetLastError returns the last error message.
// Returns a C string that must be freed by the caller.
func GetLastError() *C.char {
	lastMu.RLock()
	defer lastMu.RUnlock()

	return C.CString(lastErr)
}

func setLastError(err string) {
	lastMu.Lock()
	defer lastMu.Unlock()
	lastErr = err
}

// =====================================================
// Content Operations
// =====================================================

//export ContentCreate
// ContentCreate creates a new content item.
// Returns JSON string that must be freed by the caller.
func ContentCreate(title, contentText, sourceURL, mediaType, tags, contentHash *C.char) *C.char {
	if repo == nil {
		setLastError("Repository not initialized")
		return nil
	}

	// Convert C strings to Go strings
	item := &models.ContentItem{
		Title:       C.GoString(title),
		ContentText: C.GoString(contentText),
		SourceURL:   C.GoString(sourceURL),
		MediaType:   C.GoString(mediaType),
		Tags:        C.GoString(tags),
		ContentHash: C.GoString(contentHash),
	}

	// Create item
	if err := repo.CreateContentItem(item); err != nil {
		setLastError(fmt.Sprintf("Failed to create content: %v", err))
		return nil
	}

	// Serialize to JSON
	data, err := json.Marshal(item)
	if err != nil {
		setLastError(fmt.Sprintf("Failed to serialize: %v", err))
		return nil
	}

	return C.CString(string(data))
}

//export ContentList
// ContentList lists content items with pagination.
// Returns JSON array that must be freed by the caller.
func ContentList(limit, offset int32) *C.char {
	if repo == nil {
		setLastError("Repository not initialized")
		return nil
	}

	items, err := repo.ListContentItems(int(limit), int(offset), "")
	if err != nil {
		setLastError(fmt.Sprintf("Failed to list items: %v", err))
		return nil
	}

	// Build response
	response := map[string]interface{}{
		"items": items,
		"total": len(items),
	}

	data, err := json.Marshal(response)
	if err != nil {
		setLastError(fmt.Sprintf("Failed to serialize: %v", err))
		return nil
	}

	return C.CString(string(data))
}

//export ContentGet
// ContentGet gets a content item by ID.
// Returns JSON string that must be freed by the caller.
func ContentGet(id *C.char) *C.char {
	if repo == nil {
		setLastError("Repository not initialized")
		return nil
	}

	item, err := repo.GetContentItem(C.GoString(id))
	if err != nil {
		if err == sql.ErrNoRows {
			setLastError("Content item not found")
		} else {
			setLastError(fmt.Sprintf("Failed to get item: %v", err))
		}
		return nil
	}

	data, err := json.Marshal(item)
	if err != nil {
		setLastError(fmt.Sprintf("Failed to serialize: %v", err))
		return nil
	}

	return C.CString(string(data))
}

//export ContentUpdate
// ContentUpdate updates a content item.
// Returns JSON string that must be freed by the caller.
func ContentUpdate(id *C.char, title *C.char, tags *C.char) *C.char {
	if repo == nil {
		setLastError("Repository not initialized")
		return nil
	}

	// Get existing item
	item, err := repo.GetContentItem(C.GoString(id))
	if err != nil {
		if err == sql.ErrNoRows {
			setLastError("Content item not found")
		} else {
			setLastError(fmt.Sprintf("Failed to get item: %v", err))
		}
		return nil
	}

	// Update fields
	if title != nil {
		item.Title = C.GoString(title)
	}
	if tags != nil {
		item.Tags = C.GoString(tags)
	}

	if err := repo.UpdateContentItem(item); err != nil {
		setLastError(fmt.Sprintf("Failed to update item: %v", err))
		return nil
	}

	data, err := json.Marshal(item)
	if err != nil {
		setLastError(fmt.Sprintf("Failed to serialize: %v", err))
		return nil
	}

	return C.CString(string(data))
}

//export ContentDelete
// ContentDelete deletes a content item.
// Returns 0 on success, non-zero on error.
func ContentDelete(id *C.char) int32 {
	if repo == nil {
		setLastError("Repository not initialized")
		return 1
	}

	if err := repo.DeleteContentItem(C.GoString(id)); err != nil {
		if err == sql.ErrNoRows {
			setLastError("Content item not found")
		} else {
			setLastError(fmt.Sprintf("Failed to delete item: %v", err))
		}
		return 1
	}

	return 0
}

// =====================================================
// Search Operations
// =====================================================

//export Search
// Search performs full-text search using FTS5.
// Returns JSON array that must be freed by the caller.
func Search(query *C.char, limit int32) *C.char {
	if database == nil {
		setLastError("Database not initialized")
		return nil
	}

	queryStr := C.GoString(query)

	// Build FTS5 query with BM25 ranking
	sqlQuery := `
		SELECT
			ci.id, ci.title, ci.content_text, ci.source_url, ci.media_type,
			ci.tags, ci.summary, ci.created_at, ci.updated_at, ci.version,
			bm25(content_fts) as relevance
		FROM content_items ci
		INNER JOIN content_fts ON ci.rowid = content_fts.rowid
		WHERE content_fts MATCH ? AND ci.is_deleted = 0
		ORDER BY bm25(content_fts) LIMIT ?
	`

	rows, err := database.DB.Query(sqlQuery, queryStr, limit)
	if err != nil {
		setLastError(fmt.Sprintf("Search failed: %v", err))
		return nil
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
			setLastError(fmt.Sprintf("Failed to scan row: %v", err))
			return nil
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
			"relevance":   relevance,
		}

		if sourceURL.Valid {
			item["source_url"] = sourceURL.String
		}
		if summary.Valid {
			item["summary"] = summary.String
		}

		results = append(results, item)
	}

	if err := rows.Err(); err != nil {
		setLastError(fmt.Sprintf("Failed to iterate rows: %v", err))
		return nil
	}

	// Build response
	response := map[string]interface{}{
		"results": results,
		"total":   len(results),
		"query":   queryStr,
	}

	data, err := json.Marshal(response)
	if err != nil {
		setLastError(fmt.Sprintf("Failed to serialize: %v", err))
		return nil
	}

	return C.CString(string(data))
}

// =====================================================
// Tag Operations
// =====================================================

//export TagList
// TagList lists all tags.
// Returns JSON array that must be freed by the caller.
func TagList() *C.char {
	if repo == nil {
		setLastError("Repository not initialized")
		return nil
	}

	tags, err := repo.ListTags()
	if err != nil {
		setLastError(fmt.Sprintf("Failed to list tags: %v", err))
		return nil
	}

	data, err := json.Marshal(tags)
	if err != nil {
		setLastError(fmt.Sprintf("Failed to serialize: %v", err))
		return nil
	}

	return C.CString(string(data))
}

//export TagCreate
// TagCreate creates a new tag.
// Returns JSON string that must be freed by the caller.
func TagCreate(name *C.char, color *C.char) *C.char {
	if repo == nil {
		setLastError("Repository not initialized")
		return nil
	}

	tag := &models.Tag{
		Name:  C.GoString(name),
		Color: C.GoString(color),
	}

	if err := repo.CreateTag(tag); err != nil {
		setLastError(fmt.Sprintf("Failed to create tag: %v", err))
		return nil
	}

	data, err := json.Marshal(tag)
	if err != nil {
		setLastError(fmt.Sprintf("Failed to serialize: %v", err))
		return nil
	}

	return C.CString(string(data))
}

// =====================================================
// Analysis Operations
// =====================================================

//export AnalyzeContent
// AnalyzeContent analyzes content using TF-IDF.
// Returns JSON string that must be freed by the caller.
func AnalyzeContent(contentText *C.char) *C.char {
	analyzer := analysis.NewTFIDFAnalyzer()

	result, err := analyzer.Analyze(C.GoString(contentText))
	if err != nil {
		setLastError(fmt.Sprintf("Analysis failed: %v", err))
		return nil
	}

	data, err := json.Marshal(result)
	if err != nil {
		setLastError(fmt.Sprintf("Failed to serialize: %v", err))
		return nil
	}

	return C.CString(string(data))
}

// =====================================================
// Memory Management Helpers
// =====================================================

//export FreeString
// FreeString frees a string allocated by Go.
func FreeString(ptr *C.char) {
	if ptr != nil {
		C.free(unsafe.Pointer(ptr))
	}
}

func main() {
	// Main entry point for shared library
	// Not used when loaded as library
}
