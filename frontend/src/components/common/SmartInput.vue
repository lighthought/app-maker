<template>
  <div class="smart-input">
    <div class="input-container">
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
      <n-button
        :type="buttonType"
        :size="size"
        :disabled="!inputValue.trim()"
        @click="handleSend"
        class="send-button"
      >
        <template #icon>
          <n-icon><SendIcon /></n-icon>
        </template>
      </n-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, h } from 'vue'
import { NInputGroup, NInput, NButton, NIcon } from 'naive-ui'

interface Props {
  modelValue?: string
  placeholder?: string
  size?: 'small' | 'medium' | 'large'
  buttonType?: 'default' | 'primary' | 'info' | 'success' | 'warning' | 'error'
  autosize?: { minRows: number; maxRows: number }
}

interface Emits {
  (e: 'update:modelValue', value: string): void
  (e: 'send', value: string): void
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: '',
  placeholder: '请输入内容...',
  size: 'large',
  buttonType: 'default',
  autosize: () => ({ minRows: 3, maxRows: 6 })
})

const emit = defineEmits<Emits>()

// 主题覆盖配置
const inputThemeOverrides = {
  borderHover: 'none',
  borderFocus: 'none',
  border: '1px solid #E2E8F0',
  color: 'white',
  textColor: '#2D3748'
}

// 内部输入值
const inputValue = ref(props.modelValue)

// 监听外部值变化
watch(() => props.modelValue, (newVal) => {
  inputValue.value = newVal
})

// 监听内部值变化，同步到外部
watch(inputValue, (newVal) => {
  emit('update:modelValue', newVal)
})

// 发送图标组件
const SendIcon = () => h('svg', { 
  viewBox: '0 0 24 24', 
  fill: 'currentColor',
  style: 'width: 1em; height: 1em;'
}, [
  h('path', { d: 'M2.01 21L23 12 2.01 3 2 10l15 2-15 2z' })
])

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
  if (!inputValue.value.trim()) return
  
  emit('send', inputValue.value.trim())
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
  overflow: hidden;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  background: white;
  border-radius: var(--border-radius-lg);
}

.input-field {
  width: 100%;
  border: none;
  outline: none;
  background: transparent;
}

.input-field:focus {
  border-color: var(--primary-color);
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.1);
}

.send-button {
  position: absolute;
  bottom: 12px;
  right: 12px;
  width: 40px;
  height: 40px;
  background: #000000;
  
  color: white;
  transition: all 0.2s ease;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0;
  border-radius: 50%;
  z-index: 10;
}

.send-button:hover:not(:disabled) {
  background: #333333;
  border-color: #333333;
  transform: translateY(-1px);
}

.send-button:disabled {
  background: #e5e7eb;
  border-color: #e5e7eb;
  color: #9ca3af;
  cursor: not-allowed;
}

.send-button .n-icon {
  font-size: 20px;
  transform: rotate(-90deg);
}

/* 输入框样式优化 */
:deep(.n-input__textarea-el) {
  background: transparent;
  color: var(--text-primary);
  font-size: 20px;
  line-height: 1.5;
  padding: 16px 20px;
  padding-right: 60px;
  border: none;
  outline: none;
  text-align: left;
  resize: none;
}

:deep(.n-input__textarea-el::placeholder) {
  color: var(--text-secondary);
}

/* 确保主题覆盖生效 */
:deep(.n-input) {
  background: transparent;
  border: none;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .send-button {
    width: 36px;
    height: 36px;
    bottom: 10px;
    right: 10px;
  }
  
  :deep(.n-input__textarea-el) {
    padding: 14px 16px;
    padding-right: 50px;
    font-size: 18px;
  }
}
</style>
