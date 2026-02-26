package auth

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"ai-sre/tools/mcp/internal/config"
	"ai-sre/tools/mcp/pkg/logger"
)

// AuthMiddleware 鉴权中间件结构
type AuthMiddleware struct {
	config *config.AuthConfig
}

// NewAuthMiddleware 创建新的鉴权中间件
func NewAuthMiddleware(authConfig *config.AuthConfig) *AuthMiddleware {
	return &AuthMiddleware{
		config: authConfig,
	}
}

// Handler 鉴权中间件处理函数
func (am *AuthMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 如果未启用鉴权，直接通过
		if !am.config.Enabled {
			next.ServeHTTP(w, r)
			return
		}

		// 记录鉴权尝试
		clientIP := getClientIP(r)
		logger.WithFields(logrus.Fields{
			"client_ip": clientIP,
			"method":    r.Method,
			"path":      r.URL.Path,
			"auth_type": am.config.Type,
		}).Debug("Authentication attempt")

		// 检查IP白名单
		if len(am.config.AllowedIPs) > 0 && !am.isIPAllowed(clientIP) {
			am.logAuthFailure(clientIP, "IP not in whitelist")
			http.Error(w, "Forbidden: IP not allowed", http.StatusForbidden)
			return
		}

		// 根据鉴权类型进行验证
		var authResult bool
		var authError string

		switch am.config.Type {
		case "bearer":
			authResult, authError = am.validateBearerToken(r)
		case "api_key":
			authResult, authError = am.validateAPIKey(r)
		case "basic":
			authResult, authError = am.validateBasicAuth(r)
		default:
			authError = fmt.Sprintf("unsupported auth type: %s", am.config.Type)
		}

		if !authResult {
			am.logAuthFailure(clientIP, authError)
			w.Header().Set("WWW-Authenticate", am.getAuthChallenge())
			http.Error(w, "Unauthorized: "+authError, http.StatusUnauthorized)
			return
		}

		// 鉴权成功，记录日志并继续处理
		logger.WithFields(logrus.Fields{
			"client_ip": clientIP,
			"auth_type": am.config.Type,
		}).Info("Authentication successful")

		next.ServeHTTP(w, r)
	})
}

// validateBearerToken 验证Bearer Token
func (am *AuthMiddleware) validateBearerToken(r *http.Request) (bool, string) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return false, "missing Authorization header"
	}

	// 检查Bearer前缀
	const bearerPrefix = "Bearer "
	if !strings.HasPrefix(authHeader, bearerPrefix) {
		return false, "invalid Authorization header format, expected 'Bearer <token>'"
	}

	// 提取token
	token := strings.TrimPrefix(authHeader, bearerPrefix)
	if token == "" {
		return false, "empty bearer token"
	}

	// 验证token
	if token != am.config.BearerToken {
		return false, "invalid bearer token"
	}

	return true, ""
}

// validateAPIKey 验证API Key
func (am *AuthMiddleware) validateAPIKey(r *http.Request) (bool, string) {
	// 尝试从多个位置获取API Key
	apiKey := r.Header.Get("X-API-Key")
	if apiKey == "" {
		apiKey = r.Header.Get("X-Api-Key")
	}
	if apiKey == "" {
		apiKey = r.URL.Query().Get("api_key")
	}
	if apiKey == "" {
		apiKey = r.URL.Query().Get("apikey")
	}

	if apiKey == "" {
		return false, "missing API key (check X-API-Key header or api_key query parameter)"
	}

	// 验证API Key
	if apiKey != am.config.APIKey {
		return false, "invalid API key"
	}

	return true, ""
}

// validateBasicAuth 验证Basic认证
func (am *AuthMiddleware) validateBasicAuth(r *http.Request) (bool, string) {
	username, password, ok := r.BasicAuth()
	if !ok {
		return false, "missing or invalid Basic Auth credentials"
	}

	if username != am.config.Username || password != am.config.Password {
		return false, "invalid username or password"
	}

	return true, ""
}

// isIPAllowed 检查IP是否在允许列表中
func (am *AuthMiddleware) isIPAllowed(clientIP string) bool {
	if len(am.config.AllowedIPs) == 0 {
		return true // 如果没有配置白名单，允许所有IP
	}

	clientIPNet := net.ParseIP(clientIP)
	if clientIPNet == nil {
		return false
	}

	for _, allowedIP := range am.config.AllowedIPs {
		// 支持CIDR格式
		if strings.Contains(allowedIP, "/") {
			_, ipNet, err := net.ParseCIDR(allowedIP)
			if err == nil && ipNet.Contains(clientIPNet) {
				return true
			}
		} else {
			// 支持单个IP
			if allowedIP == clientIP {
				return true
			}
		}
	}

	return false
}

// getAuthChallenge 获取认证挑战头
func (am *AuthMiddleware) getAuthChallenge() string {
	switch am.config.Type {
	case "bearer":
		return "Bearer realm=\"MCP Server\""
	case "basic":
		return "Basic realm=\"MCP Server\""
	case "api_key":
		return "API-Key realm=\"MCP Server\""
	default:
		return "Bearer realm=\"MCP Server\""
	}
}

// logAuthFailure 记录鉴权失败日志
func (am *AuthMiddleware) logAuthFailure(clientIP, reason string) {
	logger.WithFields(logrus.Fields{
		"client_ip":   clientIP,
		"auth_type":   am.config.Type,
		"failure_reason": reason,
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
	}).Warn("Authentication failed")
}

// getClientIP 获取客户端真实IP
func getClientIP(r *http.Request) string {
	// 尝试从各种头部获取真实IP
	headers := []string{
		"X-Forwarded-For",
		"X-Real-IP",
		"X-Client-IP",
		"CF-Connecting-IP", // Cloudflare
	}

	for _, header := range headers {
		if ip := r.Header.Get(header); ip != "" {
			// X-Forwarded-For可能包含多个IP，取第一个
			if strings.Contains(ip, ",") {
				ip = strings.TrimSpace(strings.Split(ip, ",")[0])
			}
			if net.ParseIP(ip) != nil {
				return ip
			}
		}
	}

	// 如果没有找到，使用RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}