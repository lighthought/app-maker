# AutoCodeWeb Backend Architecture

## 系统架构概览

AutoCodeWeb 后端采用分层架构设计，结合异步任务处理机制，实现多Agent协作的智能项目开发系统。

## 核心架构图

```mermaid
graph TB
    subgraph "Client Layer"
        Web[Web Frontend]
        API[API Client]
    end
    
    subgraph "API Gateway"
        Router[Gin Router]
        Auth[Auth Middleware]
        CORS[CORS Middleware]
        Logger[Logger Middleware]
    end
    
    subgraph "Application Layer"
        PH[Project Handler]
        UH[User Handler]
        CH[Cache Handler]
    end
    
    subgraph "Business Logic Layer"
        PS[Project Service]
        TES[Task Execution Service]
        PDS[Project Dev Service]
        US[User Service]
    end
    
    subgraph "Data Access Layer"
        PR[Project Repository]
        UR[User Repository]
    end
    
    subgraph "Data Layer"
        PG[(PostgreSQL)]
        RD[(Redis)]
        FS[File System]
    end
    
    subgraph "External Services"
        Cursor[Cursor CLI]
        BMad[BMad-Method]
        Node[Node.js/npx]
    end
    
    Web --> Router
    API --> Router
    Router --> Auth
    Router --> CORS
    Router --> Logger
    
    Router --> PH
    Router --> TH
    Router --> UH
    Router --> TagH
    Router --> CH
    
    PH --> PS
    TH --> TS
    UH --> US
    TagH --> TagS
    CH --> RD
    
    PS --> PR
    TS --> TR
    TS --> PR
    US --> UR
    TagS --> TagR
    
    PR --> PG
    TR --> PG
    UR --> PG
    TagR --> PG
    
    TES --> PS
    TES --> PR
    TES --> TR
    TES --> PDS
    
    PDS --> Cursor
    PDS --> BMad
    PDS --> Node
    PDS --> FS
```

## 详细 UML 类图

### 1. 核心模型层

```mermaid
classDiagram
    class Project {
        +ID: string
        +Name: string
        +Description: string
        +Requirements: string
        +Status: string
        +DevStatus: string
        +DevProgress: int
        +CurrentTaskID: string
        +UserID: string
        +ProjectPath: string
        +BackendPort: int
        +FrontendPort: int
        +ApiBaseUrl: string
        +AppSecretKey: string
        +DatabasePassword: string
        +RedisPassword: string
        +JwtSecretKey: string
        +Subnetwork: string
        +CreatedAt: time.Time
        +UpdatedAt: time.Time
        +GetDevStageProgress() int
        +GetDevStageDescription() string
    }

    class User {
        +ID: string
        +Username: string
        +Email: string
        +Password: string
        +CreatedAt: time.Time
        +UpdatedAt: time.Time
    }

    User --> Project : has many
```

### 2. 数据访问层

```mermaid
classDiagram
    class ProjectRepository {
        <<interface>>
        +Create(ctx, project) error
        +GetByID(ctx, id) *Project
        +Update(ctx, project) error
        +Delete(ctx, id) error
        +List(ctx, userID, limit, offset) []*Project
        +GetNextAvailablePorts(ctx) (int, int, error)
        +GetByPath(ctx, path, userID) *Project
    }


    class UserRepository {
        <<interface>>
        +Create(ctx, user) error
        +GetByID(ctx, id) *User
        +GetByEmail(ctx, email) *User
        +Update(ctx, user) error
        +GetByUsername(ctx, username) *User
    }

```

### 3. 业务逻辑层

```mermaid
classDiagram
    class ProjectService {
        <<interface>>
        +CreateProject(ctx, req, userID) *ProjectInfo
        +GetProject(ctx, projectID, userID) *ProjectInfo
        +DeleteProject(ctx, projectID, userID) error
        +ListProjects(ctx, req, userID) ([]*ProjectInfo, *PaginationResponse)
        +DownloadProject(ctx, projectID, userID) []byte
    }

    class ProjectStageService {
        +projectService: ProjectService
        +projectRepo: ProjectRepository
        +taskRepo: TaskRepository
        +projectDevService: *ProjectDevService
        +baseProjectsDir: string
        +semaphore: *semaphore.Weighted
        +maxConcurrency: int64
        +mu: sync.Mutex
        +StartProjectDevelopment(ctx, projectID) error
        +executeWithSemaphore(ctx, project, task)
        +executeProjectDevelopment(ctx, project, task)
        +generatePRD(ctx, project, task) error
        +defineUXStandards(ctx, project, task) error
        +designArchitecture(ctx, project, task) error
        +defineDataModel(ctx, project, task) error
        +defineAPIs(ctx, project, task) error
        +planEpicsAndStories(ctx, project, task) error
        +developStories(ctx, project, task) error
        +fixBugs(ctx, project, task) error
        +runTests(ctx, project, task) error
        +packageProject(ctx, project, task) error
        +addTaskLog(ctx, taskID, level, message)
    }

    class ProjectDevService {
        +baseProjectsDir: string
    }

    class ProjectTemplateService {
        +templatePath: string
        +InitializeProject(project) error
        +extractTemplate() error
        +copyTemplateFiles(projectPath) error
        +updateProjectConfig(project) error
        +createDockerCompose(project) error
        +createNginxConfig(project) error
    }

    class ProjectNameGenerator {
        +GenerateProjectConfig(requirements, config) bool
        +generateProjectName(requirements) string
        +generateDescription(requirements) string
        +generateApiBaseUrl() string
        +generateSecretKey() string
        +generateDatabasePassword() string
        +generateRedisPassword() string
        +generateJwtSecretKey() string
    }

    class UserService {
        <<interface>>
        +Register(ctx, req) *UserInfo
        +Login(ctx, req) *LoginResponse
        +GetProfile(ctx, userID) *UserInfo
        +UpdateProfile(ctx, userID, req) *UserInfo
        +ChangePassword(ctx, userID, req) error
        +ValidateToken(token) (string, error)
    }

```

### 4. API 控制器层

```mermaid
classDiagram
    class ProjectHandler {
        +projectService: ProjectService
        +CreateProject(c) *gin.Context
        +GetProject(c) *gin.Context
        +DeleteProject(c) *gin.Context
        +ListProjects(c) *gin.Context
        +DownloadProject(c) *gin.Context
    }


    class UserHandler {
        +userService: UserService
        +Register(c) *gin.Context
        +Login(c) *gin.Context
        +GetProfile(c) *gin.Context
        +UpdateProfile(c) *gin.Context
        +ChangePassword(c) *gin.Context
    }

    class CacheHandler {
        +cache: cache.Cache
        +monitor: *cache.Monitor
        +HealthCheck(c) *gin.Context
        +GetStats(c) *gin.Context
        +GetMemoryUsage(c) *gin.Context
        +GetKeyspaceStats(c) *gin.Context
        +GetPerformanceMetrics(c) *gin.Context
    }
```

### 5. 中间件和工具层

```mermaid
classDiagram
    class JWTService {
        +secretKey: string
        +expireTime: time.Duration
        +GenerateToken(userID) string
        +ValidateToken(token) (string, error)
        +RefreshToken(token) string
    }

    class AuthMiddleware {
        +jwtService: *JWTService
        +AuthMiddleware(jwtService) gin.HandlerFunc
        +extractToken(c) string
        +validateToken(token) (string, error)
    }

    class CORSMiddleware {
        +CORS() gin.HandlerFunc
    }

    class LoggerMiddleware {
        +Logger() gin.HandlerFunc
        +logRequest(c) *gin.Context
    }

    class Cache {
        <<interface>>
        +Get(key) interface{}
        +Set(key, value, ttl) error
        +Delete(key) error
        +Exists(key) bool
        +Incr(key) int64
        +Decr(key) int64
        +Expire(key, ttl) error
    }

    class RedisCache {
        +client: *redis.Client
        +Get(key) interface{}
        +Set(key, value, ttl) error
        +Delete(key) error
        +Exists(key) bool
        +Incr(key) int64
        +Decr(key) int64
        +Expire(key, ttl) error
    }

    class CacheFactory {
        +CreateCache(config) cache.Cache
        +createRedisCache(config) *RedisCache
    }

    class CacheMonitor {
        +cache: cache.Cache
        +GetStats() *Stats
        +GetMemoryUsage() *MemoryUsage
        +GetKeyspaceStats() *KeyspaceStats
        +GetPerformanceMetrics() *PerformanceMetrics
    }

    class Logger {
        +Info(msg, fields...)
        +Error(msg, fields...)
        +Warn(msg, fields...)
        +Debug(msg, fields...)
        +String(key, value) Field
        +Int(key, value) Field
        +Error(key, err) Field
    }
```

## 数据流图

### 项目创建流程

```mermaid
sequenceDiagram
    participant Client
    participant ProjectHandler
    participant ProjectService
    participant ProjectStageService
    participant ProjectTemplateService
    participant ProjectRepository
    participant Database

    Client->>ProjectHandler: POST /api/v1/projects
    ProjectHandler->>ProjectService: CreateProject(req, userID)
    ProjectService->>ProjectRepository: Create(project)
    ProjectRepository->>Database: INSERT project
    ProjectService->>ProjectTemplateService: InitializeProject(project)
    ProjectTemplateService->>ProjectService: Project initialized
    ProjectService->>ProjectStageService: StartProjectDevelopment(projectID)
    ProjectStageService->>ProjectRepository: Update project status
    ProjectStageService->>TaskRepository: Create task
    ProjectStageService-->>ProjectService: Development started
    ProjectService-->>ProjectHandler: Project created
    ProjectHandler-->>Client: 201 Created
```

### 异步任务执行流程

```mermaid
sequenceDiagram
    participant ProjectStageService
    participant ProjectDevService
    participant CursorCLI
    participant ProjectRepository
    participant TaskRepository
    participant Database

    ProjectDevService->>CursorCLI: Install cursor-cli
    ProjectDevService->>CursorCLI: Install bmad-method
    ProjectDevService-->>ProjectStageService: Environment ready
    
    loop Development Stages
        ProjectStageService->>ProjectRepository: Update project status
        ProjectStageService->>CursorCLI: Execute development stage
        CursorCLI-->>ProjectStageService: Stage completed
        ProjectStageService->>TaskRepository: Add task log
        ProjectStageService->>ProjectRepository: Update progress
    end
    
    ProjectStageService->>ProjectRepository: Mark project completed
    ProjectStageService->>TaskRepository: Mark task completed
```

## 部署架构

```mermaid
graph TB
    subgraph "Load Balancer"
        LB[Nginx Load Balancer]
    end
    
    subgraph "Application Tier"
        App1[Backend App 1]
        App2[Backend App 2]
        App3[Backend App 3]
    end
    
    subgraph "Data Tier"
        PG[(PostgreSQL Primary)]
        PG_Slave[(PostgreSQL Slave)]
        Redis[(Redis Cluster)]
    end
    
    subgraph "Storage Tier"
        FS[File Storage]
        Backup[Backup Storage]
    end
    
    subgraph "Monitoring"
        Prometheus[Prometheus]
        Grafana[Grafana]
        ELK[ELK Stack]
    end
    
    LB --> App1
    LB --> App2
    LB --> App3
    
    App1 --> PG
    App2 --> PG
    App3 --> PG
    PG --> PG_Slave
    
    App1 --> Redis
    App2 --> Redis
    App3 --> Redis
    
    App1 --> FS
    App2 --> FS
    App3 --> FS
    
    App1 --> Prometheus
    App2 --> Prometheus
    App3 --> Prometheus
    
    Prometheus --> Grafana
    App1 --> ELK
    App2 --> ELK
    App3 --> ELK
```

## 技术决策说明

### 1. 分层架构
- **API层**: 处理HTTP请求，参数验证，响应格式化
- **服务层**: 业务逻辑处理，事务管理
- **仓库层**: 数据访问抽象，支持多种数据源
- **模型层**: 数据结构定义，业务规则封装

### 2. 异步任务处理
- **信号量控制**: 限制并发任务数量，防止资源耗尽
- **状态机**: 清晰的任务状态转换，便于监控和调试
- **日志记录**: 详细的任务执行日志，支持问题排查

### 3. AI集成
- **Cursor CLI**: 作为主要的AI代码生成工具
- **BMad-Method**: 提供结构化的开发方法论
- **Node.js环境**: 支持现代前端开发工具链

### 4. 缓存策略
- **Redis**: 高性能缓存，支持复杂数据结构
- **监控**: 实时缓存性能监控，支持容量规划
- **分层缓存**: 支持多级缓存策略

### 5. 安全设计
- **JWT认证**: 无状态认证，支持分布式部署
- **权限控制**: 基于用户的项目访问控制
- **输入验证**: 严格的参数验证和SQL注入防护

## 性能考虑

### 1. 数据库优化
- 索引优化：为常用查询字段建立索引
- 连接池：复用数据库连接，减少连接开销
- 读写分离：支持主从数据库架构

### 2. 缓存策略
- 热点数据缓存：项目信息、用户会话等
- 缓存预热：系统启动时预加载关键数据
- 缓存失效：合理的TTL设置，保证数据一致性

### 3. 异步处理
- 非阻塞操作：耗时操作异步处理
- 并发控制：防止资源竞争和系统过载
- 任务队列：支持任务优先级和重试机制

## 扩展性设计

### 1. 水平扩展
- 无状态设计：支持多实例部署
- 负载均衡：支持多种负载均衡策略
- 数据分片：支持大规模数据存储

### 2. 功能扩展
- 插件架构：支持功能模块化扩展
- 配置驱动：支持运行时配置变更
- API版本管理：支持API向后兼容

### 3. 集成能力
- 标准化接口：支持第三方系统集成
- Webhook支持：支持事件驱动集成
- 消息队列：支持异步消息处理
