// Package handlers tests for AI REST API endpoints.
// These tests verify configuration management, validation, and error handling.
// Note: Full AI functionality tests would require mocking external AI services.
package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kimhsiao/memonexus/backend/internal/db"
	"github.com/kimhsiao/memonexus/backend/internal/services"
	_ "modernc.org/sqlite"
)

// setupTestDBWithAI creates an in-memory database with AI config table
func setupTestDBWithAI(t *testing.T) (*sql.DB, func()) {
	testDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create AI config table
	_, err = testDB.Exec(`
		CREATE TABLE IF NOT EXISTS ai_config (
			id TEXT PRIMARY KEY,
			provider TEXT NOT NULL CHECK(provider IN ('openai', 'claude', 'ollama')),
			api_endpoint TEXT NOT NULL CHECK(length(api_endpoint) > 0),
			api_key_encrypted TEXT NOT NULL CHECK(length(api_key_encrypted) > 0),
			model_name TEXT NOT NULL CHECK(length(model_name) > 0),
			max_tokens INTEGER DEFAULT 1000 CHECK(max_tokens > 0),
			is_enabled INTEGER NOT NULL DEFAULT 0 CHECK(is_enabled IN (0, 1)),
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		);

		CREATE INDEX idx_ai_config_is_enabled ON ai_config(is_enabled);
	`)
	if err != nil {
		testDB.Close()
		t.Fatalf("Failed to create ai_config table: %v", err)
	}

	// Create content_items table for analysis tests
	_, err = testDB.Exec(`
		CREATE TABLE IF NOT EXISTS content_items (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			content_text TEXT NOT NULL DEFAULT '',
			source_url TEXT,
			media_type TEXT NOT NULL DEFAULT 'web',
			tags TEXT DEFAULT '',
			summary TEXT,
			is_deleted INTEGER DEFAULT 0,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL,
			version INTEGER NOT NULL DEFAULT 1,
			content_hash TEXT
		);
	`)
	if err != nil {
		testDB.Close()
		t.Fatalf("Failed to create content_items table: %v", err)
	}

	cleanup := func() {
		testDB.Close()
	}

	return testDB, cleanup
}

func TestNewAIHandler(t *testing.T) {
	testDB, cleanup := setupTestDBWithAI(t)
	defer cleanup()

	repo := db.NewRepository(testDB)
	analysisSvc := services.NewAnalysisService(nil)
	machineID := "test-machine"

	handler := NewAIHandler(repo, analysisSvc, machineID)

	if handler == nil {
		t.Error("NewAIHandler should return non-nil handler")
	}

	if handler.repo != repo {
		t.Error("Handler repo should match provided repo")
	}

	if handler.machineID != machineID {
		t.Errorf("Handler machineID should be %s, got %s", machineID, handler.machineID)
	}
}

func TestNewAIHandler_DefaultMachineID(t *testing.T) {
	testDB, cleanup := setupTestDBWithAI(t)
	defer cleanup()

	repo := db.NewRepository(testDB)
	analysisSvc := services.NewAnalysisService(nil)

	handler := NewAIHandler(repo, analysisSvc, "")

	if handler.machineID != "default" {
		t.Errorf("Expected default machineID 'default', got %s", handler.machineID)
	}
}

func TestAIHandler_SetWebSocketHub(t *testing.T) {
	testDB, cleanup := setupTestDBWithAI(t)
	defer cleanup()

	repo := db.NewRepository(testDB)
	analysisSvc := services.NewAnalysisService(nil)
	handler := NewAIHandler(repo, analysisSvc, "test")

	if handler.wsHub != nil {
		t.Error("wsHub should be nil initially")
	}

	// Create a mock broadcaster
	mockHub := &mockWSBroadcaster{}
	handler.SetWebSocketHub(mockHub)

	if handler.wsHub != mockHub {
		t.Error("wsHub should be set to mockHub")
	}
}

func TestAIHandler_GetAIConfig_NoConfig(t *testing.T) {
	testDB, cleanup := setupTestDBWithAI(t)
	defer cleanup()

	repo := db.NewRepository(testDB)
	analysisSvc := services.NewAnalysisService(nil) // No AI configured
	handler := NewAIHandler(repo, analysisSvc, "test")

	req := httptest.NewRequest(http.MethodGet, "/ai/config", nil)
	w := httptest.NewRecorder()

	handler.GetAIConfig(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["enabled"] != false {
		t.Errorf("Expected enabled=false, got %v", response["enabled"])
	}
}

func TestAIHandler_GetAIConfig_MethodNotAllowed(t *testing.T) {
	testDB, cleanup := setupTestDBWithAI(t)
	defer cleanup()

	repo := db.NewRepository(testDB)
	analysisSvc := services.NewAnalysisService(nil)
	handler := NewAIHandler(repo, analysisSvc, "test")

	req := httptest.NewRequest(http.MethodPost, "/ai/config", nil)
	w := httptest.NewRecorder()

	handler.GetAIConfig(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestAIHandler_SetAIConfig_InvalidJSON(t *testing.T) {
	testDB, cleanup := setupTestDBWithAI(t)
	defer cleanup()

	repo := db.NewRepository(testDB)
	analysisSvc := services.NewAnalysisService(nil)
	handler := NewAIHandler(repo, analysisSvc, "test")

	req := httptest.NewRequest(http.MethodPost, "/ai/config", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.SetAIConfig(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	if !bytes.Contains(w.Body.Bytes(), []byte("Invalid request body")) {
		t.Errorf("Expected error about invalid body, got: %s", w.Body.String())
	}
}

func TestAIHandler_SetAIConfig_MissingProvider(t *testing.T) {
	testDB, cleanup := setupTestDBWithAI(t)
	defer cleanup()

	repo := db.NewRepository(testDB)
	analysisSvc := services.NewAnalysisService(nil)
	handler := NewAIHandler(repo, analysisSvc, "test")

	requestBody := map[string]string{
		"api_endpoint": "https://api.openai.com",
		"api_key":      "sk-test",
		"model_name":   "gpt-4",
	}
	body, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(http.MethodPost, "/ai/config", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.SetAIConfig(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestAIHandler_SetAIConfig_InvalidProvider(t *testing.T) {
	testDB, cleanup := setupTestDBWithAI(t)
	defer cleanup()

	repo := db.NewRepository(testDB)
	analysisSvc := services.NewAnalysisService(nil)
	handler := NewAIHandler(repo, analysisSvc, "test")

	requestBody := map[string]string{
		"provider":     "invalid",
		"api_endpoint": "https://api.openai.com",
		"api_key":      "sk-test",
		"model_name":   "gpt-4",
	}
	body, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(http.MethodPost, "/ai/config", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.SetAIConfig(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	if !bytes.Contains(w.Body.Bytes(), []byte("Invalid provider")) {
		t.Errorf("Expected error about invalid provider, got: %s", w.Body.String())
	}
}

func TestAIHandler_SetAIConfig_MissingEndpoint(t *testing.T) {
	testDB, cleanup := setupTestDBWithAI(t)
	defer cleanup()

	repo := db.NewRepository(testDB)
	analysisSvc := services.NewAnalysisService(nil)
	handler := NewAIHandler(repo, analysisSvc, "test")

	requestBody := map[string]string{
		"provider":   "openai",
		"api_key":    "sk-test",
		"model_name": "gpt-4",
	}
	body, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(http.MethodPost, "/ai/config", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.SetAIConfig(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	if !bytes.Contains(w.Body.Bytes(), []byte("api_endpoint is required")) {
		t.Errorf("Expected error about missing endpoint, got: %s", w.Body.String())
	}
}

func TestAIHandler_SetAIConfig_MissingAPIKey(t *testing.T) {
	testDB, cleanup := setupTestDBWithAI(t)
	defer cleanup()

	repo := db.NewRepository(testDB)
	analysisSvc := services.NewAnalysisService(nil)
	handler := NewAIHandler(repo, analysisSvc, "test")

	requestBody := map[string]string{
		"provider":     "openai",
		"api_endpoint": "https://api.openai.com",
		"model_name":   "gpt-4",
	}
	body, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(http.MethodPost, "/ai/config", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.SetAIConfig(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	if !bytes.Contains(w.Body.Bytes(), []byte("api_key is required")) {
		t.Errorf("Expected error about missing API key, got: %s", w.Body.String())
	}
}

func TestAIHandler_SetAIConfig_MissingModelName(t *testing.T) {
	testDB, cleanup := setupTestDBWithAI(t)
	defer cleanup()

	repo := db.NewRepository(testDB)
	analysisSvc := services.NewAnalysisService(nil)
	handler := NewAIHandler(repo, analysisSvc, "test")

	requestBody := map[string]string{
		"provider":     "openai",
		"api_endpoint": "https://api.openai.com",
		"api_key":      "sk-test",
	}
	body, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(http.MethodPost, "/ai/config", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.SetAIConfig(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	if !bytes.Contains(w.Body.Bytes(), []byte("model_name is required")) {
		t.Errorf("Expected error about missing model name, got: %s", w.Body.String())
	}
}

func TestAIHandler_SetAIConfig_DefaultMaxTokens(t *testing.T) {
	testDB, cleanup := setupTestDBWithAI(t)
	defer cleanup()

	repo := db.NewRepository(testDB)
	analysisSvc := services.NewAnalysisService(nil)
	handler := NewAIHandler(repo, analysisSvc, "test")

	// Request with max_tokens = 0, should default to 1000
	requestBody := map[string]interface{}{
		"provider":    "openai",
		"api_endpoint": "https://api.openai.com",
		"api_key":     "sk-test",
		"model_name":  "gpt-4",
		"max_tokens":  0,
	}
	body, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(http.MethodPost, "/ai/config", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// This will fail at encryption or database step, but we can verify max_tokens is processed
	handler.SetAIConfig(w, req)

	// The handler should process max_tokens and attempt encryption
	// Error will occur at encryption or database, but validation should pass
	if w.Code == http.StatusBadRequest {
		// Check if error is about max_tokens (would be wrong)
		if bytes.Contains(w.Body.Bytes(), []byte("max_tokens")) {
			t.Error("max_tokens should default to 1000, not cause validation error")
		}
	}
}

func TestAIHandler_SetAIConfig_ValidProviders(t *testing.T) {
	testDB, cleanup := setupTestDBWithAI(t)
	defer cleanup()

	repo := db.NewRepository(testDB)
	analysisSvc := services.NewAnalysisService(nil)
	handler := NewAIHandler(repo, analysisSvc, "test")

	validProviders := []string{"openai", "claude", "ollama"}

	for _, provider := range validProviders {
		requestBody := map[string]string{
			"provider":     provider,
			"api_endpoint": "https://api.example.com",
			"api_key":      "sk-test",
			"model_name":   "test-model",
		}
		body, _ := json.Marshal(requestBody)

		req := httptest.NewRequest(http.MethodPost, "/ai/config", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.SetAIConfig(w, req)

		// Should fail at encryption or database save, but NOT at provider validation
		if w.Code == http.StatusBadRequest {
			if bytes.Contains(w.Body.Bytes(), []byte("Invalid provider")) {
				t.Errorf("Provider %s should be valid, got error: %s", provider, w.Body.String())
			}
		}
	}
}

func TestAIHandler_SetAIConfig_MethodNotAllowed(t *testing.T) {
	testDB, cleanup := setupTestDBWithAI(t)
	defer cleanup()

	repo := db.NewRepository(testDB)
	analysisSvc := services.NewAnalysisService(nil)
	handler := NewAIHandler(repo, analysisSvc, "test")

	req := httptest.NewRequest(http.MethodGet, "/ai/config", nil)
	w := httptest.NewRecorder()

	handler.SetAIConfig(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestAIHandler_DeleteAIConfig_MethodNotAllowed(t *testing.T) {
	testDB, cleanup := setupTestDBWithAI(t)
	defer cleanup()

	repo := db.NewRepository(testDB)
	analysisSvc := services.NewAnalysisService(nil)
	handler := NewAIHandler(repo, analysisSvc, "test")

	req := httptest.NewRequest(http.MethodGet, "/ai/config", nil)
	w := httptest.NewRecorder()

	handler.DeleteAIConfig(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestAIHandler_DeleteAIConfig_Success(t *testing.T) {
	testDB, cleanup := setupTestDBWithAI(t)
	defer cleanup()

	repo := db.NewRepository(testDB)
	analysisSvc := services.NewAnalysisService(nil)
	handler := NewAIHandler(repo, analysisSvc, "test")

	req := httptest.NewRequest(http.MethodDelete, "/ai/config", nil)
	w := httptest.NewRecorder()

	handler.DeleteAIConfig(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["status"] != "success" {
		t.Errorf("Expected status success, got %v", response["status"])
	}
}

func TestAIHandler_GenerateSummary_MethodNotAllowed(t *testing.T) {
	testDB, cleanup := setupTestDBWithAI(t)
	defer cleanup()

	repo := db.NewRepository(testDB)
	analysisSvc := services.NewAnalysisService(nil)
	handler := NewAIHandler(repo, analysisSvc, "test")

	req := httptest.NewRequest(http.MethodGet, "/content/summary", nil)
	w := httptest.NewRecorder()

	handler.GenerateSummary(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestAIHandler_GenerateSummary_MissingID(t *testing.T) {
	testDB, cleanup := setupTestDBWithAI(t)
	defer cleanup()

	repo := db.NewRepository(testDB)
	analysisSvc := services.NewAnalysisService(nil)
	handler := NewAIHandler(repo, analysisSvc, "test")

	req := httptest.NewRequest(http.MethodPost, "/content/summary", nil)
	w := httptest.NewRecorder()

	handler.GenerateSummary(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	if !bytes.Contains(w.Body.Bytes(), []byte("content id is required")) {
		t.Errorf("Expected error about missing content ID, got: %s", w.Body.String())
	}
}

func TestAIHandler_GenerateSummary_ContentNotFound(t *testing.T) {
	testDB, cleanup := setupTestDBWithAI(t)
	defer cleanup()

	repo := db.NewRepository(testDB)
	analysisSvc := services.NewAnalysisService(nil)
	handler := NewAIHandler(repo, analysisSvc, "test")

	req := httptest.NewRequest(http.MethodPost, "/content/summary?id=nonexistent", nil)
	w := httptest.NewRecorder()

	handler.GenerateSummary(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestAIHandler_ExtractKeywords_MethodNotAllowed(t *testing.T) {
	testDB, cleanup := setupTestDBWithAI(t)
	defer cleanup()

	repo := db.NewRepository(testDB)
	analysisSvc := services.NewAnalysisService(nil)
	handler := NewAIHandler(repo, analysisSvc, "test")

	req := httptest.NewRequest(http.MethodGet, "/content/keywords", nil)
	w := httptest.NewRecorder()

	handler.ExtractKeywords(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestAIHandler_ExtractKeywords_MissingID(t *testing.T) {
	testDB, cleanup := setupTestDBWithAI(t)
	defer cleanup()

	repo := db.NewRepository(testDB)
	analysisSvc := services.NewAnalysisService(nil)
	handler := NewAIHandler(repo, analysisSvc, "test")

	req := httptest.NewRequest(http.MethodPost, "/content/keywords", nil)
	w := httptest.NewRecorder()

	handler.ExtractKeywords(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	if !bytes.Contains(w.Body.Bytes(), []byte("content id is required")) {
		t.Errorf("Expected error about missing content ID, got: %s", w.Body.String())
	}
}

func TestAIHandler_ExtractKeywords_ContentNotFound(t *testing.T) {
	testDB, cleanup := setupTestDBWithAI(t)
	defer cleanup()

	repo := db.NewRepository(testDB)
	analysisSvc := services.NewAnalysisService(nil)
	handler := NewAIHandler(repo, analysisSvc, "test")

	req := httptest.NewRequest(http.MethodPost, "/content/keywords?id=nonexistent", nil)
	w := httptest.NewRecorder()

	handler.ExtractKeywords(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

// mockWSBroadcaster is a mock WebSocket broadcaster for testing
type mockWSBroadcaster struct {
	startedCount    int
	completedCount  int
	failedCount     int
	lastContentID   string
	lastOperation   string
	lastError        string
	lastFallback     string
}

func (m *mockWSBroadcaster) BroadcastAnalysisStarted(contentID string, operation string) {
	m.startedCount++
	m.lastContentID = contentID
	m.lastOperation = operation
}

func (m *mockWSBroadcaster) BroadcastAnalysisCompleted(contentID string, result map[string]interface{}) {
	m.completedCount++
	m.lastContentID = contentID
}

func (m *mockWSBroadcaster) BroadcastAnalysisFailed(contentID string, errMsg string, fallbackMethod string) {
	m.failedCount++
	m.lastContentID = contentID
	m.lastError = errMsg
	m.lastFallback = fallbackMethod
}

func TestAIHandler_GenerateSummary_WebSocketBroadcast(t *testing.T) {
	testDB, cleanup := setupTestDBWithAI(t)
	defer cleanup()

	// Insert test content using the helper from search_test.go
	contentID := insertTestContentItem(t, testDB, "Test Item", "Test content for summary", "web", "", time.Now().Unix())

	repo := db.NewRepository(testDB)
	analysisSvc := services.NewAnalysisService(nil)
	mockHub := &mockWSBroadcaster{}
	handler := NewAIHandler(repo, analysisSvc, "test")
	handler.SetWebSocketHub(mockHub)

	req := httptest.NewRequest(http.MethodPost, "/content/summary?id="+contentID, nil)
	w := httptest.NewRecorder()

	handler.GenerateSummary(w, req)

	// Verify broadcast events were called
	if mockHub.startedCount == 0 {
		t.Error("Expected BroadcastAnalysisStarted to be called")
	}
	if mockHub.completedCount == 0 {
		t.Error("Expected BroadcastAnalysisCompleted to be called")
	}
	if mockHub.lastContentID != contentID {
		t.Errorf("Expected contentID %s, got %s", contentID, mockHub.lastContentID)
	}
	if mockHub.lastOperation != "summary" {
		t.Errorf("Expected operation 'summary', got %s", mockHub.lastOperation)
	}
}
