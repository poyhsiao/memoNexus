// Package main provides WebSocket server for real-time events (desktop only).
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Only allow connections from localhost
		return r.Host == "localhost" || r.Host == "localhost:8090"
	},
}

// WSClient represents a WebSocket client connection.
type WSClient struct {
	id     string
	conn   *websocket.Conn
	send   chan []byte
	hub    *WSHub
	subscriptions map[string]bool
}

// WSHub maintains active client connections and broadcasts messages.
type WSHub struct {
	clients    map[string]*WSClient
	broadcast  chan []byte
	register   chan *WSClient
	unregister chan *WSClient
	mu         sync.RWMutex
}

// WSEnvelope wraps all WebSocket messages.
type WSEnvelope struct {
	Type      string                 `json:"type"`
	Data      map[string]interface{} `json:"data"`
	Timestamp int64                  `json:"timestamp"`
}

// =====================================================
// WebSocket Event Types
// =====================================================

const (
	// Analysis events (T145-T147)
	EventAnalysisStarted  = "analysis.started"
	EventAnalysisCompleted = "analysis.completed"
	EventAnalysisFailed    = "analysis.failed"

	// Sync events (T164-T168)
	EventSyncStarted         = "sync.started"
	EventSyncProgress        = "sync.progress"
	EventSyncCompleted       = "sync.completed"
	EventSyncFailed          = "sync.failed"
	EventSyncConflictDetected = "sync.conflict_detected"

	// Export events (T188-T189)
	EventExportStarted  = "export.started"
	EventExportProgress = "export.progress"
	EventExportCompleted = "export.completed"
	EventExportFailed    = "export.failed"

	// Import events (T188-T189)
	EventImportStarted  = "import.started"
	EventImportProgress = "import.progress"
	EventImportCompleted = "import.completed"
	EventImportFailed    = "import.failed"
)

// NewWSHub creates a new WebSocket hub.
func NewWSHub() *WSHub {
	hub := &WSHub{
		clients:    make(map[string]*WSClient),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *WSClient),
		unregister: make(chan *WSClient),
	}
	go hub.run()
	return hub
}

// run manages client connections and broadcasts.
func (h *WSHub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.id] = client
			h.mu.Unlock()
			log.Printf("[WS] Client connected: %s (total: %d)", client.id, len(h.clients))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.id]; ok {
				delete(h.clients, client.id)
				close(client.send)
			}
			h.mu.Unlock()
			log.Printf("[WS] Client disconnected: %s (total: %d)", client.id, len(h.clients))

		case message := <-h.broadcast:
			h.mu.RLock()
			for _, client := range h.clients {
				select {
				case client.send <- message:
				default:
					// Client send buffer is full, close connection
					close(client.send)
					delete(h.clients, client.id)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Broadcast sends a message to all subscribed clients.
func (h *WSHub) Broadcast(messageType string, data map[string]interface{}) {
	envelope := WSEnvelope{
		Type:      messageType,
		Data:      data,
		Timestamp: time.Now().Unix(),
	}

	bytes, err := json.Marshal(envelope)
	if err != nil {
		log.Printf("[WS] Failed to marshal message: %v", err)
		return
	}

	h.broadcast <- bytes
}

// =====================================================
// Analysis Event Broadcasters (T145-T147)
// =====================================================

// BroadcastAnalysisStarted notifies clients that content analysis has started.
func (h *WSHub) BroadcastAnalysisStarted(contentID string, operation string) {
	h.Broadcast(EventAnalysisStarted, map[string]interface{}{
		"content_id": contentID,
		"operation":  operation, // "summary" or "keywords"
	})
}

// BroadcastAnalysisCompleted notifies clients that analysis completed successfully.
func (h *WSHub) BroadcastAnalysisCompleted(contentID string, result map[string]interface{}) {
	h.Broadcast(EventAnalysisCompleted, map[string]interface{}{
		"content_id": contentID,
		"result":     result,
	})
}

// BroadcastAnalysisFailed notifies clients that analysis failed with graceful degradation.
func (h *WSHub) BroadcastAnalysisFailed(contentID string, errMsg string, fallbackMethod string) {
	h.Broadcast(EventAnalysisFailed, map[string]interface{}{
		"content_id":      contentID,
		"error":           errMsg,
		"fallback_method": fallbackMethod, // "tfidf" or "textrank"
	})
}

// =====================================================
// Sync Event Broadcasters (T164-T168)
// =====================================================

// BroadcastSyncStarted notifies clients that sync has started (T164).
func (h *WSHub) BroadcastSyncStarted() {
	h.Broadcast(EventSyncStarted, map[string]interface{}{
		"status": "started",
	})
}

// BroadcastSyncProgress notifies clients of sync progress (T165).
func (h *WSHub) BroadcastSyncProgress(percent int, completed int, total int, currentItem string) {
	h.Broadcast(EventSyncProgress, map[string]interface{}{
		"percent":     percent,
		"completed":   completed,
		"total":       total,
		"current_item": currentItem,
	})
}

// BroadcastSyncCompleted notifies clients that sync completed successfully (T166).
func (h *WSHub) BroadcastSyncCompleted(uploaded int, downloaded int, duration time.Duration) {
	h.Broadcast(EventSyncCompleted, map[string]interface{}{
		"uploaded":   uploaded,
		"downloaded": downloaded,
		"duration":   duration.Milliseconds(), // Duration in milliseconds
		"status":     "completed",
	})
}

// BroadcastSyncFailed notifies clients that sync failed (T167).
func (h *WSHub) BroadcastSyncFailed(errorCode string, retryable bool, retryAfter int) {
	h.Broadcast(EventSyncFailed, map[string]interface{}{
		"error_code": errorCode,
		"retryable":  retryable,
		"retry_after": retryAfter, // Seconds until retry suggested
		"status":     "failed",
	})
}

// BroadcastSyncConflictDetected notifies clients that sync conflicts were detected (T168).
func (h *WSHub) BroadcastSyncConflictDetected(conflicts []map[string]interface{}, resolution string) {
	h.Broadcast(EventSyncConflictDetected, map[string]interface{}{
		"conflicts":  conflicts,
		"resolution": resolution, // "last_write_wins", "manual_review", etc.
	})
}

// =====================================================
// Export Event Broadcasters (T188-T189)
// =====================================================

// BroadcastExportStarted notifies clients that export has started (T188).
func (h *WSHub) BroadcastExportStarted(includeMedia bool, encrypted bool) {
	h.Broadcast(EventExportStarted, map[string]interface{}{
		"include_media": includeMedia,
		"encrypted":     encrypted,
	})
}

// BroadcastExportProgress notifies clients of export progress (T189).
func (h *WSHub) BroadcastExportProgress(stage string, percent int, currentItem string) {
	h.Broadcast(EventExportProgress, map[string]interface{}{
		"stage":         stage, // "creating", "compressing", "encrypting", etc.
		"percent":       percent,
		"current_file":  currentItem,
	})
}

// BroadcastExportCompleted notifies clients that export completed successfully (T189).
func (h *WSHub) BroadcastExportCompleted(filePath string, sizeBytes int64, itemCount int, checksum string) {
	h.Broadcast(EventExportCompleted, map[string]interface{}{
		"file_path":  filePath,
		"size_bytes": sizeBytes,
		"item_count": itemCount,
		"checksum":   checksum,
	})
}

// BroadcastExportFailed notifies clients that export failed (T189).
func (h *WSHub) BroadcastExportFailed(errorMsg string) {
	h.Broadcast(EventExportFailed, map[string]interface{}{
		"error": errorMsg,
	})
}

// =====================================================
// Import Event Broadcasters (T188-T189)
// =====================================================

// BroadcastImportStarted notifies clients that import has started (T188).
func (h *WSHub) BroadcastImportStarted(archivePath string) {
	h.Broadcast(EventImportStarted, map[string]interface{}{
		"archive_path": archivePath,
	})
}

// BroadcastImportProgress notifies clients of import progress (T189).
func (h *WSHub) BroadcastImportProgress(stage string, percent int, currentItem string) {
	h.Broadcast(EventImportProgress, map[string]interface{}{
		"stage":         stage, // "validating", "extracting", "decrypting", "restoring", etc.
		"percent":       percent,
		"current_item":  currentItem,
	})
}

// BroadcastImportCompleted notifies clients that import completed successfully (T189).
func (h *WSHub) BroadcastImportCompleted(importedCount int, skippedCount int) {
	h.Broadcast(EventImportCompleted, map[string]interface{}{
		"imported_count": importedCount,
		"skipped_count":  skippedCount,
	})
}

// BroadcastImportFailed notifies clients that import failed (T189).
func (h *WSHub) BroadcastImportFailed(errorMsg string) {
	h.Broadcast(EventImportFailed, map[string]interface{}{
		"error": errorMsg,
	})
}

// readPump pumps messages from the WebSocket connection.
func (c *WSClient) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[WS] Read error: %v", err)
			}
			break
		}

		// Handle client messages
		var msg map[string]interface{}
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("[WS] Invalid message format: %v", err)
			continue
		}

		action, ok := msg["action"].(string)
		if !ok {
			continue
		}

		switch action {
		case "subscribe":
			if events, ok := msg["events"].([]interface{}); ok {
				for _, e := range events {
					if eventStr, ok := e.(string); ok {
						c.subscriptions[eventStr] = true
					}
				}
				// Send acknowledgment
				c.sendAck("subscribe_ack", events)
			}

		case "unsubscribe":
			if events, ok := msg["events"].([]interface{}); ok {
				for _, e := range events {
					if eventStr, ok := e.(string); ok {
						delete(c.subscriptions, eventStr)
					}
				}
			}

		case "ping":
			// Respond with pong
			c.sendPong()
		}
	}
}

// writePump pumps messages to the WebSocket connection.
func (c *WSClient) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// sendAck sends a subscription acknowledgment.
func (c *WSClient) sendAck(action string, events []interface{}) {
	envelope := map[string]interface{}{
		"action": action,
		"subscribed": events,
		"timestamp": time.Now().Unix(),
	}

	bytes, _ := json.Marshal(envelope)
	c.send <- bytes
}

// sendPong sends a pong response.
func (c *WSClient) sendPong() {
	envelope := map[string]interface{}{
		"action": "pong",
		"timestamp": time.Now().Unix(),
	}

	bytes, _ := json.Marshal(envelope)
	c.send <- bytes
}

// HandleWebSocket handles WebSocket connections.
func HandleWebSocket(hub *WSHub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Upgrade to WebSocket
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("[WS] Failed to upgrade: %v", err)
			return
		}

		// Generate client ID
		clientID := time.Now().Format("20060102150405") + "-" + r.RemoteAddr

		client := &WSClient{
			id:     clientID,
			conn:   conn,
			send:   make(chan []byte, 256),
			hub:    hub,
			subscriptions: make(map[string]bool),
		}

		hub.register <- client

		// Start pumps
		go client.writePump()
		go client.readPump()
	}
}
