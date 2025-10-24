package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/lighthought/app-maker/shared-models/agent"
	"github.com/lighthought/app-maker/shared-models/common"
	"github.com/lighthought/app-maker/shared-models/logger"
	"github.com/lighthought/app-maker/shared-models/tasks"
	"github.com/lighthought/app-maker/shared-models/utils"

	"github.com/lighthought/app-maker/agents/internal/api/models"

	"github.com/hibiken/asynq"
)

type AgentTaskService interface {
	// 处理任务
	ProcessTask(ctx context.Context, task *asynq.Task) error
	// Agent 执行任务（带CLI工具）
	EnqueueWithCli(projectGuid, agentType, message, cliTool string, stageName common.DevStatus) (*asynq.TaskInfo, error)
	// 项目环境准备
	EnqueueSetupReq(req *agent.SetupProjEnvReq) (*asynq.TaskInfo, error)
	// 部署项目
	EnqueueDeployReq(req *agent.DeployReq) (*asynq.TaskInfo, error)
	// 与 Agent 对话
	EnqueueChatWithAgent(req *agent.ChatReq) (*asynq.TaskInfo, error)
	// 与指定代理对话
	ChatWithAgent(ctx context.Context, projectGuid, agentType, message string) (*models.CommandResult, error)
}

type agentTaskService struct {
	commandService CommandService
	fileService    FileService
	gitService     GitService
	redisService   RedisService
	asyncClient    *asynq.Client
}

const (
	ASYNC_IS_NIL         = "async client is nil"
	UNEXPECTED_TASK_TYPE = "unexpected task type "
)

func NewAgentTaskService(commandService CommandService,
	fileService FileService,
	gitService GitService,
	redisService RedisService,
	asyncClient *asynq.Client) AgentTaskService {
	return &agentTaskService{
		commandService: commandService,
		fileService:    fileService,
		gitService:     gitService,
		asyncClient:    asyncClient,
		redisService:   redisService,
	}
}

// Enqueue 创建代理执行任务
func (h *agentTaskService) Enqueue(projectGuid, agentType, message string, stageName common.DevStatus) (*asynq.TaskInfo, error) {
	if h.asyncClient == nil {
		return nil, fmt.Errorf("%s", ASYNC_IS_NIL)
	}
	return h.asyncClient.Enqueue(tasks.NewAgentExecuteTask(projectGuid, agentType, message, stageName))
}

// EnqueueWithCli 创建带CLI工具的代理执行任务
func (h *agentTaskService) EnqueueWithCli(projectGuid, agentType, message, cliTool string, stageName common.DevStatus) (*asynq.TaskInfo, error) {
	if h.asyncClient == nil {
		return nil, fmt.Errorf("%s", ASYNC_IS_NIL)
	}
	return h.asyncClient.Enqueue(tasks.NewAgentExecuteTaskWithCli(projectGuid, agentType, message, cliTool, stageName))
}

// EnqueueReq 创建项目环境准备任务
func (h *agentTaskService) EnqueueSetupReq(req *agent.SetupProjEnvReq) (*asynq.TaskInfo, error) {
	if h.asyncClient == nil {
		return nil, fmt.Errorf("%s", ASYNC_IS_NIL)
	}
	if req == nil {
		return nil, fmt.Errorf("EnqueueSetupReq, req is nil")
	}
	return h.asyncClient.Enqueue(tasks.NewProjectSetupTask(req))
}

// 部署项目
func (h *agentTaskService) EnqueueDeployReq(req *agent.DeployReq) (*asynq.TaskInfo, error) {
	if h.asyncClient == nil {
		return nil, fmt.Errorf("%s", ASYNC_IS_NIL)
	}
	if req == nil {
		return nil, fmt.Errorf("EnqueueDeployReq, req is nil")
	}
	return h.asyncClient.Enqueue(tasks.NewProjectDeployTask(req))
}

// 与 Agent 对话
func (h *agentTaskService) EnqueueChatWithAgent(req *agent.ChatReq) (*asynq.TaskInfo, error) {
	if h.asyncClient == nil {
		return nil, fmt.Errorf("%s", ASYNC_IS_NIL)
	}
	if req == nil {
		return nil, fmt.Errorf("EnqueueDeployReq, req is nil")
	}
	return h.asyncClient.Enqueue(tasks.NewAgentChatTask(req))
}

// ProcessTask 处理代理执行任务
func (h *agentTaskService) ProcessTask(ctx context.Context, task *asynq.Task) error {
	switch task.Type() {
	case common.TaskTypeAgentExecute:
		return h.innerProcessAgentExecuteTask(ctx, task)
	case common.TaskTypeAgentChat:
		return h.innerProcessAgentChatTask(ctx, task)
	default:
		return fmt.Errorf("%s%s", UNEXPECTED_TASK_TYPE, task.Type())
	}
}

// 处理代理执行任务
func (s *agentTaskService) innerProcessAgentExecuteTask(ctx context.Context, task *asynq.Task) error {
	if task.Type() != common.TaskTypeAgentExecute {
		return fmt.Errorf("%s%s", UNEXPECTED_TASK_TYPE, task.Type())
	}

	payload := tasks.AgentExecuteTaskPayload{}
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}
	tasks.UpdateResult(task.ResultWriter(), common.CommonStatusInProgress, 5, "正在执行代理任务...")
	s.redisService.PublishTaskStatus(&payload, task.ResultWriter().TaskID(), common.CommonStatusInProgress, "正在执行代理任务...")

	_, err := s.innerProcessTask(ctx, payload, task)
	if err != nil {
		return err
	}
	return nil
}

// 处理与 Agent 对话任务
func (h *agentTaskService) innerProcessAgentChatTask(ctx context.Context, task *asynq.Task) error {
	if task.Type() != common.TaskTypeAgentChat {
		return fmt.Errorf("%s%s", UNEXPECTED_TASK_TYPE, task.Type())
	}

	var req agent.ChatReq
	if err := json.Unmarshal(task.Payload(), &req); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	payload := tasks.AgentExecuteTaskPayload{
		ProjectGUID: req.ProjectGuid,
		AgentType:   req.AgentType,
		Message:     req.Message,
		DevStage:    common.DevStatus(req.DevStage),
		CliTool:     req.CliTool,
	}
	_, err := h.innerProcessTask(ctx, payload, task)
	if err != nil {
		return err
	}
	return nil
}

// 构建 CLI 命令
func (h *agentTaskService) buildCliCommand(cliTool, sessionID, message string) (string, []string, bool) {
	var cliCommand string
	var args []string
	var useJsonOutput bool

	switch cliTool {
	case common.CliToolQwenCode:
		cliCommand = "qwen"
		useJsonOutput = false
		args = []string{"-y", "-p", "\"" + message + "\""}

	case common.CliToolGemini:
		cliCommand = "gemini"
		useJsonOutput = false
		args = []string{"-y", "-p", "\"" + message + "\""}

	default:
		cliCommand = "claude"
		useJsonOutput = true
		if sessionID == "" {
			args = []string{"--dangerously-skip-permissions", "--output-format", "json", "-p", "\"" + message + "\""}
		} else {
			args = []string{"--dangerously-skip-permissions", "--resume", sessionID, "--output-format", "json", "-p", "\"" + message + "\""}
		}
	}

	return cliCommand, args, useJsonOutput
}

// 处理情况
func (h *agentTaskService) handleAgentExecuteFailed(task *asynq.Task, payload tasks.AgentExecuteTaskPayload, result models.CommandResult) {
	if task != nil {
		tasks.UpdateResult(task.ResultWriter(), common.CommonStatusFailed, 0, result.Error)
		// 发布任务失败状态
		h.redisService.PublishTaskStatus(&payload, task.ResultWriter().TaskID(), common.CommonStatusFailed, result.Error)
		logger.Error("代理任务执行失败",
			logger.String("taskID", task.ResultWriter().TaskID()),
			logger.String("agentType", payload.AgentType),
			logger.String("message", payload.Message),
			logger.String("error", result.Error))
	} else {
		logger.Error("代理任务执行失败",
			logger.String("agentType", payload.AgentType),
			logger.String("message", payload.Message),
			logger.String("error", result.Error))
	}
}

// 异步任务本身、对话方法公用这个内部方法
func (h *agentTaskService) innerProcessTask(ctx context.Context, payload tasks.AgentExecuteTaskPayload, task *asynq.Task) (*models.CommandResult, error) {
	var result models.CommandResult
	timeBefor := utils.GetTimeNow()
	sessionID := h.redisService.GetSessionByProjectGuid(payload.ProjectGUID, payload.AgentType)

	logger.Info("\n===> 开始执行代理任务",
		logger.String("startTime", utils.GetCurrentTime()),
		logger.String("projectGUID", payload.ProjectGUID),
		logger.String("agentType", payload.AgentType),
		logger.String("message", payload.Message))

	// 发布任务开始状态
	if task != nil {
		h.redisService.PublishTaskStatus(&payload, task.ResultWriter().TaskID(), common.CommonStatusInProgress, "任务开始执行")
	}

	// 从 payload 或项目检测获取 CLI 类型
	cliTool := payload.CliTool
	if cliTool == "" {
		cliTool = h.fileService.DetectCliTool(payload.ProjectGUID)
	}

	// 根据 CLI 类型构建命令
	cliCommand, args, useJsonOutput := h.buildCliCommand(cliTool, sessionID, payload.Message)
	result = h.commandService.SimpleExecute(ctx, payload.ProjectGUID, cliCommand, args...)

	logger.Info("\n===> 代理任务执行完成",
		logger.String("endTime", utils.GetCurrentTime()),
		logger.String("projectGUID", payload.ProjectGUID),
		logger.String("agentType", payload.AgentType),
		logger.String("message", payload.Message))

	timeAfter := utils.GetTimeNow()
	duration := timeAfter.Sub(timeBefor)
	durationMs := int(duration.Milliseconds())
	if !result.Success {
		h.handleAgentExecuteFailed(task, payload, result)
		return nil, fmt.Errorf("agent execute task failed: %s", result.Error)
	}

	claudeResponse := models.ClaudeResponse{
		Type:          "result",
		Subtype:       "success",
		DurationMs:    durationMs,
		DurationApiMs: durationMs,
		IsError:       result.Error != "",
		Result:        result.Output,
	}

	// 根据输出格式处理结果
	if useJsonOutput {
		// 处理 JSON 输出（Claude）
		if err := json.Unmarshal([]byte(result.Output), &claudeResponse); err != nil {
			logger.Error(" ===> CLI 结果解析失败",
				logger.String("agentType", payload.AgentType),
				logger.String("message", payload.Message),
				logger.String("error", result.Error))
		} else {
			// 转换成功了，就可以直接取执行的结果了，去掉外层的 json 包装
			result.Output = claudeResponse.Result
		}
	} else {
		// 处理纯文本输出（qwen、gemini）
		// 直接使用原始输出文本
		claudeResponse.Result = result.Output
	}

	if claudeResponse.IsError {
		h.handleAgentExecuteFailed(task, payload, result)
		return nil, fmt.Errorf("agent execute task, claude failed: %s", claudeResponse.Result)
	}

	// 保存会话ID
	if claudeResponse.SessionID != "" {
		h.redisService.SaveSessionByProjectGuid(payload.ProjectGUID, payload.AgentType, claudeResponse.SessionID)
	}
	logger.Info(" ===> 代理任务执行成功",
		logger.String("agentType", payload.AgentType),
		logger.String("message", payload.Message),
		logger.String("claudeResponse", claudeResponse.ToJsonString()))

	err := h.gitService.CommitAndPush(ctx, payload.ProjectGUID, claudeResponse.Result)
	if err != nil {
		logger.Error("项目文档、代码提交并推送失败",
			logger.String("GUID", payload.ProjectGUID),
			logger.String("error", err.Error()))

		if task != nil {
			tasks.UpdateResult(task.ResultWriter(), common.CommonStatusFailed, 0, err.Error())
			// 发布任务失败状态
			h.redisService.PublishTaskStatus(&payload, task.ResultWriter().TaskID(), common.CommonStatusFailed, "项目文档、代码提交并推送失败: "+err.Error())
		}
		return nil, fmt.Errorf("项目文档、代码提交并推送失败: %w", err)
	}

	if task != nil {
		tasks.UpdateResult(task.ResultWriter(), common.CommonStatusDone, 100, claudeResponse.Result)
		// 发布任务完成状态
		h.redisService.PublishTaskStatus(&payload, task.ResultWriter().TaskID(), common.CommonStatusDone, "任务执行完成")
	}
	return &result, nil
}

// 与指定代理对话
func (h *agentTaskService) ChatWithAgent(ctx context.Context, projectGuid, agentType, message string) (*models.CommandResult, error) {
	if h.commandService == nil {
		return nil, fmt.Errorf("command service is nil")
	}

	payload := tasks.AgentExecuteTaskPayload{
		ProjectGUID: projectGuid,
		AgentType:   agentType,
		Message:     message,
		DevStage:    common.DevStatusUnknown, // 阵列用 Unknown 表示聊天
	}
	return h.innerProcessTask(ctx, payload, nil)
}
