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
          style="width: 120px; margin-right: 8px;"
          @update:value="handleEncodingChange"
        />
        <!-- Markdown预览切换按钮 -->
        <n-button 
          v-if="isMarkdownFile" 
          text 
          size="tiny" 
          @click="togglePreview"
          :type="previewMode ? 'primary' : 'default'"
        >
          <template #icon>
            <n-icon><EyeIcon v-if="!previewMode" /><EditIcon v-else /></n-icon>
          </template>
        </n-button>
        <n-button text size="tiny" @click="copyCode">
          <template #icon>
            <n-icon><CopyIcon /></n-icon>
          </template>
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
        <!-- Markdown预览模式 -->
        <div v-if="fileContent && !loadError && !isLoading && isMarkdownFile && previewMode" 
             class="markdown-preview">
          <div class="markdown-content" v-html="renderedMarkdown"></div>
        </div>
        <!-- 普通编辑器模式 -->
        <div v-else-if="fileContent && !loadError && !isLoading && (!isMarkdownFile || !previewMode)" 
             ref="editorContainer" class="monaco-editor-container"></div>
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
import { ref, onMounted, onUnmounted, watch, nextTick, computed } from 'vue'
import { NIcon, NButton, NTooltip, NSelect, NSpin, useMessage } from 'naive-ui'
import { useFilesStore } from '@/stores/file'
import { useI18n } from 'vue-i18n'
import { marked } from 'marked'
import mermaid from 'mermaid'
// 导入图标
import { CopyIcon, LoadingIcon, FileIcon, WarningIcon, EyeIcon, EditIcon } from '@/components/icon'
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
  { value: 'utf-8', label: 'UTF-8' },
  { value: 'gbk', label: 'GBK' },
  { value: 'gb18030', label: 'GB18030' },
  { value: 'ascii', label: 'ASCII' }
]

const currentEncoding = ref('utf-8')

// Markdown预览相关状态
const previewMode = ref(true)

// 是否Markdown文件的判断
const isMarkdownFile = computed(() => {
  if (!props.filePath) return false
  const extension = props.filePath.split('.').pop()?.toLowerCase()
  return extension === 'md' || extension === 'markdown'
})

// 初始化Mermaid
const initMermaid = () => {
  mermaid.initialize({
    startOnLoad: false,
    theme: 'default',
    securityLevel: 'loose',
    fontFamily: 'Arial, sans-serif',
    fontSize: 14,
    flowchart: {
      useMaxWidth: true,
      htmlLabels: true,
      curve: 'basis'
    },
    sequence: {
      useMaxWidth: true,
      diagramMarginX: 50,
      diagramMarginY: 10
    },
    gantt: {
      useMaxWidth: true
    }
  })
  console.log('Mermaid初始化完成')
}

// 渲染的Markdown内容
const renderedMarkdown = computed(() => {
  if (!fileContent.value || !isMarkdownFile.value) return ''
  
  try {    
    // 配置marked选项
    marked.setOptions({
      breaks: true,
      gfm: true
    })
    
    let html = marked.parse(fileContent.value) as string
    console.log('Marked解析后的HTML长度:', html.length)
    
    // 处理Mermaid图表 - 匹配Marked解析后的HTML格式
    const beforeReplace = html
    html = html.replace(/<pre><code class="language-mermaid">([\s\S]*?)<\/code><\/pre>/g, (match: string, diagram: string) => {
      const id = 'mermaid-' + Math.random().toString(36).substr(2, 9)
      console.log('发现Mermaid图表:', id, diagram.substring(0, 100) + '...')
      return `<div class="mermaid-diagram" id="${id}">${diagram.trim()}</div>`
    })
    
    // 也尝试匹配没有class的情况
    html = html.replace(/<pre><code>([\s\S]*?)<\/code><\/pre>/g, (match: string, diagram: string) => {
      // 检查是否是Mermaid内容（简单检查是否包含graph、flowchart等关键字）
      if (diagram.includes('graph') || diagram.includes('flowchart') || diagram.includes('sequenceDiagram')) {
        const id = 'mermaid-' + Math.random().toString(36).substr(2, 9)
        console.log('发现Mermaid图表(无class):', id, diagram.substring(0, 100) + '...')
        return `<div class="mermaid-diagram" id="${id}">${diagram.trim()}</div>`
      }
      return match // 不是Mermaid内容，保持原样
    })
    
    if (html !== beforeReplace) {
      console.log('Mermaid图表已替换到HTML中')
    } else {
      console.log('没有找到Mermaid图表进行替换')
    }
    
    return html
  } catch (error) {
    console.error('Markdown渲染失败:', error)
    return `<pre>${fileContent.value}</pre>`
  }
})

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
let loadTimeout: ReturnType<typeof setTimeout> | null = null

// 获取stores
const fileStore = useFilesStore()
const messageApi = useMessage()

// 渲染Mermaid图表
const renderMermaidDiagrams = async () => {
  console.log('开始渲染Mermaid图表...')
  
  // 等待DOM更新
  await nextTick()
  
  // 再次等待，确保DOM完全渲染
  await new Promise(resolve => setTimeout(resolve, 100))
  
  const diagrams = document.querySelectorAll('.mermaid-diagram')
  console.log('找到Mermaid图表数量:', diagrams.length)
  
  for (const diagram of diagrams) {
    try {
      const id = diagram.id
      const content = diagram.textContent || ''
      console.log('渲染图表:', id, '内容长度:', content.length)
      
      if (content.trim()) {
        // 使用更安全的渲染方法
        const { svg } = await mermaid.render(id + '-svg', content)
        diagram.innerHTML = svg
        console.log('图表渲染成功:', id)
      } else {
        console.warn('图表内容为空:', id)
      }
    } catch (error) {
      console.error('Mermaid图表渲染失败:', error)
      diagram.innerHTML = `<div class="mermaid-error">图表渲染失败: ${error}</div>`
    }
  }
}

// 切换Markdown预览模式
const togglePreview = async () => {
  previewMode.value = !previewMode.value
  console.log('切换Markdown预览模式:', previewMode.value)
  
  // 如果切换到预览模式，渲染Mermaid图表
  if (previewMode.value && isMarkdownFile.value) {
    // 等待DOM更新后再渲染
    await nextTick()
    setTimeout(async () => {
      await renderMermaidDiagrams()
    }, 200)
  }
  
  // 如果切换到编辑模式且没有编辑器，需要初始化
  if (!previewMode.value && !editor && fileContent.value) {
    console.log('切换到编辑模式，初始化编辑器...')
    await nextTick()
    if (editorContainer.value && editorContainer.value.parentNode) {
      await initEditor()
      if (editor && fileContent.value) {
        editor.setValue(fileContent.value)
        const language = getLanguage(props.filePath)
        editor.getModel()?.setLanguage(language)
      }
    }
  }
}

// 获取文件内容
const loadFileContent = async (encoding: string = 'utf-8', retry: boolean = false) => {
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
      const result = await fileStore.getFileContent(props.projectGuid!, props.filePath!, currentEncoding.value)
      if (result) {        
        fileContent.value = result.content
        loadError.value = false
      } else {
        fileContent.value = ''
        loadError.value = !retry
      }
    } catch (error) {
      console.error(t('editor.loadingFileFailed'), error)
      fileContent.value = ''
      loadError.value = true
    } finally {
      isLoading.value = false
    }
  }, retry ? 0 : 300) // 重试时不延迟
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
    filePath: props.filePath,
    previewMode: previewMode.value,
    isMarkdownFile: isMarkdownFile.value
  })
  
  // 如果是Markdown文件且处于预览模式，跳过编辑器更新
  if (isMarkdownFile.value && previewMode.value) {
    console.log('Markdown预览模式，跳过编辑器更新')
    return
  }
  
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
  } else if (!editor && fileContent.value && (!isMarkdownFile.value || !previewMode.value)) {
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
watch(() => fileContent.value, async (newContent, oldContent) => {
  console.log('文件内容变化:', { 
    newLength: newContent?.length || 0, 
    oldLength: oldContent?.length || 0,
    hasEditor: !!editor 
  })
  updateEditor()
  
  // 如果是Markdown文件且处于预览模式，重新渲染Mermaid图表
  if (isMarkdownFile.value && previewMode.value && newContent) {
    // 延迟渲染，确保DOM更新完成
    setTimeout(async () => {
      await renderMermaidDiagrams()
    }, 300)
  }
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
    currentEncoding.value = 'utf-8' // 切换文件时重置编码
    fileContent.value = '' // 清空内容，强制重新渲染
    
    // 等待一个tick确保DOM更新和编辑器销毁
    await nextTick()
    
    // 检测新文件的类型并设置预览模式
    const newFileExt = props.filePath?.split('.').pop()?.toLowerCase()
    const isNewMarkdownFile = newFileExt === 'md' || newFileExt === 'markdown'
    previewMode.value = isNewMarkdownFile
    
    // 重新加载文件内容
    if (props.projectGuid && props.filePath) {
      console.log('加载新文件内容:', props.filePath, '预览模式:', previewMode.value)
      await loadFileContent()
    }
  }
})

onMounted(async () => {
  // 初始化Mermaid
  initMermaid()
  
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

.markdown-preview {
  width: 100%;
  height: 100%;
  overflow: auto;
  background: white;
  padding: var(--spacing-lg);
}

.markdown-content {
  max-width: none;
  line-height: 1.6;
  color: #333;
}

.markdown-content h1,
.markdown-content h2,
.markdown-content h3,
.markdown-content h4,
.markdown-content h5,
.markdown-content h6 {
  margin-top: var(--spacing-lg);
  margin-bottom: var(--spacing-md);
  font-weight: 600;
  line-height: 1.25;
}

.markdown-content h1 { font-size: 2rem; border-bottom: 1px solid #eaecef; padding-bottom: 0.3rem; }
.markdown-content h2 { font-size: 1.5rem; border-bottom: 1px solid #eaecef; padding-bottom: 0.3rem; }
.markdown-content h3 { font-size: 1.25rem; }
.markdown-content h4 { font-size: 1rem; }
.markdown-content h5 { font-size: 0.875rem; }
.markdown-content h6 { font-size: 0.85rem; color: #6a737d; }

.markdown-content p {
  margin-bottom: var(--spacing-md);
}

.markdown-content ul,
.markdown-content ol {
  margin-bottom: var(--spacing-md);
  padding-left: var(--spacing-lg);
}

.markdown-content li {
  margin-bottom: var(--spacing-xs);
}

.markdown-content blockquote {
  border-left: 4px solid #dfe2e5;
  padding: 0 var(--spacing-md);
  margin: var(--spacing-md) 0;
  color: #6a737d;
}

.markdown-content code {
  background: #f6f8fa;
  border-radius: 3px;
  padding: 0.2rem 0.4rem;
  font-size: 0.85rem;
  font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, monospace;
}

.markdown-content pre {
  background: #f6f8fa;
  border-radius: 6px;
  padding: var(--spacing-md);
  margin-bottom: var(--spacing-md);
  overflow-x: auto;
  font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, monospace;
}

.markdown-content pre code {
  background: none;
  padding: 0;
  border-radius: 0;
}

.markdown-content table {
  border-spacing: 0;
  border-collapse: collapse;
  margin-bottom: var(--spacing-md);
  width: 100%;
}

.markdown-content table th,
.markdown-content table td {
  border: 1px solid #d0d7de;
  padding: var(--spacing-xs) var(--spacing-sm);
  text-align: left;
}

.markdown-content table th {
  background: #f6f8fa;
  font-weight: 600;
}

.markdown-content table tr:nth-child(2n) {
  background: #f6f8fa;
}

.markdown-content a {
  color: #0969da;
  text-decoration: none;
}

.markdown-content a:hover {
  text-decoration: underline;
}

.markdown-content img {
  max-width: 100%;
  height: auto;
}

.markdown-content hr {
  border: none;
  border-top: 1px solid #d0d7de;
  margin: var(--spacing-lg) 0;
}

/* Mermaid图表样式 */
.markdown-content .mermaid-diagram {
  margin: var(--spacing-lg) 0;
  text-align: center;
  background: #f8f9fa;
  border-radius: var(--border-radius-md);
  padding: var(--spacing-md);
  overflow-x: auto;
}

.markdown-content .mermaid-diagram svg {
  max-width: 100%;
  height: auto;
}

.markdown-content .mermaid-error {
  color: #d73a49;
  background: #ffeef0;
  border: 1px solid #f1c0c7;
  border-radius: var(--border-radius-sm);
  padding: var(--spacing-sm);
  font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, monospace;
  font-size: 0.85rem;
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

.markdown-preview::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}

.markdown-preview::-webkit-scrollbar-track {
  background: transparent;
}

.markdown-preview::-webkit-scrollbar-thumb {
  background: #c1c1c1;
  border-radius: 3px;
}

.markdown-preview::-webkit-scrollbar-thumb:hover {
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
