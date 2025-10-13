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
            {{ t('editor.code') }}
          </n-button>
          <n-button
            :type="activeTab === 'preview' ? 'primary' : 'default'"
            size="small"
            @click="activeTab = 'preview'"
          >
            <template #icon>
              <n-icon><PreviewIcon /></n-icon>
            </template>
            {{ t('editor.preview') }}
          </n-button>
        </n-button-group>
      </div>
      
      <!-- 预览控制按钮 - 只在预览标签激活时显示 -->
      <div v-if="activeTab === 'preview' && project?.preview_url" class="header-right">
        <!-- 设备视图切换 - 居中 -->
        <div class="device-view-controls">
          <n-button-group size="small">
            <n-tooltip placement="bottom">
              <template #trigger>
                <n-button
                  :type="deviceView === 'desktop' ? 'primary' : 'default'"
                  @click="deviceView = 'desktop'"
                >
                  <template #icon>
                    <n-icon><DesktopIcon /></n-icon>
                  </template>
                </n-button>
              </template>
              {{ t('preview.desktop') }}
            </n-tooltip>
            <n-tooltip placement="bottom">
              <template #trigger>
                <n-button
                  :type="deviceView === 'tablet' ? 'primary' : 'default'"
                  @click="deviceView = 'tablet'"
                >
                  <template #icon>
                    <n-icon><TabletIcon /></n-icon>
                  </template>
                </n-button>
              </template>
              {{ t('preview.tablet') }}
            </n-tooltip>
            <n-tooltip placement="bottom">
              <template #trigger>
                <n-button
                  :type="deviceView === 'mobile' ? 'primary' : 'default'"
                  @click="deviceView = 'mobile'"
                >
                  <template #icon>
                    <n-icon><PhoneIcon /></n-icon>
                  </template>
                </n-button>
              </template>
              {{ t('preview.mobile') }}
            </n-tooltip>
          </n-button-group>
        </div>
        
        <!-- 操作按钮 - 右侧 -->
        <div class="preview-actions-header">
          <n-tooltip placement="bottom">
            <template #trigger>
              <n-button text size="small" @click="copyPreviewUrl">
                <template #icon>
                  <n-icon><CopyIcon /></n-icon>
                </template>
              </n-button>
            </template>
            {{ t('preview.copyUrl') }}
          </n-tooltip>
          <n-tooltip placement="bottom">
            <template #trigger>
              <n-button text size="small" @click="showShareModal = true">
                <template #icon>
                  <n-icon><ShareIcon /></n-icon>
                </template>
              </n-button>
            </template>
            {{ t('preview.sharePreview') }}
          </n-tooltip>
          <n-tooltip placement="bottom">
            <template #trigger>
              <n-button text size="small" @click="refreshPreview">
                <template #icon>
                  <n-icon><RefreshIcon /></n-icon>
                </template>
              </n-button>
            </template>
            {{ t('common.refresh') }}
          </n-tooltip>
          <n-tooltip placement="bottom">
            <template #trigger>
              <n-button text size="small" @click="openInNewTab">
                <template #icon>
                  <n-icon><ExternalLinkIcon /></n-icon>
                </template>
              </n-button>
            </template>
            {{ t('editor.openInNewWindow') }}
          </n-tooltip>
        </div>
      </div>
    </div>

    <!-- 代码面板 -->
    <div v-if="activeTab === 'code'" class="code-panel">
      <!-- 文件树 -->
      <div class="file-tree">
        <div class="tree-header">
          <h4>{{ t('project.projectFiles') }}</h4>
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
            <span>{{ t('project.loadingFiles') }}</span>
          </div>
          
          <!-- 暂无数据状态 -->
          <div v-else-if="fileTree.length === 0 && !isLoadingFiles" class="empty-state">
            <n-icon size="32" color="#CBD5E0">
              <FolderIcon />
            </n-icon>
            <p>{{ t('project.noFileData') }}</p>
            <n-button text size="small" @click="refreshFiles">
              {{ t('common.manualRefresh') }}
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
        <MonacoEditor
          :project-guid="project?.guid"
          :file-path="selectedFile?.path"
          :language="getLanguage(selectedFile?.path)"
          :read-only="true"
          theme="vs"
          height="100%"
        />
      </div>
    </div>

    <!-- 预览面板 -->
    <div v-else-if="activeTab === 'preview'" class="preview-panel">
      <div v-if="project?.preview_url" class="preview-content">
        <div class="preview-frame-container">
          <div class="preview-frame" :class="`device-${deviceView}`">
            <!-- 移动端手机外框 -->
            <div v-if="deviceView === 'mobile'" class="phone-frame">
              <div class="phone-notch"></div>
              <iframe
                :key="iframeKey"
                :src="project.preview_url"
                frameborder="0"
                class="preview-iframe"
                @load="onPreviewLoad"
                @error="onPreviewError"
              ></iframe>
            </div>
            <!-- 非移动端直接显示 iframe -->
            <iframe
              v-else
              :key="iframeKey"
              :src="project.preview_url"
              frameborder="0"
              class="preview-iframe"
              @load="onPreviewLoad"
              @error="onPreviewError"
            ></iframe>
          </div>
        </div>
      </div>
      
      <div v-else class="preview-empty">
        <n-icon size="64" color="#CBD5E0">
          <GlobeIcon />
        </n-icon>
        <h3>{{ t('editor.previewUnavailable') }}</h3>
        <p>{{ t('editor.previewDevelopingNote') }}</p>
        <n-button type="primary" @click="activeTab = 'code'">
          {{ t('editor.viewCode') }}
        </n-button>
      </div>
    </div>

    <!-- 分享预览模态框 -->
    <SharePreviewModal
      v-model:show="showShareModal"
      :project-guid="project?.guid || ''"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, h, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { NIcon, NButton, NButtonGroup, NTag, NSpin, NTooltip, useMessage } from 'naive-ui'
import { useProjectStore } from '@/stores/project'
import { useFilesStore, type FileTreeNode } from '@/stores/file'
import type { Project } from '@/types/project'
import MonacoEditor from './MonacoEditor.vue'
import FileTreeNodeComponent from './FileTreeNode.vue'
import SharePreviewModal from './SharePreviewModal.vue'
// 导入图标
import {
  DesktopIcon,
  TabletIcon,
  PhoneIcon,
  ShareIcon,
  CodeIcon,
  PreviewIcon,
  RefreshIcon,
  FileIcon,
  CopyIcon,
  GlobeIcon,
  ExternalLinkIcon
} from '@/components/icon'

interface Props {
  project?: Project
}

const props = defineProps<Props>()
const { t } = useI18n()

// 获取message实例
const messageApi = useMessage()

const fileStore = useFilesStore()

// 响应式数据
const activeTab = ref<'code' | 'preview'>('code')
const selectedFile = ref<FileTreeNode | null>(null)
const fileTree = ref<FileTreeNode[]>([])
const previewLoading = ref(false)
const isLoadingFiles = ref(false)

// 预览相关状态
const deviceView = ref<'desktop' | 'tablet' | 'mobile'>('desktop')
const showShareModal = ref(false)
const iframeKey = ref(0)


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
    pending: t('common.draft'),
    in_progress: t('common.inProgress'),
    done: t('common.completed'),
    failed: t('common.failed')
  }
  return statusMap[status || 'pending'] || t('common.draft')
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


// 在新窗口打开预览
const openInNewTab = () => {
  if (props.project?.preview_url) {
    window.open(props.project.preview_url, '_blank')
  }
}

// 复制预览 URL
const copyPreviewUrl = async () => {
  if (!props.project?.preview_url) return
  
  try {
    await navigator.clipboard.writeText(props.project.preview_url)
    messageApi.success(t('preview.urlCopied'))
  } catch (error) {
    console.error('复制失败:', error)
    messageApi.error(t('preview.copyFailed'))
  }
}

// 刷新预览
const refreshPreview = () => {
  previewLoading.value = true
  // 通过改变 key 强制重新加载 iframe
  iframeKey.value++
}

// 预览加载完成
const onPreviewLoad = () => {
  previewLoading.value = false
}

// 预览加载错误
const onPreviewError = () => {
  previewLoading.value = false
  console.error(t('project.previewLoadFailed'))
}

// 监听项目数据变化，当项目加载完成后自动加载文件
watch(() => props.project, (newProject) => {
  if (newProject?.guid) {
    console.log(t('project.projectDataLoaded'), newProject.guid)
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
    console.log(t('project.projectDataExists'), props.project.guid)
    await loadProjectFiles()
  } else {
    console.log(t('project.projectDataNotLoaded'))
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
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid #e2e8f0;
  background: white;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  z-index: 10;
  gap: var(--spacing-md);
}

.header-left {
  display: flex;
  align-items: center;
  height: var(--height-sm);
}

.header-right {
  display: flex;
  align-items: center;
  flex: 1;
  position: relative;
  justify-content: flex-end;
}

.device-view-controls {
  position: absolute;
  left: 50%;
  transform: translateX(-50%);
  display: flex;
  align-items: center;
}

.preview-actions-header {
  display: flex;
  align-items: center;
  gap: 4px;
  z-index: 1;
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
  min-width: 0; /* 允许收缩 */
  overflow: hidden; /* 防止内容溢出 */
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

.preview-frame-container {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f5f5f5;
  padding: var(--spacing-lg);
  overflow: auto;
}

.preview-frame {
  background: white;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  border-radius: 8px;
  overflow: hidden;
  transition: all 0.3s ease;
}

/* 设备视图尺寸 */
.preview-frame.device-desktop {
  width: 100%;
  height: 100%;
  max-width: none;
}

.preview-frame.device-tablet {
  width: 768px;
  height: 1024px;
  max-width: 100%;
  max-height: 100%;
}

.preview-frame.device-mobile {
  width: 375px;
  height: 812px;
  max-width: 100%;
  max-height: 100%;
  background: #1f1f1f;
  border-radius: 40px;
  padding: 12px;
  box-shadow: 
    0 0 0 2px #1f1f1f,
    0 0 0 4px #3a3a3a,
    0 20px 60px rgba(0, 0, 0, 0.3),
    inset 0 0 6px rgba(255, 255, 255, 0.1);
  position: relative;
}

/* 手机外框容器 */
.phone-frame {
  width: 100%;
  height: 100%;
  background: white;
  border-radius: 32px;
  overflow: hidden;
  position: relative;
  display: flex;
  flex-direction: column;
}

/* iPhone 刘海 */
.phone-notch {
  position: absolute;
  top: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 180px;
  height: 30px;
  background: #1f1f1f;
  border-radius: 0 0 20px 20px;
  z-index: 10;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
}

/* 刘海内的扬声器 */
.phone-notch::before {
  content: '';
  position: absolute;
  top: 10px;
  left: 50%;
  transform: translateX(-50%);
  width: 60px;
  height: 6px;
  background: #2a2a2a;
  border-radius: 3px;
}

/* 移动端的 iframe 需要适应刘海 */
.phone-frame .preview-iframe {
  width: 100%;
  height: 100%;
  border: none;
  border-radius: 32px;
}

/* 非移动端的 iframe */
.preview-frame:not(.device-mobile) .preview-iframe {
  width: 100%;
  height: 100%;
  border: none;
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
.file-tree .tree-content::-webkit-scrollbar {
  width: 6px;
}

.file-tree .tree-content::-webkit-scrollbar-track {
  background: transparent;
}

.file-tree .tree-content::-webkit-scrollbar-thumb {
  background: #c1c1c1;
  border-radius: 3px;
}

.file-tree .tree-content::-webkit-scrollbar-thumb:hover {
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