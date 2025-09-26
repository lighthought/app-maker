## Development

This folder contains a TypeScript-based Express server. To run it, ensure you have a `package.json` and TypeScript config. If missing, create them and install dependencies:

```
pnpm install
pnpm ts-node src/server.ts
```

Environment variables:

- `PORT` (default 3001)
- `REDIS_URL` (optional; if absent, queues run in fallback inline mode)
- `CLAUDE_CLI_PATH` (default `claude`)
- `NPM_PATH` (default `npm`)
- `GIT_PATH` (default `git`)
- `BACKEND_API_URL` (default `http://localhost:8080`)

# Agents Server - Agent协作服务

## 服务定位

Agents Server 专注于多Agent协作处理，**不负责项目创建**。项目在进入 agents-server 时已经是完整可运行的状态（前后端都已存在且可本地编译通过）。

## 核心功能

- **PM Agent**: 基于现有项目生成PRD文档
- **UX Agent**: 生成UX设计规范和设计系统
- **Architect Agent**: 设计系统架构和技术选型
- **PO Agent**: 创建Epic和Story规划
- **Dev Agent**: 实现具体功能代码

## 工作流程

```
现有项目 → Backend API → Agents Server → Agent处理 → 生成文档/代码 → 更新项目
```

## 前置条件

1. **Node.js**: 版本 >= 18.0.0
2. **pnpm**: 包管理器
3. **Redis**: 后端容器已提供，端口6379
4. **Backend API**: 运行在端口8080
5. **现有项目**: 项目已创建且可运行

## 快速启动

### 方法1: 使用启动脚本（推荐）

```batch
# 双击运行
start.bat
```

### 方法2: 手动启动

```batch
# 1. 安装依赖
pnpm install

# 2. 构建项目
pnpm build

# 3. 启动服务
pnpm start
```

## 开发模式

```batch
# 开发模式（热重载）
pnpm dev
```

## 服务验证

启动成功后，可以通过以下方式验证服务：

1. **健康检查**: http://localhost:3001/api/v1/agents/health
2. **根路径**: http://localhost:3001/
3. **API文档**: http://localhost:3001/api/v1/agents/

## 测试API

### 执行Agent任务
```bash
curl -X POST http://localhost:3001/api/v1/agents/execute \
  -H "Content-Type: application/json" \
  -d '{
    "projectId": "existing-project-001",
    "userId": "user-001",
    "agentType": "pm",
    "stage": "prd_generating",
    "context": {
      "projectPath": "F:/app-maker/app_data/projects/existing-project-001",
      "projectName": "My Existing Project",
      "artifacts": [],
      "stageInput": {
        "requirements": "基于现有项目生成PRD文档"
      }
    },
    "parameters": {}
  }'
```

### 同步执行（立即返回结果）
```bash
curl -X POST http://localhost:3001/api/v1/agents/execute-sync \
  -H "Content-Type: application/json" \
  -d '{
    "projectId": "<project-id>",
    "userId": "<user-id>",
    "agentType": "pm",              # pm | ux | architect | po | dev
    "stage": "prd_generating",      # prd_generating | ux_defining | arch_designing | data_modeling | api_defining | epic_planning | story_developing | bug_fixing | testing | packaging
    "context": {
      "projectPath": "/abs/path/to/project",
      "stageInput": { "requirements": "..." }
    },
    "parameters": {}
  }'
```

### 产物路径约定（每步都会 Git 提交并推送）
- PRD → `docs/prd.md`
- UX Spec → `docs/ux-spec.md`
- Architecture → `docs/architecture.md`
- Epics → `docs/epics/`
- Stories → `docs/stories/`

> 首次推送时会自动创建 `.gitlab-ci.yml`，以触发 GitLab Runner。

### 获取队列状态
```bash
curl http://localhost:3001/api/v1/agents/queues/pm/stats
```

## 目录结构

```
agents-server/
├── src/
│   ├── controllers/     # Agent控制器
│   ├── services/        # 核心服务
│   ├── queues/          # 任务队列
│   ├── models/          # 数据模型
│   ├── utils/           # 工具函数
│   ├── routes/          # 路由
│   ├── config/          # 配置
│   ├── app.ts           # 应用入口
│   └── server.ts        # 服务器启动
├── templates/           # 模板文件
├── package.json
├── tsconfig.json
├── .env.example
└── start.bat           # 启动脚本
```

## 环境变量

复制 `.env.example` 为 `.env` 并根据需要修改：

```env
NODE_ENV=development
PORT=3001
REDIS_URL=redis://localhost:6379
BACKEND_API_URL=http://localhost:8080
PROJECT_DATA_PATH=F:/app-maker/app_data/projects
LOG_LEVEL=info
GITLAB_URL=http://gitlab.local
GITLAB_TOKEN=xxxxxxxx
GITLAB_SSH_KEY_PATH=~/.ssh/id_rsa
AGENTS_SERVER_LOG=./logs/agents-server.log
```

## 故障排除

### 1. Redis连接失败
- 确保后端容器正在运行
- 检查Redis端口6379是否可访问

### 2. 端口占用
- 修改 `.env` 文件中的 `PORT` 变量
- 或停止占用3001端口的其他服务

### 3. 依赖安装失败
- 确保网络连接正常
- 尝试使用 `npm install` 替代 `pnpm install`

### 4. 构建失败
- 检查TypeScript语法错误
- 确保所有依赖都已安装

## 重构完成

✅ **已完成的重构**：
1. **删除templates目录** - 不再管理模板文件
2. **TemplateService → DocumentService** - 专注于读取现有项目文档
3. **更新PM控制器** - 基于现有项目文档生成输出
4. **统一文档结构** - 与后端template.zip保持一致

## 下一步

服务启动成功后，可以：

1. **安装依赖包** - 创建package.json并安装所需依赖
2. **实现其他Agent控制器** - UX、Architect、PO、Dev
3. **完善文档读取逻辑** - 基于现有项目文档结构
4. **集成cursor.cli** - 进行代码生成和AI协作
5. **添加单元测试** - 确保Agent协作流程正确
6. **优化文档输出** - 按固定提示词生成结构化内容