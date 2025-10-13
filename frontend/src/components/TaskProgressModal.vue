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
                    <n-icon><DownloadIcon /></n-icon>
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
                    <n-icon><ReloadIcon /></n-icon>
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
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { NModal, NSpace, NText, NProgress, NButton, NIcon } from 'naive-ui'
// 导入图标
import { 
  DownloadIcon, 
  ReloadIcon, 
  CheckCircleIcon, 
  CloseCircleIcon, 
  LoadingIcon 
} from '@/components/icon'
import { useTaskStore } from '@/stores/task'
import { useFilesStore } from '@/stores/file'
import type { TaskResult } from '@/types/task'

interface Props {
  show: boolean
  taskId: string
  projectGuid: string
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
      return CheckCircleIcon
    case 'failed':
      return CloseCircleIcon
    case 'in_progress':
    case 'pending':
      return LoadingIcon
    default:
      return LoadingIcon
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
