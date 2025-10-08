import { httpService } from '@/utils/http'
import { defineStore } from 'pinia'
import { ref } from 'vue'
import { TaskResult } from '@/types/task'

export const useTaskStore = defineStore('task', () => {

    const taskResult = ref<TaskResult | null>(null)

    const getTaskStatus = async (taskId: string) => {
        try {
            const response = await httpService.get<{
                code: number
                message: string
                data: TaskResult
            }>(`/tasks/${taskId}`)

            if (response.code === 0 && response.data) {
                taskResult.value = response.data
                return response.data
            } else {
                console.error('获取任务状态失败:', response.message)
                return null
            }
        } catch (error) {
            console.error('获取任务状态失败:', error)
            return null
        }
    }

    const retryTask = async (taskId: string) => {
        try {
            const response = await httpService.post<{
                code: number
                message: string
                data: any
            }>(`/tasks/${taskId}/retry`)

            if (response.code === 0) {
                return { success: true, message: response.message }
            } else {
                return { success: false, message: response.message || '重试任务失败' }
            }
        } catch (error: any) {
            console.error('重试任务失败:', error)
            return { 
                success: false, 
                message: error.message || '重试任务失败' 
            }
        }
    }

    return {
        taskResult,
        getTaskStatus,
        retryTask
    }
})
