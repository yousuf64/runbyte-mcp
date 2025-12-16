package client

import (
	"context"
	"fmt"
	"sync"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yousuf/codebraid-mcp/internal/config"
)

// ClientBox manages multiple MCP client connections
type ClientBox struct {
	clients map[string]*MCPClient
	mu      sync.RWMutex
}

// NewClientBox creates a new ClientBox
func NewClientBox() *ClientBox {
	return &ClientBox{
		clients: make(map[string]*MCPClient),
	}
}

// Connect establishes connections to all configured MCP servers
func (cb *ClientBox) Connect(ctx context.Context, cfg *config.Config) error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	for name, serverCfg := range cfg.McpServers {
		client, err := NewMCPClient(ctx, name, serverCfg)
		if err != nil {
			return fmt.Errorf("failed to connect to server %q: %w", name, err)
		}
		cb.clients[name] = client
	}

	return nil
}

// CallTool calls a tool on a specific MCP server
func (cb *ClientBox) CallTool(ctx context.Context, serverName, toolName string, args map[string]interface{}) (*mcp.CallToolResult, error) {
	cb.mu.RLock()
	client, exists := cb.clients[serverName]
	cb.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("server %q not found", serverName)
	}

	return client.CallTool(ctx, toolName, args)
}

// FindToolServer finds which server has a specific tool
func (cb *ClientBox) FindToolServer(toolName string) (string, error) {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	for name, client := range cb.clients {
		for _, tool := range client.GetTools() {
			if tool.Name == toolName {
				return name, nil
			}
		}
	}

	return "", fmt.Errorf("tool %q not found in any server", toolName)
}

// ListTools returns all tools from all servers
func (cb *ClientBox) ListTools() map[string][]*mcp.Tool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	result := make(map[string][]*mcp.Tool)
	for name, client := range cb.clients {
		result[name] = client.GetTools()
	}

	return result
}

// GetToolsWithDescription returns tools with their descriptions
func (cb *ClientBox) GetToolsWithDescription() map[string][]ToolInfo {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	result := make(map[string][]ToolInfo)
	for name, client := range cb.clients {
		tools := make([]ToolInfo, 0, len(client.GetTools()))
		for _, tool := range client.GetTools() {
			tools = append(tools, ToolInfo{
				Name:        tool.Name,
				Description: tool.Description,
			})
		}
		result[name] = tools
	}

	return result
}

// Close closes all client connections
func (cb *ClientBox) Close() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	var errs []error
	for name, client := range cb.clients {
		if err := client.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close client %q: %w", name, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing clients: %v", errs)
	}

	return nil
}

// ToolInfo contains basic tool information
type ToolInfo struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}
