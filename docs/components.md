# 组件详细说明

## 用户入口层组件

### 1. 企业微信机器人 (WeChat Bot)

**功能描述**: 
- 接收企业微信消息和指令
- 推送告警和通知信息
- 支持交互式对话和指令执行

**技术实现**:
```python
# 企业微信API集成示例
class WeChatBot:
    def __init__(self, corp_id, agent_id, secret):
        self.corp_id = corp_id
        self.agent_id = agent_id 
        self.secret = secret
        
    async def send_message(self, user_id, message):
        # 发送消息到企业微信
        pass
        
    async def handle_callback(self, request):
        # 处理企业微信回调
        pass
```

**配置文件**:
```yaml
wechat:
  corp_id: "your_corp_id"
  agent_id: "your_agent_id"
  secret: "your_secret"
  callback_url: "https://your-domain.com/wechat/callback"
```

### 2. Web控制台 (Web Console)

**功能描述**:
- 可视化运维仪表板
- Agent状态监控和管理
- 任务执行历史查看
- 系统配置管理

**技术栈**:
- 前端: React + TypeScript + Ant Design
- 状态管理: Redux Toolkit
- 图表: ECharts
- 实时通信: WebSocket

**目录结构**:
```
src/interfaces/web/
├── frontend/
│   ├── src/
│   │   ├── components/     # 通用组件
│   │   ├── pages/         # 页面组件
│   │   ├── services/      # API服务
│   │   ├── store/         # 状态管理
│   │   └── utils/         # 工具函数
│   ├── public/
│   └── package.json
└── backend/
    ├── api/               # API路由
    ├── middleware/        # 中间件
    └── models/           # 数据模型
```

### 3. API网关 (API Gateway)

**功能描述**:
- RESTful API接口提供
- 请求路由和负载均衡
- 身份认证和权限控制
- 请求限流和监控

**API设计**:
```yaml
# OpenAPI 3.0 规范
paths:
  /api/v1/agents:
    get:
      summary: 获取Agent列表
      responses:
        200:
          description: 成功返回Agent列表
          
  /api/v1/tasks:
    post:
      summary: 创建新任务
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/TaskRequest'
```

## Agent编排层组件

### 1. 主控Agent (Master Agent)

**职责**:
- 任务接收和理解
- 任务分解和分发
- 结果汇总和响应
- Agent状态管理

**核心算法**:
```python
class MasterAgent(BaseAgent):
    async def process_task(self, task):
        # 1. 任务理解和分类
        task_type = await self.classify_task(task)
        
        # 2. 选择合适的Agent
        target_agents = await self.select_agents(task_type)
        
        # 3. 任务分发
        results = await self.distribute_task(task, target_agents)
        
        # 4. 结果汇总
        final_result = await self.aggregate_results(results)
        
        return final_result
```

### 2. 监控Agent (Monitoring Agent)

**职责**:
- 系统指标收集和分析
- 异常检测和预警
- 性能趋势分析
- 监控规则管理

**监控指标**:
```python
MONITORING_METRICS = {
    'system': ['cpu_usage', 'memory_usage', 'disk_usage', 'network_io'],
    'application': ['response_time', 'error_rate', 'throughput', 'availability'],
    'business': ['user_count', 'transaction_volume', 'revenue', 'conversion_rate']
}
```

### 3. 诊断Agent (Diagnosis Agent)

**职责**:
- 故障检测和分析
- 根因分析
- 解决方案推荐
- 诊断报告生成

**诊断流程**:
```python
class DiagnosisAgent(BaseAgent):
    async def diagnose(self, symptoms):
        # 1. 症状分析
        symptom_analysis = await self.analyze_symptoms(symptoms)
        
        # 2. 数据收集
        diagnostic_data = await self.collect_diagnostic_data(symptom_analysis)
        
        # 3. 根因分析
        root_causes = await self.analyze_root_causes(diagnostic_data)
        
        # 4. 解决方案推荐
        solutions = await self.recommend_solutions(root_causes)
        
        return DiagnosisReport(symptoms, root_causes, solutions)
```

### 4. 自动化Agent (Automation Agent)

**职责**:
- 自动化脚本执行
- 批量操作处理
- 工作流编排
- 执行结果验证

**自动化任务类型**:
```python
AUTOMATION_TASKS = {
    'deployment': ['deploy_application', 'rollback_deployment', 'scale_service'],
    'maintenance': ['restart_service', 'clear_cache', 'backup_database'],
    'security': ['update_certificates', 'rotate_keys', 'patch_vulnerabilities'],
    'optimization': ['tune_parameters', 'clean_logs', 'optimize_queries']
}
```

## 工具服务层组件

### 1. MCP工具集

> **技术栈**: 基于 Golang + Gin 框架构建的高性能微服务工具集

#### MCP工具架构设计

**核心特性**:
- 高性能并发处理
- 标准化MCP协议接口
- 插件化架构设计
- 统一配置管理
- 完善的错误处理和日志记录

**项目结构**:
```
tools/mcp/
├── cmd/                    # 各工具服务入口
│   ├── monitoring/         # 监控工具服务
│   ├── cloud/             # 云平台工具服务
│   ├── container/         # 容器工具服务
│   └── database/          # 数据库工具服务
├── internal/              # 内部包
│   ├── config/           # 配置管理
│   ├── middleware/       # 中间件
│   ├── models/          # 数据模型
│   └── utils/           # 工具函数
├── pkg/                  # 公共包
│   ├── mcp/             # MCP协议实现
│   ├── client/          # 客户端SDK
│   └── logger/          # 日志组件
└── api/                 # API定义
    └── proto/           # gRPC协议定义
```

#### 监控工具 (Monitoring Tools)

**Prometheus客户端**:
```go
package monitoring

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
    
    "github.com/gin-gonic/gin"
    "github.com/prometheus/client_golang/api"
    v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

type PrometheusClient struct {
    client api.Client
    api    v1.API
}

func NewPrometheusClient(url string) (*PrometheusClient, error) {
    client, err := api.NewClient(api.Config{
        Address: url,
    })
    if err != nil {
        return nil, err
    }
    
    return &PrometheusClient{
        client: client,
        api:    v1.NewAPI(client),
    }, nil
}

func (p *PrometheusClient) Query(ctx context.Context, query string) (interface{}, error) {
    result, warnings, err := p.api.Query(ctx, query, time.Now())
    if err != nil {
        return nil, fmt.Errorf("prometheus query failed: %w", err)
    }
    
    if len(warnings) > 0 {
        // 记录警告日志
        for _, warning := range warnings {
            fmt.Printf("Warning: %s\n", warning)
        }
    }
    
    return result, nil
}

func (p *PrometheusClient) QueryRange(ctx context.Context, query string, start, end time.Time, step time.Duration) (interface{}, error) {
    result, warnings, err := p.api.QueryRange(ctx, query, v1.Range{
        Start: start,
        End:   end,
        Step:  step,
    })
    
    if err != nil {
        return nil, fmt.Errorf("prometheus range query failed: %w", err)
    }
    
    if len(warnings) > 0 {
        for _, warning := range warnings {
            fmt.Printf("Warning: %s\n", warning)
        }
    }
    
    return result, nil
}

// Gin路由处理器
func (p *PrometheusClient) SetupRoutes(r *gin.Engine) {
    api := r.Group("/api/v1/monitoring")
    {
        api.POST("/query", p.handleQuery)
        api.POST("/query_range", p.handleQueryRange)
    }
}

func (p *PrometheusClient) handleQuery(c *gin.Context) {
    var req struct {
        Query string `json:"query" binding:"required"`
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    result, err := p.Query(c.Request.Context(), req.Query)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{"data": result})
}
```

**Grafana客户端**:
```go
package monitoring

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    
    "github.com/gin-gonic/gin"
)

type GrafanaClient struct {
    baseURL string
    apiKey  string
    client  *http.Client
}

func NewGrafanaClient(baseURL, apiKey string) *GrafanaClient {
    return &GrafanaClient{
        baseURL: baseURL,
        apiKey:  apiKey,
        client:  &http.Client{Timeout: 30 * time.Second},
    }
}

type Dashboard struct {
    ID          int    `json:"id,omitempty"`
    Title       string `json:"title"`
    Description string `json:"description"`
    Tags        []string `json:"tags"`
    Panels      []Panel  `json:"panels"`
}

type Panel struct {
    ID          int    `json:"id"`
    Title       string `json:"title"`
    Type        string `json:"type"`
    GridPos     GridPos `json:"gridPos"`
    Targets     []Target `json:"targets"`
}

func (g *GrafanaClient) CreateDashboard(ctx context.Context, dashboard Dashboard) (*Dashboard, error) {
    payload := map[string]interface{}{
        "dashboard": dashboard,
        "overwrite": true,
    }
    
    jsonData, err := json.Marshal(payload)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal dashboard: %w", err)
    }
    
    req, err := http.NewRequestWithContext(ctx, "POST", 
        fmt.Sprintf("%s/api/dashboards/db", g.baseURL), 
        bytes.NewBuffer(jsonData))
    if err != nil {
        return nil, err
    }
    
    req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", g.apiKey))
    req.Header.Set("Content-Type", "application/json")
    
    resp, err := g.client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("grafana API returned status %d", resp.StatusCode)
    }
    
    var result struct {
        Dashboard Dashboard `json:"dashboard"`
        Status    string    `json:"status"`
    }
    
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }
    
    return &result.Dashboard, nil
}
```

#### 云平台工具 (Cloud Tools)

**AWS客户端**:
```go
package cloud

import (
    "context"
    "net/http"
    
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/ec2"
    "github.com/gin-gonic/gin"
)

type AWSClient struct {
    ec2Client *ec2.Client
}

func NewAWSClient(ctx context.Context, region string) (*AWSClient, error) {
    cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
    if err != nil {
        return nil, err
    }
    
    return &AWSClient{
        ec2Client: ec2.NewFromConfig(cfg),
    }, nil
}

func (a *AWSClient) ListInstances(ctx context.Context) ([]Instance, error) {
    result, err := a.ec2Client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{})
    if err != nil {
        return nil, err
    }
    
    var instances []Instance
    for _, reservation := range result.Reservations {
        for _, instance := range reservation.Instances {
            instances = append(instances, Instance{
                ID:       *instance.InstanceId,
                State:    string(instance.State.Name),
                Type:     string(instance.InstanceType),
                LaunchTime: *instance.LaunchTime,
            })
        }
    }
    
    return instances, nil
}

func (a *AWSClient) StartInstance(ctx context.Context, instanceID string) error {
    _, err := a.ec2Client.StartInstances(ctx, &ec2.StartInstancesInput{
        InstanceIds: []string{instanceID},
    })
    return err
}

// Gin路由设置
func (a *AWSClient) SetupRoutes(r *gin.Engine) {
    api := r.Group("/api/v1/aws")
    {
        api.GET("/instances", a.handleListInstances)
        api.POST("/instances/:id/start", a.handleStartInstance)
        api.POST("/instances/:id/stop", a.handleStopInstance)
    }
}
```

**Kubernetes客户端**:
```go
package container

import (
    "context"
    "fmt"
    "net/http"
    "strconv"
    
    "github.com/gin-gonic/gin"
    appsv1 "k8s.io/api/apps/v1"
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/tools/clientcmd"
)

type KubernetesClient struct {
    clientset *kubernetes.Clientset
}

func NewKubernetesClient(kubeconfig string) (*KubernetesClient, error) {
    config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
    if err != nil {
        return nil, err
    }
    
    clientset, err := kubernetes.NewForConfig(config)
    if err != nil {
        return nil, err
    }
    
    return &KubernetesClient{
        clientset: clientset,
    }, nil
}

func (k *KubernetesClient) ListPods(ctx context.Context, namespace string) ([]corev1.Pod, error) {
    pods, err := k.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
    if err != nil {
        return nil, err
    }
    
    return pods.Items, nil
}

func (k *KubernetesClient) ScaleDeployment(ctx context.Context, name, namespace string, replicas int32) error {
    deployment, err := k.clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
    if err != nil {
        return err
    }
    
    deployment.Spec.Replicas = &replicas
    
    _, err = k.clientset.AppsV1().Deployments(namespace).Update(ctx, deployment, metav1.UpdateOptions{})
    return err
}

// Gin路由设置
func (k *KubernetesClient) SetupRoutes(r *gin.Engine) {
    api := r.Group("/api/v1/k8s")
    {
        api.GET("/namespaces/:namespace/pods", k.handleListPods)
        api.POST("/namespaces/:namespace/deployments/:name/scale", k.handleScaleDeployment)
    }
}

func (k *KubernetesClient) handleScaleDeployment(c *gin.Context) {
    namespace := c.Param("namespace")
    name := c.Param("name")
    
    var req struct {
        Replicas int32 `json:"replicas" binding:"required"`
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    if err := k.ScaleDeployment(c.Request.Context(), name, namespace, req.Replicas); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{"message": "Deployment scaled successfully"})
}
```

#### MCP工具服务配置

**主配置文件** (`tools/mcp/configs/config.yaml`):
```yaml
server:
  host: "0.0.0.0"
  port: 8080
  mode: "release"  # debug, release, test
  
mcp:
  protocol_version: "1.0"
  timeout: 30s
  max_concurrent_requests: 100
  
monitoring:
  prometheus:
    url: "http://prometheus:9090"
    timeout: 30s
  grafana:
    url: "http://grafana:3000"
    api_key: "${GRAFANA_API_KEY}"
    
cloud:
  aws:
    region: "us-west-2"
    access_key_id: "${AWS_ACCESS_KEY_ID}"
    secret_access_key: "${AWS_SECRET_ACCESS_KEY}"
  
kubernetes:
    kubeconfig: "/etc/kubeconfig/config"
    timeout: 60s
    
database:
  postgresql:
    host: "postgres"
    port: 5432
    database: "monitoring"
    username: "${DB_USERNAME}"
    password: "${DB_PASSWORD}"
    
logging:
  level: "info"
  format: "json"
  output: "stdout"
```

**Docker配置** (`tools/mcp/Dockerfile`):
```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/monitoring

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/main .
COPY --from=builder /app/configs ./configs

EXPOSE 8080
CMD ["./main"]
```

**服务启动示例**:
```go
package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"
    
    "github.com/gin-gonic/gin"
    "ai-sre/tools/mcp/internal/config"
    "ai-sre/tools/mcp/pkg/monitoring"
    "ai-sre/tools/mcp/pkg/cloud"
)

func main() {
    // 加载配置
    cfg, err := config.Load("configs/config.yaml")
    if err != nil {
        log.Fatal("Failed to load config:", err)
    }
    
    // 设置Gin模式
    gin.SetMode(cfg.Server.Mode)
    
    // 创建路由
    r := gin.Default()
    
    // 添加中间件
    r.Use(gin.Logger())
    r.Use(gin.Recovery())
    r.Use(corsMiddleware())
    
    // 初始化MCP工具客户端
    promClient, err := monitoring.NewPrometheusClient(cfg.Monitoring.Prometheus.URL)
    if err != nil {
        log.Fatal("Failed to create Prometheus client:", err)
    }
    
    grafanaClient := monitoring.NewGrafanaClient(
        cfg.Monitoring.Grafana.URL, 
        cfg.Monitoring.Grafana.APIKey,
    )
    
    awsClient, err := cloud.NewAWSClient(context.Background(), cfg.Cloud.AWS.Region)
    if err != nil {
        log.Fatal("Failed to create AWS client:", err)
    }
    
    // 设置路由
    promClient.SetupRoutes(r)
    grafanaClient.SetupRoutes(r)
    awsClient.SetupRoutes(r)
    
    // 健康检查
    r.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"status": "healthy"})
    })
    
    // 启动服务器
    srv := &http.Server{
        Addr:    fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
        Handler: r,
    }
    
    go func() {
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Failed to start server: %v", err)
        }
    }()
    
    // 优雅关闭
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    
    log.Println("Shutting down server...")
    
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    if err := srv.Shutdown(ctx); err != nil {
        log.Fatal("Server forced to shutdown:", err)
    }
    
    log.Println("Server exited")
}

func corsMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Header("Access-Control-Allow-Origin", "*")
        c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
        
        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(http.StatusNoContent)
            return
        }
        
        c.Next()
    }
}
```

#### 知识库结构
```
knowledge/
├── base/                  # 基础知识库
│   ├── troubleshooting/   # 故障排查手册
│   ├── best_practices/    # 最佳实践
│   ├── procedures/        # 操作程序
│   └── configurations/    # 配置模板
├── cases/                 # 历史案例库
│   ├── incidents/         # 故障案例
│   ├── changes/          # 变更案例
│   └── optimizations/    # 优化案例
└── docs/                 # 文档库
    ├── runbooks/         # 运行手册
    ├── architecture/     # 架构文档
    └── apis/            # API文档
```

#### 知识检索引擎
```python
class KnowledgeRetriever:
    def __init__(self, vector_store, llm):
        self.vector_store = vector_store
        self.llm = llm
        
    async def search(self, query, top_k=5):
        # 向量相似度搜索
        similar_docs = await self.vector_store.similarity_search(query, k=top_k)
        
        # 重排序和过滤
        ranked_docs = await self.rerank_documents(query, similar_docs)
        
        return ranked_docs
        
    async def generate_answer(self, query, context_docs):
        # 基于检索到的文档生成答案
        context = "\n".join([doc.page_content for doc in context_docs])
        
        prompt = f"""
        基于以下知识库内容回答问题：
        
        问题: {query}
        
        知识库内容:
        {context}
        
        请提供准确、详细的答案：
        """
        
        answer = await self.llm.agenerate([prompt])
        return answer
```

### 3. 子Agent池

#### 专项诊断Agent
```python
class DatabaseDiagnosisAgent(BaseAgent):
    """数据库诊断专家Agent"""
    
    async def diagnose_performance(self, db_metrics):
        # 数据库性能诊断
        pass
        
    async def analyze_slow_queries(self, slow_query_log):
        # 慢查询分析
        pass

class NetworkDiagnosisAgent(BaseAgent):
    """网络诊断专家Agent"""
    
    async def diagnose_connectivity(self, network_info):
        # 网络连通性诊断
        pass
        
    async def analyze_latency(self, latency_data):
        # 网络延迟分析
        pass
```

## 基础设施层组件

### 1. 消息队列系统

**Redis + Celery配置**:
```python
# celery_config.py
from celery import Celery

app = Celery('ai_sre')
app.config_from_object({
    'broker_url': 'redis://localhost:6379/0',
    'result_backend': 'redis://localhost:6379/0',
    'task_serializer': 'json',
    'accept_content': ['json'],
    'result_serializer': 'json',
    'timezone': 'UTC',
    'enable_utc': True,
})

@app.task
def process_monitoring_data(data):
    # 处理监控数据
    pass

@app.task
def execute_automation_task(task_config):
    # 执行自动化任务
    pass
```

### 2. 数据存储系统

#### 时序数据库 (InfluxDB)
```python
from influxdb_client import InfluxDBClient

class TimeSeriesDB:
    def __init__(self, url, token, org, bucket):
        self.client = InfluxDBClient(url=url, token=token, org=org)
        self.bucket = bucket
        
    async def write_metrics(self, metrics):
        # 写入时序数据
        write_api = self.client.write_api()
        write_api.write(bucket=self.bucket, record=metrics)
        
    async def query_metrics(self, query):
        # 查询时序数据
        query_api = self.client.query_api()
        result = query_api.query(query)
        return result
```

#### 关系数据库 (PostgreSQL)
```python
from sqlalchemy import create_engine
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import sessionmaker

Base = declarative_base()

class Task(Base):
    __tablename__ = 'tasks'
    
    id = Column(Integer, primary_key=True)
    name = Column(String(255), nullable=False)
    status = Column(String(50), nullable=False)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

class Agent(Base):
    __tablename__ = 'agents'
    
    id = Column(Integer, primary_key=True)
    name = Column(String(255), nullable=False)
    type = Column(String(100), nullable=False)
    status = Column(String(50), nullable=False)
    config = Column(JSON)
```

### 3. 配置管理

**配置文件结构**:
```yaml
# config.yaml
app:
  name: "AI SRE Assistant"
  version: "1.0.0"
  debug: false

database:
  postgresql:
    host: "localhost"
    port: 5432
    database: "ai_sre"
    username: "ai_sre_user"
    password: "${DB_PASSWORD}"
  
  influxdb:
    url: "http://localhost:8086"
    token: "${INFLUXDB_TOKEN}"
    org: "ai-sre"
    bucket: "metrics"

redis:
  host: "localhost"
  port: 6379
  db: 0
  password: "${REDIS_PASSWORD}"

agents:
  master:
    enabled: true
    max_concurrent_tasks: 10
  
  monitoring:
    enabled: true
    check_interval: 30
    
  diagnosis:
    enabled: true
    timeout: 300
    
  automation:
    enabled: true
    dry_run: false

integrations:
  wechat:
    corp_id: "${WECHAT_CORP_ID}"
    agent_id: "${WECHAT_AGENT_ID}"
    secret: "${WECHAT_SECRET}"
  
  prometheus:
    url: "http://localhost:9090"
    
  grafana:
    url: "http://localhost:3000"
    api_key: "${GRAFANA_API_KEY}"

logging:
  level: "INFO"
  format: "%(asctime)s - %(name)s - %(levelname)s - %(message)s"
  handlers:
    - type: "console"
    - type: "file"
      filename: "logs/ai-sre.log"
      max_bytes: 10485760
      backup_count: 5
```

## 部署组件

### Docker配置

**Dockerfile**:
```dockerfile
FROM python:3.9-slim

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY src/ ./src/
COPY configs/ ./configs/

EXPOSE 8000

CMD ["uvicorn", "src.main:app", "--host", "0.0.0.0", "--port", "8000"]
```

**docker-compose.yml**:
```yaml
version: '3.8'

services:
  ai-sre-api:
    build: .
    ports:
      - "8000:8000"
    environment:
      - DB_PASSWORD=${DB_PASSWORD}
      - REDIS_PASSWORD=${REDIS_PASSWORD}
    depends_on:
      - postgres
      - redis
      - influxdb

  postgres:
    image: postgres:13
    environment:
      POSTGRES_DB: ai_sre
      POSTGRES_USER: ai_sre_user
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:6-alpine
    command: redis-server --requirepass ${REDIS_PASSWORD}

  influxdb:
    image: influxdb:2.0
    environment:
      DOCKER_INFLUXDB_INIT_MODE: setup
      DOCKER_INFLUXDB_INIT_USERNAME: admin
      DOCKER_INFLUXDB_INIT_PASSWORD: ${INFLUXDB_PASSWORD}
      DOCKER_INFLUXDB_INIT_ORG: ai-sre
      DOCKER_INFLUXDB_INIT_BUCKET: metrics

volumes:
  postgres_data:
```

### Kubernetes部署

**Helm Chart结构**:
```
charts/ai-sre/
├── Chart.yaml
├── values.yaml
├── values.prod.yaml
├── templates/
│   ├── deployment.yaml
│   ├── service.yaml
│   ├── ingress.yaml
│   ├── configmap.yaml
│   ├── secret.yaml
│   └── hpa.yaml
└── charts/
    ├── postgresql/
    ├── redis/
    └── influxdb/
```

这个组件说明文档详细描述了AI SRE分身助理系统中各个组件的功能、实现方式和配置方法，为开发和部署提供了完整的技术参考。