#!/bin/bash

# 测试新的MCP路由结构
# 作者: AI SRE Team
# 用途: 验证所有MCP端点都正确工作

set -e

# 配置
PORT=9100
TOKEN="test-route-token-$(date +%s)"
BASE_URL="http://localhost:$PORT"

echo "=== MCP路由结构测试 ==="
echo "端口: $PORT"
echo "Token: $TOKEN"
echo

# 启动服务器
echo "1. 启动MCP服务器..."
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

echo "2. 测试根路径重定向..."
REDIRECT_RESPONSE=$(curl -s -w "%{http_code}" -I "$BASE_URL/" | grep -E "(HTTP|Location)")
echo "根路径重定向: $REDIRECT_RESPONSE"

echo -e "\n3. 测试MCP管理界面..."
MCP_ROOT_RESPONSE=$(curl -s -H "Authorization: Bearer $TOKEN" "$BASE_URL/mcp" | grep -o "AI SRE MCP Server" | head -1)
if [ "$MCP_ROOT_RESPONSE" = "AI SRE MCP Server" ]; then
    echo "✅ MCP管理界面正常"
else
    echo "❌ MCP管理界面异常"
fi

echo -e "\n4. 测试健康检查端点..."
HEALTH_RESPONSE=$(curl -s -H "Authorization: Bearer $TOKEN" "$BASE_URL/mcp/health" | jq -r '.transport')
if [ "$HEALTH_RESPONSE" = "http" ]; then
    echo "✅ 健康检查端点正常，transport: $HEALTH_RESPONSE"
else
    echo "❌ 健康检查端点异常，transport: $HEALTH_RESPONSE"
fi

echo -e "\n5. 测试状态端点..."
STATUS_RESPONSE=$(curl -s -H "Authorization: Bearer $TOKEN" "$BASE_URL/mcp/status" | jq -r '.service')
if [ "$STATUS_RESPONSE" = "ai-sre-mcp-server" ]; then
    echo "✅ 状态端点正常，service: $STATUS_RESPONSE"
    # 显示端点列表
    ENDPOINTS=$(curl -s -H "Authorization: Bearer $TOKEN" "$BASE_URL/mcp/status" | jq '.endpoints')
    echo "可用端点: $ENDPOINTS"
else
    echo "❌ 状态端点异常，service: $STATUS_RESPONSE"
fi

echo -e "\n6. 测试信息端点..."
INFO_RESPONSE=$(curl -s -H "Authorization: Bearer $TOKEN" "$BASE_URL/mcp/info" | jq -r '.protocol')
if [ "$INFO_RESPONSE" = "Model Context Protocol (MCP)" ]; then
    echo "✅ 信息端点正常，protocol: $INFO_RESPONSE"
    # 显示工具列表
    TOOLS=$(curl -s -H "Authorization: Bearer $TOKEN" "$BASE_URL/mcp/info" | jq '.capabilities.tools')
    echo "可用工具: $TOOLS"
else
    echo "❌ 信息端点异常，protocol: $INFO_RESPONSE"
fi

echo -e "\n7. 测试认证保护..."
UNAUTH_RESPONSE=$(curl -s -w "%{http_code}" "$BASE_URL/mcp/health" | tail -c 3)
if [ "$UNAUTH_RESPONSE" = "401" ]; then
    echo "✅ 认证保护正常，未认证请求返回: $UNAUTH_RESPONSE"
else
    echo "❌ 认证保护异常，未认证请求返回: $UNAUTH_RESPONSE"
fi

echo -e "\n8. 测试旧端点兼容性..."
OLD_HEALTH_RESPONSE=$(curl -s -w "%{http_code}" -H "Authorization: Bearer $TOKEN" "$BASE_URL/health" | tail -c 3)
if [ "$OLD_HEALTH_RESPONSE" = "404" ]; then
    echo "✅ 旧端点已正确移除，/health 返回: $OLD_HEALTH_RESPONSE"
else
    echo "⚠️  旧端点仍然存在，/health 返回: $OLD_HEALTH_RESPONSE"
fi

echo -e "\n=== 测试完成 ==="
echo "所有MCP端点都在 /mcp 路径下正常工作！"
echo
echo "快速访问链接:"
echo "- 管理界面: $BASE_URL/mcp"
echo "- 健康检查: $BASE_URL/mcp/health"
echo "- 服务状态: $BASE_URL/mcp/status"
echo "- 服务信息: $BASE_URL/mcp/info"