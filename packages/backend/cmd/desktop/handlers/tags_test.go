// Package handlers tests for tag REST API endpoints.
// These tests verify HTTP request handling, status codes, and responses.
package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kimhsiao/memonexus/backend/internal/db"
	"github.com/kimhsiao/memonexus/backend/internal/models"
	_ "modernc.org/sqlite"
)

// setupTestDB creates an in-memory database for testing
func setupTestDB(t *testing.T) (*sql.DB, func()) {
	testDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create basic schema
	_, err = testDB.Exec(`
		CREATE TABLE IF NOT EXISTS tags (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			color TEXT NOT NULL DEFAULT '#3B82F6',
			is_deleted INTEGER DEFAULT 0,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		);
	`)
	if err != nil {
		t.Fatalf("Failed to create tags table: %v", err)
	}

	cleanup := func() {
		testDB.Close()
	}

	return testDB, cleanup
}

func TestNewTagHandler(t *testing.T) {
	testDB, cleanup := setupTestDB(t)
	defer cleanup()

	repo := db.NewRepository(testDB)
	handler := NewTagHandler(repo)

	if handler == nil {
		t.Error("NewTagHandler should return non-nil handler")
	}

	if handler.repo != repo {
		t.Error("Handler repo should match provided repo")
	}
}

func TestTagHandler_ListTags(t *testing.T) {
	testDB, cleanup := setupTestDB(t)
	defer cleanup()

	repo := db.NewRepository(testDB)
	handler := NewTagHandler(repo)

	// Create a test tag
	testTag := &models.Tag{
		ID:    "test-tag-1",
		Name:  "Test Tag",
		Color: "#FF0000",
	}
	if err := repo.CreateTag(testTag); err != nil {
		t.Fatalf("Failed to create test tag: %v", err)
	}

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/tags", nil)
	w := httptest.NewRecorder()

	// Call handler
	handler.ListTags(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var tags []models.Tag
	if err := json.NewDecoder(w.Body).Decode(&tags); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(tags) != 1 {
		t.Errorf("Expected 1 tag, got %d", len(tags))
	}

	if tags[0].Name != "Test Tag" {
		t.Errorf("Expected tag name 'Test Tag', got '%s'", tags[0].Name)
	}
}

func TestTagHandler_ListTags_MethodNotAllowed(t *testing.T) {
	testDB, cleanup := setupTestDB(t)
	defer cleanup()

	repo := db.NewRepository(testDB)
	handler := NewTagHandler(repo)

	// Create POST request (should fail)
	req := httptest.NewRequest(http.MethodPost, "/tags", nil)
	w := httptest.NewRecorder()

	handler.ListTags(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestTagHandler_CreateTag(t *testing.T) {
	testDB, cleanup := setupTestDB(t)
	defer cleanup()

	repo := db.NewRepository(testDB)
	handler := NewTagHandler(repo)

	// Create request body
	requestBody := map[string]string{
		"name":  "New Tag",
		"color": "#00FF00",
	}
	body, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(http.MethodPost, "/tags", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call handler
	handler.CreateTag(w, req)

	// Check response
	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	var tag models.Tag
	if err := json.NewDecoder(w.Body).Decode(&tag); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if tag.Name != "New Tag" {
		t.Errorf("Expected name 'New Tag', got '%s'", tag.Name)
	}

	if tag.Color != "#00FF00" {
		t.Errorf("Expected color '#00FF00', got '%s'", tag.Color)
	}
}

func TestTagHandler_CreateTag_DefaultColor(t *testing.T) {
	testDB, cleanup := setupTestDB(t)
	defer cleanup()

	repo := db.NewRepository(testDB)
	handler := NewTagHandler(repo)

	// Create request without color
	requestBody := map[string]string{
		"name": "Tag Without Color",
	}
	body, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(http.MethodPost, "/tags", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreateTag(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	var tag models.Tag
	if err := json.NewDecoder(w.Body).Decode(&tag); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if tag.Color != "#3B82F6" {
		t.Errorf("Expected default color '#3B82F6', got '%s'", tag.Color)
	}
}

func TestTagHandler_CreateTag_MissingName(t *testing.T) {
	testDB, cleanup := setupTestDB(t)
	defer cleanup()

	repo := db.NewRepository(testDB)
	handler := NewTagHandler(repo)

	// Create request without name
	requestBody := map[string]string{
		"color": "#00FF00",
	}
	body, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(http.MethodPost, "/tags", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreateTag(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestTagHandler_CreateTag_NameTooLong(t *testing.T) {
	testDB, cleanup := setupTestDB(t)
	defer cleanup()

	repo := db.NewRepository(testDB)
	handler := NewTagHandler(repo)

	// Create request with name > 50 characters
	requestBody := map[string]string{
		"name": string(make([]byte, 51)), // 51 characters
	}
	body, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(http.MethodPost, "/tags", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreateTag(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	if !bytes.Contains(w.Body.Bytes(), []byte("50 characters")) {
		t.Errorf("Expected error about name length, got: %s", w.Body.String())
	}
}

func TestTagHandler_CreateTag_InvalidJSON(t *testing.T) {
	testDB, cleanup := setupTestDB(t)
	defer cleanup()

	repo := db.NewRepository(testDB)
	handler := NewTagHandler(repo)

	req := httptest.NewRequest(http.MethodPost, "/tags", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreateTag(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestTagHandler_GetTag(t *testing.T) {
	testDB, cleanup := setupTestDB(t)
	defer cleanup()

	repo := db.NewRepository(testDB)
	handler := NewTagHandler(repo)

	// Create a test tag
	testTag := &models.Tag{
		Name:  "Test Tag",
		Color: "#FF0000",
	}
	if err := repo.CreateTag(testTag); err != nil {
		t.Fatalf("Failed to create test tag: %v", err)
	}

	// Create request - use the ID returned by CreateTag
	req := httptest.NewRequest(http.MethodGet, "/tags/"+string(testTag.ID), nil)
	w := httptest.NewRecorder()

	handler.GetTag(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var tag models.Tag
	if err := json.NewDecoder(w.Body).Decode(&tag); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if tag.ID != testTag.ID {
		t.Errorf("Expected ID '%s', got '%s'", testTag.ID, tag.ID)
	}
}

func TestTagHandler_GetTag_NotFound(t *testing.T) {
	testDB, cleanup := setupTestDB(t)
	defer cleanup()

	repo := db.NewRepository(testDB)
	handler := NewTagHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/tags/nonexistent", nil)
	w := httptest.NewRecorder()

	handler.GetTag(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestTagHandler_UpdateTag(t *testing.T) {
	testDB, cleanup := setupTestDB(t)
	defer cleanup()

	repo := db.NewRepository(testDB)
	handler := NewTagHandler(repo)

	// Create a test tag
	testTag := &models.Tag{
		Name:  "Original Name",
		Color: "#FF0000",
	}
	if err := repo.CreateTag(testTag); err != nil {
		t.Fatalf("Failed to create test tag: %v", err)
	}

	// Create update request
	requestBody := map[string]string{
		"name":  "Updated Name",
		"color": "#0000FF",
	}
	body, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(http.MethodPut, "/tags/"+string(testTag.ID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.UpdateTag(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var tag models.Tag
	if err := json.NewDecoder(w.Body).Decode(&tag); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if tag.Name != "Updated Name" {
		t.Errorf("Expected name 'Updated Name', got '%s'", tag.Name)
	}

	if tag.Color != "#0000FF" {
		t.Errorf("Expected color '#0000FF', got '%s'", tag.Color)
	}
}

func TestTagHandler_UpdateTag_NameTooLong(t *testing.T) {
	testDB, cleanup := setupTestDB(t)
	defer cleanup()

	repo := db.NewRepository(testDB)
	handler := NewTagHandler(repo)

	// Create a test tag
	testTag := &models.Tag{
		Name:  "Original",
		Color: "#FF0000",
	}
	if err := repo.CreateTag(testTag); err != nil {
		t.Fatalf("Failed to create test tag: %v", err)
	}

	// Create update request with name too long
	longName := string(make([]byte, 51))
	requestBody := map[string]string{
		"name": longName,
	}
	body, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(http.MethodPut, "/tags/"+string(testTag.ID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.UpdateTag(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestTagHandler_UpdateTag_NotFound(t *testing.T) {
	testDB, cleanup := setupTestDB(t)
	defer cleanup()

	repo := db.NewRepository(testDB)
	handler := NewTagHandler(repo)

	requestBody := map[string]string{
		"name": "New Name",
	}
	body, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(http.MethodPut, "/tags/nonexistent", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.UpdateTag(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestTagHandler_DeleteTag(t *testing.T) {
	testDB, cleanup := setupTestDB(t)
	defer cleanup()

	repo := db.NewRepository(testDB)
	handler := NewTagHandler(repo)

	// Create a test tag
	testTag := &models.Tag{
		Name:  "To Delete",
		Color: "#FF0000",
	}
	if err := repo.CreateTag(testTag); err != nil {
		t.Fatalf("Failed to create test tag: %v", err)
	}

	// Create delete request
	req := httptest.NewRequest(http.MethodDelete, "/tags/"+string(testTag.ID), nil)
	w := httptest.NewRecorder()

	handler.DeleteTag(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", w.Code)
	}

	// Verify tag is soft-deleted
	deletedTag, err := repo.GetTag(string(testTag.ID))
	if err != nil {
		t.Errorf("Should still be able to retrieve soft-deleted tag: %v", err)
	}
	if !deletedTag.IsDeleted {
		t.Error("Tag should be marked as deleted (soft delete)")
	}
}

func TestTagHandler_DeleteTag_NotFound(t *testing.T) {
	testDB, cleanup := setupTestDB(t)
	defer cleanup()

	repo := db.NewRepository(testDB)
	handler := NewTagHandler(repo)

	req := httptest.NewRequest(http.MethodDelete, "/tags/nonexistent", nil)
	w := httptest.NewRecorder()

	handler.DeleteTag(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}
