# WebSocket 服务设计文档

## 1. 概述

本文档描述了 AutoCode WebSocket 服务的架构设计，用于实现实时项目状态更新和消息推送功能。

## 2. 设计目标

- 实时推送项目开发阶段更新
- 实时推送项目对话消息
- 支持与 agents-server 的 WebSocket 通信
- 支持用户反馈和交互
- 保持与现有服务架构的兼容性

## 3. 架构设计

### 3.1 服务层次结构

```
WebSocketService (接口)
├── WebSocketManager (连接管理)
├── MessageRouter (消息路由)
├── ProjectEventBroadcaster (项目事件广播)
└── AgentCommunicationHandler (Agent通信处理)
```

### 3.2 核心组件

#### 3.2.1 WebSocketService
- 管理 WebSocket 连接生命周期
- 处理客户端连接/断开
- 消息序列化/反序列化
- 错误处理和重连机制

#### 3.2.2 WebSocketManager
- 维护活跃连接映射
- 按项目分组管理连接
- 连接状态监控
- 心跳检测

#### 3.2.3 MessageRouter
- 消息类型路由
- 权限验证
- 消息格式验证
- 错误响应处理

#### 3.2.4 ProjectEventBroadcaster
- 项目阶段更新广播
- 项目消息广播
- 项目状态变更通知
- 事件过滤和分发

#### 3.2.5 AgentCommunicationHandler
- 与 agents-server 的 WebSocket 通信
- Agent 消息转发
- 用户反馈处理
- Agent 状态同步

## 4. 消息协议设计

### 4.1 消息格式

```json
{
  "type": "message_type",
  "projectGuid": "project_guid",
  "data": {},
  "timestamp": "2025-01-01T00:00:00Z",
  "id": "message_id"
}
```

### 4.2 消息类型

#### 4.2.1 客户端 -> 服务端
- `join_project`: 加入项目房间
- `leave_project`: 离开项目房间
- `user_feedback`: 用户反馈
- `ping`: 心跳检测

#### 4.2.2 服务端 -> 客户端
- `project_stage_update`: 项目阶段更新
- `project_message`: 新项目消息
- `project_status_change`: 项目状态变更
- `agent_message`: Agent 消息
- `user_feedback_response`: 用户反馈响应
- `pong`: 心跳响应
- `error`: 错误消息

#### 4.2.3 服务端 -> agents-server
- `project_update`: 项目更新通知
- `user_feedback`: 用户反馈转发
- `project_status_sync`: 项目状态同步

## 5. 数据流设计

### 5.1 项目阶段更新流程

```
ProjectStageService -> ProjectEventBroadcaster -> WebSocketManager -> Client
```

### 5.2 项目消息流程

```
MessageService -> ProjectEventBroadcaster -> WebSocketManager -> Client
```

### 5.3 Agent 通信流程

```
Agent -> AgentCommunicationHandler -> WebSocketManager -> Client
Client -> WebSocketManager -> AgentCommunicationHandler -> Agent
```

## 6. 安全设计

### 6.1 连接认证
- JWT Token 验证
- 项目访问权限检查
- 用户身份验证

### 6.2 消息安全
- 消息格式验证
- 权限检查
- 频率限制
- 恶意消息过滤

## 7. 性能考虑

### 7.1 连接管理
- 连接池管理
- 内存使用优化
- 连接超时处理

### 7.2 消息处理
- 异步消息处理
- 消息队列缓冲
- 批量消息发送

### 7.3 扩展性
- 水平扩展支持
- 负载均衡
- 集群部署

## 8. 监控和日志

### 8.1 监控指标
- 活跃连接数
- 消息发送速率
- 错误率
- 响应时间

### 8.2 日志记录
- 连接事件日志
- 消息处理日志
- 错误日志
- 性能日志

## 9. 部署和配置

### 9.1 配置参数
- WebSocket 端口
- 心跳间隔
- 连接超时
- 消息缓冲区大小

### 9.2 环境要求
- Redis (用于集群通信)
- 负载均衡器支持
- 反向代理配置

## 10. 实现计划

### 阶段 1: 核心服务实现
- WebSocketService 接口定义
- WebSocketManager 基础功能
- MessageRouter 消息路由

### 阶段 2: 项目集成
- ProjectEventBroadcaster 实现
- 与现有服务集成
- 项目阶段更新推送

### 阶段 3: Agent 通信
- AgentCommunicationHandler 实现
- 与 agents-server 通信
- 用户反馈处理

### 阶段 4: 前端集成
- 前端 WebSocket 客户端
- 消息处理逻辑
- 错误处理和重连

### 阶段 5: 优化和测试
- 性能优化
- 压力测试
- 监控和日志完善
