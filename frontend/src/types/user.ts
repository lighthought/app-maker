export interface User {
  id: string
  username: string
  email: string
  avatar?: string
  role: 'admin' | 'user'
  createdAt: string
  updatedAt: string
}

export interface LoginCredentials {
  username: string
  password: string
}

export interface RegisterData {
  username: string
  email: string
  password: string
  confirmPassword: string
}

export interface UserProfile {
  username?: string
  email?: string
  avatar?: string
}