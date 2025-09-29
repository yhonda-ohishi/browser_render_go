package tests

import (
	"testing"
	"time"

	"github.com/yourusername/browser_render_go/src/browser"
	"github.com/yourusername/browser_render_go/src/config"
	"github.com/yourusername/browser_render_go/src/storage"
)

func TestBrowserRenderer(t *testing.T) {
	// Create test config
	cfg := &config.Config{
		UserName:        "test_user",
		CompID:          "test_company",
		UserPass:        "test_pass",
		BrowserHeadless: true,
		BrowserTimeout:  30 * time.Second,
		BrowserDebug:    false,
		SQLitePath:      ":memory:", // Use in-memory database for testing
		SessionTTL:      10 * time.Minute,
		CookieTTL:       24 * time.Hour,
	}

	// Initialize storage
	store, err := storage.NewStorage(cfg.SQLitePath)
	if err != nil {
		t.Fatalf("Failed to initialize storage: %v", err)
	}
	defer store.Close()

	// Initialize renderer
	renderer, err := browser.NewRenderer(cfg, store)
	if err != nil {
		t.Fatalf("Failed to initialize renderer: %v", err)
	}
	defer renderer.Close()

	t.Run("CheckSession", func(t *testing.T) {
		// Test with empty session ID
		isValid, message := renderer.CheckSession("")
		if isValid {
			t.Error("Expected invalid session for empty ID")
		}
		if message == "" {
			t.Error("Expected error message for empty session ID")
		}

		// Test with non-existent session
		isValid, message = renderer.CheckSession("non_existent_session")
		if isValid {
			t.Error("Expected invalid session for non-existent ID")
		}
	})

	t.Run("ClearSession", func(t *testing.T) {
		// Test clearing non-existent session (should not error)
		err := renderer.ClearSession("non_existent_session")
		if err != nil {
			t.Errorf("Unexpected error clearing non-existent session: %v", err)
		}
	})
}