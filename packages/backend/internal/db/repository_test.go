// Package db provides unit tests for CRUD repository operations.
package db

import (
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/kimhsiao/memonexus/backend/internal/models"
)

// setupTestDB creates an in-memory SQLite database for testing.
func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Create test schema
	_, err = db.Exec(`
		CREATE TABLE content_items (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			content_text TEXT NOT NULL DEFAULT '',
			source_url TEXT,
			media_type TEXT NOT NULL,
			tags TEXT DEFAULT '',
			summary TEXT,
			is_deleted INTEGER NOT NULL DEFAULT 0,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL,
			version INTEGER NOT NULL DEFAULT 1,
			content_hash TEXT
		);

		CREATE TABLE tags (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			color TEXT DEFAULT '#3B82F6',
			is_deleted INTEGER NOT NULL DEFAULT 0,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		);

		CREATE TABLE change_log (
			id TEXT PRIMARY KEY,
			item_id TEXT NOT NULL,
			operation TEXT NOT NULL,
			version INTEGER NOT NULL,
			timestamp INTEGER NOT NULL
		);

		CREATE TABLE conflict_log (
			id TEXT PRIMARY KEY,
			item_id TEXT NOT NULL,
			local_timestamp INTEGER NOT NULL,
			remote_timestamp INTEGER NOT NULL,
			resolution TEXT NOT NULL DEFAULT 'last_write_wins',
			detected_at INTEGER NOT NULL
		);

		CREATE TABLE sync_queue (
			id TEXT PRIMARY KEY,
			operation TEXT NOT NULL,
			payload TEXT NOT NULL,
			retry_count INTEGER NOT NULL DEFAULT 0,
			max_retries INTEGER NOT NULL DEFAULT 3,
			next_retry_at INTEGER NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending',
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		);

		CREATE TABLE ai_config (
			id TEXT PRIMARY KEY,
			provider TEXT NOT NULL,
			api_endpoint TEXT,
			api_key_encrypted TEXT,
			model_name TEXT,
			max_tokens INTEGER,
			is_enabled INTEGER NOT NULL DEFAULT 1,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		);

		CREATE TABLE sync_credentials (
			id TEXT PRIMARY KEY,
			endpoint TEXT NOT NULL,
			bucket_name TEXT NOT NULL,
			region TEXT,
			access_key_encrypted TEXT,
			secret_key_encrypted TEXT,
			is_enabled INTEGER NOT NULL DEFAULT 1,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		);
	`)
	if err != nil {
		db.Close()
		t.Fatalf("Failed to create test schema: %v", err)
	}

	return db
}

// =====================================================
// ContentItem Repository Tests (T083)
// =====================================================

func TestCreateContentItem(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewRepository(db)

	item := &models.ContentItem{
		Title:       "Test Article",
		ContentText: "This is test content",
		MediaType:   "web",
		Tags:        "test,article",
	}

	err := repo.CreateContentItem(item)
	if err != nil {
		t.Fatalf("CreateContentItem failed: %v", err)
	}

	if item.ID == "" {
		t.Error("Expected ID to be generated")
	}
	if item.CreatedAt == 0 {
		t.Error("Expected CreatedAt to be set")
	}
	if item.Version != 1 {
		t.Errorf("Expected Version to be 1, got %d", item.Version)
	}
}

func TestGetContentItem(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewRepository(db)

	// Create a test item
	created := &models.ContentItem{
		Title:       "Test Article",
		ContentText: "This is test content",
		MediaType:   "web",
	}
	err := repo.CreateContentItem(created)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Retrieve the item
	retrieved, err := repo.GetContentItem(string(created.ID))
	if err != nil {
		t.Fatalf("GetContentItem failed: %v", err)
	}

	if retrieved.Title != created.Title {
		t.Errorf("Expected title %s, got %s", created.Title, retrieved.Title)
	}
	if retrieved.ContentText != created.ContentText {
		t.Errorf("Expected content %s, got %s", created.ContentText, retrieved.ContentText)
	}
}

func TestListContentItems(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewRepository(db)

	// Create test items
	for i := 0; i < 5; i++ {
		item := &models.ContentItem{
			Title:       "Article",
			ContentText: "Content",
			MediaType:   "web",
		}
		err := repo.CreateContentItem(item)
		if err != nil {
			t.Fatalf("Setup failed: %v", err)
		}
	}

	// List items
	items, err := repo.ListContentItems(10, 0, "")
	if err != nil {
		t.Fatalf("ListContentItems failed: %v", err)
	}

	if len(items) != 5 {
		t.Errorf("Expected 5 items, got %d", len(items))
	}
}

func TestUpdateContentItem(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewRepository(db)

	// Create a test item
	created := &models.ContentItem{
		Title:       "Original Title",
		ContentText: "Original content",
		MediaType:   "web",
	}
	err := repo.CreateContentItem(created)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Update the item
	created.Title = "Updated Title"
	created.ContentText = "Updated content"
	err = repo.UpdateContentItem(created)
	if err != nil {
		t.Fatalf("UpdateContentItem failed: %v", err)
	}

	// Verify update
	retrieved, err := repo.GetContentItem(string(created.ID))
	if err != nil {
		t.Fatalf("GetContentItem failed: %v", err)
	}

	if retrieved.Title != "Updated Title" {
		t.Errorf("Expected title 'Updated Title', got %s", retrieved.Title)
	}
	if retrieved.Version != 2 {
		t.Errorf("Expected Version to be 2, got %d", retrieved.Version)
	}
}

func TestDeleteContentItem(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewRepository(db)

	// Create a test item
	created := &models.ContentItem{
		Title:       "To Delete",
		ContentText: "Content",
		MediaType:   "web",
	}
	err := repo.CreateContentItem(created)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Delete the item
	err = repo.DeleteContentItem(string(created.ID))
	if err != nil {
		t.Fatalf("DeleteContentItem failed: %v", err)
	}

	// Verify soft delete - item should not be found
	_, err = repo.GetContentItem(string(created.ID))
	if err == nil {
		t.Error("Expected error when retrieving deleted item")
	}
}

func TestGetContentItemNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewRepository(db)

	_, err := repo.GetContentItem("non-existent-id")
	if err == nil {
		t.Error("Expected error for non-existent item")
	}
}

func TestListContentItemsWithMediaTypeFilter(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewRepository(db)

	// Create items with different media types
	webItem := &models.ContentItem{Title: "Web", ContentText: "Content", MediaType: "web"}
	imageItem := &models.ContentItem{Title: "Image", ContentText: "Content", MediaType: "image"}

	repo.CreateContentItem(webItem)
	repo.CreateContentItem(imageItem)

	// Filter by media type
	items, err := repo.ListContentItems(10, 0, "web")
	if err != nil {
		t.Fatalf("ListContentItems failed: %v", err)
	}

	if len(items) != 1 {
		t.Errorf("Expected 1 item with media_type 'web', got %d", len(items))
	}
	if items[0].MediaType != "web" {
		t.Errorf("Expected media_type 'web', got %s", items[0].MediaType)
	}
}

func TestListContentItemsWithPagination(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewRepository(db)

	// Create 25 items
	for i := 0; i < 25; i++ {
		item := &models.ContentItem{
			Title:       "Article",
			ContentText: "Content",
			MediaType:   "web",
		}
		repo.CreateContentItem(item)
	}

	// Test first page
	page1, err := repo.ListContentItems(10, 0, "")
	if err != nil {
		t.Fatalf("ListContentItems failed: %v", err)
	}
	if len(page1) != 10 {
		t.Errorf("Expected 10 items on page 1, got %d", len(page1))
	}

	// Test second page
	page2, err := repo.ListContentItems(10, 10, "")
	if err != nil {
		t.Fatalf("ListContentItems failed: %v", err)
	}
	if len(page2) != 10 {
		t.Errorf("Expected 10 items on page 2, got %d", len(page2))
	}

	// Test third page (partial)
	page3, err := repo.ListContentItems(10, 20, "")
	if err != nil {
		t.Fatalf("ListContentItems failed: %v", err)
	}
	if len(page3) != 5 {
		t.Errorf("Expected 5 items on page 3, got %d", len(page3))
	}
}

// =====================================================
// Tag Repository Tests (T084)
// =====================================================

func TestCreateTag(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewRepository(db)

	tag := &models.Tag{
		Name:  "technology",
		Color: "#3B82F6",
	}

	err := repo.CreateTag(tag)
	if err != nil {
		t.Fatalf("CreateTag failed: %v", err)
	}

	if tag.ID == "" {
		t.Error("Expected ID to be generated")
	}
	if tag.CreatedAt == 0 {
		t.Error("Expected CreatedAt to be set")
	}
}

func TestGetTag(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewRepository(db)

	// Create a test tag
	created := &models.Tag{
		Name:  "technology",
		Color: "#3B82F6",
	}
	err := repo.CreateTag(created)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Retrieve the tag
	retrieved, err := repo.GetTag(string(created.ID))
	if err != nil {
		t.Fatalf("GetTag failed: %v", err)
	}

	if retrieved.Name != created.Name {
		t.Errorf("Expected name %s, got %s", created.Name, retrieved.Name)
	}
	if retrieved.Color != created.Color {
		t.Errorf("Expected color %s, got %s", created.Color, retrieved.Color)
	}
}

func TestListTags(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewRepository(db)

	// Create test tags
	tags := []*models.Tag{
		{Name: "technology", Color: "#3B82F6"},
		{Name: "design", Color: "#EC4899"},
		{Name: "business", Color: "#10B981"},
	}
	for _, tag := range tags {
		err := repo.CreateTag(tag)
		if err != nil {
			t.Fatalf("Setup failed: %v", err)
		}
	}

	// List all tags
	retrieved, err := repo.ListTags()
	if err != nil {
		t.Fatalf("ListTags failed: %v", err)
	}

	if len(retrieved) != 3 {
		t.Errorf("Expected 3 tags, got %d", len(retrieved))
	}
}

func TestUpdateTag(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewRepository(db)

	// Create a test tag
	created := &models.Tag{
		Name:  "technology",
		Color: "#3B82F6",
	}
	err := repo.CreateTag(created)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Update the tag
	created.Name = "tech"
	created.Color = "#F59E0B"
	err = repo.UpdateTag(created)
	if err != nil {
		t.Fatalf("UpdateTag failed: %v", err)
	}

	// Verify update
	retrieved, err := repo.GetTag(string(created.ID))
	if err != nil {
		t.Fatalf("GetTag failed: %v", err)
	}

	if retrieved.Name != "tech" {
		t.Errorf("Expected name 'tech', got %s", retrieved.Name)
	}
	if retrieved.Color != "#F59E0B" {
		t.Errorf("Expected color '#F59E0B', got %s", retrieved.Color)
	}
}

func TestDeleteTag(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewRepository(db)

	// Create a test tag
	created := &models.Tag{
		Name:  "to-delete",
		Color: "#3B82F6",
	}
	err := repo.CreateTag(created)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Delete the tag
	err = repo.DeleteTag(string(created.ID))
	if err != nil {
		t.Fatalf("DeleteTag failed: %v", err)
	}

	// Verify soft delete - tag should not be in list
	tags, err := repo.ListTags()
	if err != nil {
		t.Fatalf("ListTags failed: %v", err)
	}

	for _, tag := range tags {
		if tag.ID == created.ID {
			t.Error("Expected deleted tag to not be in list")
		}
	}
}

func TestCreateTagDuplicateName(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewRepository(db)

	// Create first tag
	tag1 := &models.Tag{Name: "technology", Color: "#3B82F6"}
	err := repo.CreateTag(tag1)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Try to create duplicate
	tag2 := &models.Tag{Name: "technology", Color: "#EC4899"}
	err = repo.CreateTag(tag2)
	if err == nil {
		t.Error("Expected error for duplicate tag name")
	}
}

// =====================================================
// ChangeLog Repository Tests
// =====================================================

func TestCreateChangeLog(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewRepository(db)

	log := &models.ChangeLog{
		ItemID:    "test-item-id",
		Operation: "create",
		Version:   1,
	}

	err := repo.CreateChangeLog(log)
	if err != nil {
		t.Fatalf("CreateChangeLog failed: %v", err)
	}

	if log.ID == "" {
		t.Error("Expected ID to be generated")
	}
	if log.Timestamp == 0 {
		t.Error("Expected Timestamp to be set")
	}
}

// =====================================================
// ConflictLog Repository Tests
// =====================================================

func TestCreateConflictLog(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewRepository(db)

	log := &models.ConflictLog{
		ItemID:         "test-item-id",
		LocalTimestamp: 1234567890,
		RemoteTimestamp: 1234567900,
		Resolution:     "last_write_wins",
	}

	err := repo.CreateConflictLog(log)
	if err != nil {
		t.Fatalf("CreateConflictLog failed: %v", err)
	}

	if log.ID == "" {
		t.Error("Expected ID to be generated")
	}
	if log.DetectedAt == 0 {
		t.Error("Expected DetectedAt to be set")
	}
}

// =====================================================
// SyncQueue Repository Tests
// =====================================================

func TestCreateSyncQueue(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewRepository(db)

	entry := &models.SyncQueue{
		Operation:  "upload",
		Payload:    json.RawMessage(`{"item_id": "test-id"}`),
		MaxRetries: 3,
		Status:     "pending",
	}

	err := repo.CreateSyncQueue(entry)
	if err != nil {
		t.Fatalf("CreateSyncQueue failed: %v", err)
	}

	if entry.ID == "" {
		t.Error("Expected ID to be generated")
	}
	if entry.CreatedAt == 0 {
		t.Error("Expected CreatedAt to be set")
	}
}

// =====================================================
// AIConfig Repository Tests
// =====================================================

func TestSaveAIConfig_create(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewRepository(db)

	config := &models.AIConfig{
		Provider:       "openai",
		APIEndpoint:    "https://api.openai.com/v1",
		APIKeyEncrypted: "encrypted_key",
		ModelName:      "gpt-4",
		MaxTokens:      4096,
		IsEnabled:      true,
	}

	err := repo.SaveAIConfig(config)
	if err != nil {
		t.Fatalf("SaveAIConfig failed: %v", err)
	}

	if config.ID == "" {
		t.Error("Expected ID to be generated")
	}
	if config.CreatedAt == 0 {
		t.Error("Expected CreatedAt to be set")
	}
	if config.UpdatedAt == 0 {
		t.Error("Expected UpdatedAt to be set")
	}
}

func TestSaveAIConfig_update(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewRepository(db)

	// Create a config
	config := &models.AIConfig{
		Provider:    "openai",
		ModelName:   "gpt-4",
		MaxTokens:   4096,
		IsEnabled:   true,
	}
	err := repo.SaveAIConfig(config)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Update the config
	config.ModelName = "gpt-4-turbo"
	config.MaxTokens = 8192
	originalUpdatedAt := config.UpdatedAt

	// Wait to ensure timestamp changes (time.Now() may return same Unix timestamp)
	time.Sleep(time.Second)

	err = repo.SaveAIConfig(config)
	if err != nil {
		t.Fatalf("SaveAIConfig update failed: %v", err)
	}

	if config.UpdatedAt == originalUpdatedAt {
		t.Error("Expected UpdatedAt to change on update")
	}
}

func TestGetAIConfig(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewRepository(db)

	// Create a config
	config := &models.AIConfig{
		Provider:       "openai",
		APIEndpoint:    "https://api.openai.com/v1",
		APIKeyEncrypted: "encrypted_key",
		ModelName:      "gpt-4",
		MaxTokens:      4096,
		IsEnabled:      true,
	}
	err := repo.SaveAIConfig(config)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Retrieve the config
	retrieved, err := repo.GetAIConfig()
	if err != nil {
		t.Fatalf("GetAIConfig failed: %v", err)
	}

	if retrieved.Provider != config.Provider {
		t.Errorf("Expected provider %s, got %s", config.Provider, retrieved.Provider)
	}
	if retrieved.ModelName != config.ModelName {
		t.Errorf("Expected model %s, got %s", config.ModelName, retrieved.ModelName)
	}
	if retrieved.MaxTokens != config.MaxTokens {
		t.Errorf("Expected MaxTokens %d, got %d", config.MaxTokens, retrieved.MaxTokens)
	}
}

func TestGetAIConfig_notFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewRepository(db)

	_, err := repo.GetAIConfig()
	if err == nil {
		t.Error("Expected error when no AI config exists")
	}
}

func TestDeleteAIConfig(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewRepository(db)

	// Create a config
	config := &models.AIConfig{
		Provider:  "openai",
		ModelName: "gpt-4",
		IsEnabled: true,
	}
	err := repo.SaveAIConfig(config)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Delete the config
	err = repo.DeleteAIConfig(string(config.ID))
	if err != nil {
		t.Fatalf("DeleteAIConfig failed: %v", err)
	}

	// Verify it's deleted
	_, err = repo.GetAIConfig()
	if err == nil {
		t.Error("Expected error when retrieving deleted config")
	}
}

func TestDisableAllAIConfig(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewRepository(db)

	// Create multiple configs
	config1 := &models.AIConfig{
		Provider:  "openai",
		ModelName: "gpt-4",
		IsEnabled: true,
	}
	config2 := &models.AIConfig{
		Provider:  "claude",
		ModelName: "claude-3",
		IsEnabled: true,
	}

	repo.SaveAIConfig(config1)
	repo.SaveAIConfig(config2)

	// Disable all
	err := repo.DisableAllAIConfig()
	if err != nil {
		t.Fatalf("DisableAllAIConfig failed: %v", err)
	}

	// Verify all are disabled (GetAIConfig should return error)
	_, err = repo.GetAIConfig()
	if err == nil {
		t.Error("Expected error when all configs are disabled")
	}
}

// =====================================================
// SyncCredential Repository Tests
// =====================================================

func TestSaveSyncCredential(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewRepository(db)

	cred := &models.SyncCredential{
		Endpoint:           "https://s3.amazonaws.com",
		BucketName:         "test-bucket",
		Region:             "us-east-1",
		AccessKeyEncrypted: "encrypted_access",
		SecretKeyEncrypted: "encrypted_secret",
		IsEnabled:          true,
	}

	err := repo.SaveSyncCredential(cred)
	if err != nil {
		t.Fatalf("SaveSyncCredential failed: %v", err)
	}

	if cred.ID == "" {
		t.Error("Expected ID to be generated")
	}
	if cred.CreatedAt == 0 {
		t.Error("Expected CreatedAt to be set")
	}
	if cred.UpdatedAt == 0 {
		t.Error("Expected UpdatedAt to be set")
	}
}

func TestGetSyncCredentials(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewRepository(db)

	// Create a credential
	cred := &models.SyncCredential{
		Endpoint:           "https://s3.amazonaws.com",
		BucketName:         "test-bucket",
		Region:             "us-east-1",
		AccessKeyEncrypted: "encrypted_access",
		SecretKeyEncrypted: "encrypted_secret",
		IsEnabled:          true,
	}
	err := repo.SaveSyncCredential(cred)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Retrieve the credential
	retrieved, err := repo.GetSyncCredentials()
	if err != nil {
		t.Fatalf("GetSyncCredentials failed: %v", err)
	}

	if retrieved.Endpoint != cred.Endpoint {
		t.Errorf("Expected endpoint %s, got %s", cred.Endpoint, retrieved.Endpoint)
	}
	if retrieved.BucketName != cred.BucketName {
		t.Errorf("Expected bucket %s, got %s", cred.BucketName, retrieved.BucketName)
	}
	if retrieved.Region != cred.Region {
		t.Errorf("Expected region %s, got %s", cred.Region, retrieved.Region)
	}
}

func TestGetSyncCredentials_notFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewRepository(db)

	_, err := repo.GetSyncCredentials()
	if err == nil {
		t.Error("Expected error when no sync credential exists")
	}
}

func TestDeleteSyncCredential(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewRepository(db)

	// Create a credential
	cred := &models.SyncCredential{
		Endpoint:   "https://s3.amazonaws.com",
		BucketName: "test-bucket",
		IsEnabled:  true,
	}
	err := repo.SaveSyncCredential(cred)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Delete the credential
	err = repo.DeleteSyncCredential(string(cred.ID))
	if err != nil {
		t.Fatalf("DeleteSyncCredential failed: %v", err)
	}

	// Verify it's deleted
	_, err = repo.GetSyncCredentials()
	if err == nil {
		t.Error("Expected error when retrieving deleted credential")
	}
}

func TestDisableAllSyncCredentials(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewRepository(db)

	// Create multiple credentials
	cred1 := &models.SyncCredential{
		Endpoint:   "https://s3.amazonaws.com",
		BucketName: "bucket1",
		IsEnabled:  true,
	}
	cred2 := &models.SyncCredential{
		Endpoint:   "https://s3.amazonaws.com",
		BucketName: "bucket2",
		IsEnabled:  true,
	}

	repo.SaveSyncCredential(cred1)
	repo.SaveSyncCredential(cred2)

	// Disable all
	err := repo.DisableAllSyncCredentials()
	if err != nil {
		t.Fatalf("DisableAllSyncCredentials failed: %v", err)
	}

	// Verify all are disabled
	_, err = repo.GetSyncCredentials()
	if err == nil {
		t.Error("Expected error when all credentials are disabled")
	}
}
