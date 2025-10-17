import { httpService } from '@/utils/http'
import { defineStore } from 'pinia'

export const useBackendStore = defineStore('backend', () => {
    // 健康检查方法
    const healthCheck = async () => {
        try {
            const response = await httpService.get('/health')
            return response as any
        } catch (error) {
            throw new Error('后端服务健康检查失败')
        }
    }

    return {
        healthCheck
    }
})