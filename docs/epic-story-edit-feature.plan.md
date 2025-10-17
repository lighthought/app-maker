# Epic 和 Story 编辑功能实现计划

## 一、数据库与模型扩展

### 1.1 用户配置表扩展

- 在 `users` 表添加 `auto_go_next` 字段（boolean，默认 false）
- 更新 `backend/internal/models/user.go` 添加 `AutoGoNext` 字段
- 更新 `backend/internal/models/request.go` 中的 `UpdateUserSettingsRequest` 添加 `AutoGoNext` 字段
- 创建数据库迁移脚本 `backend/scripts/migration-add-auto-go-next-fields.sql`

### 1.2 项目表扩展

- 在 `projects` 表添加 `waiting_for_user_confirm` 字段（boolean）标识是否等待用户确认
- 在 `projects` 表添加 `confirm_stage` 字段（varchar）记录等待确认的阶段
- **在 `projects` 表添加 `auto_go_next` 字段（boolean）**，允许单个项目覆盖用户全局设置
- 更新 `backend/internal/models/project.go` 添加相关字段

### 1.3 Epic/Story 模型扩展

- 在 `project_epics` 表添加 `display_order` 字段（int）用于排序
- 在 `epic_stories` 表添加 `display_order` 字段（int）用于排序
- 更新 `backend/internal/models/epic.go` 添加排序字段

## 二、Backend 状态机重构

### 2.1 新增任务类型常量

在 `shared-models/common/constants.go` 添加：

- `TaskTypeStageCheckRequirement = "stage:check_requirement"`
- `TaskTypeStageGeneratePRD = "stage:generate_prd"`
- `TaskTypeStageDefineUXStandard = "stage:define_ux_standard"`
- `TaskTypeStageDesignArchitecture = "stage:design_architecture"`
- `TaskTypeStagePlanEpicAndStory = "stage:plan_epic_and_story"`
- `TaskTypeStageDefineDataModel = "stage:define_data_model"`
- `TaskTypeStageDefineAPI = "stage:define_api"`
- `TaskTypeStageGeneratePages = "stage:generate_pages"`
- `TaskTypeStageDevelopStory = "stage:develop_story"`
- `TaskTypeStageRunTest = "stage:run_test"`
- `TaskTypeStageDeploy = "stage:deploy"`

### 2.2 创建任务构造函数

在 `shared-models/tasks/task.go` 添加每个阶段的任务构造函数

### 2.3 重构 ProjectStageService

修改 `backend/internal/services/project_stage_service.go`：

**需要用户确认的阶段**（按优先级）：

- `handleGeneratePRDTask`：生成 PRD，**必须等待用户确认**
- `handleDefineUXStandardTask`：定义 UX 标准，**需要用户确认**
- `handleDesignArchitectureTask`：设计架构，**需要用户确认**
- `handlePlanEpicAndStoryTask`：划分 Epic/Story，**必须等待用户确认**（Epic/Story 编辑界面）
- `handleDefineDataModelTask`：定义数据模型，**需要用户确认**
- `handleDefineAPITask`：定义 API，**需要用户确认**
- `handleDevelopStoryTask`：开发 Story，**需要用户确认**

**确认逻辑**：

```go
func (s *projectStageService) proceedToNextStage(ctx context.Context, project *models.Project, currentStage common.DevStatus, requireConfirm bool) error {
    // 优先使用项目级配置，其次用户级配置
    autoGoNext := project.AutoGoNext
    if !autoGoNext {
        autoGoNext = project.User.AutoGoNext
    }
    
    if requireConfirm && !autoGoNext {
        project.WaitingForUserConfirm = true
        project.ConfirmStage = string(currentStage)
        s.projectRepo.Update(ctx, project)
        s.webSocketService.NotifyUserConfirmRequired(project.GUID, currentStage)
        return nil
    }
    
    nextStage := getNextStage(currentStage)
    task := createTaskForStage(nextStage, project)
    s.asyncClient.Enqueue(task)
    return nil
}
```

### 2.4 更新 ProcessTask 分发

在 `ProcessTask` 方法中添加所有新任务类型的 case 分支

## 三、Redis Pub/Sub 机制

### 3.1 定义消息格式

在 `shared-models/common/constants.go` 定义频道常量

在 `shared-models/agent/response.go` 添加 `AgentTaskStatusMessage` 结构体

### 3.2 Agent 端发布消息

在 `agents/internal/services/` 各 handler 执行完任务后发布状态消息

### 3.3 Backend 端订阅消息

创建 `backend/internal/services/redis_pubsub_service.go`

在 `backend/cmd/server/main.go` 启动时初始化并启动订阅服务

### 3.4 取消轮询改为订阅

修改调用 `agentClient.ChatWithAgent` 的地方，移除轮询，改为等待 Pub/Sub 消息

## 四、Epic/Story 编辑 API

### 4.1 后端接口

在 `backend/internal/api/handlers/epic_handler.go` 添加：

- `UpdateEpicOrder`：更新 Epic 排序
- `UpdateEpic`：更新 Epic 内容
- `DeleteEpic`：删除 Epic
- `UpdateStoryOrder`：更新 Story 排序  
- `UpdateStory`：更新 Story 内容
- `DeleteStory`：删除 Story（**支持删除单个 Story**）
- `BatchDeleteStories`：**批量删除 Stories**
- `ConfirmEpicsAndStories`：用户确认 Epics 和 Stories，继续执行流程

### 4.2 Service 层

在 `backend/internal/services/epic_service.go` 添加相应的业务逻辑方法

### 4.3 Repository 层

在 `backend/internal/repositories/epic_repository.go` 和 `story_repository.go` 添加数据库操作方法

### 4.4 路由注册

在 `backend/internal/api/routes/routes.go` 注册新接口

## 五、同步到项目文件

### 5.1 Epic/Story 文件同步服务

在 `backend/internal/services/file_service.go` 添加 `SyncEpicsToFiles` 方法：

**关键逻辑**：

```go
func (s *fileService) SyncEpicsToFiles(ctx context.Context, projectPath string, epics []*models.Epic) error {
    storiesDir := filepath.Join(projectPath, "docs/stories")
    
    // 1. 获取当前存在的所有 epic 文件
    existingFiles, _ := filepath.Glob(filepath.Join(storiesDir, "epic*.md"))
    existingFileMap := make(map[string]bool)
    for _, f := range existingFiles {
        existingFileMap[filepath.Base(f)] = true
    }
    
    // 2. 写入数据库中的 epics
    for _, epic := range epics {
        filePath := filepath.Join(storiesDir, epic.FilePath)
        content := generateEpicMarkdown(epic)
        os.WriteFile(filePath, []byte(content), 0644)
        delete(existingFileMap, epic.FilePath)
    }
    
    // 3. 删除用户已删除的 epic 文件
    for fileName := range existingFileMap {
        os.Remove(filepath.Join(storiesDir, fileName))
    }
    
    return nil
}
```

### 5.2 确认时触发同步

在 `ConfirmEpicsAndStories` 接口中，保存到数据库后调用文件同步

## 六、前端编辑界面

### 6.1 Epic/Story 编辑组件

创建 `frontend/src/components/EpicStoryEditor.vue`：

功能包括：

- 使用 `vue-draggable-next` 实现拖拽排序
- **响应式布局设计，支持移动端操作**
- Epic 列表展示和编辑（名称、描述、优先级、预估天数）
- Story 列表展示和编辑（编号、标题、描述、优先级、依赖、技术栈）
- 删除 Epic/Story（带确认弹窗）
- **支持批量选择和删除 Stories**
- **在对话界面中美观展示，良好的用户体验**
- 底部操作按钮：
  - "重新生成"：调用 PO Agent 重新生成
  - "确认并继续"：保存并触发下一阶段
  - "跳过确认"：直接继续（不修改）

### 6.2 集成到 ProjectEdit 页面

修改 `frontend/src/pages/ProjectEdit.vue`：

- 根据 `project.confirm_stage` 在对话容器中显示不同的确认界面
- 当 `confirm_stage === 'plan_epic_and_story'` 时在对话流中插入 Epic/Story 编辑器
- 当 `confirm_stage === 'generate_prd'` 时显示 PRD 确认界面
- 当 `confirm_stage === 'define_ux_standard'` 时显示 UX 标准确认界面
- 当 `confirm_stage === 'design_architecture'` 时显示架构设计确认界面
- 当 `confirm_stage === 'define_data_model'` 时显示数据模型确认界面
- 当 `confirm_stage === 'define_api'` 时显示 API 定义确认界面
- 当 `confirm_stage === 'develop_story'` 时显示 Story 完成确认界面

### 6.3 WebSocket 消息处理

在 `frontend/src/utils/websocket.ts` 或相关组件中：

- 监听 `user_feedback_required` 消息类型
- 收到消息后更新项目状态，在对话流中显示确认界面

### 6.4 API 调用

在 `frontend/src/utils/http.ts` 或相关 API 文件添加：

- Epic 相关：`updateEpicOrder`, `updateEpic`, `deleteEpic`
- Story 相关：`updateStoryOrder`, `updateStory`, `deleteStory`, `batchDeleteStories`
- 确认接口：`confirmEpicsAndStories`

## 七、用户设置界面

### 7.1 添加自动继续开关

在用户设置页面添加：

- "自动进入下一阶段（YOLO 模式）"开关
- 说明：开启后将跳过所有确认步骤，自动执行到项目完成

### 7.2 使用已有的 UpdateUserSettings 接口

- 后端已有 `UpdateUserSettings` 接口
- 更新 `UpdateUserSettingsRequest` 模型添加 `auto_go_next` 字段
- 更新 `UserService.UpdateUserSettings` 方法处理新字段

## 八、测试与验证

### 8.1 数据库迁移测试

- 执行迁移脚本
- 验证所有新字段正确添加

### 8.2 状态机流程测试

- 创建新项目
- 验证每个阶段按顺序执行
- 验证关键阶段正确暂停等待用户确认
- 测试项目级和用户级 `auto_go_next` 配置优先级

### 8.3 Epic/Story 编辑测试

- 测试拖拽排序
- 测试编辑保存
- 测试删除单个 Epic/Story
- 测试批量删除 Stories
- 测试移动端响应式布局
- 验证确认后同步到文件，文件正确生成和删除

### 8.4 Redis Pub/Sub 测试

- 验证 Agent 正确发布消息
- 验证 Backend 正确接收并处理消息
- 测试异常情况（消息丢失、重复等）

### 8.5 YOLO 模式测试

- 开启用户级 auto_go_next
- 创建新项目
- 验证自动执行到完成，无需用户干预
- 测试项目级配置覆盖用户级配置


# To-Dos
- [ ] 1.数据库扩展：添加用户 auto_go_next 字段、项目等待确认字段、Epic/Story 排序字段，创建并执行迁移脚本
- [ ] 2.新增任务类型常量和任务构造函数
- [ ] 3.重构 ProjectStageService：拆分 handleProjectDevelopmentTask 为独立的阶段处理方法，实现状态机流程控制
- [ ] 4.定义 Redis Pub/Sub 消息格式和常量
- [ ] 5.Agent 端实现任务状态发布：在各 handler 完成时发布状态消息
- [ ] 6.Backend 端实现订阅服务：创建 RedisPubSubService，订阅并处理 Agent 任务状态消息
- [ ] 7.实现 Epic/Story 编辑 API：Handler、Service、Repository 三层，包括排序、更新、删除、确认接口
- [ ] 8.实现 Epic/Story 文件同步服务：将数据库数据同步到项目 stories 文件
- [ ] 9.前端实现 Epic/Story 编辑组件：拖拽排序、编辑、删除、确认功能
- [ ] 10.前端集成：在项目详情页根据确认阶段显示相应界面，处理 WebSocket 消息
- [ ] 11.用户设置：前端添加 auto_go_next 开关，后端实现用户配置更新接口
- [ ] 12.集成测试：测试完整流程、状态机、编辑功能、文件同步、YOLO 模式