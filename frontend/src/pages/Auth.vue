<template>
  <div class="auth-page">
    <!-- 背景装饰 -->
    <div class="auth-background">
      <div class="background-overlay"></div>
    </div>

    <!-- 主要内容区域 -->
    <div class="auth-container">
      <!-- 认证表单区域 -->
      <div class="auth-form-container">
        <div class="form-header">
          <h2>{{ isLogin ? t('auth.welcomeBack') : t('auth.createAccount') }}</h2>
        </div>

        <!-- 切换按钮 -->
        <div class="auth-toggle">
          <n-button
            :type="isLogin ? 'primary' : 'default'"
            :ghost="!isLogin"
            @click="isLogin = true"
            class="toggle-btn"
          >
            {{ t('auth.login') }}
          </n-button>
          <n-button
            :type="!isLogin ? 'primary' : 'default'"
            :ghost="isLogin"
            @click="isLogin = false"
            class="toggle-btn"
          >
            {{ t('auth.register') }}
          </n-button>
        </div>

        <!-- 表单 -->
        <n-form
          ref="formRef"
          :model="formData"
          :rules="formRules"
          @submit.prevent="handleSubmit"
          class="auth-form"
        >
          <!-- 邮箱 -->
          <n-form-item
            :label="t('auth.email')"
            path="email"
            class="form-item"
          >
            <n-input
              v-model:value="formData.email"
              type="text"
              :placeholder="t('auth.emailPlaceholder')"
              size="large"
              clearable
              class="form-input"
            >
              <template #prefix>
                <n-icon size="16"><MailIcon /></n-icon>
              </template>
            </n-input>
          </n-form-item>
          
          <!-- 用户名(仅注册时显示) -->
          <n-form-item
            v-if="!isLogin"
            :label="t('auth.username')"
            path="username"
            class="form-item"
          >
            <n-input
              v-model:value="formData.username"
              :placeholder="t('auth.usernamePlaceholder')"
              size="large"
              clearable
              class="form-input"
            >
              <template #prefix>
                <n-icon size="16"><UserIcon /></n-icon>
              </template>
            </n-input>
          </n-form-item>

          <!-- 密码 -->
          <n-form-item
            :label="t('auth.password')"
            path="password"
            class="form-item"
          >
            <n-input
              v-model:value="formData.password"
              type="password"
              :placeholder="t('auth.passwordPlaceholder')"
              size="large"
              show-password-on="click"
              clearable
              class="form-input"
            >
              <template #prefix>
                <n-icon size="16"><LockIcon /></n-icon>
              </template>
            </n-input>
          </n-form-item>

          <!-- 确认密码（仅注册时显示） -->
          <n-form-item
            v-if="!isLogin"
            :label="t('auth.confirmPassword')"
            path="confirmPassword"
            class="form-item"
          >
            <n-input
              v-model:value="formData.confirmPassword"
              type="password"
              :placeholder="t('auth.confirmPasswordPlaceholder')"
              size="large"
              show-password-on="click"
              clearable
              class="form-input"
            >
              <template #prefix>
                <n-icon size="16"><LockIcon /></n-icon>
              </template>
            </n-input>
          </n-form-item>

          <!-- 记住我（仅登录时显示） -->
          <div v-if="isLogin" class="form-options">
            <n-checkbox v-model:checked="formData.rememberMe">
              {{ t('auth.rememberMe') }}
            </n-checkbox>
            <n-button text type="primary" @click="forgotPassword">
              {{ t('auth.forgotPassword') }}
            </n-button>
          </div>

          <!-- 提交按钮 -->
          <n-button
            type="primary"
            size="large"
            :loading="loading"
            @click="handleSubmit"
            class="submit-btn"
            block
          >
            {{ isLogin ? t('auth.login') : t('auth.register') }}
          </n-button>

          <!-- 协议同意（仅注册时显示） -->
          <div v-if="!isLogin" class="agreement">
            <n-checkbox v-model:checked="formData.agreeTerms">
              {{ t('auth.agreeTerms') }}
              <n-button text type="primary" @click="showTerms">
                {{ t('auth.userAgreement') }}
              </n-button>
              {{ t('common.and') }}
              <n-button text type="primary" @click="showPrivacy">
                {{ t('auth.privacyPolicy') }}
              </n-button>
            </n-checkbox>
          </div>
        </n-form>

        <!-- 社交登录 -->
        <div class="social-login">
          <div class="social-buttons">
            <n-button
              ghost
              size="large"
              @click="socialLogin('github')"
              class="social-btn"
            >
              <template #icon>
                <n-icon size="16"><GithubIcon /></n-icon>
              </template>
              GitHub
            </n-button>
            <n-button
              ghost
              size="large"
              @click="socialLogin('google')"
              class="social-btn"
            >
              <template #icon>
                <n-icon size="16"><GoogleIcon /></n-icon>
              </template>
              Google
            </n-button>
          </div>
        </div>
      </div>
    </div>

    <!-- 协议弹窗 -->
    <n-modal v-model:show="showTermsModal" preset="card" :title="t('auth.userAgreement')" style="width: 600px">
      <div class="terms-content">
        <h3>{{ t('auth.userAgreement') }}</h3>
        <p>{{ t('auth.welcomeToAppMaker') }}</p>
        <p>{{ t('auth.termsDescription') }}</p>
        <!-- 更多协议内容 -->
      </div>
    </n-modal>

    <n-modal v-model:show="showPrivacyModal" preset="card" :title="t('auth.privacyPolicy')" style="width: 600px">
      <div class="privacy-content">
        <h3>{{ t('auth.privacyPolicy') }}</h3>
        <p>{{ t('auth.privacyImportance') }}</p>
        <p>{{ t('auth.privacyDescription') }}</p>
        <!-- 更多隐私政策内容 -->
      </div>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, h, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useMessage } from 'naive-ui'
import { useUserStore } from '@/stores/user'
import {
  NForm, NFormItem, NInput, NButton, NCheckbox, NIcon, NModal
} from 'naive-ui'

// 调试信息
onMounted(() => {
  console.log('Auth 页面已挂载')
  console.log('用户状态:', {
    isAuthenticated: userStore.isAuthenticated,
    hasToken: !!userStore.token,
    hasUser: !!userStore.user
  })
  console.log('页面元素检查:', {
    authPage: document.querySelector('.auth-page'),
    authContainer: document.querySelector('.auth-container'),
    authFormContainer: document.querySelector('.auth-form-container')
  })
})

// 图标组件 - 使用简单的 emoji 图标，避免外部依赖
const CodeIcon = () => h('span', { style: 'font-size: 20px;' }, '💻')
const UserIcon = () => h('span', { style: 'font-size: 16px;' }, '👤')
const LockIcon = () => h('span', { style: 'font-size: 16px;' }, '🔒')
const MailIcon = () => h('span', { style: 'font-size: 16px;' }, '📧')
const GithubIcon = () => h('span', { style: 'font-size: 16px;' }, '🐙')
const GoogleIcon = () => h('span', { style: 'font-size: 16px;' }, '🔍')

const router = useRouter()
const userStore = useUserStore()
const { t } = useI18n()

// 获取 message 实例
const message = useMessage()

// 响应式数据
const isLogin = ref(true)
const loading = ref(false)
const formRef = ref()
const showTermsModal = ref(false)
const showPrivacyModal = ref(false)

// 表单数据
const formData = reactive({
  username: '',
  email: '',
  password: '',
  confirmPassword: '',
  rememberMe: false,
  agreeTerms: false
})

// 表单验证规则
const formRules = computed(() => ({
  username: isLogin.value ? [] : [
    {
      required: true,
      message: t('auth.usernameRequired'),
      trigger: 'blur'
    },
    {
      validator: (rule: any, value: string) => {        
        // 注册时验证用户名格式
        if (value.length < 3) {
          return new Error(t('auth.usernameMinLength'))
        }
        if (value.length > 20) {
          return new Error(t('auth.usernameMaxLength'))
        }        
      },
      trigger: 'blur'
    }
  ],
  email: [
    {
      required: true,
      message: t('auth.emailRequired'),
      trigger: 'blur'
    },
    {
      validator: (rule: any, value: string) => {
        const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
        if (!emailRegex.test(value)) {
          return new Error(t('auth.emailFormatError'))
        }
      },
      trigger: 'blur'
    }
  ],
  password: [
    {
      required: true,
      message: t('auth.passwordRequired'),
      trigger: 'blur'
    },
    {
      min: 6,
      message: t('auth.passwordMinLength'),
      trigger: 'blur'
    }
  ],
  confirmPassword: isLogin.value ? [] : [
    {
      required: true,
      message: t('auth.confirmPasswordRequired'),
      trigger: 'blur'
    },
    {
      validator: (rule: any, value: string) => {
        if (value !== formData.password) {
          return new Error(t('auth.passwordMismatch'))
        }
      },
      trigger: 'blur'
    }
  ]
}))

// 方法
const handleSubmit = async () => {
  try {
    await formRef.value?.validate()
    loading.value = true

    if (isLogin.value) {
      // 登录逻辑
      const loginData = {
        email: formData.email,
        password: formData.password
      }
      
      const result = await userStore.login(loginData)
      if (result.success) {
        message.success(t('auth.loginSuccess'))
        router.push('/dashboard')
      } else {
        message.error(result.message || t('auth.loginFailed'))
      }
    } else {
      // 注册逻辑
      if (!formData.agreeTerms) {
        message.warning(t('auth.agreeTermsRequired'))
        return
      }

      const registerData = {
        username: formData.username,
        email: formData.email,
        password: formData.password
      }
      
      const result = await userStore.register(registerData)
      if (result.success) {
        message.success(t('auth.registerSuccess'))
        // 注册成功后直接跳转到创建项目页面，不需要再次登录
        router.push('/create-project')
      } else {
        message.error(result.message || t('auth.registerFailed'))
      }
    }
  } catch (error) {
    console.error('表单验证失败:', error)
  } finally {
    loading.value = false
  }
}

const forgotPassword = () => {
  message.info(t('auth.forgotPasswordFeature'))
}

const showTerms = () => {
  showTermsModal.value = true
}

const showPrivacy = () => {
  showPrivacyModal.value = true
}

const socialLogin = (provider: string) => {
  message.info(t('auth.socialLoginFeature', { provider }))
}
</script>

<style scoped>
.auth-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  /* 使用更丰富的渐变背景 */
  background: linear-gradient(135deg, #667eea 0%, #764ba2 50%, #f093fb 100%);
  overflow: hidden;
}

/* 背景装饰 */
.auth-background {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  /* 移除不存在的背景图片，使用纯色渐变 */
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  z-index: 0;
}

.background-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.4);
  backdrop-filter: blur(10px);
}

/* 主容器 */
.auth-container {
  position: relative;
  z-index: 1;
  width: 100%;
  max-width: 1200px;
  margin: 0 auto;
  padding: var(--spacing-xl);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--spacing-xxl);
}

/* 表单容器 */
.auth-form-container {
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(20px);
  border-radius: var(--border-radius-xl);
  padding: var(--spacing-xxl);
  box-shadow: var(--shadow-xl);
  border: 1px solid rgba(255, 255, 255, 0.2);
  width: 100%;
  max-width: 480px;
}

.form-header {
  text-align: center;
  margin-bottom: var(--spacing-xl);
}

.form-header h2 {
  color: var(--primary-color);
  font-size: 1.5rem;
  font-weight: bold;
  margin: 0 0 var(--spacing-sm) 0;
}

/* 切换按钮 */
.auth-toggle {
  display: flex;
  gap: var(--spacing-sm);
  margin-bottom: var(--spacing-xl);
  background: var(--background-color);
  border-radius: var(--border-radius-lg);
}

.toggle-btn {
  flex: 1;
  border-radius: var(--border-radius-md);
}

/* 表单 */
.auth-form {
  margin-bottom: var(--spacing-xl);
}

.form-item :deep(.n-form-item-label) {
  color: var(--text-primary);
  font-weight: 500;
  font-size: 0.9rem;
}

.form-input {
  border-radius: var(--border-radius-md);
  border: 1px solid var(--border-color);
  transition: all 0.3s ease;
}

.form-input .n-icon {
  font-style: normal !important;
  margin-right: 4px;
}

.n-input .n-input__input-el {
  padding-left: 4px;
}

.form-input:focus-within {
  border-color: var(--primary-color);
  box-shadow: 0 0 0 2px rgba(49, 130, 206, 0.2);
}

/* 表单选项 */
.form-options {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-lg);
  font-size: 0.9rem;
}

/* 提交按钮 */
.submit-btn {
  background: linear-gradient(135deg, var(--primary-color), var(--accent-color));
  border: none;
  border-radius: var(--border-radius-md);
  font-weight: 600;
  font-size: 1rem;
  height: var(--height-md);
  transition: all 0.3s ease;
}

.submit-btn:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-lg);
}

/* 协议同意 */
.agreement {
  margin-top: var(--spacing-lg);
  text-align: center;
  font-size: 0.8rem;
  color: var(--text-secondary);
}

.agreement :deep(.n-checkbox) {
  font-size: 0.8rem;
}

/* 社交登录 */
.social-login {
  text-align: center;
}

.social-buttons {
  display: flex;
  gap: var(--spacing-md);
  justify-content: center;
}

.social-btn {
  flex: 1;
  max-width: 160px;
  border-radius: var(--border-radius-md);
  border: 1px solid var(--border-color);
  transition: all 0.3s ease;
}

.social-btn:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-md);
}

.social-btn .n-icon{
  font-style: normal !important;
}

/* 弹窗内容 */
.terms-content,
.privacy-content {
  max-height: 400px;
  overflow-y: auto;
  line-height: 1.6;
}

.terms-content h3,
.privacy-content h3 {
  color: var(--primary-color);
  margin-bottom: var(--spacing-md);
}

/* 响应式设计 */
@media (max-width: 768px) {
  .auth-container {
    padding: var(--spacing-lg);
  }
  
  .auth-form-container {
    padding: var(--spacing-xl);
    margin: 0 var(--spacing-md);
  }
  
  .logo-text {
    font-size: 1.5rem;
  }
  
  .social-buttons {
    flex-direction: column;
  }
  
  .social-btn {
    max-width: none;
  }
}

@media (max-width: 480px) {
  .auth-container {
    padding: var(--spacing-md);
  }
  
  .auth-form-container {
    padding: var(--spacing-lg);
  }
  
  .form-options {
    flex-direction: column;
    gap: var(--spacing-sm);
    align-items: flex-start;
  }
}
</style>
