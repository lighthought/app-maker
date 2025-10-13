package tasks

import (
	"shared-models/agent"
	"shared-models/common"
	"shared-models/utils"
	"time"

	"shared-models/logger"

	"github.com/hibiken/asynq"
)

const (
	taskQueueDefault  = "default"
	taskMaxRetry      = 1
	taskRetentionHour = 4 * time.Hour
)

// 创建下载项目任务
func NewProjectDownloadTask(projectID, projectGuid, projectPath string) *asynq.Task {
	payload := ProjectTaskPayload{
		ProjectID:   projectID,
		ProjectGuid: projectGuid,
		ProjectPath: projectPath,
	}
	return asynq.NewTask(common.TaskTypeProjectDownload,
		payload.ToBytes(),
		asynq.Queue(taskQueueDefault),
		asynq.MaxRetry(taskMaxRetry),
		asynq.Retention(taskRetentionHour))
}

// 创建备份项目任务
func NewProjectBackupTask(projectID, projectGuid, projectPath string) *asynq.Task {
	payload := ProjectTaskPayload{
		ProjectID:   projectID,
		ProjectGuid: projectGuid,
		ProjectPath: projectPath,
	}

	return asynq.NewTask(common.TaskTypeProjectBackup,
		payload.ToBytes(),
		asynq.Queue(taskQueueDefault),
		asynq.MaxRetry(taskMaxRetry),
		asynq.Retention(taskRetentionHour))
}

// 创建项目开发任务
func NewProjectDevelopmentTask(projectID, projectGuid, gitlabRepoURL string) *asynq.Task {
	payload := ProjectTaskPayload{
		ProjectID:   projectID,
		ProjectGuid: projectGuid,
		ProjectPath: gitlabRepoURL,
	}
	return asynq.NewTask(common.TaskTypeProjectDevelopment,
		payload.ToBytes(),
		asynq.Queue(taskQueueDefault),
		asynq.MaxRetry(taskMaxRetry),
		asynq.Retention(taskRetentionHour))
}

// 创建项目初始化任务
func NewProjectInitTask(projectID, projectGuid, projectPath string) *asynq.Task {
	payload := ProjectTaskPayload{
		ProjectID:   projectID,
		ProjectGuid: projectGuid,
		ProjectPath: projectPath,
	}
	return asynq.NewTask(common.TaskTypeProjectInit,
		payload.ToBytes(),
		asynq.Queue(taskQueueDefault),
		asynq.MaxRetry(taskMaxRetry),
		asynq.Retention(taskRetentionHour))
}

// 创建WebSocket消息广播任务
func NewWebSocketBroadcastTask(projectGUID, messageType, targetID string) *asynq.Task {
	payload := WebSocketTaskPayload{
		ProjectGUID: projectGUID,
		MessageType: messageType,
	}

	switch messageType {
	case common.WebSocketMessageTypeProjectMessage:
		payload.MessageID = targetID
	case common.WebSocketMessageTypeProjectStageUpdate:
		payload.StageID = targetID
	case common.WebSocketMessageTypeProjectInfoUpdate:
		payload.ProjectID = targetID
	}

	return asynq.NewTask(common.TaskTypeWebSocketBroadcast,
		payload.ToBytes(),
		asynq.Queue(taskQueueDefault),
		asynq.MaxRetry(taskMaxRetry),
		asynq.Retention(taskRetentionHour))
}

// 创建代理执行任务
func NewAgentExecuteTask(projectGUID, agentType, message string) *asynq.Task {
	payload := AgentExecuteTaskPayload{
		ProjectGUID: projectGUID,
		AgentType:   agentType,
		Message:     message,
	}
	return asynq.NewTask(common.TaskTypeAgentExecute,
		payload.ToBytes(),
		asynq.Queue(taskQueueDefault),
		asynq.MaxRetry(taskMaxRetry),
		asynq.Retention(taskRetentionHour))
}

// 创建项目环境准备任务
func NewProjectSetupTask(req *agent.SetupProjEnvReq) *asynq.Task {
	return asynq.NewTask(common.TaskTypeAgentSetup,
		req.ToBytes(),
		asynq.Queue(taskQueueDefault),
		asynq.MaxRetry(taskMaxRetry),
		asynq.Retention(taskRetentionHour))
}

// 创建部署项目任务
func NewProjectDeployTask(req *agent.DeployReq) *asynq.Task {
	return asynq.NewTask(common.TaskTypeProjectDeploy,
		req.ToBytes(),
		asynq.Queue(taskQueueDefault),
		asynq.MaxRetry(taskMaxRetry),
		asynq.Retention(taskRetentionHour))
}

// updateResult 是一个帮助函数，用于将任务进度更新到Redis。
// 这里假设使用一个Redis Hash结构，key为`task:progress:<task_id>`。
func UpdateResult(resultWriter *asynq.ResultWriter, status string, progress int, message string) {
	if resultWriter == nil {
		logger.Error("resultWriter is nil, can't update result")
		return
	}

	data := TaskResult{
		TaskID:    resultWriter.TaskID(),
		Status:    status,
		Progress:  progress,
		Message:   message,
		UpdatedAt: utils.GetCurrentTime(),
	}
	resultWriter.Write(data.ToBytes())
}
