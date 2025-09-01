# Epic 1: 项目创建和 BMad-Method 集成 - 用户故事

## 概述
本文档包含 Epic 1 "项目创建和 BMad-Method 集成" 的详细用户故事。该 Epic 负责实现项目创建的完整流程，集成 BMad-Method 作为多 Agent 协作的核心引擎。

## 用户故事列表

### Story 1.1: 项目初始化
**优先级**: P0  
**估算工时**: 3 天  
**负责人**: 后端开发工程师

#### 用户故事
作为系统管理员，我希望能够创建新的项目目录结构，以便为后续的开发流程提供基础环境。

#### 验收标准
- [x] 支持通过 API 创建新项目
- [x] 自动生成标准的项目目录结构
- [x] 创建项目配置文件（如 .env, docker-compose.yml 等）
- [x] 支持项目名称、描述、类型等基本信息设置
- [x] 返回唯一的项目ID和创建状态
- [x] 支持后端端口和前端端口配置
- [x] 支持模板文件中的占位符替换

#### 技术要点
- 使用 Go 的 `os` 包创建目录结构
- 支持模板化的配置文件生成
- 实现项目元数据存储到数据库
- 从 `template.zip` 提取项目模板
- 支持 `${PRODUCT_NAME}`, `${PRODUCT_DESC}`, `${BACKEND_PORT}`, `${FRONTEND_PORT}` 等占位符替换
- 通过 `replace.txt` 文件定义需要替换的文件列表

#### 依赖关系
- 无外部依赖

#### Dev Agent Record

**实现状态**: ✅ 已完成  
**实现时间**: 2025-01-30  
**实现人员**: Dev Agent  

**核心实现内容**:

1. **数据模型扩展** (`backend/internal/models/project.go`)
   - 添加 `BackendPort int` 和 `FrontendPort int` 字段
   - 设置默认值：后端端口 8080，前端端口 3000
   - 添加端口范围验证约束（1024-65535）

2. **API 接口扩展** (`backend/internal/models/common.go`)
   - `CreateProjectRequest`: 添加 `BackendPort` 和 `FrontendPort` 字段
   - `UpdateProjectRequest`: 支持端口字段更新
   - `ProjectInfo`: 返回端口信息
   - 添加端口范围验证（1024-65535）

3. **业务逻辑实现** (`backend/internal/services/project_service.go`)
   - 修改 `CreateProject` 方法，支持端口参数
   - 集成 `ProjectTemplateService` 进行模板初始化
   - 更新 `UpdateProject` 方法支持端口修改

4. **模板服务实现** (`backend/internal/services/project_template_service.go`)
   - `InitializeProject`: 协调模板提取和占位符替换
   - `ExtractTemplate`: 从 `template.zip` 解压到项目目录
   - `ReplacePlaceholders`: 根据 `replace.txt` 替换文件中的占位符
   - 支持的占位符：`${PRODUCT_NAME}`, `${PRODUCT_DESC}`, `${BACKEND_PORT}`, `${FRONTEND_PORT}`, `${PROJECT_ID}`, `${USER_ID}`

5. **数据库设计更新** (`backend/scripts/init-db.sql`)
   - 在 projects 表中添加 `backend_port` 和 `frontend_port` 字段
   - 设置默认值：后端端口 8080，前端端口 3000
   - 添加端口范围检查约束（1024-65535）
   - 在开发阶段直接集成到初始化脚本中，无需迁移

6. **路由配置更新** (`backend/internal/api/routes/routes.go`)
   - 注入 `ProjectTemplateService` 到 `ProjectService`
   - 配置模板路径为 `"./data/template.zip"`

**技术细节**:

- **模板提取**: 使用 `archive/zip` 包从 `template.zip` 解压文件
- **占位符替换**: 使用 `strings.ReplaceAll` 进行文本替换
- **文件处理**: 支持 `.txt`, `.md`, `.yml`, `.yaml`, `.json`, `.env`, `.toml`, `.ini`, `.cfg`, `.conf`, `.sh`, `.bat`, `.ps1` 等文件类型
- **错误处理**: 完善的错误处理和日志记录
- **并发安全**: 使用 `sync.Mutex` 确保并发安全

**API 使用示例**:

```bash
# 创建项目（使用默认端口）
curl -X POST http://localhost:8098/api/v1/projects \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "name": "我的项目",
    "description": "项目描述",
    "requirements": "项目需求描述"
  }'

# 创建项目（指定端口）
curl -X POST http://localhost:8098/api/v1/projects \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "name": "我的项目",
    "description": "项目描述",
    "requirements": "项目需求描述",
    "backend_port": 8081,
    "frontend_port": 3001
  }'
```

**模板文件结构**:
```
template.zip
├── frontend/
├── backend/
├── docs/
├── docker-compose.yml
├── .env.example
└── replace.txt  # 定义需要替换占位符的文件列表
```

**注意事项**:
- 模板源文件为 `backend/data/template.zip`，不应手动创建或修改
- 项目创建后会自动解压模板并替换占位符
- 端口范围限制为 1024-65535，符合标准端口规范
- 支持现有项目的端口字段更新
- 数据库字段直接在 `init-db.sql` 中定义，适合开发阶段使用

---

### Story 1.2: BMad-Method 安装和配置
**优先级**: P0  
**估算工时**: 2 天  
**负责人**: 后端开发工程师

#### 用户故事
作为系统管理员，我希望系统能够自动安装和配置 BMad-Method，以便启用多 Agent 协作功能。

#### 验收标准
- [ ] 自动检测项目目录中的 BMad-Method 安装状态
- [ ] 支持通过 npm install 安装 BMad-Method
- [ ] 自动生成 BMad-Method 配置文件
- [ ] 验证安装成功并返回配置状态

#### 技术要点
- 集成 Go 的 `exec` 包执行 npm 命令
- 实现配置文件模板化生成
- 支持安装状态检查和错误处理

#### 依赖关系
- 依赖 Story 1.1 (项目初始化)

---

### Story 1.3: cursor.cli 集成
**优先级**: P0  
**估算工时**: 3 天  
**负责人**: 后端开发工程师

#### 用户故事
作为开发人员，我希望系统能够集成 cursor.cli，以便通过编程大语言模型进行代码生成和交互。

#### 验收标准
- [ ] 检测系统中 cursor.cli 的安装状态
- [ ] 支持通过 API 调用 cursor.cli 命令
- [ ] 实现与 cursor.cli 的通信接口
- [ ] 支持命令执行结果的解析和返回
- [ ] 实现错误处理和重试机制

#### 技术要点
- 集成 Go 的 `exec` 包执行 cursor.cli 命令
- 实现命令参数的安全传递
- 支持异步执行和结果回调

#### 依赖关系
- 依赖 Story 1.2 (BMad-Method 安装和配置)

---

### Story 1.4: claude code 集成
**优先级**: P0  
**估算工时**: 3 天  
**负责人**: 后端开发工程师

#### 用户故事
作为开发人员，我希望系统能够集成 claude code，以便通过 Claude 模型进行代码生成和交互。

#### 验收标准
- [ ] 配置 Claude API 密钥和参数
- [ ] 实现与 Claude API 的通信接口
- [ ] 支持代码生成请求的发送和响应处理
- [ ] 实现结果格式化和错误处理
- [ ] 支持多种编程语言和框架

#### 技术要点
- 集成 Go 的 HTTP 客户端调用 Claude API
- 实现 API 密钥的安全管理
- 支持请求重试和限流

#### 依赖关系
- 依赖 Story 1.3 (cursor.cli 集成)

---

### Story 1.5: 提示词模板管理
**优先级**: P0  
**估算工时**: 4 天  
**负责人**: 后端开发工程师

#### 用户故事
作为产品经理，我希望系统能够管理预制好的提示词模板，以便为不同 Agent 角色提供标准化的交互指导。

#### 验收标准
- [ ] 支持模板参数化（如项目名称、类型等）

#### 技术要点
- 设计模板数据结构（支持 Markdown 和变量替换）
- 实现模板引擎进行参数替换
- 支持模板的数据库存储和缓存

#### 依赖关系
- 依赖 Story 1.4 (claude code 集成)

---

### Story 1.6: PRD 文档生成
**优先级**: P0  
**估算工时**: 3 天  
**负责人**: 后端开发工程师

#### 用户故事
作为产品经理，我希望系统能够基于提示词模板自动生成 PRD 文档，以便快速完成产品需求分析。

#### 验收标准
- [ ] 调用 PM Agent 角色生成 PRD 文档
- [ ] 实现文档的自动保存和版本管理
- [ ] 支持文档的预览和编辑
- [ ] 生成结构化的 PRD 内容

#### 技术要点
- 集成 BMad-Method 的 PM Agent 角色
- 实现文档生成的工作流
- 支持 Markdown 输出格式

#### 依赖关系
- 依赖 Story 1.5 (提示词模板管理)

---

### Story 1.7: 架构设计文档生成
**优先级**: P0  
**估算工时**: 3 天  
**负责人**: 后端开发工程师

#### 用户故事
作为架构师，我希望系统能够基于提示词模板自动生成架构设计文档，以便快速完成技术架构设计。

#### 验收标准
- [ ] 调用 Architect Agent 角色生成架构设计
- [ ] 生成技术选型和技术栈建议
- [ ] 支持架构图的生成和展示
- [ ] 实现文档的版本管理和对比

#### 技术要点
- 集成 BMad-Method 的 Architect Agent 角色
- 支持架构图的自动生成（如 Mermaid 图表）
- 实现技术选型的智能推荐

#### 依赖关系
- 依赖 Story 1.6 (PRD 文档生成)

---

### Story 1.8: UX 设计文档生成
**优先级**: P0  
**估算工时**: 3 天  
**负责人**: 后端开发工程师

#### 用户故事
作为 UX 设计师，我希望系统能够基于提示词模板自动生成 UX 设计文档，以便快速完成用户体验设计。

#### 验收标准
- [ ] 调用 UX Expert Agent 角色生成 UX 设计
- [ ] 支持多种设计风格和组件库
- [ ] 生成用户流程和交互设计
- [ ] 支持设计规范和组件库推荐

#### 技术要点
- 集成 BMad-Method 的 UX Expert Agent 角色
- 支持设计规范的自动生成
- 实现组件库的智能推荐

#### 依赖关系
- 依赖 Story 1.7 (架构设计文档生成)

---

### Story 1.9: Epic 和 Story 文档生成
**优先级**: P0  
**估算工时**: 3 天  
**负责人**: 后端开发工程师

#### 用户故事
作为产品负责人，我希望系统能够基于提示词模板自动生成 Epic 和 Story 文档，以便快速完成开发任务分解。

#### 验收标准
- [ ] 调用 PO Agent 角色生成 Epic 和 Story
- [ ] 生成任务优先级和工时估算
- [ ] 支持依赖关系的自动识别
- [ ] 实现文档的导出

#### 技术要点
- 集成 BMad-Method 的 PO Agent 角色
- 实现任务依赖关系的自动分析

#### 依赖关系
- 依赖 Story 1.8 (UX 设计文档生成)

---

### Story 1.10: 项目状态管理
**优先级**: P0  
**估算工时**: 2 天  
**负责人**: 后端开发工程师

#### 用户故事
作为项目管理员，我希望能够跟踪项目的创建和配置状态，以便了解项目的准备情况。

#### 验收标准
- [ ] 实时显示项目创建进度
- [ ] 支持项目状态的查询和更新
- [ ] 实现状态变更的历史记录
- [ ] 支持状态异常的告警和通知
- [ ] 提供项目配置的完整性检查

#### 技术要点
- 设计项目状态机（创建中、配置中、就绪、异常等）
- 实现状态变更的事件通知
- 支持状态数据的持久化存储

#### 依赖关系
- 依赖 Story 1.9 (Epic 和 Story 文档生成)

---

## 技术架构设计

### 核心组件
1. **项目管理器**：负责项目目录创建和配置
2. **BMad-Method 集成器**：管理 BMad-Method 的安装和配置
3. **工具集成器**：集成 cursor.cli 和 claude code
4. **模板引擎**：管理提示词模板和参数替换
5. **文档生成器**：协调各 Agent 角色生成文档
6. **状态管理器**：跟踪项目创建和配置状态

### 数据模型
- **Project**: 项目基本信息
- **ProjectConfig**: 项目配置信息
- **PromptTemplate**: 提示词模板
- **GeneratedDocument**: 生成的文档
- **ProjectStatus**: 项目状态信息

### API 接口
- `POST /api/projects` - 创建新项目
- `GET /api/projects/{id}` - 获取项目信息
- `POST /api/projects/{id}/install-bmad` - 安装 BMad-Method
- `POST /api/projects/{id}/generate-docs` - 生成文档
- `GET /api/projects/{id}/status` - 获取项目状态

## 总结

Epic 1 包含 10 个核心用户故事，涵盖了项目创建和 BMad-Method 集成的完整流程。这些故事按照依赖关系顺序排列，确保系统能够逐步构建完整的项目创建能力。

建议按照优先级和依赖关系进行开发，优先完成核心的基础功能，然后逐步实现文档生成和状态管理功能。
