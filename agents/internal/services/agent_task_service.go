package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/lighthought/app-maker/shared-models/agent"
	"github.com/lighthought/app-maker/shared-models/cache"
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
	// Agent 执行任务
	//Enqueue(projectGuid, agentType, message, stageName string) (*asynq.TaskInfo, error)
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
	asyncClient    *asynq.Client
	cacheInstance  cache.Cache
}

const (
	ASYNC_IS_NIL         = "async client is nil"
	UNEXPECTED_TASK_TYPE = "unexpected task type "
)

func NewAgentTaskService(commandService CommandService,
	fileService FileService,
	gitService GitService,
	asyncClient *asynq.Client,
	cacheInstance cache.Cache) AgentTaskService {
	return &agentTaskService{
		commandService: commandService,
		fileService:    fileService,
		gitService:     gitService,
		asyncClient:    asyncClient,
		cacheInstance:  cacheInstance,
	}
}

// 根据projectGuid从缓存中获取 sessionId
func (h *agentTaskService) getSessionByProjectGuid(projectGuid, agentType string) string {
	if h.cacheInstance == nil {
		logger.Warn("Cache instance is nil, cannot get session", logger.String("projectGuid", projectGuid))
		return ""
	}

	key := cache.GetProjectAgentSessionCacheKey(projectGuid, agentType)
	var sessionID string
	err := h.cacheInstance.Get(key, &sessionID)
	if err != nil {
		logger.Error("Failed to get session from Redis",
			logger.String("projectGuid", projectGuid),
			logger.String("error", err.Error()))
		return ""
	}

	logger.Info("Retrieved session for project",
		logger.String("projectGuid", projectGuid),
		logger.String("sessionID", sessionID))
	return sessionID
}

// 把会话ID保存到缓存中
func (h *agentTaskService) saveSessionByProjectGuid(projectGuid, agentType, sessionID string) {
	if h.cacheInstance == nil {
		logger.Warn("Redis client is nil, cannot save session",
			logger.String("projectGuid", projectGuid),
			logger.String("sessionID", sessionID))
		return
	}

	if projectGuid == "" || sessionID == "" {
		logger.Warn("Invalid parameters for saving session",
			logger.String("projectGuid", projectGuid),
			logger.String("sessionID", sessionID))
		return
	}

	key := cache.GetProjectAgentSessionCacheKey(projectGuid, agentType)
	// 设置过期时间为 24 小时，避免会话数据永久占用内存
	expiration := common.CacheExpirationDay

	err := h.cacheInstance.Set(key, sessionID, expiration)
	if err != nil {
		logger.Error("Failed to save session to cache",
			logger.String("projectGuid", projectGuid),
			logger.String("sessionID", sessionID),
			logger.String("error", err.Error()),
		)
		return
	}

	logger.Info("Saved session for project",
		logger.String("projectGuid", projectGuid),
		logger.String("sessionID", sessionID),
		logger.String("expiration", expiration.String()),
	)
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
func (h *agentTaskService) innerProcessAgentExecuteTask(ctx context.Context, task *asynq.Task) error {
	if task.Type() != common.TaskTypeAgentExecute {
		return fmt.Errorf("%s%s", UNEXPECTED_TASK_TYPE, task.Type())
	}

	payload := tasks.AgentExecuteTaskPayload{}
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}
	tasks.UpdateResult(task.ResultWriter(), common.CommonStatusInProgress, 5, "正在执行代理任务...")

	_, err := h.innerProcessTask(ctx, payload, task)
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
		h.publishTaskStatus(task.ResultWriter().TaskID(), payload.ProjectGUID, payload.AgentType,
			common.CommonStatusFailed, result.Error, payload.DevStage)
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
	sessionID := h.getSessionByProjectGuid(payload.ProjectGUID, payload.AgentType)

	logger.Info("\n===> 开始执行代理任务",
		logger.String("startTime", utils.GetCurrentTime()),
		logger.String("projectGUID", payload.ProjectGUID),
		logger.String("agentType", payload.AgentType),
		logger.String("message", payload.Message))

	// 发布任务开始状态
	if task != nil {
		h.publishTaskStatus(task.ResultWriter().TaskID(), payload.ProjectGUID, payload.AgentType,
			common.CommonStatusInProgress, "任务开始执行", payload.DevStage)
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
		h.saveSessionByProjectGuid(payload.ProjectGUID, payload.AgentType, claudeResponse.SessionID)
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
			h.publishTaskStatus(task.ResultWriter().TaskID(), payload.ProjectGUID, payload.AgentType,
				common.CommonStatusFailed, "项目文档、代码提交并推送失败: "+err.Error(), payload.DevStage)
		}
		return nil, fmt.Errorf("项目文档、代码提交并推送失败: %w", err)
	}

	if task != nil {
		tasks.UpdateResult(task.ResultWriter(), common.CommonStatusDone, 100, claudeResponse.Result)
		// 发布任务完成状态
		h.publishTaskStatus(task.ResultWriter().TaskID(), payload.ProjectGUID, payload.AgentType,
			common.CommonStatusDone, "任务执行完成", payload.DevStage)
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

// publishTaskStatus 发布任务状态消息到 Redis Pub/Sub
func (h *agentTaskService) publishTaskStatus(taskID, projectGuid, agentType, status, message string, devStage common.DevStatus) error {
	if h.cacheInstance == nil {
		return fmt.Errorf("cache instance is nil")
	}

	statusMsg := &agent.AgentTaskStatusMessage{
		TaskID:      taskID,
		ProjectGuid: projectGuid,
		AgentType:   agentType,
		Status:      status,
		Message:     message,
		DevStage:    string(devStage),
		Timestamp:   utils.GetCurrentTime(),
	}

	// 发布到 Redis Pub/Sub
	err := h.cacheInstance.Publish(common.RedisPubSubChannelAgentTask, statusMsg)
	if err != nil {
		return fmt.Errorf("发布任务状态消息失败: %w", err)
	}

	logger.Info("任务状态消息已发布",
		logger.String("taskID", taskID),
		logger.String("projectGuid", projectGuid),
		logger.String("agentType", agentType),
		logger.String("status", status),
		logger.String("message", message),
	)
	return nil
}
