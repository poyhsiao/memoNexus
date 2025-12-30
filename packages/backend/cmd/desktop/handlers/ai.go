// Package handlers provides REST API handlers for AI configuration and analysis.
package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/kimhsiao/memonexus/backend/internal/analysis"
	"github.com/kimhsiao/memonexus/backend/internal/crypto"
	"github.com/kimhsiao/memonexus/backend/internal/db"
	"github.com/kimhsiao/memonexus/backend/internal/models"
	"github.com/kimhsiao/memonexus/backend/internal/services"
)

// AIHandler handles AI configuration and content analysis operations.
type AIHandler struct {
	repo           *db.Repository
	analysisSvc    *services.AnalysisService
	machineID       string // For encryption key derivation
}

// NewAIHandler creates a new AIHandler.
func NewAIHandler(repo *db.Repository, analysisSvc *services.AnalysisService, machineID string) *AIHandler {
	if machineID == "" {
		machineID = "default" // Fallback for development
	}
	return &AIHandler{
		repo:        repo,
		analysisSvc: analysisSvc,
		machineID:    machineID,
	}
}

// =====================================================
// AI Configuration Endpoints (T137-T139)
// =====================================================

// GetAIConfig handles GET /ai/config
// Returns the current AI configuration with API key redacted (T137).
func (h *AIHandler) GetAIConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get from service (already has API key redacted)
	config := h.analysisSvc.GetAIConfig()
	if config == nil {
		// Return empty config instead of 404
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"enabled": false,
		})
		return
	}

	response := map[string]interface{}{
		"enabled":     config.Provider != "",
		"provider":    config.Provider,
		"api_endpoint": config.APIEndpoint,
		"api_key":     config.APIKey, // Already redacted as "***REDACTED***"
		"model_name":  config.ModelName,
		"max_tokens":  config.MaxTokens,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// SetAIConfig handles POST /ai/config
// Saves encrypted AI credentials and enables AI mode (T138).
func (h *AIHandler) SetAIConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Provider   string `json:"provider"`    // openai, claude, ollama
		APIEndpoint string `json:"api_endpoint"`
		APIKey     string `json:"api_key"`     // Plain text API key from client
		ModelName  string `json:"model_name"`  // e.g., "gpt-4", "claude-3-opus"
		MaxTokens  int    `json:"max_tokens"`  // Optional, default 1000
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate provider
	validProviders := map[string]bool{
		"openai": true,
		"claude": true,
		"ollama": true,
	}
	if !validProviders[request.Provider] {
		http.Error(w, "Invalid provider: must be openai, claude, or ollama", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if request.APIEndpoint == "" {
		http.Error(w, "api_endpoint is required", http.StatusBadRequest)
		return
	}
	if request.APIKey == "" {
		http.Error(w, "api_key is required", http.StatusBadRequest)
		return
	}
	if request.ModelName == "" {
		http.Error(w, "model_name is required", http.StatusBadRequest)
		return
	}

	// Set default max_tokens
	if request.MaxTokens <= 0 {
		request.MaxTokens = 1000
	}

	// Encrypt API key
	encryptedKey, err := crypto.EncryptAPIKey(request.APIKey, h.machineID)
	if err != nil {
		http.Error(w, "Failed to encrypt API key", http.StatusInternalServerError)
		return
	}

	// Disable all existing configs first
	if err := h.repo.DisableAllAIConfig(); err != nil {
		http.Error(w, "Failed to update configuration", http.StatusInternalServerError)
		return
	}

	// Save new config
	config := &models.AIConfig{
		Provider:        request.Provider,
		APIEndpoint:     request.APIEndpoint,
		APIKeyEncrypted: encryptedKey,
		ModelName:       request.ModelName,
		MaxTokens:       request.MaxTokens,
		IsEnabled:       true,
	}

	if err := h.repo.SaveAIConfig(config); err != nil {
		http.Error(w, "Failed to save configuration", http.StatusInternalServerError)
		return
	}

	// Update analysis service
	aiConfig := &analysis.AIConfig{
		Provider:   analysis.AIProvider(request.Provider),
		APIEndpoint: request.APIEndpoint,
		APIKey:     request.APIKey, // Plain text for service use
		ModelName:  request.ModelName,
		MaxTokens:  request.MaxTokens,
	}

	if err := h.analysisSvc.ConfigureAI(aiConfig); err != nil {
		http.Error(w, "Failed to configure AI service: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   "success",
		"message":  "AI configuration updated",
		"provider": request.Provider,
		"model":    request.ModelName,
	})
}

// DeleteAIConfig handles DELETE /ai/config
// Disables AI mode and removes credentials (T139).
func (h *AIHandler) DeleteAIConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get current config to find its ID
	config, err := h.repo.GetAIConfig()
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Failed to retrieve configuration", http.StatusInternalServerError)
		return
	}

	// Delete from database
	if config != nil {
		if err := h.repo.DeleteAIConfig(string(config.ID)); err != nil {
			http.Error(w, "Failed to delete configuration", http.StatusInternalServerError)
			return
		}
	}

	// Disable AI in service
	h.analysisSvc.DisableAI()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "AI configuration deleted",
	})
}

// =====================================================
// Content Analysis Endpoints (T140)
// =====================================================

// GenerateSummary handles POST /content/{id}/summary
// Generates a summary for the specified content item (T140).
func (h *AIHandler) GenerateSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract content ID from URL path
	// URL format: /api/content/{id}/summary
	// The main router handles the /api/content/ prefix, so we get the rest
	// For now, we'll parse the ID from a query parameter for simplicity
	contentID := r.URL.Query().Get("id")
	if contentID == "" {
		http.Error(w, "content id is required", http.StatusBadRequest)
		return
	}

	// Get content item
	item, err := h.repo.GetContentItem(contentID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Content not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve content", http.StatusInternalServerError)
		}
		return
	}

	// Check if content has text
	if item.ContentText == "" {
		http.Error(w, "Content has no text to analyze", http.StatusBadRequest)
		return
	}

	// Generate summary using analysis service
	result, err := h.analysisSvc.GenerateSummary(r.Context(), item.ContentText)
	if err != nil {
		http.Error(w, "Failed to generate summary: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Update content item with summary
	item.Summary = result
	if err := h.repo.UpdateContentItem(item); err != nil {
		http.Error(w, "Failed to save summary", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"content_id":  contentID,
		"summary":     result,
		"method":      "ai", // or "extractive" if fallback used
		"language":    "detected",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ExtractKeywords handles POST /content/{id}/keywords
// Extracts keywords from the specified content item.
func (h *AIHandler) ExtractKeywords(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	contentID := r.URL.Query().Get("id")
	if contentID == "" {
		http.Error(w, "content id is required", http.StatusBadRequest)
		return
	}

	// Get content item
	item, err := h.repo.GetContentItem(contentID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Content not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve content", http.StatusInternalServerError)
		}
		return
	}

	// Extract keywords using analysis service
	keywords, err := h.analysisSvc.ExtractKeywords(r.Context(), item.ContentText)
	if err != nil {
		http.Error(w, "Failed to extract keywords: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"content_id": contentID,
		"keywords":   keywords,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
