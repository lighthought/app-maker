export interface User {
  id: string
  username: string
  email: string
  name?: string
  avatar?: string
  role?: string
  status?: 'active' | 'inactive' | 'suspended'
  default_cli_tool?: string
  default_ai_model?: string
  default_model_provider?: string
  default_model_api_url?: string
  default_api_token?: string
  auto_go_next?: boolean
  createdAt: string
  updatedAt: string
}

export interface LoginCredentials {
  email: string
  password: string
}

export interface RegisterCredentials {
  username: string
  email: string
  password: string
}

export interface UserProfile {
  name?: string
  avatar?: string
  email?: string
}

export interface ChangePasswordRequest {
  old_password: string
  new_password: string
}