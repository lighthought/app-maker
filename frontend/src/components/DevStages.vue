<template>
  <div class="dev-stages">
    <div class="stages-header">
        <h3>{{ t('project.devProgress') }}</h3>
    </div>
    
    <div ref="stagesContainer" class="stages-container" :class="{ 'horizontal': layout === 'horizontal' }">
      <div
        v-for="(stage, index) in stages"
        :key="stage.id"
        class="stage-item"
        :class="[getStageClass(stage), { 'horizontal': layout === 'horizontal' }]"
      >
        <div class="stage-circle">
          <span class="stage-number">{{ index + 1 }}</span>
        </div>
        
        <div class="stage-content">
          <div class="stage-name">{{ getStageDisplayName(stage.name) }}</div>
          <div v-if="layout === 'vertical'" class="stage-description">{{ stage.description }}</div>
        </div>        
      </div>
    </div>
    
    <!-- 当前状态信息 -->
    <div v-if="currentStage" class="current-status">
      <!-- 失败状态时显示失败原因和重试按钮 -->
      <div v-if="currentStage.status === 'failed' && getFailedReason(currentStage)" class="failed-reason">
        <div class="failed-reason-content">
          <div class="failed-reason-text">失败原因：{{ getFailedReason(currentStage) }}</div>
          <n-button 
            v-if="currentStage.task_id"
            type="error" 
            size="small" 
            :loading="retrying"
            @click="handleRetry"
            class="retry-button"
          >
            <template #icon>
              <n-icon>
                <RefreshIcon />
              </n-icon>
            </template>
            重试
          </n-button>
        </div>
      </div>
      
      <!-- 其他状态显示正常状态信息 -->
      <div v-else class="status-info">
        <n-icon 
          size="16" 
          :color="getStatusColor(currentStage.status)"
        >
          <component :is="getStatusIcon(currentStage.status)" />
        </n-icon>
        <span class="status-text">{{ getStatusText(currentStage) }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import { NIcon, NButton, useMessage } from 'naive-ui'
import { useTaskStore } from '@/stores/task'
import type { DevStage } from '@/types/project'
// 导入图标
import { CheckIcon, ErrorIcon, ClockIcon, RefreshIcon } from '@/components/icon'

interface Props {
  stages: DevStage[]
  layout?: 'vertical' | 'horizontal'
}

const props = withDefaults(defineProps<Props>(), {
  layout: 'vertical'
})
const { t } = useI18n()
const message = useMessage()
const taskStore = useTaskStore()

// 滚动容器引用
const stagesContainer = ref<HTMLElement>()

// 重试状态
const retrying = ref(false)

// 当前阶段
const currentStage = computed(() => {
  return props.stages.find(stage => stage.status === 'in_progress') || 
         props.stages[props.stages.length - 1]
})

// 获取阶段显示名称（中文翻译）
const getStageDisplayName = (stageName: string) => {
  const nameMap: Record<string, string> = {
    'initializing': t('stage.initializing') ,
    'setup_environment': t('stage.setupEnvironment'),
    'pending_agents': t('stage.pendingAgents'),
    'check_requirement': t('stage.checkRequirement'),
    'generate_prd': t('stage.generatePrd'),
    'define_ux_standard': t('stage.defineUxStandard'),
    'design_architecture': t('stage.designArchitecture'),
    'define_data_model': t('stage.defineDataModel'),
    'define_api': t('stage.defineApi'),
    'plan_epic_and_story': t('stage.planEpicAndStory'),
    'develop_story': t('stage.developStory'),
    'fix_bug': t('stage.fixBug'),
    'run_test': t('stage.runTest'),
    'deploy': t('stage.deploy'),
    'done': t('stage.done'),
    'failed': t('stage.failed')
  }
  return nameMap[stageName] || stageName
}

// 监听阶段变化，自动滚动到最新阶段
watch(() => props.stages, (newStages, oldStages) => {
  if (newStages.length > (oldStages?.length || 0)) {
    // 有新阶段添加，滚动到最新阶段
    nextTick(() => {
      scrollToLatestStage()
    })
  }
}, { deep: true })

// 滚动到最新阶段
const scrollToLatestStage = () => {
  if (!stagesContainer.value || props.layout !== 'horizontal') return
  
  const container = stagesContainer.value
  const latestStage = container.querySelector('.stage-item:last-child') as HTMLElement
  
  if (latestStage) {
    const containerWidth = container.clientWidth
    const stageLeft = latestStage.offsetLeft
    const stageWidth = latestStage.clientWidth
    
    // 计算滚动位置，让最新阶段显示在容器右侧
    const scrollLeft = Math.max(0, stageLeft + stageWidth - containerWidth + 20)
    
    container.scrollTo({
      left: scrollLeft,
      behavior: 'smooth'
    })
  }
}

// 获取阶段样式类
const getStageClass = (stage: DevStage) => ({
  'stage-done': stage.status === 'done',
  'stage-failed': stage.status === 'failed',
  'stage-in-progress': stage.status === 'in_progress',
  'stage-pending': stage.status === 'pending'
})


// 获取状态颜色
const getStatusColor = (status: string) => {
  const colorMap = {
    done: '#38A169',
    failed: '#E53E3E',
    in_progress: '#D69E2E',
    pending: '#A0AEC0'
  }
  return colorMap[status as keyof typeof colorMap] || '#A0AEC0'
}

// 获取状态图标
const getStatusIcon = (status: string) => {
  const iconMap = {
    done: CheckIcon,
    failed: ErrorIcon,
    in_progress: ClockIcon,
    pending: ClockIcon
  }
  return iconMap[status as keyof typeof iconMap] || ClockIcon
}

// 获取状态文本
const getStatusText = (stage: DevStage) => {
  const statusMap = {
    done: `${stage.name} ${t('common.completed')}`,
    failed: `${stage.name} ${t('common.failed')}`,
    in_progress: `${t('common.inProgress')} ${stage.name}...`,
    pending: `${t('common.pending')} ${stage.name}`
  }
  return statusMap[stage.status as keyof typeof statusMap] || stage.name
}

// 获取失败原因
const getFailedReason = (stage: DevStage) => {
  if (stage.status === 'failed' && stage.failed_reason) {
    return stage.failed_reason
  }
  return null
}

// 重试处理函数
const handleRetry = async () => {
  if (!currentStage.value?.task_id) {
    message.error('任务ID不存在，无法重试')
    return
  }

  try {
    retrying.value = true
    const result = await taskStore.retryTask(currentStage.value.task_id)
    
    if (result.success) {
      message.success(result.message || '重试任务成功')
      // 触发父组件刷新数据
      emit('retry-success')
    } else {
      message.error(result.message)
    }
  } catch (error: any) {
    console.error('重试任务失败:', error)
    message.error('重试任务失败')
  } finally {
    retrying.value = false
  }
}

// 定义事件
const emit = defineEmits<{
  'retry-success': []
}>()
</script>

<style scoped>
.dev-stages {
  background: white;
  border-radius: var(--border-radius-lg);
}

/* 横向布局样式 */
.dev-stages.horizontal {
  padding: var(--spacing-md) var(--spacing-lg);
  margin-bottom: 0;
}

.stages-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.stages-header h3 {
  margin: 0;
  color: var(--primary-color);
  font-size: 1.1rem;
}

.progress-percentage {
  display: flex;
  align-items: center;
}

.percentage-circle {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: #38A169;
  border: 2px solid #2F855A;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 2px 4px rgba(56, 161, 105, 0.3);
}

.percentage-text {
  font-size: 0.9rem;
  font-weight: bold;
  color: white;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.5);
}

.stages-container {
  position: relative;
}

.stages-container.horizontal {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  overflow-x: auto;
  scrollbar-width: thin;
  scrollbar-color: #CBD5E1 transparent;
}

.stages-container.horizontal::-webkit-scrollbar {
  height: 6px;
}

.stages-container.horizontal::-webkit-scrollbar-track {
  background: transparent;
}

.stages-container.horizontal::-webkit-scrollbar-thumb {
  background: #CBD5E1;
  border-radius: 3px;
}

.stages-container.horizontal::-webkit-scrollbar-thumb:hover {
  background: #A0AEC0;
}

.stage-item {
  display: flex;
  align-items: center;
  margin-bottom: var(--spacing-lg);
  position: relative;
  margin-top: var(--spacing-md);
}

.stage-item.horizontal {
  flex-direction: column;
  margin-bottom: 0;
  flex-shrink: 0;
  min-width: 60px;
  text-align: center;
}

.stage-circle {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: #E2E8F0; /* 默认灰色背景 */
  border: 2px solid #CBD5E1; /* 默认灰色边框 */
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  margin-right: var(--spacing-md);
  transition: all 0.3s ease;
  position: relative;
}

.stage-item.horizontal .stage-circle {
  margin-right: 0;
  margin-bottom: var(--spacing-sm);
}

.stage-number {
  font-size: 1.1rem;
  font-weight: bold;
  color: #4A5568; /* 深灰色，适合浅色背景 */
  text-shadow: 0 1px 2px rgba(255, 255, 255, 0.5);
  z-index: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
}


.stage-content {
  flex: 1;
}

.stage-name {
  font-weight: 500;
  color: var(--text-primary);
  margin-bottom: var(--spacing-xs);
}

.stage-description {
  font-size: 0.9rem;
  color: var(--text-secondary);
}

/* 阶段状态样式 */
.stage-done .stage-circle {
  background: #38A169;
  border: 2px solid #2F855A;
}

.stage-done .stage-number {
  color: white;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.5);
}

.stage-failed .stage-circle {
  background: #E53E3E;
  border: 2px solid #C53030;
}

.stage-failed .stage-number {
  color: white;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.5);
}

.stage-in-progress .stage-circle {
  background: #D69E2E;
  border: 2px solid #B7791F;
  animation: pulse 2s infinite;
}

.stage-in-progress .stage-number {
  color: white;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.5);
}

.stage-pending .stage-circle {
  background: #A0AEC0;
  border: 2px solid #718096;
}

.stage-pending .stage-number {
  color: white;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.5);
}

@keyframes pulse {
  0% {
    box-shadow: 0 0 0 0 rgba(214, 158, 46, 0.7);
  }
  70% {
    box-shadow: 0 0 0 10px rgba(214, 158, 46, 0);
  }
  100% {
    box-shadow: 0 0 0 0 rgba(214, 158, 46, 0);
  }
}


/* 当前状态信息 */
.current-status {
  margin-top: var(--spacing-md);
  padding-top: var(--spacing-xs);
  border-top: 1px solid var(--border-color);
}

.status-info {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
}

.status-text {
  font-size: 0.9rem;
  color: var(--text-secondary);
}

/* 失败原因样式 */
.failed-reason {
  margin-top: var(--spacing-sm);
  padding: var(--spacing-sm);
  background: #FED7D7;
  border: 1px solid #FEB2B2;
  border-radius: var(--border-radius-md);
  border-left: 4px solid #E53E3E;
}

.failed-reason-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: var(--spacing-sm);
}

.failed-reason-text {
  font-size: 0.9rem;
  color: #742A2A;
  line-height: 1.4;
  font-weight: 500;
  flex: 1;
}

.retry-button {
  flex-shrink: 0;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .stages-header {
    flex-direction: column;
    align-items: flex-start;
    gap: var(--spacing-sm);
  }
  
  .stage-item {
    margin-bottom: var(--spacing-md);
  }
  
  .stage-circle {
    width: 28px;
    height: 28px;
  }
  
  
  .stage-item.horizontal {
    min-width: 60px;
  }
  
  .stage-name {
    font-size: 0.8rem;
  }
}
</style>
