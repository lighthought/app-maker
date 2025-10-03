# App Maker - 多Agent自动实现APP和网站项目平台

## 🎯 产品概述

**App Maker** 是一个基于 Go + Vue.js 的多Agent协作开发平台，通过标准化的Agent协作流程，自动生成APP和网站项目的完整代码。核心理念是"想法即应用"，让用户通过简单的需求描述快速获得完整的应用程序。

### 目标用户
- **产品经理**: 需要快速验证产品想法的专业人士
- **创业者**: 希望快速构建MVP的初创团队  
- **设计师**: 需要将设计转化为可交互产品的创意人员
- **非技术背景用户**: 有想法但缺乏技术实现能力的个人

## 🏗️ 系统架构

### 三层架构设计
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   前端应用      │    │   后端服务       │    │   Agent服务     │
│  (Vue.js 3)     │◄──►│   (Go + Gin)    │◄──►│  (Go + Gin)     │
│                 │    │                 │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   用户 界面     │    │   数据存储       │    │   外部工具      │
│  Naive UI       │    │  PostgreSQL     │    │  Claude CLI     │
│  TypeScript     │    │  Redis          │    │  Ollama AI      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### 核心组件
- **Frontend**: Vue.js 3 + TypeScript + Naive UI + Pinia
- **Backend**: Go + Gin + GORM + PostgreSQL + Asynq + Redis
- **Agents Service**: Go + Gin + Asynq + Redis
- **Shared Models**: Go模块，共享数据模型和客户端
- **部署**: GitLab CI/CD + Group Runners

## 🤖 多Agent协作体系
### Agent角色分工
| Agent | 中文名 | 英文名 | 职责 |
|-------|--------|--------|------|
| Analyst | 需求分析师 | Mary | 分析项目需求，生成项目简介和市场研究 |
| PM | 产品经理 | John | 编写产品需求文档(PRD) |
| UX Expert | UX专家 | Sally | 设计用户体验标准和界面规范 |
| Architect | 架构师 | Winston | 设计系统\前端\后端技术架构，输出API、数据库设计文档 |
| PO | 产品负责人 | Sarah | 划分Epic和用户故事，任务分解和优先级管理 |
| Dev | 开发工程师 | James | 实现用户故事、修复Bug、测试、部署 |



### 协作流程
```
用户需求输入 → PM Agent → UX Agent → Architect Agent → PO Agent → Dev Agent → 完整项目
     ↓           ↓         ↓              ↓              ↓           ↓
   需求澄清    PRD文档    UX规范     架构/前后端设计  Epics/Stories 源代码+测试
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

### Agents服务技术栈
- **语言**: Go 1.24+
- **Web框架**: Gin
- **任务队列**: Asynq + Redis
- **HTTP客户端**: Axios (通过cursor.cli调用)
- **日志**: Zap
- **依赖注入**: Container模式

### 数据存储
- **主数据库**: PostgreSQL 15+
- **缓存数据库**: Redis 7+
- **连接池**: pgxpool
- **Redis客户端**: go-redis

### 部署和运维
- **CI/CD**: GitLab CI/CD + Group Runners
- **容器化**: Docker + Docker Compose
- **数据库**: PostgreSQL + Redis
- **代理**: Traefik

## 🚀 快速开始

### 环境要求
- **操作系统**: Windows 10/11 (开发环境，8G+显存、16G+内存)
- **Docker Desktop**: （最新版本，28.4+)
- **Go**: >= 1.24 (Backend & Agents)
- **前端**: Node.js >= 18.0.0 (Frontend)
- **数据库**: PostgreSQL >= 15, Redis >= 7

### 启动过程

- **环境变量**
重命名 .env.example 为 .env
修改 your-xxx 为实际的环境配置（密码自定）
- **gitlab管理员**
启动容器后在 http://gitlab.app-maker.localhost/ 注册 GITLAB_USERNAME 和 GITLAB_EMAIL 对应的管理员账号、邮箱

```bash
# 1. 启动本地 ollama
ollama pull deepseek-r1:14b
ollama serve

# 2. 启动前后端、数据库、redis、gitlab容器
make build-dev
make run-dev

# 3. gitlab 配置
## 3.1 gitlab 获取初始 root 密码
docker-compose exec gitlab cat /etc/gitlab/initial_root_password

## 3.2 gitlab 批准用户注册、修改为管理员
# root 账号登陆 http://gitlab.app-maker.localhost/ 
# 批准管理员账号的注册、http://gitlab.app-maker.localhost/admin/users 页面编辑为管理员账号

## 3.3 管理员用户创建 app-maker 群组
# 管理员账号登录后，创建 app-maker 群组
# http://gitlab.app-maker.localhost/groups/new

## 3.4 gitlab 管理员配置 ssh-key
# 拷贝 run-dev执行过程中输出的: ssh-rsa 开头，gmail.com 结尾的字符串，粘贴到
# http://gitlab.app-maker.localhost/-/user_settings/ssh_keys  添加新密钥的输入框中

## 3.5 主机配置 git 账号密码
# 主机上通过 git clone 获取 http://gitlab.app-maker.localhost/ 上新建的空代码仓库，输入用户名密码

# 4. 启动Agents服务
cd agents
go mod tidy
go run cmd/server/main.go
```

### 访问地址
- **前端应用**: http://app-maker.localhost
- **后端API**: http://api.app-maker.localhost/swagger/index.html
- **Gitlab CE**: http://gitlab.app-maker.localhost/
- **Agents服务**: http://localhost:8088/swagger/index.html

## 📁 项目结构

```
app-maker/
├── frontend/                 # Vue.js前端应用
│   ├── src/
│   │   ├── components/      # 组件
│   │   ├── pages/          # 页面
│   │   ├── stores/         # Pinia状态管理
│   │   ├── router/         # 路由配置
│   │   ├── utils/          # 工具函数
│   │   └── types/          # TypeScript类型定义
│   ├── architect/          # 架构设计文档
│   └── package.json
├── backend/                 # Go后端服务
│   ├── internal/
│   │   ├── api/            # API处理器
│   │   ├── services/       # 业务逻辑
│   │   ├── models/         # 数据模型
│   │   ├── repositories/   # 数据访问层
│   │   ├── config/         # 配置管理
│   │   └── container/      # 依赖注入容器
│   ├── architect/          # 架构设计文档
│   └── go.mod
├── agents/                  # Go Agents服务
│   ├── internal/
│   │   ├── api/            # API处理器
│   │   ├── services/       # Agent核心服务
│   │   ├── config/         # 配置管理
│   │   └── container/      # 依赖注入容器
│   ├── design/             # 设计文档
│   └── go.mod
├── shared-models/          # Go共享模型模块
│   ├── agent/              # Agent相关类型
│   ├── common/             # 通用类型和常量
│   ├── client/             # HTTP客户端
│   ├── auth/               # JWT认证
│   ├── tasks/              # 任务模型
│   └── go.mod
└── LICENSE                 # 许可证
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
1. **基础架构**: Go + Vue.js 架构搭建完成
2. **用户系统**: 注册、登录、JWT认证
3. **项目管理**: 项目创建、列表、删除、下载、详情
4. **Agents服务**: Go + Gin + Asynq 异步任务处理
5. **共享模块**: shared-models Go模块
6. **前端功能**: 实时对话、文件管理、代码编辑器
7. **开发阶段**: 多阶段开发流程跟踪
8. **WebSocket**: 实时状态更新和通知
9. **文档系统**: 完整的架构设计文档
10. **部署配置**: GitLab CI/CD + Group Runners

### 🚧 待实现功能
1. **AI集成**: cursor.cli 集成优化（目前的阶段独立方案上下文 token 消耗大）
2. **用户工具**: Agent反向调用前端用户工具
3. **Story编辑**: 用户可编辑Epics和stories，排序、删减、修改
4. **分阶段重试**: 项目开发过程支持分阶段重试
5. **代码生成**: 完整的代码生成流程
6. **项目预览**: 实时项目预览功能
7. **管理员页面**: 实现模板管理、用户管理、项目管理的管理员后台页面
8. **模板适配**: 适配微信小程序模板、纯前端项目模板、纯后端项目模板、安卓应用模板、跨平替 react 模板 
9. **项目状态回滚**: 实现项目回滚到指定聊天记录，再继续
10. **测试自动化**: 单元测试和CI/CD集成测试
11. **第三方认证登录**: Github、google、微信、支付宝、手机二维码登录认证
12. **后台、Agent多语言**: 后台接口、websocket消息、Agent提示词支持多语言（中、英文）
13. **前台支持文件切换编码**: 前端编辑控件支持切换文件编码打开 

## 🔧 开发指南

### 本地开发环境

#### 启动开发服务
```bash
# 1. 启动本地 ollama
ollama serve

# 2. 启动前后端、数据库、redis、gitlab容器
make build-dev
make run-dev

# 3. 启动Agents服务
cd agents
go mod tidy
go run cmd/server/main.go
```

#### 开发调试
```bash
# 后端调试
cd backend
go run cmd/server/main.go -c configs/config.yaml

# Agents服务调试
cd agents
go run cmd/server/main.go

# 前端热重载
cd frontend
pnpm dev --host
```

#### 代码检查
```bash

# TypeScript代码检查
cd frontend
pnpm lint
pnpm type-check

# Go代码检查
go vet ./...
golangci-lint run
```

### 代码规范
- **Go**: 遵循gofmt和golint规范
- **TypeScript**: 使用ESLint和Prettier
- **Vue**: 遵循Vue 3 Composition API最佳实践
- **Git**: 使用Conventional Commits规范

---

## 📄 许可证

本项目采用 AGPLv3 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。如果您希望在不遵守AGPL条款的项目中集成本代码，需要另行购买商业许可，请联系我。

## 📞 联系方式

- **维护者**: AI探趣星船长（抖音、小红书、B站同名）
- **邮箱**: qqjack2012@gmail.com
- **项目地址**: https://github.com/lighthought/app-maker

---

*App Maker - 让每个人都能通过简单的需求描述，快速获得完整的应用程序、前后端网站*
