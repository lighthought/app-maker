import { ref, watch, onUnmounted } from 'vue'
import { useUserStore } from '@/stores/user'
import { AppConfig } from './config'
import type { ConversationMessage, DevStage, ProjectInfoUpdate } from '@/types/project'
import type { WebSocketServerMessage, WebSocketClientMessage } from '@/types/websocket'

// WebSocket 连接状态
export type WebSocketStatus = 'connecting' | 'connected' | 'disconnected' | 'error' | 'reconnecting'

// WebSocket 事件处理器
export interface WebSocketEventHandlers {
  onOpen?: () => void
  onClose?: (event: CloseEvent) => void
  onError?: (error: Event) => void
  onMessage?: (message: WebSocketServerMessage) => void
  onReconnect?: (attempt: number) => void
}

// WebSocket 拦截器
export interface WebSocketInterceptor {
  onConnect?: (url: string) => string
  onMessage?: (message: any) => any
  onError?: (error: Event) => void
  onClose?: (event: CloseEvent) => void
}

// 认证拦截器
const authInterceptor: WebSocketInterceptor = {
  onConnect: (url: string) => {
    const userStore = useUserStore()
    if (!userStore.token) {
      throw new Error('No authentication token available')
    }

    if (url.includes('token=')) {
      return url
    }
    
    const token = userStore.token.startsWith('Bearer ') 
      ? userStore.token 
      : `Bearer ${userStore.token}`
    
    const separator = url.includes('?') ? '&' : '?'
    
    return `${url}${separator}token=${encodeURIComponent(token)}`
  }
}

// WebSocket 管理器类
class WebSocketManager {
  private ws: WebSocket | null = null
  private url: string = ''
  private protocols?: string | string[]
  private interceptors: WebSocketInterceptor[] = []
  private handlers: WebSocketEventHandlers = {}
  private status: WebSocketStatus = 'disconnected'
  private reconnectAttempts = 0
  private reconnectTimer: number | null = null
  private heartbeatTimer: number | null = null
  private isManualDisconnect = false

  constructor(interceptors: WebSocketInterceptor[] = []) {
    this.interceptors = [...interceptors, authInterceptor] // 默认添加认证拦截器
  }

  // 添加拦截器
  public addInterceptor(interceptor: WebSocketInterceptor) {
    this.interceptors.push(interceptor)
  }

  // 设置事件处理器
  public setHandlers(handlers: WebSocketEventHandlers) {
    this.handlers = handlers
  }

  // 连接 WebSocket
  public async connect(url: string, protocols?: string | string[]): Promise<void> {
    return new Promise((resolve, reject) => {
      // 应用连接拦截器
      let finalUrl = url
      for (const interceptor of this.interceptors) {
        if (interceptor.onConnect) {
          try {
            finalUrl = interceptor.onConnect(finalUrl)
          } catch (error) {
            reject(error)
            return
          }
        }
      }

      this.url = finalUrl
      this.protocols = protocols
      this.status = 'connecting'

      try {
        this.ws = new WebSocket(finalUrl, protocols)
        
        this.ws.onopen = () => {
          console.log('WebSocket connected')
          this.status = 'connected'
          this.reconnectAttempts = 0
          this.startHeartbeat()
          this.handlers.onOpen?.()
          resolve()
        }

        this.ws.onmessage = (event) => {
          try {
            let message = JSON.parse(event.data)
            
            // 应用消息拦截器
            for (const interceptor of this.interceptors) {
              if (interceptor.onMessage) {
                message = interceptor.onMessage(message)
              }
            }
            
            if (message.type !== 'pong') {
              console.log('WebSocket message received:', message)
            }
            this.handlers.onMessage?.(message)
          } catch (error) {
            console.error('Failed to parse WebSocket message:', error)
          }
        }

        this.ws.onerror = (event) => {
          console.error('WebSocket error:', event)
          this.status = 'error'
          
          // 应用错误拦截器
          for (const interceptor of this.interceptors) {
            if (interceptor.onError) {
              interceptor.onError(event)
            }
          }
          
          this.handlers.onError?.(event)
          reject(event)
        }

        this.ws.onclose = (event) => {
          console.log('WebSocket closed:', event.code, event.reason)
          this.status = 'disconnected'
          this.stopHeartbeat()
          
          // 应用关闭拦截器
          for (const interceptor of this.interceptors) {
            if (interceptor.onClose) {
              interceptor.onClose(event)
            }
          }
          
          this.handlers.onClose?.(event)
          
          // 如果不是手动断开，尝试重连
          if (!this.isManualDisconnect && this.reconnectAttempts < 3) {
            this.scheduleReconnect()
          }
        }

      } catch (error) {
        this.status = 'error'
        reject(error)
      }
    })
  }

  // 发送消息
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
      if (messageWithTimestamp.type !== 'ping') {
        console.log('WebSocket message sent:', messageWithTimestamp)
      }
    } catch (error) {
      console.error('Failed to send WebSocket message:', error)
    }
  }

  // 断开连接
  public disconnect(): void {
    this.isManualDisconnect = true
    this.stopHeartbeat()
    this.clearReconnectTimer()
    
    if (this.ws) {
      this.ws.close(1000, 'Manual disconnect')
      this.ws = null
    }
    
    this.status = 'disconnected'
    console.log('WebSocket disconnected manually')
  }

  // 重连
  public reconnect(): void {
    this.disconnect()
    this.reconnectAttempts = 0
    this.connect(this.url, this.protocols)
  }

  // 获取状态
  public getStatus(): WebSocketStatus {
    return this.status
  }

  // 是否已连接
  public isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN
  }

  // 生成消息 ID
  private generateMessageId(): string {
    return `msg_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
  }

  // 开始心跳
  private startHeartbeat(): void {
    this.stopHeartbeat()
    
    this.heartbeatTimer = window.setInterval(() => {
      if (this.isConnected()) {
        this.send({ type: 'ping' })
      }
    }, 30000) // 30秒心跳
  }

  // 停止心跳
  private stopHeartbeat(): void {
    if (this.heartbeatTimer) {
      clearInterval(this.heartbeatTimer)
      this.heartbeatTimer = null
    }
  }

  // 安排重连
  private scheduleReconnect(): void {
    this.clearReconnectTimer()
    
    this.reconnectAttempts++
    this.status = 'reconnecting'
    
    console.log(`Scheduling reconnect attempt ${this.reconnectAttempts}/3`)
    this.handlers.onReconnect?.(this.reconnectAttempts)
    
    this.reconnectTimer = window.setTimeout(() => {
      this.connect(this.url, this.protocols)
    }, 5000) // 5秒后重连
  }

  // 清除重连定时器
  private clearReconnectTimer(): void {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer)
      this.reconnectTimer = null
    }
  }
}

// 构建 WebSocket URL
function buildWebSocketUrl(path: string): string {
  const config = AppConfig.getInstance().getConfig()
  const baseUrl = config.apiBaseUrl.replace('/api/v1', '')
  
  // 如果是相对路径（以 / 开头），说明是通过 Traefik 代理访问
  if (baseUrl.startsWith('/')) {
    const protocol = window.location.protocol === 'https:' ? 'wss' : 'ws'
    const host = window.location.host
    return `${protocol}://${host}${path}`
  }
  
  // 如果是完整 URL，转换为 WebSocket 协议
  const wsProtocol = baseUrl.startsWith('https') ? 'wss' : 'ws'
  const wsBaseUrl = baseUrl.replace(/^https?/, wsProtocol)
  
  return `${wsBaseUrl}${path}`
}

// 创建项目 WebSocket 连接
export function createProjectWebSocket(projectGuid: string, handlers: WebSocketEventHandlers = {}) {
  const manager = new WebSocketManager()
  manager.setHandlers(handlers)
  
  return {
    // 连接方法
    connect: () => manager.connect(buildWebSocketUrl(`/ws/project/${projectGuid}`)),
    
    // 发送消息方法
    send: (message: WebSocketClientMessage) => manager.send(message),
    
    // 发送加入项目消息
    joinProject: () => manager.send({ type: 'join_project', projectGuid }),
    
    // 发送离开项目消息
    leaveProject: () => manager.send({ type: 'leave_project', projectGuid }),
    
    // 发送用户反馈
    sendUserFeedback: (feedback: any) => manager.send({ 
      type: 'user_feedback', 
      projectGuid, 
      data: feedback 
    }),
    
    // 控制方法
    disconnect: () => manager.disconnect(),
    reconnect: () => manager.reconnect(),
    
    // 状态方法
    getStatus: () => manager.getStatus(),
    isConnected: () => manager.isConnected()
  }
}

// Vue 组合式函数
export function useWebSocket(projectGuid: string) {
  const status = ref<WebSocketStatus>('disconnected')
  const isConnected = ref(false)
  const error = ref<string | null>(null)
  const reconnectAttempts = ref(0)

  // 项目相关状态
  const projectStages = ref<DevStage[]>([])
  const projectMessages = ref<ConversationMessage[]>([])
  const projectInfo = ref<ProjectInfoUpdate>({} as ProjectInfoUpdate)

  // 创建 WebSocket 连接
  const ws = createProjectWebSocket(projectGuid, {
    onOpen: () => {
      status.value = 'connected'
      isConnected.value = true
      error.value = null
      reconnectAttempts.value = 0
      // 连接成功后加入项目
      ws.joinProject()
    },
    onClose: () => {
      status.value = 'disconnected'
      isConnected.value = false
    },
    onError: () => {
      status.value = 'error'
      isConnected.value = false
    },
    onMessage: (message: WebSocketServerMessage) => {
      switch (message.type) {
        case 'project_info_update':
          const infoData = message.data;
          projectInfo.value = infoData;
          break;
        case 'project_stage_update':
          const stageData = message.data
          const existingStageIndex = projectStages.value.findIndex(stage => stage.id === stageData.id)
          
          if (existingStageIndex >= 0) {
            projectStages.value[existingStageIndex] = stageData
          } else {
            projectStages.value.push(stageData)
          }
          
          projectStages.value.sort((a, b) => a.id.localeCompare(b.id))
          break

        case 'project_message':
          const messageData = message.data
          const existingMessageIndex = projectMessages.value.findIndex(msg => msg.id === messageData.id)
          
          if (existingMessageIndex >= 0) {
            projectMessages.value[existingMessageIndex] = messageData
          } else {
            projectMessages.value.push(messageData)
          }
          
          projectMessages.value.sort((a, b) => new Date(a.created_at).getTime() - new Date(b.created_at).getTime())
          break

        case 'error':
          error.value = message.data.message || '未知错误'
          break
      }
    },
    onReconnect: (attempt: number) => {
      status.value = 'reconnecting'
      reconnectAttempts.value = attempt
    }
  })

  // 监听项目 GUID 变化
  watch(() => projectGuid, (newGuid, oldGuid) => {
    if (newGuid && newGuid !== oldGuid) {
      if (isConnected.value) {
        ws.disconnect()
      }
      ws.connect()
    }
  }, { immediate: true })

  // 组件卸载时断开连接
  onUnmounted(() => {
    ws.disconnect()
  })

  return {
    // 连接状态
    status,
    isConnected,
    error,
    reconnectAttempts,
    
    // 项目数据
    projectStages,
    projectMessages,
    projectInfo,
    
    // 方法
    connect: ws.connect,
    disconnect: ws.disconnect,
    reconnect: ws.reconnect,
    sendUserFeedback: ws.sendUserFeedback
  }
}

// 导出 WebSocket 管理器类，供高级用法
export { WebSocketManager }
