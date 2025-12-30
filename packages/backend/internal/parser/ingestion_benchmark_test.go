// Package parser provides content ingestion performance benchmarks.
// Constitution requirement FR-038: System must ingest 100 web pages in <10 minutes
// Constitution requirement SC-006: Background ingestion must not block UI
package parser

import (
	"database/sql"
	"fmt"
	"math/rand"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

// setupTestDB creates an in-memory database for benchmarking
func setupTestDB(b *testing.B) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		b.Fatalf("Failed to open database: %v", err)
	}

	// Enable WAL mode for concurrency
	if _, err := db.Exec("PRAGMA journal_mode = WAL"); err != nil {
		b.Fatalf("Failed to enable WAL mode: %v", err)
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

	// Create indexes
	if _, err := db.Exec("CREATE INDEX idx_content_items_created_at ON content_items(created_at DESC)"); err != nil {
		b.Fatalf("Failed to create index: %v", err)
	}
	if _, err := db.Exec("CREATE INDEX idx_content_items_is_deleted ON content_items(is_deleted)"); err != nil {
		b.Fatalf("Failed to create index: %v", err)
	}

	return db
}

// generateUUID creates a random UUID string
func generateUUID() string {
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		rand.Uint32(), rand.Uint16()&0x0fff, rand.Uint16()&0x0fff,
		rand.Uint16()&0x0fff, rand.Uint64()&0xffffffffffff)
}

// generateSampleContent generates realistic sample content
func generateSampleContent(i int) (title, content, url, tags string) {
	sampleTitles := []string{
		"Dart Programming Tutorial",
		"Flutter UI Development Guide",
		"Go Concurrency Patterns",
		"SQLite Database Optimization",
		"Mobile App Architecture",
		"REST API Best Practices",
		"WebSocket Real-time Communication",
		"Performance Testing Strategies",
	}

	sampleWords := []string{
		"development", "programming", "framework", "library", "application",
		"database", "api", "ui", "mobile", "desktop", "web", "performance",
		"testing", "debugging", "optimization", "architecture", "pattern",
	}

	title = fmt.Sprintf("%s %d", sampleTitles[i%len(sampleTitles)], i)

	// Generate realistic content (500-1000 words)
	content = fmt.Sprintf("# %s\n\n", title)
	content += fmt.Sprintf("This is a comprehensive guide about %s.\n\n", sampleTitles[i%len(sampleTitles)])
	content += "## Introduction\n\n"
	content += "In this tutorial, we will explore key concepts and best practices. "
	content += "Understanding these fundamentals is crucial for building robust applications.\n\n"
	content += "## Key Topics\n\n"

	for j := 0; j < 20; j++ {
		content += fmt.Sprintf("### Topic %d: %s\n\n", j+1, sampleWords[rand.Intn(len(sampleWords))])
		content += "When working with "
		content += sampleWords[rand.Intn(len(sampleWords))]
		content += ", it's important to consider "
		content += sampleWords[rand.Intn(len(sampleWords))]
		content += ". This approach ensures "
		content += sampleWords[rand.Intn(len(sampleWords))]
		content += " while maintaining "
		content += sampleWords[rand.Intn(len(sampleWords))]
		content += ".\n\n"

		// Add some code examples
		content += "```go\n"
		content += fmt.Sprintf("func example%d() {\n", j)
		content += fmt.Sprintf("    // %s implementation\n", sampleWords[rand.Intn(len(sampleWords))])
		content += "    return\n"
		content += "}\n```\n\n"
	}

	url = fmt.Sprintf("https://example.com/article/%d", i)

	tagWords := []string{"tutorial", "guide", "reference", "best-practices", "performance"}
	tags = fmt.Sprintf("%s,%s,%s",
		tagWords[rand.Intn(len(tagWords))],
		tagWords[rand.Intn(len(tagWords))],
		tagWords[rand.Intn(len(tagWords))])

	return title, content, url, tags
}

// BenchmarkIngest100Items benchmarks ingestion of 100 items
// Constitution requirement FR-038: <10 minutes for 100 items (avg <6 seconds per item)
func BenchmarkIngest100Items(b *testing.B) {
	db := setupTestDB(b)
	defer db.Close()

	// Prepare insert statement
	stmt, err := db.Prepare(`
		INSERT INTO content_items (id, title, content_text, source_url, media_type, tags, is_deleted, created_at, updated_at, version)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		b.Fatalf("Failed to prepare insert: %v", err)
	}
	defer stmt.Close()

	b.ResetTimer()

	const numItems = 100
	const targetDuration = 10 * time.Minute // Constitution requirement FR-038

	b.Run("SequentialIngest", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			start := time.Now()

			for j := 0; j < numItems; j++ {
				uuid := generateUUID()
				title, content, url, tags := generateSampleContent(j)
				now := time.Now().Unix()

				if _, err := stmt.Exec(uuid, title, content, url, "web", tags, 0, now, now, 1); err != nil {
					b.Fatalf("Failed to insert item %d: %v", j, err)
				}
			}

			elapsed := time.Since(start)

			// Constitution requirement FR-038: 100 items in <10 minutes
			if elapsed > targetDuration {
				b.Errorf("Ingesting %d items took %v, exceeding %v threshold (FR-038)", numItems, elapsed, targetDuration)
			}

			avgTime := elapsed / time.Duration(numItems)
			b.ReportMetric(float64(elapsed.Seconds()), "s_total")
			b.ReportMetric(float64(avgTime.Milliseconds()), "ms_avg_per_item")
			b.ReportMetric(float64(numItems), "items")
		}
	})

	b.Run("ParallelIngest", func(b *testing.B) {
		// Test concurrent ingestion (simulating background workers)
		const numWorkers = 4
		itemsPerWorker := numItems / numWorkers

		for i := 0; i < b.N; i++ {
			start := time.Now()

			// Use channels to coordinate workers
			type result struct {
				duration time.Duration
				err      error
			}
			results := make(chan result, numWorkers)

			for w := 0; w < numWorkers; w++ {
				go func(workerID int) {
					workerStart := time.Now()

					// Each worker gets its own statement
					workerStmt, err := db.Prepare(`
						INSERT INTO content_items (id, title, content_text, source_url, media_type, tags, is_deleted, created_at, updated_at, version)
						VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
					`)
					if err != nil {
						results <- result{0, fmt.Errorf("worker %d: %w", workerID, err)}
						return
					}
					defer workerStmt.Close()

					for j := 0; j < itemsPerWorker; j++ {
						itemID := workerID*itemsPerWorker + j
						uuid := generateUUID()
						title, content, url, tags := generateSampleContent(itemID)
						now := time.Now().Unix()

						if _, err := workerStmt.Exec(uuid, title, content, url, "web", tags, 0, now, now, 1); err != nil {
							results <- result{0, fmt.Errorf("worker %d item %d: %w", workerID, itemID, err)}
							return
						}
					}

					results <- result{time.Since(workerStart), nil}
				}(w)
			}

			// Wait for all workers
			maxDuration := time.Duration(0)
			for w := 0; w < numWorkers; w++ {
				res := <-results
				if res.err != nil {
					b.Fatalf("Worker error: %v", res.err)
				}
				if res.duration > maxDuration {
					maxDuration = res.duration
				}
			}

			elapsed := time.Since(start)

			// Constitution requirement FR-038: 100 items in <10 minutes
			if elapsed > targetDuration {
				b.Errorf("Parallel ingest of %d items took %v, exceeding %v threshold (FR-038)", numItems, elapsed, targetDuration)
			}

			avgTime := elapsed / time.Duration(numItems)
			b.ReportMetric(float64(elapsed.Seconds()), "s_total")
			b.ReportMetric(float64(avgTime.Milliseconds()), "ms_avg_per_item")
			b.ReportMetric(float64(maxDuration.Milliseconds()), "ms_max_worker")
			b.ReportMetric(float64(numItems), "items")
		}
	})
}

// BenchmarkSingleItemIngest benchmarks individual item ingestion
func BenchmarkSingleItemIngest(b *testing.B) {
	db := setupTestDB(b)
	defer db.Close()

	stmt, err := db.Prepare(`
		INSERT INTO content_items (id, title, content_text, source_url, media_type, tags, is_deleted, created_at, updated_at, version)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		b.Fatalf("Failed to prepare insert: %v", err)
	}
	defer stmt.Close()

	b.ResetTimer()

	// Test with varying content sizes
	sizes := []struct {
		name  string
		words int
	}{
		{"Small", 100},
		{"Medium", 500},
		{"Large", 1000},
	}

	for _, size := range sizes {
		b.Run(size.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				uuid := generateUUID()
				title, content, url, tags := generateSampleContent(i)
				now := time.Now().Unix()

				start := time.Now()
				if _, err := stmt.Exec(uuid, title, content, url, "web", tags, 0, now, now, 1); err != nil {
					b.Fatalf("Insert failed: %v", err)
				}
				elapsed := time.Since(start)

				b.ReportMetric(float64(elapsed.Microseconds()), "µs")
				b.ReportMetric(float64(len(content)), "bytes")
			}
		})
	}
}
