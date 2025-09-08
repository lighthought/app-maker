import { defineStore } from 'pinia'
import { ref } from 'vue'
import { httpService } from '@/utils/http'
import type { Project, CreateProjectData, UpdateProjectData, ProjectListRequest, PaginationResponse, ConversationMessage, DevStage } from '@/types/project'

export const useProjectStore = defineStore('project', () => {
  const projects = ref<Project[]>([])
  const currentProject = ref<Project | null>(null)
  const projectStatus = ref<'idle' | 'loading' | 'success' | 'error'>('idle')
  const pagination = ref<{
    total: number
    page: number
    pageSize: number
    totalPages: number
    hasNext: boolean
    hasPrevious: boolean
  }>({
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
        data: PaginationResponse<Project>
      }>('/projects', { params })

      if (response.code === 0 && response.data) {
        projects.value = response.data.data
        pagination.value = {
          total: response.data.total,
          page: response.data.page,
          pageSize: response.data.pageSize,
          totalPages: response.data.totalPages,
          hasNext: response.data.hasNext,
          hasPrevious: response.data.hasPrevious
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

  // 更新项目
  const updateProject = async (projectId: string, updates: UpdateProjectData) => {
    projectStatus.value = 'loading'
    try {
      const response = await httpService.put<{
        code: number
        message: string
        data: Project
      }>(`/projects/${projectId}`, updates)

      if (response.code === 0 && response.data) {
        const index = projects.value.findIndex(p => p.id === projectId)
        if (index !== -1) {
          projects.value[index] = response.data
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

  // 删除项目
  const deleteProject = async (projectId: string) => {
    projectStatus.value = 'loading'
    try {
      const response = await httpService.delete<{
        code: number
        message: string
      }>(`/projects/${projectId}`)

      if (response.code === 0) {
        const index = projects.value.findIndex(p => p.id === projectId)
        if (index !== -1) {
          projects.value.splice(index, 1)
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

  // 下载项目
  const downloadProject = async (projectId: string) => {
    try {
      // 使用 httpService 的 download 方法
      const blob = await httpService.download(`/projects/${projectId}/download`)
      
      // 验证blob数据
      if (!blob || blob.size === 0) {
        throw new Error('下载的文件为空')
      }
      
      // 创建下载链接
      const url = window.URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `project-${projectId}.zip`
      document.body.appendChild(a)
      a.click()
      window.URL.revokeObjectURL(url)
      document.body.removeChild(a)
      
      console.log('项目下载成功:', projectId)
    } catch (error) {
      console.error('下载项目失败:', error)
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

  // 设置当前项目
  const setCurrentProject = (project: Project | null) => {
    currentProject.value = project
  }


  // 获取项目文件列表
  const getProjectFiles = async (projectId: string, path?: string) => {
    try {
      const response = await httpService.get<{
        code: number
        message: string
        data: Array<{
          name: string
          path: string
          type: 'file' | 'folder'
          size: number
          modifiedAt: string
        }>
      }>(`/projects/${projectId}/files`, {
        params: path ? { path } : {}
      })

      if (response.code === 0 && response.data) {
        return response.data
      } else {
        console.error('获取项目文件失败:', response.message)
        return null
      }
    } catch (error) {
      console.error('获取项目文件失败:', error)
      return null
    }
  }

  // 获取文件内容
  const getFileContent = async (projectId: string, filePath: string) => {
    try {
      const response = await httpService.get<{
        code: number
        message: string
        data: {
          path: string
          content: string
          size: number
          modifiedAt: string
        }
      }>(`/projects/${projectId}/files/content`, {
        params: { filePath }
      })

      if (response.code === 0 && response.data) {
        return response.data
      } else {
        console.error('获取文件内容失败:', response.message)
        return null
      }
    } catch (error) {
      console.error('获取文件内容失败:', error)
      return null
    }
  }


  // 获取项目对话历史
  const getProjectConversations = async (projectId: string, page = 1, pageSize = 50) => {
    try {
      const response = await httpService.get<{
        code: number
        message: string
        data: PaginationResponse<ConversationMessage>
      }>(`/projects/${projectId}/conversations`, {
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
  const addConversationMessage = async (projectId: string, message: Omit<ConversationMessage, 'id' | 'timestamp'>) => {
    try {
      const response = await httpService.post<{
        code: number
        message: string
        data: ConversationMessage
      }>(`/projects/${projectId}/conversations`, message)

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
    updateProject,
    deleteProject,
    downloadProject,
    getProject,
    setCurrentProject,
    getProjectFiles,
    getFileContent,
    getProjectConversations,
    addConversationMessage,
    getProjectStages
  }
})