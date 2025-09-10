<template>
  <n-modal
    v-model:show="showModal"
    preset="card"
    title="项目下载进度"
    style="width: 600px;"
    :mask-closable="false"
    :closable="taskStatus !== 'pending' && taskStatus !== 'in_progress'"
  >
    <div class="task-progress-container">
      <!-- 任务信息 -->
      <div class="task-info">
        <n-space vertical>
          <div class="task-id">
            <n-text type="info">任务ID: {{ taskId }}</n-text>
          </div>
          
          <!-- 进度条 -->
          <div class="progress-section">
            <n-progress
              type="line"
              :percentage="progress"
              :status="progressStatus"
              :show-indicator="true"
            />
            <div class="progress-text">
              <n-text :type="progressTextType">{{ progressText }}</n-text>
            </div>
          </div>

          <!-- 状态信息 -->
          <div class="status-section">
            <n-space align="center">
              <n-icon
                :size="20"
                :color="statusIconColor"
              >
                <component :is="statusIcon" />
              </n-icon>
              <n-text :type="statusTextType">{{ statusText }}</n-text>
            </n-space>
            
            <!-- 下载路径信息 -->
            <div v-if="message" class="download-path">
              <n-text type="info" depth="3">
                {{ message }}
              </n-text>
            </div>
          </div>

          <!-- 操作按钮 -->
          <div class="action-section">
            <div class="action-content">
              <!-- 右侧操作按钮 -->
              <div class="action-buttons">
                <n-button
                  v-if="taskStatus === 'done'"
                  type="primary"
                  @click="handleDownload"
                  :loading="downloading"
                >
                  <template #icon>
                    <n-icon><DownloadOutlined /></n-icon>
                  </template>
                  下载文件
                </n-button>
                
                <n-button
                  v-if="taskStatus === 'failed'"
                  type="error"
                  @click="handleRetry"
                  :loading="retrying"
                >
                  <template #icon>
                    <n-icon><ReloadOutlined /></n-icon>
                  </template>
                  重试
                </n-button>
                
                <n-button
                  v-if="taskStatus === 'done' || taskStatus === 'failed'"
                  @click="handleClose"
                >
                  关闭
                </n-button>
              </div>
            </div>
          </div>
        </n-space>
      </div>
    </div>
  </n-modal>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted, h } from 'vue'
import { NModal, NSpace, NText, NProgress, NButton, NIcon } from 'naive-ui'
// 使用 SVG 图标替代 @vicons/antd
const DownloadOutlined = () => h('svg', { 
  viewBox: '0 0 24 24', 
  fill: 'currentColor',
  style: 'width: 1em; height: 1em;'
}, [
  h('path', { d: 'M19 9h-4V3H9v6H5l7 7 7-7zM5 18v2h14v-2H5z' })
])

const ReloadOutlined = () => h('svg', { 
  viewBox: '0 0 24 24', 
  fill: 'currentColor',
  style: 'width: 1em; height: 1em;'
}, [
  h('path', { d: 'M17.65 6.35C16.2 4.9 14.21 4 12 4c-4.42 0-7.99 3.58-7.99 8s3.57 8 7.99 8c3.73 0 6.84-2.55 7.73-6h-2.08c-.82 2.33-3.04 4-5.65 4-3.31 0-6-2.69-6-6s2.69-6 6-6c1.66 0 3.14.69 4.22 1.78L13 11h7V4l-2.35 2.35z' })
])

const CheckCircleOutlined = () => h('svg', { 
  viewBox: '0 0 24 24', 
  fill: 'currentColor',
  style: 'width: 1em; height: 1em;'
}, [
  h('path', { d: 'M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z' })
])

const CloseCircleOutlined = () => h('svg', { 
  viewBox: '0 0 24 24', 
  fill: 'currentColor',
  style: 'width: 1em; height: 1em;'
}, [
  h('path', { d: 'M12 2C6.47 2 2 6.47 2 12s4.47 10 10 10 10-4.47 10-10S17.53 2 12 2zm5 13.59L15.59 17 12 13.41 8.41 17 7 15.59 10.59 12 7 8.41 8.41 7 12 10.59 15.59 7 17 8.41 13.41 12 17 15.59z' })
])

const LoadingOutlined = () => h('svg', { 
  viewBox: '0 0 24 24', 
  fill: 'currentColor',
  style: 'width: 1em; height: 1em;'
}, [
  h('path', { d: 'M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm0 18c-4.42 0-8-3.58-8-8s3.58-8 8-8 8 3.58 8 8-3.58 8-8 8z' }),
  h('path', { d: 'M12 6v6l4 2' })
])
import { useTaskStore } from '@/stores/task'
import { useFilesStore } from '@/stores/file'
import type { TaskResult } from '@/types/task'

interface Props {
  show: boolean
  taskId: string
  projectId: string
}

interface Emits {
  (e: 'update:show', value: boolean): void
  (e: 'retry', taskId: string): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const taskStore = useTaskStore()
const fileStore = useFilesStore()

// 响应式数据
const showModal = computed({
  get: () => props.show,
  set: (value) => emit('update:show', value)
})

const taskResult = ref<TaskResult | null>(null)
const downloading = ref(false)
const retrying = ref(false)
const pollingInterval = ref<number | null>(null)

// 计算属性
const taskStatus = computed(() => taskResult.value?.status || 'pending')
const progress = computed(() => taskResult.value?.progress || 0)
const message = computed(() => taskResult.value?.message || '')

const progressStatus = computed(() => {
  switch (taskStatus.value) {
    case 'done':
      return 'success'
    case 'failed':
      return 'error'
    case 'in_progress':
      return 'info'
    default:
      return 'default'
  }
})

const progressText = computed(() => {
  switch (taskStatus.value) {
    case 'pending':
      return '等待中...'
    case 'in_progress':
      return `处理中... ${progress.value}%`
    case 'done':
      return '完成'
    case 'failed':
      return '失败'
    default:
      return '未知状态'
  }
})

const progressTextType = computed(() => {
  switch (taskStatus.value) {
    case 'done':
      return 'success'
    case 'failed':
      return 'error'
    case 'in_progress':
      return 'info'
    default:
      return 'default'
  }
})

const statusText = computed(() => {
  switch (taskStatus.value) {
    case 'pending':
      return '任务已创建，等待处理'
    case 'in_progress':
      return '正在处理中，请稍候'
    case 'done':
      return '任务完成，可以下载文件'
    case 'failed':
      return '任务失败，请重试'
    default:
      return '未知状态'
  }
})

const statusTextType = computed(() => {
  switch (taskStatus.value) {
    case 'done':
      return 'success'
    case 'failed':
      return 'error'
    case 'in_progress':
      return 'info'
    default:
      return 'default'
  }
})

const statusIcon = computed(() => {
  switch (taskStatus.value) {
    case 'done':
      return CheckCircleOutlined
    case 'failed':
      return CloseCircleOutlined
    case 'in_progress':
    case 'pending':
      return LoadingOutlined
    default:
      return LoadingOutlined
  }
})

const statusIconColor = computed(() => {
  switch (taskStatus.value) {
    case 'done':
      return '#18a058'
    case 'failed':
      return '#d03050'
    case 'in_progress':
    case 'pending':
      return '#2080f0'
    default:
      return '#666'
  }
})

// 方法
const startPolling = () => {
  if (pollingInterval.value) {
    clearInterval(pollingInterval.value)
  }
  
  pollingInterval.value = setInterval(async () => {
    const result = await taskStore.getTaskStatus(props.taskId)
    if (result) {
      taskResult.value = result
      
      // 如果任务完成或失败，停止轮询
      if (result.status === 'done' || result.status === 'failed') {
        stopPolling()
      }
    }
  }, 3000) // 每2秒轮询一次
}

const stopPolling = () => {
  if (pollingInterval.value) {
    clearInterval(pollingInterval.value)
    pollingInterval.value = null
  }
}

const handleDownload = async () => {
  if (!taskResult.value?.message) {
    console.error('没有可下载的文件路径')
    return
  }
  
  downloading.value = true
  try {
    await fileStore.downloadFile(taskResult.value.message)
  } catch (error) {
    console.error('下载文件失败:', error)
  } finally {
    downloading.value = false
  }
}

const handleRetry = () => {
  retrying.value = true
  emit('retry', props.taskId)
  setTimeout(() => {
    retrying.value = false
  }, 1000)
}

const handleClose = () => {
  showModal.value = false
}

// 监听器
watch(() => props.show, (newShow) => {
  if (newShow && props.taskId) {
    // 开始轮询
    startPolling()
  } else {
    // 停止轮询
    stopPolling()
  }
})

watch(() => props.taskId, (newTaskId) => {
  if (newTaskId && props.show) {
    // 重置状态
    taskResult.value = null
    // 开始轮询
    startPolling()
  }
})

// 生命周期
onMounted(() => {
  if (props.show && props.taskId) {
    startPolling()
  }
})

onUnmounted(() => {
  stopPolling()
})
</script>

<style scoped lang="scss">
.task-progress-container {
  .task-info {
    .task-id {
      margin-bottom: 16px;
    }
    
    .progress-section {      
      .progress-text {
        margin-top: 8px;
        text-align: center;
      }
    }
    
    .status-section {
      margin-bottom: 16px;
      padding: 8px 0;
      
      .download-path {
        margin-top: 8px;
        font-size: 12px;
        color: var(--n-text-color-disabled);
      }
    }
    
    .action-section {
      margin-top: 24px;
      padding-top: 16px;
      border-top: 1px solid var(--n-border-color);
      
      .action-content {
        display: flex;
        justify-content: flex-end;
        
        .action-buttons {
          display: flex;
          gap: 8px;
        }
      }
    }
  }
}
</style>
