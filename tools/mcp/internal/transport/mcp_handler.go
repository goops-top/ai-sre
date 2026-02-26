package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	mcp "github.com/metoro-io/mcp-golang"
	"github.com/sirupsen/logrus"
	"ai-sre/tools/mcp/pkg/logger"
)

// MCPMessageHandler MCP消息处理器
type MCPMessageHandler struct {
	server       *mcp.Server
	mcpServer    MCPServerInterface // 添加对MCPServer的引用
	toolRegistry ToolRegistry       // 工具注册表引用（统一处理所有工具调用）
	initialized  bool
	initMux      sync.RWMutex
	capabilities map[string]interface{}
}

// MCPServerInterface 定义MCPServer接口，避免循环依赖
type MCPServerInterface interface {
	GetRegisteredTools() []string
	GetToolCount() int
}

// ToolRegistry 工具注册表接口，避免循环依赖
type ToolRegistry interface {
	CallTool(ctx context.Context, toolName string, arguments map[string]interface{}) (string, error)
}

// NewMCPMessageHandler 创建MCP消息处理器
func NewMCPMessageHandler(server *mcp.Server) *MCPMessageHandler {
	return &MCPMessageHandler{
		server:       server,
		mcpServer:    nil, // 将在SetMCPServer中设置
		initialized:  false,
		capabilities: make(map[string]interface{}),
	}
}

// SetMCPServer 设置MCPServer引用
func (h *MCPMessageHandler) SetMCPServer(mcpServer MCPServerInterface) {
	h.mcpServer = mcpServer
}

// SetToolRegistry 设置工具注册表引用
func (h *MCPMessageHandler) SetToolRegistry(registry ToolRegistry) {
	h.toolRegistry = registry
}

// HandleMessage 处理MCP消息
func (h *MCPMessageHandler) HandleMessage(ctx context.Context, message []byte) ([]byte, error) {
	logger.WithFields(logrus.Fields{
		"message_size": len(message),
	}).Debug("Received MCP message")

	// 在 debug 级别下打印完整的请求内容
	if logger.GetLogger().GetLevel() <= logrus.DebugLevel {
		logger.WithFields(logrus.Fields{
			"full_message": string(message),
		}).Debug("Full MCP message content")
	}

	// 解析JSON-RPC消息
	var jsonRPCMsg map[string]interface{}
	if err := json.Unmarshal(message, &jsonRPCMsg); err != nil {
		logger.WithFields(logrus.Fields{
			"error":   err.Error(),
			"message": string(message),
		}).Error("Failed to parse JSON-RPC message")
		return nil, fmt.Errorf("invalid JSON-RPC message: %w", err)
	}

	logger.WithFields(logrus.Fields{
		"jsonrpc": jsonRPCMsg["jsonrpc"],
		"id":      jsonRPCMsg["id"],
		"method":  jsonRPCMsg["method"],
	}).Debug("Parsed JSON-RPC message")

	// 检查是否是初始化请求
	if method, ok := jsonRPCMsg["method"].(string); ok && method == "initialize" {
		logger.Debug("Handling initialize request")
		return h.HandleInitialize(ctx, message)
	}

	// 检查是否已初始化
	if !h.IsInitialized() {
		logger.WithFields(logrus.Fields{
			"method": jsonRPCMsg["method"],
		}).Warn("Received request before initialization")
		return h.createErrorResponse(jsonRPCMsg, -32002, "Server not initialized", nil)
	}

	// 处理其他MCP消息
	return h.handleMCPMessage(ctx, jsonRPCMsg, message)
}

// HandleInitialize 处理初始化请求
func (h *MCPMessageHandler) HandleInitialize(ctx context.Context, message []byte) ([]byte, error) {
	logger.WithFields(logrus.Fields{
		"message_size": len(message),
	}).Debug("Processing initialize request")

	var initRequest struct {
		JSONRPC string `json:"jsonrpc"`
		ID      interface{} `json:"id"`
		Method  string `json:"method"`
		Params  struct {
			ProtocolVersion string `json:"protocolVersion"`
			Capabilities    map[string]interface{} `json:"capabilities"`
			ClientInfo      struct {
				Name    string `json:"name"`
				Version string `json:"version"`
			} `json:"clientInfo"`
		} `json:"params"`
	}

	if err := json.Unmarshal(message, &initRequest); err != nil {
		logger.WithFields(logrus.Fields{
			"error":   err.Error(),
			"message": string(message),
		}).Error("Failed to unmarshal initialize request")
		return nil, fmt.Errorf("invalid initialize request: %w", err)
	}

	logger.WithFields(logrus.Fields{
		"protocol_version":    initRequest.Params.ProtocolVersion,
		"client_name":         initRequest.Params.ClientInfo.Name,
		"client_version":      initRequest.Params.ClientInfo.Version,
		"client_capabilities": initRequest.Params.Capabilities,
	}).Info("Received initialize request")

	// 标记为已初始化
	h.initMux.Lock()
	h.initialized = true
	h.capabilities = initRequest.Params.Capabilities
	h.initMux.Unlock()

	logger.Debug("Server marked as initialized")

	// 创建初始化响应
	response := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      initRequest.ID,
		"result": map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{
					"listChanged": false,
				},
				"resources": map[string]interface{}{
					"listChanged": false,
				},
				"prompts": map[string]interface{}{
					"listChanged": false,
				},
			},
			"serverInfo": map[string]interface{}{
				"name":    "ai-sre-mcp-server",
				"version": "1.0.0",
			},
		},
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Failed to marshal initialize response")
		return nil, fmt.Errorf("failed to marshal initialize response: %w", err)
	}

	logger.WithFields(logrus.Fields{
		"response_size": len(responseBytes),
	}).Info("Sent initialize response")

	// 在 debug 级别下打印完整的响应内容
	if logger.GetLogger().GetLevel() <= logrus.DebugLevel {
		logger.WithFields(logrus.Fields{
			"full_response": string(responseBytes),
		}).Debug("Full initialize response content")
	}

	return responseBytes, nil
}

// IsInitialized 检查是否已初始化
func (h *MCPMessageHandler) IsInitialized() bool {
	h.initMux.RLock()
	defer h.initMux.RUnlock()
	return h.initialized
}

// handleMCPMessage 处理具体的MCP消息
func (h *MCPMessageHandler) handleMCPMessage(ctx context.Context, jsonRPCMsg map[string]interface{}, rawMessage []byte) ([]byte, error) {
	method, ok := jsonRPCMsg["method"].(string)
	if !ok {
		logger.WithFields(logrus.Fields{
			"jsonrpc_msg": jsonRPCMsg,
		}).Error("Invalid or missing method in request")
		return h.createErrorResponse(jsonRPCMsg, -32600, "Invalid request", nil)
	}

	logger.WithFields(logrus.Fields{
		"method": method,
		"id":     jsonRPCMsg["id"],
	}).Debug("Routing MCP message")

	switch method {
	case "tools/list":
		logger.Debug("Routing to tools/list handler")
		return h.handleToolsList(jsonRPCMsg)
	case "tools/call":
		logger.Debug("Routing to tools/call handler")
		return h.handleToolsCall(ctx, jsonRPCMsg)
	case "resources/list":
		logger.Debug("Routing to resources/list handler")
		return h.handleResourcesList(jsonRPCMsg)
	case "prompts/list":
		logger.Debug("Routing to prompts/list handler")
		return h.handlePromptsList(jsonRPCMsg)
	default:
		logger.WithFields(logrus.Fields{
			"method": method,
		}).Warn("Unknown method requested")
		return h.createErrorResponse(jsonRPCMsg, -32601, "Method not found", nil)
	}
}

// handleToolsList 处理工具列表请求
func (h *MCPMessageHandler) handleToolsList(jsonRPCMsg map[string]interface{}) ([]byte, error) {
	logger.WithFields(logrus.Fields{
		"method": "tools/list",
		"id":     jsonRPCMsg["id"],
	}).Debug("Processing tools/list request")

	// 动态获取工具列表
	var tools []map[string]interface{}
	
	if h.mcpServer != nil {
		// 从MCPServer获取已注册的工具
		registeredTools := h.mcpServer.GetRegisteredTools()
		tools = make([]map[string]interface{}, 0, len(registeredTools))
		
		logger.WithFields(logrus.Fields{
			"registered_tools_count": len(registeredTools),
			"registered_tools":       registeredTools,
		}).Debug("Retrieved registered tools from MCPServer")
		
		for _, toolName := range registeredTools {
			toolInfo := h.getToolInfo(toolName)
			if toolInfo != nil {
				tools = append(tools, toolInfo)
				logger.WithFields(logrus.Fields{
					"tool_name":        toolName,
					"tool_description": toolInfo["description"],
				}).Debug("Added tool to response")
			} else {
				logger.WithFields(logrus.Fields{
					"tool_name": toolName,
				}).Warn("Failed to get tool info for registered tool")
			}
		}
	} else {
		// 如果没有MCPServer引用，使用默认的硬编码工具列表
		logger.Warn("MCPServer reference is nil, using default tools")
		tools = h.getDefaultTools()
	}

	logger.WithFields(logrus.Fields{
		"total_tools_returned": len(tools),
		"tools_summary": func() []string {
			names := make([]string, len(tools))
			for i, tool := range tools {
				names[i] = tool["name"].(string)
			}
			return names
		}(),
	}).Debug("Prepared tools list response")

	response := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      jsonRPCMsg["id"],
		"result": map[string]interface{}{
			"tools": tools,
		},
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Failed to marshal tools list response")
		return nil, fmt.Errorf("failed to marshal tools list response: %w", err)
	}

	logger.WithFields(logrus.Fields{
		"response_size": len(responseBytes),
		"tools_count":   len(tools),
	}).Debug("Successfully created tools/list response")

	// 在 debug 级别下打印完整的响应内容
	if logger.GetLogger().GetLevel() <= logrus.DebugLevel {
		logger.WithFields(logrus.Fields{
			"full_response": string(responseBytes),
		}).Debug("Full tools/list response content")
	}

	return responseBytes, nil
}

// getToolInfo 获取特定工具的信息
func (h *MCPMessageHandler) getToolInfo(toolName string) map[string]interface{} {
	switch toolName {
	case "ping":
		return map[string]interface{}{
			"name":        "ping",
			"description": "简单的ping工具，用于测试MCP服务器连接和响应。返回指定的消息或默认的'pong'响应。",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"message": map[string]interface{}{
						"type":        "string",
						"description": "要返回的消息",
						"default":     "pong",
					},
				},
			},
		}
	case "echo":
		return map[string]interface{}{
			"name":        "echo",
			"description": "高级文本处理和格式化工具，支持大小写转换、前缀后缀添加、文本重复等功能。",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"text": map[string]interface{}{
						"type":        "string",
						"description": "要处理的文本",
					},
					"uppercase": map[string]interface{}{
						"type":        "boolean",
						"description": "是否转换为大写",
						"default":     false,
					},
					"lowercase": map[string]interface{}{
						"type":        "boolean",
						"description": "是否转换为小写",
						"default":     false,
					},
					"prefix": map[string]interface{}{
						"type":        "string",
						"description": "添加前缀",
					},
					"suffix": map[string]interface{}{
						"type":        "string",
						"description": "添加后缀",
					},
					"repeat": map[string]interface{}{
						"type":        "integer",
						"description": "重复次数",
						"minimum":     1,
						"maximum":     10,
						"default":     1,
					},
				},
				"required": []string{"text"},
			},
		}
	case "system_info":
		return map[string]interface{}{
			"name":        "system_info",
			"description": "获取系统运行时信息，包括Go运行时、内存使用、环境变量、进程信息等。",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"include_env": map[string]interface{}{
						"type":        "boolean",
						"description": "是否包含环境变量",
						"default":     false,
					},
					"include_memory": map[string]interface{}{
						"type":        "boolean",
						"description": "是否包含内存信息",
						"default":     true,
					},
				},
			},
		}
	case "describe_regions":
		return map[string]interface{}{
			"name":        "describe_regions",
			"description": "查询腾讯云产品支持的地域信息。支持多种产品(如tke、cvm、cos等)，支持 JSON 和表格两种输出格式。",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"product": map[string]interface{}{
						"type":        "string",
						"description": "产品名称，如 tke、cvm、cos 等",
						"enum":        []string{"tke", "cvm", "cos", "clb", "vpc", "cdb"},
						"default":     "cvm",
					},
					"format": map[string]interface{}{
						"type":        "string",
						"description": "输出格式：json 或 table",
						"enum":        []string{"json", "table"},
						"default":     "table",
					},
				},
			},
		}
	case "get_region":
		return map[string]interface{}{
			"name":        "get_region",
			"description": "根据地域ID查询腾讯云产品特定地域的详细信息。支持多种产品(如tke、cvm、cos等)，支持 JSON 和表格两种输出格式。",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"region_id": map[string]interface{}{
						"type":        "string",
						"description": "地域ID，如 ap-beijing",
					},
					"product": map[string]interface{}{
						"type":        "string",
						"description": "产品名称，如 tke、cvm、cos 等",
						"enum":        []string{"tke", "cvm", "cos", "clb", "vpc", "cdb"},
						"default":     "cvm",
					},
					"format": map[string]interface{}{
						"type":        "string",
						"description": "输出格式：json 或 table",
						"enum":        []string{"json", "table"},
						"default":     "table",
					},
				},
				"required": []string{"region_id"},
			},
		}
	case "tencentcloud_validate":
		return map[string]interface{}{
			"name":        "tencentcloud_validate",
			"description": "验证腾讯云 API 连接和权限配置。检查 SecretID、SecretKey 是否正确以及相关服务权限。",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"service": map[string]interface{}{
						"type":        "string",
						"description": "要验证的服务名称，如 tke",
						"default":     "tke",
					},
				},
			},
		}
	case "tke_describe_clusters":
		return map[string]interface{}{
			"name":        "tke_describe_clusters",
			"description": "查询指定地域的 TKE 集群列表。支持按集群类型过滤：all(全部)、tke(普通集群)、serverless(弹性集群)。默认查询全部集群。",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"region": map[string]interface{}{
						"type":        "string",
						"description": "地域ID，如 ap-beijing、ap-shanghai 等",
					},
					"cluster_type": map[string]interface{}{
						"type":        "string",
						"description": "集群类型：all(全部集群)、tke(普通集群)、serverless(弹性集群)",
						"enum":        []string{"all", "tke", "serverless"},
						"default":     "all",
					},
					"format": map[string]interface{}{
						"type":        "string",
						"description": "输出格式：json 或 table",
						"enum":        []string{"json", "table"},
						"default":     "table",
					},
				},
				"required": []string{"region"},
			},
		}
	case "tke_describe_cluster_extra_args":
		return map[string]interface{}{
			"name":        "tke_describe_cluster_extra_args",
			"description": "查询指定地域下指定 TKE 集群的自定义参数(Etcd、KubeAPIServer、KubeControllerManager、KubeScheduler)。",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"region": map[string]interface{}{
						"type":        "string",
						"description": "地域ID，如 ap-beijing、ap-shanghai 等",
					},
					"cluster_id": map[string]interface{}{
						"type":        "string",
						"description": "集群ID",
					},
					"format": map[string]interface{}{
						"type":        "string",
						"description": "输出格式：json 或 table",
						"enum":        []string{"json", "table"},
						"default":     "table",
					},
				},
				"required": []string{"region", "cluster_id"},
			},
		}
	case "tke_get_cluster_level_price":
		return map[string]interface{}{
			"name":        "tke_get_cluster_level_price",
			"description": "获取指定地域下指定集群等级的价格信息。集群等级可选：L20、L50、L100、L200、L500、L1000、L3000、L5000。",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"region": map[string]interface{}{
						"type":        "string",
						"description": "地域ID，如 ap-beijing、ap-shanghai 等",
					},
					"cluster_level": map[string]interface{}{
						"type":        "string",
						"description": "集群等级",
						"enum":        []string{"L20", "L50", "L100", "L200", "L500", "L1000", "L3000", "L5000"},
					},
					"format": map[string]interface{}{
						"type":        "string",
						"description": "输出格式：json 或 table",
						"enum":        []string{"json", "table"},
						"default":     "table",
					},
				},
				"required": []string{"region", "cluster_level"},
			},
		}
	case "tke_describe_addon":
		return map[string]interface{}{
			"name":        "tke_describe_addon",
			"description": "查询指定地域下指定 TKE 集群已安装的 addon 列表。可选指定 addon 名称查询特定 addon。",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"region": map[string]interface{}{
						"type":        "string",
						"description": "地域ID，如 ap-beijing、ap-shanghai 等",
					},
					"cluster_id": map[string]interface{}{
						"type":        "string",
						"description": "集群ID",
					},
					"addon_name": map[string]interface{}{
						"type":        "string",
						"description": "addon 名称，不传时返回集群下全部 addon",
					},
					"format": map[string]interface{}{
						"type":        "string",
						"description": "输出格式：json 或 table",
						"enum":        []string{"json", "table"},
						"default":     "table",
					},
				},
				"required": []string{"region", "cluster_id"},
			},
		}
	case "tke_get_app_chart_list":
		return map[string]interface{}{
			"name":        "tke_get_app_chart_list",
			"description": "获取指定地域可安装的 TKE addon 列表。支持按类型(kind)、架构(arch)、集群类型(cluster_type)过滤。",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"region": map[string]interface{}{
						"type":        "string",
						"description": "地域ID，如 ap-beijing、ap-shanghai 等",
					},
					"kind": map[string]interface{}{
						"type":        "string",
						"description": "app 类型",
						"enum":        []string{"log", "scheduler", "network", "storage", "monitor", "dns", "image", "other", "invisible"},
					},
					"arch": map[string]interface{}{
						"type":        "string",
						"description": "支持的操作系统架构",
						"enum":        []string{"arm32", "arm64", "amd64"},
					},
					"cluster_type": map[string]interface{}{
						"type":        "string",
						"description": "集群类型",
						"enum":        []string{"tke", "eks"},
					},
					"format": map[string]interface{}{
						"type":        "string",
						"description": "输出格式：json 或 table",
						"enum":        []string{"json", "table"},
						"default":     "table",
					},
				},
				"required": []string{"region"},
			},
		}
	case "tke_describe_images":
		return map[string]interface{}{
			"name":        "tke_describe_images",
			"description": "获取指定地域支持的 TKE 节点 OS 镜像列表。",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"region": map[string]interface{}{
						"type":        "string",
						"description": "地域ID，如 ap-beijing、ap-shanghai 等",
					},
					"format": map[string]interface{}{
						"type":        "string",
						"description": "输出格式：json 或 table",
						"enum":        []string{"json", "table"},
						"default":     "table",
					},
				},
				"required": []string{"region"},
			},
		}
	case "tke_describe_versions":
		return map[string]interface{}{
			"name":        "tke_describe_versions",
			"description": "获取指定地域支持的 TKE 集群 Kubernetes 版本列表。",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"region": map[string]interface{}{
						"type":        "string",
						"description": "地域ID，如 ap-beijing、ap-shanghai 等",
					},
					"format": map[string]interface{}{
						"type":        "string",
						"description": "输出格式：json 或 table",
						"enum":        []string{"json", "table"},
						"default":     "table",
					},
				},
				"required": []string{"region"},
			},
		}
	case "tke_describe_log_switches":
		return map[string]interface{}{
			"name":        "tke_describe_log_switches",
			"description": "查询指定地域下指定 TKE 集群的日志采集开关状态，包括审计日志、事件日志、普通日志和 Master 日志。",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"region": map[string]interface{}{
						"type":        "string",
						"description": "地域ID，如 ap-beijing、ap-shanghai 等",
					},
					"cluster_id": map[string]interface{}{
						"type":        "string",
						"description": "集群ID",
					},
					"format": map[string]interface{}{
						"type":        "string",
						"description": "输出格式：json 或 table",
						"enum":        []string{"json", "table"},
						"default":     "table",
					},
				},
				"required": []string{"region", "cluster_id"},
			},
		}
	case "tke_describe_master_component":
		return map[string]interface{}{
			"name":        "tke_describe_master_component",
			"description": "查询指定地域下指定 TKE 集群的 master 组件运行状态。支持 kube-apiserver、kube-scheduler、kube-controller-manager，默认查询 kube-apiserver。",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"region": map[string]interface{}{
						"type":        "string",
						"description": "地域ID，如 ap-beijing、ap-shanghai 等",
					},
					"cluster_id": map[string]interface{}{
						"type":        "string",
						"description": "集群ID",
					},
					"component": map[string]interface{}{
						"type":        "string",
						"description": "master 组件名称",
						"enum":        []string{"kube-apiserver", "kube-scheduler", "kube-controller-manager"},
						"default":     "kube-apiserver",
					},
					"format": map[string]interface{}{
						"type":        "string",
						"description": "输出格式：json 或 table",
						"enum":        []string{"json", "table"},
						"default":     "table",
					},
				},
				"required": []string{"region", "cluster_id"},
			},
		}
	case "tke_describe_cluster_instances":
		return map[string]interface{}{
			"name":        "tke_describe_cluster_instances",
			"description": "查询指定地域下指定 TKE 集群的节点实例列表，包含节点IP、角色、状态、封锁状态、节点池等信息。支持按节点角色过滤。",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"region": map[string]interface{}{
						"type":        "string",
						"description": "地域ID，如 ap-beijing、ap-shanghai 等",
					},
					"cluster_id": map[string]interface{}{
						"type":        "string",
						"description": "集群ID",
					},
					"instance_role": map[string]interface{}{
						"type":        "string",
						"description": "节点角色",
						"enum":        []string{"WORKER", "MASTER", "ETCD", "MASTER_ETCD", "ALL"},
						"default":     "WORKER",
					},
					"format": map[string]interface{}{
						"type":        "string",
						"description": "输出格式：json 或 table",
						"enum":        []string{"json", "table"},
						"default":     "table",
					},
				},
				"required": []string{"region", "cluster_id"},
			},
		}
	case "tke_describe_cluster_virtual_node":
		return map[string]interface{}{
			"name":        "tke_describe_cluster_virtual_node",
			"description": "查询指定地域下指定 TKE 集群的超级节点列表。可选指定节点池ID过滤。",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"region": map[string]interface{}{
						"type":        "string",
						"description": "地域ID，如 ap-beijing、ap-shanghai 等",
					},
					"cluster_id": map[string]interface{}{
						"type":        "string",
						"description": "集群ID",
					},
					"node_pool_id": map[string]interface{}{
						"type":        "string",
						"description": "节点池ID，不传时返回集群下全部超级节点",
					},
					"format": map[string]interface{}{
						"type":        "string",
						"description": "输出格式：json 或 table",
						"enum":        []string{"json", "table"},
						"default":     "table",
					},
				},
				"required": []string{"region", "cluster_id"},
			},
		}
	case "cvm_describe_instances":
		return map[string]interface{}{
			"name":        "cvm_describe_instances",
			"description": "查询指定地域的 CVM 实例列表。支持按实例ID、实例名称、可用区、项目ID等过滤。返回实例的基本信息、网络配置、磁盘信息等。",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"region": map[string]interface{}{
						"type":        "string",
						"description": "地域ID，如 ap-beijing、ap-shanghai 等",
					},
					"instance_ids": map[string]interface{}{
						"type":        "string",
						"description": "实例ID列表，多个用逗号分隔，如 ins-xxx1,ins-xxx2",
					},
					"instance_name": map[string]interface{}{
						"type":        "string",
						"description": "实例名称，支持模糊匹配",
					},
					"zone": map[string]interface{}{
						"type":        "string",
						"description": "可用区，如 ap-beijing-1",
					},
					"project_id": map[string]interface{}{
						"type":        "string",
						"description": "项目ID",
					},
					"limit": map[string]interface{}{
						"type":        "string",
						"description": "返回数量，默认20，最大100",
					},
					"format": map[string]interface{}{
						"type":        "string",
						"description": "输出格式：json 或 table",
						"enum":        []string{"json", "table"},
						"default":     "table",
					},
				},
				"required": []string{"region"},
			},
		}
	case "cvm_describe_instances_status":
		return map[string]interface{}{
			"name":        "cvm_describe_instances_status",
			"description": "查询指定地域的 CVM 实例状态列表。返回实例ID和对应的运行状态(RUNNING/STOPPED/PENDING等)。",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"region": map[string]interface{}{
						"type":        "string",
						"description": "地域ID，如 ap-beijing、ap-shanghai 等",
					},
					"instance_ids": map[string]interface{}{
						"type":        "string",
						"description": "实例ID列表，多个用逗号分隔，如 ins-xxx1,ins-xxx2",
					},
					"limit": map[string]interface{}{
						"type":        "string",
						"description": "返回数量，默认20，最大100",
					},
					"format": map[string]interface{}{
						"type":        "string",
						"description": "输出格式：json 或 table",
						"enum":        []string{"json", "table"},
						"default":     "table",
					},
				},
				"required": []string{"region"},
			},
		}
	case "clb_describe_load_balancers":
		return map[string]interface{}{
			"name":        "clb_describe_load_balancers",
			"description": "查询指定地域的 CLB 负载均衡实例列表。支持按实例ID、名称、类型(OPEN/INTERNAL)、VIP等过滤。",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"region": map[string]interface{}{
						"type":        "string",
						"description": "地域ID，如 ap-beijing、ap-shanghai 等",
					},
					"load_balancer_ids": map[string]interface{}{
						"type":        "string",
						"description": "负载均衡实例ID列表，多个用逗号分隔",
					},
					"load_balancer_name": map[string]interface{}{
						"type":        "string",
						"description": "负载均衡实例名称",
					},
					"load_balancer_type": map[string]interface{}{
						"type":        "string",
						"description": "负载均衡类型：OPEN(公网)、INTERNAL(内网)",
						"enum":        []string{"OPEN", "INTERNAL"},
					},
					"load_balancer_vip": map[string]interface{}{
						"type":        "string",
						"description": "负载均衡实例VIP",
					},
					"limit": map[string]interface{}{
						"type":        "string",
						"description": "返回数量，默认20，最大100",
					},
					"format": map[string]interface{}{
						"type":        "string",
						"description": "输出格式：json 或 table",
						"enum":        []string{"json", "table"},
						"default":     "table",
					},
				},
				"required": []string{"region"},
			},
		}
	case "clb_describe_listeners":
		return map[string]interface{}{
			"name":        "clb_describe_listeners",
			"description": "查询指定地域下指定 CLB 实例的监听器列表。返回监听器的协议、端口、健康检查配置等信息。",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"region": map[string]interface{}{
						"type":        "string",
						"description": "地域ID，如 ap-beijing、ap-shanghai 等",
					},
					"load_balancer_id": map[string]interface{}{
						"type":        "string",
						"description": "负载均衡实例ID",
					},
					"listener_ids": map[string]interface{}{
						"type":        "string",
						"description": "监听器ID列表，多个用逗号分隔",
					},
					"protocol": map[string]interface{}{
						"type":        "string",
						"description": "监听器协议类型",
						"enum":        []string{"TCP", "UDP", "HTTP", "HTTPS", "TCP_SSL", "QUIC"},
					},
					"port": map[string]interface{}{
						"type":        "string",
						"description": "监听器端口",
					},
					"format": map[string]interface{}{
						"type":        "string",
						"description": "输出格式：json 或 table",
						"enum":        []string{"json", "table"},
						"default":     "table",
					},
				},
				"required": []string{"region", "load_balancer_id"},
			},
		}
	case "clb_describe_targets":
		return map[string]interface{}{
			"name":        "clb_describe_targets",
			"description": "查询指定地域下指定 CLB 实例绑定的后端目标(RS)列表。可选指定监听器ID过滤。",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"region": map[string]interface{}{
						"type":        "string",
						"description": "地域ID，如 ap-beijing、ap-shanghai 等",
					},
					"load_balancer_id": map[string]interface{}{
						"type":        "string",
						"description": "负载均衡实例ID",
					},
					"listener_ids": map[string]interface{}{
						"type":        "string",
						"description": "监听器ID列表，多个用逗号分隔",
					},
					"format": map[string]interface{}{
						"type":        "string",
						"description": "输出格式：json 或 table",
						"enum":        []string{"json", "table"},
						"default":     "table",
					},
				},
				"required": []string{"region", "load_balancer_id"},
			},
		}
	case "clb_describe_target_health":
		return map[string]interface{}{
			"name":        "clb_describe_target_health",
			"description": "查询指定地域下指定 CLB 实例后端目标的健康检查状态。支持查询多个 CLB 实例(逗号分隔)。",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"region": map[string]interface{}{
						"type":        "string",
						"description": "地域ID，如 ap-beijing、ap-shanghai 等",
					},
					"load_balancer_ids": map[string]interface{}{
						"type":        "string",
						"description": "负载均衡实例ID列表，多个用逗号分隔",
					},
					"format": map[string]interface{}{
						"type":        "string",
						"description": "输出格式：json 或 table",
						"enum":        []string{"json", "table"},
						"default":     "table",
					},
				},
				"required": []string{"region", "load_balancer_ids"},
			},
		}
	case "cdb_describe_db_instances":
		return map[string]interface{}{
			"name":        "cdb_describe_db_instances",
			"description": "查询指定地域的 CDB (MySQL) 实例列表。支持按实例ID、实例名称、状态等过滤。返回实例基本信息、配置、网络等。",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"region": map[string]interface{}{
						"type":        "string",
						"description": "地域ID，如 ap-beijing、ap-shanghai 等",
					},
					"instance_ids": map[string]interface{}{
						"type":        "string",
						"description": "实例ID列表，多个用逗号分隔",
					},
					"instance_name": map[string]interface{}{
						"type":        "string",
						"description": "实例名称，支持模糊匹配",
					},
					"status": map[string]interface{}{
						"type":        "string",
						"description": "实例状态：0-创建中 1-运行中 4-隔离中 5-已隔离",
					},
					"limit": map[string]interface{}{
						"type":        "string",
						"description": "返回数量，默认20，最大2000",
					},
					"format": map[string]interface{}{
						"type":        "string",
						"description": "输出格式：json 或 table",
						"enum":        []string{"json", "table"},
						"default":     "table",
					},
				},
				"required": []string{"region"},
			},
		}
	case "cdb_describe_db_instance_info":
		return map[string]interface{}{
			"name":        "cdb_describe_db_instance_info",
			"description": "查询指定地域下指定 CDB (MySQL) 实例的详细信息，包括实例配置、网络信息、参数等。",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"region": map[string]interface{}{
						"type":        "string",
						"description": "地域ID，如 ap-beijing、ap-shanghai 等",
					},
					"instance_id": map[string]interface{}{
						"type":        "string",
						"description": "CDB 实例ID",
					},
					"format": map[string]interface{}{
						"type":        "string",
						"description": "输出格式：json 或 table",
						"enum":        []string{"json", "table"},
						"default":     "table",
					},
				},
				"required": []string{"region", "instance_id"},
			},
		}
	case "cdb_describe_slow_logs":
		return map[string]interface{}{
			"name":        "cdb_describe_slow_logs",
			"description": "查询指定地域下指定 CDB (MySQL) 实例的慢查询日志文件列表。返回慢日志文件名、大小、时间等信息。",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"region": map[string]interface{}{
						"type":        "string",
						"description": "地域ID，如 ap-beijing、ap-shanghai 等",
					},
					"instance_id": map[string]interface{}{
						"type":        "string",
						"description": "CDB 实例ID",
					},
					"limit": map[string]interface{}{
						"type":        "string",
						"description": "返回数量，默认20，最大100",
					},
					"format": map[string]interface{}{
						"type":        "string",
						"description": "输出格式：json 或 table",
						"enum":        []string{"json", "table"},
						"default":     "table",
					},
				},
				"required": []string{"region", "instance_id"},
			},
		}
	case "cdb_describe_error_log":
		return map[string]interface{}{
			"name":        "cdb_describe_error_log",
			"description": "查询指定地域下指定 CDB (MySQL) 实例的错误日志数据。支持按时间范围和关键字过滤。默认查询最近1小时。",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"region": map[string]interface{}{
						"type":        "string",
						"description": "地域ID，如 ap-beijing、ap-shanghai 等",
					},
					"instance_id": map[string]interface{}{
						"type":        "string",
						"description": "CDB 实例ID",
					},
					"start_time": map[string]interface{}{
						"type":        "string",
						"description": "开始时间，格式：2006-01-02 15:04:05 或 Unix时间戳",
					},
					"end_time": map[string]interface{}{
						"type":        "string",
						"description": "结束时间，格式：2006-01-02 15:04:05 或 Unix时间戳",
					},
					"key_word": map[string]interface{}{
						"type":        "string",
						"description": "搜索关键字",
					},
					"limit": map[string]interface{}{
						"type":        "string",
						"description": "返回数量，默认20，最大400",
					},
					"format": map[string]interface{}{
						"type":        "string",
						"description": "输出格式：json 或 table",
						"enum":        []string{"json", "table"},
						"default":     "table",
					},
				},
				"required": []string{"region", "instance_id"},
			},
		}
	// ========== VPC 工具 Schema ==========
	case "vpc_describe_vpcs":
		return map[string]interface{}{
			"name":        "vpc_describe_vpcs",
			"description": "查询指定地域的 VPC 列表。返回 VPC ID、名称、CIDR、是否默认、DHCP、DNS 等信息。",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"region": map[string]interface{}{
						"type":        "string",
						"description": "地域ID，如 ap-beijing、ap-shanghai 等",
					},
					"format": map[string]interface{}{
						"type":        "string",
						"description": "输出格式：json 或 table",
						"enum":        []string{"json", "table"},
						"default":     "table",
					},
				},
				"required": []string{"region"},
			},
		}
	case "vpc_describe_subnets":
		return map[string]interface{}{
			"name":        "vpc_describe_subnets",
			"description": "查询指定地域的子网列表。支持按 VPC ID 过滤。返回子网ID、CIDR、可用区、可用IP数等信息。",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"region": map[string]interface{}{
						"type":        "string",
						"description": "地域ID，如 ap-beijing、ap-shanghai 等",
					},
					"vpc_id": map[string]interface{}{
						"type":        "string",
						"description": "VPC 实例ID，不传则查询所有子网",
					},
					"format": map[string]interface{}{
						"type":        "string",
						"description": "输出格式：json 或 table",
						"enum":        []string{"json", "table"},
						"default":     "table",
					},
				},
				"required": []string{"region"},
			},
		}
	case "vpc_describe_security_groups":
		return map[string]interface{}{
			"name":        "vpc_describe_security_groups",
			"description": "查询指定地域的安全组列表。返回安全组ID、名称、描述、是否默认等信息。",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"region": map[string]interface{}{
						"type":        "string",
						"description": "地域ID，如 ap-beijing、ap-shanghai 等",
					},
					"format": map[string]interface{}{
						"type":        "string",
						"description": "输出格式：json 或 table",
						"enum":        []string{"json", "table"},
						"default":     "table",
					},
				},
				"required": []string{"region"},
			},
		}
	case "vpc_describe_network_interfaces":
		return map[string]interface{}{
			"name":        "vpc_describe_network_interfaces",
			"description": "查询指定地域的弹性网卡(ENI)列表。支持按 VPC ID 过滤。返回网卡ID、MAC、状态、内网IP等信息。",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"region": map[string]interface{}{
						"type":        "string",
						"description": "地域ID，如 ap-beijing、ap-shanghai 等",
					},
					"vpc_id": map[string]interface{}{
						"type":        "string",
						"description": "VPC 实例ID，不传则查询所有网卡",
					},
					"format": map[string]interface{}{
						"type":        "string",
						"description": "输出格式：json 或 table",
						"enum":        []string{"json", "table"},
						"default":     "table",
					},
				},
				"required": []string{"region"},
			},
		}
	case "vpc_describe_addresses":
		return map[string]interface{}{
			"name":        "vpc_describe_addresses",
			"description": "查询指定地域的弹性公网IP(EIP)列表。返回 EIP ID、公网IP、状态、绑定实例、带宽等信息。",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"region": map[string]interface{}{
						"type":        "string",
						"description": "地域ID，如 ap-beijing、ap-shanghai 等",
					},
					"format": map[string]interface{}{
						"type":        "string",
						"description": "输出格式：json 或 table",
						"enum":        []string{"json", "table"},
						"default":     "table",
					},
				},
				"required": []string{"region"},
			},
		}
	case "vpc_describe_bandwidth_packages":
		return map[string]interface{}{
			"name":        "vpc_describe_bandwidth_packages",
			"description": "查询指定地域的带宽包列表。返回带宽包ID、名称、网络类型、计费类型、带宽、状态等信息。",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"region": map[string]interface{}{
						"type":        "string",
						"description": "地域ID，如 ap-beijing、ap-shanghai 等",
					},
					"format": map[string]interface{}{
						"type":        "string",
						"description": "输出格式：json 或 table",
						"enum":        []string{"json", "table"},
						"default":     "table",
					},
				},
				"required": []string{"region"},
			},
		}
	case "vpc_describe_vpc_endpoint":
		return map[string]interface{}{
			"name":        "vpc_describe_vpc_endpoint",
			"description": "查询指定地域的终端节点列表。返回终端节点ID、名称、VPC、VIP、服务ID、状态等信息。",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"region": map[string]interface{}{
						"type":        "string",
						"description": "地域ID，如 ap-beijing、ap-shanghai 等",
					},
					"format": map[string]interface{}{
						"type":        "string",
						"description": "输出格式：json 或 table",
						"enum":        []string{"json", "table"},
						"default":     "table",
					},
				},
				"required": []string{"region"},
			},
		}
	case "vpc_describe_vpc_endpoint_service":
		return map[string]interface{}{
			"name":        "vpc_describe_vpc_endpoint_service",
			"description": "查询指定地域的终端节点服务列表。返回服务ID、名称、VPC、VIP、服务类型、终端节点数等信息。",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"region": map[string]interface{}{
						"type":        "string",
						"description": "地域ID，如 ap-beijing、ap-shanghai 等",
					},
					"format": map[string]interface{}{
						"type":        "string",
						"description": "输出格式：json 或 table",
						"enum":        []string{"json", "table"},
						"default":     "table",
					},
				},
				"required": []string{"region"},
			},
		}
	case "vpc_describe_vpc_peering_connections":
		return map[string]interface{}{
			"name":        "vpc_describe_vpc_peering_connections",
			"description": "查询指定地域的对等连接列表。返回对等连接ID、名称、本端/对端VPC、地域、状态、带宽等信息。",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"region": map[string]interface{}{
						"type":        "string",
						"description": "地域ID，如 ap-beijing、ap-shanghai 等",
					},
					"format": map[string]interface{}{
						"type":        "string",
						"description": "输出格式：json 或 table",
						"enum":        []string{"json", "table"},
						"default":     "table",
					},
				},
				"required": []string{"region"},
			},
		}
	default:
		// 对于未知工具，返回基本信息
		return map[string]interface{}{
			"name":        toolName,
			"description": fmt.Sprintf("工具: %s", toolName),
			"inputSchema": map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		}
	}
}

// getDefaultTools 获取默认工具列表（作为后备）
func (h *MCPMessageHandler) getDefaultTools() []map[string]interface{} {
	return []map[string]interface{}{
		h.getToolInfo("ping"),
		h.getToolInfo("echo"),
		h.getToolInfo("system_info"),
	}
}

// handleToolsCall 处理工具调用请求
func (h *MCPMessageHandler) handleToolsCall(ctx context.Context, jsonRPCMsg map[string]interface{}) ([]byte, error) {
	logger.WithFields(logrus.Fields{
		"method": "tools/call",
		"id":     jsonRPCMsg["id"],
	}).Debug("Processing tools/call request")

	params, ok := jsonRPCMsg["params"].(map[string]interface{})
	if !ok {
		logger.WithFields(logrus.Fields{
			"error": "Invalid params type",
			"params": jsonRPCMsg["params"],
		}).Error("Invalid params in tools/call request")
		return h.createErrorResponse(jsonRPCMsg, -32602, "Invalid params", nil)
	}

	toolName, ok := params["name"].(string)
	if !ok {
		logger.WithFields(logrus.Fields{
			"error": "Missing or invalid tool name",
			"params": params,
		}).Error("Missing tool name in tools/call request")
		return h.createErrorResponse(jsonRPCMsg, -32602, "Missing tool name", nil)
	}

	arguments, _ := params["arguments"].(map[string]interface{})

	logger.WithFields(logrus.Fields{
		"tool_name":  toolName,
		"arguments":  arguments,
	}).Debug("Extracted tool call parameters")

	// 调用具体的工具
	result, err := h.callTool(ctx, toolName, arguments)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"tool_name": toolName,
			"arguments": arguments,
			"error":     err.Error(),
		}).Error("Tool execution failed")
		return h.createErrorResponse(jsonRPCMsg, -32603, "Tool execution failed", map[string]interface{}{
			"details": err.Error(),
		})
	}

	logger.WithFields(logrus.Fields{
		"tool_name":     toolName,
		"result_length": len(result),
		"result_preview": func() string {
			if len(result) > 200 {
				return result[:200] + "..."
			}
			return result
		}(),
	}).Debug("Tool execution completed successfully")

	response := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      jsonRPCMsg["id"],
		"result": map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": result,
				},
			},
		},
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Failed to marshal tool call response")
		return nil, fmt.Errorf("failed to marshal tool call response: %w", err)
	}

	logger.WithFields(logrus.Fields{
		"response_size": len(responseBytes),
		"tool_name":     toolName,
	}).Debug("Successfully created tools/call response")

	// 在 debug 级别下打印完整的响应内容
	if logger.GetLogger().GetLevel() <= logrus.DebugLevel {
		logger.WithFields(logrus.Fields{
			"full_response": string(responseBytes),
		}).Debug("Full tools/call response content")
	}

	return responseBytes, nil
}

// handleResourcesList 处理资源列表请求
func (h *MCPMessageHandler) handleResourcesList(jsonRPCMsg map[string]interface{}) ([]byte, error) {
	response := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      jsonRPCMsg["id"],
		"result": map[string]interface{}{
			"resources": []interface{}{},
		},
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal resources list response: %w", err)
	}

	return responseBytes, nil
}

// handlePromptsList 处理提示列表请求
func (h *MCPMessageHandler) handlePromptsList(jsonRPCMsg map[string]interface{}) ([]byte, error) {
	response := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      jsonRPCMsg["id"],
		"result": map[string]interface{}{
			"prompts": []interface{}{},
		},
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal prompts list response: %w", err)
	}

	return responseBytes, nil
}

// callTool 调用具体的工具（统一通过全局注册表调用）
func (h *MCPMessageHandler) callTool(ctx context.Context, toolName string, arguments map[string]interface{}) (string, error) {
	// 统一通过工具注册表调用所有工具
	if h.toolRegistry != nil {
		result, err := h.toolRegistry.CallTool(ctx, toolName, arguments)
		if err != nil {
			return "", err
		}
		if result != "" {
			return result, nil
		}
	}

	return "", fmt.Errorf("unknown tool: %s (工具未在全局注册表中注册)", toolName)
}

// HandleRequest 处理HTTP请求 (实现MCPHandler接口)
func (h *MCPMessageHandler) HandleRequest(request map[string]interface{}) map[string]interface{} {
	ctx := context.Background()
	
	// 将请求转换为JSON字节
	requestBytes, err := json.Marshal(request)
	if err != nil {
		return h.createErrorResponseMap(request, -32700, "Parse error", nil)
	}
	
	// 使用现有的HandleMessage方法
	responseBytes, err := h.HandleMessage(ctx, requestBytes)
	if err != nil {
		return h.createErrorResponseMap(request, -32603, "Internal error", map[string]interface{}{
			"details": err.Error(),
		})
	}
	
	// 将响应转换回map
	var response map[string]interface{}
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return h.createErrorResponseMap(request, -32603, "Response encoding error", nil)
	}
	
	return response
}

// createErrorResponseMap 创建错误响应map (用于HTTP接口)
func (h *MCPMessageHandler) createErrorResponseMap(request map[string]interface{}, code int, message string, data interface{}) map[string]interface{} {
	errorResponse := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      request["id"],
		"error": map[string]interface{}{
			"code":    code,
			"message": message,
		},
	}

	if data != nil {
		errorResponse["error"].(map[string]interface{})["data"] = data
	}

	return errorResponse
}

// createErrorResponse 创建错误响应
func (h *MCPMessageHandler) createErrorResponse(jsonRPCMsg map[string]interface{}, code int, message string, data interface{}) ([]byte, error) {
	errorResponse := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      jsonRPCMsg["id"],
		"error": map[string]interface{}{
			"code":    code,
			"message": message,
		},
	}

	if data != nil {
		errorResponse["error"].(map[string]interface{})["data"] = data
	}

	responseBytes, err := json.Marshal(errorResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal error response: %w", err)
	}

	return responseBytes, nil
}