<template>
  <PageLayout>
    <div class="project-edit-page">
      <div class="page-header">
        <div class="header-left">
          <n-button text @click="goBack">
            <template #icon>
              <n-icon><ArrowLeftIcon /></n-icon>
            </template>
            返回
          </n-button>
          <div class="project-info">
            <h1>{{ project?.name || '项目编辑' }}</h1>
            <p v-if="project?.description">{{ project.description }}</p>
          </div>
        </div>
        <div class="header-right">
          <n-button-group>
            <n-button
              :type="viewMode === 'conversation' ? 'primary' : 'default'"
              @click="viewMode = 'conversation'"
            >
              <template #icon>
                <n-icon><ChatIcon /></n-icon>
              </template>
              对话
            </n-button>
            <n-button
              :type="viewMode === 'panel' ? 'primary' : 'default'"
              @click="viewMode = 'panel'"
            >
              <template #icon>
                <n-icon><CodeIcon /></n-icon>
              </template>
              面板
            </n-button>
          </n-button-group>
        </div>
      </div>

      <div class="page-content">
        <!-- 对话模式 -->
        <div v-if="viewMode === 'conversation'" class="conversation-mode">
          <div class="conversation-layout">
            <div class="conversation-left">
              <ConversationContainer
                :project-id="projectId"
                :requirements="project?.requirements || ''"
              />
            </div>
            <div class="conversation-right">
              <ProjectPanel :project="project" />
            </div>
          </div>
        </div>

        <!-- 面板模式 -->
        <div v-else-if="viewMode === 'panel'" class="panel-mode">
          <ProjectPanel :project="project" />
        </div>
      </div>
    </div>
  </PageLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, h } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { NButton, NButtonGroup, NIcon } from 'naive-ui'
import PageLayout from '@/components/layout/PageLayout.vue'
import ConversationContainer from '@/components/ConversationContainer.vue'
import ProjectPanel from '@/components/ProjectPanel.vue'
import { useProjectStore } from '@/stores/project'
import type { Project } from '@/types/project'

const route = useRoute()
const router = useRouter()
const projectStore = useProjectStore()

// 响应式数据
const viewMode = ref<'conversation' | 'panel'>('conversation')
const project = ref<Project | null>(null)
const loading = ref(false)

// 计算属性
const projectId = computed(() => route.params.id as string)

// 图标组件
const ArrowLeftIcon = () => h('svg', { 
  viewBox: '0 0 24 24', 
  fill: 'currentColor',
  style: 'width: 1em; height: 1em;'
}, [
  h('path', { d: 'M20 11H7.83l5.59-5.59L12 4l-8 8 8 8 1.41-1.41L7.83 13H20v-2z' })
])

const ChatIcon = () => h('svg', { 
  viewBox: '0 0 24 24', 
  fill: 'currentColor',
  style: 'width: 1em; height: 1em;'
}, [
  h('path', { d: 'M20 2H4c-1.1 0-2 .9-2 2v12c0 1.1.9 2 2 2h4l4 4 4-4h4c1.1 0 2-.9 2-2V4c0-1.1-.9-2-2-2zm-2 12H6v-2h12v2zm0-3H6V9h12v2zm0-3H6V6h12v2z' })
])

const CodeIcon = () => h('svg', { 
  viewBox: '0 0 24 24', 
  fill: 'currentColor',
  style: 'width: 1em; height: 1em;'
}, [
  h('path', { d: 'M9.4 16.6L4.8 12l4.6-4.6L8 6l-6 6 6 6 1.4-1.4zm5.2 0L19.2 12l-4.6-4.6L16 6l6 6-6 6-1.4-1.4z' })
])

// 方法
const goBack = () => {
  router.back()
}

const loadProject = async () => {
  if (!projectId.value) return
  
  loading.value = true
  try {
    const projectData = await projectStore.getProject(projectId.value)
    if (projectData) {
      project.value = projectData
    } else {
      // 项目不存在，跳转到仪表板
      router.push('/dashboard')
    }
  } catch (error) {
    console.error('加载项目失败:', error)
    // 跳转到仪表板
    router.push('/dashboard')
  } finally {
    loading.value = false
  }
}

// 生命周期
onMounted(() => {
  loadProject()
})
</script>

<style scoped>
.project-edit-page {
  height: 100vh;
  display: flex;
  flex-direction: column;
  background: var(--background-color);
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--spacing-lg);
  background: white;
  border-bottom: 1px solid var(--border-color);
}

.header-left {
  display: flex;
  align-items: center;
  gap: var(--spacing-lg);
}

.project-info h1 {
  margin: 0 0 var(--spacing-xs) 0;
  color: var(--primary-color);
  font-size: 1.5rem;
}

.project-info p {
  margin: 0;
  color: var(--text-secondary);
  font-size: 0.9rem;
}

.page-content {
  flex: 1;
  overflow: hidden;
}

/* 对话模式布局 */
.conversation-mode {
  height: 100%;
}

.conversation-layout {
  display: flex;
  height: 100%;
}

.conversation-left {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.conversation-right {
  width: 400px;
  border-left: 1px solid var(--border-color);
  background: white;
}

/* 面板模式布局 */
.panel-mode {
  height: 100%;
  padding: var(--spacing-lg);
}

/* 响应式设计 */
@media (max-width: 1024px) {
  .conversation-layout {
    flex-direction: column;
  }
  
  .conversation-right {
    width: 100%;
    height: 300px;
    border-left: none;
    border-top: 1px solid var(--border-color);
  }
}

@media (max-width: 768px) {
  .page-header {
    flex-direction: column;
    align-items: flex-start;
    gap: var(--spacing-md);
  }
  
  .header-left {
    flex-direction: column;
    align-items: flex-start;
    gap: var(--spacing-md);
  }
  
  .conversation-right {
    height: 250px;
  }
}
</style>