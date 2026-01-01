// Package sync tests for sync engine functionality.
package sync

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/kimhsiao/memonexus/backend/internal/models"
	"github.com/kimhsiao/memonexus/backend/internal/uuid"
)

// testEventHandler is a test implementation of SyncEventHandler.
type testEventHandler struct {
	events []SyncEvent
}

func (h *testEventHandler) OnSyncEvent(event SyncEvent) {
	h.events = append(h.events, event)
}

// TestNewSyncEngine verifies engine creation.
func TestNewSyncEngine(t *testing.T) {
	engine := NewSyncEngine(nil, nil)

	if engine == nil {
		t.Fatal("NewSyncEngine() returned nil")
	}

	if engine.Status() != SyncStatusIdle {
		t.Errorf("status = %v, want SyncStatusIdle", engine.Status())
	}

	if engine.LastSync() != nil {
		t.Error("lastSync should be nil initially")
	}

	if engine.PendingChanges() != 0 {
		t.Error("pending should be 0 initially")
	}

	if engine.LastError() != nil {
		t.Error("lastErr should be nil initially")
	}
}

// TestSetEventHandler verifies event handler setting.
func TestSetEventHandler(t *testing.T) {
	engine := NewSyncEngine(nil, nil)

	handler := &testEventHandler{}

	engine.SetEventHandler(handler)

	// Verify handler is set by emitting an event
	engine.emitEvent(SyncEvent{Type: SyncEventStarted})

	// Give goroutine time to execute
	time.Sleep(10 * time.Millisecond)

	if len(handler.events) != 1 {
		t.Errorf("handler events count = %d, want 1", len(handler.events))
	}

	if handler.events[0].Type != SyncEventStarted {
		t.Errorf("event type = %v, want SyncEventStarted", handler.events[0].Type)
	}
}

// TestSetEventHandler_nil verifies nil handler is handled.
func TestSetEventHandler_nil(t *testing.T) {
	engine := NewSyncEngine(nil, nil)

	// Should not panic with nil handler
	engine.SetEventHandler(nil)

	// Emit event - should not panic
	engine.emitEvent(SyncEvent{Type: SyncEventStarted})

	time.Sleep(10 * time.Millisecond)
}

// TestGetErrorHistory verifies error history retrieval.
func TestGetErrorHistory(t *testing.T) {
	engine := NewSyncEngine(nil, nil)

	// Initially empty
	history := engine.GetErrorHistory()
	if history == nil {
		t.Error("history should not be nil")
	}
	if len(history) != 0 {
		t.Errorf("history length = %d, want 0", len(history))
	}

	// Record some errors
	testErr := errors.New("test error")
	engine.recordError("item1", "upload", testErr)
	engine.recordError("item2", "download", errors.New("another error"))

	history = engine.GetErrorHistory()
	if len(history) != 2 {
		t.Errorf("history length = %d, want 2", len(history))
	}

	// Verify it returns a copy, not the original slice
	history[0] = SyncErrorEntry{}
	newHistory := engine.GetErrorHistory()
	if len(newHistory) != 2 {
		t.Errorf("modifying returned history affected original")
	}
}

// TestClearErrorHistory verifies error history clearing.
func TestClearErrorHistory(t *testing.T) {
	engine := NewSyncEngine(nil, nil)

	// Record some errors
	engine.recordError("item1", "upload", errors.New("test error"))
	engine.recordError("item2", "download", errors.New("another error"))

	if len(engine.GetErrorHistory()) != 2 {
		t.Error("errors not recorded")
	}

	// Clear history
	engine.ClearErrorHistory()

	history := engine.GetErrorHistory()
	if len(history) != 0 {
		t.Errorf("history length after clear = %d, want 0", len(history))
	}
}

// TestStatus verifies status getter.
func TestStatus(t *testing.T) {
	engine := NewSyncEngine(nil, nil)

	if engine.Status() != SyncStatusIdle {
		t.Errorf("Status() = %v, want SyncStatusIdle", engine.Status())
	}
}

// TestLastSync verifies last sync timestamp getter.
func TestLastSync(t *testing.T) {
	engine := NewSyncEngine(nil, nil)

	if engine.LastSync() != nil {
		t.Error("LastSync() should be nil initially")
	}
}

// TestPendingChanges verifies pending changes count getter.
func TestPendingChanges(t *testing.T) {
	engine := NewSyncEngine(nil, nil)

	if engine.PendingChanges() != 0 {
		t.Errorf("PendingChanges() = %d, want 0", engine.PendingChanges())
	}
}

// TestRecordError verifies error recording with history limits.
func TestRecordError(t *testing.T) {
	engine := NewSyncEngine(nil, nil)

	// Record more errors than maxErrorHistory
	for i := 0; i < 150; i++ {
		engine.recordError("item", "test", errors.New("test error"))
	}

	history := engine.GetErrorHistory()
	// Should be capped at maxErrorHistory
	if len(history) > maxErrorHistory {
		t.Errorf("history length = %d, should be capped at %d", len(history), maxErrorHistory)
	}
}

// TestEmitEvent verifies event emission with timestamp.
func TestEmitEvent(t *testing.T) {
	engine := NewSyncEngine(nil, nil)
	handler := &testEventHandler{}

	engine.SetEventHandler(handler)

	// Emit event without timestamp
	engine.emitEvent(SyncEvent{Type: SyncEventStarted, Message: "Test"})

	time.Sleep(10 * time.Millisecond)

	if len(handler.events) != 1 {
		t.Fatalf("events count = %d, want 1", len(handler.events))
	}

	if handler.events[0].Type != SyncEventStarted {
		t.Errorf("event type = %v, want SyncEventStarted", handler.events[0].Type)
	}

	if handler.events[0].Message != "Test" {
		t.Errorf("event message = %q, want 'Test'", handler.events[0].Message)
	}

	if handler.events[0].Timestamp.IsZero() {
		t.Error("event timestamp should be set automatically")
	}
}

// TestEmitEvent_preservesTimestamp verifies existing timestamps are preserved.
func TestEmitEvent_preservesTimestamp(t *testing.T) {
	engine := NewSyncEngine(nil, nil)
	handler := &testEventHandler{}

	engine.SetEventHandler(handler)

	// Emit event with timestamp
	testTime := time.Now().Add(-1 * time.Hour)
	engine.emitEvent(SyncEvent{Type: SyncEventStarted, Timestamp: testTime})

	time.Sleep(10 * time.Millisecond)

	if len(handler.events) != 1 {
		t.Fatalf("events count = %d, want 1", len(handler.events))
	}

	if !handler.events[0].Timestamp.Equal(testTime) {
		t.Errorf("timestamp was not preserved, got %v, want %v", handler.events[0].Timestamp, testTime)
	}
}

// TestEmitEvent_nilHandler verifies nil handler doesn't cause panic.
func TestEmitEvent_nilHandler(t *testing.T) {
	engine := NewSyncEngine(nil, nil)

	// Should not panic
	engine.emitEvent(SyncEvent{Type: SyncEventStarted})

	time.Sleep(10 * time.Millisecond)
}

// =====================================================
// Serialization Tests
// =====================================================

// TestSerializeItem verifies JSON serialization of content items.
func TestSerializeItem(t *testing.T) {
	engine := NewSyncEngine(nil, nil)

	// Create a test content item
	item := &models.ContentItem{
		ID:          models.UUID(uuid.New()),
		Title:       "Test Article",
		ContentText: "Test content",
		SourceURL:   "https://example.com/test",
		MediaType:   "web",
		Tags:        "test,example",
		Summary:     "Test summary",
		IsDeleted:   false,
		CreatedAt:   time.Now().Unix(),
		UpdatedAt:   time.Now().Unix(),
		Version:     1,
		ContentHash: "abc123",
	}

	// Serialize the item
	data := engine.serializeItem(item)

	// Verify the data is valid JSON
	var decoded models.ContentItem
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("serializeItem produced invalid JSON: %v", err)
	}

	// Verify the decoded item matches the original
	if decoded.ID != item.ID {
		t.Errorf("ID = %s, want %s", decoded.ID, item.ID)
	}
	if decoded.Title != item.Title {
		t.Errorf("Title = %s, want %s", decoded.Title, item.Title)
	}
	if decoded.ContentText != item.ContentText {
		t.Errorf("ContentText = %s, want %s", decoded.ContentText, item.ContentText)
	}
	if decoded.MediaType != item.MediaType {
		t.Errorf("MediaType = %s, want %s", decoded.MediaType, item.MediaType)
	}
	if decoded.Tags != item.Tags {
		t.Errorf("Tags = %s, want %s", decoded.Tags, item.Tags)
	}
	if decoded.Version != item.Version {
		t.Errorf("Version = %d, want %d", decoded.Version, item.Version)
	}
}

// TestSerializeItem_nilItem verifies nil item handling.
func TestSerializeItem_nilItem(t *testing.T) {
	engine := NewSyncEngine(nil, nil)

	// Should not panic with nil item
	data := engine.serializeItem(nil)

	// json.Marshal(nil) returns "null"
	if string(data) != "null" {
		t.Errorf("serializeItem(nil) = %s, want 'null'", string(data))
	}
}

// TestDeserializeItem verifies JSON deserialization of content items.
func TestDeserializeItem(t *testing.T) {
	engine := NewSyncEngine(nil, nil)

	// Create valid JSON data
	item := &models.ContentItem{
		ID:          models.UUID(uuid.New()),
		Title:       "Test Article",
		ContentText: "Test content",
		MediaType:   "web",
		Tags:        "test,example",
		CreatedAt:   time.Now().Unix(),
		UpdatedAt:   time.Now().Unix(),
		Version:     1,
	}

	// Serialize to JSON
	data, _ := json.Marshal(item)

	// Deserialize the item
	decoded, err := engine.deserializeItem(data)
	if err != nil {
		t.Fatalf("deserializeItem failed: %v", err)
	}

	// Verify the decoded item matches
	if decoded.ID != item.ID {
		t.Errorf("ID = %s, want %s", decoded.ID, item.ID)
	}
	if decoded.Title != item.Title {
		t.Errorf("Title = %s, want %s", decoded.Title, item.Title)
	}
	if decoded.ContentText != item.ContentText {
		t.Errorf("ContentText = %s, want %s", decoded.ContentText, item.ContentText)
	}
}

// TestDeserializeItem_invalidJSON verifies error handling for invalid JSON.
func TestDeserializeItem_invalidJSON(t *testing.T) {
	engine := NewSyncEngine(nil, nil)

	// Test with invalid JSON
	invalidData := []byte("{invalid json}")

	_, err := engine.deserializeItem(invalidData)
	if err == nil {
		t.Error("deserializeItem with invalid JSON should return error")
	}
}

// TestDeserializeItem_emptyJSON verifies empty JSON handling.
func TestDeserializeItem_emptyJSON(t *testing.T) {
	engine := NewSyncEngine(nil, nil)

	// Test with empty JSON object
	emptyData := []byte("{}")

	item, err := engine.deserializeItem(emptyData)
	if err != nil {
		t.Fatalf("deserializeItem with empty JSON failed: %v", err)
	}

	// Verify item is created with zero values
	if item.ID != "" {
		t.Errorf("ID should be empty, got %s", item.ID)
	}
	if item.Title != "" {
		t.Errorf("Title should be empty, got %s", item.Title)
	}
}

// TestSerializeDeserializeRoundTrip verifies round-trip serialization.
func TestSerializeDeserializeRoundTrip(t *testing.T) {
	engine := NewSyncEngine(nil, nil)

	// Create a test item with all fields
	original := &models.ContentItem{
		ID:          models.UUID(uuid.New()),
		Title:       "Round Trip Test",
		ContentText: "Testing serialization round trip",
		SourceURL:   "https://example.com/roundtrip",
		MediaType:   "pdf",
		Tags:        "serialization,test",
		Summary:     "Testing summary",
		IsDeleted:   false,
		CreatedAt:   1234567890,
		UpdatedAt:   1234567900,
		Version:     5,
		ContentHash: "hash789",
	}

	// Serialize then deserialize
	data := engine.serializeItem(original)
	restored, err := engine.deserializeItem(data)

	if err != nil {
		t.Fatalf("round trip failed: %v", err)
	}

	// Verify all fields match
	if restored.ID != original.ID {
		t.Errorf("ID: got %s, want %s", restored.ID, original.ID)
	}
	if restored.Title != original.Title {
		t.Errorf("Title: got %s, want %s", restored.Title, original.Title)
	}
	if restored.ContentText != original.ContentText {
		t.Errorf("ContentText: got %s, want %s", restored.ContentText, original.ContentText)
	}
	if restored.SourceURL != original.SourceURL {
		t.Errorf("SourceURL: got %s, want %s", restored.SourceURL, original.SourceURL)
	}
	if restored.MediaType != original.MediaType {
		t.Errorf("MediaType: got %s, want %s", restored.MediaType, original.MediaType)
	}
	if restored.Tags != original.Tags {
		t.Errorf("Tags: got %s, want %s", restored.Tags, original.Tags)
	}
	if restored.Summary != original.Summary {
		t.Errorf("Summary: got %s, want %s", restored.Summary, original.Summary)
	}
	if restored.IsDeleted != original.IsDeleted {
		t.Errorf("IsDeleted: got %v, want %v", restored.IsDeleted, original.IsDeleted)
	}
	if restored.CreatedAt != original.CreatedAt {
		t.Errorf("CreatedAt: got %d, want %d", restored.CreatedAt, original.CreatedAt)
	}
	if restored.UpdatedAt != original.UpdatedAt {
		t.Errorf("UpdatedAt: got %d, want %d", restored.UpdatedAt, original.UpdatedAt)
	}
	if restored.Version != original.Version {
		t.Errorf("Version: got %d, want %d", restored.Version, original.Version)
	}
	if restored.ContentHash != original.ContentHash {
		t.Errorf("ContentHash: got %s, want %s", restored.ContentHash, original.ContentHash)
	}
}

// TestResolveConflicts verifies conflict resolution (currently placeholder).
func TestResolveConflicts(t *testing.T) {
	engine := NewSyncEngine(nil, nil)

	// resolveConflicts is currently a placeholder that returns nil
	// This test verifies it doesn't panic and returns expected value
	ctx := context.Background()

	conflicts := engine.resolveConflicts(ctx)

	// Should return nil (no conflicts resolved)
	if conflicts != nil {
		t.Errorf("resolveConflicts = %v, want nil", conflicts)
	}
}

// TestResolveConflicts_contextCancellation verifies context cancellation handling.
func TestResolveConflicts_contextCancellation(t *testing.T) {
	engine := NewSyncEngine(nil, nil)

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Should handle cancelled context gracefully
	conflicts := engine.resolveConflicts(ctx)

	// Should return nil (no conflicts resolved)
	if conflicts != nil {
		t.Errorf("resolveConflicts with cancelled context = %v, want nil", conflicts)
	}
}
