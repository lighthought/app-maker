# 简单使用指南

## ✅ 设置完成

现在你可以在 `backend` 和 `agents` 项目中直接使用共享模块了：

```go
import (
    "shared-models/agent"      // Agent 请求响应模型
    "shared-models/common"     // 通用响应和常量
    "shared-models/client"     // HTTP 客户端
    "shared-models/project"    // 项目相关模型
)
```

## 🔧 在 Backend 中使用

### 1. 更新 project_stage_service.go

```go
// backend/internal/services/project_stage_service.go
import (
    // ... 其他导入
    "shared-models/agent"
    "shared-models/client" 
    "shared-models/common"
)

type projectStageService struct {
    // ... 其他字段
    agentClient *client.AgentClient
}

func NewProjectStageService(...) ProjectStageService {
    // agents 服务地址
    agentsURL := utils.GetEnvOrDefault("AGENTS_SERVER_URL", "http://localhost:9090")
    agentClient := client.NewAgentClient(agentsURL, 5*time.Minute)
    
    return &projectStageService{
        // ... 其他字段
        agentClient: agentClient,
    }
}

// 简化的 PM Agent 调用
func (s *projectStageService) generatePRD(ctx context.Context, project *models.Project) (*common.AgentResult, error) {
    req := &agent.GetPRDReq{
        ProjectGuid:  project.GUID,
        Requirements: project.Requirements,
    }
    return s.agentClient.GetPRD(ctx, req)
}
```

### 2. 删除旧的 buildAgentRequest 方法

现在可以删除复杂的 `buildAgentRequest` 方法，直接使用客户端：

```go
// 替换旧的 invokeAgentSync 调用
func (s *projectStageService) generatePRD(ctx context.Context, project *models.Project, resultWriter *asynq.ResultWriter) error {
    // 使用共享客户端
    result, err := s.agentClient.GetPRD(ctx, &agent.GetPRDReq{
        ProjectGuid:  project.GUID,
        Requirements: project.Requirements,
    })
    
    if err != nil {
        return err
    }
    
    // 处理结果...
    return nil
}
```

## 🤖 在 Agents 中使用

### 1. 更新 Handler 模型

```go
// agents/internal/api/handlers/pm_handler.go
import (
    "shared-models/agent"  // 使用共享模型
    // 删除 "app-maker-agents/internal/models"
)

func (s *PmHandler) GetPRD(c *gin.Context) {
    var req agent.GetPRDReq  // 使用共享模型
    if err := c.ShouldBindJSON(&req); err != nil {
        utils.Error(c, http.StatusBadRequest, "参数校验失败: "+err.Error())
        return
    }
    // ... 处理逻辑
}
```

### 2. 删除重复模型

可以删除 `agents/internal/models/req.go` 文件，直接使用共享模型。

## 🚀 优势

1. **类型统一**：Backend 和 Agents 使用相同的数据结构
2. **维护简单**：只需要维护一份模型定义
3. **开发便捷**：修改模型后，两个项目自动同步
4. **无网络依赖**：本地引用，不依赖外部仓库
5. **调试友好**：可以直接修改共享模块进行调试

## 📝 日常使用

### 添加新的 Agent 接口

1. 在 `shared-models/agent/requests.go` 中添加请求模型
2. 在 `shared-models/client/agent_client.go` 中添加客户端方法
3. 在对应的项目中使用新接口

### 修改现有模型

直接在 `shared-models` 中修改，两个项目会自动使用最新版本。

## 🔧 故障排除

如果遇到导入问题：

```bash
# 清理模块缓存
go clean -modcache

# 重新下载依赖
cd backend && go mod tidy
cd agents && go mod tidy
```

## 📁 项目结构

```
app-maker/
├── shared-models/          # 共享模块
│   ├── go.mod             # module shared-models
│   ├── agent/             # Agent 请求模型
│   ├── common/            # 通用响应
│   ├── client/            # HTTP 客户端
│   └── project/           # 项目模型
├── backend/
│   ├── go.mod             # replace shared-models => ../shared-models
│   └── internal/
└── agents/
    ├── go.mod             # replace shared-models => ../shared-models
    └── internal/
```
