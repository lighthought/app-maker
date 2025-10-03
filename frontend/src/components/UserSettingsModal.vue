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
import { ref, reactive, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useMessage, NModal, NForm, NFormItem, NInput, NButton, NTag, type FormRules } from 'naive-ui'
import { useUserStore } from '@/stores/user'
import { httpService } from '@/utils/http'

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

// 表单数据
const formData = reactive({
  username: '',
  email: ''
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
watch(() => props.show, (newVal) => {
  if (newVal && userStore.user) {
    formData.username = userStore.user.username || ''
    formData.email = userStore.user.email || ''
  }
})

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
    
    // 调用后端接口
    const response = await httpService.put<{
      code: number
      message: string
      data?: any
    }>('/users/profile', {
      username: formData.username,
      email: formData.email
    })

    if (response.code === 0) {
      message.success(t('userSettings.saveSuccess'))
      
      // 更新用户 store 中的用户信息
      if (userStore.user) {
        userStore.user.username = formData.username
        userStore.user.email = formData.email
        localStorage.setItem('user', JSON.stringify(userStore.user))
      }
      
      handleClose()
    } else {
      message.error(response.message || t('userSettings.saveFailed'))
    }
  } catch (error: any) {
    console.error('保存用户设置失败:', error)
    
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
