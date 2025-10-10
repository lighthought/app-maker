# 贡献指南

欢迎参与 App Maker 项目的开发！App Maker 是一个基于 AI 的多 Agent 协作平台，能够根据用户需求自动生成完整的前后端项目。

## 📋 目录

- [项目简介](#项目简介)
- [技术栈](#技术栈)
- [开发环境搭建](#开发环境搭建)
- [项目结构](#项目结构)
- [开发流程](#开发流程)
- [贡献方式](#贡献方式)
- [待实现功能清单](#待实现功能清单)
- [代码规范](#代码规范)
- [问题反馈](#问题反馈)
- [许可证](#许可证)

## 🚀 项目简介

App Maker 是一个创新的 AI 驱动开发平台，它通过多个专业 Agent（产品经理、架构师、开发工程师、测试工程师等）的协作，能够根据用户的简单需求描述，自动生成完整的企业级前后端项目。

### 核心特性

- 🤖 **多 Agent 协作**: PM、UX、Architect、Dev、PO 等专业角色协同工作
- 🎯 **技术模板驱动**: 基于预定义的技术模板生成标准化项目
- 📦 **容器化部署**: 生成的项目可直接容器化部署
- 🔄 **GitLab 集成**: 支持企业 GitLab 实例或默认容器
- ⚡ **实时进度**: WebSocket 实时展示开发进度
- 🎨 **现代 UI**: Vue 3 + TypeScript + Naive UI 构建的现代化界面

## 🛠 技术栈

### 后端
- **Go 1.21+**: 主要开发语言
- **Gin**: HTTP Web 框架
- **GORM**: ORM 库
- **PostgreSQL**: 主数据库
- **Redis**: 缓存和任务队列
- **Asynq**: 异步任务队列

### 前端
- **Vue 3**: 前端框架
- **TypeScript**: 类型安全
- **Naive UI**: UI 组件库
- **Pinia**: 状态管理
- **Axios**: HTTP 客户端
- **Monaco Editor**: 代码编辑器

### 基础设施
- **Docker & Docker Compose**: 容器化
- **GitLab CE**: 代码管理和 CI/CD
- **Traefik**: 反向代理和负载均衡

## 🏗 开发环境搭建

### 环境要求

- **操作系统**: Windows 10/11 (开发环境，8G+显存、16G+内存)
- **Docker Desktop**: >= 28.4+
- **Claude Code**: >= 2.0.11 (Agents)
- **Go**: >= 1.24 (Backend & Agents)
- **Node.js**: >= 18.0.0 (Frontend)

### 快速开始

1. **克隆项目**
```bash
git clone https://github.com/lighthought/app-maker.git
cd app-maker
```

2. **配置环境变量**
```bash
# 重命名 .env.example 为 .env
cp .env.example .env

# 修改 your-xxx 为实际的环境配置（密码自定）
# 编辑 .env 文件，设置数据库连接等信息
```

3. **启动本地 Ollama**
```bash
# 下载并启动 Ollama 模型
ollama pull deepseek-r1:14b
ollama serve
```

4. **构建并启动服务**
```bash
# 构建开发环境
make build-dev

# 启动前后端、数据库、redis、gitlab容器
make run-dev
```

5. **配置 GitLab**
```bash
# 获取 GitLab 初始 root 密码
docker-compose exec gitlab cat /etc/gitlab/initial_root_password

# 访问 http://gitlab.app-maker.localhost/ 使用 root 账号登录
# 批准管理员账号的注册，并设置为管理员
# 访问 http://gitlab.app-maker.localhost/admin/users 编辑用户为管理员

# 创建 app-maker 群组
# 访问 http://gitlab.app-maker.localhost/groups/new 创建群组

# 配置 SSH 密钥
# 拷贝 run-dev 执行过程中输出的 ssh-rsa 开头的字符串
# 访问 http://gitlab.app-maker.localhost/-/user_settings/ssh_keys 添加密钥

# 配置主机 Git 账号密码
# 通过 git clone 获取 GitLab 上的空代码仓库，输入用户名密码
```

6. **启动 Agents 服务**
```bash
# 进入 agents 目录
cd agents

# 安装依赖
go mod tidy

# 启动 agents 服务
go run cmd/server/main.go
```

### 访问地址

- **前端应用**: http://app-maker.localhost
- **后端API**: http://api.app-maker.localhost/swagger/index.html
- **GitLab CE**: http://gitlab.app-maker.localhost/
- **Agents服务**: http://localhost:8088/swagger/index.html

### 使用 Makefile（推荐）

项目提供了便捷的 Makefile 命令：

```bash
# 查看所有可用命令
make help

# 构建开发环境
make build-dev

# 启动开发环境
make run-dev

# 停止开发环境
make stop-dev

# 清理开发环境
make clean-dev

# 查看服务状态
make status

# 查看日志
make logs

# 进入容器
make shell-backend
make shell-frontend
make shell-gitlab
```

## 📁 项目结构

```
app-maker/
├── backend/                 # 后端服务
│   ├── cmd/server/         # 服务入口
│   ├── internal/           # 内部包
│   │   ├── api/           # API 处理器
│   │   ├── models/        # 数据模型
│   │   ├── services/      # 业务逻辑
│   │   └── database/      # 数据库相关
│   ├── configs/           # 配置文件
│   └── scripts/           # 脚本文件
├── frontend/               # 前端应用
│   ├── src/
│   │   ├── components/    # Vue 组件
│   │   ├── pages/         # 页面组件
│   │   ├── stores/        # Pinia 状态管理
│   │   ├── types/         # TypeScript 类型定义
│   │   └── utils/         # 工具函数
│   └── public/            # 静态资源
├── agents/                 # Agents 服务
│   ├── cmd/server/        # 服务入口
│   ├── internal/
│   │   ├── api/          # API 处理器
│   │   ├── services/     # Agent 服务
│   │   └── worker/       # 任务处理器
│   └── design/           # 设计文档
├── shared-models/          # 共享模型
│   ├── agent/            # Agent 相关模型
│   ├── common/           # 通用模型
│   └── utils/            # 工具函数
├── docs/                  # 项目文档
├── scripts/               # 项目脚本
└── docker-compose.yml     # Docker 编排文件
```

## 🔄 开发流程

### 1. 创建 Issue

在开始开发之前，请先创建一个 Issue 来描述您要解决的问题或添加的功能。

### 2. Fork 项目

点击 GitHub 页面右上角的 "Fork" 按钮，将项目 fork 到您的账户。

### 3. 创建分支

```bash
# 克隆您的 fork
git clone https://github.com/your-username/app-maker.git
cd app-maker

# 添加上游仓库
git remote add upstream https://github.com/original-owner/app-maker.git

# 创建功能分支
git checkout -b feature/your-feature-name
```

### 4. 开发

- 编写代码
- 添加测试
- 更新文档
- 确保代码通过所有检查

### 5. 提交代码

```bash
# 添加更改
git add .

# 提交更改（使用约定式提交格式）
git commit -m "feat: add new feature"

# 推送到您的 fork
git push origin feature/your-feature-name
```

### 6. 创建 Pull Request

在 GitHub 上创建 Pull Request，详细描述您的更改。

## 🤝 贡献方式

### 代码贡献

- **Bug 修复**: 修复已知问题
- **功能开发**: 添加新功能
- **性能优化**: 提升系统性能
- **代码重构**: 改善代码质量
- **测试覆盖**: 增加测试用例

### 文档贡献

- **README 更新**: 完善项目说明
- **API 文档**: 补充接口文档
- **开发指南**: 编写开发教程
- **用户手册**: 完善使用说明

### 其他贡献

- **问题反馈**: 报告 Bug 或建议改进
- **社区支持**: 帮助其他开发者
- **设计建议**: 提供 UI/UX 改进建议

## 🎯 待实现功能清单

我们欢迎开发者参与以下功能的开发，这些功能将大大提升 App Maker 的完整性和用户体验：

### 🚧 高优先级功能

#### 1. **Story 编辑系统** 📝
- **功能描述**: 用户可编辑 Epics 和 Stories，支持排序、删减、修改
- **技术要点**: 前端拖拽排序、后端 CRUD 接口、数据同步
- **技能要求**: Vue.js、Go、数据库设计
- **预估工作量**: 2-3 周

#### 2. **AI 集成优化** 🤖
- **功能描述**: 优化 code cli 集成，减少上下文 token 消耗
- **技术要点**: 上下文管理、token 优化、缓存策略
- **技能要求**: AI 集成、Go、性能优化
- **预估工作量**: 1-2 周

#### 3. **用户工具系统** 🛠️
- **功能描述**: Agent 反向调用前端用户工具
- **技术要点**: WebSocket 双向通信、工具注册机制、权限控制
- **技能要求**: WebSocket、Vue.js、Go
- **预估工作量**: 2-3 周

#### 4. **代码生成流程** 💻
- **功能描述**: 完整的代码生成流程实现
- **技术要点**: 模板引擎、代码生成器、文件系统操作
- **技能要求**: Go、模板引擎、文件操作
- **预估工作量**: 3-4 周

### 🎨 中优先级功能

#### 5. **项目预览功能** 👀
- **功能描述**: 实时项目预览功能
- **技术要点**: 热重载、代理服务、文件监控
- **技能要求**: Node.js、Docker、网络编程
- **预估工作量**: 2-3 周

#### 6. **管理员后台** 👨‍💼
- **功能描述**: 模板管理、用户管理、项目管理的管理员后台页面
- **技术要点**: 权限系统、数据统计、批量操作
- **技能要求**: Vue.js、Go、权限设计
- **预估工作量**: 3-4 周

#### 7. **模板适配** 📱
- **功能描述**: 适配微信小程序、纯前端、纯后端、安卓应用、React 等模板
- **技术要点**: 模板引擎、项目结构分析、代码生成
- **技能要求**: 多平台开发经验、模板设计
- **预估工作量**: 4-6 周

#### 8. **项目状态回滚** 🔄
- **功能描述**: 实现项目回滚到指定聊天记录，再继续
- **技术要点**: 版本控制、状态管理、数据恢复
- **技能要求**: Git、数据库设计、状态管理
- **预估工作量**: 2-3 周

### 🔧 低优先级功能

#### 9. **测试自动化** 🧪
- **功能描述**: 单元测试和 CI/CD 集成测试
- **技术要点**: 测试框架、CI/CD 集成、覆盖率统计
- **技能要求**: 测试框架、CI/CD、Go/Vue.js
- **预估工作量**: 2-3 周

#### 10. **第三方认证登录** 🔐
- **功能描述**: Github、Google、微信、支付宝、手机二维码登录认证
- **技术要点**: OAuth 2.0、JWT、第三方 API 集成
- **技能要求**: 认证系统、第三方集成、安全设计
- **预估工作量**: 3-4 周

#### 11. **多语言支持** 🌍
- **功能描述**: 后台接口、WebSocket 消息、Agent 提示词支持多语言（中、英文）
- **技术要点**: i18n、语言包管理、动态切换
- **技能要求**: 国际化、Vue.js、Go
- **预估工作量**: 2-3 周

### 🎯 如何参与功能开发

1. **选择功能**: 从上述清单中选择您感兴趣的功能
2. **创建 Issue**: 在 GitHub 上创建 Issue 讨论实现方案
3. **Fork 项目**: Fork 项目到您的账户
4. **创建分支**: 创建功能分支开始开发
5. **提交 PR**: 完成开发后提交 Pull Request

### 💡 贡献建议

- **新手友好**: 建议从 Bug 修复或文档完善开始
- **功能开发**: 选择与您技能匹配的功能
- **协作开发**: 复杂功能可以多人协作完成
- **测试优先**: 开发功能时请同时编写测试用例

### 🏆 贡献奖励

- **代码贡献**: 您的名字将出现在贡献者列表中
- **功能完成**: 完成重要功能的贡献者将获得特殊标识
- **社区认可**: 在项目社区中获得认可和感谢

## 📝 代码规范

### Go 代码规范

- 遵循 [Go 官方代码规范](https://golang.org/doc/effective_go.html)
- 使用 `gofmt` 格式化代码
- 使用 `golint` 检查代码质量
- 添加必要的注释和文档

```bash
# 格式化代码
go fmt ./...

# 检查代码质量
golint ./...

# 运行测试
go test ./...
```

### Vue/TypeScript 代码规范

- 遵循 [Vue 3 官方风格指南](https://vuejs.org/style-guide/)
- 使用 TypeScript 严格模式
- 使用 ESLint 和 Prettier 格式化代码
- 组件使用 Composition API

```bash
# 检查代码质量
npm run lint

# 格式化代码
npm run format

# 运行测试
npm run test
```

### 提交信息规范

使用 [约定式提交](https://www.conventionalcommits.org/) 格式：

```
<type>(<scope>): <description>

[optional body]

[optional footer(s)]
```

类型包括：
- `feat`: 新功能
- `fix`: Bug 修复
- `docs`: 文档更新
- `style`: 代码格式调整
- `refactor`: 代码重构
- `test`: 测试相关
- `chore`: 构建过程或辅助工具的变动

## 🐛 问题反馈

### 报告 Bug

1. 确认问题是否已在 Issues 中存在
2. 使用 Bug 报告模板创建 Issue
3. 提供详细的复现步骤
4. 包含环境信息和错误日志

### 功能建议

1. 检查是否已有类似建议
2. 使用功能请求模板创建 Issue
3. 详细描述功能需求和使用场景
4. 提供可能的实现方案

## 📄 许可证

本项目采用 [AGPLv3 许可证](LICENSE)。

## 🙏 致谢

感谢所有为 App Maker 项目做出贡献的开发者！

## 📞 联系方式

- 项目维护者: AI探趣星船长
- 项目地址: https://github.com/lighthought/app-maker
- 问题反馈: https://github.com/lighthought/app-maker/issues

---

再次感谢您的贡献！让我们一起打造更好的 App Maker！🚀
