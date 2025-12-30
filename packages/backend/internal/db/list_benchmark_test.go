// Package db provides content list performance benchmarks.
// Constitution requirement FR-039: Content list must render in <500ms
// Constitution requirement SC-005: Pagination must limit queries to 50 items per page
package db

import (
	"database/sql"
	"fmt"
	"math/rand"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

// BenchmarkListRender1000Items benchmarks list rendering performance with 1,000 items
// Constitution requirement FR-039: <500ms to render list view
func BenchmarkListRender1000Items(b *testing.B) {
	db := setupTestDB(b)
	defer db.Close()

	// Populate with 1,000 items
	b.Log("Populating database with 1,000 items...")
	if err := populateContentItems(db, 1000); err != nil {
		b.Fatalf("Failed to populate items: %v", err)
	}

	b.ResetTimer()

	// Simulate list view rendering with pagination
	b.Run("FirstPage", func(b *testing.B) {
		const pageSize = 50
		for i := 0; i < b.N; i++ {
			start := time.Now()

			// Query first page
			rows, err := db.Query(`
				SELECT id, title, content_text, media_type, tags, created_at, updated_at
				FROM content_items
				WHERE is_deleted = 0
				ORDER BY created_at DESC
				LIMIT ? OFFSET 0
			`, pageSize)
			if err != nil {
				b.Fatalf("Query failed: %v", err)
			}

			// Fetch all items
			items := make([]map[string]interface{}, 0, pageSize)
			for rows.Next() {
				var id, title, contentText, mediaType, tags string
				var createdAt, updatedAt int64
				if err := rows.Scan(&id, &title, &contentText, &mediaType, &tags, &createdAt, &updatedAt); err != nil {
					b.Fatalf("Scan failed: %v", err)
				}
				items = append(items, map[string]interface{}{
					"id":         id,
					"title":      title,
					"content":    contentText,
					"media_type": mediaType,
					"tags":       tags,
					"created_at": createdAt,
					"updated_at": updatedAt,
				})
			}
			rows.Close()

			elapsed := time.Since(start)

			// Verify we got items
			if len(items) == 0 {
				b.Error("No items returned")
			}

			// Constitution requirement FR-039: <500ms
			if elapsed > 500*time.Millisecond {
				b.Errorf("List render took %v, exceeding 500ms threshold (FR-039)", elapsed)
			}

			b.ReportMetric(float64(elapsed.Milliseconds()), "ms")
			b.ReportMetric(float64(len(items)), "items")
		}
	})

	b.Run("Pagination", func(b *testing.B) {
		const pageSize = 50
		const totalPages = 5 // Test first 5 pages

		for page := 0; page < totalPages; page++ {
			b.Run(fmt.Sprintf("Page%d", page+1), func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					offset := page * pageSize
					start := time.Now()

					rows, err := db.Query(`
						SELECT id, title, content_text, media_type, tags, created_at
						FROM content_items
						WHERE is_deleted = 0
						ORDER BY created_at DESC
						LIMIT ? OFFSET ?
					`, pageSize, offset)
					if err != nil {
						b.Fatalf("Query failed: %v", err)
					}

					count := 0
					for rows.Next() {
						count++
					}
					rows.Close()

					elapsed := time.Since(start)

					// Constitution requirement FR-039: <500ms
					if elapsed > 500*time.Millisecond {
						b.Errorf("Page %d render took %v, exceeding 500ms threshold (FR-039)", page+1, elapsed)
					}

					b.ReportMetric(float64(elapsed.Milliseconds()), "ms")
					b.ReportMetric(float64(count), "items")
				}
			})
		}
	})

	b.Run("FilterByMediaType", func(b *testing.B) {
		const pageSize = 50
		mediaTypes := []string{"web", "image", "video", "pdf", "markdown"}

		for _, mt := range mediaTypes {
			b.Run(mt, func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					start := time.Now()

					rows, err := db.Query(`
						SELECT id, title, content_text, media_type, tags, created_at
						FROM content_items
						WHERE is_deleted = 0 AND media_type = ?
						ORDER BY created_at DESC
						LIMIT ? OFFSET 0
					`, mt, pageSize)
					if err != nil {
						b.Fatalf("Query failed: %v", err)
					}

					count := 0
					for rows.Next() {
						count++
					}
					rows.Close()

					elapsed := time.Since(start)

					if elapsed > 500*time.Millisecond {
						b.Errorf("Filtered list render took %v, exceeding 500ms threshold (FR-039)", elapsed)
					}

					b.ReportMetric(float64(elapsed.Milliseconds()), "ms")
					b.ReportMetric(float64(count), "items")
				}
			})
		}
	})
}

// BenchmarkListRender10000Items benchmarks list rendering with 10,000 items
func BenchmarkListRender10000Items(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping large benchmark in short mode")
	}

	db := setupTestDB(b)
	defer db.Close()

	b.Log("Populating database with 10,000 items...")
	if err := populateContentItems(db, 10000); err != nil {
		b.Fatalf("Failed to populate items: %v", err)
	}

	b.ResetTimer()

	const pageSize = 50 // Constitution requirement SC-005

	for i := 0; i < b.N; i++ {
		start := time.Now()

		rows, err := db.Query(`
			SELECT id, title, content_text, media_type, tags, created_at
			FROM content_items
			WHERE is_deleted = 0
			ORDER BY created_at DESC
			LIMIT ? OFFSET 0
		`, pageSize)
		if err != nil {
			b.Fatalf("Query failed: %v", err)
		}

		count := 0
		for rows.Next() {
			count++
		}
		rows.Close()

		elapsed := time.Since(start)

		// Even with 10K items, first page should be fast
		if elapsed > 500*time.Millisecond {
			b.Errorf("List render took %v, exceeding 500ms threshold (FR-039)", elapsed)
		}

		b.ReportMetric(float64(elapsed.Milliseconds()), "ms")
		b.ReportMetric(float64(count), "items")
	}
}
