# AI SRE MCP Server - 使用示例

本文档提供了AI SRE MCP服务器的详细使用示例，包括不同传输模式和认证配置。

## 目录

- [基本使用](#基本使用)
- [传输模式](#传输模式)
- [认证配置](#认证配置)
- [环境变量配置](#环境变量配置)
- [客户端连接示例](#客户端连接示例)
- [故障排除](#故障排除)

## 基本使用

### 查看版本信息

```bash
./tools/mcp/bin/mcp-server -version
```

### 查看帮助信息

```bash
./tools/mcp/bin/mcp-server -help
```

### 默认启动（stdio模式）

```bash
./tools/mcp/bin/mcp-server
```

## 传输模式

### 1. Stdio模式（默认）

Stdio模式是MCP协议的标准模式，通过标准输入输出进行通信。

```bash
# 显式指定stdio模式
./tools/mcp/bin/mcp-server -transport stdio

# 或者使用环境变量
MCP_TRANSPORT=stdio ./tools/mcp/bin/mcp-server
```

**特点：**
- 适用于本地进程间通信
- 最轻量级的通信方式
- 无需网络配置
- 支持JSON-RPC协议

### 2. HTTP模式

HTTP模式提供基于HTTP的通信接口，同时包含管理界面。

```bash
# 启动HTTP模式服务器
./tools/mcp/bin/mcp-server -transport http -port 8080

# 使用环境变量配置
MCP_TRANSPORT=http MCP_PORT=8080 ./tools/mcp/bin/mcp-server
```

**可用端点：**
- `GET /mcp/mcp/health` - 健康检查
- `GET /` - 管理界面
- `POST /mcp` - MCP通信端点（计划中）

**测试HTTP服务器：**

```bash
# 健康检查
curl http://localhost:8080/mcp/mcp/health

# 访问管理界面
open http://localhost:8080  # macOS
# 或在浏览器中访问 http://localhost:8080
```

### 3. SSE模式（计划中）

SSE（Server-Sent Events）模式支持实时事件流。

```bash
# 当前会回退到stdio模式
./tools/mcp/bin/mcp-server -transport sse -port 8080
```

## 认证配置

### Bearer Token认证

#### 命令行配置

```bash
# 使用命令行参数
./tools/mcp/bin/mcp-server -transport http -port 8080 -auth-token "your-secret-token"

# 启用认证但不指定token（需要环境变量）
./tools/mcp/bin/mcp-server -transport http -enable-auth
```

#### 环境变量配置

```bash
# 设置Bearer Token
export MCP_AUTH_ENABLED=true
export MCP_AUTH_TYPE=bearer
export MCP_AUTH_BEARER_TOKEN="your-secret-token-123"

./tools/mcp/bin/mcp-server -transport http -port 8080
```

#### 客户端请求示例

```bash
# 无认证请求（会失败）
curl http://localhost:8080/mcp/mcp/health
# 返回: 401 Unauthorized

# 正确的认证请求
curl -H "Authorization: Bearer your-secret-token-123" http://localhost:8080/mcp/mcp/health
# 返回: {"status": "healthy", ...}

# 错误的token（会失败）
curl -H "Authorization: Bearer wrong-token" http://localhost:8080/mcp/mcp/health
# 返回: 401 Unauthorized
```

### API Key认证（计划中）

```bash
# 环境变量配置
export MCP_AUTH_ENABLED=true
export MCP_AUTH_TYPE=api_key
export MCP_AUTH_API_KEY="your-api-key-456"

./tools/mcp/bin/mcp-server -transport http
```

```bash
# 客户端请求
curl -H "X-API-Key: your-api-key-456" http://localhost:8080/mcp/mcp/health
# 或使用查询参数
curl "http://localhost:8080/mcp/mcp/health?api_key=your-api-key-456"
```

### Basic认证（计划中）

```bash
# 环境变量配置
export MCP_AUTH_ENABLED=true
export MCP_AUTH_TYPE=basic
export MCP_AUTH_USERNAME="admin"
export MCP_AUTH_PASSWORD="password123"

./tools/mcp/bin/mcp-server -transport http
```

```bash
# 客户端请求
curl -u admin:password123 http://localhost:8080/mcp/mcp/health
```

### IP白名单

```bash
# 限制访问IP
export MCP_AUTH_ENABLED=true
export MCP_AUTH_BEARER_TOKEN="secret123"
export MCP_AUTH_ALLOWED_IPS="127.0.0.1,192.168.1.0/24,10.0.0.1"

./tools/mcp/bin/mcp-server -transport http
```

## 环境变量配置

### 完整配置示例

```bash
# 服务器配置
export MCP_HOST=0.0.0.0
export MCP_PORT=8080
export MCP_READ_TIMEOUT=30s
export MCP_WRITE_TIMEOUT=30s

# 传输和协议
export MCP_TRANSPORT=http
export MCP_SERVER_NAME=my-mcp-server
export MCP_PROTOCOL_VERSION=2024-11-05

# 认证配置
export MCP_AUTH_ENABLED=true
export MCP_AUTH_TYPE=bearer
export MCP_AUTH_BEARER_TOKEN=super-secret-token-789
export MCP_AUTH_ALLOWED_IPS=127.0.0.1,192.168.1.0/24

# 日志配置
export MCP_LOG_LEVEL=info
export MCP_LOG_FORMAT=json
export MCP_LOG_FILE=/var/log/mcp-server.log

# 工具配置
export MCP_TOOL_TIMEOUT=60s
export MCP_ENABLE_TOOLS=true

# 启动服务器
./tools/mcp/bin/mcp-server
```

### 开发环境配置

```bash
# 开发环境 - 启用调试日志
export MCP_LOG_LEVEL=debug
export MCP_LOG_FORMAT=text
export MCP_AUTH_ENABLED=false

./tools/mcp/bin/mcp-server -transport http -port 9090
```

### 生产环境配置

```bash
# 生产环境 - 安全配置
export MCP_HOST=0.0.0.0
export MCP_PORT=8080
export MCP_AUTH_ENABLED=true
export MCP_AUTH_TYPE=bearer
export MCP_AUTH_BEARER_TOKEN=$(openssl rand -base64 32)
export MCP_AUTH_ALLOWED_IPS=10.0.0.0/8,172.16.0.0/12,192.168.0.0/16
export MCP_LOG_LEVEL=info
export MCP_LOG_FORMAT=json
export MCP_LOG_FILE=/var/log/mcp-server.log

./tools/mcp/bin/mcp-server -transport http
```

## 客户端连接示例

### Stdio模式客户端

```python
import json
import subprocess
import sys

# 启动MCP服务器
process = subprocess.Popen(
    ['./tools/mcp/bin/mcp-server'],
    stdin=subprocess.PIPE,
    stdout=subprocess.PIPE,
    stderr=subprocess.PIPE,
    text=True
)

# 发送ping请求
request = {
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/call",
    "params": {
        "name": "ping",
        "arguments": {"message": "Hello from client!"}
    }
}

# 发送请求
process.stdin.write(json.dumps(request) + '\n')
process.stdin.flush()

# 读取响应
response = process.stdout.readline()
print("Response:", response)

# 清理
process.terminate()
```

### HTTP模式客户端

```python
import requests
import json

# 服务器配置
BASE_URL = "http://localhost:8080"
AUTH_TOKEN = "your-secret-token"

# 设置认证头
headers = {
    "Authorization": f"Bearer {AUTH_TOKEN}",
    "Content-Type": "application/json"
}

# 健康检查
response = requests.get(f"{BASE_URL}/mcp/health", headers=headers)
print("Health check:", response.json())

# MCP工具调用（计划中的功能）
# mcp_request = {
#     "jsonrpc": "2.0",
#     "id": 1,
#     "method": "tools/call",
#     "params": {
#         "name": "ping",
#         "arguments": {"message": "Hello via HTTP!"}
#     }
# }
# 
# response = requests.post(f"{BASE_URL}/mcp", 
#                         headers=headers, 
#                         json=mcp_request)
# print("MCP Response:", response.json())
```

### JavaScript/Node.js客户端

```javascript
const axios = require('axios');

const client = axios.create({
  baseURL: 'http://localhost:8080',
  headers: {
    'Authorization': 'Bearer your-secret-token',
    'Content-Type': 'application/json'
  }
});

// 健康检查
async function healthCheck() {
  try {
    const response = await client.get('/mcp/health');
    console.log('Health check:', response.data);
  } catch (error) {
    console.error('Health check failed:', error.response?.data);
  }
}

healthCheck();
```

## 故障排除

### 常见问题

#### 1. 认证失败

**问题：** 收到401 Unauthorized错误

**解决方案：**
```bash
# 检查token是否正确
curl -v -H "Authorization: Bearer your-token" http://localhost:8080/mcp/mcp/health

# 检查服务器日志
MCP_LOG_LEVEL=debug ./tools/mcp/bin/mcp-server -transport http -auth-token "your-token"
```

#### 2. 端口被占用

**问题：** 服务器启动失败，端口被占用

**解决方案：**
```bash
# 检查端口使用情况
lsof -i :8080

# 使用不同端口
./tools/mcp/bin/mcp-server -transport http -port 9090
```

#### 3. 配置验证失败

**问题：** 配置参数无效

**解决方案：**
```bash
# 检查配置
./tools/mcp/bin/mcp-server -transport invalid
# 会显示: invalid transport mode: invalid, valid options: [stdio sse http]

# 使用正确的配置
./tools/mcp/bin/mcp-server -transport http
```

### 调试技巧

#### 启用调试日志

```bash
# 详细日志输出
MCP_LOG_LEVEL=debug MCP_LOG_FORMAT=text ./tools/mcp/bin/mcp-server -transport http
```

#### 检查服务器状态

```bash
# 健康检查
curl http://localhost:8080/mcp/mcp/health

# 查看管理界面
curl http://localhost:8080/
```

#### 网络连接测试

```bash
# 测试端口连通性
nc -zv localhost 8080

# 测试HTTP响应
curl -I http://localhost:8080/mcp/mcp/health
```

## 性能调优

### 并发配置

```bash
# 增加并发请求数
export MCP_MAX_CONCURRENT_REQUESTS=200

# 调整超时时间
export MCP_REQUEST_TIMEOUT=120s
export MCP_TOOL_TIMEOUT=60s

./tools/mcp/bin/mcp-server -transport http
```

### 日志优化

```bash
# 生产环境 - 减少日志输出
export MCP_LOG_LEVEL=warn
export MCP_LOG_FORMAT=json
export MCP_LOG_FILE=/var/log/mcp-server.log

# 开发环境 - 详细日志
export MCP_LOG_LEVEL=debug
export MCP_LOG_FORMAT=text
```

## 部署建议

### Docker部署

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o mcp-server ./cmd/mcp-server

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/mcp-server .
EXPOSE 8080
CMD ["./mcp-server", "-transport", "http", "-port", "8080"]
```

### 系统服务

```ini
# /etc/systemd/system/mcp-server.service
[Unit]
Description=AI SRE MCP Server
After=network.target

[Service]
Type=simple
User=mcp
WorkingDirectory=/opt/mcp
ExecStart=/opt/mcp/bin/mcp-server -transport http -port 8080
Environment=MCP_AUTH_ENABLED=true
Environment=MCP_AUTH_BEARER_TOKEN=your-production-token
Environment=MCP_LOG_FILE=/var/log/mcp-server.log
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

---

更多信息请参考项目文档或提交Issue。