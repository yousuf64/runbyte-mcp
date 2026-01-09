package server

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yousuf/runbyte/internal/bundler"
	"github.com/yousuf/runbyte/internal/sandbox"
	"github.com/yousuf/runbyte/internal/session"
	"github.com/yousuf/runbyte/internal/strutil"
)

// ExecuteCodeArgs represents the arguments for the execute_code tool
type ExecuteCodeArgs struct {
	Code string `json:"code" jsonschema:"TypeScript code to execute in sandbox"`
}

// ListDirectoryArgs represents the arguments for the list_directory tool
type ListDirectoryArgs struct {
	Path             string `json:"path" jsonschema:"Path to directory (e.g., '/', '/servers', '/servers/github'). Defaults to '/' if not provided."`
	WithDescriptions bool   `json:"withDescriptions,omitempty" jsonschema:"Include descriptions for functions (default: false)"`
}

// ReadFileArgs represents the arguments for the read_file tool
type ReadFileArgs struct {
	Path string `json:"path" jsonschema:"Required. Path to file in virtual filesystem (e.g., '/servers/github/listRepos.ts', '/servers/github/index.ts')"`
}

// NewMcpServer creates and configures the MCP server
func NewMcpServer(wasmBytes []byte, sessionMgr *session.Manager) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "runbyte",
		Version: "1.0.0",
	}, &mcp.ServerOptions{
		Instructions: `
# Runbyte: Execute MCP Tools as TypeScript Code

Runbyte implements the code execution pattern for MCP, enabling you to call any MCP tool by writing TypeScript code instead of traditional tool calling. This dramatically reduces token consumption (up to 98.7%) for complex workflows by processing data in a sandboxed environment rather than passing everything through your context.

## Virtual Filesystem Architecture

All MCP servers are exposed as TypeScript modules in a virtual filesystem. Explore this filesystem to discover what's available:

/
├── servers/              MCP tools compiled to TypeScript
│   ├── github/          All GitHub MCP tools as async functions
│   ├── filesystem/      File operations
│   ├── slack/           Slack integrations
│   └── ...              (All configured MCP servers)
└── workspace/           Your persistent workspace (read/write)

## Efficient Discovery Pattern

Don't load all tools upfront. Discover on-demand:

1. **Start at root**: list_directory({ path: "/" })
2. **Browse servers**: list_directory({ path: "/servers" })
3. **Explore specific server**: list_directory({ path: "/servers/github" })
4. **Read tool signatures**: read_file({ path: "/servers/github/createIssue.ts" })

Each tool file contains complete TypeScript types, JSDoc, and function signatures. Read only what you need.

## Code Execution Pattern

Write TypeScript code that calls MCP tools as normal async functions. All code must export an exec() function as the entry point:

### Simple Example: Call One Tool
` + "```typescript" + `
import * as github from './servers/github';

async function exec() {
  const repos = await github.listRepos({ 
    owner: "octocat",
    type: "public"
  });
  
  return repos.filter(r => r.stargazers_count > 100);
}
` + "```" + `

### Realistic Example: Multi-Step Workflow with Workspace
` + "```typescript" + `
import * as github from './servers/github';
import * as slack from './servers/slack';
import * as fs from '@runbyte/fs';

async function exec() {
  // Step 1: Fetch all repos
  const repos = await github.listRepos({ owner: "myorg" });
  
  // Step 2: Process and filter (happens in sandbox, not in your context)
  const criticalRepos = repos.filter(r => 
    r.open_issues_count > 10 && 
    r.pushed_at > Date.now() - 7 * 24 * 60 * 60 * 1000
  );
  
  // Step 3: Get detailed issues for each repo
  const repoIssues = [];
  for (const repo of criticalRepos) {
    const issues = await github.listIssues({ 
      owner: "myorg", 
      repo: repo.name,
      state: "open"
    });
    repoIssues.push({ repo: repo.name, issues });
  }
  
  // Step 4: Store full results in workspace for later analysis
  fs.writeFile(
    "./workspace/issues-report.json", 
    JSON.stringify(repoIssues, null, 2)
  );
  
  // Step 5: Generate summary for Slack (only summary goes through your context)
  const summary = ` + "`" + `Found ${criticalRepos.length} repos needing attention with ${
    repoIssues.reduce((sum, r) => sum + r.issues.length, 0)
  } total open issues.` + "`" + `;
  
  await slack.postMessage({
    channel: "#dev-alerts",
    text: summary
  });
  
  return { summary, criticalRepos: criticalRepos.length };
}
` + "```" + `

Notice: Full repo and issue data never passes through your context. Only the final summary does.

IMPORTANT: Since most MCP server tools do not explicitly document their output schema (return type), learn and memorize their return types from the interactions.  

### Complex Example: Data Aggregation with State Management
` + "```typescript" + `
import * as github from './servers/github';
import * as filesystem from './servers/filesystem';
import * as fs from '@runbyte/fs';

async function exec() {
  // Check if we have cached results from previous run
  try {
    const cached = fs.readFile("./workspace/metrics.json");
    const data = JSON.parse(cached);
    if (Date.now() - data.timestamp < 3600000) { // 1 hour
      return data.metrics;
    }
  } catch {}
  
  // No cache or expired - fetch and compute
  const repos = await github.listRepos({ owner: "myorg" });
  
  const metrics = {
    total: repos.length,
    byLanguage: repos.reduce((acc, r) => {
      acc[r.language] = (acc[r.language] || 0) + 1;
      return acc;
    }, {}),
    avgStars: repos.reduce((sum, r) => sum + r.stargazers_count, 0) / repos.length,
    recentlyUpdated: repos.filter(r => 
      new Date(r.updated_at) > new Date(Date.now() - 30 * 24 * 60 * 60 * 1000)
    ).length
  };
  
  // Store for next time
  fs.writeFile("./workspace/metrics.json", JSON.stringify({
    timestamp: Date.now(),
    metrics
  }));
  
  // Also append to history log
  const logEntry = ` + "`" + `[${new Date().toISOString()}] Metrics: ${JSON.stringify(metrics)}\n` + "`" + `;
  try {
    const existing = fs.readFile("./workspace/metrics-history.log");
    fs.writeFile("./workspace/metrics-history.log", existing + logEntry);
  } catch {
    fs.writeFile("./workspace/metrics-history.log", logEntry);
  }
  
  return metrics;
}
` + "```" + `

## Available Tools

- **list_directory** - Explore the virtual filesystem to discover MCP servers and tools
- **read_file** - Read tool definitions, workspace files, or cached data
- **execute_code** - Run your TypeScript code with automatic bundling

## Filesystem API (@runbyte/fs)

For multi-step workflows, use the workspace to store intermediate results:

` + "```typescript" + `
import * as fs from '@runbyte/fs';

// Write files to workspace
fs.writeFile("./workspace/data.json", JSON.stringify(data));

// Read files from workspace
const content = fs.readFile("./workspace/data.json");

// List workspace contents
const files = fs.listFiles("./workspace");

// Delete files from workspace
fs.deleteFile("./workspace/old-data.json");
` + "```" + `

## Key Principles

1. **Discover efficiently**: Start broad, narrow down. Don't read all tools.
2. **Process in sandbox**: Loops, filtering, aggregation happen in exec(), not in your context.
3. **Use workspace strategically**: Store full datasets, keep summaries for your context.
4. **Chain operations**: Call multiple MCP tools in sequence within one exec().
5. **Return what matters**: Only final results/summaries should return to your context.

## Technical Details

- All paths start with '/'
- Namespace imports work best: ` + "`" + `import * as github from './servers/github'` + "`" + `
- exec() can be sync or async
- Execution timeout: 30 seconds
- Automatic bundling with full TypeScript support
- Each tool file has complete type definitions and JSDoc
`,
	})

	server.AddReceivingMiddleware(createSessionInjectionMiddleware(sessionMgr))
	server.AddReceivingMiddleware(createLoggingMiddleware())

	// Register execute_code tool
	mcp.AddTool(server, &mcp.Tool{
		Name: "execute_code",
		Description: `Execute TypeScript code that calls MCP tools as functions. Process data in the sandbox, chain multiple operations, and use the workspace for multi-step workflows.

Your code must export an exec() function as the entry point:

    async function exec() {
        // Your workflow here
        return result;
    }

The exec() function:
- Required entry point (async or sync)
- Returns any JSON-serializable value
- Has access to all MCP servers as typed modules
- Can read/write workspace for persistent data

Import MCP servers as TypeScript modules:
    import * as github from './servers/github';
    import * as slack from './servers/slack';
    import * as fs from '@runbyte/fs';

Single-step example:
    import * as github from './servers/github';
    
    async function exec() {
        const repos = await github.listRepos({ owner: "octocat" });
        return repos.filter(r => r.stargazers_count > 100);
    }

Multi-step workflow example (process in sandbox, store intermediate results):
    import * as github from './servers/github';
    import * as slack from './servers/slack';
    import * as fs from '@runbyte/fs';
    
    async function exec() {
        // Fetch data
        const repos = await github.listRepos({ owner: "myorg" });
        
        // Process in sandbox (doesn't pass through your context)
        const active = repos.filter(r => r.open_issues_count > 5);
        
        // Store full data for later
        fs.writeFile("./workspace/repos.json", JSON.stringify(active));
        
        // Get detailed info for each
        const details = [];
        for (const repo of active) {
            const issues = await github.listIssues({ 
                owner: "myorg", 
                repo: repo.name 
            });
            details.push({ name: repo.name, issueCount: issues.length });
        }
        
        // Notify team
        await slack.postMessage({
            channel: "#eng",
            text: ` + "`" + `Found ${active.length} repos with issues` + "`" + `
        });
        
        // Return only summary (not all data)
        return { summary: details };
    }

Filesystem API for workflows:
    import * as fs from '@runbyte/fs';
    
    // All operations use the workspace/ directory
    fs.writeFile("./workspace/data.json", jsonString);
    const content = fs.readFile("./workspace/data.json");
    const files = fs.listFiles("./workspace");
    fs.deleteFile("./workspace/old.json");

Sandbox environment:
- Execution timeout: 30 seconds
- Automatic bundling with TypeScript support
- No Node.js built-ins or DOM APIs
- All MCP tool calls are async
`,
	}, func(ctx context.Context, req *mcp.CallToolRequest, args ExecuteCodeArgs) (*mcp.CallToolResult, any, error) {
		sessionCtx, err := getSessionFromContext(ctx)
		if err != nil {
			return nil, nil, err
		}

		// Step 1: Bundle the code using session's bundle directory
		b, err := bundler.New()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create bundler: %w", err)
		}

		codeWithCaller := fmt.Sprintf(`%s
exec();
`, args.Code)
		bundledCode, sourceMap, err := b.Bundle(sessionCtx.BundleDir, codeWithCaller)
		if err != nil {
			return nil, nil, fmt.Errorf("bundling failed: %w", err)
		}

		// Step 2: Create sandbox with filesystem access
		sb, err := sandbox.NewSandbox(ctx, wasmBytes, sessionCtx.ClientHub, sessionCtx.SandboxFS)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create sandbox: %w", err)
		}
		defer sb.Close()

		// Step 3: Execute bundled code
		result, err := sb.ExecuteCode(bundledCode, sourceMap)
		if err != nil {
			return nil, nil, fmt.Errorf("execution failed: %w", err)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: result},
			},
		}, nil, nil
	})

	// Register list_directory tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_directory",
		Description: "List contents of a directory in the virtual filesystem. Returns directories and files with their types. Supports both /servers/* and /workspace/* directories.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args ListDirectoryArgs) (*mcp.CallToolResult, any, error) {
		sessionCtx, err := getSessionFromContext(ctx)
		if err != nil {
			return nil, nil, err
		}

		var output bytes.Buffer

		// Normalize path
		path := strings.TrimPrefix(args.Path, "/")
		path = strings.TrimSuffix(path, "/")

		if path == "" {
			// Root directory - show both servers and filesystem directories
			output.WriteString("/\n")
			output.WriteString("├── servers/ (MCP servers)\n")

			// List filesystem directories
			if sessionCtx.SandboxFS != nil {
				dirs := sessionCtx.SandboxFS.GetDirectories()
				for i, dir := range dirs {
					prefix := "├──"
					if i == len(dirs)-1 {
						prefix = "└──"
					}
					output.WriteString(fmt.Sprintf("%s %s/\n", prefix, dir))
				}
			}

			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: output.String()},
				},
			}, nil, nil
		}

		if path == "servers" {
			// List all MCP servers
			allTools := sessionCtx.ClientHub.Tools()
			output.WriteString("/servers/\n")

			serverCount := 0
			for svr, toolList := range allTools {
				serverCount++
				prefix := "├──"
				output.WriteString(fmt.Sprintf("%s %s/ (%d functions)\n", prefix, svr, len(toolList)))
			}
			output.WriteString("└── index.ts\n")

			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: output.String()},
				},
			}, nil, nil
		}

		if strings.HasPrefix(path, "servers/") {
			// List specific server directory
			serverName := strings.TrimPrefix(path, "servers/")
			tools, ok := sessionCtx.ClientHub.ServerTools(serverName)
			if !ok {
				availableServers := sessionCtx.ClientHub.Servers()
				return nil, nil, fmt.Errorf("directory '/servers/%s/' not found. Available servers: %v",
					serverName, availableServers)
			}

			output.WriteString(fmt.Sprintf("/servers/%s/\n", serverName))
			for i, tool := range tools {
				prefix := "├──"
				if i == len(tools)-1 {
					prefix = "├──"
				}

				funcName := strutil.ToCamelCase(tool.Name)
				if args.WithDescriptions && tool.Description != "" {
					output.WriteString(fmt.Sprintf("%s %s.ts - %s\n", prefix, funcName, tool.Description))
				} else {
					output.WriteString(fmt.Sprintf("%s %s.ts\n", prefix, funcName))
				}
			}
			output.WriteString("└── index.ts\n")
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: output.String()},
				},
			}, nil, nil
		}

		// Check if it's a filesystem directory (workspace)
		if sessionCtx.SandboxFS != nil {
			dirs := sessionCtx.SandboxFS.GetDirectories()
			for _, dir := range dirs {
				if path == dir || strings.HasPrefix(path, dir+"/") {
					// List filesystem directory
					fsPath := "./" + path
					files, err := sessionCtx.SandboxFS.ListFiles(fsPath)
					if err != nil {
						return nil, nil, fmt.Errorf("failed to list directory '/%s': %w", path, err)
					}

					output.WriteString(fmt.Sprintf("/%s/\n", path))
					for i, file := range files {
						prefix := "├──"
						if i == len(files)-1 {
							prefix = "└──"
						}
						output.WriteString(fmt.Sprintf("%s %s\n", prefix, file))
					}

					return &mcp.CallToolResult{
						Content: []mcp.Content{
							&mcp.TextContent{Text: output.String()},
						},
					}, nil, nil
				}
			}
		}

		return nil, nil, fmt.Errorf("directory '/%s' not found", path)
	})

	// Register read_file tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "read_file",
		Description: "Read a file from the virtual filesystem. Supports /servers/* paths and /workspace/* paths. Examples: '/servers/github/listRepos.ts', '/workspace/config.json'.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args ReadFileArgs) (*mcp.CallToolResult, any, error) {
		sessionCtx, err := getSessionFromContext(ctx)
		if err != nil {
			return nil, nil, err
		}

		// Normalize path - remove leading slash
		path := strings.TrimPrefix(args.Path, "/")

		// Check if it's a filesystem path (workspace)
		if sessionCtx.SandboxFS != nil {
			dirs := sessionCtx.SandboxFS.GetDirectories()
			for _, dir := range dirs {
				if strings.HasPrefix(path, dir+"/") {
					// Read from sandbox filesystem
					fsPath := "./" + path
					content, err := sessionCtx.SandboxFS.ReadFile(fsPath)
					if err != nil {
						return nil, nil, fmt.Errorf("failed to read file '/%s': %w", path, err)
					}

					return &mcp.CallToolResult{
						Content: []mcp.Content{
							&mcp.TextContent{Text: content},
						},
					}, nil, nil
				}
			}
		}

		// If path starts with "servers/", read from bundle directory
		if strings.HasPrefix(path, "servers/") {
			filePath := fmt.Sprintf("%s/%s", sessionCtx.BundleDir, path)

			// Read the file from disk
			content, err := os.ReadFile(filePath)
			if err != nil {
				return nil, nil, fmt.Errorf("file '/%s' not found", path)
			}

			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: string(content)},
				},
			}, nil, nil
		}

		return nil, nil, fmt.Errorf("file '/%s' not found - path must start with 'servers/', 'workspace/'", path)
	})

	return server
}
