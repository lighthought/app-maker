<template>
  <div class="project-panel">
    <!-- 面板头部 -->
    <div class="panel-header">
      <div class="header-left">
        <n-button-group>
          <n-button
            :type="activeTab === 'code' ? 'primary' : 'default'"
            size="small"
            @click="activeTab = 'code'"
          >
            <template #icon>
              <n-icon><CodeIcon /></n-icon>
            </template>
            代码
          </n-button>
          <n-button
            :type="activeTab === 'preview' ? 'primary' : 'default'"
            size="small"
            @click="activeTab = 'preview'"
          >
            <template #icon>
              <n-icon><PreviewIcon /></n-icon>
            </template>
            预览
          </n-button>
        </n-button-group>
      </div>
    </div>

    <!-- 代码面板 -->
    <div v-if="activeTab === 'code'" class="code-panel">
      <!-- 文件树 -->
      <div class="file-tree">
        <div class="tree-header">
          <h4>项目文件</h4>
          <n-button text size="tiny" @click="refreshFiles">
            <template #icon>
              <n-icon><RefreshIcon /></n-icon>
            </template>
          </n-button>
        </div>
        
        <div class="tree-content">
          <!-- 加载状态 -->
          <div v-if="isLoadingFiles" class="loading-state">
            <n-spin size="small" />
            <span>加载文件列表中...</span>
          </div>
          
          <!-- 暂无数据状态 -->
          <div v-else-if="fileTree.length === 0 && !isLoadingFiles" class="empty-state">
            <n-icon size="32" color="#CBD5E0">
              <FolderIcon />
            </n-icon>
            <p>暂无文件数据</p>
            <n-button text size="small" @click="refreshFiles">
              手动刷新
            </n-button>
          </div>
          
          <!-- 文件树 -->
          <FileTreeNodeComponent
            v-else
            v-for="file in fileTree"
            :key="file.path"
            :node="file"
            :selected-file="selectedFile"
            :project-guid="project?.guid"
            @select-file="selectFile"
            @expand-folder="selectFile"
          />
        </div>
      </div>

      <!-- 代码编辑器 -->
      <div class="code-editor">
        <div class="editor-header">
          <div class="file-info">
            <span class="file-path">{{ selectedFile?.path || '选择文件查看代码' }}</span>
          </div>
          <div class="editor-actions">
            <n-button text size="tiny" @click="copyCode">
              <template #icon>
                <n-icon><CopyIcon /></n-icon>
              </template>
              复制
            </n-button>
          </div>
        </div>
        
        <div class="editor-content">
          <MonacoEditor
            v-if="selectedFile?.content"
            :value="selectedFile.content"
            :language="getLanguage(selectedFile.path)"
            :read-only="true"
            theme="vs"
            height="100%"
          />
          <div v-else class="empty-editor">
            <n-icon size="48" color="#CBD5E0">
              <FileIcon />
            </n-icon>
            <p>选择一个文件查看代码内容</p>
          </div>
        </div>
      </div>
    </div>

    <!-- 预览面板 -->
    <div v-else-if="activeTab === 'preview'" class="preview-panel">
      <div v-if="project?.previewUrl" class="preview-content">
        <div class="preview-header">
          <div class="preview-info">
            <n-icon size="16" color="#38A169">
              <GlobeIcon />
            </n-icon>
            <span class="preview-url">{{ project.previewUrl }}</span>
          </div>
          <div class="preview-actions">
            <n-button text size="tiny" @click="openInNewTab">
              <template #icon>
                <n-icon><ExternalLinkIcon /></n-icon>
              </template>
              新窗口打开
            </n-button>
            <n-button text size="tiny" @click="refreshPreview">
              <template #icon>
                <n-icon><RefreshIcon /></n-icon>
              </template>
              刷新
            </n-button>
          </div>
        </div>
        
        <div class="preview-frame">
          <iframe
            :src="project.previewUrl"
            frameborder="0"
            class="preview-iframe"
            @load="onPreviewLoad"
            @error="onPreviewError"
          ></iframe>
        </div>
      </div>
      
      <div v-else class="preview-empty">
        <n-icon size="64" color="#CBD5E0">
          <GlobeIcon />
        </n-icon>
        <h3>预览暂不可用</h3>
        <p>项目正在开发中，预览功能将在部署完成后可用</p>
        <n-button type="primary" @click="activeTab = 'code'">
          查看代码
        </n-button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, h, onMounted, watch } from 'vue'
import { NIcon, NButton, NButtonGroup, NTag, NSpin, useMessage } from 'naive-ui'
import { useProjectStore } from '@/stores/project'
import { useFilesStore, type FileTreeNode } from '@/stores/file'
import type { Project } from '@/types/project'
import MonacoEditor from './MonacoEditor.vue'
import FileTreeNodeComponent from './FileTreeNode.vue'

interface Props {
  project?: Project
}

const props = defineProps<Props>()

// 获取message实例
const messageApi = useMessage()

const fileStore = useFilesStore()

// 响应式数据
const activeTab = ref<'code' | 'preview'>('code')
const selectedFile = ref<FileTreeNode | null>(null)
const fileTree = ref<FileTreeNode[]>([])
const previewLoading = ref(false)
const isLoadingFiles = ref(false)


// 加载项目文件树
const loadProjectFiles = async () => {
  if (!props.project?.guid) {
    console.log('项目GUID不存在，跳过文件加载')
    return
  }
  
  isLoadingFiles.value = true
  try {
    const tree = await fileStore.getProjectFileTree(props.project.guid)
    fileTree.value = tree
    console.log('文件树加载完成:', tree.length, '个文件/文件夹')
  } catch (error) {
    console.error('加载项目文件失败:', error)
    fileTree.value = []
  } finally {
    isLoadingFiles.value = false
  }
}

// 获取状态类型
const getStatusType = (status?: string): 'default' | 'primary' | 'info' | 'success' | 'warning' | 'error' => {
  const statusMap: Record<string, 'default' | 'primary' | 'info' | 'success' | 'warning' | 'error'> = {
    pending: 'default',
    in_progress: 'warning',
    done: 'success',
    failed: 'error'
  }
  return statusMap[status || 'pending'] || 'default'
}

// 获取状态文本
const getStatusText = (status?: string) => {
  const statusMap: Record<string, string> = {
    pending: '草稿',
    in_progress: '进行中',
    done: '已完成',
    failed: '失败'
  }
  return statusMap[status || 'pending'] || '草稿'
}


// 获取语言类型（用于Monaco Editor）
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

// 选择文件或展开文件夹
const selectFile = async (file: FileTreeNode) => {
  if (file.type === 'file') {
    selectedFile.value = file
    
    // 如果文件内容未加载，则加载内容
    if (!file.content && props.project?.guid) {
      try {
        const fileContent = await fileStore.getFileContent(props.project.guid, file.path)
        if (fileContent) {
          file.content = fileContent.content
        }
      } catch (error) {
        console.error('加载文件内容失败:', error)
      }
    }
  } else if (file.type === 'folder') {
    // 展开或收起文件夹
    if (!file.expanded && !file.loaded) {
      // 展开文件夹
      try {
        await fileStore.expandFolder(props.project!.guid, file.path, fileTree.value)
        file.expanded = true
        console.log('展开文件夹:', file.path)
      } catch (error) {
        console.error('展开文件夹失败:', error)
      }
    } else {
      // 切换展开状态
      file.expanded = !file.expanded
      console.log('切换展开状态:', file.path, file.expanded)
    }
  }
}

// 刷新文件
const refreshFiles = async () => {
  await loadProjectFiles()
}

// 复制代码
const copyCode = async () => {
  if (selectedFile.value?.content) {
    try {
      await navigator.clipboard.writeText(selectedFile.value.content)
      // 显示复制成功提示
      messageApi.success('代码复制成功', {
        duration: 2000,
        closable: false
      })
    } catch (err) {
      console.error('复制失败:', err)
      // 显示复制失败提示
      messageApi.error('复制失败，请重试', {
        duration: 2000,
        closable: false
      })
    }
  }
}

// 在新窗口打开预览
const openInNewTab = () => {
  if (props.project?.previewUrl) {
    window.open(props.project.previewUrl, '_blank')
  }
}

// 刷新预览
const refreshPreview = () => {
  previewLoading.value = true
  // 这里可以重新加载iframe
}

// 预览加载完成
const onPreviewLoad = () => {
  previewLoading.value = false
}

// 预览加载错误
const onPreviewError = () => {
  previewLoading.value = false
  console.error('预览加载失败')
}

// 图标组件
const CodeIcon = () => h('svg', { 
  viewBox: '0 0 24 24', 
  fill: 'currentColor',
  style: 'width: 1em; height: 1em;'
}, [
  h('path', { d: 'M9.4 16.6L4.8 12l4.6-4.6L8 6l-6 6 6 6 1.4-1.4zm5.2 0L19.2 12l-4.6-4.6L16 6l6 6-6 6-1.4-1.4z' })
])

const PreviewIcon = () => h('svg', { 
  viewBox: '0 0 24 24', 
  fill: 'currentColor',
  style: 'width: 1em; height: 1em;'
}, [
  h('path', { d: 'M12 4.5C7 4.5 2.73 7.61 1 12c1.73 4.39 6 7.5 11 7.5s9.27-3.11 11-7.5c-1.73-4.39-6-7.5-11-7.5zM12 17c-2.76 0-5-2.24-5-5s2.24-5 5-5 5 2.24 5 5-2.24 5-5 5zm0-8c-1.66 0-3 1.34-3 3s1.34 3 3 3 3-1.34 3-3-1.34-3-3-3z' })
])

const FileIcon = () => h('svg', { 
  viewBox: '0 0 24 24', 
  fill: 'currentColor',
  style: 'width: 1em; height: 1em;'
}, [
  h('path', { d: 'M14,2H6A2,2 0 0,0 4,4V20A2,2 0 0,0 6,22H18A2,2 0 0,0 20,20V8L14,2M18,20H6V4H13V9H18V20Z' })
])

const RefreshIcon = () => h('svg', { 
  viewBox: '0 0 24 24', 
  fill: 'currentColor',
  style: 'width: 1em; height: 1em;'
}, [
  h('path', { d: 'M17.65,6.35C16.2,4.9 14.21,4 12,4A8,8 0 0,0 4,12A8,8 0 0,0 12,20C15.73,20 18.84,17.45 19.73,14H17.65C16.83,16.33 14.61,18 12,18A6,6 0 0,1 6,12A6,6 0 0,1 12,6C13.66,6 15.14,6.69 16.22,7.78L13,11H20V4L17.65,6.35Z' })
])

const CopyIcon = () => h('svg', { 
  viewBox: '0 0 24 24', 
  fill: 'currentColor',
  style: 'width: 1em; height: 1em;'
}, [
  h('path', { d: 'M16 1H4c-1.1 0-2 .9-2 2v14h2V3h12V1zm3 4H8c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h11c1.1 0 2-.9 2-2V7c0-1.1-.9-2-2-2zm0 16H8V7h11v14z' })
])

const GlobeIcon = () => h('svg', { 
  viewBox: '0 0 24 24', 
  fill: 'currentColor',
  style: 'width: 1em; height: 1em;'
}, [
  h('path', { d: 'M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-1 17.93c-3.94-.49-7-3.85-7-7.93 0-.62.08-1.21.21-1.79L9 15v1c0 1.1.9 2 2 2v1.93zm6.9-2.54c-.26-.81-1-1.39-1.9-1.39h-1v-3c0-.55-.45-1-1-1H8v-2h2c.55 0 1-.45 1-1V7h2c1.1 0 2-.9 2-2v-.41c2.93 1.19 5 4.06 5 7.41 0 2.08-.8 3.97-2.1 5.39z' })
])

const ExternalLinkIcon = () => h('svg', { 
  viewBox: '0 0 24 24', 
  fill: 'currentColor',
  style: 'width: 1em; height: 1em;'
}, [
  h('path', { d: 'M19 19H5V5h7V3H5c-1.11 0-2 .9-2 2v14c0 1.1.89 2 2 2h14c1.1 0 2-.9 2-2v-7h-2v7zM14 3v2h3.59l-9.83 9.83 1.41 1.41L19 6.41V10h2V3h-7z' })
])


// 监听项目数据变化，当项目加载完成后自动加载文件
watch(() => props.project, (newProject) => {
  if (newProject?.guid) {
    console.log('项目数据已加载，开始加载文件:', newProject.guid)
    loadProjectFiles()
  }
}, { immediate: true })

// 暴露方法给父组件
defineExpose({
  refreshFiles
})

// 初始化
onMounted(async () => {
  // 如果项目数据已经存在，直接加载文件
  if (props.project?.guid) {
    console.log('组件挂载时项目数据已存在，开始加载文件:', props.project.guid)
    await loadProjectFiles()
  } else {
    console.log('组件挂载时项目数据尚未加载，等待项目数据...')
  }
})
</script>

<style scoped>
.project-panel {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: white;
  border-radius: var(--border-radius-lg);
  overflow: hidden;
}

.panel-header {
  display: flex;
  justify-content: flex-start;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid #e2e8f0;
  background: white;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  z-index: 10;
}

.header-left {
  display: flex;
  align-items: center;
  height: var(--height-sm);
}

/* 代码面板样式 */
.code-panel {
  flex: 1;
  display: flex;
  overflow: hidden;
}

.file-tree {
  width: 250px;
  border-right: 1px solid var(--border-color);
  display: flex;
  flex-direction: column;
}

.tree-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--spacing-md) var(--spacing-lg);
  border-bottom: 1px solid var(--border-color);
  background: var(--background-color);
  height: var(--height-md);
}

.tree-header h4 {
  margin: 0;
  font-size: 0.9rem;
  color: var(--text-primary);
}

.tree-content {
  flex: 1;
  overflow-y: auto;
  padding: var(--spacing-sm);
}


.code-editor {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.editor-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--spacing-md) var(--spacing-lg);
  border-bottom: 1px solid var(--border-color);
  background: var(--background-color);
  height: var(--height-md);
}

.file-info {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
}

.file-path {
  font-size: 0.9rem;
  color: var(--text-secondary);
  font-family: 'Courier New', monospace;
}

.editor-actions {
  display: flex;
  gap: var(--spacing-sm);
}

.editor-content {
  flex: 1;
  overflow: hidden;
  background: #f8f9fa;
  border-radius: var(--border-radius-md);
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

/* 预览面板样式 */
.preview-panel {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.preview-content {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.preview-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--spacing-md) var(--spacing-lg);
  border-bottom: 1px solid var(--border-color);
  background: var(--background-color);
}

.preview-info {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
}

.preview-url {
  font-size: 0.9rem;
  color: var(--text-secondary);
  font-family: 'Courier New', monospace;
}

.preview-actions {
  display: flex;
  gap: var(--spacing-sm);
}

.preview-frame {
  flex: 1;
  position: relative;
}

.preview-iframe {
  width: 100%;
  height: 100%;
  border: none;
}

.preview-empty {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: var(--text-secondary);
  text-align: center;
}

.preview-empty h3 {
  margin: var(--spacing-md) 0 var(--spacing-sm) 0;
  color: var(--text-primary);
}

.preview-empty p {
  margin: 0 0 var(--spacing-lg) 0;
  font-size: 0.9rem;
}

/* 滚动条样式 */
.file-tree .tree-content::-webkit-scrollbar,
.editor-content::-webkit-scrollbar {
  width: 6px;
}

.file-tree .tree-content::-webkit-scrollbar-track,
.editor-content::-webkit-scrollbar-track {
  background: transparent;
}

.file-tree .tree-content::-webkit-scrollbar-thumb,
.editor-content::-webkit-scrollbar-thumb {
  background: #c1c1c1;
  border-radius: 3px;
}

.file-tree .tree-content::-webkit-scrollbar-thumb:hover,
.editor-content::-webkit-scrollbar-thumb:hover {
  background: #a8a8a8;
}

/* 加载状态和空状态样式 */
.loading-state {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--spacing-sm);
  padding: var(--spacing-xl);
  color: var(--text-secondary);
  font-size: 0.9rem;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: var(--spacing-xl);
  color: var(--text-secondary);
  text-align: center;
}

.empty-state p {
  margin: var(--spacing-md) 0 var(--spacing-sm) 0;
  font-size: 0.9rem;
}

.empty-state .n-button {
  margin-top: var(--spacing-sm);
}
</style>