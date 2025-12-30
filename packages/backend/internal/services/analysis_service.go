// Package services provides content analysis orchestration.
// T132: AnalysisService orchestration layer coordinating TF-IDF and AI analysis.
package services

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/kimhsiao/memonexus/backend/internal/analysis"
	"github.com/kimhsiao/memonexus/backend/internal/analysis/textrank"
	"github.com/kimhsiao/memonexus/backend/internal/models"
)

// AnalysisService coordinates content analysis using both TF-IDF (offline)
// and AI (optional) methods.
type AnalysisService struct {
	// TF-IDF analyzer (always available)
	tfidfAnalyzer *analysis.TFIDFAnalyzer

	// TextRank extractor for keyword extraction
	textrankExtractor *textrank.TextRankExtractor

	// AI analyzer (optional, requires configuration)
	aiAnalyzer *analysis.AIAnalyzer
	aiConfig   *analysis.AIConfig

	// Configuration
	config *AnalysisConfig

	// Event callbacks for WebSocket notifications
	onAnalysisStarted func(contentID string)
	onAnalysisCompleted func(contentID string, result *AnalysisResult)
	onAnalysisFailed func(contentID string, err error)

	mu sync.RWMutex
}

// AnalysisConfig holds configuration for the analysis service.
type AnalysisConfig struct {
	// Enable AI analysis (requires valid AI config)
	EnableAI bool

	// Number of keywords to extract
	NumKeywords int

	// Maximum summary length
	MaxSummaryLength int

	// Timeout for AI analysis
	AITimeoutSeconds int
}

// DefaultAnalysisConfig returns sensible defaults.
func DefaultAnalysisConfig() *AnalysisConfig {
	return &AnalysisConfig{
		EnableAI:         false, // AI is opt-in per constitution
		NumKeywords:       10,
		MaxSummaryLength:  500,
		AITimeoutSeconds:  60,
	}
}

// AnalysisResult represents the outcome of content analysis.
type AnalysisResult struct {
	ContentID   string   `json:"content_id"`
	Keywords    []string `json:"keywords"`
	Summary     string   `json:"summary,omitempty"`
	Language    string   `json:"language"`
	Method      string   `json:"method"` // "tfidf", "textrank", "ai"
	Confidence  float64  `json:"confidence,omitempty"`
	AIUsed      bool     `json:"ai_used"`
	ProcessedAt int64    `json:"processed_at"`
}

// NewAnalysisService creates a new AnalysisService.
func NewAnalysisService(config *AnalysisConfig) *AnalysisService {
	if config == nil {
		config = DefaultAnalysisConfig()
	}

	return &AnalysisService{
		tfidfAnalyzer:    analysis.NewTFIDFAnalyzer(),
		textrankExtractor: textrank.NewTextRankExtractor(),
		config:           config,
	}
}

// ConfigureAI sets up AI analysis with the given configuration.
// Returns an error if the configuration is invalid.
func (s *AnalysisService) ConfigureAI(config *analysis.AIConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if config == nil {
		s.aiAnalyzer = nil
		s.aiConfig = nil
		s.config.EnableAI = false
		return nil
	}

	// Validate configuration
	if config.APIKey == "" {
		return fmt.Errorf("API key is required for AI analysis")
	}

	if config.Provider == "" {
		return fmt.Errorf("AI provider must be specified")
	}

	// Create AI analyzer
	s.aiAnalyzer = analysis.NewAIAnalyzer(config)
	s.aiConfig = config
	s.config.EnableAI = true

	log.Printf("[AnalysisService] AI configured: provider=%s, model=%s",
		config.Provider, config.ModelName)

	return nil
}

// DisableAI disables AI analysis and falls back to TF-IDF.
func (s *AnalysisService) DisableAI() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.aiAnalyzer = nil
	s.aiConfig = nil
	s.config.EnableAI = false

	log.Printf("[AnalysisService] AI analysis disabled")
}

// GetAIConfig returns the current AI configuration (with API key redacted).
func (s *AnalysisService) GetAIConfig() *analysis.AIConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.aiConfig == nil {
		return nil
	}

	// Return a copy with API key redacted for security
	return &analysis.AIConfig{
		Provider:   s.aiConfig.Provider,
		APIEndpoint: s.aiConfig.APIEndpoint,
		APIKey:     "***REDACTED***",
		ModelName:  s.aiConfig.ModelName,
		MaxTokens:  s.aiConfig.MaxTokens,
	}
}

// IsAIEnabled returns true if AI analysis is enabled and configured.
func (s *AnalysisService) IsAIEnabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config.EnableAI && s.aiAnalyzer != nil
}

// AnalyzeContent performs content analysis using the best available method.
// Priority: AI (if enabled) → TextRank → TF-IDF.
// T132: Implements analysis orchestration with graceful degradation.
func (s *AnalysisService) AnalyzeContent(ctx context.Context, contentID string, text string) (*AnalysisResult, error) {
	if text == "" {
		return nil, fmt.Errorf("content text is empty")
	}

	// Notify analysis started
	if s.onAnalysisStarted != nil {
		s.onAnalysisStarted(contentID)
	}

	// Try AI first if enabled
	if s.config.EnableAI && s.aiAnalyzer != nil {
		result, err := s.analyzeWithAI(ctx, contentID, text)
		if err == nil {
			// Notify success
			if s.onAnalysisCompleted != nil {
				s.onAnalysisCompleted(contentID, result)
			}
			return result, nil
		}
		// AI failed, fall through to TF-IDF (graceful degradation per FR-056)
		log.Printf("[AnalysisService] AI analysis failed for %s, falling back to TF-IDF: %v", contentID, err)
	}

	// Fallback to offline analysis (TF-IDF + TextRank)
	result, err := s.analyzeOffline(contentID, text)
	if err != nil {
		// Notify failure
		if s.onAnalysisFailed != nil {
			s.onAnalysisFailed(contentID, err)
		}
		return nil, fmt.Errorf("offline analysis failed: %w", err)
	}

	// Notify success
	if s.onAnalysisCompleted != nil {
		s.onAnalysisCompleted(contentID, result)
	}

	return result, nil
}

// analyzeWithAI performs AI-powered analysis.
func (s *AnalysisService) analyzeWithAI(ctx context.Context, contentID string, text string) (*AnalysisResult, error) {
	s.mu.RLock()
	analyzer := s.aiAnalyzer
	s.mu.RUnlock()

	if analyzer == nil {
		return nil, fmt.Errorf("AI analyzer not configured")
	}

	// Generate summary
	summary, err := analyzer.Summarize(text)
	if err != nil {
		return nil, fmt.Errorf("AI summarization failed: %w", err)
	}

	// Extract keywords
	keywords, err := analyzer.ExtractKeywords(text)
	if err != nil {
		return nil, fmt.Errorf("AI keyword extraction failed: %w", err)
	}

	// Detect language
	language := s.detectLanguage(text)

	return &AnalysisResult{
		ContentID:  contentID,
		Keywords:   keywords,
		Summary:    summary,
		Language:   language,
		Method:     "ai",
		Confidence: 0.9, // AI methods are considered high confidence
		AIUsed:     true,
	}, nil
}

// analyzeOffline performs offline analysis using TF-IDF and TextRank.
func (s *AnalysisService) analyzeOffline(contentID string, text string) (*AnalysisResult, error) {
	// Use TextRank for keyword extraction (T131)
	keywords, err := s.textrankExtractor.Extract(text)
	if err != nil {
		// Fallback to TF-IDF
		tfidfResult, _ := s.tfidfAnalyzer.Analyze(text)
		if tfidfResult != nil {
			keywords = tfidfResult.Keywords
		}
	}

	// Generate summary using TF-IDF (first few sentences)
	summary := s.generateSummary(text)

	// Detect language
	language := s.detectLanguage(text)

	return &AnalysisResult{
		ContentID: contentID,
		Keywords:  keywords,
		Summary:   summary,
		Language:  language,
		Method:    "tfidf",
		Confidence: 0.7, // Offline methods have moderate confidence
		AIUsed:    false,
	}, nil
}

// ExtractKeywords extracts keywords from text using available methods.
func (s *AnalysisService) ExtractKeywords(ctx context.Context, text string) ([]string, error) {
	if text == "" {
		return []string{}, nil
	}

	// Try AI first if enabled
	if s.config.EnableAI && s.aiAnalyzer != nil {
		keywords, err := s.aiAnalyzer.ExtractKeywords(text)
		if err == nil && len(keywords) > 0 {
			return keywords, nil
		}
		// Fall through to offline methods
	}

	// Use TextRank for better quality
	keywords, err := s.textrankExtractor.Extract(text)
	if err == nil && len(keywords) > 0 {
		return keywords, nil
	}

	// Final fallback to TF-IDF
	result, err := s.tfidfAnalyzer.Analyze(text)
	if err != nil {
		return nil, err
	}

	return result.Keywords, nil
}

// GenerateSummary generates a summary for content.
// Uses AI if available, otherwise extracts first few sentences.
func (s *AnalysisService) GenerateSummary(ctx context.Context, text string) (string, error) {
	if text == "" {
		return "", nil
	}

	// Try AI if enabled
	if s.config.EnableAI && s.aiAnalyzer != nil {
		summary, err := s.aiAnalyzer.Summarize(text)
		if err == nil {
			return summary, nil
		}
		log.Printf("[AnalysisService] AI summarization failed, using extractive: %v", err)
	}

	// Extractive summary: first few sentences
	return s.generateSummary(text), nil
}

// generateSummary creates an extractive summary by taking the first few sentences.
func (s *AnalysisService) generateSummary(text string) string {
	// Simple extractive summary: take first 3 sentences or first 500 chars
	maxSentences := 3
	maxChars := s.config.MaxSummaryLength

	if len(text) <= maxChars {
		return text
	}

	// Split by sentence boundaries (rough approximation)
	sentences := splitSentences(text)
	if len(sentences) == 0 {
		return text[:maxChars] + "..."
	}

	// Take first few sentences
	result := ""
	for i, sentence := range sentences {
		if i >= maxSentences || len(result)+len(sentence) > maxChars {
			break
		}
		if result != "" {
			result += " "
		}
		result += sentence
	}

	if len(result) < len(text) {
		result += "..."
	}

	return result
}

// detectLanguage detects the language of the text.
func (s *AnalysisService) detectLanguage(text string) string {
	if textrank.HasCJKText(text) {
		return "cjk"
	}
	return "en"
}

// splitSentences roughly splits text into sentences.
func splitSentences(text string) []string {
	// Simple sentence boundary detection
	sentences := make([]string, 0)
	current := 0

	for i, r := range text {
		if r == '.' || r == '!' || r == '?' || r == '\n' {
			sentence := strings.TrimSpace(text[current:i])
			if sentence != "" {
				sentences = append(sentences, sentence)
			}
			current = i + 1
		}
	}

	// Don't forget the last sentence
	if current < len(text) {
		sentence := strings.TrimSpace(text[current:])
		if sentence != "" {
			sentences = append(sentences, sentence)
		}
	}

	return sentences
}

// SetEventCallbacks sets callbacks for analysis events.
func (s *AnalysisService) SetEventCallbacks(
	started func(contentID string),
	completed func(contentID string, result *AnalysisResult),
	failed func(contentID string, err error),
) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.onAnalysisStarted = started
	s.onAnalysisCompleted = completed
	s.onAnalysisFailed = failed
}

// AnalyzeContentItem is a convenience method that analyzes a ContentItem.
func (s *AnalysisService) AnalyzeContentItem(ctx context.Context, item *models.ContentItem) (*AnalysisResult, error) {
	if item == nil {
		return nil, fmt.Errorf("content item is nil")
	}

	return s.AnalyzeContent(ctx, string(item.ID), item.ContentText)
}

// BatchAnalyze analyzes multiple content items in parallel.
func (s *AnalysisService) BatchAnalyze(ctx context.Context, items []*models.ContentItem) ([]*AnalysisResult, error) {
	results := make([]*AnalysisResult, len(items))
	errors := make([]error, len(items))

	var wg sync.WaitGroup
	for i, item := range items {
		wg.Add(1)
		go func(idx int, itm *models.ContentItem) {
			defer wg.Done()
			result, err := s.AnalyzeContentItem(ctx, itm)
			results[idx] = result
			errors[idx] = err
		}(i, item)
	}

	wg.Wait()

	// Check for errors
	for _, err := range errors {
		if err != nil {
			return results, fmt.Errorf("batch analysis had errors: %w", err)
		}
	}

	return results, nil
}
