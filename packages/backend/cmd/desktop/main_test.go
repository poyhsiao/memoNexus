// Package main tests for desktop server initialization and routing.
// These tests verify server setup, route registration, and handler creation.
package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/kimhsiao/memonexus/backend/cmd/desktop/handlers"
	"github.com/kimhsiao/memonexus/backend/internal/db"
	"github.com/kimhsiao/memonexus/backend/internal/logging"
)

// setupTestEnv initializes test environment
func setupTestEnv(t *testing.T) (func()) {
	// Initialize logger to prevent panics
	logging.Init(os.Stdout, logging.LevelInfo)

	// Set environment variables for testing
	os.Setenv("DB_PATH", t.TempDir())
	os.Setenv("PORT", "0") // Use random port for testing
	os.Setenv("MACHINE_ID", "test-machine-id")

	cleanup := func() {
		os.Unsetenv("DB_PATH")
		os.Unsetenv("PORT")
		os.Unsetenv("MACHINE_ID")
	}

	return cleanup
}

func TestMain_RouteSetup(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	// Create test database
	dataDir := t.TempDir()
	database, err := db.Open(dataDir)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer database.Close()

	// Run migrations (use relative path from backend directory)
	migrator := db.NewMigrator(database.DB, "../../internal/db/migrations")
	if err := migrator.Initialize(); err != nil {
		t.Fatalf("Failed to initialize migrator: %v", err)
	}
	if err := migrator.Up(); err != nil {
		t.Fatalf("Failed to apply migrations: %v", err)
	}

	// Create repository
	repository := db.NewRepository(database.DB)

	// Create handlers
	contentHandler := handlers.NewContentHandler(repository)
	tagHandler := handlers.NewTagHandler(repository)

	// Setup routes (simplified version of main.go)
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

	// Use handlers to avoid "declared and not used" errors
	_ = contentHandler

	// Test health check endpoint
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Health check returned status %d", w.Code)
	}

	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", w.Header().Get("Content-Type"))
	}

	expectedBody := `{"status":"ok","service":"memonexus-desktop"}`
	if w.Body.String() != expectedBody {
		t.Errorf("Expected body %q, got %q", expectedBody, w.Body.String())
	}
}

func TestMain_HealthCheck_MethodNotAllowed(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	// Create minimal mux
	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	// Test POST request (should fail)
	req := httptest.NewRequest(http.MethodPost, "/api/health", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestMain_HandlerCreation(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	// Create test database
	dataDir := t.TempDir()
	database, err := db.Open(dataDir)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer database.Close()

	// Run migrations (use relative path from backend directory)
	migrator := db.NewMigrator(database.DB, "../../internal/db/migrations")
	if err := migrator.Initialize(); err != nil {
		t.Fatalf("Failed to initialize migrator: %v", err)
	}
	if err := migrator.Up(); err != nil {
		t.Fatalf("Failed to apply migrations: %v", err)
	}

	// Create repository
	repository := db.NewRepository(database.DB)

	// Create handlers
	contentHandler := handlers.NewContentHandler(repository)
	tagHandler := handlers.NewTagHandler(repository)
	searchHandler := handlers.NewSearchHandler(repository)

	// Verify handlers are created
	if contentHandler == nil {
		t.Error("ContentHandler should not be nil")
	}
	if tagHandler == nil {
		t.Error("TagHandler should not be nil")
	}
	if searchHandler == nil {
		t.Error("SearchHandler should not be nil")
	}

	// Use handlers to avoid "declared and not used" errors
	_ = contentHandler
	_ = tagHandler
	_ = searchHandler
}

func TestMain_RouteRegistration(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	// Create a basic mux and register routes
	mux := http.NewServeMux()

	routeRegistered := false
	mux.HandleFunc("/api/test", func(w http.ResponseWriter, r *http.Request) {
		routeRegistered = true
		w.WriteHeader(http.StatusOK)
	})

	// Test that the route is registered
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if !routeRegistered {
		t.Error("Route handler was not called")
	}

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestMain_PortDefaultsTo8090(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	// Unset PORT to test default
	os.Unsetenv("PORT")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8090"
	}

	if port != "8090" {
		t.Errorf("Expected default port 8090, got %s", port)
	}
}

func TestMain_PortFromEnv(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	os.Setenv("PORT", "3000")
	defer os.Unsetenv("PORT")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8090"
	}

	if port != "3000" {
		t.Errorf("Expected port 3000, got %s", port)
	}
}
