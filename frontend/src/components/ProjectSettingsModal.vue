<template>
  <n-modal
    v-model:show="localShow"
    preset="card"
    :title="t('project.projectSettings')"
    :style="{ width: '600px' }"
    :mask-closable="false"
    :closable="true"
    @update:show="handleModalClose"
  >
    <n-form
      ref="formRef"
      :model="formData"
      :rules="rules"
      label-placement="left"
      label-width="120px"
      require-mark-placement="right-hanging"
    >
      <n-form-item :label="t('project.projectName')" path="name">
        <n-input
          v-model:value="formData.name"
          :placeholder="t('project.projectName')"
        />
      </n-form-item>

      <n-form-item :label="t('project.projectDescription')" path="description">
        <n-input
          v-model:value="formData.description"
          type="textarea"
          :placeholder="t('project.projectDescription')"
          :rows="3"
        />
      </n-form-item>

      <n-divider />

      <h3 style="margin-bottom: 16px;">{{ t('project.devConfiguration') }}</h3>
      
      <n-form-item :label="t('userSettings.cliTool')" path="cliTool">
        <n-select
          v-model:value="formData.cliTool"
          :options="cliToolOptions"
          :placeholder="t('userSettings.cliToolPlaceholder')"
          clearable
        />
      </n-form-item>

      <n-form-item :label="t('userSettings.modelProvider')" path="modelProvider">
        <n-select
          v-model:value="formData.modelProvider"
          :options="modelProviderOptions"
          :placeholder="t('userSettings.modelProviderPlaceholder')"
          clearable
          @update:value="handleProviderChange"
        />
      </n-form-item>

      <n-form-item :label="t('userSettings.aiModel')" path="aiModel">
        <n-input
          v-model:value="formData.aiModel"
          :placeholder="t('userSettings.aiModelPlaceholder')"
        />
      </n-form-item>

      <n-form-item :label="t('userSettings.modelApiUrl')" path="modelApiUrl">
        <n-input
          v-model:value="formData.modelApiUrl"
          :placeholder="t('userSettings.modelApiUrlPlaceholder')"
          type="text"
        />
      </n-form-item>

      <n-alert type="info" :show-icon="false" style="margin-top: 8px;">
        {{ t('project.devConfigNote') }}
      </n-alert>
    </n-form>

    <template #footer>
      <div class="modal-footer">
        <n-button @click="handleClose">{{ t('common.cancel') }}</n-button>
        <n-button
          type="primary"
          :loading="loading"
          @click="handleSave"
        >
          {{ t('common.save') }}
        </n-button>
      </div>
    </template>
  </n-modal>
</template>

<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useMessage, NModal, NForm, NFormItem, NInput, NSelect, NButton, NAlert, NDivider, type FormRules } from 'naive-ui'
import type { Project } from '@/types/project'
import { httpService } from '@/utils/http'

interface Props {
  show: boolean
  project?: Project
}

interface Emits {
  (e: 'update:show', value: boolean): void
  (e: 'saved'): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const { t } = useI18n()
const message = useMessage()

// 表单引用
const formRef = ref()

// 加载状态
const loading = ref(false)

// 本地显示状态
const localShow = computed({
  get: () => props.show,
  set: (value) => emit('update:show', value)
})

// CLI 工具选项
const cliToolOptions = [
  { label: 'Claude Code', value: 'claude-code' },
  { label: 'Qwen Code', value: 'qwen-code' },
  { label: 'iFlow CLI', value: 'iflow-cli' },
  { label: 'Auggie CLI', value: 'auggie-cli' },
  { label: 'Gemini', value: 'gemini' }
]

// 模型提供商选项
const modelProviderOptions = [
  { label: 'Ollama (本地)', value: 'ollama' },
  { label: 'Zhipu AI (智谱)', value: 'zhipu' },
  { label: 'Anthropic (Claude)', value: 'anthropic' },
  { label: 'OpenAI (GPT)', value: 'openai' },
  { label: 'vLLM (本地)', value: 'vllm' }
]

// 默认模型映射
const defaultModelByProvider: Record<string, string> = {
  'ollama': 'qwen2.5-coder:14b',
  'zhipu': 'glm-4.6',
  'anthropic': 'claude-sonnet-4',
  'openai': 'gpt-4o',
  'vllm': 'deepseek-coder:14b'
}

// 默认 API URL 映射
const defaultApiUrlByProvider: Record<string, string> = {
  'ollama': 'http://localhost:11434',
  'zhipu': 'https://open.bigmodel.cn/api/anthropic',
  'anthropic': 'https://api.anthropic.com',
  'openai': 'https://api.openai.com/v1',
  'vllm': 'http://localhost:8000'
}

// 表单数据
const formData = ref({
  name: '',
  description: '',
  cliTool: '',
  aiModel: '',
  modelProvider: '',
  modelApiUrl: ''
})

// 表单验证规则
const rules: FormRules = {
  name: [
    { required: true, message: t('project.nameRequired'), trigger: 'blur' }
  ]
}

// 监听显示状态和项目变化，初始化表单数据
watch(() => [props.show, props.project], ([newShow, newProject]) => {
  if (newShow && newProject) {
    formData.value = {
      name: newProject.name || '',
      description: newProject.description || '',
      cliTool: newProject.cli_tool || '',
      aiModel: newProject.ai_model || '',
      modelProvider: newProject.model_provider || '',
      modelApiUrl: newProject.model_api_url || ''
    }
  }
}, { immediate: true })

// 处理模型提供商变更
const handleProviderChange = (value: string) => {
  if (!value) return
  // 自动填充默认模型和 API URL（如果当前为空）
  if (!formData.value.aiModel) {
    formData.value.aiModel = defaultModelByProvider[value] || ''
  }
  if (!formData.value.modelApiUrl) {
    formData.value.modelApiUrl = defaultApiUrlByProvider[value] || ''
  }
}

// 关闭弹窗
const handleClose = () => {
  emit('update:show', false)
}

// 处理模态框关闭
const handleModalClose = (value: boolean) => {
  if (!value) {
    emit('update:show', false)
  }
}

// 保存设置
const handleSave = async () => {
  if (!props.project?.guid) return

  try {
    // 验证表单
    await formRef.value?.validate()
    
    loading.value = true
    
    // 调用后端接口
    const response = await httpService.put<{
      code: number
      message: string
      data?: any
    }>(`/projects/${props.project.guid}`, {
      name: formData.value.name,
      description: formData.value.description,
      cli_tool: formData.value.cliTool || null,
      ai_model: formData.value.aiModel || null,
      model_provider: formData.value.modelProvider || null,
      model_api_url: formData.value.modelApiUrl || null
    })

    if (response.code === 0) {
      message.success(t('project.projectUpdated'))
      emit('saved')
      handleClose()
    } else {
      message.error(response.message || t('project.projectUpdateError'))
    }
  } catch (error: any) {
    console.error('保存项目设置失败:', error)
    
    if (error.response?.data?.message) {
      message.error(error.response.data.message)
    } else {
      message.error(t('userSettings.networkError'))
    }
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: var(--spacing-md);
}
</style>

