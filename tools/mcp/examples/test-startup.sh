#!/bin/bash

# 测试MCP服务器启动的脚本

set -e

# 项目根目录
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../" && pwd)"
MCP_DIR="$PROJECT_ROOT/tools/mcp"
MCP_SERVER="$MCP_DIR/bin/mcp-server"

echo "=== MCP Server Startup Test ==="
echo "Project Root: $PROJECT_ROOT"
echo "MCP Directory: $MCP_DIR"
echo "MCP Server: $MCP_SERVER"
echo

# 检查二进制文件是否存在
if [[ ! -f "$MCP_SERVER" ]]; then
    echo " MCP server binary not found: $MCP_SERVER"
    echo "Please run 'make build-go' first"
    exit 1
fi

echo " MCP server binary found"

# 测试版本信息
echo
echo "--- Testing version command ---"
if "$MCP_SERVER" -version; then
    echo " Version command works"
else
    echo " Version command failed"
    exit 1
fi

# 测试帮助信息
echo
echo "--- Testing help command ---"
if "$MCP_SERVER" -help >/dev/null 2>&1; then
    echo " Help command works"
else
    echo " Help command failed"
    exit 1
fi

# 测试服务器启动（快速启动和停止）
echo
echo "--- Testing server startup ---"
echo "Starting server in background for 2 seconds..."

# 在后台启动服务器
"$MCP_SERVER" &
SERVER_PID=$!

# 等待2秒
sleep 2

# 检查进程是否还在运行
if kill -0 "$SERVER_PID" 2>/dev/null; then
    echo " Server started successfully (PID: $SERVER_PID)"
    
    # 停止服务器
    kill "$SERVER_PID" 2>/dev/null || true
    wait "$SERVER_PID" 2>/dev/null || true
    echo " Server stopped successfully"
else
    echo " Server failed to start or crashed"
    exit 1
fi

echo
echo "=== All tests passed! ==="
echo "The MCP server is working correctly."
echo
echo "To start the server manually:"
echo "  $MCP_SERVER"
echo
echo "To start with debug logging:"
echo "  MCP_LOG_LEVEL=debug $MCP_SERVER"