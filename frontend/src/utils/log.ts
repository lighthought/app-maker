import type { AxiosRequestConfig, AxiosResponse } from 'axios'
import { AppConfig } from './config'

// API 日志工具类
export class ApiLogger {
  private enabled: boolean

  constructor() {
    // 从配置工具读取日志开关
    this.enabled = AppConfig.getInstance().isApiLogEnabled()
  }

  private formatTime(): string {
    return new Date().toISOString()
  }

  private formatRequest(config: AxiosRequestConfig): string {
    const { method, url, data, params, headers } = config
    return `[${this.formatTime()}] ${method?.toUpperCase()} ${url}
Headers: ${JSON.stringify(headers, null, 2)}
${data ? `Body: ${JSON.stringify(data, null, 2)}` : ''}
${params ? `Params: ${JSON.stringify(params, null, 2)}` : ''}`
  }

  private formatResponse(response: AxiosResponse): string {
    const { status, statusText, data, headers } = response
    return `[${this.formatTime()}] Response ${status} ${statusText}
Headers: ${JSON.stringify(headers, null, 2)}
Data: ${JSON.stringify(data, null, 2)}`
  }

  private formatError(error: any): string {
    const { response, request, message } = error
    let errorInfo = `[${this.formatTime()}] Error: ${message}`
    
    if (response) {
      errorInfo += `
Status: ${response.status} ${response.statusText}
Data: ${JSON.stringify(response.data, null, 2)}`
    } else if (request) {
      errorInfo += `
Request: ${JSON.stringify(request, null, 2)}`
    }
    
    return errorInfo
  }

  logRequest(config: AxiosRequestConfig): void {
    if (!this.enabled) return
    console.group(`🚀 API Request`)
    console.log(this.formatRequest(config))
    console.groupEnd()
  }

  logResponse(response: AxiosResponse): void {
    if (!this.enabled) return
    console.group(`✅ API Response`)
    console.log(this.formatResponse(response))
    console.groupEnd()
  }

  logError(error: any): void {
    if (!this.enabled) return
    console.group(`❌ API Error`)
    console.error(this.formatError(error))
    console.groupEnd()
  }

  // 切换日志开关
  setEnabled(enabled: boolean): void {
    this.enabled = enabled
    // 同步更新配置工具
    apiLogConfig.setEnabled(enabled)
  }

  // 获取当前状态
  isEnabled(): boolean {
    return this.enabled
  }
}

// 导出单例实例
export const apiLogger = new ApiLogger()
