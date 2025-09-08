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
  tags: TagInfo[]
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

export interface TagInfo {
  id: string
  name: string
  color: string
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
  pageSize: number
  totalPages: number
  data: T[]
  hasNext: boolean
  hasPrevious: boolean
}

// 任务相关类型
export interface Task {
  id: string
  projectId: string
  type: string
  status: 'pending' | 'in_progress' | 'completed' | 'failed'
  priority: number
  description: string
  startedAt?: string
  completedAt?: string
  createdAt: string
}

export interface TaskLog {
  id: string
  taskId: string
  level: 'info' | 'success' | 'warning' | 'error'
  message: string
  createdAt: string
}

// 对话消息类型
export interface ConversationMessage {
  id: string
  type: 'user' | 'agent' | 'system'
  agentRole?: 'dev' | 'pm' | 'arch' | 'ux' | 'qa' | 'ops'
  agentName?: string
  content: string
  timestamp: string
  isMarkdown?: boolean
  markdownContent?: string
  isExpanded?: boolean
}

// 开发阶段类型
export interface DevStage {
  id: string
  name: string
  status: 'pending' | 'in_progress' | 'completed' | 'failed'
  progress: number
  description: string
}