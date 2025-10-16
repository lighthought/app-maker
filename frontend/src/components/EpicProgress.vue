<template>
  <div class="epic-progress">
    <n-card title="MVP Stories 开发进度" :bordered="false">
      <n-spin :show="loading">
        <div v-if="epics.length === 0" class="empty-state">
          <n-empty description="暂无 Epic 数据" />
        </div>
        <div v-else class="epics-container">
          <n-collapse v-model:expanded-names="expandedNames" accordion>
            <n-collapse-item
              v-for="epic in epics"
              :key="epic.id"
              :name="epic.id"
              :title="`Epic ${epic.epic_number}: ${epic.name}`"
            >
              <template #header-extra>
                <n-tag
                  :type="getEpicStatusType(epic)"
                  size="small"
                  :bordered="false"
                >
                  {{ getEpicStatusText(epic) }}
                </n-tag>
              </template>
              
              <div class="epic-details">
                <div class="epic-info">
                  <n-space vertical>
                    <div class="info-item">
                      <span class="label">描述:</span>
                      <span>{{ epic.description || '无' }}</span>
                    </div>
                    <div class="info-item">
                      <span class="label">优先级:</span>
                      <n-tag :type="getPriorityType(epic.priority)" size="small">
                        {{ epic.priority }}
                      </n-tag>
                    </div>
                    <div class="info-item">
                      <span class="label">预估天数:</span>
                      <span>{{ epic.estimated_days }} 天</span>
                    </div>
                  </n-space>
                </div>
                
                <n-divider />
                
                <div class="stories-section">
                  <h4>Stories ({{ epic.stories?.length || 0 }})</h4>
                  <n-space vertical>
                    <div
                      v-for="story in epic.stories"
                      :key="story.id"
                      class="story-item"
                      :class="{ completed: story.status === 'done' }"
                    >
                      <n-checkbox
                        :checked="story.status === 'done'"
                        :disabled="true"
                      >
                        <span class="story-number">{{ story.story_number }}</span>
                        <span class="story-title">{{ story.title }}</span>
                      </n-checkbox>
                      
                      <n-tag
                        :type="getStoryStatusType(story.status)"
                        size="tiny"
                        class="story-status"
                      >
                        {{ getStoryStatusText(story.status) }}
                      </n-tag>
                    </div>
                  </n-space>
                </div>
              </div>
            </n-collapse-item>
          </n-collapse>
        </div>
      </n-spin>
    </n-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { NCard, NSpin, NEmpty, NCollapse, NCollapseItem, NTag, NSpace, NDivider, NCheckbox, useMessage } from 'naive-ui'
import { http } from '@/utils/http'

interface Story {
  id: string
  story_number: string
  title: string
  description: string
  priority: string
  status: string
  estimated_days: number
}

interface Epic {
  id: string
  epic_number: number
  name: string
  description: string
  priority: string
  status: string
  estimated_days: number
  stories: Story[]
}

interface Props {
  projectGuid: string
}

const props = defineProps<Props>()
const message = useMessage()

const loading = ref(false)
const epics = ref<Epic[]>([])
const expandedNames = ref<string[]>([])

// 获取 MVP Epics
const fetchMvpEpics = async () => {
  loading.value = true
  try {
    const response = await http.get(`/projects/${props.projectGuid}/mvp-epics`)
    if (response.data.code === 200) {
      epics.value = response.data.data || []
      // 默认展开第一个 epic
      if (epics.value.length > 0) {
        expandedNames.value = [epics.value[0].id]
      }
    } else {
      message.warning(response.data.message || '获取 MVP Epics 失败')
    }
  } catch (error: any) {
    console.error('获取 MVP Epics 失败:', error)
    message.error('获取 MVP Epics 失败: ' + (error.message || '未知错误'))
  } finally {
    loading.value = false
  }
}

// Epic 状态类型
const getEpicStatusType = (epic: Epic) => {
  const doneCount = epic.stories?.filter(s => s.status === 'done').length || 0
  const totalCount = epic.stories?.length || 0
  
  if (doneCount === totalCount && totalCount > 0) return 'success'
  if (doneCount > 0) return 'info'
  return 'default'
}

// Epic 状态文本
const getEpicStatusText = (epic: Epic) => {
  const doneCount = epic.stories?.filter(s => s.status === 'done').length || 0
  const totalCount = epic.stories?.length || 0
  
  return `${doneCount}/${totalCount} 已完成`
}

// 优先级类型
const getPriorityType = (priority: string) => {
  switch (priority) {
    case 'P0':
      return 'error'
    case 'P1':
      return 'warning'
    case 'P2':
      return 'info'
    default:
      return 'default'
  }
}

// Story 状态类型
const getStoryStatusType = (status: string) => {
  switch (status) {
    case 'done':
      return 'success'
    case 'in_progress':
      return 'info'
    case 'failed':
      return 'error'
    default:
      return 'default'
  }
}

// Story 状态文本
const getStoryStatusText = (status: string) => {
  switch (status) {
    case 'done':
      return '已完成'
    case 'in_progress':
      return '进行中'
    case 'failed':
      return '失败'
    case 'pending':
      return '待开始'
    default:
      return status
  }
}

onMounted(() => {
  fetchMvpEpics()
})
</script>

<style scoped lang="scss">
.epic-progress {
  width: 100%;
  
  .empty-state {
    padding: 40px 0;
    text-align: center;
  }
  
  .epics-container {
    width: 100%;
  }
  
  .epic-details {
    padding: 16px 0;
    
    .epic-info {
      margin-bottom: 16px;
      
      .info-item {
        display: flex;
        align-items: center;
        gap: 8px;
        
        .label {
          font-weight: 500;
          color: var(--n-text-color-2);
          min-width: 80px;
        }
      }
    }
    
    .stories-section {
      h4 {
        margin: 0 0 12px 0;
        font-size: 14px;
        font-weight: 500;
      }
      
      .story-item {
        display: flex;
        align-items: center;
        justify-content: space-between;
        padding: 8px 12px;
        background: var(--n-color);
        border-radius: 6px;
        transition: all 0.3s ease;
        
        &.completed {
          opacity: 0.7;
          
          .story-title {
            text-decoration: line-through;
          }
        }
        
        &:hover {
          background: var(--n-color-hover);
        }
        
        .story-number {
          font-weight: 500;
          color: var(--n-primary-color);
          margin-right: 8px;
        }
        
        .story-title {
          color: var(--n-text-color);
        }
        
        .story-status {
          margin-left: auto;
        }
      }
    }
  }
}
</style>

