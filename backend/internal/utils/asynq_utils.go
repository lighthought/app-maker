package utils

import (
	"autocodeweb-backend/internal/models"
	"autocodeweb-backend/pkg/logger"

	"github.com/hibiken/asynq"
)

// updateResult 是一个帮助函数，用于将任务进度更新到Redis。
// 这里假设使用一个Redis Hash结构，key为`task:progress:<task_id>`。
func UpdateResult(resultWriter *asynq.ResultWriter, status string, progress int, message string) {
	if resultWriter == nil {
		logger.Error("resultWriter is nil, can't update result")
		return
	}

	data := models.TaskResult{
		TaskID:    resultWriter.TaskID(),
		Status:    status,
		Progress:  progress,
		Message:   message,
		UpdatedAt: GetCurrentTime(),
	}
	resultWriter.Write(data.ToBytes())
}
