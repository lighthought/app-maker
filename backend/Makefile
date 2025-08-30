.PHONY: help build build-dev build-prod run-dev run-prod test clean validate-config

# 默认目标
help:
	@echo "AutoCodeWeb Backend 构建工具"
	@echo "=========================="
	@echo "可用命令:"
	@echo "  build-dev     - 构建开发环境镜像"
	@echo "  build-prod    - 构建生产环境镜像"
	@echo "  run-dev       - 启动开发环境"
	@echo "  run-prod      - 启动生产环境"
	@echo "  test          - 运行测试"
	@echo "  clean         - 清理构建文件（⚠️ 会清理所有未使用的Docker资源）"
	@echo "  clean-safe    - 安全清理（只清理当前项目）"
	@echo "  validate-config - 验证配置文件"
	@echo "  swagger       - 生成Swagger文档"
	@echo "  jenkins-build - Jenkins自动化构建"
	@echo "  deploy        - 部署服务"
	@echo "  health-check  - 健康检查"

# 生成Swagger文档
swagger:
	@echo "📚 生成Swagger文档..."
	swag init -g cmd/server/main.go -o docs

# 构建开发环境镜像
build-dev: swagger
	@echo "🔨 构建开发环境镜像..."
	docker-compose build backend

# 构建生产环境镜像
build-prod:
	@echo "🔨 构建生产环境镜像..."
	docker build -f Dockerfile.prod -t autocodeweb-backend:prod .

# 启动开发环境
run-dev:
	@echo "🚀 启动开发环境..."
	docker-compose up -d

# 启动生产环境
run-prod:
	@echo "🚀 启动生产环境..."
	docker-compose -f docker-compose.prod.yml up -d

# 停止开发环境
stop-dev:
	@echo "🛑 停止开发环境..."
	docker-compose down

# 停止生产环境
stop-prod:
	@echo "🛑 停止生产环境..."
	docker-compose -f docker-compose.prod.yml down

# 查看日志
logs-dev:
	docker-compose logs -f backend

logs-prod:
	docker-compose -f docker-compose.prod.yml logs -f backend

# 验证配置
validate-config:
	@echo "🔍 验证开发环境配置..."
	APP_ENVIRONMENT=development go run cmd/server/main.go --validate-only
	@echo "🔍 验证生产环境配置..."
	APP_ENVIRONMENT=production go run cmd/server/main.go --validate-only

# 清理
clean:
	@echo "🧹 清理构建文件..."
	docker-compose down -v
	docker system prune -f
	docker image prune -f

# 安全清理（只清理当前项目）
clean-safe:
	@echo "🧹 安全清理当前项目..."
	docker-compose down -v
	@echo "⚠️  注意：只清理了当前项目的数据卷和容器"
	@echo "💡 如需清理镜像，请手动执行: docker rmi backend-backend"

# 测试
test:
	@echo "🧪 运行测试..."
	go test ./...

# 格式化代码
fmt:
	@echo "✨ 格式化代码..."
	go fmt ./...

# 代码检查
lint:
	@echo "🔍 代码检查..."
	golangci-lint run

# 构建二进制文件
build-bin:
	@echo "🔨 构建二进制文件..."
	go build -o bin/server ./cmd/server

# 运行二进制文件
run-bin:
	@echo "🚀 运行二进制文件..."
	./bin/server

# Jenkins自动化构建
jenkins-build:
	@echo "🔧 Jenkins自动化构建..."
	@echo "用法: make jenkins-build ENV=dev TAG=v1.0.0 PUSH=true"
	@if [ "$(ENV)" = "" ]; then \
		echo "错误: 请指定环境 (ENV=dev 或 ENV=prod)"; \
		exit 1; \
	fi
	@chmod +x scripts/jenkins-build.sh
	@./scripts/jenkins-build.sh -e $(ENV) -t $(TAG) $(if $(PUSH),-p,)

# 部署服务
deploy:
	@echo "🚀 部署服务..."
	@echo "用法: make deploy ENV=dev TAG=latest FORCE=false"
	@if [ "$(ENV)" = "" ]; then \
		echo "错误: 请指定环境 (ENV=dev 或 ENV=prod)"; \
		exit 1; \
	fi
	@chmod +x scripts/deploy.sh
	@./scripts/deploy.sh -e $(ENV) -t $(TAG) $(if $(FORCE),-f,)

# 健康检查
health-check:
	@echo "🏥 执行健康检查..."
	@if [ "$(ENV)" = "prod" ]; then \
		docker-compose -f docker-compose.prod.yml ps; \
	else \
		docker-compose ps; \
	fi
	@echo "健康检查端点:"
	@if [ "$(ENV)" = "prod" ]; then \
		echo "  - http://localhost:8080/api/v1/health"; \
		echo "  - http://localhost:8080/api/v1/cache/health"; \
	else \
		echo "  - http://localhost:8098/api/v1/health"; \
		echo "  - http://localhost:8098/api/v1/cache/health"; \
	fi
