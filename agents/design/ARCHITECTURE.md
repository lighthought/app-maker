# App Maker Agents Service 架构设计

## 概述

App Maker Agents Service 是基于 Go + Gin + Asynq 架构的多Agent协作开发服务，为自动化软件开发平台提供统一的Agent执行环境。服务通过HTTP API与backend系统集成，支持异步任务执行、实时状态反馈和Git工作流管理。

## 技术架构

### 核心框架
- **Go 1.24**: 高性能的系统编程语言
- **Gin**: 轻量级、高性能的Web框架
- **Asynq**: Redis基础上的异步任务队列系统
- **Zap**: 高性能结构化日志库
- **Swagger**: 自动API文档生成

### 基础设施
- **Redis**: 任务存储、缓存和队列管理
- **Git**: 版本控制和代码提交管理
- **共享模块**: shared-models提供统一的API接口

## 系统架构

### 整体架构图

```mermaid
graph TB
    subgraph "Backend Service"
        API[Go Backend API]
        DB[(PostgreSQL)]
        REDIS[(Redis Cache)]
    end
    
    subgraph "Agents Service"
        HTTP[HTTP Server<br/>Gin Router]
        
        subgraph "Handlers"
            PROJECT[Project Handler]
            PM[PM Handler]
            UX[UX Handler]
            ARCH[Architect Handler]
            PO[PO Handler]
            DEV[Dev Handler]
            TASK[Task Handler]
        end
        
        subgraph "Services"
            AGENT_TASK[Agent Task Service]
            PROJECT_SVC[Project Service]
            COMMAND[Command Service]
            GIT[Git Service]
        end
        
        subgraph "Async Workers"
            WORKER[Asynq Workers]
            QUEUE[Task Queue<br/>Redis]
        end
        
        subgraph "External Tools"
            BMAD[BMad CLI]
            NPM[NPM/Node.js]
            GIT_CLI[Git CLI]
        end
    end
    
    subgraph "Project Workspace"
        WORKSPACE[Project Directory]
        DOCS[Docs Folder]
        CODE[Source Code]
    end
    
    API --> HTTP
    HTTP --> PROJECT
    HTTP --> PM
    HTTP --> UX
    HTTP --> ARCH
    HTTP --> PO
    HTTP --> DEV
    HTTP --> TASK
    
    PROJECT --> AGENT_TASK
    PM --> AGENT_TASK
    UX --> AGENT_TASK
    ARCH --> AGENT_TASK
    PO --> AGENT_TASK
    DEV --> AGENT_TASK
    
    AGENT_TASK --> QUEUE
    QUEUE --> WORKER
    WORKER --> PROJECT_SVC
    WORKER --> COMMAND
    WORKER --> GIT
    
    COMMAND --> BMAD
    COMMAND --> NPM
    COMMAND --> GIT_CLI
    
    PROJECT_SVC --> WORKSPACE
    WORKER --> WORKSPACE
    
    WORKER --> REDIS
```

## Agent协作流程

### 开发阶段流程图

```mermaid
sequenceDiagram
    participant Backend as Backend API
    participant Agents as Agents Service
    participant Queue as Task Queue
    participant Agent as Agent Worker
    participant Git as Git Service
    participant Tools as External Tools

    Backend->>Agents: POST /project/setup
    Agents->>Queue: Create setup task
    Queue->>Agent: Execute setup task
    Agent->>Tools: Clone repository, install dependencies
    Agent->>Git: Commit setup changes
    Agent->>Backend: Update project status

    Backend->>Agents: POST /agent/pm/prd
    Agents->>Queue: Create PRD task
    Queue->>Agent: Execute PM Agent
    Agent->>Tools: Run BMad CLI with PRD prompt
    Agent->>Git: Commit PRD document
    Agent->>Backend: Update PRD status

    Backend->>Agents: POST /agent/dev/implstory
    Agents->>Queue: Create development task
    Queue->>Agent: Execute Dev Agent
    Agent->>Tools: Run BMad CLI with dev prompt
    Agent->>Git: Commit source code
    Agent->>Backend: Update development status
```

### 任务状态流转

```mermaid
stateDiagram-v2
    [*] --> PENDING: 创建任务
    
    PENDING --> IN_PROGRESS: 开始执行
    IN_PROGRESS --> DONE: 执行成功
    IN_PROGRESS --> FAILED: 执行失败
    
    FAILED --> IN_PROGRESS: 重试执行
    FAILED --> [*]: 重试超限
    
    DONE --> [*]: 任务完成
```

## API接口设计

### REST API端点

#### 项目管理
```http
POST /api/v1/project/setup
```

#### Agent任务接口
```http
POST /api/v1/agent/pm/prd                  # PRD生成
POST /api/v1/agent/ux-expert/ux-standard  # UX标准设计
POST /api/v1/agent/architect/architect     # 架构设计
POST /api/v1/agent/po/epicsandstories      # Epic和Story
POST /api/v1/agent/dev/implstory           # 实现Story
POST /api/v1/agent/dev/fixbug              # 修复Bug
```

### 任务状态查询
```http
GET /api/v1/tasks/{task_id}
```

## 系统组件关系图

```mermaid
classDiagram
    class Container {
        +AsyncClient: *asynq.Client
        +AgentTaskService: AgentTaskService
        +ProjectService: ProjectService
        +CommandService: CommanderService
        +ProjectHandler: *ProjectHandler
        +PmHandler: *PmHandler
        +DevHandler: *DevHandler
    }

    class ProjectHandler {
        +agentTaskService: AgentTaskService
        +SetupProjectEnvironment(ctx)
    }

    class AgentTaskService {
        +commandService: CommandService
        +gitService: GitService
        +Enqueue(projectGuid, agentType, message)
        +ProcessTask(ctx, task)
    }

    class CommandService {
        +SimpleExecute(ctx, workDir, command, args)
    }

    class GitService {
        +CommitAndPush(ctx, projectGuid, message)
    }

    Container --> ProjectHandler
    Container --> AgentTaskService
    Container --> ProjectService
    
    ProjectHandler --> AgentTaskService
    AgentTaskService --> CommandService
    AgentTaskService --> GitService
```

## 总结

App Maker Agents Service 采用现代化的微服务架构，具有以下特点：

1. **高性能**: 基于Go语言构建，支持高并发处理
2. **异步处理**: 使用Asynq实现任务队列，支持可靠的任务执行
3. **模块化设计**: 清晰的层次结构，便于维护和扩展
4. **集成简单**: 通过shared-models与Backend无缝集成
5. **工具链支持**: 灵活集成BMad CLI、Git等外部工具
6. **错误处理**: 完善的错误处理和重试机制
7. **实时反馈**: 支持任务状态实时查询和进度更新

---
## 联系方式

- 维护者: AI探趣星船长（抖音、小红书、B站同名）
- 邮箱: qqjack2012@gmail.com
- 项目地址: https://github.com/lighthought/app-maker
