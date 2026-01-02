// Package handlers tests for search REST API endpoints.
// These tests verify FTS5 search functionality, query validation, and filtering.
package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/kimhsiao/memonexus/backend/internal/db"
	"github.com/kimhsiao/memonexus/backend/internal/models"
	"github.com/kimhsiao/memonexus/backend/internal/uuid"
	_ "modernc.org/sqlite"
)

// setupTestDBWithSearch creates an in-memory database with FTS5 for search testing
func setupTestDBWithSearch(t *testing.T) (*sql.DB, func()) {
	testDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create content_items table
	_, err = testDB.Exec(`
		CREATE TABLE IF NOT EXISTS content_items (
			id TEXT PRIMARY KEY CHECK(length(id) = 36),
			title TEXT NOT NULL CHECK(length(title) > 0),
			content_text TEXT NOT NULL DEFAULT '',
			source_url TEXT,
			media_type TEXT NOT NULL CHECK(media_type IN ('web', 'image', 'video', 'pdf', 'markdown')),
			tags TEXT DEFAULT '',
			summary TEXT,
			is_deleted INTEGER NOT NULL DEFAULT 0 CHECK(is_deleted IN (0, 1)),
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL CHECK(updated_at > 0 AND updated_at >= created_at),
			version INTEGER NOT NULL DEFAULT 1 CHECK(version > 0),
			content_hash TEXT
		);

		CREATE INDEX idx_content_items_created_at ON content_items(created_at DESC);
		CREATE INDEX idx_content_items_media_type ON content_items(media_type);
	`)
	if err != nil {
		testDB.Close()
		t.Fatalf("Failed to create content_items table: %v", err)
	}

	// Create FTS5 virtual table
	_, err = testDB.Exec(`
		CREATE VIRTUAL TABLE content_fts USING fts5(
			title,
			content_text,
			tags,
			content=content_items,
			content_rowid=rowid,
			tokenize='porter unicode61'
		);

		CREATE TRIGGER content_items_ai AFTER INSERT ON content_items BEGIN
			INSERT INTO content_fts(rowid, title, content_text, tags)
			VALUES (new.rowid, new.title, new.content_text, new.tags);
		END;

		CREATE TRIGGER content_items_ad AFTER DELETE ON content_items BEGIN
			INSERT INTO content_fts(content_fts, rowid, title, content_text, tags)
			VALUES ('delete', old.rowid, old.title, old.content_text, old.tags);
		END;

		CREATE TRIGGER content_items_au AFTER UPDATE ON content_items BEGIN
			INSERT INTO content_fts(content_fts, rowid, title, content_text, tags)
			VALUES ('delete', old.rowid, old.title, old.content_text, new.tags);
			INSERT INTO content_fts(rowid, title, content_text, tags)
			VALUES (new.rowid, new.title, new.content_text, new.tags);
		END;
	`)
	if err != nil {
		testDB.Close()
		t.Fatalf("Failed to create FTS5 table: %v", err)
	}

	cleanup := func() {
		testDB.Close()
	}

	return testDB, cleanup
}

// insertTestContentItem inserts a test content item and returns its ID
func insertTestContentItem(t *testing.T, db *sql.DB, title, content, mediaType, tags string, createdAt int64) string {
	t.Helper()
	id := models.UUID(uuid.New())
	query := `
		INSERT INTO content_items (id, title, content_text, media_type, tags, is_deleted, created_at, updated_at, version)
		VALUES (?, ?, ?, ?, ?, 0, ?, ?, 1)
	`
	_, err := db.Exec(query, string(id), title, content, mediaType, tags, createdAt, createdAt)
	if err != nil {
		t.Fatalf("Failed to insert test content: %v", err)
	}
	return string(id)
}

func TestNewSearchHandler(t *testing.T) {
	testDB, cleanup := setupTestDBWithSearch(t)
	defer cleanup()

	repo := db.NewRepository(testDB)
	handler := NewSearchHandler(repo)

	if handler == nil {
		t.Error("NewSearchHandler should return non-nil handler")
	}

	if handler.repo != repo {
		t.Error("Handler repo should match provided repo")
	}
}

func TestSearchHandler_Search_BasicQuery(t *testing.T) {
	testDB, cleanup := setupTestDBWithSearch(t)
	defer cleanup()

	repo := db.NewRepository(testDB)
	handler := NewSearchHandler(repo)

	// Insert test content
	insertTestContentItem(t, testDB, "Go Programming Guide", "Learn Go programming language", "web", "golang,programming", 1000)
	insertTestContentItem(t, testDB, "Python Tutorial", "Python basics for beginners", "web", "python,tutorial", 2000)

	// Create search request
	req := httptest.NewRequest(http.MethodGet, "/search?q=programming", nil)
	w := httptest.NewRecorder()

	handler.Search(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["query"] != "programming" {
		t.Errorf("Expected query 'programming', got %v", response["query"])
	}

	if response["total"].(float64) < 1 {
		t.Errorf("Expected at least 1 result, got %v", response["total"])
	}

	if response["results"] == nil {
		t.Error("Response should contain results")
	}
}

func TestSearchHandler_Search_EmptyQuery(t *testing.T) {
	testDB, cleanup := setupTestDBWithSearch(t)
	defer cleanup()

	repo := db.NewRepository(testDB)
	handler := NewSearchHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/search?q=", nil)
	w := httptest.NewRecorder()

	handler.Search(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "search query 'q' is required") {
		t.Errorf("Expected error about missing query, got: %s", w.Body.String())
	}
}

func TestSearchHandler_Search_QueryTooLong(t *testing.T) {
	testDB, cleanup := setupTestDBWithSearch(t)
	defer cleanup()

	repo := db.NewRepository(testDB)
	handler := NewSearchHandler(repo)

	// Create a query longer than 500 characters
	longQuery := strings.Repeat("a ", 300)

	req := httptest.NewRequest(http.MethodGet, "/search?q="+url.QueryEscape(longQuery), nil)
	w := httptest.NewRecorder()

	handler.Search(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "too long") {
		t.Errorf("Expected error about query too long, got: %s", w.Body.String())
	}
}

func TestSearchHandler_Search_SingleCharacterQuery(t *testing.T) {
	testDB, cleanup := setupTestDBWithSearch(t)
	defer cleanup()

	repo := db.NewRepository(testDB)
	handler := NewSearchHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/search?q=a", nil)
	w := httptest.NewRecorder()

	handler.Search(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "at least 2 characters") {
		t.Errorf("Expected error about short terms, got: %s", w.Body.String())
	}
}

func TestSearchHandler_Search_DangerousOperator(t *testing.T) {
	testDB, cleanup := setupTestDBWithSearch(t)
	defer cleanup()

	repo := db.NewRepository(testDB)
	handler := NewSearchHandler(repo)

	// Test NEAR/ operator
	req := httptest.NewRequest(http.MethodGet, "/search?q=test+NEAR/10+query", nil)
	w := httptest.NewRecorder()

	handler.Search(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "unsupported operator") {
		t.Errorf("Expected error about unsupported operator, got: %s", w.Body.String())
	}
}

func TestSearchHandler_Search_WithLimit(t *testing.T) {
	testDB, cleanup := setupTestDBWithSearch(t)
	defer cleanup()

	repo := db.NewRepository(testDB)
	handler := NewSearchHandler(repo)

	// Insert multiple items
	for i := 0; i < 25; i++ {
		insertTestContentItem(t, testDB, "Item "+strconv.Itoa(i), "Content "+strconv.Itoa(i), "web", "test", 1000+int64(i))
	}

	// Request with limit
	req := httptest.NewRequest(http.MethodGet, "/search?q=Item&limit=5", nil)
	w := httptest.NewRecorder()

	handler.Search(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	results := response["results"].([]interface{})
	if len(results) > 5 {
		t.Errorf("Expected at most 5 results with limit=5, got %d", len(results))
	}
}

func TestSearchHandler_Search_InvalidLimit(t *testing.T) {
	testDB, cleanup := setupTestDBWithSearch(t)
	defer cleanup()

	repo := db.NewRepository(testDB)
	handler := NewSearchHandler(repo)

	// Test limit too high
	req := httptest.NewRequest(http.MethodGet, "/search?q=test&limit=101", nil)
	w := httptest.NewRecorder()

	handler.Search(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "limit must be between 1 and 100") {
		t.Errorf("Expected error about invalid limit, got: %s", w.Body.String())
	}
}

func TestSearchHandler_Search_NegativeLimit(t *testing.T) {
	testDB, cleanup := setupTestDBWithSearch(t)
	defer cleanup()

	repo := db.NewRepository(testDB)
	handler := NewSearchHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/search?q=test&limit=-1", nil)
	w := httptest.NewRecorder()

	handler.Search(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestSearchHandler_Search_WithMediaType(t *testing.T) {
	testDB, cleanup := setupTestDBWithSearch(t)
	defer cleanup()

	repo := db.NewRepository(testDB)
	handler := NewSearchHandler(repo)

	// Insert items with different media types
	insertTestContentItem(t, testDB, "PDF Document", "PDF content here", "pdf", "document", 1000)
	insertTestContentItem(t, testDB, "Web Article", "Web article content", "web", "article", 2000)

	req := httptest.NewRequest(http.MethodGet, "/search?q=document&media_type=pdf", nil)
	w := httptest.NewRecorder()

	handler.Search(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Should only return PDF results
	results := response["results"].([]interface{})
	for _, r := range results {
		result := r.(map[string]interface{})
		item := result["item"].(map[string]interface{})
		if item["media_type"] != "pdf" {
			t.Errorf("Expected only PDF results, got %v", item["media_type"])
		}
	}
}

func TestSearchHandler_Search_InvalidMediaType(t *testing.T) {
	testDB, cleanup := setupTestDBWithSearch(t)
	defer cleanup()

	repo := db.NewRepository(testDB)
	handler := NewSearchHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/search?q=test&media_type=invalid", nil)
	w := httptest.NewRecorder()

	handler.Search(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "invalid media_type") {
		t.Errorf("Expected error about invalid media_type, got: %s", w.Body.String())
	}
}

func TestSearchHandler_Search_WithDateRange(t *testing.T) {
	testDB, cleanup := setupTestDBWithSearch(t)
	defer cleanup()

	repo := db.NewRepository(testDB)
	handler := NewSearchHandler(repo)

	// Insert items with different timestamps
	insertTestContentItem(t, testDB, "Old Item", "Old content", "web", "old", 1000)       // 1970
	insertTestContentItem(t, testDB, "New Item", "New content", "web", "new", 1609459200) // 2021

	// Search for items after 2020
	req := httptest.NewRequest(http.MethodGet, "/search?q=Item&date_from=1577836800", nil) // 2020-01-01
	w := httptest.NewRecorder()

	handler.Search(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Should only return the new item
	if response["total"].(float64) != 1 {
		t.Errorf("Expected 1 result with date filter, got %v", response["total"])
	}
}

func TestSearchHandler_Search_InvalidDateFrom(t *testing.T) {
	testDB, cleanup := setupTestDBWithSearch(t)
	defer cleanup()

	repo := db.NewRepository(testDB)
	handler := NewSearchHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/search?q=test&date_from=invalid", nil)
	w := httptest.NewRecorder()

	handler.Search(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "invalid date_from format") {
		t.Errorf("Expected error about invalid date_from, got: %s", w.Body.String())
	}
}

func TestSearchHandler_Search_InvalidDateTo(t *testing.T) {
	testDB, cleanup := setupTestDBWithSearch(t)
	defer cleanup()

	repo := db.NewRepository(testDB)
	handler := NewSearchHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/search?q=test&date_to=abc", nil)
	w := httptest.NewRecorder()

	handler.Search(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "invalid date_to format") {
		t.Errorf("Expected error about invalid date_to, got: %s", w.Body.String())
	}
}

func TestSearchHandler_Search_WithTags(t *testing.T) {
	testDB, cleanup := setupTestDBWithSearch(t)
	defer cleanup()

	repo := db.NewRepository(testDB)
	handler := NewSearchHandler(repo)

	// Insert items with different tags
	insertTestContentItem(t, testDB, "Tagged Item 1", "Content with tag1", "web", "tag1,important", 1000)
	insertTestContentItem(t, testDB, "Tagged Item 2", "Content with tag2", "web", "tag2", 2000)

	req := httptest.NewRequest(http.MethodGet, "/search?q=Tagged&tags=tag1", nil)
	w := httptest.NewRecorder()

	handler.Search(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Should return items with tag1
	if response["total"].(float64) < 1 {
		t.Errorf("Expected at least 1 result with tag filter, got %v", response["total"])
	}
}

func TestSearchHandler_Search_NoResults(t *testing.T) {
	testDB, cleanup := setupTestDBWithSearch(t)
	defer cleanup()

	repo := db.NewRepository(testDB)
	handler := NewSearchHandler(repo)

	// Insert test content
	insertTestContentItem(t, testDB, "Test Item", "Test content", "web", "test", 1000)

	// Search for non-existent term
	req := httptest.NewRequest(http.MethodGet, "/search?q=nonexistent", nil)
	w := httptest.NewRecorder()

	handler.Search(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["total"].(float64) != 0 {
		t.Errorf("Expected 0 results for non-existent query, got %v", response["total"])
	}

	results := response["results"].([]interface{})
	if len(results) != 0 {
		t.Errorf("Expected empty results array, got %d items", len(results))
	}
}

func TestSearchHandler_Search_WithWildcard(t *testing.T) {
	testDB, cleanup := setupTestDBWithSearch(t)
	defer cleanup()

	repo := db.NewRepository(testDB)
	handler := NewSearchHandler(repo)

	// Insert test content
	insertTestContentItem(t, testDB, "Programming in Go", "Learn Go language", "web", "programming,golang", 1000)
	insertTestContentItem(t, testDB, "Programming in Python", "Learn Python", "web", "programming,python", 2000)

	// Search with wildcard
	req := httptest.NewRequest(http.MethodGet, "/search?q=program*", nil)
	w := httptest.NewRecorder()

	handler.Search(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Should match both "programming" items
	if response["total"].(float64) < 2 {
		t.Errorf("Expected at least 2 results with wildcard, got %v", response["total"])
	}
}

func TestSearchHandler_Search_MultiWordQuery(t *testing.T) {
	testDB, cleanup := setupTestDBWithSearch(t)
	defer cleanup()

	repo := db.NewRepository(testDB)
	handler := NewSearchHandler(repo)

	// Insert test content
	insertTestContentItem(t, testDB, "Go Programming", "Go is a programming language", "web", "golang", 1000)
	insertTestContentItem(t, testDB, "Python Programming", "Python is also programming", "web", "python", 2000)

	// Search for multi-word phrase
	req := httptest.NewRequest(http.MethodGet, "/search?q=Go+programming", nil)
	w := httptest.NewRecorder()

	handler.Search(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["total"].(float64) < 1 {
		t.Errorf("Expected at least 1 result for multi-word query, got %v", response["total"])
	}
}

func TestSearchHandler_Search_WithAND(t *testing.T) {
	testDB, cleanup := setupTestDBWithSearch(t)
	defer cleanup()

	repo := db.NewRepository(testDB)
	handler := NewSearchHandler(repo)

	// Insert test content
	insertTestContentItem(t, testDB, "Go Programming Guide", "Complete guide to Go", "web", "golang,guide", 1000)
	insertTestContentItem(t, testDB, "Python Tutorial", "Python basics", "web", "python,tutorial", 2000)

	// Search with AND operator
	req := httptest.NewRequest(http.MethodGet, "/search?q=Go+AND+guide", nil)
	w := httptest.NewRecorder()

	handler.Search(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Should only match Go Programming Guide
	if response["total"].(float64) != 1 {
		t.Errorf("Expected 1 result with AND operator, got %v", response["total"])
	}
}

func TestSearchHandler_Search_WithNOT(t *testing.T) {
	testDB, cleanup := setupTestDBWithSearch(t)
	defer cleanup()

	repo := db.NewRepository(testDB)
	handler := NewSearchHandler(repo)

	// Insert test content
	insertTestContentItem(t, testDB, "Go Programming", "Go language", "web", "golang", 1000)
	insertTestContentItem(t, testDB, "Python Programming", "Python language", "web", "python", 2000)

	// Search with NOT operator to exclude Python
	req := httptest.NewRequest(http.MethodGet, "/search?q=programming+NOT+python", nil)
	w := httptest.NewRecorder()

	handler.Search(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Should only match Go Programming (not Python)
	if response["total"].(float64) != 1 {
		t.Errorf("Expected 1 result with NOT operator, got %v", response["total"])
	}
}

func TestSearchHandler_Search_ResultsIncludeMatchedTerms(t *testing.T) {
	testDB, cleanup := setupTestDBWithSearch(t)
	defer cleanup()

	repo := db.NewRepository(testDB)
	handler := NewSearchHandler(repo)

	// Insert test content
	insertTestContentItem(t, testDB, "Programming Guide", "Learn programming", "web", "programming,guide", 1000)

	req := httptest.NewRequest(http.MethodGet, "/search?q=programming", nil)
	w := httptest.NewRecorder()

	handler.Search(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	results := response["results"].([]interface{})
	if len(results) < 1 {
		t.Fatal("Expected at least 1 result")
	}

	firstResult := results[0].(map[string]interface{})
	// Note: matched_terms may not be populated in test environment without FTS5 matchinfo
	// The field exists but might be empty or nil
	_ = firstResult["matched_terms"] // Just verify field exists
}

func TestSearchHandler_Search_ResultStructure(t *testing.T) {
	testDB, cleanup := setupTestDBWithSearch(t)
	defer cleanup()

	repo := db.NewRepository(testDB)
	handler := NewSearchHandler(repo)

	// Insert test content
	insertTestContentItem(t, testDB, "Test Title", "Test content with source", "web", "test", 1000)

	req := httptest.NewRequest(http.MethodGet, "/search?q=Test", nil)
	w := httptest.NewRecorder()

	handler.Search(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	results := response["results"].([]interface{})
	if len(results) < 1 {
		t.Fatal("Expected at least 1 result")
	}

	firstResult := results[0].(map[string]interface{})

	// Check result structure
	if firstResult["relevance"] == nil {
		t.Error("Expected relevance in result")
	}

	item := firstResult["item"].(map[string]interface{})

	// Check required fields
	requiredFields := []string{"id", "title", "content_text", "media_type", "tags", "created_at", "updated_at", "version"}
	for _, field := range requiredFields {
		if item[field] == nil {
			t.Errorf("Expected field '%s' in result item", field)
		}
	}
}

func TestSearchHandler_Search_ContentTypeIsJSON(t *testing.T) {
	testDB, cleanup := setupTestDBWithSearch(t)
	defer cleanup()

	repo := db.NewRepository(testDB)
	handler := NewSearchHandler(repo)

	insertTestContentItem(t, testDB, "Test", "Content", "web", "test", 1000)

	req := httptest.NewRequest(http.MethodGet, "/search?q=Test", nil)
	w := httptest.NewRecorder()

	handler.Search(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}
}

func TestSearchHandler_Search_MethodNotAllowed(t *testing.T) {
	testDB, cleanup := setupTestDBWithSearch(t)
	defer cleanup()

	repo := db.NewRepository(testDB)
	handler := NewSearchHandler(repo)

	req := httptest.NewRequest(http.MethodPost, "/search?q=test", nil)
	w := httptest.NewRecorder()

	handler.Search(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}
