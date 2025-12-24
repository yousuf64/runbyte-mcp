package session

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/yousuf/codebraid-mcp/internal/bundler"
	"github.com/yousuf/codebraid-mcp/internal/client"
	"github.com/yousuf/codebraid-mcp/internal/codegen"
	"github.com/yousuf/codebraid-mcp/internal/config"
)

// Manager manages session contexts
type Manager struct {
	sessions map[string]*SessionContext
	mu       sync.RWMutex
	config   *config.Config
}

// NewManager creates a new session manager
func NewManager(cfg *config.Config) *Manager {
	return &Manager{
		sessions: make(map[string]*SessionContext),
		config:   cfg,
	}
}

// GetOrCreateSession gets an existing session or creates a new one
func (m *Manager) GetOrCreateSession(ctx context.Context, sessionID string) (*SessionContext, error) {
	// Try to get existing session
	m.mu.RLock()
	session, exists := m.sessions[sessionID]
	m.mu.RUnlock()

	if exists {
		return session, nil
	}

	// Create new session
	m.mu.Lock()
	defer m.mu.Unlock()

	// Double-check after acquiring write lock
	if session, exists := m.sessions[sessionID]; exists {
		return session, nil
	}

	// Create new McpClientHub and connect to all servers
	clientHub := client.NewMcpClientHub()
	if err := clientHub.Connect(ctx, m.config); err != nil {
		return nil, fmt.Errorf("failed to connect client hub: %w", err)
	}

	// Initialize session context
	session = NewSessionContext(sessionID, clientHub)

	// Setup bundle directory and generate library files
	if err := m.initializeSessionBundleDir(ctx, session); err != nil {
		// Clean up client hub on error
		clientHub.Close()
		return nil, fmt.Errorf("failed to initialize session bundle directory: %w", err)
	}

	m.sessions[sessionID] = session

	return session, nil
}

// GetSession retrieves an existing session
func (m *Manager) GetSession(sessionID string) *SessionContext {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.sessions[sessionID]
}

// DeleteSession removes a session and cleans up its resources
func (m *Manager) DeleteSession(sessionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, exists := m.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session %q not found", sessionID)
	}

	// Close all client connections
	if err := session.ClientHub.Close(); err != nil {
		return fmt.Errorf("failed to close client hub: %w", err)
	}

	// Clean up bundle directory
	if session.BundleDir != "" {
		if err := os.RemoveAll(session.BundleDir); err != nil {
			log.Printf("Warning: failed to clean up bundle dir %s: %v", session.BundleDir, err)
		}
	}

	delete(m.sessions, sessionID)
	return nil
}

// CloseAll closes all sessions
func (m *Manager) CloseAll() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errs []error
	for sessionID, session := range m.sessions {
		if err := session.ClientHub.Close(); err != nil {
			errs = append(errs, fmt.Errorf("session %q: %w", sessionID, err))
		}

		// Clean up bundle directory
		if session.BundleDir != "" {
			if err := os.RemoveAll(session.BundleDir); err != nil {
				log.Printf("Warning: failed to clean up bundle dir %s: %v", session.BundleDir, err)
			}
		}
	}

	m.sessions = make(map[string]*SessionContext)

	if len(errs) > 0 {
		return fmt.Errorf("errors closing sessions: %v", errs)
	}

	return nil
}

// initializeSessionBundleDir creates the bundle directory and writes library files
func (m *Manager) initializeSessionBundleDir(ctx context.Context, session *SessionContext) error {
	// Create persistent bundle directory for this session
	bundleDir, err := os.MkdirTemp("", fmt.Sprintf("codebraid-%s-", session.SessionID))
	if err != nil {
		return fmt.Errorf("failed to create bundle dir: %w", err)
	}

	// Create lib directory
	libDir := filepath.Join(bundleDir, "lib")
	if err := os.Mkdir(libDir, 0755); err != nil {
		os.RemoveAll(bundleDir)
		return fmt.Errorf("failed to create lib dir: %w", err)
	}

	// Introspect all MCP servers and generate TypeScript libraries
	intr := codegen.NewIntrospector(session.ClientHub)
	allTools, err := intr.IntrospectAll(ctx)
	if err != nil {
		os.RemoveAll(bundleDir)
		return fmt.Errorf("failed to introspect tools: %w", err)
	}

	groupedTools := codegen.GroupByServer(allTools)
	generator := codegen.NewTypeScriptGenerator()

	// Generate and write library files
	libs := make(map[string]string)
	for serverName, tools := range groupedTools {
		file, err := generator.GenerateFile(serverName, tools)
		if err != nil {
			os.RemoveAll(bundleDir)
			return fmt.Errorf("failed to generate lib for %s: %w", serverName, err)
		}

		libs[serverName] = file

		// Write to disk
		libPath := filepath.Join(libDir, fmt.Sprintf("%s.ts", serverName))
		if err := os.WriteFile(libPath, []byte(file), 0644); err != nil {
			os.RemoveAll(bundleDir)
			return fmt.Errorf("failed to write lib %s: %w", serverName, err)
		}
	}

	// Write mcp-types.ts
	mcpTypesContent := generator.GenerateMCPTypesFile()
	mcpTypesPath := filepath.Join(libDir, "mcp-types.ts")
	if err := os.WriteFile(mcpTypesPath, []byte(mcpTypesContent), 0644); err != nil {
		os.RemoveAll(bundleDir)
		return fmt.Errorf("failed to write mcp-types.ts: %w", err)
	}

	// Write rspack config
	rspackConfigPath := filepath.Join(bundleDir, "rspack.config.ts")
	rspackConfig := bundler.GetEmbeddedConfig()
	if err := os.WriteFile(rspackConfigPath, []byte(rspackConfig), 0644); err != nil {
		os.RemoveAll(bundleDir)
		return fmt.Errorf("failed to write rspack config: %w", err)
	}

	// Update session
	session.Libs = libs
	session.BundleDir = bundleDir

	return nil
}
