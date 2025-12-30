// Package memory provides memory profiling and leak detection tests.
// Constitution requirement: Identify memory leaks through profiling
package memory

import (
	"database/sql"
	"fmt"
	"runtime"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

// testHelper is a minimal interface for *testing.T and *testing.B
type testHelper interface {
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
}

// setupTestDB creates an in-memory database for memory profiling
func setupTestDB(t testHelper) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// Enable WAL mode
	if _, err := db.Exec("PRAGMA journal_mode = WAL"); err != nil {
		t.Fatalf("Failed to enable WAL mode: %v", err)
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
		t.Fatalf("Failed to create content_items table: %v", err)
	}

	return db
}

// getMemoryStats returns current memory statistics
func getMemoryStats() runtime.MemStats {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	return stats
}

// formatBytes formats bytes to human-readable string
func formatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}

// TestMemoryLeakSearch tests for memory leaks during repeated search operations
func TestMemoryLeakSearch(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Insert test data
	stmt, _ := db.Prepare(`
		INSERT INTO content_items (id, title, content_text, media_type, is_deleted, created_at, updated_at, version)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`)
	defer stmt.Close()

	now := time.Now().Unix()
	for i := 0; i < 1000; i++ {
		uuid := fmt.Sprintf("%08d-0000-0000-0000-000000000000", i)
		title := fmt.Sprintf("Test Content %d", i)
		content := fmt.Sprintf("This is test content number %d with some keywords", i)
		stmt.Exec(uuid, title, content, "web", 0, now, now, 1)
	}

	// Force GC before starting
	runtime.GC()
	initialStats := getMemoryStats()

	t.Log("Initial memory stats:")
	t.Logf("  Alloc: %s", formatBytes(initialStats.Alloc))
	t.Logf("  TotalAlloc: %s", formatBytes(initialStats.TotalAlloc))
	t.Logf("  Sys: %s", formatBytes(initialStats.Sys))
	t.Logf("  NumGC: %d", initialStats.NumGC)

	// Perform many search operations
	const iterations = 1000
	for i := 0; i < iterations; i++ {
		rows, _ := db.Query(`
			SELECT id, title, content_text FROM content_items
			WHERE is_deleted = 0 AND content_text LIKE '%keywords%'
			LIMIT 20
		`)

		// Always close rows to prevent connection leaks
		for rows.Next() {
			var id, title, content string
			rows.Scan(&id, &title, &content)
		}
		rows.Close()

		// Check memory every 100 iterations
		if (i+1)%100 == 0 {
			runtime.GC()
			currentStats := getMemoryStats()
			allocatedDiff := currentStats.TotalAlloc - initialStats.TotalAlloc
			allocDiff := currentStats.Alloc - initialStats.Alloc

			t.Logf("After %d iterations:", i+1)
			t.Logf("  Alloc: %s (diff: %s)", formatBytes(currentStats.Alloc), formatBytes(uint64(allocDiff)))
			t.Logf("  TotalAlloc: %s (diff: %s)", formatBytes(currentStats.TotalAlloc), formatBytes(allocatedDiff))
			t.Logf("  Sys: %s", formatBytes(currentStats.Sys))

			// Check for potential leak (Alloc should not grow unbounded)
			// Allow some growth for caches, but it should stabilize
			if allocDiff > 10*1024*1024 { // 10MB threshold
				t.Logf("WARNING: Allocated memory grew by %s, potential leak detected", formatBytes(uint64(allocDiff)))
			}
		}
	}

	// Final GC and check
	runtime.GC()
	finalStats := getMemoryStats()

	t.Log("\nFinal memory stats:")
	t.Logf("  Alloc: %s", formatBytes(finalStats.Alloc))
	t.Logf("  TotalAlloc: %s", formatBytes(finalStats.TotalAlloc))
	t.Logf("  Sys: %s", formatBytes(finalStats.Sys))
	t.Logf("  NumGC: %d", finalStats.NumGC)

	totalAllocated := finalStats.TotalAlloc - initialStats.TotalAlloc

	// Handle Alloc change (can be negative due to GC)
	var allocChange int64
	if finalStats.Alloc > initialStats.Alloc {
		allocChange = int64(finalStats.Alloc - initialStats.Alloc)
	} else {
		allocChange = 0
	}

	t.Logf("\nMemory change after %d iterations:", iterations)
	t.Logf("  TotalAlloc: + %s", formatBytes(totalAllocated))
	if allocChange > 0 {
		t.Logf("  Alloc: + %s", formatBytes(uint64(allocChange)))
	} else {
		t.Logf("  Alloc: - %s (GC reclaimed memory)", formatBytes(initialStats.Alloc-finalStats.Alloc))
	}

	// Constitution requirement: Identify memory leaks
	// If Alloc keeps growing, it indicates a leak
	if allocChange > 5*1024*1024 { // 5MB threshold for potential leak
		t.Errorf("Potential memory leak detected: allocated memory grew by %s", formatBytes(uint64(allocChange)))
	}
}

// TestMemoryLeakConnectionPool tests for database connection leaks
func TestMemoryLeakConnectionPool(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Set max connections to detect leaks
	const maxOpenConns = 10
	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(5)

	runtime.GC()
	initialStats := getMemoryStats()

	t.Log("Testing connection pool for leaks...")

	// Open many connections and queries
	const iterations = 500
	for i := 0; i < iterations; i++ {
		// Query that could leak connections if not properly closed
		rows, err := db.Query("SELECT COUNT(*) FROM content_items")
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}

		var count int
		if !rows.Next() {
			rows.Close()
			t.Fatal("No rows returned")
		}
		rows.Scan(&count)
		rows.Close() // Critical: always close rows

		// Check stats periodically
		if (i+1)%100 == 0 {
			stats := db.Stats()
			t.Logf("Iteration %d: OpenConnections=%d, InUse=%d, Idle=%d",
				i+1, stats.OpenConnections, stats.InUse, stats.Idle)

			if stats.OpenConnections > maxOpenConns+2 {
				t.Errorf("Connection pool growing unexpectedly: %d open", stats.OpenConnections)
			}
		}
	}

	runtime.GC()
	finalStats := getMemoryStats()

	t.Log("\nConnection pool stats:")
	stats := db.Stats()
	t.Logf("  OpenConnections: %d", stats.OpenConnections)
	t.Logf("  InUse: %d", stats.InUse)
	t.Logf("  Idle: %d", stats.Idle)
	t.Logf("  WaitCount: %d", stats.WaitCount)
	t.Logf("  WaitDuration: %v", stats.WaitDuration)

	// Handle memory increase (ignore decreases which are fine with GC)
	var allocDiff int64
	if finalStats.Alloc > initialStats.Alloc {
		allocDiff = int64(finalStats.Alloc - initialStats.Alloc)
		t.Logf("  Memory change: +%s", formatBytes(uint64(allocDiff)))
	} else {
		allocDiff = 0
		t.Logf("  Memory change: -%s (GC reclaimed memory)", formatBytes(uint64(initialStats.Alloc-finalStats.Alloc)))
	}

	if stats.InUse > 0 {
		t.Errorf("Connection leak detected: %d connections still in use", stats.InUse)
	}

	if allocDiff > 5*1024*1024 {
		t.Errorf("Potential memory leak: allocated memory grew by %s", formatBytes(uint64(allocDiff)))
	}
}

// BenchmarkMemoryAllocationSearch benchmarks memory allocation during search
func BenchmarkMemoryAllocationSearch(b *testing.B) {
	db := setupTestDB(b)
	defer db.Close()

	// Insert test data
	stmt, _ := db.Prepare(`
		INSERT INTO content_items (id, title, content_text, media_type, is_deleted, created_at, updated_at, version)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`)
	defer stmt.Close()

	now := time.Now().Unix()
	for i := 0; i < 1000; i++ {
		uuid := fmt.Sprintf("%08d-0000-0000-0000-000000000000", i)
		title := fmt.Sprintf("Test Content %d", i)
		content := fmt.Sprintf("This is test content number %d with some keywords", i)
		stmt.Exec(uuid, title, content, "web", 0, now, now, 1)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		rows, _ := db.Query(`
			SELECT id, title, content_text FROM content_items
			WHERE is_deleted = 0 AND content_text LIKE '%keywords%'
			LIMIT 20
		`)

		for rows.Next() {
			var id, title, content string
			rows.Scan(&id, &title, &content)
		}
		rows.Close()
	}
}

// BenchmarkMemoryAllocationList benchmarks memory allocation during list queries
func BenchmarkMemoryAllocationList(b *testing.B) {
	db := setupTestDB(b)
	defer db.Close()

	// Insert test data
	stmt, _ := db.Prepare(`
		INSERT INTO content_items (id, title, content_text, media_type, is_deleted, created_at, updated_at, version)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`)
	defer stmt.Close()

	now := time.Now().Unix()
	for i := 0; i < 1000; i++ {
		uuid := fmt.Sprintf("%08d-0000-0000-0000-000000000000", i)
		title := fmt.Sprintf("Test Content %d", i)
		content := fmt.Sprintf("Content %d", i)
		stmt.Exec(uuid, title, content, "web", 0, now, now, 1)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		rows, _ := db.Query(`
			SELECT id, title, content_text FROM content_items
			WHERE is_deleted = 0
			ORDER BY id
			LIMIT 50
		`)

		for rows.Next() {
			var id, title, content string
			rows.Scan(&id, &title, &content)
		}
		rows.Close()
	}
}
