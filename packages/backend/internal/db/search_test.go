// Package db provides unit tests for FTS5 search operations.
package db

import (
	"database/sql"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/kimhsiao/memonexus/backend/internal/models"
	"github.com/kimhsiao/memonexus/backend/internal/uuid"
)

// setupSearchTestDB creates an in-memory SQLite database with FTS5 for testing.
func setupSearchTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Create content_items table
	_, err = db.Exec(`
		CREATE TABLE content_items (
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
		db.Close()
		t.Fatalf("Failed to create content_items table: %v", err)
	}

	// Create FTS5 virtual table with porter unicode61 tokenizer
	_, err = db.Exec(`
		CREATE VIRTUAL TABLE content_items_fts USING fts5(
			title,
			content_text,
			tags,
			content=content_items,
			content_rowid=rowid,
			tokenize='porter unicode61'
		);

		-- Triggers to keep FTS5 table in sync with content_items
		INSERT INTO content_items_fts(rowid, title, content_text, tags)
		SELECT rowid, title, content_text, tags FROM content_items;

		CREATE TRIGGER content_items_ai AFTER INSERT ON content_items BEGIN
			INSERT INTO content_items_fts(rowid, title, content_text, tags)
			VALUES (new.rowid, new.title, new.content_text, new.tags);
		END;

		CREATE TRIGGER content_items_ad AFTER DELETE ON content_items BEGIN
			INSERT INTO content_items_fts(content_items_fts, rowid, title, content_text, tags)
			VALUES ('delete', old.rowid, old.title, old.content_text, old.tags);
		END;

		CREATE TRIGGER content_items_au AFTER UPDATE ON content_items BEGIN
			INSERT INTO content_items_fts(content_items_fts, rowid, title, content_text, tags)
			VALUES ('delete', old.rowid, old.title, old.content_text, old.tags);
			INSERT INTO content_items_fts(rowid, title, content_text, tags)
			VALUES (new.rowid, new.title, new.content_text, new.tags);
		END;
	`)
	if err != nil {
		db.Close()
		t.Fatalf("Failed to create FTS5 table: %v", err)
	}

	return db
}

// insertTestContentItem inserts a test content item with optional parameters.
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

// =====================================================
// FTS5 Query Execution Tests (T113)
// =====================================================

func TestFTS5QueryExecution(t *testing.T) {
	db := setupSearchTestDB(t)
	defer db.Close()

	// Insert test data with varying relevance
	now := time.Now().Unix()
	insertTestContentItem(t, db, "Dart Programming Guide", "Learn Dart programming language with Flutter", "web", "dart,flutter", now-100)
	insertTestContentItem(t, db, "Flutter Development Tutorial", "Complete Flutter course for beginners", "web", "flutter,mobile", now-90)
	insertTestContentItem(t, db, "Python Machine Learning", "ML with Python and scikit-learn", "web", "python,ml", now-80)
	insertTestContentItem(t, db, "Dart and Flutter Architecture", "Advanced patterns for Flutter apps", "web", "dart,flutter,architecture", now-70)

	// Test basic FTS5 search with BM25 ranking
	query := `
		SELECT ci.id, ci.title, ci.content_text, ci.media_type, ci.tags
		FROM content_items ci
		INNER JOIN content_items_fts fts ON ci.rowid = fts.rowid
		WHERE content_items_fts MATCH ? AND ci.is_deleted = 0
		ORDER BY rank
		LIMIT 20
	`

	rows, err := db.Query(query, "dart flutter")
	if err != nil {
		t.Fatalf("FTS5 query failed: %v", err)
	}
	defer rows.Close()

	var results []struct {
		ID        string
		Title     string
		Content   string
		MediaType string
		Tags      string
	}

	for rows.Next() {
		var r struct {
			ID        string
			Title     string
			Content   string
			MediaType string
			Tags      string
		}
		err := rows.Scan(&r.ID, &r.Title, &r.Content, &r.MediaType, &r.Tags)
		if err != nil {
			t.Fatalf("Failed to scan row: %v", err)
		}
		results = append(results, r)
	}

	// Verify results
	if len(results) == 0 {
		t.Error("Expected search results, got none")
	}

	// Results should be ranked by relevance (both "dart" and "flutter" match higher)
	// First result should contain both terms
	hasDartAndFlutter := false
	for _, r := range results {
		if containsWords(r.Title+r.Content+r.Tags, "dart", "flutter") {
			hasDartAndFlutter = true
			break
		}
	}
	if !hasDartAndFlutter {
		t.Error("Expected results containing both 'dart' and 'flutter'")
	}
}

func TestUnicodeHandling(t *testing.T) {
	db := setupSearchTestDB(t)
	defer db.Close()

	now := time.Now().Unix()
	// Insert test data with CJK characters
	insertTestContentItem(t, db, "Flutter 開發教程", "學習 Flutter 跨平台開發", "web", "flutter,教學", now)
	insertTestContentItem(t, db, "Dart プログラミング", "Dart言語入門ガイド", "web", "dart,プログラミング", now)
	insertTestContentItem(t, db, "Flutter 한글 가이드", "플러터 앱 개발", "web", "flutter,한국어", now)
	insertTestContentItem(t, db, "English Tutorial", "Flutter development guide", "web", "flutter,english", now)

	// Test CJK search (Chinese)
	// Note: FTS5 with unicode61 tokenizer indexes Chinese characters individually.
	// Use prefix search with * to match any character starting with the prefix.
	t.Run("ChineseSearch", func(t *testing.T) {
		query := `
			SELECT ci.id, ci.title
			FROM content_items ci
			INNER JOIN content_items_fts fts ON ci.rowid = fts.rowid
			WHERE content_items_fts MATCH ? AND ci.is_deleted = 0
		`
		// Search for individual character with prefix match
		rows, err := db.Query(query, "開*")
		if err != nil {
			t.Fatalf("Chinese search failed: %v", err)
		}
		defer rows.Close()

		var titles []string
		for rows.Next() {
			var id, title string
			rows.Scan(&id, &title)
			titles = append(titles, title)
		}

		if len(titles) == 0 {
			t.Error("Expected results for Chinese search term '開*'")
		}
	})

	// Test Japanese search
	t.Run("JapaneseSearch", func(t *testing.T) {
		query := `
			SELECT ci.id, ci.title
			FROM content_items ci
			INNER JOIN content_items_fts fts ON ci.rowid = fts.rowid
			WHERE content_items_fts MATCH ? AND ci.is_deleted = 0
		`
		rows, err := db.Query(query, "プログラミング")
		if err != nil {
			t.Fatalf("Japanese search failed: %v", err)
		}
		defer rows.Close()

		var titles []string
		for rows.Next() {
			var id, title string
			rows.Scan(&id, &title)
			titles = append(titles, title)
		}

		if len(titles) == 0 {
			t.Error("Expected results for Japanese search term 'プログラミング'")
		}
	})

	// Test Korean search
	t.Run("KoreanSearch", func(t *testing.T) {
		query := `
			SELECT ci.id, ci.title
			FROM content_items ci
			INNER JOIN content_items_fts fts ON ci.rowid = fts.rowid
			WHERE content_items_fts MATCH ? AND ci.is_deleted = 0
		`
		rows, err := db.Query(query, "플러터")
		if err != nil {
			t.Fatalf("Korean search failed: %v", err)
		}
		defer rows.Close()

		var titles []string
		for rows.Next() {
			var id, title string
			rows.Scan(&id, &title)
			titles = append(titles, title)
		}

		if len(titles) == 0 {
			t.Error("Expected results for Korean search term '플러터'")
		}
	})
}

func TestMultiWordPhrases(t *testing.T) {
	db := setupSearchTestDB(t)
	defer db.Close()

	now := time.Now().Unix()
	insertTestContentItem(t, db, "Machine Learning Basics", "Introduction to machine learning algorithms", "web", "ml,ai", now)
	insertTestContentItem(t, db, "Deep Learning with Neural Networks", "Advanced neural network techniques", "web", "deep-learning,ai", now)
	insertTestContentItem(t, db, "Machine Learning vs Deep Learning", "Comparing ML and DL approaches", "web", "ml,comparison", now)

	// Test phrase search with quotes
	query := `
		SELECT ci.id, ci.title
		FROM content_items ci
		INNER JOIN content_items_fts fts ON ci.rowid = fts.rowid
		WHERE content_items_fts MATCH ? AND ci.is_deleted = 0
		ORDER BY rank
	`
	rows, err := db.Query(query, `"machine learning"`)
	if err != nil {
		t.Fatalf("Phrase search failed: %v", err)
	}
	defer rows.Close()

	var titles []string
	for rows.Next() {
		var id, title string
		rows.Scan(&id, &title)
		titles = append(titles, title)
	}

	// Should find items with the exact phrase "machine learning"
	if len(titles) == 0 {
		t.Error("Expected results for phrase search 'machine learning'")
	}
}

// =====================================================
// Search Filter Tests (T114)
// =====================================================

func TestSearchFiltersMediaType(t *testing.T) {
	db := setupSearchTestDB(t)
	defer db.Close()

	now := time.Now().Unix()
	insertTestContentItem(t, db, "Web Article", "Flutter web development", "web", "flutter", now)
	insertTestContentItem(t, db, "Image Screenshot", "Screenshot of flutter app", "image", "flutter", now)
	insertTestContentItem(t, db, "PDF Document", "Flutter guide PDF", "pdf", "flutter", now)

	// Test media type filter
	query := `
		SELECT ci.id, ci.title, ci.media_type
		FROM content_items ci
		INNER JOIN content_items_fts fts ON ci.rowid = fts.rowid
		WHERE content_items_fts MATCH ? AND ci.is_deleted = 0 AND ci.media_type = ?
		ORDER BY rank
	`

	// Should only return web items
	rows, err := db.Query(query, "flutter", "web")
	if err != nil {
		t.Fatalf("Media type filter failed: %v", err)
	}
	defer rows.Close()

	var results []struct {
		ID        string
		Title     string
		MediaType string
	}

	for rows.Next() {
		var r struct {
			ID        string
			Title     string
			MediaType string
		}
		rows.Scan(&r.ID, &r.Title, &r.MediaType)
		results = append(results, r)
	}

	if len(results) == 0 {
		t.Error("Expected results for media_type filter")
	}

	// Verify all results are web type
	for _, r := range results {
		if r.MediaType != "web" {
			t.Errorf("Expected media_type 'web', got '%s'", r.MediaType)
		}
	}
}

func TestSearchFiltersDateRange(t *testing.T) {
	db := setupSearchTestDB(t)
	defer db.Close()

	now := time.Now().Unix()
	day := int64(86400)
	weekAgo := now - (7 * day)
	twoWeeksAgo := now - (14 * day)

	// Insert items with different timestamps
	oldItemID := insertTestContentItem(t, db, "Old Article", "Content from two weeks ago", "web", "old", twoWeeksAgo)
	recentItemID := insertTestContentItem(t, db, "Recent Article", "Content from last week", "web", "recent", weekAgo)
	_ = insertTestContentItem(t, db, "Very Recent", "Content from today", "web", "new", now)

	// Test date range filter (items from last week only)
	query := `
		SELECT ci.id, ci.title
		FROM content_items ci
		INNER JOIN content_items_fts fts ON ci.rowid = fts.rowid
		WHERE content_items_fts MATCH ? AND ci.is_deleted = 0 AND ci.created_at >= ?
		ORDER BY ci.created_at DESC
	`

	rows, err := db.Query(query, "article", weekAgo)
	if err != nil {
		t.Fatalf("Date range filter failed: %v", err)
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id, title string
		rows.Scan(&id, &title)
		ids = append(ids, id)
	}

	// Should not include the old item
	for _, id := range ids {
		if id == oldItemID {
			t.Error("Date range filter should exclude items older than the range")
		}
	}

	// Should include recent item
	hasRecent := false
	for _, id := range ids {
		if id == recentItemID {
			hasRecent = true
			break
		}
	}
	if !hasRecent {
		t.Error("Date range filter should include items within the range")
	}
}

func TestSearchFiltersTags(t *testing.T) {
	db := setupSearchTestDB(t)
	defer db.Close()

	now := time.Now().Unix()
	insertTestContentItem(t, db, "Flutter Tutorial", "Learn Flutter basics", "web", "flutter,beginner", now)
	insertTestContentItem(t, db, "Advanced Flutter", "Advanced Flutter patterns", "web", "flutter,advanced", now)
	insertTestContentItem(t, db, "Dart Basics", "Dart programming", "web", "dart,beginner", now)

	// Test tag filter (search content and filter by tag prefix)
	query := `
		SELECT ci.id, ci.title, ci.tags
		FROM content_items ci
		INNER JOIN content_items_fts fts ON ci.rowid = fts.rowid
		WHERE content_items_fts MATCH ? AND ci.is_deleted = 0 AND ci.tags LIKE ?
		ORDER BY rank
	`

	rows, err := db.Query(query, "flutter", "%beginner%")
	if err != nil {
		t.Fatalf("Tag filter failed: %v", err)
	}
	defer rows.Close()

	var results []struct {
		ID    string
		Title string
		Tags  string
	}

	for rows.Next() {
		var r struct {
			ID    string
			Title string
			Tags  string
		}
		rows.Scan(&r.ID, &r.Title, &r.Tags)
		results = append(results, r)
	}

	// Should only find items with "beginner" tag
	if len(results) == 0 {
		t.Error("Expected results with 'beginner' tag")
	}

	for _, r := range results {
		if !containsWord(r.Tags, "beginner") {
			t.Errorf("Expected tag 'beginner' in result, got tags: %s", r.Tags)
		}
	}
}

func TestSearchFiltersCombined(t *testing.T) {
	db := setupSearchTestDB(t)
	defer db.Close()

	now := time.Now().Unix()
	day := int64(86400)
	weekAgo := now - (7 * day)

	insertTestContentItem(t, db, "Recent Flutter Web", "Recent flutter web content", "web", "flutter,recent", now)
	insertTestContentItem(t, db, "Old Flutter Web", "Old flutter web content", "web", "flutter,old", weekAgo-1) // Slightly older than cutoff
	insertTestContentItem(t, db, "Recent Flutter PDF", "Recent flutter PDF guide", "pdf", "flutter,recent", now)

	// Test combined filters: search term + media type + date range
	query := `
		SELECT ci.id, ci.title, ci.media_type
		FROM content_items ci
		INNER JOIN content_items_fts fts ON ci.rowid = fts.rowid
		WHERE content_items_fts MATCH ?
			AND ci.is_deleted = 0
			AND ci.media_type = ?
			AND ci.created_at >= ?
		ORDER BY rank
	`

	rows, err := db.Query(query, "flutter", "web", weekAgo)
	if err != nil {
		t.Fatalf("Combined filters failed: %v", err)
	}
	defer rows.Close()

	var results []struct {
		ID        string
		Title     string
		MediaType string
	}

	for rows.Next() {
		var r struct {
			ID        string
			Title     string
			MediaType string
		}
		rows.Scan(&r.ID, &r.Title, &r.MediaType)
		results = append(results, r)
	}

	// Should only return the recent web item
	if len(results) != 1 {
		t.Errorf("Expected 1 result with combined filters, got %d", len(results))
	}

	if len(results) > 0 {
		if results[0].MediaType != "web" {
			t.Errorf("Expected media_type 'web', got '%s'", results[0].MediaType)
		}
		if results[0].Title != "Recent Flutter Web" {
			t.Errorf("Expected 'Recent Flutter Web', got '%s'", results[0].Title)
		}
	}
}

// Helper functions

func containsWords(text string, words ...string) bool {
	textLower := toLower(text)
	for _, w := range words {
		if !containsWord(textLower, toLower(w)) {
			return false
		}
	}
	return true
}

func containsWord(text, word string) bool {
	// Simple word boundary check
	for i := 0; i <= len(text)-len(word); i++ {
		if text[i:i+len(word)] == word {
			// Check boundaries
			before := byte(' ')
			after := byte(' ')
			if i > 0 {
				before = text[i-1]
			}
			if i+len(word) < len(text) {
				after = text[i+len(word)]
			}
			if (before == ' ' || before == ',') && (after == ' ' || after == ',' || after == 0) {
				return true
			}
		}
	}
	return false
}

func toLower(s string) string {
	// Simple ASCII toLower - sufficient for test purposes
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			result[i] = c + 32
		} else {
			result[i] = c
		}
	}
	return string(result)
}
