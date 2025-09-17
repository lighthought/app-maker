package tasks

import (
	"autocodeweb-backend/internal/models"
	"time"

	"github.com/hibiken/asynq"
)

// 创建发送邮件的任务
func NewEmailDeliveryTask(userID string, content string) *asynq.Task {
	payload := models.EmailTaskPayload{
		UserID:  userID,
		Content: content,
	}
	// 通常我们会返回一个唯一的任务ID，方便后续查询，Asynq会自动生成
	return asynq.NewTask(models.TypeEmailDelivery,
		payload.ToBytes(),
		asynq.Queue("default"),
		asynq.MaxRetry(1),
		asynq.Retention(1*time.Hour))
}

// 创建下载项目任务
func NewProjectDownloadTask(projectID, projectPath string) *asynq.Task {
	payload := models.ProjectTaskPayload{
		ProjectID:   projectID,
		ProjectPath: projectPath,
	}
	return asynq.NewTask(models.TypeProjectDownload,
		payload.ToBytes(),
		asynq.Queue("default"),
		asynq.MaxRetry(1),
		asynq.Retention(1*time.Hour))
}

// 创建备份项目任务
func NewProjectBackupTask(projectID, projectPath string) *asynq.Task {
	payload := models.ProjectTaskPayload{
		ProjectID:   projectID,
		ProjectPath: projectPath,
	}

	return asynq.NewTask(models.TypeProjectBackup,
		payload.ToBytes(),
		asynq.Queue("default"),
		asynq.MaxRetry(1),
		asynq.Retention(1*time.Hour))
}

// 创建项目开发任务
func NewProjectDevelopmentTask(projectID, projectPath string) *asynq.Task {
	payload := models.ProjectTaskPayload{
		ProjectID:   projectID,
		ProjectPath: projectPath,
	}
	return asynq.NewTask(models.TypeProjectDevelopment,
		payload.ToBytes(),
		asynq.Queue("default"),
		asynq.MaxRetry(1),
		asynq.Retention(1*time.Hour))
}
