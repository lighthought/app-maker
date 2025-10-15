
// 项目信息更新
export interface ProjectInfoUpdate {
  id: string
  guid: string
  name: string
  status: string
  description: string
  preview_url?: string
}

export interface Project {
  id: string
  guid: string
  name: string
  description: string
  status: 'pending' | 'in_progress' | 'done' | 'failed' | 'paused'
  requirements: string
  projectPath: string
  backendPort: number
  frontendPort: number
  preview_url?: string
  userId: string
  user?: UserInfo
  cli_tool?: string
  ai_model?: string
  model_provider?: string
  model_api_url?: string
  api_token?: string
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

export interface UpdateProjectFormData {
  name: string
  description: string
  cliTool: string
  aiModel: string
  modelProvider: string
  modelApiUrl: string
}

export interface ProjectListRequest {
  page?: number
  pageSize?: number
  status?: string
  tagIds?: string[]
  userId?: string
  search?: string
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
  agent_role?: 'user' | 'dev' | 'pm' | 'po' | 'architect' | 'ux-expert' | 'analyst' | 'qa' | 'ops'
  agent_name?: string
  content: string
  is_markdown?: boolean
  markdown_content?: string
  is_expanded?: boolean
  has_question?: boolean
  waiting_user_response?: boolean
  created_at: string
  updated_at: string
}

// 开发阶段类型
export interface DevStage {
  id: string
  name: string
  status: 'pending' | 'in_progress' | 'done' | 'failed' | 'paused'
  progress: number
  description: string
  failed_reason: string
  task_id: string
}
