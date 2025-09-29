package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/yourusername/browser_render_go/src/browser"
	"github.com/yourusername/browser_render_go/src/config"
	"github.com/yourusername/browser_render_go/src/server"
	"github.com/yourusername/browser_render_go/src/storage"
)

func main() {
	// Command line flags
	var (
		grpcPort   = flag.String("grpc-port", "", "gRPC server port (overrides env)")
		httpPort   = flag.String("http-port", "", "HTTP server port (overrides env)")
		dbPath     = flag.String("db-path", "", "SQLite database path (overrides env)")
		headless   = flag.Bool("headless", true, "Run browser in headless mode")
		debugMode  = flag.Bool("debug", false, "Enable debug mode")
		serverType = flag.String("server", "both", "Server type: grpc, http, or both")
	)
	flag.Parse()

	// Load configuration
	cfg := config.Load()

	// Override config with command line flags
	if *grpcPort != "" {
		cfg.GRPCPort = *grpcPort
	}
	if *httpPort != "" {
		cfg.HTTPPort = *httpPort
	}
	if *dbPath != "" {
		cfg.SQLitePath = *dbPath
	}
	cfg.BrowserHeadless = *headless
	cfg.BrowserDebug = *debugMode

	// Setup logging
	if cfg.BrowserDebug {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	} else {
		log.SetFlags(log.LstdFlags)
	}

	log.Println("Starting Browser Render Go Server...")
	log.Printf("Configuration loaded: gRPC=%s, HTTP=%s", cfg.GRPCPort, cfg.HTTPPort)

	// Initialize storage
	store, err := storage.NewStorage(cfg.SQLitePath)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer store.Close()
	log.Println("Storage initialized successfully")

	// Initialize browser renderer
	renderer, err := browser.NewRenderer(cfg, store)
	if err != nil {
		log.Fatalf("Failed to initialize browser renderer: %v", err)
	}
	defer renderer.Close()
	log.Println("Browser renderer initialized successfully")

	// Create servers
	grpcServer := server.NewGRPCServer(cfg, store, renderer)
	httpServer := server.NewHTTPServer(cfg, store, renderer)

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start cleanup goroutine
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := store.CleanupExpired(); err != nil {
					log.Printf("Error cleaning up expired data: %v", err)
				}
			}
		}
	}()

	// Start servers based on server type
	var wg sync.WaitGroup
	errChan := make(chan error, 2)

	switch *serverType {
	case "grpc":
		wg.Add(1)
		go func() {
			defer wg.Done()
			addr := fmt.Sprintf(":%s", cfg.GRPCPort)
			log.Printf("Starting gRPC server on %s", addr)
			if err := grpcServer.Start(addr); err != nil {
				errChan <- fmt.Errorf("gRPC server error: %w", err)
			}
		}()

	case "http":
		wg.Add(1)
		go func() {
			defer wg.Done()
			addr := fmt.Sprintf(":%s", cfg.HTTPPort)
			log.Printf("Starting HTTP server on %s", addr)
			if err := httpServer.Start(addr); err != nil {
				errChan <- fmt.Errorf("HTTP server error: %w", err)
			}
		}()

	case "both":
		// Start gRPC server
		wg.Add(1)
		go func() {
			defer wg.Done()
			addr := fmt.Sprintf(":%s", cfg.GRPCPort)
			log.Printf("Starting gRPC server on %s", addr)
			if err := grpcServer.Start(addr); err != nil {
				errChan <- fmt.Errorf("gRPC server error: %w", err)
			}
		}()

		// Start HTTP server
		wg.Add(1)
		go func() {
			defer wg.Done()
			addr := fmt.Sprintf(":%s", cfg.HTTPPort)
			log.Printf("Starting HTTP server on %s", addr)
			if err := httpServer.Start(addr); err != nil {
				errChan <- fmt.Errorf("HTTP server error: %w", err)
			}
		}()

	default:
		log.Fatalf("Invalid server type: %s (must be grpc, http, or both)", *serverType)
	}

	// Print startup information
	printStartupInfo(cfg, *serverType)

	// Wait for shutdown signal or error
	select {
	case sig := <-sigChan:
		log.Printf("Received signal: %v. Shutting down gracefully...", sig)
		cancel()
	case err := <-errChan:
		log.Printf("Server error: %v. Shutting down...", err)
		cancel()
	}

	// Wait for servers to stop
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	// Wait with timeout
	select {
	case <-done:
		log.Println("All servers stopped successfully")
	case <-time.After(30 * time.Second):
		log.Println("Timeout waiting for servers to stop")
	}

	log.Println("Shutdown complete")
}

func printStartupInfo(cfg *config.Config, serverType string) {
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("🚀 Browser Render Go Server Started")
	fmt.Println(strings.Repeat("=", 50))

	switch serverType {
	case "grpc":
		fmt.Printf("gRPC Server:  http://localhost:%s\n", cfg.GRPCPort)
	case "http":
		fmt.Printf("HTTP Server:  http://localhost:%s\n", cfg.HTTPPort)
		fmt.Printf("Health Check: http://localhost:%s/health\n", cfg.HTTPPort)
	case "both":
		fmt.Printf("gRPC Server:  localhost:%s\n", cfg.GRPCPort)
		fmt.Printf("HTTP Server:  http://localhost:%s\n", cfg.HTTPPort)
		fmt.Printf("Health Check: http://localhost:%s/health\n", cfg.HTTPPort)
	}

	fmt.Printf("\nConfiguration:\n")
	fmt.Printf("  Browser Mode: %s\n", func() string {
		if cfg.BrowserHeadless {
			return "Headless"
		}
		return "GUI"
	}())
	fmt.Printf("  Debug Mode:   %v\n", cfg.BrowserDebug)
	fmt.Printf("  Database:     %s\n", cfg.SQLitePath)

	fmt.Println("\nPress Ctrl+C to stop the server")
	fmt.Println(strings.Repeat("=", 50) + "\n")
}