package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	
	"github.com/sirupsen/logrus"
	"ai-sre/tools/mcp/internal/config"
	"ai-sre/tools/mcp/internal/server"
	"ai-sre/tools/mcp/internal/tools"
	"ai-sre/tools/mcp/pkg/logger"
)

var (
	// 版本信息，在构建时通过ldflags设置
	version   = "dev"
	commit    = "unknown"
	buildTime = "unknown"
)

func main() {
	// 解析命令行参数
	var (
		showVersion = flag.Bool("version", false, "显示版本信息")
		showHelp    = flag.Bool("help", false, "显示帮助信息")
		configFile  = flag.String("config", "", "配置文件路径")
		transport   = flag.String("transport", "", "传输模式 (stdio|sse|http)")
		authToken   = flag.String("auth-token", "", "Bearer认证令牌")
		enableAuth  = flag.Bool("enable-auth", false, "启用认证")
		port        = flag.Int("port", 0, "服务器端口")
		logLevel    = flag.String("log-level", "", "日志级别 (debug|info|warn|error)")
	)
	flag.Parse()
	
	// 显示版本信息
	if *showVersion {
		fmt.Printf("AI SRE MCP Server\n")
		fmt.Printf("Version: %s\n", version)
		fmt.Printf("Commit: %s\n", commit)
		fmt.Printf("Build Time: %s\n", buildTime)
		os.Exit(0)
	}
	
	// 显示帮助信息
	if *showHelp {
		showUsage()
		os.Exit(0)
	}
	
	// 加载配置
	cfg := config.LoadConfig()
	
	// 命令行参数覆盖配置
	if *transport != "" {
		cfg.MCP.Transport = *transport
	}
	if *authToken != "" {
		cfg.MCP.Auth.Enabled = true
		cfg.MCP.Auth.Type = "bearer"
		cfg.MCP.Auth.BearerToken = *authToken
	}
	if *enableAuth {
		cfg.MCP.Auth.Enabled = true
	}
	if *port > 0 {
		cfg.Server.Port = *port
	}
	if *logLevel != "" {
		cfg.Logging.Level = *logLevel
	}
	
	// TODO: 如果指定了配置文件，从文件加载配置
	if *configFile != "" {
		fmt.Printf("Warning: Config file loading not implemented yet: %s\n", *configFile)
	}
	
	// 初始化日志系统（在配置处理之后）
	if err := logger.Init(&cfg.Logging); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	
	logger.WithFields(logrus.Fields{
		"version":    version,
		"commit":     commit,
		"build_time": buildTime,
	}).Info("Starting AI SRE MCP Server")
	
	// 验证配置
	if err := cfg.Validate(); err != nil {
		logger.WithError(err).Fatal("Invalid configuration")
	}
	
	logger.WithFields(logrus.Fields{
		"server_name":             cfg.MCP.Name,
		"server_version":          cfg.MCP.Version,
		"protocol_version":        cfg.MCP.ProtocolVersion,
		"transport_mode":          cfg.MCP.Transport,
		"max_concurrent_requests": cfg.MCP.MaxConcurrentRequests,
		"tool_execution_timeout":  cfg.Tools.ExecutionTimeout,
		"auth_enabled":            cfg.MCP.Auth.Enabled,
		"auth_type":               cfg.MCP.Auth.Type,
	}).Info("Configuration loaded")
	
	// 创建MCP服务器
	mcpServer := server.NewMCPServer(cfg)
	
	// 创建工具管理器
	toolManager := tools.NewToolManager(mcpServer)
	
	// 注册默认工具
	if err := toolManager.RegisterDefaultTools(); err != nil {
		logger.WithError(err).Error("Failed to register some default tools")
		// 不退出，继续运行已成功注册的工具
	}
	
	// 显示注册的工具信息
	registeredTools := mcpServer.GetRegisteredTools()
	logger.WithFields(logrus.Fields{
		"tool_count": len(registeredTools),
		"tools":      registeredTools,
	}).Info("Tools registered")
	
	// 显示服务器信息
	if cfg.MCP.Transport != "stdio" {
		logger.WithFields(logrus.Fields{
			"address":      cfg.GetServerAddress(),
			"transport":    cfg.MCP.Transport,
			"auth_enabled": cfg.MCP.Auth.Enabled,
		}).Info("Server will listen on HTTP")
	}
	
	// 启动服务器
	logger.Info("Starting MCP server...")
	ctx := context.Background()
	if err := mcpServer.Start(ctx); err != nil {
		logger.WithError(err).Fatal("Failed to start MCP server")
	}
}

// showUsage 显示使用说明
func showUsage() {
	fmt.Printf(`AI SRE MCP Server

一个基于Go语言实现的MCP (Model Context Protocol) 服务器，为AI SRE系统提供工具调用能力。

用法:
  %s [选项]

选项:
  -version              显示版本信息并退出
  -help                 显示此帮助信息并退出
  -config <file>        指定配置文件路径 (暂未实现)
  -transport <mode>     传输模式 (stdio|sse|http, 默认: stdio)
  -auth-token <token>   Bearer认证令牌 (自动启用认证)
  -enable-auth          启用认证 (需要配置认证参数)
  -port <port>          服务器端口 (仅HTTP/SSE模式, 默认: 8080)

环境变量:
  服务器配置:
    MCP_HOST                    服务器监听地址 (默认: localhost)
    MCP_PORT                    服务器监听端口 (默认: 8080)
    MCP_READ_TIMEOUT            读取超时时间 (默认: 30s)
    MCP_WRITE_TIMEOUT           写入超时时间 (默认: 30s)
    MCP_IDLE_TIMEOUT            空闲超时时间 (默认: 60s)
    MCP_SHUTDOWN_TIMEOUT        关闭超时时间 (默认: 10s)

  日志配置:
    MCP_LOG_LEVEL               日志级别 (debug|info|warn|error, 默认: info)
    MCP_LOG_FORMAT              日志格式 (json|text, 默认: json)
    MCP_LOG_FILE                日志文件路径 (默认: 输出到stdout)
    MCP_LOG_ROTATE              是否启用日志轮转 (默认: true)
    MCP_LOG_MAX_SIZE            日志文件最大大小MB (默认: 100)
    MCP_LOG_MAX_BACKUPS         保留的日志文件数量 (默认: 3)
    MCP_LOG_MAX_AGE             日志文件保留天数 (默认: 7)

  MCP协议配置:
    MCP_SERVER_NAME             服务器名称 (默认: ai-sre-mcp-server)
    MCP_SERVER_VERSION          服务器版本 (默认: 1.0.0)
    MCP_PROTOCOL_VERSION        协议版本 (默认: 2024-11-05)
    MCP_TRANSPORT               传输模式 (stdio|sse|http, 默认: stdio)
    MCP_REQUEST_TIMEOUT         请求超时时间 (默认: 60s)
    MCP_MAX_CONCURRENT_REQUESTS 最大并发请求数 (默认: 100)

  认证配置:
    MCP_AUTH_ENABLED            是否启用认证 (默认: false)
    MCP_AUTH_TYPE               认证类型 (bearer|basic|api_key, 默认: bearer)
    MCP_AUTH_BEARER_TOKEN       Bearer令牌
    MCP_AUTH_API_KEY            API密钥
    MCP_AUTH_USERNAME           用户名 (Basic认证)
    MCP_AUTH_PASSWORD           密码 (Basic认证)
    MCP_AUTH_ALLOWED_IPS        允许的IP地址列表 (逗号分隔)

  功能特性:
    MCP_ENABLE_TOOLS            是否启用工具调用 (默认: true)
    MCP_ENABLE_RESOURCES        是否启用资源访问 (默认: false)
    MCP_ENABLE_PROMPTS          是否启用提示模板 (默认: false)
    MCP_ENABLE_LOGGING          是否启用日志记录 (默认: true)

  工具配置:
    MCP_TOOL_TIMEOUT            工具执行超时时间 (默认: 30s)
    MCP_TOOL_CACHE              是否启用工具缓存 (默认: false)
    MCP_TOOL_CACHE_EXPIRY       缓存过期时间 (默认: 5m)

内置工具:
  ping          简单的连接测试工具
  echo          高级文本处理和格式化工具
  system_info   系统运行时信息查询工具

示例:
  # 使用默认配置启动服务器 (stdio模式)
  %s

  # 启动HTTP模式服务器
  %s -transport http -port 9090

  # 启动SSE模式服务器并启用认证
  %s -transport sse -auth-token "your-secret-token"

  # 启动HTTP模式并通过环境变量配置认证
  MCP_AUTH_ENABLED=true MCP_AUTH_BEARER_TOKEN="secret123" %s -transport http

  # 启用调试日志
  MCP_LOG_LEVEL=debug %s

  # 显示版本信息
  %s -version

更多信息请参考项目文档。
`, os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0])
}