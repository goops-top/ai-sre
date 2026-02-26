package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	mcp "github.com/metoro-io/mcp-golang"
	"github.com/metoro-io/mcp-golang/transport/stdio"
	"github.com/sirupsen/logrus"
	"ai-sre/tools/mcp/internal/auth"
	"ai-sre/tools/mcp/internal/config"
	"ai-sre/tools/mcp/internal/tools"
	"ai-sre/tools/mcp/internal/transport"
	"ai-sre/tools/mcp/pkg/logger"
)

// MCPServer MCP服务器结构
type MCPServer struct {
	config         *config.Config
	server         *mcp.Server
	transport      *stdio.StdioServerTransport
	httpServer     *http.Server // HTTP服务器（用于HTTP和SSE模式）
	httpTransport  *transport.HTTPTransport // HTTP MCP传输层
	mcpHandler     *transport.MCPMessageHandler // MCP消息处理器
	authMiddleware *auth.AuthMiddleware
	tools          map[string]interface{}
	mutex          sync.RWMutex
}

// NewMCPServer 创建新的MCP服务器实例
func NewMCPServer(cfg *config.Config) *MCPServer {
	var stdioTransport *stdio.StdioServerTransport
	var httpServer *http.Server
	var httpTransport *transport.HTTPTransport
	var mcpHandler *transport.MCPMessageHandler
	var authMiddleware *auth.AuthMiddleware

	// 根据配置的传输模式创建相应的传输层
	switch cfg.MCP.Transport {
	case "stdio":
		stdioTransport = stdio.NewStdioServerTransport()
	case "sse", "http":
		// 创建鉴权中间件
		if cfg.MCP.Auth.Enabled {
			authMiddleware = auth.NewAuthMiddleware(&cfg.MCP.Auth)
		}
		
		// 创建HTTP服务器，集成MCP传输和管理端点
		mux := http.NewServeMux()
		
		// 创建MCP消息处理器
		mcpHandler = transport.NewMCPMessageHandler(nil) // 暂时传nil，稍后设置
		
		// 添加MCP协议端点（不需要认证，MCP协议自己处理）
		mux.HandleFunc("/mcp", func(w http.ResponseWriter, r *http.Request) {
			handleMCPRequest(w, r, mcpHandler)
		})
		
		// 添加通用管理端点（应用认证中间件）
		if authMiddleware != nil {
			mux.Handle("/", authMiddleware.Handler(rootHandler(cfg)))
			mux.Handle("/health", authMiddleware.Handler(generalHealthHandler(cfg)))
			mux.Handle("/status", authMiddleware.Handler(generalStatusHandler(cfg)))
		} else {
			mux.HandleFunc("/", rootHandler(cfg))
			mux.HandleFunc("/health", generalHealthHandler(cfg))
			mux.HandleFunc("/status", generalStatusHandler(cfg))
		}
		
		// 添加MCP管理端点（应用认证中间件）
		if authMiddleware != nil {
			mux.Handle("/mcp/manage", authMiddleware.Handler(mcpRootHandler(cfg)))
			mux.Handle("/mcp/manage/", authMiddleware.Handler(mcpRootHandler(cfg)))
			mux.Handle("/mcp/manage/health", authMiddleware.Handler(mcpHealthHandler(cfg)))
			mux.Handle("/mcp/manage/status", authMiddleware.Handler(mcpStatusHandler(cfg)))
			mux.Handle("/mcp/manage/info", authMiddleware.Handler(mcpInfoHandler(cfg)))
			mux.Handle("/mcp/manage/tools", authMiddleware.Handler(mcpToolsHandler(cfg)))
		} else {
			mux.HandleFunc("/mcp/manage", mcpRootHandler(cfg))
			mux.HandleFunc("/mcp/manage/", mcpRootHandler(cfg))
			mux.HandleFunc("/mcp/manage/health", mcpHealthHandler(cfg))
			mux.HandleFunc("/mcp/manage/status", mcpStatusHandler(cfg))
			mux.HandleFunc("/mcp/manage/info", mcpInfoHandler(cfg))
			mux.HandleFunc("/mcp/manage/tools", mcpToolsHandler(cfg))
		}
		
		httpServer = &http.Server{
			Addr:         cfg.GetServerAddress(),
			Handler:      mux,
			ReadTimeout:  cfg.Server.ReadTimeout,
			WriteTimeout: cfg.Server.WriteTimeout,
			IdleTimeout:  cfg.Server.IdleTimeout,
		}
		
		// 保存mcpHandler引用，稍后设置MCPServer引用
		httpTransport = &transport.HTTPTransport{}
		
		logger.WithFields(logrus.Fields{
			"transport": cfg.MCP.Transport,
			"address":   cfg.GetServerAddress(),
		}).Info("HTTP MCP transport configured")
		
	default:
		// 默认使用stdio
		logger.WithFields(logrus.Fields{
			"invalid_transport": cfg.MCP.Transport,
			"fallback_transport": "stdio",
		}).Warn("Invalid transport mode, falling back to stdio")
		cfg.MCP.Transport = "stdio"
		stdioTransport = stdio.NewStdioServerTransport()
	}

	// 创建MCP服务器
	var mcpServer *mcp.Server
	if stdioTransport != nil {
		mcpServer = mcp.NewServer(stdioTransport)
	} else {
		// HTTP模式下创建一个虚拟的stdio服务器用于工具注册
		mcpServer = mcp.NewServer(stdio.NewStdioServerTransport())
	}

	server := &MCPServer{
		config:         cfg,
		server:         mcpServer,
		transport:      stdioTransport,
		httpServer:     httpServer,
		httpTransport:  httpTransport,
		mcpHandler:     mcpHandler,
		authMiddleware: authMiddleware,
		tools:          make(map[string]interface{}),
	}

	// 如果有mcpHandler，设置MCPServer引用
	if mcpHandler != nil {
		mcpHandler.SetMCPServer(server)
		// 设置工具注册表（所有工具统一通过全局注册表调用）
		mcpHandler.SetToolRegistry(tools.GetGlobalRegistry())
	}

	return server
}

// RegisterTool 注册工具到MCP服务器
func (s *MCPServer) RegisterTool(name, description string, handler interface{}) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 使用mcp-golang库的RegisterTool方法
	if err := s.server.RegisterTool(name, description, handler); err != nil {
		logger.WithFields(logrus.Fields{
			"tool_name": name,
			"error":     err.Error(),
		}).Error("Failed to register tool")
		return fmt.Errorf("failed to register tool %s: %w", name, err)
	}

	// 记录到本地工具映射
	s.tools[name] = handler

	logger.WithFields(logrus.Fields{
		"tool_name":        name,
		"tool_description": description,
	}).Info("Tool registered successfully")

	return nil
}

// GetRegisteredTools 获取已注册的工具列表
func (s *MCPServer) GetRegisteredTools() []string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	tools := make([]string, 0, len(s.tools))
	for name := range s.tools {
		tools = append(tools, name)
	}
	return tools
}

// GetToolCount 获取工具数量
func (s *MCPServer) GetToolCount() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return len(s.tools)
}

// Start 启动MCP服务器
func (s *MCPServer) Start(ctx context.Context) error {
	logger.WithFields(logrus.Fields{
		"server_name":      s.config.MCP.Name,
		"server_version":   s.config.MCP.Version,
		"protocol_version": s.config.MCP.ProtocolVersion,
		"transport_mode":   s.config.MCP.Transport,
		"tool_count":       len(s.tools),
		"auth_enabled":     s.config.MCP.Auth.Enabled,
	}).Info("Starting MCP server")

	// 创建一个带取消的上下文
	serverCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// 启动错误通道
	errChan := make(chan error, 3)

	// 根据传输模式启动相应的服务
	switch s.config.MCP.Transport {
	case "stdio":
		// 启动stdio MCP服务器
		go func() {
			logger.Info("Starting MCP protocol server (stdio)")
			if err := s.server.Serve(); err != nil {
				errChan <- fmt.Errorf("MCP server error: %w", err)
			}
		}()
		
	case "http", "sse":
		// 启动HTTP服务器（包含MCP传输和管理端点）
		if s.httpServer != nil {
			go func() {
				logger.WithFields(logrus.Fields{
					"address":      s.httpServer.Addr,
					"transport":    s.config.MCP.Transport,
					"auth_enabled": s.config.MCP.Auth.Enabled,
				}).Info("Starting HTTP server with MCP transport")
				
				if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					errChan <- fmt.Errorf("HTTP server error: %w", err)
				}
			}()
		}
		
		logger.WithFields(logrus.Fields{
			"address":   s.config.GetServerAddress(),
			"mcp_endpoint": "/mcp",
			"management_endpoints": []string{"/", "/health", "/status", "/mcp/manage"},
		}).Info("HTTP server with MCP transport started")
	}

	logger.WithFields(logrus.Fields{
		"transport": s.config.MCP.Transport,
		"mode":      "ready",
	}).Info("MCP server started and ready for connections")

	// 等待信号或错误
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-serverCtx.Done():
		logger.Info("Server context cancelled")
		return s.shutdown()
	case sig := <-sigChan:
		logger.WithFields(logrus.Fields{
			"signal": sig.String(),
		}).Info("Received shutdown signal")
		return s.shutdown()
	case err := <-errChan:
		logger.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Server error")
		return err
	}
}

// shutdown 优雅关闭服务器
func (s *MCPServer) shutdown() error {
	logger.Info("Shutting down MCP server")

	// 如果有HTTP服务器，优雅关闭
	if s.httpServer != nil {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), s.config.Server.ShutdownTimeout)
		defer cancel()

		if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
			logger.WithError(err).Error("HTTP server shutdown error")
			return err
		}
		logger.Info("HTTP server stopped")
	}

	// 关闭MCP传输
	if s.transport != nil {
		if err := s.transport.Close(); err != nil {
			logger.WithError(err).Error("MCP transport close error")
		}
	}

	logger.Info("MCP server stopped")
	return nil
}

// rootHandler 通用根路径处理器
func rootHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		
		authStatus := "disabled"
		if cfg.MCP.Auth.Enabled {
			authStatus = fmt.Sprintf("enabled (%s)", cfg.MCP.Auth.Type)
		}
		
		html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>AI SRE Server Management</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; background: #f5f5f5; }
        .container { max-width: 800px; margin: 0 auto; background: white; padding: 30px; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        .header { color: #333; margin-bottom: 20px; }
        .info { background: #f8f9fa; padding: 20px; border-radius: 5px; margin: 20px 0; }
        .status { color: #28a745; font-weight: bold; }
        .warning { color: #ffc107; font-weight: bold; }
        .endpoint { margin: 10px 0; padding: 10px; background: #e9ecef; border-radius: 3px; }
        .code { background: #eee; padding: 2px 5px; border-radius: 3px; font-family: monospace; }
        .note { background: #d1ecf1; border: 1px solid #bee5eb; color: #0c5460; padding: 15px; border-radius: 5px; margin: 20px 0; }
        .section { margin: 30px 0; padding: 20px; border: 1px solid #dee2e6; border-radius: 5px; }
    </style>
</head>
<body>
    <div class="container">
        <h1 class="header">AI SRE Server Management</h1>
        
        <div class="info">
            <p><strong>Status:</strong> <span class="status">Running</span></p>
            <p><strong>Transport Mode:</strong> %s</p>
            <p><strong>Version:</strong> %s</p>
            <p><strong>Authentication:</strong> %s</p>
        </div>

        <div class="note">
            <strong>Note:</strong> This is the general server management interface. 
            For MCP-specific tools and functionality, visit <a href="/mcp">/mcp</a>.
        </div>
        
        <div class="section">
            <h3>General Management Endpoints:</h3>
            <div class="endpoint">
                <strong>Health Check:</strong> <span class="code">GET /health</span>
                <br><small>General server health status</small>
            </div>
            <div class="endpoint">
                <strong>Status:</strong> <span class="code">GET /status</span>
                <br><small>General server status and configuration</small>
            </div>
        </div>
        
        <div class="section">
            <h3>MCP Tools and Services:</h3>
            <div class="endpoint">
                <strong>MCP Management:</strong> <a href="/mcp"><span class="code">GET /mcp</span></a>
                <br><small>Access MCP-specific tools, health checks, and capabilities</small>
            </div>
        </div>
        
        <h3>Documentation:</h3>
        <p>For more information:</p>
        <ul>
            <li><a href="/mcp">MCP Tools and Services</a></li>
            <li><a href="https://modelcontextprotocol.io" target="_blank">Model Context Protocol Documentation</a></li>
        </ul>
    </div>
</body>
</html>`, cfg.MCP.Transport, cfg.MCP.Version, authStatus)
		
		w.Write([]byte(html))
	}
}

// generalHealthHandler 通用健康检查处理器
func generalHealthHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		
		response := fmt.Sprintf(`{
		"status": "healthy",
		"timestamp": "%s",
		"service": "ai-sre-server",
		"transport": "%s",
		"note": "General server health check. For MCP-specific health, use /mcp/health"
	}`, time.Now().UTC().Format(time.RFC3339), cfg.MCP.Transport)
		
		w.Write([]byte(response))
	}
}

// generalStatusHandler 通用状态处理器
func generalStatusHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		
		response := fmt.Sprintf(`{
		"service": "ai-sre-server",
		"status": "running",
		"timestamp": "%s",
		"transport": "%s",
		"version": "%s",
		"auth": {
			"enabled": %t,
			"type": "%s"
		},
		"endpoints": {
			"root": "/",
			"health": "/health",
			"status": "/status",
			"mcp": "/mcp"
		}
	}`, time.Now().UTC().Format(time.RFC3339), cfg.MCP.Transport, cfg.MCP.Version, cfg.MCP.Auth.Enabled, cfg.MCP.Auth.Type)
		
		w.Write([]byte(response))
	}
}

// mcpHealthHandler MCP专用健康检查处理器
func mcpHealthHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		
		note := "This is a management endpoint."
		if cfg.MCP.Transport == "stdio" {
			note = "This is a management endpoint. MCP communication happens via stdio."
		} else {
			note = fmt.Sprintf("This is a management endpoint. MCP communication happens via %s.", cfg.MCP.Transport)
		}
		
		response := fmt.Sprintf(`{
		"status": "healthy",
		"timestamp": "%s",
		"service": "ai-sre-mcp-server",
		"transport": "%s",
		"note": "%s"
	}`, time.Now().UTC().Format(time.RFC3339), cfg.MCP.Transport, note)
		
		w.Write([]byte(response))
	}
}

// mcpRootHandler MCP根路径处理器
func mcpRootHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		
		authStatus := "disabled"
		if cfg.MCP.Auth.Enabled {
			authStatus = fmt.Sprintf("enabled (%s)", cfg.MCP.Auth.Type)
		}
		
		html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>AI SRE MCP Server</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; background: #f5f5f5; }
        .container { max-width: 800px; margin: 0 auto; background: white; padding: 30px; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        .header { color: #333; margin-bottom: 20px; }
        .info { background: #f8f9fa; padding: 20px; border-radius: 5px; margin: 20px 0; }
        .status { color: #28a745; font-weight: bold; }
        .warning { color: #ffc107; font-weight: bold; }
        .endpoint { margin: 10px 0; padding: 10px; background: #e9ecef; border-radius: 3px; }
        .code { background: #eee; padding: 2px 5px; border-radius: 3px; font-family: monospace; }
        .note { background: #d1ecf1; border: 1px solid #bee5eb; color: #0c5460; padding: 15px; border-radius: 5px; margin: 20px 0; }
    </style>
</head>
<body>
    <div class="container">
        <h1 class="header">AI SRE MCP Server</h1>
        
        <div class="info">
            <p><strong>Status:</strong> <span class="status">Running</span></p>
            <p><strong>Transport Mode:</strong> %s</p>
            <p><strong>Protocol:</strong> Model Context Protocol (MCP)</p>
            <p><strong>Version:</strong> %s</p>
            <p><strong>Authentication:</strong> %s</p>
        </div>

        <div class="note">
            <strong>Note:</strong> This is a management interface. The actual MCP communication happens via <strong>%s</strong>. 
            This web interface is provided for monitoring and health checks only.
        </div>
        
        <h3>Available Management Endpoints:</h3>
        <div class="endpoint">
            <strong>Health Check:</strong> <span class="code">GET /mcp/health</span>
            <br><small>Returns MCP server health status in JSON format</small>
        </div>
        <div class="endpoint">
            <strong>Status:</strong> <span class="code">GET /mcp/status</span>
            <br><small>Returns detailed MCP server status and configuration</small>
        </div>
        <div class="endpoint">
            <strong>Info:</strong> <span class="code">GET /mcp/info</span>
            <br><small>Returns MCP server capabilities and documentation links</small>
        </div>
        <div class="endpoint">
            <strong>Tools:</strong> <span class="code">GET /mcp/tools</span>
            <br><small>Returns list of available MCP tools and their descriptions</small>
        </div>
        
        <h3>MCP Communication:</h3>
        <div class="info">
            <p>To communicate with this MCP server:</p>
            <ul>
                <li><strong>Transport Mode:</strong> %s</li>
                <li><strong>Protocol:</strong> JSON-RPC over %s</li>
                <li><strong>Format:</strong> Model Context Protocol (MCP) messages</li>
            </ul>
        </div>
        
        <h3>Available Tools:</h3>
        <div class="info">
            <ul>
                <li><strong>ping</strong> - Connection test tool</li>
                <li><strong>echo</strong> - Advanced text processing tool</li>
                <li><strong>system_info</strong> - System runtime information</li>
                <li><strong>describe_regions</strong> - Query Tencent Cloud product regions</li>
                <li><strong>get_region</strong> - Query specific region details</li>
                <li><strong>tencentcloud_validate</strong> - Validate Tencent Cloud API connection</li>
                <li><strong>tke_describe_clusters</strong> - Query TKE cluster list (tke/serverless/all)</li>
                <li><strong>tke_describe_cluster_extra_args</strong> - Query TKE cluster custom extra args</li>
                <li><strong>tke_get_cluster_level_price</strong> - Get TKE cluster level price</li>
                <li><strong>tke_describe_addon</strong> - Query TKE cluster installed addons</li>
                <li><strong>tke_get_app_chart_list</strong> - Get available TKE addon chart list</li>
                <li><strong>tke_describe_images</strong> - Get TKE supported OS images</li>
                <li><strong>tke_describe_versions</strong> - Get TKE supported cluster versions</li>
                <li><strong>tke_describe_log_switches</strong> - Query TKE cluster log switches</li>
                <li><strong>tke_describe_master_component</strong> - Query TKE master component status</li>
                <li><strong>tke_describe_cluster_instances</strong> - Query TKE cluster node instances</li>
                <li><strong>tke_describe_cluster_virtual_node</strong> - Query TKE cluster virtual nodes</li>
                <li><strong>cvm_describe_instances</strong> - Query CVM instance list</li>
                <li><strong>cvm_describe_instances_status</strong> - Query CVM instance status</li>
                <li><strong>clb_describe_load_balancers</strong> - Query CLB load balancer list</li>
                <li><strong>clb_describe_listeners</strong> - Query CLB listener list</li>
                <li><strong>clb_describe_targets</strong> - Query CLB backend targets</li>
                <li><strong>clb_describe_target_health</strong> - Query CLB target health status</li>
                <li><strong>cdb_describe_db_instances</strong> - Query CDB (MySQL) instance list</li>
                <li><strong>cdb_describe_db_instance_info</strong> - Query CDB instance details</li>
                <li><strong>cdb_describe_slow_logs</strong> - Query CDB slow query logs</li>
                <li><strong>cdb_describe_error_log</strong> - Query CDB error logs</li>
                <li><strong>vpc_describe_vpcs</strong> - Query VPC list</li>
                <li><strong>vpc_describe_subnets</strong> - Query subnet list</li>
                <li><strong>vpc_describe_security_groups</strong> - Query security group list</li>
                <li><strong>vpc_describe_network_interfaces</strong> - Query ENI list</li>
                <li><strong>vpc_describe_addresses</strong> - Query EIP list</li>
                <li><strong>vpc_describe_bandwidth_packages</strong> - Query bandwidth package list</li>
                <li><strong>vpc_describe_vpc_endpoint</strong> - Query VPC endpoint list</li>
                <li><strong>vpc_describe_vpc_endpoint_service</strong> - Query VPC endpoint service list</li>
                <li><strong>vpc_describe_vpc_peering_connections</strong> - Query VPC peering connection list</li>
            </ul>
        </div>
        
        <h3>Documentation:</h3>
        <p>For more information about the Model Context Protocol:</p>
        <ul>
            <li><a href="https://modelcontextprotocol.io" target="_blank">Official MCP Documentation</a></li>
            <li><a href="https://github.com/modelcontextprotocol" target="_blank">MCP GitHub Organization</a></li>
        </ul>
    </div>
</body>
</html>`, cfg.MCP.Transport, cfg.MCP.Version, authStatus, cfg.MCP.Transport, cfg.MCP.Transport, cfg.MCP.Transport)
		
		w.Write([]byte(html))
	}
}

// mcpStatusHandler MCP专用状态处理器
func mcpStatusHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		
		response := fmt.Sprintf(`{
		"service": "ai-sre-mcp-server",
		"status": "running",
		"timestamp": "%s",
		"transport": "%s",
		"version": "%s",
		"auth": {
			"enabled": %t,
			"type": "%s"
		},
		"endpoints": {
			"root": "/mcp",
			"health": "/mcp/health",
			"status": "/mcp/status",
			"info": "/mcp/info",
			"tools": "/mcp/tools"
		}
	}`, time.Now().UTC().Format(time.RFC3339), cfg.MCP.Transport, cfg.MCP.Version, cfg.MCP.Auth.Enabled, cfg.MCP.Auth.Type)
		
		w.Write([]byte(response))
	}
}

// mcpInfoHandler MCP专用信息处理器
func mcpInfoHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		
		// 这里应该从实际的MCPServer获取工具列表，但由于架构限制，暂时使用默认列表
		tools := []string{"ping", "echo", "system_info", "describe_regions", "get_region", "tencentcloud_validate", "tke_describe_clusters", "tke_describe_cluster_extra_args", "tke_get_cluster_level_price", "tke_describe_addon", "tke_get_app_chart_list", "tke_describe_images", "tke_describe_versions", "tke_describe_log_switches", "tke_describe_master_component", "tke_describe_cluster_instances", "tke_describe_cluster_virtual_node", "cvm_describe_instances", "cvm_describe_instances_status", "clb_describe_load_balancers", "clb_describe_listeners", "clb_describe_targets", "clb_describe_target_health", "cdb_describe_db_instances", "cdb_describe_db_instance_info", "cdb_describe_slow_logs", "cdb_describe_error_log", "vpc_describe_vpcs", "vpc_describe_subnets", "vpc_describe_security_groups", "vpc_describe_network_interfaces", "vpc_describe_addresses", "vpc_describe_bandwidth_packages", "vpc_describe_vpc_endpoint", "vpc_describe_vpc_endpoint_service", "vpc_describe_vpc_peering_connections"}
		toolsJSON, _ := json.Marshal(tools)
		
		response := fmt.Sprintf(`{
		"service": "ai-sre-mcp-server",
		"description": "AI SRE Model Context Protocol Server",
		"version": "%s",
		"protocol": "Model Context Protocol (MCP)",
		"transport": "%s",
		"capabilities": {
			"tools": %s,
			"resources": [],
			"prompts": []
		},
		"documentation": {
			"mcp_spec": "https://modelcontextprotocol.io",
			"github": "https://github.com/modelcontextprotocol"
		}
	}`, cfg.MCP.Version, cfg.MCP.Transport, string(toolsJSON))
		
		w.Write([]byte(response))
	}
}

// mcpToolsHandler MCP工具列表处理器
func mcpToolsHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		
		tools := []map[string]interface{}{
			{
				"name": "ping",
				"description": "简单的ping工具，用于测试MCP服务器连接和响应。返回指定的消息或默认的'pong'响应。",
				"endpoint": "/mcp/tools/ping",
			},
			{
				"name": "echo",
				"description": "高级文本处理和格式化工具，支持大小写转换、前缀后缀添加、文本重复等功能。",
				"endpoint": "/mcp/tools/echo",
			},
			{
				"name": "system_info",
				"description": "获取系统运行时信息，包括Go运行时、内存使用、环境变量、进程信息等。",
				"endpoint": "/mcp/tools/system_info",
			},
			{
				"name": "describe_regions",
				"description": "查询腾讯云产品支持的地域信息。支持多种产品(如tke、cvm、cos等)，支持 JSON 和表格两种输出格式。",
				"endpoint": "/mcp/tools/describe_regions",
			},
			{
				"name": "get_region",
				"description": "根据地域ID查询腾讯云产品特定地域的详细信息。支持多种产品(如tke、cvm、cos等)，支持 JSON 和表格两种输出格式。",
				"endpoint": "/mcp/tools/get_region",
			},
			{
				"name": "tencentcloud_validate",
				"description": "验证腾讯云 API 连接和权限配置。检查 SecretID、SecretKey 是否正确以及相关服务权限。",
				"endpoint": "/mcp/tools/tencentcloud_validate",
			},
			{
				"name": "tke_describe_clusters",
				"description": "查询指定地域的 TKE 集群列表。支持按集群类型过滤：all(全部)、tke(普通集群)、serverless(弹性集群)。默认查询全部集群。",
				"endpoint": "/mcp/tools/tke_describe_clusters",
			},
			{
				"name": "tke_describe_cluster_extra_args",
				"description": "查询指定地域下指定 TKE 集群的自定义参数(Etcd、KubeAPIServer、KubeControllerManager、KubeScheduler)。",
				"endpoint": "/mcp/tools/tke_describe_cluster_extra_args",
			},
			{
				"name": "tke_get_cluster_level_price",
				"description": "获取指定地域下指定集群等级的价格信息。集群等级可选：L20、L50、L100、L200、L500、L1000、L3000、L5000。",
				"endpoint": "/mcp/tools/tke_get_cluster_level_price",
			},
			{
				"name": "tke_describe_addon",
				"description": "查询指定地域下指定 TKE 集群已安装的 addon 列表。可选指定 addon 名称查询特定 addon。",
				"endpoint": "/mcp/tools/tke_describe_addon",
			},
			{
				"name": "tke_get_app_chart_list",
				"description": "获取指定地域可安装的 TKE addon 列表。支持按类型(kind)、架构(arch)、集群类型(cluster_type)过滤。",
				"endpoint": "/mcp/tools/tke_get_app_chart_list",
			},
			{
				"name": "tke_describe_images",
				"description": "获取指定地域支持的 TKE 节点 OS 镜像列表。",
				"endpoint": "/mcp/tools/tke_describe_images",
			},
			{
				"name": "tke_describe_versions",
				"description": "获取指定地域支持的 TKE 集群 Kubernetes 版本列表。",
				"endpoint": "/mcp/tools/tke_describe_versions",
			},
			{
				"name": "tke_describe_log_switches",
				"description": "查询指定地域下指定 TKE 集群的日志采集开关状态，包括审计日志、事件日志、普通日志和 Master 日志。",
				"endpoint": "/mcp/tools/tke_describe_log_switches",
			},
			{
				"name": "tke_describe_master_component",
				"description": "查询指定地域下指定 TKE 集群的 master 组件运行状态。支持 kube-apiserver、kube-scheduler、kube-controller-manager。",
				"endpoint": "/mcp/tools/tke_describe_master_component",
			},
			{
				"name": "tke_describe_cluster_instances",
				"description": "查询指定地域下指定 TKE 集群的节点实例列表，包含节点IP、角色、状态、封锁状态、节点池等信息。",
				"endpoint": "/mcp/tools/tke_describe_cluster_instances",
			},
			{
				"name": "tke_describe_cluster_virtual_node",
				"description": "查询指定地域下指定 TKE 集群的超级节点列表。",
				"endpoint": "/mcp/tools/tke_describe_cluster_virtual_node",
			},
			{
				"name": "cvm_describe_instances",
				"description": "查询指定地域的 CVM 实例列表。支持按实例ID、名称、可用区等过滤。",
				"endpoint": "/mcp/tools/cvm_describe_instances",
			},
			{
				"name": "cvm_describe_instances_status",
				"description": "查询指定地域的 CVM 实例状态列表。",
				"endpoint": "/mcp/tools/cvm_describe_instances_status",
			},
			{
				"name": "clb_describe_load_balancers",
				"description": "查询指定地域的 CLB 负载均衡实例列表。支持按ID、名称、类型等过滤。",
				"endpoint": "/mcp/tools/clb_describe_load_balancers",
			},
			{
				"name": "clb_describe_listeners",
				"description": "查询指定 CLB 实例的监听器列表。",
				"endpoint": "/mcp/tools/clb_describe_listeners",
			},
			{
				"name": "clb_describe_targets",
				"description": "查询指定 CLB 实例绑定的后端目标(RS)列表。",
				"endpoint": "/mcp/tools/clb_describe_targets",
			},
			{
				"name": "clb_describe_target_health",
				"description": "查询指定 CLB 实例后端目标的健康检查状态。",
				"endpoint": "/mcp/tools/clb_describe_target_health",
			},
			{
				"name": "cdb_describe_db_instances",
				"description": "查询指定地域的 CDB (MySQL) 实例列表。",
				"endpoint": "/mcp/tools/cdb_describe_db_instances",
			},
			{
				"name": "cdb_describe_db_instance_info",
				"description": "查询指定 CDB (MySQL) 实例的详细信息。",
				"endpoint": "/mcp/tools/cdb_describe_db_instance_info",
			},
			{
				"name": "cdb_describe_slow_logs",
				"description": "查询指定 CDB (MySQL) 实例的慢查询日志文件列表。",
				"endpoint": "/mcp/tools/cdb_describe_slow_logs",
			},
			{
				"name": "cdb_describe_error_log",
				"description": "查询指定 CDB (MySQL) 实例的错误日志数据。",
				"endpoint": "/mcp/tools/cdb_describe_error_log",
			},
			{
				"name": "vpc_describe_vpcs",
				"description": "查询指定地域的 VPC 列表。",
				"endpoint": "/mcp/tools/vpc_describe_vpcs",
			},
			{
				"name": "vpc_describe_subnets",
				"description": "查询指定地域的子网列表，支持按 VPC ID 过滤。",
				"endpoint": "/mcp/tools/vpc_describe_subnets",
			},
			{
				"name": "vpc_describe_security_groups",
				"description": "查询指定地域的安全组列表。",
				"endpoint": "/mcp/tools/vpc_describe_security_groups",
			},
			{
				"name": "vpc_describe_network_interfaces",
				"description": "查询指定地域的弹性网卡(ENI)列表，支持按 VPC ID 过滤。",
				"endpoint": "/mcp/tools/vpc_describe_network_interfaces",
			},
			{
				"name": "vpc_describe_addresses",
				"description": "查询指定地域的弹性公网IP(EIP)列表。",
				"endpoint": "/mcp/tools/vpc_describe_addresses",
			},
			{
				"name": "vpc_describe_bandwidth_packages",
				"description": "查询指定地域的带宽包列表。",
				"endpoint": "/mcp/tools/vpc_describe_bandwidth_packages",
			},
			{
				"name": "vpc_describe_vpc_endpoint",
				"description": "查询指定地域的终端节点列表。",
				"endpoint": "/mcp/tools/vpc_describe_vpc_endpoint",
			},
			{
				"name": "vpc_describe_vpc_endpoint_service",
				"description": "查询指定地域的终端节点服务列表。",
				"endpoint": "/mcp/tools/vpc_describe_vpc_endpoint_service",
			},
			{
				"name": "vpc_describe_vpc_peering_connections",
				"description": "查询指定地域的对等连接列表。",
				"endpoint": "/mcp/tools/vpc_describe_vpc_peering_connections",
			},
		}
		
		toolsJSON, _ := json.Marshal(tools)
		
		response := fmt.Sprintf(`{
		"service": "ai-sre-mcp-server",
		"timestamp": "%s",
		"total_tools": %d,
		"tools": %s,
		"note": "These are MCP tools available for execution via the Model Context Protocol"
	}`, time.Now().UTC().Format(time.RFC3339), len(tools), string(toolsJSON))
		
		w.Write([]byte(response))
	}
}

// handleMCPRequest 处理MCP协议请求
func handleMCPRequest(w http.ResponseWriter, r *http.Request, handler *transport.MCPMessageHandler) {
	// 验证协议版本
	protocolVersion := r.Header.Get("MCP-Protocol-Version")
	if protocolVersion == "" {
		protocolVersion = "2025-03-26" // 默认版本
	}
	
	// 验证支持的版本
	supportedVersions := []string{"2024-11-05", "2025-03-26", "2025-06-18"}
	supported := false
	for _, v := range supportedVersions {
		if v == protocolVersion {
			supported = true
			break
		}
	}
	if !supported {
		http.Error(w, "Unsupported protocol version", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodPost:
		// 读取请求体
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		if len(body) == 0 {
			http.Error(w, "Empty request body", http.StatusBadRequest)
			return
		}

		// 处理MCP消息
		response, err := handler.HandleMessage(r.Context(), body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// 检查Accept头决定响应格式
		accept := r.Header.Get("Accept")
		if strings.Contains(accept, "text/event-stream") {
			// 返回SSE流
			w.Header().Set("Content-Type", "text/event-stream")
			w.Header().Set("Cache-Control", "no-cache")
			w.Header().Set("Connection", "keep-alive")
			
			fmt.Fprintf(w, "data: %s\n\n", string(response))
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
		} else {
			// 返回JSON响应
			w.Header().Set("Content-Type", "application/json")
			w.Write(response)
		}

	case http.MethodGet:
		// 建立SSE流用于接收服务器推送的消息
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// 保持连接开放
		select {
		case <-r.Context().Done():
			return
		}

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

