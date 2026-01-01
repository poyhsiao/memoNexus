// Package db provides database schema migration management.
package db

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Migration represents a database schema migration.
type Migration struct {
	Version    int
	AppliedAt  time.Time
	Description string
	Checksum    string
}

// Migrator handles database schema migrations.
type Migrator struct {
	db      *sql.DB
	migrateDir string
}

// NewMigrator creates a new Migrator instance.
func NewMigrator(db *sql.DB, migrateDir string) *Migrator {
	return &Migrator{
		db:         db,
		migrateDir: migrateDir,
	}
}

// Initialize creates the schema_migrations table if it doesn't exist.
func (m *Migrator) Initialize() error {
	query := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY CHECK(version > 0),
		applied_at INTEGER NOT NULL CHECK(applied_at > 0),
		description TEXT NOT NULL CHECK(length(description) > 0),
		checksum TEXT NOT NULL CHECK(length(checksum) = 64)
	);`
	_, err := m.db.Exec(query)
	return err
}

// CurrentVersion returns the current schema version.
func (m *Migrator) CurrentVersion() (int, error) {
	var version int
	err := m.db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_migrations").Scan(&version)
	return version, err
}

// GetAppliedMigrations returns all applied migrations.
func (m *Migrator) GetAppliedMigrations() ([]Migration, error) {
	rows, err := m.db.Query("SELECT version, applied_at, description, checksum FROM schema_migrations ORDER BY version")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var migrations []Migration
	for rows.Next() {
		var m Migration
		var appliedAt int64
		err := rows.Scan(&m.Version, &appliedAt, &m.Description, &m.Checksum)
		if err != nil {
			return nil, err
		}
		m.AppliedAt = time.Unix(appliedAt, 0)
		migrations = append(migrations, m)
	}
	return migrations, nil
}

// Up applies all pending migrations.
func (m *Migrator) Up() error {
	// Get applied versions
	applied, err := m.GetAppliedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}
	appliedVersions := make(map[int]bool)
	for _, mig := range applied {
		appliedVersions[mig.Version] = true
	}

	// List migration files
	entries, err := os.ReadDir(m.migrateDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	// Process migration files
	var migrations []struct {
		version int
		name    string
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()

		// Parse version from filename (V1__initial_schema.up.sql)
		if !strings.HasSuffix(name, ".up.sql") {
			continue
		}

		parts := strings.Split(strings.TrimSuffix(name, ".up.sql"), "__")
		if len(parts) < 2 {
			continue
		}

		versionStr := strings.TrimPrefix(parts[0], "V")
		version, err := strconv.Atoi(versionStr)
		if err != nil {
			continue
		}

		migrations = append(migrations, struct {
			version int
			name    string
		}{version, name})
	}

	// Sort by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].version < migrations[j].version
	})

	// Apply pending migrations
	for _, mig := range migrations {
		if appliedVersions[mig.version] {
			continue // Already applied
		}

		if err := m.applyMigration(mig.version, mig.name); err != nil {
			return fmt.Errorf("failed to apply migration V%d: %w", mig.version, err)
		}
	}

	return nil
}

// applyMigration applies a single migration.
func (m *Migrator) applyMigration(version int, filename string) error {
	// Read migration SQL
	path := filepath.Join(m.migrateDir, filename)
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	// Start transaction
	tx, err := m.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Execute migration SQL
	if _, err := tx.Exec(string(content)); err != nil {
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	// Record migration
	description := strings.TrimSuffix(filename, ".up.sql")
	description = strings.TrimPrefix(description, fmt.Sprintf("V%d__", version))
	query := `INSERT INTO schema_migrations (version, applied_at, description, checksum)
			  VALUES (?, ?, ?, ?)`
	// Compute SHA-256 checksum of migration SQL content
	hash := sha256.Sum256(content)
	checksum := hex.EncodeToString(hash[:])
	if _, err := tx.Exec(query, version, time.Now().Unix(), description, checksum); err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	return tx.Commit()
}

// Down rolls back the last migration.
func (m *Migrator) Down() error {
	current, err := m.CurrentVersion()
	if err != nil {
		return err
	}
	if current == 0 {
		return fmt.Errorf("no migrations to rollback")
	}

	// Find the down migration file (V%d__*.down.sql pattern)
	pattern := fmt.Sprintf("V%d__*.down.sql", current)
	matches, err := filepath.Glob(filepath.Join(m.migrateDir, pattern))
	if err != nil {
		return fmt.Errorf("failed to search for rollback migration: %w", err)
	}
	if len(matches) == 0 {
		return fmt.Errorf("no rollback migration found for version %d", current)
	}
	// Use the first match (there should be only one)
	path := matches[0]

	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read rollback migration: %w", err)
	}

	// Start transaction
	tx, err := m.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Execute rollback SQL
	if _, err := tx.Exec(string(content)); err != nil {
		return fmt.Errorf("failed to execute rollback SQL: %w", err)
	}

	// Remove migration record
	if _, err := tx.Exec("DELETE FROM schema_migrations WHERE version = ?", current); err != nil {
		return fmt.Errorf("failed to remove migration record: %w", err)
	}

	return tx.Commit()
}
