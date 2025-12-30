// Package main provides the embedded PocketBase server for desktop platforms.
// Desktop clients communicate via REST/WebSocket on localhost:8090.
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/kimhsiao/memonexus/backend/internal/db"
	"github.com/kimhsiao/memonexus/backend/internal/logging"
	"github.com/kimhsiao/memonexus/backend/internal/services"
	"github.com/kimhsiao/memonexus/backend/cmd/desktop/handlers"
)

func main() {
	// Initialize logger
	logging.Init(os.Stdout, logging.LevelInfo)
	logging.Info("MemoNexus Desktop Server starting...")

	// Get data directory from environment or use default
	dataDir := os.Getenv("DB_PATH")
	if dataDir == "" {
		dataDir = "./data"
	}

	// Ensure data directory exists
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

	// Open database
	database, err := db.Open(dataDir)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer database.Close()

	logging.Info("Database opened", map[string]interface{}{"path": dataDir})

	// Run migrations
	migrator := db.NewMigrator(database.DB, "./internal/db/migrations")
	if err := migrator.Initialize(); err != nil {
		log.Fatalf("Failed to initialize migrator: %v", err)
	}

	if err := migrator.Up(); err != nil {
		log.Fatalf("Failed to apply migrations: %v", err)
	}

	currentVersion, _ := migrator.CurrentVersion()
	logging.Info("Migrations applied", map[string]interface{}{"version": currentVersion})

	// Create repository
	repository := db.NewRepository(database.DB)

	// Create analysis service
	analysisService := services.NewAnalysisService(services.DefaultAnalysisConfig())

	// Create handlers
	contentHandler := handlers.NewContentHandler(repository)
	tagHandler := handlers.NewTagHandler(repository)
	searchHandler := handlers.NewSearchHandler(repository)
	aiHandler := handlers.NewAIHandler(repository, analysisService, os.Getenv("MACHINE_ID"))

	// Create WebSocket hub
	wsHub := NewWSHub()

	// Setup routes
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","service":"memonexus-desktop"}`))
	})

	// Content routes
	mux.HandleFunc("/api/content", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			contentHandler.ListContentItems(w, r)
		case http.MethodPost:
			contentHandler.CreateContentItem(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/content/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			contentHandler.GetContentItem(w, r)
		case http.MethodPut:
			contentHandler.UpdateContentItem(w, r)
		case http.MethodDelete:
			contentHandler.DeleteContentItem(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Tag routes
	mux.HandleFunc("/api/tags", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			tagHandler.ListTags(w, r)
		case http.MethodPost:
			tagHandler.CreateTag(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/tags/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			tagHandler.GetTag(w, r)
		case http.MethodPut:
			tagHandler.UpdateTag(w, r)
		case http.MethodDelete:
			tagHandler.DeleteTag(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Search route
	mux.HandleFunc("/api/search", func(w http.ResponseWriter, r *http.Request) {
		searchHandler.Search(w, r)
	})

	// AI configuration routes (T137-T139)
	mux.HandleFunc("/api/ai/config", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			aiHandler.GetAIConfig(w, r)
		case http.MethodPost:
			aiHandler.SetAIConfig(w, r)
		case http.MethodDelete:
			aiHandler.DeleteAIConfig(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Content analysis routes (T140)
	mux.HandleFunc("/api/content/analyze", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			// Determine operation from query parameter
			operation := r.URL.Query().Get("operation")
			switch operation {
			case "summary":
				aiHandler.GenerateSummary(w, r)
			case "keywords":
				aiHandler.ExtractKeywords(w, r)
			default:
				http.Error(w, "Invalid operation: use 'summary' or 'keywords'", http.StatusBadRequest)
			}
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// WebSocket route
	mux.HandleFunc("/api/realtime", HandleWebSocket(wsHub))

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8090"
	}

	logging.Info("Server listening", map[string]interface{}{"port": port})
	log.Printf("MemoNexus Desktop Server listening on :%s", port)

	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
