package services

import (
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/lighthought/app-maker/shared-models/agent"
	"github.com/lighthought/app-maker/shared-models/tasks"
)

// asynq 异步处理业务接口
type AsyncClientService interface {
	// 创建项目初始化任务
	EnqueueProjectInitTask(projectID, projectGuid, projectPath string) (string, error)
	// 创建项目开发阶段任务
	EnqueueProjectStageTask(needConfirm bool, projectGuid, stageName string) (string, error)
	// 创建 WebSocket 消息广播任务
	EnqueueWebSocketBroadcastTask(projectGUID, messageType, targetId string) (string, error)
	// 与项目中的 Agent 进行对话
	EnqueueAgentChatTask(req *agent.ChatReq) (string, error)
	// 创建 Agent 任务响应任务
	EnqueueAgentTaskResponseTask(message *agent.AgentTaskStatusMessage) (string, error)
	// 创建项目备份任务
	EnqueueProjectBackupTask(projectID, projectGuid, projectPath string) (string, error)
	// 创建项目下载任务
	EnqueueProjectDownloadTask(projectID, projectGuid, projectPath string) (string, error)
	// 创建部署项目任务
	EnqueueProjectDeployTask(projectGuid, environment string) (string, error)
}

// asynq 异步处理业务实现
type asyncClientService struct {
	asyncClient *asynq.Client
}

// NewAsyncService 创建 asynq 异步处理业务
func NewAsyncClientService(asyncClient *asynq.Client) AsyncClientService {
	return &asyncClientService{
		asyncClient: asyncClient,
	}
}

// EnqueueProjectInitTask 创建项目初始化任务
func (s *asyncClientService) EnqueueProjectInitTask(projectID, projectGuid, projectPath string) (string, error) {
	taskInfo, err := s.asyncClient.Enqueue(tasks.NewProjectInitTask(projectID, projectGuid, projectPath))
	if err != nil {
		return "", fmt.Errorf("failed to create project init task: %s", err.Error())
	}
	return taskInfo.ID, nil
}

// EnqueueProjectStageTask 创建项目开发阶段任务
func (s *asyncClientService) EnqueueProjectStageTask(needConfirm bool, projectGuid, stageName string) (string, error) {
	taskInfo, err := s.asyncClient.Enqueue(tasks.NewProjectStageTask(needConfirm, projectGuid, stageName))
	if err != nil {
		return "", fmt.Errorf("failed to create project stage task: %s", err.Error())
	}
	return taskInfo.ID, nil
}

// EnqueueWebSocketBroadcastTask 创建 WebSocket 消息广播任务
func (s *asyncClientService) EnqueueWebSocketBroadcastTask(projectGUID, messageType, targetId string) (string, error) {
	taskInfo, err := s.asyncClient.Enqueue(
		tasks.NewWebSocketBroadcastTask(projectGUID, messageType, targetId),
	)
	if err != nil {
		return "", fmt.Errorf("failed to create web socket broadcast task: %s", err.Error())
	}
	return taskInfo.ID, nil
}

// EnqueueAgentChatTask 创建与项目中的 Agent 进行对话任务
func (s *asyncClientService) EnqueueAgentChatTask(req *agent.ChatReq) (string, error) {
	// 异步方式
	taskInfo, err := s.asyncClient.Enqueue(tasks.NewAgentChatTask(req))
	if err != nil {
		return "", fmt.Errorf("创建与 Agent 对话任务失败: %w", err)
	}
	return taskInfo.ID, nil
}

// EnqueueAgentTaskResponseTask 创建 Agent 任务响应任务
func (s *asyncClientService) EnqueueAgentTaskResponseTask(message *agent.AgentTaskStatusMessage) (string, error) {
	taskInfo, err := s.asyncClient.Enqueue(tasks.NewAgentTaskResponseTask(message))
	if err != nil {
		return "", fmt.Errorf("failed to create agent task response task: %s", err.Error())
	}
	return taskInfo.ID, nil
}

// EnqueueProjectBackupTask 创建项目备份任务
func (s *asyncClientService) EnqueueProjectBackupTask(projectID, projectGuid, projectPath string) (string, error) {
	taskInfo, err := s.asyncClient.Enqueue(tasks.NewProjectBackupTask(projectID, projectGuid, projectPath))
	if err != nil {
		return "", fmt.Errorf("failed to create project backup task: %s", err.Error())
	}
	return taskInfo.ID, nil
}

// EnqueueProjectDownloadTask 创建项目下载任务
func (s *asyncClientService) EnqueueProjectDownloadTask(projectID, projectGuid, projectPath string) (string, error) {
	taskInfo, err := s.asyncClient.Enqueue(tasks.NewProjectDownloadTask(projectID, projectGuid, projectPath))
	if err != nil {
		return "", fmt.Errorf("failed to create project download task: %s", err.Error())
	}
	return taskInfo.ID, nil
}

// EnqueueProjectDeployTask 创建部署项目任务
func (s *asyncClientService) EnqueueProjectDeployTask(projectGuid, environment string) (string, error) {
	req := &agent.DeployReq{
		ProjectGuid: projectGuid,
		Environment: environment,
	}
	// 异步方法，返回任务 ID
	info, err := s.asyncClient.Enqueue(tasks.NewProjectDeployTask(req))
	if err != nil {
		return "", fmt.Errorf("failed to create deploy project task: %s", err.Error())
	}

	return info.ID, nil
}
