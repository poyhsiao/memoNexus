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
