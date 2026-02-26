# AI SRE MCP Server

一个基于 Model Context Protocol (MCP) 的智能运维服务器，支持多种传输模式和完整的认证系统。

## 🎯 核心功能

### ✅ 已实现功能

1. **多传输模式支持**
   - ✅ **stdio** - 标准输入输出模式（适用于本地集成）
   - ✅ **http** - HTTP 模式（适用于远程访问和 Web 集成）

2. **基础工具集**
   - ✅ **ping** - 连接测试工具
   - ✅ **echo** - 高级文本处理工具（支持大小写转换、前缀后缀、重复等）
   - ✅ **system_info** - 系统运行时信息获取

3. **腾讯云工具集**
   - ✅ **tke_describe_regions** - 查询 TKE 支持的地域信息
   - ✅ **tke_get_region** - 查询特定地域的详细信息
   - ✅ **tencentcloud_validate** - 验证腾讯云 API 连接和权限

4. **动态工具发现**
   - ✅ **动态工具列表** - 客户端可以动态获取服务器实际注册的所有工具
   - ✅ **完整工具信息** - 包括工具描述、参数模式、类型定义等
   - ✅ **实时同步** - 工具注册状态与客户端可见性实时同步

5. **企业级特性**
   - ✅ **认证授权** - Bearer Token 和 API Key 认证支持
   - ✅ **配置管理** - 灵活的配置文件和环境变量支持
   - ✅ **日志记录** - 结构化日志，支持不同级别
   - ✅ **健康检查** - 完整的健康检查和状态监控端点
   - ✅ **优雅关闭** - 支持优雅关闭和资源清理

6. **开发和测试**
   - ✅ **MCP 客户端** - 内置测试客户端，支持 stdio 和 HTTP 模式
   - ✅ **管理界面** - Web 管理界面，支持工具查看和状态监控
   - ✅ **API 文档** - 完整的 REST API 文档和示例

##  快速开始

```bash
# 构建服务器
make build-go

# 默认stdio模式启动
./tools/mcp/bin/mcp-server

# HTTP模式启动（带认证）
./tools/mcp/bin/mcp-server -transport http -port 8080 -auth-token "your-secret-token"

# 查看帮助
./tools/mcp/bin/mcp-server -help

docker    run  --name ai-sre -itd --network host -e TENCENTCLOUD_SECRET_ID=sID -e TENCENTCLOUD_SECRET_KEY=sKey mirrors.tencent.com/tke-oss/ai-sre-mcp-server  /app/mcp-server -transport http -port 8081 --log-level debug

```

## 🔧 AI Chat工具配置

**重要更新**: MCP服务器现在支持真正的HTTP传输模式！

### HTTP模式配置（推荐）

1. **启动HTTP MCP服务器**:
   ```bash
   ./tools/mcp/bin/mcp-server -transport http -port 8082
   ```

2. **在AI Chat工具中配置**:
   - **服务器类型**: HTTP MCP Server
   - **URL**: `http://localhost:8082/mcp`
   - **协议版本**: `2024-11-05`

### stdio模式配置（备选）

如果AI Chat工具不支持HTTP模式：

- **服务器类型**: 可执行程序 (Executable)
- **可执行文件路径**: `/Users/cloudnativesre/Desktop/ai-sre/tools/mcp/bin/mcp-server`
- **传输协议**: stdio (默认)

### ✅ 验证连接

HTTP模式测试：
```bash
curl -X POST http://localhost:8082/mcp \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -H "MCP-Protocol-Version: 2024-11-05" \
  -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{"tools":{}},"clientInfo":{"name":"test-client","version":"1.0.0"}}}'
```

详细配置说明请参考: [MCP_CLIENT_SETUP.md](./MCP_CLIENT_SETUP.md)

##  主要特性

-  **多传输模式**: stdio（默认）、HTTP、SSE
-  **完整认证系统**: Bearer Token、IP白名单、多种认证类型
-  **内置SRE工具**: ping、echo、system_info
-  **Web管理界面**: 健康检查、状态监控
-  **优雅关闭**: 信号处理、资源清理
-  **结构化日志**: JSON/文本格式、可配置级别
-  **生产就绪**: 配置验证、错误处理、性能优化

##  传输模式

### stdio模式（默认）
```bash
./tools/mcp/bin/mcp-server
```
适用于MCP客户端直接连接。

### HTTP模式
```bash
# 无认证
./tools/mcp/bin/mcp-server -transport http -port 8080

# 带认证
./tools/mcp/bin/mcp-server -transport http -port 8080 -auth-token "secret"
```
提供HTTP接口和Web管理界面。

## 认证配置

### 方式1: 命令行参数（开发推荐）
```bash
./tools/mcp/bin/mcp-server -transport http -auth-token "dev-token-123"
```

### 方式2: 环境变量（生产推荐）
```bash
export MCP_AUTH_ENABLED=true
export MCP_AUTH_BEARER_TOKEN="prod-secret-token"
./tools/mcp/bin/mcp-server -transport http
```

### 客户端使用
```bash
# 正确认证
curl -H "Authorization: Bearer your-token" http://localhost:8080/health

# 无认证（返回401）
curl http://localhost:8080/health
```

##  健康检查

```bash
# 基本检查
curl http://localhost:8080/health

# 带认证检查
curl -H "Authorization: Bearer token" http://localhost:8080/health
```

**响应示例**:
```json
{
  "status": "healthy",
  "timestamp": "2026-02-12T07:15:40Z",
  "service": "ai-sre-mcp-server",
  "transport": "stdio"
}
```

##  内置工具

### 🔄 动态工具发现

**重要特性**: MCP 服务器现在支持动态工具发现！客户端可以实时获取服务器上实际注册的所有工具。

#### 工具列表获取示例

```bash
# 1. 初始化连接
curl -X POST http://localhost:8085/mcp \
  -H "Content-Type: application/json" \
  -H "MCP-Protocol-Version: 2024-11-05" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "initialize",
    "params": {
      "protocolVersion": "2024-11-05",
      "capabilities": {},
      "clientInfo": {
        "name": "test-client",
        "version": "1.0.0"
      }
    }
  }'

# 2. 获取动态工具列表
curl -X POST http://localhost:8085/mcp \
  -H "Content-Type: application/json" \
  -H "MCP-Protocol-Version: 2024-11-05" \
  -d '{
    "jsonrpc": "2.0",
    "id": 2,
    "method": "tools/list"
  }' | jq .
```

#### 响应示例

```json
{
  "id": 2,
  "jsonrpc": "2.0",
  "result": {
    "tools": [
      {
        "name": "ping",
        "description": "简单的ping工具，用于测试MCP服务器连接和响应",
        "inputSchema": {
          "type": "object",
          "properties": {
            "message": {
              "type": "string",
              "description": "要返回的消息",
              "default": "pong"
            }
          }
        }
      },
      {
        "name": "tke_describe_regions",
        "description": "查询腾讯云 TKE (容器服务) 支持的地域信息",
        "inputSchema": {
          "type": "object",
          "properties": {
            "format": {
              "type": "string",
              "description": "输出格式：json 或 table",
              "enum": ["json", "table"],
              "default": "json"
            }
          }
        }
      }
    ]
  }
}
```

### 📋 工具清单

| 工具 | 描述 | 用途 |
|------|------|------|
| `ping` | 连接测试 | 测试MCP服务器连接和响应 |
| `echo` | 文本处理 | 大小写转换、前缀后缀、重复 |
| `system_info` | 系统信息 | 运行时、内存、环境、进程信息 |
| `tke_describe_regions` | TKE地域查询 | 查询腾讯云TKE支持的地域信息 |
| `tke_get_region` | TKE地域详情 | 查询特定地域的详细信息 |
| `tencentcloud_validate` | 腾讯云验证 | 验证腾讯云API连接和权限 |

**注意**: 腾讯云工具需要设置环境变量 `TENCENTCLOUD_SECRET_ID` 和 `TENCENTCLOUD_SECRET_KEY` 才能正常工作。

### 🐛 调试和排障

#### 启用详细日志

为了方便排障，服务器支持详细的调试日志，可以显示所有 MCP 消息的完整内容：

```bash
# 启用 debug 日志级别
./mcp-server --transport http --port 8085 --log-level debug

# 或者通过环境变量
MCP_LOG_LEVEL=debug ./mcp-server --transport http --port 8085
```

#### Debug 日志内容

启用 debug 日志后，你将看到：

1. **完整的请求和响应内容**：
   ```json
   {
     "full_message": "{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"tools/list\"}",
     "level": "debug",
     "msg": "Full MCP message content"
   }
   ```

2. **动态工具获取过程**：
   ```json
   {
     "registered_tools_count": 6,
     "registered_tools": ["ping", "echo", "system_info", "tke_describe_regions", "tke_get_region", "tencentcloud_validate"],
     "level": "debug", 
     "msg": "Retrieved registered tools from MCPServer"
   }
   ```

3. **工具执行详情**：
   ```json
   {
     "tool_name": "ping",
     "arguments": {"message": "test"},
     "result_preview": "Ping response: test",
     "level": "debug",
     "msg": "Tool execution completed successfully"
   }
   ```

4. **完整的响应内容**：
   ```json
   {
     "full_response": "{\"id\":2,\"jsonrpc\":\"2.0\",\"result\":{\"tools\":[...]}}",
     "level": "debug",
     "msg": "Full tools/list response content"
   }
   ```

#### 排障指南

- **工具不可见**：检查 debug 日志中的 `registered_tools` 字段，确认工具是否成功注册
- **工具调用失败**：查看 `Tool execution failed` 错误信息和参数
- **协议问题**：检查 `Full MCP message content` 确认请求格式正确
- **响应异常**：查看 `Full response content` 确认服务器返回的完整内容

## 环境变量

### 核心配置
```bash
export MCP_TRANSPORT=http              # 传输模式
export MCP_PORT=8080                   # 服务端口
export MCP_LOG_LEVEL=info              # 日志级别
```

### 认证配置
```bash
export MCP_AUTH_ENABLED=true           # 启用认证
export MCP_AUTH_BEARER_TOKEN="token"   # Bearer令牌
export MCP_AUTH_ALLOWED_IPS="10.0.0.0/8" # IP白名单
```

##  使用示例

### 开发环境
```bash
./tools/mcp/bin/mcp-server -transport http -port 9090 -auth-token "dev-123"
```

### 生产环境
```bash
export MCP_AUTH_ENABLED=true
export MCP_AUTH_BEARER_TOKEN="$(openssl rand -hex 32)"
export MCP_LOG_FORMAT=json
./tools/mcp/bin/mcp-server -transport http
```

### Docker部署
```bash
docker run -d \
  -p 8080:8080 \
  -e MCP_AUTH_BEARER_TOKEN="your-token" \
  your-mcp-server:latest
```

##  测试验证

```bash
# 运行完整测试
./tools/mcp/examples/test-all-modes.sh

# 测试特定功能
./tools/mcp/examples/test-startup.sh
```

## 管理端点

当使用HTTP传输模式时，服务器提供两套独立的管理端点：

### 通用管理端点

| 端点 | 描述 | 响应格式 |
|------|------|----------|
| `GET /` | 通用服务器管理界面 | HTML |
| `GET /health` | 通用健康检查 | JSON |
| `GET /status` | 通用服务器状态 | JSON |

### MCP专用端点

| 端点 | 描述 | 响应格式 |
|------|------|----------|
| `GET /mcp` | MCP服务器管理界面 | HTML |
| `GET /mcp/health` | MCP健康检查 | JSON |
| `GET /mcp/status` | MCP服务器状态和配置信息 | JSON |
| `GET /mcp/info` | MCP服务器能力和文档链接 | JSON |
| `GET /mcp/tools` | MCP工具列表和描述 | JSON |

### 使用示例

#### 通用管理
```bash
# 访问通用管理界面
curl http://localhost:8080/

# 通用健康检查
curl http://localhost:8080/health

# 通用服务器状态
curl http://localhost:8080/status
```

#### MCP工具管理
```bash
# 访问MCP管理界面
curl http://localhost:8080/mcp

# MCP健康检查
curl http://localhost:8080/mcp/health

# 获取MCP服务器状态
curl http://localhost:8080/mcp/status

# 获取MCP服务器能力信息
curl http://localhost:8080/mcp/info

# 获取MCP工具列表
curl http://localhost:8080/mcp/tools
```

### 认证访问
```bash
# 使用Bearer Token访问任何端点
curl -H "Authorization: Bearer your-token" http://localhost:8080/health
curl -H "Authorization: Bearer your-token" http://localhost:8080/mcp/health
```

### 端点区别

#### 通用端点特点
- **服务标识**: `ai-sre-server`
- **用途**: 通用服务器管理和监控
- **范围**: 整体服务器状态，不特定于MCP

#### MCP端点特点
- **服务标识**: `ai-sre-mcp-server`
- **用途**: MCP协议相关的工具和功能
- **范围**: MCP特定的健康检查、工具管理、能力展示

##  文档

- **完整用户指南**: [docs/USER_GUIDE.md](docs/USER_GUIDE.md)
- **使用示例**: [examples/usage-examples.md](examples/usage-examples.md)
- **API文档**: 访问 `http://localhost:8080/` (通用管理) 或 `http://localhost:8080/mcp` (MCP工具) 查看Web界面

## 故障排除

### 认证失败
```bash
# 检查token
echo $MCP_AUTH_BEARER_TOKEN

# 启用调试日志
MCP_LOG_LEVEL=debug ./tools/mcp/bin/mcp-server -transport http -auth-token "test"
```

### 端口占用
```bash
# 查看端口
lsof -i :8080

# 使用其他端口
./tools/mcp/bin/mcp-server -transport http -port 9090
```

##  安全最佳实践

-  使用强随机token（32位以上）
-  生产环境使用环境变量
-  配置IP白名单
-  定期轮换认证凭据
-  避免在命令行中暴露敏感信息

##  性能配置

```bash
export MCP_MAX_CONCURRENT_REQUESTS=200  # 最大并发
export MCP_REQUEST_TIMEOUT=120s         # 请求超时
export MCP_TOOL_EXECUTION_TIMEOUT=60s   # 工具超时
```

##  版本信息

```bash
# 查看版本
./tools/mcp/bin/mcp-server -version

# 输出示例
AI SRE MCP Server
Version: dev
Commit: unknown
Build Time: unknown
```

##  构建和开发

```bash
# 构建项目
cd tools/mcp
go build -o bin/mcp-server ./cmd/mcp-server

# 或使用Makefile
make build-go

# 运行测试
go test ./...

# 代码格式化
go fmt ./...
```

## 支持

如有问题或建议，请：
1. 查看 [完整用户指南](docs/USER_GUIDE.md)
2. 启用调试日志进行诊断
3. 提交Issue到项目仓库

---

**License**: MIT  
**Version**: v1.0.0  
**Maintainer**: AI SRE Team
