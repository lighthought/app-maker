package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	cfgPkg "autocodeweb-backend/internal/config"
	"autocodeweb-backend/internal/models"
	"autocodeweb-backend/internal/repositories"
	"shared-models/agent"
	"shared-models/client"
	"shared-models/common"
	"shared-models/logger"
	"shared-models/tasks"
	"shared-models/utils"

	"github.com/hibiken/asynq"
)

type ProjectStageService interface {
	// 获取项目开发阶段
	GetProjectStages(ctx context.Context, projectGuid string) ([]*models.DevStage, error)

	// 处理项目任务
	ProcessTask(ctx context.Context, task *asynq.Task) error

	// 与项目中的 Agent 进行对话
	ChatWithAgent(ctx context.Context, req *agent.ChatReq) error
}

// ProjectStageService 任务执行服务
type projectStageService struct {
	projectRepo      repositories.ProjectRepository
	stageRepo        repositories.StageRepository
	messageRepo      repositories.MessageRepository
	webSocketService WebSocketService
	gitService       GitService
	fileService      FileService
	asyncClient      *asynq.Client
	epicRepo         repositories.EpicRepository
	storyRepo        repositories.StoryRepository
	agentsURL        string
}

// NewTaskExecutionService 创建任务执行服务
func NewProjectStageService(
	projectRepo repositories.ProjectRepository,
	stageRepo repositories.StageRepository,
	messageRepo repositories.MessageRepository,
	webSocketService WebSocketService,
	gitService GitService,
	fileService FileService,
	asyncClient *asynq.Client,
	epicRepo repositories.EpicRepository,
	storyRepo repositories.StoryRepository,
) ProjectStageService {
	// 读取配置
	var agentsURL string
	if cfg, err := cfgPkg.Load(); err == nil {
		agentsURL = cfg.Agents.URL
	}
	return &projectStageService{
		projectRepo:      projectRepo,
		stageRepo:        stageRepo,
		messageRepo:      messageRepo,
		webSocketService: webSocketService,
		gitService:       gitService,
		fileService:      fileService,
		asyncClient:      asyncClient,
		epicRepo:         epicRepo,
		storyRepo:        storyRepo,
		agentsURL:        agentsURL,
	}
}

// GetProjectStages 获取项目开发阶段
func (s *projectStageService) GetProjectStages(ctx context.Context, projectGuid string) ([]*models.DevStage, error) {
	return s.stageRepo.GetByProjectGUID(ctx, projectGuid)
}

// getCliTool 获取项目的 CLI 工具类型
func (s *projectStageService) getCliTool(project *models.Project) string {
	cliTool := project.CliTool
	if cliTool == "" {
		cliTool = project.User.DefaultCliTool
	}
	if cliTool == "" {
		cliTool = common.CliToolClaudeCode
	}
	return cliTool
}

// ProcessTask 处理项目任务
func (h *projectStageService) ProcessTask(ctx context.Context, task *asynq.Task) error {
	switch task.Type() {
	case common.TaskTypeProjectDevelopment:
		return h.handleProjectDevelopmentTask(ctx, task)
	case common.TaskTypeProjectDeploy:
		return h.handleProjectDeployTask(ctx, task)
	case common.TaskTypeAgentChat:
		return h.handleAgentChatTask(ctx, task)
	default:
		return fmt.Errorf("unexpected task type %s", task.Type())
	}
}

// HandleProjectDevelopmentTask 处理项目开发任务
func (s *projectStageService) handleProjectDevelopmentTask(ctx context.Context, t *asynq.Task) error {
	var payload tasks.ProjectTaskPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}
	resultWriter := t.ResultWriter()
	logger.Info("处理项目开发任务", logger.String("taskID", resultWriter.TaskID()))

	project, err := s.projectRepo.GetByGUID(ctx, payload.ProjectGuid)
	if err != nil {
		return fmt.Errorf("获取项目信息失败: %w", err)
	}

	if s.agentsURL != "" {
		s.agentsURL = utils.GetEnvOrDefault("AGENTS_SERVER_URL", "http://host.docker.internal:8088")
	}
	// 使用较长的超时时间，因为 Agent 执行复杂任务（如生成前端页面）可能需要 20-30 分钟
	agentClient := client.NewAgentClient(s.agentsURL, 60*time.Minute)

	// 2. 执行开发阶段
	stages := []struct {
		status      common.DevStatus
		description string
		executor    func(context.Context, *models.Project, *asynq.ResultWriter, *client.AgentClient, *models.DevStage) error
	}{
		{common.DevStatusPendingAgents, "等待Agents处理", s.pendingAgents},
		{common.DevStatusCheckRequirement, "检查需求", s.checkRequirement},
		{common.DevStatusGeneratePRD, "生成PRD文档", s.generatePRD},
		{common.DevStatusDefineUXStandard, "定义UX标准", s.defineUXStandards},
		{common.DevStatusDesignArchitecture, "设计系统架构", s.designArchitecture},
		{common.DevStatusPlanEpicAndStory, "划分Epic和Story", s.planEpicsAndStories},
		{common.DevStatusDefineDataModel, "定义数据模型", s.defineDataModel},
		{common.DevStatusDefineAPI, "定义API接口", s.defineAPIs},
		{common.DevStatusGeneratePages, "生成前端页面", s.generateFrontendPages},
		// TODO: 调试阶段注释，{common.DevStatusDevelopStory, "开发Story功能", s.developStories},
		//{common.DevStatusFixBug, "修复开发问题", s.fixBugs}, // 这个要用户前端输入，可以提供入口
		{common.DevStatusRunTest, "执行自动测试", s.runTests},
		{common.DevStatusDeploy, "打包项目", s.packageProject},
	}

	gitConfig := &GitConfig{
		UserID:        project.UserID,
		GUID:          project.GUID,
		ProjectPath:   project.ProjectPath,
		CommitMessage: fmt.Sprintf("Auto commit by App Maker - %s", project.Name),
	}

	for _, stage := range stages {
		devProjectStage, err := s.stageRepo.GetByProjectGuidAndName(ctx, project.GUID, string(stage.status))
		if err == nil && devProjectStage.Status == common.CommonStatusDone {
			tasks.UpdateResult(resultWriter, common.CommonStatusInProgress, common.GetDevStageProgress(stage.status), common.GetDevStageDescription(stage.status)+"已完成")
			continue
		}

		if err != nil {
			devProjectStage = nil
		} else if devProjectStage != nil {
			devProjectStage.TaskID = project.CurrentTaskID
			s.stageRepo.Update(ctx, devProjectStage)
			s.webSocketService.NotifyProjectStageUpdate(ctx, project.GUID, devProjectStage)
		}

		// 执行阶段
		if err := stage.executor(ctx, project, resultWriter, agentClient, devProjectStage); err != nil {
			logger.Error("开发阶段执行失败",
				logger.String("projectID", project.ID),
				logger.String("stage", string(stage.status)),
				logger.String("error", err.Error()),
			)

			// 更新项目状态为失败
			project.SetDevStatus(common.DevStatusFailed)
			s.projectRepo.Update(ctx, project)
			return err
		}

		if err := s.gitService.Pull(ctx, gitConfig); err != nil {
			logger.Error("拉取远程仓库代码失败",
				logger.String("error", err.Error()),
				logger.String("projectID", project.ID),
			)
		}

		// TODO: 等待当前阶段变成完成状态、不再是暂停的状态
	}

	// 开发完成
	project.SetDevStatus(common.DevStatusDone)
	project.Status = common.CommonStatusDone
	s.projectRepo.Update(ctx, project)
	s.webSocketService.NotifyProjectInfoUpdate(ctx, project.GUID, project)

	logger.Info("项目开发流程执行完成",
		logger.String("projectID", project.ID),
	)
	tasks.UpdateResult(resultWriter, common.CommonStatusDone, 100, "项目开发任务完成")
	return nil
}

// 处理项目部署任务
func (s *projectStageService) handleProjectDeployTask(ctx context.Context, t *asynq.Task) error {
	var req agent.DeployReq
	if err := json.Unmarshal(t.Payload(), &req); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}
	resultWriter := t.ResultWriter()
	logger.Info("处理项目部署任务", logger.String("taskID", resultWriter.TaskID()))

	project, err := s.projectRepo.GetByGUID(ctx, req.ProjectGuid)
	if err != nil {
		return fmt.Errorf("获取项目信息失败: %w", err)
	}

	// 使用较长的超时时间，因为部署任务可能需要较长时间
	agentClient := client.NewAgentClient(s.agentsURL, 60*time.Minute)
	response, err := agentClient.Deploy(ctx, &req)
	if err != nil {
		tasks.UpdateResult(resultWriter, common.CommonStatusFailed, 0, "调用 Dev Agent 打包失败: "+err.Error())
		return err
	}

	projectMsg := &models.ConversationMessage{
		ProjectGuid:     project.GUID,
		Type:            common.ConversationTypeAgent,
		AgentRole:       common.AgentDev.Role,
		AgentName:       common.AgentDev.Name,
		Content:         "项目项目已打包部署",
		IsMarkdown:      true,
		MarkdownContent: response.Message,
		IsExpanded:      true,
	}

	// 设置预览 URL
	if project.PreviewUrl == "" {
		project.PreviewUrl = fmt.Sprintf("http://%s.app-maker.localhost", project.GUID)
		if err := s.projectRepo.Update(ctx, project); err != nil {
			logger.Error("更新项目预览URL失败",
				logger.String("error", err.Error()),
				logger.String("projectID", project.ID),
			)
		} else {
			logger.Info("项目预览URL已设置",
				logger.String("projectID", project.ID),
				logger.String("previewUrl", project.PreviewUrl),
			)
			// 通知前端预览URL已设置
			s.webSocketService.NotifyProjectInfoUpdate(ctx, project.GUID, project)
		}
	}

	s.notifyProjectStatusChange(ctx, project, projectMsg, nil)

	tasks.UpdateResult(resultWriter, common.CommonStatusDone, 100, "项目项目已打包部署")
	return nil
}

// 处理与 Agent 对话任务
func (s *projectStageService) handleAgentChatTask(ctx context.Context, task *asynq.Task) error {
	var req agent.ChatReq
	if err := json.Unmarshal(task.Payload(), &req); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	resultWriter := task.ResultWriter()
	logger.Info("🔵 [AgentChat] 开始处理 Agent 对话任务",
		logger.String("taskID", resultWriter.TaskID()),
		logger.String("projectGUID", req.ProjectGuid),
		logger.String("agentType", req.AgentType),
		logger.String("message", req.Message),
	)
	tasks.UpdateResult(resultWriter, common.CommonStatusInProgress, 0, "开始处理对话任务")

	// 创建用户消息
	userMessage := &models.ConversationMessage{
		ProjectGuid:     req.ProjectGuid,
		Type:            common.ConversationTypeUser,
		AgentRole:       common.AgentTypeUser,
		AgentName:       "user",
		Content:         req.Message,
		IsMarkdown:      false,
		MarkdownContent: req.Message,
		IsExpanded:      false,
	}
	// 保存用户消息
	logger.Info("🔵 [AgentChat] 保存用户消息到数据库",
		logger.String("projectGUID", req.ProjectGuid),
	)
	if err := s.messageRepo.Create(ctx, userMessage); err != nil {
		logger.Error("保存用户消息失败",
			logger.String("error", err.Error()),
			logger.String("projectGUID", req.ProjectGuid),
		)
	} else {
		logger.Info("🔵 [AgentChat] 用户消息保存成功",
			logger.String("messageID", userMessage.ID),
		)
	}

	logger.Info("🔵 [AgentChat] 推送用户消息到前端",
		logger.String("projectGUID", req.ProjectGuid),
		logger.String("messageID", userMessage.ID),
	)
	s.webSocketService.NotifyProjectMessage(ctx, req.ProjectGuid, userMessage)
	tasks.UpdateResult(resultWriter, common.CommonStatusInProgress, 10, "处理对话数据")

	// 获取项目信息
	logger.Info("🔵 [AgentChat] 获取项目信息",
		logger.String("projectGUID", req.ProjectGuid),
	)
	project, err := s.projectRepo.GetByGUID(ctx, req.ProjectGuid)
	if err != nil {
		logger.Error("🔴 [AgentChat] 获取项目信息失败",
			logger.String("error", err.Error()),
			logger.String("projectGUID", req.ProjectGuid),
		)
		tasks.UpdateResult(resultWriter, common.CommonStatusFailed, 0, "无法获取项目信息")
		return fmt.Errorf("获取项目信息失败: %w", err)
	}
	logger.Info("🔵 [AgentChat] 项目信息获取成功",
		logger.String("projectID", project.ID),
		logger.String("projectStatus", project.Status),
		logger.String("devStatus", project.DevStatus),
	)

	if project.Status == common.CommonStatusPaused {
		logger.Info("🔵 [AgentChat] 项目处于暂停状态，恢复为进行中",
			logger.String("projectID", project.ID),
		)
		project.Status = common.CommonStatusInProgress
		s.projectRepo.Update(ctx, project)
		s.webSocketService.NotifyProjectInfoUpdate(ctx, project.GUID, project)
		tasks.UpdateResult(resultWriter, common.CommonStatusInProgress, 20, "处理项目状态")
	}

	// 恢复当前暂停的阶段
	logger.Info("🔵 [AgentChat] 检查当前阶段状态",
		logger.String("projectGUID", req.ProjectGuid),
		logger.String("devStatus", project.DevStatus),
	)
	currentStage, err := s.stageRepo.GetByProjectGuidAndName(ctx, project.GUID, project.DevStatus)
	if err == nil && currentStage != nil && currentStage.Status == common.CommonStatusPaused {
		logger.Info("🔵 [AgentChat] 阶段处于暂停状态，恢复为进行中",
			logger.String("stageID", currentStage.ID),
			logger.String("stageName", currentStage.Name),
		)
		currentStage.Status = common.CommonStatusInProgress
		if err := s.stageRepo.Update(ctx, currentStage); err != nil {
			logger.Error("恢复阶段状态失败",
				logger.String("error", err.Error()),
				logger.String("projectID", project.ID),
				logger.String("stageID", currentStage.ID),
			)
		} else {
			s.webSocketService.NotifyProjectStageUpdate(ctx, project.GUID, currentStage)
			tasks.UpdateResult(resultWriter, common.CommonStatusInProgress, 30, "恢复阶段状态")
		}
	}

	logger.Info("🟢 [AgentChat] 项目执行已恢复",
		logger.String("projectID", project.ID),
		logger.String("devStatus", project.DevStatus),
	)

	tasks.UpdateResult(resultWriter, common.CommonStatusInProgress, 35, "和 Agent 对话中...")
	logger.Info("🔵 [AgentChat] 开始调用 Agent 模块",
		logger.String("agentsURL", s.agentsURL),
		logger.String("agentType", req.AgentType),
	)
	// 使用较长的超时时间，因为 Agent 执行复杂任务可能需要 20-30 分钟甚至更长
	agentClient := client.NewAgentClient(s.agentsURL, 60*time.Minute)
	// 使用 background context 避免 HTTP 请求超时，但保留原 context 用于取消信号
	response, err := agentClient.ChatWithAgent(ctx, &req)
	if err != nil {
		logger.Error("🔴 [AgentChat] Agent 对话失败",
			logger.String("error", err.Error()),
			logger.String("agentType", req.AgentType),
		)
		tasks.UpdateResult(resultWriter, common.CommonStatusFailed, 0, "和 Agent 对话失败: "+err.Error())
		return err
	}
	logger.Info("🟢 [AgentChat] Agent 对话成功",
		logger.String("agentType", req.AgentType),
		logger.String("responseLength", fmt.Sprintf("%d", len(response.Message))),
	)

	agent := common.GetAgentByAgentType(req.AgentType)
	if agent == nil {
		agent = &common.AgentDev
	}

	projectMsg := &models.ConversationMessage{
		ProjectGuid:     project.GUID,
		Type:            common.ConversationTypeAgent,
		AgentRole:       agent.Role,
		AgentName:       agent.Name,
		Content:         "已完成",
		IsMarkdown:      true,
		MarkdownContent: response.Message,
		IsExpanded:      true,
	}

	logger.Info("🔵 [AgentChat] 保存并推送 Agent 响应消息",
		logger.String("projectGUID", project.GUID),
		logger.String("agentRole", agent.Role),
		logger.String("agentName", agent.Name),
	)
	// 支持多轮对话
	s.notifyProjectStatusChange(ctx, project, projectMsg, currentStage)

	logger.Info("🟢 [AgentChat] Agent 对话任务执行完成",
		logger.String("taskID", resultWriter.TaskID()),
		logger.String("projectGUID", req.ProjectGuid),
	)
	tasks.UpdateResult(resultWriter, common.CommonStatusDone, 100, "Agent 对话任务执行完成")
	return nil
}

// 统一由这个函数更新项目状态
func (s *projectStageService) notifyProjectStatusChange(ctx context.Context,
	project *models.Project, message *models.ConversationMessage, stage *models.DevStage) {
	if message != nil {
		// 检查是否需要暂停（Agent 消息包含问题）
		if message.Type == common.ConversationTypeAgent {
			hasQuestion := utils.ContainsQuestion(message.Content) || utils.ContainsQuestion(message.MarkdownContent)
			if hasQuestion {
				message.HasQuestion = true
				message.WaitingUserResponse = true
				message.Content = strings.Replace(message.Content, "已完成", "需要反馈", 1)

				// 暂停项目和当前阶段
				project.Status = common.CommonStatusPaused
				if stage != nil {
					stage.Status = common.CommonStatusPaused
				}

				logger.Info("检测到 Agent 问题，暂停项目执行",
					logger.String("projectID", project.ID),
					logger.String("agentRole", message.AgentRole),
					logger.String("agentName", message.AgentName),
				)
			}
		}

		// 保存用户消息
		if err := s.messageRepo.Create(ctx, message); err != nil {
			logger.Error("保存项目消息失败",
				logger.String("error", err.Error()),
				logger.String("projectID", project.ID),
			)
		}
		s.webSocketService.NotifyProjectMessage(ctx, project.GUID, message)
	}

	if stage != nil {
		if stage.ID == "" {
			// 插入项目阶段
			if err := s.stageRepo.Create(ctx, stage); err != nil {
				logger.Error("插入项目阶段失败",
					logger.String("error", err.Error()),
					logger.String("projectID", project.ID),
				)
			}

			project.SetDevStatus(common.DevStatus(stage.Name))
			s.projectRepo.Update(ctx, project)
			s.webSocketService.NotifyProjectStageUpdate(ctx, project.GUID, stage)

			logger.Info("插入项目阶段成功", logger.String("projectID", project.ID), logger.String("stageID", stage.ID))
		} else {
			stage.ProjectID = project.ID
			stage.ProjectGuid = project.GUID
			if err := s.stageRepo.Update(ctx, stage); err != nil {
				logger.Error("更新项目阶段失败",
					logger.String("error", err.Error()),
					logger.String("projectID", project.ID),
					logger.String("stageID", stage.ID),
					logger.String("stageName", stage.Name),
					logger.String("status", stage.Status),
				)
			}
			s.webSocketService.NotifyProjectStageUpdate(ctx, project.GUID, stage)
			logger.Info("更新项目阶段成功", logger.String("projectID", project.ID), logger.String("stageID", stage.ID))
		}
	}
}

// pendingAgents 准备项目开发环境
func (s *projectStageService) pendingAgents(ctx context.Context,
	project *models.Project, resultWriter *asynq.ResultWriter,
	agentClient *client.AgentClient, devStage *models.DevStage) error {
	var devProjectStage *models.DevStage
	if devStage == nil {
		devProjectStage = models.NewDevStage(project, common.DevStatusPendingAgents, common.CommonStatusInProgress)
	} else {
		devProjectStage = devStage
		devProjectStage.SetStatus(common.CommonStatusInProgress)
	}

	s.notifyProjectStatusChange(ctx, project, nil, devProjectStage)

	// 设置 CLI 工具和模型配置，如果项目没有设置则使用用户的默认设置
	cliTool := project.CliTool
	aiModel := project.AiModel
	modelProvider := project.ModelProvider
	modelApiUrl := project.ModelApiUrl

	if cliTool == "" {
		cliTool = project.User.DefaultCliTool
	}
	if aiModel == "" {
		aiModel = project.User.DefaultAiModel
	}
	if modelProvider == "" {
		modelProvider = project.User.DefaultModelProvider
	}
	if modelApiUrl == "" {
		modelApiUrl = project.User.DefaultModelApiUrl
	}

	// 如果还是空，使用系统默认值
	if cliTool == "" {
		cliTool = common.CliToolClaudeCode
	}
	if aiModel == "" {
		aiModel = common.DefaultModelByProvider[common.ModelProviderZhipu]
	}
	if modelProvider == "" {
		modelProvider = common.ModelProviderZhipu
	}
	if modelApiUrl == "" {
		modelApiUrl = common.DefaultAPIUrlByProvider[common.ModelProviderZhipu]
	}

	// 获取 API Token
	apiToken := project.ApiToken
	if apiToken == "" {
		apiToken = project.User.DefaultApiToken
	}

	result, err := agentClient.SetupProjectEnvironment(ctx, &agent.SetupProjEnvReq{
		ProjectGuid:     project.GUID,
		GitlabRepoUrl:   project.GitlabRepoURL,
		SetupBmadMethod: true,
		BmadCliType:     cliTool,
		AiModel:         aiModel,
		ModelProvider:   modelProvider,
		ModelApiUrl:     modelApiUrl,
		ApiToken:        apiToken,
	})
	if err != nil {
		tasks.UpdateResult(resultWriter, common.CommonStatusFailed, 0, "agents 项目环境准备失败: "+err.Error())
		devProjectStage.SetStatus(common.CommonStatusFailed)
		devProjectStage.FailedReason = err.Error()
		s.notifyProjectStatusChange(ctx, project, nil, devProjectStage)
		return err
	}

	projectMsg := &models.ConversationMessage{
		ProjectGuid:     project.GUID,
		Type:            common.ConversationTypeAgent,
		AgentRole:       common.AgentPM.Role,
		AgentName:       common.AgentPM.Name,
		Content:         "项目开发环境已准备完成",
		IsMarkdown:      true,
		MarkdownContent: result.Message,
		IsExpanded:      true,
	}

	devProjectStage.SetStatus(common.CommonStatusDone)
	s.notifyProjectStatusChange(ctx, project, projectMsg, devProjectStage)

	tasks.UpdateResult(resultWriter, common.CommonStatusInProgress, common.GetDevStageProgress(common.DevStatusPendingAgents), "项目开发环境已准备完成")

	return nil
}

// checkRequirement 检查需求
func (s *projectStageService) checkRequirement(ctx context.Context,
	project *models.Project, resultWriter *asynq.ResultWriter,
	agentClient *client.AgentClient, devStage *models.DevStage) error {
	var devProjectStage *models.DevStage
	if devStage == nil {
		devProjectStage = models.NewDevStage(project, common.DevStatusCheckRequirement, common.CommonStatusInProgress)
	} else {
		devProjectStage = devStage
		devProjectStage.SetStatus(common.CommonStatusInProgress)
	}

	s.notifyProjectStatusChange(ctx, project, nil, devProjectStage)

	req := &agent.GetProjBriefReq{
		ProjectGuid:  project.GUID,
		Requirements: project.Requirements,
		CliTool:      s.getCliTool(project),
	}

	response, err := agentClient.AnalyseProjectBrief(ctx, req)
	if err != nil {
		tasks.UpdateResult(resultWriter, common.CommonStatusFailed, 0, "调用 Analyst Agent 检查需求失败: "+err.Error())
		devProjectStage.SetStatus(common.CommonStatusFailed)
		devProjectStage.FailedReason = err.Error()
		s.notifyProjectStatusChange(ctx, project, nil, devProjectStage)
		return err
	}

	projectMsg := &models.ConversationMessage{
		ProjectGuid:     project.GUID,
		Type:            common.ConversationTypeAgent,
		AgentRole:       common.AgentAnalyst.Role,
		AgentName:       common.AgentAnalyst.Name,
		Content:         "项目需求已检查完成",
		IsMarkdown:      true,
		MarkdownContent: response.Message,
		IsExpanded:      true,
	}

	devProjectStage.SetStatus(common.CommonStatusDone)
	s.notifyProjectStatusChange(ctx, project, projectMsg, devProjectStage)

	tasks.UpdateResult(resultWriter, common.CommonStatusInProgress, common.GetDevStageProgress(common.DevStatusCheckRequirement), "项目需求已检查完成")
	return nil
}

// generatePRD 生成PRD文档
func (s *projectStageService) generatePRD(ctx context.Context,
	project *models.Project, resultWriter *asynq.ResultWriter,
	agentClient *client.AgentClient, devStage *models.DevStage) error {
	var devProjectStage *models.DevStage
	if devStage == nil {
		devProjectStage = models.NewDevStage(project, common.DevStatusGeneratePRD, common.CommonStatusInProgress)
	} else {
		devProjectStage = devStage
		devProjectStage.SetStatus(common.CommonStatusInProgress)
	}

	s.notifyProjectStatusChange(ctx, project, nil, devProjectStage)
	generatePrdReq := &agent.GetPRDReq{
		ProjectGuid:  project.GUID,
		Requirements: project.Requirements,
		CliTool:      s.getCliTool(project),
	}
	// 调用 agents-server 生成 PRD 文档，并提交到 GitLab
	response, err := agentClient.GetPRD(ctx, generatePrdReq)
	if err != nil {
		tasks.UpdateResult(resultWriter, common.CommonStatusFailed, 0, "调用 PM Agent 生成 PRD 文档失败: "+err.Error())
		devProjectStage.SetStatus(common.CommonStatusFailed)
		devProjectStage.FailedReason = err.Error()
		s.notifyProjectStatusChange(ctx, project, nil, devProjectStage)
		return err
	}

	projectMsg := &models.ConversationMessage{
		ProjectGuid:     project.GUID,
		Type:            common.ConversationTypeAgent,
		AgentRole:       common.AgentPM.Role,
		AgentName:       common.AgentPM.Name,
		Content:         "项目PRD文档已生成",
		IsMarkdown:      true,
		MarkdownContent: response.Message,
		IsExpanded:      true,
	}

	devProjectStage.SetStatus(common.CommonStatusDone)
	s.notifyProjectStatusChange(ctx, project, projectMsg, devProjectStage)

	tasks.UpdateResult(resultWriter, common.CommonStatusInProgress, common.GetDevStageProgress(common.DevStatusGeneratePRD), "项目PRD文档已生成")
	return nil
}

// defineUXStandards 定义UX标准
func (s *projectStageService) defineUXStandards(ctx context.Context,
	project *models.Project, resultWriter *asynq.ResultWriter,
	agentClient *client.AgentClient, devStage *models.DevStage) error {
	var devProjectStage *models.DevStage
	if devStage == nil {
		devProjectStage = models.NewDevStage(project, common.DevStatusDefineUXStandard, common.CommonStatusInProgress)
	} else {
		devProjectStage = devStage
		devProjectStage.SetStatus(common.CommonStatusInProgress)
	}

	s.notifyProjectStatusChange(ctx, project, nil, devProjectStage)

	req := &agent.GetUXStandardReq{
		ProjectGuid:  project.GUID,
		Requirements: project.Requirements,
		PrdPath:      "docs/PRD.md",
		CliTool:      s.getCliTool(project),
	}
	// 调用 agents-server 定义 UX 标准
	response, err := agentClient.GetUXStandard(ctx, req)
	if err != nil {
		tasks.UpdateResult(resultWriter, common.CommonStatusFailed, 0, "调用 UX Agent 失败: "+err.Error())
		devProjectStage.SetStatus(common.CommonStatusFailed)
		devProjectStage.FailedReason = err.Error()
		s.notifyProjectStatusChange(ctx, project, nil, devProjectStage)
		return err
	}

	projectMsg := &models.ConversationMessage{
		ProjectGuid:     project.GUID,
		Type:            common.ConversationTypeAgent,
		AgentRole:       common.AgentUXExpert.Role,
		AgentName:       common.AgentUXExpert.Name,
		Content:         "项目UX标准已定义",
		IsMarkdown:      true,
		MarkdownContent: response.Message,
		IsExpanded:      true,
	}

	devProjectStage.SetStatus(common.CommonStatusDone)
	s.notifyProjectStatusChange(ctx, project, projectMsg, devProjectStage)

	tasks.UpdateResult(resultWriter, common.CommonStatusInProgress, common.GetDevStageProgress(common.DevStatusDefineUXStandard), "项目UX标准已定义")
	return nil
}

// designArchitecture 设计系统架构
func (s *projectStageService) designArchitecture(ctx context.Context,
	project *models.Project, resultWriter *asynq.ResultWriter,
	agentClient *client.AgentClient, devStage *models.DevStage) error {
	var devProjectStage *models.DevStage
	if devStage == nil {
		devProjectStage = models.NewDevStage(project, common.DevStatusDesignArchitecture, common.CommonStatusInProgress)
	} else {
		devProjectStage = devStage
		devProjectStage.SetStatus(common.CommonStatusInProgress)
	}

	s.notifyProjectStatusChange(ctx, project, nil, devProjectStage)

	req := &agent.GetArchitectureReq{
		ProjectGuid: project.GUID,
		PrdPath:     "docs/PRD.md",
		UxSpecPath:  "docs/ux/ux-spec.md",
		// 从模板中读取架构信息
		TemplateArchDescription: "1. 前端：vue.js+ vite ；\n" +
			"2. 后端服务和 API： GO + Gin 框架实现 API、数据库用 PostgreSql、缓存用 Redis。\n" +
			"3. 部署相关的脚本已经有了，用的 docker，前端用一个 nginx ，配置 /api 重定向到 /backend:port ，这样就能在前端项目中访问后端 API 了。" +
			" 引用关系是：前端依赖后端，后端依赖 Redis 和 PostgreSql。",
		CliTool: s.getCliTool(project),
	}
	// 调用 agents-server 设计系统架构
	response, err := agentClient.GetArchitecture(ctx, req)
	if err != nil {
		tasks.UpdateResult(resultWriter, common.CommonStatusFailed, 0, "调用 Architect Agent 失败: "+err.Error())
		devProjectStage.SetStatus(common.CommonStatusFailed)
		devProjectStage.FailedReason = err.Error()
		s.notifyProjectStatusChange(ctx, project, nil, devProjectStage)
		return err
	}

	projectMsg := &models.ConversationMessage{
		ProjectGuid:     project.GUID,
		Type:            common.ConversationTypeAgent,
		AgentRole:       common.AgentArchitect.Role,
		AgentName:       common.AgentArchitect.Name,
		Content:         "项目系统架构已设计",
		IsMarkdown:      true,
		MarkdownContent: response.Message,
		IsExpanded:      true,
	}

	devProjectStage.SetStatus(common.CommonStatusDone)
	s.notifyProjectStatusChange(ctx, project, projectMsg, devProjectStage)

	tasks.UpdateResult(resultWriter, common.CommonStatusInProgress, common.GetDevStageProgress(common.DevStatusDesignArchitecture), "项目系统架构已设计")
	return nil
}

// defineDataModel 定义数据模型
func (s *projectStageService) defineDataModel(ctx context.Context,
	project *models.Project, resultWriter *asynq.ResultWriter,
	agentClient *client.AgentClient, devStage *models.DevStage) error {
	var devProjectStage *models.DevStage
	if devStage == nil {
		devProjectStage = models.NewDevStage(project, common.DevStatusDefineDataModel, common.CommonStatusInProgress)
	} else {
		devProjectStage = devStage
		devProjectStage.SetStatus(common.CommonStatusInProgress)
	}

	s.notifyProjectStatusChange(ctx, project, nil, devProjectStage)

	req := &agent.GetDatabaseDesignReq{
		ProjectGuid:   project.GUID,
		PrdPath:       "docs/PRD.md",
		ArchFolder:    "docs/arch",
		StoriesFolder: "docs/stories",
		CliTool:       s.getCliTool(project),
	}
	// 调用 agents-server 定义数据模型
	response, err := agentClient.GetDatabaseDesign(ctx, req)
	if err != nil {
		tasks.UpdateResult(resultWriter, common.CommonStatusFailed, 0, "调用 Architect Agent 失败: "+err.Error())
		devProjectStage.SetStatus(common.CommonStatusFailed)
		devProjectStage.FailedReason = err.Error()
		s.notifyProjectStatusChange(ctx, project, nil, devProjectStage)
		return err
	}

	projectMsg := &models.ConversationMessage{
		ProjectGuid:     project.GUID,
		Type:            common.ConversationTypeAgent,
		AgentRole:       common.AgentArchitect.Role,
		AgentName:       common.AgentArchitect.Name,
		Content:         "项目数据模型已定义",
		IsMarkdown:      true,
		MarkdownContent: response.Message,
		IsExpanded:      true,
	}

	devProjectStage.SetStatus(common.CommonStatusDone)
	s.notifyProjectStatusChange(ctx, project, projectMsg, devProjectStage)

	tasks.UpdateResult(resultWriter, common.CommonStatusInProgress, common.GetDevStageProgress(common.DevStatusDefineDataModel), "项目数据模型已定义")
	return nil
}

// defineAPIs 定义API接口
func (s *projectStageService) defineAPIs(ctx context.Context,
	project *models.Project, resultWriter *asynq.ResultWriter,
	agentClient *client.AgentClient, devStage *models.DevStage) error {
	var devProjectStage *models.DevStage
	if devStage == nil {
		devProjectStage = models.NewDevStage(project, common.DevStatusDefineAPI, common.CommonStatusInProgress)
	} else {
		devProjectStage = devStage
		devProjectStage.SetStatus(common.CommonStatusInProgress)
	}

	s.notifyProjectStatusChange(ctx, project, nil, devProjectStage)

	req := &agent.GetAPIDefinitionReq{
		ProjectGuid:   project.GUID,
		PrdPath:       "docs/PRD.md",
		DbFolder:      "docs/db",
		StoriesFolder: "docs/stories",
		CliTool:       s.getCliTool(project),
	}
	// 调用 agents-server 定义 API 接口
	response, err := agentClient.GetAPIDefinition(ctx, req)
	if err != nil {
		tasks.UpdateResult(resultWriter, common.CommonStatusFailed, 0, "调用 Architect Agent 失败: "+err.Error())
		devProjectStage.SetStatus(common.CommonStatusFailed)
		devProjectStage.FailedReason = err.Error()
		s.notifyProjectStatusChange(ctx, project, nil, devProjectStage)
		return err
	}

	projectMsg := &models.ConversationMessage{
		ProjectGuid:     project.GUID,
		Type:            common.ConversationTypeAgent,
		AgentRole:       common.AgentArchitect.Role,
		AgentName:       common.AgentArchitect.Name,
		Content:         "项目API接口已定义",
		IsMarkdown:      true,
		MarkdownContent: response.Message,
		IsExpanded:      true,
	}

	devProjectStage.SetStatus(common.CommonStatusDone)
	s.notifyProjectStatusChange(ctx, project, projectMsg, devProjectStage)

	tasks.UpdateResult(resultWriter, common.CommonStatusInProgress, common.GetDevStageProgress(common.DevStatusDefineAPI), "项目API接口已定义")
	return nil
}

// planEpicsAndStories 划分Epic和Story
func (s *projectStageService) planEpicsAndStories(ctx context.Context,
	project *models.Project, resultWriter *asynq.ResultWriter,
	agentClient *client.AgentClient, devStage *models.DevStage) error {
	var devProjectStage *models.DevStage
	if devStage == nil {
		devProjectStage = models.NewDevStage(project, common.DevStatusPlanEpicAndStory, common.CommonStatusInProgress)
	} else {
		devProjectStage = devStage
		devProjectStage.SetStatus(common.CommonStatusInProgress)
	}

	s.notifyProjectStatusChange(ctx, project, nil, devProjectStage)

	req := &agent.GetEpicsAndStoriesReq{
		ProjectGuid: project.GUID,
		PrdPath:     "docs/PRD.md",
		ArchFolder:  "docs/arch",
		CliTool:     s.getCliTool(project),
	}
	// 调用 agents-server 划分 Epics 和 Stories
	response, err := agentClient.GetEpicsAndStories(ctx, req)
	if err != nil {
		tasks.UpdateResult(resultWriter, common.CommonStatusFailed, 0, "调用 PO Agent 失败: "+err.Error())
		devProjectStage.SetStatus(common.CommonStatusFailed)
		devProjectStage.FailedReason = err.Error()
		s.notifyProjectStatusChange(ctx, project, nil, devProjectStage)
		return err
	}

	// 解析返回的 markdown 中的 MVP Epics JSON 信息
	mvpData, err := s.extractMvpEpicsJSON(response.Message)
	if err == nil && mvpData != nil {
		// 保存到数据库
		if err := s.saveMvpEpics(ctx, project, mvpData); err != nil {
			logger.Error("保存 MVP Epics 失败", logger.String("error", err.Error()))
		} else {
			logger.Info("MVP Epics 已保存到数据库")
		}
	} else {
		logger.Warn("未能提取 MVP Epics JSON，将依赖文件方式读取", logger.String("error", err.Error()))
	}

	projectMsg := &models.ConversationMessage{
		ProjectGuid:     project.GUID,
		Type:            common.ConversationTypeAgent,
		AgentRole:       common.AgentPO.Role,
		AgentName:       common.AgentPO.Name,
		Content:         "项目Epic和Story已划分",
		IsMarkdown:      true,
		MarkdownContent: response.Message,
		IsExpanded:      true,
	}

	devProjectStage.SetStatus(common.CommonStatusDone)
	s.notifyProjectStatusChange(ctx, project, projectMsg, devProjectStage)

	// TODO: 让用户反馈，这个部分是比较关键的，后期加入了交互以后，需要调整这一块内容
	tasks.UpdateResult(resultWriter, common.CommonStatusInProgress, common.GetDevStageProgress(common.DevStatusPlanEpicAndStory), "项目Epic和Story已划分")
	return nil
}

// developStories 开发Story功能 (只实现 MVP Stories)
func (s *projectStageService) developStories(ctx context.Context,
	project *models.Project, resultWriter *asynq.ResultWriter,
	agentClient *client.AgentClient, devStage *models.DevStage) error {
	var devProjectStage *models.DevStage
	if devStage == nil {
		devProjectStage = models.NewDevStage(project, common.DevStatusDevelopStory, common.CommonStatusInProgress)
	} else {
		devProjectStage = devStage
		devProjectStage.SetStatus(common.CommonStatusInProgress)
	}

	s.notifyProjectStatusChange(ctx, project, nil, devProjectStage)

	// 尝试从数据库获取 MVP 阶段的 Epics (P0 优先级)
	mvpEpics, err := s.epicRepo.GetMvpEpicsByProject(ctx, project.ID)

	// 如果数据库中没有 MVP Epics，fallback 到文件方式
	if err != nil || len(mvpEpics) == 0 {
		logger.Warn("数据库中未找到 MVP Epics，使用文件方式", logger.String("error", err.Error()))
		return s.developStoriesFromFiles(ctx, project, resultWriter, agentClient, devProjectStage)
	}

	logger.Info("从数据库读取到 MVP Epics", logger.Int("count", len(mvpEpics)))

	req := &agent.ImplementStoryReq{
		ProjectGuid: project.GUID,
		PrdPath:     "docs/PRD.md",
		ArchFolder:  "docs/arch/",
		DbFolder:    "docs/db/",
		ApiFolder:   "docs/api/",
		UxSpecPath:  "docs/ux/ux-spec.md",
		EpicFile:    "docs/stories/",
		StoryFile:   "",
		CliTool:     s.getCliTool(project),
	}

	bDev := (utils.GetEnvOrDefault("ENVIRONMENT", common.EnvironmentDevelopment) == common.EnvironmentDevelopment)
	developStoryCount := 0
	totalStoryCount := 0
	var lastResponse *tasks.TaskResult

	// 按 Epic 和 Story 的顺序实现
	for epicIndex, epic := range mvpEpics {
		logger.Info("开始实现 Epic",
			logger.String("epic_id", epic.ID),
			logger.String("epic_name", epic.Name),
			logger.Int("story_count", len(epic.Stories)))

		for storyIndex, story := range epic.Stories {
			totalStoryCount++

			// 跳过已完成的 Story
			if story.Status == common.CommonStatusDone {
				logger.Info("Story 已完成，跳过",
					logger.String("story_number", story.StoryNumber),
					logger.String("story_title", story.Title))
				continue
			}

			// 开发环境只实现第一个 Story
			if developStoryCount >= 1 && bDev {
				logger.Info("开发模式：跳过 Story",
					logger.String("story_number", story.StoryNumber),
					logger.String("story_title", story.Title))

				// 模拟完成
				lastResponse = &tasks.TaskResult{
					Message: fmt.Sprintf("开发模式：跳过 Story %s - %s", story.StoryNumber, story.Title),
				}
				continue
			}

			// 设置 Story 文件路径
			req.StoryFile = story.FilePath
			req.EpicFile = story.FilePath

			logger.Info("开始实现 Story",
				logger.String("story_number", story.StoryNumber),
				logger.String("story_title", story.Title),
				logger.String("story_file", story.FilePath))

			// 调用 Dev Agent 实现 Story
			response, err := agentClient.ImplementStory(ctx, req)
			if err != nil {
				logger.Error("Story 实现失败",
					logger.String("story_number", story.StoryNumber),
					logger.String("error", err.Error()))

				tasks.UpdateResult(resultWriter, common.CommonStatusFailed, 0, "调用 Dev Agent 开发失败: "+err.Error())
				devProjectStage.SetStatus(common.CommonStatusFailed)
				devProjectStage.FailedReason = err.Error()

				// 更新 Story 状态为失败
				story.Status = common.CommonStatusFailed
				s.storyRepo.Update(ctx, &story)

				s.notifyProjectStatusChange(ctx, project, nil, devProjectStage)
				return err
			}

			// 更新 Story 状态为完成
			story.Status = common.CommonStatusDone
			if err := s.storyRepo.Update(ctx, &story); err != nil {
				logger.Error("更新 Story 状态失败", logger.String("error", err.Error()))
			}

			logger.Info("Story 实现成功",
				logger.String("story_number", story.StoryNumber),
				logger.String("story_title", story.Title))

			developStoryCount++
			lastResponse = response

			// 不是最后一个 Story，发送中间消息
			if !(epicIndex == len(mvpEpics)-1 && storyIndex == len(epic.Stories)-1) {
				projectMsg := &models.ConversationMessage{
					ProjectGuid:     project.GUID,
					Type:            common.ConversationTypeAgent,
					AgentRole:       common.AgentDev.Role,
					AgentName:       common.AgentDev.Name,
					Content:         fmt.Sprintf("Story %s 已完成", story.StoryNumber),
					IsMarkdown:      true,
					MarkdownContent: response.Message,
					IsExpanded:      true,
				}
				s.notifyProjectStatusChange(ctx, project, projectMsg, devProjectStage)
			}
		}

		// Epic 完成，更新 Epic 状态
		allStoriesDone := true
		for _, story := range epic.Stories {
			if story.Status != common.CommonStatusDone {
				allStoriesDone = false
				break
			}
		}
		if allStoriesDone {
			epic.Status = common.CommonStatusDone
			if err := s.epicRepo.Update(ctx, epic); err != nil {
				logger.Error("更新 Epic 状态失败", logger.String("error", err.Error()))
			}
			logger.Info("Epic 已完成", logger.String("epic_name", epic.Name))
		}
	}

	// 发送最终完成消息
	devProjectStage.SetStatus(common.CommonStatusDone)
	finalMsg := fmt.Sprintf("MVP Stories 开发完成，共实现 %d 个 Story", developStoryCount)
	if lastResponse != nil {
		finalMsg = lastResponse.Message
	}

	projectMsg := &models.ConversationMessage{
		ProjectGuid:     project.GUID,
		Type:            common.ConversationTypeAgent,
		AgentRole:       common.AgentDev.Role,
		AgentName:       common.AgentDev.Name,
		Content:         "MVP Stories 功能已开发",
		IsMarkdown:      true,
		MarkdownContent: finalMsg,
		IsExpanded:      true,
	}
	s.notifyProjectStatusChange(ctx, project, projectMsg, devProjectStage)

	tasks.UpdateResult(resultWriter, common.CommonStatusInProgress, 60, "MVP Stories 功能已开发")
	return nil
}

// developStoriesFromFiles 从文件方式开发 Stories (fallback)
func (s *projectStageService) developStoriesFromFiles(ctx context.Context,
	project *models.Project, resultWriter *asynq.ResultWriter,
	agentClient *client.AgentClient, devProjectStage *models.DevStage) error {

	req := &agent.ImplementStoryReq{
		ProjectGuid: project.GUID,
		PrdPath:     "docs/PRD.md",
		ArchFolder:  "docs/arch/",
		DbFolder:    "docs/db/",
		ApiFolder:   "docs/api/",
		UxSpecPath:  "docs/ux/ux-spec.md",
		EpicFile:    "docs/stories/",
		StoryFile:   "",
		CliTool:     s.getCliTool(project),
	}

	storyFiles, err := s.fileService.GetRelativeFiles(project.ProjectPath, "docs/stories")
	if err != nil || len(storyFiles) == 0 {
		response, err := agentClient.ImplementStory(ctx, req)
		if err != nil {
			tasks.UpdateResult(resultWriter, common.CommonStatusFailed, 0, "调用 Dev Agent 开发失败: "+err.Error())
			devProjectStage.SetStatus(common.CommonStatusFailed)
			devProjectStage.FailedReason = err.Error()
			s.notifyProjectStatusChange(ctx, project, nil, devProjectStage)
			return err
		}

		projectMsg := &models.ConversationMessage{
			ProjectGuid:     project.GUID,
			Type:            common.ConversationTypeAgent,
			AgentRole:       common.AgentDev.Role,
			AgentName:       common.AgentDev.Name,
			Content:         "项目Story功能已开发",
			IsMarkdown:      true,
			MarkdownContent: response.Message,
			IsExpanded:      true,
		}

		devProjectStage.SetStatus(common.CommonStatusDone)
		s.notifyProjectStatusChange(ctx, project, projectMsg, devProjectStage)

		tasks.UpdateResult(resultWriter, common.CommonStatusInProgress, 60, "项目Story功能已开发")
		return nil
	}

	var response = &tasks.TaskResult{}
	developStoryCount := 0
	bDev := (utils.GetEnvOrDefault("ENVIRONMENT", common.EnvironmentDevelopment) == common.EnvironmentDevelopment)

	// 获取 stories 下的文件，循环开发每个 Story
	for index, storyFile := range storyFiles {
		// development 模式，只开发一个
		if developStoryCount < 1 || !bDev {
			req.StoryFile = storyFile
			// 调用 agents-server 开发 Story 功能
			response, err = agentClient.ImplementStory(ctx, req)
			if err != nil {
				tasks.UpdateResult(resultWriter, common.CommonStatusFailed, 0, "调用 Dev Agent 开发失败: "+err.Error())
				devProjectStage.SetStatus(common.CommonStatusFailed)
				devProjectStage.FailedReason = err.Error()
				s.notifyProjectStatusChange(ctx, project, nil, devProjectStage)
				return err
			}

			developStoryCount += 1
		} else {
			response.Message = "开发需求故事" + storyFile + "已完成"
		}

		if index < len(storyFiles)-1 {
			projectMsg := &models.ConversationMessage{
				ProjectGuid:     project.GUID,
				Type:            common.ConversationTypeAgent,
				AgentRole:       common.AgentDev.Role,
				AgentName:       common.AgentDev.Name,
				Content:         "项目Story功能已开发",
				IsMarkdown:      true,
				MarkdownContent: response.Message,
				IsExpanded:      true,
			}

			s.notifyProjectStatusChange(ctx, project, projectMsg, devProjectStage)
		}
	}

	devProjectStage.SetStatus(common.CommonStatusDone)
	projectMsg := &models.ConversationMessage{
		ProjectGuid:     project.GUID,
		Type:            common.ConversationTypeAgent,
		AgentRole:       common.AgentDev.Role,
		AgentName:       common.AgentDev.Name,
		Content:         "项目Story功能已开发",
		IsMarkdown:      true,
		MarkdownContent: response.Message,
		IsExpanded:      true,
	}
	s.notifyProjectStatusChange(ctx, project, projectMsg, devProjectStage)

	tasks.UpdateResult(resultWriter, common.CommonStatusInProgress, 60, "项目Story功能已开发")
	return nil
}

// fixBugs 修复开发问题
func (s *projectStageService) fixBugs(ctx context.Context,
	project *models.Project, resultWriter *asynq.ResultWriter,
	agentClient *client.AgentClient, devStage *models.DevStage) error {
	var devProjectStage *models.DevStage
	if devStage == nil {
		devProjectStage = models.NewDevStage(project, common.DevStatusFixBug, common.CommonStatusInProgress)
	} else {
		devProjectStage = devStage
		devProjectStage.SetStatus(common.CommonStatusInProgress)
	}

	s.notifyProjectStatusChange(ctx, project, nil, devProjectStage)

	req := &agent.FixBugReq{
		ProjectGuid:    project.GUID,
		BugDescription: "修复开发问题",
		CliTool:        s.getCliTool(project),
	}
	// 调用 agents-server 修复问题
	response, err := agentClient.FixBug(ctx, req)
	if err != nil {
		tasks.UpdateResult(resultWriter, common.CommonStatusFailed, 0, "调用 Dev Agent 修复问题失败: "+err.Error())
		devProjectStage.SetStatus(common.CommonStatusFailed)
		devProjectStage.FailedReason = err.Error()
		s.notifyProjectStatusChange(ctx, project, nil, devProjectStage)
		return err
	}

	projectMsg := &models.ConversationMessage{
		ProjectGuid:     project.GUID,
		Type:            common.ConversationTypeAgent,
		AgentRole:       common.AgentDev.Role,
		AgentName:       common.AgentDev.Name,
		Content:         "项目开发问题已修复",
		IsMarkdown:      true,
		MarkdownContent: response.Message,
		IsExpanded:      true,
	}

	devProjectStage.SetStatus(common.CommonStatusDone)
	s.notifyProjectStatusChange(ctx, project, projectMsg, devProjectStage)

	tasks.UpdateResult(resultWriter, common.CommonStatusInProgress, 65, "项目开发问题已修复")
	return nil
}

// runTests 执行自动测试
func (s *projectStageService) runTests(ctx context.Context,
	project *models.Project, resultWriter *asynq.ResultWriter,
	agentClient *client.AgentClient, devStage *models.DevStage) error {
	var devProjectStage *models.DevStage
	if devStage == nil {
		devProjectStage = models.NewDevStage(project, common.DevStatusRunTest, common.CommonStatusInProgress)
	} else {
		devProjectStage = devStage
		devProjectStage.SetStatus(common.CommonStatusInProgress)
	}

	s.notifyProjectStatusChange(ctx, project, nil, devProjectStage)

	req := &agent.RunTestReq{
		ProjectGuid: project.GUID,
		CliTool:     s.getCliTool(project),
	}
	// 调用 agents-server 执行自动测试
	response, err := agentClient.RunTest(ctx, req)
	if err != nil {
		tasks.UpdateResult(resultWriter, common.CommonStatusFailed, 0, "调用 Dev Agent 测试失败: "+err.Error())
		devProjectStage.SetStatus(common.CommonStatusFailed)
		devProjectStage.FailedReason = err.Error()
		s.notifyProjectStatusChange(ctx, project, nil, devProjectStage)
		return err
	}

	projectMsg := &models.ConversationMessage{
		ProjectGuid:     project.GUID,
		Type:            common.ConversationTypeAgent,
		AgentRole:       common.AgentDev.Role,
		AgentName:       common.AgentDev.Name,
		Content:         "项目自动测试已执行",
		IsMarkdown:      true,
		MarkdownContent: response.Message,
		IsExpanded:      true,
	}

	devProjectStage.SetStatus(common.CommonStatusDone)
	s.notifyProjectStatusChange(ctx, project, projectMsg, devProjectStage)

	tasks.UpdateResult(resultWriter, common.CommonStatusInProgress, 75, "项目自动测试已执行")
	return nil
}

// packageProject 打包项目
func (s *projectStageService) packageProject(ctx context.Context,
	project *models.Project, resultWriter *asynq.ResultWriter,
	agentClient *client.AgentClient, devStage *models.DevStage) error {
	var devProjectStage *models.DevStage
	if devStage == nil {
		devProjectStage = models.NewDevStage(project, common.DevStatusDeploy, common.CommonStatusInProgress)
	} else {
		devProjectStage = devStage
		devProjectStage.SetStatus(common.CommonStatusInProgress)
	}

	s.notifyProjectStatusChange(ctx, project, nil, devProjectStage)

	req := &agent.DeployReq{
		ProjectGuid:   project.GUID,
		Environment:   "dev",
		DeployOptions: map[string]interface{}{},
		CliTool:       s.getCliTool(project),
	}
	// 调用 agents-server 打包部署项目（提交 .gitlab-ci.yml 即可触发 runner）
	response, err := agentClient.Deploy(ctx, req)
	if err != nil {
		tasks.UpdateResult(resultWriter, common.CommonStatusFailed, 0, "调用 Dev Agent 打包失败: "+err.Error())
		devProjectStage.SetStatus(common.CommonStatusFailed)
		devProjectStage.FailedReason = err.Error()
		s.notifyProjectStatusChange(ctx, project, nil, devProjectStage)
		return err
	}

	projectMsg := &models.ConversationMessage{
		ProjectGuid:     project.GUID,
		Type:            common.ConversationTypeAgent,
		AgentRole:       common.AgentDev.Role,
		AgentName:       common.AgentDev.Name,
		Content:         "项目项目已打包部署",
		IsMarkdown:      true,
		MarkdownContent: response.Message,
		IsExpanded:      true,
	}

	// 设置预览 URL
	if project.PreviewUrl == "" {
		project.PreviewUrl = fmt.Sprintf("http://%s.app-maker.localhost", project.GUID)
		if err := s.projectRepo.Update(ctx, project); err != nil {
			logger.Error("更新项目预览URL失败",
				logger.String("error", err.Error()),
				logger.String("projectID", project.ID),
			)
		} else {
			logger.Info("项目预览URL已设置",
				logger.String("projectID", project.ID),
				logger.String("previewUrl", project.PreviewUrl),
			)
			// 通知前端预览URL已设置
			s.webSocketService.NotifyProjectInfoUpdate(ctx, project.GUID, project)
		}
	}

	devProjectStage.SetStatus(common.CommonStatusDone)
	s.notifyProjectStatusChange(ctx, project, projectMsg, devProjectStage)

	tasks.UpdateResult(resultWriter, common.CommonStatusInProgress, 80, "项目项目已打包部署")
	return nil
}

// 与项目中的 Agent 进行对话
func (s *projectStageService) ChatWithAgent(ctx context.Context, req *agent.ChatReq) error {
	// 异步方式
	_, err := s.asyncClient.Enqueue(tasks.NewAgentChatTask(req))
	if err != nil {
		return fmt.Errorf("创建与 Agent 对话任务失败: %w", err)
	}
	return nil
}

// extractMvpEpicsJSON 从 markdown 内容中提取 MVP Epics JSON
func (s *projectStageService) extractMvpEpicsJSON(content string) (*models.MvpEpicsData, error) {
	// 查找 JSON 代码块
	jsonStart := strings.Index(content, "```json")
	if jsonStart == -1 {
		logger.Warn("未找到 JSON 代码块")
		return nil, fmt.Errorf("未找到 JSON 代码块")
	}

	jsonStart += len("```json")
	jsonEnd := strings.Index(content[jsonStart:], "```")
	if jsonEnd == -1 {
		logger.Warn("JSON 代码块未闭合")
		return nil, fmt.Errorf("JSON 代码块未闭合")
	}

	jsonContent := strings.TrimSpace(content[jsonStart : jsonStart+jsonEnd])

	var mvpData models.MvpEpicsData
	if err := json.Unmarshal([]byte(jsonContent), &mvpData); err != nil {
		logger.Error("解析 MVP Epics JSON 失败", logger.String("error", err.Error()))
		return nil, fmt.Errorf("解析 MVP Epics JSON 失败: %w", err)
	}

	logger.Info("成功解析 MVP Epics JSON", logger.Int("epic_count", len(mvpData.MvpEpics)))
	return &mvpData, nil
}

// saveMvpEpics 保存 MVP Epics 到数据库
func (s *projectStageService) saveMvpEpics(ctx context.Context, project *models.Project, mvpData *models.MvpEpicsData) error {
	if mvpData == nil || len(mvpData.MvpEpics) == 0 {
		return fmt.Errorf("MVP Epics 数据为空")
	}

	// 遍历每个 Epic
	for _, epicItem := range mvpData.MvpEpics {
		// 创建 Epic
		epic := &models.Epic{
			ProjectID:     project.ID,
			ProjectGuid:   project.GUID,
			EpicNumber:    epicItem.EpicNumber,
			Name:          epicItem.Name,
			Description:   epicItem.Description,
			Priority:      epicItem.Priority,
			EstimatedDays: epicItem.EstimatedDays,
			Status:        common.CommonStatusPending,
			FilePath:      epicItem.FilePath,
		}

		// 保存 Epic
		if err := s.epicRepo.Create(ctx, epic); err != nil {
			logger.Error("保存 Epic 失败",
				logger.String("epic_name", epic.Name),
				logger.String("error", err.Error()))
			return fmt.Errorf("保存 Epic 失败: %w", err)
		}

		logger.Info("Epic 已保存",
			logger.String("epic_id", epic.ID),
			logger.String("epic_name", epic.Name))

		// 遍历 Epic 下的每个 Story
		for _, storyItem := range epicItem.Stories {
			story := &models.Story{
				EpicID:        epic.ID,
				StoryNumber:   storyItem.StoryNumber,
				Title:         storyItem.Title,
				Description:   storyItem.Description,
				Priority:      storyItem.Priority,
				EstimatedDays: storyItem.EstimatedDays,
				Status:        common.CommonStatusPending,
				FilePath:      epic.FilePath, // Story 的 FilePath 与 Epic 相同
				Depends:       storyItem.Depends,
				Techs:         storyItem.Techs,
			}

			// 保存 Story
			if err := s.storyRepo.Create(ctx, story); err != nil {
				logger.Error("保存 Story 失败",
					logger.String("story_number", story.StoryNumber),
					logger.String("story_title", story.Title),
					logger.String("error", err.Error()))
				return fmt.Errorf("保存 Story 失败: %w", err)
			}

			logger.Info("Story 已保存",
				logger.String("story_id", story.ID),
				logger.String("story_number", story.StoryNumber),
				logger.String("story_title", story.Title))
		}
	}

	logger.Info("所有 MVP Epics 和 Stories 已保存",
		logger.Int("epic_count", len(mvpData.MvpEpics)))
	return nil
}

// generateFrontendPages 生成前端关键页面 (Vibe Coding)
func (s *projectStageService) generateFrontendPages(ctx context.Context,
	project *models.Project, resultWriter *asynq.ResultWriter,
	agentClient *client.AgentClient, devStage *models.DevStage) error {

	var devProjectStage *models.DevStage
	if devStage == nil {
		devProjectStage = models.NewDevStage(project, common.DevStatusGeneratePages, common.CommonStatusInProgress)
	} else {
		devProjectStage = devStage
		devProjectStage.SetStatus(common.CommonStatusInProgress)
	}

	s.notifyProjectStatusChange(ctx, project, nil, devProjectStage)

	// 只在开发模式下执行
	bDev := (utils.GetEnvOrDefault("ENVIRONMENT", common.EnvironmentDevelopment) == common.EnvironmentDevelopment)
	if !bDev {
		logger.Info("生产环境跳过前端页面生成")
		devProjectStage.SetStatus(common.CommonStatusDone)
		s.notifyProjectStatusChange(ctx, project, nil, devProjectStage)
		tasks.UpdateResult(resultWriter, common.CommonStatusInProgress, common.GetDevStageProgress(common.DevStatusGeneratePages), "跳过前端页面生成")
		return nil
	}

	// 检查 page-prompt.md 文件是否存在
	pagePromptRelPath := "docs/ux/page-prompt.md"
	pagePromptFiles, err := s.fileService.GetRelativeFiles(project.ProjectPath, "docs/ux")
	hasPagePrompt := false
	for _, file := range pagePromptFiles {
		if strings.Contains(file, "page-prompt") || strings.Contains(file, "prompt") {
			hasPagePrompt = true
			break
		}
	}

	if !hasPagePrompt {
		logger.Warn("未找到 page-prompt.md 文件，跳过前端页面生成")
		devProjectStage.SetStatus(common.CommonStatusDone)
		s.notifyProjectStatusChange(ctx, project, nil, devProjectStage)
		tasks.UpdateResult(resultWriter, common.CommonStatusInProgress, common.GetDevStageProgress(common.DevStatusGeneratePages), "未找到页面提示词文件")
		return nil
	}

	logger.Info("开始生成前端页面", logger.String("pagePromptPath", pagePromptRelPath))

	// 根据 CLI 类型选择不同的 prompt
	var agentPrompt string
	cliTool := s.getCliTool(project)
	if cliTool == common.CliToolGemini {
		agentPrompt = "@.bmad-core/agents/dev.md"
	} else {
		agentPrompt = "@bmad/dev.mdc"
	}

	// 调用 Dev Agent 生成前端页面
	message := agentPrompt + " 请基于 @docs/ux/page-prompt.md 中的页面设计提示词," +
		"在前端项目 frontend/src/pages/ 目录下生成关键页面组件。" +
		"使用 Vue 3 + TypeScript + Naive UI,遵循现有项目的代码风格和架构。" +
		"只生成 page-prompt.md 中明确定义的页面，不要生成其他页面。" +
		"注意：始终用中文回答我。"

	req := &agent.ChatReq{
		ProjectGuid: project.GUID,
		AgentType:   common.AgentTypeDev,
		Message:     message,
	}

	response, err := agentClient.ChatWithAgent(ctx, req)
	if err != nil {
		logger.Error("生成前端页面失败", logger.String("error", err.Error()))
		tasks.UpdateResult(resultWriter, common.CommonStatusFailed, 0, "生成前端页面失败: "+err.Error())
		devProjectStage.SetStatus(common.CommonStatusFailed)
		devProjectStage.FailedReason = err.Error()
		s.notifyProjectStatusChange(ctx, project, nil, devProjectStage)
		return err
	}

	projectMsg := &models.ConversationMessage{
		ProjectGuid:     project.GUID,
		Type:            common.ConversationTypeAgent,
		AgentRole:       common.AgentDev.Role,
		AgentName:       common.AgentDev.Name,
		Content:         "前端关键页面已生成",
		IsMarkdown:      true,
		MarkdownContent: response.Message,
		IsExpanded:      true,
	}

	devProjectStage.SetStatus(common.CommonStatusDone)
	s.notifyProjectStatusChange(ctx, project, projectMsg, devProjectStage)

	logger.Info("前端页面生成完成")
	tasks.UpdateResult(resultWriter, common.CommonStatusInProgress, common.GetDevStageProgress(common.DevStatusGeneratePages), "前端页面已生成")
	return nil
}
