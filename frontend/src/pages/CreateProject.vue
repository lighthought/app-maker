<template>
  <PageLayout>
    <div class="create-project-page">
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
                :disabled="isCreating"
                @send="handleProjectCreate"
              />
            </div>
            
            <!-- 创建状态提示 -->
            <div v-if="isCreating" class="creating-status">
              <n-icon size="24" color="white">
                <LoadingIcon />
              </n-icon>
              <span>{{ t('project.creating') }}</span>
            </div>
          </div>
        </div>
      </section>
    </div>
  </PageLayout>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useMessage, NIcon } from 'naive-ui'
import PageLayout from '@/components/layout/PageLayout.vue'
import SmartInput from '@/components/common/SmartInput.vue'
// 导入图标
import { LoadingIcon } from '@/components/icon'
import { useProjectStore } from '@/stores/project'

const router = useRouter()
const route = useRoute()
const message = useMessage()
const projectStore = useProjectStore()
const { t } = useI18n()

// 响应式数据
const projectDescription = ref('')
const isCreating = ref(false)

// 方法
const handleProjectCreate = async (content?: string) => {
  const description = content || projectDescription.value
  if (!description.trim()) {
    message.warning(t('project.descriptionRequired'))
    return
  }
  
  try {
    isCreating.value = true
    message.loading(t('project.creating'), { duration: 0 })
    
    // 不再生成项目名称，让后端自动生成
    const projectData = {
      requirements: description.trim(),
    }
    
    const createdProject = await projectStore.createProject(projectData)
    
    if (createdProject) {
      message.destroyAll()
      message.success(t('project.projectCreated'))
      
      // 自动跳转到项目详情页面
      router.push(`/project/${createdProject.guid}`)
    } else {
      throw new Error(t('project.projectCreateError'))
    }
  } catch (error) {
    message.destroyAll()
    message.error(`${t('project.projectCreateError')}: ${error instanceof Error ? error.message : t('common.unknownError')}`)
    console.error('创建项目失败:', error)
  } finally {
    isCreating.value = false
  }
}

// 生命周期
onMounted(() => {
  // 检查是否有待创建的项目描述（从首页跳转过来）
  const pendingDescription = localStorage.getItem('pendingProjectDescription')
  if (pendingDescription) {
    projectDescription.value = pendingDescription
    localStorage.removeItem('pendingProjectDescription')
    
    // 自动触发创建
    setTimeout(() => {
      handleProjectCreate()
    }, 500) // 延迟500ms确保页面完全加载
  }
  
  // 检查URL查询参数
  const description = route.query.description as string
  if (description) {
    projectDescription.value = description
  }
})
</script>

<style scoped>
.create-project-page {
  min-height: 100vh;
  background: var(--background-color);
}

/* Hero 区域 */
.hero {
  padding: 120px 0 80px;
  text-align: center;
  /*color: white;*/
  /*background: linear-gradient(135deg, var(--primary-color) 0%, var(--accent-color) 100%);*/
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

/* 创建状态提示 */
.creating-status {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--spacing-md);
  margin-top: var(--spacing-lg);
  padding: var(--spacing-md);
  background: rgba(255, 255, 255, 0.1);
  border-radius: var(--border-radius-lg);
  backdrop-filter: blur(10px);
}

.creating-status span {
  font-size: 1rem;
  font-weight: 500;
}

/* 项目内容区域 */
.project-content {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 var(--spacing-lg) var(--spacing-xxl);
  text-align: center;
  color: var(--text-primary);
  background: white;
  border-radius: var(--border-radius-lg);
  margin-top: var(--spacing-xl);
  box-shadow: var(--shadow-sm);
}

.project-content h2 {
  margin: 0;
  padding: var(--spacing-lg) 0 var(--spacing-md) 0;
  color: var(--primary-color);
  font-size: 2rem;
  font-weight: bold;
}

.project-content p {
  margin: 0;
  padding-bottom: var(--spacing-lg);
  color: var(--text-secondary);
  font-size: 1.1rem;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .hero-title {
    font-size: 2rem;
  }
  
  .hero-subtitle {
    font-size: 1rem;
  }
  
  .hero {
    padding: 100px 0 60px;
  }
}
</style>