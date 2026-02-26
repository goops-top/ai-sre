#!/bin/bash

# 测试分离的路由结构
# 作者: AI SRE Team
# 用途: 验证通用管理端点和MCP专用端点的分离

set -e

# 配置
PORT=9400
TOKEN="test-separated-routes-$(date +%s)"
BASE_URL="http://localhost:$PORT"

echo "=== 分离路由结构测试 ==="
echo "端口: $PORT"
echo "Token: $TOKEN"
echo

# 启动服务器
echo "1. 启动服务器..."
./bin/mcp-server -transport http -port $PORT -auth-token "$TOKEN" &
SERVER_PID=$!
sleep 3

# 清理函数
cleanup() {
    echo -e "\n清理: 停止服务器..."
    kill $SERVER_PID 2>/dev/null || true
    wait $SERVER_PID 2>/dev/null || true
}
trap cleanup EXIT

echo "=== 通用管理端点测试 ==="

echo "2. 测试通用根路径 (/)..."
ROOT_TITLE=$(curl -s -H "Authorization: Bearer $TOKEN" "$BASE_URL/" | grep -o "AI SRE Server Management" | head -1)
if [ "$ROOT_TITLE" = "AI SRE Server Management" ]; then
    echo "✅ 通用根路径正常: $ROOT_TITLE"
else
    echo "❌ 通用根路径异常"
fi

echo -e "\n3. 测试通用健康检查 (/health)..."
GENERAL_HEALTH=$(curl -s -H "Authorization: Bearer $TOKEN" "$BASE_URL/health" | jq -r '.service')
if [ "$GENERAL_HEALTH" = "ai-sre-server" ]; then
    echo "✅ 通用健康检查正常: $GENERAL_HEALTH"
    # 显示完整响应
    echo "响应详情:"
    curl -s -H "Authorization: Bearer $TOKEN" "$BASE_URL/health" | jq '.'
else
    echo "❌ 通用健康检查异常: $GENERAL_HEALTH"
fi

echo -e "\n4. 测试通用状态 (/status)..."
GENERAL_STATUS=$(curl -s -H "Authorization: Bearer $TOKEN" "$BASE_URL/status" | jq -r '.service')
if [ "$GENERAL_STATUS" = "ai-sre-server" ]; then
    echo "✅ 通用状态端点正常: $GENERAL_STATUS"
    # 显示端点列表
    GENERAL_ENDPOINTS=$(curl -s -H "Authorization: Bearer $TOKEN" "$BASE_URL/status" | jq '.endpoints')
    echo "通用端点列表: $GENERAL_ENDPOINTS"
else
    echo "❌ 通用状态端点异常: $GENERAL_STATUS"
fi

echo -e "\n=== MCP专用端点测试 ==="

echo "5. 测试MCP根路径 (/mcp)..."
MCP_TITLE=$(curl -s -H "Authorization: Bearer $TOKEN" "$BASE_URL/mcp" | grep -o "AI SRE MCP Server" | head -1)
if [ "$MCP_TITLE" = "AI SRE MCP Server" ]; then
    echo "✅ MCP根路径正常: $MCP_TITLE"
else
    echo "❌ MCP根路径异常"
fi

echo -e "\n6. 测试MCP健康检查 (/mcp/health)..."
MCP_HEALTH=$(curl -s -H "Authorization: Bearer $TOKEN" "$BASE_URL/mcp/health" | jq -r '.service')
if [ "$MCP_HEALTH" = "ai-sre-mcp-server" ]; then
    echo "✅ MCP健康检查正常: $MCP_HEALTH"
    # 显示transport信息
    MCP_TRANSPORT=$(curl -s -H "Authorization: Bearer $TOKEN" "$BASE_URL/mcp/health" | jq -r '.transport')
    echo "MCP传输模式: $MCP_TRANSPORT"
else
    echo "❌ MCP健康检查异常: $MCP_HEALTH"
fi

echo -e "\n7. 测试MCP状态 (/mcp/status)..."
MCP_STATUS=$(curl -s -H "Authorization: Bearer $TOKEN" "$BASE_URL/mcp/status" | jq -r '.service')
if [ "$MCP_STATUS" = "ai-sre-mcp-server" ]; then
    echo "✅ MCP状态端点正常: $MCP_STATUS"
    # 显示MCP端点列表
    MCP_ENDPOINTS=$(curl -s -H "Authorization: Bearer $TOKEN" "$BASE_URL/mcp/status" | jq '.endpoints')
    echo "MCP端点列表: $MCP_ENDPOINTS"
else
    echo "❌ MCP状态端点异常: $MCP_STATUS"
fi

echo -e "\n8. 测试MCP信息 (/mcp/info)..."
MCP_PROTOCOL=$(curl -s -H "Authorization: Bearer $TOKEN" "$BASE_URL/mcp/info" | jq -r '.protocol')
if [ "$MCP_PROTOCOL" = "Model Context Protocol (MCP)" ]; then
    echo "✅ MCP信息端点正常: $MCP_PROTOCOL"
    # 显示工具列表
    MCP_TOOLS=$(curl -s -H "Authorization: Bearer $TOKEN" "$BASE_URL/mcp/info" | jq '.capabilities.tools')
    echo "MCP工具列表: $MCP_TOOLS"
else
    echo "❌ MCP信息端点异常: $MCP_PROTOCOL"
fi

echo -e "\n9. 测试MCP工具列表 (/mcp/tools)..."
MCP_TOOL_COUNT=$(curl -s -H "Authorization: Bearer $TOKEN" "$BASE_URL/mcp/tools" | jq -r '.total_tools')
if [ "$MCP_TOOL_COUNT" = "3" ]; then
    echo "✅ MCP工具列表正常: $MCP_TOOL_COUNT 个工具"
    # 显示工具详情
    echo "工具详情:"
    curl -s -H "Authorization: Bearer $TOKEN" "$BASE_URL/mcp/tools" | jq '.tools[] | {name: .name, description: .description}'
else
    echo "❌ MCP工具列表异常: $MCP_TOOL_COUNT"
fi

echo -e "\n=== 路由分离验证 ==="

echo "10. 验证路由分离..."
echo "通用服务标识: $(curl -s -H "Authorization: Bearer $TOKEN" "$BASE_URL/health" | jq -r '.service')"
echo "MCP服务标识: $(curl -s -H "Authorization: Bearer $TOKEN" "$BASE_URL/mcp/health" | jq -r '.service')"

if [ "$(curl -s -H "Authorization: Bearer $TOKEN" "$BASE_URL/health" | jq -r '.service')" != "$(curl -s -H "Authorization: Bearer $TOKEN" "$BASE_URL/mcp/health" | jq -r '.service')" ]; then
    echo "✅ 路由成功分离: 通用管理和MCP工具使用不同的服务标识"
else
    echo "❌ 路由分离失败"
fi

echo -e "\n=== 测试完成 ==="
echo "路由结构总结:"
echo "📋 通用管理端点:"
echo "  - 根路径: $BASE_URL/"
echo "  - 健康检查: $BASE_URL/health"
echo "  - 状态信息: $BASE_URL/status"
echo
echo "🔧 MCP专用端点:"
echo "  - MCP根路径: $BASE_URL/mcp"
echo "  - MCP健康检查: $BASE_URL/mcp/health"
echo "  - MCP状态: $BASE_URL/mcp/status"
echo "  - MCP信息: $BASE_URL/mcp/info"
echo "  - MCP工具: $BASE_URL/mcp/tools"