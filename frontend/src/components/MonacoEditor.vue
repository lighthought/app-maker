<template>
  <div ref="editorContainer" class="monaco-editor-container"></div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch, nextTick } from 'vue'

interface Props {
  value: string
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
      value: props.value,
      language: props.language,
      readOnly: props.readOnly,
      theme: props.theme,
      automaticLayout: true,
      minimap: { enabled: false },
      scrollBeyondLastLine: false,
      wordWrap: 'on',
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
const updateEditor = () => {
  if (editor && editor.getValue() !== props.value) {
    editor.setValue(props.value)
  }
}

// 更新语言
const updateLanguage = () => {
  if (editor && editor.getModel()) {
    editor.getModel().setLanguage(props.language)
  }
}

// 监听属性变化
watch(() => props.value, updateEditor)
watch(() => props.language, updateLanguage)

onMounted(async () => {
  await nextTick()
  await initEditor()
})

onUnmounted(() => {
  if (editor) {
    editor.dispose()
  }
})
</script>

<style scoped>
.monaco-editor-container {
  width: 100%;
  height: 100%;
  border-radius: var(--border-radius-md);
  overflow: hidden;
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
</style>
