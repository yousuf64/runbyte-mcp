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
	"github.com/yousuf/codebraid-mcp/internal/bundler"
	"github.com/yousuf/codebraid-mcp/internal/config"
	"github.com/yousuf/codebraid-mcp/internal/server"
	"github.com/yousuf/codebraid-mcp/internal/session"
	"github.com/yousuf/codebraid-mcp/pkg/wasm"
)

// getWasmBytes returns WASM bytes, preferring config path over embedded
func getWasmBytes(cfg *config.Config) ([]byte, error) {
	// Check for config override
	if wasmPath := cfg.GetWasmPath(); wasmPath != "" {
		data, err := os.ReadFile(wasmPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read WASM from config path %q: %w", wasmPath, err)
		}
		return data, nil
	}

	// Use embedded WASM
	if len(wasm.Embedded) == 0 {
		return nil, fmt.Errorf("embedded WASM not found - binary may not be built correctly")
	}
	return wasm.Embedded, nil
}

func runStdioServer(wasmBytes []byte, sessionMgr *session.Manager) {
	log.Println("CodeBraid MCP server running in stdio mode")

	// Create MCP server
	mcpServer := server.NewMcpServer(wasmBytes, sessionMgr)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Run server in goroutine
	errChan := make(chan error, 1)
	go func() {
		// Run blocks until connection is closed or context is cancelled
		errChan <- mcpServer.Run(ctx, &mcp.StdioTransport{})
	}()

	// Wait for either completion or interrupt
	select {
	case <-sigChan:
		log.Println("Interrupt received, shutting down...")
		cancel()
		// Wait for server to stop
		<-errChan
	case err := <-errChan:
		if err != nil {
			log.Printf("Server stopped with error: %v", err)
		}
	}

	// Close all sessions
	if err := sessionMgr.CloseAll(); err != nil {
		log.Printf("Error closing sessions: %v", err)
	}

	log.Println("Server stopped")
}

func runHttpServer(cfg *config.Config, wasmBytes []byte, sessionMgr *session.Manager, port int) {
	// Create HTTP handler with proper session management
	handler := mcp.NewStreamableHTTPHandler(func(req *http.Request) *mcp.Server {
		// Create a new MCP server instance for each request
		// This allows the SDK to manage sessions properly
		return server.NewMcpServer(wasmBytes, sessionMgr)
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

func main() {
	// Parse command-line flags
	var (
		configPath    = flag.String("config", os.Getenv("CODEBRAID_CONFIG"), "Path to configuration file")
		portFlag      = flag.Int("port", 0, "HTTP server port (overrides config file)")
		transportMode = flag.String("transport", "http", "Transport mode: stdio or http")
		help          = flag.Bool("help", false, "Show usage information")
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

	// Initialize bundler
	if err = bundler.Initialize(); err != nil {
		log.Fatalf("Failed to initialize bundler: %v\n\nHint: Install rspack with: npm install -g @rspack/cli @rspack/core", err)
	}
	log.Println("Bundler initialized successfully")

	// Create session manager
	sessionMgr := session.NewManager(cfg)

	// Load WASM bytes (embedded or from config)
	wasmBytes, err := getWasmBytes(cfg)
	if err != nil {
		log.Fatalf("Failed to load WASM: %v", err)
	}

	// Route to appropriate transport mode
	switch *transportMode {
	case "stdio":
		runStdioServer(wasmBytes, sessionMgr)
	case "http":
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
		runHttpServer(cfg, wasmBytes, sessionMgr, port)
	default:
		log.Fatalf("Invalid transport mode: %s (must be 'stdio' or 'http')", *transportMode)
	}
}
