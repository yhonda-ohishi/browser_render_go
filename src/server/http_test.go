package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yhonda-ohishi/browser_render_go/src/browser"
	"github.com/yhonda-ohishi/browser_render_go/src/config"
	"github.com/yhonda-ohishi/browser_render_go/src/storage"
)

func setupTestServer(t *testing.T) (*HTTPServer, func()) {
	// Create test config
	cfg := &config.Config{
		HTTPPort:        "8080",
		BrowserHeadless: true,
		BrowserTimeout:  10 * time.Second,
		SessionTTL:      10 * time.Minute,
	}

	// Create test storage
	tmpDir := t.TempDir()
	dbPath := tmpDir + "/test.db"
	store, err := storage.NewStorage(dbPath)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	// Create test renderer (will be nil for most tests)
	// In real tests, we'd use the actual browser.Renderer with mocked browser
	var renderer *browser.Renderer

	cleanup := func() {
		store.Close()
	}

	return NewHTTPServer(cfg, store, renderer), cleanup
}

func TestHTTPServer_Health(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	server.handleHealth(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)

	if body["status"] != "healthy" {
		t.Errorf("Expected status 'healthy', got %v", body["status"])
	}
}

func TestHTTPServer_Metrics(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()

	server.handleMetrics(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)

	if body["uptime_seconds"] == nil {
		t.Error("Expected uptime_seconds in metrics")
	}
	if body["timestamp"] == nil {
		t.Error("Expected timestamp in metrics")
	}
}

func TestHTTPServer_SessionCheck(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	// Test without session_id parameter
	req := httptest.NewRequest("GET", "/v1/session/check", nil)
	w := httptest.NewRecorder()
	server.handleSessionCheck(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}

	// We can't test with session_id when renderer is nil as it will panic
	// The server should check for nil renderer but currently doesn't
}

func TestHTTPServer_SessionClear(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	// Test without session_id parameter
	req := httptest.NewRequest("DELETE", "/v1/session/clear", nil)
	w := httptest.NewRecorder()
	server.handleSessionClear(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}

	// We can't test with session_id when renderer is nil as it will panic
	// The server should check for nil renderer but currently doesn't
}

func TestHTTPServer_VehicleData(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	// Test with invalid JSON
	req := httptest.NewRequest("POST", "/v1/vehicle/data", bytes.NewBufferString("invalid json"))
	w := httptest.NewRecorder()
	server.handleVehicleData(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400 for invalid JSON, got %d", resp.StatusCode)
	}

	// We can't test with valid JSON when renderer is nil as it will panic
	// The server should check for nil renderer but currently doesn't
}

func TestHTTPServer_CORSMiddleware(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	// Test OPTIONS request
	req := httptest.NewRequest("OPTIONS", "/v1/vehicle/data", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()

	handler := server.corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 for OPTIONS, got %d", resp.StatusCode)
	}

	if resp.Header.Get("Access-Control-Allow-Origin") != "*" {
		t.Error("Expected CORS header Access-Control-Allow-Origin")
	}
	if resp.Header.Get("Access-Control-Allow-Methods") == "" {
		t.Error("Expected CORS header Access-Control-Allow-Methods")
	}
}

func TestHTTPServer_NotFound(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/unknown/endpoint", nil)
	w := httptest.NewRecorder()

	server.notFound(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", resp.StatusCode)
	}
}

func TestHTTPServer_Start(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	// Create a test server
	ts := httptest.NewServer(server.mux)
	defer ts.Close()

	// Test that the server responds
	resp, err := http.Get(ts.URL + "/health")
	if err != nil {
		t.Fatalf("Failed to call health endpoint: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

// Test error scenarios and edge cases
func TestHTTPServer_ErrorResponses(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		endpoint   string
		body       string
		wantStatus int
	}{
		{
			name:       "Invalid method for vehicle data",
			method:     "GET",
			endpoint:   "/v1/vehicle/data",
			body:       "",
			wantStatus: http.StatusMethodNotAllowed,
		},
		{
			name:       "Invalid method for session clear",
			method:     "GET",
			endpoint:   "/v1/session/clear",
			body:       "",
			wantStatus: http.StatusMethodNotAllowed,
		},
		{
			name:       "Empty body for vehicle data",
			method:     "POST",
			endpoint:   "/v1/vehicle/data",
			body:       "",
			wantStatus: http.StatusBadRequest,
		},
	}

	server, cleanup := setupTestServer(t)
	defer cleanup()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.body != "" {
				req = httptest.NewRequest(tt.method, tt.endpoint, bytes.NewBufferString(tt.body))
			} else {
				req = httptest.NewRequest(tt.method, tt.endpoint, nil)
			}
			w := httptest.NewRecorder()

			// Route the request through the mux
			server.mux.ServeHTTP(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, resp.StatusCode)
			}
		})
	}
}