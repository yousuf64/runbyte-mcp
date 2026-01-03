# Runbyte MCP

A Model Context Protocol (MCP) server that enables AI agents to execute TypeScript code with access to multiple downstream MCP servers. It automatically translates MCP tools into TypeScript libraries accessible at `/servers/` in a virtual filesystem.

## Key Features

- 🔄 **Automatic MCP-to-TypeScript translation** - Access any MCP server as a typed TypeScript library
- 📁 **Virtual filesystem** - Discover and explore tools at `/servers/` 
- ⚡ **Session-based caching** - Fast performance with intelligent cache invalidation
- 🔔 **Auto-refresh** - Detects and reloads when downstream tools change
- 🔒 **Sandboxed execution** - Secure WebAssembly runtime for code execution
- 🚀 **High performance** - Built in Go for speed and reliability

## Why Runbyte?

Traditional MCP clients load all tool definitions directly into the model's context and pass every intermediate result through the model. This becomes inefficient at scale:

**The Problem:**
- **Tool overload**: With dozens of MCP servers, you might load 150,000+ tokens of tool definitions before processing any request
- **Data bloat**: Large documents and datasets (like a 10,000-row spreadsheet) flow through the model multiple times
- **Repeated processing**: Every intermediate result passes through the context window, even when just moving data between tools
- **Higher costs**: More tokens = slower responses and higher API costs

**The Solution:**

Runbyte implements the code execution pattern described in [Anthropic's research on Code Execution with MCP](https://www.anthropic.com/engineering/code-execution-with-mcp), achieving up to **98.7% token reduction** compared to traditional tool calling:

- **Load only what you need**: Explore the virtual filesystem to discover and load just the tools required for your task
- **Process data in the sandbox**: Filter, transform, and aggregate data in the execution environment before returning results
- **Write familiar code**: Use loops, conditionals, async/await, and other programming patterns instead of chaining tool calls

Instead of loading 150,000 tokens of tool definitions, agents might load just 2,000 tokens—reducing time and cost by 98.7%. [Cloudflare reported similar findings](https://blog.cloudflare.com/code-mode/) with their "Code Mode" implementation.

## Requirements

- **Docker** (recommended), or
- **Go 1.21+** and **Node.js 18+** (for building from source)
- **MCP client**: VS Code, Cursor, Windsurf, Claude Desktop, Goose, Zed, or any other MCP-compatible client

## Getting Started

### Standard Configuration

This configuration works in most MCP clients and uses Docker with stdio transport:

```json
{
  "mcpServers": {
    "runbyte": {
      "command": "docker",
      "args": [
        "run",
        "-i",
        "--rm",
        "-v",
        "${HOME}/.runbyte/config.json:/root/.runbyte/config.json",
        "yousuf64/runbyte:latest",
        "-transport",
        "stdio"
      ]
    }
  }
}
```

### Initial Setup

**Before using Runbyte, create your config file at `~/.runbyte/config.json`:**

```json
{
  "mcpServers": {
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "/tmp"]
    }
  }
}
```

This config tells Runbyte which MCP servers to connect to and make available as TypeScript libraries.

## Installation by Client

### VS Code

**stdio mode (recommended):**

```json
{
  "mcpServers": {
    "runbyte": {
      "command": "docker",
      "args": [
        "run",
        "-i",
        "--rm",
        "-v",
        "${HOME}/.runbyte/config.json:/root/.runbyte/config.json",
        "yousuf64/runbyte:latest",
        "-transport",
        "stdio"
      ]
    }
  }
}
```

**HTTP mode:**

First, start the Runbyte server:
```bash
docker run -d -p 3000:3000 \
  -v ~/.runbyte/config.json:/app/runbyte.json \
  yousuf64/runbyte:latest \
  -transport http -port 3000
```

Then configure VS Code:
```json
{
  "mcpServers": {
    "runbyte": {
      "url": "http://localhost:3000"
    }
  }
}
```

### Cursor

**stdio mode (recommended):**

```json
{
  "mcpServers": {
    "runbyte": {
      "command": "docker",
      "args": [
        "run",
        "-i",
        "--rm",
        "-v",
        "${HOME}/.runbyte/config.json:/root/.runbyte/config.json",
        "yousuf64/runbyte:latest",
        "-transport",
        "stdio"
      ]
    }
  }
}
```

**HTTP mode:**

First, start the Runbyte server:
```bash
docker run -d -p 3000:3000 \
  -v ~/.runbyte/config.json:/app/runbyte.json \
  yousuf64/runbyte:latest \
  -transport http -port 3000
```

Then configure Cursor:
```json
{
  "mcpServers": {
    "runbyte": {
      "url": "http://localhost:3000"
    }
  }
}
```

### Claude Desktop

**stdio mode:**

Add to `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS) or `%APPDATA%/Claude/claude_desktop_config.json` (Windows):

```json
{
  "mcpServers": {
    "runbyte": {
      "command": "docker",
      "args": [
        "run",
        "-i",
        "--rm",
        "-v",
        "${HOME}/.runbyte/config.json:/root/.runbyte/config.json",
        "yousuf64/runbyte:latest",
        "-transport",
        "stdio"
      ]
    }
  }
}
```

**HTTP mode:**

First, start the Runbyte server:
```bash
docker run -d -p 3000:3000 \
  -v ~/.runbyte/config.json:/app/runbyte.json \
  yousuf64/runbyte:latest \
  -transport http -port 3000
```

Then configure Claude Desktop:
```json
{
  "mcpServers": {
    "runbyte": {
      "url": "http://localhost:3000"
    }
  }
}
```

### Windsurf

**stdio mode (recommended):**

```json
{
  "mcpServers": {
    "runbyte": {
      "command": "docker",
      "args": [
        "run",
        "-i",
        "--rm",
        "-v",
        "${HOME}/.runbyte/config.json:/root/.runbyte/config.json",
        "yousuf64/runbyte:latest",
        "-transport",
        "stdio"
      ]
    }
  }
}
```

**HTTP mode:**

First, start the Runbyte server:
```bash
docker run -d -p 3000:3000 \
  -v ~/.runbyte/config.json:/app/runbyte.json \
  yousuf64/runbyte:latest \
  -transport http -port 3000
```

Then configure Windsurf:
```json
{
  "mcpServers": {
    "runbyte": {
      "url": "http://localhost:3000"
    }
  }
}
```

### Goose

**stdio mode:**

Add to your Goose configuration:

```json
{
  "mcpServers": {
    "runbyte": {
      "command": "docker",
      "args": [
        "run",
        "-i",
        "--rm",
        "-v",
        "${HOME}/.runbyte/config.json:/root/.runbyte/config.json",
        "yousuf64/runbyte:latest",
        "-transport",
        "stdio"
      ]
    }
  }
}
```

**HTTP mode:**

First, start the Runbyte server:
```bash
docker run -d -p 3000:3000 \
  -v ~/.runbyte/config.json:/app/runbyte.json \
  yousuf64/runbyte:latest \
  -transport http -port 3000
```

Then configure Goose:
```json
{
  "mcpServers": {
    "runbyte": {
      "url": "http://localhost:3000"
    }
  }
}
```

### Other MCP Clients

Runbyte works with any MCP-compatible client. Use the stdio configuration shown above, or HTTP mode if your client requires it.

## Configuration

### Config File Locations

Runbyte searches for configuration files in this order:

1. `./runbyte.json` - Current directory
2. `~/.config/runbyte/config.json` - XDG config directory
3. `~/.runbyte/config.json` - Home directory (recommended)
4. `/etc/runbyte/config.json` - System-wide config
5. Custom path via `-config` flag

### Basic Configuration Structure

```json
{
  "mcpServers": {
    "serverName": {
      "command": "node",
      "args": ["server.js"]
    }
  }
}
```

### MCP Server Configuration

#### stdio servers (command/args)

```json
{
  "mcpServers": {
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "/tmp"],
      "env": {
        "NODE_ENV": "production"
      },
      "cwd": "/path/to/working/directory"
    }
  }
}
```

#### HTTP servers (url)

```json
{
  "mcpServers": {
    "github": {
      "url": "https://api.github.com/mcp",
      "headers": {
        "Authorization": "Bearer ${GITHUB_TOKEN}"
      }
    }
  }
}
```

#### SSE servers

```json
{
  "mcpServers": {
    "sse-server": {
      "type": "sse",
      "url": "https://example.com/sse"
    }
  }
}
```

### Server Options

Configure Runbyte's HTTP server and execution timeouts:

```json
{
  "server": {
    "port": 3000,
    "timeout": 30
  },
  "mcpServers": {
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "/tmp"]
    }
  }
}
```

### Environment Variables

Use environment variable substitution in your configuration:

```json
{
  "mcpServers": {
    "github": {
      "url": "https://api.github.com/mcp",
      "headers": {
        "Authorization": "Bearer ${GITHUB_TOKEN}"
      }
    },
    "custom": {
      "command": "${CUSTOM_MCP_COMMAND}",
      "args": ["${CUSTOM_ARG}"]
    }
  }
}
```

## Tools

Runbyte provides three main tools for interacting with the virtual filesystem and executing code:

### `list_directory`

List contents of a directory in the virtual filesystem.

**Parameters:**
- `path` (string, required): Directory path (e.g., `/`, `/servers`, `/servers/github`)
- `withDescriptions` (boolean, optional): Include function descriptions (default: false)

**Examples:**

List all available servers:
```json
{
  "path": "/servers"
}
```

List tools in a specific server:
```json
{
  "path": "/servers/github"
}
```

List with descriptions:
```json
{
  "path": "/servers/filesystem",
  "withDescriptions": true
}
```

**Response:** Returns a list of files and directories with their types.

### `read_file`

Read a TypeScript library file from the virtual filesystem.

**Parameters:**
- `path` (string, required): Absolute file path (e.g., `/servers/github/index.ts`, `/servers/github/listRepos.ts`)

**Examples:**

Read server index:
```json
{
  "path": "/servers/github/index.ts"
}
```

Read specific function:
```json
{
  "path": "/servers/github/listRepos.ts"
}
```

**Response:** Returns the TypeScript source code with full type information.

### `execute_code`

Execute TypeScript code in a sandboxed environment with access to all configured MCP servers.

**Parameters:**
- `code` (string, required): TypeScript code containing an `exec()` function

**Requirements:**
- Must define an `exec()` function as the entry point
- Use namespace imports: `import * as name from './servers/name'`
- All imports are automatically bundled
- 30 second execution timeout
- No access to Node.js built-ins or filesystem
- No access to DOM or browser APIs

**Examples:**

Basic example:
```typescript
import * as github from './servers/github';

async function exec() {
  const repos = await github.listRepos({ owner: "octocat" });
  return { count: repos.length };
}
```

Multi-server workflow:
```typescript
import * as github from './servers/github';
import * as filesystem from './servers/filesystem';

async function exec() {
  // Fetch data from GitHub
  const repos = await github.listRepos({ owner: "octocat" });
  
  // Read local config
  const config = await filesystem.readFile({ 
    path: "/tmp/config.json" 
  });
  
  // Process and return
  return {
    totalRepos: repos.length,
    configSize: config.length
  };
}
```

With strong typing:
```typescript
import * as github from './servers/github';
import * as filesystem from './servers/filesystem';

interface ExecResult {
  totalRepos: number;
  readmeLength: number;
  timestamp: string;
}

async function exec(): Promise<ExecResult> {
  const repos = await github.listRepos({ owner: "octocat" });
  const readme = await filesystem.readFile({ 
    path: "/tmp/README.md" 
  });
  
  return {
    totalRepos: repos.length,
    readmeLength: readme.length,
    timestamp: new Date().toISOString()
  };
}
```

Error handling:
```typescript
import * as github from './servers/github';

async function exec() {
  try {
    const repos = await github.listRepos({ owner: "octocat" });
    return { success: true, repos };
  } catch (error) {
    return { 
      success: false, 
      error: error instanceof Error ? error.message : String(error)
    };
  }
}
```

## Benefits

### Progressive Tool Discovery

Agents explore the virtual filesystem to find tools on-demand, loading only what they need for the current task instead of all definitions upfront.

**Without Runbyte:** Load all 150,000 tokens of tool definitions before starting

**With Runbyte:**
```typescript
// 1. List available servers (minimal tokens)
{ "path": "/servers" }
// Response: ["github/", "filesystem/", "slack/"]

// 2. Load only the tools you need
{ "path": "/servers/github/index.ts" }
```

### Context-Efficient Data Processing

Process and filter data in the execution environment before returning to the model. A 10,000-row spreadsheet can be filtered down to 5 relevant rows before consuming context.

**Example: Filter large datasets**
```typescript
import * as gdrive from './servers/gdrive';

async function exec() {
  // Fetch 10,000 rows (stays in execution environment)
  const allRows = await gdrive.getSheet({ sheetId: 'abc123' });
  
  // Filter to only what matters
  const pendingOrders = allRows.filter(row => 
    row["Status"] === 'pending' && row["Amount"] > 1000
  );
  
  // Return only summary (minimal context consumption)
  return {
    total: allRows.length,
    pending: pendingOrders.length,
    topOrders: pendingOrders.slice(0, 5)
  };
}
```

The model sees a small summary instead of 10,000 rows—saving tokens and costs.

### Powerful Async Control Flow

Use familiar programming patterns—async/await, loops, conditionals, error handling—instead of chaining individual tool calls through the agent loop.

**Example: Async polling with control flow**
```typescript
import * as slack from './servers/slack';

async function exec() {
  let deploymentComplete = false;
  let attempts = 0;
  const maxAttempts = 10;
  
  while (!deploymentComplete && attempts < maxAttempts) {
    const messages = await slack.getChannelHistory({ 
      channel: 'C123456' 
    });
    
    deploymentComplete = messages.some(m => 
      m.text.includes('deployment complete')
    );
    
    if (!deploymentComplete) {
      attempts++;
      await new Promise(resolve => setTimeout(resolve, 5000));
    }
  }
  
  return { 
    found: deploymentComplete, 
    attempts,
    message: deploymentComplete 
      ? 'Deployment successful' 
      : 'Timeout waiting for deployment'
  };
}
```

**Example: Parallel async operations**
```typescript
import * as github from './servers/github';
import * as jira from './servers/jira';

async function exec() {
  // Execute multiple async operations in parallel
  const [githubIssues, jiraTickets, prList] = await Promise.all([
    github.listIssues({ owner: 'octocat', repo: 'hello-world' }),
    jira.searchIssues({ jql: 'project = PROJ AND status = Open' }),
    github.listPullRequests({ owner: 'octocat', repo: 'hello-world' })
  ]);
  
  // Process results together
  return {
    githubIssues: githubIssues.length,
    jiraTickets: jiraTickets.length,
    openPRs: prList.length,
    total: githubIssues.length + jiraTickets.length + prList.length
  };
}
```

**Example: Error handling with async/await**
```typescript
import * as gdrive from './servers/gdrive';
import * as slack from './servers/slack';

async function exec() {
  const results = [];
  const errors = [];
  
  const docIds = ['doc1', 'doc2', 'doc3'];
  
  for (const docId of docIds) {
    try {
      const doc = await gdrive.getDocument({ documentId: docId });
      results.push({ docId, success: true, length: doc.content.length });
    } catch (error) {
      errors.push({ docId, error: error.message });
    }
  }
  
  // Notify if there were errors
  if (errors.length > 0) {
    await slack.sendMessage({
      channel: '#alerts',
      text: `Failed to process ${errors.length} documents`
    });
  }
  
  return { processed: results.length, failed: errors.length, results, errors };
}
```

This is more efficient than alternating between tool calls and sleep commands through the agent loop, and saves on "time to first token" latency.

### Privacy and Security

Intermediate data stays in the sandboxed execution environment. Only explicitly returned results enter the model's context, protecting sensitive information.

**Example: Privacy-preserving data flow**
```typescript
import * as gdrive from './servers/gdrive';
import * as salesforce from './servers/salesforce';

async function exec() {
  // Customer data stays in execution environment
  const sheet = await gdrive.getSheet({ sheetId: 'customer-data' });
  
  // Process sensitive data without exposing it to the model
  for (const row of sheet.rows) {
    await salesforce.updateRecord({
      objectType: 'Lead',
      recordId: row.salesforceId,
      data: { 
        Email: row.email,      // Never enters model context
        Phone: row.phone,      // Never enters model context  
        Name: row.name         // Never enters model context
      }
    });
  }
  
  // Only return non-sensitive summary
  return { 
    message: `Updated ${sheet.rows.length} customer records`,
    count: sheet.rows.length
  };
}
```

Sensitive data (emails, phone numbers, names) flows from Google Sheets to Salesforce without ever passing through the model's context.

## Usage Workflow

### 1. Discover Available Servers

```typescript
// List all servers in the virtual filesystem
{
  "path": "/servers"
}

// Response: ["github/", "filesystem/", "index.ts"]
```

### 2. Explore Server Capabilities

```typescript
// Read the server's index file
{
  "path": "/servers/github/index.ts"
}

// Or list available functions
{
  "path": "/servers/github"
}
```

### 3. Write and Execute Code

```typescript
import * as github from './servers/github';

async function exec() {
  const repos = await github.listRepos({ owner: "octocat" });
  return repos;
}
```

### 4. Use Multiple Servers Together

```typescript
import * as github from './servers/github';
import * as filesystem from './servers/filesystem';
import * as slack from './servers/slack';

async function exec() {
  // Fetch issues from GitHub
  const issues = await github.listIssues({ 
    owner: "octocat", 
    repo: "hello-world" 
  });
  
  // Save to file
  await filesystem.writeFile({ 
    path: "/tmp/issues.json",
    content: JSON.stringify(issues, null, 2)
  });
  
  // Send notification
  await slack.sendMessage({ 
    channel: "#dev",
    text: `Found ${issues.length} open issues`
  });
  
  return { success: true, issueCount: issues.length };
}
```

## Running Runbyte

### Using Docker (Recommended)

#### stdio mode

This is the recommended mode for most MCP clients:

```bash
# Pull the image
docker pull yousuf64/runbyte:latest

# Run in stdio mode (interactive)
docker run -i --rm \
  -v ~/.runbyte/config.json:/app/runbyte.json \
  yousuf64/runbyte:latest \
  -transport stdio
```

#### HTTP mode

Use HTTP mode when running on systems without display or from IDE worker processes:

```bash
# Run as daemon on port 3000
docker run -d \
  -p 3000:3000 \
  -v ~/.runbyte/config.json:/app/runbyte.json \
  --name runbyte \
  yousuf64/runbyte:latest \
  -transport http -port 3000

# Check logs
docker logs runbyte

# Stop the server
docker stop runbyte
```

#### Custom config location

```bash
# Use a different config file
docker run -i --rm \
  -v /path/to/custom/config.json:/app/runbyte.json \
  yousuf64/runbyte:latest \
  -transport stdio
```

### From Source

#### Prerequisites

- Go 1.21 or newer
- Node.js 18 or newer
- npm or yarn

#### Build and Run

```bash
# 1. Clone the repository
git clone https://github.com/yousuf/runbyte.git
cd runbyte

# 2. Install dependencies and build
make build

# 3. Create config file (if not exists)
mkdir -p ~/.runbyte
cat > ~/.runbyte/config.json << 'EOF'
{
  "mcpServers": {
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "/tmp"]
    }
  }
}
EOF

# 4. Run in stdio mode (for MCP clients)
./runbyte -transport stdio

# Or run in HTTP mode
./runbyte -transport http -port 3000

# Or specify custom config
./runbyte -config /path/to/config.json -transport stdio
```

#### Development

```bash
# Run tests
make test

# Build for specific platform
GOOS=linux GOARCH=amd64 make build

# Clean build artifacts
make clean
```

## How It Works

Runbyte acts as a bridge between AI agents and multiple MCP servers, providing a unified TypeScript interface:

1. **Connect**: Runbyte connects to all MCP servers defined in your config file
2. **Introspect**: It queries each server for available tools and their schemas
3. **Generate**: Creates TypeScript libraries for each server at `/servers/` with full type information
4. **Cache**: Stores generated code in a session-based cache for performance
5. **Monitor**: Watches for tool changes and regenerates libraries automatically
6. **Execute**: Runs your TypeScript code in a secure WebAssembly sandbox
7. **Route**: Routes function calls to the appropriate downstream MCP server
8. **Return**: Collects results and returns them to your code

### Virtual Filesystem Structure

```
/
├── servers/
│   ├── github/              (Generated from GitHub MCP server)
│   │   ├── listRepos.ts
│   │   ├── getIssues.ts
│   │   └── index.ts
│   ├── filesystem/          (Generated from Filesystem MCP server)
│   │   ├── readFile.ts
│   │   ├── writeFile.ts
│   │   └── index.ts
│   └── index.ts             (Main server index)
```

### Session Caching

Runbyte caches generated TypeScript libraries per session for optimal performance:

- Cache is created when a session starts
- Automatically invalidated when downstream tools change
- Tools are regenerated on-demand when changes are detected
- Reduces latency for repeated tool discovery

## Architecture

### System Overview

```
┌─────────────────────────────────────────────────────────────┐
│                      MCP Client                             │
│              (VS Code / Cursor / Claude Desktop)            │
└───────────────────────────┬─────────────────────────────────┘
                            │
                    stdio or HTTP transport
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                   Runbyte Server (Go)                       │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │            MCP Client Hub                           │   │
│  │  • Manages connections to downstream MCP servers   │   │
│  │  • Handles stdio/HTTP/SSE transports               │   │
│  └──────────────────┬──────────────────────────────────┘   │
│                     │                                       │
│                     ▼                                       │
│  ┌─────────────────────────────────────────────────────┐   │
│  │         Code Generator (Codegen)                    │   │
│  │  • Introspects MCP server tools                    │   │
│  │  • Converts JSON schemas to TypeScript types       │   │
│  │  • Generates typed function wrappers               │   │
│  └──────────────────┬──────────────────────────────────┘   │
│                     │                                       │
│                     ▼                                       │
│  ┌─────────────────────────────────────────────────────┐   │
│  │      Virtual Filesystem (/servers/)                 │   │
│  │  • Stores generated TypeScript libraries           │   │
│  │  • Provides list_directory and read_file tools     │   │
│  │  • Session-based caching with invalidation         │   │
│  └──────────────────┬──────────────────────────────────┘   │
│                     │                                       │
│                     ▼                                       │
│  ┌─────────────────────────────────────────────────────┐   │
│  │           Bundler (Rspack)                          │   │
│  │  • Bundles user code with generated libraries      │   │
│  │  • Resolves imports and dependencies               │   │
│  │  • Produces single executable bundle               │   │
│  └──────────────────┬──────────────────────────────────┘   │
│                     │                                       │
│                     ▼                                       │
│  ┌─────────────────────────────────────────────────────┐   │
│  │         WASM Sandbox (QuickJS)                      │   │
│  │  • Executes bundled TypeScript/JavaScript          │   │
│  │  • Isolated execution environment                  │   │
│  │  • 30-second timeout protection                    │   │
│  │  • No filesystem or network access                 │   │
│  └──────────────────┬──────────────────────────────────┘   │
│                     │                                       │
│                     │ Routes tool calls                     │
│                     ▼                                       │
│  ┌─────────────────────────────────────────────────────┐   │
│  │          MCP Client Hub (routing)                   │   │
│  └──────────────────┬──────────────────────────────────┘   │
│                     │                                       │
└─────────────────────┼───────────────────────────────────────┘
                      │
        ┌─────────────┼─────────────┬──────────────┐
        │             │             │              │
        ▼             ▼             ▼              ▼
   ┌─────────┐  ┌─────────┐  ┌─────────┐    ┌─────────┐
   │ GitHub  │  │FileSys  │  │ Slack   │ ...│ Custom  │
   │   MCP   │  │  MCP    │  │  MCP    │    │  MCP    │
   └─────────┘  └─────────┘  └─────────┘    └─────────┘
```

### Core Components

#### MCP Client Hub
**Purpose:** Manages connections to all downstream MCP servers

**Responsibilities:**
- Establishes and maintains connections to configured MCP servers
- Supports multiple transport types: stdio (command/args), HTTP (url), SSE
- Routes tool execution requests to appropriate servers
- Handles connection lifecycle, reconnection, and error recovery
- Manages concurrent requests across multiple servers

#### Code Generator (Codegen)
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

#### Virtual Filesystem
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

#### Bundler (Rspack)
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

#### WASM Sandbox (QuickJS)
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

### Data Flows

#### Flow 1: Tool Discovery & Code Generation

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

#### Flow 2: Code Execution

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

#### Flow 3: Cache Invalidation & Updates

```
1. MCP server tool definitions change
2. Notification sent to Runbyte (if supported)
   OR detected on next introspection
3. Session Manager → invalidates cache for that server
4. Code Generator → regenerates TypeScript files
5. Virtual Filesystem → updates with new code
6. Next execution → uses updated definitions
```

### Transport Layer

#### stdio Transport (Default)
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

#### HTTP Transport
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

#### Downstream MCP Servers
- Runbyte connects to downstream servers via their configured transport
- Supports stdio, HTTP, and SSE for downstream connections
- Each server can use different transport type
- Connection pooling for HTTP/SSE servers

### Security Model

**Sandbox Isolation:**
- Code runs in WebAssembly sandbox (QuickJS)
- No access to host filesystem
- No direct network access
- No Node.js built-in modules
- Only controlled access via MCP tool calls

**Resource Limits:**
- 30-second execution timeout (configurable)
- Memory limits enforced by WASM runtime
- No infinite loops or resource exhaustion

**Data Privacy:**
- Intermediate data stays in execution environment
- Only returned results enter model context
- Sensitive data never exposed to agent unless explicitly returned

**MCP Tool Access:**
- All tool calls go through validated routing
- Type safety enforced at TypeScript level
- Schema validation on tool inputs
- Error handling prevents sandbox escapes

## Roadmap

We're actively working on expanding Runbyte's capabilities:

### OAuth Authentication for MCP Servers

Support for OAuth authentication flows when connecting to MCP servers. This will enable secure authentication with services that require OAuth (Google Drive, GitHub, Salesforce, etc.) without manual token management.

### Persistent Workspace Storage

The sandbox execution environment will support a `/workspace` folder where agents can store and retrieve files across executions. This enables:

- **State persistence**: Save progress and resume work across sessions
- **Intermediate results**: Store large datasets without consuming context
- **Reusable skills**: Save working code as functions for future tasks

Example:
```typescript
import * as fs from './workspace/fs';

async function exec() {
  // Save data for later use
  await fs.writeFile('/workspace/analysis.json', JSON.stringify(data));
  
  // Retrieve in a future execution
  const saved = await fs.readFile('/workspace/analysis.json');
}
```

This aligns with the [Skills concept](https://docs.claude.com/en/docs/agents-and-tools/agent-skills/overview), allowing agents to build a library of reusable capabilities over time.

## Troubleshooting

### Config file not found

Ensure your config file exists at one of the supported locations:
- `~/.runbyte/config.json` (recommended)
- `~/.config/runbyte/config.json`
- `./runbyte.json`

Or specify explicitly with `-config`:
```bash
./runbyte -config /path/to/config.json -transport stdio
```

### Docker volume mount issues

On macOS/Windows, ensure the path is under your home directory or explicitly shared in Docker settings.

**macOS:**
```bash
docker run -i --rm \
  -v ~/.runbyte/config.json:/app/runbyte.json \
  yousuf64/runbyte:latest -transport stdio
```

**Windows (PowerShell):**
```powershell
docker run -i --rm `
  -v ${HOME}/.runbyte/config.json:/app/runbyte.json `
  yousuf64/runbyte:latest -transport stdio
```

### HTTP mode connection refused

Ensure the port is not already in use and properly exposed:
```bash
# Check if port is available
lsof -i :3000

# Run with different port
docker run -d -p 8080:8080 \
  -v ~/.runbyte/config.json:/app/runbyte.json \
  yousuf64/runbyte:latest \
  -transport http -port 8080
```

### Execution timeout

The default timeout is 30 seconds. For longer-running operations, increase the timeout in your config:
```json
{
  "server": {
    "timeout": 60
  },
  "mcpServers": {
    "...": "..."
  }
}
```

## Acknowledgments

Runbyte implements the code execution pattern described in Anthropic's research article ["Code execution with MCP: Building more efficient agents"](https://www.anthropic.com/engineering/code-execution-with-mcp). This approach enables agents to use context more efficiently by loading tools on-demand and processing data in a sandboxed environment, achieving up to 98.7% token reduction compared to traditional tool calling.

Similar findings have been reported by [Cloudflare's "Code Mode"](https://blog.cloudflare.com/code-mode/) implementation. We're grateful to the [MCP community](https://modelcontextprotocol.io/community) for building the ecosystem that makes this possible.

## License

MIT

