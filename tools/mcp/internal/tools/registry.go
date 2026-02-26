package tools

import (
	"context"
	"encoding/json"
	"sync"

	mcp "github.com/metoro-io/mcp-golang"
	"github.com/sirupsen/logrus"
	"ai-sre/tools/mcp/pkg/logger"
)

// ToolHandlerFunc 工具处理函数类型
type ToolHandlerFunc func(arguments interface{}) (*mcp.ToolResponse, error)

// GlobalToolRegistry 全局工具注册表
type GlobalToolRegistry struct {
	handlers map[string]ToolHandlerFunc
	mutex    sync.RWMutex
}

var (
	// 全局工具注册表实例
	globalRegistry = &GlobalToolRegistry{
		handlers: make(map[string]ToolHandlerFunc),
	}
)

// GetGlobalRegistry 获取全局工具注册表
func GetGlobalRegistry() *GlobalToolRegistry {
	return globalRegistry
}

// RegisterHandler 注册工具处理函数
func (r *GlobalToolRegistry) RegisterHandler(toolName string, handler ToolHandlerFunc) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	r.handlers[toolName] = handler
	logger.WithFields(logrus.Fields{
		"tool_name": toolName,
	}).Debug("Registered tool handler in global registry")
}

// GetHandler 获取工具处理函数
func (r *GlobalToolRegistry) GetHandler(toolName string) (ToolHandlerFunc, bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	handler, exists := r.handlers[toolName]
	return handler, exists
}

// CallTool 调用工具
func (r *GlobalToolRegistry) CallTool(ctx context.Context, toolName string, arguments map[string]interface{}) (string, error) {
	handler, exists := r.GetHandler(toolName)
	if !exists {
		return "", nil // 返回空字符串表示工具不存在，让调用者处理
	}
	
	logger.WithFields(logrus.Fields{
		"tool_name": toolName,
		"arguments": arguments,
	}).Debug("Calling tool via global registry")
	
	// 调用处理函数
	response, err := handler(arguments)
	if err != nil {
		return "", err
	}
	
	// 提取文本内容
	if response != nil && len(response.Content) > 0 {
		// 检查第一个内容项的类型
		content := response.Content[0]
		if content.TextContent != nil {
			return content.TextContent.Text, nil
		}
	}
	
	return "工具执行完成，但没有返回内容", nil
}

// ListTools 列出所有注册的工具
func (r *GlobalToolRegistry) ListTools() []string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	tools := make([]string, 0, len(r.handlers))
	for toolName := range r.handlers {
		tools = append(tools, toolName)
	}
	return tools
}

// ConvertArgumentsToStruct 将map参数转换为结构体
func ConvertArgumentsToStruct(arguments map[string]interface{}, target interface{}) error {
	// 先转换为JSON，再反序列化到目标结构体
	jsonData, err := json.Marshal(arguments)
	if err != nil {
		return err
	}
	
	return json.Unmarshal(jsonData, target)
}