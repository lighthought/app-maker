<template>
  <div class="epic-story-editor">
    <n-card title="Epic 和 Story 编辑" :bordered="false" class="editor-card">
      <template #header-extra>
        <n-space>
          <n-button @click="handleRegenerate" :loading="regenerating" type="warning" size="small">
            重新生成
          </n-button>
          <n-button @click="handleSkip" type="info" size="small">
            跳过确认
          </n-button>
          <n-button @click="handleConfirm" :loading="confirming" type="primary" size="small">
            确认并继续
          </n-button>
        </n-space>
      </template>

      <n-spin :show="loading">
        <div v-if="epics.length === 0" class="empty-state">
          <n-empty description="暂无 Epic 数据" />
        </div>

        <div v-else class="epics-container">
          <!-- 移动端优化：使用折叠面板 -->
          <n-collapse v-model:expanded-names="expandedNames" accordion class="mobile-collapse">
            <n-collapse-item
              v-for="epic in sortedEpics"
              :key="epic.id"
              :name="epic.id"
              :title="`Epic ${epic.epic_number}: ${epic.name}`"
            >
              <template #header-extra>
                <n-space>
                  <n-tag :type="getEpicStatusType(epic)" size="small">
                    {{ getEpicStatusText(epic) }}
                  </n-tag>
                  <n-button
                    @click.stop="handleDeleteEpic(epic)"
                    type="error"
                    size="tiny"
                    text
                  >
                    删除
                  </n-button>
                </n-space>
              </template>

              <div class="epic-content">
                <!-- Epic 编辑区域 -->
                <n-card size="small" class="epic-edit-card">
                  <n-form :model="epic" label-placement="left" label-width="80px">
                    <n-grid :cols="isMobile ? 1 : 2" :x-gap="12">
                      <n-form-item-gi label="名称">
                        <n-input v-model:value="epic.name" placeholder="Epic 名称" />
                      </n-form-item-gi>
                      <n-form-item-gi label="优先级">
                        <n-select
                          v-model:value="epic.priority"
                          :options="priorityOptions"
                          placeholder="选择优先级"
                        />
                      </n-form-item-gi>
                      <n-form-item-gi label="预估天数">
                        <n-input-number
                          v-model:value="epic.estimated_days"
                          :min="0"
                          placeholder="预估天数"
                        />
                      </n-form-item-gi>
                      <n-form-item-gi label="状态">
                        <n-tag :type="getEpicStatusType(epic)" size="small">
                          {{ getEpicStatusText(epic) }}
                        </n-tag>
                      </n-form-item-gi>
                    </n-grid>
                    <n-form-item label="描述">
                      <n-input
                        v-model:value="epic.description"
                        type="textarea"
                        :rows="2"
                        placeholder="Epic 描述"
                      />
                    </n-form-item>
                    <n-form-item>
                      <n-space>
                        <n-button @click="handleUpdateEpic(epic)" type="primary" size="small">
                          保存 Epic
                        </n-button>
                        <n-button @click="handleDeleteEpic(epic)" type="error" size="small">
                          删除 Epic
                        </n-button>
                      </n-space>
                    </n-form-item>
                  </n-form>
                </n-card>

                <n-divider />

                <!-- Stories 编辑区域 -->
                <div class="stories-section">
                  <div class="stories-header">
                    <h4>Stories ({{ epic.stories?.length || 0 }})</h4>
                    <n-space>
                      <n-button
                        @click="toggleStorySelection(epic)"
                        size="small"
                        :type="isAllStoriesSelected(epic) ? 'primary' : 'default'"
                      >
                        {{ isAllStoriesSelected(epic) ? '取消全选' : '全选' }}
                      </n-button>
                      <n-button
                        @click="handleBatchDeleteStories(epic)"
                        :disabled="!hasSelectedStories(epic)"
                        type="error"
                        size="small"
                      >
                        批量删除 ({{ getSelectedStoriesCount(epic) }})
                      </n-button>
                    </n-space>
                  </div>

                  <!-- 拖拽排序的 Stories 列表 -->
                  <draggable
                    v-model="epic.stories"
                    :animation="200"
                    handle=".drag-handle"
                    class="stories-list"
                    @end="handleStoryOrderChange(epic)"
                  >
                    <div
                      v-for="story in epic.stories"
                      :key="story.id"
                      class="story-item"
                      :class="{
                        'selected': selectedStories.has(story.id),
                        'completed': story.status === 'done'
                      }"
                    >
                      <!-- 拖拽手柄 -->
                      <div class="drag-handle">
                        <n-icon size="16">
                          <svg viewBox="0 0 24 24">
                            <path d="M3 18h18v-2H3v2zm0-5h18v-2H3v2zm0-7v2h18V6H3z"/>
                          </svg>
                        </n-icon>
                      </div>

                      <!-- 选择框 -->
                      <n-checkbox
                        :checked="selectedStories.has(story.id)"
                        @update:checked="toggleStorySelection(story.id)"
                        class="story-checkbox"
                      />

                      <!-- Story 内容 -->
                      <div class="story-content">
                        <div class="story-header">
                          <span class="story-number">{{ story.story_number }}</span>
                          <n-input
                            v-model:value="story.title"
                            placeholder="Story 标题"
                            class="story-title-input"
                            @blur="handleUpdateStory(story)"
                          />
                          <n-tag :type="getStoryStatusType(story.status)" size="tiny">
                            {{ getStoryStatusText(story.status) }}
                          </n-tag>
                        </div>

                        <!-- Story 详细信息（可折叠） -->
                        <n-collapse>
                          <n-collapse-item title="详细信息" name="details">
                            <n-form :model="story" label-placement="left" label-width="80px">
                              <n-grid :cols="isMobile ? 1 : 2" :x-gap="12">
                                <n-form-item-gi label="优先级">
                                  <n-select
                                    v-model:value="story.priority"
                                    :options="priorityOptions"
                                    placeholder="选择优先级"
                                    @update:value="handleUpdateStory(story)"
                                  />
                                </n-form-item-gi>
                                <n-form-item-gi label="预估天数">
                                  <n-input-number
                                    v-model:value="story.estimated_days"
                                    :min="0"
                                    placeholder="预估天数"
                                    @update:value="handleUpdateStory(story)"
                                  />
                                </n-form-item-gi>
                                <n-form-item-gi label="依赖">
                                  <n-input
                                    v-model:value="story.depends"
                                    placeholder="依赖的其他 Story"
                                    @blur="handleUpdateStory(story)"
                                  />
                                </n-form-item-gi>
                                <n-form-item-gi label="技术要点">
                                  <n-input
                                    v-model:value="story.techs"
                                    placeholder="技术要点"
                                    @blur="handleUpdateStory(story)"
                                  />
                                </n-form-item-gi>
                              </n-grid>
                              <n-form-item label="描述">
                                <n-input
                                  v-model:value="story.description"
                                  type="textarea"
                                  :rows="2"
                                  placeholder="Story 描述"
                                  @blur="handleUpdateStory(story)"
                                />
                              </n-form-item>
                              <n-form-item label="验收标准">
                                <n-input
                                  v-model:value="story.acceptance_criteria"
                                  type="textarea"
                                  :rows="3"
                                  placeholder="验收标准"
                                  @blur="handleUpdateStory(story)"
                                />
                              </n-form-item>
                              <n-form-item label="详细内容">
                                <n-input
                                  v-model:value="story.content"
                                  type="textarea"
                                  :rows="3"
                                  placeholder="详细内容"
                                  @blur="handleUpdateStory(story)"
                                />
                              </n-form-item>
                            </n-form>
                          </n-collapse-item>
                        </n-collapse>
                      </div>

                      <!-- 删除按钮 -->
                      <n-button
                        @click="handleDeleteStory(story)"
                        type="error"
                        size="tiny"
                        text
                        class="delete-button"
                      >
                        删除
                      </n-button>
                    </div>
                  </draggable>
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
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { NCard, NSpin, NEmpty, NCollapse, NCollapseItem, NTag, NSpace, NDivider, NCheckbox, NButton, NForm, NFormItem, NFormItemGi, NGrid, NInput, NInputNumber, NSelect, NIcon, useMessage, useDialog } from 'naive-ui'
import draggable from 'vuedraggable'
import { useProjectStore } from '@/stores/project'

interface Story {
  id: string
  epic_id: string
  story_number: string
  title: string
  description: string
  priority: string
  status: string
  estimated_days: number
  depends: string
  techs: string
  content: string
  acceptance_criteria: string
  display_order: number
}

interface Epic {
  id: string
  epic_number: number
  name: string
  description: string
  priority: string
  status: string
  estimated_days: number
  display_order: number
  stories: Story[]
}

interface Props {
  projectGuid: string
}

const props = defineProps<Props>()
const message = useMessage()
const dialog = useDialog()
const projectStore = useProjectStore()

// 响应式数据
const loading = ref(false)
const epics = ref<Epic[]>([])
const expandedNames = ref<string[]>([])
const selectedStories = ref<Set<string>>(new Set())
const confirming = ref(false)
const regenerating = ref(false)

// 移动端检测
const isMobile = ref(false)
const checkMobile = () => {
  isMobile.value = window.innerWidth <= 768
}

// 优先级选项
const priorityOptions = [
  { label: 'P0 - 最高优先级', value: 'P0' },
  { label: 'P1 - 高优先级', value: 'P1' },
  { label: 'P2 - 中优先级', value: 'P2' },
  { label: 'P3 - 低优先级', value: 'P3' }
]

// 计算属性
const sortedEpics = computed(() => {
  return [...epics.value].sort((a, b) => a.display_order - b.display_order)
})

// 获取 MVP Epics
const fetchMvpEpics = async () => {
  loading.value = true
  try {
    epics.value = await projectStore.getMvpEpics(props.projectGuid)
    // 默认展开第一个 epic
    if (epics.value.length > 0) {
      expandedNames.value = [epics.value[0].id]
    }
  } catch (error: any) {
    console.error('获取 MVP Epics 失败:', error)
    message.error('获取 MVP Epics 失败: ' + (error.message || '未知错误'))
  } finally {
    loading.value = false
  }
}

// Epic 状态相关方法
const getEpicStatusType = (epic: Epic) => {
  const doneCount = epic.stories?.filter(s => s.status === 'done').length || 0
  const totalCount = epic.stories?.length || 0

  if (doneCount === totalCount && totalCount > 0) return 'success'
  if (doneCount > 0) return 'info'
  return 'default'
}

const getEpicStatusText = (epic: Epic) => {
  const doneCount = epic.stories?.filter(s => s.status === 'done').length || 0
  const totalCount = epic.stories?.length || 0

  return `${doneCount}/${totalCount} 已完成`
}

// Story 状态相关方法
const getStoryStatusType = (status: string) => {
  switch (status) {
    case 'done': return 'success'
    case 'in_progress': return 'info'
    case 'failed': return 'error'
    default: return 'default'
  }
}

const getStoryStatusText = (status: string) => {
  switch (status) {
    case 'done': return '已完成'
    case 'in_progress': return '进行中'
    case 'failed': return '失败'
    case 'pending': return '待开始'
    default: return status
  }
}

// Story 选择相关方法
const toggleStorySelection = (storyIdOrEpic: string | Epic) => {
  if (typeof storyIdOrEpic === 'string') {
    // 切换单个 Story 选择
    if (selectedStories.value.has(storyIdOrEpic)) {
      selectedStories.value.delete(storyIdOrEpic)
    } else {
      selectedStories.value.add(storyIdOrEpic)
    }
  } else {
    // 切换整个 Epic 的 Stories 选择
    const epic = storyIdOrEpic
    const allSelected = isAllStoriesSelected(epic)

    if (allSelected) {
      // 取消全选
      epic.stories?.forEach(story => {
        selectedStories.value.delete(story.id)
      })
    } else {
      // 全选
      epic.stories?.forEach(story => {
        selectedStories.value.add(story.id)
      })
    }
  }
}

const isAllStoriesSelected = (epic: Epic) => {
  if (!epic.stories || epic.stories.length === 0) return false
  return epic.stories.every(story => selectedStories.value.has(story.id))
}

const hasSelectedStories = (epic: Epic) => {
  if (!epic.stories || epic.stories.length === 0) return false
  return epic.stories.some(story => selectedStories.value.has(story.id))
}

const getSelectedStoriesCount = (epic: Epic) => {
  if (!epic.stories || epic.stories.length === 0) return 0
  return epic.stories.filter(story => selectedStories.value.has(story.id)).length
}

// Epic 操作方法
const handleUpdateEpic = async (epic: Epic) => {
  try {
    const success = await projectStore.updateEpic(props.projectGuid, epic.id, {
      name: epic.name,
      description: epic.description,
      priority: epic.priority,
      estimated_days: epic.estimated_days
    })

    if (success) {
      message.success('Epic 更新成功')
    } else {
      message.error('Epic 更新失败')
    }
  } catch (error: any) {
    console.error('更新 Epic 失败:', error)
    message.error('更新 Epic 失败: ' + (error.message || '未知错误'))
  }
}

const handleDeleteEpic = async (epic: Epic) => {
  dialog.warning({
    title: '确认删除',
    content: `确定要删除 Epic "${epic.name}" 吗？这将同时删除该 Epic 下的所有 Stories。`,
    positiveText: '确定',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        const success = await projectStore.deleteEpic(props.projectGuid, epic.id)

        if (success) {
          message.success('Epic 删除成功')
          // 从列表中移除
          const index = epics.value.findIndex(e => e.id === epic.id)
          if (index > -1) {
            epics.value.splice(index, 1)
          }
        } else {
          message.error('Epic 删除失败')
        }
      } catch (error: any) {
        console.error('删除 Epic 失败:', error)
        message.error('删除 Epic 失败: ' + (error.message || '未知错误'))
      }
    }
  })
}

// Story 操作方法
const handleUpdateStory = async (story: Story) => {
  try {
    // 找到 Story 所属的 Epic
    const epic = epics.value.find(e => e.stories?.some(s => s.id === story.id))
    if (!epic) {
      message.error('未找到 Story 所属的 Epic')
      return
    }

    const success = await projectStore.updateStory(props.projectGuid, epic.id, story.id, {
      title: story.title,
      description: story.description,
      priority: story.priority,
      estimated_days: story.estimated_days,
      depends: story.depends,
      techs: story.techs,
      content: story.content,
      acceptance_criteria: story.acceptance_criteria
    })

    if (success) {
      message.success('Story 更新成功')
    } else {
      message.error('Story 更新失败')
    }
  } catch (error: any) {
    console.error('更新 Story 失败:', error)
    message.error('更新 Story 失败: ' + (error.message || '未知错误'))
  }
}

const handleDeleteStory = async (story: Story) => {
  dialog.warning({
    title: '确认删除',
    content: `确定要删除 Story "${story.title}" 吗？`,
    positiveText: '确定',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        // 找到 Story 所属的 Epic
        const epic = epics.value.find(e => e.stories?.some(s => s.id === story.id))
        if (!epic) {
          message.error('未找到 Story 所属的 Epic')
          return
        }

        const success = await projectStore.deleteStory(props.projectGuid, epic.id, story.id)

        if (success) {
          message.success('Story 删除成功')
          // 从列表中移除
          if (epic.stories) {
            const index = epic.stories.findIndex(s => s.id === story.id)
            if (index > -1) {
              epic.stories.splice(index, 1)
            }
          }
          // 从选择列表中移除
          selectedStories.value.delete(story.id)
        } else {
          message.error('Story 删除失败')
        }
      } catch (error: any) {
        console.error('删除 Story 失败:', error)
        message.error('删除 Story 失败: ' + (error.message || '未知错误'))
      }
    }
  })
}

const handleBatchDeleteStories = async (epic: Epic) => {
  const selectedStoryIds = epic.stories?.filter(story => selectedStories.value.has(story.id)).map(story => story.id) || []

  if (selectedStoryIds.length === 0) {
    message.warning('请先选择要删除的 Stories')
    return
  }

  dialog.warning({
    title: '确认批量删除',
    content: `确定要删除选中的 ${selectedStoryIds.length} 个 Stories 吗？`,
    positiveText: '确定',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        const success = await projectStore.batchDeleteStories(props.projectGuid, selectedStoryIds)

        if (success) {
          message.success('批量删除 Stories 成功')
          // 从列表中移除
          selectedStoryIds.forEach(storyId => {
            selectedStories.value.delete(storyId)
            if (epic.stories) {
              const index = epic.stories.findIndex(s => s.id === storyId)
              if (index > -1) {
                epic.stories.splice(index, 1)
              }
            }
          })
        } else {
          message.error('批量删除 Stories 失败')
        }
      } catch (error: any) {
        console.error('批量删除 Stories 失败:', error)
        message.error('批量删除 Stories 失败: ' + (error.message || '未知错误'))
      }
    }
  })
}

// 拖拽排序处理
const handleStoryOrderChange = async (epic: Epic) => {
  if (!epic.stories) return

  // 更新 display_order
  epic.stories.forEach((story, index) => {
    story.display_order = index
  })

  // 批量更新排序
  try {
    const updatePromises = epic.stories.map((story, index) =>
      projectStore.updateStoryOrder(props.projectGuid, epic.id, story.id, index)
    )

    await Promise.all(updatePromises)
    message.success('Story 排序更新成功')
  } catch (error: any) {
    console.error('更新 Story 排序失败:', error)
    message.error('更新 Story 排序失败: ' + (error.message || '未知错误'))
    // 重新获取数据以恢复正确的顺序
    await fetchMvpEpics()
  }
}

// 确认操作
const handleConfirm = async () => {
  confirming.value = true
  try {
    const success = await projectStore.confirmEpicsAndStories(props.projectGuid, 'confirm')

    if (success) {
      message.success('确认成功，项目将继续执行')
      // 触发父组件更新
      emit('confirmed')
    } else {
      message.error('确认失败')
    }
  } catch (error: any) {
    console.error('确认失败:', error)
    message.error('确认失败: ' + (error.message || '未知错误'))
  } finally {
    confirming.value = false
  }
}

const handleSkip = async () => {
  try {
    const success = await projectStore.confirmEpicsAndStories(props.projectGuid, 'skip')

    if (success) {
      message.success('跳过确认成功，项目将继续执行')
      emit('confirmed')
    } else {
      message.error('跳过确认失败')
    }
  } catch (error: any) {
    console.error('跳过确认失败:', error)
    message.error('跳过确认失败: ' + (error.message || '未知错误'))
  }
}

const handleRegenerate = async () => {
  regenerating.value = true
  try {
    const success = await projectStore.confirmEpicsAndStories(props.projectGuid, 'regenerate')

    if (success) {
      message.success('重新生成请求已发送')
      // 重新获取数据
      await fetchMvpEpics()
    } else {
      message.error('重新生成失败')
    }
  } catch (error: any) {
    console.error('重新生成失败:', error)
    message.error('重新生成失败: ' + (error.message || '未知错误'))
  } finally {
    regenerating.value = false
  }
}

// 事件定义
const emit = defineEmits<{
  confirmed: []
}>()

// 生命周期
onMounted(() => {
  checkMobile()
  fetchMvpEpics()

  // 监听窗口大小变化
  window.addEventListener('resize', checkMobile)
})

// 清理
onUnmounted(() => {
  window.removeEventListener('resize', checkMobile)
})
</script>

<style scoped lang="scss">
.epic-story-editor {
  width: 100%;
  max-width: 1200px;
  margin: 0 auto;

  .editor-card {
    .n-card-header {
      .n-card-header__main {
        font-size: 18px;
        font-weight: 600;
      }
    }
  }

  .empty-state {
    padding: 40px 0;
    text-align: center;
  }

  .epics-container {
    width: 100%;
  }

  .mobile-collapse {
    .n-collapse-item__header {
      padding: 12px 16px;

      .n-collapse-item__header-main {
        font-weight: 500;
      }
    }
  }

  .epic-content {
    padding: 16px 0;

    .epic-edit-card {
      margin-bottom: 16px;
      border: 1px solid var(--n-border-color);

      .n-form {
        .n-form-item {
          margin-bottom: 16px;
        }
      }
    }

    .stories-section {
      .stories-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 16px;

        h4 {
          margin: 0;
          font-size: 16px;
          font-weight: 500;
        }

        @media (max-width: 768px) {
          flex-direction: column;
          align-items: flex-start;
          gap: 8px;
        }
      }

      .stories-list {
        .story-item {
          display: flex;
          align-items: flex-start;
          gap: 12px;
          padding: 16px;
          margin-bottom: 12px;
          background: var(--n-color);
          border: 1px solid var(--n-border-color);
          border-radius: 8px;
          transition: all 0.3s ease;

          &.selected {
            border-color: var(--n-primary-color);
            background: var(--n-primary-color-hover);
          }

          &.completed {
            opacity: 0.7;

            .story-title-input {
              text-decoration: line-through;
            }
          }

          &:hover {
            background: var(--n-color-hover);
          }

          .drag-handle {
            cursor: grab;
            color: var(--n-text-color-3);
            padding: 4px;

            &:hover {
              color: var(--n-text-color-2);
            }

            &:active {
              cursor: grabbing;
            }
          }

          .story-checkbox {
            margin-top: 4px;
          }

          .story-content {
            flex: 1;

            .story-header {
              display: flex;
              align-items: center;
              gap: 12px;
              margin-bottom: 8px;

              .story-number {
                font-weight: 600;
                color: var(--n-primary-color);
                min-width: 60px;
              }

              .story-title-input {
                flex: 1;
                font-weight: 500;
              }

              @media (max-width: 768px) {
                flex-direction: column;
                align-items: flex-start;
                gap: 8px;
              }
            }

            .n-form {
              .n-form-item {
                margin-bottom: 12px;
              }
            }
          }

          .delete-button {
            margin-top: 4px;
          }

          @media (max-width: 768px) {
            flex-direction: column;
            align-items: stretch;
            gap: 8px;

            .drag-handle {
              align-self: flex-start;
            }

            .story-checkbox {
              align-self: flex-start;
            }

            .delete-button {
              align-self: flex-end;
            }
          }
        }
      }
    }
  }
}

// 移动端优化
@media (max-width: 768px) {
  .epic-story-editor {
    .editor-card {
      .n-card-header {
        .n-card-header__extra {
          .n-space {
            flex-direction: column;
            width: 100%;

            .n-button {
              width: 100%;
              margin-bottom: 8px;
            }
          }
        }
      }
    }
  }
}
</style>