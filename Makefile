.PHONY: help build build-dev build-prod run-dev run-prod test clean validate-config network-create network-check network-clean external-services docker-check docker-start docker-stop docker-status docker-ensure docker-service-start docker-service-stop docker-service-restart docker-install-check print-ssh-pub

# 默认目标
help:
	@echo "AutoCodeWeb Full-Stack Application Build Tool"
	@echo "=========================================="
	@echo "Available Commands:"
	@echo "  docker-check   - Check Docker Desktop status"
	@echo "  docker-start   - Start Docker Desktop (Windows)"
	@echo "  docker-stop    - Stop Docker Desktop (Windows)"
	@echo "  docker-status  - Show Docker service status"
	@echo "  docker-ensure  - Ensure Docker Desktop is running (auto-start if needed)"
	@echo "  docker-service-start/stop/restart - Manage Docker Windows service"
	@echo "  docker-install-check - Check Docker Desktop installation"
	@echo "  docker-image-pull - Pull Docker images"
	@echo "  network-create - Create Docker network (app-maker-network)"
	@echo "  network-check  - Check if Docker network exists"
	@echo "  external-services - Show external services configuration"
	@echo "  build-dev     - Build local development environment (Go + Node.js)"
	@echo "  build-dev-docker - Build development environment images (Docker)"
	@echo "  build-prod    - Build production environment images"
	@echo "  run-dev       - Start hybrid dev environment (infra Docker + apps local)"
	@echo "  run-dev-docker - Start development environment (Docker)"
	@echo "  run-prod      - Start production environment"
	@echo "  stop-dev      - Stop infrastructure services (Docker)"
	@echo "  stop-dev-local - Stop application services (local)"
	@echo "  stop-dev-all  - Stop all development services"
	@echo "  restart-backend-local - Restart local backend server"
	@echo "  restart-frontend-local - Restart local frontend server"
	@echo "  status-local  - Show local development environment status"
	@echo "  start-infrastructure-services - Start infrastructure services (Docker)"
	@echo "  start-local-services - Start local services (backend and frontend)"
	@echo "  start-backend-local - Start only backend server locally"
	@echo "  start-frontend-local - Start only frontend server locally"
	@echo "  test          - Run tests"
	@echo "  clean         - Clean build files (Will clean all unused Docker resources)"
	@echo "  clean-safe    - Safe cleanup (only current project)"
	@echo "  validate-config - Validate configuration files"
	@echo "  swagger       - Generate Swagger documentation"
	@echo "  jenkins-build - Jenkins automated build"
	@echo "  deploy        - Deploy services"
	@echo "  health-check  - Health check"
	@echo "  logs-dev      - View development environment logs"
	@echo "  logs-prod     - View production environment logs"
	@echo "  restart-front-dev - Restart frontend development environment (rebuild)"

# Docker Desktop 检查和启动功能
docker-check:
	@echo "Checking Docker Desktop status..."
	@docker version >nul 2>&1 && echo "[OK] Docker Desktop is running" || echo "[ERROR] Docker Desktop is not running"
	@docker info >nul 2>&1 && echo "[OK] Docker daemon is accessible" || echo "[ERROR] Docker daemon is not accessible"

# 智能 Docker 检查 - 如果 Docker 不可用则自动启动
docker-ensure:
	@echo "Ensuring Docker Desktop is available..."
	@docker version >nul 2>&1 && echo "[OK] Docker Desktop is already running" || ( \
		echo "[WARNING] Docker Desktop is not running, attempting to start..." && \
		$(MAKE) docker-start \
	)

docker-start:
	@echo "Starting Docker Desktop..."
	@echo "Checking if Docker Desktop is already running..."
	@docker version >nul 2>&1 && (echo "[OK] Docker Desktop is already running") || ( \
		echo "[INFO] Starting Docker Desktop..." && \
		( \
			if exist "C:\Program Files\Docker\Docker\Docker Desktop.exe" ( \
				start "" "C:\Program Files\Docker\Docker\Docker Desktop.exe" \
			) else if exist "C:\Program Files (x86)\Docker\Docker\Docker Desktop.exe" ( \
				start "" "C:\Program Files (x86)\Docker\Docker\Docker Desktop.exe" \
			) else ( \
				echo "[ERROR] Docker Desktop not found in standard locations" && \
				echo "Please ensure Docker Desktop is installed" && \
				exit /b 1 \
			) \
		) && \
		echo "[INFO] Waiting for Docker Desktop to start..." && \
		timeout /t 15 /nobreak >nul && \
		echo "[INFO] Checking Docker status..." && \
		docker version >nul 2>&1 && echo "[OK] Docker Desktop started successfully!" || echo "[WARNING] Docker Desktop may still be starting. Please wait a moment and try again." \
	)

docker-stop:
	@echo "Stopping Docker Desktop..."
	@taskkill /F /IM "Docker Desktop.exe" >nul 2>&1 && echo "[OK] Docker Desktop stopped" || echo "[WARNING] Docker Desktop was not running or could not be stopped"

docker-status:
	@echo "Docker Service Status:"
	@echo "====================="
	@echo "Docker Desktop Process:"
	@tasklist /FI "IMAGENAME eq Docker Desktop.exe" 2>nul | findstr "Docker Desktop.exe" >nul && echo "[OK] Docker Desktop process is running" || echo "[ERROR] Docker Desktop process is not running"
	@echo ""
	@echo "Docker Daemon Status:"
	@docker version >nul 2>&1 && echo "[OK] Docker daemon is accessible" || echo "[ERROR] Docker daemon is not accessible"
	@echo ""
	@echo "Docker Service Status:"
	@sc query "com.docker.service" >nul 2>&1 && ( \
		sc query "com.docker.service" | findstr "RUNNING" >nul && echo "[OK] Docker service is running" || echo "[WARNING] Docker service exists but not running" \
	) || echo "[ERROR] Docker service not found"
	@echo ""
	@echo "Docker Desktop Installation:"
	@if exist "C:\Program Files\Docker\Docker\Docker Desktop.exe" ( \
		echo "[OK] Docker Desktop is installed" \
	) else ( \
		echo "[ERROR] Docker Desktop not found at default location" \
	)

# Windows 特定的 Docker 服务管理
docker-service-start:
	@echo "Starting Docker services..."
	@net start "com.docker.service" >nul 2>&1 && echo "[OK] Docker service started" || echo "[WARNING] Docker service start failed or already running"

docker-service-stop:
	@echo "Stopping Docker services..."
	@net stop "com.docker.service" >nul 2>&1 && echo "[OK] Docker service stopped" || echo "[WARNING] Docker service stop failed or not running"

docker-service-restart:
	@echo "Restarting Docker services..."
	@net stop "com.docker.service" >nul 2>&1
	@timeout /t 3 /nobreak >nul
	@net start "com.docker.service" >nul 2>&1 && echo "[OK] Docker service restarted" || echo "[WARNING] Docker service restart failed"

docker-image-pull:
	@echo "Checking Docker images..."
	@echo "Checking golang:1.24-alpine..."
	@docker image inspect golang:1.24-alpine >nul 2>&1 || docker pull golang:1.24-alpine
	@echo "Checking alpine:latest..."
	@docker image inspect alpine:latest >nul 2>&1 || docker pull alpine:latest
	@echo "Checking node:18-alpine..."
	@docker image inspect node:18-alpine >nul 2>&1 || docker pull node:18-alpine
	@echo "Checking nginx:alpine..."
	@docker image inspect nginx:alpine >nul 2>&1 || docker pull nginx:alpine
	@echo "Docker images check completed."

# 检查 Docker Desktop 安装路径
docker-install-check:
	@echo "Checking Docker Desktop installation..."
	@if exist "C:\Program Files\Docker\Docker\Docker Desktop.exe" ( \
		echo "[OK] Docker Desktop found at: C:\Program Files\Docker\Docker\Docker Desktop.exe" \
	) else if exist "C:\Program Files (x86)\Docker\Docker\Docker Desktop.exe" ( \
		echo "[OK] Docker Desktop found at: C:\Program Files (x86)\Docker\Docker\Docker Desktop.exe" \
	) else ( \
		echo "[ERROR] Docker Desktop not found in standard locations" && \
		echo "Please ensure Docker Desktop is installed" \
	)

# 检查Docker网络是否存在
network-check:
	@echo "Checking Docker network 'app-maker-network'..."
	@docker network ls --format "table {{.Name}}" | findstr "app-maker-network" >nul 2>&1 && echo "Network 'app-maker-network' already exists" || echo "Network 'app-maker-network' does not exist"

# 创建Docker网络
network-create:
	@echo "Creating Docker network 'app-maker-network'..."
	@docker network ls --format "table {{.Name}}" | findstr "app-maker-network" >nul 2>&1 && echo "Network 'app-maker-network' already exists, skipping creation" || (echo "Creating network 'app-maker-network'..." && docker network create app-maker-network && echo "Network 'app-maker-network' created successfully")

# 显示外部服务配置
external-services:
	@echo "External Services Configuration:"
	@echo "================================"
	@echo "Current external services in traefik-external.yml:"
	@if exist traefik-external.yml ( \
		echo "Ollama AI Service: http://chat.app-maker.localhost -> localhost:11434" && \
		echo "Edit traefik-external.yml to add more services" && \
		echo "Template available: traefik-external-template.yml" \
	) else ( \
		echo "traefik-external.yml not found" \
	)
	@echo ""
	@echo "To add a new service:"
	@echo "1. Edit traefik-external.yml"
	@echo "2. Add router and service configuration"
	@echo "3. Restart Traefik: docker-compose restart traefik"

# 生成Swagger文档（Docker方式）
swagger:
	@echo "Generating Swagger documentation..."
	cd backend && swag init -g cmd/server/main.go -o docs --parseDependency --parseInternal

# 检查本地开发环境
check-local-dev-env:
	@echo "Checking local development environment..."
	@echo "Checking Go installation..."
	@go version >nul 2>&1 && echo "[OK] Go is installed" || ( \
		echo "[ERROR] Go is not installed. Please install Go 1.24+" && \
		exit /b 1 \
	)
	@echo "Checking Node.js installation..."
	@node --version >nul 2>&1 && echo "[OK] Node.js is installed" || ( \
		echo "[ERROR] Node.js is not installed. Please install Node.js 18+" && \
		exit /b 1 \
	)
	@echo "Checking pnpm installation..."
	@pnpm --version >nul 2>&1 && echo "[OK] pnpm is installed" || ( \
		echo "[WARNING] pnpm not found, installing..." && \
		npm install -g pnpm \
	)
	@echo "Checking PostgreSQL connection..."
	@echo "[INFO] Please ensure PostgreSQL is running on localhost:5432"
	@echo "Checking Redis connection..."
	@echo "[INFO] Please ensure Redis is running on localhost:6379"
	@echo "Local development environment check completed!"

# 生成Swagger文档（本地方式）
swagger-local:
	@echo "Generating Swagger documentation locally..."
	@echo "Checking if swag is installed..."
	@swag version >nul 2>&1 && echo "[OK] swag is installed" || ( \
		echo "[WARNING] swag not found, installing..." && \
		go install github.com/swaggo/swag/cmd/swag@latest \
	)
	cd backend && swag init -g cmd/server/main.go -o docs --parseDependency --parseInternal

# 构建开发环境镜像（Docker方式）
build-dev-docker: docker-ensure network-create swagger docker-image-pull
	@echo "Building development environment images..."
	docker-compose build

# 本地开发环境构建（无需Docker）
build-dev: check-local-dev-env swagger-local
	@echo "Building local development environment..."
	@echo "Backend: Go modules and dependencies"
	@echo "Frontend: Node.js dependencies"
	@echo "Building backend dependencies..."
	cd backend && go mod download && go mod tidy
	@echo "Building frontend dependencies..."
	cd frontend && pnpm install
	@echo "Local development environment ready!"

# 构建生产环境镜像
build-prod: docker-ensure network-create swagger docker-image-pull
	@echo "Building production environment images..."
	docker-compose -f docker-compose.prod.yml build

# 启动开发环境（Docker方式）
run-dev-docker: docker-ensure network-create
	@echo "Starting development environment with Docker..."
	@echo "Frontend: http://localhost:3000 (Direct) or http://app-maker.localhost (via Traefik)"
	@echo "Backend API: http://localhost:8098 (Direct) or http://api.app-maker.localhost (via Traefik)"
	@echo "Traefik Dashboard: http://localhost:8080 or http://traefik.app-maker.localhost"
	@echo "Swagger Docs: http://localhost:8098/swagger/index.html" or http://api.app-maker.localhost/swagger/index.html
	docker-compose up -d
	@$(MAKE) print-ssh-pub

# 启动本地服务（后端和前端）
start-local-services:
	@echo "Starting local services..."
	@echo "Starting backend server in background..."
	@cmd /c "cd backend && go run cmd/server/main.go" &
	@timeout /t 3 /nobreak >nul
	@echo "Starting frontend development server in background..."
	@cmd /c "cd frontend && pnpm dev" &
	@echo "Services started in background"
	@echo "Backend: http://localhost:8088"
	@echo "Frontend: http://localhost:3000"
	@echo "Note: Services are running in background. Use 'make stop-dev-local' to stop them."

# 启动本地后端服务
start-backend-local:
	@echo "Starting backend server locally..."
	cd backend && go run cmd/server/main.go

# 启动本地前端服务
start-frontend-local:
	@echo "Starting frontend development server locally..."
	cd frontend && pnpm dev

# 启动基础设施服务（Docker）
start-infrastructure-services: docker-ensure network-create
	@echo "Starting infrastructure services with Docker..."
	@echo "Starting PostgreSQL, Redis, GitLab, and Traefik..."
	docker-compose up -d postgres redis gitlab traefik
	@echo "Waiting for services to be ready..."
	@timeout /t 10 /nobreak >nul
	@echo "Infrastructure services started!"

# 启动混合开发环境（基础服务Docker + 应用服务本地）
run-dev: docker-ensure network-create start-infrastructure-services check-local-dev-env start-local-services
	@echo "Starting hybrid development environment..."
	@echo "Infrastructure services (Docker):"
	@echo "  PostgreSQL: localhost:5432"
	@echo "  Redis: localhost:6379"
	@echo "  GitLab: localhost:8081"
	@echo "  Traefik: localhost:8080"
	@echo ""
	@echo "Application services (Local):"
	@echo "  Frontend: http://localhost:3000"
	@echo "  Backend API: http://localhost:8088"
	@echo "  Swagger Docs: http://localhost:8088/swagger/index.html"
	@echo "=========================================="
	@echo "Hybrid development environment is running!"
	@echo "Use 'make stop-dev' to stop infrastructure services"
	@echo "Use 'make stop-dev-local' to stop application services"
	@echo "=========================================="

# 启动生产环境
run-prod: docker-ensure network-create
	@echo "Starting production environment..."
	@echo "Frontend: http://localhost (Direct) or http://thought-light.com (via Traefik)"
	@echo "Backend API: http://localhost:8080 (Direct) or http://api.thought-light.com (via Traefik)"
	@echo "Traefik Dashboard: http://localhost:8080 or http://traefik.thought-light.com"
	@echo "Swagger Docs: http://localhost:8080/swagger/index.html" or http://api.thought-light.com/swagger/index.html
	docker-compose -f docker-compose.prod.yml up -d
	@$(MAKE) print-ssh-pub

# 打印SSH公钥
print-ssh-pub:
	@echo "=========================================="
	@echo "Checking GitLab SSH Connection..."
	@echo "=========================================="
	@docker-compose exec -T backend ssh -o StrictHostKeyChecking=no -o ConnectTimeout=10 -T git@gitlab >nul 2>&1 && ( \
		echo "[OK] SSH key is already configured for GitLab!" && \
		echo "GitLab connection is working properly." \
	) || ( \
		echo "[WARNING] SSH key not configured for GitLab" && \
		echo "Please add SSH Keys to gitlab to ensure API has the rights to operate git repositories" && \
		echo "" && \
		echo "Backend container SSH public key:" && \
		echo "----------------------------------" && \
		docker-compose exec -T backend cat /home/appuser/.ssh/id_rsa.pub 2>nul || echo "[WARNING] SSH public key not found in backend container" \
	)
	@echo ""
	@echo "=========================================="

# 停止本地开发环境
stop-dev-local:
	@echo "Stopping local development environment..."
	@taskkill /F /IM "go.exe" >nul 2>&1 && echo "[OK] Backend stopped" || echo "[INFO] Backend was not running"
	@taskkill /F /IM "node.exe" >nul 2>&1 && echo "[OK] Frontend stopped" || echo "[INFO] Frontend was not running"
	@echo "Local development environment stopped!"

# 重启本地后端
restart-backend-local:
	@echo "Restarting backend server..."
	@taskkill /F /IM "go.exe" >nul 2>&1
	@timeout /t 2 /nobreak >nul
	@cmd /c "cd backend && start cmd /k go run cmd/server/main.go"
	@echo "Backend server restarted!"

# 重启本地前端
restart-frontend-local:
	@echo "Restarting frontend development server..."
	@taskkill /F /IM "node.exe" >nul 2>&1
	@timeout /t 2 /nobreak >nul
	@cmd /c "cd frontend && start cmd /k pnpm dev"
	@echo "Frontend server restarted!"

# 查看本地服务状态
status-local:
	@echo "Local Development Environment Status:"
	@echo "===================================="
	@echo "Backend (Go):"
	@tasklist /FI "IMAGENAME eq go.exe" 2>nul | findstr "go.exe" >nul && echo "[RUNNING] Backend server is running" || echo "[STOPPED] Backend server is not running"
	@echo "Frontend (Node.js):"
	@tasklist /FI "IMAGENAME eq node.exe" 2>nul | findstr "node.exe" >nul && echo "[RUNNING] Frontend server is running" || echo "[STOPPED] Frontend server is not running"
	@echo "===================================="

# 停止开发环境（只停止基础设施服务）
stop-dev:
	@echo "Stopping infrastructure services..."
	docker-compose stop postgres redis gitlab traefik
	@echo "Infrastructure services stopped!"
	@echo "Note: Local application services are still running."
	@echo "Use 'make stop-dev-local' to stop application services."

# 停止所有开发环境服务
stop-dev-all:
	@echo "Stopping all development environment services..."
	@$(MAKE) stop-dev-local
	@$(MAKE) stop-dev
	@echo "All development services stopped!"

# 停止生产环境
stop-prod:
	@echo "Stopping production environment..."
	docker-compose -f docker-compose.prod.yml down

# 查看日志
logs-dev:
	@echo "Development environment logs..."
	docker-compose logs -f

logs-prod:
	@echo "Production environment logs..."
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
	@echo "Validating development environment configuration..."
	cd backend && APP_ENVIRONMENT=development go run cmd/server/main.go --validate-only
	@echo "Validating production environment configuration..."
	cd backend && APP_ENVIRONMENT=production go run cmd/server/main.go --validate-only

# 清理网络
network-clean:
	@echo "Cleaning Docker network 'app-maker-network'..."
	@docker network ls --format "table {{.Name}}" | findstr "app-maker-network" >nul 2>&1 && (echo "Removing network 'app-maker-network'..." && docker network rm app-maker-network && echo "Network 'app-maker-network' removed successfully") || echo "Network 'app-maker-network' does not exist, nothing to clean"

# 清理
clean: network-clean
	@echo "Cleaning build files..."
	docker-compose down -v
	docker-compose -f docker-compose.prod.yml down -v
	docker system prune -f
	docker image prune -f

# 安全清理（只清理当前项目）
clean-safe:
	@echo "Safe cleanup for current project..."
	docker-compose down -v
	@echo "Note: Only cleaned current project data volumes and containers"
	@echo "To clean images, manually run: docker rmi app-maker-backend app-maker-frontend"

# 测试
test:
	@echo "Running backend tests..."
	cd backend && go test ./...
	@echo "Running frontend tests..."
	cd frontend && pnpm run test

# 格式化代码
fmt:
	@echo "Formatting backend code..."
	cd backend && go fmt ./...
	@echo "Formatting frontend code..."
	cd frontend && pnpm run format

# 代码检查
lint:
	@echo "Backend code linting..."
	cd backend && golangci-lint run
	@echo "Frontend code linting..."
	cd frontend && pnpm run lint

# 构建二进制文件
build-bin:
	@echo "Building backend binary..."
	cd backend && go build -o bin/server ./cmd/server

# 运行二进制文件
run-bin:
	@echo "Running backend binary..."
	cd backend && ./bin/server


# 健康检查
health-check:
	@echo "Performing health check..."
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
	@echo "Restarting development environment..."
	docker-compose restart

restart-prod:
	@echo "Restarting production environment..."
	docker-compose -f docker-compose.prod.yml restart

# 重启前端开发环境（重新编译）
restart-front-dev:
	@echo "Restarting frontend development environment..."
	@echo "1.Stopping frontend container..."
	docker-compose stop frontend
	@echo "2.Removing frontend container..."
	docker-compose rm -f frontend
	@echo "3.Rebuilding frontend image..."
	docker-compose build frontend
	@echo "4.Starting frontend container..."
	docker-compose up -d frontend
	@echo "Frontend restart completed!"
	@echo "Frontend URL: http://localhost:3000"

# 重启前端生产环境（重新编译）
restart-front-prod:
	@echo "Restarting frontend production environment..."
	@echo "1.Stopping frontend container..."
	@echo "Frontend URL: http://localhost:3000"

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
	@echo "Database migration..."
	cd backend && go run cmd/server/main.go --migrate

db-seed:
	@echo "Database seed data..."
	cd backend && go run cmd/server/main.go --seed

# 缓存操作
cache-clear:
	@echo "Clearing cache..."
	docker-compose exec redis redis-cli FLUSHALL

cache-info:
	@echo "Cache information..."
	docker-compose exec redis redis-cli INFO