// Package handlers provides REST API handlers for tags.
package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/kimhsiao/memonexus/backend/internal/db"
	"github.com/kimhsiao/memonexus/backend/internal/models"
)

// TagHandler handles tag operations.
type TagHandler struct {
	repo *db.Repository
}

// NewTagHandler creates a new TagHandler.
func NewTagHandler(repo *db.Repository) *TagHandler {
	return &TagHandler{repo: repo}
}

// ListTags handles GET /tags
func (h *TagHandler) ListTags(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tags, err := h.repo.ListTags()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tags)
}

// CreateTag handles POST /tags
func (h *TagHandler) CreateTag(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Name  string `json:"name"`
		Color string `json:"color"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate name
	if request.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}
	if len(request.Name) > 50 {
		http.Error(w, "name must be 50 characters or less", http.StatusBadRequest)
		return
	}

	// Set default color
	if request.Color == "" {
		request.Color = "#3B82F6"
	}

	tag := &models.Tag{
		Name:  request.Name,
		Color: request.Color,
	}

	if err := h.repo.CreateTag(tag); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(tag)
}

// GetTag handles GET /tags/{id}
func (h *TagHandler) GetTag(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		id = r.URL.Path[len("/tags/"):]
	}

	tag, err := h.repo.GetTag(id)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Tag not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tag)
}

// UpdateTag handles PUT /tags/{id}
func (h *TagHandler) UpdateTag(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		id = r.URL.Path[len("/tags/"):]
	}

	var request struct {
		Name  *string `json:"name"`
		Color *string `json:"color"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get existing tag
	tag, err := h.repo.GetTag(id)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Tag not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Update fields
	if request.Name != nil {
		if len(*request.Name) > 50 {
			http.Error(w, "name must be 50 characters or less", http.StatusBadRequest)
			return
		}
		tag.Name = *request.Name
	}
	if request.Color != nil {
		tag.Color = *request.Color
	}

	if err := h.repo.UpdateTag(tag); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tag)
}

// DeleteTag handles DELETE /tags/{id}
func (h *TagHandler) DeleteTag(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		id = r.URL.Path[len("/tags/"):]
	}

	if err := h.repo.DeleteTag(id); err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Tag not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
