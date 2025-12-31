// Package queue provides unit tests for sync queue.
// T150: Unit test for sync queue with exponential backoff.
package queue

import (
	"testing"
	"time"

	"github.com/kimhsiao/memonexus/backend/internal/models"
)

// TestSyncQueueEnqueue tests enqueuing operations.
func TestSyncQueueEnqueue(t *testing.T) {
	q := NewSyncQueue(100)

	payload := map[string]interface{}{
		"item_id": "test-item-1",
		"data":    "test data",
	}

	item, err := q.Enqueue(OperationUpload, payload)

	if err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}

	if item == nil {
		t.Fatal("Expected non-nil item")
	}

	if item.ID == "" {
		t.Error("Expected item ID to be set")
	}

	if item.Operation != OperationUpload {
		t.Errorf("Expected Upload operation, got %s", item.Operation)
	}

	if item.Status != QueueStatusPending {
		t.Errorf("Expected Pending status, got %s", item.Status)
	}

	if item.RetryCount != 0 {
		t.Errorf("Expected RetryCount 0, got %d", item.RetryCount)
	}

	if item.MaxRetries != 3 {
		t.Errorf("Expected MaxRetries 3, got %d", item.MaxRetries)
	}
}

// TestSyncQueueFull tests queue capacity limit.
func TestSyncQueueFull(t *testing.T) {
	q := NewSyncQueue(2)

	// Fill queue to capacity
	payload := map[string]interface{}{"data": "test"}
	q.Enqueue(OperationUpload, payload)
	q.Enqueue(OperationDownload, payload)

	// Try to enqueue beyond capacity
	_, err := q.Enqueue(OperationDelete, payload)

	if err == nil {
		t.Error("Expected error when queue is full")
	}
}

// TestSyncQueueDequeue tests dequeuing operations.
func TestSyncQueueDequeue(t *testing.T) {
	q := NewSyncQueue(100)

	payload := map[string]interface{}{"item_id": "test-item-1"}
	q.Enqueue(OperationUpload, payload)

	// Dequeue should return the item
	item := q.Dequeue()

	if item == nil {
		t.Fatal("Expected non-nil item")
	}

	if item.Status != QueueStatusInProgress {
		t.Errorf("Expected InProgress status, got %s", item.Status)
	}

	// Dequeue again should return nil (no more items)
	item = q.Dequeue()
	if item != nil {
		t.Error("Expected nil when queue is empty")
	}
}

// TestSyncQueueDequeueBlocking tests blocking dequeue with timeout.
func TestSyncQueueDequeueBlocking(t *testing.T) {
	q := NewSyncQueue(100)

	// Test with empty queue and short timeout
	item := q.DequeueBlocking(100 * time.Millisecond)
	if item != nil {
		t.Error("Expected nil when queue is empty")
	}

	// Add item and test blocking dequeue
	payload := map[string]interface{}{"item_id": "test-item-1"}
	q.Enqueue(OperationUpload, payload)

	// Dequeue should return immediately
	item = q.DequeueBlocking(1 * time.Second)
	if item == nil {
		t.Fatal("Expected non-nil item")
	}

	if item.Operation != OperationUpload {
		t.Errorf("Expected Upload operation, got %s", item.Operation)
	}
}

// TestSyncQueueComplete tests marking operations as completed.
func TestSyncQueueComplete(t *testing.T) {
	q := NewSyncQueue(100)

	payload := map[string]interface{}{"item_id": "test-item-1"}
	item, _ := q.Enqueue(OperationUpload, payload)

	// Mark as completed
	err := q.Complete(item.ID)

	if err != nil {
		t.Fatalf("Complete failed: %v", err)
	}

	// Item should be removed from queue
	status, err := q.GetStatus(item.ID)
	if err == nil {
		t.Error("Expected error when getting completed item")
	}

	if status != nil {
		t.Error("Expected nil status for completed item")
	}
}

// TestSyncQueueFailed tests handling failed operations.
func TestSyncQueueFailed(t *testing.T) {
	q := NewSyncQueue(100)

	payload := map[string]interface{}{"item_id": "test-item-1"}
	item, _ := q.Enqueue(OperationUpload, payload)

	testErr := &TestError{Message: "upload failed"}

	// Mark as failed
	err := q.Failed(item.ID, testErr)

	if err != nil {
		t.Fatalf("Failed failed: %v", err)
	}

	// Check item status
	status, err := q.GetStatus(item.ID)
	if err != nil {
		t.Fatalf("GetStatus failed: %v", err)
	}

	if status.RetryCount != 1 {
		t.Errorf("Expected RetryCount 1, got %d", status.RetryCount)
	}

	if status.Status != QueueStatusPending {
		t.Errorf("Expected Pending status (for retry), got %s", status.Status)
	}

	if status.LastError != testErr.Error() {
		t.Errorf("Expected error message, got %s", status.LastError)
	}
}

// TestSyncQueueMaxRetries tests max retries limit.
func TestSyncQueueMaxRetries(t *testing.T) {
	q := NewSyncQueue(100)

	payload := map[string]interface{}{"item_id": "test-item-1"}
	item, _ := q.Enqueue(OperationUpload, payload)

	testErr := &TestError{Message: "upload failed"}

	// Fail 3 times (max retries)
	for i := 0; i < 3; i++ {
		err := q.Failed(item.ID, testErr)
		if i < 2 && err != nil {
			t.Fatalf("Failed failed on attempt %d: %v", i+1, err)
		}
	}

	// Third failure should mark as permanently failed
	status, err := q.GetStatus(item.ID)
	if err != nil {
		t.Fatalf("GetStatus failed: %v", err)
	}

	if status.Status != QueueStatusFailed {
		t.Errorf("Expected Failed status, got %s", status.Status)
	}

	if status.RetryCount != 3 {
		t.Errorf("Expected RetryCount 3, got %d", status.RetryCount)
	}
}

// TestCalculateBackoff tests exponential backoff calculation.
func TestCalculateBackoff(t *testing.T) {
	tests := []struct {
		name        string
		retryCount  int
		wantSeconds int64
	}{
		{"retry 1", 1, 120},      // 2^1 * 60 = 120
		{"retry 2", 2, 240},      // 2^2 * 60 = 240
		{"retry 3", 3, 480},      // 2^3 * 60 = 480
		{"retry 4", 4, 960},      // 2^4 * 60 = 960
		{"retry 5", 5, 1920},     // 2^5 * 60 = 1920
		{"retry 6", 6, 3600},     // 2^6 * 60 = 3840 -> capped to 3600
		{"retry 10", 10, 3600},   // Capped at 1 hour
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateBackoff(tt.retryCount)
			if got != tt.wantSeconds {
				t.Errorf("calculateBackoff(%d) = %d, want %d", tt.retryCount, got, tt.wantSeconds)
			}
		})
	}
}

// TestGetPending tests retrieving pending operations.
func TestGetPending(t *testing.T) {
	q := NewSyncQueue(100)

	// Enqueue multiple items
	payload := map[string]interface{}{"data": "test"}
	q.Enqueue(OperationUpload, payload)
	q.Enqueue(OperationDownload, payload)

	// Mark one as in progress
	item := q.Dequeue()

	pending := q.GetPending()

	// Should only have 1 pending item (the other is in progress)
	if len(pending) != 1 {
		t.Errorf("Expected 1 pending item, got %d", len(pending))
	}

	// The pending item should not be the in-progress item
	if len(pending) > 0 && pending[0].ID == item.ID {
		t.Error("Pending item should not be the in-progress item")
	}
}

// TestGetStatus tests retrieving item status.
func TestGetStatus(t *testing.T) {
	q := NewSyncQueue(100)

	payload := map[string]interface{}{"item_id": "test-item-1"}
	enqueuedItem, _ := q.Enqueue(OperationUpload, payload)

	// Get status
	status, err := q.GetStatus(enqueuedItem.ID)

	if err != nil {
		t.Fatalf("GetStatus failed: %v", err)
	}

	if status.ID != enqueuedItem.ID {
		t.Errorf("Expected ID %s, got %s", enqueuedItem.ID, status.ID)
	}

	if status.Operation != OperationUpload {
		t.Errorf("Expected Upload operation, got %s", status.Operation)
	}
}

// TestGetStatusNotFound tests getting status for non-existent item.
func TestGetStatusNotFound(t *testing.T) {
	q := NewSyncQueue(100)

	_, err := q.GetStatus("non-existent-id")

	if err == nil {
		t.Error("Expected error for non-existent item")
	}
}

// TestList tests listing all items.
func TestList(t *testing.T) {
	q := NewSyncQueue(100)

	// Enqueue multiple items
	payload := map[string]interface{}{"data": "test"}
	q.Enqueue(OperationUpload, payload)
	q.Enqueue(OperationDownload, payload)
	q.Enqueue(OperationDelete, payload)

	// List all items
	items := q.List()

	if len(items) != 3 {
		t.Errorf("Expected 3 items, got %d", len(items))
	}
}

// TestSize tests getting queue size.
func TestSize(t *testing.T) {
	q := NewSyncQueue(100)

	if q.Size() != 0 {
		t.Errorf("Expected size 0, got %d", q.Size())
	}

	payload := map[string]interface{}{"data": "test"}
	q.Enqueue(OperationUpload, payload)

	if q.Size() != 1 {
		t.Errorf("Expected size 1, got %d", q.Size())
	}
}

// TestClear tests clearing the queue.
func TestClear(t *testing.T) {
	q := NewSyncQueue(100)

	// Enqueue items
	payload := map[string]interface{}{"data": "test"}
	q.Enqueue(OperationUpload, payload)
	q.Enqueue(OperationDownload, payload)

	// Clear queue
	q.Clear()

	if q.Size() != 0 {
		t.Errorf("Expected size 0 after clear, got %d", q.Size())
	}
}

// TestRemove tests removing a specific item.
func TestRemove(t *testing.T) {
	q := NewSyncQueue(100)

	payload := map[string]interface{}{"data": "test"}
	item, _ := q.Enqueue(OperationUpload, payload)

	// Remove item
	err := q.Remove(item.ID)

	if err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	if q.Size() != 0 {
		t.Errorf("Expected size 0 after remove, got %d", q.Size())
	}

	// Remove non-existent item should error
	err = q.Remove("non-existent-id")
	if err == nil {
		t.Error("Expected error when removing non-existent item")
	}
}

// TestRetryAll tests resetting failed items.
func TestRetryAll(t *testing.T) {
	q := NewSyncQueue(100)

	payload := map[string]interface{}{"data": "test"}

	// Enqueue items
	item1, _ := q.Enqueue(OperationUpload, payload)
	item2, _ := q.Enqueue(OperationDownload, payload)

	// Mark both as failed (max retries)
	testErr := &TestError{Message: "failed"}
	for i := 0; i < 3; i++ {
		q.Failed(item1.ID, testErr)
		q.Failed(item2.ID, testErr)
	}

	// Verify both are failed
	stats := q.GetStats()
	if stats["failed"] != 2 {
		t.Errorf("Expected 2 failed items, got %d", stats["failed"])
	}

	// Retry all
	count := q.RetryAll()

	if count != 2 {
		t.Errorf("Expected 2 items retried, got %d", count)
	}

	// Verify both are now pending
	stats = q.GetStats()
	if stats["pending"] != 2 {
		t.Errorf("Expected 2 pending items after retry, got %d", stats["pending"])
	}
}

// TestGetStats tests queue statistics.
func TestGetStats(t *testing.T) {
	q := NewSyncQueue(100)

	payload := map[string]interface{}{"data": "test"}

	// Add items in different states
	_, _ = q.Enqueue(OperationUpload, payload)
	item2, _ := q.Enqueue(OperationDownload, payload)
	_, _ = q.Enqueue(OperationDelete, payload)

	// Mark first item as in progress
	q.Dequeue()

	// Mark item2 as failed
	testErr := &TestError{Message: "failed"}
	q.Failed(item2.ID, testErr)
	q.Failed(item2.ID, testErr)
	q.Failed(item2.ID, testErr) // Max retries

	// Get stats
	stats := q.GetStats()

	if stats["total"] != 3 {
		t.Errorf("Expected total 3, got %d", stats["total"])
	}

	if stats["in_progress"] != 1 {
		t.Errorf("Expected 1 in_progress, got %d", stats["in_progress"])
	}

	if stats["failed"] != 1 {
		t.Errorf("Expected 1 failed, got %d", stats["failed"])
	}

	if stats["pending"] != 1 {
		t.Errorf("Expected 1 pending, got %d", stats["pending"])
	}
}

// TestToModel tests converting QueueItem to database model.
func TestToModel(t *testing.T) {
	item := &QueueItem{
		ID:        "test-id",
		Operation: OperationUpload,
		Payload:   map[string]interface{}{"item_id": "item-1"},
		Status:    QueueStatusPending,
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}

	model := item.ToModel()

	if string(model.ID) != item.ID {
		t.Errorf("Expected ID %s, got %s", item.ID, model.ID)
	}

	if model.Operation != string(item.Operation) {
		t.Errorf("Expected operation %s, got %s", item.Operation, model.Operation)
	}
}

// TestFromModel tests creating QueueItem from database model.
func TestFromModel(t *testing.T) {
	model := &models.SyncQueue{
		ID:        "test-id",
		Operation: "upload",
		Payload:   []byte(`{"item_id": "item-1"}`),
		Status:    "pending",
	}

	item, err := FromModel(model)

	if err != nil {
		t.Fatalf("FromModel failed: %v", err)
	}

	if item.ID != string(model.ID) {
		t.Errorf("Expected ID %s, got %s", model.ID, item.ID)
	}

	if item.Operation != OperationUpload {
		t.Errorf("Expected Upload operation, got %s", item.Operation)
	}

	if item.Payload["item_id"] != "item-1" {
		t.Errorf("Expected item_id item-1, got %v", item.Payload["item_id"])
	}
}

// TestError types
type TestError struct {
	Message string
}

func (e *TestError) Error() string {
	return e.Message
}
