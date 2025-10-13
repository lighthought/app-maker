<template>
  <n-modal
    v-model:show="show"
    preset="card"
    :title="t('preview.shareLink')"
    :style="{ width: '500px' }"
    :mask-closable="true"
    :closable="true"
  >
    <div v-if="loading" class="loading-container">
      <n-spin size="medium" />
      <p>{{ t('preview.generatingLink') }}</p>
    </div>

    <div v-else-if="shareData" class="share-content">
      <n-alert type="success" :show-icon="true" style="margin-bottom: 16px;">
        {{ t('preview.linkGenerated') }}
      </n-alert>

      <div class="link-container">
        <n-input
          :value="shareData.share_link"
          readonly
          type="text"
          size="large"
        />
        <n-button
          type="primary"
          size="large"
          @click="copyLink"
          style="margin-left: 8px;"
        >
          <template #icon>
            <n-icon><CopyIcon /></n-icon>
          </template>
          {{ t('common.copy') }}
        </n-button>
      </div>

      <div class="info-section">
        <div class="info-item">
          <span class="info-label">{{ t('preview.expiresAt') }}:</span>
          <span class="info-value">{{ formatDate(shareData.expires_at) }}</span>
        </div>
        <div class="info-item">
          <span class="info-label">{{ t('preview.token') }}:</span>
          <span class="info-value token">{{ shareData.token }}</span>
        </div>
      </div>

      <n-alert type="info" :show-icon="false" style="margin-top: 16px;">
        {{ t('preview.shareNote') }}
      </n-alert>
    </div>

    <template #footer>
      <div class="modal-footer">
        <n-button @click="handleClose">{{ t('common.close') }}</n-button>
        <n-button
          v-if="shareData"
          type="primary"
          @click="openInNewTab"
        >
          {{ t('preview.openLink') }}
        </n-button>
      </div>
    </template>
  </n-modal>
</template>

<script setup lang="ts">
import { ref, watch, h } from 'vue'
import { useI18n } from 'vue-i18n'
import { useMessage, NModal, NButton, NInput, NAlert, NSpin, NIcon } from 'naive-ui'
import { httpService } from '@/utils/http'
// 导入图标
import { CopyIcon } from '@/components/icon'

interface Props {
  show: boolean
  projectGuid: string
}

interface Emits {
  (e: 'update:show', value: boolean): void
}

interface ShareData {
  token: string
  share_link: string
  expires_at: string
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const { t } = useI18n()
const message = useMessage()

const loading = ref(false)
const shareData = ref<ShareData | null>(null)

// 监听 show 变化，生成分享链接
watch(() => props.show, async (newVal) => {
  if (newVal && props.projectGuid) {
    await generateShareLink()
  } else {
    // 关闭时清空数据
    shareData.value = null
  }
})

// 生成分享链接
const generateShareLink = async () => {
  try {
    loading.value = true
    const response = await httpService.post<{
      code: number
      message: string
      data: ShareData
    }>(`/projects/${props.projectGuid}/preview-link`, {})

    if (response.code === 0 && response.data) {
      shareData.value = response.data
      message.success(t('preview.linkGenerated'))
    } else {
      message.error(response.message || t('preview.generateFailed'))
      handleClose()
    }
  } catch (error: any) {
    console.error('生成分享链接失败:', error)
    message.error(t('preview.generateError'))
    handleClose()
  } finally {
    loading.value = false
  }
}

// 复制链接
const copyLink = async () => {
  if (!shareData.value) return

  try {
    await navigator.clipboard.writeText(shareData.value.share_link)
    message.success(t('common.copySuccess'))
  } catch (error) {
    console.error('复制失败:', error)
    message.error(t('common.copyFailed'))
  }
}

// 在新标签页打开
const openInNewTab = () => {
  if (shareData.value) {
    window.open(shareData.value.share_link, '_blank')
  }
}

// 关闭弹窗
const handleClose = () => {
  emit('update:show', false)
}

// 格式化日期
const formatDate = (dateStr: string) => {
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}
</script>

<style scoped>
.loading-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px 20px;
  gap: 16px;
}

.loading-container p {
  color: var(--text-secondary);
  margin: 0;
}

.share-content {
  display: flex;
  flex-direction: column;
}

.link-container {
  display: flex;
  align-items: center;
  margin-bottom: 20px;
}

.info-section {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 16px;
  background: var(--background-color);
  border-radius: var(--border-radius-md);
}

.info-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.info-label {
  font-weight: 500;
  color: var(--text-secondary);
}

.info-value {
  color: var(--text-primary);
  font-family: monospace;
}

.info-value.token {
  font-size: 0.85em;
  max-width: 250px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: var(--spacing-md);
}
</style>

