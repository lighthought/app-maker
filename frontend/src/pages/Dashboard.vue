<template>
  <PageLayout>
    <div class="dashboard">
    <!-- 页面头部 -->
    <div class="dashboard-header">
      <div class="header-content">
        <div class="welcome-section">
          <h1>{{ t('dashboard.welcomeBack', { name: userStore.user?.username || userStore.user?.name || t('common.user') }) }}</h1>
          <p>{{ t('dashboard.todayStats', { date: currentDate, count: totalProjects }) }}</p>
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
            {{ t('buttons.createProject') }}
          </n-button>
        </div>
      </div>
    </div>

    <!-- 统计卡片 -->
    <div class="stats-grid">
      <n-card class="stat-card">
        <n-statistic
          :label="t('dashboard.totalProjects')"
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
          :label="t('common.inProgress')"
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
          :label="t('common.completed')"
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
          :label="t('dashboard.newThisMonth')"
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
          <h2>{{ t('dashboard.myProjects') }}</h2>
          <div class="filter-controls">
            <n-input
              v-model:value="searchKeyword"
              :placeholder="t('project.searchProjects')"
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
              :placeholder="t('dashboard.statusFilter')"
              class="status-filter"
            />
          </div>
        </div>

        <!-- 项目列表 -->
        <div class="projects-grid" v-if="filteredProjects.length > 0">
          <n-card
            v-for="project in filteredProjects"
            :key="project.guid"
            class="project-card"
            :class="{ 'project-card--active': currentProject?.guid === project.guid }"
            :style="{ '--double-click-text': `'${t('dashboard.doubleClickEdit')}'` }"
            @click="selectProject(project)"
            @dblclick="editProject(project.guid)"
          >
            <div class="project-header">
              <h3>{{ project.name }}</h3>
              <n-tag :type="getStatusType(project.status)" size="small">
                {{ getStatusText(project.status) }}
              </n-tag>
            </div>
            
            <p 
              class="project-description" 
              :title="project.description"
            >
              {{ project.description }}
            </p>
            
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
              <span class="created-time">{{ formatDateShort(project.created_at) }}</span>
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
          <h3>{{ t('dashboard.noProjects') }}</h3>
          <p>{{ t('dashboard.noProjectsDesc') }}</p>
          <n-button
            type="primary"
            size="large"
            @click="createNewProject"
            class="create-first-project-btn"
          >
            <template #icon>
              <n-icon><AddIcon /></n-icon>
            </template>
            {{ t('dashboard.createFirstProject') }}
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
        <!-- 系统状态 -->
        <n-card class="system-status-card">
          <template #header>
            <h3>{{ t('dashboard.systemStatus') }}</h3>
          </template>
          
          <div class="status-list">
            <div class="status-item">
              <n-icon 
                size="16" 
                :color="backendStatus === 'ok' ? '#38A169' : backendStatus === 'error' ? '#E53E3E' : '#D69E2E'"
              >
                <CheckIcon />
              </n-icon>
              <span>{{ t('dashboard.backendService') }} {{ backendStatus === 'ok' ? t('dashboard.normal') : backendStatus === 'error' ? t('dashboard.abnormal') : t('dashboard.checking') }}</span>
              <span v-if="backendVersion" class="version-info">v{{ backendVersion }}</span>
            </div>
            <div class="status-item">
              <n-icon size="16" color="#38A169">
                <CheckIcon />
              </n-icon>
              <span>{{ t('dashboard.database') }} {{ t('dashboard.normal') }}</span>
            </div>
            <div class="status-item">
              <n-icon size="16" color="#38A169">
                <CheckIcon />
              </n-icon>
              <span>{{ t('dashboard.agentOnline') }}</span>
            </div>
          </div>
        </n-card>

        <!-- 当前项目详情 -->
        <n-card v-if="currentProject" class="current-project-card">
          <template #header>
            <div class="card-header">
              <h3>{{ t('dashboard.currentProject') }}</h3>
              <n-button size="tiny" @click="currentProject = null">
                <n-icon><CloseIcon /></n-icon>
              </n-button>
            </div>
          </template>
          
          <div class="project-detail">
            <h4>{{ currentProject.name }}</h4>
            <p 
              class="project-description-short" 
              :title="currentProject.description"
            >
              {{ truncateDescription(currentProject.description) }}
            </p>
            
            <div class="project-stats">
              <div class="stat-item">
                <span class="label">{{ t('common.status') }}</span>
                <n-tag :type="getStatusType(currentProject.status)">
                  {{ getStatusText(currentProject.status) }}
                </n-tag>
              </div>
              <div class="stat-item">
                <span class="label">{{ t('dashboard.progress') }}</span>
                <span class="value">{{ getProjectProgress(currentProject) }}%</span>
              </div>
              <div class="stat-item">
                <span class="label">{{ t('dashboard.createdAt') }}</span>
                <span class="value">{{ formatDateTime(currentProject.created_at) }}</span>
              </div>
            </div>
            
            <div class="project-actions">
              <n-button type="primary" @click="editProject(currentProject.guid)">
                {{ t('dashboard.actionEdit') }}
              </n-button>
              <n-button @click="openProjectSettings(currentProject)">
                {{ t('project.projectSettings') }}
              </n-button>
              <n-button 
                v-if="currentProject.status !== 'pending'"
                @click="previewProject(currentProject.guid)"
              >
                {{ t('dashboard.actionPreview') }}
              </n-button>
              <n-button @click="downloadProject(currentProject.guid)">
                {{ t('dashboard.actionDownload') }}
              </n-button>
              <n-button type="error" @click="handleDeleteProject(currentProject.guid)">
                {{ t('dashboard.actionDelete') }}
              </n-button>
            </div>
          </div>
        </n-card>
      </div>
    </div>
    
    <!-- 任务轮询弹窗 -->
    <TaskProgressModal
      v-model:show="showTaskModal"
      :task-id="currentTaskId"
      :project-guid="currentDownloadProjectGuid"
      @retry="handleTaskRetry"
    />
    
    <!-- 项目设置弹窗 -->
    <ProjectSettingsModal
      v-model:show="showProjectSettings"
      :project="selectedProject"
      @saved="handleProjectSettingsSaved"
    />
    </div>
  </PageLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useUserStore } from '@/stores/user'
import { useProjectStore } from '@/stores/project'
import { useFilesStore } from '@/stores/file'
import TaskProgressModal from '@/components/TaskProgressModal.vue'
import ProjectSettingsModal from '@/components/ProjectSettingsModal.vue'
import { formatDateTime, formatDateShort } from '@/utils/time'
import { httpService } from '@/utils/http'
import {
  NButton, NIcon, NCard, NStatistic, NInput, NSelect, NTag, NProgress, NPagination
} from 'naive-ui'
import PageLayout from '@/components/layout/PageLayout.vue'
import type { Project, ProjectListRequest } from '@/types/project'
// 导入图标
import {
  SettingsIcon,
  AddIcon,
  FolderIcon,
  ClockIcon,
  CheckIcon,
  TrendingUpIcon,
  SearchIcon,
  CloseIcon,
  EmptyIcon
} from '@/components/icon'

const router = useRouter()
const { t, locale } = useI18n()
const userStore = useUserStore()
const projectStore = useProjectStore()
const fileStore = useFilesStore()

// 响应式数据
const searchKeyword = ref('')
const statusFilter = ref('')
const currentPage = ref(1)
const pageSize = ref(10)
const currentProject = ref<Project | null>(null)
const updateInterval = ref<number | null>(null)
const backendStatus = ref<'ok' | 'error' | 'checking'>('checking')
const backendVersion = ref('')

// 统计数据（全部项目，不受过滤影响）
const allProjectsStats = ref({
  total: 0,
  inProgress: 0,
  completed: 0,
  newThisMonth: 0
})

// 任务轮询相关状态
const showTaskModal = ref(false)
const currentTaskId = ref('')
const currentDownloadProjectGuid = ref('')

// 项目设置模态框
const showProjectSettings = ref(false)
const selectedProject = ref<Project | undefined>(undefined)

// 状态选项
const statusOptions = computed(() => [
  { label: t('dashboard.allStatus'), value: '' },
  { label: t('common.draft'), value: 'pending' },
  { label: t('common.inProgress'), value: 'in_progress' },
  { label: t('common.completed'), value: 'done' },
  { label: t('common.failed'), value: 'failed' }
])

// 计算属性
const currentDate = computed(() => {
  const currentLocale = locale.value === 'zh' ? 'zh-CN' : 'en-US'
  return new Date().toLocaleDateString(currentLocale, {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
    weekday: 'long'
  })
})

// 统计数据使用独立的 allProjectsStats，不受过滤影响
const totalProjects = computed(() => allProjectsStats.value.total)

const inProgressProjects = computed(() => allProjectsStats.value.inProgress)

const completedProjects = computed(() => allProjectsStats.value.completed)

const newThisMonth = computed(() => allProjectsStats.value.newThisMonth)

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

const previewProject = (projectGuid: string) => {
  // 跳转到编辑界面，URL hash 传递预览标识
  router.push(`/project/${projectGuid}#preview`)
}

const editProject = (projectGuid: string) => {
  router.push(`/project/${projectGuid}`)
}

const openProjectSettings = (project: Project) => {
  selectedProject.value = project
  showProjectSettings.value = true
}

const handleProjectSettingsSaved = () => {
  // 重新加载项目列表
  fetchProjectsWithFilters()
}

const downloadProject = async (projectGuid: string) => {
  try {
    const taskId = await projectStore.downloadProject(projectGuid)
    if (taskId) {
      // 显示任务轮询弹窗
      currentTaskId.value = taskId
      currentDownloadProjectGuid.value = projectGuid
      showTaskModal.value = true
    }
  } catch (error) {
    console.error('下载项目失败:', error)
  }
}

// 处理任务重试
const handleTaskRetry = async (taskId: string) => {
  try {
    const newTaskId = await projectStore.downloadProject(currentDownloadProjectGuid.value)
    if (newTaskId) {
      currentTaskId.value = newTaskId
    }
  } catch (error) {
    console.error('重试下载项目失败:', error)
  }
}

const handleDeleteProject = async (projectGuid: string) => {
  try {
    await projectStore.deleteProject(projectGuid)
    // 如果删除的是当前选中的项目，清空选中状态
    if (currentProject.value?.guid === projectGuid) {
      currentProject.value = null
    }
  } catch (error) {
    console.error('删除项目失败:', error)
  }
}

// 健康检查方法
const checkBackendHealth = async () => {
  try {
    backendStatus.value = 'checking'
    const healthData = await httpService.healthCheck()
    backendStatus.value = 'ok'
    backendVersion.value = healthData.version
    console.log('后端健康检查成功:', healthData)
  } catch (error) {
    backendStatus.value = 'error'
    console.error('后端健康检查失败:', error)
  }
}

const getStatusType = (status: string): 'default' | 'error' | 'warning' | 'success' | 'primary' | 'info' => {
  const statusMap: Record<string, 'default' | 'error' | 'warning' | 'success' | 'primary' | 'info'> = {
    draft: 'default',
    in_progress: 'warning',
    done: 'success',
    failed: 'error'
  }
  return statusMap[status] || 'default'
}

const getStatusText = (status: string) => {
  const statusMap: Record<string, string> = {
    draft: t('common.draft'),
    in_progress: t('common.inProgress'),
    done: t('common.completed'),
    failed: t('common.failed')
  }
  return statusMap[status] || status
}

const getProjectProgress = (project: Project) => {
  // 根据项目状态计算进度
  const progressMap: Record<string, number> = {
    draft: 10,
    in_progress: 60,
    done: 100,
    failed: 0
  }
  return progressMap[project.status] || 0
}

const getProgressColor = (status: string) => {
  const colorMap: Record<string, string> = {
    draft: '#A0AEC0',
    in_progress: '#D69E2E',
    done: '#38A169',
    failed: '#E53E3E'
  }
  return colorMap[status] || '#A0AEC0'
}

const truncateDescription = (description: string) => {
  if (!description) return ''
  return description.length > 100 ? description.substring(0, 100) + '...' : description
}

// 实时更新
const startRealTimeUpdates = () => {
  updateInterval.value = window.setInterval(async () => {
    try {
      // 更新统计数据（全部项目）
      await fetchAllProjectsStats()
      
      // 使用实际的API获取最新数据（带过滤）
      const params: ProjectListRequest = {
        page: currentPage.value,
        pageSize: pageSize.value,
        status: statusFilter.value || undefined,
        search: searchKeyword.value || undefined
      }
      await projectStore.fetchProjects(params)
      console.log('实时更新项目数据完成')

      // 健康检查
      await checkBackendHealth()
    } catch (error) {
      console.error('实时更新项目数据失败:', error)
    }
  }, 45000) // 45秒更新一次
}

const stopRealTimeUpdates = () => {
  if (updateInterval.value) {
    clearInterval(updateInterval.value)
    updateInterval.value = null
  }
}

// 生命周期
onMounted(async () => {
  // 初始健康检查
  await checkBackendHealth()
  
  // 获取统计数据（全部项目）
  await fetchAllProjectsStats()
  
  // 加载项目数据（带过滤）
  fetchProjectsWithFilters()
  
  // 启动实时更新
  startRealTimeUpdates()
})

// 获取全部项目的统计数据（不受过滤影响）
const fetchAllProjectsStats = async () => {
  try {
    // 获取所有项目（不分页，不过滤）用于统计
    const response = await httpService.get<{
      code: number
      message: string
      total: number
      data: Project[]
    }>('/projects', { 
      params: { 
        page: 1, 
        pageSize: 9999  // 获取所有项目
      } 
    })

    if (response.code === 0 && response.data) {
      const allProjects = response.data || []
      
      // 计算统计数据
      allProjectsStats.value = {
        total: allProjects.length,
        inProgress: allProjects.filter(p => p.status === 'in_progress').length,
        completed: allProjects.filter(p => p.status === 'done').length,
        newThisMonth: (() => {
          const now = new Date()
          const thisMonth = new Date(now.getFullYear(), now.getMonth(), 1)
          return allProjects.filter(p => new Date(p.created_at) >= thisMonth).length
        })()
      }
    }
  } catch (error) {
    console.error('获取统计数据失败:', error)
  }
}

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
/* 布局相关样式 */
.content-wrapper {
  padding: var(--spacing-lg);
  min-height: calc(100vh - 64px);
}

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
  position: relative;
}

.project-card:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-md);
}

.project-card:hover::after {
  content: var(--double-click-text);
  position: absolute;
  top: 8px;
  right: 8px;
  background: rgba(0, 0, 0, 0.7);
  color: white;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 0.7rem;
  opacity: 0;
  animation: fadeIn 0.3s ease forwards;
}

@keyframes fadeIn {
  to {
    opacity: 1;
  }
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
  cursor: help;
}

.project-description:hover {
  background-color: rgba(0, 0, 0, 0.02);
  border-radius: 4px;
  padding: 2px;
  margin: -2px -2px calc(var(--spacing-md) - 2px) -2px;
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

.project-description-short {
  margin: 0 0 var(--spacing-md) 0;
  color: var(--text-secondary);
  font-size: 0.9rem;
  line-height: 1.4;
  max-height: 3.6em;
  overflow: hidden;
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  cursor: help;
  position: relative;
}

.project-description-short:hover {
  background-color: rgba(0, 0, 0, 0.02);
  border-radius: 4px;
  padding: 4px;
  margin: -4px -4px calc(var(--spacing-md) - 4px) -4px;
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
  flex-wrap: wrap;
  gap: var(--spacing-sm);
}

.project-actions .n-button {
  /* 前4个按钮每行2个，各占约50%宽度 */
  flex: 0 0 calc(50% - var(--spacing-sm) / 2);
  min-width: 60px;
}

/* 删除按钮单独占一行 */
.project-actions .n-button:last-child {
  flex: 0 0 100%;
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

.version-info {
  margin-left: auto;
  font-size: 0.8rem;
  color: var(--text-disabled);
  font-weight: 500;
}

/* 响应式设计 */
@media (max-width: 1024px) {
  .dashboard {
    padding: var(--spacing-md);
    min-height: auto;
  }
  
  .dashboard-content {
    grid-template-columns: 1fr;
    gap: var(--spacing-lg);
  }
  
  .sidebar-panel {
    order: -1;
    display: grid;
    grid-template-columns: 1fr;
    gap: var(--spacing-md);
  }
  
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
    gap: var(--spacing-md);
  }
  
  .projects-grid {
    grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
  }
}

@media (max-width: 768px) {
  .dashboard {
    padding: var(--spacing-sm);
  }
  
  .dashboard-header {
    margin-bottom: var(--spacing-md);
  }
  
  .header-content {
    flex-direction: column;
    gap: var(--spacing-md);
    padding: var(--spacing-md);
  }
  
  .welcome-section {
    text-align: center;
    width: 100%;
  }
  
  .welcome-section h1 {
    font-size: 1.25rem;
  }
  
  .welcome-section p {
    font-size: 0.85rem;
  }
  
  .header-actions {
    width: 100%;
  }
  
  .create-btn {
    width: 100%;
  }
  
  .section-header {
    flex-direction: column;
    align-items: stretch;
    gap: var(--spacing-sm);
  }
  
  .section-header h2 {
    font-size: 1.1rem;
  }
  
  .filter-controls {
    flex-direction: column;
    width: 100%;
  }
  
  .search-input,
  .status-filter {
    width: 100%;
  }
  
  .projects-section {
    padding: var(--spacing-md);
  }
  
  .projects-grid {
    grid-template-columns: 1fr;
    gap: var(--spacing-md);
  }
  
  .project-card {
    font-size: 0.9rem;
  }
  
  .project-header h3 {
    font-size: 1rem;
  }
  
  /* 移动端隐藏双击提示 */
  .project-card:hover::after {
    display: none;
  }
  
  .pagination-wrapper {
    overflow-x: auto;
    padding: var(--spacing-sm) 0;
  }
  
  .pagination-wrapper :deep(.n-pagination) {
    flex-wrap: nowrap;
    justify-content: flex-start;
  }
  
  /* 平板模式下的按钮优化 */
  .project-actions .n-button {
    flex: 0 0 calc(50% - var(--spacing-sm) / 2);  /* 保持每行2个 */
    min-height: 40px;  /* ✨ 平板模式最小高度 40px */
    height: auto;  /* ✨ 自动高度，适应内容 */
    padding: 10px 16px;  /* ✨ 增加内边距 */
  }
  
  .project-actions .n-button:last-child {
    flex: 0 0 100%;  /* 删除按钮单独一行 */
  }
}

@media (max-width: 480px) {
  .dashboard {
    padding: 0;
  }
  
  .dashboard-header {
    margin-bottom: var(--spacing-sm);
  }
  
  .header-content {
    border-radius: 0;
    padding: var(--spacing-sm);
  }
  
  .welcome-section h1 {
    font-size: 1.1rem;
    margin-bottom: 4px;
  }
  
  .welcome-section p {
    font-size: 0.75rem;
    line-height: 1.4;
  }
  
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);  /* ✨ 2列布局 */
    gap: var(--spacing-sm);
    margin-bottom: var(--spacing-md);
  }
  
  .stat-card {
    padding: var(--spacing-sm);
  }
  
  .stat-card :deep(.n-statistic-label) {
    font-size: 0.8rem;
  }
  
  .stat-card :deep(.n-statistic-value) {
    font-size: 1.5rem !important;
  }
  
  .projects-section {
    border-radius: 0;
    padding: var(--spacing-sm);
  }
  
  .section-header {
    margin-bottom: var(--spacing-sm);
  }
  
  .projects-grid {
    gap: var(--spacing-sm);
  }
  
  .project-card {
    padding: var(--spacing-sm);
  }
  
  .project-description {
    font-size: 0.85rem;
    -webkit-line-clamp: 1;
  }
  
  .project-meta {
    flex-direction: column;
    align-items: flex-start;
    gap: 4px;
  }
  
  .created-time {
    font-size: 0.7rem;
  }
  
  .sidebar-panel {
    padding: 0 var(--spacing-sm);
  }
  
  .current-project-card,
  .system-status-card {
    border-radius: var(--border-radius-md);
  }
  
  .card-header h3 {
    font-size: 1rem;
  }
  
  .project-detail h4 {
    font-size: 1rem;
  }
  
  .project-detail p {
    font-size: 0.85rem;
  }
  
  .project-actions {
    gap: var(--spacing-sm);  /* 增加按钮间距 */
  }
  
  .project-actions .n-button {
    flex: 0 0 calc(50% - var(--spacing-sm) / 2);  /* 每行2个 */
    min-height: 44px;  /* ✨ 移动端最小触摸高度 44px */
    height: 44px;  /* ✨ 固定高度，易于点击 */
    font-size: 0.9rem;  /* ✨ 适当字体大小 */
  }
  
  .project-actions .n-button:last-child {
    flex: 0 0 100%;  /* 删除按钮单独一行 */
  }
  
  .status-item {
    font-size: 0.85rem;
  }
  
  .empty-state {
    padding: var(--spacing-lg) var(--spacing-sm);
    margin: var(--spacing-sm) 0;
  }
  
  .empty-state h3 {
    font-size: 1.1rem;
  }
  
  .empty-state p {
    font-size: 0.9rem;
  }
  
  .create-first-project-btn {
    width: 100%;
  }
  
  /* 分页优化 */
  .pagination-wrapper :deep(.n-pagination) {
    font-size: 0.75rem;
  }
  
  .pagination-wrapper :deep(.n-pagination-item) {
    min-width: 28px;
    height: 28px;
  }
}
</style>