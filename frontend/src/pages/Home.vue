<template>
  <div class="home-page">
    <!-- 顶部导航栏 -->
    <header class="header" :class="{ 'header-scrolled': isScrolled }">
      <div class="header-container">
        <div class="logo">
          <h1>AutoCodeWeb</h1>
        </div>
        <nav class="nav">
          <a href="#features" class="nav-link">功能特性</a>
          <a href="#process" class="nav-link">使用流程</a>
          <a href="#about" class="nav-link">关于我们</a>
        </nav>
        <div class="header-actions">
          <n-button
            size="small"
            @click="toggleLanguage"
            class="language-btn"
          >
            {{ currentLanguage === 'zh' ? 'EN' : '中文' }}
          </n-button>
          <n-button
            type="primary"
            @click="handleExperienceClick"
            class="experience-btn"
          >
            {{ isLoggedIn ? '进入控制台' : '立即体验' }}
          </n-button>
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
            <n-input-group>
              <n-input
                v-model:value="projectDescription"
                :placeholder="t('hero.inputPlaceholder')"
                size="large"
                @keydown.enter="handleProjectCreate"
              />
              <n-button
                type="primary"
                size="large"
                :disabled="!projectDescription.trim()"
                @click="handleProjectCreate"
              >
                {{ t('hero.createButton') }}
              </n-button>
            </n-input-group>
          </div>

          <!-- 用户项目展示 -->
          <div v-if="isLoggedIn && userProjects.length > 0" class="user-projects">
            <h3>{{ t('hero.recentProjects') }}</h3>
            <div class="project-cards">
              <div
                v-for="project in userProjects.slice(0, 5)"
                :key="project.id"
                class="project-card"
                @click="goToProject(project.id)"
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

    <!-- 功能特性展示 -->
    <section id="features" class="features">
      <div class="container">
        <h2 class="section-title">{{ t('features.title') }}</h2>
        <div class="features-grid">
          <div class="feature-card" v-for="feature in features" :key="feature.id">
            <div class="feature-icon">
              <n-icon size="48">
                <component :is="feature.icon" />
              </n-icon>
            </div>
            <h3>{{ feature.title }}</h3>
            <p>{{ feature.description }}</p>
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
            :class="{ 'active': currentStep === index }"
          >
            <div class="step-number">{{ index + 1 }}</div>
            <div class="step-content">
              <h3>{{ step.title }}</h3>
              <p>{{ step.description }}</p>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- 底部信息 -->
    <footer id="about" class="footer">
      <div class="container">
        <div class="footer-content">
          <div class="footer-section">
            <h3>AutoCodeWeb</h3>
            <p>{{ t('footer.description') }}</p>
          </div>
          <div class="footer-section">
            <h4>{{ t('footer.contact') }}</h4>
            <p>邮箱: contact@autocodeweb.com</p>
            <p>电话: +86 400-123-4567</p>
          </div>
          <div class="footer-section">
            <h4>{{ t('footer.follow') }}</h4>
            <div class="social-links">
              <a href="#" class="social-link">GitHub</a>
              <a href="#" class="social-link">Twitter</a>
              <a href="#" class="social-link">LinkedIn</a>
            </div>
          </div>
        </div>
        <div class="footer-bottom">
          <p>&copy; 2025 AutoCodeWeb. {{ t('footer.rights') }}</p>
        </div>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { useProjectStore } from '@/stores/project'
import { NButton, NInputGroup, NInput, NIcon, NTag } from 'naive-ui'
import type { Project } from '@/types/project'

// 图标组件（临时使用 emoji，后续可替换为真实图标）
const CodeIcon = () => '💻'
const RobotIcon = () => '🤖'
const RocketIcon = () => '🚀'
const ShieldIcon = () => '🛡️'
const UsersIcon = () => '👥'
const ZapIcon = () => '⚡'

const router = useRouter()
const userStore = useUserStore()
const projectStore = useProjectStore()

// 响应式数据
const isScrolled = ref(false)
const currentLanguage = ref('zh')
const projectDescription = ref('')
const currentStep = ref(0)

// 计算属性
const isLoggedIn = computed(() => userStore.isLoggedIn)
const userProjects = computed(() => projectStore.projects.slice(0, 5))

// 功能特性数据
const features = ref([
  {
    id: 1,
    icon: CodeIcon,
    title: '智能代码生成',
    description: '基于自然语言描述，自动生成高质量的代码'
  },
  {
    id: 2,
    icon: RobotIcon,
    title: '多Agent协作',
    description: '产品经理、架构师、开发工程师等多角色协作'
  },
  {
    id: 3,
    icon: RocketIcon,
    title: '快速部署',
    description: '一键部署到云端，支持多种部署方式'
  },
  {
    id: 4,
    icon: ShieldIcon,
    title: '安全可靠',
    description: '企业级安全保障，数据加密传输'
  },
  {
    id: 5,
    icon: UsersIcon,
    title: '团队协作',
    description: '支持团队协作，权限管理完善'
  },
  {
    id: 6,
    icon: ZapIcon,
    title: '高效开发',
    description: '提升开发效率，减少重复工作'
  }
])

// 使用流程数据
const processSteps = ref([
  {
    id: 1,
    title: '描述需求',
    description: '用自然语言描述你的项目需求'
  },
  {
    id: 2,
    title: 'Agent分析',
    description: '多Agent协作分析需求并制定方案'
  },
  {
    id: 3,
    title: '生成代码',
    description: '自动生成高质量的代码和文档'
  },
  {
    id: 4,
    title: '测试部署',
    description: '自动测试并部署到目标环境'
  }
])

// 国际化文本
const t = (key: string) => {
  const texts = {
    zh: {
      'hero.title': '多Agent自动实现APP和网站项目',
      'hero.subtitle': '用自然语言描述需求，AI Agent 自动生成完整项目',
      'hero.inputPlaceholder': '描述你的项目需求，例如：创建一个电商网站...',
      'hero.createButton': '开始创建',
      'hero.recentProjects': '最近项目',
      'features.title': '功能特性',
      'process.title': '使用流程',
      'footer.description': '让编程变得更简单，让创意更快实现',
      'footer.contact': '联系我们',
      'footer.follow': '关注我们',
      'footer.rights': '保留所有权利'
    },
    en: {
      'hero.title': 'Multi-Agent Auto Implementation for APP and Web Projects',
      'hero.subtitle': 'Describe requirements in natural language, AI Agents auto-generate complete projects',
      'hero.inputPlaceholder': 'Describe your project requirements, e.g.: Create an e-commerce website...',
      'hero.createButton': 'Start Creating',
      'hero.recentProjects': 'Recent Projects',
      'features.title': 'Features',
      'process.title': 'How It Works',
      'footer.description': 'Making programming simpler, making ideas come true faster',
      'footer.contact': 'Contact Us',
      'footer.follow': 'Follow Us',
      'footer.rights': 'All rights reserved'
    }
  }
  return texts[currentLanguage.value as keyof typeof texts]?.[key as keyof typeof texts.zh] || key
}

// 方法
const toggleLanguage = () => {
  currentLanguage.value = currentLanguage.value === 'zh' ? 'en' : 'zh'
}

const handleExperienceClick = () => {
  if (isLoggedIn.value) {
    router.push('/dashboard')
  } else {
    router.push('/auth/login')
  }
}

const handleProjectCreate = async () => {
  if (!projectDescription.value.trim()) return
  
  if (!isLoggedIn.value) {
    // 未登录用户跳转到登录页面，并保存输入内容
    localStorage.setItem('pendingProjectDescription', projectDescription.value)
    router.push('/auth/login')
    return
  }
  
  // 已登录用户直接跳转到创建项目页面
  router.push({
    path: '/create-project',
    query: { description: projectDescription.value }
  })
}

const goToProject = (projectId: string) => {
  router.push(`/project/${projectId}`)
}

const getStatusType = (status: string) => {
  const statusMap: Record<string, string> = {
    draft: 'default',
    in_progress: 'warning',
    completed: 'success',
    failed: 'error'
  }
  return statusMap[status] || 'default'
}

const getStatusText = (status: string) => {
  const statusMap: Record<string, string> = {
    draft: '草稿',
    in_progress: '进行中',
    completed: '已完成',
    failed: '失败'
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
}

.experience-btn {
  background: var(--accent-color);
  border: none;
  color: white;
  font-weight: 600;
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
  flex-direction: column;
  gap: var(--spacing-xl);
  max-width: 800px;
  margin: 0 auto;
}

.process-step {
  display: flex;
  align-items: center;
  gap: var(--spacing-lg);
  padding: var(--spacing-lg);
  background: white;
  border-radius: var(--border-radius-lg);
  box-shadow: var(--shadow-sm);
  transition: all 0.3s ease;
}

.process-step.active {
  border-left: 4px solid var(--accent-color);
  box-shadow: var(--shadow-md);
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