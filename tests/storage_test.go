package tests

import (
	"testing"
	"time"

	"github.com/yourusername/browser_render_go/src/storage"
)

func TestStorage(t *testing.T) {
	// Use in-memory database for testing
	store, err := storage.NewStorage(":memory:")
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer store.Close()

	t.Run("Session Operations", func(t *testing.T) {
		session := &storage.Session{
			ID:        "test_session",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			ExpiresAt: time.Now().Add(1 * time.Hour),
			UserID:    "test_user",
			CompanyID: "test_company",
		}

		// Create session
		err := store.CreateSession(session)
		if err != nil {
			t.Fatalf("Failed to create session: %v", err)
		}

		// Get session
		retrieved, err := store.GetSession("test_session")
		if err != nil {
			t.Fatalf("Failed to get session: %v", err)
		}
		if retrieved == nil {
			t.Fatal("Expected session, got nil")
		}
		if retrieved.UserID != session.UserID {
			t.Errorf("UserID mismatch: expected %s, got %s", session.UserID, retrieved.UserID)
		}

		// Delete session
		err = store.DeleteSession("test_session")
		if err != nil {
			t.Fatalf("Failed to delete session: %v", err)
		}

		// Verify deletion
		retrieved, err = store.GetSession("test_session")
		if err != nil {
			t.Fatalf("Error getting deleted session: %v", err)
		}
		if retrieved != nil {
			t.Error("Expected nil for deleted session")
		}
	})

	t.Run("Cookie Operations", func(t *testing.T) {
		// Create session first
		session := &storage.Session{
			ID:        "cookie_test_session",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			ExpiresAt: time.Now().Add(1 * time.Hour),
		}
		store.CreateSession(session)

		cookies := []storage.Cookie{
			{
				Name:      "test_cookie",
				Value:     "test_value",
				Domain:    "example.com",
				Path:      "/",
				ExpiresAt: time.Now().Add(24 * time.Hour),
				HTTPOnly:  true,
				Secure:    true,
			},
		}

		// Save cookies
		err := store.SaveCookies("cookie_test_session", cookies)
		if err != nil {
			t.Fatalf("Failed to save cookies: %v", err)
		}

		// Get cookies
		retrieved, err := store.GetCookies("cookie_test_session")
		if err != nil {
			t.Fatalf("Failed to get cookies: %v", err)
		}
		if len(retrieved) != 1 {
			t.Fatalf("Expected 1 cookie, got %d", len(retrieved))
		}
		if retrieved[0].Name != "test_cookie" {
			t.Errorf("Cookie name mismatch: expected test_cookie, got %s", retrieved[0].Name)
		}
	})

	t.Run("KV Store Operations", func(t *testing.T) {
		// Set value
		testData := map[string]string{
			"key1": "value1",
			"key2": "value2",
		}
		err := store.Set("test_key", testData)
		if err != nil {
			t.Fatalf("Failed to set value: %v", err)
		}

		// Get value
		var retrieved map[string]string
		err = store.Get("test_key", &retrieved)
		if err != nil {
			t.Fatalf("Failed to get value: %v", err)
		}
		if retrieved["key1"] != testData["key1"] {
			t.Errorf("Value mismatch: expected %s, got %s", testData["key1"], retrieved["key1"])
		}

		// Delete value
		err = store.Delete("test_key")
		if err != nil {
			t.Fatalf("Failed to delete value: %v", err)
		}
	})

	t.Run("Vehicle Cache Operations", func(t *testing.T) {
		vehicleData := map[string]interface{}{
			"VehicleCD":   "TEST001",
			"VehicleName": "Test Vehicle",
			"Status":      "Active",
		}

		// Cache vehicle data
		err := store.CacheVehicleData("TEST001", vehicleData, 5*time.Minute)
		if err != nil {
			t.Fatalf("Failed to cache vehicle data: %v", err)
		}

		// Get cached data
		cached, err := store.GetCachedVehicleData("TEST001")
		if err != nil {
			t.Fatalf("Failed to get cached data: %v", err)
		}
		if cached == "" {
			t.Error("Expected cached data, got empty string")
		}
	})

	t.Run("Cleanup Expired", func(t *testing.T) {
		// Create expired session
		expiredSession := &storage.Session{
			ID:        "expired_session",
			CreatedAt: time.Now().Add(-2 * time.Hour),
			UpdatedAt: time.Now().Add(-2 * time.Hour),
			ExpiresAt: time.Now().Add(-1 * time.Hour), // Already expired
		}
		store.CreateSession(expiredSession)

		// Run cleanup
		err := store.CleanupExpired()
		if err != nil {
			t.Fatalf("Failed to cleanup expired data: %v", err)
		}

		// Verify expired session was deleted
		retrieved, _ := store.GetSession("expired_session")
		if retrieved != nil {
			t.Error("Expected expired session to be deleted")
		}
	})
}