# AutoCodeWeb - 多Agent自动实现APP和网站项目平台

## 🎯 产品概述

**AutoCodeWeb** 是一个基于 **BMad-Method** 的多Agent协作开发平台，通过标准化的Agent协作流程，自动生成APP和网站项目的完整代码。核心理念是"想法即应用"，让用户通过简单的需求描述快速获得完整的应用程序。

### 目标用户
- **产品经理**: 需要快速验证产品想法的专业人士
- **创业者**: 希望快速构建MVP的初创团队  
- **设计师**: 需要将设计转化为可交互产品的创意人员
- **非技术背景用户**: 有想法但缺乏技术实现能力的个人

## 🏗️ 系统架构

### 三层架构设计
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   前端应用      │    │   后端服务      │    │   Agent服务     │
│  (Vue.js 3)     │◄──►│   (Go + Gin)    │◄──►│  (Node.js)      │
│                │    │                │    │                │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   用户界面      │    │   数据存储      │    │   外部工具      │
│  Naive UI       │    │  PostgreSQL     │    │  Cursor CLI     │
│  TypeScript     │    │  Redis          │    │  Ollama AI      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### 核心组件
- **Frontend**: Vue.js 3 + TypeScript + Naive UI + Pinia
- **Backend**: Go + Gin + GORM + PostgreSQL + Redis
- **Agents Server**: Node.js + Express + Bull Queue + Socket.io
- **Proxy**: Traefik 反向代理，统一入口管理
- **AI**: Ollama 本地AI模型服务

## 🤖 BMad-Method Agent体系

### Agent角色分工
1. **PM Agent（产品经理）**
   - 通过交互式问题澄清用户需求
   - 创建完整的PRD文档
   - 需求优先级排序和功能规划

2. **UX Expert Agent（用户体验专家）**
   - 结合Figma链接和PRD文档
   - 输出前端UX规范和关键页面设计
   - 生成页面设计提示词

3. **Architect Agent（架构师专家）**
   - 基于PRD和UX Spec创建整体架构
   - 设计前端、后端架构
   - 输出API设计文档和数据库设计文档

4. **PO Agent（产品负责人）**
   - 根据PRD和架构设计
   - 输出Epics和Stories文档
   - 任务分解和优先级管理

5. **Dev Agent（开发工程师）**
   - 根据Stories进行编码实现
   - 自动错误检测和修复
   - 集成测试和部署

### 协作流程
```
用户需求输入 → PM Agent → UX Agent → Architect Agent → PO Agent → Dev Agent → 完整项目
     ↓              ↓           ↓              ↓            ↓           ↓
   需求澄清      PRD文档    UX规范      架构设计      Epics/Stories   源代码+测试
```

## 🔧 技术栈

### 前端技术栈
- **框架**: Vue.js 3 + Composition API
- **构建工具**: Vite
- **UI组件库**: Naive UI
- **状态管理**: Pinia
- **路由**: Vue Router 4
- **HTTP客户端**: Axios
- **语言**: TypeScript

### 后端技术栈
- **语言**: Go 1.21+
- **Web框架**: Gin
- **ORM**: GORM
- **验证**: validator
- **配置管理**: Viper
- **日志**: Zap
- **任务队列**: Asynq + Redis

### Agent服务技术栈
- **语言**: Node.js + TypeScript
- **Web框架**: Express.js
- **任务队列**: Bull Queue + Redis
- **实时通信**: Socket.io
- **包管理**: pnpm
- **日志**: Winston

### 数据存储
- **主数据库**: PostgreSQL 15+
- **缓存数据库**: Redis 7+
- **连接池**: pgxpool
- **Redis客户端**: go-redis

### 部署和运维
- **容器化**: Docker + Docker Compose
- **反向代理**: Traefik
- **进程管理**: Supervisor
- **监控**: Prometheus + Grafana

## 🚀 快速开始

### 环境要求
- **操作系统**: Windows 10/11 (开发环境)
- **Docker**: 最新版本
- **Node.js**: >= 18.0.0 (Agents Server)
- **Go**: >= 1.21 (Backend)
- **GPU**: 支持CUDA的NVIDIA显卡（AI加速，可选）

### 一键启动
```bash
# 0. 启动 ollama 服务
ollama serve

# 1. 克隆项目
git clone <repository-url>
cd app-maker

# 2. 启动所有服务（自动检查Docker、创建网络、启动服务）
make run-dev

# 3. 启动Agents Server (本地运行)
cd agents-server
pnpm install
pnpm start
```

### 访问地址
- **前端应用**: http://app-maker.localhost
- **后端API**: http://api.app-maker.localhost
- **Agents Server**: http://localhost:3001
- **Ollama AI**: http://chat.app-maker.localhost
- **Traefik Dashboard**: http://traefik.app-maker.localhost:8080/dashboard/

## 📁 项目结构

```
app-maker/
├── frontend/                 # Vue.js前端应用
│   ├── src/
│   │   ├── components/      # 组件
│   │   ├── pages/          # 页面
│   │   ├── stores/         # Pinia状态管理
│   │   ├── router/         # 路由配置
│   │   └── utils/          # 工具函数
│   └── Dockerfile
├── backend/                 # Go后端服务
│   ├── internal/
│   │   ├── api/            # API处理器
│   │   ├── services/       # 业务逻辑
│   │   ├── models/         # 数据模型
│   │   ├── repositories/   # 数据访问层
│   │   └── utils/          # 工具函数
│   └── Dockerfile
├── agents-server/          # Node.js Agent服务
│   ├── src/
│   │   ├── controllers/    # Agent控制器
│   │   ├── services/       # 核心服务
│   │   ├── queues/         # 任务队列
│   │   └── models/         # 数据模型
│   └── package.json
├── docs/                   # 项目文档
│   ├── PRD/               # 产品需求文档
│   ├── architect/         # 架构设计文档
│   └── stories/           # 用户故事
├── docker-compose.yml      # Docker编排文件
├── traefik.yml            # Traefik配置
└── traefik-external.yml   # 外部服务配置
```

## 🎨 核心功能

### 1. 智能需求分析
- 通过PM Agent的交互式问题澄清用户需求
- 智能需求澄清问题生成
- 多轮需求确认和优化
- 自动生成功能清单和技术架构建议

### 2. 多Agent协作开发
- 标准化的文档传递流程
- 自动化的Agent任务调度
- 实时的协作状态监控
- 智能任务分配和负载均衡

### 3. 文档管理与归档
- 自动文档版本管理
- 文档关联关系追踪
- 项目信息统一存储
- 文档模板和标准化

### 4. 后台任务执行
- 框架搭建：前后端框架自动搭建
- 需求实现：根据Stories进行编码实现
- 问题调试：自动错误检测和修复
- 部署打包：自动构建和部署
- 自动测试：集成测试和回归测试

### 5. 实时状态监控
- 前端轮询任务状态
- 实时进度更新
- 任务完成通知
- 错误状态反馈

## 🔒 设计约束

### 1. 技术约束
- **本地化部署**: 支持Windows本地运行，无需云端依赖
- **GPU加速**: 支持NVIDIA GPU加速AI模型推理
- **文件系统访问**: 直接访问主机文件系统，支持项目文件操作
- **容器化**: 核心服务容器化，Agent服务本地运行

### 2. 架构约束
- **微服务架构**: 前后端分离，Agent服务独立
- **事件驱动**: 基于消息队列的异步处理机制
- **无状态设计**: 支持水平扩展和负载均衡
- **API优先**: RESTful API设计，支持多种客户端

### 3. 安全约束
- **JWT认证**: 无状态认证，支持分布式部署
- **权限控制**: 基于用户的项目访问控制
- **输入验证**: 严格的参数验证和SQL注入防护
- **HTTPS**: 生产环境强制HTTPS

### 4. 性能约束
- **响应时间**: 页面加载 < 2秒，API响应 < 500ms
- **并发处理**: 支持1000+并发用户，100+项目并行开发
- **系统可用性**: 99.9%可用性要求
- **资源限制**: 内存使用 < 8GB，CPU使用 < 80%

### 5. 扩展性约束
- **水平扩展**: 支持多实例部署和负载均衡
- **功能扩展**: 支持新Agent角色的快速添加
- **集成能力**: 标准化接口，支持第三方系统集成
- **多语言支持**: 支持中英文界面和文档

## 📊 当前状态

### ✅ 已完成功能
1. **基础架构**: 前后端框架搭建完成
2. **用户系统**: 注册、登录、JWT认证
3. **项目管理**: 项目创建、列表、删除
4. **Agent框架**: Agents Server基础架构
5. **AI集成**: Ollama集成和项目总结生成
6. **部署配置**: Docker + Traefik完整配置
9. **项目详情接口**: 获取项目完整信息
10. **文件管理接口**: 文件列表、内容读取、预览
11. **对话消息接口**: 项目对话历史管理
12. **WebSocket推送**: 实时状态更新

### 🚧 待实现功能
1. **开发阶段接口**: 实时进度跟踪

## 🔧 开发指南

### 本地开发环境

#### 使用 Make 命令（推荐）
```bash
# 查看所有可用命令
make help

# 启动开发环境（一键启动所有服务）
make run-dev

# 停止开发环境
make stop-dev

# 查看开发环境日志
make logs-dev

# 重启开发环境
make restart-dev

# 健康检查
make health-check
```

#### 手动启动各服务
```bash
# 1. 确保 Docker 运行
make docker-ensure

# 2. 创建网络
make network-create

# 3. 启动所有服务
make run-dev

# 4. 启动 Agents Server (本地运行)
cd agents-server
pnpm install
pnpm start
```

#### 开发调试命令
```bash
# 重新构建并启动前端（开发环境）
make restart-front-dev

# 进入容器调试
make shell-frontend-dev    # 进入前端容器
make shell-backend-dev     # 进入后端容器

# 查看特定服务日志
make logs-frontend-dev     # 前端日志
make logs-backend-dev      # 后端日志

# 代码格式化
make fmt

# 代码检查
make lint

# 运行测试
make test
```

#### 数据库和缓存操作
```bash
# 数据库迁移
make db-migrate

# 数据库种子数据
make db-seed

# 清理缓存
make cache-clear

# 查看缓存信息
make cache-info
```

### 代码规范
- **Go**: 遵循gofmt和golint规范
- **TypeScript**: 使用ESLint和Prettier
- **Vue**: 遵循Vue 3 Composition API最佳实践
- **Git**: 使用Conventional Commits规范

---

*AutoCodeWeb - 让每个人都能通过简单的需求描述，快速获得完整的应用程序、前后端网站*
