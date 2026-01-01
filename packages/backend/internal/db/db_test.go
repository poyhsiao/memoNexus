// Package db tests for database connection management.
package db

import (
	"os"
	"path/filepath"
	"testing"
)

// TestOpen verifies database opening with proper configuration.
func TestOpen(t *testing.T) {
	// Create temporary directory for test database
	tmpDir, err := os.MkdirTemp("", "memonexus_db_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Open database
	db, err := Open(tmpDir)
	if err != nil {
		t.Fatalf("Open() failed: %v", err)
	}
	defer db.Close()

	// Verify database file was created
	dbPath := filepath.Join(tmpDir, "memonexus.db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("Database file was not created")
	}

	// Verify connection is usable
	var result int
	err = db.QueryRow("SELECT 1").Scan(&result)
	if err != nil {
		t.Errorf("Database query failed: %v", err)
	}
	if result != 1 {
		t.Errorf("Expected 1, got %d", result)
	}

	// Verify WAL mode is enabled
	var walMode string
	err = db.QueryRow("PRAGMA journal_mode").Scan(&walMode)
	if err != nil {
		t.Errorf("Failed to check WAL mode: %v", err)
	}
	if walMode != "wal" {
		t.Errorf("WAL mode not enabled, got: %s", walMode)
	}

	// Verify foreign keys are enabled
	var fkEnabled int
	err = db.QueryRow("PRAGMA foreign_keys").Scan(&fkEnabled)
	if err != nil {
		t.Errorf("Failed to check foreign keys: %v", err)
	}
	if fkEnabled != 1 {
		t.Errorf("Foreign keys not enabled, got: %d", fkEnabled)
	}

	// Verify FTS5 is available
	var fts5Enabled bool
	err = db.QueryRow("SELECT COUNT(*) > 0 FROM pragma_compile_options WHERE compile_options = 'ENABLE_FTS5'").Scan(&fts5Enabled)
	if err != nil {
		t.Errorf("Failed to check FTS5: %v", err)
	}
	if !fts5Enabled {
		t.Error("FTS5 is not enabled")
	}
}

// TestOpen_invalidDataDir verifies error when data directory cannot be created.
func TestOpen_invalidDataDir(t *testing.T) {
	// Use a path that cannot be created as a directory
	invalidPath := "/dev/null/invalid_path/that/cannot/be/created"

	_, err := Open(invalidPath)
	if err == nil {
		t.Error("Open() with invalid path should return error")
	}
}

// TestClose verifies database closing.
func TestClose(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "memonexus_db_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	db, err := Open(tmpDir)
	if err != nil {
		t.Fatalf("Open() failed: %v", err)
	}

	// Close database
	err = db.Close()
	if err != nil {
		t.Errorf("Close() failed: %v", err)
	}

	// Close is idempotent in SQLite - second call should succeed
	err = db.Close()
	if err != nil {
		t.Errorf("Second Close() should not return error, got: %v", err)
	}

	// Try to query closed database - should fail
	var result int
	err = db.QueryRow("SELECT 1").Scan(&result)
	if err == nil {
		t.Error("Query on closed database should fail")
	}
}

// TestDB_reopen verifies database can be reopened after close.
func TestDB_reopen(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "memonexus_db_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// First open
	db1, err := Open(tmpDir)
	if err != nil {
		t.Fatalf("First Open() failed: %v", err)
	}

	// Create a test table
	_, err = db1.Exec("CREATE TABLE test_table (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	// Insert test data
	_, err = db1.Exec("INSERT INTO test_table (id, name) VALUES (1, 'test')")
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	// Close database
	err = db1.Close()
	if err != nil {
		t.Fatalf("Close() failed: %v", err)
	}

	// Reopen database
	db2, err := Open(tmpDir)
	if err != nil {
		t.Fatalf("Second Open() failed: %v", err)
	}
	defer db2.Close()

	// Verify data persisted
	var name string
	err = db2.QueryRow("SELECT name FROM test_table WHERE id = 1").Scan(&name)
	if err != nil {
		t.Errorf("Failed to query test data: %v", err)
	}
	if name != "test" {
		t.Errorf("Expected 'test', got %q", name)
	}
}

// TestDB_concurrentQueries verifies database handles multiple queries.
func TestDB_concurrentQueries(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "memonexus_db_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	db, err := Open(tmpDir)
	if err != nil {
		t.Fatalf("Open() failed: %v", err)
	}
	defer db.Close()

	// Create a test table
	_, err = db.Exec("CREATE TABLE test_table (id INTEGER PRIMARY KEY, value INTEGER)")
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	// Insert test data
	for i := 1; i <= 10; i++ {
		_, err = db.Exec("INSERT INTO test_table (id, value) VALUES (?, ?)", i, i*10)
		if err != nil {
			t.Fatalf("Failed to insert test data: %v", err)
		}
	}

	// Run multiple concurrent queries
	done := make(chan bool, 5)
	for i := 0; i < 5; i++ {
		go func() {
			rows, err := db.Query("SELECT value FROM test_table")
			if err != nil {
				t.Errorf("Concurrent query failed: %v", err)
				done <- false
				return
			}
			defer rows.Close()
			for rows.Next() {
			}
			done <- true
		}()
	}

	// Wait for all queries to complete
	for i := 0; i < 5; i++ {
		if !<-done {
			t.Error("Concurrent query failed")
		}
	}
}
