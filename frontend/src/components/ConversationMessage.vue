<template>
  <div class="conversation-message" :class="messageClass">
    <!-- 用户消息 - 右侧显示 -->
    <div v-if="message.type === 'user'" class="user-message">
      <div class="message-content">
        <div class="message-text">{{ message.content }}</div>
        <div class="message-time">{{ formatTime(message.created_at) }}</div>
      </div>
      <div class="user-avatar">
        <n-icon size="20" color="white">
          <UserIcon />
        </n-icon>
      </div>
    </div>

    <!-- Agent/系统消息 - 左侧显示 -->
    <div v-else class="agent-message">
      <div class="agent-avatar" :class="agentAvatarClass">
        <n-icon size="20" color="white">
          <component :is="agentIcon" />
        </n-icon>
      </div>
      <div class="message-content">
        <div v-if="hasAgentInfo" class="agent-header">
          <span class="agent-name">{{ message.agent_name || getAgentName(message.agent_role) }}</span>
          <span class="agent-role">{{ getAgentRoleText(message.agent_role) }}</span>
        </div>
        
        <!-- 普通文本消息 -->
        <div v-if="!message.is_markdown" class="message-text">
          {{ message.content }}
        </div>
        
        <!-- Markdown消息 -->
        <div v-else class="markdown-message">
          <div class="markdown-header">
            <div class="content-preview">
              <span class="content-text" :title="message.content">{{ message.content }}</span>
            </div>
            <div class="action-buttons">
              <n-button
                text
                size="tiny"
                @click="toggleExpanded"
                class="expand-btn"
              >
                <template #icon>
                  <n-icon>
                    <ChevronDownIcon v-if="message.is_expanded" />
                    <ChevronUpIcon v-else />
                  </n-icon>
                </template>
              </n-button>
              <n-button
                text
                size="tiny"
                @click="copyMarkdown"
                class="copy-btn"
              >
                <template #icon>
                  <n-icon><CopyIcon /></n-icon>
                </template>
              </n-button>
            </div>
          </div>
          
          <div v-if="message.is_expanded" class="markdown-content">
            <div v-html="renderedMarkdown" class="markdown-body"></div>
          </div>
        </div>
        
        <div class="message-time">{{ formatTime(message.created_at) }}</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, h } from 'vue'
import { useI18n } from 'vue-i18n'
import { NIcon, NButton, useMessage } from 'naive-ui'
import { marked } from 'marked'
import type { ConversationMessage } from '@/types/project'
import { formatDateTime } from '@/utils/time'

interface Props {
  message: ConversationMessage
}

const props = defineProps<Props>()
const { t } = useI18n()

interface Emits {
  (e: 'toggle-expanded', messageId: string): void
}

const emit = defineEmits<Emits>()

// 获取message实例
const messageApi = useMessage()

// 消息样式类
const messageClass = computed(() => ({
  'message-user': props.message.type === 'user',
  'message-agent': props.message.type === 'agent',
  'message-system': props.message.type === 'system'
}))

// Agent头像样式类
const agentAvatarClass = computed(() => ({
  'avatar-dev': props.message.agent_role === 'dev',
  'avatar-pm': props.message.agent_role === 'pm',
  'avatar-po': props.message.agent_role === 'po',
  'avatar-architect': props.message.agent_role === 'architect',
  'avatar-ux-expert': props.message.agent_role === 'ux-expert',
  'avatar-analyst': props.message.agent_role === 'analyst',
  'avatar-qa': props.message.agent_role === 'qa',
  'avatar-ops': props.message.agent_role === 'ops'
}))

// 渲染的Markdown内容
const renderedMarkdown = computed(() => {
  if (!props.message.markdown_content) return ''
  return marked(props.message.markdown_content)
})

// 图标组件
const UserIcon = () => h('svg', { 
  viewBox: '0 0 24 24', 
  fill: 'currentColor',
  style: 'width: 1em; height: 1em;'
}, [
  h('path', { d: 'M12 12c2.21 0 4-1.79 4-4s-1.79-4-4-4-4 1.79-4 4 1.79 4 4 4zm0 2c-2.67 0-8 1.34-8 4v2h16v-2c0-2.66-5.33-4-8-4z' })
])

const DevIcon = () => h('svg', { 
  viewBox: '0 0 24 24', 
  fill: 'currentColor',
  style: 'width: 1em; height: 1em;'
}, [
  h('path', { d: 'M9.4 16.6L4.8 12l4.6-4.6L8 6l-6 6 6 6 1.4-1.4zm5.2 0L19.2 12l-4.6-4.6L16 6l6 6-6 6-1.4-1.4z' })
])

const PmIcon = () => h('svg', { 
  viewBox: '0 0 24 24', 
  fill: 'currentColor',
  style: 'width: 1em; height: 1em;'
}, [
  h('path', { d: 'M19 3H5c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h14c1.1 0 2-.9 2-2V5c0-1.1-.9-2-2-2zm-5 14H7v-2h7v2zm3-4H7v-2h10v2zm0-4H7V7h10v2z' })
])

const ArchIcon = () => h('svg', { 
  viewBox: '0 0 24 24', 
  fill: 'currentColor',
  style: 'width: 1em; height: 1em;'
}, [
  h('path', { d: 'M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z' })
])

const UxIcon = () => h('svg', { 
  viewBox: '0 0 24 24', 
  fill: 'currentColor',
  style: 'width: 1em; height: 1em;'
}, [
  h('path', { d: 'M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z' })
])

const QaIcon = () => h('svg', { 
  viewBox: '0 0 24 24', 
  fill: 'currentColor',
  style: 'width: 1em; height: 1em;'
}, [
  h('path', { d: 'M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41z' })
])

const OpsIcon = () => h('svg', { 
  viewBox: '0 0 24 24', 
  fill: 'currentColor',
  style: 'width: 1em; height: 1em;'
}, [
  h('path', { d: 'M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z' })
])

const InfoIcon = () => h('svg', { 
  viewBox: '0 0 24 24', 
  fill: 'currentColor',
  style: 'width: 1em; height: 1em;'
}, [
  h('path', { d: 'M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-6h2v6zm0-8h-2V7h2v2z' })
])

const ChevronUpIcon = () => h('svg', { 
  viewBox: '0 0 24 24', 
  fill: 'currentColor',
  style: 'width: 1em; height: 1em;'
}, [
  h('path', { d: 'M7.41 15.41L12 10.83l4.59 4.58L18 14l-6-6-6 6z' })
])

const ChevronDownIcon = () => h('svg', { 
  viewBox: '0 0 24 24', 
  fill: 'currentColor',
  style: 'width: 1em; height: 1em;'
}, [
  h('path', { d: 'M7.41 8.59L12 13.17l4.59-4.58L18 10l-6 6-6-6 1.41-1.41z' })
])

const CopyIcon = () => h('svg', { 
  viewBox: '0 0 24 24', 
  fill: 'currentColor',
  style: 'width: 1em; height: 1em;'
}, [
  h('path', { d: 'M16 1H4c-1.1 0-2 .9-2 2v14h2V3h12V1zm3 4H8c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h11c1.1 0 2-.9 2-2V7c0-1.1-.9-2-2-2zm0 16H8V7h11v14z' })
])

// Agent图标映射
const agentIcon = computed(() => {
  const iconMap = {
    dev: DevIcon,
    pm: PmIcon,
    po: PmIcon, // 产品负责人使用产品经理图标
    architect: ArchIcon,
    'ux-expert': UxIcon,
    analyst: ArchIcon, // 分析师使用架构师图标
    qa: QaIcon,
    ops: OpsIcon
  }
  return iconMap[props.message.agent_role as keyof typeof iconMap] || DevIcon
})

// 获取Agent名称
const getAgentName = (role?: string) => {
  const nameMap = {
    dev: 'James',
    pm: 'Alex',
    arch: 'Sam',
    ux: 'Emma',
    qa: 'Mike',
    ops: 'Lisa'
  }
  return nameMap[role as keyof typeof nameMap] || 'Agent'
}

// 获取Agent角色文本
const getAgentRoleText = (role?: string) => {
  const roleMap = {
    dev: t('agent.devEngineer'),
    pm: t('agent.productManager'),
    po: t('agent.productOwner'),
    architect: t('agent.architect'),
    'ux-expert': t('agent.uxExpert'),
    analyst: t('agent.analyst'),
    qa: t('agent.testEngineer'),
    ops: t('agent.opsEngineer')
  }
  return roleMap[role as keyof typeof roleMap] || t('agent.devEngineer')
}

// 判断是否有Agent信息
const hasAgentInfo = computed(() => {
  return !!(props.message.agent_name || props.message.agent_role)
})

// 格式化时间
const formatTime = (timestamp: string) => {
  return formatDateTime(timestamp)
}

// 切换展开状态
const toggleExpanded = () => {
  emit('toggle-expanded', props.message.id)
}

// 复制Markdown内容
const copyMarkdown = async () => {
  if (props.message.markdown_content) {
    try {
      await navigator.clipboard.writeText(props.message.markdown_content)
      // 显示复制成功提示
      messageApi.success(t('common.copySuccess'), {
        duration: 2000,
        closable: false
      })
    } catch (err) {
      console.error(t('common.copyFailed'), err)
      // 显示复制失败提示
      messageApi.error(t('common.copyRetry'), {
        duration: 2000,
        closable: false
      })
    }
  }
}
</script>

<style scoped>
.conversation-message {
  margin-bottom: var(--spacing-lg);
}

/* 用户消息样式 - 右侧显示 */
.user-message {
  display: flex;
  justify-content: flex-end;
  align-items: flex-end;
  gap: var(--spacing-sm);
  margin-left: 8%;
}

.user-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: #3b82f6;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.user-message .message-content {
  width: 100%;
  max-width: 100%;
  min-width: 100px;
  background: #3b82f6;
  color: white;
  padding: var(--spacing-md) var(--spacing-lg);
  border-radius: 18px;
  border-bottom-right-radius: 4px;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
  box-sizing: border-box;
  word-wrap: break-word;
  overflow-wrap: break-word;
}

/* Agent/系统消息样式 - 左侧显示 */
.agent-message {
  display: flex;
  align-items: flex-start;
  gap: var(--spacing-sm);
  margin-right: 8%;
}

.agent-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.avatar-dev { background: #3182CE; }
.avatar-pm { background: #38A169; }
.avatar-po { background: #38A169; }
.avatar-architect { background: #D69E2E; }
.avatar-ux-expert { background: #E53E3E; }
.avatar-analyst { background: #D69E2E; }
.avatar-qa { background: #805AD5; }
.avatar-ops { background: #DD6B20; }

.agent-message .message-content {
  width: 100%;
  max-width: 100%;
  min-width: 100px;
  background: white;
  border: 1px solid #e2e8f0;
  padding: var(--spacing-md) var(--spacing-lg);
  border-radius: 18px;
  border-bottom-left-radius: 4px;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
  box-sizing: border-box;
  word-wrap: break-word;
  overflow-wrap: break-word;
}

.agent-header {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  margin-bottom: var(--spacing-sm);
}

.agent-name {
  font-weight: bold;
  color: var(--primary-color);
}

.agent-role {
  font-size: 0.8rem;
  color: var(--text-secondary);
  background: var(--background-color);
  padding: 2px 6px;
  border-radius: var(--border-radius-sm);
}


/* 通用样式 */
.message-text {
  line-height: 1.5;
  word-wrap: break-word;
  overflow-wrap: break-word;
  max-width: 100%;
  box-sizing: border-box;
}

.message-time {
  font-size: 0.8rem;
  color: var(--text-disabled);
  margin-top: var(--spacing-xs);
  text-align: right;
}

.agent-message .message-time {
  text-align: left;
}

/* Markdown消息样式 */
.markdown-message {
  width: 100%;
}

.markdown-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-sm);
  padding: var(--spacing-sm);
  background: #f8fafc;
  border-radius: var(--border-radius-md);
  border: 1px solid #e2e8f0;
}

.content-preview {
  flex: 1;
  min-width: 0;
  margin-right: var(--spacing-md);
  overflow: hidden;
}

.content-text {
  font-size: 0.9rem;
  color: #374151;
  display: block;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 100%;
  word-break: break-all;
  box-sizing: border-box;
}

.action-buttons {
  display: flex;
  gap: var(--spacing-xs);
  flex-shrink: 0;
  min-width: fit-content;
}

.expand-btn,
.copy-btn {
  font-size: 0.8rem;
  color: var(--text-secondary);
}

.markdown-content {
  border: 1px solid var(--border-color);
  border-radius: var(--border-radius-md);
  padding: var(--spacing-md);
  background: var(--background-color);
  width: 100%;
  box-sizing: border-box;
  overflow-x: auto;
}

.markdown-body {
  line-height: 1.6;
}

.markdown-body :deep(h1),
.markdown-body :deep(h2),
.markdown-body :deep(h3),
.markdown-body :deep(h4),
.markdown-body :deep(h5),
.markdown-body :deep(h6) {
  margin-top: var(--spacing-lg);
  margin-bottom: var(--spacing-sm);
  color: var(--primary-color);
}

.markdown-body :deep(h1) { font-size: 1.5rem; }
.markdown-body :deep(h2) { font-size: 1.3rem; }
.markdown-body :deep(h3) { font-size: 1.1rem; }

.markdown-body :deep(p) {
  margin-bottom: var(--spacing-sm);
}

.markdown-body :deep(code) {
  background: var(--background-color);
  padding: 2px 4px;
  border-radius: var(--border-radius-sm);
  font-family: 'Courier New', monospace;
}

.markdown-body :deep(pre) {
  background: var(--background-color);
  padding: var(--spacing-md);
  border-radius: var(--border-radius-md);
  overflow-x: auto;
  margin: var(--spacing-sm) 0;
  max-width: 100%;
  box-sizing: border-box;
  word-wrap: break-word;
  overflow-wrap: break-word;
}

/* 统一滚动条样式 - 与开发进度区域保持一致 */
.markdown-body :deep(pre)::-webkit-scrollbar {
  height: 6px;
}

.markdown-body :deep(pre)::-webkit-scrollbar-track {
  background: transparent;
}

.markdown-body :deep(pre)::-webkit-scrollbar-thumb {
  background: #c1c1c1;
  border-radius: 2px;
}

.markdown-body :deep(pre)::-webkit-scrollbar-thumb:hover {
  background: #a8a8a8;
}

.markdown-body :deep(ul),
.markdown-body :deep(ol) {
  margin: var(--spacing-sm) 0;
  padding-left: var(--spacing-lg);
}

.markdown-body :deep(li) {
  margin-bottom: var(--spacing-xs);
}

.markdown-body :deep(blockquote) {
  border-left: 4px solid var(--primary-color);
  padding-left: var(--spacing-md);
  margin: var(--spacing-sm) 0;
  color: var(--text-secondary);
}
</style>
