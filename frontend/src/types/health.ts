// 健康检查相关类型定义

// Agent 健康检查响应
export interface AgentHealthResponse {
  status: string
  version: string
  environment: Record<string, string>
  checked_at: string
}

// 健康检查状态
export type HealthStatus = 'checking' | 'ok' | 'error' | 'warning'
// 服务状态映射
export interface ServiceStatus {
    name: string
    status: HealthStatus
    message: string
    version?: string
    lastChecked?: string
  }
  

// Backend 健康检查响应
export interface BackendHealthResponse {
  status: 'healthy' | 'degraded' | 'unhealthy'
  service: string
  version: string
  timestamp: string
  services: ServiceStatus[]
  agent?: AgentHealthResponse
}

