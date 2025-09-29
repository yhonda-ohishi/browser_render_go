package config

import (
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	// Server settings
	GRPCPort string
	HTTPPort string

	// Auth credentials
	UserName string
	CompID   string
	UserPass string

	// Browser settings
	BrowserHeadless bool
	BrowserTimeout  time.Duration
	BrowserDebug    bool

	// Database
	SQLitePath string

	// Session settings
	SessionTTL time.Duration
	CookieTTL  time.Duration
}

func Load() *Config {
	// Load .env file if exists
	loadEnvFile()

	cfg := &Config{
		// Default values
		GRPCPort:        getEnv("GRPC_PORT", "50051"),
		HTTPPort:        getEnv("HTTP_PORT", "8080"),
		UserName:        getEnv("USER_NAME", ""),
		CompID:          getEnv("COMP_ID", ""),
		UserPass:        getEnv("USER_PASS", ""),
		BrowserHeadless: getEnvBool("BROWSER_HEADLESS", true),
		BrowserTimeout:  getEnvDuration("BROWSER_TIMEOUT", 60*time.Second),
		BrowserDebug:    getEnvBool("BROWSER_DEBUG", false),
		SQLitePath:      getEnv("SQLITE_PATH", "./data/browser_render.db"),
		SessionTTL:      getEnvDuration("SESSION_TTL", 10*time.Minute),
		CookieTTL:       getEnvDuration("COOKIE_TTL", 24*time.Hour),
	}

	// Validate required fields
	if cfg.UserName == "" || cfg.CompID == "" || cfg.UserPass == "" {
		log.Println("Warning: Authentication credentials not set in environment variables")
	}

	return cfg
}

func loadEnvFile() {
	// Try to load .env file from multiple possible locations
	possiblePaths := []string{
		".env",
		"../.env",
		filepath.Join(getCurrentDir(), ".env"),
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			if err := godotenv.Load(path); err != nil {
				log.Printf("Warning: Error loading .env file from %s: %v", path, err)
			} else {
				log.Printf("Loaded environment variables from %s", path)
				break
			}
		}
	}
}

func getCurrentDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return "."
	}
	return dir
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if b, err := strconv.ParseBool(value); err == nil {
			return b
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
		// Try parsing as milliseconds
		if ms, err := strconv.ParseInt(value, 10, 64); err == nil {
			return time.Duration(ms) * time.Millisecond
		}
	}
	return defaultValue
}