package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/yourusername/browser_render_go/src/browser"
	"github.com/yourusername/browser_render_go/src/config"
	"github.com/yourusername/browser_render_go/src/storage"
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
	if r.Method != http.MethodPost {
		s.sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		BranchID   string `json:"branch_id"`
		FilterID   string `json:"filter_id"`
		ForceLogin bool   `json:"force_login"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Default values
	if req.BranchID == "" {
		req.BranchID = "00000000"
	}
	if req.FilterID == "" {
		req.FilterID = "0"
	}

	// Get vehicle data - use extended timeout for browser operations
	ctx, cancel := context.WithTimeout(r.Context(), 90*time.Second)
	defer cancel()

	vehicleData, sessionID, err := s.renderer.GetVehicleData(
		ctx,
		"", // Session ID from cookie if needed
		req.BranchID,
		req.FilterID,
		req.ForceLogin,
	)
	if err != nil {
		log.Printf("Error getting vehicle data: %v", err)
		s.sendJSON(w, map[string]interface{}{
			"status": err.Error(),
			"data":   []interface{}{},
		}, http.StatusInternalServerError)
		return
	}

	// Convert to response format
	data := make([]map[string]interface{}, len(vehicleData))
	for i, v := range vehicleData {
		item := map[string]interface{}{
			"VehicleCD":   v.VehicleCD,
			"VehicleName": v.VehicleName,
			"Status":      v.Status,
		}
		// Add metadata
		for k, val := range v.Metadata {
			item[k] = val
		}
		data[i] = item
	}

	s.sendJSON(w, map[string]interface{}{
		"status":     "complete",
		"data":       data,
		"session_id": sessionID,
	}, http.StatusOK)
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