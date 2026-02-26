# AI SRE MCP Server 快速参考

##  快速启动

```bash
# 构建
make build-go

# 默认启动（stdio模式）
./tools/mcp/bin/mcp-server

# HTTP模式 + 认证
./tools/mcp/bin/mcp-server -transport http -port 8080 -auth-token "secret"
```

##  命令行参数

| 参数 | 描述 | 示例 |
|------|------|------|
| `-version` | 显示版本 | `./mcp-server -version` |
| `-help` | 显示帮助 | `./mcp-server -help` |
| `-transport <mode>` | 传输模式 | `-transport http` |
| `-port <port>` | 端口号 | `-port 8080` |
| `-auth-token <token>` | 认证令牌 | `-auth-token "secret"` |
| `-enable-auth` | 启用认证 | `-enable-auth` |

## 核心环境变量

```bash
# 传输和端口
export MCP_TRANSPORT=http
export MCP_PORT=8080

# 认证
export MCP_AUTH_ENABLED=true
export MCP_AUTH_BEARER_TOKEN="your-token"

# 日志
export MCP_LOG_LEVEL=info
export MCP_LOG_FORMAT=json
```

## 认证方式

### 开发环境（命令行）
```bash
./mcp-server -transport http -auth-token "dev-123"
```

### 生产环境（环境变量）
```bash
export MCP_AUTH_ENABLED=true
export MCP_AUTH_BEARER_TOKEN="prod-secret"
./mcp-server -transport http
```

### 客户端使用
```bash
curl -H "Authorization: Bearer your-token" http://localhost:8080/mcp/health
```

##  内置工具

| 工具 | 功能 | 参数 |
|------|------|------|
| `ping` | 连接测试 | `message` (可选) |
| `echo` | 文本处理 | `text`, `uppercase`, `prefix`, `suffix`, `repeat` |
| `system_info` | 系统信息 | `info_type`: `runtime`/`memory`/`env`/`process` |

##  HTTP端点

| 端点 | 方法 | 描述 |
|------|------|------|
| `/` | GET | 重定向到 `/mcp` |
| `/mcp` | GET | 管理界面 |
| `/mcp/health` | GET | 健康检查 |
| `/mcp/status` | GET | 服务器状态 |
| `/mcp/info` | GET | 服务器信息 |
| `/` | GET | Web管理界面 |

##  健康检查

```bash
# 基本检查
curl http://localhost:8080/mcp/health

# 带认证
curl -H "Authorization: Bearer token" http://localhost:8080/mcp/health

# 响应示例
{
  "status": "healthy",
  "timestamp": "2026-02-12T07:15:40Z",
  "service": "ai-sre-mcp-server"
}
```

## 常见错误

| 错误 | 原因 | 解决方案 |
|------|------|----------|
| `401 Unauthorized` | 认证失败 | 检查token是否正确 |
| `bind: address already in use` | 端口被占用 | 使用其他端口或停止占用进程 |
| `invalid transport mode` | 传输模式错误 | 使用 `stdio`/`http`/`sse` |

##  调试命令

```bash
# 启用调试日志
MCP_LOG_LEVEL=debug ./mcp-server -transport http -auth-token "debug"

# 检查端口占用
lsof -i :8080

# 测试认证
curl -v -H "Authorization: Bearer token" http://localhost:8080/mcp/health
```

##  配置示例

### 开发配置
```bash
./mcp-server \
  -transport http \
  -port 9090 \
  -auth-token "dev-token-123"
```

### 生产配置
```bash
export MCP_AUTH_ENABLED=true
export MCP_AUTH_BEARER_TOKEN="$(openssl rand -hex 32)"
export MCP_AUTH_ALLOWED_IPS="10.0.0.0/8,192.168.0.0/16"
export MCP_LOG_FORMAT=json
export MCP_LOG_LEVEL=info

./mcp-server -transport http -port 8080
```

### Docker配置
```bash
docker run -d \
  -p 8080:8080 \
  -e MCP_AUTH_BEARER_TOKEN="your-token" \
  -e MCP_LOG_FORMAT=json \
  your-mcp-server:latest
```

## 🧪 测试脚本

```bash
# 运行所有测试
./tools/mcp/examples/test-all-modes.sh

# 启动测试
./tools/mcp/examples/test-startup.sh

# 自定义测试
curl -s -w "Status: %{http_code}\n" \
  -H "Authorization: Bearer test-token" \
  http://localhost:8080/mcp/health
```

##  文档链接

- **完整指南**: [docs/USER_GUIDE.md](USER_GUIDE.md)
- **API参考**: [docs/API_REFERENCE.md](API_REFERENCE.md)
- **使用示例**: [examples/usage-examples.md](../examples/usage-examples.md)

##  安全检查清单

- [ ] 使用强随机token（32位以上）
- [ ] 生产环境使用环境变量
- [ ] 配置IP白名单
- [ ] 启用结构化日志
- [ ] 定期轮换认证凭据
- [ ] 使用HTTPS（通过反向代理）

##  性能调优

```bash
# 并发配置
export MCP_MAX_CONCURRENT_REQUESTS=200
export MCP_REQUEST_TIMEOUT=120s
export MCP_TOOL_EXECUTION_TIMEOUT=60s

# 监控系统资源
curl -H "Authorization: Bearer token" \
  -X POST http://localhost:8080/mcp/tools/call \
  -d '{"name": "system_info", "arguments": {"info_type": "memory"}}'
```

---

**提示**: 使用 `./mcp-server -help` 查看完整的命令行选项