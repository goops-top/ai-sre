# AI SRE MCP Server 用户指南

## 📖 概述

AI SRE MCP Server 是一个基于 Model Context Protocol (MCP) 的智能运维服务器，支持多种传输模式和完整的认证系统。本服务器提供了丰富的SRE工具集，可以通过不同的传输方式进行访问。

##  快速开始

### 基本启动
```bash
# 默认stdio模式启动
./tools/mcp/bin/mcp-server

# 查看版本信息
./tools/mcp/bin/mcp-server -version

# 查看帮助信息
./tools/mcp/bin/mcp-server -help
```

### HTTP模式启动
```bash
# HTTP模式，无认证
./tools/mcp/bin/mcp-server -transport http -port 8080

# HTTP模式，带认证
./tools/mcp/bin/mcp-server -transport http -port 8080 -auth-token "your-secret-token"
```

##  传输模式

### 1. stdio模式（默认）
- **描述**: 标准的MCP协议通信方式，通过标准输入输出
- **适用场景**: MCP客户端直接连接，如IDE插件、CLI工具
- **启动方式**: 
  ```bash
  ./tools/mcp/bin/mcp-server
  # 或明确指定
  ./tools/mcp/bin/mcp-server -transport stdio
  ```

### 2. HTTP模式
- **描述**: 提供HTTP接口和Web管理界面
- **适用场景**: Web应用、API调用、远程管理
- **功能特性**:
  - 健康检查端点: `/mcp/health`
  - Web管理界面: `/`
  - 支持认证和IP白名单
- **启动方式**:
  ```bash
  ./tools/mcp/bin/mcp-server -transport http -port 8080
  ```

### 3. SSE模式
- **描述**: Server-Sent Events模式（当前回退到stdio）
- **适用场景**: 实时数据推送、事件流
- **启动方式**:
  ```bash
  ./tools/mcp/bin/mcp-server -transport sse -port 8080
  ```

## 认证系统

### 认证类型

#### 1. Bearer Token认证（推荐）
最常用的认证方式，支持HTTP头认证。

**配置方式**:
```bash
# 方式1: 命令行参数（开发环境推荐）
./tools/mcp/bin/mcp-server -transport http -auth-token "your-secret-token"

# 方式2: 环境变量（生产环境推荐）
export MCP_AUTH_ENABLED=true
export MCP_AUTH_BEARER_TOKEN="your-secret-token"
./tools/mcp/bin/mcp-server -transport http
```

**客户端使用**:
```bash
# 正确的认证请求
curl -H "Authorization: Bearer your-secret-token" http://localhost:8080/mcp/health

# 错误的请求（返回401）
curl http://localhost:8080/mcp/health
```

#### 2. API Key认证（框架支持，待实现）
```bash
export MCP_AUTH_TYPE=api_key
export MCP_AUTH_API_KEY="your-api-key"
```

#### 3. Basic认证（框架支持，待实现）
```bash
export MCP_AUTH_TYPE=basic
export MCP_AUTH_USERNAME="admin"
export MCP_AUTH_PASSWORD="password"
```

### IP白名单
```bash
# 允许特定IP访问
export MCP_AUTH_ALLOWED_IPS="192.168.1.100,10.0.0.0/8"
./tools/mcp/bin/mcp-server -transport http -enable-auth
```

##  命令行参数

### 基本参数
| 参数 | 描述 | 默认值 | 示例 |
|------|------|--------|------|
| `-version` | 显示版本信息并退出 | - | `./mcp-server -version` |
| `-help` | 显示帮助信息并退出 | - | `./mcp-server -help` |
| `-config <file>` | 指定配置文件路径 | - | `./mcp-server -config config.yaml` |

### 传输配置
| 参数 | 描述 | 可选值 | 默认值 |
|------|------|--------|--------|
| `-transport <mode>` | 传输模式 | `stdio`, `sse`, `http` | `stdio` |
| `-port <port>` | 服务器端口（仅HTTP/SSE模式） | 1-65535 | `8080` |

### 认证配置
| 参数 | 描述 | 作用 |
|------|------|------|
| `-auth-token <token>` | Bearer认证令牌 | 自动启用认证并设置token |
| `-enable-auth` | 启用认证 | 仅启用认证，需配合环境变量 |

### 参数使用说明

#### `--auth-token` vs `--enable-auth`

**`--auth-token`（推荐用于开发）**:
-  一步到位：自动启用认证并设置token
-  简单直接，适合快速测试
-  token在命令行中可见

```bash
./tools/mcp/bin/mcp-server -transport http -auth-token "dev-token-123"
```

**`--enable-auth`（推荐用于生产）**:
-  仅启用认证功能
-  配合环境变量，更安全
-  适合容器化部署

```bash
export MCP_AUTH_BEARER_TOKEN="prod-secret-token"
./tools/mcp/bin/mcp-server -transport http -enable-auth
```

## 环境变量配置

### 服务器配置
| 环境变量 | 描述 | 默认值 |
|----------|------|--------|
| `MCP_SERVER_NAME` | 服务器名称 | `ai-sre-mcp-server` |
| `MCP_SERVER_VERSION` | 服务器版本 | `1.0.0` |
| `MCP_PORT` | 服务器端口 | `8080` |
| `MCP_HOST` | 服务器主机 | `localhost` |

### MCP协议配置
| 环境变量 | 描述 | 默认值 |
|----------|------|--------|
| `MCP_PROTOCOL_VERSION` | 协议版本 | `2024-11-05` |
| `MCP_TRANSPORT` | 传输模式 | `stdio` |
| `MCP_REQUEST_TIMEOUT` | 请求超时时间 | `60s` |
| `MCP_MAX_CONCURRENT_REQUESTS` | 最大并发请求数 | `100` |

### 认证配置
| 环境变量 | 描述 | 默认值 |
|----------|------|--------|
| `MCP_AUTH_ENABLED` | 是否启用认证 | `false` |
| `MCP_AUTH_TYPE` | 认证类型 | `bearer` |
| `MCP_AUTH_BEARER_TOKEN` | Bearer令牌 | - |
| `MCP_AUTH_API_KEY` | API密钥 | - |
| `MCP_AUTH_USERNAME` | 用户名（Basic认证） | - |
| `MCP_AUTH_PASSWORD` | 密码（Basic认证） | - |
| `MCP_AUTH_ALLOWED_IPS` | 允许的IP地址列表（逗号分隔） | - |
| `MCP_AUTH_TOKEN_EXPIRY` | Token过期时间 | `24h` |

### 日志配置
| 环境变量 | 描述 | 默认值 |
|----------|------|--------|
| `MCP_LOG_LEVEL` | 日志级别 | `info` |
| `MCP_LOG_FORMAT` | 日志格式 | `json` |
| `MCP_LOG_FILE` | 日志文件路径 | - |

### 工具配置
| 环境变量 | 描述 | 默认值 |
|----------|------|--------|
| `MCP_TOOL_EXECUTION_TIMEOUT` | 工具执行超时时间 | `30s` |
| `MCP_ENABLE_TOOLS` | 是否启用工具 | `true` |

##  内置工具

### 1. ping工具
**描述**: 简单的连接测试工具
**功能**: 返回指定消息或默认的'pong'响应
**参数**:
- `message` (可选): 自定义返回消息

### 2. echo工具  
**描述**: 高级文本处理和格式化工具
**功能**: 支持大小写转换、前缀后缀添加、文本重复等
**参数**:
- `text` (必需): 要处理的文本
- `uppercase` (可选): 转换为大写
- `lowercase` (可选): 转换为小写
- `prefix` (可选): 添加前缀
- `suffix` (可选): 添加后缀
- `repeat` (可选): 重复次数

### 3. system_info工具
**描述**: 获取系统运行时信息
**功能**: 包括Go运行时、内存使用、环境变量、进程信息等
**参数**:
- `info_type` (可选): 信息类型 (`runtime`, `memory`, `env`, `process`)

##  使用示例

### 开发环境快速启动
```bash
# 启动HTTP服务器，带认证
./tools/mcp/bin/mcp-server -transport http -port 9090 -auth-token "dev-123"

# 测试健康检查
curl -H "Authorization: Bearer dev-123" http://localhost:9090/mcp/health
```

### 生产环境部署
```bash
# 设置环境变量
export MCP_AUTH_ENABLED=true
export MCP_AUTH_BEARER_TOKEN="$(openssl rand -hex 32)"
export MCP_AUTH_ALLOWED_IPS="10.0.0.0/8,192.168.0.0/16"
export MCP_LOG_LEVEL=info
export MCP_LOG_FORMAT=json

# 启动服务器
./tools/mcp/bin/mcp-server -transport http -port 8080
```

### Docker容器部署
```dockerfile
FROM alpine:latest
COPY mcp-server /usr/local/bin/
EXPOSE 8080

ENV MCP_AUTH_ENABLED=true
ENV MCP_LOG_FORMAT=json

CMD ["mcp-server", "-transport", "http", "-port", "8080"]
```

```bash
# 运行容器
docker run -d \
  -p 8080:8080 \
  -e MCP_AUTH_BEARER_TOKEN="your-secret-token" \
  your-mcp-server:latest
```

### 多实例负载均衡
```bash
# 实例1
MCP_AUTH_BEARER_TOKEN="shared-token" ./mcp-server -transport http -port 8081 &

# 实例2  
MCP_AUTH_BEARER_TOKEN="shared-token" ./mcp-server -transport http -port 8082 &

# 实例3
MCP_AUTH_BEARER_TOKEN="shared-token" ./mcp-server -transport http -port 8083 &
```

##  健康检查和监控

### 健康检查端点
```bash
# 基本健康检查
curl http://localhost:8080/mcp/health

# 带认证的健康检查
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

### Web管理界面
访问 `http://localhost:8080/` 查看Web管理界面，包含：
- 服务器状态信息
- 已注册工具列表
- 配置信息展示
- 实时日志查看

## 故障排除

### 常见问题

#### 1. 认证失败
**问题**: 收到401 Unauthorized错误
**解决方案**:
```bash
# 检查token是否正确
curl -v -H "Authorization: Bearer your-token" http://localhost:8080/mcp/health

# 检查环境变量
echo $MCP_AUTH_BEARER_TOKEN

# 查看服务器日志
MCP_LOG_LEVEL=debug ./mcp-server -transport http -auth-token "test"
```

#### 2. 端口被占用
**问题**: 启动时提示端口被占用
**解决方案**:
```bash
# 查看端口占用
lsof -i :8080

# 使用其他端口
./mcp-server -transport http -port 9090
```

#### 3. 配置验证失败
**问题**: 启动时配置验证错误
**解决方案**:
```bash
# 检查传输模式
./mcp-server -transport invalid  # 会显示有效选项

# 检查认证配置
MCP_AUTH_ENABLED=true ./mcp-server -transport http  # 需要提供token
```

### 调试模式
```bash
# 启用详细日志
MCP_LOG_LEVEL=debug ./mcp-server -transport http -auth-token "debug"

# 查看所有配置
MCP_LOG_LEVEL=debug ./mcp-server -version
```

##  安全最佳实践

### 1. Token管理
-  使用强随机token（32位以上）
-  定期轮换token
-  生产环境使用环境变量
-  避免在命令行中暴露token

### 2. 网络安全
-  使用IP白名单限制访问
-  在反向代理后运行（如nginx）
-  启用HTTPS（通过反向代理）
-  避免直接暴露到公网

### 3. 日志安全
-  定期清理日志文件
-  避免在日志中记录敏感信息
-  使用结构化日志便于分析

##  性能调优

### 并发配置
```bash
# 调整最大并发请求数
export MCP_MAX_CONCURRENT_REQUESTS=200

# 调整请求超时时间
export MCP_REQUEST_TIMEOUT=120s

# 调整工具执行超时
export MCP_TOOL_EXECUTION_TIMEOUT=60s
```

### 资源监控
```bash
# 查看系统信息
curl -H "Authorization: Bearer token" \
  -X POST http://localhost:8080/mcp/tools/call \
  -d '{"name": "system_info", "arguments": {"info_type": "memory"}}'
```

##  版本升级

### 检查当前版本
```bash
./mcp-server -version
```

### 平滑升级
```bash
# 1. 备份当前版本
cp mcp-server mcp-server.backup

# 2. 替换新版本
cp new-mcp-server mcp-server

# 3. 验证新版本
./mcp-server -version

# 4. 重启服务（支持优雅关闭）
kill -TERM $MCP_PID
./mcp-server -transport http -auth-token "your-token"
```

## 支持和反馈

如果您在使用过程中遇到问题或有改进建议，请：

1. 查看本文档的故障排除部分
2. 启用调试日志进行诊断
3. 提交Issue到项目仓库
4. 联系开发团队

---

**版本**: v1.0.0  
**最后更新**: 2026-02-12  
**文档维护**: AI SRE Team