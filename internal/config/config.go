package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config represents the main configuration structure
type Config struct {
	McpServers map[string]McpServerConfig `json:"mcpServers"`
}

// McpServerConfig is the interface for all MCP server configurations
type McpServerConfig struct {
	Type string `json:"type"` // "stdio", "http", or "sse"

	// Stdio fields
	Command string            `json:"command,omitempty"`
	Args    []string          `json:"args,omitempty"`
	Cwd     string            `json:"cwd,omitempty"`
	Env     map[string]string `json:"env,omitempty"`

	// HTTP/SSE fields
	URL     string            `json:"url,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

// Load reads and parses the configuration file
func Load(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if err := validate(&config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &config, nil
}

// validate checks if the configuration is valid
func validate(config *Config) error {
	if len(config.McpServers) == 0 {
		return fmt.Errorf("no MCP servers configured")
	}

	for name, server := range config.McpServers {
		switch server.Type {
		case "stdio":
			if server.Command == "" {
				return fmt.Errorf("server %q: command is required for stdio type", name)
			}
		case "http", "sse":
			if server.URL == "" {
				return fmt.Errorf("server %q: url is required for %s type", name, server.Type)
			}
		default:
			return fmt.Errorf("server %q: invalid type %q (must be stdio, http, or sse)", name, server.Type)
		}
	}

	return nil
}
