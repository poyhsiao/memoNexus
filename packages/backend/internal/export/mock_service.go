// Package export provides mock implementations for testing.
package export

import (
	"fmt"
	"os"
	"sync"
	"time"
)

// MockExportService is a mock implementation of ExportServiceInterface for testing.
type MockExportService struct {
	mu            sync.Mutex
	shouldSucceed bool
	exportDelay   time.Duration // Simulate export delay
	exportCalled  bool
	lastConfig    *ExportConfig
	exportPath    string
	callCount     int
}

// NewMockExportService creates a new mock export service.
func NewMockExportService() *MockExportService {
	return &MockExportService{
		shouldSucceed: true,
		exportDelay:   0,
	}
}

// Export performs a mock export operation.
func (m *MockExportService) Export(config *ExportConfig) (*ExportResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.exportCalled = true
	m.callCount++
	m.lastConfig = config

	// Simulate delay if configured
	if m.exportDelay > 0 {
		time.Sleep(m.exportDelay)
	}

	if !m.shouldSucceed {
		return nil, fmt.Errorf("mock export failed")
	}

	// Determine output path
	outputPath := config.OutputPath
	if outputPath == "" {
		outputPath = "mock_export.tar.gz"
	}
	m.exportPath = outputPath

	// Create mock export file
	if err := os.WriteFile(outputPath, []byte("mock export data"), 0644); err != nil {
		return nil, fmt.Errorf("failed to create mock export file: %w", err)
	}

	return &ExportResult{
		FilePath:  outputPath,
		SizeBytes: 1024,
		ItemCount: 0,
		Checksum:  "mock-checksum-12345",
		Encrypted: len(config.Password) > 0,
		Duration:  time.Millisecond * 10,
	}, nil
}

// SetShouldSucceed controls whether the mock export will succeed.
func (m *MockExportService) SetShouldSucceed(shouldSucceed bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldSucceed = shouldSucceed
}

// SetExportDelay sets a delay for export operations (useful for testing cancellation).
func (m *MockExportService) SetExportDelay(delay time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.exportDelay = delay
}

// WasExportCalled returns true if Export was called.
func (m *MockExportService) WasExportCalled() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.exportCalled
}

// GetCallCount returns the number of times Export was called.
func (m *MockExportService) GetCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.callCount
}

// GetLastConfig returns the config passed to the last Export call.
func (m *MockExportService) GetLastConfig() *ExportConfig {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.lastConfig
}

// GetExportPath returns the path of the last export.
func (m *MockExportService) GetExportPath() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.exportPath
}

// Reset resets the mock state.
func (m *MockExportService) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.exportCalled = false
	m.callCount = 0
	m.lastConfig = nil
}
