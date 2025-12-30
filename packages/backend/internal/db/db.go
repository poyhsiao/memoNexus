// Package db provides database connection management and operations.
package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// DB wraps the sql.DB with MemoNexus-specific configuration.
type DB struct {
	*sql.DB
}

// Open opens a SQLite database with MemoNexus configuration.
// The database is opened with:
// - WAL mode for concurrent reads/writes
// - FTS5 extension enabled
// - Foreign key constraints enabled
func Open(dataDir string) (*DB, error) {
	// Ensure data directory exists
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	// Database file path
	dbPath := filepath.Join(dataDir, "memonexus.db")

	// Open database with modernc.org/sqlite (pure Go, no CGO)
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection
	db.SetMaxOpenConns(1) // SQLite doesn't support multiple writers
	db.SetMaxIdleConns(1)

	// Enable WAL mode for better concurrency
	if _, err := db.Exec("PRAGMA journal_mode=WAL;"); err != nil {
		return nil, fmt.Errorf("failed to enable WAL mode: %w", err)
	}

	// Enable foreign key constraints
	if _, err := db.Exec("PRAGMA foreign_keys=ON;"); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	// Verify FTS5 is available
	var fts5Enabled bool
	if err := db.QueryRow("SELECT COUNT(*) > 0 FROM pragma_compile_options WHERE compile_options = 'ENABLE_FTS5'").Scan(&fts5Enabled); err != nil {
		return nil, fmt.Errorf("failed to verify FTS5: %w", err)
	}
	if !fts5Enabled {
		return nil, fmt.Errorf("FTS5 is not enabled in this SQLite build")
	}

	return &DB{db}, nil
}

// Close closes the database connection.
func (db *DB) Close() error {
	return db.DB.Close()
}
