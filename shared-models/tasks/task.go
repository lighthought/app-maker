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

// 创建带CLI工具的代理执行任务
func NewAgentExecuteTaskWithCli(projectGUID, agentType, message, cliTool string) *asynq.Task {
	payload := AgentExecuteTaskPayload{
		ProjectGUID: projectGUID,
		AgentType:   agentType,
		Message:     message,
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

// 创建检查需求任务
func NewCheckRequirementTask(projectID, projectGuid string) *asynq.Task {
	payload := ProjectTaskPayload{
		ProjectID:   projectID,
		ProjectGuid: projectGuid,
	}
	return asynq.NewTask(common.TaskTypeStageCheckRequirement,
		payload.ToBytes(),
		asynq.Queue(taskQueueDefault),
		asynq.MaxRetry(taskMaxRetry),
		asynq.Retention(taskRetentionHour))
}

// 创建生成PRD任务
func NewGeneratePRDTask(projectID, projectGuid string) *asynq.Task {
	payload := ProjectTaskPayload{
		ProjectID:   projectID,
		ProjectGuid: projectGuid,
	}
	return asynq.NewTask(common.TaskTypeStageGeneratePRD,
		payload.ToBytes(),
		asynq.Queue(taskQueueDefault),
		asynq.MaxRetry(taskMaxRetry),
		asynq.Retention(taskRetentionHour))
}

// 创建定义UX标准任务
func NewDefineUXStandardTask(projectID, projectGuid string) *asynq.Task {
	payload := ProjectTaskPayload{
		ProjectID:   projectID,
		ProjectGuid: projectGuid,
	}
	return asynq.NewTask(common.TaskTypeStageDefineUXStandard,
		payload.ToBytes(),
		asynq.Queue(taskQueueDefault),
		asynq.MaxRetry(taskMaxRetry),
		asynq.Retention(taskRetentionHour))
}

// 创建设计架构任务
func NewDesignArchitectureTask(projectID, projectGuid string) *asynq.Task {
	payload := ProjectTaskPayload{
		ProjectID:   projectID,
		ProjectGuid: projectGuid,
	}
	return asynq.NewTask(common.TaskTypeStageDesignArchitecture,
		payload.ToBytes(),
		asynq.Queue(taskQueueDefault),
		asynq.MaxRetry(taskMaxRetry),
		asynq.Retention(taskRetentionHour))
}

// 创建划分Epic和Story任务
func NewPlanEpicAndStoryTask(projectID, projectGuid string) *asynq.Task {
	payload := ProjectTaskPayload{
		ProjectID:   projectID,
		ProjectGuid: projectGuid,
	}
	return asynq.NewTask(common.TaskTypeStagePlanEpicAndStory,
		payload.ToBytes(),
		asynq.Queue(taskQueueDefault),
		asynq.MaxRetry(taskMaxRetry),
		asynq.Retention(taskRetentionHour))
}

// 创建定义数据模型任务
func NewDefineDataModelTask(projectID, projectGuid string) *asynq.Task {
	payload := ProjectTaskPayload{
		ProjectID:   projectID,
		ProjectGuid: projectGuid,
	}
	return asynq.NewTask(common.TaskTypeStageDefineDataModel,
		payload.ToBytes(),
		asynq.Queue(taskQueueDefault),
		asynq.MaxRetry(taskMaxRetry),
		asynq.Retention(taskRetentionHour))
}

// 创建定义API任务
func NewDefineAPITask(projectID, projectGuid string) *asynq.Task {
	payload := ProjectTaskPayload{
		ProjectID:   projectID,
		ProjectGuid: projectGuid,
	}
	return asynq.NewTask(common.TaskTypeStageDefineAPI,
		payload.ToBytes(),
		asynq.Queue(taskQueueDefault),
		asynq.MaxRetry(taskMaxRetry),
		asynq.Retention(taskRetentionHour))
}

// 创建生成前端页面任务
func NewGeneratePagesTask(projectID, projectGuid string) *asynq.Task {
	payload := ProjectTaskPayload{
		ProjectID:   projectID,
		ProjectGuid: projectGuid,
	}
	return asynq.NewTask(common.TaskTypeStageGeneratePages,
		payload.ToBytes(),
		asynq.Queue(taskQueueDefault),
		asynq.MaxRetry(taskMaxRetry),
		asynq.Retention(taskRetentionHour))
}

// 创建开发Story任务
func NewDevelopStoryTask(projectID, projectGuid string) *asynq.Task {
	payload := ProjectTaskPayload{
		ProjectID:   projectID,
		ProjectGuid: projectGuid,
	}
	return asynq.NewTask(common.TaskTypeStageDevelopStory,
		payload.ToBytes(),
		asynq.Queue(taskQueueDefault),
		asynq.MaxRetry(taskMaxRetry),
		asynq.Retention(taskRetentionHour))
}

// 创建运行测试任务
func NewRunTestTask(projectID, projectGuid string) *asynq.Task {
	payload := ProjectTaskPayload{
		ProjectID:   projectID,
		ProjectGuid: projectGuid,
	}
	return asynq.NewTask(common.TaskTypeStageRunTest,
		payload.ToBytes(),
		asynq.Queue(taskQueueDefault),
		asynq.MaxRetry(taskMaxRetry),
		asynq.Retention(taskRetentionHour))
}

// 创建部署任务
func NewDeployTask(projectID, projectGuid string) *asynq.Task {
	payload := ProjectTaskPayload{
		ProjectID:   projectID,
		ProjectGuid: projectGuid,
	}
	return asynq.NewTask(common.TaskTypeStageDeploy,
		payload.ToBytes(),
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
