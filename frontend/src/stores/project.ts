import { defineStore } from 'pinia'
import { ref } from 'vue'
import { httpService } from '@/utils/http'
import type { Project, CreateProjectData, UpdateProjectData, ProjectListRequest, PaginationResponse } from '@/types/project'

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
    setCurrentProject
  }
})