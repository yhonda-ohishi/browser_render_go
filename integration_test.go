// +build integration

package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestIntegration_HealthEndpoint(t *testing.T) {
	// Wait for server to be ready
	time.Sleep(2 * time.Second)

	resp, err := http.Get("http://localhost:8080/health")
	if err != nil {
		t.Fatalf("Failed to call health endpoint: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if result["status"] != "healthy" {
		t.Errorf("Expected status to be 'healthy', got %v", result["status"])
	}
}

func TestIntegration_MetricsEndpoint(t *testing.T) {
	resp, err := http.Get("http://localhost:8080/metrics")
	if err != nil {
		t.Fatalf("Failed to call metrics endpoint: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if result["uptime_seconds"] == nil {
		t.Error("Expected uptime_seconds in response")
	}
}

func TestIntegration_VehicleDataEndpoint(t *testing.T) {
	reqBody := map[string]interface{}{
		"branch_id":    "",
		"filter_id":    "0",
		"force_login": false,
	}

	jsonData, _ := json.Marshal(reqBody)
	resp, err := http.Post("http://localhost:8080/v1/vehicle/data", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("Failed to call vehicle data endpoint: %v", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if result["status"] == nil {
		t.Error("Expected status in response")
	}

	// Check if data is returned
	if data, ok := result["data"].([]interface{}); ok {
		t.Logf("Got %d vehicles", len(data))
	}
}

func TestIntegration_SessionCheck(t *testing.T) {
	resp, err := http.Get("http://localhost:8080/v1/session/check?session_id=test123")
	if err != nil {
		t.Fatalf("Failed to call session check endpoint: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if result["is_valid"] == nil {
		t.Error("Expected is_valid in response")
	}
	if result["message"] == nil {
		t.Error("Expected message in response")
	}
}