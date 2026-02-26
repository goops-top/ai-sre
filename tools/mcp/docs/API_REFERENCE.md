# AI SRE MCP Server API 参考

## 📖 概述

本文档描述了AI SRE MCP Server提供的所有API接口，包括HTTP管理接口和MCP工具接口。

##  HTTP管理接口

当服务器运行在HTTP模式时，提供以下管理接口：

### 基础信息

- **Base URL**: `http://localhost:8080` (默认)
- **认证方式**: Bearer Token (如果启用认证)
- **Content-Type**: `application/json`

### 端点列表

#### 1. 健康检查 - `/mcp/health`

**描述**: 检查服务器健康状态

**方法**: `GET`

**认证**: 如果启用认证则需要

**请求示例**:
```bash
# 无认证
curl http://localhost:8080/mcp/health

# 带认证
curl -H "Authorization: Bearer your-token" http://localhost:8080/mcp/health
```

**响应示例**:
```json
{
  "status": "healthy",
  "timestamp": "2026-02-12T07:15:40Z",
  "service": "ai-sre-mcp-server",
  "transport": "stdio",
  "note": "This is a management endpoint. MCP communication happens via stdio."
}
```

**状态码**:
- `200 OK`: 服务器健康
- `401 Unauthorized`: 认证失败
- `403 Forbidden`: IP不在白名单中
- `500 Internal Server Error`: 服务器错误

#### 2. 服务器状态 - `/mcp/status`

**描述**: 获取服务器详细状态和配置信息

**方法**: `GET`

**认证**: 如果启用认证则需要

**请求示例**:
```bash
curl -H "Authorization: Bearer your-token" http://localhost:8080/mcp/status
```

**响应示例**:
```json
{
  "service": "ai-sre-mcp-server",
  "status": "running",
  "timestamp": "2026-02-12T07:49:46Z",
  "transport": "http",
  "version": "1.0.0",
  "auth": {
    "enabled": true,
    "type": "bearer"
  },
  "endpoints": {
    "root": "/mcp",
    "health": "/mcp/health",
    "status": "/mcp/status",
    "info": "/mcp/info"
  }
}
```

#### 3. 服务器信息 - `/mcp/info`

**描述**: 获取服务器能力和文档链接

**方法**: `GET`

**认证**: 如果启用认证则需要

**请求示例**:
```bash
curl -H "Authorization: Bearer your-token" http://localhost:8080/mcp/info
```

**响应示例**:
```json
{
  "service": "ai-sre-mcp-server",
  "description": "AI SRE Model Context Protocol Server",
  "version": "1.0.0",
  "protocol": "Model Context Protocol (MCP)",
  "transport": "http",
  "capabilities": {
    "tools": ["ping", "echo", "system_info"],
    "resources": [],
    "prompts": []
  },
  "documentation": {
    "mcp_spec": "https://modelcontextprotocol.io",
    "github": "https://github.com/modelcontextprotocol"
  }
}
```

#### 4. 管理界面 - `/mcp`

**描述**: 显示服务器详细信息和Web管理界面

**方法**: `GET`

**认证**: 如果启用认证则需要

**请求示例**:
```bash
curl -H "Authorization: Bearer your-token" http://localhost:8080/mcp
```

**响应**: HTML页面，包含：
- 服务器基本信息
- 配置详情
- 已注册工具列表
- 可用管理端点
- 实时状态监控

## 认证

### Bearer Token认证

**Header格式**:
```
Authorization: Bearer <token>
```

**认证流程**:
1. 客户端在请求头中包含Bearer token
2. 服务器验证token有效性
3. 检查IP白名单（如果配置）
4. 返回相应结果

**错误响应**:
```json
{
  "error": "Unauthorized",
  "message": "missing Authorization header",
  "timestamp": "2026-02-12T07:15:40Z"
}
```

### IP白名单

支持以下格式：
- 单个IP: `192.168.1.100`
- CIDR网段: `10.0.0.0/8`
- 多个地址: `192.168.1.100,10.0.0.0/8`

##  MCP工具接口

### 工具调用方式

MCP工具通过标准的MCP协议调用，支持以下传输方式：
- **stdio**: 标准输入输出（默认）
- **HTTP**: HTTP POST请求（计划支持）
- **SSE**: Server-Sent Events（计划支持）

### 内置工具

#### 1. ping工具

**描述**: 简单的连接测试工具

**参数**:
```json
{
  "message": "string (可选)"
}
```

**示例调用**:
```json
{
  "name": "ping",
  "arguments": {
    "message": "Hello MCP Server"
  }
}
```

**响应**:
```json
{
  "content": [
    {
      "type": "text",
      "text": "Hello MCP Server"
    }
  ]
}
```

#### 2. echo工具

**描述**: 高级文本处理和格式化工具

**参数**:
```json
{
  "text": "string (必需)",
  "uppercase": "boolean (可选)",
  "lowercase": "boolean (可选)", 
  "prefix": "string (可选)",
  "suffix": "string (可选)",
  "repeat": "number (可选)"
}
```

**示例调用**:
```json
{
  "name": "echo",
  "arguments": {
    "text": "hello world",
    "uppercase": true,
    "prefix": ">>> ",
    "suffix": " <<<",
    "repeat": 2
  }
}
```

**响应**:
```json
{
  "content": [
    {
      "type": "text", 
      "text": ">>> HELLO WORLD <<<\n>>> HELLO WORLD <<<"
    }
  ]
}
```

#### 3. system_info工具

**描述**: 获取系统运行时信息

**参数**:
```json
{
  "info_type": "string (可选): runtime|memory|env|process"
}
```

**示例调用**:
```json
{
  "name": "system_info",
  "arguments": {
    "info_type": "memory"
  }
}
```

**响应**:
```json
{
  "content": [
    {
      "type": "text",
      "text": "Memory Information:\n- Allocated: 2.5 MB\n- Total Allocations: 1024\n- System Memory: 16 GB\n- GC Cycles: 5"
    }
  ]
}
```

**info_type选项**:
- `runtime`: Go运行时信息
- `memory`: 内存使用情况
- `env`: 环境变量
- `process`: 进程信息
- 不指定: 返回所有信息

##  错误处理

### HTTP错误码

| 状态码 | 描述 | 原因 |
|--------|------|------|
| 200 | OK | 请求成功 |
| 400 | Bad Request | 请求格式错误 |
| 401 | Unauthorized | 认证失败 |
| 403 | Forbidden | 权限不足或IP限制 |
| 404 | Not Found | 端点不存在 |
| 405 | Method Not Allowed | HTTP方法不支持 |
| 429 | Too Many Requests | 请求频率限制 |
| 500 | Internal Server Error | 服务器内部错误 |
| 503 | Service Unavailable | 服务不可用 |

### 错误响应格式

```json
{
  "error": "错误类型",
  "message": "详细错误信息",
  "timestamp": "2026-02-12T07:15:40Z",
  "request_id": "req-123456789"
}
```

### MCP错误

MCP工具调用错误遵循MCP协议标准：

```json
{
  "error": {
    "code": -32602,
    "message": "Invalid params",
    "data": {
      "parameter": "text",
      "reason": "required parameter missing"
    }
  }
}
```

##  性能和限制

### 请求限制

| 限制类型 | 默认值 | 环境变量 |
|----------|--------|----------|
| 最大并发请求 | 100 | `MCP_MAX_CONCURRENT_REQUESTS` |
| 请求超时 | 60s | `MCP_REQUEST_TIMEOUT` |
| 工具执行超时 | 30s | `MCP_TOOL_EXECUTION_TIMEOUT` |
| 请求体大小 | 1MB | `MCP_MAX_REQUEST_SIZE` |

### 性能指标

可通过系统信息工具获取：
```bash
curl -X POST http://localhost:8080/mcp/tools/call \
  -H "Authorization: Bearer token" \
  -d '{"name": "system_info", "arguments": {"info_type": "runtime"}}'
```

##  监控和日志

### 访问日志

HTTP请求自动记录访问日志：
```json
{
  "level": "info",
  "msg": "HTTP request",
  "method": "GET",
  "path": "/health",
  "status": 200,
  "duration": "1.234ms",
  "client_ip": "127.0.0.1",
  "user_agent": "curl/7.68.0"
}
```

### 认证日志

认证事件记录：
```json
{
  "level": "warning",
  "msg": "Authentication failed",
  "auth_type": "bearer",
  "client_ip": "127.0.0.1",
  "failure_reason": "invalid bearer token",
  "timestamp": "2026-02-12T07:15:40Z"
}
```

### 工具执行日志

工具调用记录：
```json
{
  "level": "info",
  "msg": "Tool executed successfully",
  "tool_name": "ping",
  "execution_time": "5.678ms",
  "parameters": {"message": "test"}
}
```

## 🧪 测试和调试

### 健康检查测试

```bash
#!/bin/bash
# 基本健康检查
response=$(curl -s -w "%{http_code}" http://localhost:8080/mcp/health)
if [[ "$response" == *"200" ]]; then
  echo " Health check passed"
else
  echo " Health check failed: $response"
fi
```

### 认证测试

```bash
#!/bin/bash
# 测试认证
TOKEN="your-test-token"

# 无认证（应该失败）
curl -s -w "Status: %{http_code}\n" http://localhost:8080/mcp/health

# 错误token（应该失败）
curl -s -w "Status: %{http_code}\n" \
  -H "Authorization: Bearer wrong-token" \
  http://localhost:8080/mcp/health

# 正确token（应该成功）
curl -s -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/mcp/health
```

### 工具测试

```bash
#!/bin/bash
# 测试MCP工具（需要MCP客户端）
echo '{"name": "ping", "arguments": {"message": "test"}}' | \
  ./mcp-server
```

##  开发指南

### 添加新工具

1. 在 `internal/tools/` 目录创建新工具文件
2. 实现工具处理函数
3. 在 `manager.go` 中注册工具
4. 更新文档

**示例**:
```go
// internal/tools/my_tool.go
func MyToolHandler(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
    // 工具实现
    return &mcp.CallToolResult{
        Content: []interface{}{
            map[string]interface{}{
                "type": "text",
                "text": "Tool result",
            },
        },
    }, nil
}
```

### 自定义认证

1. 实现 `AuthMiddleware` 接口
2. 在服务器配置中注册
3. 更新配置验证逻辑

### 扩展传输模式

1. 实现传输接口
2. 在服务器启动逻辑中添加支持
3. 更新配置和文档

##  相关链接

- [MCP协议规范](https://modelcontextprotocol.io/)
- [项目GitHub仓库](https://github.com/your-org/ai-sre)
- [完整用户指南](USER_GUIDE.md)
- [使用示例](../examples/usage-examples.md)

---

**版本**: v1.0.0  
**最后更新**: 2026-02-12  
**API版本**: v1