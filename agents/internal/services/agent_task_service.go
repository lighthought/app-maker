package services

import (
	"context"
	"encoding/json"
	"fmt"
	"shared-models/agent"
	"shared-models/common"
	"shared-models/logger"
	"shared-models/tasks"

	"github.com/hibiken/asynq"
)

type AgentTaskService interface {
	ProcessTask(ctx context.Context, task *asynq.Task) error
	Enqueue(projectGuid, agentType, message string) (*asynq.TaskInfo, error)
	EnqueueReq(req *agent.SetupProjEnvReq) (*asynq.TaskInfo, error)
}

type agentTaskService struct {
	commandService CommandService
	gitService     GitService
	asyncClient    *asynq.Client
}

func NewAgentTaskService(commandService CommandService, gitService GitService, asyncClient *asynq.Client) AgentTaskService {
	return &agentTaskService{commandService: commandService, gitService: gitService, asyncClient: asyncClient}
}

// Enqueue 创建代理执行任务
func (h *agentTaskService) Enqueue(projectGuid, agentType, message string) (*asynq.TaskInfo, error) {
	if h.asyncClient == nil {
		return nil, fmt.Errorf("async client is nil")
	}
	return h.asyncClient.Enqueue(tasks.NewAgentExecuteTask(projectGuid, agentType, message))
}

// EnqueueReq 创建项目环境准备任务
func (h *agentTaskService) EnqueueReq(req *agent.SetupProjEnvReq) (*asynq.TaskInfo, error) {
	if h.asyncClient == nil {
		return nil, fmt.Errorf("async client is nil")
	}
	return h.asyncClient.Enqueue(tasks.NewProjectSetupTask(req))
}

// ProcessTask 处理代理执行任务
func (h *agentTaskService) ProcessTask(ctx context.Context, task *asynq.Task) error {
	if task.Type() != common.TaskTypeAgentExecute {
		return fmt.Errorf("unexpected task type %s", task.Type())
	}

	payload := tasks.AgentExecuteTaskPayload{}
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}
	tasks.UpdateResult(task.ResultWriter(), common.CommonStatusInProgress, 5, "正在执行代理任务...")

	result := h.commandService.SimpleExecute(ctx, payload.ProjectGUID, "claude", "--dangerously-skip-permissions", "-p", payload.Message)

	if !result.Success {
		tasks.UpdateResult(task.ResultWriter(), common.CommonStatusFailed, 0, result.Error)
		return fmt.Errorf("agent execute task failed: %s", result.Error)
	}

	logger.Info("代理任务执行成功", logger.String("taskID", task.ResultWriter().TaskID()))

	err := h.gitService.CommitAndPush(ctx, payload.ProjectGUID, result.Output)
	if err != nil {
		logger.Error("项目文档、代码提交并推送失败", logger.String("GUID", payload.ProjectGUID), logger.String("error", err.Error()))
		tasks.UpdateResult(task.ResultWriter(), common.CommonStatusFailed, 0, err.Error())
		return fmt.Errorf("项目文档、代码提交并推送失败: %w", err)
	}

	tasks.UpdateResult(task.ResultWriter(), common.CommonStatusDone, 100, result.Output)
	return nil
}
