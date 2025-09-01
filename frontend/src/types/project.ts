export interface Project {
  id: string
  name: string
  description: string
  status: 'draft' | 'in_progress' | 'completed' | 'failed'
  requirements: string
  projectPath: string
  backendPort: number
  frontendPort: number
  userId: string
  user?: UserInfo
  tags: TagInfo[]
  createdAt: string
  updatedAt: string
}

export interface UserInfo {
  id: string
  email: string
  username: string
  role: string
  status: string
  createdAt: string
}

export interface TagInfo {
  id: string
  name: string
  color: string
}

export interface CreateProjectData {
  name: string
  description: string
  requirements: string
  backendPort?: number
  frontendPort?: number
  tagIds?: string[]
}

export interface UpdateProjectData {
  name?: string
  description?: string
  requirements?: string
  status?: Project['status']
  backendPort?: number
  frontendPort?: number
  tagIds?: string[]
}

export interface ProjectListRequest {
  page?: number
  pageSize?: number
  status?: string
  tagIds?: string[]
  userId?: string
  search?: string
}

export interface PaginationResponse<T> {
  total: number
  page: number
  pageSize: number
  totalPages: number
  data: T[]
  hasNext: boolean
  hasPrevious: boolean
}