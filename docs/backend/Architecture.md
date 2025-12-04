# App Maker Backend Architecture

## 1. Overview

**App Maker Backend** is a multi-Agent collaboration platform backend based on Go + Gin + GORM + PostgreSQL + Redis. It provides high-performance API services for the frontend, including project management, BMad-Method integration, and task execution.

### 1.1 Core Concepts
- **Microservices Architecture**: Services are divided by business domain, supporting independent deployment and scaling.
- **Layered Architecture**: Clear separation of concerns for easier maintenance and testing.
- **Event-Driven**: Asynchronous processing mechanism based on message queues (Redis Streams/Asynq).
- **RESTful API**: Standardized API design.
- **High Concurrency**: Leveraging Go's goroutines for high concurrency handling.

### 1.2 Technology Stack
- **Language**: Go 1.24+
- **Web Framework**: Gin 1.10+
- **ORM**: GORM 1.25+
- **Database**: PostgreSQL 15+
- **Cache**: Redis 7+
- **Task Queue**: Asynq + Redis
- **Configuration**: Viper
- **Logging**: Zap (via `shared-models/logger`)
- **Validation**: validator
- **JWT**: golang-jwt

## 2. System Architecture

### 2.1 Core Architecture Diagram

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

## 3. Project Structure

### 3.1 Directory Layout

```
backend/
├── cmd/                    # Application Entry Points
│   └── server/             # Main Server Entry
│       └── main.go
├── internal/               # Internal Packages (Private to this module)
│   ├── api/                # API Layer
│   │   ├── handlers/       # HTTP Handlers
│   │   ├── middleware/     # Gin Middleware
│   │   └── routes/         # Route Definitions
│   ├── config/             # Configuration Management
│   ├── container/          # Dependency Injection Container
│   ├── database/           # Database Connection & Init
│   ├── models/             # Domain Models (Specific to backend)
│   ├── repositories/       # Data Access Layer
│   └── services/           # Business Logic Layer
├── docs/                   # Swagger/OpenAPI Docs
├── scripts/                # Utility Scripts
├── .env                    # Environment Variables
├── go.mod                  # Go Module Definition
└── Dockerfile              # Docker Build File
```

> **Note**: Shared components such as `auth`, `logger`, `common` types, and `client` utilities are located in the `shared-models` module, not in a local `pkg` directory.

### 3.2 Layered Design

```
┌─────────────────────────────────────────────────────────────┐
│                        API Layer                            │
│               HTTP Handlers, Routes, Middleware             │
│               (internal/api)                                │
└─────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────┐
│                      Service Layer                          │
│           Business Logic, Transactions, Rules               │
│               (internal/services)                           │
└─────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────┐
│                    Repository Layer                         │
│           Data Access, Query Optimization                   │
│               (internal/repositories)                       │
└─────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────┐
│                       Data Layer                            │
│           PostgreSQL, Redis, File System                    │
└─────────────────────────────────────────────────────────────┘
```

## 4. Key Components Implementation

### 4.1 Application Entry (`cmd/server/main.go`)

The entry point initializes configuration, logging, database connections, and the dependency injection container before starting the Gin server.

```go
func main() {
    cfg, err := loadConfigAndService()
    if err != nil {
        logger.Fatal("Failed to load config or connect to services", logger.String("error", err.Error()))
        os.Exit(1)
    }

    container, engine := setupContainer(cfg)
    if container == nil {
        logger.Fatal("Failed to initialize dependency container")
        os.Exit(1)
    }

    startServer(cfg, engine, container)
}
```

### 4.2 Configuration (`internal/config`)

Configuration is managed using Viper, supporting YAML files and environment variables.

### 4.3 Service Layer Example (`internal/services`)

Services encapsulate business logic and interact with repositories.

```go
type UserService struct {
    userRepo   repositories.UserRepository
    cache      cache.Cache // From shared-models/cache or internal wrapper
    authHelper auth.Helper // From shared-models/auth
}
```

### 4.4 Middleware (`internal/api/middleware`)

- **AuthMiddleware**: Validates JWT tokens using `shared-models/auth`.
- **CORS**: Handles Cross-Origin Resource Sharing.
- **RequestID**: Adds a unique ID to each request context.

## 5. Development Stage Management

The system manages the full project development lifecycle through state transitions.

```mermaid
stateDiagram-v2
    [*] --> Initializing: Create Project
    
    Initializing --> SetupEnvironment: Env Setup
    SetupEnvironment --> PendingAgents: Agents Ready
    PendingAgents --> CheckRequirement: Req Analysis
    CheckRequirement --> GeneratePRD: Generate PRD
    GeneratePRD --> DefineUXStandard: UX Design
    DefineUXStandard --> DesignArchitecture: Architecture
    DesignArchitecture --> PlanEpicAndStory: Epic Planning
    PlanEpicAndStory --> DefineDataModel: Data Model
    DefineDataModel --> DefineAPI: API Definition
    DefineAPI --> DevelopStory: Development
    DevelopStory --> FixBug: Bug Fix
    FixBug --> RunTest: Testing
    RunTest --> Deploy: Deployment
    Deploy --> Done: Complete
    
    FixBug --> Failed: Fail
    RunTest --> Failed: Fail
    Deploy --> Failed: Fail
    
    Failed --> FixBug: Retry Fix
```

## 6. WebSocket Architecture

Real-time communication is handled via WebSocket for status updates, progress bars, and chat messages.

```mermaid
graph TB
    subgraph "WebSocket Service"
        WS[WebSocket Hub]
        Rooms[Project Rooms]
    end
    
    subgraph "Message Types"
        MSG[Project Messages]
        STATUS[Status Updates]
        PROGRESS[Progress Updates]
        ERROR[Error Alerts]
    end
    
    Client[Web Client] --> WS
    WS --> Rooms
    Rooms --> MSG
    Rooms --> STATUS
    Rooms --> PROGRESS
    
    MSG --> Client
    STATUS --> Client
    PROGRESS --> Client
```

## 7. Agent Integration

The Backend interacts with the **Agents Service** via HTTP API calls.

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
```

---
*Documentation consolidated and updated for App Maker Backend.*
