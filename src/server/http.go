package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/yhonda-ohishi/browser_render_go/src/browser"
	"github.com/yhonda-ohishi/browser_render_go/src/config"
	"github.com/yhonda-ohishi/browser_render_go/src/storage"
)

// HTTPServer handles HTTP requests
type HTTPServer struct {
	config    *config.Config
	storage   *storage.Storage
	renderer  *browser.Renderer
	startTime time.Time
	mux       *http.ServeMux
}

// NewHTTPServer creates a new HTTP server instance
func NewHTTPServer(cfg *config.Config, store *storage.Storage, renderer *browser.Renderer) *HTTPServer {
	s := &HTTPServer{
		config:    cfg,
		storage:   store,
		renderer:  renderer,
		startTime: time.Now(),
		mux:       http.NewServeMux(),
	}
	s.setupRoutes()
	return s
}

func (s *HTTPServer) setupRoutes() {
	// API endpoints
	s.mux.HandleFunc("/v1/vehicle/data", s.handleVehicleData)
	s.mux.HandleFunc("/v1/session/check", s.handleSessionCheck)
	s.mux.HandleFunc("/v1/session/clear", s.handleSessionClear)

	// Health and metrics
	s.mux.HandleFunc("/health", s.handleHealth)
	s.mux.HandleFunc("/metrics", s.handleMetrics)

	// CORS middleware wrapper
	s.mux.HandleFunc("/", s.corsMiddleware(s.notFound))
}

// CORS middleware
func (s *HTTPServer) corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

// Vehicle data endpoint
func (s *HTTPServer) handleVehicleData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Fixed values
	branchID := "00000000"
	filterID := "0"
	forceLogin := false

	// Start background processing
	go func() {
		log.Printf("Starting background vehicle data retrieval...")

		// Create a context with timeout for the background operation
		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		defer cancel()

		vehicleData, sessionID, honoAPIResponse, err := s.renderer.GetVehicleData(
			ctx,
			"", // Session ID from cookie if needed
			branchID,
			filterID,
			forceLogin,
		)

		if err != nil {
			log.Printf("Error getting vehicle data in background: %v", err)
			return
		}

		log.Printf("Successfully retrieved %d vehicles in background", len(vehicleData))
		log.Printf("Session ID: %s", sessionID)

		if honoAPIResponse != nil {
			log.Printf("Hono API Response - Success: %v, Records Added: %d, Total: %d",
				honoAPIResponse.Success,
				honoAPIResponse.RecordsAdded,
				honoAPIResponse.TotalRecords)
		}
	}()

	// Return immediately
	s.sendJSON(w, map[string]interface{}{
		"status": "processing",
		"message": "Vehicle data retrieval started in background",
	}, http.StatusAccepted)
}

// Session check endpoint
func (s *HTTPServer) handleSessionCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		s.sendJSON(w, map[string]interface{}{
			"is_valid": false,
			"message":  "Session ID is required",
		}, http.StatusBadRequest)
		return
	}

	isValid, message := s.renderer.CheckSession(sessionID)
	s.sendJSON(w, map[string]interface{}{
		"is_valid": isValid,
		"message":  message,
	}, http.StatusOK)
}

// Session clear endpoint
func (s *HTTPServer) handleSessionClear(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		s.sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		s.sendJSON(w, map[string]interface{}{
			"success": false,
			"message": "Session ID is required",
		}, http.StatusBadRequest)
		return
	}

	err := s.renderer.ClearSession(sessionID)
	if err != nil {
		s.sendJSON(w, map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Failed to clear session: %v", err),
		}, http.StatusInternalServerError)
		return
	}

	s.sendJSON(w, map[string]interface{}{
		"success": true,
		"message": "Session cleared successfully",
	}, http.StatusOK)
}

// Health check endpoint
func (s *HTTPServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	uptime := time.Since(s.startTime).Seconds()
	s.sendJSON(w, map[string]interface{}{
		"status":  "healthy",
		"version": "1.0.0",
		"uptime":  uptime,
	}, http.StatusOK)
}

// Metrics endpoint
func (s *HTTPServer) handleMetrics(w http.ResponseWriter, r *http.Request) {
	// Simple metrics - can be extended with Prometheus later
	stats := map[string]interface{}{
		"uptime_seconds": time.Since(s.startTime).Seconds(),
		"timestamp":      time.Now().Unix(),
	}

	s.sendJSON(w, stats, http.StatusOK)
}

// Not found handler
func (s *HTTPServer) notFound(w http.ResponseWriter, r *http.Request) {
	s.sendError(w, "Endpoint not found", http.StatusNotFound)
}

// Helper functions
func (s *HTTPServer) sendJSON(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}
}

func (s *HTTPServer) sendError(w http.ResponseWriter, message string, status int) {
	s.sendJSON(w, map[string]interface{}{
		"error": message,
	}, status)
}

// Start starts the HTTP server
func (s *HTTPServer) Start(address string) error {
	server := &http.Server{
		Addr:         address,
		Handler:      s.mux,
		ReadTimeout:  120 * time.Second,
		WriteTimeout: 120 * time.Second,
		IdleTimeout:  180 * time.Second,
	}

	log.Printf("HTTP server starting on %s", address)
	return server.ListenAndServe()
}