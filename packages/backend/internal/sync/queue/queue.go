// Package queue provides sync queue management for offline operations.
// T157: Sync queue manager with exponential backoff and retry logic.
package queue

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/kimhsiao/memonexus/backend/internal/models"
)

// Operation represents a sync operation type.
type Operation string

const (
	OperationUpload   Operation = "upload"
	OperationDownload Operation = "download"
	OperationDelete   Operation = "delete"
)

// QueueStatus represents the status of a queued operation.
type QueueStatus string

const (
	QueueStatusPending    QueueStatus = "pending"
	QueueStatusInProgress QueueStatus = "in_progress"
	QueueStatusFailed     QueueStatus = "failed"
	QueueStatusCompleted  QueueStatus = "completed"
)

// QueueItem represents a sync operation in the queue.
type QueueItem struct {
	ID           string
	Operation    Operation
	Payload      map[string]interface{}
	RetryCount   int
	MaxRetries   int
	NextRetryAt  int64
	Status       QueueStatus
	CreatedAt    int64
	UpdatedAt    int64
	LastError    string
}

// SyncQueue manages pending sync operations with retry logic.
type SyncQueue struct {
	items    map[string]*QueueItem
	mu       sync.RWMutex
	maxSize  int
	notEmpty  *sync.Cond
}

// NewSyncQueue creates a new SyncQueue.
func NewSyncQueue(maxSize int) *SyncQueue {
	q := &SyncQueue{
		items:   make(map[string]*QueueItem),
		maxSize: maxSize,
	}
	q.notEmpty = sync.NewCond(&q.mu)
	return q
}

// Enqueue adds an operation to the queue.
func (q *SyncQueue) Enqueue(operation Operation, payload map[string]interface{}) (*QueueItem, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	// Check queue capacity
	if len(q.items) >= q.maxSize {
		return nil, fmt.Errorf("queue is full (max size: %d)", q.maxSize)
	}

	now := time.Now().Unix()

	item := &QueueItem{
		ID:          uuid.New().String(),
		Operation:   operation,
		Payload:     payload,
		RetryCount:  0,
		MaxRetries:  3,
		NextRetryAt: now,
		Status:      QueueStatusPending,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	q.items[item.ID] = item

	// Signal that queue is not empty
	q.notEmpty.Signal()

	log.Printf("[SyncQueue] Enqueued %s operation %s", item.Operation, item.ID)

	return item, nil
}

// Dequeue retrieves and removes the next pending operation.
// Returns nil if no operations are ready.
func (q *SyncQueue) Dequeue() *QueueItem {
	q.mu.Lock()
	defer q.mu.Unlock()

	now := time.Now().Unix()

	// Find the next ready item
	var readyItem *QueueItem

	for _, item := range q.items {
		if item.Status == QueueStatusPending && item.NextRetryAt <= now {
			readyItem = item
			break
		}
	}

	if readyItem == nil {
		return nil
	}

	// Remove from queue (or update status if keeping for history)
	readyItem.Status = QueueStatusInProgress
	readyItem.UpdatedAt = now

	log.Printf("[SyncQueue] Dequeued %s operation %s", readyItem.Operation, readyItem.ID)

	return readyItem
}

// DequeueBlocking waits for an item to become available.
// Returns nil if timeout expires before an item is ready.
func (q *SyncQueue) DequeueBlocking(timeout time.Duration) *QueueItem {
	q.mu.Lock()
	defer q.mu.Unlock()

	now := time.Now().Unix()
	deadline := now + int64(timeout.Seconds())

	// Check for ready item immediately
	for _, item := range q.items {
		if item.Status == QueueStatusPending && item.NextRetryAt <= now {
			item.Status = QueueStatusInProgress
			item.UpdatedAt = now
			return item
		}
	}

	// Wait with timeout using a timer goroutine
	if timeout > 0 {
		stopCh := make(chan struct{})
		defer close(stopCh)

		go func() {
			select {
			case <-time.After(timeout):
				q.mu.Lock()
				q.notEmpty.Signal()
				q.mu.Unlock()
			case <-stopCh:
				return
			}
		}()

		for {
			// Check again before waiting
			now = time.Now().Unix()
			for _, item := range q.items {
				if item.Status == QueueStatusPending && item.NextRetryAt <= now {
					item.Status = QueueStatusInProgress
					item.UpdatedAt = now
					return item
				}
			}

			// Check if we've exceeded deadline
			if now >= deadline {
				return nil
			}

			// Wait for signal
			q.notEmpty.Wait()
		}
	}

	return nil
}

// Complete marks an operation as completed and removes it from the queue.
func (q *SyncQueue) Complete(id string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	item, ok := q.items[id]
	if !ok {
		return fmt.Errorf("item %s not found", id)
	}

	item.Status = QueueStatusCompleted
	item.UpdatedAt = time.Now().Unix()

	// Remove from queue
	delete(q.items, id)

	log.Printf("[SyncQueue] Completed %s operation %s", item.Operation, id)

	return nil
}

// Failed marks an operation as failed and schedules a retry if possible.
func (q *SyncQueue) Failed(id string, err error) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	item, ok := q.items[id]
	if !ok {
		return fmt.Errorf("item %s not found", id)
	}

	item.RetryCount++
	item.LastError = err.Error()
	item.UpdatedAt = time.Now().Unix()

	if item.RetryCount >= item.MaxRetries {
		// Max retries reached, mark as failed
		item.Status = QueueStatusFailed
		log.Printf("[SyncQueue] %s operation %s failed permanently: %v", item.Operation, id, err)
		return fmt.Errorf("max retries (%d) reached: %w", item.MaxRetries, err)
	}

	// Calculate next retry time with exponential backoff
	backoffSeconds := calculateBackoff(item.RetryCount)
	item.NextRetryAt = time.Now().Unix() + int64(backoffSeconds)
	item.Status = QueueStatusPending

	log.Printf("[SyncQueue] %s operation %s failed, retry %d/%d in %d seconds: %v",
		item.Operation, id, item.RetryCount, item.MaxRetries, backoffSeconds, err)

	// Signal that queue has pending items
	q.notEmpty.Signal()

	return nil
}

// calculateBackoff calculates exponential backoff delay in seconds.
// Formula: 2^retry_count * 60, capped at 3600 seconds (1 hour).
func calculateBackoff(retryCount int) int64 {
	backoff := int64(1) << uint(retryCount) // 2^retry_count
	backoff = backoff * 60                   // Convert to seconds

	// Cap at 1 hour
	maxBackoff := int64(3600)
	if backoff > maxBackoff {
		backoff = maxBackoff
	}

	return backoff
}

// GetPending returns all pending operations.
func (q *SyncQueue) GetPending() []*QueueItem {
	q.mu.RLock()
	defer q.mu.RUnlock()

	var pending []*QueueItem
	now := time.Now().Unix()

	for _, item := range q.items {
		if item.Status == QueueStatusPending && item.NextRetryAt <= now {
			pending = append(pending, item)
		}
	}

	return pending
}

// GetStatus returns the status of a specific item.
func (q *SyncQueue) GetStatus(id string) (*QueueItem, error) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	item, ok := q.items[id]
	if !ok {
		return nil, fmt.Errorf("item %s not found", id)
	}

	// Return a copy to avoid external modification
	copy := *item
	return &copy, nil
}

// List returns all items in the queue.
func (q *SyncQueue) List() []*QueueItem {
	q.mu.RLock()
	defer q.mu.RUnlock()

	items := make([]*QueueItem, 0, len(q.items))

	for _, item := range q.items {
		copy := *item
		items = append(items, &copy)
	}

	return items
}

// Size returns the number of items in the queue.
func (q *SyncQueue) Size() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return len(q.items)
}

// Clear removes all items from the queue.
func (q *SyncQueue) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.items = make(map[string]*QueueItem)

	log.Printf("[SyncQueue] Queue cleared")
}

// Remove removes a specific item from the queue.
func (q *SyncQueue) Remove(id string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if _, ok := q.items[id]; !ok {
		return fmt.Errorf("item %s not found", id)
	}

	delete(q.items, id)
	return nil
}

// RetryAll resets all failed items to pending for retry.
func (q *SyncQueue) RetryAll() int {
	q.mu.Lock()
	defer q.mu.Unlock()

	now := time.Now().Unix()
	count := 0

	for _, item := range q.items {
		if item.Status == QueueStatusFailed {
			item.Status = QueueStatusPending
			item.RetryCount = 0
			item.NextRetryAt = now
			item.LastError = ""
			item.UpdatedAt = now
			count++
		}
	}

	if count > 0 {
		q.notEmpty.Signal()
		log.Printf("[SyncQueue] Reset %d failed items for retry", count)
	}

	return count
}

// GetStats returns queue statistics.
func (q *SyncQueue) GetStats() map[string]int {
	q.mu.RLock()
	defer q.mu.RUnlock()

	stats := map[string]int{
		"total":     0,
		"pending":   0,
		"in_progress": 0,
		"failed":    0,
		"completed": 0,
	}

	for _, item := range q.items {
		stats["total"]++
		switch item.Status {
		case QueueStatusPending:
			stats["pending"]++
		case QueueStatusInProgress:
			stats["in_progress"]++
		case QueueStatusFailed:
			stats["failed"]++
		case QueueStatusCompleted:
			stats["completed"]++
		}
	}

	return stats
}

// ToModel converts a QueueItem to a SyncQueue model for database storage.
func (item *QueueItem) ToModel() *models.SyncQueue {
	payloadJSON, _ := json.Marshal(item.Payload)

	return &models.SyncQueue{
		ID:          models.UUID(item.ID),
		Operation:   string(item.Operation),
		Payload:     json.RawMessage(payloadJSON),
		RetryCount:  item.RetryCount,
		MaxRetries:  item.MaxRetries,
		NextRetryAt: item.NextRetryAt,
		Status:      string(item.Status),
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}
}

// FromModel creates a QueueItem from a SyncQueue model.
func FromModel(model *models.SyncQueue) (*QueueItem, error) {
	var payload map[string]interface{}
	if err := json.Unmarshal(model.Payload, &payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	return &QueueItem{
		ID:          string(model.ID),
		Operation:   Operation(model.Operation),
		Payload:     payload,
		RetryCount:  model.RetryCount,
		MaxRetries:  model.MaxRetries,
		NextRetryAt: model.NextRetryAt,
		Status:      QueueStatus(model.Status),
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}, nil
}
