// Package handlers provides REST API handlers for content items.
package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/kimhsiao/memonexus/backend/internal/db"
	"github.com/kimhsiao/memonexus/backend/internal/models"
)

// ContentHandler handles content item operations.
type ContentHandler struct {
	repo *db.Repository
}

// NewContentHandler creates a new ContentHandler.
func NewContentHandler(repo *db.Repository) *ContentHandler {
	return &ContentHandler{repo: repo}
}

// ListContentItems handles GET /content
func (h *ContentHandler) ListContentItems(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	mediaType := r.URL.Query().Get("media_type")

	offset := (page - 1) * perPage

	items, err := h.repo.ListContentItems(perPage, offset, mediaType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get total count (simplified - in production, use separate COUNT query)
	total := len(items)
	totalPages := (total + perPage - 1) / perPage
	if totalPages < 1 {
		totalPages = 1
	}

	response := map[string]interface{}{
		"items":       items,
		"total":       total,
		"page":        page,
		"per_page":    perPage,
		"total_pages": totalPages,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CreateContentItem handles POST /content
func (h *ContentHandler) CreateContentItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Type      string   `json:"type"`
		SourceURL string   `json:"source_url"`
		FilePath  string   `json:"file_path"`
		Title     string   `json:"title"`
		Tags      []string `json:"tags"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if request.Type != "url" && request.Type != "file" {
		http.Error(w, "Invalid type: must be 'url' or 'file'", http.StatusBadRequest)
		return
	}
	if request.Type == "url" && request.SourceURL == "" {
		http.Error(w, "source_url is required for type 'url'", http.StatusBadRequest)
		return
	}
	if request.Type == "file" && request.FilePath == "" {
		http.Error(w, "file_path is required for type 'file'", http.StatusBadRequest)
		return
	}

	// Convert tags list to comma-separated string
	tagsStr := ""
	if len(request.Tags) > 0 {
		for i, tag := range request.Tags {
			if i > 0 {
				tagsStr += ","
			}
			tagsStr += tag
		}
	}

	// Create content item
	item := &models.ContentItem{
		Title:       request.Title,
		ContentText: "", // Will be populated by parser
		MediaType:   "web", // Will be updated by parser
		Tags:        tagsStr,
	}

	// TODO: Trigger async parsing based on type (URL or file)
	// For now, create with minimal data

	if err := h.repo.CreateContentItem(item); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(item)
}

// GetContentItem handles GET /content/{id}
func (h *ContentHandler) GetContentItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		// Fallback for older Go versions without PathValue
		id = r.URL.Path[len("/content/"):]
	}

	item, err := h.repo.GetContentItem(id)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Content item not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}

// UpdateContentItem handles PUT /content/{id}
func (h *ContentHandler) UpdateContentItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		id = r.URL.Path[len("/content/"):]
	}

	var request struct {
		Title *string  `json:"title"`
		Tags  []string `json:"tags"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get existing item
	item, err := h.repo.GetContentItem(id)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Content item not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Update fields
	if request.Title != nil {
		item.Title = *request.Title
	}
	if request.Tags != nil {
		tagsStr := ""
		for i, tag := range request.Tags {
			if i > 0 {
				tagsStr += ","
			}
			tagsStr += tag
		}
		item.Tags = tagsStr
	}

	if err := h.repo.UpdateContentItem(item); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}

// DeleteContentItem handles DELETE /content/{id}
func (h *ContentHandler) DeleteContentItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		id = r.URL.Path[len("/content/"):]
	}

	if err := h.repo.DeleteContentItem(id); err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Content item not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
