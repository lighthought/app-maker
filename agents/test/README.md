# Agent 模块本地调试测试

这个测试套件允许你在本地调试 Agent 模块，而不需要每次都通过前端创建新项目。

## 🚀 快速开始

### 1. 启动 Agent 服务

确保 Agent 服务正在运行：

```bash
# 在项目根目录
cd agents
go run cmd/server/main.go
```

默认情况下，Agent 服务会在 `http://localhost:8088` 启动。

### 2. 运行测试

```bash
cd agents/test

# 运行快速测试（推荐开始使用）
make test-quick

# 运行完整流程测试
make test-all

# 运行单个步骤测试（用于调试特定功能）
make test-single
```

## 📋 测试项目信息

- **测试项目 GUID**: `b0f152887703419e8a39d9718b024f7f`
- **Agent 服务地址**: `http://localhost:8088`
- **项目需求**: 在线书店管理系统（包含用户管理、图书管理、订单管理等）

## 🔧 测试配置

你可以通过环境变量自定义配置：

```bash
# 使用不同的 Agent 服务地址
AGENT_URL=http://localhost:9090 make test-quick

# 使用不同的项目 GUID
PROJECT_GUID=your-project-guid make test-quick
```

## 📝 测试流程

测试按照 `project_stage_service.go` 中定义的开发流程执行：

### 完整流程 (`TestCompleteProjectDevelopment`)

1. **健康检查** - 验证 Agent 服务可用性
2. **项目环境准备** - 设置项目环境和 BMad Method
3. **分析项目概览** - 生成项目简介和市场研究
4. **生成PRD文档** - 创建产品需求文档
5. **定义UX标准** - 生成用户体验规范
6. **设计系统架构** - 设计技术架构
7. **定义数据模型** - 设计数据库结构
8. **定义API接口** - 设计 API 规范
9. **划分Epic和Story** - 创建开发任务
10. **开发Story功能** - 实现具体功能
11. **修复开发问题** - 处理 Bug 和问题
12. **执行自动测试** - 运行项目测试
13. **打包部署项目** - 构建和部署

### 快速流程 (`TestQuickFlow`)

只包含前4个关键步骤，适合快速验证基本功能。

### 单步测试 (`TestSingleStep`)

可以修改代码来测试特定的单个步骤，适合调试特定功能。

## 🛠 开发和调试

### 修改测试项目需求

编辑 `project_test.go` 中的 `testRequirements` 常量来测试不同的项目需求。

### 调试特定步骤

1. 修改 `TestSingleStep` 函数中的测试步骤
2. 运行 `make test-single`

### 添加新的测试步骤

1. 在 `project_test.go` 中添加新的测试函数
2. 将函数添加到相应的测试流程中

## 📊 测试输出

测试会输出详细的执行日志，包括：

- 每个步骤的开始和完成时间
- Agent 服务的响应内容
- 任何错误或失败信息

示例输出：

```
=== 开始执行: 1. 健康检查 ===
2025/09/29 12:30:00 健康检查成功: {"status": "ok"}
=== 完成执行: 1. 健康检查 ===

=== 开始执行: 2. 项目环境准备 ===
2025/09/29 12:30:02 项目环境准备成功: {"message": "项目环境准备完成"}
=== 完成执行: 2. 项目环境准备 ===
```

## 🚨 故障排除

### Agent 服务未启动

```bash
# 检查服务状态
make check-agent

# 如果服务未运行，启动服务
cd ../
go run cmd/server/main.go
```

### 测试超时

如果某个步骤执行时间过长，可以：

1. 增加测试超时时间
2. 检查 Agent 服务日志
3. 使用单步测试调试特定问题

### 项目 GUID 不存在

确保使用的项目 GUID 在系统中存在，或者修改为实际存在的项目 GUID。

## 💡 最佳实践

1. **从快速测试开始** - 先运行 `make test-quick` 验证基本功能
2. **逐步调试** - 使用单步测试调试特定问题
3. **查看日志** - 注意 Agent 服务的日志输出
4. **适当等待** - 每个步骤之间有2秒间隔，避免请求过快
5. **修改需求** - 根据需要修改测试项目需求来测试不同场景

## 🔗 相关文件

- `project_test.go` - 主要测试文件
- `Makefile` - 测试命令定义
- `../internal/api/handlers/` - Agent 处理器实现
- `../../shared-models/client/agent_client.go` - Agent 客户端
- `../../shared-models/agent/requests.go` - 请求类型定义
