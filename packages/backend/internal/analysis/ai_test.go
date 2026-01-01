// Package analysis provides unit tests for AI client mocking.
// T129: Unit test for AI client mocking (OpenAI, Claude, Ollama).
package analysis

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestNewAIAnalyzer verifies analyzer initialization.
func TestNewAIAnalyzer(t *testing.T) {
	config := &AIConfig{
		Provider:   AIProviderOpenAI,
		APIEndpoint: "https://api.openai.com",
		APIKey:      "test-key",
		ModelName:   "gpt-4",
		MaxTokens:   1000,
	}

	analyzer := NewAIAnalyzer(config)

	if analyzer == nil {
		t.Fatal("NewAIAnalyzer() returned nil")
	}

	if analyzer.config != config {
		t.Error("config not set correctly")
	}

	if analyzer.httpClient == nil {
		t.Error("httpClient not initialized")
	}

	if analyzer.fallback == nil {
		t.Error("fallback TF-IDF analyzer not initialized")
	}
}

// TestSummarizeEmptyText verifies empty text handling.
func TestSummarizeEmptyText(t *testing.T) {
	config := &AIConfig{
		Provider:   AIProviderOpenAI,
		APIEndpoint: "https://api.openai.com",
		APIKey:      "test-key",
		ModelName:   "gpt-4",
		MaxTokens:   1000,
	}

	analyzer := NewAIAnalyzer(config)

	summary, err := analyzer.Summarize("")

	if err != nil {
		t.Fatalf("Summarize() returned error: %v", err)
	}

	if summary != "" {
		t.Errorf("expected empty summary, got '%s'", summary)
	}
}

// TestExtractKeywordsEmptyText verifies empty text handling for keywords.
func TestExtractKeywordsEmptyText(t *testing.T) {
	config := &AIConfig{
		Provider:   AIProviderOpenAI,
		APIEndpoint: "https://api.openai.com",
		APIKey:      "test-key",
		ModelName:   "gpt-4",
		MaxTokens:   1000,
	}

	analyzer := NewAIAnalyzer(config)

	keywords, err := analyzer.ExtractKeywords("")

	if err != nil {
		t.Fatalf("ExtractKeywords() returned error: %v", err)
	}

	if keywords == nil {
		t.Fatal("expected non-nil keywords")
	}

	if len(keywords) != 0 {
		t.Errorf("expected 0 keywords, got %d", len(keywords))
	}
}

// TestSummarizeFallback verifies TF-IDF fallback on AI failure.
func TestSummarizeFallback(t *testing.T) {
	// Invalid config (will trigger fallback)
	config := &AIConfig{
		Provider:   AIProvider("invalid"),
		APIEndpoint: "invalid",
		APIKey:      "invalid",
		ModelName:   "invalid",
		MaxTokens:   1000,
	}

	analyzer := NewAIAnalyzer(config)

	text := `Machine learning is a subfield of artificial intelligence.
		Machine learning algorithms build models based on sample data.`

	summary, err := analyzer.Summarize(text)

	// Should fallback to TF-IDF and not error
	if err != nil {
		t.Fatalf("Summarize() with fallback returned error: %v", err)
	}

	if summary == "" {
		t.Error("expected summary from fallback TF-IDF")
	}

	// Verify summary contains some of the input text
	if !strings.Contains(summary, "Machine") && !strings.Contains(summary, "learning") {
		t.Error("fallback summary should contain input text")
	}
}

// TestExtractKeywordsFallback verifies TF-IDF fallback for keywords.
func TestExtractKeywordsFallback(t *testing.T) {
	config := &AIConfig{
		Provider:   AIProvider("invalid"),
		APIEndpoint: "invalid",
		APIKey:      "invalid",
		ModelName:   "invalid",
		MaxTokens:   1000,
	}

	analyzer := NewAIAnalyzer(config)

	text := `Machine learning algorithms build models.`

	keywords, err := analyzer.ExtractKeywords(text)

	if err != nil {
		t.Fatalf("ExtractKeywords() with fallback returned error: %v", err)
	}

	if keywords == nil {
		t.Fatal("expected non-nil keywords from fallback")
	}

	if len(keywords) == 0 {
		t.Error("expected keywords from fallback TF-IDF")
	}
}

// =====================================================
// OpenAI Client Tests
// =====================================================

// TestSummarizeOpenAI verifies OpenAI summarization with mock server.
func TestSummarizeOpenAI(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		if !strings.Contains(r.URL.Path, "chat/completions") {
			t.Errorf("expected chat/completions path, got %s", r.URL.Path)
		}

		auth := r.Header.Get("Authorization")
		if !strings.Contains(auth, "Bearer test-key") {
			t.Errorf("expected Authorization header with Bearer token, got '%s'", auth)
		}

		// Decode request
		var req openAIRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if req.Model != "gpt-4" {
			t.Errorf("expected model 'gpt-4', got '%s'", req.Model)
		}

		// Send mock response
		response := openAIResponse{
			Choices: []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			}{
				{
					Message: struct {
						Content string `json:"content"`
					}{
						Content: "Machine learning is a subset of AI focusing on data-driven algorithms.",
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create analyzer with mock server endpoint
	config := &AIConfig{
		Provider:   AIProviderOpenAI,
		APIEndpoint: server.URL,
		APIKey:      "test-key",
		ModelName:   "gpt-4",
		MaxTokens:   1000,
	}

	analyzer := NewAIAnalyzer(config)

	summary, err := analyzer.Summarize("Machine learning text here.")

	if err != nil {
		t.Fatalf("Summarize() returned error: %v", err)
	}

	expectedSummary := "Machine learning is a subset of AI focusing on data-driven algorithms."
	if summary != expectedSummary {
		t.Errorf("expected '%s', got '%s'", expectedSummary, summary)
	}
}

// TestSummarizeOpenAIError verifies error handling.
func TestSummarizeOpenAIError(t *testing.T) {
	// Create mock server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]string{
				"message": "Invalid request",
				"type":    "invalid_request_error",
			},
		})
	}))
	defer server.Close()

	config := &AIConfig{
		Provider:   AIProviderOpenAI,
		APIEndpoint: server.URL,
		APIKey:      "test-key",
		ModelName:   "gpt-4",
		MaxTokens:   1000,
	}

	analyzer := NewAIAnalyzer(config)

	// Should fallback to TF-IDF
	summary, err := analyzer.Summarize("Some text here.")

	if err != nil {
		t.Fatalf("Summarize() should fallback, got error: %v", err)
	}

	if summary == "" {
		t.Error("expected summary from fallback TF-IDF")
	}
}

// TestExtractKeywordsOpenAI verifies OpenAI keyword extraction.
func TestExtractKeywordsOpenAI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req openAIRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		// Send mock response with keywords
		response := openAIResponse{
			Choices: []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			}{
				{
					Message: struct {
						Content string `json:"content"`
					}{
						Content: "machine, learning, algorithm, data, model",
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := &AIConfig{
		Provider:   AIProviderOpenAI,
		APIEndpoint: server.URL,
		APIKey:      "test-key",
		ModelName:   "gpt-4",
		MaxTokens:   1000,
	}

	analyzer := NewAIAnalyzer(config)

	keywords, err := analyzer.ExtractKeywords("Machine learning algorithms.")

	if err != nil {
		t.Fatalf("ExtractKeywords() returned error: %v", err)
	}

	if len(keywords) == 0 {
		t.Error("expected keywords from AI")
	}

	// Verify expected keywords
	expectedKeywords := []string{"machine", "learning", "algorithm", "data", "model"}
	for _, expected := range expectedKeywords {
		found := false
		for _, kw := range keywords {
			if strings.EqualFold(kw, expected) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected keyword '%s' not found", expected)
		}
	}
}

// =====================================================
// Claude Client Tests
// =====================================================

// TestSummarizeClaude verifies Claude summarization.
func TestSummarizeClaude(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify Claude-specific headers
		if r.Header.Get("anthropic-version") == "" {
			t.Error("expected anthropic-version header")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"text": "AI summary of machine learning concepts.",
				},
			},
		})
	}))
	defer server.Close()

	config := &AIConfig{
		Provider:   AIProviderClaude,
		APIEndpoint: server.URL,
		APIKey:      "test-key",
		ModelName:   "claude-3-opus",
		MaxTokens:   1000,
	}

	analyzer := NewAIAnalyzer(config)

	summary, err := analyzer.Summarize("Machine learning text.")

	if err != nil {
		t.Fatalf("Summarize() returned error: %v", err)
	}

	if summary == "" {
		t.Error("expected non-empty summary")
	}
}

// =====================================================
// Ollama Client Tests
// =====================================================

// TestSummarizeOllama verifies Ollama summarization.
func TestSummarizeOllama(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify Ollama-specific request
		if !strings.Contains(r.URL.Path, "api/generate") {
			t.Errorf("expected api/generate path, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"response": "Ollama generated summary of the text.",
		})
	}))
	defer server.Close()

	config := &AIConfig{
		Provider:   AIProviderOllama,
		APIEndpoint: server.URL,
		APIKey:      "", // Ollama doesn't need API key
		ModelName:   "llama2",
		MaxTokens:   1000,
	}

	analyzer := NewAIAnalyzer(config)

	summary, err := analyzer.Summarize("Some text to summarize.")

	if err != nil {
		t.Fatalf("Summarize() returned error: %v", err)
	}

	if summary == "" {
		t.Error("expected non-empty summary")
	}

	if !strings.Contains(summary, "Ollama") {
		t.Errorf("expected summary to contain 'Ollama', got '%s'", summary)
	}
}

// =====================================================
// Additional Keyword Extraction Tests
// =====================================================

// TestExtractKeywordsClaude verifies Claude keyword extraction.
func TestExtractKeywordsClaude(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify Claude-specific headers
		if r.Header.Get("anthropic-version") == "" {
			t.Error("expected anthropic-version header")
		}

		if r.Header.Get("x-api-key") == "" {
			t.Error("expected x-api-key header")
		}

		// Send mock response with keywords
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"text": "artificial, intelligence, machine, learning, neural, network",
				},
			},
		})
	}))
	defer server.Close()

	config := &AIConfig{
		Provider:   AIProviderClaude,
		APIEndpoint: server.URL,
		APIKey:      "test-key",
		ModelName:   "claude-3-opus",
		MaxTokens:   1000,
	}

	analyzer := NewAIAnalyzer(config)

	keywords, err := analyzer.ExtractKeywords("Artificial intelligence and machine learning are transforming technology.")

	if err != nil {
		t.Fatalf("ExtractKeywords() returned error: %v", err)
	}

	if len(keywords) == 0 {
		t.Error("expected keywords from Claude")
	}

	// Verify expected keywords
	expectedKeywords := []string{"artificial", "intelligence", "machine", "learning", "neural", "network"}
	for _, expected := range expectedKeywords {
		found := false
		for _, kw := range keywords {
			if strings.EqualFold(kw, expected) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected keyword '%s' not found", expected)
		}
	}
}

// TestExtractKeywordsClaudeError verifies Claude keyword extraction error handling.
func TestExtractKeywordsClaudeError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Return Claude API error response
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]string{
				"message": "Invalid request",
				"type":    "invalid_request_error",
			},
		})
	}))
	defer server.Close()

	config := &AIConfig{
		Provider:   AIProviderClaude,
		APIEndpoint: server.URL,
		APIKey:      "test-key",
		ModelName:   "claude-3-opus",
		MaxTokens:   1000,
	}

	analyzer := NewAIAnalyzer(config)

	// Should fallback to TF-IDF
	keywords, err := analyzer.ExtractKeywords("Test text for fallback.")

	if err != nil {
		t.Fatalf("ExtractKeywords() should fallback, got error: %v", err)
	}

	if len(keywords) == 0 {
		t.Error("expected keywords from fallback TF-IDF")
	}
}

// TestExtractKeywordsOllama verifies Ollama keyword extraction.
func TestExtractKeywordsOllama(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify Ollama-specific request
		if !strings.Contains(r.URL.Path, "api/generate") {
			t.Errorf("expected api/generate path, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		// Return comma-separated keywords
		json.NewEncoder(w).Encode(map[string]interface{}{
			"response": "data, science, analytics, statistics, modeling",
		})
	}))
	defer server.Close()

	config := &AIConfig{
		Provider:   AIProviderOllama,
		APIEndpoint: server.URL,
		APIKey:      "", // Ollama doesn't need API key
		ModelName:   "llama2",
		MaxTokens:   1000,
	}

	analyzer := NewAIAnalyzer(config)

	keywords, err := analyzer.ExtractKeywords("Data science involves analytics and statistics.")

	if err != nil {
		t.Fatalf("ExtractKeywords() returned error: %v", err)
	}

	if len(keywords) == 0 {
		t.Error("expected keywords from Ollama")
	}

	// Verify expected keywords
	expectedKeywords := []string{"data", "science", "analytics", "statistics", "modeling"}
	for _, expected := range expectedKeywords {
		found := false
		for _, kw := range keywords {
			if strings.EqualFold(kw, expected) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected keyword '%s' not found", expected)
		}
	}
}

// TestExtractKeywordsOllamaError verifies Ollama keyword extraction error handling.
func TestExtractKeywordsOllamaError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	config := &AIConfig{
		Provider:   AIProviderOllama,
		APIEndpoint: server.URL,
		APIKey:      "",
		ModelName:   "llama2",
		MaxTokens:   1000,
	}

	analyzer := NewAIAnalyzer(config)

	// Should fallback to TF-IDF
	keywords, err := analyzer.ExtractKeywords("Test text for fallback.")

	if err != nil {
		t.Fatalf("ExtractKeywords() should fallback, got error: %v", err)
	}

	if len(keywords) == 0 {
		t.Error("expected keywords from fallback TF-IDF")
	}
}

// =====================================================
// Error Categorization Tests (tests isRetryableError)
// =====================================================

// TestRateLimitError verifies 429 rate limit error handling.
func TestRateLimitError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Rate limit exceeded",
		})
	}))
	defer server.Close()

	config := &AIConfig{
		Provider:   AIProviderOpenAI,
		APIEndpoint: server.URL,
		APIKey:      "test-key",
		ModelName:   "gpt-4",
		MaxTokens:   1000,
	}

	analyzer := NewAIAnalyzer(config)

	// Should fallback to TF-IDF on rate limit
	summary, err := analyzer.Summarize("Test text.")

	if err != nil {
		t.Fatalf("Summarize() should fallback on rate limit, got error: %v", err)
	}

	if summary == "" {
		t.Error("expected fallback summary on rate limit")
	}
}

// TestUnauthorizedError verifies 401 unauthorized error handling.
func TestUnauthorizedError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid API key",
		})
	}))
	defer server.Close()

	config := &AIConfig{
		Provider:   AIProviderClaude,
		APIEndpoint: server.URL,
		APIKey:      "invalid-key",
		ModelName:   "claude-3-opus",
		MaxTokens:   1000,
	}

	analyzer := NewAIAnalyzer(config)

	// Should fallback to TF-IDF on auth error
	summary, err := analyzer.Summarize("Test text.")

	if err != nil {
		t.Fatalf("Summarize() should fallback on auth error, got error: %v", err)
	}

	if summary == "" {
		t.Error("expected fallback summary on auth error")
	}
}

// =====================================================
// Graceful Degradation Tests
// =====================================================

// TestUnsupportedProvider verifies error handling for unsupported provider.
func TestUnsupportedProvider(t *testing.T) {
	config := &AIConfig{
		Provider:   AIProvider("unknown"),
		APIEndpoint: "invalid",
		APIKey:      "invalid",
		ModelName:   "invalid",
		MaxTokens:   1000,
	}

	analyzer := NewAIAnalyzer(config)

	text := "Test text for unsupported provider."

	// Should fallback to TF-IDF
	summary, err := analyzer.Summarize(text)
	if err != nil {
		t.Fatalf("Summarize() should fallback gracefully, got error: %v", err)
	}

	if summary == "" {
		t.Error("expected fallback summary")
	}

	// Same for keywords
	keywords, err := analyzer.ExtractKeywords(text)
	if err != nil {
		t.Fatalf("ExtractKeywords() should fallback gracefully, got error: %v", err)
	}

	if len(keywords) == 0 {
		t.Error("expected fallback keywords")
	}
}

// TestNetworkErrorFallback verifies network error handling.
func TestNetworkErrorFallback(t *testing.T) {
	// Use invalid endpoint (will fail network call)
	config := &AIConfig{
		Provider:   AIProviderOpenAI,
		APIEndpoint: "http://invalid-endpoint-that-does-not-exist-12345.com",
		APIKey:      "test-key",
		ModelName:   "gpt-4",
		MaxTokens:   1000,
	}

	analyzer := NewAIAnalyzer(config)
	analyzer.httpClient = httpClientTimeout(100) // Short timeout for testing

	text := `Machine learning is a subfield of artificial intelligence.`

	// Should fallback to TF-IDF on network error
	summary, err := analyzer.Summarize(text)
	if err != nil {
		t.Fatalf("Summarize() should fallback on network error, got: %v", err)
	}

	if summary == "" {
		t.Error("expected fallback summary on network error")
	}
}

// =====================================================
// Config Validation Tests
// =====================================================

// TestAIProviderConstants verifies provider constants.
func TestAIProviderConstants(t *testing.T) {
	constants := map[AIProvider]bool{
		AIProviderOpenAI: true,
		AIProviderClaude: true,
		AIProviderOllama: true,
	}

	for provider := range constants {
		if provider == "" {
			t.Errorf("provider constant '%v' is empty", provider)
		}
	}
}

// TestAIConfigSerialization verifies JSON serialization.
func TestAIConfigSerialization(t *testing.T) {
	config := AIConfig{
		Provider:   AIProviderOpenAI,
		APIEndpoint: "https://api.openai.com",
		APIKey:      "secret-key",
		ModelName:   "gpt-4",
		MaxTokens:   2000,
	}

	data, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded AIConfig
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Provider != config.Provider {
		t.Errorf("Provider not preserved: expected %v, got %v", config.Provider, decoded.Provider)
	}

	if decoded.APIEndpoint != config.APIEndpoint {
		t.Errorf("APIEndpoint not preserved: expected %v, got %v", config.APIEndpoint, decoded.APIEndpoint)
	}

	if decoded.ModelName != config.ModelName {
		t.Errorf("ModelName not preserved: expected %v, got %v", config.ModelName, decoded.ModelName)
	}

	if decoded.MaxTokens != config.MaxTokens {
		t.Errorf("MaxTokens not preserved: expected %v, got %v", config.MaxTokens, decoded.MaxTokens)
	}
}

// =====================================================
// Benchmarks
// =====================================================

// BenchmarkSummarizeOpenAI benchmarks OpenAI summarization.
func BenchmarkSummarizeOpenAI(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := openAIResponse{
			Choices: []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			}{
				{
					Message: struct {
						Content string `json:"content"`
					}{
						Content: "Summary of the text.",
					},
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := &AIConfig{
		Provider:   AIProviderOpenAI,
		APIEndpoint: server.URL,
		APIKey:      "test-key",
		ModelName:   "gpt-4",
		MaxTokens:   1000,
	}

	analyzer := NewAIAnalyzer(config)
	text := "Machine learning is a subfield of artificial intelligence."

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = analyzer.Summarize(text)
	}
}

// BenchmarkSummarizeFallback benchmarks TF-IDF fallback.
func BenchmarkSummarizeFallback(b *testing.B) {
	config := &AIConfig{
		Provider:   AIProvider("invalid"),
		APIEndpoint: "invalid",
		APIKey:      "invalid",
		ModelName:   "invalid",
		MaxTokens:   1000,
	}

	analyzer := NewAIAnalyzer(config)
	text := `Machine learning is a subfield of artificial intelligence.
		Machine learning algorithms build models based on sample data.
		Artificial intelligence and machine learning are transforming technology.`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = analyzer.Summarize(text)
	}
}

// Helper to create HTTP client with custom timeout
func httpClientTimeout(ms int) *http.Client {
	return &http.Client{
		Timeout: httpClientTimeoutDuration(ms),
	}
}

func httpClientTimeoutDuration(ms int) time.Duration {
	return time.Duration(ms) * time.Millisecond
}
