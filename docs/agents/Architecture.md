# App Maker Agents Service Architecture

## 1. Overview

**App Maker Agents Service** is a multi-Agent collaboration development service based on **Go + Gin + Asynq**. It provides a unified execution environment for automated software development. The service integrates with the backend system via HTTP APIs, supporting asynchronous task execution, real-time status feedback, and Git workflow management.

## 2. Technical Architecture

### 2.1 Core Framework
- **Go 1.24**: High-performance system programming language.
- **Gin**: Lightweight, high-performance Web framework.
- **Asynq**: Asynchronous task queue system based on Redis.
- **Zap**: High-performance structured logging library.
- **Swagger**: Automatic API documentation generation.

### 2.2 Infrastructure
- **Redis**: Task storage, caching, and queue management.
- **Git**: Version control and code commit management.
- **Shared Models**: `shared-models` module providing unified API interfaces and types.

## 3. System Architecture

### 3.1 Overall Architecture Diagram

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

## 4. Agent Collaboration Flow

### 4.1 Development Stage Flow

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

### 4.2 Task State Transition

```mermaid
stateDiagram-v2
    [*] --> PENDING: Create Task
    
    PENDING --> IN_PROGRESS: Start Execution
    IN_PROGRESS --> DONE: Success
    IN_PROGRESS --> FAILED: Failure
    
    FAILED --> IN_PROGRESS: Retry
    FAILED --> [*]: Max Retries Exceeded
    
    DONE --> [*]: Task Complete
```

## 5. API Design

### 5.1 REST API Endpoints

#### Project Management
```http
POST /api/v1/project/setup
```

#### Agent Task Interfaces
```http
POST /api/v1/agent/pm/prd                  # Generate PRD
POST /api/v1/agent/ux-expert/ux-standard   # UX Standard Design
POST /api/v1/agent/architect/architect     # Architecture Design
POST /api/v1/agent/po/epicsandstories      # Epics and Stories
POST /api/v1/agent/dev/implstory           # Implement Story
POST /api/v1/agent/dev/fixbug              # Fix Bug
```

#### Task Status Query
```http
GET /api/v1/tasks/{task_id}
```

## 6. Component Relationship

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

## 7. Summary

App Maker Agents Service adopts a modern microservices architecture with the following features:

1.  **High Performance**: Built on Go, supporting high concurrency.
2.  **Asynchronous Processing**: Uses Asynq for reliable task queues.
3.  **Modular Design**: Clear layering for easy maintenance.
4.  **Simple Integration**: Seamless integration with Backend via `shared-models`.
5.  **Toolchain Support**: Flexible integration with BMad CLI, Git, etc.
6.  **Error Handling**: Robust error handling and retry mechanisms.
7.  **Real-time Feedback**: Supports real-time task status queries and progress updates.
