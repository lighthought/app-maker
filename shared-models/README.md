# App Maker 共享模块

## ✅ 已完成设置

使用简单的相对路径引用方式，backend 和 agents 项目都通过 `../shared-models` 引用此模块。

## 📁 项目结构

```
app-maker/
├── shared-models/          # 共享模块 (module shared-models)
│   ├── go.mod
│   ├── agent/             # Agent 请求响应模型
│   ├── common/            # 通用响应和常量  
│   ├── client/            # HTTP 客户端工具
│   ├── project/           # 项目相关模型
│   └── examples/          # 使用示例
├── backend/               # 后端服务
│   ├── go.mod            # replace shared-models => ../shared-models
│   └── internal/
└── agents/               # Agent 服务
    ├── go.mod            # replace shared-models => ../shared-models
    └── internal/
```

## 🚀 立即使用

在任何项目中直接导入：

```go
import (
    "shared-models/agent"      // Agent 请求响应模型
    "shared-models/common"     // 通用响应和常量
    "shared-models/client"     // HTTP 客户端
    "shared-models/project"    // 项目相关模型
)
```

## 📖 详细文档

- [简单使用指南](SIMPLE_USAGE.md) - 如何在项目中使用
- [使用示例](examples/backend_usage.go) - 完整的代码示例

## 🔧 优势

1. **简单直接**：使用相对路径，无需复杂配置
2. **本地开发**：修改即生效，无需发布版本
3. **类型安全**：编译时检查，避免运行时错误
4. **统一模型**：Backend 和 Agents 使用相同的数据结构
5. **零依赖**：不依赖外部仓库或网络

## 📝 常用操作

### 添加新的请求模型
1. 在 `agent/requests.go` 中添加结构体
2. 在 `client/agent_client.go` 中添加对应方法
3. 在项目中直接使用

### 修改现有模型
直接修改对应文件，所有引用项目自动更新

### 故障排除
如果遇到导入问题：
```bash
cd backend && go mod tidy
cd agents && go mod tidy
```

## ✨ 示例

### Backend 中使用客户端
```go
agentClient := client.NewAgentClient("http://localhost:9090", 5*time.Minute)
result, err := agentClient.GetPRD(ctx, &agent.GetPRDReq{
    ProjectGuid: "123",
    Requirements: "需求描述",
})
```

### Agents 中使用请求模型
```go
func (h *PmHandler) GetPRD(c *gin.Context) {
    var req agent.GetPRDReq
    if err := c.ShouldBindJSON(&req); err != nil {
        // 处理错误
    }
    // 处理请求
}
```

这就是最简单、最直接的共享模块方案！🎉