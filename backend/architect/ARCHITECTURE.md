# App Maker Backend 架构设计

## 系统架构概览

App Maker Backend 采用分层架构设计，通过异步任务处理机制和Agent服务集成，实现多Agent协作的智能项目开发平台。

## 核心架构图

```mermaid
graph TB
    subgraph "Client Layer"
        Web[Web Frontend]
        WS[WebSocket Client]
    end
    
    subgraph "API Gateway"
        Router[Gin Router]
        Auth[Auth Middleware]
        CORS[CORS Middleware]
        Logger[Logger Middleware]
    end
    
    subgraph "Handlers Layer"
        PH[Project Handler]
        UH[User Handler]
        FH[File Handler]
        CH[Chat Handler]
        TH[Task Handler]
        WH[WebSocket Handler]
    end
    
    subgraph "Services Layer"
        PS[Project Service]
        PSS[Project Stage Service]
        US[User Service]
        MS[Message Service]
        FS[File Service]
        WSS[WebSocket Service]
    end
    
    subgraph "Repository Layer"
        PR[Project Repository]
        UR[User Repository]
        MR[Message Repository]
        SR[Stage Repository]
    end
    
    subgraph "Data Layer"
        PG[(PostgreSQL)]
        RD[(Redis)]
        FS_FILE[File System]
    end
    
    subgraph "External Services"
        Agents[Agents Service]
        Git[Git Lab CI/CD]
    end
    
    Web --> Router
    WS --> Router
    Router --> Auth
    Router --> CORS
    Router --> Logger
    
    Router --> PH
    Router --> UH
    Router --> FH
    Router --> CH
    Router --> TH
    Router --> WH
    
    PH --> PS
    UH --> US
    CH --> MS
    FH --> FS
    TH --> PSS
    WH --> WSS
    
    PS --> PR
    US --> UR
    MS --> MR
    PSS --> SR
    
    PR --> PG
    UR --> PG
    MR --> PG
    SR --> PG
    
    PSS --> Agents
    PS --> Git
    
    WSS --> RD
    FS --> FS_FILE
```

## 开发阶段管理

系统支持完整的项目开发生命周期管理：

```mermaid
stateDiagram-v2
    [*] --> Initializing: 创建项目
    
    Initializing --> SetupEnvironment: 环境准备
    SetupEnvironment --> PendingAgents: Agents就绪
    PendingAgents --> CheckRequirement: 需求分析
    CheckRequirement --> GeneratePRD: 生成PRD
    GeneratePRD --> DefineUXStandard: UX设计
    DefineUXStandard --> DesignArchitecture: 架构设计
    DesignArchitecture --> PlanEpicAndStory: Epic规划
    PlanEpicAndStory --> DefineDataModel: 数据模型
    DefineDataModel --> DefineAPI: API定义
    DefineAPI --> DevelopStory: 功能开发
    DevelopStory --> FixBug: Bug修复
    FixBug --> RunTest: 测试验证
    RunTest --> Deploy: 部署发布
    Deploy --> Done: 完成
    
    FixBug --> Failed: 失败
    RunTest --> Failed: 失败
    Deploy --> Failed: 失败
    
    Failed --> FixBug: 重试修复
```

## WebSocket 实时通信架构

```mermaid
graph TB
    subgraph "WebSocket Service"
        WS[WebSocket Hub]
        Rooms[Project Rooms]
        Connections[Client Connections]
    end
    
    subgraph "Message Types"
        MSG[Project Messages]
        STATUS[Status Updates]
        PROGRESS[Progress Updates]
        ERROR[Error Alerts]
    end
    
    subgraph "Clients"
        WEB[Web Client]
        MOBILE[Mobile Client]
    end
    
    WEB --> CONN1[Connection 1]
    MOBILE --> CONN2[Connection 2]
    
    CONN1 --> WS
    CONN2 --> WS
    
    WS --> Rooms
    Rooms --> MSG
    Rooms --> STATUS
    Rooms --> PROGRESS
    Rooms --> ERROR
    
    MSG --> WEB
    STATUS --> WEB
    PROGRESS --> WEB
    ERROR --> WEB
```

## 数据模型关系

```mermaid
erDiagram
    users ||--o{ projects : owns
    projects ||--o{ dev_stages : has
    projects ||--o{ conversation_messages : contains
    users ||--o{ websocket_connections : maintains
    
    users {
        string id PK
        string email UK
        string username UK
        string password
        string role
        string status
        timestamp created_at
        timestamp updated_at
    }
    
    projects {
        string id PK
        string guid UK
        string name
        string description
        text requirements
        string status
        string dev_status
        int dev_progress
        string user_id FK
        string project_path
        string gitlab_repo_url
        int backend_port
        int frontend_port
        timestamp created_at
        timestamp updated_at
    }
    
    dev_stages {
        string id PK
        string project_id FK
        string project_guid FK
        string name
        string status
        int progress
        text log_data
        timestamp created_at
        timestamp updated_at
    }
    
    conversation_messages {
        string id PK
        string project_guid FK
        string type
        string agent_role
        string agent_name
        text content
        text markdown_content
        boolean is_markdown
        boolean is_expanded
        timestamp created_at
    }
    
    websocket_connections {
        string id PK
        string user_id FK
        string project_guid FK
        string connection_state
        timestamp connected_at
        timestamp last_active_at
    }
```

## Agent 集成架构

```mermaid
sequenceDiagram
    participant Backend as Backend Service
    participant Agents as Agents Service
    participant Client as Web Client
    
    Backend->>Agents: POST /project/setup
    Agents-->>Backend: Environment Ready
    
    Backend->>Agents: POST /agent/pm/prd
    Agents->>Agents: Execute PM Agent
    Agents-->>Backend: PRD Generated
    
    Backend->>Client: WebSocket Status Update
    Client->>Client: Show Progress
    
    Backend->>Agents: POST /agent/dev/implstory
    Agents->>Agents: Execute Dev Agent
    Agents-->>Backend: Story Implemented
    
    Backend->>Client: WebSocket Completion Update
```

## GitLab CI/CD 集成

```mermaid
graph TB
    subgraph "Development Flow"
        DEV[Developer]
        BMAD[Workspace]
        WORKER[GitLab Runner]
    end
    
    subgraph "GitLab"
        REPO[Repository]
        CI[CI/CD Pipeline]
        REGISTRY[Container Registry]
    end
    
    subgraph "Deployment"
        DEPLOY[Production Environment]
        MONITOR[Monitoring]
    end
    
    DEV --> BMAD
    BMAD --> REPO
    REPO --> CI
    CI --> WORKER
    WORKER --> REGISTRY
    CI --> DEPLOY
    DEPLOY --> MONITOR
```

## 技术栈说明

### 数据持久化
- **PostgreSQL**: 主数据库，支持ACID事务
- **Redis**: 缓存和会话存储，支持发布订阅
- **GORM**: ORM框架，支持数据库迁移和关联查询

### 异步处理
- **Asynq**: 基于Redis的任务队列，支持任务重试和调度
- **Goroutines**: Go原生并发，支持高并发处理

### API与通信
- **Gin**: 高性能HTTP框架，中间件丰富
- **WebSocket**: 实时双向通信，支持房间管理
- **JWT**: 无状态认证，支持分布式部署

### 外部集成
- **Agents Service**: AI Agent服务，通过HTTP API调用
- **GitLab**: 代码仓库和CI/CD流水线
- **Docker**: 容器化部署，支持多环境

---
## 联系方式

- 维护者: AI探趣星船长（抖音、小红书、B站同名）
- 邮箱: qqjack2012@gmail.com
- 项目地址: https://github.com/lighthought/app-maker
