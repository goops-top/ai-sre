package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config 包含MCP服务器的所有配置选项
type Config struct {
	// 服务器配置
	Server ServerConfig `yaml:"server"`
	
	// 日志配置
	Logging LoggingConfig `yaml:"logging"`
	
	// MCP协议配置
	MCP MCPConfig `yaml:"mcp"`
	
	// 工具配置
	Tools ToolsConfig `yaml:"tools"`
}

// ServerConfig 服务器相关配置
type ServerConfig struct {
	// 服务器监听地址
	Host string `yaml:"host"`
	
	// 服务器监听端口
	Port int `yaml:"port"`
	
	// 读取超时时间
	ReadTimeout time.Duration `yaml:"read_timeout"`
	
	// 写入超时时间
	WriteTimeout time.Duration `yaml:"write_timeout"`
	
	// 空闲超时时间
	IdleTimeout time.Duration `yaml:"idle_timeout"`
	
	// 优雅关闭超时时间
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
}

// LoggingConfig 日志相关配置
type LoggingConfig struct {
	// 日志级别 (debug, info, warn, error)
	Level string `yaml:"level"`
	
	// 日志格式 (json, text)
	Format string `yaml:"format"`
	
	// 日志输出文件路径，空则输出到stdout
	File string `yaml:"file"`
	
	// 是否启用日志轮转
	Rotate bool `yaml:"rotate"`
	
	// 日志文件最大大小 (MB)
	MaxSize int `yaml:"max_size"`
	
	// 保留的日志文件数量
	MaxBackups int `yaml:"max_backups"`
	
	// 日志文件保留天数
	MaxAge int `yaml:"max_age"`
}

// MCPConfig MCP协议相关配置
type MCPConfig struct {
	// 服务器名称
	Name string `yaml:"name"`
	
	// 服务器版本
	Version string `yaml:"version"`
	
	// 协议版本
	ProtocolVersion string `yaml:"protocol_version"`
	
	// 传输模式 (stdio, sse, http)
	Transport string `yaml:"transport"`
	
	// 支持的功能特性
	Capabilities MCPCapabilities `yaml:"capabilities"`
	
	// 请求超时时间
	RequestTimeout time.Duration `yaml:"request_timeout"`
	
	// 最大并发请求数
	MaxConcurrentRequests int `yaml:"max_concurrent_requests"`
	
	// 鉴权配置
	Auth AuthConfig `yaml:"auth"`
}

// MCPCapabilities MCP服务器支持的功能特性
type MCPCapabilities struct {
	// 是否支持工具调用
	Tools bool `yaml:"tools"`
	
	// 是否支持资源访问
	Resources bool `yaml:"resources"`
	
	// 是否支持提示模板
	Prompts bool `yaml:"prompts"`
	
	// 是否支持日志记录
	Logging bool `yaml:"logging"`
}

// AuthConfig 鉴权相关配置
type AuthConfig struct {
	// 是否启用鉴权
	Enabled bool `yaml:"enabled"`
	
	// 鉴权类型 (bearer, basic, api_key)
	Type string `yaml:"type"`
	
	// Bearer Token (用于bearer类型)
	BearerToken string `yaml:"bearer_token"`
	
	// API Key (用于api_key类型)
	APIKey string `yaml:"api_key"`
	
	// 用户名 (用于basic类型)
	Username string `yaml:"username"`
	
	// 密码 (用于basic类型)
	Password string `yaml:"password"`
	
	// Token过期时间
	TokenExpiry time.Duration `yaml:"token_expiry"`
	
	// 允许的IP地址列表
	AllowedIPs []string `yaml:"allowed_ips"`
}

// ToolsConfig 工具相关配置
type ToolsConfig struct {
	// 工具执行超时时间
	ExecutionTimeout time.Duration `yaml:"execution_timeout"`
	
	// 是否启用工具缓存
	EnableCache bool `yaml:"enable_cache"`
	
	// 缓存过期时间
	CacheExpiry time.Duration `yaml:"cache_expiry"`
	
	// 允许的工具列表，空则允许所有
	AllowedTools []string `yaml:"allowed_tools"`
	
	// 禁用的工具列表
	DisabledTools []string `yaml:"disabled_tools"`
}

// LoadConfig 从环境变量和默认值加载配置
func LoadConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host:            getEnvString("MCP_HOST", "localhost"),
			Port:            getEnvInt("MCP_PORT", 8080),
			ReadTimeout:     getEnvDuration("MCP_READ_TIMEOUT", 30*time.Second),
			WriteTimeout:    getEnvDuration("MCP_WRITE_TIMEOUT", 30*time.Second),
			IdleTimeout:     getEnvDuration("MCP_IDLE_TIMEOUT", 60*time.Second),
			ShutdownTimeout: getEnvDuration("MCP_SHUTDOWN_TIMEOUT", 10*time.Second),
		},
		Logging: LoggingConfig{
			Level:      getEnvString("MCP_LOG_LEVEL", "info"),
			Format:     getEnvString("MCP_LOG_FORMAT", "json"),
			File:       getEnvString("MCP_LOG_FILE", ""),
			Rotate:     getEnvBool("MCP_LOG_ROTATE", true),
			MaxSize:    getEnvInt("MCP_LOG_MAX_SIZE", 100),
			MaxBackups: getEnvInt("MCP_LOG_MAX_BACKUPS", 3),
			MaxAge:     getEnvInt("MCP_LOG_MAX_AGE", 7),
		},
		MCP: MCPConfig{
			Name:            getEnvString("MCP_SERVER_NAME", "ai-sre-mcp-server"),
			Version:         getEnvString("MCP_SERVER_VERSION", "1.0.0"),
			ProtocolVersion: getEnvString("MCP_PROTOCOL_VERSION", "2024-11-05"),
			Transport:       getEnvString("MCP_TRANSPORT", "stdio"),
			Capabilities: MCPCapabilities{
				Tools:     getEnvBool("MCP_ENABLE_TOOLS", true),
				Resources: getEnvBool("MCP_ENABLE_RESOURCES", false),
				Prompts:   getEnvBool("MCP_ENABLE_PROMPTS", false),
				Logging:   getEnvBool("MCP_ENABLE_LOGGING", true),
			},
			RequestTimeout:        getEnvDuration("MCP_REQUEST_TIMEOUT", 60*time.Second),
			MaxConcurrentRequests: getEnvInt("MCP_MAX_CONCURRENT_REQUESTS", 100),
			Auth: AuthConfig{
				Enabled:     getEnvBool("MCP_AUTH_ENABLED", false),
				Type:        getEnvString("MCP_AUTH_TYPE", "bearer"),
				BearerToken: getEnvString("MCP_AUTH_BEARER_TOKEN", ""),
				APIKey:      getEnvString("MCP_AUTH_API_KEY", ""),
				Username:    getEnvString("MCP_AUTH_USERNAME", ""),
				Password:    getEnvString("MCP_AUTH_PASSWORD", ""),
				TokenExpiry: getEnvDuration("MCP_AUTH_TOKEN_EXPIRY", 24*time.Hour),
				AllowedIPs:  getEnvStringSlice("MCP_AUTH_ALLOWED_IPS", []string{}),
			},
		},
		Tools: ToolsConfig{
			ExecutionTimeout: getEnvDuration("MCP_TOOL_TIMEOUT", 30*time.Second),
			EnableCache:      getEnvBool("MCP_TOOL_CACHE", false),
			CacheExpiry:      getEnvDuration("MCP_TOOL_CACHE_EXPIRY", 5*time.Minute),
			AllowedTools:     []string{}, // 默认允许所有工具
			DisabledTools:    []string{}, // 默认不禁用任何工具
		},
	}
}

// Validate 验证配置的有效性
func (c *Config) Validate() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}
	
	if c.Server.ReadTimeout <= 0 {
		return fmt.Errorf("read timeout must be positive")
	}
	
	if c.Server.WriteTimeout <= 0 {
		return fmt.Errorf("write timeout must be positive")
	}
	
	if c.MCP.RequestTimeout <= 0 {
		return fmt.Errorf("request timeout must be positive")
	}
	
	if c.MCP.MaxConcurrentRequests <= 0 {
		return fmt.Errorf("max concurrent requests must be positive")
	}
	
	// 验证传输模式
	validTransports := []string{"stdio", "sse", "http"}
	if !contains(validTransports, c.MCP.Transport) {
		return fmt.Errorf("invalid transport mode: %s, valid options: %v", c.MCP.Transport, validTransports)
	}
	
	// 验证鉴权配置
	if c.MCP.Auth.Enabled {
		validAuthTypes := []string{"bearer", "basic", "api_key"}
		if !contains(validAuthTypes, c.MCP.Auth.Type) {
			return fmt.Errorf("invalid auth type: %s, valid options: %v", c.MCP.Auth.Type, validAuthTypes)
		}
		
		// 对于HTTP传输，如果启用了鉴权，必须提供相应的凭据
		if c.MCP.Transport == "http" || c.MCP.Transport == "sse" {
			switch c.MCP.Auth.Type {
			case "bearer":
				if c.MCP.Auth.BearerToken == "" {
					return fmt.Errorf("bearer token is required when auth type is 'bearer'")
				}
			case "api_key":
				if c.MCP.Auth.APIKey == "" {
					return fmt.Errorf("api key is required when auth type is 'api_key'")
				}
			case "basic":
				if c.MCP.Auth.Username == "" || c.MCP.Auth.Password == "" {
					return fmt.Errorf("username and password are required when auth type is 'basic'")
				}
			}
		}
	}
	
	return nil
}

// GetServerAddress 获取服务器完整地址
func (c *Config) GetServerAddress() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

// 辅助函数：从环境变量获取字符串值
func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// 辅助函数：从环境变量获取整数值
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// 辅助函数：从环境变量获取布尔值
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// 辅助函数：从环境变量获取时间间隔值
func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// 辅助函数：从环境变量获取字符串切片值
func getEnvStringSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		// 简单的逗号分隔解析
		parts := make([]string, 0)
		for _, part := range splitAndTrim(value, ",") {
			if part != "" {
				parts = append(parts, part)
			}
		}
		if len(parts) > 0 {
			return parts
		}
	}
	return defaultValue
}

// 辅助函数：分割字符串并去除空白
func splitAndTrim(s, sep string) []string {
	parts := make([]string, 0)
	for _, part := range strings.Split(s, sep) {
		trimmed := strings.TrimSpace(part)
		parts = append(parts, trimmed)
	}
	return parts
}

// 辅助函数：检查切片是否包含指定元素
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}