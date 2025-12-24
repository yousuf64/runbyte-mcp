package session

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/yousuf/codebraid-mcp/internal/client"
	"github.com/yousuf/codebraid-mcp/internal/codegen"
)

// SessionContext represents a session with its associated resources and lifecycle.
// It stores session data independently of request contexts.
type SessionContext struct {
	SessionID      string
	ClientHub      *client.McpClientHub
	CreatedAt      time.Time
	Libs           map[string]string // Generated TypeScript library files
	BundleDir      string            // Persistent directory for libs and bundling workspace
	lastAccessedAt time.Time
	mu             sync.RWMutex
}

// NewSessionContext creates a new session context.
func NewSessionContext(sessionID string, clientHub *client.McpClientHub) *SessionContext {
	now := time.Now()
	return &SessionContext{
		SessionID:      sessionID,
		ClientHub:      clientHub,
		CreatedAt:      now,
		lastAccessedAt: now,
	}
}

// UpdateLastAccessed updates the last accessed timestamp (thread-safe)
func (s *SessionContext) UpdateLastAccessed() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastAccessedAt = time.Now()
}

// LastAccessedAt returns the last accessed timestamp (thread-safe)
func (s *SessionContext) LastAccessedAt() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastAccessedAt
}

// Age returns the duration since the session was created
func (s *SessionContext) Age() time.Duration {
	return time.Since(s.CreatedAt)
}

// IdleDuration returns the duration since the last access
func (s *SessionContext) IdleDuration() time.Duration {
	return time.Since(s.LastAccessedAt())
}

// RefreshLibs regenerates library files for a specific server
// Called when an MCP server notifies of tool changes
func (s *SessionContext) RefreshLibs(ctx context.Context, serverName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Re-introspect the specific server
	intr := codegen.NewIntrospector(s.ClientHub)
	tools, err := intr.IntrospectServer(ctx, serverName)
	if err != nil {
		return fmt.Errorf("failed to re-introspect server %s: %w", serverName, err)
	}

	// Regenerate TypeScript file for this server
	generator := codegen.NewTypeScriptGenerator()
	file, err := generator.GenerateFile(serverName, tools)
	if err != nil {
		return fmt.Errorf("failed to regenerate lib for %s: %w", serverName, err)
	}

	// Update the in-memory libs map
	s.Libs[serverName] = file

	// Update the file on disk so next bundle picks it up
	if s.BundleDir != "" {
		libPath := filepath.Join(s.BundleDir, "lib", fmt.Sprintf("%s.ts", serverName))
		if err = os.WriteFile(libPath, []byte(file), 0644); err != nil {
			return fmt.Errorf("failed to write updated lib to disk: %w", err)
		}
	}

	return nil
}

// RefreshAllLibs regenerates all library files
// Called when multiple servers change or on manual refresh
func (s *SessionContext) RefreshAllLibs(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	intr := codegen.NewIntrospector(s.ClientHub)
	allTools, err := intr.IntrospectAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to re-introspect: %w", err)
	}

	groupedTools := codegen.GroupByServer(allTools)
	generator := codegen.NewTypeScriptGenerator()

	newLibs := make(map[string]string)
	for server, tools := range groupedTools {
		file, err := generator.GenerateFile(server, tools)
		if err != nil {
			return fmt.Errorf("failed to regenerate lib for %s: %w", server, err)
		}
		newLibs[server] = file
	}

	// Replace entire libs map
	s.Libs = newLibs

	// Update all files on disk
	if s.BundleDir != "" {
		libDir := filepath.Join(s.BundleDir, "lib")
		for serverName, content := range newLibs {
			libPath := filepath.Join(libDir, fmt.Sprintf("%s.ts", serverName))
			if err := os.WriteFile(libPath, []byte(content), 0644); err != nil {
				return fmt.Errorf("failed to write lib %s to disk: %w", serverName, err)
			}
		}
	}

	return nil
}
