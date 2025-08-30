<template>
  <div class="auth-page">
    <!-- 背景装饰 -->
    <div class="auth-background">
      <div class="background-overlay"></div>
    </div>

    <!-- 主要内容区域 -->
    <div class="auth-container">
      <!-- Logo 区域 -->
      <div class="logo-section">
        <div class="logo-container">
          <div class="logo-icon">
            <n-icon size="48" color="#3182CE">
              <CodeIcon />
            </n-icon>
          </div>
          <h1 class="logo-text">煲应用 - AutoCode</h1>
        </div>
      </div>

      <!-- 认证表单区域 -->
      <div class="auth-form-container">
        <div class="form-header">
          <h2>{{ isLogin ? '欢迎回来' : '创建账户' }}</h2>
          <p>{{ isLogin ? '登录您的账户继续使用' : '开始您的项目之旅' }}</p>
        </div>

        <!-- 切换按钮 -->
        <div class="auth-toggle">
          <n-button
            :type="isLogin ? 'primary' : 'default'"
            :ghost="!isLogin"
            @click="isLogin = true"
            class="toggle-btn"
          >
            登录
          </n-button>
          <n-button
            :type="!isLogin ? 'primary' : 'default'"
            :ghost="isLogin"
            @click="isLogin = false"
            class="toggle-btn"
          >
            注册
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
          <!-- 用户名/邮箱 -->
          <n-form-item
            :label="isLogin ? '邮箱' : '用户名'"
            path="username"
            class="form-item"
          >
            <n-input
              v-model:value="formData.username"
              :placeholder="isLogin ? '请输入邮箱' : '请输入用户名'"
              size="large"
              clearable
              class="form-input"
            >
              <template #prefix>
                <n-icon><UserIcon /></n-icon>
              </template>
            </n-input>
          </n-form-item>

          <!-- 密码 -->
          <n-form-item
            label="密码"
            path="password"
            class="form-item"
          >
            <n-input
              v-model:value="formData.password"
              type="password"
              placeholder="请输入密码"
              size="large"
              show-password-on="click"
              clearable
              class="form-input"
            >
              <template #prefix>
                <n-icon><LockIcon /></n-icon>
              </template>
            </n-input>
          </n-form-item>

          <!-- 确认密码（仅注册时显示） -->
          <n-form-item
            v-if="!isLogin"
            label="确认密码"
            path="confirmPassword"
            class="form-item"
          >
            <n-input
              v-model:value="formData.confirmPassword"
              type="password"
              placeholder="请再次输入密码"
              size="large"
              show-password-on="click"
              clearable
              class="form-input"
            >
              <template #prefix>
                <n-icon><LockIcon /></n-icon>
              </template>
            </n-input>
          </n-form-item>

          <!-- 邮箱（仅注册时显示） -->
          <n-form-item
            v-if="!isLogin"
            label="邮箱"
            path="email"
            class="form-item"
          >
            <n-input
              v-model:value="formData.email"
              type="text"
              placeholder="请输入邮箱"
              size="large"
              clearable
              class="form-input"
            >
              <template #prefix>
                <n-icon><MailIcon /></n-icon>
              </template>
            </n-input>
          </n-form-item>

          <!-- 记住我（仅登录时显示） -->
          <div v-if="isLogin" class="form-options">
            <n-checkbox v-model:checked="formData.rememberMe">
              记住我
            </n-checkbox>
            <n-button text type="primary" @click="forgotPassword">
              忘记密码？
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
            {{ isLogin ? '登录' : '注册' }}
          </n-button>

          <!-- 协议同意（仅注册时显示） -->
          <div v-if="!isLogin" class="agreement">
            <n-checkbox v-model:checked="formData.agreeTerms">
              注册即表示您同意我们的
              <n-button text type="primary" @click="showTerms">
                《用户协议》
              </n-button>
              和
              <n-button text type="primary" @click="showPrivacy">
                《隐私政策》
              </n-button>
            </n-checkbox>
          </div>
        </n-form>

        <!-- 社交登录 -->
        <div class="social-login">
          <div class="divider">
            <span>或</span>
          </div>
          <div class="social-buttons">
            <n-button
              ghost
              size="large"
              @click="socialLogin('github')"
              class="social-btn"
            >
              <template #icon>
                <n-icon><GithubIcon /></n-icon>
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
                <n-icon><GoogleIcon /></n-icon>
              </template>
              Google
            </n-button>
          </div>
        </div>
      </div>
    </div>

    <!-- 协议弹窗 -->
    <n-modal v-model:show="showTermsModal" preset="card" title="用户协议" style="width: 600px">
      <div class="terms-content">
        <h3>用户协议</h3>
        <p>欢迎使用煲应用 - AutoCode！</p>
        <p>本协议是您与煲应用平台之间的法律协议，请您仔细阅读。</p>
        <!-- 更多协议内容 -->
      </div>
    </n-modal>

    <n-modal v-model:show="showPrivacyModal" preset="card" title="隐私政策" style="width: 600px">
      <div class="privacy-content">
        <h3>隐私政策</h3>
        <p>我们非常重视您的隐私保护。</p>
        <p>本政策说明了我们如何收集、使用和保护您的个人信息。</p>
        <!-- 更多隐私政策内容 -->
      </div>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, h } from 'vue'
import { useRouter } from 'vue-router'
import { useMessage } from 'naive-ui'
import { useUserStore } from '@/stores/user'
import {
  NForm, NFormItem, NInput, NButton, NCheckbox, NIcon, NModal
} from 'naive-ui'

// 图标组件 - 使用简单的 emoji 图标，避免外部依赖
const CodeIcon = () => h('span', { style: 'font-size: 20px;' }, '💻')
const UserIcon = () => h('span', { style: 'font-size: 16px;' }, '👤')
const LockIcon = () => h('span', { style: 'font-size: 16px;' }, '🔒')
const MailIcon = () => h('span', { style: 'font-size: 16px;' }, '📧')
const GithubIcon = () => h('span', { style: 'font-size: 16px;' }, '🐙')
const GoogleIcon = () => h('span', { style: 'font-size: 16px;' }, '🔍')

const router = useRouter()
const userStore = useUserStore()

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
  username: [
    {
      required: true,
      message: isLogin.value ? '请输入邮箱' : '请输入用户名',
      trigger: 'blur'
    },
    {
      validator: (rule: any, value: string) => {
        if (isLogin.value) {
          // 登录时验证邮箱格式
          const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
          if (!emailRegex.test(value)) {
            return new Error('请输入有效的邮箱地址')
          }
        } else {
          // 注册时验证用户名格式
          if (value.length < 3) {
            return new Error('用户名至少需要3个字符')
          }
          if (value.length > 20) {
            return new Error('用户名不能超过20个字符')
          }
        }
      },
      trigger: 'blur'
    }
  ],
  email: isLogin.value ? [] : [
    {
      required: true,
      message: '请输入邮箱',
      trigger: 'blur'
    },
    {
      validator: (rule: any, value: string) => {
        const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
        if (!emailRegex.test(value)) {
          return new Error('请输入有效的邮箱地址')
        }
      },
      trigger: 'blur'
    }
  ],
  password: [
    {
      required: true,
      message: '请输入密码',
      trigger: 'blur'
    },
    {
      min: 6,
      message: '密码至少需要6个字符',
      trigger: 'blur'
    }
  ],
  confirmPassword: isLogin.value ? [] : [
    {
      required: true,
      message: '请确认密码',
      trigger: 'blur'
    },
    {
      validator: (rule: any, value: string) => {
        if (value !== formData.password) {
          return new Error('两次输入的密码不一致')
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
        email: formData.username, // 登录时使用邮箱
        password: formData.password
      }
      
      const result = await userStore.login(loginData)
      if (result.success) {
        message.success('登录成功')
        router.push('/dashboard')
      } else {
        message.error(result.message || '登录失败')
      }
    } else {
      // 注册逻辑
      if (!formData.agreeTerms) {
        message.warning('请先同意用户协议和隐私政策')
        return
      }

      const registerData = {
        username: formData.username,
        email: formData.email,
        password: formData.password
      }
      
      const result = await userStore.register(registerData)
      if (result.success) {
        message.success('注册成功')
        isLogin.value = true
        // 清空表单
        Object.assign(formData, {
          username: '',
          email: '',
          password: '',
          confirmPassword: '',
          rememberMe: false,
          agreeTerms: false
        })
      } else {
        message.error(result.message || '注册失败')
      }
    }
  } catch (error) {
    console.error('表单验证失败:', error)
  } finally {
    loading.value = false
  }
}

const forgotPassword = () => {
  message.info('密码重置功能开发中...')
}

const showTerms = () => {
  showTermsModal.value = true
}

const showPrivacy = () => {
  showPrivacyModal.value = true
}

const socialLogin = (provider: string) => {
  message.info(`${provider} 登录功能开发中...`)
}
</script>

<style scoped>
.auth-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  background: linear-gradient(135deg, var(--primary-color), var(--accent-color));
  overflow: hidden;
}

/* 背景装饰 */
.auth-background {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: url('/images/auth-bg.jpg') center/cover;
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

/* Logo 区域 */
.logo-section {
  text-align: center;
}

.logo-container {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--spacing-md);
}

.logo-text {
  color: white;
  font-size: 2rem;
  font-weight: bold;
  margin: 0;
  text-shadow: 0 2px 4px rgba(0, 0, 0, 0.3);
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

.form-header p {
  color: var(--text-secondary);
  margin: 0;
  font-size: 0.9rem;
}

/* 切换按钮 */
.auth-toggle {
  display: flex;
  gap: var(--spacing-sm);
  margin-bottom: var(--spacing-xl);
  background: var(--background-color);
  padding: var(--spacing-sm);
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

.form-item {
  margin-bottom: var(--spacing-lg);
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
  height: 48px;
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

.divider {
  position: relative;
  margin: var(--spacing-lg) 0;
  text-align: center;
}

.divider::before {
  content: '';
  position: absolute;
  top: 50%;
  left: 0;
  right: 0;
  height: 1px;
  background: var(--border-color);
}

.divider span {
  background: white;
  padding: 0 var(--spacing-md);
  color: var(--text-secondary);
  font-size: 0.9rem;
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
