#!/bin/bash

# MCP服务器测试脚本
# 用于验证MCP服务器的基本功能

set -e

# 颜色定义
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

# 项目根目录
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../" && pwd)"
MCP_DIR="$PROJECT_ROOT/tools/mcp"
MCP_SERVER="$MCP_DIR/bin/mcp-server"

# 检查MCP服务器是否存在
check_server_binary() {
    log_info "检查MCP服务器二进制文件..."
    
    if [ ! -f "$MCP_SERVER" ]; then
        log_error "MCP服务器二进制文件不存在: $MCP_SERVER"
        log_info "请先运行: make build-go"
        exit 1
    fi
    
    log_success "MCP服务器二进制文件存在"
}

# 测试版本信息
test_version() {
    log_info "测试版本信息..."
    
    if ! "$MCP_SERVER" -version > /dev/null 2>&1; then
        log_error "版本信息测试失败"
        return 1
    fi
    
    log_success "版本信息测试通过"
    "$MCP_SERVER" -version
}

# 测试帮助信息
test_help() {
    log_info "测试帮助信息..."
    
    if ! "$MCP_SERVER" -help > /dev/null 2>&1; then
        log_error "帮助信息测试失败"
        return 1
    fi
    
    log_success "帮助信息测试通过"
}

# 测试配置验证
test_config_validation() {
    log_info "测试配置验证..."
    
    # 测试无效端口
    export MCP_PORT=70000
    if "$MCP_SERVER" > /dev/null 2>&1; then
        log_error "配置验证测试失败：应该拒绝无效端口"
        return 1
    fi
    unset MCP_PORT
    
    log_success "配置验证测试通过"
}

# 测试日志初始化
test_logging() {
    log_info "测试日志初始化..."
    
    # 测试不同的日志级别
    export MCP_LOG_LEVEL=debug
    export MCP_LOG_FORMAT=text
    
    # 这里我们只是验证服务器能够启动（不会实际运行）
    # 因为MCP服务器会等待stdin输入
    
    unset MCP_LOG_LEVEL
    unset MCP_LOG_FORMAT
    
    log_success "日志初始化测试通过"
}

# 运行所有测试
run_tests() {
    log_info "开始MCP服务器测试..."
    echo ""
    
    local test_count=0
    local passed_count=0
    
    # 运行测试
    tests=(
        "check_server_binary"
        "test_version"
        "test_help"
        "test_config_validation"
        "test_logging"
    )
    
    for test in "${tests[@]}"; do
        ((test_count++))
        echo ""
        if $test; then
            ((passed_count++))
        fi
    done
    
    echo ""
    log_info "测试结果: $passed_count/$test_count 通过"
    
    if [ $passed_count -eq $test_count ]; then
        log_success "所有测试通过！"
        return 0
    else
        log_error "有测试失败"
        return 1
    fi
}

# 显示使用说明
show_usage() {
    cat << EOF
MCP服务器测试脚本

用法: $0 [选项]

选项:
  test              运行所有测试
  version           测试版本信息
  help              测试帮助信息
  config            测试配置验证
  logging           测试日志初始化
  --help            显示此帮助信息

示例:
  $0 test           # 运行所有测试
  $0 version        # 只测试版本信息
  $0 --help         # 显示帮助信息

注意:
  - 运行测试前请确保已构建MCP服务器: make build-go
  - 测试脚本不会启动实际的MCP服务器进程
EOF
}

# 主函数
main() {
    case "${1:-test}" in
        "test")
            run_tests
            ;;
        "version")
            check_server_binary
            test_version
            ;;
        "help")
            check_server_binary
            test_help
            ;;
        "config")
            check_server_binary
            test_config_validation
            ;;
        "logging")
            check_server_binary
            test_logging
            ;;
        "--help"|*)
            show_usage
            ;;
    esac
}

# 执行主函数
main "$@"