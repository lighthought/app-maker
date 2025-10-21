// 应用配置工具
export class AppConfig {
  private static instance: AppConfig
  private apiLogEnabled: boolean = true

  private constructor() {
    // 从环境变量读取配置
    this.apiLogEnabled = import.meta.env.VITE_API_LOG_CONSOLE
  }

  static getInstance(): AppConfig {
    if (!AppConfig.instance) {
      AppConfig.instance = new AppConfig()
    }
    return AppConfig.instance
  }

  // API 日志相关配置
  isApiLogEnabled(): boolean {
    return this.apiLogEnabled
  }

  setApiLogEnabled(enabled: boolean): void {
    this.apiLogEnabled = enabled
  }

  // 获取 API 基础 URL
  getApiBaseUrl(): string {
    return import.meta.env.VITE_API_BASE_URL || '/api/v1'
  }

  // 获取 API 超时时间
  getApiTimeout(): number {
    return 1000 * (import.meta.env.VITE_API_TIMEOUT || 60)
  }

  // 获取当前配置信息
  getConfig(): { apiLogEnabled: boolean; apiBaseUrl: string; apiTimeout: number } {
    return {
      apiLogEnabled: this.apiLogEnabled,
      apiBaseUrl: this.getApiBaseUrl(),
      apiTimeout: this.getApiTimeout()
    }
  }
}

// 导出单例实例
export const appConfig = AppConfig.getInstance()

// 为了向后兼容，导出别名
export const apiLogConfig = {
  isEnabled: () => appConfig.isApiLogEnabled(),
  setEnabled: (enabled: boolean) => appConfig.setApiLogEnabled(enabled)
}
