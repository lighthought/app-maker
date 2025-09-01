<template>
  <div class="dashboard">
    <!-- 页面头部 -->
    <div class="dashboard-header">
      <div class="header-content">
        <div class="welcome-section">
          <h1>欢迎回来，{{ userStore.user?.username || userStore.user?.name || '用户' }}</h1>
          <p>今天是 {{ currentDate }}，您有 {{ totalProjects }} 个项目</p>
        </div>
        <div class="header-actions">
          <n-button
            type="primary"
            size="large"
            @click="createNewProject"
            class="create-btn"
          >
            <template #icon>
              <n-icon><AddIcon /></n-icon>
            </template>
            创建新项目
          </n-button>
        </div>
      </div>
    </div>

    <!-- 统计卡片 -->
    <div class="stats-grid">
      <n-card class="stat-card">
        <n-statistic
          label="总项目数"
          :value="totalProjects"
          :value-style="{ color: '#3182CE' }"
        >
          <template #prefix>
            <n-icon size="24" color="#3182CE">
              <FolderIcon />
            </n-icon>
          </template>
        </n-statistic>
      </n-card>

      <n-card class="stat-card">
        <n-statistic
          label="进行中"
          :value="inProgressProjects"
          :value-style="{ color: '#D69E2E' }"
        >
          <template #prefix>
            <n-icon size="24" color="#D69E2E">
              <ClockIcon />
            </n-icon>
          </template>
        </n-statistic>
      </n-card>

      <n-card class="stat-card">
        <n-statistic
          label="已完成"
          :value="completedProjects"
          :value-style="{ color: '#38A169' }"
        >
          <template #prefix>
            <n-icon size="24" color="#38A169">
              <CheckIcon />
            </n-icon>
          </template>
        </n-statistic>
      </n-card>

      <n-card class="stat-card">
        <n-statistic
          label="本月新增"
          :value="newThisMonth"
          :value-style="{ color: '#E53E3E' }"
        >
          <template #prefix>
            <n-icon size="24" color="#E53E3E">
              <TrendingUpIcon />
            </n-icon>
          </template>
        </n-statistic>
      </n-card>
    </div>

    <!-- 主要内容区域 -->
    <div class="dashboard-content">
      <!-- 左侧项目列表 -->
      <div class="projects-section">
        <div class="section-header">
          <h2>我的项目</h2>
          <div class="filter-controls">
            <n-input
              v-model:value="searchKeyword"
              placeholder="搜索项目..."
              clearable
              class="search-input"
            >
              <template #prefix>
                <n-icon><SearchIcon /></n-icon>
              </template>
            </n-input>
            <n-select
              v-model:value="statusFilter"
              :options="statusOptions"
              placeholder="状态筛选"
              class="status-filter"
            />
          </div>
        </div>

        <!-- 项目列表 -->
        <div class="projects-grid" v-if="filteredProjects.length > 0">
          <n-card
            v-for="project in filteredProjects"
            :key="project.id"
            class="project-card"
            :class="{ 'project-card--active': currentProject?.id === project.id }"
            @click="selectProject(project)"
          >
            <div class="project-header">
              <h3>{{ project.name }}</h3>
              <n-tag :type="getStatusType(project.status)" size="small">
                {{ getStatusText(project.status) }}
              </n-tag>
            </div>
            
            <p class="project-description">{{ project.description }}</p>
            
            <div class="project-progress">
              <n-progress
                :percentage="getProjectProgress(project)"
                :color="getProgressColor(project.status)"
                :show-indicator="false"
                size="small"
              />
              <span class="progress-text">{{ getProjectProgress(project) }}%</span>
            </div>
            
            <div class="project-meta">
              <span class="created-time">{{ formatDate(project.createdAt) }}</span>
              <div class="project-actions">
                <n-button
                  size="tiny"
                  @click.stop="previewProject(project.id)"
                >
                  预览
                </n-button>
                <n-button
                  size="tiny"
                  type="primary"
                  @click.stop="editProject(project.id)"
                >
                  编辑
                </n-button>
              </div>
            </div>
          </n-card>
        </div>

        <!-- 空状态 -->
        <div v-else class="empty-state">
          <div class="empty-icon">
            <n-icon size="64" color="#CBD5E0">
              <EmptyIcon />
            </n-icon>
          </div>
          <h3>暂无项目</h3>
          <p>您还没有创建任何项目，开始您的第一个项目吧！</p>
          <n-button
            type="primary"
            size="large"
            @click="createNewProject"
            class="create-first-project-btn"
          >
            <template #icon>
              <n-icon><AddIcon /></n-icon>
            </template>
            创建第一个项目
          </n-button>
        </div>

        <!-- 分页 -->
        <div class="pagination-wrapper">
          <n-pagination
            v-model:page="currentPage"
            v-model:page-size="pageSize"
            :item-count="projectStore.pagination.total"
            :page-sizes="[5, 10, 20]"
            show-size-picker
            show-quick-jumper
          />
        </div>
      </div>

      <!-- 右侧面板 -->
      <div class="sidebar-panel">
        <!-- 当前项目详情 -->
        <n-card v-if="currentProject" class="current-project-card">
          <template #header>
            <div class="card-header">
              <h3>当前项目</h3>
              <n-button size="tiny" @click="currentProject = null">
                <n-icon><CloseIcon /></n-icon>
              </n-button>
            </div>
          </template>
          
          <div class="project-detail">
            <h4>{{ currentProject.name }}</h4>
            <p>{{ currentProject.description }}</p>
            
            <div class="project-stats">
              <div class="stat-item">
                <span class="label">状态</span>
                <n-tag :type="getStatusType(currentProject.status)">
                  {{ getStatusText(currentProject.status) }}
                </n-tag>
              </div>
              <div class="stat-item">
                <span class="label">进度</span>
                <span class="value">{{ getProjectProgress(currentProject) }}%</span>
              </div>
              <div class="stat-item">
                <span class="label">创建时间</span>
                <span class="value">{{ formatDate(currentProject.createdAt) }}</span>
              </div>
            </div>
            
            <div class="project-actions">
              <n-button type="primary" @click="editProject(currentProject.id)">
                继续编辑
              </n-button>
              <n-button @click="previewProject(currentProject.id)">
                预览项目
              </n-button>
            </div>
          </div>
        </n-card>

        <!-- 系统状态 -->
        <n-card class="system-status-card">
          <template #header>
            <h3>系统状态</h3>
          </template>
          
          <div class="status-list">
            <div class="status-item">
              <n-icon size="16" color="#38A169">
                <CheckIcon />
              </n-icon>
              <span>后端服务正常</span>
            </div>
            <div class="status-item">
              <n-icon size="16" color="#38A169">
                <CheckIcon />
              </n-icon>
              <span>数据库连接正常</span>
            </div>
            <div class="status-item">
              <n-icon size="16" color="#38A169">
                <CheckIcon />
              </n-icon>
              <span>AI Agent 在线</span>
            </div>
          </div>
        </n-card>

        <!-- 快速操作 -->
        <n-card class="quick-actions-card">
          <template #header>
            <h3>快速操作</h3>
          </template>
          
          <div class="quick-actions">
            <n-button
              v-for="action in quickActions"
              :key="action.key"
              :type="action.type"
              size="small"
              @click="handleQuickAction(action.key)"
              class="quick-action-btn"
            >
              <template #icon>
                <n-icon><component :is="action.icon" /></n-icon>
              </template>
              {{ action.label }}
            </n-button>
          </div>
        </n-card>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch, h } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { useProjectStore } from '@/stores/project'
import {
  NButton, NIcon, NCard, NStatistic, NInput, NSelect, NTag, NProgress, NPagination
} from 'naive-ui'
import type { Project, ProjectListRequest } from '@/types/project'

// 图标组件 - 使用 SVG 图标替代 emoji
const AddIcon = () => h('svg', { 
  viewBox: '0 0 24 24', 
  fill: 'currentColor',
  style: 'width: 1em; height: 1em;'
}, [
  h('path', { d: 'M19 13h-6v6h-2v-6H5v-2h6V5h2v6h6v2z' })
])

const FolderIcon = () => h('svg', { 
  viewBox: '0 0 24 24', 
  fill: 'currentColor',
  style: 'width: 1em; height: 1em;'
}, [
  h('path', { d: 'M10 4H4c-1.1 0-1.99.9-1.99 2L2 18c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V8c0-1.1-.9-2-2-2h-8l-2-2z' })
])

const ClockIcon = () => h('svg', { 
  viewBox: '0 0 24 24', 
  fill: 'currentColor',
  style: 'width: 1em; height: 1em;'
}, [
  h('path', { d: 'M11.99 2C6.47 2 2 6.48 2 12s4.47 10 9.99 10C17.52 22 22 17.52 22 12S17.52 2 11.99 2zM12 20c-4.42 0-8-3.58-8-8s3.58-8 8-8 8 3.58 8 8-3.58 8-8 8z' }),
  h('path', { d: 'M12.5 7H11v6l5.25 3.15.75-1.23-4.5-2.67z' })
])

const CheckIcon = () => h('svg', { 
  viewBox: '0 0 24 24', 
  fill: 'currentColor',
  style: 'width: 1em; height: 1em;'
}, [
  h('path', { d: 'M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41z' })
])

const TrendingUpIcon = () => h('svg', { 
  viewBox: '0 0 24 24', 
  fill: 'currentColor',
  style: 'width: 1em; height: 1em;'
}, [
  h('path', { d: 'M16 6l2.29 2.29-4.88 4.88-4-4L2 16.59 3.41 18l6-6 4 4 6.3-6.29L22 12V6z' })
])

const SearchIcon = () => h('svg', { 
  viewBox: '0 0 24 24', 
  fill: 'currentColor',
  style: 'width: 1em; height: 1em;'
}, [
  h('path', { d: 'M15.5 14h-.79l-.28-.27C15.41 12.59 16 11.11 16 9.5 16 5.91 13.09 3 9.5 3S3 5.91 3 9.5 5.91 16 9.5 16c1.61 0 3.09-.59 4.23-1.57l.27.28v.79l5 4.99L20.49 19l-4.99-5zm-6 0C7.01 14 5 11.99 5 9.5S7.01 5 9.5 5 14 7.01 14 9.5 11.99 14 9.5 14z' })
])

const CloseIcon = () => h('svg', { 
  viewBox: '0 0 24 24', 
  fill: 'currentColor',
  style: 'width: 1em; height: 1em;'
}, [
  h('path', { d: 'M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z' })
])

const EmptyIcon = () => h('svg', { 
  viewBox: '0 0 24 24', 
  fill: 'currentColor',
  style: 'width: 1em; height: 1em;'
}, [
  h('path', { d: 'M20 6h-8l-2-2H4c-1.1 0-1.99.9-1.99 2L2 18c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V8c0-1.1-.9-2-2-2zm0 12H4V8h16v10z' })
])

const router = useRouter()
const userStore = useUserStore()
const projectStore = useProjectStore()

// 响应式数据
const searchKeyword = ref('')
const statusFilter = ref('')
const currentPage = ref(1)
const pageSize = ref(8)
const currentProject = ref<Project | null>(null)
const updateInterval = ref<number | null>(null)

// 状态选项
const statusOptions = [
  { label: '全部状态', value: '' },
  { label: '草稿', value: 'draft' },
  { label: '进行中', value: 'in_progress' },
  { label: '已完成', value: 'completed' },
  { label: '失败', value: 'failed' }
]

// 快速操作
const quickActions = [
  { key: 'create', label: '创建项目', icon: AddIcon, type: 'primary' as const },
  { key: 'import', label: '导入项目', icon: FolderIcon, type: 'default' as const },
  { key: 'export', label: '导出数据', icon: TrendingUpIcon, type: 'default' as const },
  { key: 'settings', label: '设置', icon: ClockIcon, type: 'default' as const }
]

// 计算属性
const currentDate = computed(() => {
  return new Date().toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
    weekday: 'long'
  })
})

const totalProjects = computed(() => projectStore.projects.length)

const inProgressProjects = computed(() => 
  projectStore.projects.filter(p => p.status === 'in_progress').length
)

const completedProjects = computed(() => 
  projectStore.projects.filter(p => p.status === 'completed').length
)

const newThisMonth = computed(() => {
  const now = new Date()
  const thisMonth = new Date(now.getFullYear(), now.getMonth(), 1)
  return projectStore.projects.filter(p => 
    new Date(p.createdAt) >= thisMonth
  ).length
})

const filteredProjects = computed(() => {
  return projectStore.projects
})

const totalFilteredProjects = computed(() => {
  return projectStore.pagination.total
})

// 方法
const createNewProject = () => {
  router.push('/create-project')
}

const selectProject = (project: Project) => {
  currentProject.value = project
}

const previewProject = (projectId: string) => {
  router.push(`/preview/${projectId}`)
}

const editProject = (projectId: string) => {
  router.push(`/project/${projectId}`)
}

const getStatusType = (status: string): 'default' | 'error' | 'warning' | 'success' | 'primary' | 'info' => {
  const statusMap: Record<string, 'default' | 'error' | 'warning' | 'success' | 'primary' | 'info'> = {
    draft: 'default',
    in_progress: 'warning',
    completed: 'success',
    failed: 'error'
  }
  return statusMap[status] || 'default'
}

const getStatusText = (status: string) => {
  const statusMap: Record<string, string> = {
    draft: '草稿',
    in_progress: '进行中',
    completed: '已完成',
    failed: '失败'
  }
  return statusMap[status] || status
}

const getProjectProgress = (project: Project) => {
  // 根据项目状态计算进度
  const progressMap: Record<string, number> = {
    draft: 10,
    in_progress: 60,
    completed: 100,
    failed: 0
  }
  return progressMap[project.status] || 0
}

const getProgressColor = (status: string) => {
  const colorMap: Record<string, string> = {
    draft: '#A0AEC0',
    in_progress: '#D69E2E',
    completed: '#38A169',
    failed: '#E53E3E'
  }
  return colorMap[status] || '#A0AEC0'
}

const formatDate = (date: string) => {
  return new Date(date).toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: 'short',
    day: 'numeric'
  })
}

const handleQuickAction = (action: string) => {
  switch (action) {
    case 'create':
      createNewProject()
      break
    case 'import':
      // TODO: 实现导入功能
      console.log('导入项目')
      break
    case 'export':
      // TODO: 实现导出功能
      console.log('导出数据')
      break
    case 'settings':
      // TODO: 跳转到设置页面
      console.log('设置')
      break
  }
}

// 实时更新
const startRealTimeUpdates = () => {
  updateInterval.value = window.setInterval(() => {
    // 这里可以调用 API 获取最新数据
    // 目前使用模拟数据，实际项目中应该调用 projectStore.fetchProjects()
    console.log('实时更新项目数据...')
  }, 30000) // 30秒更新一次
}

const stopRealTimeUpdates = () => {
  if (updateInterval.value) {
    clearInterval(updateInterval.value)
    updateInterval.value = null
  }
}

// 生命周期
onMounted(() => {
  // 加载项目数据
  fetchProjectsWithFilters()
  
  // 启动实时更新
  startRealTimeUpdates()
})

// 获取项目数据（带筛选和分页）
const fetchProjectsWithFilters = async () => {
  const params: ProjectListRequest = {
    page: currentPage.value,
    pageSize: pageSize.value,
    status: statusFilter.value || undefined,
    search: searchKeyword.value || undefined
  }
  await projectStore.fetchProjects(params)
}

// 监听筛选条件变化
watch([searchKeyword, statusFilter], () => {
  currentPage.value = 1 // 重置到第一页
  fetchProjectsWithFilters()
})

// 监听分页变化
watch([currentPage, pageSize], () => {
  fetchProjectsWithFilters()
})

onUnmounted(() => {
  stopRealTimeUpdates()
})
</script>

<style scoped>
.dashboard {
  padding: var(--spacing-lg);
  background: var(--background-color);
  min-height: 100vh;
}

/* 页面头部 */
.dashboard-header {
  margin-bottom: var(--spacing-xl);
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--spacing-lg);
  background: white;
  border-radius: var(--border-radius-lg);
  box-shadow: var(--shadow-sm);
}

.welcome-section h1 {
  margin: 0 0 var(--spacing-sm) 0;
  color: var(--primary-color);
  font-size: 1.5rem;
  font-weight: bold;
}

.welcome-section p {
  margin: 0;
  color: var(--text-secondary);
}

.create-btn {
  background: linear-gradient(135deg, var(--primary-color), var(--accent-color));
  border: none;
  font-weight: 600;
}

/* 统计卡片 */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: var(--spacing-lg);
  margin-bottom: var(--spacing-xl);
}

.stat-card {
  text-align: center;
  transition: transform 0.3s ease;
}

.stat-card:hover {
  transform: translateY(-4px);
}

/* 统计卡片图标样式 */
.stat-card .n-statistic .n-statistic-label {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--spacing-sm);
  margin-bottom: var(--spacing-sm);
}

.stat-card .n-statistic .n-statistic-value {
  font-size: 2rem;
  font-weight: bold;
  margin-top: var(--spacing-sm);
}

/* 主要内容区域 */
.dashboard-content {
  display: grid;
  grid-template-columns: 1fr 300px;
  gap: var(--spacing-xl);
}

/* 项目列表区域 */
.projects-section {
  background: white;
  border-radius: var(--border-radius-lg);
  padding: var(--spacing-lg);
  box-shadow: var(--shadow-sm);
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-lg);
  flex-wrap: wrap;
  gap: var(--spacing-md);
}

.section-header h2 {
  margin: 0;
  color: var(--primary-color);
  font-size: 1.25rem;
  font-weight: bold;
}

.filter-controls {
  display: flex;
  gap: var(--spacing-md);
  align-items: center;
}

.search-input {
  width: 200px;
}

.status-filter {
  width: 120px;
}

/* 项目网格 */
.projects-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: var(--spacing-lg);
  margin-bottom: var(--spacing-lg);
}

.project-card {
  cursor: pointer;
  transition: all 0.3s ease;
  border: 1px solid var(--border-color);
}

.project-card:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-md);
}

.project-card--active {
  border-color: var(--primary-color);
  box-shadow: 0 0 0 2px rgba(49, 130, 206, 0.2);
}

.project-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: var(--spacing-sm);
}

.project-header h3 {
  margin: 0;
  font-size: 1.1rem;
  color: var(--primary-color);
  flex: 1;
}

.project-description {
  margin: 0 0 var(--spacing-md) 0;
  color: var(--text-secondary);
  font-size: 0.9rem;
  line-height: 1.4;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.project-progress {
  margin-bottom: var(--spacing-md);
}

.progress-text {
  font-size: 0.8rem;
  color: var(--text-secondary);
  margin-left: var(--spacing-sm);
}

.project-meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.created-time {
  font-size: 0.8rem;
  color: var(--text-disabled);
}

.project-actions {
  display: flex;
  gap: var(--spacing-sm);
}

/* 空状态 */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: var(--spacing-xl) var(--spacing-lg);
  text-align: center;
  background: white;
  border-radius: var(--border-radius-lg);
  border: 2px dashed var(--border-color);
  margin: var(--spacing-lg) 0;
}

.empty-icon {
  margin-bottom: var(--spacing-lg);
}

.empty-state h3 {
  margin: 0 0 var(--spacing-sm) 0;
  color: var(--primary-color);
  font-size: 1.25rem;
  font-weight: bold;
}

.empty-state p {
  margin: 0 0 var(--spacing-lg) 0;
  color: var(--text-secondary);
  font-size: 1rem;
  line-height: 1.5;
}

.create-first-project-btn {
  background: linear-gradient(135deg, var(--primary-color), var(--accent-color));
  border: none;
  font-weight: 600;
  padding: var(--spacing-md) var(--spacing-lg);
}

.create-first-project-btn:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-md);
}

/* 分页 */
.pagination-wrapper {
  display: flex;
  justify-content: center;
  margin-top: var(--spacing-lg);
}

/* 右侧面板 */
.sidebar-panel {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-lg);
}

.current-project-card,
.system-status-card,
.quick-actions-card {
  background: white;
  border-radius: var(--border-radius-lg);
  box-shadow: var(--shadow-sm);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-header h3 {
  margin: 0;
  color: var(--primary-color);
  font-size: 1.1rem;
  font-weight: bold;
}

.project-detail h4 {
  margin: 0 0 var(--spacing-sm) 0;
  color: var(--primary-color);
  font-size: 1.1rem;
}

.project-detail p {
  margin: 0 0 var(--spacing-md) 0;
  color: var(--text-secondary);
  font-size: 0.9rem;
}

.project-stats {
  margin-bottom: var(--spacing-lg);
}

.stat-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-sm);
}

.stat-item .label {
  color: var(--text-secondary);
  font-size: 0.9rem;
}

.stat-item .value {
  color: var(--primary-color);
  font-weight: 500;
}

.project-actions {
  display: flex;
  gap: var(--spacing-sm);
}

.status-list {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-sm);
}

.status-item {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  font-size: 0.9rem;
  color: var(--text-secondary);
}

.quick-actions {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: var(--spacing-sm);
}

.quick-action-btn {
  width: 100%;
}

/* 响应式设计 */
@media (max-width: 1024px) {
  .dashboard-content {
    grid-template-columns: 1fr;
  }
  
  .sidebar-panel {
    order: -1;
  }
}

@media (max-width: 768px) {
  .dashboard {
    padding: var(--spacing-md);
  }
  
  .header-content {
    flex-direction: column;
    gap: var(--spacing-md);
    text-align: center;
  }
  
  .section-header {
    flex-direction: column;
    align-items: stretch;
  }
  
  .filter-controls {
    flex-direction: column;
  }
  
  .search-input,
  .status-filter {
    width: 100%;
  }
  
  .projects-grid {
    grid-template-columns: 1fr;
  }
  
  .quick-actions {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 480px) {
  .stats-grid {
    grid-template-columns: 1fr;
  }
  
  .project-meta {
    flex-direction: column;
    align-items: flex-start;
    gap: var(--spacing-sm);
  }
}
</style>