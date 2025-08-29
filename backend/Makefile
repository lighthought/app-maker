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
	@echo "  clean         - 清理构建文件"
	@echo "  validate-config - 验证配置文件"

# 构建开发环境镜像
build-dev:
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
