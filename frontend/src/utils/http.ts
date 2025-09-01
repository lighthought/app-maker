import axios from 'axios'
import type { AxiosInstance, AxiosRequestConfig, AxiosResponse } from 'axios'
import { useUserStore } from '@/stores/user'

class HttpService {
  private instance: AxiosInstance
  private isRefreshing = false
  private failedQueue: Array<{
    resolve: (value?: any) => void
    reject: (error?: any) => void
  }> = []
  
  constructor() {
    this.instance = axios.create({
      baseURL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080',
      timeout: 10000,
      headers: {
        'Content-Type': 'application/json'
      }
    })
    
    this.setupInterceptors()
  }
  
  private setupInterceptors() {
    // 请求拦截器
    this.instance.interceptors.request.use(
      (config) => {
        const userStore = useUserStore()
        if (userStore.token) {
          config.headers.Authorization = `Bearer ${userStore.token}`
        }
        return config
      },
      (error) => Promise.reject(error)
    )
    
    // 响应拦截器
    this.instance.interceptors.response.use(
      (response: AxiosResponse) => {
        // 直接返回响应数据，让业务层处理成功/失败逻辑
        return response.data
      },
      async (error) => {
        const originalRequest = error.config
        
        // 如果是401错误且不是刷新token的请求，尝试刷新token
        if (error.response?.status === 401 && !originalRequest._retry) {
          if (this.isRefreshing) {
            // 如果正在刷新，将请求加入队列
            return new Promise((resolve, reject) => {
              this.failedQueue.push({ resolve, reject })
            }).then(() => {
              return this.instance(originalRequest)
            }).catch((err) => {
              return Promise.reject(err)
            })
          }
          
          originalRequest._retry = true
          this.isRefreshing = true
          
          const userStore = useUserStore()
          
          try {
            // 检查是否有刷新令牌，如果没有则直接返回错误响应
            if (!userStore.refreshToken) {
              // 没有刷新令牌，直接返回错误响应数据
              if (error.response?.data) {
                return Promise.resolve(error.response.data)
              }
              return Promise.reject(error)
            }
            
            const refreshed = await userStore.refreshAuth()
            if (refreshed) {
              // 处理队列中的请求
              this.failedQueue.forEach(({ resolve }) => {
                resolve()
              })
              this.failedQueue = []
              
              // 重试原始请求
              return this.instance(originalRequest)
            } else {
              // 刷新失败，清除认证状态
              userStore.clearAuth()
              window.location.href = '/auth'
              return Promise.reject(error)
            }
          } catch (refreshError) {
            // 刷新失败，清除认证状态
            userStore.clearAuth()
            window.location.href = '/auth'
            return Promise.reject(refreshError)
          } finally {
            this.isRefreshing = false
          }
        }
        
        // 对于其他错误，返回错误响应数据而不是抛出异常
        // 这样业务层可以统一处理成功和失败的情况
        if (error.response?.data) {
          return Promise.resolve(error.response.data)
        }
        
        return Promise.reject(error)
      }
    )
  }
  
  public get<T>(url: string, config?: AxiosRequestConfig): Promise<T> {
    return this.instance.get(url, config)
  }
  
  public post<T>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> {
    return this.instance.post(url, data, config)
  }
  
  public put<T>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> {
    return this.instance.put(url, data, config)
  }
  
  public delete<T>(url: string, config?: AxiosRequestConfig): Promise<T> {
    return this.instance.delete(url, config)
  }
  
  public patch<T>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> {
    return this.instance.patch(url, data, config)
  }
}

export const httpService = new HttpService()