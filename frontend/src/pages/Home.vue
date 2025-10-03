<template>
  <div class="home-page">
    <!-- 顶部导航栏 -->
    <header class="header" :class="{ 'header-scrolled': isScrolled }">
      <div class="header-container">
        <div class="logo">
          <img src="@/assets/logo.svg" alt="App-Maker" class="logo-icon" />
          <h1>App-Maker</h1>
        </div>
        <nav class="nav">
          <a href="#process" class="nav-link">{{ t('process.title') }}</a>
          <a href="#features" class="nav-link">{{ t('features.title') }}</a>          
          <a href="#about" class="nav-link">{{ t('about.title') }}</a>
        </nav>
        <div class="header-actions">
          <n-button
            size="small"
            @click="toggleLanguage"
            class="language-btn"
          >
            {{ locale === 'zh' ? 'EN' : '中文' }}
          </n-button>
          <a :href="isLoggedIn ? '/dashboard' : '/auth'" class="experience-btn">
            <n-button type="primary">
              {{ isLoggedIn ? t('buttons.enterConsole') : t('buttons.experience') }}
            </n-button>
          </a>
        </div>
      </div>
    </header>

    <!-- Hero 区域 -->
    <section class="hero">
      <div class="hero-container">
        <div class="hero-content">
          <h1 class="hero-title">
            {{ t('hero.title') }}
          </h1>
          <p class="hero-subtitle">
            {{ t('hero.subtitle') }}
          </p>
          
          <!-- 智能输入框 -->
          <div class="hero-input">
            <SmartInput
              v-model="projectDescription"
              :placeholder="t('hero.inputPlaceholder')"
              size="large"
              @send="handleProjectCreate"
            />
          </div>

          <!-- 用户项目展示 -->
          <div v-if="isLoggedIn && userProjects.length > 0" class="user-projects">
            <h3>{{ t('hero.recentProjects') }}</h3>
            <div class="project-cards">
              <div
                v-for="project in userProjects.slice(0, 5)"
                :key="project.guid"
                class="project-card"
                @click="goToProject(project.guid)"
              >
                <h4>{{ project.name }}</h4>
                <p>{{ project.description }}</p>
                <div class="project-status">
                  <n-tag :type="getStatusType(project.status)">
                    {{ getStatusText(project.status) }}
                  </n-tag>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- 使用流程说明 -->
    <section id="process" class="process">
      <div class="container">
        <h2 class="section-title">{{ t('process.title') }}</h2>
        <div class="process-timeline">
          <div
            v-for="(step, index) in processSteps"
            :key="step.id"
            class="process-step"
          >
            <div class="step-number">{{ index + 1 }}</div>
            <div class="step-content">
              <h3>{{ step.title() }}</h3>
              <p>{{ step.description() }}</p>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- 功能特性展示 -->
    <section id="features" class="features">
      <div class="container">
        <h2 class="section-title">{{ t('features.title') }}</h2>
        <div class="features-grid">
          <div class="feature-card" v-for="feature in features" :key="feature.id">
            <div class="feature-icon">
              <component :is="feature.icon" />
            </div>
            <h3>{{ feature.title() }}</h3>
            <p>{{ feature.description() }}</p>
          </div>
        </div>
      </div>
    </section>

    <!-- 底部信息 -->
    <footer id="about" class="footer">
      <div class="container">
        <div class="footer-content">
          <div class="footer-section">
            <h3>App-Maker</h3>
            <p>{{ t('footer.description') }}</p>
          </div>
          <div class="footer-section">
            <h4>{{ t('footer.contact') }}</h4>
            <p>{{ t('footer.email') }}: qqjack2012@gmail.com</p>
            <p>{{ t('footer.account') }}: AI 探趣星船长</p>
          </div>
          <div class="footer-section">
            <h4>{{ t('footer.follow') }}</h4>
            <div class="social-links">
              <a href="https://github.com/lighthought/app-maker" target="_blank" rel="noopener noreferrer" class="social-link">GitHub</a>
              <a href="https://www.xiaohongshu.com/user/profile/62033e59000000001000aa0d" target="_blank" rel="noopener noreferrer" class="social-link">{{ t('footer.xiaohongshu') }}</a>
              <a href="https://space.bilibili.com/44060402" target="_blank" rel="noopener noreferrer" class="social-link">{{ t('footer.bilibili') }}</a>
              <a href="https://www.douyin.com/user/MS4wLjABAAAA9Dl00eOUp2iD1zKY-Gdlr1uGovve-8ky7Fntl_kM5VA" target="_blank" rel="noopener noreferrer" class="social-link">{{ t('footer.douyin') }}</a>
            </div>
          </div>
        </div>
        <div class="footer-bottom">
          <p>&copy; 2025 thought-light.com. {{ t('footer.rights') }}</p>
        </div>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, h } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useUserStore } from '@/stores/user'
import { useProjectStore } from '@/stores/project'
import { NButton, NIcon, NTag } from 'naive-ui'
import SmartInput from '@/components/common/SmartInput.vue'

// 图标组件（使用 div 标签避免倾斜）
const CodeIcon = () => h('div', { 
  style: 'font-size: 48px; line-height: 1; text-align: center;'
}, '💻')
const RobotIcon = () => h('div', { 
  style: 'font-size: 48px; line-height: 1; text-align: center;'
}, '🤖')
const RocketIcon = () => h('div', { 
  style: 'font-size: 48px; line-height: 1; text-align: center;'
}, '🚀')
const ShieldIcon = () => h('div', { 
  style: 'font-size: 48px; line-height: 1; text-align: center;'
}, '🛡️')
const RepositoryIcon = () => h('div', { 
  style: 'font-size: 48px; line-height: 1; text-align: center;'
}, '📂')
const ZapIcon = () => h('div', { 
  style: 'font-size: 48px; line-height: 1; text-align: center;'
}, '⚡')

const router = useRouter()
const userStore = useUserStore()
const projectStore = useProjectStore()
const { locale, t } = useI18n()

// 响应式数据
const isScrolled = ref(false)
const projectDescription = ref('')
const currentStep = ref(0)

// 计算属性
const isLoggedIn = computed(() => userStore.isAuthenticated)
const userProjects = computed(() => projectStore.projects.slice(0, 5))
// 文本使用新的 i18n t 函数

// 功能特性数据
const features = ref([
  {
    id: 1,
    icon: CodeIcon,
    title: () => t('features.smartCodeGeneration'),
    description: () => t('features.smartCodeGenerationDescription')
  },
  {
    id: 2,
    icon: RobotIcon,
    title: () => t('features.multiAgentCollaboration'),
    description: () => t('features.multiAgentCollaborationDescription')
  },
  {
    id: 3,
    icon: RocketIcon,
    title: () => t('features.fastDeployment'),
    description: () => t('features.fastDeploymentDescription')
  },
  {
    id: 4,
    icon: ShieldIcon,
    title: () => t('features.secureReliable'),
    description: () => t('features.secureReliableDescription')
  },
  {
    id: 5,
    icon: RepositoryIcon,
    title: () => t('features.codeRepository'),
    description: () => t('features.codeRepositoryDescription')
  },
  {
    id: 6,
    icon: ZapIcon,
    title: () => t('features.efficientDevelopment'),
    description: () => t('features.efficientDevelopmentDescription')
  }
])

// 使用流程数据
const processSteps = ref([
  {
    id: 1,
    title: () => t('process.describe'),
    description: () => t('process.describeDescription')
  },
  {
    id: 2,
    title: () => t('process.agentAnalysis'),
    description: () => t('process.agentAnalysisDescription')
  },
  {
    id: 3,
    title: () => t('process.generateCode'),
    description: () => t('process.generateCodeDescription')
  },
  {
    id: 4,
    title: () => t('process.testDeploy'),
    description: () => t('process.testDeployDescription')
  }
])

// 方法
const toggleLanguage = () => {
  locale.value = locale.value === 'zh' ? 'en' : 'zh'
  localStorage.setItem('preferred-language', locale.value)
}

const handleProjectCreate = async () => {
  if (!projectDescription.value.trim()) return
  
  if (!isLoggedIn.value) {
    // 未登录用户跳转到登录页面，并保存输入内容
    localStorage.setItem('pendingProjectDescription', projectDescription.value)
    router.push('/auth')
    return
  }
  
  // 已登录用户直接跳转到创建项目页面
  router.push({
    path: '/create-project',
    query: { description: projectDescription.value }
  })
}

const goToProject = (projectGuid: string) => {
  router.push(`/project/${projectGuid}`)
}

const getStatusType = (status: string): 'default' | 'error' | 'warning' | 'success' | 'primary' | 'info' => {
  const statusMap: Record<string, 'default' | 'error' | 'warning' | 'success' | 'primary' | 'info'> = {
    draft: 'default',
    in_progress: 'warning',
    done: 'success',
    failed: 'error'
  }
  return statusMap[status] || 'default'
}

const getStatusText = (status: string) => {
  const statusMap: Record<string, string> = {
    draft: t('common.draft'),
    in_progress: t('common.inProgress'),
    done: t('common.completed'),
    failed: t('common.failed')
  }
  return statusMap[status] || status
}

// 滚动监听
const handleScroll = () => {
  isScrolled.value = window.scrollY > 50
}

// 生命周期
onMounted(() => {
  window.addEventListener('scroll', handleScroll)
  
  // 检查是否有待创建的项目描述
  const pendingDescription = localStorage.getItem('pendingProjectDescription')
  if (pendingDescription && isLoggedIn.value) {
    projectDescription.value = pendingDescription
    localStorage.removeItem('pendingProjectDescription')
  }
})

onUnmounted(() => {
  window.removeEventListener('scroll', handleScroll)
})
</script>

<style scoped>
.home-page {
  min-height: 100vh;
  background: linear-gradient(135deg, var(--primary-color) 0%, var(--accent-color) 100%);
}

/* 顶部导航栏 */
.header {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 1000;
  background: rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(10px);
  border-bottom: 1px solid rgba(255, 255, 255, 0.2);
  transition: all 0.3s ease;
}

.header-scrolled {
  background: rgba(255, 255, 255, 0.95);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.header-container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 var(--spacing-lg);
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 64px;
}

.logo {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
}

.logo-icon {
  width: 32px;
  height: 32px;
  flex-shrink: 0;
}

.logo h1 {
  color: white;
  font-size: 1.5rem;
  font-weight: bold;
  margin: 0;
}

.header-scrolled .logo h1 {
  color: var(--primary-color);
}

.nav {
  display: flex;
  gap: var(--spacing-lg);
}

.nav-link {
  color: white;
  text-decoration: none;
  font-weight: 500;
  transition: color 0.3s ease;
}

.header-scrolled .nav-link {
  color: var(--primary-color);
}

.nav-link:hover {
  color: var(--accent-color);
}

.header-actions {
  display: flex;
  gap: var(--spacing-md);
  align-items: center;
}

.language-btn {
  background: rgba(255, 255, 255, 0.2);
  border: 1px solid rgba(255, 255, 255, 0.3);
  color: white;
  transition: all 0.3s ease;
  width: 50px;
}

.header-scrolled .language-btn {
  background: var(--background-color);
  border: 1px solid var(--border-color);
  color: var(--text-primary);
}

.experience-btn {
  text-decoration: none;
  display: inline-block;
  width: 98px;
}

.experience-btn .n-button {
  background: var(--accent-color);
  border: none;
  color: white;
  font-weight: 600;
  width: 100%;
}

/* Hero 区域 */
.hero {
  padding: 120px 0 80px;
  text-align: center;
  color: white;
}

.hero-container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 var(--spacing-lg);
}

.hero-title {
  font-size: 3rem;
  font-weight: bold;
  margin-bottom: var(--spacing-lg);
  line-height: 1.2;
}

.hero-subtitle {
  font-size: 1.25rem;
  margin-bottom: var(--spacing-xxl);
  opacity: 0.9;
  max-width: 600px;
  margin-left: auto;
  margin-right: auto;
}

.hero-input {
  max-width: 600px;
  margin: 0 auto var(--spacing-xxl);
}

.user-projects {
  margin-top: var(--spacing-xxl);
  text-align: left;
}

.user-projects h3 {
  margin-bottom: var(--spacing-lg);
  font-size: 1.5rem;
}

.project-cards {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: var(--spacing-lg);
}

.project-card {
  background: rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: var(--border-radius-lg);
  padding: var(--spacing-lg);
  cursor: pointer;
  transition: all 0.3s ease;
}

.project-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 8px 25px rgba(0, 0, 0, 0.2);
}

.project-card h4 {
  margin: 0 0 var(--spacing-sm) 0;
  font-size: 1.1rem;
}

.project-card p {
  margin: 0 0 var(--spacing-md) 0;
  opacity: 0.8;
  font-size: 0.9rem;
}

/* 功能特性 */
.features {
  padding: 80px 0;
  background: white;
}

.container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 var(--spacing-lg);
}

.section-title {
  text-align: center;
  font-size: 2.5rem;
  font-weight: bold;
  margin-bottom: var(--spacing-xxl);
  color: var(--primary-color);
}

.features-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: var(--spacing-xl);
}

.feature-card {
  text-align: center;
  padding: var(--spacing-xl);
  border-radius: var(--border-radius-lg);
  background: white;
  box-shadow: var(--shadow-md);
  transition: all 0.3s ease;
}

.feature-card:hover {
  transform: translateY(-8px);
  box-shadow: var(--shadow-lg);
}

.feature-icon {
  margin-bottom: var(--spacing-lg);
  color: var(--accent-color);
  display: flex;
  align-items: center;
  justify-content: center;
}

.feature-icon .n-icon {
  transform: none !important;
  font-style: normal !important;
}

.feature-icon .n-icon i {
  transform: none !important;
  font-style: normal !important;
}

.feature-card h3 {
  margin-bottom: var(--spacing-md);
  color: var(--primary-color);
  font-size: 1.25rem;
}

.feature-card p {
  color: var(--text-secondary);
  line-height: 1.6;
}

/* 使用流程 */
.process {
  padding: 80px 0;
  background: var(--background-color);
}

.process-timeline {
  display: flex;
  flex-direction: row;
  gap: var(--spacing-lg);
  max-width: 1200px;
  margin: 0 auto;
  justify-content: center;
  flex-wrap: wrap;
}

.process-step {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  gap: var(--spacing-md);
  padding: var(--spacing-lg);
  background: white;
  border-radius: var(--border-radius-lg);
  box-shadow: var(--shadow-sm);
  transition: all 0.3s ease;
  flex: 1;
  min-width: 200px;
  max-width: 280px;
}

.step-number {
  width: 60px;
  height: 60px;
  border-radius: 50%;
  background: var(--accent-color);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.5rem;
  font-weight: bold;
  flex-shrink: 0;
}

.step-content h3 {
  margin: 0 0 var(--spacing-sm) 0;
  color: var(--primary-color);
  font-size: 1.25rem;
}

.step-content p {
  margin: 0;
  color: var(--text-secondary);
  line-height: 1.6;
}

/* 底部 */
.footer {
  background: var(--primary-color);
  color: white;
  padding: 60px 0 20px;
}

.footer-content {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: var(--spacing-xl);
  margin-bottom: var(--spacing-xl);
}

.footer-section h3,
.footer-section h4 {
  margin-bottom: var(--spacing-md);
}

.footer-section p {
  margin-bottom: var(--spacing-sm);
  opacity: 0.8;
}

.social-links {
  display: flex;
  gap: var(--spacing-md);
}

.social-link {
  color: white;
  text-decoration: none;
  opacity: 0.8;
  transition: opacity 0.3s ease;
}

.social-link:hover {
  opacity: 1;
}

.footer-bottom {
  border-top: 1px solid rgba(255, 255, 255, 0.2);
  padding-top: var(--spacing-lg);
  text-align: center;
  opacity: 0.8;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .hero-title {
    font-size: 2rem;
  }
  
  .hero-subtitle {
    font-size: 1rem;
  }
  
  .section-title {
    font-size: 2rem;
  }
  
  .features-grid {
    grid-template-columns: 1fr;
  }
  
  .process-step {
    flex-direction: column;
    text-align: center;
    min-width: 150px;
    max-width: 200px;
  }
  
  .header-container {
    padding: 0 var(--spacing-md);
  }
  
  .nav {
    display: none;
  }
}

@media (max-width: 480px) {
  .hero {
    padding: 100px 0 60px;
  }
  
  .hero-input {
    margin-bottom: var(--spacing-xl);
  }
  
  .project-cards {
    grid-template-columns: 1fr;
  }
}
</style>