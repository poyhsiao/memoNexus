// Integration tests for offline functionality.
// Constitution requirement: 100% of features must work without network connectivity
package integration

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kimhsiao/memonexus/backend/internal/db"
	_ "modernc.org/sqlite"
)

// setupOfflineDB creates a local SQLite database for offline testing
func setupOfflineDB(t *testing.T) (*sql.DB, string) {
	// Create temp directory for test database
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	// Open database
	database, err := db.Open(tempDir)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// Run migrations
	migrator := db.NewMigrator(database.DB, "./../../packages/backend/internal/db/migrations")
	if err := migrator.Initialize(); err != nil {
		t.Fatalf("Failed to initialize migrator: %v", err)
	}
	if err := migrator.Up(); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	return database, dbPath
}

// TestOfflineContentCRUD tests content operations work completely offline
func TestOfflineContentCRUD(t *testing.T) {
	database, dbPath := setupOfflineDB(t)
	defer database.Close()
	defer os.Remove(dbPath)

	repository := db.NewRepository(database.DB)

	t.Log("Testing offline content CRUD operations...")

	// Test CREATE
	t.Run("Create", func(t *testing.T) {
		item := &db.ContentItem{
			Title:       "Offline Test Content",
			ContentText: "This content was created completely offline",
			MediaType:   "markdown",
			Tags:        "offline,test",
		}

		err := repository.CreateContentItem(item)
		if err != nil {
			t.Fatalf("Failed to create content: %v", err)
		}

		if item.ID == "" {
			t.Error("ID was not generated")
		}

		t.Logf("Created content with ID: %s", item.ID)
	})

	// Test READ
	var itemID string
	t.Run("Read", func(t *testing.T) {
		item := &db.ContentItem{
			Title:       "Read Test Content",
			ContentText: "Content for read test",
			MediaType:   "web",
		}

		if err := repository.CreateContentItem(item); err != nil {
			t.Fatalf("Failed to create test item: %v", err)
		}
		itemID = item.ID

		readItem, err := repository.GetContentItem(item.ID)
		if err != nil {
			t.Fatalf("Failed to read content: %v", err)
		}

		if readItem.Title != item.Title {
			t.Errorf("Title mismatch: got %s, want %s", readItem.Title, item.Title)
		}

		t.Logf("Successfully read content: %s", readItem.Title)
	})

	// Test UPDATE
	t.Run("Update", func(t *testing.T) {
		if itemID == "" {
			t.Skip("No item ID available")
		}

		item, err := repository.GetContentItem(itemID)
		if err != nil {
			t.Fatalf("Failed to get item: %v", err)
		}

		item.Title = "Updated Title Offline"
		item.Summary = "This was updated without network"

		err = repository.UpdateContentItem(item)
		if err != nil {
			t.Fatalf("Failed to update content: %v", err)
		}

		updatedItem, err := repository.GetContentItem(itemID)
		if err != nil {
			t.Fatalf("Failed to read updated content: %v", err)
		}

		if updatedItem.Title != item.Title {
			t.Errorf("Update failed: got %s, want %s", updatedItem.Title, item.Title)
		}

		t.Logf("Successfully updated content to: %s", updatedItem.Title)
	})

	// Test LIST
	t.Run("List", func(t *testing.T) {
		// Create multiple items
		for i := 0; i < 5; i++ {
			item := &db.ContentItem{
				Title:       fmt.Sprintf("List Test Item %d", i),
				ContentText: fmt.Sprintf("Content %d", i),
				MediaType:   "web",
			}
			if err := repository.CreateContentItem(item); err != nil {
				t.Fatalf("Failed to create item %d: %v", i, err)
			}
		}

		items, err := repository.ListContentItems(10, 0, "")
		if err != nil {
			t.Fatalf("Failed to list content: %v", err)
		}

		if len(items) < 5 {
			t.Errorf("Expected at least 5 items, got %d", len(items))
		}

		t.Logf("Successfully listed %d items", len(items))
	})

	// Test DELETE
	t.Run("Delete", func(t *testing.T) {
		item := &db.ContentItem{
			Title:       "Delete Test Item",
			ContentText: "This will be deleted",
			MediaType:   "web",
		}

		if err := repository.CreateContentItem(item); err != nil {
			t.Fatalf("Failed to create item: %v", err)
		}

		deleteID := item.ID

		err := repository.DeleteContentItem(deleteID)
		if err != nil {
			t.Fatalf("Failed to delete content: %v", err)
		}

		// Verify deletion
		_, err = repository.GetContentItem(deleteID)
		if err == nil {
			t.Error("Expected error when reading deleted item, got nil")
		}

		t.Log("Successfully deleted content")
	})
}

// TestOfflineTagCRUD tests tag operations work completely offline
func TestOfflineTagCRUD(t *testing.T) {
	database, dbPath := setupOfflineDB(t)
	defer database.Close()
	defer os.Remove(dbPath)

	repository := db.NewRepository(database.DB)

	t.Log("Testing offline tag CRUD operations...")

	// Test CREATE tag
	t.Run("CreateTag", func(t *testing.T) {
		tag := &db.Tag{
			Name:  "offline-tag",
			Color: "#FF5722",
		}

		err := repository.CreateTag(tag)
		if err != nil {
			t.Fatalf("Failed to create tag: %v", err)
		}

		t.Logf("Created tag: %s (color: %s)", tag.Name, tag.Color)
	})

	// Test LIST tags
	t.Run("ListTags", func(t *testing.T) {
		// Create multiple tags
		colors := []string{"#3B82F6", "#10B981", "#F59E0B"}
		for i, color := range colors {
			tag := &db.Tag{
				Name:  fmt.Sprintf("tag-%d", i),
				Color: color,
			}
			if err := repository.CreateTag(tag); err != nil {
				t.Fatalf("Failed to create tag %d: %v", i, err)
			}
		}

		tags, err := repository.ListTags()
		if err != nil {
			t.Fatalf("Failed to list tags: %v", err)
		}

		if len(tags) < 3 {
			t.Errorf("Expected at least 3 tags, got %d", len(tags))
		}

		t.Logf("Successfully listed %d tags", len(tags))
	})

	// Test GET tag
	var tagID string
	t.Run("GetTag", func(t *testing.T) {
		tag := &db.Tag{
			Name:  "test-tag",
			Color: "#8B5CF6",
		}

		if err := repository.CreateTag(tag); err != nil {
			t.Fatalf("Failed to create tag: %v", err)
		}
		tagID = tag.ID

		readTag, err := repository.GetTag(tag.ID)
		if err != nil {
			t.Fatalf("Failed to read tag: %v", err)
		}

		if readTag.Name != tag.Name {
			t.Errorf("Tag name mismatch: got %s, want %s", readTag.Name, tag.Name)
		}

		t.Logf("Successfully read tag: %s", readTag.Name)
	})
}

// TestOfflineSearch tests full-text search works completely offline
func TestOfflineSearch(t *testing.T) {
	database, dbPath := setupOfflineDB(t)
	defer database.Close()
	defer os.Remove(dbPath)

	repository := db.NewRepository(database.DB)

	t.Log("Testing offline full-text search...")

	// Create test content with searchable terms
	testContent := []struct {
		title   string
		content string
		tags    string
	}{
		{"Dart Programming Guide", "Learn Dart programming language for Flutter development", "dart,flutter"},
		{"Go Concurrency", "Understanding goroutines and channels in Go", "go,golang"},
		{"SQLite Tutorial", "Master SQLite database with FTS5 full-text search", "database,sqlite"},
		{"Web Development", "Modern web development with HTML, CSS, and JavaScript", "web,frontend"},
		{"Mobile App Development", "Building cross-platform mobile applications", "mobile,app"},
	}

	for _, tc := range testContent {
		item := &db.ContentItem{
			Title:       tc.title,
			ContentText: tc.content,
			MediaType:   "markdown",
			Tags:        tc.tags,
		}
		if err := repository.CreateContentItem(item); err != nil {
			t.Fatalf("Failed to create content: %v", err)
		}
	}

	// Test full-text search
	t.Run("FTSSearch", func(t *testing.T) {
		// Search for "Dart programming"
		rows, err := database.DB.Query(`
			SELECT ci.id, ci.title, ci.content_text, ci.tags
			FROM content_items ci
			INNER JOIN content_items_fts fts ON ci.rowid = fts.rowid
			WHERE content_items_fts MATCH 'Dart programming'
			AND ci.is_deleted = 0
			ORDER BY rank
			LIMIT 10
		`)
		if err != nil {
			t.Fatalf("Search query failed: %v", err)
		}
		defer rows.Close()

		results := 0
		for rows.Next() {
			var id, title, content, tags string
			if err := rows.Scan(&id, &title, &content, &tags); err != nil {
				t.Fatalf("Scan failed: %v", err)
			}
			results++
			t.Logf("Found: %s (tags: %s)", title, tags)
		}

		if results == 0 {
			t.Error("FTS search returned no results")
		}

		t.Logf("Full-text search returned %d results", results)
	})

	// Test search with ranking
	t.Run("FTSSearchRanking", func(t *testing.T) {
		searchTerms := []string{"development", "programming", "database"}

		for _, term := range searchTerms {
			rows, err := database.DB.Query(`
				SELECT ci.title, rank
				FROM content_items ci
				INNER JOIN content_items_fts fts ON ci.rowid = fts.rowid
				WHERE content_items_fts MATCH ?
				AND ci.is_deleted = 0
				ORDER BY rank
				LIMIT 5
			`, term)
			if err != nil {
				t.Fatalf("Search failed for '%s': %v", term, err)
			}

			count := 0
			for rows.Next() {
				var title string
				var rank float64
				rows.Scan(&title, &rank)
				count++
				t.Logf("Term '%s': %s (rank: %.2f)", term, title, rank)
			}
			rows.Close()

			if count == 0 {
				t.Errorf("No results for search term: %s", term)
			}
		}
	})
}

// TestOfflinePersistence tests data persists across database restarts
func TestOfflinePersistence(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "persist.db")

	// Phase 1: Create data
	t.Log("Phase 1: Creating data...")
	database1, err := db.Open(tempDir)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	migrator1 := db.NewMigrator(database1.DB, "./../../packages/backend/internal/db/migrations")
	if err := migrator1.Initialize(); err != nil {
		t.Fatalf("Failed to initialize migrator: %v", err)
	}
	if err := migrator1.Up(); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	repository1 := db.NewRepository(database1.DB)

	item := &db.ContentItem{
		Title:       "Persistent Test Content",
		ContentText: "This should persist across database restarts",
		MediaType:   "markdown",
		Tags:        "persistence,test",
	}

	if err := repository1.CreateContentItem(item); err != nil {
		t.Fatalf("Failed to create item: %v", err)
	}

	itemID := item.ID
	t.Logf("Created item with ID: %s", itemID)

	database1.Close()

	// Phase 2: Reopen database and verify data
	t.Log("Phase 2: Reopening database...")
	time.Sleep(100 * time.Millisecond) // Brief pause

	database2, err := db.Open(tempDir)
	if err != nil {
		t.Fatalf("Failed to reopen database: %v", err)
	}
	defer database2.Close()

	repository2 := db.NewRepository(database2.DB)

	readItem, err := repository2.GetContentItem(itemID)
	if err != nil {
		t.Fatalf("Failed to read item after restart: %v", err)
	}

	if readItem.Title != item.Title {
		t.Errorf("Title mismatch after restart: got %s, want %s", readItem.Title, item.Title)
	}

	if readItem.ContentText != item.ContentText {
		t.Errorf("Content mismatch after restart: got %s, want %s", readItem.ContentText, item.ContentText)
	}

	t.Log("Data successfully persisted across database restart")
}

// TestOfflineConcurrency tests concurrent operations work offline
func TestOfflineConcurrency(t *testing.T) {
	database, dbPath := setupOfflineDB(t)
	defer database.Close()
	defer os.Remove(dbPath)

	repository := db.NewRepository(database.DB)

	t.Log("Testing offline concurrent operations...")

	// Enable WAL mode for concurrent access
	if _, err := database.DB.Exec("PRAGMA journal_mode = WAL"); err != nil {
		t.Fatalf("Failed to enable WAL mode: %v", err)
	}

	const numGoroutines = 10
	const itemsPerGoroutine = 5

	done := make(chan bool, numGoroutines)
	errors := make(chan error, numGoroutines*itemsPerGoroutine)

	// Concurrent writes
	for g := 0; g < numGoroutines; g++ {
		go func(goroutineID int) {
			for i := 0; i < itemsPerGoroutine; i++ {
				item := &db.ContentItem{
					Title:       fmt.Sprintf("Concurrent Item %d-%d", goroutineID, i),
					ContentText: fmt.Sprintf("Content from goroutine %d, item %d", goroutineID, i),
					MediaType:   "web",
				}
				if err := repository.CreateContentItem(item); err != nil {
					errors <- fmt.Errorf("goroutine %d item %d: %w", goroutineID, i, err)
				}
			}
			done <- true
		}(g)
	}

	// Wait for all goroutines
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Check for errors
	close(errors)
	errorList := make([]error, 0)
	for err := range errors {
		errorList = append(errorList, err)
	}

	if len(errorList) > 0 {
		t.Fatalf("Encountered %d errors during concurrent writes: %v", len(errorList), errorList[0])
	}

	// Verify all items were created
	items, err := repository.ListContentItems(100, 0, "")
	if err != nil {
		t.Fatalf("Failed to list items: %v", err)
	}

	expectedCount := numGoroutines * itemsPerGoroutine
	if len(items) != expectedCount {
		t.Errorf("Expected %d items, got %d", expectedCount, len(items))
	}

	t.Logf("Successfully handled %d concurrent writes, created %d items", numGoroutines, len(items))
}

// TestOfflinePerformance100Items tests performance of ingesting 100 items offline
func TestOfflinePerformance100Items(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	database, dbPath := setupOfflineDB(t)
	defer database.Close()
	defer os.Remove(dbPath)

	repository := db.NewRepository(database.DB)

	t.Log("Testing offline ingestion performance for 100 items...")

	start := time.Now()

	for i := 0; i < 100; i++ {
		item := &db.ContentItem{
			Title:       fmt.Sprintf("Performance Test Item %d", i),
			ContentText: fmt.Sprintf("This is test content number %d for performance testing", i),
			MediaType:   "web",
			Tags:        "performance,test",
		}

		if err := repository.CreateContentItem(item); err != nil {
			t.Fatalf("Failed to create item %d: %v", i, err)
		}
	}

	elapsed := time.Since(start)
	avgTime := elapsed / 100

	t.Logf("Ingested 100 items in %v (avg: %v per item)", elapsed, avgTime)

	// Constitution requirement FR-038: 100 items in <10 minutes
	// Our benchmark should be much faster: <10 seconds total
	if elapsed > 10*time.Minute {
		t.Errorf("Ingestion took %v, exceeding 10 minute threshold (FR-038)", elapsed)
	}

	if elapsed > 10*time.Second {
		t.Logf("WARNING: Ingestion took %v, consider optimization", elapsed)
	}
}
