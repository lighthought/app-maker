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
        <!-- 编码选择器 -->
        <n-select
          v-if="fileContent"
          v-model:value="currentEncoding"
          :options="ENCODING_OPTIONS"
          size="tiny"
          style="width: 100px; margin-right: 8px;"
          @update:value="handleEncodingChange"
        />
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
      <!-- 加载状态层叠显示 -->
      <div v-if="isLoading" class="loading-overlay">
        <n-spin size="large">
          <template #description>
            <div class="loading-info">
              <p>{{ t('editor.loadingFile') }}</p>
              <p class="file-name">{{ filePath || 'unknown' }}</p>
            </div>
          </template>
        </n-spin>
      </div>
      
      <!-- 编辑器和状态显示 -->
      <div class="editor-state-container">
        <div v-if="fileContent && !loadError && !isLoading" ref="editorContainer" class="monaco-editor-container"></div>
        <div v-else-if="loadError && !isLoading" class="error-editor">
          <n-icon size="48" color="#f56565">
            <WarningIcon />
          </n-icon>
          <p>{{ t('editor.loadError') }}</p>
          <n-button size="small" @click="retryLoad">
            {{ t('common.retry') }}
          </n-button>
          <div v-if="failedFileContent" class="fallback-content">
            <p>{{ t('editor.rawContent') }}</p>
            <pre class="raw-text">{{ failedFileContent }}</pre>
          </div>
        </div>
        <div v-else-if="!isLoading" class="empty-editor">
          <n-icon size="48" color="#CBD5E0">
            <FileIcon />
          </n-icon>
          <p>{{ t('editor.selectFileToView') }}</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch, nextTick, h } from 'vue'
import { NIcon, NButton, NTooltip, NSelect, NSpin, useMessage } from 'naive-ui'
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

// 支持的编码格式
const ENCODING_OPTIONS = [
  { value: 'utf8', label: 'UTF-8' },
  { value: 'gbk', label: 'GBK' },
  { value: 'gb2312', label: 'GB2312' },
  { value: 'big5', label: 'Big5' },
  { value: 'utf16le', label: 'UTF-16 LE' },
  { value: 'utf16be', label: 'UTF-16 BE' },
  { value: 'latin1', label: 'Latin1' },
  { value: 'ascii', label: 'ASCII' }
]

const currentEncoding = ref('utf8')

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
const loadError = ref(false)
const failedFileContent = ref<string>('')

// 防抖处理
let loadTimeout: number | null = null

// 获取stores
const fileStore = useFilesStore()
const messageApi = useMessage()

// 获取文件内容
const loadFileContent = async (encoding: string = 'utf8', retry: boolean = false) => {
  if (!props.projectGuid || !props.filePath) {
    fileContent.value = ''
    loadError.value = false
    failedFileContent.value = ''
    return
  }
  
  // 取消之前的加载请求
  if (loadTimeout) {
    clearTimeout(loadTimeout)
  }
  
  // 防抖：300ms后开始加载
  loadTimeout = setTimeout(async () => {
    isLoading.value = true
    loadError.value = false
    failedFileContent.value = ''
    
    try {
      // 如果有编码需求，可以在这里处理
      const result = await fileStore.getFileContent(props.projectGuid!, props.filePath!)
      if (result) {
        let content = result.content
        // 如果内容包含乱码，尝试使用备用编码
        if (encoding !== 'utf8' && isGarbledText(content)) {
          try {
            // 这里可以实现编码转换逻辑
            // 暂时直接使用原内容
            content = result.content
            failedFileContent.value = content // 保存原始内容作为备用
          } catch (encodingError) {
            console.warn(t('editor.encodingConversionFailed'), encodingError)
          }
        }
        fileContent.value = content
        loadError.value = false
      } else {
        fileContent.value = ''
        loadError.value = !retry
      }
    } catch (error) {
      console.error(t('editor.loadingFileFailed'), error)
      fileContent.value = ''
      loadError.value = true
      
      // 如果是重试请求，再次尝试获取内容
      if (!retry) {
        try {
          const fallbackResult = await fileStore.getFileContent(props.projectGuid!, props.filePath!)
          if (fallbackResult && fallbackResult.content) {
            failedFileContent.value = fallbackResult.content
          }
        } catch (fallbackError) {
          console.error(t('editor.fallbackLoadFailed'), fallbackError)
        }
      }
    } finally {
      isLoading.value = false
    }
  }, retry ? 0 : 300) // 重试时不延迟
}

// 检测文本是否包含乱码
const isGarbledText = (text: string): boolean => {
  // 检测常见的乱码字符
  const garbledPatterns = [
    /�+/g, // 替换字符
    /[\uFFFD]+/g, // Unicode 替换字符
    /[^\x00-\x7F\u4E00-\u9FFF]/g // 包含非 ASCII 和非中文字符
  ]
  
  return garbledPatterns.some(pattern => pattern.test(text))
}

// 重试加载
const retryLoad = () => {
  loadFileContent(currentEncoding.value, true)
}

// 处理编码切换
const handleEncodingChange = (encoding: string) => {
  currentEncoding.value = encoding
  // 使用新编码重新加载文件
  loadFileContent(encoding, false)
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
  console.log('initEditor 开始初始化')
  
  if (!editorContainer.value || !editorContainer.value.parentNode) {
    console.warn('Monaco Editor container 不存在或没有父元素')
    return
  }

  try {
    // 确保清理已存在的编辑器
    if (editor) {
      console.log('清理已存在的编辑器实例')
      editor.dispose()
      editor = null
    }
    
    // 动态加载 Monaco Editor
    const monaco = await import('monaco-editor')
    
    // 确保容器可见且有尺寸
    const container = editorContainer.value
    
    // 清空容器内容
    container.innerHTML = ''
    
    if (container.clientWidth === 0 || container.clientHeight === 0) {
      console.warn('Monaco Editor container 尺寸为 0:', {
        width: container.clientWidth,
        height: container.clientHeight
      })
      // 给容器设置最小尺寸
      container.style.minHeight = '200px'
      container.style.minWidth = '100px'
    }
    
    console.log('容器状态:', {
      width: container.clientWidth,
      height: container.clientHeight,
      hasContent: container.children.length > 0
    })
    
    console.log('创建 Monaco Editor 实例')
    editor = monaco.editor.create(container, {
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

    // 如果有内容，立即设置
    if (fileContent.value) {
      editor.setValue(fileContent.value)
      console.log('Monaco Editor 初始化时设置内容:', fileContent.value.length, '字符')
    } else {
      console.log('Monaco Editor 初始化时内容为空，等待内容加载')
    }

    console.log('Monaco Editor 初始化成功，语言:', props.language, '文件路径:', props.filePath)

  } catch (error) {
    console.error('Monaco Editor 初始化失败:', error)
  }
}

// 更新编辑器内容
const updateEditor = async () => {
  console.log('updateEditor 调用:', { 
    hasContent: !!fileContent.value,
    contentLength: fileContent.value?.length || 0,
    hasEditor: !!editor,
    hasContainer: !!editorContainer.value,
    filePath: props.filePath
  })
  
  // 如果有内容但编辑器未初始化，先初始化
  if (fileContent.value && !editor && editorContainer.value) {
    console.log('初始化编辑器...')
    await nextTick() // 等待DOM更新
    if (editorContainer.value && editorContainer.value.parentNode) {
      await initEditor()
      // 初始化完成后，再次设置内容
      if (editor && fileContent.value) {
        console.log('编辑器初始化后设置内容:', fileContent.value.length, '字符')
        editor.setValue(fileContent.value)
        const language = getLanguage(props.filePath)
        editor.getModel()?.setLanguage(language)
      }
    }
    return
  }
  
  // 如果有编辑器且内容已加载，更新内容
  if (editor && fileContent.value) {
    try {
      const currentValue = editor.getValue()
      if (currentValue !== fileContent.value) {
        console.log('更新编辑器内容:', fileContent.value.length, '字符')
        editor.setValue(fileContent.value)
        // 设置语言
        const language = getLanguage(props.filePath)
        editor.getModel()?.setLanguage(language)
        console.log('编辑器内容更新完成')
      } else {
        console.log('编辑器内容无变化，跳过更新')
      }
    } catch (error) {
      console.error('更新Monaco编辑器内容失败:', error)
    }
  } else if (!editor && fileContent.value) {
    console.log('有内容但没有编辑器，尝试初始化...')
    // 如果没有编辑器但有内容，尝试初始化
    await nextTick()
    if (editorContainer.value && editorContainer.value.parentNode) {
      await initEditor()
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
watch(() => fileContent.value, (newContent, oldContent) => {
  console.log('文件内容变化:', { 
    newLength: newContent?.length || 0, 
    oldLength: oldContent?.length || 0,
    hasEditor: !!editor 
  })
  updateEditor()
})
watch(() => props.language, updateLanguage)

// 监听projectGuid和filePath变化
watch(() => [props.projectGuid, props.filePath],async (newValues, oldValues) => {
  const [newGuid, newPath] = newValues
  const [oldGuid, oldPath] = oldValues || [null, null]
  
  // 只有在真正变化时才处理
  if (newGuid !== oldGuid || newPath !== oldPath) {
    console.log('文件切换:', { oldPath, newPath, oldGuid, newGuid })
    
    // 清理旧的编辑器实例
    if (editor) {
      console.log('清理旧的Monaco Editor实例')
      editor.dispose()
      editor = null
    }
    
    // 重置状态
    loadError.value = false
    failedFileContent.value = ''
    currentEncoding.value = 'utf8' // 切换文件时重置编码
    fileContent.value = '' // 清空内容，强制重新渲染
    
    // 等待一个tick确保DOM更新和编辑器销毁
    await nextTick()
    
    // 重新加载文件内容
    if (props.projectGuid && props.filePath) {
      console.log('加载新文件内容:', props.filePath)
      await loadFileContent()
    }
  }
})

onMounted(async () => {
  await nextTick()
  // 如果有文件路径，加载文件内容
  if (props.projectGuid && props.filePath) {
    console.log('组件挂载时加载文件:', props.filePath)
    await loadFileContent()
    // 延迟初始化编辑器，确保DOM已经渲染
    setTimeout(() => {
      updateEditor()
    }, 50)
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

const FileIcon = () => h('svg', { 
  viewBox: '0 0 24 24', 
  fill: 'currentColor',
  style: 'width: 1em; height: 1em;'
}, [
  h('path', { d: 'M14,2H6A2,2 0 0,0 4,4V20A2,2 0 0,0 6,22H18A2,2 0 0,0 20,20V8L14,2M18,20H6V4H13V9H18V20Z' })
])

const WarningIcon = () => h('svg', {
  viewBox: '0 0 24 24',
  fill: 'currentColor',
  style: 'width: 1em; height: 1em;'
}, [
  h('path', { d: 'M1 21h22L12 2 1 21zm12-3h-2v-2h2v2zm0-4h-2v-4h2v4z' })
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
  position: relative;
}

.editor-state-container {
  width: 100%;
  height: 100%;
}

.loading-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(2px);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  z-index: 100;
}

.loading-info {
  text-align: center;
  color: var(--text-primary);
}

.loading-info p {
  margin: var(--spacing-xs) 0;
  font-size: 0.9rem;
}

.loading-info .file-name {
  font-family: 'Courier New', monospace;
  color: var(--text-secondary);
  font-size: 0.8rem !important;
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

.error-editor {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: #f56565;
  padding: var(--spacing-lg);
}

.error-editor p {
  margin: var(--spacing-md) 0 var(--spacing-sm) 0;
  font-size: 0.9rem;
}

.fallback-content {
  margin-top: var(--spacing-md);
  width: 100%;
  max-width: 100%;
}

.fallback-content p {
  color: var(--text-secondary);
  font-size: 0.8rem;
  margin-bottom: var(--spacing-sm);
}

.raw-text {
  background: #f7fafc;
  border: 1px solid #e2e8f0;
  border-radius: var(--border-radius-sm);
  padding: var(--spacing-md);
  font-family: 'Courier New', monospace;
  font-size: 0.8rem;
  white-space: pre-wrap;
  word-break: break-all;
  max-height: 300px;
  overflow-y: auto;
  color: #2d3748;
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
