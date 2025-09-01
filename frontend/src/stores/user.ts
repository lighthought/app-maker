import { defineStore } from 'pinia'
import { ref } from 'vue'
import { httpService } from '@/utils/http'
import type { User, LoginCredentials, RegisterCredentials } from '@/types/user'

export const useUserStore = defineStore('user', () => {
  const user = ref<User | null>(null)
  const token = ref<string>('')
  const refreshToken = ref<string>('')
  const permissions = ref<string[]>([])
  const isAuthenticated = ref(false)

  // 从 localStorage 恢复状态
  const initFromStorage = () => {
    const storedToken = localStorage.getItem('token')
    const storedRefreshToken = localStorage.getItem('refreshToken')
    const storedUser = localStorage.getItem('user')
    
    if (storedToken && storedUser && storedUser !== 'null' && storedUser !== 'undefined') {
      try {
        token.value = storedToken
        refreshToken.value = storedRefreshToken || ''
        user.value = JSON.parse(storedUser)
        isAuthenticated.value = true
      } catch (error) {
        console.warn('解析用户数据失败，清除本地存储:', error)
        clearAuth()
      }
    }
  }

  // 登录
  const login = async (credentials: LoginCredentials) => {
    try {
      const response = await httpService.post<{
        code: number
        message: string
        data?: {
          user: User
          access_token: string
          refresh_token: string
          expires_in: number
        }
      }>('/auth/login', credentials)

      // 检查响应码
      if (response.code === 0 && response.data) {
        const { access_token, refresh_token, user: userData } = response.data
        
        // 保存到 store
        token.value = access_token
        refreshToken.value = refresh_token
        user.value = userData
        isAuthenticated.value = true

        // 保存到 localStorage
        localStorage.setItem('token', access_token)
        localStorage.setItem('refreshToken', refresh_token)
        localStorage.setItem('user', JSON.stringify(userData))

        return { success: true, message: '登录成功' }
      } else {
        // 处理业务逻辑错误（如用户不存在、密码错误等）
        return { success: false, message: response.message || '登录失败' }
      }
    } catch (error: any) {
      console.error('登录失败:', error)
      // 处理网络错误或其他异常
      const message = error.response?.data?.message || '登录失败，请检查网络连接'
      return { success: false, message }
    }
  }

  // 注册
  const register = async (credentials: RegisterCredentials) => {
    try {
      const response = await httpService.post<{
        code: number
        message: string
        data?: {
          user: User
          access_token: string
          refresh_token: string
          expires_in: number
        }
      }>('/auth/register', credentials)

      // 检查响应码
      if (response.code === 0 && response.data) {
        const { access_token, refresh_token, user: userData } = response.data
        
        // 保存到 store
        token.value = access_token
        refreshToken.value = refresh_token
        user.value = userData
        isAuthenticated.value = true

        // 保存到 localStorage
        localStorage.setItem('token', access_token)
        localStorage.setItem('refreshToken', refresh_token)
        localStorage.setItem('user', JSON.stringify(userData))

        return { success: true, message: '注册成功' }
      } else {
        // 处理业务逻辑错误（如邮箱已存在、用户名已存在等）
        return { success: false, message: response.message || '注册失败' }
      }
    } catch (error: any) {
      console.error('注册失败:', error)
      // 处理网络错误或其他异常
      const message = error.response?.data?.message || '注册失败，请检查网络连接'
      return { success: false, message }
    }
  }

  // 登出
  const logout = async () => {
    try {
      // 调用后端登出接口
      if (token.value) {
        await httpService.post('/users/logout')
      }
    } catch (error) {
      console.error('登出请求失败:', error)
    } finally {
      // 清除本地状态
      clearAuth()
    }
  }

  // 刷新令牌
  const refreshAuth = async () => {
    try {
      if (!refreshToken.value) {
        console.warn('没有刷新令牌，无法刷新认证')
        return false
      }

      const response = await httpService.post<{
        code: number
        message: string
        data?: {
          access_token: string
          refresh_token: string
          expires_in: number
        }
      }>('/auth/refresh', null, {
        params: { refresh_token: refreshToken.value }
      })

      // 检查响应码
      if (response.code === 0 && response.data) {
        const { access_token, refresh_token } = response.data
        
        // 更新令牌
        token.value = access_token
        refreshToken.value = refresh_token

        // 更新 localStorage
        localStorage.setItem('token', access_token)
        localStorage.setItem('refreshToken', refresh_token)

        return true
      } else {
        console.error('刷新令牌失败:', response.message)
        return false
      }
    } catch (error) {
      console.error('刷新令牌失败:', error)
      return false
    }
  }

  // 清除认证状态
  const clearAuth = () => {
    user.value = null
    token.value = ''
    refreshToken.value = ''
    permissions.value = []
    isAuthenticated.value = false

    // 清除 localStorage
    localStorage.removeItem('token')
    localStorage.removeItem('refreshToken')
    localStorage.removeItem('user')
  }

  // 更新用户信息
  const updateProfile = async (profile: Partial<User>) => {
    try {
      const response = await httpService.put<{ user: User }>('/users/profile', profile)
      
      // 直接使用响应数据
      user.value = response.user
      localStorage.setItem('user', JSON.stringify(response.user))
      return { success: true, message: '更新成功' }
    } catch (error: any) {
      console.error('更新用户信息失败:', error)
      const message = error.response?.data?.message || '更新失败'
      return { success: false, message }
    }
  }

  // 修改密码
  const changePassword = async (oldPassword: string, newPassword: string) => {
    try {
      const response = await httpService.put<{ message: string }>('/users/change-password', {
        old_password: oldPassword,
        new_password: newPassword
      })
      
      return { success: true, message: '密码修改成功' }
    } catch (error: any) {
      console.error('修改密码失败:', error)
      const message = error.response?.data?.message || '修改密码失败'
      return { success: false, message }
    }
  }

  // 检查权限
  const hasPermission = (permission: string) => {
    return permissions.value.includes(permission)
  }

  // 初始化
  initFromStorage()

  return {
    user,
    token,
    refreshToken,
    permissions,
    isAuthenticated,
    login,
    register,
    logout,
    refreshAuth,
    updateProfile,
    changePassword,
    hasPermission,
    clearAuth
  }
})