package services

import (
	"app-maker-agents/internal/api/models"
	"context"
	"encoding/json"
	"fmt"
	"shared-models/agent"
	"shared-models/common"
	"shared-models/logger"
	"shared-models/tasks"
	"shared-models/utils"
	"time"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

type AgentTaskService interface {
	// 处理任务
	ProcessTask(ctx context.Context, task *asynq.Task) error
	// Agent 执行任务
	Enqueue(projectGuid, agentType, message string) (*asynq.TaskInfo, error)
	// Agent 执行任务（带CLI工具）
	EnqueueWithCli(projectGuid, agentType, message, cliTool string) (*asynq.TaskInfo, error)
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
	redisClient    *redis.Client
	keyFormat      string
	// 缓存 guid,session_id 映射关系
}

func NewAgentTaskService(commandService CommandService,
	fileService FileService,
	gitService GitService,
	asyncClient *asynq.Client,
	redisClient *redis.Client) AgentTaskService {
	return &agentTaskService{
		commandService: commandService,
		fileService:    fileService,
		gitService:     gitService,
		asyncClient:    asyncClient,
		redisClient:    redisClient,
		keyFormat:      "project:sessions:%s:%s",
	}
}

// 根据projectGuid从缓存中获取 sessionId
func (h *agentTaskService) getSessionByProjectGuid(projectGuid, agentType string) string {
	if h.redisClient == nil {
		logger.Warn("Redis client is nil, cannot get session", logger.String("projectGuid", projectGuid))
		return ""
	}

	ctx := context.Background()
	key := fmt.Sprintf(h.keyFormat, projectGuid, agentType)

	sessionID, err := h.redisClient.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			logger.Debug("No session found for project", logger.String("projectGuid", projectGuid))
		} else {
			logger.Error("Failed to get session from Redis",
				logger.String("projectGuid", projectGuid),
				logger.String("error", err.Error()))
		}
		return ""
	}

	logger.Info("Retrieved session for project",
		logger.String("projectGuid", projectGuid),
		logger.String("sessionID", sessionID))
	return sessionID
}

func (h *agentTaskService) saveSessionByProjectGuid(projectGuid, agentType, sessionID string) {
	if h.redisClient == nil {
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

	ctx := context.Background()
	key := fmt.Sprintf(h.keyFormat, projectGuid, agentType)

	// 设置过期时间为 24 小时，避免会话数据永久占用内存
	expiration := 24 * time.Hour
	err := h.redisClient.Set(ctx, key, sessionID, expiration).Err()
	if err != nil {
		logger.Error("Failed to save session to Redis",
			logger.String("projectGuid", projectGuid),
			logger.String("sessionID", sessionID),
			logger.String("error", err.Error()))
		return
	}

	logger.Info("Saved session for project",
		logger.String("projectGuid", projectGuid),
		logger.String("sessionID", sessionID),
		logger.String("expiration", expiration.String()))
}

// Enqueue 创建代理执行任务
func (h *agentTaskService) Enqueue(projectGuid, agentType, message string) (*asynq.TaskInfo, error) {
	if h.asyncClient == nil {
		return nil, fmt.Errorf("async client is nil")
	}
	return h.asyncClient.Enqueue(tasks.NewAgentExecuteTask(projectGuid, agentType, message))
}

// EnqueueWithCli 创建带CLI工具的代理执行任务
func (h *agentTaskService) EnqueueWithCli(projectGuid, agentType, message, cliTool string) (*asynq.TaskInfo, error) {
	if h.asyncClient == nil {
		return nil, fmt.Errorf("async client is nil")
	}
	return h.asyncClient.Enqueue(tasks.NewAgentExecuteTaskWithCli(projectGuid, agentType, message, cliTool))
}

// EnqueueReq 创建项目环境准备任务
func (h *agentTaskService) EnqueueSetupReq(req *agent.SetupProjEnvReq) (*asynq.TaskInfo, error) {
	if h.asyncClient == nil {
		return nil, fmt.Errorf("async client is nil")
	}
	if req == nil {
		return nil, fmt.Errorf("EnqueueSetupReq, req is nil")
	}
	return h.asyncClient.Enqueue(tasks.NewProjectSetupTask(req))
}

// 部署项目
func (h *agentTaskService) EnqueueDeployReq(req *agent.DeployReq) (*asynq.TaskInfo, error) {
	if h.asyncClient == nil {
		return nil, fmt.Errorf("async client is nil")
	}
	if req == nil {
		return nil, fmt.Errorf("EnqueueDeployReq, req is nil")
	}
	return h.asyncClient.Enqueue(tasks.NewProjectDeployTask(req))
}

// 与 Agent 对话
func (h *agentTaskService) EnqueueChatWithAgent(req *agent.ChatReq) (*asynq.TaskInfo, error) {
	if h.asyncClient == nil {
		return nil, fmt.Errorf("async client is nil")
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
		return fmt.Errorf("unexpected task type %s", task.Type())
	}
}

// 处理代理执行任务
func (h *agentTaskService) innerProcessAgentExecuteTask(ctx context.Context, task *asynq.Task) error {
	if task.Type() != common.TaskTypeAgentExecute {
		return fmt.Errorf("unexpected task type %s", task.Type())
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

func (h *agentTaskService) innerProcessAgentChatTask(ctx context.Context, task *asynq.Task) error {
	if task.Type() != common.TaskTypeAgentChat {
		return fmt.Errorf("unexpected task type %s", task.Type())
	}

	var req agent.ChatReq
	if err := json.Unmarshal(task.Payload(), &req); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	payload := tasks.AgentExecuteTaskPayload{
		ProjectGUID: req.ProjectGuid,
		AgentType:   req.AgentType,
		Message:     req.Message,
	}
	_, err := h.innerProcessTask(ctx, payload, task)
	if err != nil {
		return err
	}
	return nil
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

	// 从 payload 或项目检测获取 CLI 类型
	cliTool := payload.CliTool
	if cliTool == "" {
		cliTool = h.fileService.DetectCliTool(payload.ProjectGUID)
	}

	// 根据 CLI 类型构建命令
	var cliCommand string
	var args []string
	var useJsonOutput bool

	switch cliTool {
	case common.CliToolClaudeCode:
		cliCommand = "claude"
		useJsonOutput = true
		if sessionID == "" {
			args = []string{"--dangerously-skip-permissions", "--output-format", "json", "-p", "\"" + payload.Message + "\""}
		} else {
			args = []string{"--dangerously-skip-permissions", "--resume", sessionID, "--output-format", "json", "-p", "\"" + payload.Message + "\""}
		}

	case common.CliToolQwenCode:
		cliCommand = "qwen"
		useJsonOutput = false
		args = []string{"-y", "-p", "\"" + payload.Message + "\""}

	case common.CliToolGemini:
		cliCommand = "gemini"
		useJsonOutput = false
		args = []string{"-y", "-p", "\"" + payload.Message + "\""}

	default:
		cliCommand = "claude"
		useJsonOutput = true
		if sessionID == "" {
			args = []string{"--dangerously-skip-permissions", "--output-format", "json", "-p", "\"" + payload.Message + "\""}
		} else {
			args = []string{"--dangerously-skip-permissions", "--resume", sessionID, "--output-format", "json", "-p", "\"" + payload.Message + "\""}
		}
	}

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
		if task != nil {
			tasks.UpdateResult(task.ResultWriter(), common.CommonStatusFailed, 0, result.Error)
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
		if task != nil {
			tasks.UpdateResult(task.ResultWriter(), common.CommonStatusFailed, 0, claudeResponse.Result)
			logger.Error("代理任务执行失败, claude failed",
				logger.String("taskID", task.ResultWriter().TaskID()),
				logger.String("agentType", payload.AgentType),
				logger.String("message", payload.Message),
				logger.String("error", claudeResponse.Result))
		} else {
			logger.Error("代理任务执行失败, claude failed",
				logger.String("agentType", payload.AgentType),
				logger.String("message", payload.Message),
				logger.String("error", claudeResponse.Result))
		}
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
		}
		return nil, fmt.Errorf("项目文档、代码提交并推送失败: %w", err)
	}

	if task != nil {
		tasks.UpdateResult(task.ResultWriter(), common.CommonStatusDone, 100, claudeResponse.Result)
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
	}
	return h.innerProcessTask(ctx, payload, nil)
}
