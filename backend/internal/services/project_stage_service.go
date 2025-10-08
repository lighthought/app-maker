package services

import (
	"context"
	"encoding/json"
	"fmt"
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
	GetProjectStages(ctx context.Context, projectGuid string) ([]*models.DevStage, error)
	ProcessTask(ctx context.Context, task *asynq.Task) error
}

// ProjectStageService 任务执行服务
type projectStageService struct {
	projectRepo      repositories.ProjectRepository
	stageRepo        repositories.StageRepository
	messageRepo      repositories.MessageRepository
	webSocketService WebSocketService
	gitService       GitService
	fileService      FileService
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
		agentsURL:        agentsURL,
	}
}

// GetProjectStages 获取项目开发阶段
func (s *projectStageService) GetProjectStages(ctx context.Context, projectGuid string) ([]*models.DevStage, error) {
	return s.stageRepo.GetByProjectGUID(ctx, projectGuid)
}

// ProcessTask 处理项目任务
func (h *projectStageService) ProcessTask(ctx context.Context, task *asynq.Task) error {
	switch task.Type() {
	case common.TaskTypeProjectDevelopment:
		return h.HandleProjectDevelopmentTask(ctx, task)
	default:
		return fmt.Errorf("unexpected task type %s", task.Type())
	}
}

// HandleProjectDevelopmentTask 处理项目开发任务
func (s *projectStageService) HandleProjectDevelopmentTask(ctx context.Context, t *asynq.Task) error {
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
	agentClient := client.NewAgentClient(s.agentsURL, 10*time.Second)

	// TODO: 检查项目是否已有正在进行中（未完成、未失败）的任务，且任务 ID 和当前不同，不同，则直接跳过。

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
		{common.DevStatusDevelopStory, "开发Story功能", s.developStories},
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

// 统一由这个函数更新项目状态
func (s *projectStageService) notifyProjectStatusChange(ctx context.Context,
	project *models.Project, message *models.ConversationMessage, stage *models.DevStage) {
	if message != nil {
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

	result, err := agentClient.SetupProjectEnvironment(ctx, &agent.SetupProjEnvReq{
		ProjectGuid:     project.GUID,
		GitlabRepoUrl:   project.GitlabRepoURL,
		SetupBmadMethod: true,
		BmadCliType:     "claude",
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

	// TODO: git 拉新代码，通过文件解析 epics 和 stories 这个关键信息
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

// developStories 开发Story功能
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

	req := &agent.ImplementStoryReq{
		ProjectGuid: project.GUID,
		PrdPath:     "docs/PRD.md",
		ArchFolder:  "docs/arch/",
		DbFolder:    "docs/db/",
		ApiFolder:   "docs/api/",
		UxSpecPath:  "docs/ux/ux-spec.md",
		EpicFile:    "docs/stories/",
		StoryFile:   "",
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
		// development 模式，只开发一个，其他的都直接打印结果就可以了
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

	devProjectStage.SetStatus(common.CommonStatusDone)
	s.notifyProjectStatusChange(ctx, project, projectMsg, devProjectStage)

	tasks.UpdateResult(resultWriter, common.CommonStatusInProgress, 80, "项目项目已打包部署")
	return nil
}
