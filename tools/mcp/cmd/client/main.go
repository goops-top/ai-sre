package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run cmd/client/main.go <mode>")
		fmt.Println("Modes:")
		fmt.Println("  stdio - Test stdio transport")
		fmt.Println("  http  - Test HTTP transport")
		os.Exit(1)
	}

	mode := os.Args[1]
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	switch mode {
	case "stdio":
		testStdioTransport(ctx)
	case "http":
		testHTTPTransport(ctx)
	default:
		fmt.Printf("Unknown mode: %s\n", mode)
		os.Exit(1)
	}
}

func testStdioTransport(ctx context.Context) {
	fmt.Println("=== Testing MCP Server with stdio transport ===")
	
	// 创建 stdio 客户端
	mcpClient, err := client.NewStdioMCPClient("go", []string{"run", "cmd/mcp-server/main.go", "-transport", "stdio"})
	if err != nil {
		log.Fatalf("Failed to create stdio client: %v", err)
	}
	defer mcpClient.Close()
	
	// 初始化连接
	initRequest := mcp.InitializeRequest{
		Params: mcp.InitializeParams{
			ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION,
			ClientInfo: mcp.Implementation{
				Name:    "test-client",
				Version: "1.0.0",
			},
			Capabilities: mcp.ClientCapabilities{},
		},
	}
	
	_, err = mcpClient.Initialize(ctx, initRequest)
	if err != nil {
		log.Fatalf("Failed to initialize: %v", err)
	}
	
	fmt.Println("✓ Connected to MCP server via stdio")
	
	// 测试获取工具列表
	testToolsList(ctx, mcpClient)
	
	// 测试调用工具
	testToolCalls(ctx, mcpClient)
}

func testHTTPTransport(ctx context.Context) {
	fmt.Println("=== Testing MCP Server with HTTP transport ===")
	
	// 创建 HTTP 客户端
	mcpClient, err := client.NewStreamableHttpClient("http://localhost:8084/mcp")
	if err != nil {
		log.Fatalf("Failed to create HTTP client: %v", err)
	}
	defer mcpClient.Close()
	
	// 启动客户端
	if err := mcpClient.Start(ctx); err != nil {
		log.Fatalf("Failed to start HTTP client: %v", err)
	}
	
	// 初始化连接
	initRequest := mcp.InitializeRequest{
		Params: mcp.InitializeParams{
			ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION,
			ClientInfo: mcp.Implementation{
				Name:    "test-client",
				Version: "1.0.0",
			},
			Capabilities: mcp.ClientCapabilities{},
		},
	}
	
	_, err = mcpClient.Initialize(ctx, initRequest)
	if err != nil {
		log.Fatalf("Failed to initialize HTTP client: %v", err)
		fmt.Println("Make sure the server is running with: go run cmd/mcp-server/main.go -transport http -port 8082")
		return
	}
	
	fmt.Println("✓ Connected to MCP server via HTTP")
	
	// 测试获取工具列表
	testToolsList(ctx, mcpClient)
	
	// 测试调用工具
	testToolCalls(ctx, mcpClient)
}

func testToolsList(ctx context.Context, mcpClient client.MCPClient) {
	fmt.Println("\n--- Testing tools/list ---")
	
	// 获取工具列表
	toolsRequest := mcp.ListToolsRequest{}
	toolsResult, err := mcpClient.ListTools(ctx, toolsRequest)
	if err != nil {
		log.Printf("Failed to list tools: %v", err)
		return
	}
	
	fmt.Printf("✓ Found %d tools:\n", len(toolsResult.Tools))
	for i, tool := range toolsResult.Tools {
		fmt.Printf("  %d. %s - %s\n", i+1, tool.Name, tool.Description)
		// 检查 InputSchema 是否为空结构体
		if tool.InputSchema.Type != "" || len(tool.InputSchema.Properties) > 0 {
			fmt.Printf("     Input Schema Type: %s\n", tool.InputSchema.Type)
			if len(tool.InputSchema.Properties) > 0 {
				fmt.Printf("     Properties: %+v\n", tool.InputSchema.Properties)
			}
		}
	}
}

func testToolCalls(ctx context.Context, mcpClient client.MCPClient) {
	fmt.Println("\n--- Testing tool calls ---")
	
	// 测试 ping 工具
	fmt.Println("\n1. Testing ping tool:")
	pingRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "ping",
			Arguments: map[string]any{},
		},
	}
	pingResult, err := mcpClient.CallTool(ctx, pingRequest)
	if err != nil {
		log.Printf("Failed to call ping tool: %v", err)
	} else {
		fmt.Printf("✓ Ping result: %+v\n", pingResult)
		if len(pingResult.Content) > 0 {
			if textContent, ok := pingResult.Content[0].(mcp.TextContent); ok {
				fmt.Printf("  Content: %s\n", textContent.Text)
			}
		}
	}
	
	// 测试 echo 工具
	fmt.Println("\n2. Testing echo tool:")
	echoRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "echo",
			Arguments: map[string]any{
				"text": "Hello from MCP client!",
			},
		},
	}
	echoResult, err := mcpClient.CallTool(ctx, echoRequest)
	if err != nil {
		log.Printf("Failed to call echo tool: %v", err)
	} else {
		fmt.Printf("✓ Echo result: %+v\n", echoResult)
		if len(echoResult.Content) > 0 {
			if textContent, ok := echoResult.Content[0].(mcp.TextContent); ok {
				fmt.Printf("  Content: %s\n", textContent.Text)
			}
		}
	}
	
	// 测试 system_info 工具
	fmt.Println("\n3. Testing system_info tool:")
	sysInfoRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "system_info",
			Arguments: map[string]any{},
		},
	}
	sysInfoResult, err := mcpClient.CallTool(ctx, sysInfoRequest)
	if err != nil {
		log.Printf("Failed to call system_info tool: %v", err)
	} else {
		fmt.Printf("✓ System info result: %+v\n", sysInfoResult)
		if len(sysInfoResult.Content) > 0 {
			if textContent, ok := sysInfoResult.Content[0].(mcp.TextContent); ok {
				fmt.Printf("  Content: %s\n", textContent.Text)
			}
		}
	}
}