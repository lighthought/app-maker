import { httpService } from '@/utils/http'
import { defineStore } from 'pinia'
import type { BackendHealthResponse } from '@/types/health'

export const useBackendStore = defineStore('backend', () => {
    // 健康检查方法
    const healthCheck = async (): Promise<BackendHealthResponse> => {
        try {
            const response = await httpService.get('/health')
            // 根据实际接口响应结构，data 字段包含健康检查数据
            return (response as any).data as BackendHealthResponse
        } catch (error) {
            throw new Error('后端服务健康检查失败')
        }
    }

    return {
        healthCheck
    }
})