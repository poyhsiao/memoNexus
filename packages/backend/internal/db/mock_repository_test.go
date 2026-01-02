// Package db provides mock repository implementations for testing.
package db

import (
	"database/sql"
	"sync"

	"github.com/kimhsiao/memonexus/backend/internal/models"
)

// MockContentItemRepository is a mock implementation of ContentItemRepository.
type MockContentItemRepository struct {
	mu             sync.RWMutex
	items          map[string]*models.ContentItem
	createError    error
	getError       error
	listError      error
	updateError    error
	deleteError    error
	createCalled   bool
	getCalled      bool
	listCalled     bool
	updateCalled   bool
	deleteCalled   bool
	createIDArg    *models.ContentItem
	getIDArg       string
	listLimitArg   int
	listOffsetArg  int
	listFilterArg  string
	updateItemArg  *models.ContentItem
	deleteIDArg    string
}

// NewMockContentItemRepository creates a new mock repository.
func NewMockContentItemRepository() *MockContentItemRepository {
	return &MockContentItemRepository{
		items: make(map[string]*models.ContentItem),
	}
}

// CreateContentItem creates a new content item in the mock.
func (m *MockContentItemRepository) CreateContentItem(item *models.ContentItem) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.createCalled = true
	m.createIDArg = item

	if m.createError != nil {
		return m.createError
	}

	m.items[string(item.ID)] = item
	return nil
}

// GetContentItem retrieves a content item by ID from the mock.
func (m *MockContentItemRepository) GetContentItem(id string) (*models.ContentItem, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	m.getCalled = true
	m.getIDArg = id

	if m.getError != nil {
		return nil, m.getError
	}

	item, ok := m.items[id]
	if !ok {
		return nil, sql.ErrNoRows
	}
	return item, nil
}

// ListContentItems returns content items from the mock.
func (m *MockContentItemRepository) ListContentItems(limit, offset int, mediaType string) ([]*models.ContentItem, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	m.listCalled = true
	m.listLimitArg = limit
	m.listOffsetArg = offset
	m.listFilterArg = mediaType

	if m.listError != nil {
		return nil, m.listError
	}

	// Filter by media type if specified
	var result []*models.ContentItem
	for _, item := range m.items {
		if mediaType == "" || item.MediaType == mediaType {
			result = append(result, item)
		}
	}

	// Apply pagination
	start := offset
	if start > len(result) {
		start = len(result)
	}
	end := start + limit
	if end > len(result) {
		end = len(result)
	}

	if start >= end {
		return []*models.ContentItem{}, nil
	}

	return result[start:end], nil
}

// UpdateContentItem updates a content item in the mock.
func (m *MockContentItemRepository) UpdateContentItem(item *models.ContentItem) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.updateCalled = true
	m.updateItemArg = item

	if m.updateError != nil {
		return m.updateError
	}

	if _, ok := m.items[string(item.ID)]; !ok {
		return sql.ErrNoRows
	}

	m.items[string(item.ID)] = item
	return nil
}

// DeleteContentItem soft deletes a content item from the mock.
func (m *MockContentItemRepository) DeleteContentItem(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.deleteCalled = true
	m.deleteIDArg = id

	if m.deleteError != nil {
		return m.deleteError
	}

	if _, ok := m.items[id]; !ok {
		return sql.ErrNoRows
	}
	delete(m.items, id)
	return nil
}

// SetCreateError sets the error to return from CreateContentItem.
func (m *MockContentItemRepository) SetCreateError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.createError = err
}

// SetGetError sets the error to return from GetContentItem.
func (m *MockContentItemRepository) SetGetError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.getError = err
}

// SetListError sets the error to return from ListContentItems.
func (m *MockContentItemRepository) SetListError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.listError = err
}

// SetUpdateError sets the error to return from UpdateContentItem.
func (m *MockContentItemRepository) SetUpdateError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.updateError = err
}

// SetDeleteError sets the error to return from DeleteContentItem.
func (m *MockContentItemRepository) SetDeleteError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.deleteError = err
}

// WasCreateCalled returns true if CreateContentItem was called.
func (m *MockContentItemRepository) WasCreateCalled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.createCalled
}

// WasGetCalled returns true if GetContentItem was called.
func (m *MockContentItemRepository) WasGetCalled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.getCalled
}

// WasListCalled returns true if ListContentItems was called.
func (m *MockContentItemRepository) WasListCalled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.listCalled
}

// WasUpdateCalled returns true if UpdateContentItem was called.
func (m *MockContentItemRepository) WasUpdateCalled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.updateCalled
}

// WasDeleteCalled returns true if DeleteContentItem was called.
func (m *MockContentItemRepository) WasDeleteCalled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.deleteCalled
}

// GetCreateIDArg returns the item passed to the last CreateContentItem call.
func (m *MockContentItemRepository) GetCreateIDArg() *models.ContentItem {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.createIDArg
}

// GetGetIDArg returns the ID passed to the last GetContentItem call.
func (m *MockContentItemRepository) GetGetIDArg() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.getIDArg
}

// GetListArgs returns the arguments passed to the last ListContentItems call.
func (m *MockContentItemRepository) GetListArgs() (limit, offset int, mediaType string) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.listLimitArg, m.listOffsetArg, m.listFilterArg
}

// GetUpdateItemArg returns the item passed to the last UpdateContentItem call.
func (m *MockContentItemRepository) GetUpdateItemArg() *models.ContentItem {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.updateItemArg
}

// GetDeleteIDArg returns the ID passed to the last DeleteContentItem call.
func (m *MockContentItemRepository) GetDeleteIDArg() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.deleteIDArg
}

// Count returns the number of items in the mock repository.
func (m *MockContentItemRepository) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.items)
}

// =====================================================
// Mock ChangeLog Repository
// =====================================================

// MockChangeLogRepository is a mock implementation of ChangeLogRepository.
type MockChangeLogRepository struct {
	mu          sync.Mutex
	createError error
	createCalled bool
	createLogArg *models.ChangeLog
}

// NewMockChangeLogRepository creates a new mock change log repository.
func NewMockChangeLogRepository() *MockChangeLogRepository {
	return &MockChangeLogRepository{}
}

// CreateChangeLog creates a new change log entry in the mock.
func (m *MockChangeLogRepository) CreateChangeLog(log *models.ChangeLog) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.createCalled = true
	m.createLogArg = log

	if m.createError != nil {
		return m.createError
	}
	return nil
}

// SetCreateError sets the error to return from CreateChangeLog.
func (m *MockChangeLogRepository) SetCreateError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.createError = err
}

// WasCreateCalled returns true if CreateChangeLog was called.
func (m *MockChangeLogRepository) WasCreateCalled() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.createCalled
}

// GetCreateLogArg returns the log passed to the last CreateChangeLog call.
func (m *MockChangeLogRepository) GetCreateLogArg() *models.ChangeLog {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.createLogArg
}

// =====================================================
// Mock ConflictLog Repository
// =====================================================

// MockConflictLogRepository is a mock implementation of ConflictLogRepository.
type MockConflictLogRepository struct {
	mu          sync.Mutex
	createError error
	createCalled bool
	createLogArg *models.ConflictLog
}

// NewMockConflictLogRepository creates a new mock conflict log repository.
func NewMockConflictLogRepository() *MockConflictLogRepository {
	return &MockConflictLogRepository{}
}

// CreateConflictLog creates a new conflict log entry in the mock.
func (m *MockConflictLogRepository) CreateConflictLog(log *models.ConflictLog) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.createCalled = true
	m.createLogArg = log

	if m.createError != nil {
		return m.createError
	}
	return nil
}

// SetCreateError sets the error to return from CreateConflictLog.
func (m *MockConflictLogRepository) SetCreateError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.createError = err
}

// WasCreateCalled returns true if CreateConflictLog was called.
func (m *MockConflictLogRepository) WasCreateCalled() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.createCalled
}

// GetCreateLogArg returns the log passed to the last CreateConflictLog call.
func (m *MockConflictLogRepository) GetCreateLogArg() *models.ConflictLog {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.createLogArg
}

// =====================================================
// Mock Sync Repository
// =====================================================

// MockSyncRepository combines all mock repositories for sync operations.
type MockSyncRepository struct {
	*MockContentItemRepository
	*MockChangeLogRepository
	*MockConflictLogRepository
}

// NewMockSyncRepository creates a new mock sync repository.
func NewMockSyncRepository() *MockSyncRepository {
	return &MockSyncRepository{
		MockContentItemRepository:  NewMockContentItemRepository(),
		MockChangeLogRepository:    NewMockChangeLogRepository(),
		MockConflictLogRepository:  NewMockConflictLogRepository(),
	}
}
