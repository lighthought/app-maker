import { defineStore } from 'pinia'
import { ref } from 'vue'
import { httpService } from '@/utils/http'
import type { User, LoginCredentials, RegisterCredentials } from '@/types/user'
import axios from 'axios'
import { AppConfig } from '@/utils/config'

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
        
        // 启动时验证 token 有效性
        validateTokenOnStartup()
      } catch (error) {
        console.warn('解析用户数据失败，清除本地存储:', error)
        clearAuth()
      }
    }
  }

  // 启动时验证 token 有效性
  const validateTokenOnStartup = async () => {
    try {
      // 尝试调用一个需要认证的接口来验证 token
      await httpService.get('/users/profile')
      return true
    } catch (error: any) {
      console.warn('启动时 token 验证失败:', error)
      
      // 如果验证失败，尝试刷新 token
      if (refreshToken.value) {
        const refreshed = await refreshAuth()
        if (refreshed) {
          return true
        }
      }
      
      // 刷新也失败，抛出错误让路由守卫处理
      throw new Error('Token validation failed')
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
        // 使用原始的 axios 实例，避免被拦截器处理
        const directAxios = axios.create({
          baseURL: `${AppConfig.getInstance().getApiBaseUrl()}`,
          timeout: 5000
        })
        
        await directAxios.post('/users/logout', null, {
          headers: {
            'Authorization': `Bearer ${token.value}`,
            'Content-Type': 'application/json'
          }
        })
      }
    } catch (error: any) {
      console.error('登出请求失败:', error)
      // 如果登出失败（比如401错误），也要清除本地状态
      if (error.response?.status === 401) {
        console.warn('登出时token已失效，直接清除本地状态')
      }
    } finally {
      // 无论成功还是失败，都要清除本地状态
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

  // 获取用户设置
  const getUserSettings = async () => {
    try {
      const response = await httpService.get<{
        code: number
        message: string
        data: {
          default_cli_tool: string
          default_ai_model: string
          default_model_provider: string
          default_model_api_url: string
          default_api_token: string
          auto_go_next: boolean
        }
      }>('/users/settings')
      
      if (response.code === 0 && response.data) {
        return { success: true, data: response.data }
      } else {
        return { success: false, message: response.message || '获取设置失败', data: null }
      }
    } catch (error: any) {
      console.error('获取用户设置失败:', error)
      const message = error.response?.data?.message || '获取设置失败'
      return { success: false, message, data: null }
    }
  }

  // 更新用户设置
  const updateUserSettings = async (settings: {
    default_cli_tool?: string
    default_ai_model?: string
    default_model_provider?: string
    default_model_api_url?: string
    default_api_token?: string
    auto_go_next?: boolean
  }) => {
    try {
      const response = await httpService.put<{
        code: number
        message: string
        data?: any
      }>('/users/settings', settings)

      if (response.code === 0) {
        return { success: true, message: '设置保存成功' }
      } else {
        return { success: false, message: response.message || '设置保存失败' }
      }
    } catch (error: any) {
      console.error('更新用户设置失败:', error)
      const message = error.response?.data?.message || '设置保存失败'
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
    validateTokenOnStartup,
    updateProfile,
    changePassword,
    getUserSettings,
    updateUserSettings,
    hasPermission,
    clearAuth
  }
})