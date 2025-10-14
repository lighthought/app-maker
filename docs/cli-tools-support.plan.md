# CLI 工具支持实现计划

## 1. 移除不支持的 CLI 工具

### 1.1 前端移除

**文件**: `frontend/src/components/UserSettingsModal.vue`

- 从 `cliToolOptions` 数组中移除 `iflow-cli` 和 `auggie-cli`
- 保留 `claude-code`, `qwen-code`, `gemini`

### 1.2 Shared-models 更新

**文件**: `shared-models/common/constants.go`

- 移除 `CliToolIFlowCli` 和 `CliToolAuggieCli` 常量
- 更新 `SupportedCliTools` 数组，只包含 `claude-code`, `qwen-code`, `gemini`

## 2. 在 Agent 请求中添加 CLI 类型参数

### 2.1 更新请求结构体

**文件**: `shared-models/agent/requests.go`

- 在所有请求结构体中添加 `CliTool string` 字段：
  - `GetProjBriefReq`
  - `GetPRDReq` 
  - `GetUXStandardReq`
  - `GetArchitectureReq`
  - `GetDatabaseDesignReq`
  - `GetAPIDefinitionReq`
  - `GetEpicsAndStoriesReq`
  - `ImplementStoryReq`
  - `FixBugReq`
  - `RunTestReq`
  - `DeployReq`

### 2.2 Backend 传递 CLI 类型

**文件**: `backend/internal/services/project_stage_service.go`

- 在所有调用 Agent 的方法中，从 `project` 获取 CLI 类型并传递：
  - `checkRequirement` → `agentClient.AnalyseProjectBrief`
  - `generatePRD` → `agentClient.GetPRD`
  - `defineUXStandards` → `agentClient.GetUXStandard`
  - `designArchitecture` → `agentClient.GetArchitecture`
  - `defineDataModel` → `agentClient.GetDatabaseDesign`
  - `defineAPIs` → `agentClient.GetAPIDefinition`
  - `planEpicsAndStories` → `agentClient.GetEpicsAndStories`
  - `developStories` → `agentClient.ImplementStory`
  - `fixBugs` → `agentClient.FixBug`
  - `runTests` → `agentClient.RunTest`
  - `packageProject` → `agentClient.Deploy`

获取 CLI 类型逻辑（复用现有的优先级）：

```go
cliTool := project.CliTool
if cliTool == "" {
    cliTool = project.User.DefaultCliTool
}
if cliTool == "" {
    cliTool = common.CliToolClaudeCode
}
```

## 3. Agents 服务支持不同 CLI 工具

### 3.1 CLI 类型检测和默认值

**文件**: `agents/internal/services/project_service.go`

在 `SetupProjectEnvironment` 方法中的 CLI 类型处理逻辑：

1. **优先使用请求参数**: 如果 `req.BmadCliType` 不为空，使用该值
2. **检测本地目录**: 如果请求参数为空，检测项目根目录：

   - 存在 `.claude` 目录 → `claude-code`
   - 存在 `.qwen` 目录 → `qwen-code`
   - 存在 `.gemini` 目录 → `gemini`
   - 都不存在 → 默认 `claude-code`

3. **验证 CLI 工具**: 检查请求的 CLI 类型对应的目录是否存在：

   - 如果不存在，先执行 `npx bmad-method install -f -i <cli-type> -d .` 安装
   - 然后继续后续操作
```go
func (s *projectService) SetupProjectEnvironment(ctx context.Context, req *agent.SetupProjEnvReq) error {
    cliTool := req.BmadCliType
    
    // 如果请求参数为空，检测本地目录
    if cliTool == "" {
        cliTool = s.DetectCliTool(req.ProjectGuid)
    }
    
    // 检查 CLI 工具对应的目录是否存在
    projectPath := s.getProjectPath(req.ProjectGuid)
    cliDirMap := map[string]string{
        common.CliToolClaudeCode: ".claude",
        common.CliToolQwenCode:   ".qwen",
        common.CliToolGemini:     ".gemini",
    }
    
    cliDir := cliDirMap[cliTool]
    if !utils.FileExists(filepath.Join(projectPath, cliDir)) {
        // 执行 bmad-method install
        s.commandService.SimpleExecute(ctx, req.ProjectGuid, "npx", "bmad-method", "install", "-f", "-i", cliTool, "-d", ".")
    }
    
    // ... 后续操作
}
```


### 3.2 命令执行适配

**文件**: `agents/internal/services/agent_task_service.go`

在 `innerProcessTask` 方法中，根据 CLI 类型构建不同的命令和处理输出：

```go
// 从 payload 或项目检测获取 CLI 类型
cliTool := payload.CliTool
if cliTool == "" {
    cliTool = h.projectService.DetectCliTool(payload.ProjectGUID)
}

// 根据 CLI 类型构建命令
var cliCommand string
var args []string
var useJsonOutput bool

switch cliTool {
case common.CliToolClaudeCode:
    cliCommand = "claude"
    useJsonOutput = true
    args = []string{"--dangerously-skip-permissions", "--output-format", "json", "-p", payload.Message}
    if sessionID != "" {
        args = []string{"--dangerously-skip-permissions", "--resume", sessionID, "--output-format", "json", "-p", payload.Message}
    }

case common.CliToolQwenCode:
    cliCommand = "qwen"
    useJsonOutput = false
    args = []string{"-y", "-p", payload.Message}

case common.CliToolGemini:
    cliCommand = "gemini"
    useJsonOutput = false
    args = []string{"-y", "-p", payload.Message}

default:
    cliCommand = "claude"
    useJsonOutput = true
    args = []string{"--dangerously-skip-permissions", "--output-format", "json", "-p", payload.Message}
}

result = h.commandService.SimpleExecute(ctx, payload.ProjectGUID, cliCommand, args...)

// 根据输出格式处理结果
if useJsonOutput {
    // 处理 JSON 输出（现有逻辑）
    // 尝试解析 JSON，失败则使用原始输出
} else {
    // 处理纯文本输出（qwen、gemini）
    // 直接使用原始输出文本
}
```

### 3.3 Prompt 适配不同 CLI

**文件**: `agents/internal/api/handlers/analyse_handler.go`, `pm_handler.go`, `ux_handler.go`, `architect_handler.go`, `po_handler.go`, `dev_handler.go`

在每个 Handler 的方法中，根据 CLI 类型调整 prompt：

```go
func (h *AnalyseHandler) ProjectBrief(c *gin.Context) {
    var req agent.GetProjBriefReq
    // ... 参数绑定
    
    var message string
    if req.CliTool == common.CliToolGemini {
        message = "@.bmad-core/agents/analyst.md 请你为我生成项目简介，再执行市场研究。输出对应的文档到 docs/analyse/ 目录下。\n" + 
            "项目需求：" + req.Requirements
    } else {
        message = "@bmad/analyst.mdc 请你为我生成项目简介，再执行市场研究。输出对应的文档到 docs/analyse/ 目录下。\n" + 
            "项目需求：" + req.Requirements
    }
    
    // ... 后续处理
}
```

需要更新的 Handler 方法：

- `analyse_handler.go`: `ProjectBrief` → `@.bmad-core/agents/analyst.md` (gemini)
- `pm_handler.go`: `GetPRD` → `@.bmad-core/agents/pm.md` (gemini)
- `ux_handler.go`: `GetUXStandard` → `@.bmad-core/agents/ux-expert.md` (gemini)
- `architect_handler.go`: `GetArchitecture`, `GetDatabase`, `GetAPIDefinition` → `@.bmad-core/agents/architect.md` (gemini)
- `po_handler.go`: `GetEpicsAndStories` → `@.bmad-core/agents/po.md` (gemini)
- `dev_handler.go`: `ImplementStory`, `FixBug`, `RunTest`, `Deploy` → `@.bmad-core/agents/dev.md` (gemini)

## 4. 健康检查实现

### 4.1 Agents 健康检查

**文件**: `agents/internal/api/handlers/health.go`

实现完整的健康检查：

```go
type HealthCheckResult struct {
    Status      string            `json:"status"`
    Version     string            `json:"version"`
    Environment map[string]string `json:"environment"`
}

func HealthCheck(c *gin.Context) {
    result := HealthCheckResult{
        Status:      "running",
        Version:     "1.0.0",
        Environment: make(map[string]string),
    }
    
    // 检查 Node.js
    if version, err := exec.Command("node", "--version").Output(); err == nil {
        result.Environment["node"] = strings.TrimSpace(string(version))
    } else {
        result.Environment["node"] = "not found"
    }
    
    // 检查 npx
    if version, err := exec.Command("npx", "--version").Output(); err == nil {
        result.Environment["npx"] = strings.TrimSpace(string(version))
    } else {
        result.Environment["npx"] = "not found"
    }
    
    // 检查 claude-code
    if version, err := exec.Command("claude", "--version").Output(); err == nil {
        result.Environment["claude-code"] = strings.TrimSpace(string(version))
    } else {
        result.Environment["claude-code"] = "not installed"
    }
    
    // 检查 qwen-code
    if version, err := exec.Command("qwen", "--version").Output(); err == nil {
        result.Environment["qwen-code"] = strings.TrimSpace(string(version))
    } else {
        result.Environment["qwen-code"] = "not installed"
    }
    
    // 检查 gemini
    if version, err := exec.Command("gemini", "--version").Output(); err == nil {
        result.Environment["gemini"] = strings.TrimSpace(string(version))
    } else {
        result.Environment["gemini"] = "not installed"
    }
    
    c.JSON(http.StatusOK, utils.GetSuccessResponse("App Maker Agents is running", result))
}
```

### 4.2 更新健康检查响应结构

**文件**: `shared-models/agent/response.go`

```go
type AgentHealthResp struct {
    Status      string            `json:"status"`
    Version     string            `json:"version"`
    Environment map[string]string `json:"environment"`
}
```

### 4.3 Backend 调用 Agent 健康检查

**文件**: `backend/internal/api/handlers/health.go`

```go
type BackendHealthResp struct {
    Status       string                   `json:"status"`
    Version      string                   `json:"version"`
    AgentService *agent.AgentHealthResp   `json:"agent_service"`
}

func HealthCheck(c *gin.Context) {
    resp := BackendHealthResp{
        Status:  "running",
        Version: "1.0.0",
    }
    
    // 调用 Agent 健康检查
    agentsURL := utils.GetEnvOrDefault("AGENTS_SERVER_URL", "http://host.docker.internal:8088")
    agentClient := client.NewAgentClient(agentsURL, 5*time.Second)
    
    if agentHealth, err := agentClient.HealthCheck(context.Background()); err == nil {
        resp.AgentService = agentHealth
    } else {
        resp.AgentService = &agent.AgentHealthResp{
            Status: "unavailable",
        }
    }
    
    c.JSON(http.StatusOK, utils.GetSuccessResponse("AutoCodeWeb Backend is running", resp))
}
```

### 4.4 前端显示 Agent 状态

**文件**: `frontend/src/pages/Dashboard.vue`

在仪表板页面添加 Agent 状态显示卡片：

- 显示 Agent 服务状态
- 显示已安装的 CLI 工具和版本
- 显示 Node.js、npx 版本

## 5. 补充修改

### 5.1 Agent 任务 Payload 结构

**文件**: `shared-models/tasks/task.go`

- 在 `AgentExecuteTaskPayload` 中添加 `CliTool string` 字段

### 5.2 项目服务添加 CLI 检测方法

**文件**: `agents/internal/services/project_service.go`

```go
func (s *projectService) DetectCliTool(projectGuid string) string {
    projectPath := s.getProjectPath(projectGuid)
    
    if utils.FileExists(filepath.Join(projectPath, ".claude")) {
        return common.CliToolClaudeCode
    }
    if utils.FileExists(filepath.Join(projectPath, ".qwen")) {
        return common.CliToolQwenCode
    }
    if utils.FileExists(filepath.Join(projectPath, ".gemini")) {
        return common.CliToolGemini
    }
    
    return common.CliToolClaudeCode // 默认
}
```