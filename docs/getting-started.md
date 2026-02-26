# 快速开始指南

本指南将帮助你快速搭建和运行AI SRE分身助理系统。

## 前置条件

### 系统要求
- **操作系统**: Linux/macOS/Windows
- **内存**: 最少8GB，推荐16GB+
- **存储**: 最少20GB可用空间
- **网络**: 稳定的互联网连接

### 软件依赖
- **Python**: 3.9或更高版本 (Agent编排层)
- **Go**: 1.21或更高版本 (MCP工具层)
- **Docker**: 20.10或更高版本
- **Docker Compose**: 2.0或更高版本
- **Node.js**: 16或更高版本 (前端开发)
- **Git**: 用于代码管理

## 快速部署

### 1. 获取代码

```bash
# 克隆项目
git clone https://github.com/your-org/ai-sre.git
cd ai-sre

# 检查项目结构
ls -la
```

### 2. 环境配置

```bash
# 复制配置文件模板
cp configs/config.example.yaml configs/config.yaml

# 复制环境变量文件
cp .env.example .env
```

编辑 `.env` 文件，填入必要的配置：

```bash
# 数据库密码
DB_PASSWORD=your_secure_password

# Redis密码
REDIS_PASSWORD=your_redis_password

# InfluxDB配置
INFLUXDB_PASSWORD=your_influxdb_password
INFLUXDB_TOKEN=your_influxdb_token

# AI服务配置
OPENAI_API_KEY=your_openai_api_key
CLAUDE_API_KEY=your_claude_api_key

# 企业微信配置
WECHAT_CORP_ID=your_corp_id
WECHAT_AGENT_ID=your_agent_id
WECHAT_SECRET=your_secret

# Grafana配置
GRAFANA_API_KEY=your_grafana_api_key
```

### 3. 一键启动

```bash
# 启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f ai-sre-api
```

### 4. 验证部署

访问以下地址验证服务是否正常运行：

- **API文档**: http://localhost:8000/docs
- **Web控制台**: http://localhost:3000
- **Grafana监控**: http://localhost:3001 (admin/admin)
- **健康检查**: http://localhost:8000/health

## 开发环境搭建

### 1. Python环境 (Agent编排层)

```bash
# 创建虚拟环境
python -m venv venv

# 激活虚拟环境
# Linux/macOS:
source venv/bin/activate
# Windows:
venv\Scripts\activate

# 安装依赖
pip install -r requirements.txt
pip install -r requirements-dev.txt
```

### 2. Go环境 (MCP工具层)

```bash
# 进入MCP工具目录
cd tools/mcp

# 下载依赖
go mod download

# 构建所有工具服务
make build-all

# 或单独构建
go build -o bin/monitoring ./cmd/monitoring
go build -o bin/cloud ./cmd/cloud
go build -o bin/container ./cmd/container
go build -o bin/database ./cmd/database
```

### 3. 前端环境

```bash
# 进入前端目录
cd src/interfaces/web/frontend

# 安装依赖
npm install

# 启动开发服务器
npm run dev
```

### 3. 数据库初始化

```bash
# 启动数据库服务
docker-compose up -d postgres redis influxdb

# 运行数据库迁移
python -m alembic upgrade head

# 初始化基础数据
python scripts/init_data.py
```

### 4. 启动开发服务

**Python Agent服务**:
```bash
# 启动API服务
uvicorn src.main:app --reload --host 0.0.0.0 --port 8000

# 启动Celery Worker (新终端)
celery -A src.core.celery worker --loglevel=info

# 启动Celery Beat (新终端)
celery -A src.core.celery beat --loglevel=info
```

**Go MCP工具服务**:
```bash
# 启动监控工具服务 (新终端)
cd tools/mcp && go run cmd/monitoring/main.go

# 启动云平台工具服务 (新终端)
cd tools/mcp && go run cmd/cloud/main.go

# 启动容器工具服务 (新终端)
cd tools/mcp && go run cmd/container/main.go

# 启动数据库工具服务 (新终端)
cd tools/mcp && go run cmd/database/main.go
```

**验证服务**:
```bash
# 检查Python API服务
curl http://localhost:8000/health

# 检查MCP工具服务
curl http://localhost:8081/health  # 监控工具
curl http://localhost:8082/health  # 云平台工具
curl http://localhost:8083/health  # 容器工具
curl http://localhost:8084/health  # 数据库工具
```

## 配置说明

### 核心配置文件

#### `configs/config.yaml`
```yaml
# 应用基础配置
app:
  name: "AI SRE Assistant"
  version: "1.0.0"
  debug: true
  log_level: "DEBUG"

# 数据库配置
database:
  postgresql:
    host: "localhost"
    port: 5432
    database: "ai_sre"
    username: "ai_sre_user"
    password: "${DB_PASSWORD}"
    
# Agent配置
agents:
  master:
    enabled: true
    max_concurrent_tasks: 10
    timeout: 300
    
  monitoring:
    enabled: true
    check_interval: 30
    alert_threshold: 0.8
    
  diagnosis:
    enabled: true
    max_diagnosis_time: 600
    confidence_threshold: 0.7
    
  automation:
    enabled: true
    dry_run: false
    approval_required: true

# 外部服务集成
integrations:
  openai:
    api_key: "${OPENAI_API_KEY}"
    model: "gpt-4"
    max_tokens: 2000
    
  prometheus:
    url: "http://localhost:9090"
    timeout: 30
    
  grafana:
    url: "http://localhost:3000"
    api_key: "${GRAFANA_API_KEY}"
```

### 环境变量说明

| 变量名 | 说明 | 示例值 |
|--------|------|--------|
| `DB_PASSWORD` | PostgreSQL数据库密码 | `secure_password_123` |
| `REDIS_PASSWORD` | Redis密码 | `redis_password_456` |
| `OPENAI_API_KEY` | OpenAI API密钥 | `sk-...` |
| `WECHAT_CORP_ID` | 企业微信企业ID | `ww123456789` |
| `WECHAT_AGENT_ID` | 企业微信应用ID | `1000001` |
| `WECHAT_SECRET` | 企业微信应用密钥 | `secret_key_789` |

## 功能验证

### 1. API接口测试

```bash
# 健康检查
curl http://localhost:8000/health

# 获取Agent列表
curl http://localhost:8000/api/v1/agents

# 创建测试任务
curl -X POST http://localhost:8000/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "type": "monitoring",
    "description": "检查系统状态",
    "priority": "normal"
  }'
```

### 2. 企业微信集成测试

```bash
# 测试企业微信消息发送
python scripts/test_wechat.py
```

### 3. 监控数据测试

```bash
# 生成测试监控数据
python scripts/generate_test_data.py

# 查看Grafana仪表板
# 访问 http://localhost:3001
# 用户名: admin, 密码: admin
```

## 常见问题

### 1. 服务启动失败

**问题**: Docker容器启动失败
```bash
# 检查日志
docker-compose logs ai-sre-api

# 常见原因：
# - 端口被占用
# - 环境变量未设置
# - 数据库连接失败
```

**解决方案**:
```bash
# 检查端口占用
netstat -tulpn | grep :8000

# 重新生成配置
cp configs/config.example.yaml configs/config.yaml

# 重启服务
docker-compose down
docker-compose up -d
```

### 2. 数据库连接问题

**问题**: 无法连接到PostgreSQL
```bash
# 检查数据库状态
docker-compose ps postgres

# 检查数据库日志
docker-compose logs postgres
```

**解决方案**:
```bash
# 重置数据库
docker-compose down -v
docker-compose up -d postgres

# 等待数据库启动完成
sleep 30

# 运行数据库迁移
python -m alembic upgrade head
```

### 3. AI服务配置问题

**问题**: AI API调用失败
```bash
# 检查API密钥配置
echo $OPENAI_API_KEY

# 测试API连接
python scripts/test_ai_connection.py
```

**解决方案**:
```bash
# 更新API密钥
export OPENAI_API_KEY="your_new_api_key"

# 重启服务
docker-compose restart ai-sre-api
```

## 下一步

完成基础部署后，你可以：

1. **配置监控源**: 连接你的Prometheus、Grafana等监控系统
2. **添加知识库**: 导入你的运维文档和最佳实践
3. **配置告警规则**: 设置适合你环境的告警策略
4. **训练专项Agent**: 基于你的业务场景训练专门的Agent
5. **集成企业系统**: 连接你的ITSM、CMDB等企业系统

详细的配置和使用指南请参考：
- [架构文档](architecture.md)
- [组件说明](components.md)
- [API文档](api.md)
- [部署指南](deployment.md)