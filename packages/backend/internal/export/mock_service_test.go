package export

import (
	"testing"
	"time"
)

// TestNewMockExportService verifies mock service creation.
func TestNewMockExportService(t *testing.T) {
	mock := NewMockExportService()
	if mock == nil {
		t.Fatal("NewMockExportService() should not return nil")
	}
	if !mock.shouldSucceed {
		t.Error("NewMockExportService() should default to success")
	}
}

// TestMockExportService_Export verifies mock export functionality.
func TestMockExportService_Export(t *testing.T) {
	mock := NewMockExportService()

	config := &ExportConfig{
		OutputPath: "/tmp/test_mock_export.tar.gz",
	}

	result, err := mock.Export(config)
	if err != nil {
		t.Fatalf("Export() should succeed, got: %v", err)
	}
	if result == nil {
		t.Fatal("Export() should return result")
	}
	if !mock.exportCalled {
		t.Error("Export() should set exportCalled flag")
	}
	if result.FilePath != config.OutputPath {
		t.Errorf("Export() should set correct path, got: %s", result.FilePath)
	}
}

// TestMockExportService_Export_defaultPath verifies default path handling.
func TestMockExportService_Export_defaultPath(t *testing.T) {
	mock := NewMockExportService()

	result, err := mock.Export(&ExportConfig{})
	if err != nil {
		t.Fatalf("Export() should succeed, got: %v", err)
	}
	if result.FilePath != "mock_export.tar.gz" {
		t.Errorf("Export() should use default path, got: %s", result.FilePath)
	}
}

// TestMockExportService_Export_encrypted verifies encrypted export tracking.
func TestMockExportService_Export_encrypted(t *testing.T) {
	mock := NewMockExportService()

	config := &ExportConfig{
		Password: "test-password-123",
	}

	result, err := mock.Export(config)
	if err != nil {
		t.Fatalf("Export() should succeed, got: %v", err)
	}
	if !result.Encrypted {
		t.Error("Export() should mark result as encrypted when password provided")
	}
}

// TestMockExportService_SetShouldSucceed verifies failure simulation.
func TestMockExportService_SetShouldSucceed(t *testing.T) {
	mock := NewMockExportService()
	mock.SetShouldSucceed(false)

	_, err := mock.Export(&ExportConfig{})
	if err == nil {
		t.Error("Export() should fail when SetShouldSucceed(false)")
	}
}

// TestMockExportService_SetExportDelay verifies delay simulation.
func TestMockExportService_SetExportDelay(t *testing.T) {
	mock := NewMockExportService()
	delay := 10 * time.Millisecond
	mock.SetExportDelay(delay)

	start := time.Now()
	mock.Export(&ExportConfig{})
	elapsed := time.Since(start)

	if elapsed < delay {
		t.Errorf("Export() should respect delay, expected >= %v, got: %v", delay, elapsed)
	}
}

// TestMockExportService_WasExportCalled verifies call tracking.
func TestMockExportService_WasExportCalled(t *testing.T) {
	mock := NewMockExportService()

	if mock.WasExportCalled() {
		t.Error("WasExportCalled() should be false initially")
	}

	mock.Export(&ExportConfig{})

	if !mock.WasExportCalled() {
		t.Error("WasExportCalled() should be true after Export()")
	}
}

// TestMockExportService_GetCallCount verifies call counting.
func TestMockExportService_GetCallCount(t *testing.T) {
	mock := NewMockExportService()

	if mock.GetCallCount() != 0 {
		t.Errorf("GetCallCount() should be 0 initially, got: %d", mock.GetCallCount())
	}

	mock.Export(&ExportConfig{})
	if mock.GetCallCount() != 1 {
		t.Errorf("GetCallCount() should be 1 after one export, got: %d", mock.GetCallCount())
	}

	mock.Export(&ExportConfig{})
	if mock.GetCallCount() != 2 {
		t.Errorf("GetCallCount() should be 2 after two exports, got: %d", mock.GetCallCount())
	}
}

// TestMockExportService_GetLastConfig verifies config tracking.
func TestMockExportService_GetLastConfig(t *testing.T) {
	mock := NewMockExportService()

	config := &ExportConfig{
		OutputPath: "/tmp/test.tar.gz",
		Password:   "test-password",
	}

	mock.Export(config)

	lastConfig := mock.GetLastConfig()
	if lastConfig == nil {
		t.Fatal("GetLastConfig() should return config")
	}
	if lastConfig.OutputPath != config.OutputPath {
		t.Errorf("GetLastConfig() should return correct path, got: %s", lastConfig.OutputPath)
	}
	if lastConfig.Password != config.Password {
		t.Errorf("GetLastConfig() should return correct password")
	}
}

// TestMockExportService_GetExportPath verifies path tracking.
func TestMockExportService_GetExportPath(t *testing.T) {
	mock := NewMockExportService()
	expectedPath := "/tmp/test_export.tar.gz"

	mock.Export(&ExportConfig{OutputPath: expectedPath})

	actualPath := mock.GetExportPath()
	if actualPath != expectedPath {
		t.Errorf("GetExportPath() should return %s, got: %s", expectedPath, actualPath)
	}
}

// TestMockExportService_Reset verifies state reset.
func TestMockExportService_Reset(t *testing.T) {
	mock := NewMockExportService()

	mock.Export(&ExportConfig{})
	mock.SetShouldSucceed(false)

	mock.Reset()

	if mock.WasExportCalled() {
		t.Error("Reset() should clear exportCalled flag")
	}
	if mock.GetCallCount() != 0 {
		t.Errorf("Reset() should clear call count, got: %d", mock.GetCallCount())
	}
	if mock.GetLastConfig() != nil {
		t.Error("Reset() should clear last config")
	}
	// Note: Reset() does not reset shouldSucceed flag - it preserves configuration
	// To restore success state, explicitly call SetShouldSucceed(true)
}

