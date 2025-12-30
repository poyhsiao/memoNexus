// Package db provides search performance benchmarks.
// Constitution requirement SC-002: FTS5 search must return results in <100ms for 10,000 items
// Constitution requirement SC-007: System must maintain <100ms search response time with 100K items
package db

import (
	"database/sql"
	"fmt"
	"math/rand"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

// setupBenchmarkDB creates an in-memory database for benchmarking
func setupBenchmarkDB(b *testing.B) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		b.Fatalf("Failed to open database: %v", err)
	}

	// Enable WAL mode for concurrency
	if _, err := db.Exec("PRAGMA journal_mode = WAL"); err != nil {
		b.Fatalf("Failed to enable WAL mode: %v", err)
	}

	// Enable FTS5
	if _, err := db.Exec("PRAGMA journal_mode = WAL"); err != nil {
		b.Fatalf("Failed to enable WAL: %v", err)
	}

	// Create content_items table
	if _, err := db.Exec(`
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
			updated_at INTEGER NOT NULL,
			version INTEGER NOT NULL DEFAULT 1,
			content_hash TEXT
		)
	`); err != nil {
		b.Fatalf("Failed to create content_items table: %v", err)
	}

	// Create FTS5 virtual table
	if _, err := db.Exec(`
		CREATE VIRTUAL TABLE content_items_fts USING fts5(
			title,
			content_text,
			tags,
			content=content_items,
			content_rowid=rowid,
			tokenize='porter unicode61'
		)
	`); err != nil {
		b.Fatalf("Failed to create FTS5 table: %v", err)
	}

	// Create triggers for FTS5 sync
	triggers := []string{
		`CREATE TRIGGER content_items_ai AFTER INSERT ON content_items BEGIN
			INSERT INTO content_items_fts(rowid, title, content_text, tags)
			VALUES (new.rowid, new.title, new.content_text, new.tags);
		END`,
		`CREATE TRIGGER content_items_ad AFTER DELETE ON content_items BEGIN
			INSERT INTO content_items_fts(content_items_fts, rowid, title, content_text, tags)
			VALUES ('delete', old.rowid, old.title, old.content_text, old.tags);
		END`,
		`CREATE TRIGGER content_items_au AFTER UPDATE ON content_items BEGIN
			INSERT INTO content_items_fts(content_items_fts, rowid, title, content_text, tags)
			VALUES ('delete', old.rowid, old.title, old.content_text, old.tags);
			INSERT INTO content_items_fts(rowid, title, content_text, tags)
			VALUES (new.rowid, new.title, new.content_text, new.tags);
		END`,
	}

	for _, trigger := range triggers {
		if _, err := db.Exec(trigger); err != nil {
			b.Fatalf("Failed to create trigger: %v", err)
		}
	}

	return db
}

// generateUUID creates a random UUID string
func generateUUID() string {
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		rand.Uint32(), uint16(rand.Uint32())&0x0fff, uint16(rand.Uint32())&0x0fff,
		uint16(rand.Uint32())&0x0fff, rand.Uint64()&0xffffffffffff)
}

// populateContentItems inserts n test content items
func populateContentItems(db *sql.DB, n int) error {
	stmt, err := db.Prepare(`
		INSERT INTO content_items (id, title, content_text, media_type, tags, is_deleted, created_at, updated_at, version)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare insert: %w", err)
	}
	defer stmt.Close()

	now := time.Now().Unix()
	sampleWords := []string{
		"dart", "flutter", "go", "golang", "database", "sqlite", "ffi", "mobile",
		"desktop", "api", "rest", "websocket", "sync", "offline", "performance",
		"benchmark", "testing", "code", "development", "framework", "library",
		"search", "index", "query", "storage", "memory", "concurrency", "parallel",
	}

	for i := 0; i < n; i++ {
		uuid := generateUUID()
		title := fmt.Sprintf("Test Content %d: %s %s", i,
			sampleWords[rand.Intn(len(sampleWords))],
			sampleWords[rand.Intn(len(sampleWords))])
		content := fmt.Sprintf("This is test content number %d with some keywords: %s, %s, %s",
			i,
			sampleWords[rand.Intn(len(sampleWords))],
			sampleWords[rand.Intn(len(sampleWords))],
			sampleWords[rand.Intn(len(sampleWords))])
		tags := fmt.Sprintf("%s,%s",
			sampleWords[rand.Intn(len(sampleWords))],
			sampleWords[rand.Intn(len(sampleWords))])

		if _, err := stmt.Exec(uuid, title, content, "web", tags, 0, now, now, 1); err != nil {
			return fmt.Errorf("failed to insert item %d: %w", i, err)
		}
	}

	return nil
}

// BenchmarkSearch10000Items benchmarks search performance with 10,000 items
// Constitution requirement SC-002: <100ms for 10,000 items
func BenchmarkSearch10000Items(b *testing.B) {
	db := setupBenchmarkDB(b)
	defer db.Close()

	// Populate with 10,000 items
	b.Log("Populating database with 10,000 items...")
	if err := populateContentItems(db, 10000); err != nil {
		b.Fatalf("Failed to populate items: %v", err)
	}

	// Reset timer for actual benchmark
	b.ResetTimer()

	b.Run("SimpleQuery", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			start := time.Now()
			rows, err := db.Query(`
				SELECT ci.id, ci.title, ci.content_text, ci.media_type, ci.tags
				FROM content_items ci
				INNER JOIN content_items_fts fts ON ci.rowid = fts.rowid
				WHERE content_items_fts MATCH 'dart flutter'
				AND ci.is_deleted = 0
				ORDER BY rank
				LIMIT 20
			`)
			if err != nil {
				b.Fatalf("Query failed: %v", err)
			}

			count := 0
			for rows.Next() {
				count++
			}
			rows.Close()

			elapsed := time.Since(start)
			if elapsed > 100*time.Millisecond {
				b.Errorf("Search took %v, exceeding 100ms threshold (SC-002)", elapsed)
			}

			b.ReportMetric(float64(elapsed.Milliseconds()), "ms")
		}
	})

	b.Run("ComplexQuery", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			start := time.Now()
			rows, err := db.Query(`
				SELECT ci.id, ci.title, ci.content_text, ci.media_type, ci.tags
				FROM content_items ci
				INNER JOIN content_items_fts fts ON ci.rowid = fts.rowid
				WHERE content_items_fts MATCH 'database OR search OR performance'
				AND ci.is_deleted = 0
				AND ci.media_type = 'web'
				ORDER BY rank
				LIMIT 20
			`)
			if err != nil {
				b.Fatalf("Query failed: %v", err)
			}

			count := 0
			for rows.Next() {
				count++
			}
			rows.Close()

			elapsed := time.Since(start)
			if elapsed > 100*time.Millisecond {
				b.Errorf("Search took %v, exceeding 100ms threshold (SC-002)", elapsed)
			}

			b.ReportMetric(float64(elapsed.Milliseconds()), "ms")
		}
	})
}

// BenchmarkSearch100000Items benchmarks search performance with 100,000 items
// Constitution requirement SC-007: <100ms for 100,000 items
func BenchmarkSearch100000Items(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping large benchmark in short mode")
	}

	db := setupBenchmarkDB(b)
	defer db.Close()

	// Populate with 100,000 items
	b.Log("Populating database with 100,000 items...")
	if err := populateContentItems(db, 100000); err != nil {
		b.Fatalf("Failed to populate items: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		start := time.Now()
		rows, err := db.Query(`
			SELECT ci.id, ci.title, ci.content_text, ci.media_type, ci.tags
			FROM content_items ci
			INNER JOIN content_items_fts fts ON ci.rowid = fts.rowid
			WHERE content_items_fts MATCH 'dart flutter'
			AND ci.is_deleted = 0
			ORDER BY rank
			LIMIT 20
		`)
		if err != nil {
			b.Fatalf("Query failed: %v", err)
		}

		count := 0
		for rows.Next() {
			count++
		}
		rows.Close()

		elapsed := time.Since(start)
		if elapsed > 100*time.Millisecond {
			b.Errorf("Search took %v, exceeding 100ms threshold (SC-007)", elapsed)
		}

		b.ReportMetric(float64(elapsed.Milliseconds()), "ms")
	}
}
