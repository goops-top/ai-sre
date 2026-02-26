# AI SRE 项目 Makefile
# 提供统一的构建、测试、部署命令

.PHONY: help setup clean build test lint generate validate deploy mcp

# 默认目标
help: ## 显示帮助信息
	@echo "AI SRE 分身助理项目"
	@echo ""
	@echo "可用命令:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# 环境变量
PYTHON := python3
GO := go
NODE := node
NPM := npm
DOCKER := docker
KUBECTL := kubectl
HELM := helm

# 项目路径
PROJECT_ROOT := $(shell pwd)
SRC_DIR := $(PROJECT_ROOT)/src
TOOLS_DIR := $(PROJECT_ROOT)/tools
SPECS_DIR := $(PROJECT_ROOT)/specs
DOCS_DIR := $(PROJECT_ROOT)/docs

# 版本信息
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

#==============================================================================
# 环境设置
#==============================================================================

setup: setup-python setup-go setup-node setup-tools ## 设置开发环境
	@echo " 开发环境设置完成"

setup-python: ## 设置Python环境
	@echo " 设置Python环境..."
	$(PYTHON) -m venv venv
	./venv/bin/pip install --upgrade pip
	./venv/bin/pip install -r requirements.txt
	./venv/bin/pip install -r requirements-dev.txt

setup-go: ## 设置Go环境
	@echo " 设置Go环境..."
	cd $(TOOLS_DIR)/mcp && $(GO) mod download
	cd $(TOOLS_DIR)/mcp && $(GO) mod tidy

setup-node: ## 设置Node.js环境
	@echo " 设置Node.js环境..."
	$(NPM) install -g @apidevtools/swagger-cli
	$(NPM) install -g @openapitools/openapi-generator-cli
	$(NPM) install -g @stoplight/spectral-cli
	cd $(SRC_DIR)/interfaces/web/frontend && $(NPM) install

setup-tools: ## 安装开发工具
	@echo " 安装开发工具..."
	# 安装spec-kit CLI
	@if ! command -v specify >/dev/null 2>&1; then \
		echo "安装spec-kit CLI..."; \
		uv tool install specify-cli --from git+https://github.com/github/spec-kit.git; \
	fi
	# Protocol Buffers
	@if ! command -v protoc >/dev/null 2>&1; then \
		echo "请安装 Protocol Buffers: https://grpc.io/docs/protoc-installation/"; \
		exit 1; \
	fi
	# buf工具
	@if ! command -v buf >/dev/null 2>&1; then \
		echo "安装buf工具..."; \
		curl -sSL "https://github.com/bufbuild/buf/releases/latest/download/buf-$$(uname -s)-$$(uname -m)" -o /usr/local/bin/buf; \
		chmod +x /usr/local/bin/buf; \
	fi

#==============================================================================
# 代码生成
#==============================================================================

generate: generate-proto generate-openapi ## 生成所有代码
	@echo " 代码生成完成"

generate-proto: ## 生成Protocol Buffers代码
	@echo " 生成Protocol Buffers代码..."
	# 生成Go代码
	cd $(SPECS_DIR) && buf generate
	# 生成Python代码
	cd $(SPECS_DIR) && python -m grpc_tools.protoc \
		--proto_path=. \
		--python_out=../src/generated \
		--grpc_python_out=../src/generated \
		proto/agent/*.proto proto/mcp/*.proto

generate-openapi: ## 生成OpenAPI客户端代码
	@echo " 生成OpenAPI客户端代码..."
	# 生成Python客户端
	openapi-generator-cli generate \
		-i $(SPECS_DIR)/openapi/agent-api.yaml \
		-g python \
		-o $(SRC_DIR)/generated/agent_client \
		--package-name agent_client
	# 生成Go客户端
	openapi-generator-cli generate \
		-i $(SPECS_DIR)/openapi/mcp-tools.yaml \
		-g go \
		-o $(TOOLS_DIR)/generated/mcp_client \
		--package-name mcp_client
	# 生成TypeScript客户端
	openapi-generator-cli generate \
		-i $(SPECS_DIR)/openapi/agent-api.yaml \
		-g typescript-axios \
		-o $(SRC_DIR)/interfaces/web/frontend/src/generated/api

#==============================================================================
# 规范验证
#==============================================================================

validate: validate-specs validate-quality validate-openapi validate-proto validate-schemas validate-emoji ## 验证所有规范
	@echo " 规范验证通过"

validate-specs: ## 验证spec-kit功能规范
	@echo " 验证功能规范..."
	./.specify/scripts/speckit-workflow.sh validate-all

validate-quality: ## 验证代码质量和文档规范
	@echo " 验证代码质量和文档规范..."
	./.specify/scripts/speckit-workflow.sh validate-quality

validate-emoji: ## 严格验证emoji规范（所有文件类型）
	@echo " 严格验证emoji规范..."
	./.specify/scripts/strict-emoji-check.sh check

clean-emoji: ## 清理所有文件中的emoji表情
	@echo " 清理所有文件中的emoji表情..."
	./.specify/scripts/strict-emoji-check.sh clean

validate-openapi: ## 验证OpenAPI规范
	@echo " 验证OpenAPI规范..."
	swagger-cli validate $(SPECS_DIR)/openapi/agent-api.yaml
	swagger-cli validate $(SPECS_DIR)/openapi/mcp-tools.yaml
	spectral lint $(SPECS_DIR)/openapi/*.yaml

validate-proto: ## 验证Protocol Buffers
	@echo " 验证Protocol Buffers..."
	cd $(SPECS_DIR) && buf lint
	cd $(SPECS_DIR) && buf breaking --against '.git#branch=main'

validate-schemas: ## 验证JSON Schema
	@echo " 验证JSON Schema..."
	# 这里可以添加JSON Schema验证逻辑

#==============================================================================
# 构建
#==============================================================================

build: build-python build-go build-frontend ## 构建所有组件
	@echo " 构建完成"

build-python: ## 构建Python组件
	@echo " 构建Python组件..."
	cd $(SRC_DIR) && $(PYTHON) -m py_compile **/*.py

build-go: ## 构建Go MCP工具
	@echo " 构建Go MCP工具..."
	cd $(TOOLS_DIR)/mcp && $(GO) build -ldflags="-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.commit=$(GIT_COMMIT)" -o bin/mcp-server ./cmd/mcp-server

build-frontend: ## 构建前端
	@echo " 构建前端..."
	cd $(SRC_DIR)/interfaces/web/frontend && $(NPM) run build

#==============================================================================
# 测试
#==============================================================================

test: test-python test-go test-frontend ## 运行所有测试
	@echo " 所有测试通过"

test-python: ## 运行Python测试
	@echo " 运行Python测试..."
	cd $(SRC_DIR) && $(PYTHON) -m pytest tests/ -v --cov=. --cov-report=html

test-go: ## 运行Go测试
	@echo " 运行Go测试..."
	cd $(TOOLS_DIR)/mcp && $(GO) test -v -race -coverprofile=coverage.out ./...
	cd $(TOOLS_DIR)/mcp && $(GO) tool cover -html=coverage.out -o coverage.html

test-frontend: ## 运行前端测试
	@echo " 运行前端测试..."
	cd $(SRC_DIR)/interfaces/web/frontend && $(NPM) test

test-integration: ## 运行集成测试
	@echo " 运行集成测试..."
	cd $(PROJECT_ROOT) && $(PYTHON) -m pytest tests/integration/ -v

test-contract: ## 运行契约测试
	@echo " 运行契约测试..."
	# 这里可以添加契约测试逻辑

#==============================================================================
# 代码质量
#==============================================================================

lint: lint-python lint-go lint-frontend lint-docs lint-comments ## 运行所有代码检查
	@echo " 代码检查通过"

lint-python: ## Python代码检查
	@echo " Python代码检查..."
	cd $(SRC_DIR) && black --check .
	cd $(SRC_DIR) && isort --check-only .
	cd $(SRC_DIR) && flake8 .
	cd $(SRC_DIR) && mypy .

lint-go: ## Go代码检查
	@echo " Go代码检查..."
	cd $(TOOLS_DIR)/mcp && $(GO) fmt ./...
	cd $(TOOLS_DIR)/mcp && $(GO) vet ./...
	cd $(TOOLS_DIR)/mcp && golangci-lint run

lint-frontend: ## 前端代码检查
	@echo " 前端代码检查..."
	cd $(SRC_DIR)/interfaces/web/frontend && $(NPM) run lint

lint-docs: ## 文档规范检查
	@echo " 检查文档emoji规范..."
	./.specify/scripts/clean-emoji.sh check

lint-comments: ## 代码注释检查
	@echo " 检查代码注释覆盖率..."
	python3 ./.specify/scripts/check-comments.py $(SRC_DIR) --min-coverage 80
	@if [ -d "$(TOOLS_DIR)/mcp" ]; then \
		python3 ./.specify/scripts/check-comments.py $(TOOLS_DIR)/mcp --min-coverage 80; \
	fi

format: format-python format-go format-frontend ## 格式化所有代码
	@echo " 代码格式化完成"

format-python: ## 格式化Python代码
	@echo " 格式化Python代码..."
	cd $(SRC_DIR) && black .
	cd $(SRC_DIR) && isort .

format-go: ## 格式化Go代码
	@echo " 格式化Go代码..."
	cd $(TOOLS_DIR)/mcp && $(GO) fmt ./...

format-frontend: ## 格式化前端代码
	@echo " 格式化前端代码..."
	cd $(SRC_DIR)/interfaces/web/frontend && $(NPM) run format

#==============================================================================
# Docker
#==============================================================================

# MCP镜像相关变量
MCP_IMAGE_NAME := ai-sre-mcp-server
MCP_REGISTRY := ccr.ccs.tencentyun.com/goops

mcp: ## 构建、标签和推送MCP服务器镜像到腾讯云镜像仓库
	@echo "🚀 构建并推送MCP服务器镜像..."
	@echo "📦 执行构建脚本..."
	cd $(TOOLS_DIR)/mcp && sh deploy.sh build
	@echo "🏷️  添加镜像标签..."
	$(DOCKER) tag $(MCP_IMAGE_NAME):latest $(MCP_REGISTRY)/$(MCP_IMAGE_NAME):latest
	$(DOCKER) tag $(MCP_IMAGE_NAME):latest $(MCP_REGISTRY)/$(MCP_IMAGE_NAME):$(VERSION)
	@echo "⬆️  推送镜像到腾讯云镜像仓库..."
	$(DOCKER) push $(MCP_REGISTRY)/$(MCP_IMAGE_NAME):latest
	$(DOCKER) push $(MCP_REGISTRY)/$(MCP_IMAGE_NAME):$(VERSION)
	@echo "✅ MCP镜像构建和推送完成!"
	@echo "   镜像地址: $(MCP_REGISTRY)/$(MCP_IMAGE_NAME):latest"
	@echo "   版本标签: $(MCP_REGISTRY)/$(MCP_IMAGE_NAME):$(VERSION)"

docker-build: ## 构建Docker镜像
	@echo "🐳 构建Docker镜像..."
	$(DOCKER) build -t ai-sre/agent:$(VERSION) -f docker/agent.Dockerfile .
	$(DOCKER) build -t ai-sre/mcp-monitoring:$(VERSION) -f $(TOOLS_DIR)/mcp/docker/monitoring.Dockerfile $(TOOLS_DIR)/mcp
	$(DOCKER) build -t ai-sre/mcp-cloud:$(VERSION) -f $(TOOLS_DIR)/mcp/docker/cloud.Dockerfile $(TOOLS_DIR)/mcp
	$(DOCKER) build -t ai-sre/mcp-container:$(VERSION) -f $(TOOLS_DIR)/mcp/docker/container.Dockerfile $(TOOLS_DIR)/mcp
	$(DOCKER) build -t ai-sre/mcp-database:$(VERSION) -f $(TOOLS_DIR)/mcp/docker/database.Dockerfile $(TOOLS_DIR)/mcp
	$(DOCKER) build -t ai-sre/web:$(VERSION) -f docker/web.Dockerfile .

docker-push: ## 推送Docker镜像
	@echo " 推送Docker镜像..."
	$(DOCKER) push ai-sre/agent:$(VERSION)
	$(DOCKER) push ai-sre/mcp-monitoring:$(VERSION)
	$(DOCKER) push ai-sre/mcp-cloud:$(VERSION)
	$(DOCKER) push ai-sre/mcp-container:$(VERSION)
	$(DOCKER) push ai-sre/mcp-database:$(VERSION)
	$(DOCKER) push ai-sre/web:$(VERSION)

#==============================================================================
# 开发服务
#==============================================================================

dev-start: ## 启动开发环境
	@echo " 启动开发环境..."
	docker-compose -f docker-compose.dev.yml up -d

dev-stop: ## 停止开发环境
	@echo " 停止开发环境..."
	docker-compose -f docker-compose.dev.yml down

dev-logs: ## 查看开发环境日志
	docker-compose -f docker-compose.dev.yml logs -f

dev-restart: dev-stop dev-start ## 重启开发环境

#==============================================================================
# 部署
#==============================================================================

deploy-staging: ## 部署到测试环境
	@echo " 部署到测试环境..."
	$(HELM) upgrade --install ai-sre-staging ./charts/ai-sre \
		--namespace ai-sre-staging \
		--create-namespace \
		--values values.staging.yaml \
		--set image.tag=$(VERSION)

deploy-prod: ## 部署到生产环境
	@echo " 部署到生产环境..."
	$(HELM) upgrade --install ai-sre ./charts/ai-sre \
		--namespace ai-sre \
		--create-namespace \
		--values values.prod.yaml \
		--set image.tag=$(VERSION)

#==============================================================================
# 文档
#==============================================================================

docs-generate: ## 生成API文档
	@echo " 生成API文档..."
	# 生成OpenAPI文档
	swagger-cli bundle $(SPECS_DIR)/openapi/agent-api.yaml -o $(DOCS_DIR)/api/agent-api.html -t html
	swagger-cli bundle $(SPECS_DIR)/openapi/mcp-tools.yaml -o $(DOCS_DIR)/api/mcp-tools.html -t html
	# 生成Protocol Buffers文档
	cd $(SPECS_DIR) && buf generate --template buf.gen.docs.yaml

docs-serve: ## 启动文档服务器
	@echo " 启动文档服务器..."
	cd $(DOCS_DIR) && $(PYTHON) -m http.server 8080

#==============================================================================
# 清理
#==============================================================================

clean: ## 清理构建产物
	@echo " 清理构建产物..."
	rm -rf $(SRC_DIR)/generated/
	rm -rf $(TOOLS_DIR)/generated/
	rm -rf $(TOOLS_DIR)/mcp/bin/
	rm -rf $(SRC_DIR)/interfaces/web/frontend/dist/
	rm -rf $(SRC_DIR)/interfaces/web/frontend/build/
	find . -type d -name "__pycache__" -exec rm -rf {} +
	find . -type f -name "*.pyc" -delete
	find . -type f -name "coverage.out" -delete
	find . -type f -name "coverage.html" -delete

clean-all: clean ## 清理所有文件（包括依赖）
	rm -rf venv/
	rm -rf node_modules/
	rm -rf $(SRC_DIR)/interfaces/web/frontend/node_modules/

#==============================================================================
# 实用工具
#==============================================================================

check-deps: ## 检查依赖更新
	@echo " 检查依赖更新..."
	cd $(SRC_DIR) && pip list --outdated
	cd $(TOOLS_DIR)/mcp && $(GO) list -u -m all
	cd $(SRC_DIR)/interfaces/web/frontend && $(NPM) outdated

security-scan: ## 安全扫描
	@echo " 运行安全扫描..."
	cd $(SRC_DIR) && safety check
	cd $(TOOLS_DIR)/mcp && gosec ./...
	cd $(SRC_DIR)/interfaces/web/frontend && $(NPM) audit

performance-test: ## 性能测试
	@echo " 运行性能测试..."
	# 这里可以添加性能测试逻辑

load-test: ## 负载测试
	@echo " 运行负载测试..."
	# 这里可以添加负载测试逻辑

#==============================================================================
# CI/CD辅助
#==============================================================================

ci-setup: setup generate validate ## CI环境设置
	@echo " CI环境设置完成"

ci-test: test lint ## CI测试流程
	@echo " CI测试完成"

ci-build: build docker-build ## CI构建流程
	@echo " CI构建完成"

ci-deploy: docker-push deploy-staging ## CI部署流程
	@echo " CI部署完成"

#==============================================================================
# 版本管理
#==============================================================================

version: ## 显示版本信息
	@echo "版本: $(VERSION)"
	@echo "构建时间: $(BUILD_TIME)"
	@echo "Git提交: $(GIT_COMMIT)"

tag: ## 创建版本标签
	@read -p "输入版本号 (例如: v1.0.0): " version; \
	git tag -a $$version -m "Release $$version"; \
	git push origin $$version

#==============================================================================
# 帮助和状态
#==============================================================================

status: ## 显示项目状态
	@echo " 项目状态:"
	@echo "  Python版本: $$($(PYTHON) --version)"
	@echo "  Go版本: $$($(GO) version)"
	@echo "  Node版本: $$($(NODE) --version)"
	@echo "  Docker版本: $$($(DOCKER) --version)"
	@echo "  项目版本: $(VERSION)"
	@echo ""
	@echo " 目录结构:"
	@find . -type d -name ".*" -prune -o -type d -print | head -20

info: ## 显示项目信息
	@echo " AI SRE 分身助理项目"
	@echo ""
	@echo " 项目概述:"
	@echo "  基于AI技术构建的智能化SRE运维助理系统"
	@echo "  提供7x24小时的自动化运维服务"
	@echo ""
	@echo " 技术栈:"
	@echo "  Agent编排层: Python + FastAPI + LangChain"
	@echo "  MCP工具层: Go + Gin + gRPC"
	@echo "  前端界面: React + TypeScript + Ant Design"
	@echo "  基础设施: Docker + Kubernetes + Helm"
	@echo ""
	@echo " 文档:"
	@echo "  架构设计: docs/architecture.md"
	@echo "  组件说明: docs/components.md"
	@echo "  快速开始: docs/getting-started.md"
	@echo "  API规范: specs/openapi/"

#==============================================================================
# Spec-Kit 规范驱动开发
#==============================================================================

spec-init: ## 初始化新功能规范 (用法: make spec-init FEATURE=功能名称)
	@if [ -z "$(FEATURE)" ]; then \
		echo " 请指定功能名称: make spec-init FEATURE=功能名称"; \
		exit 1; \
	fi
	@echo " 初始化功能规范: $(FEATURE)"
	./.specify/scripts/speckit-workflow.sh init-feature $(FEATURE)

spec-validate: ## 验证功能规范 (用法: make spec-validate FEATURE=功能名称)
	@if [ -z "$(FEATURE)" ]; then \
		echo " 验证所有功能规范..."; \
		./.specify/scripts/speckit-workflow.sh validate-all; \
	else \
		echo " 验证功能规范: $(FEATURE)"; \
		./.specify/scripts/speckit-workflow.sh validate-spec $(FEATURE); \
	fi

spec-plan: ## 生成技术计划 (用法: make spec-plan FEATURE=功能名称)
	@if [ -z "$(FEATURE)" ]; then \
		echo " 请指定功能名称: make spec-plan FEATURE=功能名称"; \
		exit 1; \
	fi
	@echo " 生成技术计划: $(FEATURE)"
	./.specify/scripts/speckit-workflow.sh generate-plan $(FEATURE)

spec-tasks: ## 生成任务分解 (用法: make spec-tasks FEATURE=功能名称)
	@if [ -z "$(FEATURE)" ]; then \
		echo " 请指定功能名称: make spec-tasks FEATURE=功能名称"; \
		exit 1; \
	fi
	@echo " 生成任务分解: $(FEATURE)"
	./.specify/scripts/speckit-workflow.sh generate-tasks $(FEATURE)

spec-implement: ## 实施功能开发 (用法: make spec-implement FEATURE=功能名称)
	@if [ -z "$(FEATURE)" ]; then \
		echo " 请指定功能名称: make spec-implement FEATURE=功能名称"; \
		exit 1; \
	fi
	@echo " 实施功能开发: $(FEATURE)"
	./.specify/scripts/speckit-workflow.sh implement $(FEATURE)

spec-check-emoji: ## 检查文档emoji规范
	@echo " 检查文档emoji规范..."
	./.specify/scripts/clean-emoji.sh check

spec-check-comments: ## 检查代码注释覆盖率
	@echo " 检查代码注释覆盖率..."
	python3 ./.specify/scripts/check-comments.py $(SRC_DIR) --min-coverage 80 --verbose

spec-clean: ## 清理spec-kit临时文件
	@echo " 清理spec-kit临时文件..."
	./.specify/scripts/speckit-workflow.sh clean

spec-help: ## 显示spec-kit使用帮助
	@echo " Spec-Kit 规范驱动开发工具"
	@echo ""
	@echo " 开发新功能的完整流程:"
	@echo "  1. make spec-init FEATURE=功能名称     # 初始化功能规范"
	@echo "  2. 编辑 .specify/specs/功能名称/spec.md  # 编写详细规范"
	@echo "  3. make spec-validate FEATURE=功能名称  # 验证规范"
	@echo "  4. 在AI助手中使用spec-kit命令:"
	@echo "     /speckit.constitution"
	@echo "     /speckit.specify"
	@echo "     /speckit.clarify"
	@echo "     /speckit.plan"
	@echo "     /speckit.tasks"
	@echo "     /speckit.implement"
	@echo ""
	@echo " 相关文档:"
	@echo "  项目宪章: .specify/memory/constitution.md"
	@echo "  AI助手配置: .specify/ai-assistant-config.md"
	@echo "  使用指南: .specify/README.md"
