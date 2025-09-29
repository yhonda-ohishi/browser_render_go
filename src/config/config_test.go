package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	// Save original env vars
	originalEnvs := map[string]string{
		"GRPC_PORT":        os.Getenv("GRPC_PORT"),
		"HTTP_PORT":        os.Getenv("HTTP_PORT"),
		"USER_NAME":        os.Getenv("USER_NAME"),
		"COMP_ID":          os.Getenv("COMP_ID"),
		"USER_PASS":        os.Getenv("USER_PASS"),
		"BROWSER_HEADLESS": os.Getenv("BROWSER_HEADLESS"),
		"BROWSER_TIMEOUT":  os.Getenv("BROWSER_TIMEOUT"),
		"BROWSER_DEBUG":    os.Getenv("BROWSER_DEBUG"),
		"SQLITE_PATH":      os.Getenv("SQLITE_PATH"),
		"SESSION_TTL":      os.Getenv("SESSION_TTL"),
		"COOKIE_TTL":       os.Getenv("COOKIE_TTL"),
	}

	// Restore env vars after test
	defer func() {
		for k, v := range originalEnvs {
			if v != "" {
				os.Setenv(k, v)
			} else {
				os.Unsetenv(k)
			}
		}
	}()

	tests := []struct {
		name     string
		envVars  map[string]string
		validate func(*testing.T, *Config)
	}{
		{
			name:    "Default values",
			envVars: map[string]string{},
			validate: func(t *testing.T, cfg *Config) {
				if cfg.GRPCPort != "50051" {
					t.Errorf("Expected GRPCPort to be 50051, got %s", cfg.GRPCPort)
				}
				if cfg.HTTPPort != "8080" {
					t.Errorf("Expected HTTPPort to be 8080, got %s", cfg.HTTPPort)
				}
				if !cfg.BrowserHeadless {
					t.Errorf("Expected BrowserHeadless to be true")
				}
				if cfg.BrowserTimeout != 60*time.Second {
					t.Errorf("Expected BrowserTimeout to be 60s, got %v", cfg.BrowserTimeout)
				}
				if cfg.BrowserDebug {
					t.Errorf("Expected BrowserDebug to be false")
				}
				if cfg.SQLitePath != "./data/browser_render.db" {
					t.Errorf("Expected SQLitePath to be ./data/browser_render.db, got %s", cfg.SQLitePath)
				}
				if cfg.SessionTTL != 10*time.Minute {
					t.Errorf("Expected SessionTTL to be 10m, got %v", cfg.SessionTTL)
				}
				if cfg.CookieTTL != 24*time.Hour {
					t.Errorf("Expected CookieTTL to be 24h, got %v", cfg.CookieTTL)
				}
			},
		},
		{
			name: "Custom values",
			envVars: map[string]string{
				"GRPC_PORT":        "9090",
				"HTTP_PORT":        "3000",
				"USER_NAME":        "testuser",
				"COMP_ID":          "12345",
				"USER_PASS":        "testpass",
				"BROWSER_HEADLESS": "false",
				"BROWSER_TIMEOUT":  "120s",
				"BROWSER_DEBUG":    "true",
				"SQLITE_PATH":      "/tmp/test.db",
				"SESSION_TTL":      "20m",
				"COOKIE_TTL":       "48h",
			},
			validate: func(t *testing.T, cfg *Config) {
				if cfg.GRPCPort != "9090" {
					t.Errorf("Expected GRPCPort to be 9090, got %s", cfg.GRPCPort)
				}
				if cfg.HTTPPort != "3000" {
					t.Errorf("Expected HTTPPort to be 3000, got %s", cfg.HTTPPort)
				}
				if cfg.UserName != "testuser" {
					t.Errorf("Expected UserName to be testuser, got %s", cfg.UserName)
				}
				if cfg.CompID != "12345" {
					t.Errorf("Expected CompID to be 12345, got %s", cfg.CompID)
				}
				if cfg.UserPass != "testpass" {
					t.Errorf("Expected UserPass to be testpass, got %s", cfg.UserPass)
				}
				if cfg.BrowserHeadless {
					t.Errorf("Expected BrowserHeadless to be false")
				}
				if cfg.BrowserTimeout != 120*time.Second {
					t.Errorf("Expected BrowserTimeout to be 120s, got %v", cfg.BrowserTimeout)
				}
				if !cfg.BrowserDebug {
					t.Errorf("Expected BrowserDebug to be true")
				}
				if cfg.SQLitePath != "/tmp/test.db" {
					t.Errorf("Expected SQLitePath to be /tmp/test.db, got %s", cfg.SQLitePath)
				}
				if cfg.SessionTTL != 20*time.Minute {
					t.Errorf("Expected SessionTTL to be 20m, got %v", cfg.SessionTTL)
				}
				if cfg.CookieTTL != 48*time.Hour {
					t.Errorf("Expected CookieTTL to be 48h, got %v", cfg.CookieTTL)
				}
			},
		},
		{
			name: "Duration as milliseconds",
			envVars: map[string]string{
				"BROWSER_TIMEOUT": "5000",
				"SESSION_TTL":     "600000",
				"COOKIE_TTL":      "86400000",
			},
			validate: func(t *testing.T, cfg *Config) {
				if cfg.BrowserTimeout != 5*time.Second {
					t.Errorf("Expected BrowserTimeout to be 5s, got %v", cfg.BrowserTimeout)
				}
				if cfg.SessionTTL != 10*time.Minute {
					t.Errorf("Expected SessionTTL to be 10m, got %v", cfg.SessionTTL)
				}
				if cfg.CookieTTL != 24*time.Hour {
					t.Errorf("Expected CookieTTL to be 24h, got %v", cfg.CookieTTL)
				}
			},
		},
		{
			name: "Invalid boolean values",
			envVars: map[string]string{
				"BROWSER_HEADLESS": "invalid",
				"BROWSER_DEBUG":    "not-a-bool",
			},
			validate: func(t *testing.T, cfg *Config) {
				// Should use defaults for invalid values
				if !cfg.BrowserHeadless {
					t.Errorf("Expected BrowserHeadless to be true (default)")
				}
				if cfg.BrowserDebug {
					t.Errorf("Expected BrowserDebug to be false (default)")
				}
			},
		},
		{
			name: "Invalid duration values",
			envVars: map[string]string{
				"BROWSER_TIMEOUT": "invalid",
				"SESSION_TTL":     "not-a-duration",
				"COOKIE_TTL":      "abc",
			},
			validate: func(t *testing.T, cfg *Config) {
				// Should use defaults for invalid values
				if cfg.BrowserTimeout != 60*time.Second {
					t.Errorf("Expected BrowserTimeout to be 60s (default), got %v", cfg.BrowserTimeout)
				}
				if cfg.SessionTTL != 10*time.Minute {
					t.Errorf("Expected SessionTTL to be 10m (default), got %v", cfg.SessionTTL)
				}
				if cfg.CookieTTL != 24*time.Hour {
					t.Errorf("Expected CookieTTL to be 24h (default), got %v", cfg.CookieTTL)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all env vars
			for k := range originalEnvs {
				os.Unsetenv(k)
			}

			// Set test env vars
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			cfg := Load()
			tt.validate(t, cfg)
		})
	}
}

func TestLoadEnvFile(t *testing.T) {
	// Create a temporary .env file
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	content := `GRPC_PORT=9999
HTTP_PORT=8888
USER_NAME=envuser
COMP_ID=99999
USER_PASS=envpass
`
	if err := os.WriteFile(envFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create .env file: %v", err)
	}

	// Change to temp directory
	originalWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalWd)

	// Clear env vars
	os.Unsetenv("GRPC_PORT")
	os.Unsetenv("HTTP_PORT")
	os.Unsetenv("USER_NAME")
	os.Unsetenv("COMP_ID")
	os.Unsetenv("USER_PASS")

	cfg := Load()

	if cfg.GRPCPort != "9999" {
		t.Errorf("Expected GRPCPort to be 9999, got %s", cfg.GRPCPort)
	}
	if cfg.HTTPPort != "8888" {
		t.Errorf("Expected HTTPPort to be 8888, got %s", cfg.HTTPPort)
	}
	if cfg.UserName != "envuser" {
		t.Errorf("Expected UserName to be envuser, got %s", cfg.UserName)
	}
	if cfg.CompID != "99999" {
		t.Errorf("Expected CompID to be 99999, got %s", cfg.CompID)
	}
	if cfg.UserPass != "envpass" {
		t.Errorf("Expected UserPass to be envpass, got %s", cfg.UserPass)
	}
}

func TestGetCurrentDir(t *testing.T) {
	dir := getCurrentDir()
	if dir == "" {
		t.Error("getCurrentDir should not return empty string")
	}

	// Test that it returns current directory
	expected, _ := os.Getwd()
	if dir != expected {
		t.Errorf("Expected getCurrentDir to return %s, got %s", expected, dir)
	}
}

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		value        string
		defaultValue string
		expected     string
	}{
		{
			name:         "Existing env var",
			key:          "TEST_VAR",
			value:        "test_value",
			defaultValue: "default",
			expected:     "test_value",
		},
		{
			name:         "Non-existing env var",
			key:          "NON_EXISTING_VAR",
			value:        "",
			defaultValue: "default",
			expected:     "default",
		},
		{
			name:         "Empty env var",
			key:          "EMPTY_VAR",
			value:        "",
			defaultValue: "default",
			expected:     "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != "" {
				os.Setenv(tt.key, tt.value)
				defer os.Unsetenv(tt.key)
			}

			result := getEnv(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestGetEnvBool(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		value        string
		defaultValue bool
		expected     bool
	}{
		{
			name:         "True string",
			key:          "BOOL_TRUE",
			value:        "true",
			defaultValue: false,
			expected:     true,
		},
		{
			name:         "False string",
			key:          "BOOL_FALSE",
			value:        "false",
			defaultValue: true,
			expected:     false,
		},
		{
			name:         "1 as true",
			key:          "BOOL_1",
			value:        "1",
			defaultValue: false,
			expected:     true,
		},
		{
			name:         "0 as false",
			key:          "BOOL_0",
			value:        "0",
			defaultValue: true,
			expected:     false,
		},
		{
			name:         "Invalid bool",
			key:          "BOOL_INVALID",
			value:        "invalid",
			defaultValue: true,
			expected:     true,
		},
		{
			name:         "Non-existing var",
			key:          "BOOL_NON_EXISTING",
			value:        "",
			defaultValue: false,
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != "" {
				os.Setenv(tt.key, tt.value)
				defer os.Unsetenv(tt.key)
			}

			result := getEnvBool(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestGetEnvDuration(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		value        string
		defaultValue time.Duration
		expected     time.Duration
	}{
		{
			name:         "Valid duration string",
			key:          "DUR_VALID",
			value:        "10s",
			defaultValue: 5 * time.Second,
			expected:     10 * time.Second,
		},
		{
			name:         "Milliseconds as number",
			key:          "DUR_MS",
			value:        "5000",
			defaultValue: 1 * time.Second,
			expected:     5 * time.Second,
		},
		{
			name:         "Invalid duration",
			key:          "DUR_INVALID",
			value:        "invalid",
			defaultValue: 30 * time.Second,
			expected:     30 * time.Second,
		},
		{
			name:         "Non-existing var",
			key:          "DUR_NON_EXISTING",
			value:        "",
			defaultValue: 60 * time.Second,
			expected:     60 * time.Second,
		},
		{
			name:         "Complex duration",
			key:          "DUR_COMPLEX",
			value:        "1h30m",
			defaultValue: 1 * time.Hour,
			expected:     90 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != "" {
				os.Setenv(tt.key, tt.value)
				defer os.Unsetenv(tt.key)
			}

			result := getEnvDuration(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}