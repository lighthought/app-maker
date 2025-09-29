package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	cfgPkg "autocodeweb-backend/internal/config"
	"autocodeweb-backend/internal/models"
	"autocodeweb-backend/internal/repositories"
	"autocodeweb-backend/internal/utils"
	"autocodeweb-backend/pkg/logger"

	"shared-models/agent"
	"shared-models/client"
	"shared-models/common"

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
	agentsURL        string
}

// NewTaskExecutionService 创建任务执行服务
func NewProjectStageService(
	projectRepo repositories.ProjectRepository,
	stageRepo repositories.StageRepository,
	messageRepo repositories.MessageRepository,
	webSocketService WebSocketService,
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
	case models.TypeProjectDevelopment:
		return h.HandleProjectDevelopmentTask(ctx, task)
	default:
		return fmt.Errorf("unexpected task type %s", task.Type())
	}
}

// HandleProjectDevelopmentTask 处理项目开发任务
func (s *projectStageService) HandleProjectDevelopmentTask(ctx context.Context, t *asynq.Task) error {

	var payload models.ProjectTaskPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}
	resultWriter := t.ResultWriter()
	logger.Info("处理项目开发任务", logger.String("taskID", resultWriter.TaskID()))

	// 更新 pending agents 的 stage 为 done，表示 API 内部这个 async 调用成功，也获取到了合法的数据
	if stagePendingAgents, err := s.stageRepo.UpdateStageToDone(ctx, payload.ProjectID, string(common.DevStatusPendingAgents)); err != nil {
		logger.Error("更新项目阶段失败",
			logger.String("error", err.Error()),
			logger.String("projectID", payload.ProjectID),
			logger.String("projectGuid", payload.ProjectGuid),
		)
	} else {
		if stagePendingAgents != nil {
			s.webSocketService.NotifyProjectStageUpdate(ctx, payload.ProjectGuid, stagePendingAgents)
		}
	}

	project, err := s.projectRepo.GetByGUID(ctx, payload.ProjectGuid)
	if err != nil {
		return fmt.Errorf("获取项目信息失败: %w", err)
	}

	s.executeProjectDevelopment(ctx, project, resultWriter)
	utils.UpdateResult(resultWriter, common.CommandStatusDone, 100, "项目开发任务完成")
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

			project.SetDevStatus(common.DevStage(stage.Name))
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

// executeProjectDevelopment 执行项目开发流程
func (s *projectStageService) executeProjectDevelopment(ctx context.Context,
	project *models.Project, resultWriter *asynq.ResultWriter) {
	logger.Info("开始执行项目开发流程",
		logger.String("projectID", project.ID),
	)

	if s.agentsURL != "" {
		s.agentsURL = utils.GetEnvOrDefault("AGENTS_SERVER_URL", "http://host.docker.internal:8088")
	}
	agentClient := client.NewAgentClient(s.agentsURL, 10*time.Second)

	// 2. 执行开发阶段
	stages := []struct {
		status      common.DevStage
		description string
		executor    func(context.Context, *models.Project, *asynq.ResultWriter, *client.AgentClient) error
	}{
		{common.DevStatusCheckRequirement, "检查需求", s.checkRequirement},
		{common.DevStatusGeneratePRD, "生成PRD文档", s.generatePRD},
		{common.DevStatusDefineUXStandard, "定义UX标准", s.defineUXStandards},
		{common.DevStatusDesignArchitecture, "设计系统架构", s.designArchitecture},
		{common.DevStatusPlanEpicAndStory, "划分Epic和Story", s.planEpicsAndStories},
		{common.DevStatusDefineDataModel, "定义数据模型", s.defineDataModel},
		{common.DevStatusDefineAPI, "定义API接口", s.defineAPIs},
		{common.DevStatusDevelopStory, "开发Story功能", s.developStories},
		{common.DevStatusFixBug, "修复开发问题", s.fixBugs},
		{common.DevStatusRunTest, "执行自动测试", s.runTests},
		{common.DevStatusDeploy, "打包项目", s.packageProject},
	}

	for _, stage := range stages {
		// 更新项目状态
		project.SetDevStatus(stage.status)
		s.projectRepo.Update(ctx, project)

		// 执行阶段
		if err := stage.executor(ctx, project, resultWriter, agentClient); err != nil {
			logger.Error("开发阶段执行失败",
				logger.String("projectID", project.ID),
				logger.String("stage", string(stage.status)),
				logger.String("error", err.Error()),
			)

			// 更新项目状态为失败
			project.SetDevStatus(common.DevStatusFailed)
			s.projectRepo.Update(ctx, project)

			return
		}

	}

	// 开发完成
	project.SetDevStatus(common.DevStatusDone)
	project.Status = common.CommandStatusDone
	s.projectRepo.Update(ctx, project)
	s.webSocketService.NotifyProjectInfoUpdate(ctx, project.GUID, project)

	logger.Info("项目开发流程执行完成",
		logger.String("projectID", project.ID),
	)
}

// checkRequirement 检查需求
func (s *projectStageService) checkRequirement(ctx context.Context,
	project *models.Project, resultWriter *asynq.ResultWriter, agentClient *client.AgentClient) error {
	devProjectStage := models.NewDevStage(project, common.DevStatusCheckRequirement, common.CommandStatusInProgress)
	s.notifyProjectStatusChange(ctx, project, nil, devProjectStage)

	req := &agent.GetProjBriefReq{
		ProjectGuid:  project.GUID,
		Requirements: project.Requirements,
	}

	response, err := agentClient.AnalyseProjectBrief(ctx, req)
	if err != nil {
		utils.UpdateResult(resultWriter, common.CommandStatusFailed, 0, "调用 Analyst Agent 检查需求失败: "+err.Error())
		return err
	}

	projectMsg := &models.ConversationMessage{
		ProjectGuid:     project.GUID,
		Type:            common.ConversationTypeAgent,
		AgentRole:       common.AgentAnalyst.Role,
		AgentName:       common.AgentAnalyst.Name,
		Content:         "项目需求已检查完成",
		IsMarkdown:      false,
		MarkdownContent: response.MarkdownContent,
		IsExpanded:      false,
	}

	devProjectStage.SetStatus(common.CommandStatusDone)
	s.notifyProjectStatusChange(ctx, project, projectMsg, devProjectStage)

	utils.UpdateResult(resultWriter, common.CommandStatusInProgress, 10, "项目需求已检查完成")
	return nil
}

// generatePRD 生成PRD文档
func (s *projectStageService) generatePRD(ctx context.Context,
	project *models.Project, resultWriter *asynq.ResultWriter, agentClient *client.AgentClient) error {
	devProjectStage := models.NewDevStage(project, common.DevStatusGeneratePRD, common.CommandStatusInProgress)
	s.notifyProjectStatusChange(ctx, project, nil, devProjectStage)
	generatePrdReq := &agent.GetPRDReq{
		ProjectGuid:  project.GUID,
		Requirements: project.Requirements,
	}
	// 调用 agents-server 生成 PRD 文档，并提交到 GitLab
	response, err := agentClient.GetPRD(ctx, generatePrdReq)
	if err != nil {
		utils.UpdateResult(resultWriter, common.CommandStatusFailed, 0, "调用 PM Agent 生成 PRD 文档失败: "+err.Error())
		return err
	}

	projectMsg := &models.ConversationMessage{
		ProjectGuid:     project.GUID,
		Type:            common.ConversationTypeAgent,
		AgentRole:       common.AgentPM.Role,
		AgentName:       common.AgentPM.Name,
		Content:         "项目PRD文档已生成",
		IsMarkdown:      true,
		MarkdownContent: response.GetMarkdownContent(),
		IsExpanded:      false,
	}

	devProjectStage.SetStatus(common.CommandStatusDone)
	s.notifyProjectStatusChange(ctx, project, projectMsg, devProjectStage)

	utils.UpdateResult(resultWriter, common.CommandStatusInProgress, 20, "项目PRD文档已生成")
	return nil
}

// defineUXStandards 定义UX标准
func (s *projectStageService) defineUXStandards(ctx context.Context,
	project *models.Project, resultWriter *asynq.ResultWriter, agentClient *client.AgentClient) error {
	devProjectStage := models.NewDevStage(project, common.DevStatusDefineUXStandard, common.CommandStatusInProgress)
	s.notifyProjectStatusChange(ctx, project, nil, devProjectStage)

	req := &agent.GetUXStandardReq{
		ProjectGuid:  project.GUID,
		Requirements: project.Requirements,
		PrdPath:      "docs/PRD.md",
	}
	// 调用 agents-server 定义 UX 标准
	response, err := agentClient.GetUXStandard(ctx, req)
	if err != nil {
		utils.UpdateResult(resultWriter, common.CommandStatusFailed, 0, "调用 UX Agent 失败: "+err.Error())
		return err
	}

	projectMsg := &models.ConversationMessage{
		ProjectGuid:     project.GUID,
		Type:            common.ConversationTypeAgent,
		AgentRole:       common.AgentUXExpert.Role,
		AgentName:       common.AgentUXExpert.Name,
		Content:         "项目UX标准已定义",
		IsMarkdown:      true,
		MarkdownContent: response.GetMarkdownContent(),
		IsExpanded:      false,
	}

	devProjectStage.SetStatus(common.CommandStatusDone)
	s.notifyProjectStatusChange(ctx, project, projectMsg, devProjectStage)

	utils.UpdateResult(resultWriter, common.CommandStatusInProgress, 30, "项目UX标准已定义")
	return nil
}

// designArchitecture 设计系统架构
func (s *projectStageService) designArchitecture(ctx context.Context,
	project *models.Project, resultWriter *asynq.ResultWriter, agentClient *client.AgentClient) error {
	devProjectStage := models.NewDevStage(project, common.DevStatusDesignArchitecture, common.CommandStatusInProgress)
	s.notifyProjectStatusChange(ctx, project, nil, devProjectStage)

	req := &agent.GetArchitectureReq{
		ProjectGuid:             project.GUID,
		PrdPath:                 "docs/PRD.md",
		UxSpecPath:              "docs/ux/ux-spec.md",
		TemplateArchDescription: "Vue.js + Vite 前端，Go + Gin 后端，PostgreSQL 数据库，Redis 缓存，Docker 部署",
	}
	// 调用 agents-server 设计系统架构
	response, err := agentClient.GetArchitecture(ctx, req)
	if err != nil {
		utils.UpdateResult(resultWriter, common.CommandStatusFailed, 0, "调用 Architect Agent 失败: "+err.Error())
		return err
	}

	projectMsg := &models.ConversationMessage{
		ProjectGuid:     project.GUID,
		Type:            common.ConversationTypeAgent,
		AgentRole:       common.AgentArchitect.Role,
		AgentName:       common.AgentArchitect.Name,
		Content:         "项目系统架构已设计",
		IsMarkdown:      true,
		MarkdownContent: response.GetMarkdownContent(),
		IsExpanded:      false,
	}

	devProjectStage.SetStatus(common.CommandStatusDone)
	s.notifyProjectStatusChange(ctx, project, projectMsg, devProjectStage)

	utils.UpdateResult(resultWriter, common.CommandStatusInProgress, 40, "项目系统架构已设计")
	return nil
}

// defineDataModel 定义数据模型
func (s *projectStageService) defineDataModel(ctx context.Context,
	project *models.Project, resultWriter *asynq.ResultWriter, agentClient *client.AgentClient) error {
	devProjectStage := models.NewDevStage(project, common.DevStatusDefineDataModel, common.CommandStatusInProgress)
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
		utils.UpdateResult(resultWriter, common.CommandStatusFailed, 0, "调用 Architect Agent 失败: "+err.Error())
		return err
	}

	projectMsg := &models.ConversationMessage{
		ProjectGuid:     project.GUID,
		Type:            common.ConversationTypeAgent,
		AgentRole:       common.AgentArchitect.Role,
		AgentName:       common.AgentArchitect.Name,
		Content:         "项目数据模型已定义",
		IsMarkdown:      true,
		MarkdownContent: response.GetMarkdownContent(),
		IsExpanded:      false,
	}

	devProjectStage.SetStatus(common.CommandStatusDone)
	s.notifyProjectStatusChange(ctx, project, projectMsg, devProjectStage)

	utils.UpdateResult(resultWriter, common.CommandStatusInProgress, 45, "项目数据模型已定义")
	return nil
}

// defineAPIs 定义API接口
func (s *projectStageService) defineAPIs(ctx context.Context,
	project *models.Project, resultWriter *asynq.ResultWriter, agentClient *client.AgentClient) error {
	devProjectStage := models.NewDevStage(project, common.DevStatusDefineAPI, common.CommandStatusInProgress)
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
		utils.UpdateResult(resultWriter, common.CommandStatusFailed, 0, "调用 Architect Agent 失败: "+err.Error())
		return err
	}

	projectMsg := &models.ConversationMessage{
		ProjectGuid:     project.GUID,
		Type:            common.ConversationTypeAgent,
		AgentRole:       common.AgentArchitect.Role,
		AgentName:       common.AgentArchitect.Name,
		Content:         "项目API接口已定义",
		IsMarkdown:      true,
		MarkdownContent: response.GetMarkdownContent(),
		IsExpanded:      false,
	}

	devProjectStage.SetStatus(common.CommandStatusDone)
	s.notifyProjectStatusChange(ctx, project, projectMsg, devProjectStage)

	utils.UpdateResult(resultWriter, common.CommandStatusInProgress, 50, "项目API接口已定义")
	return nil
}

// planEpicsAndStories 划分Epic和Story
func (s *projectStageService) planEpicsAndStories(ctx context.Context,
	project *models.Project, resultWriter *asynq.ResultWriter, agentClient *client.AgentClient) error {
	devProjectStage := models.NewDevStage(project, common.DevStatusPlanEpicAndStory, common.CommandStatusInProgress)
	s.notifyProjectStatusChange(ctx, project, nil, devProjectStage)

	req := &agent.GetEpicsAndStoriesReq{
		ProjectGuid: project.GUID,
		PrdPath:     "docs/PRD.md",
		ArchFolder:  "docs/arch",
	}
	// 调用 agents-server 划分 Epics 和 Stories
	response, err := agentClient.GetEpicsAndStories(ctx, req)
	if err != nil {
		utils.UpdateResult(resultWriter, common.CommandStatusFailed, 0, "调用 PO Agent 失败: "+err.Error())
		return err
	}

	projectMsg := &models.ConversationMessage{
		ProjectGuid:     project.GUID,
		Type:            common.ConversationTypeAgent,
		AgentRole:       common.AgentPO.Role,
		AgentName:       common.AgentPO.Name,
		Content:         "项目Epic和Story已划分",
		IsMarkdown:      true,
		MarkdownContent: response.GetMarkdownContent(),
		IsExpanded:      false,
	}

	devProjectStage.SetStatus(common.CommandStatusDone)
	s.notifyProjectStatusChange(ctx, project, projectMsg, devProjectStage)

	utils.UpdateResult(resultWriter, common.CommandStatusInProgress, 55, "项目Epic和Story已划分")
	return nil
}

// developStories 开发Story功能
func (s *projectStageService) developStories(ctx context.Context,
	project *models.Project, resultWriter *asynq.ResultWriter, agentClient *client.AgentClient) error {
	devProjectStage := models.NewDevStage(project, common.DevStatusDevelopStory, common.CommandStatusInProgress)
	s.notifyProjectStatusChange(ctx, project, nil, devProjectStage)

	req := &agent.ImplementStoryReq{
		ProjectGuid: project.GUID,
		PrdPath:     "docs/PRD.md",
		ArchFolder:  "docs/arch",
		DbFolder:    "docs/db",
		ApiFolder:   "docs/api",
	}
	// 调用 agents-server 开发 Story 功能
	response, err := agentClient.ImplementStory(ctx, req)
	if err != nil {
		utils.UpdateResult(resultWriter, common.CommandStatusFailed, 0, "调用 Dev Agent 开发失败: "+err.Error())
		return err
	}

	projectMsg := &models.ConversationMessage{
		ProjectGuid:     project.GUID,
		Type:            common.ConversationTypeAgent,
		AgentRole:       common.AgentDev.Role,
		AgentName:       common.AgentDev.Name,
		Content:         "项目Story功能已开发",
		IsMarkdown:      true,
		MarkdownContent: response.GetMarkdownContent(),
		IsExpanded:      false,
	}

	devProjectStage.SetStatus(common.CommandStatusDone)
	s.notifyProjectStatusChange(ctx, project, projectMsg, devProjectStage)

	utils.UpdateResult(resultWriter, common.CommandStatusInProgress, 60, "项目Story功能已开发")
	return nil
}

// fixBugs 修复开发问题
func (s *projectStageService) fixBugs(ctx context.Context,
	project *models.Project, resultWriter *asynq.ResultWriter, agentClient *client.AgentClient) error {
	devProjectStage := models.NewDevStage(project, common.DevStatusFixBug, common.CommandStatusInProgress)
	s.notifyProjectStatusChange(ctx, project, nil, devProjectStage)

	req := &agent.FixBugReq{
		ProjectGuid:    project.GUID,
		BugDescription: "修复开发问题",
	}
	// 调用 agents-server 修复问题
	response, err := agentClient.FixBug(ctx, req)
	if err != nil {
		utils.UpdateResult(resultWriter, common.CommandStatusFailed, 0, "调用 Dev Agent 修复问题失败: "+err.Error())
		return err
	}

	projectMsg := &models.ConversationMessage{
		ProjectGuid:     project.GUID,
		Type:            common.ConversationTypeAgent,
		AgentRole:       common.AgentDev.Role,
		AgentName:       common.AgentDev.Name,
		Content:         "项目开发问题已修复",
		IsMarkdown:      true,
		MarkdownContent: response.GetMarkdownContent(),
		IsExpanded:      false,
	}

	devProjectStage.SetStatus(common.CommandStatusDone)
	s.notifyProjectStatusChange(ctx, project, projectMsg, devProjectStage)

	utils.UpdateResult(resultWriter, common.CommandStatusInProgress, 65, "项目开发问题已修复")
	return nil
}

// runTests 执行自动测试
func (s *projectStageService) runTests(ctx context.Context,
	project *models.Project, resultWriter *asynq.ResultWriter, agentClient *client.AgentClient) error {
	devProjectStage := models.NewDevStage(project, common.DevStatusRunTest, common.CommandStatusInProgress)
	s.notifyProjectStatusChange(ctx, project, nil, devProjectStage)

	req := &agent.RunTestReq{
		ProjectGuid: project.GUID,
	}
	// 调用 agents-server 执行自动测试
	response, err := agentClient.RunTest(ctx, req)
	if err != nil {
		utils.UpdateResult(resultWriter, common.CommandStatusFailed, 0, "调用 Dev Agent 测试失败: "+err.Error())
		return err
	}

	projectMsg := &models.ConversationMessage{
		ProjectGuid:     project.GUID,
		Type:            common.ConversationTypeAgent,
		AgentRole:       common.AgentDev.Role,
		AgentName:       common.AgentDev.Name,
		Content:         "项目自动测试已执行",
		IsMarkdown:      true,
		MarkdownContent: response.GetMarkdownContent(),
		IsExpanded:      false,
	}

	devProjectStage.SetStatus(common.CommandStatusDone)
	s.notifyProjectStatusChange(ctx, project, projectMsg, devProjectStage)

	utils.UpdateResult(resultWriter, common.CommandStatusInProgress, 75, "项目自动测试已执行")
	return nil
}

// packageProject 打包项目
func (s *projectStageService) packageProject(ctx context.Context,
	project *models.Project, resultWriter *asynq.ResultWriter, agentClient *client.AgentClient) error {
	devProjectStage := models.NewDevStage(project, common.DevStatusDeploy, common.CommandStatusInProgress)
	s.notifyProjectStatusChange(ctx, project, nil, devProjectStage)

	req := &agent.DeployReq{
		ProjectGuid:   project.GUID,
		Environment:   "dev",
		DeployOptions: map[string]interface{}{},
	}
	// 调用 agents-server 打包部署项目（提交 .gitlab-ci.yml 即可触发 runner）
	response, err := agentClient.Deploy(ctx, req)
	if err != nil {
		utils.UpdateResult(resultWriter, common.CommandStatusFailed, 0, "调用 Dev Agent 打包失败: "+err.Error())
		return err
	}

	projectMsg := &models.ConversationMessage{
		ProjectGuid:     project.GUID,
		Type:            common.ConversationTypeAgent,
		AgentRole:       common.AgentDev.Role,
		AgentName:       common.AgentDev.Name,
		Content:         "项目项目已打包部署",
		IsMarkdown:      true,
		MarkdownContent: response.GetMarkdownContent(),
		IsExpanded:      false,
	}

	devProjectStage.SetStatus(common.CommandStatusDone)
	s.notifyProjectStatusChange(ctx, project, projectMsg, devProjectStage)

	utils.UpdateResult(resultWriter, common.CommandStatusInProgress, 80, "项目项目已打包部署")
	return nil
}
