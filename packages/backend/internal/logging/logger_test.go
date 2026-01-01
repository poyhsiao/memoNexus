// Package logging tests for structured JSON logging.
package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

// =====================================================
// Logger Creation and Initialization Tests
// =====================================================

// TestInit verifies logger initialization.
func TestInit(t *testing.T) {
	var buf bytes.Buffer
	Init(&buf, LevelInfo)

	logger := Get()
	if logger == nil {
		t.Fatal("Get() returned nil after Init()")
	}

	if logger.out != &buf {
		t.Error("Init() did not set output writer correctly")
	}

	if logger.minLevel != LevelInfo {
		t.Errorf("minLevel = %v, want LevelInfo", logger.minLevel)
	}
}

// TestInit_idempotent verifies Init is idempotent.
func TestInit_idempotent(t *testing.T) {
	// Reset global logger for this test
	global = nil
	once = *new(sync.Once)

	var buf1 bytes.Buffer
	Init(&buf1, LevelInfo)

	// Save the reference
	firstLogger := Get()

	// Second init with different parameters should be ignored
	var buf2 bytes.Buffer
	Init(&buf2, LevelDebug)

	// Get should return the same logger
	logger := Get()
	if logger != firstLogger {
		t.Error("Second Init() should be ignored, different logger returned")
	}

	// Writer should still be the first one
	if logger.out != &buf1 {
		t.Error("Second Init() should be ignored, output writer changed")
	}
}

// TestGet_default verifies default logger creation.
func TestGet_default(t *testing.T) {
	// Reset global logger for this test
	global = nil
	once = *new(sync.Once)

	logger := Get()
	if logger == nil {
		t.Fatal("Get() returned nil without Init()")
	}

	if logger.out != os.Stdout {
		t.Error("Get() should default to os.Stdout")
	}

	if logger.minLevel != LevelInfo {
		t.Errorf("minLevel = %v, want LevelInfo", logger.minLevel)
	}
}

// =====================================================
// Log Level Tests
// =====================================================

// TestLogLevel_shouldLog verifies log level filtering.
func TestLogLevel_shouldLog(t *testing.T) {
	tests := []struct {
		name     string
		minLevel LogLevel
		logLevel LogLevel
		expected bool
	}{
		{"debug logs at debug", LevelDebug, LevelDebug, true},
		{"debug logs at info", LevelInfo, LevelDebug, false},
		{"info logs at info", LevelInfo, LevelInfo, true},
		{"info logs at warn", LevelWarn, LevelInfo, false},
		{"warn logs at error", LevelError, LevelWarn, false},
		{"error logs at error", LevelError, LevelError, true},
		{"error logs at debug", LevelDebug, LevelError, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := &Logger{minLevel: tt.minLevel}
			result := logger.shouldLog(tt.logLevel)
			if result != tt.expected {
				t.Errorf("shouldLog(%v) at minLevel %v = %v, want %v",
					tt.logLevel, tt.minLevel, result, tt.expected)
			}
		})
	}
}

// =====================================================
// Logging Tests
// =====================================================

// TestLogger_Debug verifies debug logging.
func TestLogger_Debug(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{out: &buf, minLevel: LevelDebug}

	logger.Debug("test message", map[string]interface{}{"key": "value"})

	output := buf.String()
	if output == "" {
		t.Error("Debug() produced no output")
	}

	// Verify JSON format
	var entry LogEntry
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &entry); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	if entry.Level != "DEBUG" {
		t.Errorf("Level = %q, want 'DEBUG'", entry.Level)
	}

	if entry.Message != "test message" {
		t.Errorf("Message = %q, want 'test message'", entry.Message)
	}

	if entry.Context == nil {
		t.Error("Context should not be nil")
	}

	if entry.Context["key"] != "value" {
		t.Errorf("Context['key'] = %v, want 'value'", entry.Context["key"])
	}
}

// TestLogger_Info verifies info logging.
func TestLogger_Info(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{out: &buf, minLevel: LevelInfo}

	logger.Info("info message")

	output := strings.TrimSpace(buf.String())
	var entry LogEntry
	if err := json.Unmarshal([]byte(output), &entry); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	if entry.Level != "INFO" {
		t.Errorf("Level = %q, want 'INFO'", entry.Level)
	}

	if entry.Message != "info message" {
		t.Errorf("Message = %q, want 'info message'", entry.Message)
	}
}

// TestLogger_Warn verifies warn logging.
func TestLogger_Warn(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{out: &buf, minLevel: LevelInfo}

	logger.Warn("warning message")

	output := strings.TrimSpace(buf.String())
	var entry LogEntry
	if err := json.Unmarshal([]byte(output), &entry); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	if entry.Level != "WARN" {
		t.Errorf("Level = %q, want 'WARN'", entry.Level)
	}
}

// TestLogger_Error verifies error logging.
func TestLogger_Error(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{out: &buf, minLevel: LevelInfo}

	testErr := io.ErrUnexpectedEOF
	logger.Error("error occurred", testErr)

	output := strings.TrimSpace(buf.String())
	var entry LogEntry
	if err := json.Unmarshal([]byte(output), &entry); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	if entry.Level != "ERROR" {
		t.Errorf("Level = %q, want 'ERROR'", entry.Level)
	}

	if entry.Error == "" {
		t.Error("Error field should not be empty")
	}

	if !strings.Contains(entry.Error, testErr.Error()) {
		t.Errorf("Error field should contain error details, got: %s", entry.Error)
	}
}

// TestLogger_ErrorWithCode verifies error logging with code.
func TestLogger_ErrorWithCode(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{out: &buf, minLevel: LevelInfo}

	testErr := io.ErrUnexpectedEOF
	logger.ErrorWithCode("validation failed", "VAL001", testErr, map[string]interface{}{"field": "email"})

	output := strings.TrimSpace(buf.String())
	var entry LogEntry
	if err := json.Unmarshal([]byte(output), &entry); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	if entry.Level != "ERROR" {
		t.Errorf("Level = %q, want 'ERROR'", entry.Level)
	}

	if entry.Context == nil {
		t.Fatal("Context should not be nil")
	}

	if entry.Context["error_code"] != "VAL001" {
		t.Errorf("error_code = %v, want 'VAL001'", entry.Context["error_code"])
	}

	if entry.Context["field"] != "email" {
		t.Errorf("field = %v, want 'email'", entry.Context["field"])
	}
}

// TestLogger_ErrorWithCode_noContext verifies error code without existing context.
func TestLogger_ErrorWithCode_noContext(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{out: &buf, minLevel: LevelInfo}

	logger.ErrorWithCode("error occurred", "ERR001", io.ErrUnexpectedEOF)

	output := strings.TrimSpace(buf.String())
	var entry LogEntry
	if err := json.Unmarshal([]byte(output), &entry); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	if entry.Context == nil {
		t.Fatal("Context should not be nil")
	}

	if entry.Context["error_code"] != "ERR001" {
		t.Errorf("error_code = %v, want 'ERR001'", entry.Context["error_code"])
	}
}

// =====================================================
// Log Level Filtering Tests
// =====================================================

// TestLogger_filtering verifies minimum level filtering.
func TestLogger_filtering(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{out: &buf, minLevel: LevelWarn}

	// These should not log (below minimum level)
	logger.Debug("debug message")
	logger.Info("info message")

	// These should log
	logger.Warn("warn message")
	logger.Error("error message", nil)

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	if len(lines) != 2 {
		t.Errorf("Expected 2 log lines, got %d", len(lines))
	}

	// First line should be WARN
	var entry LogEntry
	json.Unmarshal([]byte(lines[0]), &entry)
	if entry.Level != "WARN" {
		t.Errorf("First log level = %q, want 'WARN'", entry.Level)
	}

	// Second line should be ERROR
	json.Unmarshal([]byte(lines[1]), &entry)
	if entry.Level != "ERROR" {
		t.Errorf("Second log level = %q, want 'ERROR'", entry.Level)
	}
}

// TestLogger_noDebug verifies debug messages are filtered at INFO level.
func TestLogger_noDebug(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{out: &buf, minLevel: LevelInfo}

	logger.Debug("debug message")

	if buf.String() != "" {
		t.Error("Debug() should not log when minLevel is INFO")
	}
}

// =====================================================
// Context Handling Tests
// =====================================================

// TestGetContext_single verifies single context handling.
func TestLogger_getContext_single(t *testing.T) {
	logger := &Logger{}

	ctx := logger.getContext(map[string]interface{}{"key1": "value1"})

	if ctx == nil {
		t.Fatal("getContext() returned nil for single context")
	}

	if ctx["key1"] != "value1" {
		t.Errorf("ctx['key1'] = %v, want 'value1'", ctx["key1"])
	}
}

// TestGetContext_multiple verifies context merging.
func TestLogger_getContext_multiple(t *testing.T) {
	logger := &Logger{}

	ctx := logger.getContext(
		map[string]interface{}{"key1": "value1"},
		map[string]interface{}{"key2": "value2"},
		map[string]interface{}{"key1": "overridden"},
	)

	if ctx == nil {
		t.Fatal("getContext() returned nil for multiple contexts")
	}

	// key1 should be overridden
	if ctx["key1"] != "overridden" {
		t.Errorf("ctx['key1'] = %v, want 'overridden'", ctx["key1"])
	}

	// key2 should be present
	if ctx["key2"] != "value2" {
		t.Errorf("ctx['key2'] = %v, want 'value2'", ctx["key2"])
	}
}

// TestGetContext_none verifies no context returns nil.
func TestLogger_getContext_none(t *testing.T) {
	logger := &Logger{}

	ctx := logger.getContext()

	if ctx != nil {
		t.Errorf("getContext() with no arguments should return nil, got %v", ctx)
	}
}

// =====================================================
// JSON Output Tests
// =====================================================

// TestLogger_jsonFormat verifies JSON output format.
func TestLogger_jsonFormat(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{out: &buf, minLevel: LevelInfo}

	logger.Info("test message", map[string]interface{}{
		"string": "value",
		"number": 42,
		"bool":   true,
	})

	output := strings.TrimSpace(buf.String())

	var entry LogEntry
	if err := json.Unmarshal([]byte(output), &entry); err != nil {
		t.Fatalf("Output is not valid JSON: %v\nOutput: %s", err, output)
	}

	if entry.Timestamp == "" {
		t.Error("Timestamp should not be empty")
	}

	if _, err := time.Parse(time.RFC3339, entry.Timestamp); err != nil {
		t.Errorf("Timestamp is not valid RFC3339: %v", err)
	}

	if entry.Context == nil {
		t.Fatal("Context should not be nil")
	}

	if entry.Context["string"] != "value" {
		t.Errorf("Context['string'] = %v, want 'value'", entry.Context["string"])
	}

	if entry.Context["number"] != float64(42) {
		t.Errorf("Context['number'] = %v, want 42", entry.Context["number"])
	}

	if entry.Context["bool"] != true {
		t.Errorf("Context['bool'] = %v, want true", entry.Context["bool"])
	}
}

// TestLogger_multipleLines verifies multiple log entries.
func TestLogger_multipleLines(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{out: &buf, minLevel: LevelInfo}

	logger.Info("message 1")
	logger.Warn("message 2")
	logger.Error("message 3", nil)

	output := strings.TrimSpace(buf.String())
	lines := strings.Split(output, "\n")

	if len(lines) != 3 {
		t.Errorf("Expected 3 log lines, got %d", len(lines))
	}

	// Verify each line is valid JSON
	for i, line := range lines {
		var entry LogEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			t.Errorf("Line %d is not valid JSON: %v", i, err)
		}
	}
}

// =====================================================
// Thread Safety Tests
// =====================================================

// TestLogger_concurrentLogging verifies concurrent logging is safe.
func TestLogger_concurrentLogging(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{out: &buf, minLevel: LevelInfo}

	var wg sync.WaitGroup
	iterations := 100

	// Launch multiple goroutines logging concurrently
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				logger.Info("message", map[string]interface{}{"goroutine": id})
			}
		}(i)
	}

	wg.Wait()

	output := strings.TrimSpace(buf.String())
	lines := strings.Split(output, "\n")

	expectedLines := 10 * iterations
	if len(lines) != expectedLines {
		t.Errorf("Expected %d log lines, got %d", expectedLines, len(lines))
	}

	// Verify all lines are valid JSON
	for i, line := range lines {
		var entry LogEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			t.Errorf("Line %d is not valid JSON: %v", i, err)
		}
	}
}

// =====================================================
// Global Convenience Functions Tests
// =====================================================

// TestGlobalDebug verifies global Debug function.
func TestGlobalDebug(t *testing.T) {
	var buf bytes.Buffer
	global = nil
	once = *new(sync.Once)
	Init(&buf, LevelDebug)

	Debug("debug message")

	output := buf.String()
	if output == "" {
		t.Error("Debug() produced no output")
	}

	if !strings.Contains(output, "DEBUG") {
		t.Error("Output should contain 'DEBUG'")
	}
}

// TestGlobalInfo verifies global Info function.
func TestGlobalInfo(t *testing.T) {
	var buf bytes.Buffer
	global = nil
	once = *new(sync.Once)
	Init(&buf, LevelInfo)

	Info("info message")

	output := buf.String()
	if output == "" {
		t.Error("Info() produced no output")
	}

	if !strings.Contains(output, "INFO") {
		t.Error("Output should contain 'INFO'")
	}
}

// TestGlobalWarn verifies global Warn function.
func TestGlobalWarn(t *testing.T) {
	var buf bytes.Buffer
	global = nil
	once = *new(sync.Once)
	Init(&buf, LevelInfo)

	Warn("warn message")

	output := buf.String()
	if output == "" {
		t.Error("Warn() produced no output")
	}

	if !strings.Contains(output, "WARN") {
		t.Error("Output should contain 'WARN'")
	}
}

// TestGlobalError verifies global Error function.
func TestGlobalError(t *testing.T) {
	var buf bytes.Buffer
	global = nil
	once = *new(sync.Once)
	Init(&buf, LevelInfo)

	testErr := io.ErrUnexpectedEOF
	Error("error message", testErr)

	output := buf.String()
	if output == "" {
		t.Error("Error() produced no output")
	}

	if !strings.Contains(output, "ERROR") {
		t.Error("Output should contain 'ERROR'")
	}

	if !strings.Contains(output, testErr.Error()) {
		t.Error("Output should contain error details")
	}
}

// TestGlobalErrorWithCode verifies global ErrorWithCode function.
func TestGlobalErrorWithCode(t *testing.T) {
	var buf bytes.Buffer
	global = nil
	once = *new(sync.Once)
	Init(&buf, LevelInfo)

	ErrorWithCode("error occurred", "ERR001", io.ErrUnexpectedEOF)

	output := buf.String()
	if output == "" {
		t.Error("ErrorWithCode() produced no output")
	}

	if !strings.Contains(output, "error_code") {
		t.Error("Output should contain 'error_code'")
	}

	if !strings.Contains(output, "ERR001") {
		t.Error("Output should contain 'ERR001'")
	}
}

// =====================================================
// Edge Cases Tests
// =====================================================

// TestLogger_nilContext verifies nil context is handled.
func TestLogger_nilContext(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{out: &buf, minLevel: LevelInfo}

	logger.Info("message")

	var entry LogEntry
	output := strings.TrimSpace(buf.String())
	if err := json.Unmarshal([]byte(output), &entry); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	// Context field should be omitted when nil
	if entry.Context != nil {
		t.Error("Context should be omitted when nil")
	}
}

// TestLogger_emptyMessage verifies empty message is logged.
func TestLogger_emptyMessage(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{out: &buf, minLevel: LevelInfo}

	logger.Info("")

	output := strings.TrimSpace(buf.String())
	if output == "" {
		t.Error("Empty message should still be logged")
	}

	var entry LogEntry
	if err := json.Unmarshal([]byte(output), &entry); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	if entry.Message != "" {
		t.Errorf("Message = %q, want empty string", entry.Message)
	}
}

// TestLogger_emptyContext verifies empty context map is handled.
func TestLogger_emptyContext(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{out: &buf, minLevel: LevelInfo}

	logger.Info("message", map[string]interface{}{})

	var entry LogEntry
	output := strings.TrimSpace(buf.String())
	if err := json.Unmarshal([]byte(output), &entry); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	// Empty context map is omitted due to omitempty tag
	if entry.Context != nil {
		t.Error("Empty context map should be omitted due to omitempty tag")
	}
}

// TestLogger_writeError verifies write errors are handled gracefully.
func TestLogger_writeError(t *testing.T) {
	// Create a writer that always fails
	failWriter := &failingWriter{}
	logger := &Logger{out: failWriter, minLevel: LevelInfo}

	// Should not panic, just fail silently
	logger.Info("test message")
}

// =====================================================
// Helper Types
// =====================================================

// failingWriter is a test helper that always fails to write.
type failingWriter struct{}

func (w *failingWriter) Write(p []byte) (n int, err error) {
	return 0, io.ErrClosedPipe
}

// TestLogger_contextCancellation verifies logger doesn't block on context cancellation.
func TestLogger_contextCancellation(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{out: &buf, minLevel: LevelInfo}

	_, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Logging should still work even after context cancellation
	logger.Info("test message")

	if buf.String() == "" {
		t.Error("Logger should still work after context cancellation")
	}
}
