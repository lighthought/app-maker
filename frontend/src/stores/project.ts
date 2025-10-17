import { defineStore } from 'pinia'
import { ref } from 'vue'
import { httpService } from '@/utils/http'
import type { Project, CreateProjectData, ProjectListRequest, PaginationInfo, ConversationMessage, DevStage } from '@/types/project'

export const useProjectStore = defineStore('project', () => {
  const projects = ref<Project[]>([])
  const currentProject = ref<Project | null>(null)
  const projectStatus = ref<'idle' | 'loading' | 'success' | 'error'>('idle')
  const pagination = ref<PaginationInfo>({
    total: 0,
    page: 1,
    pageSize: 10,
    totalPages: 0,
    hasNext: false,
    hasPrevious: false
  })

  // 获取项目列表
  const fetchProjects = async (params?: ProjectListRequest) => {
    projectStatus.value = 'loading'
    try {
      const response = await httpService.get<{
        code: number
        message: string
        total: number
        page: number
        page_size: number
        total_pages: number
        data: Project[]
        has_next: boolean
        has_previous: boolean
        timestamp: string
      }>('/projects', { params })

      if (response.code === 0 && response.data) {
        // 确保 projects.value 是数组
        if (!Array.isArray(projects.value)) {
          projects.value = []
        }
        projects.value = response.data || []
        pagination.value = {
          total: response.total,
          page: response.page,
          pageSize: response.page_size,
          totalPages: response.total_pages,
          hasNext: response.has_next,
          hasPrevious: response.has_previous
        }
        projectStatus.value = 'success'
      } else {
        console.error('获取项目列表失败:', response.message)
        projectStatus.value = 'error'
      }
    } catch (error) {
      console.error('获取项目列表失败:', error)
      projectStatus.value = 'error'
    }
  }

  // 创建项目
  const createProject = async (projectData: CreateProjectData) => {
    projectStatus.value = 'loading'
    try {
      const response = await httpService.post<{
        code: number
        message: string
        data: Project
      }>('/projects', projectData)

      if (response.code === 0 && response.data) {
        // 确保 projects.value 是数组
        if (!Array.isArray(projects.value)) {
          projects.value = []
        }
        projects.value.unshift(response.data)
        projectStatus.value = 'success'
        return response.data
      } else {
        console.error('创建项目失败:', response.message)
        projectStatus.value = 'error'
        throw new Error(response.message || '创建项目失败')
      }
    } catch (error) {
      console.error('创建项目失败:', error)
      projectStatus.value = 'error'
      throw error
    }
  }

  // 删除项目
  const deleteProject = async (projectGuid: string) => {
    projectStatus.value = 'loading'
    try {
      const response = await httpService.delete<{
        code: number
        message: string
      }>(`/projects/${projectGuid}`)

      if (response.code === 0) {
        // 确保 projects.value 是数组
        if (Array.isArray(projects.value)) {
          const index = projects.value.findIndex(p => p.guid === projectGuid)
          if (index !== -1) {
            projects.value.splice(index, 1)
          }
        }
        projectStatus.value = 'success'
      } else {
        console.error('删除项目失败:', response.message)
        projectStatus.value = 'error'
        throw new Error(response.message || '删除项目失败')
      }
    } catch (error) {
      console.error('删除项目失败:', error)
      projectStatus.value = 'error'
      throw error
    }
  }

  // 获取单个项目
  const getProject = async (projectGuid: string) => {
    try {
      const response = await httpService.get<{
        code: number
        message: string
        data: Project
      }>(`/projects/${projectGuid}`)

      if (response.code === 0 && response.data) {
        return response.data
      } else {
        console.error('获取项目详情失败:', response.message)
        return null
      }
    } catch (error) {
      console.error('获取项目详情失败:', error)
      return null
    }
  }

  // 更新项目
  const updateProject = async (projectGuid: string, updateData: {
    name?: string
    description?: string
    cli_tool?: string | null
    ai_model?: string | null
    model_provider?: string | null
    model_api_url?: string | null
  }) => {
    projectStatus.value = 'loading'
    try {
      const response = await httpService.put<{
        code: number
        message: string
        data: Project
      }>(`/projects/${projectGuid}`, updateData)

      if (response.code === 0 && response.data) {
        // 更新本地项目列表中的项目
        if (Array.isArray(projects.value)) {
          const index = projects.value.findIndex(p => p.guid === projectGuid)
          if (index !== -1) {
            projects.value[index] = response.data
          }
        }
        // 更新当前项目
        if (currentProject.value?.guid === projectGuid) {
          currentProject.value = response.data
        }
        projectStatus.value = 'success'
        return response.data
      } else {
        console.error('更新项目失败:', response.message)
        projectStatus.value = 'error'
        throw new Error(response.message || '更新项目失败')
      }
    } catch (error) {
      console.error('更新项目失败:', error)
      projectStatus.value = 'error'
      throw error
    }
  }

  // 下载项目
  const downloadProject = async (projectGuid: string) => {
    try {
      // 使用 httpService 的 download 方法
      const response = await httpService.get<{
        code: number
        message: string
        data: string
      }>(`/projects/download/${projectGuid}`)

      if (response.code === 0 && response.data) {
        return response.data
      } else {
        console.error('下载项目失败:', response.message)
        return null
      }
    } catch (error) {
      console.error('下载项目失败:', error)
      return null
    }
  }

  // 设置当前项目
  const setCurrentProject = (project: Project | null) => {
    currentProject.value = project
  }

  
  // 获取项目对话历史
  const getProjectMessages = async (projectGuid: string, page = 1, pageSize = 50) => {
    try {
      const response = await httpService.get<{
        code: number
        message: string
        total: number
        page: number
        page_size: number
        total_pages: number
        data: ConversationMessage[]
        has_next: boolean
        has_previous: boolean
        timestamp: string
      }>(`/chat/messages/${projectGuid}`, {
        params: { page, pageSize }
      })

      if (response.code === 0) {
        return {
          total: response.total,
          page: response.page,
          page_size: response.page_size,
          total_pages: response.total_pages,
          data: response.data || [],
          has_next: response.has_next,
          has_previous: response.has_previous
        }
      } else {
        console.error('获取对话历史失败:', response.message)
        return null
      }
    } catch (error) {
      console.error('获取对话历史失败:', error)
      return null
    }
  }

  // 添加对话消息
  const addChatMessage = async (projectGuid: string, message: Omit<ConversationMessage, 'id' | 'created_at' | 'updated_at' | 'project_id'>) => {
    try {
      const response = await httpService.post<{
        code: number
        message: string
        data: ConversationMessage
      }>(`/chat/chat/${projectGuid}`, message)

      if (response.code === 0 && response.data) {
        return response.data
      } else {
        console.error('添加对话消息失败:', response.message)
        return null
      }
    } catch (error) {
      console.error('添加对话消息失败:', error)
      return null
    }
  }

  // 向指定 Agent 发送消息
  const sendMessageToAgent = async (projectGuid: string, agentType: string, content: string) => {
    try {
      const response = await httpService.post<{
        code: number
        message: string
        data: any
      }>(`/chat/send-to-agent/${projectGuid}`, {
        agent_type: agentType,
        content: content
      })

      if (response.code === 0) {
        return response.data
      } else {
        console.error('向 Agent 发送消息失败:', response.message)
        return null
      }
    } catch (error) {
      console.error('向 Agent 发送消息失败:', error)
      return null
    }
  }

  // 获取项目开发阶段
  const getProjectStages = async (projectGuid: string) => {
    try {
      const response = await httpService.get<{
        code: number
        message: string
        data: DevStage[]
      }>(`/projects/${projectGuid}/stages`)

      if (response.code === 0 && response.data) {
        return response.data
      } else {
        console.error('获取开发阶段失败:', response.message)
        return null
      }
    } catch (error) {
      console.error('获取开发阶段失败:', error)
      return null
    }
  }

  // 一键部署项目
  const deployProject = async (projectGuid: string) => {
    try {
      const response = await httpService.post<{
        code: number
        message: string
        data: any
      }>(`/projects/${projectGuid}/deploy`, {})

      if (response.code === 0) {
        return response.data
      } else {
        console.error('部署项目失败:', response.message)
        throw new Error(response.message || '部署项目失败')
      }
    } catch (error) {
      console.error('部署项目失败:', error)
      throw error
    }
  }

  // Epic/Story 相关方法
  const getMvpEpics = async (projectGuid: string) => {
    try {
      const response = await httpService.get<{
        code: number
        message: string
        data: any[]
      }>(`/projects/${projectGuid}/mvp-epics`)

      if (response.code === 0) {
        return response.data || []
      } else {
        console.error('获取 MVP Epics 失败:', response.message)
        return []
      }
    } catch (error) {
      console.error('获取 MVP Epics 失败:', error)
      return []
    }
  }

  const updateEpicOrder = async (projectGuid: string, epicId: string, order: number) => {
    try {
      const response = await httpService.put<{
        code: number
        message: string
      }>(`/projects/${projectGuid}/epics/${epicId}/order`, { order })

      if (response.code === 0) {
        return true
      } else {
        console.error('更新 Epic 排序失败:', response.message)
        return false
      }
    } catch (error) {
      console.error('更新 Epic 排序失败:', error)
      return false
    }
  }

  const updateEpic = async (projectGuid: string, epicId: string, updateData: {
    name?: string
    description?: string
    priority?: string
    estimated_days?: number
  }) => {
    try {
      const response = await httpService.put<{
        code: number
        message: string
      }>(`/projects/${projectGuid}/epics/${epicId}`, updateData)

      if (response.code === 0) {
        return true
      } else {
        console.error('更新 Epic 失败:', response.message)
        return false
      }
    } catch (error) {
      console.error('更新 Epic 失败:', error)
      return false
    }
  }

  const deleteEpic = async (projectGuid: string, epicId: string) => {
    try {
      const response = await httpService.delete<{
        code: number
        message: string
      }>(`/projects/${projectGuid}/epics/${epicId}`)

      if (response.code === 0) {
        return true
      } else {
        console.error('删除 Epic 失败:', response.message)
        return false
      }
    } catch (error) {
      console.error('删除 Epic 失败:', error)
      return false
    }
  }

  const updateStoryOrder = async (projectGuid: string, epicId: string, storyId: string, order: number) => {
    try {
      const response = await httpService.put<{
        code: number
        message: string
      }>(`/projects/${projectGuid}/epics/${epicId}/stories/${storyId}/order`, { order })

      if (response.code === 0) {
        return true
      } else {
        console.error('更新 Story 排序失败:', response.message)
        return false
      }
    } catch (error) {
      console.error('更新 Story 排序失败:', error)
      return false
    }
  }

  const updateStory = async (projectGuid: string, epicId: string, storyId: string, updateData: {
    title?: string
    description?: string
    priority?: string
    estimated_days?: number
    depends?: string
    techs?: string
    content?: string
    acceptance_criteria?: string
  }) => {
    try {
      const response = await httpService.put<{
        code: number
        message: string
      }>(`/projects/${projectGuid}/epics/${epicId}/stories/${storyId}`, updateData)

      if (response.code === 0) {
        return true
      } else {
        console.error('更新 Story 失败:', response.message)
        return false
      }
    } catch (error) {
      console.error('更新 Story 失败:', error)
      return false
    }
  }

  const deleteStory = async (projectGuid: string, epicId: string, storyId: string) => {
    try {
      const response = await httpService.delete<{
        code: number
        message: string
      }>(`/projects/${projectGuid}/epics/${epicId}/stories/${storyId}`)

      if (response.code === 0) {
        return true
      } else {
        console.error('删除 Story 失败:', response.message)
        return false
      }
    } catch (error) {
      console.error('删除 Story 失败:', error)
      return false
    }
  }

  const batchDeleteStories = async (projectGuid: string, storyIds: string[]) => {
    try {
      const response = await httpService.delete<{
        code: number
        message: string
      }>(`/projects/${projectGuid}/epics/stories/batch-delete`, { data: { story_ids: storyIds } })
      
      if (response.code === 0) {
        return true
      } else {
        console.error('批量删除 Stories 失败:', response.message)
        return false
      }
    } catch (error) {
      console.error('批量删除 Stories 失败:', error)
      return false
    }
  }

  const confirmEpicsAndStories = async (projectGuid: string, action: 'confirm' | 'skip' | 'regenerate') => {
    try {
      const response = await httpService.post<{
        code: number
        message: string
      }>(`/projects/${projectGuid}/epics/confirm`, { action })

      if (response.code === 0) {
        return true
      } else {
        console.error('确认 Epics 和 Stories 失败:', response.message)
        return false
      }
    } catch (error) {
      console.error('确认 Epics 和 Stories 失败:', error)
      return false
    }
  }

  return {
    projects,
    currentProject,
    projectStatus,
    pagination,
    fetchProjects,
    createProject,
    updateProject,
    deleteProject,
    getProject,
    setCurrentProject,
    getProjectMessages,
    addChatMessage,
    sendMessageToAgent,
    getProjectStages,
    downloadProject,
    deployProject,
    // Epic/Story 相关方法
    getMvpEpics,
    updateEpicOrder,
    updateEpic,
    deleteEpic,
    updateStoryOrder,
    updateStory,
    deleteStory,
    batchDeleteStories,
    confirmEpicsAndStories
  }
})