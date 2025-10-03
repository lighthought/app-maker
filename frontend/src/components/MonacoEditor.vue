<template>
  <div class="monaco-editor-wrapper">
    <!-- 编辑器头部 -->
    <div class="editor-header">
      <div class="file-info">
        <n-tooltip :disabled="!filePath || filePath.length <= 50">
          <template #trigger>
            <span class="file-path">{{ filePath || t('editor.selectFileToView') }}</span>
          </template>
          {{ filePath }}
        </n-tooltip>
      </div>
      <div class="editor-actions">
        <n-button text size="tiny" @click="copyCode">
          <template #icon>
            <n-icon><CopyIcon /></n-icon>
          </template>
          {{ t('common.copy') }}
        </n-button>
      </div>
    </div>
    
    <!-- 编辑器内容 -->
    <div class="editor-content">
      <div v-if="fileContent" ref="editorContainer" class="monaco-editor-container"></div>
      <div v-else-if="isLoading" class="loading-editor">
        <n-icon size="48" color="#CBD5E0">
          <LoadingIcon />
        </n-icon>
        <p>{{ t('editor.loadingFile') }}</p>
      </div>
      <div v-else class="empty-editor">
        <n-icon size="48" color="#CBD5E0">
          <FileIcon />
        </n-icon>
        <p>{{ t('editor.selectFileToView') }}</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch, nextTick, h } from 'vue'
import { NIcon, NButton, NTooltip, useMessage } from 'naive-ui'
import { useFilesStore } from '@/stores/file'
import { useI18n } from 'vue-i18n'
const { t } = useI18n()
interface Props {
  projectGuid?: string
  filePath?: string
  language?: string
  readOnly?: boolean
  theme?: 'vs' | 'vs-dark' | 'hc-black'
  height?: string
}

interface Emits {
  (e: 'update:value', value: string): void
  (e: 'change', value: string): void
}

const props = withDefaults(defineProps<Props>(), {
  language: 'javascript',
  readOnly: true,
  theme: 'vs',
  height: '100%'
})

const emit = defineEmits<Emits>()

const editorContainer = ref<HTMLElement>()
let editor: any = null

// 文件内容状态
const fileContent = ref<string>('')
const isLoading = ref(false)

// 获取stores
const fileStore = useFilesStore()
const messageApi = useMessage()

// 获取文件内容
const loadFileContent = async () => {
  if (!props.projectGuid || !props.filePath) {
    fileContent.value = ''
    return
  }
  
  isLoading.value = true
  try {
    const result = await fileStore.getFileContent(props.projectGuid, props.filePath)
    if (result) {
      fileContent.value = result.content
    } else {
      fileContent.value = ''
    }
  } catch (error) {
    console.error(t('editor.loadingFileFailed'), error)
    fileContent.value = ''
  } finally {
    isLoading.value = false
  }
}

// 复制代码
const copyCode = async () => {
  if (fileContent.value) {
    try {
      await navigator.clipboard.writeText(fileContent.value)
      messageApi.success(t('common.copySuccess'), {
        duration: 2000,
        closable: false
      })
    } catch (err) {
      console.error(t('common.copyFailed'), err)
      messageApi.error(t('common.copyRetry'), {
        duration: 2000,
        closable: false
      })
    }
  }
}

// 获取语言类型
const getLanguage = (filePath?: string): string => {
  if (!filePath) return 'plaintext'
  
  const extension = filePath.split('.').pop()?.toLowerCase()
  const languageMap: Record<string, string> = {
    'js': 'javascript',
    'jsx': 'javascript',
    'ts': 'typescript',
    'tsx': 'typescript',
    'vue': 'html',
    'html': 'html',
    'css': 'css',
    'scss': 'scss',
    'sass': 'sass',
    'less': 'less',
    'json': 'json',
    'xml': 'xml',
    'yaml': 'yaml',
    'yml': 'yaml',
    'md': 'markdown',
    'py': 'python',
    'java': 'java',
    'go': 'go',
    'rs': 'rust',
    'php': 'php',
    'rb': 'ruby',
    'sh': 'shell',
    'bash': 'shell',
    'sql': 'sql',
    'dockerfile': 'dockerfile',
    'gitignore': 'plaintext',
    'env': 'plaintext'
  }
  
  return languageMap[extension || ''] || 'plaintext'
}

// 初始化编辑器
const initEditor = async () => {
  if (!editorContainer.value) return

  try {
    // 动态加载 Monaco Editor
    const monaco = await import('monaco-editor')
    
    editor = monaco.editor.create(editorContainer.value, {
      value: fileContent.value,
      language: props.language,
      readOnly: props.readOnly,
      theme: props.theme,
      automaticLayout: true,
      minimap: { enabled: false },
      scrollBeyondLastLine: false,
      wordWrap: 'off', // 关闭自动换行，允许横向滚动
      lineNumbers: 'on',
      folding: true,
      lineDecorationsWidth: 0,
      lineNumbersMinChars: 3,
      renderLineHighlight: 'line',
      fontSize: 14,
      fontFamily: 'Consolas, "Courier New", monospace',
      tabSize: 2,
      insertSpaces: true,
      detectIndentation: false,
      cursorBlinking: 'blink',
      cursorSmoothCaretAnimation: "on",
      smoothScrolling: true,
      mouseWheelZoom: true,
      contextmenu: !props.readOnly,
      selectOnLineNumbers: true,
      roundedSelection: false,
      occurrencesHighlight: "singleFile",
      selectionHighlight: false,
      codeLens: false,
      foldingStrategy: 'indentation',
      showFoldingControls: 'always',
      bracketPairColorization: {
        enabled: true
      },
      guides: {
        bracketPairs: true,
        indentation: true
      },
      // 禁用需要 Web Worker 的功能
      quickSuggestions: false,
      suggestOnTriggerCharacters: false,
      acceptSuggestionOnEnter: 'off',
      tabCompletion: 'off',
      wordBasedSuggestions: 'off',
      parameterHints: { enabled: false },
      hover: { enabled: false },
      formatOnPaste: false,
      formatOnType: false
    })

    // 监听内容变化
    editor.onDidChangeModelContent(() => {
      const value = editor.getValue()
      emit('update:value', value)
      emit('change', value)
    })

    // 设置容器高度
    if (props.height !== '100%') {
      editorContainer.value.style.height = props.height
    }

  } catch (error) {
    console.error('Monaco Editor 初始化失败:', error)
  }
}

// 更新编辑器内容
const updateEditor = async () => {
  if (fileContent.value) {
    // 如果有内容但编辑器未初始化，先初始化
    if (!editor) {
      await initEditor()
      return
    }
    // 如果编辑器已存在，更新内容
    if (editor.getValue() !== fileContent.value) {
      editor.setValue(fileContent.value)
    }
  }
}

// 更新语言
const updateLanguage = () => {
  if (editor && editor.getModel()) {
    editor.getModel().setLanguage(props.language)
  }
}

// 监听属性变化
watch(() => fileContent.value, updateEditor)
watch(() => props.language, updateLanguage)
watch(() => [props.projectGuid, props.filePath], async () => {
  // 当项目GUID或文件路径变化时，重新加载文件内容
  await loadFileContent()
})

onMounted(async () => {
  await nextTick()
  // 如果有文件路径，加载文件内容
  if (props.projectGuid && props.filePath) {
    await loadFileContent()
  }
})

onUnmounted(() => {
  if (editor) {
    editor.dispose()
  }
})

// 图标组件
const CopyIcon = () => h('svg', { 
  viewBox: '0 0 24 24', 
  fill: 'currentColor',
  style: 'width: 1em; height: 1em;'
}, [
  h('path', { d: 'M16 1H4c-1.1 0-2 .9-2 2v14h2V3h12V1zm3 4H8c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h11c1.1 0 2-.9 2-2V7c0-1.1-.9-2-2-2zm0 16H8V7h11v14z' })
])

const FileIcon = () => h('svg', { 
  viewBox: '0 0 24 24', 
  fill: 'currentColor',
  style: 'width: 1em; height: 1em;'
}, [
  h('path', { d: 'M14,2H6A2,2 0 0,0 4,4V20A2,2 0 0,0 6,22H18A2,2 0 0,0 20,20V8L14,2M18,20H6V4H13V9H18V20Z' })
])

const LoadingIcon = () => h('svg', { 
  viewBox: '0 0 24 24', 
  fill: 'currentColor',
  style: 'width: 1em; height: 1em; animation: spin 1s linear infinite;'
}, [
  h('path', { d: 'M12 2A10 10 0 0 0 2 12a10 10 0 0 0 10 10 10 10 0 0 0 10-10A10 10 0 0 0 12 2zm0 18a8 8 0 0 1-8-8 8 8 0 0 1 8-8 8 8 0 0 1 8 8 8 8 0 0 1-8 8z' }),
  h('path', { 
    d: 'M12 4a8 8 0 0 1 8 8 8 8 0 0 1-8 8',
    style: 'opacity: 0.3;'
  })
])
</script>

<style scoped>
.monaco-editor-wrapper {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: white;
  border-radius: var(--border-radius-md);
  overflow: hidden;
  min-width: 0; /* 允许收缩 */
  width: 100%; /* 确保不超出父容器 */
}

.editor-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--spacing-md) var(--spacing-lg);
  border-bottom: 1px solid var(--border-color);
  background: var(--background-color);
  height: var(--height-md);
  min-width: 0;
  flex-shrink: 0;
}

.file-info {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  flex: 1;
  min-width: 0;
  margin-right: var(--spacing-md);
}

.file-path {
  font-size: 0.9rem;
  color: var(--text-secondary);
  font-family: 'Courier New', monospace;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex: 1;
}

.editor-actions {
  display: flex;
  gap: var(--spacing-sm);
  flex-shrink: 0;
}

.editor-content {
  flex: 1;
  overflow: auto;
  background: #f8f9fa;
  border-radius: var(--border-radius-md);
}

.monaco-editor-container {
  width: 100%;
  height: 100%;
  border-radius: var(--border-radius-md);
  overflow: hidden;
  min-width: 0; /* 允许收缩 */
}

.empty-editor {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: var(--text-secondary);
}

.empty-editor p {
  margin: var(--spacing-md) 0 0 0;
  font-size: 0.9rem;
}

.loading-editor {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: var(--text-secondary);
}

.loading-editor p {
  margin: var(--spacing-md) 0 0 0;
  font-size: 0.9rem;
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

/* Monaco Editor 样式覆盖 */
:deep(.monaco-editor) {
  border-radius: var(--border-radius-md);
}

:deep(.monaco-editor .margin) {
  background-color: #f8f9fa;
}

:deep(.monaco-editor .monaco-editor-background) {
  background-color: #ffffff;
}

/* 暗色主题 */
:deep(.monaco-editor.vs-dark .margin) {
  background-color: #1e1e1e;
}

:deep(.monaco-editor.vs-dark .monaco-editor-background) {
  background-color: #1e1e1e;
}

/* 滚动条样式 */
.editor-content::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}

.editor-content::-webkit-scrollbar-track {
  background: transparent;
}

.editor-content::-webkit-scrollbar-thumb {
  background: #c1c1c1;
  border-radius: 3px;
}

.editor-content::-webkit-scrollbar-thumb:hover {
  background: #a8a8a8;
}

/* Monaco Editor 内部滚动条样式覆盖 */
:deep(.monaco-editor .monaco-scrollable-element > .scrollbar) {
  background: transparent !important;
}

:deep(.monaco-editor .monaco-scrollable-element > .scrollbar > .slider) {
  background: #c1c1c1 !important;
  border-radius: 3px !important;
}

:deep(.monaco-editor .monaco-scrollable-element > .scrollbar > .slider:hover) {
  background: #a8a8a8 !important;
}

:deep(.monaco-editor .monaco-scrollable-element > .scrollbar > .slider.active) {
  background: #a8a8a8 !important;
}

/* 横向滚动条 */
:deep(.monaco-editor .monaco-scrollable-element > .scrollbar.horizontal) {
  height: 6px !important;
}

:deep(.monaco-editor .monaco-scrollable-element > .scrollbar.horizontal > .slider) {
  height: 6px !important;
}

/* 纵向滚动条 */
:deep(.monaco-editor .monaco-scrollable-element > .scrollbar.vertical) {
  width: 6px !important;
}

:deep(.monaco-editor .monaco-scrollable-element > .scrollbar.vertical > .slider) {
  width: 6px !important;
}
</style>
