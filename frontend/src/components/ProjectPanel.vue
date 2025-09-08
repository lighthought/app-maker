<template>
  <div class="project-panel">
    <!-- 面板头部 -->
    <div class="panel-header">
      <div class="header-left">
        <h3>{{ project?.name || '项目面板' }}</h3>
        <n-tag :type="getStatusType(project?.status)" size="small">
          {{ getStatusText(project?.status) }}
        </n-tag>
      </div>
      <div class="header-right">
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
          <div
            v-for="file in fileTree"
            :key="file.path"
            class="tree-item"
            :class="{ 'tree-item--active': selectedFile?.path === file.path }"
            @click="selectFile(file)"
          >
            <n-icon size="16" :color="getFileIconColor(file.type)">
              <component :is="getFileIcon(file.type)" />
            </n-icon>
            <span class="file-name">{{ file.name }}</span>
          </div>
        </div>
      </div>

      <!-- 代码编辑器 -->
      <div class="code-editor">
        <div class="editor-header">
          <div class="file-info">
            <n-icon size="16" :color="getFileIconColor(selectedFile?.type)">
              <component :is="getFileIcon(selectedFile?.type)" />
            </n-icon>
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
          <pre v-if="selectedFile?.content" class="code-content"><code :class="getLanguageClass(selectedFile.type)">{{ selectedFile.content }}</code></pre>
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
import { ref, computed, h, onMounted } from 'vue'
import { NIcon, NButton, NButtonGroup, NTag } from 'naive-ui'
import { useProjectStore } from '@/stores/project'
import type { Project } from '@/types/project'

interface Props {
  project?: Project
}

interface FileItem {
  name: string
  path: string
  type: 'file' | 'folder'
  content?: string
}

const props = defineProps<Props>()
const projectStore = useProjectStore()

// 响应式数据
const activeTab = ref<'code' | 'preview'>('code')
const selectedFile = ref<FileItem | null>(null)
const fileTree = ref<FileItem[]>([])
const previewLoading = ref(false)


// 加载项目文件
const loadProjectFiles = async () => {
  if (!props.project?.id) return
  
  try {
    const files = await projectStore.getProjectFiles(props.project.id)
    if (files) {
      fileTree.value = files.map(file => ({
        name: file.name,
        path: file.path,
        type: file.type,
        content: undefined // 内容按需加载
      }))
    }
  } catch (error) {
    console.error('加载项目文件失败:', error)
  }
}

// 获取状态类型
const getStatusType = (status?: string): 'default' | 'primary' | 'info' | 'success' | 'warning' | 'error' => {
  const statusMap: Record<string, 'default' | 'primary' | 'info' | 'success' | 'warning' | 'error'> = {
    draft: 'default',
    in_progress: 'warning',
    completed: 'success',
    failed: 'error'
  }
  return statusMap[status || 'draft'] || 'default'
}

// 获取状态文本
const getStatusText = (status?: string) => {
  const statusMap: Record<string, string> = {
    draft: '草稿',
    in_progress: '进行中',
    completed: '已完成',
    failed: '失败'
  }
  return statusMap[status || 'draft'] || '草稿'
}

// 获取文件图标
const getFileIcon = (type?: string) => {
  const iconMap = {
    file: FileIcon,
    folder: FolderIcon
  }
  return iconMap[type as keyof typeof iconMap] || FileIcon
}

// 获取文件图标颜色
const getFileIconColor = (type?: string) => {
  const colorMap = {
    file: '#666',
    folder: '#3182CE'
  }
  return colorMap[type as keyof typeof colorMap] || '#666'
}

// 获取语言类名
const getLanguageClass = (type?: string) => {
  const languageMap: Record<string, string> = {
    'src/App.vue': 'vue',
    'src/main.ts': 'typescript',
    'package.json': 'json'
  }
  return languageMap[selectedFile.value?.path || ''] || 'text'
}

// 选择文件
const selectFile = async (file: FileItem) => {
  if (file.type === 'file') {
    selectedFile.value = file
    
    // 如果文件内容未加载，则加载内容
    if (!file.content && props.project?.id) {
      try {
        const fileContent = await projectStore.getFileContent(props.project.id, file.path)
        if (fileContent) {
          file.content = fileContent.content
        }
      } catch (error) {
        console.error('加载文件内容失败:', error)
      }
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
      // 可以添加成功提示
    } catch (err) {
      console.error('复制失败:', err)
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

const FolderIcon = () => h('svg', { 
  viewBox: '0 0 24 24', 
  fill: 'currentColor',
  style: 'width: 1em; height: 1em;'
}, [
  h('path', { d: 'M10,4H4C2.89,4 2,4.89 2,6V18A2,2 0 0,0 4,20H20A2,2 0 0,0 22,18V8C22,6.89 21.1,6 20,6H12L10,4Z' })
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

// 初始化
onMounted(async () => {
  await loadProjectFiles()
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
  justify-content: space-between;
  align-items: center;
  padding: var(--spacing-lg);
  border-bottom: 1px solid var(--border-color);
  background: var(--background-color);
}

.header-left {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
}

.header-left h3 {
  margin: 0;
  color: var(--primary-color);
  font-size: 1.1rem;
}

.header-right {
  display: flex;
  align-items: center;
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

.tree-item {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  padding: var(--spacing-sm) var(--spacing-md);
  border-radius: var(--border-radius-sm);
  cursor: pointer;
  transition: background-color 0.2s ease;
}

.tree-item:hover {
  background: var(--background-color);
}

.tree-item--active {
  background: var(--primary-color);
  color: white;
}

.file-name {
  font-size: 0.9rem;
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
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
  overflow: auto;
  background: #f8f9fa;
}

.code-content {
  margin: 0;
  padding: var(--spacing-lg);
  font-family: 'Courier New', monospace;
  font-size: 0.9rem;
  line-height: 1.5;
  color: var(--text-primary);
  white-space: pre-wrap;
  word-wrap: break-word;
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
</style>