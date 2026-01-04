# Runbyte MCP

A Model Context Protocol (MCP) server implementing the **code execution pattern** for efficient agent-tool interactions. Instead of calling MCP tools directly through the model's context, Runbyte enables AI agents to write and execute TypeScript code in a sandboxed environment. It automatically translates downstream MCP servers into typed TypeScript modules accessible at `/servers/` in a virtual filesystem, allowing agents to discover tools on-demand, process data efficiently, and compose complex workflows all while dramatically reducing token consumption.

## Key Features

- **Compile MCP to TypeScript** - Access any MCP server as a typed TypeScript module
- **Virtual File System** - Discover and explore tools at `/servers/`
- **Auto Recompilation** - Detects and recompiles when connected MCP tools change
- **Sandboxed execution** - Secure WebAssembly runtime for code execution
- **High performance** - Built in Go for speed and reliability

## Table of Contents

- [Why Runbyte?](#why-runbyte)
- [Requirements](#requirements)
- [Getting Started](#getting-started)
- [Quick Example](#quick-example)
- [Installation by Client](#installation-by-client)
- [Runbyte Configuration](#runbyte-configuration)
- [Tools](#tools)
- [Benefits](#benefits)
- [Usage Workflow](#usage-workflow)
- [Running Runbyte](#running-runbyte)
- [How It Works](#how-it-works)
- [Architecture](#architecture)
- [Roadmap](#roadmap)
- [Troubleshooting](#troubleshooting)
- [Acknowledgments](#acknowledgments)
- [License](#license)

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
- **Go 1.21+**, **Node.js 18+** and **[extism-js compiler](https://github.com/extism/js-pdk?tab=readme-ov-file#install-the-compiler)** (for building from source)
- **MCP client**: VS Code, Cursor, Windsurf, Claude Desktop, Goose, Zed, or any other MCP-compatible client

## Getting Started

### Step 1: Create Runbyte Configuration

**First, create your Runbyte config file at `~/.runbyte/config.json`:**

This file tells Runbyte which downstream MCP servers to connect to and make available as TypeScript modules.

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

**Using a custom config location:**

You can specify a custom config file path using the `-config` flag:

```bash
./runbyte -config /path/to/runbyte.json -transport stdio
```

Or with Docker:

```bash
docker run -i --rm \
  -v /path/to/runbyte.json:/app/runbyte.json \
  yousuf64/runbyte:latest \
  -config /app/runbyte.json \
  -transport stdio
```

### Step 2: Configure Your MCP Client

**Add Runbyte to your MCP client configuration:**

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
        "${env:HOME}/.runbyte/config.json:/root/.runbyte/config.json",
        "yousuf64/runbyte:latest",
        "-transport",
        "stdio"
      ]
    }
  }
}
```

See [Installation by Client](#installation-by-client) below for client-specific instructions.

## Quick Example

Here's a complete workflow showing how to use Runbyte:

**1. Discover available servers:**

Ask your AI agent: "What MCP servers are available?"

The agent uses the `list_directory` tool:
```json
{
  "path": "/servers"
}
```

Response:
```
/servers/
  ├── filesystem/ (14 functions)
  ├── github/ (40 functions)
  ├── google-drive/ (20 functions)
  ├── slack/ (21 functions)
  └── index.ts
```

**2. Explore the GitHub server:**

Ask: "What tools does the GitHub server have?"

The agent lists the `/servers/github` directory:
```json
{
  "path": "/servers/github"
}
```

Response:
```
/servers/github/
  ├── listCommits.ts
  ├── issueRead.ts
  ├── createPullRequest.ts
  ├── createRepository.ts
  └── index.ts
```

**3. Read a specific tool definition from the `filesystem` server:**

The agent reads the `/servers/filesystem/readTextFile.ts` file to see its signature:
```json
{
  "path": "/servers/filesystem/readTextFile.ts"
}
```

Response:
```typescript
export interface ReadTextFileArgs {
    path: string;
    /** If provided, returns only the last N lines of the file */
    tail?: number;
    /** If provided, returns only the first N lines of the file */
    head?: number;
}

export interface ReadTextFileResult {
    content: string;
}

/**
 * Read the complete contents of a file from the file system as text. Handles various text encodings and provides detailed error messages if the file cannot be read. Use this tool when you need to examine the contents of a single file. Use the 'head' parameter to read only the first N lines of a file, or the 'tail' parameter to read only the last N lines of a file. Operates on the file as text regardless of extension. Only works within allowed directories.
 *
 * Returns parsed response - structure depends on tool implementation.
 */
export async function readTextFile(args: ReadTextFileArgs): Promise<ReadTextFileResult> {
    return await callTool("filesystem", "read_text_file", args);
}
```

**4. Execute code to use the tools:**

Ask: "Get all public repositories for 'octocat' and show me the top 3 most starred"

The agent executes:
```typescript
import * as github from './servers/github';

async function exec() {
  const repos = await github.listRepos({ 
    owner: "octocat",
    type: "public"
  });
  
  // Sort by stars and get top 3
  const topRepos = repos
    .sort((a, b) => b.stargazers_count - a.stargazers_count)
    .slice(0, 3)
    .map(r => ({
      name: r.name,
      stars: r.stargazers_count,
      url: r.html_url
    }));
  
  return { 
    total: repos.length,
    topRepos
  };
}
```

Result:
```json
{
  "total": 8,
  "topRepos": [
    {
      "name": "Hello-World",
      "stars": 2150,
      "url": "https://github.com/octocat/Hello-World"
    },
    {
      "name": "Spoon-Knife",
      "stars": 543,
      "url": "https://github.com/octocat/Spoon-Knife"
    },
    {
      "name": "test-repo",
      "stars": 89,
      "url": "https://github.com/octocat/test-repo"
    }
  ]
}
```

The agent sees only the summary (8 total repos, top 3 with filtered fields) instead of all repository data with every field, saving context tokens while providing the exact information needed.

## Installation by Client

Runbyte works with any MCP-compatible client. Choose your client below for specific setup instructions:

<details>
<summary><strong>VS Code</strong></summary>

#### stdio mode (recommended)

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
        "${env:HOME}/.runbyte/config.json:/root/.runbyte/config.json",
        "yousuf64/runbyte:latest",
        "-transport",
        "stdio"
      ]
    }
  }
}
```

#### HTTP mode

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

</details>

<details>
<summary><strong>Cursor</strong></summary>

#### stdio mode (recommended)

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
        "${env:HOME}/.runbyte/config.json:/root/.runbyte/config.json",
        "yousuf64/runbyte:latest",
        "-transport",
        "stdio"
      ]
    }
  }
}
```

#### HTTP mode

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

</details>

<details>
<summary><strong>Claude Desktop</strong></summary>

#### stdio mode

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
        "${env:HOME}/.runbyte/config.json:/root/.runbyte/config.json",
        "yousuf64/runbyte:latest",
        "-transport",
        "stdio"
      ]
    }
  }
}
```

#### HTTP mode

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

</details>

<details>
<summary><strong>Windsurf</strong></summary>

#### stdio mode (recommended)

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
        "${env:HOME}/.runbyte/config.json:/root/.runbyte/config.json",
        "yousuf64/runbyte:latest",
        "-transport",
        "stdio"
      ]
    }
  }
}
```

#### HTTP mode

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

</details>

<details>
<summary><strong>Goose</strong></summary>

#### stdio mode

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
        "${env:HOME}/.runbyte/config.json:/root/.runbyte/config.json",
        "yousuf64/runbyte:latest",
        "-transport",
        "stdio"
      ]
    }
  }
}
```

#### HTTP mode

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

</details>

<details>
<summary><strong>Other MCP Clients</strong></summary>

Runbyte works with any MCP-compatible client. Use the stdio configuration shown above, or HTTP mode if your client requires it.

**stdio mode template:**
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
        "${env:HOME}/.runbyte/config.json:/root/.runbyte/config.json",
        "yousuf64/runbyte:latest",
        "-transport",
        "stdio"
      ]
    }
  }
}
```

**HTTP mode template:**
```json
{
  "mcpServers": {
    "runbyte": {
      "url": "http://localhost:3000"
    }
  }
}
```

</details>

## Runbyte Configuration

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
3. **Generate**: Creates TypeScript modules for each server at `/servers/` with full type information
4. **Cache**: Stores generated code in a session-based cache for performance
5. **Monitor**: Watches for tool changes and regenerates modules automatically
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
│   └── index.ts
```

### Session Caching

Runbyte caches generated TypeScript modules per session for optimal performance:

- Cache is created when a session starts
- Automatically invalidated when downstream tools change
- Tools are regenerated on-demand when changes are detected
- Reduces latency for repeated tool discovery

## Architecture

```
┌──────────────┐
│  MCP Client  │  (VS Code / Cursor / Claude Desktop)
└──────┬───────┘
       │ stdio/HTTP/SSE
       ▼
┌─────────────────────────────────────────────────┐
│              Runbyte Server                     │
│                                                 │
│  Agent Code → Bundler → WASM Sandbox           │
│                  ↓           ↓                  │
│            Virtual FS    MCP Tools              │
│            /servers/    (isolated)              │
└─────────────────────┬───────────────────────────┘
                      │
        ┌─────────────┼─────────────┐
        ▼             ▼             ▼
   ┌────────┐    ┌────────┐    ┌────────┐
   │ GitHub │    │FileSys │    │ Slack  │  ...
   │  MCP   │    │  MCP   │    │  MCP   │
   └────────┘    └────────┘    └────────┘
```

Runbyte processes AI-generated code through five core components:

1. **MCP Client Hub** - Manages connections to downstream MCP servers (stdio/HTTP/SSE)
2. **Code Generator** - Introspects MCP tools and generates typed TypeScript modules
3. **Virtual Filesystem** - Stores generated code at `/servers/` with session-based caching
4. **Bundler (Rspack)** - Bundles user code with modules using ultra-fast SWC transpilation
5. **WASM Sandbox (QuickJS)** - Executes code securely with 30s timeout and no host access

The sandbox executes code in complete isolation—no filesystem, no network, no Node.js built-ins—routing all tool calls through validated MCP channels. This ensures secure, efficient execution while dramatically reducing context token consumption.

**For detailed architecture documentation including component responsibilities, data flows, transport layers, and security model, see [ARCHITECTURE.md](ARCHITECTURE.md).**

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
import * as fs from '@runbyte/fs';

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
  -v ${env:HOME}/.runbyte/config.json:/app/runbyte.json `
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

Similar findings have been reported by [Cloudflare's "Code Mode"](https://blog.cloudflare.com/code-mode/) implementation.

## License

MIT

