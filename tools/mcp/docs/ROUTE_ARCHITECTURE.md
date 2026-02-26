# AI SRE服务器路由架构

## 概述

AI SRE服务器采用分离式路由架构，将通用管理功能和MCP专用工具分别组织在不同的路径下，提供清晰的功能边界和更好的用户体验。

## 架构设计

### 双层路由结构

```
AI SRE Server
├── 通用管理层 (/)
│   ├── 服务器管理界面
│   ├── 通用健康检查
│   └── 通用状态监控
└── MCP工具层 (/mcp)
    ├── MCP管理界面
    ├── MCP健康检查
    ├── MCP状态信息
    ├── MCP能力展示
    └── MCP工具列表
```

## 路由详情

### 通用管理端点 (/)

#### 根路径 `/`
- **功能**: 通用服务器管理界面
- **服务标识**: `ai-sre-server`
- **用途**: 整体服务器状态概览，提供到MCP工具的导航
- **特点**: 不特定于任何协议，通用管理功能

#### 健康检查 `/health`
- **功能**: 通用服务器健康状态
- **响应示例**:
```json
{
  "status": "healthy",
  "timestamp": "2026-02-12T08:14:01Z",
  "service": "ai-sre-server",
  "transport": "http",
  "note": "General server health check. For MCP-specific health, use /mcp/health"
}
```

#### 状态信息 `/status`
- **功能**: 通用服务器状态和配置
- **响应示例**:
```json
{
  "service": "ai-sre-server",
  "status": "running",
  "transport": "http",
  "version": "1.0.0",
  "endpoints": {
    "root": "/",
    "health": "/health",
    "status": "/status",
    "mcp": "/mcp"
  }
}
```

### MCP专用端点 (/mcp)

#### MCP根路径 `/mcp`
- **功能**: MCP服务器管理界面
- **服务标识**: `ai-sre-mcp-server`
- **用途**: MCP协议相关的工具管理和监控
- **特点**: 专门针对Model Context Protocol的功能

#### MCP健康检查 `/mcp/health`
- **功能**: MCP服务器健康状态
- **响应示例**:
```json
{
  "status": "healthy",
  "timestamp": "2026-02-12T08:14:01Z",
  "service": "ai-sre-mcp-server",
  "transport": "http",
  "note": "This is a management endpoint. MCP communication happens via http."
}
```

#### MCP状态 `/mcp/status`
- **功能**: MCP服务器详细状态
- **响应示例**:
```json
{
  "service": "ai-sre-mcp-server",
  "status": "running",
  "transport": "http",
  "version": "1.0.0",
  "endpoints": {
    "root": "/mcp",
    "health": "/mcp/health",
    "status": "/mcp/status",
    "info": "/mcp/info",
    "tools": "/mcp/tools"
  }
}
```

#### MCP信息 `/mcp/info`
- **功能**: MCP服务器能力和文档
- **响应示例**:
```json
{
  "service": "ai-sre-mcp-server",
  "protocol": "Model Context Protocol (MCP)",
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

#### MCP工具 `/mcp/tools`
- **功能**: MCP工具列表和详细描述
- **响应示例**:
```json
{
  "service": "ai-sre-mcp-server",
  "total_tools": 3,
  "tools": [
    {
      "name": "ping",
      "description": "简单的ping工具，用于测试MCP服务器连接和响应。",
      "endpoint": "/mcp/tools/ping"
    }
  ]
}
```

## 设计原则

### 1. 关注点分离
- **通用管理**: 处理服务器级别的管理任务
- **MCP工具**: 专注于Model Context Protocol相关功能

### 2. 清晰的命名空间
- `/` 路径: 通用服务器功能
- `/mcp` 路径: MCP协议专用功能

### 3. 一致的认证
- 所有端点使用相同的认证机制
- 支持Bearer Token和其他认证方式

### 4. 独立的服务标识
- 通用端点: `ai-sre-server`
- MCP端点: `ai-sre-mcp-server`

## 使用场景

### 系统管理员
- 使用 `/health` 和 `/status` 监控整体服务器状态
- 使用 `/` 获取服务器概览和导航

### MCP开发者
- 使用 `/mcp/tools` 查看可用工具
- 使用 `/mcp/info` 了解服务器能力
- 使用 `/mcp/health` 检查MCP服务状态

### 监控系统
- 通用监控: 监控 `/health`
- MCP监控: 监控 `/mcp/health`
- 分别设置不同的告警规则

## 扩展性

### 未来扩展点
- `/api` - RESTful API端点
- `/ws` - WebSocket连接
- `/metrics` - Prometheus指标
- `/mcp/v2` - MCP协议版本管理

### 插件架构
- 每个功能模块可以注册自己的路径前缀
- 保持路由的模块化和可维护性

## 最佳实践

### 客户端集成
```bash
# 检查服务器是否运行
curl http://localhost:8080/health

# 检查MCP功能是否可用
curl http://localhost:8080/mcp/health

# 获取可用的MCP工具
curl http://localhost:8080/mcp/tools
```

### 监控配置
```yaml
# 通用监控
- name: server_health
  url: http://localhost:8080/health
  
# MCP监控  
- name: mcp_health
  url: http://localhost:8080/mcp/health
```

这种分离式架构提供了清晰的功能边界，便于维护和扩展，同时为不同类型的用户提供了专门的接口。