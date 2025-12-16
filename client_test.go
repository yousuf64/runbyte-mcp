package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestClient(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 99999999*time.Second)
	defer cancel()

	// Get server URL from environment or use default
	serverURL := os.Getenv("MCP_SERVER_URL")
	if serverURL == "" {
		serverURL = "http://localhost:3000/mcp"
	}

	log.Printf("Connecting to CodeBraid MCP server at %s", serverURL)

	// Create a new MCP client
	client := mcp.NewClient(&mcp.Implementation{
		Name:    "codebraid-test-client",
		Version: "1.0.0",
	}, &mcp.ClientOptions{})

	// Connect to the server using HTTP transport
	transport := &mcp.StreamableClientTransport{Endpoint: serverURL}
	session, err := client.Connect(ctx, transport, &mcp.ClientSessionOptions{})
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer session.Close()

	log.Printf("Connected! Session ID: %s", session.ID())

	// Test 1: List all tools
	log.Println("\n=== Test 1: Listing all available tools ===")
	toolsResult, err := session.ListTools(ctx, &mcp.ListToolsParams{})
	if err != nil {
		log.Fatalf("Failed to list tools: %v", err)
	}

	log.Printf("Found %d tools:", len(toolsResult.Tools))
	for _, tool := range toolsResult.Tools {
		log.Printf("  - %s: %s", tool.Name, tool.Description)
	}

	// Test 2: Call get_mcp_tools
	log.Println("\n=== Test 2: Calling get_mcp_tools ===")
	getMcpToolsResult, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "get_mcp_tools",
		Arguments: map[string]interface{}{
			"withDescription": true,
		},
	})
	if err != nil {
		log.Fatalf("Failed to call get_mcp_tools: %v", err)
	}

	if getMcpToolsResult.IsError {
		log.Fatalf("Tool returned error")
	}

	for _, content := range getMcpToolsResult.Content {
		if textContent, ok := content.(*mcp.TextContent); ok {
			log.Println("Downstream MCP tools:")

			// Pretty print the JSON
			var toolsMap map[string]interface{}
			if err := json.Unmarshal([]byte(textContent.Text), &toolsMap); err == nil {
				prettyJSON, _ := json.MarshalIndent(toolsMap, "", "  ")
				fmt.Println(string(prettyJSON))
			} else {
				fmt.Println(textContent.Text)
			}
		}
	}

	// Test 3: Call get_mcp_tool_details
	log.Println("\n=== Test 3: Calling get_mcp_tool_details ===")

	// First, let's get a tool name from the downstream servers
	// We'll try to get details for the first tool we find
	var firstToolName string
	if len(getMcpToolsResult.Content) > 0 {
		if textContent, ok := getMcpToolsResult.Content[0].(*mcp.TextContent); ok {
			var toolsMap map[string]interface{}
			if err := json.Unmarshal([]byte(textContent.Text), &toolsMap); err == nil {
				// Get the first server's first tool
				for _, tools := range toolsMap {
					if toolsList, ok := tools.(map[string]interface{}); ok {
						for toolName := range toolsList {
							firstToolName = toolName
							break
						}
						if firstToolName != "" {
							break
						}
					}
				}
			}
		}
	}

	if firstToolName != "" {
		log.Printf("Getting details for tool: %s", firstToolName)
		toolDetailsResult, err := session.CallTool(ctx, &mcp.CallToolParams{
			Name: "get_mcp_tool_details",
			Arguments: map[string]interface{}{
				"toolName": firstToolName,
			},
		})
		if err != nil {
			log.Printf("Failed to get tool details: %v", err)
		} else if toolDetailsResult.IsError {
			log.Println("Tool returned error")
		} else {
			for _, content := range toolDetailsResult.Content {
				if textContent, ok := content.(*mcp.TextContent); ok {
					log.Println("Tool details:")

					var detailsMap map[string]interface{}
					if err := json.Unmarshal([]byte(textContent.Text), &detailsMap); err == nil {
						prettyJSON, _ := json.MarshalIndent(detailsMap, "", "  ")
						fmt.Println(string(prettyJSON))
					} else {
						fmt.Println(textContent.Text)
					}
				}
			}
		}
	} else {
		log.Println("No downstream tools found to get details for")
	}

	// Test 4: Try execute_code (will fail if plugin.wasm not built)
	log.Println("\n=== Test 4: Calling execute_code (optional) ===")
	executeCodeResult, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "execute_code",
		Arguments: map[string]interface{}{
			"code": `
				return (async () => {
					try {	
						console.log("Starting Google search for Gustave...");
						
						// Navigate to Google
						console.log("Step 1: Navigating to Google...");
						const navResult = await callTool("playwright", "browser_navigate", { url: "https://www.google.com" });
						console.log("Navigation result:", JSON.stringify(navResult));
						
						// Wait for page to load
						console.log("Step 2: Waiting for page to load...");
						await callTool("playwright", "browser_wait_for", { time: 2 });
						
						// Take snapshot to see page structure
						console.log("Step 3: Taking snapshot...");
						const snapshot = await callTool("playwright", "browser_snapshot", {});
						console.log("Snapshot received");
						
						return {
						  status: "success",
						  message: "Navigation completed",
						  snapshot: snapshot
						};
					} catch (error) {
						console.error("Error:", error);
						return {
							status: "error",
							error: error.message || String(error)
						};
					}
				})();
	
				// const result = {
				//	message: "Hello from CodeBraid MCP!",
				//	timestamp: new Date().toISOString(),
				//	test: "success"
				//};
				//return result;
			`,
		},
	})
	if err != nil {
		log.Printf("⚠️  execute_code failed (expected if plugin.wasm not built): %v", err)
	} else if executeCodeResult.IsError {
		log.Println("⚠️  execute_code returned error (expected if plugin.wasm not built)")
	} else {
		log.Println("✅ execute_code succeeded!")
		for _, content := range executeCodeResult.Content {
			if textContent, ok := content.(*mcp.TextContent); ok {
				log.Println("Result:")

				var resultMap map[string]interface{}
				if err := json.Unmarshal([]byte(textContent.Text), &resultMap); err == nil {
					prettyJSON, _ := json.MarshalIndent(resultMap, "", "  ")
					fmt.Println(string(prettyJSON))
				} else {
					fmt.Println(textContent.Text)
				}
			}
		}
	}

	log.Println("\n=== All tests completed! ===")
}
