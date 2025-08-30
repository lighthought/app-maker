export interface User {
  id: string
  username: string
  email: string
  name?: string
  avatar?: string
  role?: string
  status?: 'active' | 'inactive' | 'suspended'
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