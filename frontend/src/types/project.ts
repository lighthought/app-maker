export interface Project {
  id: string
  name: string
  description: string
  status: 'draft' | 'in_progress' | 'completed' | 'failed'
  userId: string
  tags: string[]
  createdAt: string
  updatedAt: string
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
  tags?: string[]
}