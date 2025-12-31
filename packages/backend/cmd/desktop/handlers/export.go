// Package handlers provides REST API handlers for data export/import.
// T186-T187: Export/Import REST API endpoints.
package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/kimhsiao/memonexus/backend/internal/db"
	"github.com/kimhsiao/memonexus/backend/internal/export"
)

// ExportHandler handles export/import operations.
type ExportHandler struct {
	repo   *db.Repository
	export *export.ExportService
}

// NewExportHandler creates a new ExportHandler.
func NewExportHandler(repo *db.Repository) *ExportHandler {
	return &ExportHandler{
		repo:   repo,
		export: export.NewExportService(repo),
	}
}

// ExportRequest represents the export request body.
type ExportRequest struct {
	Password     string `json:"password"`      // Optional password for encryption
	IncludeMedia bool   `json:"include_media"`  // Whether to include media files
	OutputPath   string `json:"output_path"`    // Optional custom output path
}

// ImportRequest represents the import request body.
type ImportRequest struct {
	ArchivePath string `json:"archive_path"` // Path to the archive file
	Password     string `json:"password"`     // Password for decryption (if encrypted)
}

// Export handles POST /export
// T186: Trigger export with password and include_media option.
func (h *ExportHandler) Export(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ExportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Create export configuration
	config := &export.ExportConfig{
		Password:     req.Password,
		IncludeMedia: req.IncludeMedia,
		OutputPath:   req.OutputPath,
	}

	// Perform export
	result, err := h.export.Export(config)
	if err != nil {
		http.Error(w, "Export failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// Import handles POST /import
// T187: Import archive with password, validation, and restoration.
func (h *ExportHandler) Import(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ImportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Validate archive path
	if req.ArchivePath == "" {
		http.Error(w, "archive_path is required", http.StatusBadRequest)
		return
	}

	// Create import configuration
	config := &export.ImportConfig{
		ArchivePath: req.ArchivePath,
		Password:     req.Password,
	}

	// Perform import
	result, err := h.export.Import(config)
	if err != nil {
		http.Error(w, "Import failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// ExportStatus handles GET /export/status
// Returns the current export status and recent exports.
func (h *ExportHandler) ExportStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Implement export status tracking
	// For now, return empty status
	status := map[string]interface{}{
		"active":       false,
		"recent":       []interface{}{},
		"exports_dir":  "exports/",
		"total_exports": 0,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}
