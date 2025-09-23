// WebSocket 消息类型
export interface WebSocketMessage {
  type: string
  projectGuid: string
  data: any
  timestamp: string
  id: string
}

// WebSocket 客户端消息类型（发送给服务端）
export interface WebSocketClientMessage {
  type: 'ping' | 'join_project' | 'leave_project' | 'user_feedback'
  projectGuid?: string
  data?: any
  timestamp?: string
  id?: string
}

// WebSocket 服务端消息类型（接收自服务端）
export interface WebSocketServerMessage {
  type: 'project_stage_update' | 'project_message' | 'project_status_change' | 'agent_message' | 'user_feedback_response' | 'pong' | 'error' | 'project_joined' | 'project_left'
  projectGuid: string
  data: any
  timestamp: string
  id: string
}

// WebSocket 连接状态
export type WebSocketStatus = 'connecting' | 'connected' | 'disconnected' | 'error' | 'reconnecting'

// WebSocket 连接配置
export interface WebSocketConfig {
  url: string
  protocols?: string[]
  reconnectInterval?: number
  maxReconnectAttempts?: number
  heartbeatInterval?: number
  timeout?: number
}

// WebSocket 事件回调
export interface WebSocketEventHandlers {
  onOpen?: () => void
  onClose?: (event: CloseEvent) => void
  onError?: (error: Event) => void
  onMessage?: (message: WebSocketServerMessage) => void
  onReconnect?: (attempt: number) => void
  onReconnectFailed?: () => void
}

// WebSocket 连接实例
export interface WebSocketConnection {
  connect: () => Promise<void>
  disconnect: () => void
  send: (message: WebSocketClientMessage) => void
  getStatus: () => WebSocketStatus
  isConnected: () => boolean
  reconnect: () => void
}

// 项目阶段更新消息数据
export interface ProjectStageUpdateData {
  id: string
  name: string
  status: 'pending' | 'in_progress' | 'done' | 'failed'
  progress: number
  description: string
  failed_reason: string
  task_id: string
}

// 项目消息更新数据
export interface ProjectMessageUpdateData {
  id: string
  project_id: string
  type: 'user' | 'agent' | 'system'
  agent_role?: string
  agent_name?: string
  content: string
  is_markdown?: boolean
  markdown_content?: string
  is_expanded?: boolean
  created_at: string
  updated_at: string
}

// 项目状态变更数据
export interface ProjectStatusChangeData {
  status: string
}

// 用户反馈响应数据
export interface UserFeedbackResponseData {
  message: string
}

// 错误消息数据
export interface ErrorMessageData {
  message: string
  details: string
}
