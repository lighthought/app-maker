import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { User } from '@/types/user'

export const useUserStore = defineStore('user', () => {
  // 状态
  const user = ref<User | null>(null)
  const token = ref<string>('')
  const permissions = ref<string[]>([])
  
  // 计算属性
  const isLoggedIn = computed(() => !!token.value)
  const hasPermission = computed(() => (permission: string) => 
    permissions.value.includes(permission)
  )
  
  // 动作
  const login = async (credentials: { username: string; password: string }) => {
    // TODO: 实现登录逻辑
    console.log('Login:', credentials)
  }
  
  const logout = () => {
    user.value = null
    token.value = ''
    permissions.value = []
    localStorage.removeItem('token')
  }
  
  const updateProfile = async (profile: Partial<User>) => {
    // TODO: 实现更新用户资料逻辑
    console.log('Update profile:', profile)
  }
  
  return {
    user, token, permissions, isLoggedIn, hasPermission,
    login, logout, updateProfile
  }
})