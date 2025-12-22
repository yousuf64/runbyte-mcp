package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yousuf/codebraid-mcp/internal/client"
	"github.com/yousuf/codebraid-mcp/internal/codegen"
	"github.com/yousuf/codebraid-mcp/internal/config"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Parse flags
	configPath := flag.String("config", "", "Path to config.json (required)")
	outputDir := flag.String("output-dir", "./generated", "Directory to write TypeScript files")
	serverFilter := flag.String("server", "", "Generate only for specific server(s), comma-separated")
	verbose := flag.Bool("verbose", false, "Enable verbose output")
	flag.Parse()

	if *configPath == "" {
		flag.Usage()
		return fmt.Errorf("--config is required")
	}

	ctx := context.Background()

	// Load config
	if *verbose {
		fmt.Printf("Loading config from: %s\n", *configPath)
	}
	cfg, err := config.Load(*configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create ClientBox and connect to MCP servers
	if *verbose {
		fmt.Println("Connecting to MCP servers...")
	}
	clientBox := client.NewClientBox()
	if err := clientBox.Connect(ctx, cfg); err != nil {
		return fmt.Errorf("failed to connect to MCP servers: %w", err)
	}
	defer clientBox.Close()

	// Create introspector
	introspector := codegen.NewIntrospector(clientBox)

	// Determine which servers to process
	var serversToProcess []string
	if *serverFilter != "" {
		serversToProcess = strings.Split(*serverFilter, ",")
		for i := range serversToProcess {
			serversToProcess[i] = strings.TrimSpace(serversToProcess[i])
		}
	} else {
		serversToProcess = introspector.ListServers()
	}

	if *verbose {
		fmt.Printf("Processing servers: %v\n", serversToProcess)
	}

	// Introspect tools
	if *verbose {
		fmt.Println("Discovering tools...")
	}
	allTools := make([]codegen.ToolDefinition, 0)
	for _, serverName := range serversToProcess {
		tools, err := introspector.IntrospectServer(ctx, serverName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to introspect server %q: %v\n", serverName, err)
			continue
		}
		allTools = append(allTools, tools...)
		if *verbose {
			fmt.Printf("  %s: %d tools\n", serverName, len(tools))
		}
	}

	if len(allTools) == 0 {
		return fmt.Errorf("no tools found")
	}

	// Group by server
	grouped := codegen.GroupByServer(allTools)

	// Create output directory
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate TypeScript files
	generator := codegen.NewTypeScriptGenerator()

	generatedServers := make([]string, 0, len(grouped))

	for serverName, tools := range grouped {
		if *verbose {
			fmt.Printf("Generating %s.ts...\n", serverName)
		}

		content, err := generator.GenerateFile(serverName, tools)
		if err != nil {
			return fmt.Errorf("failed to generate file for %q: %w", serverName, err)
		}

		outputPath := filepath.Join(*outputDir, serverName+".ts")
		if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", outputPath, err)
		}

		generatedServers = append(generatedServers, serverName)
	}

	// Generate mcp-types.ts
	if *verbose {
		fmt.Println("Generating mcp-types.ts...")
	}
	mcpTypesContent := generator.GenerateMCPTypesFile()
	mcpTypesPath := filepath.Join(*outputDir, "mcp-types.ts")
	if err := os.WriteFile(mcpTypesPath, []byte(mcpTypesContent), 0644); err != nil {
		return fmt.Errorf("failed to write mcp-types.ts: %w", err)
	}

	// Generate index.ts
	if *verbose {
		fmt.Println("Generating index.ts...")
	}
	indexContent := generator.GenerateIndexFile(generatedServers)
	indexPath := filepath.Join(*outputDir, "index.ts")
	if err := os.WriteFile(indexPath, []byte(indexContent), 0644); err != nil {
		return fmt.Errorf("failed to write index.ts: %w", err)
	}

	fmt.Printf("\n✓ Successfully generated TypeScript definitions for %d servers\n", len(generatedServers))
	fmt.Printf("  Output directory: %s\n", *outputDir)
	fmt.Println("\nGenerated files:")
	fmt.Println("  - mcp-types.ts")
	fmt.Println("  - index.ts")
	for _, server := range generatedServers {
		fmt.Printf("  - %s.ts\n", server)
	}

	return nil
}
