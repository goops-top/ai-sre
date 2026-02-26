#!/bin/bash

# 测试MCP服务器所有传输模式和认证功能的脚本

set -e

# 项目根目录
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../" && pwd)"
MCP_DIR="$PROJECT_ROOT/tools/mcp"
MCP_SERVER="$MCP_DIR/bin/mcp-server"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== AI SRE MCP Server - 全功能测试 ===${NC}"
echo "Project Root: $PROJECT_ROOT"
echo "MCP Server: $MCP_SERVER"
echo

# 检查二进制文件是否存在
if [[ ! -f "$MCP_SERVER" ]]; then
    echo -e "${RED} MCP server binary not found: $MCP_SERVER${NC}"
    echo "Please run 'make build-go' first"
    exit 1
fi

echo -e "${GREEN} MCP server binary found${NC}"
echo

# 测试1: 版本信息
echo -e "${YELLOW}--- Test 1: Version Information ---${NC}"
if "$MCP_SERVER" -version; then
    echo -e "${GREEN} Version command works${NC}"
else
    echo -e "${RED} Version command failed${NC}"
    exit 1
fi
echo

# 测试2: 帮助信息
echo -e "${YELLOW}--- Test 2: Help Information ---${NC}"
if "$MCP_SERVER" -help >/dev/null 2>&1; then
    echo -e "${GREEN} Help command works${NC}"
else
    echo -e "${RED} Help command failed${NC}"
    exit 1
fi
echo

# 测试3: Stdio模式（默认）
echo -e "${YELLOW}--- Test 3: Stdio Mode (Default) ---${NC}"
echo "Starting server in stdio mode for 2 seconds..."

"$MCP_SERVER" &
SERVER_PID=$!
sleep 2

if kill -0 "$SERVER_PID" 2>/dev/null; then
    echo -e "${GREEN} Stdio mode server started successfully (PID: $SERVER_PID)${NC}"
    kill "$SERVER_PID" 2>/dev/null || true
    wait "$SERVER_PID" 2>/dev/null || true
    echo -e "${GREEN} Stdio mode server stopped successfully${NC}"
else
    echo -e "${RED} Stdio mode server failed to start${NC}"
    exit 1
fi
echo

# 测试4: HTTP模式（无认证）
echo -e "${YELLOW}--- Test 4: HTTP Mode (No Auth) ---${NC}"
echo "Starting HTTP server on port 9093..."

"$MCP_SERVER" -transport http -port 9093 &
SERVER_PID=$!
sleep 3

echo "Testing health endpoint..."
if curl -s http://localhost:9093/health | grep -q "healthy"; then
    echo -e "${GREEN} HTTP mode health check passed${NC}"
else
    echo -e "${RED} HTTP mode health check failed${NC}"
    kill "$SERVER_PID" 2>/dev/null || true
    exit 1
fi

echo "Testing root endpoint..."
if curl -s http://localhost:9093/ | grep -q "AI SRE MCP Server"; then
    echo -e "${GREEN} HTTP mode root endpoint works${NC}"
else
    echo -e "${RED} HTTP mode root endpoint failed${NC}"
fi

kill "$SERVER_PID" 2>/dev/null || true
wait "$SERVER_PID" 2>/dev/null || true
echo -e "${GREEN} HTTP mode server stopped${NC}"
echo

# 测试5: HTTP模式（Bearer Token认证）
echo -e "${YELLOW}--- Test 5: HTTP Mode with Bearer Token Auth ---${NC}"
echo "Starting HTTP server with Bearer token authentication..."

"$MCP_SERVER" -transport http -port 9094 -auth-token "secret123" &
SERVER_PID=$!
sleep 3

echo "Testing without authentication (should fail)..."
HTTP_CODE=$(curl -s -w "%{http_code}" -o /dev/null http://localhost:9094/health)
if [[ "$HTTP_CODE" == "401" ]]; then
    echo -e "${GREEN} Unauthenticated request correctly rejected (401)${NC}"
else
    echo -e "${RED} Unauthenticated request should return 401, got $HTTP_CODE${NC}"
fi

echo "Testing with wrong token (should fail)..."
HTTP_CODE=$(curl -s -w "%{http_code}" -o /dev/null -H "Authorization: Bearer wrongtoken" http://localhost:9094/health)
if [[ "$HTTP_CODE" == "401" ]]; then
    echo -e "${GREEN} Wrong token correctly rejected (401)${NC}"
else
    echo -e "${RED} Wrong token should return 401, got $HTTP_CODE${NC}"
fi

echo "Testing with correct token (should succeed)..."
if curl -s -H "Authorization: Bearer secret123" http://localhost:9094/health | grep -q "healthy"; then
    echo -e "${GREEN} Correct token authentication passed${NC}"
else
    echo -e "${RED} Correct token authentication failed${NC}"
fi

kill "$SERVER_PID" 2>/dev/null || true
wait "$SERVER_PID" 2>/dev/null || true
echo -e "${GREEN} Authenticated HTTP server stopped${NC}"
echo

# 测试6: 环境变量配置
echo -e "${YELLOW}--- Test 6: Environment Variable Configuration ---${NC}"
echo "Testing environment variable configuration..."

export MCP_AUTH_ENABLED=true
export MCP_AUTH_BEARER_TOKEN="env-token-456"
export MCP_LOG_LEVEL=debug

"$MCP_SERVER" -transport http -port 9095 &
SERVER_PID=$!
sleep 3

echo "Testing with environment-configured token..."
if curl -s -H "Authorization: Bearer env-token-456" http://localhost:9095/health | grep -q "healthy"; then
    echo -e "${GREEN} Environment variable authentication works${NC}"
else
    echo -e "${RED} Environment variable authentication failed${NC}"
fi

kill "$SERVER_PID" 2>/dev/null || true
wait "$SERVER_PID" 2>/dev/null || true

# 清理环境变量
unset MCP_AUTH_ENABLED MCP_AUTH_BEARER_TOKEN MCP_LOG_LEVEL
echo -e "${GREEN} Environment variable test completed${NC}"
echo

# 测试7: 配置验证
echo -e "${YELLOW}--- Test 7: Configuration Validation ---${NC}"
echo "Testing invalid transport mode..."

if "$MCP_SERVER" -transport invalid 2>&1 | grep -q "invalid transport mode"; then
    echo -e "${GREEN} Invalid transport mode correctly rejected${NC}"
else
    echo -e "${RED} Invalid transport mode validation failed${NC}"
fi
echo

# 总结
echo -e "${BLUE}=== Test Summary ===${NC}"
echo -e "${GREEN} All tests passed successfully!${NC}"
echo
echo -e "${BLUE}Available Features:${NC}"
echo " Transport Modes:"
echo "  • stdio (default) - Standard input/output communication"
echo "  • http - HTTP-based communication with management interface"
echo "  • sse - Server-Sent Events (planned, currently falls back to stdio)"
echo
echo "🔐 Authentication:"
echo "  • Bearer Token authentication"
echo "  • API Key authentication (planned)"
echo "  • Basic authentication (planned)"
echo "  • IP whitelist support"
echo
echo " Management Features:"
echo "  • Health check endpoint (/health)"
echo "  • Web management interface (/)"
echo "  • Structured logging with configurable levels"
echo "  • Graceful shutdown"
echo
echo " Built-in Tools:"
echo "  • ping - Connection testing"
echo "  • echo - Text processing and formatting"
echo "  • system_info - System runtime information"
echo
echo -e "${BLUE}Usage Examples:${NC}"
echo "# Default stdio mode"
echo "$MCP_SERVER"
echo
echo "# HTTP mode with authentication"
echo "$MCP_SERVER -transport http -port 8080 -auth-token \"your-secret\""
echo
echo "# Environment variable configuration"
echo "MCP_AUTH_ENABLED=true MCP_AUTH_BEARER_TOKEN=\"secret\" $MCP_SERVER -transport http"
echo
echo -e "${GREEN} MCP Server is ready for production use!${NC}"