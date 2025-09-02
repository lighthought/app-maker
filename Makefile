.PHONY: help build build-dev build-prod run-dev run-prod test clean validate-config

# 默认目标
help:
	@echo "AutoCodeWeb Full-Stack Application Build Tool"
	@echo "=========================================="
	@echo "Available Commands:"
	@echo "  build-dev     - Build development environment images"
	@echo "  build-prod    - Build production environment images"
	@echo "  run-dev       - Start development environment"
	@echo "  run-prod      - Start production environment"
	@echo "  stop-dev      - Stop development environment"
	@echo "  stop-prod     - Stop production environment"
	@echo "  test          - Run tests"
	@echo "  clean         - Clean build files (⚠️  Will clean all unused Docker resources)"
	@echo "  clean-safe    - Safe cleanup (only current project)"
	@echo "  validate-config - Validate configuration files"
	@echo "  swagger       - Generate Swagger documentation"
	@echo "  jenkins-build - Jenkins automated build"
	@echo "  deploy        - Deploy services"
	@echo "  health-check  - Health check"
	@echo "  logs-dev      - View development environment logs"
	@echo "  logs-prod     - View production environment logs"
	@echo "  restart-front-dev - Restart frontend development environment (rebuild)"

# 生成Swagger文档
swagger:
	@echo "📚 Generating Swagger documentation..."
	cd backend && swag init -g cmd/server/main.go -o docs

# 构建开发环境镜像
build-dev: swagger
	@echo "🔨 Building development environment images..."
	docker-compose build

# 构建生产环境镜像
build-prod: swagger
	@echo "🔨 Building production environment images..."
	docker-compose -f docker-compose.prod.yml build

# 启动开发环境
run-dev:
	@echo "🚀 Starting development environment..."
	@echo "Frontend: http://localhost:3000"
	@echo "Backend API: http://localhost:8098"
	@echo "Swagger Docs: http://localhost:8098/swagger/index.html"
	docker-compose up -d

# 启动生产环境
run-prod:
	@echo "🚀 Starting production environment..."
	@echo "Frontend: http://localhost"
	@echo "Backend API: http://localhost:8080"
	@echo "Swagger Docs: http://localhost:8080/swagger/index.html"
	docker-compose -f docker-compose.prod.yml up -d

# 停止开发环境
stop-dev:
	@echo "🛑 Stopping development environment..."
	docker-compose down

# 停止生产环境
stop-prod:
	@echo "🛑 Stopping production environment..."
	docker-compose -f docker-compose.prod.yml down

# 查看日志
logs-dev:
	@echo "📋 Development environment logs..."
	docker-compose logs -f

logs-prod:
	@echo "📋 Production environment logs..."
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
	@echo "🔍 Validating development environment configuration..."
	cd backend && APP_ENVIRONMENT=development go run cmd/server/main.go --validate-only
	@echo "🔍 Validating production environment configuration..."
	cd backend && APP_ENVIRONMENT=production go run cmd/server/main.go --validate-only

# 清理
clean:
	@echo "🧹 Cleaning build files..."
	docker-compose down -v
	docker-compose -f docker-compose.prod.yml down -v
	docker system prune -f
	docker image prune -f

# 安全清理（只清理当前项目）
clean-safe:
	@echo "🧹 Safe cleanup for current project..."
	docker-compose down -v
	@echo "⚠️  Note: Only cleaned current project data volumes and containers"
	@echo "💡 To clean images, manually run: docker rmi app-maker-backend app-maker-frontend"

# 测试
test:
	@echo "🧪 Running backend tests..."
	cd backend && go test ./...
	@echo "🧪 Running frontend tests..."
	cd frontend && pnpm run test

# 格式化代码
fmt:
	@echo "✨ Formatting backend code..."
	cd backend && go fmt ./...
	@echo "✨ Formatting frontend code..."
	cd frontend && pnpm run format

# 代码检查
lint:
	@echo "🔍 Backend code linting..."
	cd backend && golangci-lint run
	@echo "🔍 Frontend code linting..."
	cd frontend && pnpm run lint

# 构建二进制文件
build-bin:
	@echo "🔨 Building backend binary..."
	cd backend && go build -o bin/server ./cmd/server

# 运行二进制文件
run-bin:
	@echo "🚀 Running backend binary..."
	cd backend && ./bin/server

# Jenkins自动化构建
jenkins-build:
	@echo "🔧 Jenkins automated build..."
	@echo "Usage: make jenkins-build ENV=dev TAG=v1.0.0 PUSH=true"
	@if [ "$(ENV)" = "" ]; then \
		echo "Error: Please specify environment (ENV=dev or ENV=prod)"; \
		exit 1; \
	fi
	@chmod +x backend/scripts/jenkins-build.sh
	@./backend/scripts/jenkins-build.sh -e $(ENV) -t $(TAG) $(if $(PUSH),-p,)

# 部署服务
deploy:
	@echo "🚀 Deploying services..."
	@echo "Usage: make deploy ENV=dev TAG=latest FORCE=false"
	@if [ "$(ENV)" = "" ]; then \
		echo "Error: Please specify environment (ENV=dev or ENV=prod)"; \
		exit 1; \
	fi
	@chmod +x backend/scripts/deploy.sh
	@./backend/scripts/deploy.sh -e $(ENV) -t $(TAG) $(if $(FORCE),-f,)

# 健康检查
health-check:
	@echo "🏥 Performing health check..."
	@if [ "$(ENV)" = "prod" ]; then \
		docker-compose -f docker-compose.prod.yml ps; \
		echo "Health check endpoints:"; \
		echo "  - http://localhost/api/v1/health"; \
		echo "  - http://localhost/api/v1/cache/health"; \
	else \
		docker-compose ps; \
		echo "Health check endpoints:"; \
		echo "  - http://localhost:3000"; \
		echo "  - http://localhost:8098/api/v1/health"; \
		echo "  - http://localhost:8098/api/v1/cache/health"; \
	fi

# 重启服务
restart-dev:
	@echo "🔄 Restarting development environment..."
	docker-compose restart

restart-prod:
	@echo "🔄 Restarting production environment..."
	docker-compose -f docker-compose.prod.yml restart

# 重启前端开发环境（重新编译）
restart-front-dev:
	@echo "🔄 Restarting frontend development environment..."
	@echo "1️⃣ Stopping frontend container..."
	docker-compose stop frontend
	@echo "2️⃣ Removing frontend container..."
	docker-compose rm -f frontend
	@echo "3️⃣ Rebuilding frontend image..."
	docker-compose build frontend
	@echo "4️⃣ Starting frontend container..."
	docker-compose up -d frontend
	@echo "✅ Frontend restart completed!"
	@echo "🌐 Frontend URL: http://localhost:3000"

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
	@echo "🗄️  Database migration..."
	cd backend && go run cmd/server/main.go --migrate

db-seed:
	@echo "🌱 Database seed data..."
	cd backend && go run cmd/server/main.go --seed

# 缓存操作
cache-clear:
	@echo "🗑️  Clearing cache..."
	docker-compose exec redis redis-cli FLUSHALL

cache-info:
	@echo "📊 Cache information..."
	docker-compose exec redis redis-cli INFO