.PHONY: help build build-dev build-prod run-dev run-prod test clean validate-config

# 默认目标
help:
	@echo "AutoCodeWeb 全栈应用构建工具"
	@echo "=========================="
	@echo "可用命令:"
	@echo "  build-dev     - 构建开发环境镜像"
	@echo "  build-prod    - 构建生产环境镜像"
	@echo "  run-dev       - 启动开发环境"
	@echo "  run-prod      - 启动生产环境"
	@echo "  stop-dev      - 停止开发环境"
	@echo "  stop-prod     - 停止生产环境"
	@echo "  test          - 运行测试"
	@echo "  clean         - 清理构建文件（⚠️ 会清理所有未使用的Docker资源）"
	@echo "  clean-safe    - 安全清理（只清理当前项目）"
	@echo "  validate-config - 验证配置文件"
	@echo "  swagger       - 生成Swagger文档"
	@echo "  jenkins-build - Jenkins自动化构建"
	@echo "  deploy        - 部署服务"
	@echo "  health-check  - 健康检查"
	@echo "  logs-dev      - 查看开发环境日志"
	@echo "  logs-prod     - 查看生产环境日志"

# 生成Swagger文档
swagger:
	@echo "📚 生成Swagger文档..."
	cd backend && swag init -g cmd/server/main.go -o docs

# 构建开发环境镜像
build-dev: swagger
	@echo "🔨 构建开发环境镜像..."
	docker-compose build

# 构建生产环境镜像
build-prod: swagger
	@echo "🔨 构建生产环境镜像..."
	docker-compose -f docker-compose.prod.yml build

# 启动开发环境
run-dev:
	@echo "🚀 启动开发环境..."
	@echo "前端: http://localhost:3000"
	@echo "后端API: http://localhost:8098"
	@echo "Swagger文档: http://localhost:8098/swagger/index.html"
	docker-compose up -d

# 启动生产环境
run-prod:
	@echo "🚀 启动生产环境..."
	@echo "前端: http://localhost"
	@echo "后端API: http://localhost:8080"
	@echo "Swagger文档: http://localhost:8080/swagger/index.html"
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
	@echo "📋 开发环境日志..."
	docker-compose logs -f

logs-prod:
	@echo "📋 生产环境日志..."
	docker-compose -f docker-compose.prod.yml logs -f

# 查看前端日志
logs-frontend-dev:
	docker-compose logs -f frontend

logs-frontend-prod:
	docker-compose -f docker-compose.prod.yml logs -f frontend

# 查看后端日志
logs-backend-dev:
	docker-compose logs -f backend

logs-backend-prod:
	docker-compose -f docker-compose.prod.yml logs -f backend

# 验证配置
validate-config:
	@echo "🔍 验证开发环境配置..."
	cd backend && APP_ENVIRONMENT=development go run cmd/server/main.go --validate-only
	@echo "🔍 验证生产环境配置..."
	cd backend && APP_ENVIRONMENT=production go run cmd/server/main.go --validate-only

# 清理
clean:
	@echo "🧹 清理构建文件..."
	docker-compose down -v
	docker-compose -f docker-compose.prod.yml down -v
	docker system prune -f
	docker image prune -f

# 安全清理（只清理当前项目）
clean-safe:
	@echo "🧹 安全清理当前项目..."
	docker-compose down -v
	@echo "⚠️  注意：只清理了当前项目的数据卷和容器"
	@echo "💡 如需清理镜像，请手动执行: docker rmi app-maker-backend app-maker-frontend"

# 测试
test:
	@echo "🧪 运行后端测试..."
	cd backend && go test ./...
	@echo "🧪 运行前端测试..."
	cd frontend && pnpm run test

# 格式化代码
fmt:
	@echo "✨ 格式化后端代码..."
	cd backend && go fmt ./...
	@echo "✨ 格式化前端代码..."
	cd frontend && pnpm run format

# 代码检查
lint:
	@echo "🔍 后端代码检查..."
	cd backend && golangci-lint run
	@echo "🔍 前端代码检查..."
	cd frontend && pnpm run lint

# 构建二进制文件
build-bin:
	@echo "🔨 构建后端二进制文件..."
	cd backend && go build -o bin/server ./cmd/server

# 运行二进制文件
run-bin:
	@echo "🚀 运行后端二进制文件..."
	cd backend && ./bin/server

# Jenkins自动化构建
jenkins-build:
	@echo "🔧 Jenkins自动化构建..."
	@echo "用法: make jenkins-build ENV=dev TAG=v1.0.0 PUSH=true"
	@if [ "$(ENV)" = "" ]; then \
		echo "错误: 请指定环境 (ENV=dev 或 ENV=prod)"; \
		exit 1; \
	fi
	@chmod +x backend/scripts/jenkins-build.sh
	@./backend/scripts/jenkins-build.sh -e $(ENV) -t $(TAG) $(if $(PUSH),-p,)

# 部署服务
deploy:
	@echo "🚀 部署服务..."
	@echo "用法: make deploy ENV=dev TAG=latest FORCE=false"
	@if [ "$(ENV)" = "" ]; then \
		echo "错误: 请指定环境 (ENV=dev 或 ENV=prod)"; \
		exit 1; \
	fi
	@chmod +x backend/scripts/deploy.sh
	@./backend/scripts/deploy.sh -e $(ENV) -t $(TAG) $(if $(FORCE),-f,)

# 健康检查
health-check:
	@echo "🏥 执行健康检查..."
	@if [ "$(ENV)" = "prod" ]; then \
		docker-compose -f docker-compose.prod.yml ps; \
		echo "健康检查端点:"; \
		echo "  - http://localhost/api/v1/health"; \
		echo "  - http://localhost/api/v1/cache/health"; \
	else \
		docker-compose ps; \
		echo "健康检查端点:"; \
		echo "  - http://localhost:3000"; \
		echo "  - http://localhost:8098/api/v1/health"; \
		echo "  - http://localhost:8098/api/v1/cache/health"; \
	fi

# 重启服务
restart-dev:
	@echo "🔄 重启开发环境..."
	docker-compose restart

restart-prod:
	@echo "🔄 重启生产环境..."
	docker-compose -f docker-compose.prod.yml restart

# 进入容器
shell-frontend-dev:
	docker-compose exec frontend sh

shell-backend-dev:
	docker-compose exec backend sh

shell-frontend-prod:
	docker-compose -f docker-compose.prod.yml exec frontend sh

shell-backend-prod:
	docker-compose -f docker-compose.prod.yml exec backend sh

# 数据库操作
db-migrate:
	@echo "🗄️  数据库迁移..."
	cd backend && go run cmd/server/main.go --migrate

db-seed:
	@echo "🌱 数据库种子数据..."
	cd backend && go run cmd/server/main.go --seed

# 缓存操作
cache-clear:
	@echo "🗑️  清理缓存..."
	docker-compose exec redis redis-cli FLUSHALL

cache-info:
	@echo "📊 缓存信息..."
	docker-compose exec redis redis-cli INFO