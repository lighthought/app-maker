# AutoCodeWeb API 设计文档

## 1. API 设计概述

### 1.1 设计理念
- **RESTful 设计**：遵循 REST 架构原则，使用标准 HTTP 方法
- **版本控制**：API 版本化管理，支持向后兼容
- **统一响应格式**：标准化的响应结构和错误处理
- **认证授权**：JWT Token 认证，基于角色的权限控制
- **限流保护**：API 访问频率限制，防止恶意请求

### 1.2 技术规范
- **协议**：HTTP/HTTPS
- **数据格式**：JSON
- **字符编码**：UTF-8
- **认证方式**：Bearer Token (JWT)
- **状态码**：标准 HTTP 状态码
- **时间格式**：ISO 8601 (RFC 3339)

### 1.3 基础URL
```
开发环境: http://localhost:8080/api/v1
生产环境: https://api.autocodeweb.com/api/v1
```

## 2. 通用响应格式

### 2.1 成功响应
```json
{
  "success": true,
  "data": {
    // 具体数据内容
  },
  "message": "操作成功",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### 2.2 错误响应
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "请求参数验证失败",
    "details": [
      {
        "field": "email",
        "message": "邮箱格式不正确"
      }
    ]
  },
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### 2.3 分页响应
```json
{
  "success": true,
  "data": {
    "items": [
      // 数据项列表
    ],
    "pagination": {
      "page": 1,
      "page_size": 20,
      "total": 100,
      "total_pages": 5,
      "has_next": true,
      "has_prev": false
    }
  },
  "message": "查询成功"
}
```

## 3. 认证与授权

### 3.1 认证接口
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123"
}
```

**响应示例：**
```json
{
  "success": true,
  "data": {
    "user": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "email": "user@example.com",
      "name": "张三",
      "role": "user",
      "avatar_url": "https://example.com/avatar.jpg"
    },
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 3600
  },
  "message": "登录成功"
}
```

```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "email": "newuser@example.com",
  "password": "password123",
  "name": "新用户",
  "confirm_password": "password123"
}
```

```http
POST /api/v1/auth/refresh
Authorization: Bearer {refresh_token}
```

```http
POST /api/v1/auth/logout
Authorization: Bearer {access_token}
```

### 3.2 认证中间件
所有需要认证的接口都需要在请求头中包含：
```http
Authorization: Bearer {access_token}
```

## 4. 用户管理 API

### 4.1 用户信息管理
```http
GET /api/v1/users/profile
Authorization: Bearer {access_token}
```

**响应示例：**
```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "name": "张三",
    "avatar_url": "https://example.com/avatar.jpg",
    "role": "user",
    "status": "active",
    "email_verified": true,
    "last_login_at": "2024-01-15T10:30:00Z",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  }
}
```

```http
PUT /api/v1/users/profile
Authorization: Bearer {access_token}
Content-Type: application/json

{
  "name": "新名字",
  "avatar_url": "https://example.com/new-avatar.jpg"
}
```

```http
PUT /api/v1/users/password
Authorization: Bearer {access_token}
Content-Type: application/json

{
  "current_password": "oldpassword",
  "new_password": "newpassword",
  "confirm_password": "newpassword"
}
```

### 4.2 用户权限管理
```http
GET /api/v1/users/permissions
Authorization: Bearer {access_token}
```

```http
POST /api/v1/users/permissions
Authorization: Bearer {access_token}
Content-Type: application/json

{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "permission": "project:create",
  "expires_at": "2024-12-31T23:59:59Z"
}
```

## 5. 项目管理 API

### 5.1 项目基础操作
```http
GET /api/v1/projects
Authorization: Bearer {access_token}
Query Parameters:
  - page: 1 (默认)
  - page_size: 20 (默认)
  - status: draft|planning|development|testing|completed|deployed|archived
  - project_type: web|mobile|desktop|api
  - search: 搜索关键词
```

**响应示例：**
```json
{
  "success": true,
  "data": {
    "items": [
      {
        "id": "660e8400-e29b-41d4-a716-446655440001",
        "name": "电商网站",
        "description": "一个现代化的电商网站",
        "status": "development",
        "project_type": "web",
        "user_id": "550e8400-e29b-41d4-a716-446655440000",
        "owner_name": "张三",
        "task_count": 15,
        "completed_tasks": 8,
        "created_at": "2024-01-10T00:00:00Z",
        "updated_at": "2024-01-15T10:30:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "page_size": 20,
      "total": 1,
      "total_pages": 1,
      "has_next": false,
      "has_prev": false
    }
  }
}
```

```http
POST /api/v1/projects
Authorization: Bearer {access_token}
Content-Type: application/json

{
  "name": "新项目",
  "description": "项目描述",
  "project_type": "web",
  "figma_url": "https://figma.com/file/xxx",
  "requirements": "项目需求描述"
}
```

```http
GET /api/v1/projects/{project_id}
Authorization: Bearer {access_token}
```

```http
PUT /api/v1/projects/{project_id}
Authorization: Bearer {access_token}
Content-Type: application/json

{
  "name": "更新后的项目名",
  "description": "更新后的描述",
  "status": "planning"
}
```

```http
DELETE /api/v1/projects/{project_id}
Authorization: Bearer {access_token}
```

### 5.2 项目成员管理
```http
GET /api/v1/projects/{project_id}/members
Authorization: Bearer {access_token}
```

```http
POST /api/v1/projects/{project_id}/members
Authorization: Bearer {access_token}
Content-Type: application/json

{
  "user_id": "660e8400-e29b-41d4-a716-446655440002",
  "role": "developer"
}
```

```http
PUT /api/v1/projects/{project_id}/members/{user_id}
Authorization: Bearer {access_token}
Content-Type: application/json

{
  "role": "admin"
}
```

```http
DELETE /api/v1/projects/{project_id}/members/{user_id}
Authorization: Bearer {access_token}
```

### 5.3 项目标签管理
```http
GET /api/v1/projects/{project_id}/tags
Authorization: Bearer {access_token}
```

```http
POST /api/v1/projects/{project_id}/tags
Authorization: Bearer {access_token}
Content-Type: application/json

{
  "tag": "前端"
}
```

```http
DELETE /api/v1/projects/{project_id}/tags/{tag}
Authorization: Bearer {access_token}
```

## 6. Agent 协作 API

### 6.1 Agent 会话管理
```http
GET /api/v1/projects/{project_id}/agent-sessions
Authorization: Bearer {access_token}
Query Parameters:
  - session_type: pm|ux|architect|po|qa
  - status: active|completed|failed
```

**响应示例：**
```json
{
  "success": true,
  "data": [
    {
      "id": "770e8400-e29b-41d4-a716-446655440003",
      "session_type": "pm",
      "status": "active",
      "started_at": "2024-01-15T10:00:00Z",
      "metadata": {
        "agent_name": "产品经理",
        "current_step": "需求澄清"
      }
    }
  ]
}
```

```http
POST /api/v1/projects/{project_id}/agent-sessions
Authorization: Bearer {access_token}
Content-Type: application/json

{
  "session_type": "pm",
  "requirements": "用户需求描述"
}
```

```http
GET /api/v1/agent-sessions/{session_id}
Authorization: Bearer {access_token}
```

### 6.2 Agent 对话管理
```http
GET /api/v1/agent-sessions/{session_id}/messages
Authorization: Bearer {access_token}
Query Parameters:
  - page: 1
  - page_size: 50
  - message_type: user|agent|system
```

**响应示例：**
```json
{
  "success": true,
  "data": {
    "items": [
      {
        "id": "880e8400-e29b-41d4-a716-446655440004",
        "message_type": "user",
        "content": "我需要一个电商网站",
        "created_at": "2024-01-15T10:00:00Z"
      },
      {
        "id": "990e8400-e29b-41d4-a716-446655440005",
        "message_type": "agent",
        "agent_role": "pm",
        "content": "好的，我来帮您分析需求。请告诉我更多细节...",
        "created_at": "2024-01-15T10:00:05Z"
      }
    ],
    "pagination": {
      "page": 1,
      "page_size": 50,
      "total": 2,
      "total_pages": 1,
      "has_next": false,
      "has_prev": false
    }
  }
}
```

```http
POST /api/v1/agent-sessions/{session_id}/messages
Authorization: Bearer {access_token}
Content-Type: application/json

{
  "content": "用户的新消息",
  "message_type": "user"
}
```

### 6.3 Agent 工作成果管理
```http
GET /api/v1/agent-sessions/{session_id}/artifacts
Authorization: Bearer {access_token}
Query Parameters:
  - artifact_type: prd|ux_spec|architecture|epics|stories|test_plan
  - status: draft|review|approved|rejected
```

```http
GET /api/v1/agent-sessions/{session_id}/artifacts/{artifact_id}
Authorization: Bearer {access_token}
```

```http
PUT /api/v1/agent-sessions/{session_id}/artifacts/{artifact_id}
Authorization: Bearer {access_token}
Content-Type: application/json

{
  "content": {
    "title": "产品需求文档",
    "sections": [
      {
        "title": "功能需求",
        "content": "详细的功能描述"
      }
    ]
  },
  "status": "review"
}
```

## 7. 文档管理 API

### 7.1 文档基础操作
```http
GET /api/v1/projects/{project_id}/documents
Authorization: Bearer {access_token}
Query Parameters:
  - document_type: prd|ux_spec|architecture|api_spec|database_schema|test_plan
  - status: draft|review|approved|published
  - format: markdown|json|yaml|html
  - page: 1
  - page_size: 20
```

```http
POST /api/v1/projects/{project_id}/documents
Authorization: Bearer {access_token}
Content-Type: application/json

{
  "title": "产品需求文档",
  "content": "# 产品需求文档\n\n## 1. 概述\n...",
  "document_type": "prd",
  "format": "markdown"
}
```

```http
GET /api/v1/documents/{document_id}
Authorization: Bearer {access_token}
```

```http
PUT /api/v1/documents/{document_id}
Authorization: Bearer {access_token}
Content-Type: application/json

{
  "title": "更新后的标题",
  "content": "更新后的内容",
  "status": "review"
}
```

```http
DELETE /api/v1/documents/{document_id}
Authorization: Bearer {access_token}
```

### 7.2 文档版本管理
```http
GET /api/v1/documents/{document_id}/versions
Authorization: Bearer {access_token}
```

```http
GET /api/v1/documents/{document_id}/versions/{version}
Authorization: Bearer {access_token}
```

```http
POST /api/v1/documents/{document_id}/versions
Authorization: Bearer {access_token}
Content-Type: application/json

{
  "content": "新版本的内容",
  "change_log": "版本变更说明"
}
```

### 7.3 文档评论管理
```http
GET /api/v1/documents/{document_id}/comments
Authorization: Bearer {access_token}
Query Parameters:
  - page: 1
  - page_size: 20
```

```http
POST /api/v1/documents/{document_id}/comments
Authorization: Bearer {access_token}
Content-Type: application/json

{
  "content": "评论内容",
  "parent_id": "parent_comment_id" // 可选，用于回复
}
```

```http
PUT /api/v1/comments/{comment_id}
Authorization: Bearer {access_token}
Content-Type: application/json

{
  "content": "更新后的评论内容"
}
```

```http
DELETE /api/v1/comments/{comment_id}
Authorization: Bearer {access_token}
```

## 9. 部署和预览 API

### 9.1 部署环境管理
```http
GET /api/v1/projects/{project_id}/environments
Authorization: Bearer {access_token}
```

```http
POST /api/v1/projects/{project_id}/environments
Authorization: Bearer {access_token}
Content-Type: application/json

{
  "name": "开发环境",
  "environment_type": "development",
  "url": "https://dev.example.com",
  "settings": {
    "auto_deploy": true,
    "branch": "develop"
  }
}
```

```http
GET /api/v1/environments/{environment_id}
Authorization: Bearer {access_token}
```

```http
PUT /api/v1/environments/{environment_id}
Authorization: Bearer {access_token}
Content-Type: application/json

{
  "status": "maintenance",
  "settings": {
    "auto_deploy": false
  }
}
```

### 9.2 部署管理
```http
GET /api/v1/projects/{project_id}/deployments
Authorization: Bearer {access_token}
Query Parameters:
  - environment_id: environment_id
  - status: pending|building|deploying|success|failed|rolled_back
  - page: 1
  - page_size: 20
```

```http
POST /api/v1/projects/{project_id}/deployments
Authorization: Bearer {access_token}
Content-Type: application/json

{
  "environment_id": "environment_id",
  "version": "v1.0.0"
}
```

```http
GET /api/v1/deployments/{deployment_id}
Authorization: Bearer {access_token}
```

```http
POST /api/v1/deployments/{deployment_id}/rollback
Authorization: Bearer {access_token}
```

### 9.3 预览配置管理
```http
GET /api/v1/projects/{project_id}/preview-configs
Authorization: Bearer {access_token}
```

```http
POST /api/v1/projects/{project_id}/preview-configs
Authorization: Bearer {access_token}
Content-Type: application/json

{
  "name": "桌面端预览",
  "config_type": "desktop",
  "settings": {
    "viewport": "1920x1080",
    "theme": "light",
    "device": "desktop"
  }
}
```

```http
GET /api/v1/preview-configs/{config_id}
Authorization: Bearer {access_token}
```

```http
PUT /api/v1/preview-configs/{config_id}
Authorization: Bearer {access_token}
Content-Type: application/json

{
  "settings": {
    "viewport": "1366x768",
    "theme": "dark"
  }
}
```

## 10. 系统管理 API

### 10.1 系统状态
```http
GET /api/v1/system/status
Authorization: Bearer {access_token}
```

**响应示例：**
```json
{
  "success": true,
  "data": {
    "status": "healthy",
    "version": "1.0.0",
    "uptime": "72h30m15s",
    "database": {
      "status": "connected",
      "connections": 15,
      "max_connections": 100
    },
    "redis": {
      "status": "connected",
      "memory_usage": "256MB",
      "keys": 1250
    },
    "agents": {
      "total": 5,
      "active": 3,
      "idle": 2
    }
  }
}
```

### 10.2 系统配置
```http
GET /api/v1/system/config
Authorization: Bearer {access_token}
```

```http
PUT /api/v1/system/config
Authorization: Bearer {access_token}
Content-Type: application/json

{
  "max_projects_per_user": 50,
  "max_file_size": "10MB",
  "auto_cleanup_days": 30
}
```

## 11. 错误码定义

### 11.1 通用错误码
| 错误码 | HTTP状态码 | 说明 |
|--------|------------|------|
| `VALIDATION_ERROR` | 400 | 请求参数验证失败 |
| `UNAUTHORIZED` | 401 | 未认证或认证失败 |
| `FORBIDDEN` | 403 | 权限不足 |
| `NOT_FOUND` | 404 | 资源不存在 |
| `CONFLICT` | 409 | 资源冲突 |
| `RATE_LIMIT_EXCEEDED` | 429 | 请求频率超限 |
| `INTERNAL_ERROR` | 500 | 服务器内部错误 |

### 11.2 业务错误码
| 错误码 | HTTP状态码 | 说明 |
|--------|------------|------|
| `PROJECT_NOT_FOUND` | 404 | 项目不存在 |
| `PROJECT_ACCESS_DENIED` | 403 | 项目访问权限不足 |
| `AGENT_SESSION_EXPIRED` | 410 | Agent会话已过期 |
| `TASK_DEPENDENCY_CYCLE` | 400 | 任务依赖关系存在循环 |
| `DEPLOYMENT_IN_PROGRESS` | 409 | 部署正在进行中 |
| `INSUFFICIENT_QUOTA` | 429 | 配额不足 |

## 12. 限流策略

### 12.1 限流规则
```json
{
  "anonymous": {
    "rate_limit": "10/minute",
    "burst_limit": 20
  },
  "authenticated": {
    "rate_limit": "100/minute",
    "burst_limit": 200
  },
  "admin": {
    "rate_limit": "1000/minute",
    "burst_limit": 2000
  }
}
```

### 12.2 限流响应头
```http
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1642233600
Retry-After: 60
```

---

*本文档为 AutoCodeWeb 项目的 API 设计，由架构师 Winston 创建*
