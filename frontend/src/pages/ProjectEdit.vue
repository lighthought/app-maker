<template>
  <div class="project-edit-page">
    <!-- 主内容区域 - 分屏布局 -->
    <div class="main-content">
      <div class="split-container">
        <!-- 左侧对话窗口 -->
        <div class="left-panel" :style="{ width: leftWidth + '%' }">
          <!-- 左侧顶部导航 -->
          <div class="left-header">
            <n-button text @click="goBack" class="back-button">
              <template #icon>
                <n-icon><ArrowLeftIcon /></n-icon>
              </template>
              {{ t('common.back') }}
            </n-button>
            <div class="project-info">
              <h2 class="project-title">{{ getProjectDisplayName() }}</h2>
              <n-tag :type="getStatusType(project?.status)" size="small">
                {{ getStatusText(project?.status) }}
              </n-tag>
            </div>
          </div>
          
          <!-- 对话容器 -->
          <div class="conversation-wrapper">
            <ConversationContainer
              :project-guid="projectGuid"
              :requirements="project?.requirements || ''"
              :project="project"
              @project-info-update="handleProjectInfoUpdate"
              @project-env-setup="handleProjectEnvSetup"
            />
          </div>
        </div>
        
        <!-- 分割器 -->
        <div 
          class="splitter" 
          @mousedown="startResize"
          @touchstart="startResize"
        >
          <div class="splitter-handle"></div>
        </div>
        
        <!-- 右侧项目面板 -->
        <div class="right-panel" :style="{ width: rightWidth + '%' }">
          <ProjectPanel ref="projectPanelRef" :project="project" />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, h, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { NButton, NIcon, NTag } from 'naive-ui'
import ConversationContainer from '@/components/ConversationContainer.vue'
import ProjectPanel from '@/components/ProjectPanel.vue'
import { useProjectStore } from '@/stores/project'
import type { Project, ProjectInfoUpdate } from '@/types/project'

const route = useRoute()
const router = useRouter()
const projectStore = useProjectStore()
const { t } = useI18n()

// 响应式数据
const project = ref<Project | undefined>(undefined)
const loading = ref(false)

// 分割器相关
const leftWidth = ref(50)
const rightWidth = ref(50)
const isResizing = ref(false)

// 组件引用
const projectPanelRef = ref<any>(null)

// 计算属性
const projectGuid = computed(() => route.params.guid as string)

// 获取状态类型
const getStatusType = (status?: string): 'default' | 'primary' | 'info' | 'success' | 'warning' | 'error' => {
  const statusMap: Record<string, 'default' | 'primary' | 'info' | 'success' | 'warning' | 'error'> = {
    pending: 'default',
    in_progress: 'primary',
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

// 获取项目显示名称
const getProjectDisplayName = () => {
  if (!project.value) return t('project.editProject')
  
  // 如果项目名称是默认的"new-project"，尝试从对话消息中获取实际项目名称
  if (project.value.name === 'new-project' || project.value.name === t('project.editProject')) {
    // 这里可以从对话消息中提取项目名称，或者使用项目ID
    return t('project.projectWithId', { id: project.value.id.slice(-6) }) // 显示项目ID的后6位
  }
  
  return project.value.name || t('project.editProject')
}

// 处理项目信息更新
const handleProjectInfoUpdate = (info: ProjectInfoUpdate) => {
  if (!project.value) return
  
  // 更新项目信息
  if (info.name) {
    project.value.name = info.name
  }
  if (info.status) {
    project.value.status = info.status as any
  }
  if (info.description) {
    project.value.description = info.description
  }
  if (info.previewUrl) {
    project.value.previewUrl = info.previewUrl
  }
  
  console.log(t('project.projectInfoUpdated'), info)
}

// 处理项目环境配置完成
const handleProjectEnvSetup = () => {
  if (projectPanelRef.value && projectPanelRef.value.refreshFiles) {
    console.log(t('project.envSetupCompleted'))
    projectPanelRef.value.refreshFiles()
  }
}

// 图标组件
const ArrowLeftIcon = () => h('svg', { 
  viewBox: '0 0 24 24', 
  fill: 'currentColor',
  style: 'width: 1em; height: 1em;'
}, [
  h('path', { d: 'M20 11H7.83l5.59-5.59L12 4l-8 8 8 8 1.41-1.41L7.83 13H20v-2z' })
])

// 分割器拖拽逻辑
const startResize = (e: MouseEvent | TouchEvent) => {
  isResizing.value = true
  e.preventDefault()
  
  const handleMouseMove = (e: MouseEvent | TouchEvent) => {
    if (!isResizing.value) return
    
    const clientX = 'touches' in e ? e.touches[0].clientX : e.clientX
    const containerWidth = window.innerWidth
    const newLeftWidth = (clientX / containerWidth) * 100
    
    // 限制最小和最大宽度
    const minWidth = 20
    const maxWidth = 80
    
    if (newLeftWidth >= minWidth && newLeftWidth <= maxWidth) {
      leftWidth.value = newLeftWidth
      rightWidth.value = 100 - newLeftWidth
    }
  }
  
  const handleMouseUp = () => {
    isResizing.value = false
    document.removeEventListener('mousemove', handleMouseMove)
    document.removeEventListener('mouseup', handleMouseUp)
    document.removeEventListener('touchmove', handleMouseMove)
    document.removeEventListener('touchend', handleMouseUp)
  }
  
  document.addEventListener('mousemove', handleMouseMove)
  document.addEventListener('mouseup', handleMouseUp)
  document.addEventListener('touchmove', handleMouseMove)
  document.addEventListener('touchend', handleMouseUp)
}

// 方法
const goBack = () => {
  // 返回 dashboard 面板页面
  router.push('/dashboard')
}

const loadProject = async () => {
  if (!projectGuid.value) return
  
  loading.value = true
  try {
    const projectData = await projectStore.getProject(projectGuid.value)
    if (projectData) {
      project.value = projectData
    } else {
      // 项目不存在，跳转到仪表板
      router.push('/dashboard')
    }
  } catch (error) {
    console.error(t('project.loadProjectFailed'), error)
    // 跳转到仪表板
    router.push('/dashboard')
  } finally {
    loading.value = false
  }
}

// 生命周期
onMounted(() => {
  loadProject()
})

onUnmounted(() => {
  // 清理事件监听器
  document.removeEventListener('mousemove', () => {})
  document.removeEventListener('mouseup', () => {})
  document.removeEventListener('touchmove', () => {})
  document.removeEventListener('touchend', () => {})
})
</script>

<style scoped>
.project-edit-page {
  height: 100vh;
  display: flex;
  flex-direction: column;
  background: #f8fafc;
  overflow: hidden;
}

/* 左侧头部 */
.left-header {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 16px 20px;
  background: white;
  border-bottom: 1px solid #e2e8f0;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  z-index: 10;
}

.project-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.project-title {
  font-size: 18px;
  font-weight: 600;
  color: #1e293b;
  margin: 0;
}

.back-button {
  color: #64748b;
  font-size: 14px;
  height: var(--height-sm);
}

.back-button:hover {
  color: #334155;
}

.conversation-wrapper {
  flex: 1;
  overflow: hidden;
}

/* 主内容区域 - 分屏布局 */
.main-content {
  flex: 1;
  overflow: hidden;
}

.split-container {
  display: flex;
  height: 100%;
  position: relative;
}

/* 左侧面板 */
.left-panel {
  background: white;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

/* 右侧面板 */
.right-panel {
  background: #f8fafc;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

/* 分割器 */
.splitter {
  width: 8px;
  background: #e2e8f0;
  cursor: col-resize;
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background-color 0.2s ease;
}

.splitter:hover {
  background: #cbd5e1;
}

.splitter-handle {
  width: 4px;
  height: 40px;
  background: #94a3b8;
  border-radius: 2px;
  opacity: 0.6;
  transition: opacity 0.2s ease;
}

.splitter:hover .splitter-handle {
  opacity: 1;
}

/* 拖拽时的样式 */
.splitter:active {
  background: #3b82f6;
}

.splitter:active .splitter-handle {
  background: white;
  opacity: 1;
}

/* 响应式设计 */
@media (max-width: 1024px) {
  .split-container {
    flex-direction: column;
  }
  
  .left-panel,
  .right-panel {
    width: 100% !important;
    height: 50%;
  }
  
  .splitter {
    width: 100%;
    height: 8px;
    cursor: row-resize;
  }
  
  .splitter-handle {
    width: 40px;
    height: 4px;
  }
}

@media (max-width: 768px) {
  .top-navbar {
    padding: 8px 16px;
  }
  
  .navbar-left {
    gap: 12px;
  }
  
  .project-title {
    font-size: 16px;
  }
  
  .left-panel,
  .right-panel {
    height: 50%;
  }
}

@media (max-width: 480px) {
  .top-navbar {
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
    padding: 12px 16px;
  }
  
  .navbar-left {
    width: 100%;
    justify-content: space-between;
  }
}
</style>