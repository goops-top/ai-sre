# 腾讯云 MCP 工具使用指南

本文档介绍如何使用 AI SRE MCP 服务器中的腾讯云工具集。

## 概述

腾讯云工具集为 MCP 服务器提供了与腾讯云服务交互的能力，目前支持以下产品：

- **TKE (腾讯云容器服务)**: 查询地域信息、集群管理等
- 更多产品支持正在开发中...

## 配置要求

### 环境变量配置

在使用腾讯云工具之前，需要设置以下环境变量：

```bash
# 必需配置
export TENCENTCLOUD_SECRET_ID="your-secret-id"
export TENCENTCLOUD_SECRET_KEY="your-secret-key"

# 可选配置
export TENCENTCLOUD_REGION="ap-beijing"          # 默认地域，默认为 ap-beijing
export TENCENTCLOUD_ENDPOINT=""                  # 自定义端点，通常不需要设置
```

### 获取腾讯云密钥

1. 登录 [腾讯云控制台](https://console.cloud.tencent.com/)
2. 访问 [访问管理 - API密钥管理](https://console.cloud.tencent.com/cam/capi)
3. 创建新的 API 密钥或使用现有密钥
4. 记录 SecretId 和 SecretKey

### 权限要求

确保您的腾讯云账号具有以下权限：

- **TKE 相关权限**:
  - `tke:DescribeRegions` - 查询地域信息

## 可用工具

### 1. tke_describe_regions

查询腾讯云 TKE 支持的所有地域信息。

**参数**:
- `format` (可选): 输出格式
  - `table` (默认): 表格格式
  - `json`: JSON 格式

**示例调用**:
```json
{
  "name": "tke_describe_regions",
  "arguments": {
    "format": "table"
  }
}
```

**输出示例** (表格格式):
```
TKE 支持的地域信息:
┌─────────────────┬──────────────────────────┬──────────────┐
│ 地域ID          │ 地域名称                 │ 状态         │
├─────────────────┼──────────────────────────┼──────────────┤
│ 1               │ 华南地区(广州)           │ alluser      │
│ 4               │ 华东地区(上海)           │ alluser      │
│ 8               │ 华北地区(北京)           │ alluser      │
└─────────────────┴──────────────────────────┴──────────────┘
总计: 3 个地域
```

### 2. tke_get_region

根据地域ID或地域名称查询特定地域的详细信息。

**参数**:
- `region_id` (必需): 地域ID或地域名称
- `format` (可选): 输出格式
  - `table` (默认): 表格格式
  - `json`: JSON 格式

**示例调用**:
```json
{
  "name": "tke_get_region",
  "arguments": {
    "region_id": "1",
    "format": "json"
  }
}
```

**输出示例** (JSON 格式):
```json
{
  "region_id": 1,
  "region_name": "华南地区(广州)",
  "status": "alluser"
}
```

### 3. tencentcloud_validate

验证腾讯云 API 连接和权限配置。

**参数**: 无

**示例调用**:
```json
{
  "name": "tencentcloud_validate",
  "arguments": {}
}
```

**输出示例**:
```json
{
  "status": "success",
  "message": "腾讯云连接验证成功",
  "services": ["TKE"]
}
```

## 使用示例

### 启动服务器

```bash
# 设置环境变量
export TENCENTCLOUD_SECRET_ID="your-secret-id"
export TENCENTCLOUD_SECRET_KEY="your-secret-key"

# 启动 HTTP 模式服务器
go run cmd/mcp-server/main.go -transport http -port 8080

# 或启动 stdio 模式服务器
go run cmd/mcp-server/main.go -transport stdio
```

### 使用客户端测试

```bash
# 测试 HTTP 模式
go run cmd/client/main.go http

# 测试 stdio 模式
go run cmd/client/main.go stdio
```

### 在 AI Chat 中配置

**HTTP 模式配置**:
- URL: `http://localhost:8080/mcp`
- 类型: HTTP

**stdio 模式配置**:
- 命令: `go run cmd/mcp-server/main.go -transport stdio`
- 工作目录: `/path/to/ai-sre/tools/mcp`

## 故障排除

### 常见错误

1. **"腾讯云工具未初始化，请检查配置"**
   - 检查环境变量 `TENCENTCLOUD_SECRET_ID` 和 `TENCENTCLOUD_SECRET_KEY` 是否正确设置
   - 确保密钥有效且未过期

2. **"TKE API 错误 [AuthFailure]: ..."**
   - 检查 SecretId 和 SecretKey 是否正确
   - 确保账号具有相应的 TKE 权限

3. **"TKE 权限验证失败"**
   - 检查账号是否具有 `tke:DescribeRegions` 权限
   - 联系管理员添加相应权限

### 调试模式

启用调试日志以获取更详细的错误信息：

```bash
MCP_LOG_LEVEL=debug go run cmd/mcp-server/main.go -transport http
```

## 扩展开发

### 添加新的腾讯云产品支持

1. 在 `internal/tencentcloud/` 目录下创建新的产品客户端
2. 实现 `ProductClient` 接口
3. 在 `TencentCloudTools` 中添加新的客户端
4. 创建相应的工具处理函数
5. 在 `RegisterTencentCloudTools` 中注册新工具

### 目录结构

```
internal/tencentcloud/
├── client.go          # 通用客户端管理器
├── config.go          # 配置管理
├── tke/              # TKE 产品客户端
│   └── client.go
└── [product]/        # 其他产品客户端
    └── client.go
```

## 版本历史

- **v1.0.0**: 初始版本，支持 TKE 地域查询
- 更多功能正在开发中...

## 支持与反馈

如有问题或建议，请通过以下方式联系：

- 创建 Issue
- 提交 Pull Request
- 联系开发团队

---

**注意**: 请妥善保管您的腾讯云密钥，不要在代码中硬编码或提交到版本控制系统中。