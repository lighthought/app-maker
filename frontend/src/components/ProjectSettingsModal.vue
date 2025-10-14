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

      <!-- 开发配置区域：只在项目未完成时显示 -->
      <template v-if="!isProjectCompleted">
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
      </template>
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
import type { Project, UpdateProjectFormData } from '@/types/project'
import { useProjectStore } from '@/stores/project'

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
const projectStore = useProjectStore()

// 表单引用
const formRef = ref()

// 加载状态
const loading = ref(false)

// 本地显示状态
const localShow = computed({
  get: () => props.show,
  set: (value) => emit('update:show', value)
})

// 判断项目是否已完成
const isProjectCompleted = computed(() => {
  if (!props.project) return false
  // 项目状态为 done 时视为已完成
  return props.project.status === 'done'
})

// CLI 工具选项
const cliToolOptions = [
  { label: 'Claude Code', value: 'claude-code' },
  { label: 'Qwen Code', value: 'qwen-code' },
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
const formData = ref<UpdateProjectFormData>({
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
watch(() => [props.show, props.project] as const, ([newShow, newProject]) => {
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
    
    // 构建更新数据：已完成的项目只更新基本信息
    const updateData: {
      name: string
      description: string
      cli_tool?: string | null
      ai_model?: string | null
      model_provider?: string | null
      model_api_url?: string | null
    } = {
      name: formData.value.name,
      description: formData.value.description
    }
    
    // 未完成的项目可以更新开发配置
    if (!isProjectCompleted.value) {
      updateData.cli_tool = formData.value.cliTool || null
      updateData.ai_model = formData.value.aiModel || null
      updateData.model_provider = formData.value.modelProvider || null
      updateData.model_api_url = formData.value.modelApiUrl || null
    }
    
    // 使用 project store 更新项目
    await projectStore.updateProject(props.project.guid, updateData)
    
    message.success(t('project.projectUpdated'))
    emit('saved')
    handleClose()
  } catch (error: any) {
    console.error('保存项目设置失败:', error)
    
    if (error.message) {
      message.error(error.message)
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

