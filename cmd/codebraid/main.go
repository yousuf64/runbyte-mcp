package main

import (
	"context"
	"flag"
	"fmt"
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
	// Parse command-line flags
	var (
		configPath = flag.String("config", os.Getenv("CODEBRAID_CONFIG"), "Path to configuration file")
		portFlag   = flag.Int("port", 0, "HTTP server port (overrides config file)")
		help       = flag.Bool("help", false, "Show usage information")
	)
	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	// Load configuration with flexible options
	cfg, err := config.LoadWithOptions(config.LoadOptions{
		ConfigPath:        *configPath,
		SearchPaths:       config.DefaultSearchPaths(),
		AllowEnvOverrides: true,
	})
	if err != nil {
		log.Fatalf("Failed to load config: %v\n\nHint: Specify a config file with -config flag or CODEBRAID_CONFIG env var", err)
	}

	log.Printf("Loaded configuration with %d MCP server(s)", len(cfg.McpServers))

	// Determine server port (priority: flag > env > config > default)
	port := *portFlag
	if port == 0 {
		if envPort := os.Getenv("CODEBRAID_PORT"); envPort != "" {
			fmt.Sscanf(envPort, "%d", &port)
		}
	}
	if port == 0 {
		port = cfg.GetServerPort()
	}

	// Create session manager
	sessionMgr := session.NewManager(cfg)

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
	timeout := time.Duration(cfg.GetServerTimeout()) * time.Second
	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      handler,
		ReadTimeout:  timeout,
		WriteTimeout: timeout,
		IdleTimeout:  timeout * 4,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("CodeBraid MCP server listening on port %d", port)
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
