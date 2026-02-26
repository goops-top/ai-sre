#!/bin/bash

# AI SRE MCP Server 部署脚本
# 使用方法: ./deploy.sh [build|run|stop|logs|clean]

set -e

# 配置变量
IMAGE_NAME="ai-sre-mcp-server"
CONTAINER_NAME="ai-sre-mcp-server"
VERSION="1.0.0"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查Docker是否安装
check_docker() {
    if ! command -v docker &> /dev/null; then
        log_error "Docker 未安装，请先安装 Docker"
        exit 1
    fi
    
    if ! docker info &> /dev/null; then
        log_error "Docker 服务未运行，请启动 Docker 服务"
        exit 1
    fi
}

# 构建镜像
build_image() {
    log_info "开始构建 Docker 镜像..."
    
    # 检查是否存在 Dockerfile
    if [ ! -f "Dockerfile" ]; then
        log_error "Dockerfile 不存在"
        exit 1
    fi
    
    # 构建镜像
    docker build -t ${IMAGE_NAME}:${VERSION} -t ${IMAGE_NAME}:latest .
    
    if [ $? -eq 0 ]; then
        log_success "Docker 镜像构建成功"
        docker images | grep ${IMAGE_NAME}
    else
        log_error "Docker 镜像构建失败"
        exit 1
    fi
}

# 运行容器
run_container() {
    log_info "启动 MCP Server 容器..."
    
    # 停止已存在的容器
    if docker ps -a | grep -q ${CONTAINER_NAME}; then
        log_warning "发现已存在的容器，正在停止并删除..."
        docker stop ${CONTAINER_NAME} 2>/dev/null || true
        docker rm ${CONTAINER_NAME} 2>/dev/null || true
    fi
    
    # 检查环境变量
    if [ -z "$TENCENTCLOUD_SECRET_ID" ] || [ -z "$TENCENTCLOUD_SECRET_KEY" ]; then
        log_warning "腾讯云认证环境变量未设置，请确保设置了以下环境变量:"
        log_warning "  - TENCENTCLOUD_SECRET_ID"
        log_warning "  - TENCENTCLOUD_SECRET_KEY"
        log_warning "  - TENCENTCLOUD_REGION (可选，默认: ap-beijing)"
    fi
    
    # 创建日志目录
    mkdir -p logs configs
    
    # 运行容器
    docker run -d \
        --name ${CONTAINER_NAME} \
        --restart unless-stopped \
        -p 8080:8080 \
        -e TENCENTCLOUD_SECRET_ID="${TENCENTCLOUD_SECRET_ID}" \
        -e TENCENTCLOUD_SECRET_KEY="${TENCENTCLOUD_SECRET_KEY}" \
        -e TENCENTCLOUD_REGION="${TENCENTCLOUD_REGION:-ap-beijing}" \
        -e MCP_LOG_LEVEL="${MCP_LOG_LEVEL:-info}" \
        -v $(pwd)/logs:/app/logs \
        -v $(pwd)/configs:/app/configs \
        ${IMAGE_NAME}:latest
    
    if [ $? -eq 0 ]; then
        log_success "MCP Server 容器启动成功"
        log_info "容器状态:"
        docker ps | grep ${CONTAINER_NAME}
        log_info "服务地址: http://localhost:8080"
        log_info "查看日志: ./deploy.sh logs"
    else
        log_error "容器启动失败"
        exit 1
    fi
}

# 使用docker-compose运行
run_compose() {
    log_info "使用 docker-compose 启动服务..."
    
    if [ ! -f "docker-compose.yml" ]; then
        log_error "docker-compose.yml 文件不存在"
        exit 1
    fi
    
    # 检查环境变量
    if [ -z "$TENCENTCLOUD_SECRET_ID" ] || [ -z "$TENCENTCLOUD_SECRET_KEY" ]; then
        log_warning "请在 docker-compose.yml 中设置腾讯云认证环境变量"
    fi
    
    docker-compose up -d
    
    if [ $? -eq 0 ]; then
        log_success "服务启动成功"
        docker-compose ps
    else
        log_error "服务启动失败"
        exit 1
    fi
}

# 停止容器
stop_container() {
    log_info "停止 MCP Server 容器..."
    
    if docker ps | grep -q ${CONTAINER_NAME}; then
        docker stop ${CONTAINER_NAME}
        log_success "容器已停止"
    else
        log_warning "容器未运行"
    fi
}

# 停止docker-compose服务
stop_compose() {
    log_info "停止 docker-compose 服务..."
    docker-compose down
    log_success "服务已停止"
}

# 查看日志
show_logs() {
    if docker ps | grep -q ${CONTAINER_NAME}; then
        log_info "显示容器日志 (按 Ctrl+C 退出):"
        docker logs -f ${CONTAINER_NAME}
    else
        log_error "容器未运行"
        exit 1
    fi
}

# 清理资源
clean_resources() {
    log_info "清理 Docker 资源..."
    
    # 停止并删除容器
    if docker ps -a | grep -q ${CONTAINER_NAME}; then
        docker stop ${CONTAINER_NAME} 2>/dev/null || true
        docker rm ${CONTAINER_NAME} 2>/dev/null || true
        log_success "容器已删除"
    fi
    
    # 删除镜像
    if docker images | grep -q ${IMAGE_NAME}; then
        docker rmi ${IMAGE_NAME}:latest ${IMAGE_NAME}:${VERSION} 2>/dev/null || true
        log_success "镜像已删除"
    fi
    
    # 清理未使用的资源
    docker system prune -f
    log_success "清理完成"
}

# 显示帮助信息
show_help() {
    echo "AI SRE MCP Server 部署脚本"
    echo ""
    echo "使用方法:"
    echo "  ./deploy.sh [命令]"
    echo ""
    echo "命令:"
    echo "  build       构建 Docker 镜像"
    echo "  run         运行容器 (单容器模式)"
    echo "  compose     使用 docker-compose 运行"
    echo "  stop        停止容器"
    echo "  stop-compose 停止 docker-compose 服务"
    echo "  logs        查看容器日志"
    echo "  clean       清理所有资源"
    echo "  help        显示此帮助信息"
    echo ""
    echo "环境变量:"
    echo "  TENCENTCLOUD_SECRET_ID    腾讯云 SecretID (必需)"
    echo "  TENCENTCLOUD_SECRET_KEY   腾讯云 SecretKey (必需)"
    echo "  TENCENTCLOUD_REGION       腾讯云地域 (可选，默认: ap-beijing)"
    echo "  MCP_LOG_LEVEL            日志级别 (可选，默认: info)"
    echo ""
    echo "示例:"
    echo "  # 构建并运行"
    echo "  export TENCENTCLOUD_SECRET_ID=\"your_secret_id\""
    echo "  export TENCENTCLOUD_SECRET_KEY=\"your_secret_key\""
    echo "  ./deploy.sh build"
    echo "  ./deploy.sh run"
    echo ""
    echo "  # 使用 docker-compose"
    echo "  ./deploy.sh compose"
}

# 主函数
main() {
    check_docker
    
    case "${1:-help}" in
        build)
            build_image
            ;;
        run)
            run_container
            ;;
        compose)
            run_compose
            ;;
        stop)
            stop_container
            ;;
        stop-compose)
            stop_compose
            ;;
        logs)
            show_logs
            ;;
        clean)
            clean_resources
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            log_error "未知命令: $1"
            show_help
            exit 1
            ;;
    esac
}

# 执行主函数
main "$@"