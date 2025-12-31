// Package handlers provides REST API handlers for sync configuration and operations.
package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/kimhsiao/memonexus/backend/internal/crypto"
	"github.com/kimhsiao/memonexus/backend/internal/db"
	"github.com/kimhsiao/memonexus/backend/internal/models"
	"github.com/kimhsiao/memonexus/backend/internal/sync"
	"github.com/kimhsiao/memonexus/backend/internal/sync/queue"
)

// SyncHandler handles sync configuration and operations.
type SyncHandler struct {
	repo     *db.Repository
	engine   *sync.SyncEngine
	queue    *queue.SyncQueue
	machineID string // For encryption key derivation
	wsHub    WSSyncBroadcaster // WebSocket broadcaster for T164-T168
}

// WSSyncBroadcaster interface for sync WebSocket events.
type WSSyncBroadcaster interface {
	BroadcastSyncStarted()
	BroadcastSyncProgress(percent int, completed int, total int, currentItem string)
	BroadcastSyncCompleted(uploaded int, downloaded int, duration time.Duration)
	BroadcastSyncFailed(errorCode string, retryable bool, retryAfter int)
	BroadcastSyncConflictDetected(conflicts []map[string]interface{}, resolution string)
}

// NewSyncHandler creates a new SyncHandler.
func NewSyncHandler(repo *db.Repository, engine *sync.SyncEngine, queue *queue.SyncQueue, machineID string) *SyncHandler {
	if machineID == "" {
		machineID = "default"
	}
	return &SyncHandler{
		repo:     repo,
		engine:   engine,
		queue:    queue,
		machineID: machineID,
		wsHub:    nil, // Set via SetWebSocketHub
	}
}

// SetWebSocketHub sets the WebSocket hub for broadcasting sync events.
func (h *SyncHandler) SetWebSocketHub(wsHub WSSyncBroadcaster) {
	h.wsHub = wsHub
}

// =====================================================
// Sync Configuration Endpoints (T159-T161)
// =====================================================

// GetCredentials handles GET /sync/credentials
// Returns the current S3 configuration with secrets redacted (T159).
func (h *SyncHandler) GetCredentials(w http.ResponseWriter, r *http.Request) {
	// Get sync credentials from database
	creds, err := h.repo.GetSyncCredentials()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// No credentials configured
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"configured": false,
			})
			return
		}
		http.Error(w, "Failed to retrieve credentials", http.StatusInternalServerError)
		return
	}

	// Check if enabled
	if !creds.IsEnabled {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"configured": false,
		})
		return
	}

	response := map[string]interface{}{
		"configured":   true,
		"endpoint":     creds.Endpoint,
		"bucket_name":   creds.BucketName,
		"region":       creds.Region,
		"access_key":   "***REDACTED***",
		"secret_key":   "***REDACTED***",
		"last_tested":  creds.UpdatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// SetCredentials handles POST /sync/credentials
// Saves encrypted S3 credentials and enables sync (T160).
func (h *SyncHandler) SetCredentials(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Endpoint   string `json:"endpoint"`
		BucketName string `json:"bucket_name"`
		Region     string `json:"region"`
		AccessKey  string `json:"access_key"`
		SecretKey  string `json:"secret_key"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if request.Endpoint == "" {
		http.Error(w, "endpoint is required", http.StatusBadRequest)
		return
	}
	if request.BucketName == "" {
		http.Error(w, "bucket_name is required", http.StatusBadRequest)
		return
	}
	if request.AccessKey == "" {
		http.Error(w, "access_key is required", http.StatusBadRequest)
		return
	}
	if request.SecretKey == "" {
		http.Error(w, "secret_key is required", http.StatusBadRequest)
		return
	}

	// Set default region
	if request.Region == "" {
		request.Region = "us-east-1"
	}

	// Encrypt credentials
	encryptedAccessKey, err := crypto.EncryptAPIKey(request.AccessKey, h.machineID)
	if err != nil {
		http.Error(w, "Failed to encrypt access key", http.StatusInternalServerError)
		return
	}

	encryptedSecretKey, err := crypto.EncryptAPIKey(request.SecretKey, h.machineID)
	if err != nil {
		http.Error(w, "Failed to encrypt secret key", http.StatusInternalServerError)
		return
	}

	// Disable existing credentials first
	if err := h.repo.DisableAllSyncCredentials(); err != nil {
		log.Printf("Failed to disable existing credentials: %v", err)
	}

	// Save new credentials
	creds := &models.SyncCredential{
		Endpoint:           request.Endpoint,
		BucketName:         request.BucketName,
		Region:             request.Region,
		AccessKeyEncrypted: encryptedAccessKey,
		SecretKeyEncrypted: encryptedSecretKey,
		IsEnabled:          true,
		CreatedAt:          time.Now().Unix(),
		UpdatedAt:          time.Now().Unix(),
	}

	if err := h.repo.SaveSyncCredential(creds); err != nil {
		http.Error(w, "Failed to save credentials", http.StatusInternalServerError)
		return
	}

	// Update sync engine with new S3 client
	// Note: In production, you'd update the engine's storage client here
	// For now, credentials are saved and can be used on next sync

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Sync credentials saved",
	})
}

// DeleteCredentials handles DELETE /sync/credentials
// Disables sync and removes credentials (T161).
func (h *SyncHandler) DeleteCredentials(w http.ResponseWriter, r *http.Request) {
	// Get current credentials
	creds, err := h.repo.GetSyncCredentials()
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Failed to retrieve credentials", http.StatusInternalServerError)
		return
	}

	// Delete if exists
	if creds != nil {
		if err := h.repo.DeleteSyncCredential(string(creds.ID)); err != nil {
			http.Error(w, "Failed to delete credentials", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Sync credentials removed",
	})
}

// =====================================================
// Sync Status and Trigger Endpoints (T162-T163)
// =====================================================

// GetStatus handles GET /sync/status
// Returns current sync status, last sync time, and pending changes (T162).
func (h *SyncHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	status := h.engine.Status()

	response := map[string]interface{}{
		"status": status,
	}

	// Add last sync time if available
	if lastSync := h.engine.LastSync(); lastSync != nil {
		response["last_sync"] = lastSync.Unix()
	}

	// Add pending changes count
	response["pending_changes"] = h.engine.PendingChanges()

	// Add queue statistics
	if h.queue != nil {
		queueStats := h.queue.GetStats()
		response["queue_stats"] = queueStats
	}

	// Check if credentials are configured
	creds, err := h.repo.GetSyncCredentials()
	if err == nil && creds.IsEnabled {
		response["configured"] = true
	} else {
		response["configured"] = false
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// TriggerSync handles POST /sync/now
// Triggers immediate sync operation (T163).
func (h *SyncHandler) TriggerSync(w http.ResponseWriter, r *http.Request) {
	// T164: Broadcast sync started event
	if h.wsHub != nil {
		h.wsHub.BroadcastSyncStarted()
	}

	// Perform sync
	ctx := r.Context()
	result, err := h.engine.Sync(ctx)

	if err != nil {
		// T167: Broadcast sync failed event
		if h.wsHub != nil {
			retryable := true // Most sync errors are retryable
			retryAfter := 60  // Suggest retry after 60 seconds
			h.wsHub.BroadcastSyncFailed("SYNC_ERROR", retryable, retryAfter)
		}

		http.Error(w, "Sync failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// T166: Broadcast sync completed event
	if h.wsHub != nil {
		h.wsHub.BroadcastSyncCompleted(result.Uploaded, result.Downloaded, result.Duration)
	}

	response := map[string]interface{}{
		"status":    "success",
		"uploaded":  result.Uploaded,
		"downloaded": result.Downloaded,
		"conflicts":  result.Conflicts,
		"duration":  result.Duration.Milliseconds(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
