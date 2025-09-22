export interface Project {
  id: string
  name: string
  description: string
  status: 'draft' | 'in_progress' | 'completed' | 'failed'
  requirements: string
  projectPath: string
  backendPort: number
  frontendPort: number
  previewUrl?: string
  userId: string
  user?: UserInfo
  created_at: string
  updated_at: string
}

export interface UserInfo {
  id: string
  email: string
  username: string
  role: string
  status: string
  createdAt: string
}

export interface CreateProjectData {
  requirements: string
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
  page_size: number
  total_pages: number
  data: T[]
  has_next: boolean
  has_previous: boolean
}

// 前端使用的分页信息接口（驼峰命名）
export interface PaginationInfo {
  total: number
  page: number
  pageSize: number
  totalPages: number
  hasNext: boolean
  hasPrevious: boolean
}


// 对话消息类型
export interface ConversationMessage {
  id: string
  project_id: string
  type: 'user' | 'agent' | 'system'
  agent_role?: 'dev' | 'pm' | 'po' | 'architect' | 'ux-expert' | 'analyst' | 'qa' | 'ops'
  agent_name?: string
  content: string
  is_markdown?: boolean
  markdown_content?: string
  is_expanded?: boolean
  created_at: string
  updated_at: string
}

// 开发阶段类型
export interface DevStage {
  id: string
  name: string
  status: 'pending' | 'in_progress' | 'completed' | 'failed'
  progress: number
  description: string
}