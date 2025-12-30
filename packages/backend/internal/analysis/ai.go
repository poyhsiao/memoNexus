// Package analysis provides AI-powered content analysis.
package analysis

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// AIProvider represents supported AI providers.
type AIProvider string

const (
	AIProviderOpenAI AIProvider = "openai"
	AIProviderClaude AIProvider = "claude"
	AIProviderOllama AIProvider = "ollama"
)

// AIConfig holds AI service configuration.
type AIConfig struct {
	Provider   AIProvider `json:"provider"`
	APIEndpoint string   `json:"api_endpoint"`
	APIKey      string   `json:"api_key"`
	ModelName   string   `json:"model_name"`
	MaxTokens   int      `json:"max_tokens"`
}

// AIAnalyzer provides AI-powered content analysis.
type AIAnalyzer struct {
	config     *AIConfig
	httpClient *http.Client
	fallback   *TFIDFAnalyzer // Fallback to TF-IDF
}

// NewAIAnalyzer creates a new AIAnalyzer.
func NewAIAnalyzer(config *AIConfig) *AIAnalyzer {
	return &AIAnalyzer{
		config: config,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		fallback: NewTFIDFAnalyzer(),
	}
}

// Summarize generates an AI summary for the content.
// Falls back to TF-IDF if AI service fails.
func (a *AIAnalyzer) Summarize(text string) (string, error) {
	if text == "" {
		return "", nil
	}

	// Try AI summarization
	summary, err := a.aiSummarize(text)
	if err != nil {
		// Fallback to TF-IDF
		result, _ := a.fallback.Analyze(text)
		if result.Summary != "" {
			return result.Summary, nil
		}
		return "", err
	}

	return summary, nil
}

// ExtractKeywords extracts keywords using AI.
// Falls back to TF-IDF if AI service fails.
func (a *AIAnalyzer) ExtractKeywords(text string) ([]string, error) {
	if text == "" {
		return []string{}, nil
	}

	// Try AI keyword extraction
	keywords, err := a.aiExtractKeywords(text)
	if err != nil {
		// Fallback to TF-IDF
		result, _ := a.fallback.Analyze(text)
		return result.Keywords, nil
	}

	return keywords, nil
}

// aiSummarize performs AI-based summarization.
func (a *AIAnalyzer) aiSummarize(text string) (string, error) {
	switch a.config.Provider {
	case AIProviderOpenAI:
		return a.summarizeOpenAI(text)
	case AIProviderClaude:
		return a.summarizeClaude(text)
	case AIProviderOllama:
		return a.summarizeOllama(text)
	default:
		return "", fmt.Errorf("unsupported AI provider: %s", a.config.Provider)
	}
}

// aiExtractKeywords performs AI-based keyword extraction.
func (a *AIAnalyzer) aiExtractKeywords(text string) ([]string, error) {
	switch a.config.Provider {
	case AIProviderOpenAI:
		return a.extractKeywordsOpenAI(text)
	case AIProviderClaude:
		return a.extractKeywordsClaude(text)
	case AIProviderOllama:
		return a.extractKeywordsOllama(text)
	default:
		return []string{}, fmt.Errorf("unsupported AI provider: %s", a.config.Provider)
	}
}

// =====================================================
// OpenAI Integration
// =====================================================

type openAIRequest struct {
	Model    string    `json:"model"`
	Messages []message `json:"messages"`
	MaxTokens int      `json:"max_tokens,omitempty"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

func (a *AIAnalyzer) summarizeOpenAI(text string) (string, error) {
	// Truncate text if too long
	maxInput := 12000
	if len(text) > maxInput {
		text = text[:maxInput] + "..."
	}

	reqBody := openAIRequest{
		Model: a.config.ModelName,
		Messages: []message{
			{
				Role: "system",
				Content: "You are a helpful assistant that summarizes text concisely. " +
					"Provide a 2-3 sentence summary capturing the main points.",
			},
			{
				Role:    "user",
				Content: "Summarize the following text:\n\n" + text,
			},
		},
		MaxTokens: a.config.MaxTokens,
	}

	resp, err := a.doOpenAIRequest(reqBody)
	if err != nil {
		return "", err
	}

	if resp.Error != nil {
		return "", fmt.Errorf("OpenAI API error: %s", resp.Error.Message)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI")
	}

	return resp.Choices[0].Message.Content, nil
}

func (a *AIAnalyzer) extractKeywordsOpenAI(text string) ([]string, error) {
	maxInput := 12000
	if len(text) > maxInput {
		text = text[:maxInput] + "..."
	}

	reqBody := openAIRequest{
		Model: a.config.ModelName,
		Messages: []message{
			{
				Role: "system",
				Content: "Extract 5-10 key topics or keywords from the text. " +
					"Respond with a comma-separated list only, no explanation.",
			},
			{
				Role:    "user",
				Content: "Extract keywords from:\n\n" + text,
			},
		},
		MaxTokens: 100,
	}

	resp, err := a.doOpenAIRequest(reqBody)
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("OpenAI API error: %s", resp.Error.Message)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	// Parse comma-separated response
	content := resp.Choices[0].Message.Content
	return parseCommaSeparatedList(content), nil
}

func (a *AIAnalyzer) doOpenAIRequest(reqBody openAIRequest) (*openAIResponse, error) {
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", a.config.APIEndpoint+"/chat/completions", bytes.NewReader(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+a.config.APIKey)

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("OpenAI API returned %d: %s", resp.StatusCode, string(body))
	}

	var openAIResp openAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
		return nil, err
	}

	return &openAIResp, nil
}

// =====================================================
// Claude Integration
// =====================================================

type claudeRequest struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens,omitempty"`
	Messages  []message `json:"messages"`
}

type claudeResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

func (a *AIAnalyzer) summarizeClaude(text string) (string, error) {
	maxInput := 12000
	if len(text) > maxInput {
		text = text[:maxInput] + "..."
	}

	reqBody := claudeRequest{
		Model:     a.config.ModelName,
		MaxTokens: a.config.MaxTokens,
		Messages: []message{
			{
				Role: "user",
				Content: "Summarize the following text in 2-3 sentences:\n\n" + text,
			},
		},
	}

	resp, err := a.doClaudeRequest(reqBody)
	if err != nil {
		return "", err
	}

	if resp.Error != nil {
		return "", fmt.Errorf("Claude API error: %s", resp.Error.Message)
	}

	if len(resp.Content) == 0 {
		return "", fmt.Errorf("no response from Claude")
	}

	return resp.Content[0].Text, nil
}

func (a *AIAnalyzer) extractKeywordsClaude(text string) ([]string, error) {
	maxInput := 12000
	if len(text) > maxInput {
		text = text[:maxInput] + "..."
	}

	reqBody := claudeRequest{
		Model:     a.config.ModelName,
		MaxTokens: 100,
		Messages: []message{
			{
				Role: "user",
				Content: "Extract 5-10 key topics or keywords from this text. " +
					"Respond with comma-separated list only:\n\n" + text,
			},
		},
	}

	resp, err := a.doClaudeRequest(reqBody)
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("Claude API error: %s", resp.Error.Message)
	}

	if len(resp.Content) == 0 {
		return nil, fmt.Errorf("no response from Claude")
	}

	return parseCommaSeparatedList(resp.Content[0].Text), nil
}

func (a *AIAnalyzer) doClaudeRequest(reqBody claudeRequest) (*claudeResponse, error) {
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", a.config.APIEndpoint+"/v1/messages", bytes.NewReader(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", a.config.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Claude API returned %d: %s", resp.StatusCode, string(body))
	}

	var claudeResp claudeResponse
	if err := json.NewDecoder(resp.Body).Decode(&claudeResp); err != nil {
		return nil, err
	}

	return &claudeResp, nil
}

// =====================================================
// Ollama Integration (Local)
// =====================================================

type ollamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type ollamaResponse struct {
	Response string `json:"response"`
	Error    string `json:"error,omitempty"`
}

func (a *AIAnalyzer) summarizeOllama(text string) (string, error) {
	maxInput := 12000
	if len(text) > maxInput {
		text = text[:maxInput] + "..."
	}

	reqBody := ollamaRequest{
		Model: a.config.ModelName,
		Prompt: "Summarize the following text in 2-3 sentences:\n\n" + text,
		Stream: false,
	}

	resp, err := a.doOllamaRequest(reqBody)
	if err != nil {
		return "", err
	}

	if resp.Error != "" {
		return "", fmt.Errorf("Ollama error: %s", resp.Error)
	}

	return resp.Response, nil
}

func (a *AIAnalyzer) extractKeywordsOllama(text string) ([]string, error) {
	maxInput := 12000
	if len(text) > maxInput {
		text = text[:maxInput] + "..."
	}

	reqBody := ollamaRequest{
		Model: a.config.ModelName,
		Prompt: "Extract 5-10 key topics or keywords from this text. " +
			"Respond with comma-separated list only:\n\n" + text,
		Stream: false,
	}

	resp, err := a.doOllamaRequest(reqBody)
	if err != nil {
		return nil, err
	}

	if resp.Error != "" {
		return nil, fmt.Errorf("Ollama error: %s", resp.Error)
	}

	return parseCommaSeparatedList(resp.Response), nil
}

func (a *AIAnalyzer) doOllamaRequest(reqBody ollamaRequest) (*ollamaResponse, error) {
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", a.config.APIEndpoint+"/api/generate", bytes.NewReader(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Ollama returned %d: %s", resp.StatusCode, string(body))
	}

	var ollamaResp ollamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, err
	}

	return &ollamaResp, nil
}

// =====================================================
// Helpers
// =====================================================

// parseCommaSeparatedList parses a comma-separated string into a slice.
func parseCommaSeparatedList(s string) []string {
	var result []string

	for _, item := range strings.Split(s, ",") {
		item = strings.TrimSpace(item)
		item = strings.Trim(item, `"'`) // Trim quotes
		if item != "" {
			result = append(result, item)
		}
	}

	return result
}
