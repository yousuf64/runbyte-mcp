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

// NewSandbox creates a new sandbox instance
func NewSandbox(ctx context.Context, wasmPath string, clientHub *client.McpClientHub) (*Sandbox, error) {
	manifest := extism.Manifest{
		Wasm: []extism.Wasm{
			extism.WasmFile{
				Path: wasmPath,
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

	var outputMap map[string]interface{}
	err = json.Unmarshal(output, &outputMap)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal output: %w", err)
	}

	if errVal, ok := outputMap["error"]; ok && errVal != nil && errVal.(string) != "" {
		if stackVal, ok := outputMap["stack"]; ok && stackVal != nil && stackVal.(string) != "" {
			mappedStack, err := sourcemap.Map(sourceMap, stackVal.(string), true)
			if err != nil {
				return "", fmt.Errorf("failed to map error stack trace: %w", err)
			}

			errorOutput := map[string]interface{}{
				"error": errVal.(string),
				"stack": mappedStack,
			}
			errorOutputJson, err := json.Marshal(errorOutput)
			if err != nil {
				return "", fmt.Errorf("failed to marshal error output: %w", err)
			}
			return string(errorOutputJson), err
		}
	}

	return string(output), nil
}

// Close closes the sandbox and frees resources
func (s *Sandbox) Close() {
	if s.plugin != nil {
		s.plugin.Close(s.ctx)
	}
}
