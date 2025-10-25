<template>
  <div class="conversation-container">
    <!-- WebSocket 状态指示器 -->
    <div v-if="wsError || wsStatus === 'reconnecting'" class="ws-status-bar">
      <NAlert 
        v-if="wsError" 
        type="error" 
        :title="`${t('common.websocketError')}: ${wsError}`"
        closable
        @close="wsError = null"
      >
        <div style="margin-top: 8px;">
          <n-button size="small" @click="wsReconnect">{{ t('common.reconnect') }}</n-button>
        </div>
      </NAlert>
      <NAlert 
        v-else-if="wsStatus === 'reconnecting'" 
        type="warning" 
        :title="t('common.reconnecting', { attempts: wsReconnectAttempts, max: 5 })"
      />
    </div>

    <!-- 开发阶段进度 - 横向展示 -->
    <div class="progress-section">
      <DevStages 
        :stages="devStages" 
        layout="horizontal"
        @retry-success="handleRetrySuccess"
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
      
      <!-- 用户确认界面 -->
      <div v-if="showConfirmInterface" class="confirm-interface">
        <div class="confirm-message">
          <div class="confirm-avatar">
            <n-icon size="20" color="white">
              <svg viewBox="0 0 24 24">
                <path fill="currentColor" d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z"/>
              </svg>
            </n-icon>
          </div>
          <div class="confirm-content">
            <div class="confirm-title">{{ getConfirmTitle() }}</div>
            <div class="confirm-description">{{ getConfirmDescription() }}</div>
            
            <!-- Epic/Story 编辑界面 -->
            <div v-if="confirmStage === 'plan_epic_and_story'" class="epic-story-editor-wrapper">
              <EpicStoryEditor 
                :project-guid="projectGuid"
                @confirmed="handleConfirmCompleted"
              />
            </div>
            
            <!-- 其他确认界面 -->
            <div v-else class="simple-confirm">
              <n-space>
                <n-button @click="handleSimpleConfirm" type="primary" size="small">
                  确认并继续
                </n-button>
                <n-button @click="handleSimpleSkip" type="info" size="small">
                  跳过确认
                </n-button>
              </n-space>
            </div>
          </div>
        </div>
      </div>
      
      <!-- 加载状态 -->
      <div v-if="isLoading" class="loading-message">
        <div class="loading-avatar">
          <n-icon size="20" color="white">
            <LoadingIcon />
          </n-icon>
        </div>
        <div class="loading-content">
          <div class="loading-text">{{ t('common.aiThinking') }}</div>
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
        v-model:selectedAgent="selectedAgent"
        :agentOptions="agentOptions"
        :agentLocked="isAgentLocked"
        :placeholder="t('common.inputRequirements')"
        @send="handleSendMessage"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { NIcon, NAlert, NButton, NSpace, useMessage } from 'naive-ui'
import ConversationMessage from './ConversationMessage.vue'
import DevStages from './DevStages.vue'
import SmartInput from './common/SmartInput.vue'
import EpicStoryEditor from './EpicStoryEditor.vue'
import { useProjectStore } from '@/stores/project'
import { useWebSocket } from '@/utils/websocket'
import type { ConversationMessage as ConversationMessageType, DevStage, ProjectInfoUpdate } from '@/types/project'
// 导入图标
import { LoadingIcon } from '@/components/icon'

interface Props {
  projectGuid: string
  requirements: string
  project?: any // 接收项目信息
}

const props = defineProps<Props>()
const { t } = useI18n()
const message = useMessage()

// 定义事件
const emit = defineEmits<{
  projectInfoUpdate: [info: ProjectInfoUpdate]
  projectEnvSetup: []
}>()
const projectStore = useProjectStore()

// WebSocket 集成
const {
  status: wsStatus,
  isConnected: wsConnected,
  error: wsError,
  reconnectAttempts: wsReconnectAttempts,
  projectStages: wsProjectStages,
  projectMessages: wsProjectMessages,
  projectInfo: wsProjectInfo,
  connect: wsConnect,
  disconnect: wsDisconnect,
  reconnect: wsReconnect,
  sendUserFeedback: wsSendUserFeedback
} = useWebSocket(props.projectGuid)

// 响应式数据
const messages = ref<ConversationMessageType[]>([])
const devStages = ref<DevStage[]>([])
const isLoading = ref(false)
const messagesContainer = ref<HTMLElement>()
const inputValue = ref('')

// Agent 相关状态
const lastAgentType = ref<string>('') // 最后一次提问的 Agent
const isAgentLocked = ref(false) // 是否锁定 Agent 选择
const selectedAgent = ref<string>('') // 当前选择的 Agent

// 确认界面相关状态
const showConfirmInterface = ref(false)
const confirmStage = ref<string>('')

// Agent 选项列表 - 使用国际化
const agentOptions = computed(() => [
  { label: t('agent.analyst'), value: 'analyst' },
  { label: t('agent.productManager'), value: 'pm' },
  { label: t('agent.uxExpert'), value: 'ux-expert' },
  { label: t('agent.architect'), value: 'architect' },
  { label: t('agent.productOwner'), value: 'po' },
  { label: t('agent.devEngineer'), value: 'dev' },
  { label: t('agent.testEngineer'), value: 'qa' },
])

// 定时刷新（作为 WebSocket 的备用方案）
let refreshTimer: number | null = null

// 清理 WebSocket 缓存数据
const clearWebSocketCache = () => {
  console.log('🧹 [WebSocket] 清理 WebSocket 缓存数据...')
  // 注意：这里不能直接清空 wsProjectStages 和 wsProjectMessages，因为它们是来自 useWebSocket 的响应式数据
  // 我们只能清理本地的合并数据，让系统重新从接口获取
  console.log('🧹 [WebSocket] 缓存清理完成')
}

// 加载开发阶段
const loadDevStages = async () => {
  try {
    console.log('🔄 [DevStages] 开始从接口获取开发阶段数据...')
    const stages = await projectStore.getProjectStages(props.projectGuid)
    if (stages) {
      console.log('✅ [DevStages] 接口获取成功，数据量:', stages.length, '数据:', stages)
      devStages.value = stages
    } else {
      console.log('⚠️ [DevStages] 接口返回空数据')
    }
  } catch (error) {
    console.error('❌ [DevStages] 接口获取失败:', error)
  }
}

// 加载对话历史
const loadConversations = async () => {
  try {
    console.log('🔄 [Messages] 开始从接口获取对话历史数据...')
    const conversations = await projectStore.getProjectMessages(props.projectGuid)
    if (conversations) {
      console.log('✅ [Messages] 接口获取成功，数据量:', conversations.data?.length || 0, '数据:', conversations.data)
      messages.value = conversations.data
      scrollToBottom()
    } else {
      console.log('⚠️ [Messages] 接口返回空数据')
    }
  } catch (error) {
    console.error('❌ [Messages] 接口获取失败:', error)
  }
}

// 同步 WebSocket 数据到本地状态 - 增量追加
const syncWebSocketData = () => {
  console.log('🔄 [WebSocket] 开始同步 WebSocket 数据...')
  
  // 增量同步项目阶段数据
  if (wsProjectStages.value.length > 0) {
    console.log('📊 [DevStages] WebSocket 数据:', wsProjectStages.value.length, '条')
    console.log('📊 [DevStages] 当前本地数据:', devStages.value.length, '条')
    
    // 获取当前本地已有的阶段ID
    const existingStageIds = new Set(devStages.value.map(stage => stage.id))
    
    // 找出需要追加的新阶段
    const newStages = wsProjectStages.value.filter(stage => !existingStageIds.has(stage.id))
    
    // 找出需要更新的已有阶段（ID相同但状态或其他字段有变化）
    const updatedStages = wsProjectStages.value.filter(wsStage => {
      const existingStage = devStages.value.find(stage => stage.id === wsStage.id)
      if (!existingStage) return false
      
      // 检查关键字段是否有变化
      return (
        existingStage.status !== wsStage.status ||
        existingStage.failed_reason !== wsStage.failed_reason ||
        existingStage.progress !== wsStage.progress ||
        existingStage.description !== wsStage.description ||
        existingStage.task_id !== wsStage.task_id
      )
    })
    
    // 追加新阶段
    if (newStages.length > 0) {
      console.log('➕ [DevStages] 发现新阶段:', newStages.length, '条，追加到本地数据')
      devStages.value.push(...newStages)
    }
    
    // 更新已有阶段的状态
    if (updatedStages.length > 0) {
      console.log('🔄 [DevStages] 发现阶段状态更新:', updatedStages.length, '条')
      updatedStages.forEach(wsStage => {
        const existingStageIndex = devStages.value.findIndex(stage => stage.id === wsStage.id)
        if (existingStageIndex !== -1) {
          const oldStatus = devStages.value[existingStageIndex].status
          console.log(`🔄 [DevStages] 更新阶段 ${wsStage.id}: ${oldStatus} → ${wsStage.status}`)
          // 直接更新阶段数据
          devStages.value[existingStageIndex] = wsStage
        }
      })
    }
    
    // 按ID排序保持顺序
    devStages.value.sort((a, b) => a.id.localeCompare(b.id))
    
    console.log('✅ [DevStages] 同步完成，最终数据量:', devStages.value.length, '条')
    if (newStages.length > 0 || updatedStages.length > 0) {
      console.log(`📈 [DevStages] 变更统计: 新增 ${newStages.length} 条，更新 ${updatedStages.length} 条`)
    }
  }
  
  // 增量同步项目消息数据
  if (wsProjectMessages.value.length > 0) {
    console.log('💬 [Messages] WebSocket 数据:', wsProjectMessages.value.length, '条')
    console.log('💬 [Messages] 当前本地数据:', messages.value.length, '条')
    
    // 获取当前本地已有的消息ID
    const existingMessageIds = new Set(messages.value.map(msg => msg.id))
    
    // 找出需要追加的新消息
    const newMessages = wsProjectMessages.value.filter(msg => !existingMessageIds.has(msg.id))
    
    if (newMessages.length > 0) {
      console.log('➕ [Messages] 发现新消息:', newMessages.length, '条，追加到本地数据')
      messages.value.push(...newMessages)
      // 按时间排序
      messages.value.sort((a, b) => new Date(a.created_at).getTime() - new Date(b.created_at).getTime())
      scrollToBottom()
    } else {
      console.log('ℹ️ [Messages] 没有新消息需要追加')
    }
    
    console.log('✅ [Messages] 同步完成，最终数据量:', messages.value.length, '条')
  }
}

// 智能合并对话历史（保持用户操作状态）
const mergeConversations = async () => {
  try {
    console.log('🔄 [Merge] 开始智能合并对话历史...')
    const conversations = await projectStore.getProjectMessages(props.projectGuid)
    if (!conversations || !conversations.data) {
      console.log('⚠️ [Merge] 接口返回空数据，跳过合并')
      return
    }
    
    const newMessages = conversations.data
    const currentMessages = messages.value
    
    console.log('📊 [Merge] 合并前状态:')
    console.log('  - 接口数据:', newMessages.length, '条')
    console.log('  - 本地数据:', currentMessages.length, '条')
    
    // 如果消息数量相同，检查是否有内容更新
    if (newMessages.length === currentMessages.length) {
      console.log('🔍 [Merge] 消息数量相同，检查内容更新...')
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
          console.log('🔄 [Merge] 发现消息更新:', newMsg.id)
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
        console.log('✅ [Merge] 内容更新完成')
        messages.value = updatedMessages
      } else {
        console.log('ℹ️ [Merge] 没有内容更新')
      }
      return
    }
    
    // 消息数量不同，进行完整合并
    console.log('🔀 [Merge] 消息数量不同，进行完整合并...')
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
    
    console.log('📈 [Merge] 合并结果:')
    console.log('  - 最终数据量:', mergedMessages.length, '条')
    console.log('  - 是否有新消息:', hasNewMessages)
    
    // 更新消息列表
    messages.value = mergedMessages
    
    // 如果有新消息，滚动到底部
    if (hasNewMessages && lastMessageId !== newLastMessageId) {
      console.log('📜 [Merge] 检测到新消息，滚动到底部')
      scrollToBottom()
    }
    
    console.log('✅ [Merge] 智能合并完成')
    
  } catch (error) {
    console.error('❌ [Merge] 合并对话历史失败:', error)
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

// 处理重试成功事件
const handleRetrySuccess = async () => {
  console.log('🔄 [Retry] 收到重试成功事件，重新加载开发阶段数据...')
  await loadDevStages()
  console.log('✅ [Retry] 开发阶段数据重新加载完成')
}

// 发送消息
const handleSendMessage = async (content: string, agentType: string) => {
  if (!content.trim() || !agentType) return
  
  isLoading.value = true
  
  try {
    // 如果项目已完成且没有 WebSocket 连接，重新启动连接
    if (isProjectCompleted() && !wsConnected.value) {
      console.log('用户发送新消息，重新启动 WebSocket 连接')
      try {
        await wsConnect()
      } catch (error) {
        console.error('重新启动 WebSocket 连接失败:', error)
      }
    }
    
    // 调用新的 API：向指定 Agent 发送消息
    await projectStore.sendMessageToAgent(
      props.projectGuid, 
      agentType, 
      content.trim()
    )
    
    // 清空输入框
    inputValue.value = ''
    
    // WebSocket 会推送新消息
    
  } catch (error) {
    console.error('发送消息失败:', error)
  } finally {
    isLoading.value = false
  }
}

// 定时刷新数据（作为 WebSocket 的备用方案）
const startAutoRefresh = () => {
  // 只有在 WebSocket 未连接时才启动定时刷新
  if (!wsConnected.value) {
    refreshTimer = window.setInterval(async () => {
      await mergeConversations() // 使用智能合并而不是完全替换
      await loadDevStages()
    }, 5000) // 每5秒刷新一次
  }
}

const stopAutoRefresh = () => {
  if (refreshTimer) {
    window.clearInterval(refreshTimer)
    refreshTimer = null
  }
}

// 监听 WebSocket 连接状态
watch(wsConnected, (connected) => {
  if (connected) {
    console.log('🔗 [WebSocket] 连接成功，停止定时刷新，开始同步数据')
    // WebSocket 连接成功，停止定时刷新
    stopAutoRefresh()
    // 同步 WebSocket 数据
    syncWebSocketData()
  } else {
    console.log('🔌 [WebSocket] 连接断开，清理缓存并启动定时刷新')
    // WebSocket 断开，清理缓存数据
    clearWebSocketCache()
    // 重新从接口获取最新数据
    loadDevStages()
    loadConversations()
    // 启动定时刷新作为备用
    startAutoRefresh()
  }
})

// 监听 WebSocket 数据变化
watch([wsProjectStages, wsProjectMessages], (newValues, oldValues) => {
  if (wsConnected.value) {
    const [newStages, newMessages] = newValues
    const [oldStages, oldMessages] = oldValues || [[], []]
    
    console.log('📡 [WebSocket] 数据变化检测:')
    console.log('  - DevStages: 旧数据', oldStages?.length || 0, '条 → 新数据', newStages?.length || 0, '条')
    console.log('  - Messages: 旧数据', oldMessages?.length || 0, '条 → 新数据', newMessages?.length || 0, '条')
    
    syncWebSocketData()
  }
}, { deep: true })

// 监听消息变化，管理 Agent 锁定状态
watch(messages, (newMessages) => {
  // 找到最后一条包含问题且等待回复的 Agent 消息
  const lastQuestionMsg = [...newMessages]
    .reverse()
    .find(msg => msg.type === 'agent' && msg.has_question && msg.waiting_user_response)
  
  if (lastQuestionMsg) {
    lastAgentType.value = lastQuestionMsg.agent_role || ''
    selectedAgent.value = lastQuestionMsg.agent_role || ''
    isAgentLocked.value = true
    console.log('🔒 [Agent] 锁定 Agent 选择:', lastQuestionMsg.agent_role)
  } else {
    isAgentLocked.value = false
    // 默认选择最后一条 Agent 消息的类型
    const lastAgentMsg = [...newMessages]
      .reverse()
      .find(msg => msg.type === 'agent')
    if (lastAgentMsg && !selectedAgent.value) {
      selectedAgent.value = lastAgentMsg.agent_role || ''
      console.log('🔄 [Agent] 默认选择 Agent:', lastAgentMsg.agent_role)
    }
  }
}, { deep: true })

// 监听项目信息更新
watch(wsProjectInfo, (newInfo) => {
  if (newInfo && Object.keys(newInfo).length > 0) {
    emit('projectInfoUpdate', {
      id: newInfo.id,
      guid: newInfo.guid,
      name: newInfo.name,
      status: newInfo.status,
      description: newInfo.description,
      preview_url: newInfo.preview_url,
      waiting_for_user_confirm: newInfo.waiting_for_user_confirm,
      confirm_stage: newInfo.confirm_stage,
    })
  }
}, { deep: true })

// 监听 setup_environment 阶段状态变化
const previousSetupStatus = ref<string | null>(null)
watch(wsProjectStages, (newStages) => {
  if (newStages && newStages.length > 0) {
    const setupStage = newStages.find(stage => stage.name === 'setup_environment')
    if (setupStage) {
      const currentStatus = setupStage.status
      // 只有当状态从 in_progress 变为 done 时才触发
      if (previousSetupStatus.value === 'in_progress' && currentStatus === 'done') {
        console.log('setup_environment 阶段已完成，通知刷新文件树')
        emit('projectEnvSetup')
      }
      previousSetupStatus.value = currentStatus
    }
  }
}, { deep: true })

// 检查项目是否已完成
const isProjectCompleted = () => {
  // 统一使用 project.value?.status 来判断项目是否完成或失败
  return props.project?.status === 'done' || props.project?.status === 'failed'
}


// 初始化
const initialize = async () => {
  console.log('🚀 [Init] 开始初始化 ConversationContainer...')
  
  // 1. 先加载初始数据（接口获取）
  console.log('📡 [Init] 步骤1: 从接口获取初始数据...')
  await loadDevStages()
  await loadConversations()
  console.log('✅ [Init] 步骤1完成: 接口数据已展示')
  
  // 检查项目是否已完成
  if (isProjectCompleted()) {
    console.log('ℹ️ [Init] 项目已完成，不启动 WebSocket 连接和定时刷新')
    return
  }
  
  // 2. 接口数据展示后，启动 WebSocket 连接
  console.log('🔗 [Init] 步骤2: 启动 WebSocket 连接...')
  try {
    await wsConnect()
    console.log('✅ [Init] 步骤2完成: WebSocket 连接成功')
  } catch (error) {
    console.error('❌ [Init] WebSocket 连接失败，将使用定时刷新:', error)
    // WebSocket 连接失败，启动定时刷新作为备用
    startAutoRefresh()
  }
  
  console.log('🎉 [Init] 初始化完成')
}

// 生命周期钩子
onMounted(() => {
  initialize()
})

// 确认界面相关方法
const getConfirmTitle = () => {
  switch (confirmStage.value) {
    case 'generate_prd':
      return 'PRD 文档生成完成'
    case 'define_ux_standard':
      return 'UX 标准定义完成'
    case 'design_architecture':
      return '系统架构设计完成'
    case 'plan_epic_and_story':
      return 'Epic 和 Story 划分完成'
    case 'define_data_model':
      return '数据模型定义完成'
    case 'define_api':
      return 'API 接口定义完成'
    case 'develop_story':
      return 'Story 开发完成'
    default:
      return '阶段完成，需要确认'
  }
}

const getConfirmDescription = () => {
  switch (confirmStage.value) {
    case 'generate_prd':
      return 'PRD 文档已生成，请确认内容是否符合要求'
    case 'define_ux_standard':
      return 'UX 标准已定义，请确认设计规范是否合适'
    case 'design_architecture':
      return '系统架构已设计，请确认架构方案是否合理'
    case 'plan_epic_and_story':
      return 'Epic 和 Story 已划分，请确认并编辑后继续'
    case 'define_data_model':
      return '数据模型已定义，请确认数据表结构是否正确'
    case 'define_api':
      return 'API 接口已定义，请确认接口设计是否完整'
    case 'develop_story':
      return 'Story 开发已完成，请确认功能实现是否满足需求'
    default:
      return '当前阶段已完成，请确认是否继续下一阶段'
  }
}

const handleSimpleConfirm = async () => {
  try {
    const success = await projectStore.confirmEpicsAndStories(props.projectGuid, 'confirm')
   
    if (success) {
      showConfirmInterface.value = false
      confirmStage.value = ''
      // 重新加载项目信息
      await loadDevStages()
      await loadConversations()
    } else {
      //message.error(response.data.message || '确认失败')
    }
  } catch (error: any) {
    console.error('确认失败:', error)
    message.error('确认失败: ' + (error.message || '未知错误'))
  }
}

const handleSimpleSkip = async () => {
  try {
    const success = await projectStore.confirmEpicsAndStories(props.projectGuid, 'skip')
    
    if (success) {
      showConfirmInterface.value = false
      confirmStage.value = ''
      // 重新加载项目信息
      await loadDevStages()
      await loadConversations()
    } else {
      //message.error(response.data.message || '跳过确认失败')
    }
  } catch (error: any) {
    console.error('跳过确认失败:', error)
    message.error('跳过确认失败: ' + (error.message || '未知错误'))
  }
}

const handleConfirmCompleted = () => {
  showConfirmInterface.value = false
  confirmStage.value = ''
  // 重新加载项目信息
  loadDevStages()
  loadConversations()
}

// 监听项目状态变化，显示确认界面
watch(() => props.project, (newProject) => {
  if (newProject && newProject.waiting_for_user_confirm && newProject.confirm_stage) {
    showConfirmInterface.value = true
    confirmStage.value = newProject.confirm_stage
    // 滚动到底部显示确认界面
    nextTick(() => {
      scrollToBottom()
    })
  } else {
    showConfirmInterface.value = false
    confirmStage.value = ''
  }
}, { immediate: true, deep: true })

onUnmounted(() => {
  stopAutoRefresh()
  wsDisconnect()
})
</script>

<style scoped>
.conversation-container {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: #f8fafc;
}

.ws-status-bar {
  flex-shrink: 0;
  padding: var(--spacing-sm) var(--spacing-lg);
  background: white;
  border-bottom: 1px solid #e2e8f0;
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
  overflow-x: hidden;
  padding: var(--spacing-lg);
  background: #f8fafc;
  width: 100%;
  box-sizing: border-box;
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

/* 滚动条样式 - 与开发进度区域保持一致 */
.conversation-messages::-webkit-scrollbar {
  width: 6px;
}

.conversation-messages::-webkit-scrollbar-track {
  background: transparent;
}

.conversation-messages::-webkit-scrollbar-thumb {
  background: #c1c1c1;
  border-radius: 2px;
}

.conversation-messages::-webkit-scrollbar-thumb:hover {
  background: #a8a8a8;
}

/* 确认界面样式 */
.confirm-interface {
  margin: 16px 0;
  
  .confirm-message {
    display: flex;
    gap: 12px;
    padding: 16px;
    background: var(--n-color);
    border: 1px solid var(--n-border-color);
    border-radius: 8px;
    
    .confirm-avatar {
      width: 40px;
      height: 40px;
      background: var(--n-primary-color);
      border-radius: 50%;
      display: flex;
      align-items: center;
      justify-content: center;
      flex-shrink: 0;
    }
    
    .confirm-content {
      flex: 1;
      
      .confirm-title {
        font-size: 16px;
        font-weight: 600;
        margin-bottom: 8px;
        color: var(--n-text-color);
      }
      
      .confirm-description {
        font-size: 14px;
        color: var(--n-text-color-2);
        margin-bottom: 16px;
        line-height: 1.5;
      }
      
      .epic-story-editor-wrapper {
        margin-top: 16px;
      }
      
      .simple-confirm {
        margin-top: 16px;
      }
    }
  }
}

/* 响应式设计 */
@media (max-width: 1024px) {
  .conversation-container {
    height: 100%;
  }
  
  .conversation-messages {
    padding: var(--spacing-sm);
  }
  
  .confirm-interface {
    .confirm-message {
      flex-direction: column;
      gap: 8px;
      
      .confirm-avatar {
        align-self: flex-start;
      }
    }
  }
}

@media (max-width: 768px) {
  .conversation-header {
    padding: 8px 12px;
  }
  
  .conversation-title {
    font-size: 1rem;
  }
  
  .conversation-messages {
    padding: 8px;
    gap: var(--spacing-sm);
  }
  
  .input-section {
    padding: 8px;
  }
  
  .loading-container {
    padding: 12px 8px;
  }
}

@media (max-width: 480px) {
  .conversation-header {
    padding: 6px 8px;
    min-height: 44px;
  }
  
  .conversation-title {
    font-size: 0.9rem;
  }
  
  .conversation-messages {
    padding: 6px;
    gap: 8px;
  }
  
  .input-section {
    padding: 6px;
  }
  
  .loading-container {
    padding: 8px 6px;
  }
  
  .loading-text {
    font-size: 0.85rem;
  }
}
</style>
