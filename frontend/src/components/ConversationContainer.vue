<template>
  <div class="conversation-container">
    <!-- 开发阶段进度 - 横向展示 -->
    <div class="progress-section">
      <DevStages 
        :stages="devStages" 
        :current-progress="currentProgress"
        layout="horizontal"
      />
    </div>
    
    <!-- 对话消息列表 -->
    <div class="conversation-messages" ref="messagesContainer">
      <ConversationMessage
        v-for="message in messages"
        :key="message.id"
        :message="message"
        @toggle-expanded="toggleMessageExpanded"
      />
      
      <!-- 加载状态 -->
      <div v-if="isLoading" class="loading-message">
        <div class="loading-avatar">
          <n-icon size="20" color="white">
            <LoadingIcon />
          </n-icon>
        </div>
        <div class="loading-content">
          <div class="loading-text">AI Agent 正在思考中...</div>
          <div class="loading-dots">
            <span></span>
            <span></span>
            <span></span>
          </div>
        </div>
      </div>
    </div>
    
    <!-- 底部输入框 -->
    <div class="input-section">
      <SmartInput
        v-model="inputValue"
        placeholder="输入您的需求或问题..."
        @send="handleSendMessage"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick, h, onMounted, onUnmounted } from 'vue'
import { NIcon } from 'naive-ui'
import ConversationMessage from './ConversationMessage.vue'
import DevStages from './DevStages.vue'
import SmartInput from './common/SmartInput.vue'
import { useProjectStore } from '@/stores/project'
import type { ConversationMessage as ConversationMessageType, DevStage } from '@/types/project'

interface Props {
  projectGuid: string
  requirements: string
}

const props = defineProps<Props>()
const projectStore = useProjectStore()

// 响应式数据
const messages = ref<ConversationMessageType[]>([])
const devStages = ref<DevStage[]>([])
const currentProgress = ref(0)
const isLoading = ref(false)
const messagesContainer = ref<HTMLElement>()
const inputValue = ref('')

// 定时刷新
let refreshTimer: number | null = null

// 加载开发阶段
const loadDevStages = async () => {
  try {
    const stages = await projectStore.getProjectStages(props.projectGuid)
    if (stages) {
      devStages.value = stages
      updateCurrentProgress()
    }
  } catch (error) {
    console.error('加载开发阶段失败:', error)
  }
}

// 加载对话历史
const loadConversations = async () => {
  try {
    const conversations = await projectStore.getProjectMessages(props.projectGuid)
    if (conversations) {
      messages.value = conversations.data
      scrollToBottom()
    }
  } catch (error) {
    console.error('加载对话历史失败:', error)
  }
}

// 智能合并对话历史（保持用户操作状态）
const mergeConversations = async () => {
  try {
    const conversations = await projectStore.getProjectMessages(props.projectGuid)
    if (!conversations || !conversations.data) return
    
    const newMessages = conversations.data
    const currentMessages = messages.value
    
    // 如果消息数量相同，检查是否有内容更新
    if (newMessages.length === currentMessages.length) {
      let hasUpdates = false
      const updatedMessages = newMessages.map((newMsg, index) => {
        const currentMsg = currentMessages[index]
        
        // 检查消息内容是否有更新
        if (currentMsg && (
          currentMsg.content !== newMsg.content ||
          currentMsg.markdown_content !== newMsg.markdown_content ||
          currentMsg.updated_at !== newMsg.updated_at
        )) {
          hasUpdates = true
          // 保持用户的展开/折叠状态
          return {
            ...newMsg,
            is_expanded: currentMsg.is_expanded
          }
        }
        
        // 没有更新，保持原消息（包括用户状态）
        return currentMsg || newMsg
      })
      
      if (hasUpdates) {
        messages.value = updatedMessages
      }
      return
    }
    
    // 消息数量不同，进行完整合并
    const existingMessagesMap = new Map()
    currentMessages.forEach(msg => {
      existingMessagesMap.set(msg.id, {
        is_expanded: msg.is_expanded,
        // 可以保存其他用户操作状态
      })
    })
    
    // 合并新消息，保持用户状态
    const mergedMessages = newMessages.map(newMsg => {
      const existingState = existingMessagesMap.get(newMsg.id)
      if (existingState) {
        // 保持用户的展开/折叠状态
        return {
          ...newMsg,
          is_expanded: existingState.is_expanded
        }
      }
      return newMsg
    })
    
    // 检查是否有新消息
    const hasNewMessages = mergedMessages.length > currentMessages.length
    const lastMessageId = currentMessages.length > 0 ? currentMessages[currentMessages.length - 1].id : null
    const newLastMessageId = mergedMessages.length > 0 ? mergedMessages[mergedMessages.length - 1].id : null
    
    // 更新消息列表
    messages.value = mergedMessages
    
    // 如果有新消息，滚动到底部
    if (hasNewMessages && lastMessageId !== newLastMessageId) {
      scrollToBottom()
    }
    
  } catch (error) {
    console.error('合并对话历史失败:', error)
  }
}


// 更新开发阶段状态
const updateDevStage = (stageId: string, status: 'pending' | 'in_progress' | 'done' | 'failed') => {
  const stage = devStages.value.find(s => s.id === stageId)
  if (stage) {
    stage.status = status
    updateCurrentProgress()
  }
}

// 更新当前进度
const updateCurrentProgress = () => {
  const completedStages = devStages.value.filter(s => s.status === 'done')
  const inProgressStage = devStages.value.find(s => s.status === 'in_progress')
  
  if (inProgressStage) {
    currentProgress.value = inProgressStage.progress
  } else if (completedStages.length > 0) {
    const lastCompleted = completedStages[completedStages.length - 1]
    currentProgress.value = lastCompleted.progress
  } else {
    currentProgress.value = 0
  }
}

// 滚动到底部
const scrollToBottom = () => {
  nextTick(() => {
    if (messagesContainer.value) {
      messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
    }
  })
}

// 切换消息展开状态
const toggleMessageExpanded = (messageId: string) => {
  const message = messages.value.find(m => m.id === messageId)
  if (message) {
    message.is_expanded = !message.is_expanded
  }
}

// 发送消息
const handleSendMessage = async (content: string) => {
  if (!content.trim()) return
  
  isLoading.value = true
  
  try {
    // 添加用户消息
    const userMessage = await projectStore.addChatMessage(props.projectGuid, {
      type: 'user',
      content: content.trim(),
      is_expanded: false
    })
    
    if (userMessage) {
      messages.value.push(userMessage)
      scrollToBottom()
    }
    
    // 清空输入框
    inputValue.value = ''
    
    // 这里可以添加发送到后端的逻辑
    // 后端会通过WebSocket推送AI回复
    
  } catch (error) {
    console.error('发送消息失败:', error)
  } finally {
    isLoading.value = false
  }
}

// 定时刷新数据
const startAutoRefresh = () => {
  refreshTimer = window.setInterval(async () => {
    await mergeConversations() // 使用智能合并而不是完全替换
    await loadDevStages()
  }, 5000) // 每5秒刷新一次
}

const stopAutoRefresh = () => {
  if (refreshTimer) {
    window.clearInterval(refreshTimer)
    refreshTimer = null
  }
}


// 图标组件
const LoadingIcon = () => h('svg', { 
  viewBox: '0 0 24 24', 
  fill: 'currentColor',
  style: 'width: 1em; height: 1em;'
}, [
  h('path', { d: 'M12 2A10 10 0 0 0 2 12a10 10 0 0 0 10 10 10 10 0 0 0 10-10A10 10 0 0 0 12 2zm0 18a8 8 0 0 1-8-8 8 8 0 0 1 8-8 8 8 0 0 1 8 8 8 8 0 0 1-8 8z' }),
  h('path', { 
    d: 'M12 4a8 8 0 0 1 8 8 8 8 0 0 1-8 8',
    style: 'opacity: 0.3;'
  })
])

// 初始化
const initialize = async () => {
  await loadDevStages()
  await loadConversations()
  
  // 如果没有对话历史，添加初始消息
  if (messages.value.length === 0) {
    // 添加用户需求消息
    const userMessage = await projectStore.addChatMessage(props.projectGuid, {
      type: 'user',
      content: props.requirements,
      is_expanded: false
    })
    if (userMessage) {
      messages.value.push(userMessage)
    }
    
    // 系统消息将通过WebSocket推送
  }
  
  // 启动定时刷新
  startAutoRefresh()
}

// 生命周期钩子
onMounted(() => {
  initialize()
})

onUnmounted(() => {
  stopAutoRefresh()
})
</script>

<style scoped>
.conversation-container {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: #f8fafc;
}

.progress-section {
  flex-shrink: 0;
  padding: var(--spacing-sm) var(--spacing-lg);
  background: white;
  border-bottom: 1px solid #e2e8f0;
}

.conversation-messages {
  flex: 1;
  overflow-y: auto;
  padding: var(--spacing-lg);
  background: #f8fafc;
}

.input-section {
  flex-shrink: 0;
  padding: var(--spacing-md) var(--spacing-lg);
  background: white;
  border-top: 1px solid #e2e8f0;
}

/* 加载状态样式 */
.loading-message {
  display: flex;
  align-items: flex-start;
  gap: var(--spacing-sm);
  margin-bottom: var(--spacing-lg);
}

.loading-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: #D69E2E;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  animation: pulse 2s infinite;
}

.loading-content {
  background: white;
  border: 1px solid var(--border-color);
  padding: var(--spacing-md) var(--spacing-lg);
  border-radius: var(--border-radius-lg);
  border-bottom-left-radius: var(--border-radius-sm);
}

.loading-text {
  font-size: 0.9rem;
  color: var(--text-secondary);
  margin-bottom: var(--spacing-xs);
}

.loading-dots {
  display: flex;
  gap: 4px;
}

.loading-dots span {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #D69E2E;
  animation: bounce 1.4s infinite ease-in-out both;
}

.loading-dots span:nth-child(1) {
  animation-delay: -0.32s;
}

.loading-dots span:nth-child(2) {
  animation-delay: -0.16s;
}

@keyframes pulse {
  0% {
    box-shadow: 0 0 0 0 rgba(214, 158, 46, 0.7);
  }
  70% {
    box-shadow: 0 0 0 10px rgba(214, 158, 46, 0);
  }
  100% {
    box-shadow: 0 0 0 0 rgba(214, 158, 46, 0);
  }
}

@keyframes bounce {
  0%, 80%, 100% {
    transform: scale(0);
  }
  40% {
    transform: scale(1);
  }
}

/* 滚动条样式 */
.conversation-messages::-webkit-scrollbar {
  width: 6px;
}

.conversation-messages::-webkit-scrollbar-track {
  background: transparent;
}

.conversation-messages::-webkit-scrollbar-thumb {
  background: #c1c1c1;
  border-radius: 3px;
}

.conversation-messages::-webkit-scrollbar-thumb:hover {
  background: #a8a8a8;
}
</style>
