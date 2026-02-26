# MCP路由迁移指南

## 概述

为了更好地组织API结构，所有MCP相关的管理端点已从根路径迁移到 `/mcp` 路径下。

## 路由变更

### 旧路由结构
```
GET /           # 服务器管理界面
GET /health     # 健康检查
```

### 新路由结构
```
GET /           # 自动重定向到 /mcp
GET /mcp        # 服务器管理界面
GET /mcp/       # 同 /mcp
GET /mcp/health # 健康检查
GET /mcp/status # 服务器状态和配置信息
GET /mcp/info   # 服务器能力和文档链接
```

## 端点详情

### 1. 根路径重定向 `/`
- **行为**: 自动重定向到 `/mcp`
- **状态码**: 302 Found
- **认证**: 不需要认证即可重定向

### 2. MCP管理界面 `/mcp`
- **功能**: 显示服务器详细信息和Web管理界面
- **响应**: HTML页面
- **认证**: 如果启用认证则需要

### 3. 健康检查 `/mcp/health`
- **功能**: 返回服务器健康状态
- **响应**: JSON格式
- **认证**: 如果启用认证则需要

### 4. 服务器状态 `/mcp/status`
- **功能**: 返回详细的服务器状态和配置信息
- **响应**: JSON格式，包含端点列表、认证状态等
- **认证**: 如果启用认证则需要

### 5. 服务器信息 `/mcp/info`
- **功能**: 返回服务器能力、工具列表和文档链接
- **响应**: JSON格式，包含capabilities、tools等
- **认证**: 如果启用认证则需要

## 迁移影响

### 兼容性
- **向后兼容**: 旧的 `/health` 端点已移除，返回404
- **自动重定向**: 根路径 `/` 自动重定向到 `/mcp`
- **认证保护**: 所有端点都受到相同的认证保护

### 客户端更新
需要更新客户端代码中的端点URL：

```bash
# 旧的调用方式
curl http://localhost:8080/health

# 新的调用方式
curl http://localhost:8080/mcp/health
```

## 测试验证

使用提供的测试脚本验证新路由：

```bash
# 运行路由测试
./examples/test-new-routes.sh

# 运行完整测试
./examples/test-all-modes.sh
```

## 文档更新

以下文档已更新以反映新的路由结构：
- README.md
- docs/USER_GUIDE.md
- docs/API_REFERENCE.md
- docs/QUICK_REFERENCE.md
- examples/usage-examples.md

## 优势

1. **更好的组织**: 所有MCP相关功能集中在 `/mcp` 路径下
2. **清晰的结构**: 根路径用于重定向，MCP功能有专门的命名空间
3. **扩展性**: 便于未来添加更多MCP相关端点
4. **一致性**: 所有管理端点都在同一路径前缀下

## 迁移日期

- **实施日期**: 2026-02-12
- **影响版本**: v1.0.0+
- **测试状态**: 全部通过

## 支持

如有问题，请参考：
- [用户指南](USER_GUIDE.md)
- [API参考](API_REFERENCE.md)
- [快速参考](QUICK_REFERENCE.md)