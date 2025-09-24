<template>
  <div class="websocket-debug">
    <n-card title="WebSocket 调试工具" class="debug-card">
      <!-- 连接状态 -->
      <n-space vertical>
        <n-alert 
          :type="connectionStatus.type" 
          :title="connectionStatus.title"
          :description="connectionStatus.description"
        />
        
        <!-- 连接控制 -->
        <n-space>
          <n-button 
            type="primary" 
            @click="connect" 
            :disabled="isConnected"
            :loading="isConnecting"
          >
            连接
          </n-button>
          <n-button 
            type="error" 
            @click="disconnect" 
            :disabled="!isConnected"
          >
            断开
          </n-button>
          <n-button 
            type="warning" 
            @click="reconnect" 
            :disabled="isConnecting"
          >
            重连
          </n-button>
        </n-space>

        <!-- 项目 GUID 输入 -->
        <n-form-item label="项目 GUID">
          <n-input 
            v-model:value="projectGuid" 
            placeholder="输入项目 GUID"
            :disabled="isConnected"
          />
        </n-form-item>

        <!-- 消息发送 -->
        <n-card title="发送消息" size="small">
          <n-space vertical>
            <n-select 
              v-model:value="messageType" 
              :options="messageTypeOptions"
              placeholder="选择消息类型"
            />
            <n-input 
              v-model:value="messageData" 
              type="textarea" 
              placeholder="消息数据 (JSON 格式)"
              :rows="3"
            />
            <n-button 
              type="primary" 
              @click="sendMessage" 
              :disabled="!isConnected || !messageType"
            >
              发送消息
            </n-button>
          </n-space>
        </n-card>

        <!-- 消息历史 -->
        <n-card title="消息历史" size="small">
          <template #header-extra>
            <n-button size="small" @click="clearMessages">清空</n-button>
          </template>
          <div class="message-list">
            <div 
              v-for="(msg, index) in messageHistory" 
              :key="index"
              class="message-item"
              :class="msg.direction"
            >
              <div class="message-header">
                <n-tag :type="msg.direction === 'incoming' ? 'success' : 'info'">
                  {{ msg.direction === 'incoming' ? '接收' : '发送' }}
                </n-tag>
                <span class="message-time">{{ msg.timestamp }}</span>
                <span class="message-type">{{ msg.type }}</span>
              </div>
              <div class="message-content">
                <pre>{{ JSON.stringify(msg.data, null, 2) }}</pre>
              </div>
            </div>
          </div>
        </n-card>

        <!-- 统计信息 -->
        <n-card title="连接统计" size="small">
          <n-descriptions :column="2" size="small">
            <n-descriptions-item label="连接状态">
              <n-tag :type="isConnected ? 'success' : 'error'">
                {{ isConnected ? '已连接' : '未连接' }}
              </n-tag>
            </n-descriptions-item>
            <n-descriptions-item label="重连次数">
              {{ reconnectAttempts }}
            </n-descriptions-item>
            <n-descriptions-item label="接收消息数">
              {{ incomingMessageCount }}
            </n-descriptions-item>
            <n-descriptions-item label="发送消息数">
              {{ outgoingMessageCount }}
            </n-descriptions-item>
          </n-descriptions>
        </n-card>
      </n-space>
    </n-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { NCard, NSpace, NButton, NAlert, NFormItem, NInput, NSelect, NTag, NDescriptions, NDescriptionsItem, useMessage } from 'naive-ui'
import { useUserStore } from '@/stores/user'
import type { WebSocketClientMessage } from '@/types/websocket'

// 消息提示
const messageApi = useMessage()
const userStore = useUserStore()

// 响应式数据
const projectGuid = ref('')
const messageType = ref('')
const messageData = ref('')
const messageHistory = ref<Array<{
  direction: 'incoming' | 'outgoing'
  type: string
  data: any
  timestamp: string
}>>([])
const incomingMessageCount = ref(0)
const outgoingMessageCount = ref(0)

// 消息提示
const message = useMessage()

// WebSocket 相关状态
const wsStatus = ref<'connecting' | 'connected' | 'disconnected' | 'error' | 'reconnecting'>('disconnected')
const isConnected = ref(false)
const wsError = ref<string | null>(null)
const reconnectAttempts = ref(0)
const ws = ref<any>(null)

// 监听 projectGuid 变化
watch(projectGuid, (newGuid: string) => {
  if (newGuid && newGuid !== '') {
    console.log('Project GUID changed to:', newGuid)
  }
})

// 计算属性
const isConnecting = computed(() => wsStatus.value === 'connecting' || wsStatus.value === 'reconnecting')

const connectionStatus = computed(() => {
  if (isConnected.value) {
    return {
      type: 'success' as const,
      title: 'WebSocket 已连接',
      description: `项目 GUID: ${projectGuid.value}`
    }
  } else if (wsError.value) {
    return {
      type: 'error' as const,
      title: 'WebSocket 连接失败',
      description: wsError.value
    }
  } else if (isConnecting.value) {
    return {
      type: 'warning' as const,
      title: 'WebSocket 连接中...',
      description: `重连次数: ${reconnectAttempts.value}`
    }
  } else {
    return {
      type: 'info' as const,
      title: 'WebSocket 未连接',
      description: '点击连接按钮开始调试'
    }
  }
})

// 消息类型选项
const messageTypeOptions = [
  { label: 'Ping', value: 'ping' },
  { label: '加入项目', value: 'join_project' },
  { label: '离开项目', value: 'leave_project' },
  { label: '用户反馈', value: 'user_feedback' }
]

// 方法
const connect = async () => {
  if (!projectGuid.value) {
    messageApi.warning('请输入项目 GUID')
    return
  }
  
  try {
    wsStatus.value = 'connecting'
    wsError.value = null
    
    // 使用新的 WebSocket 工具
    const { createProjectWebSocket } = await import('@/utils/websocket')
    ws.value = createProjectWebSocket(projectGuid.value, {
      onOpen: () => {
        console.log('WebSocket connected successfully')
        wsStatus.value = 'connected'
        isConnected.value = true
        reconnectAttempts.value = 0
        messageApi.success('WebSocket 连接成功')
      },
      onClose: (event) => {
        console.log('WebSocket closed:', event.code, event.reason)
        wsStatus.value = 'disconnected'
        isConnected.value = false
        
        if (!event.wasClean) {
          wsError.value = `WebSocket 连接异常关闭: ${event.code} - ${event.reason}`
          messageApi.error(`连接异常关闭: ${event.code}`)
        }
      },
      onError: (error) => {
        console.error('WebSocket error:', error)
        wsStatus.value = 'error'
        isConnected.value = false
        wsError.value = 'WebSocket 连接错误'
        messageApi.error('WebSocket 连接失败')
      },
      onMessage: (receivedMessage) => {
        try {
          console.log('Received WebSocket message:', receivedMessage)
          
          // 添加到消息历史
          addToHistory('incoming', receivedMessage.type, receivedMessage)
          incomingMessageCount.value++
        } catch (error) {
          console.error('Failed to parse WebSocket message:', error)
        }
      },
      onReconnect: (attempt) => {
        wsStatus.value = 'reconnecting'
        reconnectAttempts.value = attempt
        messageApi.info(`正在重连... (${attempt}/3)`)
      }
    })
    
    await ws.value.connect()
    
  } catch (error) {
    wsStatus.value = 'error'
    wsError.value = error instanceof Error ? error.message : '连接失败'
    messageApi.error(`连接失败: ${error}`)
  }
}

const disconnect = () => {
  if (ws.value) {
    ws.value.disconnect()
    ws.value = null
  }
  wsStatus.value = 'disconnected'
  isConnected.value = false
  messageApi.info('WebSocket 已断开')
}

const reconnect = () => {
  if (ws.value) {
    ws.value.reconnect()
  }
}


const sendMessage = () => {
  if (!isConnected.value || !messageType.value || !ws.value) {
    messageApi.warning('请先连接并选择消息类型')
    return
  }

  try {
    let data: any = {}
    
    // 根据消息类型构建数据
    switch (messageType.value) {
      case 'ping':
        data = {}
        break
      case 'join_project':
        data = { projectGuid: projectGuid.value }
        break
      case 'leave_project':
        data = { projectGuid: projectGuid.value }
        break
      case 'user_feedback':
        data = { feedback: messageData.value || '测试反馈' }
        break
    }

    // 如果有自定义数据，尝试解析 JSON
    if (messageData.value && messageType.value !== 'user_feedback') {
      try {
        data = { ...data, ...JSON.parse(messageData.value) }
      } catch (e) {
        messageApi.warning('自定义数据不是有效的 JSON 格式')
        return
      }
    }

    const wsMessage = {
      type: messageType.value,
      projectGuid: projectGuid.value,
      data,
      timestamp: new Date().toISOString(),
      id: `debug_${Date.now()}`
    }

    // 使用新的 WebSocket 工具发送消息
    ws.value.send(wsMessage)
    
    // 添加到消息历史
    addToHistory('outgoing', wsMessage.type, wsMessage)
    outgoingMessageCount.value++
    
    messageApi.success('消息发送成功')
  } catch (error) {
    messageApi.error(`发送失败: ${error}`)
  }
}

const addToHistory = (direction: 'incoming' | 'outgoing', type: string, data: any) => {
  messageHistory.value.unshift({
    direction,
    type,
    data,
    timestamp: new Date().toLocaleTimeString()
  })
  
  // 限制历史记录数量
  if (messageHistory.value.length > 100) {
    messageHistory.value = messageHistory.value.slice(0, 100)
  }
}

const clearMessages = () => {
  messageHistory.value = []
  incomingMessageCount.value = 0
  outgoingMessageCount.value = 0
}

// 监听 WebSocket 消息
onMounted(() => {
  // 这里可以添加消息监听逻辑
})

onUnmounted(() => {
  disconnect()
})
</script>

<style scoped>
.websocket-debug {
  padding: 20px;
  max-width: 1200px;
  margin: 0 auto;
}

.debug-card {
  margin-bottom: 20px;
}

.message-list {
  max-height: 400px;
  overflow-y: auto;
  border: 1px solid #e0e0e6;
  border-radius: 6px;
  padding: 10px;
}

.message-item {
  margin-bottom: 15px;
  padding: 10px;
  border-radius: 6px;
  border-left: 4px solid;
}

.message-item.incoming {
  background-color: #f6ffed;
  border-left-color: #52c41a;
}

.message-item.outgoing {
  background-color: #e6f7ff;
  border-left-color: #1890ff;
}

.message-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 8px;
}

.message-time {
  font-size: 12px;
  color: #666;
}

.message-type {
  font-weight: bold;
  color: #333;
}

.message-content {
  background-color: #fafafa;
  padding: 8px;
  border-radius: 4px;
}

.message-content pre {
  margin: 0;
  font-size: 12px;
  white-space: pre-wrap;
  word-break: break-all;
}
</style>
