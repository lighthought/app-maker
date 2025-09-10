import { defineStore } from 'pinia'
import { ref } from 'vue'
import { httpService } from '@/utils/http'
import type { Project, CreateProjectData, UpdateProjectData, ProjectListRequest, PaginationResponse, PaginationInfo, ConversationMessage, DevStage } from '@/types/project'

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
  const deleteProject = async (projectId: string) => {
    projectStatus.value = 'loading'
    try {
      const response = await httpService.delete<{
        code: number
        message: string
      }>(`/projects/${projectId}`)

      if (response.code === 0) {
        // 确保 projects.value 是数组
        if (Array.isArray(projects.value)) {
          const index = projects.value.findIndex(p => p.id === projectId)
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
  const getProject = async (projectId: string) => {
    try {
      const response = await httpService.get<{
        code: number
        message: string
        data: Project
      }>(`/projects/${projectId}`)

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

  // 下载项目
  const downloadProject = async (projectId: string) => {
    try {
      // 使用 httpService 的 download 方法
      const response = await httpService.get<{
        code: number
        message: string
        data: string
      }>(`/projects/download/${projectId}`)

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
  const getProjectMessages = async (projectId: string, page = 1, pageSize = 50) => {
    try {
      const response = await httpService.get<{
        code: number
        message: string
        data: PaginationResponse<ConversationMessage>
      }>(`/chat/messages/${projectId}`, {
        params: { page, pageSize }
      })

      if (response.code === 0 && response.data) {
        return response.data
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
  const addChatMessage = async (projectId: string, message: Omit<ConversationMessage, 'id' | 'timestamp'>) => {
    try {
      const response = await httpService.post<{
        code: number
        message: string
        data: ConversationMessage
      }>(`/chat/chat/${projectId}`, message)

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

  // 获取项目开发阶段
  const getProjectStages = async (projectId: string) => {
    try {
      const response = await httpService.get<{
        code: number
        message: string
        data: DevStage[]
      }>(`/projects/${projectId}/stages`)

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

  return {
    projects,
    currentProject,
    projectStatus,
    pagination,
    fetchProjects,
    createProject,
    deleteProject,
    getProject,
    setCurrentProject,
    getProjectMessages,
    addChatMessage,
    getProjectStages,
    downloadProject
  }
})