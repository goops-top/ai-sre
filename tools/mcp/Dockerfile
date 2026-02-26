# 第一阶段：构建阶段
FROM golang:1.24-alpine AS builder

# 安装必要的工具
RUN apk add --no-cache git ca-certificates tzdata

# 设置工作目录
WORKDIR /build

# 设置Go环境变量
ENV GO111MODULE=on \
    GOPROXY=https://goproxy.cn,direct \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# 复制go mod文件并下载依赖
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# 复制源代码
COPY . .

# 构建应用程序
RUN go build -ldflags="-w -s -X main.version=1.0.0 -X main.commit=$(git rev-parse --short HEAD 2>/dev/null || echo 'unknown') -X main.buildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    -o mcp-server ./cmd/mcp-server

# 第二阶段：运行阶段，使用指定的腾讯云基础镜像
FROM ccr.ccs.tencentyun.com/goops/smarttools:v1

# 设置工作目录
WORKDIR /app

# 从构建阶段复制编译好的二进制文件
COPY --from=builder /build/mcp-server .

# 从构建阶段复制时区信息和CA证书
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo


# 创建必要的目录并设置权限
RUN mkdir -p /app/logs /app/configs 


# 暴露端口（默认8080，可通过环境变量覆盖）
EXPOSE 8080

# 设置默认环境变量
ENV MCP_LOG_LEVEL=info \
    MCP_LOG_FORMAT=json \
    MCP_TRANSPORT=http \
    MCP_PORT=8080 \
    MCP_HOST=0.0.0.0 \
    MCP_ENABLE_TOOLS=true \
    MCP_ENABLE_LOGGING=true

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:${MCP_PORT}/health || exit 1

# 启动命令
CMD ["./mcp-server", "-transport", "http"]
