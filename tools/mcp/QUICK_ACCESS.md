# AI SRE服务器快速访问指南

## 启动服务器

```bash
# 开发模式（无认证）
./tools/mcp/bin/mcp-server -transport http -port 8080

# 生产模式（带认证）
./tools/mcp/bin/mcp-server -transport http -port 8080 -auth-token "your-secret-token"
```

## 快速访问链接

### 通用管理端点
- **管理界面**: http://localhost:8080/
- **健康检查**: http://localhost:8080/health
- **服务状态**: http://localhost:8080/status

### MCP专用端点
- **MCP管理界面**: http://localhost:8080/mcp
- **MCP健康检查**: http://localhost:8080/mcp/health
- **MCP服务状态**: http://localhost:8080/mcp/status
- **MCP服务信息**: http://localhost:8080/mcp/info
- **MCP工具列表**: http://localhost:8080/mcp/tools

### 认证模式访问
```bash
# 设置认证头
AUTH_HEADER="Authorization: Bearer your-secret-token"

# 访问通用端点
curl -H "$AUTH_HEADER" http://localhost:8080/health
curl -H "$AUTH_HEADER" http://localhost:8080/status

# 访问MCP端点
curl -H "$AUTH_HEADER" http://localhost:8080/mcp/health
curl -H "$AUTH_HEADER" http://localhost:8080/mcp/tools
```

## 一键测试

```bash
# 测试分离路由结构
./tools/mcp/examples/test-separated-routes.sh

# 完整功能测试
./tools/mcp/examples/test-all-modes.sh
```

## 端点对比

### 通用管理端点

| 端点 | 功能 | 服务标识 | 响应格式 |
|------|------|----------|----------|
| `/` | 通用管理界面 | `ai-sre-server` | HTML |
| `/health` | 通用健康检查 | `ai-sre-server` | JSON |
| `/status` | 通用服务状态 | `ai-sre-server` | JSON |

### MCP专用端点

| 端点 | 功能 | 服务标识 | 响应格式 |
|------|------|----------|----------|
| `/mcp` | MCP管理界面 | `ai-sre-mcp-server` | HTML |
| `/mcp/health` | MCP健康检查 | `ai-sre-mcp-server` | JSON |
| `/mcp/status` | MCP服务状态 | `ai-sre-mcp-server` | JSON |
| `/mcp/info` | MCP服务信息 | `ai-sre-mcp-server` | JSON |
| `/mcp/tools` | MCP工具列表 | `ai-sre-mcp-server` | JSON |

## 响应示例

### 通用健康检查
```json
{
  "status": "healthy",
  "timestamp": "2026-02-12T08:14:01Z",
  "service": "ai-sre-server",
  "transport": "http",
  "note": "General server health check. For MCP-specific health, use /mcp/health"
}
```

### MCP健康检查
```json
{
  "status": "healthy",
  "timestamp": "2026-02-12T08:14:01Z",
  "service": "ai-sre-mcp-server",
  "transport": "http",
  "note": "This is a management endpoint. MCP communication happens via http."
}
```

### MCP工具列表
```json
{
  "service": "ai-sre-mcp-server",
  "total_tools": 3,
  "tools": [
    {
      "name": "ping",
      "description": "简单的ping工具，用于测试MCP服务器连接和响应。",
      "endpoint": "/mcp/tools/ping"
    },
    {
      "name": "echo", 
      "description": "高级文本处理和格式化工具。",
      "endpoint": "/mcp/tools/echo"
    },
    {
      "name": "system_info",
      "description": "获取系统运行时信息。",
      "endpoint": "/mcp/tools/system_info"
    }
  ]
}
```

## 使用建议

### 系统管理员
- 使用通用端点监控整体服务器状态
- 定期检查 `/health` 和 `/status`

### MCP开发者
- 使用MCP端点管理和测试工具
- 查看 `/mcp/tools` 了解可用功能

### 监控系统
- 分别监控通用和MCP端点
- 设置不同的告警规则