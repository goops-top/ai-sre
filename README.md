# AI SRE 智能运维助理

> 基于 MCP (Model Context Protocol) 构建的智能化 SRE 运维工具系统，为 AI 模型提供腾讯云基础设施查询能力

[![Go](https://img.shields.io/badge/go-1.24+-blue.svg)](https://golang.org)
[![Docker](https://img.shields.io/badge/docker-ready-blue.svg)](https://docker.com)
[![MCP](https://img.shields.io/badge/MCP-2025--06--18-green.svg)](https://modelcontextprotocol.io)

## 项目简介

AI SRE 是一个面向 SRE 场景的 MCP 工具服务器，通过标准化的 MCP 协议为 AI 模型（如 Claude、GPT 等）提供腾讯云基础设施的实时查询能力。当前已实现 **36 个只读查询工具**，覆盖 TKE、CVM、CLB、CDB、VPC 等核心云产品，同时提供独立的 Kubernetes 集群操作公共库。

### 核心特性

- **36 个 MCP 工具** — 覆盖腾讯云 6 大产品线的只读查询
- **3 种传输模式** — 支持 stdio / HTTP / SSE 三种协议通信
- **多认证方式** — Bearer Token / Basic Auth / API Key + IP 白名单
- **Kubernetes 公共库** — 独立模块，支持命名空间、工作负载、Pod 状态查询
- **双输出格式** — 所有云产品工具支持 JSON 和 Table 两种输出

## 系统架构

```
┌─────────────────────────────────────────────────────────┐
│  AI 模型层    │  Claude  │  GPT  │  其他 LLM            │
├─────────────────────────────────────────────────────────┤
│  协议层       │  stdio  │  HTTP JSON-RPC  │  SSE        │
├─────────────────────────────────────────────────────────┤
│  MCP Server   │  工具管理  │  认证中间件  │  路由分发    │
├─────────────────────────────────────────────────────────┤
│  工具层       │ TKE(11) │ CVM(2) │ CLB(4) │ CDB(4)     │
│               │ VPC(9)  │ Region(3) │ 内置(3)           │
├─────────────────────────────────────────────────────────┤
│  SDK/客户端层 │  腾讯云 SDK  │  K8s client-go           │
├─────────────────────────────────────────────────────────┤
│  基础设施     │  腾讯云  │  Kubernetes 集群             │
└─────────────────────────────────────────────────────────┘
```

## 项目结构

```
ai-sre/
├── tools/mcp/                          # MCP 服务器主模块
│   ├── cmd/mcp-server/main.go          # 服务入口
│   ├── internal/
│   │   ├── auth/middleware.go           # 认证中间件 (Bearer/Basic/APIKey)
│   │   ├── config/config.go            # 配置管理 (环境变量 + 命令行参数)
│   │   ├── server/server.go            # HTTP 服务器 & 路由 & 管理端点
│   │   ├── tools/
│   │   │   ├── manager.go              # 工具注册管理器
│   │   │   ├── registry.go             # 全局工具注册表
│   │   │   ├── ping.go                 # ping 工具
│   │   │   ├── echo.go                 # echo 工具
│   │   │   ├── system_info.go          # 系统信息工具
│   │   │   ├── tencentcloud_handlers.go # 腾讯云工具处理函数
│   │   │   └── tencentcloud_tools.go   # 腾讯云工具业务逻辑
│   │   ├── transport/
│   │   │   ├── http.go                 # HTTP 传输实现
│   │   │   └── mcp_handler.go          # MCP JSON-RPC 处理器 & Schema
│   │   └── tencentcloud/               # 腾讯云 SDK 封装
│   │       ├── client.go               # 客户端管理器
│   │       ├── config.go               # 凭证配置
│   │       ├── region/client.go        # 地域服务
│   │       ├── tke/client.go           # TKE 容器服务
│   │       ├── cvm/client.go           # CVM 云服务器
│   │       ├── clb/client.go           # CLB 负载均衡
│   │       ├── cdb/client.go           # CDB 云数据库
│   │       └── vpc/client.go           # VPC 私有网络
│   ├── pkg/logger/                     # 日志工具包
│   ├── Dockerfile                      # Docker 镜像构建
│   ├── docker-compose.yml              # Docker Compose 部署
│   └── deploy.sh                       # 部署脚本
├── pkg/kubernetes/                     # 公共 Kubernetes 客户端库 (独立模块)
│   ├── client.go                       # K8s 客户端 (kubeconfig/in-cluster)
│   ├── namespaces.go                   # 命名空间操作
│   ├── workloads.go                    # 工作负载操作 (6 种类型)
│   └── pods.go                         # Pod 状态查询
├── configs/                            # 配置文件模板
├── specs/                              # API 规范定义 (OpenAPI/Protobuf)
├── docs/                               # 项目文档
├── scripts/                            # 运维脚本
└── Makefile                            # 统一构建命令
```

## 快速开始

### 环境要求

- Go 1.24+
- Docker (可选，用于容器化部署)
- 腾讯云 SecretID / SecretKey (使用云产品工具时必需)

### 本地编译运行

```bash
# 克隆项目
git clone https://github.com/goops-top/ai-sre.git
cd ai-sre

# 编译 MCP Server
cd tools/mcp
go build -o bin/mcp-server ./cmd/mcp-server

# stdio 模式运行 (默认)
./bin/mcp-server

# HTTP 模式运行
export TENCENTCLOUD_SECRET_ID="your_secret_id"
export TENCENTCLOUD_SECRET_KEY="your_secret_key"
./bin/mcp-server -transport http -port 8080
```

### Docker 部署

```bash
cd tools/mcp

# 构建镜像
docker build -t ai-sre-mcp-server:latest .

# 运行
docker run -d \
  --name ai-sre-mcp-server \
  -p 8080:8080 \
  -e TENCENTCLOUD_SECRET_ID="your_secret_id" \
  -e TENCENTCLOUD_SECRET_KEY="your_secret_key" \
  -e TENCENTCLOUD_REGION="ap-beijing" \
  ai-sre-mcp-server:latest
```

### Docker Compose 部署

```bash
cd tools/mcp

# 编辑 docker-compose.yml 填入腾讯云凭证
docker-compose up -d
```

### 推送到腾讯云镜像仓库

```bash
# 在项目根目录
make mcp
```

## MCP 工具清单

### 内置工具 (3 个)

| 工具 | 说明 |
|------|------|
| `ping` | 连接测试，返回 pong 响应 |
| `echo` | 文本处理，支持大小写转换、前缀后缀、重复 |
| `system_info` | 系统运行时信息 (Go 运行时、内存、环境) |

### 腾讯云通用工具 (3 个)

| 工具 | 说明 |
|------|------|
| `describe_regions` | 查询产品支持的地域列表 (支持 tke/cvm/cos/clb/vpc/cdb) |
| `get_region` | 根据地域 ID 查询详细信息 |
| `tencentcloud_validate` | 验证腾讯云 API 连接和权限 |

### TKE 容器服务工具 (11 个)

| 工具 | 说明 |
|------|------|
| `tke_describe_clusters` | 查询集群列表，支持按类型过滤 (普通/弹性) |
| `tke_describe_cluster_extra_args` | 查询集群自定义参数 |
| `tke_get_cluster_level_price` | 获取集群等级价格信息 |
| `tke_describe_addon` | 查询集群已安装的 addon 列表 |
| `tke_get_app_chart_list` | 获取可安装的 addon 列表 |
| `tke_describe_images` | 获取支持的节点 OS 镜像 |
| `tke_describe_versions` | 获取支持的 K8s 版本 |
| `tke_describe_log_switches` | 查询集群日志开关状态 |
| `tke_describe_master_component` | 查询 Master 组件运行状态 |
| `tke_describe_cluster_instances` | 查询集群节点列表 |
| `tke_describe_cluster_virtual_node` | 查询超级节点列表 |

### CVM 云服务器工具 (2 个)

| 工具 | 说明 |
|------|------|
| `cvm_describe_instances` | 查询实例列表，支持按 ID/名称/可用区过滤 |
| `cvm_describe_instances_status` | 查询实例运行状态 |

### CLB 负载均衡工具 (4 个)

| 工具 | 说明 |
|------|------|
| `clb_describe_load_balancers` | 查询 CLB 实例列表 |
| `clb_describe_listeners` | 查询监听器列表 |
| `clb_describe_targets` | 查询后端目标 (RS) 列表 |
| `clb_describe_target_health` | 查询后端目标健康状态 |

### CDB 云数据库工具 (4 个)

| 工具 | 说明 |
|------|------|
| `cdb_describe_db_instances` | 查询 MySQL 实例列表 |
| `cdb_describe_db_instance_info` | 查询实例详细信息 |
| `cdb_describe_slow_logs` | 查询慢查询日志文件 |
| `cdb_describe_error_log` | 查询错误日志 (支持时间范围和关键字) |

### VPC 私有网络工具 (9 个)

| 工具 | 说明 |
|------|------|
| `vpc_describe_vpcs` | 查询 VPC 列表 |
| `vpc_describe_subnets` | 查询子网列表 |
| `vpc_describe_security_groups` | 查询安全组列表 |
| `vpc_describe_network_interfaces` | 查询弹性网卡列表 |
| `vpc_describe_addresses` | 查询弹性公网 IP 列表 |
| `vpc_describe_bandwidth_packages` | 查询带宽包列表 |
| `vpc_describe_vpc_endpoint` | 查询终端节点列表 |
| `vpc_describe_vpc_endpoint_service` | 查询终端节点服务列表 |
| `vpc_describe_vpc_peering_connections` | 查询对等连接列表 |

## 公共 Kubernetes 库

`pkg/kubernetes/` 是一个独立的 Go 模块，提供 Kubernetes 集群操作的原子化公共函数，可被 MCP 工具和 Agent 等上层模块引用。

### 客户端初始化

```go
import "ai-sre/pkg/kubernetes"

// 支持 kubeconfig 文件 / KUBECONFIG 环境变量 / ~/.kube/config / in-cluster 四种模式
client, err := kubernetes.NewClient("/path/to/kubeconfig", logger)
```

### 核心 API

| 方法 | 说明 |
|------|------|
| `ListNamespaces(ctx, opts)` | 获取命名空间列表，支持标签选择器 |
| `ListWorkloads(ctx, namespace, opts)` | 获取工作负载列表 (Deployment/StatefulSet/DaemonSet/CronJob/Job/StatefulSetPlus) |
| `GetWorkloadDetail(ctx, namespace, name, workloadType, opts)` | 获取工作负载详情 (容器规格、资源、探针、条件等) |
| `ListPodStatus(ctx, namespace, opts)` | 获取 Pod 状态列表，支持只返回异常 Pod |

> StatefulSetPlus 通过 dynamic client 查询 CRD，兼容 `apps.kruise.io` 和 `platform.tkestack.io` 两种 GVR。

## 传输模式

| 模式 | 启动参数 | 说明 |
|------|---------|------|
| **stdio** | `-transport stdio` (默认) | 通过 stdin/stdout 通信，适合 CLI 工具集成 |
| **HTTP** | `-transport http -port 8080` | JSON-RPC over HTTP，适合服务端部署 |
| **SSE** | `-transport sse -port 8080` | Server-Sent Events 流式通信 |

### HTTP 模式端点

| 端点 | 方法 | 说明 |
|------|------|------|
| `/mcp` | POST | MCP JSON-RPC 通信入口 |
| `/mcp` | GET | SSE 流 (SSE 模式) |
| `/health` | GET | 健康检查 |
| `/status` | GET | 服务状态 |
| `/` | GET | 管理主页 |
| `/mcp/manage/info` | GET | MCP 服务信息 |
| `/mcp/manage/tools` | GET | 工具列表 |

## 认证配置

通过环境变量配置，支持三种方式 (可同时启用)：

| 认证方式 | 环境变量 |
|---------|---------|
| Bearer Token | `MCP_AUTH_BEARER_TOKEN` |
| Basic Auth | `MCP_AUTH_USERNAME` + `MCP_AUTH_PASSWORD` |
| API Key | `MCP_AUTH_API_KEY` |
| IP 白名单 | `MCP_AUTH_ALLOWED_IPS` |

启动时通过 `-enable-auth` 参数或 `MCP_ENABLE_AUTH=true` 环境变量开启认证。

## 环境变量

### MCP Server 配置

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `MCP_TRANSPORT` | `stdio` | 传输模式 (stdio/http/sse) |
| `MCP_PORT` | `8080` | HTTP/SSE 端口 |
| `MCP_HOST` | `0.0.0.0` | 监听地址 |
| `MCP_LOG_LEVEL` | `info` | 日志级别 |
| `MCP_LOG_FORMAT` | `json` | 日志格式 (json/text) |
| `MCP_ENABLE_AUTH` | `false` | 是否启用认证 |

### 腾讯云配置

| 变量 | 必需 | 说明 |
|------|------|------|
| `TENCENTCLOUD_SECRET_ID` | 是 | 腾讯云 SecretID |
| `TENCENTCLOUD_SECRET_KEY` | 是 | 腾讯云 SecretKey |
| `TENCENTCLOUD_REGION` | 否 | 地域，默认 `ap-beijing` |
| `TENCENTCLOUD_ENDPOINT` | 否 | 自定义 API 端点 (优先级最高) |
| `TENCENTCLOUD_USE_INTERNAL` | 否 | 是否使用内网域名访问云 API，设置为 `true` 启用 |

> **内网访问说明**：当 MCP Server 部署在腾讯云 CVM/TKE 等云上环境时，可设置 `TENCENTCLOUD_USE_INTERNAL=true` 通过内网域名 (`{product}.internal.tencentcloudapi.com`) 调用云 API，无需公网带宽，延迟更低。域名解析优先级：`TENCENTCLOUD_ENDPOINT` > `TENCENTCLOUD_USE_INTERNAL` > SDK 默认公网域名。

## 构建命令

```bash
# 编译
make build-go

# 构建 Docker 镜像并推送到腾讯云
make mcp

# 运行测试
make test-go

# 代码检查
make lint-go

# 部署到 staging
make deploy-staging

# 部署到 production
make deploy-prod
```

## 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情
