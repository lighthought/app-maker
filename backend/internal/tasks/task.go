package tasks

import (
	"autocodeweb-backend/internal/constants"
	"autocodeweb-backend/internal/models"
	"time"

	"github.com/hibiken/asynq"
)

const (
	taskQueueDefault  = "default"
	taskMaxRetry      = 1
	taskRetentionHour = 4 * time.Hour
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
		asynq.Queue(taskQueueDefault),
		asynq.MaxRetry(taskMaxRetry),
		asynq.Retention(taskRetentionHour))
}

// 创建下载项目任务
func NewProjectDownloadTask(projectID, projectGuid, projectPath string) *asynq.Task {
	payload := models.ProjectTaskPayload{
		ProjectID:   projectID,
		ProjectGuid: projectGuid,
		ProjectPath: projectPath,
	}
	return asynq.NewTask(models.TypeProjectDownload,
		payload.ToBytes(),
		asynq.Queue(taskQueueDefault),
		asynq.MaxRetry(taskMaxRetry),
		asynq.Retention(taskRetentionHour))
}

// 创建备份项目任务
func NewProjectBackupTask(projectID, projectGuid, projectPath string) *asynq.Task {
	payload := models.ProjectTaskPayload{
		ProjectID:   projectID,
		ProjectGuid: projectGuid,
		ProjectPath: projectPath,
	}

	return asynq.NewTask(models.TypeProjectBackup,
		payload.ToBytes(),
		asynq.Queue(taskQueueDefault),
		asynq.MaxRetry(taskMaxRetry),
		asynq.Retention(taskRetentionHour))
}

// 创建项目开发任务
func NewProjectDevelopmentTask(projectID, projectGuid, gitlabRepoURL string) *asynq.Task {
	payload := models.ProjectTaskPayload{
		ProjectID:   projectID,
		ProjectGuid: projectGuid,
		ProjectPath: gitlabRepoURL,
	}
	return asynq.NewTask(models.TypeProjectDevelopment,
		payload.ToBytes(),
		asynq.Queue(taskQueueDefault),
		asynq.MaxRetry(taskMaxRetry),
		asynq.Retention(taskRetentionHour))
}

// 创建项目初始化任务
func NewProjectInitTask(projectID, projectGuid, projectPath string) *asynq.Task {
	payload := models.ProjectTaskPayload{
		ProjectID:   projectID,
		ProjectGuid: projectGuid,
		ProjectPath: projectPath,
	}
	return asynq.NewTask(models.TypeProjectInit,
		payload.ToBytes(),
		asynq.Queue(taskQueueDefault),
		asynq.MaxRetry(taskMaxRetry),
		asynq.Retention(taskRetentionHour))
}

// 创建WebSocket消息广播任务
func NewWebSocketBroadcastTask(projectGUID, messageType, targetID string) *asynq.Task {
	payload := models.WebSocketTaskPayload{
		ProjectGUID: projectGUID,
		MessageType: messageType,
	}

	switch messageType {
	case constants.WebSocketMessageTypeProjectMessage:
		payload.MessageID = targetID
	case constants.WebSocketMessageTypeProjectStageUpdate:
		payload.StageID = targetID
	case constants.WebSocketMessageTypeProjectInfoUpdate:
		payload.ProjectID = targetID
	}

	return asynq.NewTask(models.TypeWebSocketBroadcast,
		payload.ToBytes(),
		asynq.Queue(taskQueueDefault),
		asynq.MaxRetry(taskMaxRetry),
		asynq.Retention(taskRetentionHour))
}
