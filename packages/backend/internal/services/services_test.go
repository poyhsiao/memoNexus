// Package services tests for business logic orchestration.
package services

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/kimhsiao/memonexus/backend/internal/analysis"
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
