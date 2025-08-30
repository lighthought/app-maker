export interface Project {
  id: string
  name: string
  description: string
  status: 'draft' | 'in_progress' | 'completed' | 'failed'
  createdAt: string
  updatedAt: string
  userId: string
  progress: number
  tags: string[]
}

export interface CreateProjectData {
  name: string
  description: string
  tags?: string[]
}

export interface UpdateProjectData {
  name?: string
  description?: string
  status?: Project['status']
  progress?: number
  tags?: string[]
}