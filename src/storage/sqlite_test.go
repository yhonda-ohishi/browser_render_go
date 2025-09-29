package storage

import (
	"encoding/json"
	"path/filepath"
	"testing"
	"time"
)

func setupTestDB(t *testing.T) *Storage {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := NewStorage(dbPath)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	return store
}

func TestNewStorage(t *testing.T) {
	tests := []struct {
		name    string
		dbPath  string
		wantErr bool
	}{
		{
			name:    "Valid path",
			dbPath:  filepath.Join(t.TempDir(), "test.db"),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, err := NewStorage(tt.dbPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewStorage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if store != nil {
				defer store.Close()
			}
		})
	}
}

func TestStorage_Sessions(t *testing.T) {
	store := setupTestDB(t)
	defer store.Close()

	// Test CreateSession
	session := &Session{
		ID:        "test-session-1",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		ExpiresAt: time.Now().Add(10 * time.Minute),
		UserID:    "test-user",
		CompanyID: "test-company",
	}

	err := store.CreateSession(session)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Test GetSession
	retrievedSession, err := store.GetSession(session.ID)
	if err != nil {
		t.Fatalf("Failed to get session: %v", err)
	}

	if retrievedSession.ID != session.ID {
		t.Errorf("Expected session ID %s, got %s", session.ID, retrievedSession.ID)
	}
	if retrievedSession.UserID != session.UserID {
		t.Errorf("Expected user ID %s, got %s", session.UserID, retrievedSession.UserID)
	}
	if retrievedSession.CompanyID != session.CompanyID {
		t.Errorf("Expected company ID %s, got %s", session.CompanyID, retrievedSession.CompanyID)
	}

	// Test GetSession with non-existing ID
	nonExistingSession, err := store.GetSession("non-existing")
	if err != nil {
		t.Errorf("Expected no error for non-existing session, got %v", err)
	}
	if nonExistingSession != nil {
		t.Error("Expected nil session for non-existing ID")
	}

	// Test DeleteSession
	err = store.DeleteSession(session.ID)
	if err != nil {
		t.Fatalf("Failed to delete session: %v", err)
	}

	deletedSession, err := store.GetSession(session.ID)
	if err != nil {
		t.Errorf("Expected no error after deletion, got %v", err)
	}
	if deletedSession != nil {
		t.Error("Expected nil session after deletion")
	}
}

func TestStorage_Cookies(t *testing.T) {
	store := setupTestDB(t)
	defer store.Close()

	sessionID := "test-session-cookies"
	cookies := []Cookie{
		{
			Name:      "cookie1",
			Value:     "value1",
			Domain:    ".example.com",
			Path:      "/",
			ExpiresAt: time.Now().Add(24 * time.Hour),
			HTTPOnly:  true,
			Secure:    true,
		},
		{
			Name:      "cookie2",
			Value:     "value2",
			Domain:    ".example.com",
			Path:      "/api",
			ExpiresAt: time.Now().Add(48 * time.Hour),
			HTTPOnly:  false,
			Secure:    true,
		},
	}

	// Test SaveCookies
	err := store.SaveCookies(sessionID, cookies)
	if err != nil {
		t.Fatalf("Failed to save cookies: %v", err)
	}

	// Test GetCookies
	retrievedCookies, err := store.GetCookies(sessionID)
	if err != nil {
		t.Fatalf("Failed to get cookies: %v", err)
	}

	if len(retrievedCookies) != len(cookies) {
		t.Errorf("Expected %d cookies, got %d", len(cookies), len(retrievedCookies))
	}

	// Verify cookie values
	cookieMap := make(map[string]Cookie)
	for _, c := range retrievedCookies {
		cookieMap[c.Name] = c
	}

	for _, expectedCookie := range cookies {
		actualCookie, exists := cookieMap[expectedCookie.Name]
		if !exists {
			t.Errorf("Cookie %s not found", expectedCookie.Name)
			continue
		}

		if actualCookie.Value != expectedCookie.Value {
			t.Errorf("Cookie %s: expected value %s, got %s",
				expectedCookie.Name, expectedCookie.Value, actualCookie.Value)
		}
		if actualCookie.Domain != expectedCookie.Domain {
			t.Errorf("Cookie %s: expected domain %s, got %s",
				expectedCookie.Name, expectedCookie.Domain, actualCookie.Domain)
		}
	}

	// Test SaveCookies update (replaces cookies)
	updatedCookies := []Cookie{
		{
			Name:  "cookie1",
			Value: "updated_value",
		},
		{
			Name:  "cookie3",
			Value: "new_cookie",
		},
	}

	err = store.SaveCookies(sessionID, updatedCookies)
	if err != nil {
		t.Fatalf("Failed to save updated cookies: %v", err)
	}

	// Retrieve and verify
	retrievedUpdated, _ := store.GetCookies(sessionID)
	if len(retrievedUpdated) != len(updatedCookies) {
		t.Errorf("Expected %d cookies after update, got %d", len(updatedCookies), len(retrievedUpdated))
	}
}

func TestStorage_VehicleCache(t *testing.T) {
	store := setupTestDB(t)
	defer store.Close()

	vehicleData := map[string]interface{}{
		"VehicleCD":   "TEST001",
		"VehicleName": "Test Vehicle",
		"Status":      "Active",
	}

	// Test CacheVehicleData
	err := store.CacheVehicleData("TEST001", vehicleData, 5*time.Minute)
	if err != nil {
		t.Fatalf("Failed to cache vehicle data: %v", err)
	}

	// Test GetCachedVehicleData
	retrievedData, err := store.GetCachedVehicleData("TEST001")
	if err != nil {
		t.Fatalf("Failed to get cached vehicle data: %v", err)
	}

	// Parse JSON string
	var parsedData map[string]interface{}
	if err := json.Unmarshal([]byte(retrievedData), &parsedData); err != nil {
		t.Fatalf("Failed to parse cached data: %v", err)
	}

	if parsedData["VehicleCD"] != vehicleData["VehicleCD"] {
		t.Errorf("Expected VehicleCD %s, got %v", vehicleData["VehicleCD"], parsedData["VehicleCD"])
	}

	// Test GetCachedVehicleData with non-existing vehicle
	nonExistingData, err := store.GetCachedVehicleData("NON-EXISTING")
	if err != nil {
		t.Error("Should not error for non-existing vehicle")
	}
	if nonExistingData != "" {
		t.Error("Expected empty string for non-existing vehicle")
	}

	// Test cache update
	updatedData := map[string]interface{}{
		"VehicleCD":   "TEST001",
		"VehicleName": "Updated Vehicle",
		"Status":      "Inactive",
	}

	err = store.CacheVehicleData("TEST001", updatedData, 10*time.Minute)
	if err != nil {
		t.Fatalf("Failed to update cached vehicle data: %v", err)
	}
}

func TestStorage_SetGetDelete(t *testing.T) {
	store := setupTestDB(t)
	defer store.Close()

	// Test Set and Get
	testData := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	err := store.Set("test-key", testData)
	if err != nil {
		t.Fatalf("Failed to set value: %v", err)
	}

	var retrievedData map[string]string
	err = store.Get("test-key", &retrievedData)
	if err != nil {
		t.Fatalf("Failed to get value: %v", err)
	}

	if retrievedData["key1"] != testData["key1"] {
		t.Errorf("Expected key1 to be %s, got %s", testData["key1"], retrievedData["key1"])
	}

	// Test Delete
	err = store.Delete("test-key")
	if err != nil {
		t.Fatalf("Failed to delete value: %v", err)
	}

	// Try to get deleted value
	var deletedData map[string]string
	err = store.Get("test-key", &deletedData)
	if err == nil {
		t.Error("Expected error when getting deleted key")
	}
}

func TestStorage_CleanupExpired(t *testing.T) {
	store := setupTestDB(t)
	defer store.Close()

	// Create expired session
	expiredSession := &Session{
		ID:        "expired-session",
		CreatedAt: time.Now().Add(-2 * time.Hour),
		UpdatedAt: time.Now().Add(-1 * time.Hour),
		ExpiresAt: time.Now().Add(-30 * time.Minute),
		UserID:    "test-user",
		CompanyID: "test-company",
	}
	store.CreateSession(expiredSession)

	// Create valid session
	validSession := &Session{
		ID:        "valid-session",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		ExpiresAt: time.Now().Add(30 * time.Minute),
		UserID:    "test-user",
		CompanyID: "test-company",
	}
	store.CreateSession(validSession)

	// Clean expired
	err := store.CleanupExpired()
	if err != nil {
		t.Fatalf("Failed to cleanup expired: %v", err)
	}

	// Check that expired session is deleted (GetSession already filters expired sessions)
	expiredResult, err := store.GetSession(expiredSession.ID)
	if err != nil {
		t.Errorf("Expected no error for expired session, got %v", err)
	}
	if expiredResult != nil {
		t.Error("Expired session should return nil due to expiration check")
	}

	// Check that valid session still exists
	retrieved, err := store.GetSession(validSession.ID)
	if err != nil {
		t.Error("Valid session should still exist")
	}
	if retrieved == nil {
		t.Error("Valid session should not be nil")
	}
}

func TestStorage_Close(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_close.db")

	store, err := NewStorage(dbPath)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	// Close the storage
	err = store.Close()
	if err != nil {
		t.Errorf("Failed to close storage: %v", err)
	}

	// Try to use after closing - should fail
	err = store.CreateSession(&Session{ID: "test"})
	if err == nil {
		t.Error("Expected error when using closed database")
	}
}