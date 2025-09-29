# Agent 模块测试 Makefile

# 默认配置
AGENT_URL ?= http://localhost:8088
PROJECT_GUID ?= b0f152887703419e8a39d9718b024f7f

.PHONY: help test-all test-single test-quick clean

help: ## 显示帮助信息
	@echo "Agent 模块测试命令："
	@echo ""
	@echo "  make test-all      - 运行完整的项目开发流程测试"
	@echo "  make test-single   - 运行单个步骤测试（用于调试）"
	@echo "  make test-quick    - 运行快速流程测试（只测试关键步骤）"
	@echo "  make clean         - 清理测试缓存"
	@echo ""
	@echo "环境变量："
	@echo "  AGENT_URL      - Agent 服务地址 (默认: http://localhost:8088)"
	@echo "  PROJECT_GUID   - 测试项目 GUID (默认: b0f152887703419e8a39d9718b024f7f)"

test-all: ## 运行完整的项目开发流程测试
	@echo "=== 运行完整项目开发流程测试 ==="
	@echo "Agent URL: $(AGENT_URL)"
	@echo "Project GUID: $(PROJECT_GUID)"
	@echo ""
	go test -v -run TestCompleteProjectDevelopment -timeout 30m

test-single: ## 运行单个步骤测试
	@echo "=== 运行单个步骤测试 ==="
	@echo "Agent URL: $(AGENT_URL)"
	@echo "Project GUID: $(PROJECT_GUID)"
	@echo ""
	go test -v -run TestSingleStep -timeout 5m

test-quick: ## 运行快速流程测试
	@echo "=== 运行快速流程测试 ==="
	@echo "Agent URL: $(AGENT_URL)"
	@echo "Project GUID: $(PROJECT_GUID)"
	@echo ""
	go test -v -run TestQuickFlow -timeout 15m

clean: ## 清理测试缓存
	@echo "=== 清理测试缓存 ==="
	go clean -testcache

# 检查 Agent 服务是否运行
check-agent: ## 检查 Agent 服务状态
	@echo "=== 检查 Agent 服务状态 ==="
	@curl -f $(AGENT_URL)/api/v1/health || (echo "Agent 服务未运行，请先启动 Agent 服务" && exit 1)
	@echo "Agent 服务运行正常"

# 启动测试前的检查
pre-test: check-agent ## 测试前检查
	@echo "=== 测试环境检查完成 ==="

# 带检查的完整测试
test-all-safe: pre-test test-all ## 安全的完整测试（带环境检查）

# 带检查的快速测试
test-quick-safe: pre-test test-quick ## 安全的快速测试（带环境检查）
