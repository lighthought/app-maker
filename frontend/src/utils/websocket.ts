import { useUserStore } from '@/stores/user'
import { AppConfig } from './config'
import type {
  WebSocketConfig,
  WebSocketEventHandlers,
  WebSocketConnection,
  WebSocketStatus,
  WebSocketClientMessage,
  WebSocketServerMessage
} from '@/types/websocket'

export class WebSocketService implements WebSocketConnection {
  private ws: WebSocket | null = null
  private config: WebSocketConfig
  private handlers: WebSocketEventHandlers
  private status: WebSocketStatus = 'disconnected'
  private reconnectAttempts = 0
  private reconnectTimer: number | null = null
  private heartbeatTimer: number | null = null
  private isManualDisconnect = false

  constructor(config: WebSocketConfig, handlers: WebSocketEventHandlers = {}) {
    this.config = {
      reconnectInterval: 5000,
      maxReconnectAttempts: 5,
      heartbeatInterval: 30000,
      timeout: 10000,
      ...config
    }
    this.handlers = handlers
  }

  public async connect(): Promise<void> {
    if (this.ws?.readyState === WebSocket.OPEN) {
      console.log('WebSocket already connected')
      return
    }

    this.isManualDisconnect = false
    this.setStatus('connecting')

    try {
      const userStore = useUserStore()
      if (!userStore.token) {
        throw new Error('No authentication token available')
      }

      // 构建 WebSocket URL
      const wsUrl = this.buildWebSocketUrl()
      
      console.log(`Connecting to WebSocket: ${wsUrl}`)
      
      this.ws = new WebSocket(wsUrl, this.config.protocols)
      
      // 设置连接超时
      const timeout = setTimeout(() => {
        if (this.ws?.readyState === WebSocket.CONNECTING) {
          this.ws.close()
          this.setStatus('error')
          this.handlers.onError?.(new Event('Connection timeout'))
        }
      }, this.config.timeout)

      this.ws.onopen = () => {
        clearTimeout(timeout)
        this.setStatus('connected')
        this.reconnectAttempts = 0
        this.startHeartbeat()
        console.log('WebSocket connected successfully')
        this.handlers.onOpen?.()
      }

      this.ws.onclose = (event) => {
        clearTimeout(timeout)
        this.stopHeartbeat()
        this.setStatus('disconnected')
        console.log(`WebSocket closed: ${event.code} - ${event.reason}`)
        this.handlers.onClose?.(event)
        
        // 如果不是手动断开，尝试重连
        if (!this.isManualDisconnect && this.reconnectAttempts < this.config.maxReconnectAttempts!) {
          this.scheduleReconnect()
        }
      }

      this.ws.onerror = (error) => {
        clearTimeout(timeout)
        this.setStatus('error')
        console.error('WebSocket error:', error)
        this.handlers.onError?.(error)
      }

      this.ws.onmessage = (event) => {
        try {
          const message: WebSocketServerMessage = JSON.parse(event.data)
          console.log('WebSocket message received:', message)
          this.handlers.onMessage?.(message)
        } catch (error) {
          console.error('Failed to parse WebSocket message:', error)
        }
      }

    } catch (error) {
      this.setStatus('error')
      console.error('WebSocket connection failed:', error)
      throw error
    }
  }

  public disconnect(): void {
    this.isManualDisconnect = true
    this.stopHeartbeat()
    this.clearReconnectTimer()
    
    if (this.ws) {
      this.ws.close(1000, 'Manual disconnect')
      this.ws = null
    }
    
    this.setStatus('disconnected')
    console.log('WebSocket disconnected manually')
  }

  public send(message: WebSocketClientMessage): void {
    if (!this.isConnected()) {
      console.error('Cannot send message: WebSocket not connected')
      return
    }

    try {
      const messageWithTimestamp = {
        ...message,
        timestamp: new Date().toISOString(),
        id: this.generateMessageId()
      }
      
      this.ws!.send(JSON.stringify(messageWithTimestamp))
      console.log('WebSocket message sent:', messageWithTimestamp)
    } catch (error) {
      console.error('Failed to send WebSocket message:', error)
    }
  }

  public getStatus(): WebSocketStatus {
    return this.status
  }

  public isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN
  }

  public reconnect(): void {
    this.disconnect()
    this.reconnectAttempts = 0
    this.connect()
  }

  private setStatus(status: WebSocketStatus): void {
    this.status = status
  }

  private buildWebSocketUrl(): string {
    const config = AppConfig.getInstance().getConfig()
    const baseUrl = config.apiBaseUrl.replace('/api/v1', '')
    const wsProtocol = baseUrl.startsWith('https') ? 'wss' : 'ws'
    const wsBaseUrl = baseUrl.replace(/^https?/, wsProtocol)
    
    return `${wsBaseUrl}${this.config.url}`
  }

  private generateMessageId(): string {
    return `msg_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
  }

  private startHeartbeat(): void {
    this.stopHeartbeat()
    
    this.heartbeatTimer = window.setInterval(() => {
      if (this.isConnected()) {
        this.send({ type: 'ping' })
      }
    }, this.config.heartbeatInterval)
  }

  private stopHeartbeat(): void {
    if (this.heartbeatTimer) {
      clearInterval(this.heartbeatTimer)
      this.heartbeatTimer = null
    }
  }

  private scheduleReconnect(): void {
    this.clearReconnectTimer()
    
    this.reconnectAttempts++
    this.setStatus('reconnecting')
    
    console.log(`Scheduling reconnect attempt ${this.reconnectAttempts}/${this.config.maxReconnectAttempts}`)
    this.handlers.onReconnect?.(this.reconnectAttempts)
    
    this.reconnectTimer = window.setTimeout(() => {
      this.connect()
    }, this.config.reconnectInterval)
  }

  private clearReconnectTimer(): void {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer)
      this.reconnectTimer = null
    }
  }
}

// 项目 WebSocket 服务类
export class ProjectWebSocketService {
  private wsService: WebSocketService | null = null
  private projectGuid: string | null = null
  private handlers: WebSocketEventHandlers

  constructor(handlers: WebSocketEventHandlers = {}) {
    this.handlers = handlers
  }

  public async connectToProject(projectGuid: string): Promise<void> {
    // 如果已经连接到其他项目，先断开
    if (this.wsService && this.projectGuid !== projectGuid) {
      this.disconnect()
    }

    this.projectGuid = projectGuid

    const config: WebSocketConfig = {
      url: `/ws/project/${projectGuid}`,
      protocols: []
    }

    const eventHandlers: WebSocketEventHandlers = {
      onOpen: () => {
        // 连接成功后加入项目
        this.wsService?.send({
          type: 'join_project',
          projectGuid: projectGuid
        })
        this.handlers.onOpen?.()
      },
      onClose: (event) => {
        this.handlers.onClose?.(event)
      },
      onError: (error) => {
        this.handlers.onError?.(error)
      },
      onMessage: (message) => {
        this.handleMessage(message)
        this.handlers.onMessage?.(message)
      },
      onReconnect: (attempt) => {
        this.handlers.onReconnect?.(attempt)
      },
      onReconnectFailed: () => {
        this.handlers.onReconnectFailed?.()
      }
    }

    this.wsService = new WebSocketService(config, eventHandlers)
    await this.wsService.connect()
  }

  public disconnect(): void {
    if (this.wsService) {
      // 发送离开项目消息
      if (this.projectGuid) {
        this.wsService.send({
          type: 'leave_project',
          projectGuid: this.projectGuid
        })
      }
      
      this.wsService.disconnect()
      this.wsService = null
      this.projectGuid = null
    }
  }

  public sendUserFeedback(feedback: any): void {
    if (this.wsService && this.projectGuid) {
      this.wsService.send({
        type: 'user_feedback',
        projectGuid: this.projectGuid,
        data: feedback
      })
    }
  }

  public getStatus(): WebSocketStatus {
    return this.wsService?.getStatus() || 'disconnected'
  }

  public isConnected(): boolean {
    return this.wsService?.isConnected() || false
  }

  public reconnect(): void {
    if (this.wsService && this.projectGuid) {
      this.wsService.reconnect()
    }
  }

  private handleMessage(message: WebSocketServerMessage): void {
    switch (message.type) {
      case 'project_stage_update':
        // 项目阶段更新
        console.log('Project stage updated:', message.data)
        break
      case 'project_message':
        // 项目新消息
        console.log('New project message:', message.data)
        break
      case 'project_status_change':
        // 项目状态变更
        console.log('Project status changed:', message.data)
        break
      case 'agent_message':
        // Agent 消息
        console.log('Agent message:', message.data)
        break
      case 'user_feedback_response':
        // 用户反馈响应
        console.log('User feedback response:', message.data)
        break
      case 'pong':
        // 心跳响应
        console.log('Heartbeat response received')
        break
      case 'error':
        // 错误消息
        console.error('WebSocket error message:', message.data)
        break
      default:
        console.log('Unknown message type:', message.type)
    }
  }
}

// 导出单例实例
export const projectWebSocketService = new ProjectWebSocketService()
