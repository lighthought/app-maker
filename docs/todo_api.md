# 后端API接口待实现清单

## 概述

本文档列出了前端当前界面实际需要但后端尚未实现的API接口。前端代码已经按照这些接口规范编写，假设后端接口已经实现。

## 1. 项目相关接口

### 1.1 获取项目详情
- **接口**: `GET /api/v1/projects/{projectId}`
- **描述**: 获取指定项目的详细信息
- **请求参数**: 
  - `projectId` (path): 项目ID
- **响应数据**:
```json
{
  "id": "string",
  "name": "string",
  "description": "string",
  "status": "draft|in_progress|completed|failed",
  "requirements": "string",
  "projectPath": "string",
  "backendPort": 8080,
  "frontendPort": 3000,
  "previewUrl": "string",
  "userId": "string",
  "user": {
    "id": "string",
    "email": "string",
    "username": "string"
  },
  "tags": [
    {
      "id": "string",
      "name": "string",
      "color": "string"
    }
  ],
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```


## 2. 文件相关接口

### 2.1 获取项目文件列表
- **接口**: `GET /api/v1/projects/{projectId}/files`
- **描述**: 获取项目的文件树结构
- **请求参数**:
  - `projectId` (path): 项目ID
  - `path` (query): 可选，指定目录路径
- **响应数据**:
```json
[
  {
    "name": "string",
    "path": "string",
    "type": "file|folder",
    "size": 1024,
    "modifiedAt": "2024-01-01T00:00:00Z"
  }
]
```

### 2.2 获取文件内容
- **接口**: `GET /api/v1/projects/{projectId}/files/content`
- **描述**: 获取指定文件的内容
- **请求参数**:
  - `projectId` (path): 项目ID
  - `filePath` (query): 文件路径
- **响应数据**:
```json
{
  "path": "string",
  "content": "string",
  "size": 1024,
  "modifiedAt": "2024-01-01T00:00:00Z"
}
```


## 3. 对话消息相关接口

### 3.1 获取项目对话历史
- **接口**: `GET /api/v1/projects/{projectId}/conversations`
- **描述**: 获取项目的对话历史记录
- **请求参数**:
  - `projectId` (path): 项目ID
  - `page` (query): 页码，默认1
  - `pageSize` (query): 每页数量，默认50
- **响应数据**:
```json
{
  "total": 100,
  "page": 1,
  "pageSize": 50,
  "totalPages": 2,
  "data": [
    {
      "id": "string",
      "type": "user|agent|system",
      "agentRole": "dev|pm|arch|ux|qa|ops",
      "agentName": "string",
      "content": "string",
      "timestamp": "2024-01-01T00:00:00Z",
      "isMarkdown": false,
      "markdownContent": "string",
      "isExpanded": false
    }
  ],
  "hasNext": true,
  "hasPrevious": false
}
```

### 3.2 添加对话消息
- **接口**: `POST /api/v1/projects/{projectId}/conversations`
- **描述**: 添加新的对话消息（系统消息）
- **请求参数**:
  - `projectId` (path): 项目ID
- **请求体**:
```json
{
  "type": "user|system|agent",
  "content": "string",
  "isMarkdown": false,
}
```

## 4. 开发阶段相关接口

### 4.1 获取项目开发阶段
- **接口**: `GET /api/v1/projects/{projectId}/stages`
- **描述**: 获取项目的开发阶段信息
- **请求参数**:
  - `projectId` (path): 项目ID
- **响应数据**:
```json
[
  {
    "id": "string",
    "name": "string",
    "status": "pending|in_progress|completed|failed",
    "progress": 25,
    "description": "string",
    "startedAt": "2024-01-01T00:00:00Z",
    "completedAt": "2024-01-01T00:00:00Z"
  }
]
```

## 5. WebSocket接口（未来扩展）

### 5.1 对话消息实时推送
- **接口**: `WS /ws/projects/{projectId}/conversations`
- **描述**: WebSocket连接，实时推送对话消息和接收用户反馈
- **连接参数**:
  - `projectId` (path): 项目ID
  - `token` (query): JWT认证token
- **后端推送消息格式**:
```json
{
  "type": "agent_message|system_message|stage_update",
  "data": {
    "messageId": "string",
    "type": "agent|system",
    "agentRole": "pm|dev|designer|qa",
    "agentName": "string",
    "content": "string",
    "isMarkdown": false,
    "markdownContent": "string",
    "timestamp": "2024-01-01T00:00:00Z"
  }
}
```
- **前端发送消息格式**:
```json
{
  "type": "user_feedback|user_message",
  "data": {
    "content": "string",
    "stageId": "string",
    "feedbackType": "approve|reject|modify"
  }
}
```

### 5.2 开发阶段状态推送
- **接口**: `WS /ws/projects/{projectId}/stages`
- **描述**: WebSocket连接，实时推送开发阶段状态更新
- **连接参数**:
  - `projectId` (path): 项目ID
  - `token` (query): JWT认证token
- **推送消息格式**:
```json
{
  "type": "stage_update",
  "data": {
    "stageId": "string",
    "stageName": "string",
    "status": "pending|in_progress|completed|failed",
    "progress": 25,
    "description": "string",
    "timestamp": "2024-01-01T00:00:00Z"
  }
}
```

## 6. 错误处理

所有API接口都应该遵循统一的错误响应格式：

```json
{
  "error": {
    "code": "string",
    "message": "string",
    "details": "string"
  }
}
```

常见错误码：
- `400`: 请求参数错误
- `401`: 未授权
- `403`: 禁止访问
- `404`: 资源不存在
- `500`: 服务器内部错误

## 7. 认证

所有API接口都需要JWT认证，除了WebSocket连接外，其他接口都需要在请求头中包含：

```
Authorization: Bearer <jwt_token>
```

## 8. 实现优先级

### 高优先级（当前界面必需）
1. 项目详情接口
2. 文件相关接口（文件列表、文件内容、预览配置）
3. 对话消息接口（获取历史、添加消息）
4. 开发阶段接口

### 中优先级（增强功能）
1. WebSocket实时推送
2. 项目预览URL更新

## 9. 注意事项

1. **分页**: 对话历史接口支持分页
2. **权限**: 确保用户只能访问自己的项目
3. **安全**: 文件路径需要防止目录遍历攻击
4. **性能**: 大文件内容可能需要流式传输
5. **实时性**: WebSocket连接需要处理断线重连