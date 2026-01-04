# Runbyte Architecture

This document provides detailed technical architecture documentation for Runbyte, including system components, data flows, transport mechanisms, and security model.

## System Overview

```
┌─────────────────────────────────────────────────────────────┐
│                      MCP Client                             │
│              (VS Code / Cursor / Claude Desktop)            │
└───────────────────────────┬─────────────────────────────────┘
                            │
                    stdio/HTTP/SSE transport
                            │
                            ▼
┌──────────────────────────────────────────────────────────┐
│                   Runbyte Server (Go)                    │
│                                                          │
│  ┌────────────────────────────────────────────────────┐  │
│  │            MCP Client Hub                          │  │
│  │  • Manages connections to downstream MCP servers   │  │
│  │  • Handles stdio/HTTP/SSE transports               │  │
│  └──────────────────┬─────────────────────────────────┘  │
│                     │                                    │
│                     ▼                                    │
│  ┌────────────────────────────────────────────────────┐  │
│  │         Code Generator (Codegen)                   │  │
│  │  • Introspects MCP server tools                    │  │
│  │  • Converts JSON schemas to TypeScript types       │  │
│  │  • Generates typed function wrappers               │  │
│  └──────────────────┬─────────────────────────────────┘  │
│                     │                                    │
│                     ▼                                    │
│  ┌────────────────────────────────────────────────────┐  │
│  │      Virtual Filesystem (/servers/)                │  │
│  │  • Stores generated TypeScript libraries           │  │
│  │  • Provides list_directory and read_file tools     │  │
│  │  • Session-based caching with invalidation         │  │
│  └──────────────────┬─────────────────────────────────┘  │
│                     │                                    │
│                     ▼                                    │
│  ┌────────────────────────────────────────────────────┐  │
│  │           Bundler (Rspack)                         │  │
│  │  • Bundles user code with generated libraries      │  │
│  │  • Resolves imports and dependencies               │  │
│  │  • Produces single executable bundle               │  │
│  └──────────────────┬─────────────────────────────────┘  │
│                     │                                    │
│                     ▼                                    │
│  ┌────────────────────────────────────────────────────┐  │
│  │      WebAssembly Sandbox (QuickJS)                 │  │
│  │  • Executes bundled TypeScript/JavaScript          │  │
│  │  • Isolated execution environment                  │  │
│  │  • 30-second timeout protection                    │  │
│  │  • No filesystem or network access                 │  │
│  └──────────────────┬─────────────────────────────────┘  │
│                     │                                    │
│                     │ Routes tool calls                  │
│                     ▼                                    │
│  ┌────────────────────────────────────────────────────┐  │
│  │          MCP Client Hub (routing)                  │  │
│  └──────────────────┬─────────────────────────────────┘  │
│                     │                                    │
└─────────────────────┼────────────────────────────────────┘
                      │
        ┌─────────────┼─────────────┬──────────────┐
        │             │             │              │
        ▼             ▼             ▼              ▼
   ┌─────────┐  ┌─────────┐  ┌─────────┐    ┌─────────┐
   │ GitHub  │  │FileSys  │  │ Slack   │ ...│ Custom  │
   │   MCP   │  │  MCP    │  │  MCP    │    │  MCP    │
   └─────────┘  └─────────┘  └─────────┘    └─────────┘
```

## Core Components

### MCP Client Hub
**Purpose:** Manages connections to all downstream MCP servers

**Responsibilities:**
- Establishes and maintains connections to configured MCP servers
- Supports multiple transport types: stdio (command/args), HTTP (url), SSE
- Routes tool execution requests to appropriate servers
- Handles connection lifecycle, reconnection, and error recovery
- Manages concurrent requests across multiple servers

### Code Generator (Codegen)
**Purpose:** Translates MCP tool schemas into TypeScript libraries

**Responsibilities:**
- Introspects each MCP server to discover available tools
- Parses JSON schemas and converts them to TypeScript types
- Generates type-safe function wrappers for each tool
- Creates index files with all exports
- Produces documentation comments from tool descriptions
- Validates schema compatibility and handles edge cases

**Output Example:**
```typescript
// /servers/github/listRepos.ts
export async function listRepos(input: {
  owner: string;
  page?: number;
}): Promise<Repository[]> {
  return callMCPTool('github', 'listRepos', input);
}
```

### Virtual Filesystem
**Purpose:** Provides discoverable access to generated TypeScript libraries

**Responsibilities:**
- Stores generated code in a hierarchical structure (`/servers/`)
- Implements `list_directory` tool for filesystem exploration
- Implements `read_file` tool for reading TypeScript source
- Session-based caching for fast repeated access
- Cache invalidation when downstream tools change
- Serves as the discovery interface for AI agents

**Structure:**
```
/servers/
  ├── github/
  │   ├── listRepos.ts
  │   ├── getIssues.ts
  │   └── index.ts
  ├── filesystem/
  │   └── index.ts
  └── index.ts
```

### Bundler (Rspack)
**Purpose:** Bundles user code with generated libraries into executable form

**Responsibilities:**
- Resolves import statements from user code
- Bundles all dependencies into a single file
- Transpiles TypeScript to JavaScript using SWC (Speedy Web Compiler)
- Performs tree-shaking and optimization
- Produces code compatible with WASM sandbox
- Generates source maps for debugging

**Technology:**
- **Rspack**: High-performance bundler written in Rust
- **SWC**: Ultra-fast TypeScript/JavaScript compiler and transpiler
- Provides near-instant bundling for fast execution cycles

### WASM Sandbox (QuickJS)
**Purpose:** Securely executes user-provided TypeScript/JavaScript code

**Responsibilities:**
- Runs bundled code in isolated WebAssembly environment
- Enforces 30-second execution timeout
- Prevents access to Node.js built-ins (fs, net, etc.)
- Blocks filesystem and network operations
- Provides controlled access only to MCP tool calls
- Returns execution results or errors

**Security Features:**
- No file system access
- No network access (except via MCP tool calls)
- Memory limits and execution timeout
- Isolated from host system
- Deterministic execution environment

## Data Flows

### Flow 1: Tool Discovery & Code Generation

```
1. Runbyte starts → connects to MCP servers
2. MCP servers → return tool list + schemas
3. Code Generator → parses schemas
4. Code Generator → generates TypeScript files
5. Virtual Filesystem → stores generated code
6. Cache → stores for session
```

**Example:**
```
GitHub MCP lists tools: [listRepos, getIssues, createPR]
        ↓
Code Generator creates:
  /servers/github/listRepos.ts
  /servers/github/getIssues.ts  
  /servers/github/createPR.ts
  /servers/github/index.ts
        ↓
Agent can list_directory("/servers/github")
        ↓
Agent can read_file("/servers/github/listRepos.ts")
```

### Flow 2: Code Execution

```
1. Agent submits code via execute_code tool
2. Bundler → resolves imports from /servers/
3. Bundler → produces single JavaScript bundle
4. WASM Sandbox → executes bundle
5. Code calls MCP tools → routed via Client Hub
6. Client Hub → forwards to appropriate MCP server
7. MCP server → returns results
8. Results → flow back to sandbox
9. Sandbox → returns final result to agent
```

**Example:**
```typescript
// Agent's code
import * as github from './servers/github';
const repos = await github.listRepos({ owner: "octocat" });
```

**Execution path:**
```
Bundler resolves './servers/github' → /servers/github/index.ts
        ↓
WASM executes: github.listRepos(...)
        ↓
Sandbox calls: callMCPTool('github', 'listRepos', {...})
        ↓
Client Hub routes to GitHub MCP server
        ↓
GitHub MCP returns repository data
        ↓
Data flows back to sandbox
        ↓
Result returned to agent
```

### Flow 3: Cache Invalidation & Updates

```
1. MCP server tool definitions change
2. Notification sent to Runbyte (if supported)
   OR detected on next introspection
3. Session Manager → invalidates cache for that server
4. Code Generator → regenerates TypeScript files
5. Virtual Filesystem → updates with new code
6. Next execution → uses updated definitions
```

## Transport Layer

### stdio Transport (Default)
- Used by most MCP clients (VS Code, Cursor, etc.)
- Bidirectional JSON-RPC over stdin/stdout
- Process-to-process communication
- Runbyte spawned as child process by client

**Flow:**
```
MCP Client → spawns Runbyte process
         → sends JSON-RPC via stdin
         → receives JSON-RPC via stdout
```

### HTTP Transport
- Used when stdio isn't feasible
- RESTful HTTP endpoints
- Runbyte runs as standalone server
- Client connects via HTTP

**Flow:**
```
Runbyte Server → listens on port (e.g., 3000)
MCP Client → sends HTTP POST with JSON-RPC
          → receives HTTP response with result
```

### Downstream MCP Servers
- Runbyte connects to downstream servers via their configured transport
- Supports stdio, HTTP, and SSE for downstream connections
- Each server can use different transport type
- Connection pooling for HTTP/SSE servers

## Security Model

### Sandbox Isolation

**Code runs in WebAssembly sandbox (QuickJS):**
- No access to host filesystem
- No direct network access
- No Node.js built-in modules
- Only controlled access via MCP tool calls

### Resource Limits

**Enforced constraints:**
- 30-second execution timeout (configurable)
- Memory limits enforced by WASM runtime
- No infinite loops or resource exhaustion

### Data Privacy

**Controlled data exposure:**
- Intermediate data stays in execution environment
- Only returned results enter model context
- Sensitive data never exposed to agent unless explicitly returned

### MCP Tool Access

**Secure tool routing:**
- All tool calls go through validated routing
- Type safety enforced at TypeScript level
- Schema validation on tool inputs
- Error handling prevents sandbox escapes

## Implementation Details

### Session Management

Each client connection establishes a session with:
- Unique session ID
- Dedicated cache for generated TypeScript libraries
- Connection pool for downstream MCP servers
- Resource cleanup on session end

### Code Generation Process

1. **Introspection**: Query MCP server for tool list
2. **Schema Parsing**: Extract JSON schemas for each tool
3. **Type Conversion**: Convert JSON Schema to TypeScript types
4. **Wrapper Generation**: Create async function wrappers
5. **Index Generation**: Create index.ts with exports
6. **Caching**: Store in session cache

### Bundling Process

1. **Import Resolution**: Map imports to virtual filesystem paths
2. **Dependency Graph**: Build complete dependency tree
3. **Transpilation**: Convert TypeScript to JavaScript (SWC)
4. **Tree Shaking**: Remove unused code
5. **Optimization**: Minify and optimize
6. **Output**: Single executable JavaScript bundle

### Execution Process

1. **Bundle Loading**: Load compiled JavaScript into WASM sandbox
2. **Entry Point**: Call the `exec()` function
3. **Tool Calls**: Route `callMCPTool()` calls to Client Hub
4. **Result Collection**: Gather return value from `exec()`
5. **Cleanup**: Release sandbox resources
6. **Response**: Return result to MCP client

## Performance Considerations

### Caching Strategy

- Generated TypeScript libraries cached per session
- Cache invalidated only when tools change
- Reduces introspection overhead on repeated tool discovery

### Bundling Performance

- Rspack provides near-instant bundling (written in Rust)
- SWC transpiles TypeScript faster than traditional tools
- Incremental builds when possible

### Concurrent Execution

- Multiple sessions can run simultaneously
- Each session isolated with its own resources
- Client Hub manages concurrent requests to downstream servers