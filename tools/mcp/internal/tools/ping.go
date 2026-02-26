package tools

import (
	"fmt"
	"time"

	mcp "github.com/metoro-io/mcp-golang"
)

// PingArguments ping工具的参数结构
type PingArguments struct {
	Message *string `json:"message" jsonschema:"description=要回显的消息内容,default=pong"`
}

// PingHandler ping工具的处理函数
func PingHandler(arguments PingArguments) (*mcp.ToolResponse, error) {
	startTime := time.Now()
	
	// 如果没有提供消息，使用默认值
	message := "pong"
	if arguments.Message != nil && *arguments.Message != "" {
		message = *arguments.Message
	}
	
	result := map[string]interface{}{
		"message":   message,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"server":    "ai-sre-mcp-server",
		"status":    "success",
		"duration_ms": time.Since(startTime).Milliseconds(),
	}
	
	// 创建响应内容
	responseText := fmt.Sprintf("Ping successful! Message: %s, Timestamp: %s, Duration: %dms", 
		result["message"], result["timestamp"], result["duration_ms"])
	
	return mcp.NewToolResponse(mcp.NewTextContent(responseText)), nil
}