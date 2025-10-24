package tasks

import (
	"time"

	"github.com/lighthought/app-maker/shared-models/agent"
	"github.com/lighthought/app-maker/shared-models/common"
	"github.com/lighthought/app-maker/shared-models/utils"

	"github.com/lighthought/app-maker/shared-models/logger"

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
func NewAgentExecuteTask(projectGUID, agentType, message string, stageName common.DevStatus) *asynq.Task {
	payload := AgentExecuteTaskPayload{
		ProjectGUID: projectGUID,
		AgentType:   agentType,
		Message:     message,
		DevStage:    stageName,
	}
	return asynq.NewTask(common.TaskTypeAgentExecute,
		payload.ToBytes(),
		asynq.Queue(taskQueueDefault),
		asynq.MaxRetry(taskMaxRetry),
		asynq.Retention(taskRetentionHour))
}

// 创建带CLI工具的代理执行任务
func NewAgentExecuteTaskWithCli(projectGUID, agentType, message, cliTool string, stageName common.DevStatus) *asynq.Task {
	payload := AgentExecuteTaskPayload{
		ProjectGUID: projectGUID,
		AgentType:   agentType,
		Message:     message,
		DevStage:    stageName,
		CliTool:     cliTool,
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

// 创建与 Agent 对话任务
func NewAgentChatTask(req *agent.ChatReq) *asynq.Task {
	return asynq.NewTask(common.TaskTypeAgentChat,
		req.ToBytes(),
		asynq.Queue(taskQueueDefault),
		asynq.MaxRetry(taskMaxRetry),
		asynq.Retention(taskRetentionHour))
}

// 创建项目开发阶段任务构造函数
func NewProjectStageTask(requireConfirm bool, projectGuid, stageName string) *asynq.Task {
	payload := ProjectStageTaskPayload{
		ProjectGuid:    projectGuid,
		StageName:      stageName,
		RequireConfirm: requireConfirm,
	}
	return asynq.NewTask(common.TaskTypeProjectStage,
		payload.ToBytes(),
		asynq.Queue(taskQueueDefault),
		asynq.MaxRetry(taskMaxRetry),
		asynq.Retention(taskRetentionHour))
}

// 创建 Agent 任务状态消息任务
func NewAgentTaskResponseTask(message *agent.AgentTaskStatusMessage) *asynq.Task {
	return asynq.NewTask(common.TaskTypeAgentTaskResponse,
		message.ToBytes(),
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
	logger.Info("更新任务进度",
		logger.String("taskID", resultWriter.TaskID()),
		logger.String("status", status),
		logger.Int("progress", progress),
		logger.String("message", message))
}
