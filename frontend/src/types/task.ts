export interface TaskResult {
    task_id: string
    status: string
    progress: number
    message: string
    updated_at: string
}

// 任务重试请求
export interface RetryTaskReq {
    task_id: string
    stage_id: string
}