<template>
  <div class="smart-input">
    <div class="input-container">
      <!-- 输入框区域 -->
      <n-input
        v-model:value="inputValue"
        :placeholder="placeholder"
        :size="size"
        type="textarea"
        :autosize="autosize"
        @keydown.enter="handleEnterKey"
        class="input-field"
        :theme-overrides="inputThemeOverrides"
      />
      
      <!-- 底部工具栏 -->
      <div class="toolbar">
        <!-- Agent 选择器 - 左下角 -->
        <div v-if="showAgentSelector" class="agent-selector-wrapper">
          <n-select 
            v-model:value="currentAgent"
            :options="agentOptions"
            :disabled="agentLocked"
            class="agent-selector"
            size="small"
            :placeholder="t('common.selectAgent')"
            :theme-overrides="selectThemeOverrides"
            :consistent-menu-width="false"
            @update:show="handleSelectShow"
          >
          </n-select>
          <div v-if="agentLocked" class="lock-indicator" :title="t('common.agentLocked')">
            <n-icon :component="LockIcon" size="14" color="#f59e0b" />
          </div>
        </div>
        
        <!-- 占位符 - 保持布局平衡 -->
        <div v-else class="toolbar-spacer"></div>
        
        <!-- 发送按钮 - 右下角 -->
        <n-button
          :type="buttonType"
          :size="size"
          :disabled="isSendDisabled"
          @click="handleSend"
          class="send-button"
        >
          <template #icon>
            <n-icon><SendIcon /></n-icon>
          </template>
        </n-button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { NInput, NButton, NIcon, NSelect } from 'naive-ui'
import { useI18n } from 'vue-i18n'
// 导入图标
import { SendIcon, ChevronDownIcon, LockIcon } from '@/components/icon'

interface Props {
  modelValue?: string
  placeholder?: string
  size?: 'small' | 'medium' | 'large'
  buttonType?: 'default' | 'primary' | 'info' | 'success' | 'warning' | 'error'
  autosize?: { minRows: number; maxRows: number }
  agentOptions?: Array<{ label: string; value: string }>
  selectedAgent?: string
  agentLocked?: boolean
}

interface Emits {
  (e: 'update:modelValue', value: string): void
  (e: 'update:selectedAgent', value: string): void
  (e: 'send', value: string, agentType: string): void
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: '',
  placeholder: '请输入内容...',
  size: 'large',
  buttonType: 'default',
  autosize: () => ({ minRows: 3, maxRows: 6 }),
  agentOptions: () => [],
  selectedAgent: '',
  agentLocked: false
})

const emit = defineEmits<Emits>()
const { t } = useI18n()

// 主题覆盖配置
const inputThemeOverrides = {
  borderHover: 'none',
  borderFocus: 'none',
  border: '1px solid #E2E8F0',
  color: 'white',
  textColor: '#2D3748'
}

const selectThemeOverrides = {
  peers: {
    InternalSelection: {
      border: '1px solid #E2E8F0',
      borderHover: '1px solid #CBD5E1',
      borderActive: '1px solid var(--primary-color)',
      borderFocus: '1px solid var(--primary-color)',
      color: '#F9FAFB',
      textColor: '#2D3748'
    }
  }
}

// 内部输入值
const inputValue = ref(props.modelValue)
const currentAgent = ref(props.selectedAgent)
const isSelectOpen = ref(false)

// 计算属性
const showAgentSelector = computed(() => {
  return props.agentOptions && props.agentOptions.length > 0
})

// 处理下拉框展开/收起
const handleSelectShow = (show: boolean) => {
  isSelectOpen.value = show
}

const isSendDisabled = computed(() => {
  if (!inputValue.value.trim()) return true
  // 只有在显示 Agent 选择器时才检查 Agent 是否选中
  if (showAgentSelector.value && !currentAgent.value) return true
  return false
})

// 监听外部值变化
watch(() => props.modelValue, (newVal) => {
  inputValue.value = newVal
})

watch(() => props.selectedAgent, (newVal) => {
  currentAgent.value = newVal
})

// 监听内部值变化，同步到外部
watch(inputValue, (newVal) => {
  emit('update:modelValue', newVal)
})

watch(currentAgent, (newVal) => {
  emit('update:selectedAgent', newVal)
})

// 键盘事件处理
const handleEnterKey = (event: KeyboardEvent) => {
  if (event.shiftKey) {
    // Shift + Enter: 换行
    return
  } else {
    // Enter: 发送
    event.preventDefault()
    handleSend()
  }
}

// 发送处理
const handleSend = () => {
  if (isSendDisabled.value) return
  
  // 如果有 Agent 选择器，发送时传递 agentType；否则传递空字符串
  const agentType = showAgentSelector.value ? currentAgent.value : ''
  emit('send', inputValue.value.trim(), agentType)
  // 发送后清空输入框
  inputValue.value = ''
}
</script>

<style scoped>
.smart-input {
  width: 100%;
}

.input-container {
  position: relative;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  background: white;
  border-radius: 12px;
  border: 1px solid #e5e7eb;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  padding: 16px;
}

.input-container:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.12);
  border-color: #cbd5e1;
}

.input-container:focus-within {
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.15);
  border-color: var(--primary-color);
}

/* 输入框样式 */
.input-field {
  width: 100%;
  border: none;
  outline: none;
  background: transparent;
  margin-bottom: 12px;
}

/* 底部工具栏 */
.toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding-top: 8px;
}

/* Agent 选择器区域 - 左下角 */
.agent-selector-wrapper {
  display: flex;
  align-items: center;
  gap: 6px;
  flex: 0 0 auto;
}

.agent-selector {
  width: 140px;
}

.lock-indicator {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  animation: pulse-lock 2s infinite;
}

@keyframes pulse-lock {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
}

.toolbar-spacer {
  flex: 1;
}

/* 发送按钮 - 右下角 */
.send-button {
  flex-shrink: 0;
  width: 40px;
  height: 40px;
  background: #000;
  color: white;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0;
  border-radius: 8px;
  border: none;
  cursor: pointer;
}

.send-button:hover:not(:disabled) {
  background: #333;
  transform: translateY(-1px);
}

.send-button:active:not(:disabled) {
  transform: translateY(0);
}

.send-button:disabled {
  background: #e5e7eb;
  color: #9ca3af;
  cursor: not-allowed;
}

.send-button .n-icon {
  font-size: 20px;
  transform: rotate(-90deg);
}

/* 输入框内部样式优化 */
:deep(.n-input__textarea-el) {
  background: transparent;
  color: var(--text-primary);
  font-size: 15px;
  line-height: 1.5;
  padding: 0;
  border: none;
  outline: none;
  text-align: left;
  resize: none;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
}

:deep(.n-input__textarea-el::placeholder) {
  color: #9ca3af;
}

/* 选择器样式优化 - 扁平化设计 */
:deep(.n-select) {
  background: transparent;
}

:deep(.n-base-selection) {
  background: white;
  border: 1px solid #e5e7eb;
  border-radius: 4px;
  transition: all 0.15s ease;
  min-height: 30px;
  padding: 0 8px;
  box-shadow: none;
}

:deep(.n-base-selection:hover) {
  border-color: #d1d5db;
  background: white;
}

:deep(.n-base-selection.n-base-selection--active),
:deep(.n-base-selection.n-base-selection--focus) {
  border-color: #cbd5e1;
  background: white;
  box-shadow: none;
}

:deep(.n-base-selection.n-base-selection--disabled) {
  background: #f3f4f6;
  border-color: #e5e7eb;
  opacity: 0.5;
}

:deep(.n-base-selection-label) {
  font-weight: 400;
  font-size: 12px;
  color: #374151;
}

:deep(.n-base-selection-placeholder) {
  font-size: 12px;
  color: #9ca3af;
}

/* 选择器箭头 */
:deep(.n-base-selection .n-base-suffix) {
  color: #6b7280;
  transition: transform 0.2s ease;
}

:deep(.n-base-selection .n-base-suffix__arrow) {
  transition: transform 0.2s ease;
}

/* 选择器展开时箭头旋转 */
:deep(.n-base-selection.n-base-selection--active .n-base-suffix__arrow) {
  transform: rotate(180deg);
}

/* 选择器内边距优化 */
:deep(.n-base-selection .n-base-selection-label) {
  padding: 0;
}

/* 确保主题覆盖生效 */
:deep(.n-input) {
  background: transparent;
  border: none;
}

:deep(.n-input__border),
:deep(.n-input__state-border) {
  display: none;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .input-container {
    border-radius: 10px;
    padding: 12px;
  }
  
  .toolbar {
    gap: 8px;
    padding-top: 6px;
  }
  
  .agent-selector {
    width: 120px;
  }
  
  .send-button {
    width: 36px;
    height: 36px;
  }
  
  :deep(.n-input__textarea-el) {
    font-size: 14px;
  }
  
  :deep(.n-base-selection) {
    min-height: 28px;
  }
  
  :deep(.n-base-selection-label) {
    font-size: 12px;
  }
}
</style>
