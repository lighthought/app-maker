# App Maker Frontend Architecture

## 系统架构概览

App Maker 前端采用现代化的 Vue 3 + TypeScript + Naive UI 技术栈，结合 Pinia 状态管理和 Axios HTTP 客户端，实现响应式、类型安全的多页面应用。系统支持多 Agent 协作的项目创建和管理，提供完整的用户认证、项目管理、实时对话等功能。

## 核心架构图

```mermaid
graph TB
    subgraph "用户界面层"
        Pages[页面组件]
        Layout[布局组件]
        Common[通用组件]
    end
    
    subgraph "状态管理层"
        Pinia[Pinia Store]
        UserStore[用户状态]
        ProjectStore[项目状态]
        FileStore[文件状态]
    end
    
    subgraph "服务层"
        HttpService[HTTP服务]
        ConfigService[配置服务]
        LoggerService[日志服务]
        TimeService[时间服务]
    end
    
    subgraph "工具层"
        Utils[工具函数]
        Types[类型定义]
        Router[路由管理]
    end
    
    subgraph "样式层"
        SCSS[SCSS样式]
        Variables[CSS变量]
        Mixins[样式混入]
    end
    
    subgraph "外部依赖"
        NaiveUI[Naive UI]
        Axios[Axios]
        VueRouter[Vue Router]
        Marked[Marked]
    end
    
    Pages --> Layout
    Pages --> Common
    Layout --> Common
    
    Pages --> Pinia
    Layout --> Pinia
    
    Pinia --> UserStore
    Pinia --> ProjectStore
    Pinia --> FileStore
    
    UserStore --> HttpService
    ProjectStore --> HttpService
    FileStore --> HttpService
    
    HttpService --> Axios
    HttpService --> ConfigService
    HttpService --> LoggerService
    
    Pages --> Router
    Router --> VueRouter
    
    Pages --> SCSS
    Layout --> SCSS
    Common --> SCSS
    
    SCSS --> Variables
    SCSS --> Mixins
    
    Pages --> Types
    Pinia --> Types
    HttpService --> Types
```

## 详细 UML 类图

### 1. 核心组件层

```mermaid
classDiagram
    class App {
        +router: Router
        +theme: Theme
        +beforeEnter()
        +enter()
        +leave()
    }

    class PageLayout {
        +collapsed: boolean
        +showSettingsModal: boolean
        +toggleSidebar()
        +openSettings()
    }

    class Header {
        +showNotifications: boolean
        +showUserMenu: boolean
        +handleSettings()
        +handleLogout()
    }

    class Sidebar {
        +collapsed: boolean
        +activeKey: string
        +menuOptions: MenuOption[]
        +handleMenuUpdate()
        +goToHome()
    }

    class SmartInput {
        +inputValue: string
        +placeholder: string
        +size: string
        +buttonType: string
        +handleEnterKey()
        +handleSend()
    }

    class UserSettingsModal {
        +show: boolean
        +formData: object
        +loading: boolean
        +handleSave()
        +handleClose()
    }

    App --> PageLayout
    PageLayout --> Header
    PageLayout --> Sidebar
    PageLayout --> UserSettingsModal
```

### 2. 页面组件层

```mermaid
classDiagram
    class Home {
        +isScrolled: boolean
        +currentLanguage: string
        +projectDescription: string
        +userProjects: Project[]
        +handleProjectCreate()
        +goToProject()
        +toggleLanguage()
    }

    class Auth {
        +isLogin: boolean
        +loading: boolean
        +formData: object
        +formRules: object
        +handleSubmit()
        +forgotPassword()
        +socialLogin()
    }

    class Dashboard {
        +searchKeyword: string
        +statusFilter: string
        +currentPage: number
        +currentProject: Project
        +backendStatus: string
        +createNewProject()
        +selectProject()
        +editProject()
        +downloadProject()
    }

    class CreateProject {
        +projectDescription: string
        +isCreating: boolean
        +handleProjectCreate()
    }

    class ProjectEdit {
        +projectId: string
        +requirements: string
        +activeTab: string
        +loadProject()
        +switchTab()
    }

    Home --> ProjectEdit
    Dashboard --> ProjectEdit
    CreateProject --> ProjectEdit
```

### 3. 状态管理层

```mermaid
classDiagram
    class UserStore {
        +user: User
        +token: string
        +refreshToken: string
        +isAuthenticated: boolean
        +initFromStorage()
        +validateTokenOnStartup()
        +login(credentials)
        +register(credentials)
        +logout()
        +refreshAuth()
        +updateProfile(profile)
        +changePassword()
        +hasPermission(permission)
    }

    class ProjectStore {
        +projects: Project[]
        +currentProject: Project
        +pagination: PaginationResponse
        +loading: boolean
        +error: string
        +fetchProjects(params)
        +createProject(data)
        +deleteProject(id)
        +getProject(id)
        +setCurrentProject(project)
        +getProjectMessages(projectId)
        +addChatMessage(projectId, message)
        +getProjectStages(projectId)
    }

    class FileStore {
        +downloadProject(projectId)
        +getFileContent(projectId, filePath)
        +getProjectFiles(projectId, path)
    }

    class User {
        +id: string
        +username: string
        +email: string
        +name: string
        +avatar: string
        +role: string
        +status: string
        +createdAt: string
        +updatedAt: string
    }

    class Project {
        +id: string
        +name: string
        +description: string
        +status: string
        +requirements: string
        +projectPath: string
        +backendPort: number
        +frontendPort: number
        +previewUrl: string
        +userId: string
        +user: UserInfo
        +created_at: string
        +updated_at: string
    }

    UserStore --> User
    ProjectStore --> Project
    Project --> User
```

### 4. 服务层

```mermaid
classDiagram
    class HttpService {
        +instance: AxiosInstance
        +isRefreshing: boolean
        +failedQueue: Array
        +setupInterceptors()
        +get(url, config)
        +post(url, data, config)
        +put(url, data, config)
        +delete(url, config)
        +patch(url, data, config)
        +download(url)
        +healthCheck()
        +setLogEnabled(enabled)
        +isLogEnabled()
    }

    class AppConfig {
        -instance: AppConfig
        -apiLogEnabled: boolean
        +getInstance()
        +isApiLogEnabled()
        +setApiLogEnabled(enabled)
        +getApiBaseUrl()
        +getConfig()
    }

    class ApiLogger {
        -enabled: boolean
        +formatTime()
        +formatRequest(config)
        +formatResponse(response)
        +formatError(error)
        +logRequest(config)
        +logResponse(response)
        +logError(error)
        +setEnabled(enabled)
        +isEnabled()
    }

    class TimeUtils {
        +formatDateTime(date)
        +formatDate(date)
        +formatDateShort(date)
        +getRelativeTime(date)
    }

    HttpService --> AppConfig
    HttpService --> ApiLogger
```

### 5. 工具层

```mermaid
classDiagram
    class Router {
        +routes: RouteRecordRaw[]
        +createRouter()
        +beforeEach()
        +afterEach()
    }

    class Types {
        +Project: interface
        +User: interface
        +CreateProjectData: interface
        +UpdateProjectData: interface
        +ProjectListRequest: interface
        +PaginationResponse: interface
        +LoginCredentials: interface
        +RegisterCredentials: interface
        +ConversationMessage: interface
        +DevStage: interface
        +FileItem: interface
    }

    Router --> Types
    TimeUtils --> Types
```

## 数据流图

### 用户认证流程

```mermaid
sequenceDiagram
    participant User
    participant AuthPage
    participant UserStore
    participant HttpService
    participant Backend

    User->>AuthPage: 输入登录信息
    AuthPage->>UserStore: login(credentials)
    UserStore->>HttpService: post('/auth/login', credentials)
    HttpService->>Backend: HTTP请求
    Backend-->>HttpService: 返回token和用户信息
    HttpService-->>UserStore: 响应数据
    UserStore->>UserStore: 保存token和用户信息
    UserStore-->>AuthPage: 登录成功
    AuthPage->>AuthPage: 跳转到Dashboard
```

### 项目创建流程

```mermaid
sequenceDiagram
    participant User
    participant CreateProjectPage
    participant ProjectStore
    participant HttpService
    participant Backend

    User->>CreateProjectPage: 输入项目描述
    CreateProjectPage->>ProjectStore: createProject(data)
    ProjectStore->>HttpService: post('/projects', data)
    HttpService->>Backend: HTTP请求
    Backend-->>HttpService: 返回项目信息
    HttpService-->>ProjectStore: 响应数据
    ProjectStore->>ProjectStore: 更新项目列表
    ProjectStore-->>CreateProjectPage: 创建成功
    CreateProjectPage->>CreateProjectPage: 跳转到项目详情
```

### 项目编辑页面交互流程

```mermaid
sequenceDiagram
    participant User
    participant ProjectEdit
    participant ConversationContainer
    participant ProjectPanel
    participant ProjectStore
    participant Backend

    User->>ProjectEdit: 访问项目编辑页面
    ProjectEdit->>ProjectStore: getProject(projectId)
    ProjectStore->>Backend: GET /projects/{id}
    Backend-->>ProjectStore: 返回项目信息
    ProjectStore-->>ProjectEdit: 项目数据
    
    ProjectEdit->>ConversationContainer: 传递projectId和requirements
    ConversationContainer->>ProjectStore: getProjectStages(projectId)
    ProjectStore->>Backend: GET /projects/{id}/stages
    Backend-->>ProjectStore: 返回开发阶段
    ProjectStore-->>ConversationContainer: 阶段数据
    
    ConversationContainer->>ProjectStore: getProjectMessages(projectId)
    ProjectStore->>Backend: GET /projects/{id}/conversations
    Backend-->>ProjectStore: 返回对话历史
    ProjectStore-->>ConversationContainer: 对话数据
    
    ProjectEdit->>ProjectPanel: 传递project数据
    ProjectPanel->>FileStore: getProjectFiles(projectId)
    FileStore->>Backend: GET /projects/{id}/files
    Backend-->>FileStore: 返回文件列表
    FileStore-->>ProjectPanel: 文件数据
    
    User->>ProjectPanel: 点击文件
    ProjectPanel->>FileStore: getFileContent(projectId, filePath)
    FileStore->>Backend: GET /projects/{id}/files/content
    Backend-->>FileStore: 返回文件内容
    FileStore-->>ProjectPanel: 文件内容
```

## 技术栈说明

### 1. 核心框架
- **Vue 3**: 采用 Composition API，提供更好的类型推导和逻辑复用
- **TypeScript**: 提供类型安全，减少运行时错误
- **Vite**: 快速的构建工具，支持热重载

### 2. UI 框架
- **Naive UI**: 现代化的 Vue 3 UI 组件库
- **SCSS**: CSS 预处理器，支持变量、混入、嵌套等特性

### 3. 状态管理
- **Pinia**: Vue 3 官方推荐的状态管理库
- **响应式设计**: 基于 Vue 3 的响应式系统

### 4. 路由管理
- **Vue Router**: 官方路由管理器
- **路由守卫**: 支持认证和权限控制

### 5. HTTP 客户端
- **Axios**: 基于 Promise 的 HTTP 客户端
- **拦截器**: 支持请求/响应拦截和错误处理
- **自动重试**: 支持 token 刷新和请求重试

### 6. 工具库
- **@vueuse/core**: Vue 组合式 API 工具集
- **marked**: Markdown 解析器
- **@iconify/vue**: 图标库组件

## 目录结构说明

```
frontend/src/
├── App.vue                 # 根组件
├── main.ts                 # 应用入口
├── vite-env.d.ts          # Vite 环境类型声明
├── assets/                 # 静态资源
│   └── logo.svg           # 应用Logo
├── components/             # 组件目录
│   ├── common/            # 通用组件
│   │   ├── index.ts       # 组件导出
│   │   └── SmartInput.vue # 智能输入组件
│   ├── layout/            # 布局组件
│   │   ├── Header.vue     # 顶部导航
│   │   ├── PageLayout.vue # 页面布局
│   │   └── Sidebar.vue    # 侧边栏
│   ├── ConversationMessage.vue # 对话消息组件
│   ├── ConversationContainer.vue # 对话容器组件
│   ├── DevStages.vue      # 开发阶段组件
│   ├── FileTreeNode.vue   # 文件树组件
│   ├── MonacoEditor.vue   # 代码编辑器组件
│   ├── ProjectPanel.vue   # 项目面板组件
│   ├── TaskProgressModal.vue # 任务进度弹窗
│   └── UserSettingsModal.vue # 用户设置弹窗
├── pages/                  # 页面组件
│   ├── Auth.vue           # 认证页面
│   ├── CreateProject.vue  # 创建项目页面
│   ├── Dashboard.vue      # 仪表板页面
│   ├── Home.vue           # 首页
│   ├── ProjectEdit.vue    # 项目编辑页面
│   └── WebSocketDebug.vue # WebSocket 调试菜单
├── router/                 # 路由配置
│   └── index.ts           # 路由定义
├── stores/                 # 状态管理
│   ├── file.ts            # 文件状态
│   ├── project.ts         # 项目状态
│   ├── task.ts            # 任务状态
│   └── user.ts            # 用户状态
├── styles/                 # 样式文件
│   ├── main.scss          # 主样式文件
│   ├── mixins.scss        # SCSS混入
│   └── variables.scss     # CSS变量
├── types/                  # 类型定义
│   ├── project.ts         # 项目相关类型
│   ├── task.ts            # 任务相关类型
│   ├── user.ts            # 用户相关类型
│   └── websocket.ts       # WebSocket 相关类型
└── utils/                  # 工具函数
    ├── config.ts          # 配置管理
    ├── http.ts            # HTTP服务
    ├── log.ts             # 日志工具
    ├── time.ts            # 时间工具
    └── websocket.ts       # WebSocket 工具
├── userWorker.ts           # Web Worker 用户任务处理
```

## 开发约束和规范

### 1. 组件开发规范
- **组件命名**: 使用 PascalCase 命名
- **Props 定义**: 使用 TypeScript 接口定义 Props 类型
- **事件命名**: 使用 kebab-case 命名自定义事件
- **样式隔离**: 使用 scoped 样式避免样式污染

### 2. 状态管理规范
- **Store 结构**: 每个功能模块对应一个 Store
- **数据获取**: 通过 Store 的 actions 获取数据
- **状态更新**: 通过 mutations 更新状态
- **类型安全**: 所有 Store 数据都有对应的 TypeScript 类型

### 3. HTTP 请求规范
- **统一封装**: 所有 HTTP 请求通过 HttpService 发送
- **错误处理**: 统一的错误处理和用户提示
- **请求拦截**: 自动添加认证 token
- **响应拦截**: 统一处理响应数据和错误

### 4. 路由管理规范
- **路由守卫**: 使用 beforeEach 进行认证检查
- **懒加载**: 页面组件使用动态导入
- **路由命名**: 使用有意义的 route name

### 5. 样式开发规范
- **CSS 变量**: 使用 CSS 变量定义主题色彩
- **SCSS 混入**: 使用混入定义通用样式
- **响应式设计**: 使用媒体查询适配不同屏幕
- **组件样式**: 每个组件使用独立的 scoped 样式

## 性能优化策略

### 1. 组件优化
- **懒加载**: 路由组件使用动态导入
- **组件缓存**: 使用 keep-alive 缓存组件
- **虚拟滚动**: 大列表使用虚拟滚动

### 2. 状态管理优化
- **数据缓存**: Store 中缓存常用数据
- **按需加载**: 只加载必要的数据
- **状态持久化**: 重要状态保存到 localStorage

### 3. 网络优化
- **请求合并**: 合并多个相关请求
- **请求缓存**: 缓存不经常变化的数据
- **错误重试**: 网络错误自动重试

### 4. 构建优化
- **代码分割**: 按路由分割代码
- **Tree Shaking**: 移除未使用的代码
- **资源压缩**: 压缩 CSS 和 JS 文件

## 安全考虑

### 1. 认证安全
- **Token 管理**: 安全的 token 存储和刷新
- **路由守卫**: 保护需要认证的页面
- **自动登出**: token 过期自动登出

### 2. 数据安全
- **输入验证**: 前端输入验证和过滤
- **XSS 防护**: 防止跨站脚本攻击
- **CSRF 防护**: 使用 CSRF token

### 3. 通信安全
- **HTTPS**: 强制使用 HTTPS 通信
- **请求签名**: 重要请求使用签名验证
- **敏感数据**: 敏感数据不存储在客户端

## 扩展性设计

### 1. 组件扩展
- **组件库**: 基于 Naive UI 扩展自定义组件
- **主题系统**: 支持多主题切换
- **国际化**: 支持多语言切换

### 2. 功能扩展
- **插件系统**: 支持功能插件扩展
- **模块化**: 功能模块独立开发和部署
- **API 扩展**: 支持新的 API 接口

### 3. 部署扩展
- **多环境**: 支持开发、测试、生产环境
- **CDN 部署**: 静态资源使用 CDN 加速
- **容器化**: 支持 Docker 容器化部署

## 实际组件架构

### 1. 页面组件实现
```typescript
// 实际实现的页面组件
- Home.vue: 首页，包含产品介绍、快速创建、用户项目展示
- Auth.vue: 认证页，支持登录/注册切换、表单验证、社交登录
- Dashboard.vue: 仪表板，项目统计、搜索筛选、分页展示
- CreateProject.vue: 创建项目，智能输入框、需求输入
- ProjectEdit.vue: 项目编辑，分屏布局、对话交互、文件查看
```

### 2. 布局组件实现
```typescript
// 实际实现的布局组件
- PageLayout.vue: 主布局，包含侧边栏、顶部导航、内容区域
- Header.vue: 顶部导航，用户菜单、通知、设置
- Sidebar.vue: 侧边栏，菜单导航、Logo展示
```

### 3. 业务组件实现
```typescript
// 实际实现的业务组件
- SmartInput.vue: 智能输入框，支持多行输入、发送按钮
- ConversationContainer.vue: 对话容器，消息列表、开发阶段
- ConversationMessage.vue: 对话消息，支持Markdown渲染
- DevStages.vue: 开发阶段，进度可视化
- ProjectPanel.vue: 项目面板，文件树、代码展示、预览
- UserSettingsModal.vue: 用户设置弹窗
```

### 4. 状态管理实现
```typescript
// 实际实现的Store
- user.ts: 用户状态，认证、权限、个人信息
- project.ts: 项目状态，项目列表、详情、对话、阶段
- file.ts: 文件状态，文件列表、内容、下载
```

### 5. 工具函数实现
```typescript
// 实际实现的工具函数
- http.ts: HTTP服务，拦截器、错误处理、token管理
- config.ts: 配置管理，API地址、日志开关
- log.ts: 日志工具，请求/响应日志
- time.ts: 时间工具，格式化、相对时间
```

## 实际数据流示例

### 数据流管理
- 用户认证：Token 管理和自动刷新
- 项目管理：CRUD 操作和状态同步  
- 实时通信：WebSocket 连接和消息处理
- 任务跟踪：异步任务状态监控
- **用户认证**: 完整的登录、注册、登出流程，支持token自动刷新和验证
- **项目管理**: 项目创建、列表、详情、删除功能，支持搜索和筛选
- **实时对话**: 与AI Agent的交互界面，支持Markdown渲染和消息展开
- **文件管理**: 项目文件查看、内容展示、项目下载功能
- **开发进度**: 可视化项目开发阶段和进度跟踪
- **项目预览**: 通过iframe嵌入实时预览项目效果

### 2. 页面功能实现
- **首页**: 产品介绍、快速创建项目、用户项目展示、中英文切换
- **认证页**: 登录/注册切换、表单验证、社交登录按钮、协议弹窗
- **仪表板**: 项目统计卡片、搜索筛选、分页展示、系统状态监控
- **创建项目**: 智能输入框、项目需求输入、自动跳转到项目编辑
- **项目编辑**: 分屏布局、对话交互、文件查看、代码展示

### 3. 组件特性实现
- **响应式设计**: 适配桌面、平板、手机各种屏幕尺寸
- **主题系统**: 统一的色彩和间距系统，支持CSS变量
- **国际化**: 中英文切换支持，动态语言切换
- **无障碍**: 符合WCAG标准的可访问性设计

### 4. 技术实现细节
- **TypeScript**: 完整的类型定义和类型安全，严格模式
- **Pinia**: 现代化的状态管理，支持响应式数据
- **Vue 3**: Composition API 和响应式系统，更好的逻辑复用
- **Naive UI**: 丰富的UI组件库，支持主题定制
- **SCSS**: 模块化的样式系统，支持变量和混入
- **Axios**: 统一的HTTP客户端，支持拦截器和错误处理
- **路由守卫**: 认证检查和权限控制，自动跳转
- **Docker**: 容器化部署，支持开发和生产环境

## 部署架构

### 1. 开发环境
- **Vite Dev Server**: 开发服务器，支持热重载
- **Nginx**: 反向代理，API转发
- **Docker**: 容器化部署

### 2. 生产环境
- **静态构建**: Vite 构建优化
- **Nginx**: 静态文件服务和API代理
- **CDN**: 静态资源加速
- **Docker**: 生产环境容器

### 3. 构建优化
- **代码分割**: 按需加载
- **资源压缩**: Gzip压缩
- **缓存策略**: 静态资源长期缓存
- **安全头**: 安全响应头设置