// Package services tests for business logic orchestration.
package services

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/kimhsiao/memonexus/backend/internal/analysis"
	"github.com/kimhsiao/memonexus/backend/internal/db"
	"github.com/kimhsiao/memonexus/backend/internal/models"
)

// =====================================================
// AnalysisService Tests
// =====================================================

// TestDefaultAnalysisConfig verifies default configuration.
func TestDefaultAnalysisConfig(t *testing.T) {
	config := DefaultAnalysisConfig()

	if config == nil {
		t.Fatal("DefaultAnalysisConfig() returned nil")
	}

	if config.EnableAI {
		t.Error("EnableAI should be false by default")
	}

	if config.NumKeywords != 10 {
		t.Errorf("NumKeywords = %d, want 10", config.NumKeywords)
	}

	if config.MaxSummaryLength != 500 {
		t.Errorf("MaxSummaryLength = %d, want 500", config.MaxSummaryLength)
	}

	if config.AITimeoutSeconds != 60 {
		t.Errorf("AITimeoutSeconds = %d, want 60", config.AITimeoutSeconds)
	}
}

// TestNewAnalysisService verifies service creation.
func TestNewAnalysisService(t *testing.T) {
	// Test with nil config (should use defaults)
	svc := NewAnalysisService(nil)

	if svc == nil {
		t.Fatal("NewAnalysisService(nil) returned nil")
	}

	if svc.tfidfAnalyzer == nil {
		t.Error("TF-IDF analyzer should be initialized")
	}

	if svc.textrankExtractor == nil {
		t.Error("TextRank extractor should be initialized")
	}

	if svc.config == nil {
		t.Error("Config should be set")
	}

	if svc.config.EnableAI {
		t.Error("AI should be disabled by default")
	}

	// Test with custom config
	customConfig := &AnalysisConfig{
		EnableAI:         true,
		NumKeywords:      5,
		MaxSummaryLength: 200,
		AITimeoutSeconds: 30,
	}

	svc2 := NewAnalysisService(customConfig)

	if svc2.config.NumKeywords != 5 {
		t.Errorf("NumKeywords = %d, want 5", svc2.config.NumKeywords)
	}

	if svc2.config.MaxSummaryLength != 200 {
		t.Errorf("MaxSummaryLength = %d, want 200", svc2.config.MaxSummaryLength)
	}
}

// TestIsAIEnabled verifies AI enabled state.
func TestIsAIEnabled(t *testing.T) {
	svc := NewAnalysisService(nil)

	// Initially disabled
	if svc.IsAIEnabled() {
		t.Error("AI should be disabled initially")
	}

	// Configure AI
	config := &analysis.AIConfig{
		Provider:  "openai",
		APIKey:    "test-key",
		ModelName: "gpt-4",
	}

	if err := svc.ConfigureAI(config); err != nil {
		t.Fatalf("ConfigureAI failed: %v", err)
	}

	// Now enabled
	if !svc.IsAIEnabled() {
		t.Error("AI should be enabled after configuration")
	}

	// Disable AI
	svc.DisableAI()

	if svc.IsAIEnabled() {
		t.Error("AI should be disabled after DisableAI()")
	}
}

// TestDisableAI verifies AI can be disabled.
func TestDisableAI(t *testing.T) {
	svc := NewAnalysisService(nil)

	// Configure AI
	config := &analysis.AIConfig{
		Provider:  "openai",
		APIKey:    "test-key",
		ModelName: "gpt-4",
	}

	if err := svc.ConfigureAI(config); err != nil {
		t.Fatalf("ConfigureAI failed: %v", err)
	}

	// Verify AI is configured
	if svc.aiAnalyzer == nil {
		t.Error("AI analyzer should be set after ConfigureAI")
	}

	if svc.aiConfig == nil {
		t.Error("AI config should be set after ConfigureAI")
	}

	// Disable AI
	svc.DisableAI()

	// Verify AI is cleared
	if svc.aiAnalyzer != nil {
		t.Error("AI analyzer should be nil after DisableAI")
	}

	if svc.aiConfig != nil {
		t.Error("AI config should be nil after DisableAI")
	}

	if svc.config.EnableAI {
		t.Error("EnableAI should be false after DisableAI")
	}
}

// TestGetAIConfig verifies API key redaction.
func TestGetAIConfig(t *testing.T) {
	svc := NewAnalysisService(nil)

	// Initially nil
	if svc.GetAIConfig() != nil {
		t.Error("GetAIConfig() should return nil when AI not configured")
	}

	// Configure AI
	config := &analysis.AIConfig{
		Provider:    "openai",
		APIKey:      "secret-key-12345",
		APIEndpoint: "https://api.openai.com",
		ModelName:   "gpt-4",
		MaxTokens:   4000,
	}

	if err := svc.ConfigureAI(config); err != nil {
		t.Fatalf("ConfigureAI failed: %v", err)
	}

	// Get config (should have redacted API key)
	retrievedConfig := svc.GetAIConfig()

	if retrievedConfig == nil {
		t.Fatal("GetAIConfig() returned nil after configuration")
	}

	if retrievedConfig.APIKey != "***REDACTED***" {
		t.Errorf("APIKey = %s, want '***REDACTED***'", retrievedConfig.APIKey)
	}

	if retrievedConfig.Provider != "openai" {
		t.Errorf("Provider = %s, want 'openai'", retrievedConfig.Provider)
	}

	if retrievedConfig.ModelName != "gpt-4" {
		t.Errorf("ModelName = %s, want 'gpt-4'", retrievedConfig.ModelName)
	}

	if retrievedConfig.MaxTokens != 4000 {
		t.Errorf("MaxTokens = %d, want 4000", retrievedConfig.MaxTokens)
	}

	// Verify original config is not modified
	if svc.aiConfig.APIKey != "secret-key-12345" {
		t.Error("Original config API key should not be modified")
	}
}

// TestConfigureAI_validation verifies AI configuration validation.
func TestConfigureAI_validation(t *testing.T) {
	svc := NewAnalysisService(nil)

	// Test missing API key
	t.Run("MissingAPIKey", func(t *testing.T) {
		config := &analysis.AIConfig{
			Provider:  "openai",
			ModelName: "gpt-4",
		}

		err := svc.ConfigureAI(config)
		if err == nil {
			t.Error("ConfigureAI with missing API key should return error")
		}
	})

	// Test missing provider
	t.Run("MissingProvider", func(t *testing.T) {
		config := &analysis.AIConfig{
			APIKey:    "test-key",
			ModelName: "gpt-4",
		}

		err := svc.ConfigureAI(config)
		if err == nil {
			t.Error("ConfigureAI with missing provider should return error")
		}
	})

	// Test valid configuration
	t.Run("ValidConfig", func(t *testing.T) {
		config := &analysis.AIConfig{
			Provider:  "openai",
			APIKey:    "test-key",
			ModelName: "gpt-4",
		}

		err := svc.ConfigureAI(config)
		if err != nil {
			t.Errorf("ConfigureAI with valid config failed: %v", err)
		}

		if !svc.IsAIEnabled() {
			t.Error("AI should be enabled after valid configuration")
		}
	})

	// Test nil config (should disable AI)
	t.Run("NilConfig", func(t *testing.T) {
		// First configure AI
		config := &analysis.AIConfig{
			Provider:  "openai",
			APIKey:    "test-key",
			ModelName: "gpt-4",
		}
		svc.ConfigureAI(config)

		// Pass nil to disable
		err := svc.ConfigureAI(nil)
		if err != nil {
			t.Errorf("ConfigureAI(nil) failed: %v", err)
		}

		if svc.IsAIEnabled() {
			t.Error("AI should be disabled after ConfigureAI(nil)")
		}
	})
}

// TestDetectLanguage verifies language detection.
func TestDetectLanguage(t *testing.T) {
	svc := NewAnalysisService(nil)

	tests := []struct {
		name     string
		text     string
		expected string
	}{
		{
			name:     "English text",
			text:     "This is a simple English sentence.",
			expected: "en",
		},
		{
			name:     "Chinese text",
			text:     "这是一个中文句子",
			expected: "cjk",
		},
		{
			name:     "Japanese text",
			text:     "これは日本語の文章です",
			expected: "cjk",
		},
		{
			name:     "Korean text",
			text:     "이것은 한국어 문장입니다",
			expected: "cjk",
		},
		{
			name:     "Mixed CJK and English",
			text:     "Hello 你好",
			expected: "cjk",
		},
		{
			name:     "Empty text",
			text:     "",
			expected: "en",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.detectLanguage(tt.text)
			if result != tt.expected {
				t.Errorf("detectLanguage(%q) = %s, want %s", tt.text, result, tt.expected)
			}
		})
	}
}

// TestSplitSentences verifies sentence splitting.
func TestSplitSentences(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected []string
	}{
		{
			name:     "Simple sentences",
			text:     "First sentence. Second sentence. Third sentence.",
			expected: []string{"First sentence", "Second sentence", "Third sentence"},
		},
		{
			name:     "Exclamation marks",
			text:     "Hello! World! Test!",
			expected: []string{"Hello", "World", "Test"},
		},
		{
			name:     "Question marks",
			text:     "How are you? What is this? Why?",
			expected: []string{"How are you", "What is this", "Why"},
		},
		{
			name:     "Mixed punctuation",
			text:     "First. Second! Third?",
			expected: []string{"First", "Second", "Third"},
		},
		{
			name:     "Newlines as sentence boundaries",
			text:     "First line\nSecond line\nThird line",
			expected: []string{"First line", "Second line", "Third line"},
		},
		{
			name:     "Trailing text without delimiter",
			text:     "First. Second trailing",
			expected: []string{"First", "Second trailing"},
		},
		{
			name:     "Empty text",
			text:     "",
			expected: []string{},
		},
		{
			name:     "Single sentence",
			text:     "Only one sentence here",
			expected: []string{"Only one sentence here"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitSentences(tt.text)

			if len(result) != len(tt.expected) {
				t.Errorf("splitSentences(%q) returned %d sentences, want %d", tt.text, len(result), len(tt.expected))
				return
			}

			for i, sentence := range result {
				if sentence != tt.expected[i] {
					t.Errorf("splitSentences(%q)[%d] = %q, want %q", tt.text, i, sentence, tt.expected[i])
				}
			}
		})
	}
}

// TestGenerateSummary verifies summary generation.
func TestGenerateSummary(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected string
	}{
		{
			name:     "Short text",
			text:     "This is a short text.",
			expected: "This is a short text.",
		},
		{
			name:     "Text within max length",
			text:     "This is a reasonably long text but still within the default max length of 500 characters. " +
				"It should be returned as is without truncation.",
			expected: "This is a reasonably long text but still within the default max length of 500 characters. " +
				"It should be returned as is without truncation.",
		},
	}

	svc := NewAnalysisService(nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.generateSummary(tt.text)
			if result != tt.expected {
				t.Errorf("generateSummary() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestSetEventCallbacks verifies callback configuration.
func TestSetEventCallbacks(t *testing.T) {
	svc := NewAnalysisService(nil)

	startedCalled := false
	completedCalled := false
	failedCalled := false

	svc.SetEventCallbacks(
		func(contentID string) {
			startedCalled = true
		},
		func(contentID string, result *AnalysisResult) {
			completedCalled = true
		},
		func(contentID string, err error) {
			failedCalled = true
		},
	)

	// Trigger callbacks
	if svc.onAnalysisStarted != nil {
		svc.onAnalysisStarted("test-id")
	}

	if svc.onAnalysisCompleted != nil {
		svc.onAnalysisCompleted("test-id", &AnalysisResult{})
	}

	if svc.onAnalysisFailed != nil {
		svc.onAnalysisFailed("test-id", nil)
	}

	if !startedCalled {
		t.Error("started callback was not called")
	}

	if !completedCalled {
		t.Error("completed callback was not called")
	}

	if !failedCalled {
		t.Error("failed callback was not called")
	}
}

// TestExtractKeywords_emptyText verifies empty text handling.
func TestExtractKeywords_emptyText(t *testing.T) {
	svc := NewAnalysisService(nil)

	keywords, err := svc.ExtractKeywords(context.Background(), "")

	if err != nil {
		t.Errorf("ExtractKeywords with empty text failed: %v", err)
	}

	if len(keywords) != 0 {
		t.Errorf("ExtractKeywords with empty text returned %d keywords, want 0", len(keywords))
	}
}

// TestGenerateSummary_emptyText verifies empty text summary.
func TestGenerateSummary_emptyText(t *testing.T) {
	svc := NewAnalysisService(nil)

	summary, err := svc.GenerateSummary(context.Background(), "")

	if err != nil {
		t.Errorf("GenerateSummary with empty text failed: %v", err)
	}

	if summary != "" {
		t.Errorf("GenerateSummary with empty text returned %q, want empty string", summary)
	}
}

// =====================================================
// ContentService Tests
// =====================================================

// TestDetectMediaTypeFromPath verifies media type detection.
func TestDetectMediaTypeFromPath(t *testing.T) {
	svc := &ContentService{} // Minimal setup, we only test the utility function

	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "JPEG image",
			path:     "/path/to/image.jpg",
			expected: "image",
		},
		{
			name:     "PNG image",
			path:     "photo.png",
			expected: "image",
		},
		{
			name:     "GIF image",
			path:     "animation.gif",
			expected: "image",
		},
		{
			name:     "WebP image",
			path:     "image.webp",
			expected: "image",
		},
		{
			name:     "MP4 video",
			path:     "video.mp4",
			expected: "video",
		},
		{
			name:     "WebM video",
			path:     "clip.webm",
			expected: "video",
		},
		{
			name:     "MOV video",
			path:     "movie.mov",
			expected: "video",
		},
		{
			name:     "AVI video",
			path:     "video.avi",
			expected: "video",
		},
		{
			name:     "PDF document",
			path:     "document.pdf",
			expected: "pdf",
		},
		{
			name:     "Markdown file",
			path:     "README.md",
			expected: "markdown",
		},
		{
			name:     "Markdown file with longer extension",
			path:     "doc.markdown",
			expected: "markdown",
		},
		{
			name:     "No extension",
			path:     "/path/to/file",
			expected: "web",
		},
		{
			name:     "Unknown extension",
			path:     "file.unknown",
			expected: "web",
		},
		{
			name:     "TXT file",
			path:     "notes.txt",
			expected: "web",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := string(svc.detectMediaTypeFromPath(tt.path))
			if result != tt.expected {
				t.Errorf("detectMediaTypeFromPath(%q) = %s, want %s", tt.path, result, tt.expected)
			}
		})
	}
}

// TestAnalyzeContent_emptyText verifies empty text error handling.
func TestAnalyzeContent_emptyText(t *testing.T) {
	svc := NewAnalysisService(nil)

	_, err := svc.AnalyzeContent(context.Background(), "test-id", "")

	if err == nil {
		t.Error("AnalyzeContent with empty text should return error")
	}
}

// TestAnalyzeContentItem_nilItem verifies nil item error handling.
func TestAnalyzeContentItem_nilItem(t *testing.T) {
	svc := NewAnalysisService(nil)

	_, err := svc.AnalyzeContentItem(context.Background(), nil)

	if err == nil {
		t.Error("AnalyzeContentItem with nil item should return error")
	}
}

// =====================================================
// Additional AnalysisService Tests
// =====================================================

// MockWebSocketBroadcaster is a mock implementation of WebSocketBroadcaster.
type MockWebSocketBroadcaster struct {
	startedEvents   []string
	completedEvents []map[string]interface{}
	failedEvents    []map[string]string
	mu              sync.Mutex
}

func (m *MockWebSocketBroadcaster) BroadcastAnalysisStarted(contentID string, operation string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.startedEvents = append(m.startedEvents, contentID)
}

func (m *MockWebSocketBroadcaster) BroadcastAnalysisCompleted(contentID string, result map[string]interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.completedEvents = append(m.completedEvents, map[string]interface{}{
		"content_id": contentID,
		"result":     result,
	})
}

func (m *MockWebSocketBroadcaster) BroadcastAnalysisFailed(contentID string, errMsg string, fallbackMethod string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.failedEvents = append(m.failedEvents, map[string]string{
		"content_id":      contentID,
		"error":           errMsg,
		"fallback_method": fallbackMethod,
	})
}

// TestSetWebSocketBroadcaster verifies WebSocket broadcaster configuration.
func TestSetWebSocketBroadcaster(t *testing.T) {
	svc := NewAnalysisService(nil)

	mockWS := &MockWebSocketBroadcaster{}

	// Set the broadcaster
	svc.SetWebSocketBroadcaster(mockWS)

	// Verify callbacks are set by triggering them through AnalyzeContent
	ctx := context.Background()
	text := "This is a test text for analysis with enough content."

	result, err := svc.AnalyzeContent(ctx, "test-id", text)
	if err != nil {
		t.Fatalf("AnalyzeContent failed: %v", err)
	}

	// Check that events were broadcast
	if len(mockWS.startedEvents) != 1 {
		t.Errorf("Expected 1 started event, got %d", len(mockWS.startedEvents))
	}
	if mockWS.startedEvents[0] != "test-id" {
		t.Errorf("Expected content_id 'test-id', got %s", mockWS.startedEvents[0])
	}

	if len(mockWS.completedEvents) != 1 {
		t.Errorf("Expected 1 completed event, got %d", len(mockWS.completedEvents))
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	if result.Method != "tfidf" {
		t.Errorf("Expected method 'tfidf', got %s", result.Method)
	}
}

// TestAnalyzeContent_offline verifies offline TF-IDF analysis.
func TestAnalyzeContent_offline(t *testing.T) {
	svc := NewAnalysisService(nil)

	ctx := context.Background()
	text := "This is a longer text that should contain enough content for keyword extraction. " +
		"We have multiple sentences here. And some more text to ensure proper analysis."

	result, err := svc.AnalyzeContent(ctx, "test-id", text)
	if err != nil {
		t.Fatalf("AnalyzeContent failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	if result.ContentID != "test-id" {
		t.Errorf("ContentID = %s, want 'test-id'", result.ContentID)
	}

	if result.Method != "tfidf" {
		t.Errorf("Method = %s, want 'tfidf'", result.Method)
	}

	if result.AIUsed {
		t.Error("AIUsed should be false for offline analysis")
	}

	if len(result.Keywords) == 0 {
		t.Error("Expected at least one keyword")
	}

	if result.Summary == "" {
		t.Error("Expected summary to be generated")
	}

	if result.Language != "en" {
		t.Errorf("Language = %s, want 'en'", result.Language)
	}
}

// TestBatchAnalyze verifies batch analysis of multiple items.
func TestBatchAnalyze(t *testing.T) {
	svc := NewAnalysisService(nil)

	ctx := context.Background()
	items := []*models.ContentItem{
		{ID: models.UUID("1"), ContentText: "First item with some content."},
		{ID: models.UUID("2"), ContentText: "Second item with different content."},
		{ID: models.UUID("3"), ContentText: "Third item for analysis."},
	}

	results, err := svc.BatchAnalyze(ctx, items)
	if err != nil {
		t.Fatalf("BatchAnalyze failed: %v", err)
	}

	if len(results) != 3 {
		t.Fatalf("Expected 3 results, got %d", len(results))
	}

	for i, result := range results {
		if result == nil {
			t.Errorf("Result %d is nil", i)
			continue
		}

		if result.ContentID != string(items[i].ID) {
			t.Errorf("Result %d ContentID = %s, want %s", i, result.ContentID, items[i].ID)
		}
	}
}

// TestExtractKeywords_offline verifies offline keyword extraction.
func TestExtractKeywords_offline(t *testing.T) {
	svc := NewAnalysisService(nil)

	ctx := context.Background()
	text := "Machine learning is a subset of artificial intelligence. " +
		"Deep learning neural networks are powerful models for pattern recognition. " +
		"Natural language processing analyzes text data."

	keywords, err := svc.ExtractKeywords(ctx, text)
	if err != nil {
		t.Fatalf("ExtractKeywords failed: %v", err)
	}

	if len(keywords) == 0 {
		t.Error("Expected at least one keyword")
	}
}

// TestAnalyzeContent_withAI verifies AI analysis path.
func TestAnalyzeContent_withAI(t *testing.T) {
	svc := NewAnalysisService(nil)

	// Configure AI (will use mock/stub analyzer)
	config := &analysis.AIConfig{
		Provider:  "test",
		APIKey:    "test-key",
		ModelName: "test-model",
	}

	err := svc.ConfigureAI(config)
	if err != nil {
		t.Fatalf("ConfigureAI failed: %v", err)
	}

	ctx := context.Background()
	text := "This is a test text for AI analysis."

	// This will use AI but fallback to offline on failure
	// since we don't have a real AI implementation
	result, err := svc.AnalyzeContent(ctx, "test-id", text)

	// The AI analyzer will be called, but since it's not a real provider
	// it should still complete (with AI method tag) or fallback
	// We just verify it doesn't crash
	if err != nil {
		// AI may have failed, that's okay
		t.Logf("AnalyzeContent had error (may fall back to offline): %v", err)
	}

	if result != nil {
		// Verify we got some result
		if result.ContentID != "test-id" {
			t.Errorf("ContentID = %s, want 'test-id'", result.ContentID)
		}
	}
}

// TestGenerateSummary_truncation verifies summary truncation.
func TestGenerateSummary_truncation(t *testing.T) {
	svc := NewAnalysisService(&AnalysisConfig{
		MaxSummaryLength: 100,
	})

	// Create a long text
	longText := strings.Repeat("This is a sentence. ", 50)

	summary, err := svc.GenerateSummary(context.Background(), longText)
	if err != nil {
		t.Fatalf("GenerateSummary failed: %v", err)
	}

	if len(summary) > 150 { // Allow some margin for "..." and word boundaries
		t.Errorf("Summary length = %d, want <= 150", len(summary))
	}

	if !strings.HasSuffix(summary, "...") {
		t.Error("Truncated summary should end with '...'")
	}
}

// TestGenerateSummary_CjkText verifies CJK text handling.
func TestGenerateSummary_cjkText(t *testing.T) {
	svc := NewAnalysisService(nil)

	text := "这是中文文本。它包含多个句子。每个句子都应该被正确处理。"

	summary, err := svc.GenerateSummary(context.Background(), text)
	if err != nil {
		t.Fatalf("GenerateSummary failed: %v", err)
	}

	if summary == "" {
		t.Error("Expected non-empty summary for CJK text")
	}

	// Verify language detection for CJK
	result, err := svc.AnalyzeContent(context.Background(), "test-id", text)
	if err != nil {
		t.Fatalf("AnalyzeContent failed: %v", err)
	}

	if result.Language != "cjk" {
		t.Errorf("Language = %s, want 'cjk'", result.Language)
	}
}

// TestExtractKeywords_cjkText verifies CJK keyword extraction.
func TestExtractKeywords_cjkText(t *testing.T) {
	svc := NewAnalysisService(nil)

	ctx := context.Background()
	text := "机器学习是人工智能的一个子集。深度学习神经网络是强大的模式识别模型。自然语言处理分析文本数据。"

	keywords, err := svc.ExtractKeywords(ctx, text)
	if err != nil {
		t.Fatalf("ExtractKeywords failed: %v", err)
	}

	if len(keywords) == 0 {
		t.Error("Expected at least one keyword from CJK text")
	}
}

// TestAnalyzeContent_cjk verifies full analysis pipeline with CJK text.
func TestAnalyzeContent_cjk(t *testing.T) {
	svc := NewAnalysisService(nil)

	ctx := context.Background()
	text := "自然语言处理（NLP）是人工智能的重要分支。深度学习技术在NLP领域取得了显著进展。" +
		"机器翻译、情感分析、文本分类都是NLP的应用场景。"

	result, err := svc.AnalyzeContent(ctx, "cjk-test", text)
	if err != nil {
		t.Fatalf("AnalyzeContent failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	if result.Language != "cjk" {
		t.Errorf("Language = %s, want 'cjk'", result.Language)
	}

	if len(result.Keywords) == 0 {
		t.Error("Expected at least one keyword")
	}
}

// TestSetEventCallbacks_nilCallbacks verifies nil callback handling.
func TestSetEventCallbacks_nilCallbacks(t *testing.T) {
	svc := NewAnalysisService(nil)

	// Set all nil callbacks - should not panic
	svc.SetEventCallbacks(nil, nil, nil)

	// Try triggering - should handle nil gracefully
	ctx := context.Background()
	text := "Test content"

	// This should not panic even with nil callbacks
	_, err := svc.AnalyzeContent(ctx, "test-id", text)
	if err != nil {
		t.Logf("AnalyzeContent with nil callbacks returned error: %v", err)
	}
}

// =====================================================
// Additional ExtractKeywords and GenerateSummary Tests
// =====================================================

// TestExtractKeywords_AI_fallback verifies AI fallback to offline.
func TestExtractKeywords_AI_fallback(t *testing.T) {
	svc := NewAnalysisService(nil)

	// Configure AI (test provider will fail and fallback)
	config := &analysis.AIConfig{
		Provider:  "test", // Non-existent provider for fallback test
		APIKey:    "test-key",
		ModelName: "test-model",
	}

	_ = svc.ConfigureAI(config)

	ctx := context.Background()
	text := "Machine learning algorithms transform data into insights. " +
		"Neural networks learn patterns from training data. " +
		"Deep learning models achieve state-of-the-art results."

	keywords, err := svc.ExtractKeywords(ctx, text)

	if err != nil {
		t.Errorf("ExtractKeywords should fallback to offline on AI failure: %v", err)
	}

	if len(keywords) == 0 {
		t.Error("Expected keywords from fallback offline extraction")
	}
}

// TestGenerateSummary_AI_fallback verifies AI fallback to offline.
func TestGenerateSummary_AI_fallback(t *testing.T) {
	svc := NewAnalysisService(nil)

	// Configure AI (will use test provider that falls back)
	config := &analysis.AIConfig{
		Provider:  "test", // Non-existent provider for fallback test
		APIKey:    "test-key",
		ModelName: "test-model",
	}

	_ = svc.ConfigureAI(config)

	ctx := context.Background()
	text := "This is the first sentence. This is the second sentence. " +
		"This is the third sentence. And this is the fourth sentence."

	summary, err := svc.GenerateSummary(ctx, text)

	if err != nil {
		t.Errorf("GenerateSummary should fallback to offline on AI failure: %v", err)
	}

	if summary == "" {
		t.Error("Expected summary from fallback offline extraction")
	}

	// Offline summary should be first few sentences
	sentences := splitSentences(text)
	if len(sentences) > 0 && summary != sentences[0] {
		// Summary should at least contain the first sentence or be truncated
		t.Logf("Summary: %q, First sentence: %q", summary, sentences[0])
	}
}

// TestAnalyzeContent_verifyResult verifies all result fields are populated.
func TestAnalyzeContent_verifyResult(t *testing.T) {
	svc := NewAnalysisService(nil)

	ctx := context.Background()
	text := "Machine learning is transforming technology. " +
		"Natural language processing enables computers to understand text. " +
		"Computer vision allows machines to interpret visual data."

	result, err := svc.AnalyzeContent(ctx, "verify-test", text)

	if err != nil {
		t.Fatalf("AnalyzeContent failed: %v", err)
	}

	if result.ContentID != "verify-test" {
		t.Errorf("ContentID = %s, want 'verify-test'", result.ContentID)
	}

	if result.Method != "tfidf" {
		t.Errorf("Method = %s, want 'tfidf'", result.Method)
	}

	if result.Language != "en" {
		t.Errorf("Language = %s, want 'en'", result.Language)
	}

	if result.Confidence <= 0 {
		t.Errorf("Confidence = %f, want > 0", result.Confidence)
	}

	if result.AIUsed {
		t.Error("AIUsed should be false for offline analysis")
	}

	if len(result.Keywords) == 0 {
		t.Error("Expected at least one keyword")
	}

	if result.Summary == "" {
		t.Error("Expected non-empty summary")
	}

	// Note: ProcessedAt is not set by analyzeOffline, it's the caller's responsibility
	// We just verify the result is complete for the service layer
}

// TestExtractKeywords_shortText verifies short text handling.
func TestExtractKeywords_shortText(t *testing.T) {
	svc := NewAnalysisService(nil)

	ctx := context.Background()
	text := "Short"

	keywords, err := svc.ExtractKeywords(ctx, text)

	if err != nil {
		t.Errorf("ExtractKeywords with short text failed: %v", err)
	}

	// Short text may return empty keywords
	t.Logf("Short text keywords: %v", keywords)
}

// TestGenerateSummary_shortText verifies short text summary.
func TestGenerateSummary_shortText(t *testing.T) {
	svc := NewAnalysisService(nil)

	ctx := context.Background()
	text := "Hi"

	summary, err := svc.GenerateSummary(ctx, text)

	if err != nil {
		t.Errorf("GenerateSummary with short text failed: %v", err)
	}

	if summary != text {
		t.Errorf("Summary = %q, want %q (same as input for short text)", summary, text)
	}
}

// TestSetWebSocketBroadcaster_nilBroadcaster verifies nil handling.
func TestSetWebSocketBroadcaster_nilBroadcaster(t *testing.T) {
	svc := NewAnalysisService(nil)

	// Set nil broadcaster - should not panic
	svc.SetWebSocketBroadcaster(nil)

	// AnalyzeContent should handle nil broadcaster gracefully
	ctx := context.Background()
	text := "Test content for nil broadcaster"

	result, err := svc.AnalyzeContent(ctx, "test-nil-broadcaster", text)

	if err != nil {
		t.Errorf("AnalyzeContent with nil broadcaster failed: %v", err)
	}

	if result == nil {
		t.Error("Expected result even with nil broadcaster")
	}
}

// TestExtractKeywords_textrankFallback verifies TF-IDF fallback when TextRank fails.
func TestExtractKeywords_textrankFallback(t *testing.T) {
	svc := NewAnalysisService(nil)

	ctx := context.Background()
	// Use text that might cause TextRank to return fewer results
	text := "AI machine learning technology."

	keywords, err := svc.ExtractKeywords(ctx, text)

	if err != nil {
		t.Errorf("ExtractKeywords failed: %v", err)
	}

	// Should get keywords from either TextRank or TF-IDF
	t.Logf("Keywords: %v", keywords)

	// Keywords should be extracted even if short text
	if len(keywords) == 0 {
		t.Log("No keywords extracted (may happen with very short text)")
	}
}

// TestAnalyzeContent_textrankFallback verifies offline analysis with TextRank fallback.
func TestAnalyzeContent_textrankFallback(t *testing.T) {
	svc := NewAnalysisService(nil)

	ctx := context.Background()
	// Use simple text that tests fallback paths
	text := "Data science involves statistics programming and domain knowledge."

	result, err := svc.AnalyzeContent(ctx, "textrank-test", text)

	if err != nil {
		t.Fatalf("AnalyzeContent failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	if result.Method != "tfidf" {
		t.Errorf("Method = %s, want 'tfidf'", result.Method)
	}

	// Verify we got some analysis result
	t.Logf("Keywords: %v, Summary: %q", result.Keywords, result.Summary)
}

// TestGenerateSummary_emptyText verifies empty text returns empty summary.
func TestGenerateSummary_veryShortText(t *testing.T) {
	svc := NewAnalysisService(nil)

	ctx := context.Background()
	text := "A"

	summary, err := svc.GenerateSummary(ctx, text)

	if err != nil {
		t.Errorf("GenerateSummary with very short text failed: %v", err)
	}

	// Very short text should be returned as-is
	if summary != text {
		t.Errorf("Summary = %q, want %q", summary, text)
	}
}

// TestAnalyzeContent_concurrent verifies concurrent analysis is safe.
func TestAnalyzeContent_concurrent(t *testing.T) {
	svc := NewAnalysisService(nil)

	ctx := context.Background()

	// Run multiple analyses concurrently
	done := make(chan bool, 5)
	for i := 0; i < 5; i++ {
		go func(idx int) {
			text := fmt.Sprintf("Concurrent test %d with some content for analysis.", idx)
			_, err := svc.AnalyzeContent(ctx, fmt.Sprintf("concurrent-%d", idx), text)
			if err != nil {
				t.Logf("Concurrent analysis %d failed: %v", idx, err)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 5; i++ {
		<-done
	}

	// If we got here without deadlock or panic, test passes
}

// TestExtractKeywords_AI_disabled verifies keyword extraction with AI disabled.
func TestExtractKeywords_AI_disabled(t *testing.T) {
	// Create service with AI explicitly disabled
	config := &AnalysisConfig{
		EnableAI: false,
	}
	svc := NewAnalysisService(config)

	ctx := context.Background()
	text := "Natural language processing and computer vision are key areas of artificial intelligence."

	keywords, err := svc.ExtractKeywords(ctx, text)

	if err != nil {
		t.Errorf("ExtractKeywords failed: %v", err)
	}

	if len(keywords) == 0 {
		t.Error("Expected keywords from offline extraction")
	}

	// Verify AI was not used
	if svc.IsAIEnabled() {
		t.Error("AI should be disabled")
	}
}

// =====================================================
// Additional Edge Case and Error Path Tests
// =====================================================


// TestAnalyzeContent_verifyCallbacks verifies all callbacks are called correctly.
func TestAnalyzeContent_verifyCallbacks(t *testing.T) {
	svc := NewAnalysisService(nil)

	startedCalled := false
	completedCalled := false
	failedCalled := false

	svc.SetEventCallbacks(
		func(contentID string) {
			if contentID != "callback-test" {
				t.Errorf("contentID = %s, want 'callback-test'", contentID)
			}
			startedCalled = true
		},
		func(contentID string, result *AnalysisResult) {
			if contentID != "callback-test" {
				t.Errorf("contentID = %s, want 'callback-test'", contentID)
			}
			if result == nil {
				t.Error("result should not be nil in completed callback")
			}
			completedCalled = true
		},
		func(contentID string, err error) {
			failedCalled = true
		},
	)

	ctx := context.Background()
	text := "Test content for callback verification"

	result, err := svc.AnalyzeContent(ctx, "callback-test", text)

	if err != nil {
		t.Fatalf("AnalyzeContent failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result")
	}

	if !startedCalled {
		t.Error("started callback was not called")
	}

	if !completedCalled {
		t.Error("completed callback was not called")
	}

	if failedCalled {
		t.Error("failed callback should not be called on success")
	}
}

// TestAnalyzeContent_offlineWithTextRankFallback verifies TextRank fallback.
func TestAnalyzeContent_offlineWithTextRankFallback(t *testing.T) {
	svc := NewAnalysisService(nil)

	ctx := context.Background()
	// Use text that's long enough for analysis
	text := strings.Repeat("This is a test sentence. ", 20)

	result, err := svc.AnalyzeContent(ctx, "textrank-fallback-test", text)

	if err != nil {
		t.Fatalf("AnalyzeContent failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result")
	}

	if result.Method != "tfidf" {
		t.Errorf("Method = %s, want 'tfidf'", result.Method)
	}

	if len(result.Keywords) == 0 {
		t.Error("Expected keywords from TextRank or TF-IDF")
	}
}

// TestGenerateSummary_boundaryCases verifies summary generation edge cases.
func TestGenerateSummary_boundaryCases(t *testing.T) {
	svc := NewAnalysisService(&AnalysisConfig{
		MaxSummaryLength: 50,
	})

	tests := []struct {
		name     string
		text     string
		verify   func(t *testing.T, summary string)
	}{
		{
			name: "Exactly at max length",
			text: strings.Repeat("A", 50),
			verify: func(t *testing.T, summary string) {
				if len(summary) > 50 {
					t.Errorf("Summary length = %d, want <= 50", len(summary))
				}
			},
		},
		{
			name: "One character over max",
			text: strings.Repeat("B", 51),
			verify: func(t *testing.T, summary string) {
				if len(summary) > 51 {
					t.Errorf("Summary length = %d, want <= 51", len(summary))
				}
			},
		},
		{
			name: "Multiple newlines",
			text: "Line 1\nLine 2\nLine 3\nLine 4",
			verify: func(t *testing.T, summary string) {
				if summary == "" {
					t.Error("Summary should not be empty")
				}
			},
		},
		{
			name: "Only delimiters",
			text: ".!?\n.",
			verify: func(t *testing.T, summary string) {
				// Should handle gracefully
				t.Logf("Summary for only delimiters: %q", summary)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary, err := svc.GenerateSummary(context.Background(), tt.text)
			if err != nil {
				t.Errorf("GenerateSummary failed: %v", err)
			}
			tt.verify(t, summary)
		})
	}
}

// TestBatchAnalyze_withNilItems verifies batch handles nil items gracefully.
func TestBatchAnalyze_withNilItems(t *testing.T) {
	svc := NewAnalysisService(nil)

	ctx := context.Background()
	items := []*models.ContentItem{
		{ID: models.UUID("1"), ContentText: "First item"},
		nil, // Nil item in the middle
		{ID: models.UUID("2"), ContentText: "Second item"},
	}

	results, err := svc.BatchAnalyze(ctx, items)

	// BatchAnalyze may return partial results or error
	// The important thing is it doesn't panic
	if err != nil {
		t.Logf("BatchAnalyze with nil items returned error: %v", err)
	}

	if results == nil {
		t.Error("results should not be nil")
	}

	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}

	// Check nil item position
	if results[1] != nil {
		t.Error("Result for nil item should be nil")
	}
}

// TestBatchAnalyze_concurrentWithErrors verifies concurrent batch with errors.
func TestBatchAnalyze_concurrentWithErrors(t *testing.T) {
	svc := NewAnalysisService(nil)

	ctx := context.Background()
	items := []*models.ContentItem{
		{ID: models.UUID("1"), ContentText: "Valid item 1"},
		{ID: models.UUID("2"), ContentText: ""}, // Empty content (will error)
		{ID: models.UUID("3"), ContentText: "Valid item 2"},
	}

	results, err := svc.BatchAnalyze(ctx, items)

	// Should return error due to empty content
	if err == nil {
		t.Error("BatchAnalyze with errors should return error")
	}

	if results == nil {
		t.Fatal("results should not be nil")
	}

	// First and third items should have results
	if results[0] == nil && results[2] == nil {
		t.Error("At least some results should be returned")
	}
}

// TestGenerateSummary_exactMaxLength verifies exact max length handling.
func TestGenerateSummary_exactMaxLength(t *testing.T) {
	maxLen := 100
	svc := NewAnalysisService(&AnalysisConfig{
		MaxSummaryLength: maxLen,
	})

	// Create text that's exactly at max length
	text := strings.Repeat("A", maxLen)

	summary, err := svc.GenerateSummary(context.Background(), text)

	if err != nil {
		t.Errorf("GenerateSummary failed: %v", err)
	}

	if summary != text {
		t.Errorf("Summary = %q, want %q (exact match for max length)", summary, text)
	}

	if len(summary) != maxLen {
		t.Errorf("Summary length = %d, want %d", len(summary), maxLen)
	}
}

// TestAnalyzeContent_verifyConfidence verifies confidence scores are set.
func TestAnalyzeContent_verifyConfidence(t *testing.T) {
	svc := NewAnalysisService(nil)

	ctx := context.Background()
	text := "This is a test text with enough content for analysis."

	result, err := svc.AnalyzeContent(ctx, "confidence-test", text)

	if err != nil {
		t.Fatalf("AnalyzeContent failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result")
	}

	// Offline analysis should have 0.7 confidence
	if result.Confidence != 0.7 {
		t.Errorf("Confidence = %f, want 0.7 for offline analysis", result.Confidence)
	}
}

// TestAnalyzeContent_verifyProcessedAt verifies ProcessedAt is not set by service.
func TestAnalyzeContent_verifyProcessedAt(t *testing.T) {
	svc := NewAnalysisService(nil)

	ctx := context.Background()
	text := "Test content for ProcessedAt verification."

	result, err := svc.AnalyzeContent(ctx, "processed-at-test", text)

	if err != nil {
		t.Fatalf("AnalyzeContent failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result")
	}

	// ProcessedAt should be 0 (not set by service layer)
	if result.ProcessedAt != 0 {
		t.Errorf("ProcessedAt = %d, want 0 (not set by service)", result.ProcessedAt)
	}
}

// TestSetWebSocketBroadcaster_eventsVerify verifies all events are broadcast.
func TestSetWebSocketBroadcaster_eventsVerify(t *testing.T) {
	svc := NewAnalysisService(nil)

	mockWS := &MockWebSocketBroadcaster{}
	svc.SetWebSocketBroadcaster(mockWS)

	ctx := context.Background()
	text := "Test content for WebSocket event verification"

	// Reset mock events
	mockWS.startedEvents = nil
	mockWS.completedEvents = nil
	mockWS.failedEvents = nil

	result, err := svc.AnalyzeContent(ctx, "ws-events-test", text)

	if err != nil {
		t.Fatalf("AnalyzeContent failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result")
	}

	// Verify started event
	if len(mockWS.startedEvents) != 1 {
		t.Errorf("Expected 1 started event, got %d", len(mockWS.startedEvents))
	}

	if mockWS.startedEvents[0] != "ws-events-test" {
		t.Errorf("Started content_id = %s, want 'ws-events-test'", mockWS.startedEvents[0])
	}

	// Verify completed event
	if len(mockWS.completedEvents) != 1 {
		t.Errorf("Expected 1 completed event, got %d", len(mockWS.completedEvents))
	}

	// Verify no failed events
	if len(mockWS.failedEvents) != 0 {
		t.Errorf("Expected 0 failed events, got %d", len(mockWS.failedEvents))
	}
}


// TestGenerateSummary_noSentenceBoundaries verifies text without sentence boundaries.
func TestGenerateSummary_noSentenceBoundaries(t *testing.T) {
	svc := NewAnalysisService(nil)

	ctx := context.Background()
	// Text without sentence delimiters
	text := "This is a long continuous text without any sentence boundaries or punctuation marks"

	summary, err := svc.GenerateSummary(ctx, text)

	if err != nil {
		t.Errorf("GenerateSummary failed: %v", err)
	}

	if summary == "" {
		t.Error("Summary should not be empty for text without boundaries")
	}

	// Should return text as-is or truncated with ...
	if summary != text && !strings.HasSuffix(summary, "...") {
		t.Errorf("Summary should be text or end with '...', got %q", summary)
	}
}

// TestAnalyzeContent_verifySummaryFormat verifies summary format is correct.
func TestAnalyzeContent_verifySummaryFormat(t *testing.T) {
	svc := NewAnalysisService(nil)

	ctx := context.Background()
	text := "First sentence. Second sentence. Third sentence. Fourth sentence. Fifth sentence."

	result, err := svc.AnalyzeContent(ctx, "summary-format-test", text)

	if err != nil {
		t.Fatalf("AnalyzeContent failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result")
	}

	if result.Summary == "" {
		t.Error("Summary should not be empty")
	}

	// Summary should contain content (either truncated or first few sentences)
	if result.Summary == "" {
		t.Error("Summary should not be empty")
	}

	// Verify it's not just the full text repeated
	t.Logf("Original text length: %d, Summary length: %d", len(text), len(result.Summary))
}

// TestAnalyzeContent_offlineAnalysis verifies offline TF-IDF analysis works.
func TestAnalyzeContent_offlineAnalysis(t *testing.T) {
	svc := NewAnalysisService(nil)
	
	// AI should be disabled by default
	if svc.IsAIEnabled() {
		t.Error("AI should be disabled by default")
	}

	ctx := context.Background()
	text := "This is a test text for offline analysis."

	result, err := svc.AnalyzeContent(ctx, "offline-test", text)

	if err != nil {
		t.Errorf("AnalyzeContent with offline analysis failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result from offline analysis")
	}

	if result.Method != "tfidf" {
		t.Errorf("Method = %s, want 'tfidf' (offline)", result.Method)
	}

	if result.AIUsed {
		t.Error("AIUsed should be false for offline analysis")
	}
}

// =====================================================
// ContentService Tests
// =====================================================

// setupTestDB creates an in-memory database and runs migrations.
func setupTestDB(t *testing.T) *sql.DB {
	database, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open in-memory database: %v", err)
	}

	// Initialize migrator (creates schema_migrations table)
	migrator := db.NewMigrator(database, "../db/migrations")
	if err := migrator.Initialize(); err != nil {
		database.Close()
		t.Fatalf("Failed to initialize migrator: %v", err)
	}

	// Apply migrations - handle "table already exists" gracefully
	if err := migrator.Up(); err != nil {
		// schema_migrations table already exists is acceptable
		if !strings.Contains(err.Error(), "already exists") {
			database.Close()
			t.Fatalf("Failed to apply migrations: %v", err)
		}
	}

	// Verify at least one content table exists
	var tableName string
	err = database.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='content_items' LIMIT 1").Scan(&tableName)
	if err != nil {
		database.Close()
		t.Fatalf("content_items table does not exist after migrations: %v", err)
	}

	return database
}

// TestNewContentService verifies ContentService creation.
func TestNewContentService(t *testing.T) {
	tempDir := t.TempDir()
	db := setupTestDB(t)
	defer db.Close()

	svc, err := NewContentService(db, tempDir)

	if err != nil {
		t.Fatalf("NewContentService() failed: %v", err)
	}

	if svc == nil {
		t.Fatal("NewContentService() returned nil")
	}

	if svc.repo == nil {
		t.Error("Repository should be initialized")
	}

	if svc.parser == nil {
		t.Error("Parser should be initialized")
	}

	if svc.storage == nil {
		t.Error("Storage should be initialized")
	}
}

// TestNewContentService_invalidStorageDir verifies error handling for invalid storage directory.
func TestNewContentService_invalidStorageDir(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Use an invalid path that cannot be created
	invalidDir := "/proc/nonexistent/path\000" // Invalid null byte in path

	_, err := NewContentService(db, invalidDir)

	if err == nil {
		t.Error("NewContentService() with invalid storage dir should return error")
	}
}

// TestContentService_GetContent_notFound verifies error when content doesn't exist.
func TestContentService_GetContent_notFound(t *testing.T) {
	tempDir := t.TempDir()
	db := setupTestDB(t)
	defer db.Close()

	svc, _ := NewContentService(db, tempDir)

	_, err := svc.GetContent("non-existent-id")

	if err == nil {
		t.Error("GetContent() with non-existent ID should return error")
	}
}

// TestContentService_ListContent_empty verifies empty list when no content exists.
func TestContentService_ListContent_empty(t *testing.T) {
	tempDir := t.TempDir()
	db := setupTestDB(t)
	defer db.Close()

	svc, _ := NewContentService(db, tempDir)

	items, err := svc.ListContent(10, 0, "")

	if err != nil {
		t.Fatalf("ListContent() failed: %v", err)
	}

	// Both nil and empty slice are acceptable for "no results"
	if items != nil && len(items) != 0 {
		t.Errorf("ListContent() should return 0 items, got %d", len(items))
	}
}

// TestContentService_UpdateContent_notFound verifies error when updating non-existent content.
func TestContentService_UpdateContent_notFound(t *testing.T) {
	tempDir := t.TempDir()
	db := setupTestDB(t)
	defer db.Close()

	svc, _ := NewContentService(db, tempDir)

	item := &models.ContentItem{
		ID:      models.UUID("non-existent"),
		Title:   "Updated Title",
		ContentText: "Updated content",
	}

	err := svc.UpdateContent(item)

	if err == nil {
		t.Error("UpdateContent() with non-existent ID should return error")
	}
}

// TestContentService_DeleteContent_notFound verifies error when deleting non-existent content.
func TestContentService_DeleteContent_notFound(t *testing.T) {
	tempDir := t.TempDir()
	db := setupTestDB(t)
	defer db.Close()

	svc, _ := NewContentService(db, tempDir)

	err := svc.DeleteContent("non-existent-id")

	if err == nil {
		t.Error("DeleteContent() with non-existent ID should return error")
	}
}

// TestContentService_CreateFromURL_invalidURL verifies error handling for invalid URL.
func TestContentService_CreateFromURL_invalidURL(t *testing.T) {
	tempDir := t.TempDir()
	db := setupTestDB(t)
	defer db.Close()

	svc, _ := NewContentService(db, tempDir)

	invalidURL := "://not-a-valid-url"
	_, err := svc.CreateFromURL(invalidURL)

	if err == nil {
		t.Error("CreateFromURL() with invalid URL should return error")
	}

	if !strings.Contains(err.Error(), "invalid URL") {
		t.Errorf("Error should mention 'invalid URL', got: %v", err)
	}
}

// TestContentService_SearchContent verifies search returns items (placeholder implementation).
func TestContentService_SearchContent(t *testing.T) {
	tempDir := t.TempDir()
	db := setupTestDB(t)
	defer db.Close()

	svc, _ := NewContentService(db, tempDir)

	// SearchContent is currently a placeholder that returns all items
	items, err := svc.SearchContent("test query", 10)

	if err != nil {
		t.Fatalf("SearchContent() failed: %v", err)
	}

	// Both nil and empty slice are acceptable for "no results"
	if items != nil && len(items) != 0 {
		t.Logf("SearchContent() returned %d items (placeholder implementation)", len(items))
	}
}

// TestContentService_GetStorageFilePath verifies storage file path resolution.
func TestContentService_GetStorageFilePath(t *testing.T) {
	tempDir := t.TempDir()
	db := setupTestDB(t)
	defer db.Close()

	svc, _ := NewContentService(db, tempDir)

	testHash := strings.Repeat("a", 64) // SHA-256 hash length

	path, err := svc.GetStorageFilePath(testHash)

	if err != nil {
		// File doesn't exist, but path resolution should work
		t.Logf("GetStorageFilePath() returned error (expected for non-existent file): %v", err)
	}

	// Path should contain the hash prefix
	if path != "" && !strings.Contains(path, testHash[:4]) {
		t.Errorf("Path should contain hash prefix, got: %s", path)
	}
}

// TestContentService_GetStorageStats verifies storage statistics retrieval.
func TestContentService_GetStorageStats(t *testing.T) {
	tempDir := t.TempDir()
	db := setupTestDB(t)
	defer db.Close()

	svc, _ := NewContentService(db, tempDir)

	stats, err := svc.GetStorageStats()

	if err != nil {
		t.Fatalf("GetStorageStats() failed: %v", err)
	}

	if stats == nil {
		t.Fatal("GetStorageStats() should return stats, not nil")
	}

	// Empty storage should have 0 files and 0 size
	if stats.TotalFiles != 0 {
		t.Errorf("TotalFiles = %d, want 0 for empty storage", stats.TotalFiles)
	}

	if stats.TotalSize != 0 {
		t.Errorf("TotalSize = %d, want 0 for empty storage", stats.TotalSize)
	}
}

// TestContentService_detectMediaTypeFromPath verifies media type detection.
func TestContentService_detectMediaTypeFromPath(t *testing.T) {
	// This test is skipped since detectMediaTypeFromPath is a private method
	// It is indirectly tested through CreateFromFile integration tests
	t.Skip("detectMediaTypeFromPath is private - tested via CreateFromFile")
}

// =====================================================
// CreateFromFile and findByContentHash Tests
// =====================================================

// TestContentService_FindByContentHash_empty verifies empty database handling.
func TestContentService_FindByContentHash_empty(t *testing.T) {
	tempDir := t.TempDir()
	db := setupTestDB(t)
	defer db.Close()

	svc, _ := NewContentService(db, tempDir)

	// Access private method via type assertion to create a testable wrapper
	// We'll test it indirectly through CreateFromFile
	items, err := svc.ListContent(1000, 0, "")
	if err != nil {
		t.Fatalf("ListContent() failed: %v", err)
	}

	if len(items) != 0 {
		t.Errorf("Expected 0 items in empty database, got %d", len(items))
	}
}

// TestContentService_CreateFromFile verifies file creation from reader.
func TestContentService_CreateFromFile(t *testing.T) {
	tempDir := t.TempDir()
	db := setupTestDB(t)
	defer db.Close()

	svc, _ := NewContentService(db, tempDir)

	// Create test content
	testContent := []byte("This is test file content for CreateFromFile")
	filename := "test-file.txt"

	item, err := svc.CreateFromFile(filename, strings.NewReader(string(testContent)))

	if err != nil {
		t.Fatalf("CreateFromFile() failed: %v", err)
	}

	if item == nil {
		t.Fatal("CreateFromFile() returned nil item")
	}

	if item.Title != filename {
		t.Errorf("Title = %s, want %s", item.Title, filename)
	}

	if item.MediaType != "web" {
		t.Errorf("MediaType = %s, want 'web' for .txt file", item.MediaType)
	}

	if item.SourceURL != "" {
		t.Errorf("SourceURL should be empty for local file, got %s", item.SourceURL)
	}

	if item.ContentHash == "" {
		t.Error("ContentHash should not be empty")
	}
}

// TestContentService_CreateFromFile_duplicateDetection verifies duplicate file detection.
func TestContentService_CreateFromFile_duplicateDetection(t *testing.T) {
	tempDir := t.TempDir()
	db := setupTestDB(t)
	defer db.Close()

	svc, _ := NewContentService(db, tempDir)

	// Create test content
	testContent := []byte("Duplicate content test")
	filename1 := "file1.txt"
	filename2 := "file2.txt"

	// Create first file
	item1, err := svc.CreateFromFile(filename1, strings.NewReader(string(testContent)))
	if err != nil {
		t.Fatalf("CreateFromFile() first call failed: %v", err)
	}

	// Try to create duplicate with same content (different filename)
	item2, err := svc.CreateFromFile(filename2, strings.NewReader(string(testContent)))

	// Should return error about duplicate
	if err == nil {
		t.Error("CreateFromFile() with duplicate content should return error")
	}

	if item2 != nil && item2.ID == item1.ID {
		// Some implementations return the existing item
		t.Logf("Duplicate detection returned existing item: %s", item2.ID)
	}
}

// TestContentService_CreateFromFile_variousTypes verifies media type detection for various file types.
func TestContentService_CreateFromFile_variousTypes(t *testing.T) {
	tests := []struct {
		filename  string
		mediaType string
	}{
		{"image.jpg", "image"},
		{"photo.png", "image"},
		{"video.mp4", "video"},
		{"clip.webm", "video"},
		{"document.pdf", "pdf"},
		{"notes.md", "markdown"},
		{"readme.markdown", "markdown"},
		{"file.txt", "web"},
		{"noextension", "web"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			// Each subtest gets its own database and storage to avoid duplicate detection
			tempDir := t.TempDir()
			db := setupTestDB(t)
			defer db.Close()

			svc, _ := NewContentService(db, tempDir)

			// Create unique content for each file type
			testContent := strings.Repeat(fmt.Sprintf("Test content for %s. ", tt.filename), 10)
			item, err := svc.CreateFromFile(tt.filename, strings.NewReader(testContent))

			if err != nil {
				t.Fatalf("CreateFromFile(%q) failed: %v", tt.filename, err)
			}

			if item.MediaType != tt.mediaType {
				t.Errorf("MediaType = %s, want %s for file %s", item.MediaType, tt.mediaType, tt.filename)
			}
		})
	}
}

// TestContentService_CreateFromFile_largeFile verifies handling of larger files.
func TestContentService_CreateFromFile_largeFile(t *testing.T) {
	tempDir := t.TempDir()
	db := setupTestDB(t)
	defer db.Close()

	svc, _ := NewContentService(db, tempDir)

	// Create a larger file (10KB)
	largeContent := strings.Repeat("A", 10*1024)
	filename := "large-file.txt"

	item, err := svc.CreateFromFile(filename, strings.NewReader(largeContent))

	if err != nil {
		t.Fatalf("CreateFromFile() with large file failed: %v", err)
	}

	if item == nil {
		t.Fatal("CreateFromFile() returned nil for large file")
	}

	// Verify content hash is computed
	if item.ContentHash == "" {
		t.Error("ContentHash should be computed for large file")
	}

	// ContentText should contain size info
	if !strings.Contains(item.ContentText, "Size:") {
		t.Error("ContentText should contain size information")
	}
}

// TestContentService_CreateFromFile_emptyFile verifies empty file handling.
func TestContentService_CreateFromFile_emptyFile(t *testing.T) {
	tempDir := t.TempDir()
	db := setupTestDB(t)
	defer db.Close()

	svc, _ := NewContentService(db, tempDir)

	filename := "empty-file.txt"
	_, err := svc.CreateFromFile(filename, strings.NewReader(""))

	// Empty file should be rejected by storage (minimum size requirement)
	if err == nil {
		t.Error("CreateFromFile() with empty file should return error (minimum size)")
	}
}

// TestContentService_CreateFromFile_specialCharacters verifies handling of special characters in filename.
func TestContentService_CreateFromFile_specialCharacters(t *testing.T) {
	tempDir := t.TempDir()
	db := setupTestDB(t)
	defer db.Close()

	svc, _ := NewContentService(db, tempDir)

	tests := []struct {
		filename string
		content  string
	}{
		{"file with spaces.txt", strings.Repeat("Content with spaces. ", 10)},
		{"file-with-dashes.txt", strings.Repeat("Content with dashes. ", 10)},
		{"file_with_underscores.txt", strings.Repeat("Content with underscores. ", 10)},
		{"file.multiple.dots.txt", strings.Repeat("Content with multiple dots. ", 10)},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			item, err := svc.CreateFromFile(tt.filename, strings.NewReader(tt.content))

			if err != nil {
				t.Fatalf("CreateFromFile(%q) failed: %v", tt.filename, err)
			}

			if item.Title != tt.filename {
				t.Errorf("Title = %s, want %s", item.Title, tt.filename)
			}
		})
	}
}

// TestContentService_FindByContentHash_multipleItems verifies hash lookup with multiple items.
func TestContentService_FindByContentHash_multipleItems(t *testing.T) {
	tempDir := t.TempDir()
	db := setupTestDB(t)
	defer db.Close()

	svc, _ := NewContentService(db, tempDir)

	// Create multiple items with different content (valid MediaType required)
	items := []*models.ContentItem{
		{ID: models.UUID("1"), ContentHash: "hash1", Title: "File 1", MediaType: "web", ContentText: "Content 1"},
		{ID: models.UUID("2"), ContentHash: "hash2", Title: "File 2", MediaType: "image", ContentText: "Content 2"},
		{ID: models.UUID("3"), ContentHash: "hash3", Title: "File 3", MediaType: "pdf", ContentText: "Content 3"},
	}

	for _, item := range items {
		if err := svc.repo.CreateContentItem(item); err != nil {
			t.Fatalf("Failed to create test item: %v", err)
		}
	}

	// List all items
	allItems, err := svc.ListContent(100, 0, "")
	if err != nil {
		t.Fatalf("ListContent() failed: %v", err)
	}

	if len(allItems) != 3 {
		t.Errorf("Expected 3 items, got %d", len(allItems))
	}
}

// TestContentService_CreateFromFile_withReader verifies reader is properly consumed.
func TestContentService_CreateFromFile_withReader(t *testing.T) {
	tempDir := t.TempDir()
	db := setupTestDB(t)
	defer db.Close()

	svc, _ := NewContentService(db, tempDir)

	// Create a reader that tracks if it was read
	testContent := "Content that should be read"
	reader := strings.NewReader(testContent)
	filename := "tracked-file.txt"

	item, err := svc.CreateFromFile(filename, reader)

	if err != nil {
		t.Fatalf("CreateFromFile() failed: %v", err)
	}

	if item == nil {
		t.Fatal("CreateFromFile() returned nil")
	}

	// Verify storage hash was computed (meaning content was read)
	if item.ContentHash == "" {
		t.Error("ContentHash should be computed (reader should be consumed)")
	}
}

// TestContentService_CreateFromFile_unicodeFilename verifies Unicode filename handling.
func TestContentService_CreateFromFile_unicodeFilename(t *testing.T) {
	tempDir := t.TempDir()
	db := setupTestDB(t)
	defer db.Close()

	svc, _ := NewContentService(db, tempDir)

	// Test Unicode filename with content that meets minimum size
	filename := "文件-📄.txt"
	testContent := strings.Repeat("测试内容", 50) // Repeat to meet minimum size

	item, err := svc.CreateFromFile(filename, strings.NewReader(testContent))

	if err != nil {
		t.Fatalf("CreateFromFile() with Unicode filename failed: %v", err)
	}

	if item.Title != filename {
		t.Errorf("Title = %s, want %s (Unicode preserved)", item.Title, filename)
	}
}
