// Package db tests for database migration management.
package db

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestNewMigrator verifies Migrator initialization.
func TestNewMigrator(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open in-memory database: %v", err)
	}
	defer db.Close()

	migrateDir := "/test/migrations"
	m := NewMigrator(db, migrateDir)

	if m == nil {
		t.Fatal("NewMigrator() returned nil")
	}

	if m.db != db {
		t.Error("Migrator.db not set correctly")
	}

	if m.migrateDir != migrateDir {
		t.Errorf("Migrator.migrateDir = %q, want %q", m.migrateDir, migrateDir)
	}
}

// TestInitialize verifies schema_migrations table creation.
func TestInitialize(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open in-memory database: %v", err)
	}
	defer db.Close()

	m := NewMigrator(db, "/test/migrations")

	err = m.Initialize()
	if err != nil {
		t.Fatalf("Initialize() failed: %v", err)
	}

	// Verify table exists
	var tableName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='schema_migrations'").Scan(&tableName)
	if err != nil {
		t.Errorf("schema_migrations table not found: %v", err)
	}

	// Verify table structure by inserting a test row
	_, err = db.Exec("INSERT INTO schema_migrations (version, applied_at, description, checksum) VALUES (?, ?, ?, ?)",
		1, 123456, "test_migration", strings.Repeat("a", 64))
	if err != nil {
		t.Errorf("Failed to insert test row: %v", err)
	}
}

// TestCurrentVersion verifies version tracking.
func TestCurrentVersion(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open in-memory database: %v", err)
	}
	defer db.Close()

	m := NewMigrator(db, "/test/migrations")

	// Before initialization
	_, err = m.CurrentVersion()
	if err == nil {
		t.Error("CurrentVersion() should fail before Initialize()")
	}

	// Initialize and check version 0
	err = m.Initialize()
	if err != nil {
		t.Fatalf("Initialize() failed: %v", err)
	}

	version, err := m.CurrentVersion()
	if err != nil {
		t.Errorf("CurrentVersion() failed: %v", err)
	}
	if version != 0 {
		t.Errorf("CurrentVersion() = %d, want 0", version)
	}

	// Insert a migration
	_, err = db.Exec("INSERT INTO schema_migrations (version, applied_at, description, checksum) VALUES (?, ?, ?, ?)",
		1, 123456, "V1__initial", strings.Repeat("a", 64))
	if err != nil {
		t.Fatalf("Failed to insert migration: %v", err)
	}

	version, err = m.CurrentVersion()
	if err != nil {
		t.Errorf("CurrentVersion() failed: %v", err)
	}
	if version != 1 {
		t.Errorf("CurrentVersion() = %d, want 1", version)
	}
}

// TestGetAppliedMigrations verifies migration listing.
func TestGetAppliedMigrations(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open in-memory database: %v", err)
	}
	defer db.Close()

	m := NewMigrator(db, "/test/migrations")

	err = m.Initialize()
	if err != nil {
		t.Fatalf("Initialize() failed: %v", err)
	}

	// Initially empty
	migrations, err := m.GetAppliedMigrations()
	if err != nil {
		t.Errorf("GetAppliedMigrations() failed: %v", err)
	}
	if len(migrations) != 0 {
		t.Errorf("GetAppliedMigrations() = %d, want 0", len(migrations))
	}

	// Insert test migrations
	checksum := strings.Repeat("a", 64)
	_, err = db.Exec("INSERT INTO schema_migrations (version, applied_at, description, checksum) VALUES (?, ?, ?, ?)",
		1, 1000, "V1__initial", checksum)
	if err != nil {
		t.Fatalf("Failed to insert migration 1: %v", err)
	}
	_, err = db.Exec("INSERT INTO schema_migrations (version, applied_at, description, checksum) VALUES (?, ?, ?, ?)",
		2, 2000, "V2__add_column", checksum)
	if err != nil {
		t.Fatalf("Failed to insert migration 2: %v", err)
	}

	migrations, err = m.GetAppliedMigrations()
	if err != nil {
		t.Errorf("GetAppliedMigrations() failed: %v", err)
	}
	if len(migrations) != 2 {
		t.Errorf("GetAppliedMigrations() = %d, want 2", len(migrations))
	}

	// Verify order (should be sorted by version)
	if migrations[0].Version != 1 {
		t.Errorf("First migration version = %d, want 1", migrations[0].Version)
	}
	if migrations[1].Version != 2 {
		t.Errorf("Second migration version = %d, want 2", migrations[1].Version)
	}
}

// TestUp_noMigrations verifies Up succeeds when no migrations exist.
func TestUp_noMigrations(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "migrate_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	db, err := sql.Open("sqlite", filepath.Join(tmpDir, "test.db"))
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	m := NewMigrator(db, tmpDir)

	err = m.Initialize()
	if err != nil {
		t.Fatalf("Initialize() failed: %v", err)
	}

	// Up should succeed with no migrations
	err = m.Up()
	if err != nil {
		t.Errorf("Up() with no migrations failed: %v", err)
	}
}

// TestDown_noMigrations verifies error when no migrations to rollback.
func TestDown_noMigrations(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open in-memory database: %v", err)
	}
	defer db.Close()

	m := NewMigrator(db, "/test/migrations")

	err = m.Initialize()
	if err != nil {
		t.Fatalf("Initialize() failed: %v", err)
	}

	err = m.Down()
	if err == nil {
		t.Error("Down() with no migrations should return error")
	}
	if !strings.Contains(err.Error(), "no migrations to rollback") {
		t.Errorf("Error message should mention 'no migrations to rollback', got: %v", err)
	}
}

// TestUp_appliesMigration verifies migration files are applied.
func TestUp_appliesMigration(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "migrate_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	db, err := sql.Open("sqlite", filepath.Join(tmpDir, "test.db"))
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	m := NewMigrator(db, tmpDir)

	err = m.Initialize()
	if err != nil {
		t.Fatalf("Initialize() failed: %v", err)
	}

	// Create a test migration file
	migrationSQL := `CREATE TABLE test_table (id INTEGER PRIMARY KEY, name TEXT);`
	err = os.WriteFile(filepath.Join(tmpDir, "V1__test_migration.up.sql"), []byte(migrationSQL), 0644)
	if err != nil {
		t.Fatalf("Failed to create migration file: %v", err)
	}

	// Apply migration
	err = m.Up()
	if err != nil {
		t.Errorf("Up() failed: %v", err)
	}

	// Verify table was created
	var tableName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='test_table'").Scan(&tableName)
	if err != nil {
		t.Errorf("Migration not applied: %v", err)
	}

	// Verify migration was recorded
	version, err := m.CurrentVersion()
	if err != nil {
		t.Errorf("CurrentVersion() failed: %v", err)
	}
	if version != 1 {
		t.Errorf("CurrentVersion() = %d, want 1", version)
	}

	// Running Up again should skip already applied migration
	err = m.Up()
	if err != nil {
		t.Errorf("Up() second time failed: %v", err)
	}
}
