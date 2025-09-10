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

    return {
        taskResult,
        getTaskStatus
    }
})
