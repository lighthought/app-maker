<template>
  <n-modal
    v-model:show="show"
    preset="card"
    :title="t('header.userSettings')"
    :style="{ width: '500px' }"
    :mask-closable="false"
    :closable="true"
    :auto-focus="false"
    :trap-focus="false"
    @close="handleClose"
  >
    <n-form
      ref="formRef"
      :model="formData"
      :rules="rules"
      label-placement="left"
      label-width="auto"
      require-mark-placement="right-hanging"
    >
      <n-form-item :label="t('userSettings.nickname')" path="username" required>
        <n-input
          v-model:value="formData.username"
          :placeholder="t('userSettings.nicknamePlaceholder')"
          maxlength="20"
          show-count
        />
      </n-form-item>

      <n-form-item :label="t('userSettings.email')" path="email" required>
        <n-input
          v-model:value="formData.email"
          :placeholder="t('userSettings.emailPlaceholder')"
          type="text"
        />
      </n-form-item>

      <n-form-item :label="t('userSettings.phoneBinding')">
         <div class="binding-status">
           <n-tag type="success" size="small">{{ t('userSettings.phoneBound') }}</n-tag>
           <span class="binding-text">138****8888</span>
         </div>
       </n-form-item>

      <n-divider />

      <!-- 开发设置 -->
      <h3 style="margin-bottom: 16px;">{{ t('userSettings.developmentSettings') }}</h3>
      
      <n-form-item :label="t('userSettings.cliTool')" path="defaultCliTool">
        <n-select
          v-model:value="formData.defaultCliTool"
          :options="cliToolOptions"
          :placeholder="t('userSettings.cliToolPlaceholder')"
        />
      </n-form-item>

      <n-form-item :label="t('userSettings.modelProvider')" path="defaultModelProvider">
        <n-select
          v-model:value="formData.defaultModelProvider"
          :options="modelProviderOptions"
          :placeholder="t('userSettings.modelProviderPlaceholder')"
          @update:value="handleProviderChange"
        />
      </n-form-item>

      <n-form-item :label="t('userSettings.aiModel')" path="defaultAiModel">
        <n-input
          v-model:value="formData.defaultAiModel"
          :placeholder="t('userSettings.aiModelPlaceholder')"
        />
      </n-form-item>

      <n-form-item :label="t('userSettings.modelApiUrl')" path="defaultModelApiUrl">
        <n-input
          v-model:value="formData.defaultModelApiUrl"
          :placeholder="t('userSettings.modelApiUrlPlaceholder')"
          type="text"
        />
      </n-form-item>

      <n-form-item :label="t('userSettings.apiToken')" path="defaultApiToken">
        <n-input
          v-model:value="formData.defaultApiToken"
          :placeholder="t('userSettings.apiTokenPlaceholder')"
          :type="showApiToken ? 'text' : 'password'"
          clearable
        >
          <template #suffix>
            <component 
              :is="renderIcon(showApiToken ? EyeOffIcon : EyeIcon)" 
              @click="showApiToken = !showApiToken"
              style="cursor: pointer;"
            />
          </template>
        </n-input>
      </n-form-item>

      <n-alert type="info" :show-icon="false" style="margin-top: 8px;">
        {{ t('userSettings.devSettingsNote') }}
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
import { ref, reactive, watch, h } from 'vue'
import { useI18n } from 'vue-i18n'
import { useMessage, NModal, NForm, NFormItem, NInput, NButton, NTag, NSelect, NAlert, NDivider, NIcon, type FormRules } from 'naive-ui'
import { useUserStore } from '@/stores/user'
// 导入图标
import { EyeIcon, EyeOffIcon } from '@/components/icon'

interface Props {
  show: boolean
}

interface Emits {
  (e: 'update:show', value: boolean): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const userStore = useUserStore()
const message = useMessage()
const { t } = useI18n()

// 表单引用
const formRef = ref()

// 加载状态
const loading = ref(false)

// API Token 显示状态
const showApiToken = ref(false)

// 渲染图标
const renderIcon = (icon: any) => {
  return () => h(NIcon, null, { default: icon })
}

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
  'ollama': 'deepcoder:14b',
  'zhipu': 'glm-4.6',
  'anthropic': 'claude-sonnet-4',
  'openai': 'gpt-4o',
  'vllm': 'qwen2.5-coder:14b'
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
const formData = reactive({
  username: '',
  email: '',
  defaultCliTool: 'claude-code',
  defaultAiModel: 'glm-4.6',
  defaultModelProvider: 'zhipu',
  defaultModelApiUrl: 'https://open.bigmodel.cn/api/anthropic',
  defaultApiToken: ''
})

// 表单验证规则
const rules: FormRules = {
  username: [
    { required: true, message: t('userSettings.nicknameRequired'), trigger: 'blur' },
    { min: 3, max: 20, message: t('userSettings.nicknameLength'), trigger: 'blur' }
  ],
  email: [
    { required: true, message: t('userSettings.emailRequired'), trigger: 'blur' },
    { pattern: /^[^\s@]+@[^\s@]+\.[^\s@]+$/, message: t('userSettings.emailFormat'), trigger: 'blur' }
  ]
}

// 监听显示状态，初始化表单数据
watch(() => props.show, async (newVal) => {
  if (newVal && userStore.user) {
    formData.username = userStore.user.username || ''
    formData.email = userStore.user.email || ''
    
    // 加载用户开发设置
    const result = await userStore.getUserSettings()
    if (result.success && result.data) {
      formData.defaultCliTool = result.data.default_cli_tool || 'claude-code'
      formData.defaultAiModel = result.data.default_ai_model || 'glm-4.6'
      formData.defaultModelProvider = result.data.default_model_provider || 'zhipu'
      formData.defaultModelApiUrl = result.data.default_model_api_url || 'https://open.bigmodel.cn/api/anthropic'
      formData.defaultApiToken = result.data.default_api_token || ''
    } else {
      console.error('加载用户设置失败:', result.message)
    }
  }
})

// 处理模型提供商变更
const handleProviderChange = (value: string) => {
  // 自动填充默认模型和 API URL
  formData.defaultAiModel = defaultModelByProvider[value] || ''
  formData.defaultModelApiUrl = defaultApiUrlByProvider[value] || ''
}

// 关闭弹窗
const handleClose = () => {
  emit('update:show', false)
}

// 保存设置
const handleSave = async () => {
  try {
    // 验证表单
    await formRef.value?.validate()
    
    loading.value = true
    
    // 保存基本信息
    const profileResult = await userStore.updateProfile({
      username: formData.username,
      email: formData.email
    })

    if (!profileResult.success) {
      message.error(profileResult.message || t('userSettings.saveFailed'))
      return
    }

    // 保存开发设置
    const settingsResult = await userStore.updateUserSettings({
      default_cli_tool: formData.defaultCliTool,
      default_ai_model: formData.defaultAiModel,
      default_model_provider: formData.defaultModelProvider,
      default_model_api_url: formData.defaultModelApiUrl,
      default_api_token: formData.defaultApiToken
    })

    if (settingsResult.success) {
      message.success(t('userSettings.saveSuccess'))
      handleClose()
    } else {
      message.error(settingsResult.message || t('userSettings.saveFailed'))
    }
  } catch (error: any) {
    console.error('保存用户设置失败:', error)
    message.error(t('userSettings.networkError'))
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.binding-status {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
}

.binding-text {
  color: var(--text-secondary);
  font-size: 0.9rem;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: var(--spacing-md);
}

/* 自定义模态框样式 */
:deep(.n-modal) {
  border-radius: var(--border-radius-lg);
}

:deep(.n-modal .n-card__header) {
  border-bottom: 1px solid var(--border-color);
  padding: var(--spacing-lg);
}

:deep(.n-modal .n-card__content) {
  padding: var(--spacing-lg);
}

:deep(.n-modal .n-card__footer) {
  border-top: 1px solid var(--border-color);
  padding: var(--spacing-lg);
  background: var(--background-color);
}

/* 表单样式优化 */
:deep(.n-form-item) {
  margin-bottom: var(--spacing-lg);
}

:deep(.n-form-item-label) {
  font-weight: 500;
  color: var(--text-primary);
}

:deep(.n-input) {
  border-radius: var(--border-radius-md);
}

:deep(.n-button) {
  border-radius: var(--border-radius-md);
  font-weight: 500;
}
</style>
