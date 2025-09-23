import { ref, onUnmounted, watch } from 'vue'
import { projectWebSocketService } from '@/utils/websocket'
import type { WebSocketStatus, WebSocketServerMessage } from '@/types/websocket'
import type { DevStage, ConversationMessage } from '@/types/project'

export function useWebSocket(projectGuid: string) {
  const status = ref<WebSocketStatus>('disconnected')
  const isConnected = ref(false)
  const error = ref<string | null>(null)
  const reconnectAttempts = ref(0)

  // 项目相关状态
  const projectStages = ref<DevStage[]>([])
  const projectMessages = ref<ConversationMessage[]>([])
  const projectStatus = ref<string>('')
  const projectInfo = ref<any>(null)

  // 连接 WebSocket
  const connect = async () => {
    try {
      error.value = null
      await projectWebSocketService.connectToProject(projectGuid)
    } catch (err) {
      error.value = err instanceof Error ? err.message : '连接失败'
      console.error('WebSocket connection failed:', err)
    }
  }

  // 断开连接
  const disconnect = () => {
    projectWebSocketService.disconnect()
  }

  // 发送用户反馈
  const sendUserFeedback = (feedback: any) => {
    projectWebSocketService.sendUserFeedback(feedback)
  }

  // 重连
  const reconnect = () => {
    projectWebSocketService.reconnect()
  }

  // 处理 WebSocket 消息
  const handleMessage = (message: WebSocketServerMessage) => {
    switch (message.type) {
      case 'project_stage_update':
        // 更新项目阶段
        const stageData = message.data as DevStage
        const existingStageIndex = projectStages.value.findIndex(stage => stage.id === stageData.id)
        
        if (existingStageIndex >= 0) {
          // 更新现有阶段
          projectStages.value[existingStageIndex] = stageData
        } else {
          // 添加新阶段
          projectStages.value.push(stageData)
        }
        
        // 按创建时间排序
        projectStages.value.sort((a, b) => a.id.localeCompare(b.id))
        break

      case 'project_message':
        // 更新项目消息
        const messageData = message.data as ConversationMessage
        const existingMessageIndex = projectMessages.value.findIndex(msg => msg.id === messageData.id)
        
        if (existingMessageIndex >= 0) {
          // 更新现有消息
          projectMessages.value[existingMessageIndex] = messageData
        } else {
          // 添加新消息
          projectMessages.value.push(messageData)
        }
        
        // 按创建时间排序
        projectMessages.value.sort((a, b) => new Date(a.created_at).getTime() - new Date(b.created_at).getTime())
        break

      case 'project_status_change':
        // 更新项目状态
        projectStatus.value = message.data.status
        break

      case 'agent_message':
        // 处理 Agent 消息
        console.log('Agent message received:', message.data)
        break

      case 'user_feedback_response':
        // 处理用户反馈响应
        console.log('User feedback response:', message.data)
        break

      case 'error':
        // 处理错误消息
        error.value = message.data.message || '未知错误'
        console.error('WebSocket error:', message.data)
        break

      default:
        console.log('Unknown message type:', message.type)
    }
  }

  // 设置事件处理器
  const eventHandlers = {
    onOpen: () => {
      status.value = 'connected'
      isConnected.value = true
      error.value = null
      reconnectAttempts.value = 0
    },
    onClose: () => {
      status.value = 'disconnected'
      isConnected.value = false
    },
    onError: () => {
      status.value = 'error'
      isConnected.value = false
    },
    onMessage: handleMessage,
    onReconnect: (attempt: number) => {
      status.value = 'reconnecting'
      reconnectAttempts.value = attempt
    },
    onReconnectFailed: () => {
      status.value = 'error'
      error.value = '重连失败'
    }
  }

  // 监听项目 GUID 变化
  watch(() => projectGuid, (newGuid, oldGuid) => {
    if (newGuid && newGuid !== oldGuid) {
      // 如果已经连接，先断开
      if (isConnected.value) {
        disconnect()
      }
      // 连接到新项目
      connect()
    }
  }, { immediate: true })

  // 组件卸载时断开连接
  onUnmounted(() => {
    disconnect()
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
    projectStatus,
    projectInfo,
    
    // 方法
    connect,
    disconnect,
    reconnect,
    sendUserFeedback
  }
}
