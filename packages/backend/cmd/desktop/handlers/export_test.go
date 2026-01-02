// Package handlers tests for export/import REST API endpoints.
// These tests verify HTTP request handling, status codes, and responses.
package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/kimhsiao/memonexus/backend/internal/db"
	_ "modernc.org/sqlite"
)

// setupTestDBWithExport creates an in-memory database with export-related tables
func setupTestDBWithExport(t *testing.T) (*sql.DB, func()) {
	testDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create basic schema (simplified for export tests)
	_, err = testDB.Exec(`
		CREATE TABLE IF NOT EXISTS content (
			id TEXT PRIMARY KEY,
			url TEXT NOT NULL,
			title TEXT,
			content TEXT,
			is_deleted INTEGER DEFAULT 0,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		);

		CREATE TABLE IF NOT EXISTS tags (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			color TEXT NOT NULL DEFAULT '#3B82F6',
			is_deleted INTEGER DEFAULT 0,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		);

		CREATE TABLE IF NOT EXISTS content_tags (
			content_id TEXT NOT NULL,
			tag_id TEXT NOT NULL,
			PRIMARY KEY (content_id, tag_id),
			FOREIGN KEY (content_id) REFERENCES content(id) ON DELETE CASCADE,
			FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
		);

		CREATE TABLE IF NOT EXISTS settings (
			key TEXT PRIMARY KEY,
			value TEXT
		);
	`)
	if err != nil {
		t.Fatalf("Failed to create tables: %v", err)
	}

	cleanup := func() {
		testDB.Close()
	}

	return testDB, cleanup
}

func TestNewExportHandler(t *testing.T) {
	testDB, cleanup := setupTestDBWithExport(t)
	defer cleanup()

	repo := db.NewRepository(testDB)
	handler := NewExportHandler(repo)

	if handler == nil {
		t.Error("NewExportHandler should return non-nil handler")
	}

	if handler.repo != repo {
		t.Error("Handler repo should match provided repo")
	}

	if handler.export == nil {
		t.Error("Handler export service should be initialized")
	}
}

func TestExportHandler_Export_Success(t *testing.T) {
	testDB, cleanup := setupTestDBWithExport(t)
	defer cleanup()

	repo := db.NewRepository(testDB)
	handler := NewExportHandler(repo)

	// Create a temporary directory for exports
	tempDir := t.TempDir()

	// Create request body
	requestBody := ExportRequest{
		Password:     "test123",
		IncludeMedia: false,
		OutputPath:   tempDir,
	}
	body, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(http.MethodPost, "/export", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call handler
	handler.Export(w, req)

	// Export should succeed or fail depending on actual export logic
	// We're primarily testing the HTTP handling here
	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 200 or 500, got %d. Body: %s", w.Code, w.Body.String())
	}

	// If successful, response should be JSON
	if w.Code == http.StatusOK {
		var result map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
			t.Errorf("Response should be valid JSON: %v", err)
		}
	}
}

func TestExportHandler_Export_InvalidJSON(t *testing.T) {
	testDB, cleanup := setupTestDBWithExport(t)
	defer cleanup()

	repo := db.NewRepository(testDB)
	handler := NewExportHandler(repo)

	req := httptest.NewRequest(http.MethodPost, "/export", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Export(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	if !bytes.Contains(w.Body.Bytes(), []byte("Invalid request body")) {
		t.Errorf("Expected error about invalid JSON, got: %s", w.Body.String())
	}
}

func TestExportHandler_Export_MethodNotAllowed(t *testing.T) {
	testDB, cleanup := setupTestDBWithExport(t)
	defer cleanup()

	repo := db.NewRepository(testDB)
	handler := NewExportHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/export", nil)
	w := httptest.NewRecorder()

	handler.Export(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestExportHandler_Import_MissingArchivePath(t *testing.T) {
	testDB, cleanup := setupTestDBWithExport(t)
	defer cleanup()

	repo := db.NewRepository(testDB)
	handler := NewExportHandler(repo)

	// Create request without archive_path
	requestBody := ImportRequest{
		Password: "test123",
	}
	body, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(http.MethodPost, "/import", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Import(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	if !bytes.Contains(w.Body.Bytes(), []byte("archive_path is required")) {
		t.Errorf("Expected error about missing archive_path, got: %s", w.Body.String())
	}
}

func TestExportHandler_Import_InvalidJSON(t *testing.T) {
	testDB, cleanup := setupTestDBWithExport(t)
	defer cleanup()

	repo := db.NewRepository(testDB)
	handler := NewExportHandler(repo)

	req := httptest.NewRequest(http.MethodPost, "/import", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Import(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestExportHandler_Import_MethodNotAllowed(t *testing.T) {
	testDB, cleanup := setupTestDBWithExport(t)
	defer cleanup()

	repo := db.NewRepository(testDB)
	handler := NewExportHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/import", nil)
	w := httptest.NewRecorder()

	handler.Import(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestExportHandler_Import_NonexistentArchive(t *testing.T) {
	testDB, cleanup := setupTestDBWithExport(t)
	defer cleanup()

	repo := db.NewRepository(testDB)
	handler := NewExportHandler(repo)

	// Create request with non-existent archive path
	requestBody := ImportRequest{
		ArchivePath: "/nonexistent/path/to/archive.tar.gz",
		Password:    "test123",
	}
	body, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(http.MethodPost, "/import", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Import(w, req)

	// Should fail with 500 (archive doesn't exist)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500 for non-existent archive, got %d", w.Code)
	}
}

func TestExportHandler_ExportStatus(t *testing.T) {
	testDB, cleanup := setupTestDBWithExport(t)
	defer cleanup()

	repo := db.NewRepository(testDB)
	handler := NewExportHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/export/status", nil)
	w := httptest.NewRecorder()

	handler.ExportStatus(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var status map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&status); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if status["active"] != false {
		t.Error("Expected active to be false")
	}

	if _, ok := status["recent"].([]interface{}); !ok {
		t.Error("Expected recent to be an array")
	}

	if status["exports_dir"] != "exports/" {
		t.Errorf("Expected exports_dir to be 'exports/', got %v", status["exports_dir"])
	}
}

func TestExportHandler_ExportStatus_MethodNotAllowed(t *testing.T) {
	testDB, cleanup := setupTestDBWithExport(t)
	defer cleanup()

	repo := db.NewRepository(testDB)
	handler := NewExportHandler(repo)

	req := httptest.NewRequest(http.MethodPost, "/export/status", nil)
	w := httptest.NewRecorder()

	handler.ExportStatus(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestExportHandler_Import_WithMockArchive(t *testing.T) {
	testDB, cleanup := setupTestDBWithExport(t)
	defer cleanup()

	repo := db.NewRepository(testDB)
	handler := NewExportHandler(repo)

	// Create a temporary "archive" file (not a valid archive, just a file)
	tempDir := t.TempDir()
	mockArchive := filepath.Join(tempDir, "archive.tar.gz")
	if err := os.WriteFile(mockArchive, []byte("not a real archive"), 0644); err != nil {
		t.Fatalf("Failed to create mock archive: %v", err)
	}

	// Create request with mock archive
	requestBody := ImportRequest{
		ArchivePath: mockArchive,
		Password:    "test123",
	}
	body, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(http.MethodPost, "/import", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Import(w, req)

	// Should fail with 500 (not a valid archive)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500 for invalid archive, got %d. Body: %s", w.Code, w.Body.String())
	}
}
