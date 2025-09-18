# Agents Server Architecture Document

## 概述

Agents Server 是基于 Node.js + TypeScript 的多Agent协作开发服务，采用自定义实现 + MCP SDK 混合方案。该服务作为 AutoCodeWeb 平台的核心组件，负责执行 BMad-Method 标准流程中的各个Agent角色任务，通过标准化的Agent协作流程实现项目开发的自动化。

## 技术架构

### 整体架构图

```mermaid
graph TB
    subgraph "External Systems"
        Backend[Go Backend API]
        Ollama[Ollama AI Service]
        Cursor[Cursor CLI]
        Figma[Figma API]
    end
    
    subgraph "Agents Server Core"
        HTTP[HTTP Server<br/>Express.js]
        WS[WebSocket Server<br/>Socket.io]
        Queue[Task Queue<br/>Bull Queue]
        
        subgraph "Agent Orchestrator"
            Orchestrator[Agent Orchestrator]
            Workflow[Workflow Engine]
            State[State Manager]
        end
        
        subgraph "MCP Integration"
            MCPManager[MCP Manager]
            MCPClient[MCP Client]
            MCPServers[MCP Servers]
        end
        
        subgraph "Agent Controllers"
            PM[PM Agent Controller]
            UX[UX Expert Controller]
            ARCH[Architect Controller]
            PO[PO Agent Controller]
            DEV[Dev Agent Controller]
        end
        
        subgraph "Core Services"
            FS[File System Service]
            CMD[Command Execution Service]
            TEMPLATE[Template Service]
            NOTIFY[Notification Service]
        end
    end
    
    subgraph "Project Workspace"
        PROJECT[Project Directory]
        DOCS[Docs Folder]
        CODE[Source Code]
    end
    
    Backend --> HTTP
    HTTP --> Queue
    Queue --> Orchestrator
    Orchestrator --> PM
    Orchestrator --> UX
    Orchestrator --> ARCH
    Orchestrator --> PO
    Orchestrator --> DEV
    
    PM --> MCPManager
    UX --> MCPManager
    ARCH --> MCPManager
    PO --> MCPManager
    DEV --> MCPManager
    
    MCPManager --> MCPClient
    MCPClient --> MCPServers
    
    PM --> FS
    UX --> FS
    ARCH --> FS
    PO --> FS
    DEV --> FS
    
    FS --> PROJECT
    CMD --> Cursor
    CMD --> Ollama
    
    WS --> Backend
    NOTIFY --> WS
```

## 核心组件设计

### 1. Agent Orchestrator (编排器)

```mermaid
classDiagram
    class AgentOrchestrator {
        -agents: Map~string, MCPAgent~
        -workflow: WorkflowStep[]
        -stateManager: StateManager
        -mcpManager: MCPManager
        +executeWorkflow(context: ProjectContext): Promise~void~
        +addAgent(agent: MCPAgent): void
        +removeAgent(agentName: string): void
        +pauseWorkflow(): void
        +resumeWorkflow(): void
        +requestHumanInput(result: AgentResult): Promise~any~
    }
    
    class WorkflowStep {
        +id: string
        +agentName: string
        +dependencies: string[]
        +condition: WorkflowCondition
        +retryPolicy: RetryPolicy
        +timeout: number
    }
    
    class StateManager {
        -state: Map~string, any~
        -history: StateHistory[]
        +getState(key: string): any
        +setState(key: string, value: any): void
        +saveCheckpoint(): void
        +restoreCheckpoint(checkpointId: string): void
    }
    
    AgentOrchestrator --> WorkflowStep
    AgentOrchestrator --> StateManager
```

### 2. MCP Integration Layer

```mermaid
classDiagram
    class MCPManager {
        -servers: Map~string, Client~
        -connections: Map~string, Transport~
        +connectServer(serverName: string): Promise~Client~
        +disconnectServer(serverName: string): Promise~void~
        +callTool(serverName: string, toolName: string, args: any): Promise~any~
        +listTools(serverName: string): Promise~Tool[]
        +getServerStatus(serverName: string): ServerStatus
    }
    
    class MCPClient {
        -client: Client
        -transport: Transport
        +connect(): Promise~void~
        +disconnect(): Promise~void~
        +callTool(toolName: string, args: any): Promise~any~
        +listTools(): Promise~Tool[]
    }
    
    class MCPServerRegistry {
        -servers: Map~string, ServerConfig~
        +registerServer(config: ServerConfig): void
        +getServerConfig(name: string): ServerConfig
        +listServers(): string[]
    }
    
    MCPManager --> MCPClient
    MCPManager --> MCPServerRegistry
```

### 3. Agent Base Classes

```mermaid
classDiagram
    class MCPAgent {
        <<abstract>>
        #name: string
        #instruction: string
        #serverNames: string[]
        #functions: AgentFunction[]
        #mcpManager: MCPManager
        +execute(context: ProjectContext): Promise~AgentResult~
        +validate(context: ProjectContext): Promise~ValidationResult~
        +rollback(context: ProjectContext): Promise~void~
        #callTool(serverName: string, toolName: string, args: any): Promise~any~
        #updateProgress(progress: number): void
        #requestHumanInput(message: string): Promise~any~
    }
    
    class PMAgent {
        +generatePRD(requirements: string): Promise~PRDDocument~
        +clarifyRequirements(questions: Question[]): Promise~ClarificationResult~
        +validateRequirements(prd: PRDDocument): Promise~ValidationResult~
    }
    
    class UXAgent {
        +generateUXSpec(prd: PRDDocument, figmaUrl?: string): Promise~UXSpecification~
        +createDesignSystem(uxSpec: UXSpecification): Promise~DesignSystem~
        +generatePageDesigns(uxSpec: UXSpecification): Promise~PageDesign[]
    }
    
    class ArchitectAgent {
        +designArchitecture(prd: PRDDocument, uxSpec: UXSpecification): Promise~ArchitectureDesign~
        +selectTechStack(requirements: string): Promise~TechStack~
        +designAPIs(architecture: ArchitectureDesign): Promise~APISpecification~
        +designDatabase(architecture: ArchitectureDesign): Promise~DatabaseSchema~
    }
    
    class POAgent {
        +createEpics(prd: PRDDocument, architecture: ArchitectureDesign): Promise~Epic[]
        +createStories(epics: Epic[]): Promise~Story[]
        +prioritizeStories(stories: Story[]): Promise~Story[]
    }
    
    class DevAgent {
        +implementStory(story: Story, context: ProjectContext): Promise~ImplementationResult~
        +runTests(projectPath: string): Promise~TestResult~
        +fixBugs(bugReport: BugReport): Promise~FixResult~
        +deployProject(projectPath: string): Promise~DeploymentResult~
    }
    
    MCPAgent <|-- PMAgent
    MCPAgent <|-- UXAgent
    MCPAgent <|-- ArchitectAgent
    MCPAgent <|-- POAgent
    MCPAgent <|-- DevAgent
```

## 数据模型设计

### 核心数据模型

```typescript
// 项目上下文
interface ProjectContext {
  projectId: string;
  userId: string;
  projectPath: string;
  projectName: string;
  requirements: string;
  status: ProjectStatus;
  currentStage: DevStage;
  artifacts: ProjectArtifact[];
  dependencies: string[];
  metadata: Record<string, any>;
}

// Agent任务
interface AgentTask {
  id: string;
  projectId: string;
  userId: string;
  agentType: AgentType;
  stage: DevStage;
  status: TaskStatus;
  progress: number;
  parameters: Record<string, any>;
  context: ProjectContext;
  createdAt: Date;
  startedAt?: Date;
  completedAt?: Date;
  error?: string;
  result?: any;
  retryCount: number;
  maxRetries: number;
}

// Agent结果
interface AgentResult {
  success: boolean;
  artifacts: ProjectArtifact[];
  nextStage?: DevStage;
  dependencies?: string[];
  error?: string;
  metadata: Record<string, any>;
  requiresHumanInput?: boolean;
  humanInputMessage?: string;
}

// 项目工件
interface ProjectArtifact {
  id: string;
  type: ArtifactType;
  name: string;
  path: string;
  content: string;
  format: DocumentFormat;
  createdAt: Date;
  updatedAt: Date;
  version: number;
  dependencies: string[];
}

// 工作流步骤
interface WorkflowStep {
  id: string;
  agentName: string;
  dependencies: string[];
  condition?: WorkflowCondition;
  retryPolicy: RetryPolicy;
  timeout: number;
  parallel: boolean;
}

// 重试策略
interface RetryPolicy {
  maxRetries: number;
  retryDelay: number;
  backoffMultiplier: number;
  retryableErrors: string[];
}
```

## 关键流程设计

### 1. Agent协作流程

```mermaid
sequenceDiagram
    participant Backend as Backend API
    participant Orchestrator as Agent Orchestrator
    participant PM as PM Agent
    participant UX as UX Agent
    participant ARCH as Architect Agent
    participant PO as PO Agent
    participant DEV as Dev Agent
    participant MCP as MCP Manager
    participant FS as File System
    
    Backend->>Orchestrator: 启动项目开发流程
    Orchestrator->>PM: 执行需求分析任务
    PM->>MCP: 连接MCP服务器
    MCP-->>PM: 服务器连接成功
    PM->>MCP: 调用AI工具生成PRD
    MCP-->>PM: 返回PRD文档
    PM->>FS: 保存PRD文档
    PM->>Orchestrator: 任务完成，触发下一阶段
    
    Orchestrator->>UX: 执行UX设计任务
    UX->>MCP: 调用Figma API和AI工具
    MCP-->>UX: 返回UX规范
    UX->>FS: 保存UX规范
    UX->>Orchestrator: 任务完成，触发下一阶段
    
    Orchestrator->>ARCH: 执行架构设计任务
    ARCH->>MCP: 调用AI工具设计架构
    MCP-->>ARCH: 返回架构设计
    ARCH->>FS: 保存架构文档
    ARCH->>Orchestrator: 任务完成，触发下一阶段
    
    Orchestrator->>PO: 执行任务分解
    PO->>MCP: 调用AI工具创建Epics/Stories
    MCP-->>PO: 返回Epics和Stories
    PO->>FS: 保存任务文档
    PO->>Orchestrator: 任务完成，触发下一阶段
    
    Orchestrator->>DEV: 执行代码开发
    DEV->>MCP: 调用Cursor CLI和AI工具
    MCP-->>DEV: 返回源代码
    DEV->>FS: 保存源代码
    DEV->>Orchestrator: 任务完成
    Orchestrator->>Backend: 项目开发完成
```

### 2. MCP服务器管理流程

```mermaid
sequenceDiagram
    participant Agent as MCP Agent
    participant MCPManager as MCP Manager
    participant MCPClient as MCP Client
    participant MCPServer as MCP Server
    participant Transport as Transport Layer
    
    Agent->>MCPManager: 请求连接服务器
    MCPManager->>MCPClient: 创建客户端连接
    MCPClient->>Transport: 建立传输连接
    Transport->>MCPServer: 启动服务器进程
    MCPServer-->>Transport: 服务器就绪
    Transport-->>MCPClient: 连接建立
    MCPClient-->>MCPManager: 客户端就绪
    MCPManager-->>Agent: 服务器连接成功
    
    Agent->>MCPManager: 调用工具
    MCPManager->>MCPClient: 转发工具调用
    MCPClient->>Transport: 发送工具请求
    Transport->>MCPServer: 执行工具
    MCPServer-->>Transport: 返回结果
    Transport-->>MCPClient: 工具执行结果
    MCPClient-->>MCPManager: 返回结果
    MCPManager-->>Agent: 工具调用完成
    
    Agent->>MCPManager: 断开连接
    MCPManager->>MCPClient: 关闭客户端
    MCPClient->>Transport: 关闭传输
    Transport->>MCPServer: 终止服务器进程
    MCPServer-->>Transport: 服务器已关闭
    Transport-->>MCPClient: 传输已关闭
    MCPClient-->>MCPManager: 客户端已关闭
    MCPManager-->>Agent: 连接已断开
```

### 3. 错误处理和重试流程

```mermaid
stateDiagram-v2
    [*] --> PENDING: 创建任务
    
    PENDING --> RUNNING: 开始执行
    RUNNING --> COMPLETED: 执行成功
    RUNNING --> FAILED: 执行失败
    RUNNING --> CANCELLED: 用户取消
    RUNNING --> PAUSED: 需要人工干预
    
    FAILED --> RETRYING: 自动重试
    RETRYING --> RUNNING: 重试执行
    RETRYING --> FAILED: 重试次数超限
    
    PAUSED --> RUNNING: 人工干预完成
    PAUSED --> CANCELLED: 人工取消
    
    COMPLETED --> [*]: 任务结束
    CANCELLED --> [*]: 任务结束
    FAILED --> [*]: 任务结束
    
    note right of RUNNING
        执行过程中会定期更新进度
        支持实时状态推送
        支持人工干预请求
    end note
    
    note right of RETRYING
        支持指数退避重试
        可配置重试策略
        记录重试历史
    end note
```

## 服务层设计

### 1. 文件系统服务

```mermaid
classDiagram
    class FileSystemService {
        -basePath: string
        +readFile(path: string): Promise~string~
        +writeFile(path: string, content: string): Promise~void~
        +createDirectory(path: string): Promise~void~
        +copyTemplate(templatePath: string, targetPath: string): Promise~void~
        +replacePlaceholders(filePath: string, variables: Record~string, string~): Promise~void~
        +listFiles(directory: string): Promise~FileInfo[]~
        +deleteFile(path: string): Promise~void~
        +moveFile(source: string, destination: string): Promise~void~
    }
    
    class FileInfo {
        +name: string
        +path: string
        +type: 'file' | 'directory'
        +size: number
        +modifiedAt: Date
        +permissions: string
    }
```

### 2. 命令执行服务

```mermaid
classDiagram
    class CommandExecutionService {
        -config: CommandConfig
        +executeCommand(command: string, options: ExecOptions): Promise~ExecResult~
        +executeCursorCommand(projectPath: string, message: string): Promise~string~
        +executeNPMCommand(command: string, projectPath: string): Promise~string~
        +executeGitCommand(command: string, projectPath: string): Promise~string~
        +executeOllamaCommand(prompt: string, model: string): Promise~string~
        +executeFigmaCommand(figmaUrl: string, action: string): Promise~any~
    }
    
    class ExecResult {
        +stdout: string
        +stderr: string
        +exitCode: number
        +duration: number
        +success: boolean
    }
    
    class CommandConfig {
        +cursorCliPath: string
        +npmPath: string
        +gitPath: string
        +ollamaUrl: string
        +figmaApiKey: string
        +timeout: number
    }
```

### 3. 通知服务

```mermaid
classDiagram
    class NotificationService {
        -backendApiUrl: string
        -io: SocketIOServer
        +notifyBackend(event: NotificationEvent): Promise~void~
        +broadcastProgress(projectId: string, progress: ProgressUpdate): Promise~void~
        +sendErrorAlert(error: Error, context: ErrorContext): Promise~void~
        +sendTaskComplete(taskId: string, result: any): Promise~void~
        +sendHumanInputRequest(taskId: string, message: string): Promise~void~
    }
    
    class NotificationEvent {
        +type: NotificationType
        +projectId: string
        +taskId: string
        +data: any
        +timestamp: Date
    }
    
    class ProgressUpdate {
        +taskId: string
        +projectId: string
        +progress: number
        +message: string
        +stage: string
        +estimatedTimeRemaining?: number
    }
```

## 配置管理

### 1. 应用配置

```typescript
interface AppConfig {
  app: {
    port: number;
    nodeEnv: 'development' | 'production' | 'test';
    cors: CorsOptions;
    rateLimit: RateLimitOptions;
  };
  redis: {
    url: string;
    host: string;
    port: number;
    password?: string;
    db: number;
  };
  mcp: {
    servers: Record<string, ServerConfig>;
    timeout: number;
    retryAttempts: number;
  };
  projectDataPath: string;
  backendApiUrl: string;
  tools: CommandConfig;
  logging: {
    level: string;
    format: string;
    file?: string;
  };
}

interface ServerConfig {
  command: string;
  args: string[];
  description: string;
  timeout?: number;
  retryAttempts?: number;
  env?: Record<string, string>;
}
```

### 2. 环境变量配置

```bash
# 应用配置
NODE_ENV=development
PORT=3001
LOG_LEVEL=info

# Redis配置
REDIS_URL=redis://localhost:6379
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# 后端API配置
BACKEND_API_URL=http://localhost:8080

# 项目数据路径
PROJECT_DATA_PATH=F:/app-maker/app_data/projects

# MCP服务器配置
MCP_TIMEOUT=30000
MCP_RETRY_ATTEMPTS=3

# 工具配置
CURSOR_CLI_PATH=C:\Program Files\Cursor\cursor.exe
NPM_PATH=C:\Program Files\nodejs\npm.cmd
GIT_PATH=C:\Program Files\Git\bin\git.exe
OLLAMA_URL=http://localhost:11434
FIGMA_API_KEY=your_figma_api_key
```

## 部署架构

### 1. 本地部署架构

```mermaid
graph TB
    subgraph "Host Machine (Windows)"
        subgraph "Docker Containers"
            Backend[Go Backend<br/>:8080]
            Redis[Redis<br/>:6379]
            Postgres[PostgreSQL<br/>:5432]
        end
        
        subgraph "Local Services"
            AgentsServer[Agents Server<br/>:3001]
            Ollama[Ollama AI<br/>:11434]
            Cursor[Cursor CLI]
            Git[Git]
            NPM[NPM]
        end
        
        subgraph "File System"
            Projects[Project Files<br/>F:/app-maker/app_data/projects]
            Templates[Templates<br/>F:/app-maker/app_data/templates]
            Logs[Logs<br/>F:/app-maker/logs]
        end
    end
    
    AgentsServer --> Backend
    AgentsServer --> Redis
    AgentsServer --> Ollama
    AgentsServer --> Cursor
    AgentsServer --> Projects
    AgentsServer --> Templates
    
    Backend --> Postgres
    Backend --> Redis
```

### 2. 进程管理

```mermaid
graph TB
    subgraph "Process Management"
        PM2[PM2 Process Manager]
        Supervisor[Supervisor]
        Systemd[Systemd Service]
    end
    
    subgraph "Agents Server Process"
        MainProcess[Main Process<br/>Express + Socket.io]
        WorkerProcesses[Worker Processes<br/>Agent Controllers]
        MCPProcesses[MCP Server Processes<br/>Dynamic Spawning]
    end
    
    subgraph "Monitoring"
        HealthCheck[Health Check Endpoint]
        Metrics[Metrics Collection]
        Logging[Structured Logging]
    end
    
    PM2 --> MainProcess
    MainProcess --> WorkerProcesses
    MainProcess --> MCPProcesses
    
    MainProcess --> HealthCheck
    MainProcess --> Metrics
    MainProcess --> Logging
```

## 安全设计

### 1. 安全架构

```mermaid
graph TB
    subgraph "Security Layers"
        Auth[Authentication<br/>JWT Tokens]
        Authz[Authorization<br/>Role-based Access]
        Validation[Input Validation<br/>Schema Validation]
        Sanitization[Data Sanitization<br/>XSS Prevention]
        Encryption[Data Encryption<br/>TLS/HTTPS]
        Audit[Audit Logging<br/>Security Events]
    end
    
    subgraph "Agent Security"
        Isolation[Process Isolation<br/>Sandboxed Execution]
        ResourceLimits[Resource Limits<br/>CPU/Memory Limits]
        NetworkRestriction[Network Restrictions<br/>Allowed Hosts Only]
        FileSystemAccess[File System Access<br/>Restricted Paths]
    end
    
    subgraph "MCP Security"
        ServerValidation[MCP Server Validation<br/>Signature Verification]
        ToolRestrictions[Tool Restrictions<br/>Allowed Tools Only]
        DataFiltering[Data Filtering<br/>Sensitive Data Protection]
    end
```

### 2. 权限控制

```typescript
interface SecurityConfig {
  authentication: {
    jwtSecret: string;
    tokenExpiry: string;
    refreshTokenExpiry: string;
  };
  authorization: {
    roles: string[];
    permissions: Record<string, string[]>;
  };
  mcp: {
    allowedServers: string[];
    allowedTools: Record<string, string[]>;
    serverValidation: boolean;
  };
  execution: {
    maxConcurrentTasks: number;
    maxExecutionTime: number;
    memoryLimit: string;
    cpuLimit: string;
  };
  fileSystem: {
    allowedPaths: string[];
    restrictedPaths: string[];
    maxFileSize: number;
  };
}
```

## 监控和日志

### 1. 监控架构

```mermaid
graph TB
    subgraph "Metrics Collection"
        AppMetrics[Application Metrics<br/>Task Count, Success Rate]
        SystemMetrics[System Metrics<br/>CPU, Memory, Disk]
        BusinessMetrics[Business Metrics<br/>Project Completion Rate]
    end
    
    subgraph "Logging"
        StructuredLogs[Structured Logging<br/>JSON Format]
        LogLevels[Log Levels<br/>DEBUG, INFO, WARN, ERROR]
        LogRotation[Log Rotation<br/>Size-based Rotation]
    end
    
    subgraph "Alerting"
        ErrorAlerts[Error Alerts<br/>Critical Failures]
        PerformanceAlerts[Performance Alerts<br/>Slow Tasks]
        ResourceAlerts[Resource Alerts<br/>High Usage]
    end
    
    subgraph "Dashboards"
        HealthDashboard[Health Dashboard<br/>Service Status]
        PerformanceDashboard[Performance Dashboard<br/>Metrics Visualization]
        BusinessDashboard[Business Dashboard<br/>Project Analytics]
    end
```

### 2. 日志格式

```typescript
interface LogEntry {
  timestamp: string;
  level: 'DEBUG' | 'INFO' | 'WARN' | 'ERROR';
  message: string;
  context: {
    projectId?: string;
    taskId?: string;
    agentType?: string;
    userId?: string;
    requestId?: string;
  };
  metadata?: Record<string, any>;
  error?: {
    name: string;
    message: string;
    stack: string;
  };
}
```

## 性能优化

### 1. 性能策略

```mermaid
graph TB
    subgraph "并发控制"
        Semaphore[信号量控制<br/>限制并发任务数]
        Queue[任务队列<br/>Bull Queue + Redis]
        LoadBalancing[负载均衡<br/>多Worker进程]
    end
    
    subgraph "缓存策略"
        RedisCache[Redis缓存<br/>任务状态缓存]
        MemoryCache[内存缓存<br/>配置和模板缓存]
        FileCache[文件缓存<br/>生成结果缓存]
    end
    
    subgraph "资源优化"
        ConnectionPooling[连接池<br/>数据库和MCP连接]
        ResourceReuse[资源复用<br/>MCP服务器连接复用]
        LazyLoading[懒加载<br/>按需加载Agent]
    end
    
    subgraph "异步处理"
        NonBlocking[非阻塞操作<br/>异步I/O]
        Streaming[流式处理<br/>大文件处理]
        BackgroundTasks[后台任务<br/>定时清理]
    end
```

### 2. 性能指标

```typescript
interface PerformanceMetrics {
  taskExecution: {
    averageExecutionTime: number;
    successRate: number;
    failureRate: number;
    retryRate: number;
  };
  system: {
    cpuUsage: number;
    memoryUsage: number;
    diskUsage: number;
    networkLatency: number;
  };
  mcp: {
    serverConnectionTime: number;
    toolExecutionTime: number;
    serverAvailability: number;
  };
  business: {
    projectsCompletedPerHour: number;
    averageProjectCompletionTime: number;
    userSatisfactionScore: number;
  };
}
```

## 扩展性设计

### 1. 水平扩展

```mermaid
graph TB
    subgraph "负载均衡器"
        LB[Load Balancer<br/>Nginx/Traefik]
    end
    
    subgraph "Agents Server集群"
        AS1[Agents Server 1<br/>:3001]
        AS2[Agents Server 2<br/>:3002]
        AS3[Agents Server 3<br/>:3003]
    end
    
    subgraph "共享存储"
        Redis[Redis Cluster<br/>任务队列和状态]
        FileStorage[共享文件存储<br/>NFS/对象存储]
    end
    
    subgraph "MCP服务器池"
        MCPServer1[MCP Server 1]
        MCPServer2[MCP Server 2]
        MCPServer3[MCP Server 3]
    end
    
    LB --> AS1
    LB --> AS2
    LB --> AS3
    
    AS1 --> Redis
    AS2 --> Redis
    AS3 --> Redis
    
    AS1 --> FileStorage
    AS2 --> FileStorage
    AS3 --> FileStorage
    
    AS1 --> MCPServer1
    AS2 --> MCPServer2
    AS3 --> MCPServer3
```

### 2. 插件架构

```mermaid
classDiagram
    class PluginManager {
        -plugins: Map~string, Plugin~
        +registerPlugin(plugin: Plugin): void
        +unregisterPlugin(pluginName: string): void
        +executePlugin(pluginName: string, context: any): Promise~any~
        +listPlugins(): string[]
    }
    
    class Plugin {
        <<interface>>
        +name: string
        +version: string
        +execute(context: any): Promise~any~
        +validate(context: any): boolean
        +cleanup(): Promise~void~
    }
    
    class CustomAgentPlugin {
        +name: string
        +version: string
        +execute(context: any): Promise~any~
        +validate(context: any): boolean
        +cleanup(): Promise~void~
    }
    
    class CustomToolPlugin {
        +name: string
        +version: string
        +execute(context: any): Promise~any~
        +validate(context: any): boolean
        +cleanup(): Promise~void~
    }
    
    PluginManager --> Plugin
    Plugin <|-- CustomAgentPlugin
    Plugin <|-- CustomToolPlugin
```

## 总结

Agents Server 架构设计遵循以下核心原则：

1. **模块化设计**: 每个组件职责清晰，便于维护和扩展
2. **异步处理**: 基于队列的任务处理，支持高并发和容错
3. **实时通信**: WebSocket支持实时状态更新和进度反馈
4. **MCP集成**: 标准化的MCP服务器连接和管理
5. **可扩展性**: 支持水平扩展和插件化扩展
6. **本地部署**: 直接运行在主机上，支持GPU加速和文件系统访问
7. **安全设计**: 多层安全防护，确保系统安全
8. **监控完善**: 全面的监控和日志系统，便于运维管理

该架构为后续的开发工程师提供了清晰的实现指导，确保系统能够稳定、高效地运行，并支持未来的功能扩展和性能优化。