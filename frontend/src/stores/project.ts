import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { Project } from '@/types/project'

export const useProjectStore = defineStore('project', () => {
  const projects = ref<Project[]>([])
  const currentProject = ref<Project | null>(null)
  const projectStatus = ref<'idle' | 'loading' | 'success' | 'error'>('idle')

  // 模拟项目数据
  const mockProjects: Project[] = [
    {
      id: '1',
      name: '电商网站',
      description: '一个现代化的电商网站，包含商品展示、购物车、支付功能',
      status: 'in_progress',
      createdAt: '2025-01-15T10:30:00Z',
      updatedAt: '2025-01-20T14:20:00Z',
      userId: 'user1',
      progress: 60,
      tags: ['电商', 'React', 'Node.js']
    },
    {
      id: '2',
      name: '任务管理应用',
      description: '团队协作的任务管理工具，支持任务分配、进度跟踪',
      status: 'completed',
      createdAt: '2025-01-10T09:15:00Z',
      updatedAt: '2025-01-18T16:45:00Z',
      userId: 'user1',
      progress: 100,
      tags: ['协作', 'Vue', 'TypeScript']
    },
    {
      id: '3',
      name: '博客系统',
      description: '个人博客平台，支持文章发布、评论、用户管理',
      status: 'draft',
      createdAt: '2025-01-25T11:00:00Z',
      updatedAt: '2025-01-25T11:00:00Z',
      userId: 'user1',
      progress: 10,
      tags: ['博客', 'Next.js', 'MongoDB']
    },
    {
      id: '4',
      name: '在线教育平台',
      description: '在线学习平台，包含课程管理、视频播放、作业提交',
      status: 'in_progress',
      createdAt: '2025-01-05T08:45:00Z',
      updatedAt: '2025-01-22T13:30:00Z',
      userId: 'user1',
      progress: 75,
      tags: ['教育', 'Vue', 'Node.js']
    },
    {
      id: '5',
      name: '社交媒体应用',
      description: '类似微博的社交媒体应用，支持发布动态、关注用户',
      status: 'failed',
      createdAt: '2025-01-12T15:20:00Z',
      updatedAt: '2025-01-19T10:15:00Z',
      userId: 'user1',
      progress: 0,
      tags: ['社交', 'React Native', 'Firebase']
    },
    {
      id: '6',
      name: '库存管理系统',
      description: '企业库存管理解决方案，支持商品入库、出库、盘点',
      status: 'completed',
      createdAt: '2025-01-08T14:30:00Z',
      updatedAt: '2025-01-16T17:00:00Z',
      userId: 'user1',
      progress: 100,
      tags: ['企业', 'Vue', 'Java']
    },
    {
      id: '7',
      name: '在线聊天应用',
      description: '实时聊天应用，支持私聊、群聊、文件传输',
      status: 'in_progress',
      createdAt: '2025-01-20T12:00:00Z',
      updatedAt: '2025-01-24T09:30:00Z',
      userId: 'user1',
      progress: 45,
      tags: ['聊天', 'React', 'Socket.io']
    },
    {
      id: '8',
      name: '个人理财应用',
      description: '个人财务管理工具，支持收入支出记录、预算管理',
      status: 'draft',
      createdAt: '2025-01-28T16:15:00Z',
      updatedAt: '2025-01-28T16:15:00Z',
      userId: 'user1',
      progress: 5,
      tags: ['理财', 'Vue', 'SQLite']
    }
  ]

  // 获取项目列表
  const fetchProjects = async () => {
    projectStatus.value = 'loading'
    try {
      // 模拟 API 调用延迟
      await new Promise(resolve => setTimeout(resolve, 500))
      projects.value = mockProjects
      projectStatus.value = 'success'
    } catch (error) {
      console.error('获取项目列表失败:', error)
      projectStatus.value = 'error'
    }
  }

  // 创建项目
  const createProject = async (projectData: Omit<Project, 'id' | 'createdAt' | 'updatedAt'>) => {
    projectStatus.value = 'loading'
    try {
      const newProject: Project = {
        ...projectData,
        id: Date.now().toString(),
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString()
      }
      projects.value.unshift(newProject)
      projectStatus.value = 'success'
      return newProject
    } catch (error) {
      console.error('创建项目失败:', error)
      projectStatus.value = 'error'
      throw error
    }
  }

  // 更新项目
  const updateProject = async (projectId: string, updates: Partial<Project>) => {
    projectStatus.value = 'loading'
    try {
      const index = projects.value.findIndex(p => p.id === projectId)
      if (index !== -1) {
        projects.value[index] = {
          ...projects.value[index],
          ...updates,
          updatedAt: new Date().toISOString()
        }
      }
      projectStatus.value = 'success'
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
      const index = projects.value.findIndex(p => p.id === projectId)
      if (index !== -1) {
        projects.value.splice(index, 1)
      }
      projectStatus.value = 'success'
    } catch (error) {
      console.error('删除项目失败:', error)
      projectStatus.value = 'error'
      throw error
    }
  }

  // 获取单个项目
  const getProject = (projectId: string) => {
    return projects.value.find(p => p.id === projectId) || null
  }

  // 设置当前项目
  const setCurrentProject = (project: Project | null) => {
    currentProject.value = project
  }

  return {
    projects,
    currentProject,
    projectStatus,
    fetchProjects,
    createProject,
    updateProject,
    deleteProject,
    getProject,
    setCurrentProject
  }
})