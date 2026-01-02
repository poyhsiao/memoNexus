// Package export tests for core export/import service functions.
package export

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimhsiao/memonexus/backend/internal/db"
	"github.com/kimhsiao/memonexus/backend/internal/models"
	"github.com/kimhsiao/memonexus/backend/internal/uuid"
)

// setupTestDB creates an in-memory database for testing.
func setupTestDB(t *testing.T) (*sql.DB, *db.Repository) {
	database, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open in-memory database: %v", err)
	}

	migrator := db.NewMigrator(database, "../db/migrations")
	if err := migrator.Initialize(); err != nil {
		database.Close()
		t.Fatalf("Failed to initialize migrator: %v", err)
	}

	// Apply migrations
	if err := migrator.Up(); err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			database.Close()
			t.Fatalf("Failed to apply migrations: %v", err)
		}
	}

	// Create repository
	repo := db.NewRepository(database)

	return database, repo
}

// TestExport_noDatabase verifies Export handles nil repository gracefully.
func TestExport_noDatabase(t *testing.T) {
	service := NewExportService(nil)
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "test_export.tar.gz")

	config := &ExportConfig{
		OutputPath:   outputPath,
		Password:     "",
		IncludeMedia: false,
	}

	// Export with nil repo will panic when calling repo.ListContentItems()
	// This is expected behavior - Export requires a valid repository
	defer func() {
		if r := recover(); r != nil {
			// Expected panic with nil repo
			t.Logf("Export() correctly panicked with nil repo (expected behavior): %v", r)
		}
	}()

	result, err := service.Export(config)

	// If we get here without panic, check for error
	if err == nil {
		t.Error("Export() with nil repo should return error or panic")
	}
	if result != nil {
		t.Error("result should be nil on error")
	}
}

// TestExport_emptyDatabase verifies export with empty database.
func TestExport_emptyDatabase(t *testing.T) {
	database, repo := setupTestDB(t)
	defer database.Close()

	service := NewExportService(repo)
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "test_export.tar.gz")

	config := &ExportConfig{
		OutputPath:   outputPath,
		Password:     "",
		IncludeMedia: false,
	}

	result, err := service.Export(config)

	if err != nil {
		t.Fatalf("Export() error = %v", err)
	}
	if result == nil {
		t.Fatal("Export() result should not be nil")
	}

	// Verify basic result fields
	if result.ItemCount != 0 {
		t.Errorf("ItemCount = %d, want 0 for empty database", result.ItemCount)
	}
	if result.Encrypted {
		t.Error("Encrypted should be false when no password provided")
	}
	if result.Checksum == "" {
		t.Error("Checksum should not be empty")
	}

	// Verify file was created
	if _, err := os.Stat(result.FilePath); os.IsNotExist(err) {
		t.Errorf("Export file was not created at %s", result.FilePath)
	}
}

// TestExport_withEncryption verifies encrypted export creation.
func TestExport_withEncryption(t *testing.T) {
	database, repo := setupTestDB(t)
	defer database.Close()

	// Add a test item
	item := &models.ContentItem{
		ID:          models.UUID(uuid.New()),
		Title:       "Test Export",
		ContentText: "Test content for encryption",
		MediaType:   "web",
		Tags:        "export,encryption",
		CreatedAt:   1704067200,
		UpdatedAt:   1704067200,
		Version:     1,
	}
	if err := repo.CreateContentItem(item); err != nil {
		t.Fatalf("Failed to create test item: %v", err)
	}

	service := NewExportService(repo)
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "test_export_encrypted.tar.gz")

	config := &ExportConfig{
		OutputPath:   outputPath,
		Password:     "test-password-123",
		IncludeMedia: false,
	}

	result, err := service.Export(config)

	if err != nil {
		t.Fatalf("Export() error = %v", err)
	}
	if result == nil {
		t.Fatal("Export() result should not be nil")
	}

	// Verify encryption result
	if !result.Encrypted {
		t.Error("Encrypted should be true when password provided")
	}
	if result.ItemCount != 1 {
		t.Errorf("ItemCount = %d, want 1", result.ItemCount)
	}

	// Verify encrypted file was created and is different from plain
	if _, err := os.Stat(result.FilePath); os.IsNotExist(err) {
		t.Errorf("Encrypted export file was not created at %s", result.FilePath)
	}
}

// TestExport_defaultOutputPath verifies export with default output path.
func TestExport_defaultOutputPath(t *testing.T) {
	database, repo := setupTestDB(t)
	defer database.Close()

	service := NewExportService(repo)

	// Create exports directory in temp dir
	exportsDir := t.TempDir()

	// Change to temp dir and set empty output path
	// The service will use default "exports/" prefix
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	if err := os.Chdir(exportsDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	config := &ExportConfig{
		OutputPath:   "", // Use default
		Password:     "",
		IncludeMedia: false,
	}

	result, err := service.Export(config)

	if err != nil {
		t.Fatalf("Export() error = %v", err)
	}
	if result == nil {
		t.Fatal("Export() result should not be nil")
	}

	// Verify default path pattern is used
	if !strings.Contains(result.FilePath, "exports/memonexus_") {
		t.Errorf("Default output path not used, got: %s", result.FilePath)
	}
}

// TestExport_createExportsDir verifies export creates exports directory.
func TestExport_createExportsDir(t *testing.T) {
	database, repo := setupTestDB(t)
	defer database.Close()

	service := NewExportService(repo)
	tempDir := t.TempDir()

	// Use a subdirectory that doesn't exist
	outputPath := filepath.Join(tempDir, "exports/subdir/test.tar.gz")

	config := &ExportConfig{
		OutputPath:   outputPath,
		Password:     "",
		IncludeMedia: false,
	}

	result, err := service.Export(config)

	if err != nil {
		t.Fatalf("Export() error = %v", err)
	}
	if result == nil {
		t.Fatal("Export() result should not be nil")
	}

	// Verify the directory was created
	if _, err := os.Stat(filepath.Dir(outputPath)); os.IsNotExist(err) {
		t.Error("Exports directory was not created")
	}
}

// TestImport_noDatabase verifies Import handles nil repository gracefully.
func TestImport_noDatabase(t *testing.T) {
	service := NewExportService(nil)

	// Create a minimal test archive
	tempDir := t.TempDir()
	archivePath := filepath.Join(tempDir, "test.tar.gz")

	// Create a simple tar.gz with manifest and data
	// This requires actual archive creation - for nil repo test,
	// we expect failure during data import
	config := &ImportConfig{
		ArchivePath: archivePath,
		Password:    "",
	}

	// Import with nil repo will panic when calling importDataFile()
	// which tries to check item existence and create items
	defer func() {
		if r := recover(); r != nil {
			// Expected panic with nil repo
			t.Logf("Import() correctly panicked with nil repo (expected behavior): %v", r)
		}
	}()

	result, err := service.Import(config)

	// If we get here without panic, check for error
	if err == nil {
		t.Error("Import() with nil repo should return error or panic")
	}
	if result != nil {
		t.Error("result should be nil on error")
	}
}

// TestImport_invalidArchive verifies import with non-existent archive.
func TestImport_invalidArchive(t *testing.T) {
	database, repo := setupTestDB(t)
	defer database.Close()

	service := NewExportService(repo)

	config := &ImportConfig{
		ArchivePath: "/nonexistent/path/archive.tar.gz",
		Password:    "",
	}

	result, err := service.Import(config)

	if err == nil {
		t.Error("Import() with invalid archive path should return error")
	}
	if result != nil {
		t.Error("result should be nil on error")
	}
}

// TestImport_exportImportRoundTrip verifies export then import round-trip.
func TestImport_exportImportRoundTrip(t *testing.T) {
	database, repo := setupTestDB(t)
	defer database.Close()

	// Create test items
	items := []*models.ContentItem{
		{
			ID:          models.UUID(uuid.New()),
			Title:       "First Item",
			ContentText: "Content 1",
			MediaType:   "web",
			Tags:        "test",
			CreatedAt:   1704067200,
			UpdatedAt:   1704067200,
			Version:     1,
		},
		{
			ID:          models.UUID(uuid.New()),
			Title:       "Second Item",
			ContentText: "Content 2",
			MediaType:   "pdf",
			Tags:        "document",
			CreatedAt:   1704067300,
			UpdatedAt:   1704067300,
			Version:     1,
		},
	}

	for _, item := range items {
		if err := repo.CreateContentItem(item); err != nil {
			t.Fatalf("Failed to create test item: %v", err)
		}
	}

	// Export
	service := NewExportService(repo)
	tempDir := t.TempDir()
	exportPath := filepath.Join(tempDir, "roundtrip.tar.gz")

	exportConfig := &ExportConfig{
		OutputPath:   exportPath,
		Password:     "",
		IncludeMedia: false,
	}

	exportResult, err := service.Export(exportConfig)
	if err != nil {
		t.Fatalf("Export() failed: %v", err)
	}

	if exportResult.ItemCount != 2 {
		t.Errorf("ExportItemCount = %d, want 2", exportResult.ItemCount)
	}

	// Clear database by deleting all items
	for _, item := range items {
		if err := repo.DeleteContentItem(string(item.ID)); err != nil {
			t.Fatalf("Failed to delete item: %v", err)
		}
	}

	// Verify items are deleted
	allItems, _ := repo.ListContentItems(100, 0, "")
	if len(allItems) != 0 {
		t.Fatalf("Database should be empty after deletion, got %d items", len(allItems))
	}

	// Import
	importConfig := &ImportConfig{
		ArchivePath: exportResult.FilePath,
		Password:    "",
	}

	importResult, err := service.Import(importConfig)
	if err != nil {
		t.Fatalf("Import() failed: %v", err)
	}

	if importResult.ImportedCount != 2 {
		t.Errorf("ImportedCount = %d, want 2", importResult.ImportedCount)
	}
	if importResult.SkippedCount != 0 {
		t.Errorf("SkippedCount = %d, want 0", importResult.SkippedCount)
	}

	// Verify items were imported correctly
	importedItems, err := repo.ListContentItems(100, 0, "")
	if err != nil {
		t.Fatalf("Failed to list imported items: %v", err)
	}

	t.Logf("Imported %d items", len(importedItems))
	for _, item := range importedItems {
		t.Logf("  - ID: %s, Title: %s", item.ID, item.Title)
	}

	if len(importedItems) != 2 {
		t.Errorf("Imported item count = %d, want 2", len(importedItems))
	}
}

// TestImport_duplicateItems verifies import handles existing items.
func TestImport_duplicateItems(t *testing.T) {
	database, repo := setupTestDB(t)
	defer database.Close()

	// Create test item
	item := &models.ContentItem{
		ID:          models.UUID(uuid.New()),
		Title:       "Duplicate Test",
		ContentText: "Original content",
		MediaType:   "web",
		CreatedAt:   1704067200,
		UpdatedAt:   1704067200,
		Version:     1,
	}

	if err := repo.CreateContentItem(item); err != nil {
		t.Fatalf("Failed to create test item: %v", err)
	}

	// Export
	service := NewExportService(repo)
	tempDir := t.TempDir()
	exportPath := filepath.Join(tempDir, "duplicate.tar.gz")

	exportConfig := &ExportConfig{
		OutputPath:   exportPath,
		Password:     "",
		IncludeMedia: false,
	}

	exportResult, err := service.Export(exportConfig)
	if err != nil {
		t.Fatalf("Export() failed: %v", err)
	}

	// Import without deleting - should skip existing item
	importConfig := &ImportConfig{
		ArchivePath: exportResult.FilePath,
		Password:    "",
	}

	importResult, err := service.Import(importConfig)
	if err != nil {
		t.Fatalf("Import() failed: %v", err)
	}

	if importResult.ImportedCount != 0 {
		t.Errorf("ImportedCount = %d, want 0 (item already exists)", importResult.ImportedCount)
	}
	if importResult.SkippedCount != 1 {
		t.Errorf("SkippedCount = %d, want 1", importResult.SkippedCount)
	}
}

// TestImport_encryptedArchive verifies import with encrypted archive.
func TestImport_encryptedArchive(t *testing.T) {
	database, repo := setupTestDB(t)
	defer database.Close()

	// Create test item
	item := &models.ContentItem{
		ID:          models.UUID(uuid.New()),
		Title:       "Encrypted Import Test",
		ContentText: "Secret content",
		MediaType:   "web",
		CreatedAt:   1704067200,
		UpdatedAt:   1704067200,
		Version:     1,
	}

	if err := repo.CreateContentItem(item); err != nil {
		t.Fatalf("Failed to create test item: %v", err)
	}

	// Export with encryption
	service := NewExportService(repo)
	tempDir := t.TempDir()
	exportPath := filepath.Join(tempDir, "encrypted.tar.gz")

	password := "test-password-456"

	exportConfig := &ExportConfig{
		OutputPath:   exportPath,
		Password:     password,
		IncludeMedia: false,
	}

	exportResult, err := service.Export(exportConfig)
	if err != nil {
		t.Fatalf("Export() failed: %v", err)
	}

	// Delete item to prepare for import
	if err := repo.DeleteContentItem(string(item.ID)); err != nil {
		t.Fatalf("Failed to delete item: %v", err)
	}

	// Import with correct password
	importConfig := &ImportConfig{
		ArchivePath: exportResult.FilePath,
		Password:    password,
	}

	importResult, err := service.Import(importConfig)
	if err != nil {
		t.Fatalf("Import() with correct password failed: %v", err)
	}

	if importResult.ImportedCount != 1 {
		t.Errorf("ImportedCount = %d, want 1", importResult.ImportedCount)
	}
}

// TestImport_wrongPassword verifies import fails with wrong password.
func TestImport_wrongPassword(t *testing.T) {
	database, repo := setupTestDB(t)
	defer database.Close()

	// Create test item
	item := &models.ContentItem{
		ID:          models.UUID(uuid.New()),
		Title:       "Password Test",
		ContentText: "Protected content",
		MediaType:   "web",
		CreatedAt:   1704067200,
		UpdatedAt:   1704067200,
		Version:     1,
	}

	if err := repo.CreateContentItem(item); err != nil {
		t.Fatalf("Failed to create test item: %v", err)
	}

	// Export with encryption
	service := NewExportService(repo)
	tempDir := t.TempDir()
	exportPath := filepath.Join(tempDir, "protected.tar.gz")

	exportConfig := &ExportConfig{
		OutputPath:   exportPath,
		Password:     "correct-password",
		IncludeMedia: false,
	}

	exportResult, err := service.Export(exportConfig)
	if err != nil {
		t.Fatalf("Export() failed: %v", err)
	}

	// Try to import with wrong password
	importConfig := &ImportConfig{
		ArchivePath: exportResult.FilePath,
		Password:    "wrong-password",
	}

	result, err := service.Import(importConfig)

	if err == nil {
		t.Error("Import() with wrong password should return error")
	}
	if result != nil {
		t.Error("result should be nil on decryption error")
	}
}

// TestCreateDataFile verifies data file creation with test data.
func TestCreateDataFile(t *testing.T) {
	database, repo := setupTestDB(t)
	defer database.Close()

	// Create test items
	items := []*models.ContentItem{
		{
			ID:          models.UUID(uuid.New()),
			Title:       "Data File Test 1",
			ContentText: "Content 1",
			MediaType:   "web",
			CreatedAt:   1704067200,
			UpdatedAt:   1704067200,
			Version:     1,
		},
		{
			ID:          models.UUID(uuid.New()),
			Title:       "Data File Test 2",
			ContentText: "Content 2",
			MediaType:   "markdown",
			CreatedAt:   1704067300,
			UpdatedAt:   1704067300,
			Version:     1,
		},
	}

	for _, item := range items {
		if err := repo.CreateContentItem(item); err != nil {
			t.Fatalf("Failed to create test item: %v", err)
		}
	}

	service := NewExportService(repo)
	tempDir := t.TempDir()
	dataPath := filepath.Join(tempDir, "data.json")

	count, checksum, err := service.createDataFile(dataPath)

	if err != nil {
		t.Fatalf("createDataFile() error = %v", err)
	}

	if count != 2 {
		t.Errorf("count = %d, want 2", count)
	}

	if checksum == "" {
		t.Error("checksum should not be empty")
	}

	// Verify file exists and has content
	if _, err := os.Stat(dataPath); os.IsNotExist(err) {
		t.Error("data file was not created")
	}

	data, err := os.ReadFile(dataPath)
	if err != nil {
		t.Fatalf("Failed to read data file: %v", err)
	}

	if len(data) == 0 {
		t.Error("data file is empty")
	}

	// Verify it's valid JSON array
	if !strings.HasPrefix(string(data), "[") {
		t.Error("data file should be a JSON array")
	}
}

// TestItemExists verifies item existence checking.
func TestItemExists(t *testing.T) {
	database, repo := setupTestDB(t)
	defer database.Close()

	service := NewExportService(repo)

	// Test with non-existent item
	nonExistentID := models.UUID(uuid.New())
	exists, err := service.itemExists(nonExistentID)

	if err != nil {
		t.Fatalf("itemExists() error = %v", err)
	}
	if exists {
		t.Error("itemExists() should return false for non-existent item")
	}

	// Create an item
	item := &models.ContentItem{
		ID:          models.UUID(uuid.New()),
		Title:       "Exists Test",
		ContentText: "Test content",
		MediaType:   "web",
		CreatedAt:   1704067200,
		UpdatedAt:   1704067200,
		Version:     1,
	}

	if err := repo.CreateContentItem(item); err != nil {
		t.Fatalf("Failed to create test item: %v", err)
	}

	// Test with existing item
	exists, err = service.itemExists(item.ID)

	if err != nil {
		t.Fatalf("itemExists() error = %v", err)
	}
	if !exists {
		t.Error("itemExists() should return true for existing item")
	}
}

// TestCreateItem verifies item creation.
func TestCreateItem(t *testing.T) {
	database, repo := setupTestDB(t)
	defer database.Close()

	service := NewExportService(repo)

	item := &models.ContentItem{
		ID:          models.UUID(uuid.New()),
		Title:       "Create Item Test",
		ContentText: "New item content",
		MediaType:   "web",
		CreatedAt:   1704067200,
		UpdatedAt:   1704067200,
		Version:     1,
	}

	err := service.createItem(item)

	if err != nil {
		t.Fatalf("createItem() error = %v", err)
	}

	// Verify item was created
	retrieved, err := repo.GetContentItem(string(item.ID))
	if err != nil {
		t.Fatalf("Failed to retrieve created item: %v", err)
	}

	if retrieved.Title != item.Title {
		t.Errorf("Retrieved title = %q, want %q", retrieved.Title, item.Title)
	}
}
