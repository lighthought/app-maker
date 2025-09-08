<template>
  <div class="dev-stages">
    <div class="stages-header">
      <h3>开发进度</h3>
      <div class="progress-text">{{ currentProgress }}%</div>
    </div>
    
    <div class="stages-container">
      <div
        v-for="(stage, index) in stages"
        :key="stage.id"
        class="stage-item"
        :class="getStageClass(stage)"
      >
        <div class="stage-circle">
          <n-icon v-if="stage.status === 'completed'" size="16" color="white">
            <CheckIcon />
          </n-icon>
          <n-icon v-else-if="stage.status === 'failed'" size="16" color="white">
            <ErrorIcon />
          </n-icon>
          <span v-else class="stage-number">{{ index + 1 }}</span>
        </div>
        
        <div class="stage-content">
          <div class="stage-name">{{ stage.name }}</div>
          <div class="stage-description">{{ stage.description }}</div>
        </div>
        
        <!-- 连接线 -->
        <div
          v-if="index < stages.length - 1"
          class="stage-connector"
          :class="getConnectorClass(stage, stages[index + 1])"
        ></div>
      </div>
    </div>
    
    <!-- 当前状态信息 -->
    <div v-if="currentStage" class="current-status">
      <div class="status-info">
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
import { computed, h } from 'vue'
import { NIcon } from 'naive-ui'
import type { DevStage } from '@/types/project'

interface Props {
  stages: DevStage[]
  currentProgress: number
}

const props = defineProps<Props>()

// 当前阶段
const currentStage = computed(() => {
  return props.stages.find(stage => stage.status === 'in_progress') || 
         props.stages[props.stages.length - 1]
})

// 获取阶段样式类
const getStageClass = (stage: DevStage) => ({
  'stage-completed': stage.status === 'completed',
  'stage-failed': stage.status === 'failed',
  'stage-in-progress': stage.status === 'in_progress',
  'stage-pending': stage.status === 'pending'
})

// 获取连接线样式类
const getConnectorClass = (currentStage: DevStage, nextStage: DevStage) => ({
  'connector-completed': currentStage.status === 'completed',
  'connector-failed': currentStage.status === 'failed',
  'connector-in-progress': currentStage.status === 'in_progress',
  'connector-pending': currentStage.status === 'pending'
})

// 获取状态颜色
const getStatusColor = (status: string) => {
  const colorMap = {
    completed: '#38A169',
    failed: '#E53E3E',
    in_progress: '#D69E2E',
    pending: '#A0AEC0'
  }
  return colorMap[status as keyof typeof colorMap] || '#A0AEC0'
}

// 获取状态图标
const getStatusIcon = (status: string) => {
  const iconMap = {
    completed: CheckIcon,
    failed: ErrorIcon,
    in_progress: ClockIcon,
    pending: ClockIcon
  }
  return iconMap[status as keyof typeof iconMap] || ClockIcon
}

// 获取状态文本
const getStatusText = (stage: DevStage) => {
  const statusMap = {
    completed: `${stage.name}已完成`,
    failed: `${stage.name}执行失败`,
    in_progress: `正在${stage.name}...`,
    pending: `等待${stage.name}`
  }
  return statusMap[stage.status as keyof typeof statusMap] || stage.name
}

// 图标组件
const CheckIcon = () => h('svg', { 
  viewBox: '0 0 24 24', 
  fill: 'currentColor',
  style: 'width: 1em; height: 1em;'
}, [
  h('path', { d: 'M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41z' })
])

const ErrorIcon = () => h('svg', { 
  viewBox: '0 0 24 24', 
  fill: 'currentColor',
  style: 'width: 1em; height: 1em;'
}, [
  h('path', { d: 'M12 2C6.47 2 2 6.47 2 12s4.47 10 10 10 10-4.47 10-10S17.53 2 12 2zm5 13.59L15.59 17 12 13.41 8.41 17 7 15.59 10.59 12 7 8.41 8.41 7 12 10.59 15.59 7 17 8.41 13.41 12 17 15.59z' })
])

const ClockIcon = () => h('svg', { 
  viewBox: '0 0 24 24', 
  fill: 'currentColor',
  style: 'width: 1em; height: 1em;'
}, [
  h('path', { d: 'M11.99 2C6.47 2 2 6.48 2 12s4.47 10 9.99 10C17.52 22 22 17.52 22 12S17.52 2 11.99 2zM12 20c-4.42 0-8-3.58-8-8s3.58-8 8-8 8 3.58 8 8-3.58 8-8 8z' }),
  h('path', { d: 'M12.5 7H11v6l5.25 3.15.75-1.23-4.5-2.67z' })
])
</script>

<style scoped>
.dev-stages {
  background: white;
  border-radius: var(--border-radius-lg);
  padding: var(--spacing-lg);
  margin-bottom: var(--spacing-lg);
}

.stages-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-lg);
}

.stages-header h3 {
  margin: 0;
  color: var(--primary-color);
  font-size: 1.1rem;
}

.progress-text {
  font-size: 1.2rem;
  font-weight: bold;
  color: var(--primary-color);
}

.stages-container {
  position: relative;
}

.stage-item {
  display: flex;
  align-items: center;
  margin-bottom: var(--spacing-lg);
  position: relative;
}

.stage-circle {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  margin-right: var(--spacing-md);
  transition: all 0.3s ease;
}

.stage-number {
  font-size: 0.9rem;
  font-weight: bold;
  color: white;
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
.stage-completed .stage-circle {
  background: #38A169;
}

.stage-failed .stage-circle {
  background: #E53E3E;
}

.stage-in-progress .stage-circle {
  background: #D69E2E;
  animation: pulse 2s infinite;
}

.stage-pending .stage-circle {
  background: #A0AEC0;
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

/* 连接线样式 */
.stage-connector {
  position: absolute;
  left: 15px;
  top: 32px;
  width: 2px;
  height: var(--spacing-lg);
  transition: all 0.3s ease;
}

.connector-completed {
  background: #38A169;
}

.connector-failed {
  background: #E53E3E;
}

.connector-in-progress {
  background: linear-gradient(to bottom, #D69E2E, #A0AEC0);
}

.connector-pending {
  background: #A0AEC0;
}

/* 当前状态信息 */
.current-status {
  margin-top: var(--spacing-lg);
  padding-top: var(--spacing-lg);
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
  
  .stage-connector {
    left: 13px;
  }
}
</style>
