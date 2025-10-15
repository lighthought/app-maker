<template>
  <div class="project-panel">
    <!-- 面板头部 - 单行布局 -->
    <div class="panel-header">
      <!-- 左侧：设备视图切换 - 只在预览模式下显示 -->
      <div v-if="activeTab === 'preview' && project?.preview_url" class="header-left">
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
      
      <!-- 中间：代码/预览切换（居中） -->
      <div class="header-center">
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
      
      <!-- 右侧：操作按钮 + 部署按钮 - 只在预览模式下显示 -->
      <div v-if="activeTab === 'preview' && project?.preview_url" class="header-right">
        <div class="preview-actions">
          <!-- 桌面端显示所有操作按钮 -->
          <template v-if="!isMobile">
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
          </template>
          
          <!-- 刷新按钮：桌面和移动端都显示 -->
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
          
          <!-- 桌面端显示新标签页打开按钮 -->
          <n-tooltip v-if="!isMobile" placement="bottom">
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

        <!-- 部署按钮 -->
        <n-button type="primary" size="small" @click="deployProject">
          {{ t('preview.deploy') }}
        </n-button>
      </div>
    </div>

    <!-- 代码面板 -->
    <div v-if="activeTab === 'code'" class="code-panel">
      <!-- 文件树展开按钮（当树被收起时显示） -->
      <div v-if="isFileTreeCollapsed" class="expand-file-tree-btn">
        <n-button text @click="toggleFileTree" class="expand-btn">
          <template #icon>
            <n-icon><ChevronRightIcon /></n-icon>
          </template>
        </n-button>
      </div>
      
      <!-- 文件树 -->
      <div class="file-tree" :class="{ 'collapsed': isFileTreeCollapsed }">
        <div class="tree-header">
          <h4>{{ t('project.projectFiles') }}</h4>
          <div class="tree-header-actions">
            <n-button text size="tiny" @click="refreshFiles">
              <template #icon>
                <n-icon><RefreshIcon /></n-icon>
              </template>
            </n-button>
            <n-button text size="tiny" @click="toggleFileTree" class="toggle-tree-btn">
              <template #icon>
                <n-icon><ChevronLeftIcon /></n-icon>
              </template>
            </n-button>
          </div>
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
import { ref, computed, h, onMounted, onUnmounted, watch } from 'vue'
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
  ExternalLinkIcon,
  ChevronLeftIcon,
  ChevronRightIcon,
  FolderIcon
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
const isFileTreeCollapsed = ref(false)

// 预览相关状态
const deviceView = ref<'desktop' | 'tablet' | 'mobile'>('desktop')
const showShareModal = ref(false)
const iframeKey = ref(0)

// 移动端检测
const isMobile = ref(false)
const checkMobile = () => {
  isMobile.value = window.innerWidth <= 768
}



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

// 切换文件树折叠状态
const toggleFileTree = () => {
  isFileTreeCollapsed.value = !isFileTreeCollapsed.value
}

// 一键部署项目
const deployProject = async () => {
  if (!props.project?.guid) return
  
  try {
    messageApi.loading(t('preview.deploying'), { duration: 0 })
    const projectStore = useProjectStore()
    await projectStore.deployProject(props.project.guid)
    messageApi.destroyAll()
    messageApi.success(t('preview.deploySuccess'))
    // 部署成功后刷新预览
    setTimeout(() => {
      refreshPreview()
    }, 2000)
  } catch (error: any) {
    messageApi.destroyAll()
    messageApi.error(error.message || t('preview.deployFailed'))
    console.error('部署项目失败:', error)
  }
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

// 监听外部触发的预览切换事件
const handleSwitchToPreview = () => {
  activeTab.value = 'preview'
  console.log('自动切换到预览视图')
}

// 初始化
onMounted(async () => {
  // 检测移动端
  checkMobile()
  window.addEventListener('resize', checkMobile)
  
  // 监听切换到预览的自定义事件
  const el = document.querySelector('.project-panel')
  if (el) {
    el.addEventListener('switch-to-preview', handleSwitchToPreview as EventListener)
  }
  
  // 如果项目数据已经存在，直接加载文件
  if (props.project?.guid) {
    console.log(t('project.projectDataExists'), props.project.guid)
    await loadProjectFiles()
  } else {
    console.log(t('project.projectDataNotLoaded'))
  }
})

// 组件卸载时移除监听
onUnmounted(() => {
  window.removeEventListener('resize', checkMobile)
  
  // 移除自定义事件监听
  const el = document.querySelector('.project-panel')
  if (el) {
    el.removeEventListener('switch-to-preview', handleSwitchToPreview as EventListener)
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
  justify-content: center;  /* 居中对齐 */
  align-items: center;
  padding: 12px 20px;
  border-bottom: 1px solid #e2e8f0;
  background: white;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  z-index: 10;
  gap: var(--spacing-md);
  position: relative;
}

/* 左侧（预览模式）：设备切换 */
.header-left {
  display: flex;
  align-items: center;
  flex-shrink: 0;
  position: absolute;
  left: 20px;
}

/* 中间：代码/预览切换（居中） */
.header-center {
  display: flex;
  align-items: center;
  flex-shrink: 0;
}

/* 右侧：操作按钮 + 部署按钮 */
.header-right {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  flex-shrink: 0;
  position: absolute;
  right: 20px;
}

.preview-actions {
  display: flex;
  align-items: center;
  gap: 4px;
}

/* 代码面板样式 */
.code-panel {
  flex: 1;
  display: flex;
  overflow: hidden;
  position: relative;
}

/* 文件树展开按钮 */
.expand-file-tree-btn {
  position: absolute;
  left: 0;
  top: 50%;
  transform: translateY(-50%);
  z-index: 100;
  background: white;
  border: 1px solid var(--border-color);
  border-left: none;
  border-radius: 0 8px 8px 0;
  box-shadow: 2px 0 8px rgba(0, 0, 0, 0.1);
}

.expand-btn {
  padding: 8px 4px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.expand-btn :deep(.n-icon) {
  font-size: 20px;
  color: var(--primary-color);
}

.file-tree {
  width: 250px;
  min-width: 200px;
  border-right: 1px solid var(--border-color);
  display: flex;
  flex-direction: column;
  transition: all 0.3s ease;
}

.file-tree.collapsed {
  width: 0;
  min-width: 0;
  border-right: none;
  overflow: hidden;
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

.tree-header-actions {
  display: flex;
  align-items: center;
  gap: 4px;
}

.toggle-tree-btn {
  transition: transform 0.3s ease;
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
  padding: var(--spacing-lg);  /* 上下左右相同边距 */
  overflow: auto;
  box-sizing: border-box;  /* 包含 padding 在内的盒模型 */
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
  height: calc(100% - var(--spacing-lg) * 2);  /* ✨ 自适应高度，上下留边距 */
  max-width: 100%;
  max-height: 900px;  /* ✨ 最大高度限制 */
  min-height: 600px;  /* ✨ 最小高度保证 */
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
  overflow-y: auto;  /* ✨ 允许垂直滚动 */
  overflow-x: hidden;  /* 隐藏横向滚动 */
  position: relative;
  display: flex;
  flex-direction: column;
  -webkit-overflow-scrolling: touch;  /* ✨ iOS 平滑滚动 */
  scroll-behavior: smooth;  /* ✨ 平滑滚动 */
}

/* 手机外框的滚动条样式 */
.phone-frame::-webkit-scrollbar {
  width: 4px;
}

.phone-frame::-webkit-scrollbar-track {
  background: rgba(0, 0, 0, 0.05);
  border-radius: 32px;
}

.phone-frame::-webkit-scrollbar-thumb {
  background: rgba(0, 0, 0, 0.2);
  border-radius: 32px;
}

.phone-frame::-webkit-scrollbar-thumb:hover {
  background: rgba(0, 0, 0, 0.3);
}

/* iPhone 刘海 */
.phone-notch {
  position: sticky;  /* ✨ sticky 定位，滚动时固定在顶部 */
  top: 0;
  width: 180px;
  height: 30px;
  background: #1f1f1f;
  border-radius: 0 0 20px 20px;
  z-index: 10;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
  flex-shrink: 0;  /* ✨ 不收缩 */
  margin: 0 auto;  /* ✨ 水平居中 */
  align-self: center;  /* ✨ flex 容器中居中 */
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
  flex: 1;  /* ✨ 占据剩余空间 */
  min-height: 100%;  /* ✨ 最小高度100%，允许内容更长 */
  border: none;
  border-radius: 0 0 32px 32px;  /* 只圆角底部 */
  display: block;
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

/* 响应式设计 */
@media (max-width: 1024px) {
  .project-panel {
    height: 100%;
  }
  
  .panel-header {
    padding: 10px 16px;
    gap: 8px;
  }
  
  /* 移动端文件树优化 */
  .file-tree {
    width: 180px;
    min-width: 150px;
  }
  
  .tree-header {
    padding: var(--spacing-sm);
  }
  
  .tree-header h4 {
    font-size: 0.85rem;
  }
  
  /* 展开按钮优化 */
  .expand-btn {
    padding: 6px 3px;
  }
  
  .expand-btn :deep(.n-icon) {
    font-size: 18px;
  }
}

@media (max-width: 768px) {
  .panel-header {
    padding: 8px 12px;
    gap: 6px;
    justify-content: center;  /* 居中对齐 */
    flex-wrap: nowrap;  /* 不换行 */
    min-height: 48px;  /* 保证最小高度 */
  }
  
  .header-left {
    position: absolute;  /* 绝对定位到左侧 */
    left: 12px;
  }
  
  .header-center {
    position: static;  /* 静态定位，居中显示 */
    /* 移除 absolute 定位，让它自然居中 */
  }
  
  .header-right {
    position: absolute;  /* 绝对定位到右侧 */
    right: 12px;
  }
  
  .panel-header :deep(.n-button) {
    font-size: 0.85rem;
    padding: 0 8px;
  }
  
  .panel-header :deep(.n-button-group .n-button) {
    padding: 0 6px;  /* 按钮组更紧凑 */
  }
  
  .preview-actions {
    gap: 4px;
  }
  
  .preview-actions :deep(.n-button) {
    padding: 0 4px;  /* 操作按钮更紧凑 */
  }
  
  /* 文件树更紧凑 */
  .file-tree {
    width: 150px;
    min-width: 120px;
  }
  
  .tree-header {
    padding: 8px;
  }
  
  .tree-header h4 {
    font-size: 0.8rem;
  }
  
  .tree-content {
    padding: 4px;
  }
  
  /* 展开按钮 */
  .expand-btn {
    padding: 4px 2px;
  }
  
  .expand-btn :deep(.n-icon) {
    font-size: 16px;
  }
  
  .code-editor,
  .preview-container {
    font-size: 0.85rem;
  }
  
  .empty-state {
    padding: var(--spacing-lg);
  }
  
  .empty-state h3 {
    font-size: 1rem;
  }
  
  .empty-state p {
    font-size: 0.85rem;
  }
}

@media (max-width: 480px) {
  .panel-header {
    padding: 6px 8px;
    gap: 4px;
    justify-content: center;
    flex-wrap: nowrap;
    min-height: 44px;  /* 保证最小高度 */
  }
  
  .header-left {
    position: absolute;
    left: 8px;
  }
  
  .header-center {
    position: static;  /* 静态定位，自然居中 */
  }
  
  .header-right {
    position: absolute;
    right: 8px;
  }
  
  .panel-header :deep(.n-button) {
    font-size: 0.75rem;
    padding: 0 4px;
    min-width: auto;
  }
  
  .panel-header :deep(.n-button-group .n-button) {
    padding: 0 4px;
  }
  
  .panel-header :deep(.n-button .n-icon) {
    font-size: 14px;
  }
  
  .preview-actions {
    gap: 2px;
  }
  
  .preview-actions :deep(.n-button) {
    padding: 0 3px;
  }
  
  /* 部署按钮文字可能需要隐藏，只显示图标 */
  .header-right :deep(.n-button:not(.n-button--text)) {
    padding: 0 8px;
  }
  
  /* 小手机文件树默认折叠 */
  .file-tree {
    width: 0;
    min-width: 0;
    border-right: none;
  }
  
  .file-tree:not(.collapsed) {
    width: 200px;
    min-width: 180px;
    border-right: 1px solid var(--border-color);
  }
  
  .tree-header {
    padding: 6px 8px;
  }
  
  .tree-header h4 {
    font-size: 0.75rem;
  }
  
  .tree-content {
    padding: 2px;
  }
  
  .empty-state {
    padding: var(--spacing-md);
  }
  
  .empty-state h3 {
    font-size: 0.95rem;
  }
  
  .empty-state p {
    font-size: 0.8rem;
  }
}
</style>