package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yousuf/codebraid-mcp/internal/config"
	"github.com/yousuf/codebraid-mcp/internal/server"
	"github.com/yousuf/codebraid-mcp/internal/session"
)

func main() {
	// Get config path from environment
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH environment variable is required")
	}

	// Load configuration
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Loaded configuration with %d MCP server(s)", len(cfg.McpServers))

	// Create session manager
	sessionMgr := session.NewManager(cfg)

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	// Create HTTP handler with proper session management
	handler := mcp.NewStreamableHTTPHandler(func(req *http.Request) *mcp.Server {
		// Create a new MCP server instance for each request
		// This allows the SDK to manage sessions properly
		return server.NewMCPServer(sessionMgr)
	}, &mcp.StreamableHTTPOptions{
		Stateless:      false,
		JSONResponse:   false,
		Logger:         nil,
		EventStore:     nil,
		SessionTimeout: 0,
	})

	// Setup HTTP server
	httpServer := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("CodeBraid MCP server listening on port %s", port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	// Close all sessions
	if err := sessionMgr.CloseAll(); err != nil {
		log.Printf("Error closing sessions: %v", err)
	}

	log.Println("Server stopped")
}
