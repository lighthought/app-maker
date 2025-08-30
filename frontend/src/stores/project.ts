import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { Project } from '@/types/project'

export const useProjectStore = defineStore('project', () => {
  const projects = ref<Project[]>([])
  const currentProject = ref<Project | null>(null)
  const projectStatus = ref<'idle' | 'loading' | 'error'>('idle')
  
  const createProject = async (projectData: Partial<Project>) => {
    // TODO: 实现创建项目逻辑
    console.log('Create project:', projectData)
  }
  
  const updateProject = async (projectId: string, updates: Partial<Project>) => {
    // TODO: 实现更新项目逻辑
    console.log('Update project:', projectId, updates)
  }
  
  const deleteProject = async (projectId: string) => {
    // TODO: 实现删除项目逻辑
    console.log('Delete project:', projectId)
  }
  
  return {
    projects, currentProject, projectStatus,
    createProject, updateProject, deleteProject
  }
})