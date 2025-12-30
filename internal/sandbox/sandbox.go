package sandbox

import (
	"context"
	"encoding/json"
	"fmt"

	extism "github.com/extism/go-sdk"
	"github.com/yousuf/codebraid-mcp/internal/client"
	"github.com/yousuf/codebraid-mcp/internal/sourcemap"
)

// Sandbox provides a WebAssembly execution environment for user code
type Sandbox struct {
	plugin    *extism.Plugin
	clientHub *client.McpClientHub
	ctx       context.Context
}

type ExecuteCodeResult struct {
	Error  string
	Stack  string
	Result string
}

// NewSandbox creates a new sandbox instance from WASM bytes
func NewSandbox(ctx context.Context, wasmBytes []byte, clientHub *client.McpClientHub) (*Sandbox, error) {
	manifest := extism.Manifest{
		Wasm: []extism.Wasm{
			extism.WasmData{
				Data: wasmBytes,
			},
		},
	}

	config := extism.PluginConfig{
		EnableWasi: true,
	}

	sb := &Sandbox{
		clientHub: clientHub,
		ctx:       ctx,
	}

	// Create host functions
	hostFunctions := []extism.HostFunction{
		createCallMcpToolHostFunc(sb),
	}

	plugin, err := extism.NewPlugin(ctx, manifest, config, hostFunctions)
	if err != nil {
		return nil, fmt.Errorf("failed to create plugin: %w", err)
	}

	sb.plugin = plugin
	return sb, nil
}

// ExecuteCode executes bundled JavaScript code in the sandbox
func (s *Sandbox) ExecuteCode(bundledCode, sourceMap string) (string, error) {
	// Call the executeCode function exported by the JavaScript plugin
	exit, output, err := s.plugin.Call("executeCode", []byte(bundledCode))
	if err != nil {
		return "", fmt.Errorf("plugin execution failed: %w", err)
	}
	if exit != 0 {
		return "", fmt.Errorf("plugin exited with code %d", exit)
	}

	var result ExecuteCodeResult
	err = json.Unmarshal(output, &result)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal output: %w", err)
	}

	if result.Error != "" {
		if result.Stack != "" {
			mappedStack, err := sourcemap.Map(sourceMap, result.Stack, true)
			if err != nil {
				return "", fmt.Errorf("failed to map error stack trace: %w", err)
			}

			return "", fmt.Errorf("failed to execute code\nerror:\n%s\nstack trace:\n%s", result.Error, mappedStack)
		}

		return "", fmt.Errorf("failed to execute code:\n%s", result.Error)
	}

	return result.Result, nil
}

// Close closes the sandbox and frees resources
func (s *Sandbox) Close() {
	if s.plugin != nil {
		s.plugin.Close(s.ctx)
	}
}
