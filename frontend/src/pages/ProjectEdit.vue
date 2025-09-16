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
              返回
            </n-button>
            <div class="project-info">
              <h2 class="project-title">{{ project?.name || '项目编辑' }}</h2>
              <n-tag :type="getStatusType(project?.status)" size="small">
                {{ getStatusText(project?.status) }}
              </n-tag>
            </div>
          </div>
          
          <!-- 对话容器 -->
          <div class="conversation-wrapper">
            <ConversationContainer
              :project-id="projectId"
              :requirements="project?.requirements || ''"
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
          <ProjectPanel :project="project" />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, h, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { NButton, NIcon, NTag } from 'naive-ui'
import ConversationContainer from '@/components/ConversationContainer.vue'
import ProjectPanel from '@/components/ProjectPanel.vue'
import { useProjectStore } from '@/stores/project'
import type { Project } from '@/types/project'

const route = useRoute()
const router = useRouter()
const projectStore = useProjectStore()

// 响应式数据
const project = ref<Project | undefined>(undefined)
const loading = ref(false)

// 分割器相关
const leftWidth = ref(50)
const rightWidth = ref(50)
const isResizing = ref(false)

// 计算属性
const projectId = computed(() => route.params.id as string)

// 获取状态类型
const getStatusType = (status?: string): 'default' | 'primary' | 'info' | 'success' | 'warning' | 'error' => {
  const statusMap: Record<string, 'default' | 'primary' | 'info' | 'success' | 'warning' | 'error'> = {
    draft: 'default',
    in_progress: 'primary',
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
  if (!projectId.value) return
  
  loading.value = true
  try {
    const projectData = await projectStore.getProject(projectId.value)
    if (projectData) {
      project.value = projectData
    } else {
      // 项目不存在，跳转到仪表板
      router.push('/dashboard')
    }
  } catch (error) {
    console.error('加载项目失败:', error)
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