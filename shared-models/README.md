# App Maker 共享模块

App Maker 是一个基于多Agent协作的自动化软件开发平台。此 shared-models 模块为整个项目提供通用的数据模型、API客户端、工具函数等共享组件。

## 📦 模块特性

- **统一数据模型**：为 agents 和 backend 服务提供一致的请求响应结构
- **HTTP 客户端**：封装了与 agents 服务的通信逻辑
- **工具函数库**：提供密码、文件、时间、UUID等常用工具
- **认证服务**：JWT令牌生成和验证
- **异步任务**：基于 Asynq 的任务队列模型
- **日志管理**：统一的日志接口
- **类型安全**：编译时检查，避免运行时错误

## 📁 项目结构

```
shared-models/
├── agent/              # Agent 请求响应模型
│   ├── requests.go     # 请求结构体定义
│   ├── response.go     # 响应结构体定义
│   └── roles.go        # Agent 角色定义
├── auth/               # 认证相关
│   └── jwt.go          # JWT认证服务
├── client/             # HTTP 客户端工具
│   ├── agent_client.go # Agent 服务客户端
│   └── http_client.go   # HTTP 客户端封装
├── common/             # 通用常量和响应结构
│   ├── constants.go    # 常量定义
│   └── response.go     # 通用响应结构
├── logger/             # 日志管理
│   └── logger.go       # 结构化日志服务
├── tasks/              # 异步任务模型
│   ├── model.go        # 任务结果模型
│   └── task.go         # 任务创建和管理函数
├── utils/              # 工具函数集合
│   ├── ai_utils.go     # AI相关工具
│   ├── env_utils.go    # 环境变量工具
│   ├── file_utils.go   # 文件操作工具
│   ├── password_utils.go # 密码工具
│   ├── response_utils.go # 响应工具
│   ├── time_utils.go   # 时间工具
│   ├── uuid_utils.go   # UUID工具
│   └── zip_utils.go    # 压缩工具
└── go.mod             # Go 模块定义
```

## 🚀 快速开始

### 安装依赖

```bash
cd shared-models
go mod tidy
```

### 在项目中使用

在任何项目中直接导入：

```go
import (
    "shared-models/agent"      // Agent 请求响应模型
    "shared-models/common"      // 通用响应和常量
    "shared-models/client"      // HTTP 客户端
    "shared-models/auth"        // JWT认证服务
    "shared-models/utils"       // 工具函数
)
```

### Backend 中使用客户端

```go
// 创建 Agent 客户端
agentClient := client.NewAgentClient("http://localhost:8090", 5*time.Minute)

// 调用 Agent 服务生成 PRD
result, err := agentClient.GetPRD(ctx, &agent.GetPRDReq{
    ProjectGuid: "1234567890",
    Requirements: "创建一个在线购物平台",
})

if err != nil {
    log.Printf("PRD 生成失败: %v", err)
    return
}

log.Printf("PRD 生成成功: %s", result.Message)
```

### Agents 中使用请求模型

```go
func (h *PmHandler) GetPRD(c *gin.Context) {
    var req agent.GetPRDReq
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "参数验证失败"))
        return
    }
    
    // 处理 PRD 生成逻辑
    taskID := uuid.NewString()
    // ...
    
    c.JSON(http.StatusOK, utils.GetSuccessResponse("PRD 生成任务已提交", taskID))
}
```

## 📖 API 参考

### Agent 请求类型

| 请求类型 | 描述 | 对应 Agent |
|---------|------|-----------|
| `SetupProjEnvReq` | 项目环境准备 | BMad Master |
| `GetProjBriefReq` | 项目简介分析 | Analyst |
| `GetPRDReq` | 产品需求文档 | PM |
| `GetUXStandardReq` | UX标准设计 | UX Expert |
| `GetArchitectureReq` | 系统架构设计 | Architect |
| `GetDatabaseDesignReq` | 数据库设计 | Architect |
| `GetAPIDefinitionReq` | API接口定义 | Architect |
| `GetEpicsAndStoriesReq` | 史诗和故事划分 | PO |
| `ImplementStoryReq` | 用户故事实现 | Dev |
| `FixBugReq` | Bug修复 | Dev |
| `.RunTestReq` | 测试执行 | Dev |
| `DeployReq` | 项目部署 | Dev |

### 开发阶段常量

系统定义了完整的项目开发阶段：

- `initializing`: 等待开始开发
- `setup_environment`: 正在初始化开发环境
- `check_requirement`: 正在检查需求
- `generate_prd`: 正在生成PRD文档
- `define_ux_standard`: 正在定义UX标准
- `design_architecture`: 正在设计系统架构
- `define_data_model`: 正在定义数据模型
- `define_api`: 正在定义API接口
- `plan_epic_and_story`: 正在划分Epic和Story
- `develop_story`: 正在开发Story功能
- `fix_bug`: 正在修复开发问题
- `run_test`: 正在执行自动测试
- `deploy`: 正在部署项目
- `done`: 项目开发完成

### Agent 角色定义

系统支持多种AI Agent角色：

- **Analyst (Mary)**: 需求分析师
- **PM (John)**: 产品经理  
- **UX Expert (Sally)**: 用户体验专家
- **Architect (Winston)**: 架构师
- **PO (Sarah)**: 产品负责人
- **Dev (James)**: 开发工程师
- **QA (Quinn)**: 测试和质量工程师
- **SM (Bob)**: 敏捷教练
- **BMAD Master**: BMad管理员

## 🔧 技术栈

- **Go 1.24+**: 主要编程语言
- **ASInq**: 异步任务队列
- **JWT**: JSON Web Token认证
- **Zap**: 高性能日志库
- **UUID**: 全局唯一标识符
- **Viper**: 配置管理
- **Ollama**: 本地AI模型支持
- **DeepSeek**: AI API集成

## 📝 开发指南

### 添加新的请求模型

1. 在 `agent/requests.go` 中添加新的请求结构体
2. 在 `client/agent_client.go` 中添加对应的客户端方法
3. 如需处理响应，在 `agent/response.go` 中添加响应结构体
4. 更新路由和处理器

### 修改现有模型

直接修改对应文件，所有引用项目会自动更新（相对路径依赖）。

### 故障排除

如果遇到导入问题：

```bash
# 清理依赖缓存
cd backend && go clean -modcache && go mod tidy
cd agents && go clean -modcache && go mod tidy
```

## 🌟 优势

1. **简单直接**：使用相对路径，无需复杂配置
2. **本地开发**：修改即生效，无需发布版本  
3. **类型安全**：编译时检查，避免运行时错误
4. **统一模型**：Backend 和 Agents 使用相同的数据结构
5. **零依赖冲突**：独立模块，不干扰主项目依赖
6. **高可维护性**：单一职责，易于理解和修改

## 📄 许可证

本项目采用 AGPLv3 许可证 - 查看 [LICENSE](..\LICENSE) 文件了解详情。如果您希望在不遵守AGPL条款的项目中集成本代码，需要另行购买商业许可，请联系我。

---
## 联系方式

- 维护者: AI探趣星船长
- 邮箱: qqjack2012@gmail.com
- 项目地址: https://github.com/lighthought/app-maker