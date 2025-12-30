// Package main provides the embedded PocketBase server for desktop platforms.
// Desktop clients communicate via REST/WebSocket on localhost:8090.
package main

import (
	"log"
	"net/http"
	"os"

	_ "modernc.org/sqlite"
)

func main() {
	// TODO: Integrate PocketBase embedded server
	// For now, provide a simple health check

	port := "8090"
	log.Printf("MemoNexus Desktop Server starting on port %s...", port)

	// Simple health check endpoint
	http.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","service":"memonexus-desktop"}`))
	})

	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func init() {
	// Ensure data directory exists
	dataDir := os.Getenv("DB_PATH")
	if dataDir == "" {
		dataDir = "./data"
	}
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}
}
