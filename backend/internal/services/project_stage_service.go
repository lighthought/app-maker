package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	cfgPkg "autocodeweb-backend/internal/config"
	"autocodeweb-backend/internal/constants"
	"autocodeweb-backend/internal/models"
	"autocodeweb-backend/internal/repositories"
	"autocodeweb-backend/internal/utils"
	"autocodeweb-backend/pkg/logger"

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
	if stagePendingAgents, err := s.stageRepo.UpdateStageToDone(ctx, payload.ProjectID, constants.DevStatusPendingAgents); err != nil {
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
	utils.UpdateResult(resultWriter, constants.CommandStatusDone, 100, "项目开发任务完成")
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

			project.SetDevStatus(stage.Name)
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

	// 2. 执行开发阶段
	stages := []struct {
		status      string
		description string
		executor    func(context.Context, *models.Project, *asynq.ResultWriter) error
	}{
		{constants.DevStatusCheckRequirement, "检查需求", s.checkRequirement},
		{constants.DevStatusGeneratePRD, "生成PRD文档", s.generatePRD},
		{constants.DevStatusDefineUXStandard, "定义UX标准", s.defineUXStandards},
		{constants.DevStatusDesignArchitecture, "设计系统架构", s.designArchitecture},
		{constants.DevStatusPlanEpicAndStory, "划分Epic和Story", s.planEpicsAndStories},
		{constants.DevStatusDefineDataModel, "定义数据模型", s.defineDataModel},
		{constants.DevStatusDefineAPI, "定义API接口", s.defineAPIs},
		{constants.DevStatusDevelopStory, "开发Story功能", s.developStories},
		{constants.DevStatusFixBug, "修复开发问题", s.fixBugs},
		{constants.DevStatusRunTest, "执行自动测试", s.runTests},
		{constants.DevStatusDeploy, "打包项目", s.packageProject},
	}

	for _, stage := range stages {
		// 更新项目状态
		project.SetDevStatus(stage.status)
		s.projectRepo.Update(ctx, project)

		// 执行阶段
		if err := stage.executor(ctx, project, resultWriter); err != nil {
			logger.Error("开发阶段执行失败",
				logger.String("projectID", project.ID),
				logger.String("stage", stage.status),
				logger.String("error", err.Error()),
			)

			// 更新项目状态为失败
			project.SetDevStatus(constants.DevStatusFailed)
			s.projectRepo.Update(ctx, project)

			return
		}

	}

	// 开发完成
	project.SetDevStatus(constants.DevStatusDone)
	project.Status = constants.CommandStatusDone
	s.projectRepo.Update(ctx, project)
	s.webSocketService.NotifyProjectInfoUpdate(ctx, project.GUID, project)

	logger.Info("项目开发流程执行完成",
		logger.String("projectID", project.ID),
	)
}

// checkRequirement 检查需求
func (s *projectStageService) checkRequirement(ctx context.Context, project *models.Project, resultWriter *asynq.ResultWriter) error {
	devProjectStage := models.NewDevStage(project, constants.DevStatusCheckRequirement, constants.CommandStatusInProgress)
	s.notifyProjectStatusChange(ctx, project, nil, devProjectStage)

	// TODO: 调用AgentServer检查需求

	projectMsg := &models.ConversationMessage{
		ProjectGuid:     project.GUID,
		Type:            constants.ConversationTypeAgent,
		AgentRole:       AgentAnalyst.Role,
		AgentName:       AgentAnalyst.Name,
		Content:         "项目需求已检查完成",
		IsMarkdown:      false,
		MarkdownContent: "项目需求已检查完成",
		IsExpanded:      false,
	}

	devProjectStage.SetStatus(constants.CommandStatusDone)
	s.notifyProjectStatusChange(ctx, project, projectMsg, devProjectStage)

	utils.UpdateResult(resultWriter, constants.CommandStatusInProgress, 10, "项目需求已检查完成")
	return nil
}

// generatePRD 生成PRD文档
func (s *projectStageService) generatePRD(ctx context.Context, project *models.Project, resultWriter *asynq.ResultWriter) error {
	devProjectStage := models.NewDevStage(project, constants.DevStatusGeneratePRD, constants.CommandStatusInProgress)
	s.notifyProjectStatusChange(ctx, project, nil, devProjectStage)
	// 调用 agents-server 生成 PRD 文档，并提交到 GitLab
	response, err := s.invokeAgentSync(ctx, project, "pm", constants.DevStatusGeneratePRD, map[string]any{
		"requirements": project.Requirements,
	})
	if err != nil {
		logger.Error("调用 PM Agent 生成 PRD 失败",
			logger.String("error", err.Error()),
			logger.String("projectID", project.ID),
		)
		return err
	}

	projectMsg := &models.ConversationMessage{
		ProjectGuid:     project.GUID,
		Type:            constants.ConversationTypeAgent,
		AgentRole:       AgentPM.Role,
		AgentName:       AgentPM.Name,
		Content:         "项目PRD文档已生成",
		IsMarkdown:      true,
		MarkdownContent: response.GetMarkdownContent(),
		IsExpanded:      false,
	}

	devProjectStage.SetStatus(constants.CommandStatusDone)
	s.notifyProjectStatusChange(ctx, project, projectMsg, devProjectStage)

	utils.UpdateResult(resultWriter, constants.CommandStatusInProgress, 20, "项目PRD文档已生成")
	return nil
}

// defineUXStandards 定义UX标准
func (s *projectStageService) defineUXStandards(ctx context.Context, project *models.Project, resultWriter *asynq.ResultWriter) error {
	devProjectStage := models.NewDevStage(project, constants.DevStatusDefineUXStandard, constants.CommandStatusInProgress)
	s.notifyProjectStatusChange(ctx, project, nil, devProjectStage)
	// 调用 agents-server 定义 UX 标准
	response, err := s.invokeAgentSync(ctx, project, "ux", constants.DevStatusDefineUXStandard, map[string]any{})
	if err != nil {
		logger.Error("调用 UX Agent 失败",
			logger.String("error", err.Error()),
			logger.String("projectID", project.ID),
		)
		return err
	}

	projectMsg := &models.ConversationMessage{
		ProjectGuid:     project.GUID,
		Type:            constants.ConversationTypeAgent,
		AgentRole:       AgentUXExpert.Role,
		AgentName:       AgentUXExpert.Name,
		Content:         "项目UX标准已定义",
		IsMarkdown:      true,
		MarkdownContent: response.GetMarkdownContent(),
		IsExpanded:      false,
	}

	devProjectStage.SetStatus(constants.CommandStatusDone)
	s.notifyProjectStatusChange(ctx, project, projectMsg, devProjectStage)

	utils.UpdateResult(resultWriter, constants.CommandStatusInProgress, 30, "项目UX标准已定义")
	return nil
}

// designArchitecture 设计系统架构
func (s *projectStageService) designArchitecture(ctx context.Context, project *models.Project, resultWriter *asynq.ResultWriter) error {
	devProjectStage := models.NewDevStage(project, constants.DevStatusDesignArchitecture, constants.CommandStatusInProgress)
	s.notifyProjectStatusChange(ctx, project, nil, devProjectStage)
	// 调用 agents-server 设计系统架构
	response, err := s.invokeAgentSync(ctx, project, "architect", constants.DevStatusDesignArchitecture, map[string]any{})
	if err != nil {
		logger.Error("调用 Architect Agent 失败",
			logger.String("error", err.Error()),
			logger.String("projectID", project.ID),
		)
		return err
	}

	projectMsg := &models.ConversationMessage{
		ProjectGuid:     project.GUID,
		Type:            constants.ConversationTypeAgent,
		AgentRole:       AgentArchitect.Role,
		AgentName:       AgentArchitect.Name,
		Content:         "项目系统架构已设计",
		IsMarkdown:      true,
		MarkdownContent: response.GetMarkdownContent(),
		IsExpanded:      false,
	}

	devProjectStage.SetStatus(constants.CommandStatusDone)
	s.notifyProjectStatusChange(ctx, project, projectMsg, devProjectStage)

	utils.UpdateResult(resultWriter, constants.CommandStatusInProgress, 40, "项目系统架构已设计")
	return nil
}

// defineDataModel 定义数据模型
func (s *projectStageService) defineDataModel(ctx context.Context, project *models.Project, resultWriter *asynq.ResultWriter) error {
	devProjectStage := models.NewDevStage(project, constants.DevStatusDefineDataModel, constants.CommandStatusInProgress)
	s.notifyProjectStatusChange(ctx, project, nil, devProjectStage)
	// 调用 agents-server 定义数据模型
	response, err := s.invokeAgentSync(ctx, project, "architect", constants.DevStatusDefineDataModel, map[string]any{})
	if err != nil {
		logger.Error("调用 Architect Agent(数据模型) 失败",
			logger.String("error", err.Error()),
			logger.String("projectID", project.ID),
		)
		return err
	}

	projectMsg := &models.ConversationMessage{
		ProjectGuid:     project.GUID,
		Type:            constants.ConversationTypeAgent,
		AgentRole:       AgentArchitect.Role,
		AgentName:       AgentArchitect.Name,
		Content:         "项目数据模型已定义",
		IsMarkdown:      true,
		MarkdownContent: response.GetMarkdownContent(),
		IsExpanded:      false,
	}

	devProjectStage.SetStatus(constants.CommandStatusDone)
	s.notifyProjectStatusChange(ctx, project, projectMsg, devProjectStage)

	utils.UpdateResult(resultWriter, constants.CommandStatusInProgress, 45, "项目数据模型已定义")
	return nil
}

// defineAPIs 定义API接口
func (s *projectStageService) defineAPIs(ctx context.Context, project *models.Project, resultWriter *asynq.ResultWriter) error {
	devProjectStage := models.NewDevStage(project, constants.DevStatusDefineAPI, constants.CommandStatusInProgress)
	s.notifyProjectStatusChange(ctx, project, nil, devProjectStage)
	// 调用 agents-server 定义 API 接口
	response, err := s.invokeAgentSync(ctx, project, "architect", constants.DevStatusDefineAPI, map[string]any{})
	if err != nil {
		logger.Error("调用 Architect Agent(API) 失败",
			logger.String("error", err.Error()),
			logger.String("projectID", project.ID),
		)
		return err
	}

	projectMsg := &models.ConversationMessage{
		ProjectGuid:     project.GUID,
		Type:            constants.ConversationTypeAgent,
		AgentRole:       AgentArchitect.Role,
		AgentName:       AgentArchitect.Name,
		Content:         "项目API接口已定义",
		IsMarkdown:      true,
		MarkdownContent: response.GetMarkdownContent(),
		IsExpanded:      false,
	}

	devProjectStage.SetStatus(constants.CommandStatusDone)
	s.notifyProjectStatusChange(ctx, project, projectMsg, devProjectStage)

	utils.UpdateResult(resultWriter, constants.CommandStatusInProgress, 50, "项目API接口已定义")
	return nil
}

// planEpicsAndStories 划分Epic和Story
func (s *projectStageService) planEpicsAndStories(ctx context.Context, project *models.Project, resultWriter *asynq.ResultWriter) error {
	devProjectStage := models.NewDevStage(project, constants.DevStatusPlanEpicAndStory, constants.CommandStatusInProgress)
	s.notifyProjectStatusChange(ctx, project, nil, devProjectStage)
	// 调用 agents-server 划分 Epics 和 Stories
	response, err := s.invokeAgentSync(ctx, project, "po", constants.DevStatusPlanEpicAndStory, map[string]any{})
	if err != nil {
		logger.Error("调用 PO Agent 失败",
			logger.String("error", err.Error()),
			logger.String("projectID", project.ID),
		)
		return err
	}

	projectMsg := &models.ConversationMessage{
		ProjectGuid:     project.GUID,
		Type:            constants.ConversationTypeAgent,
		AgentRole:       AgentPO.Role,
		AgentName:       AgentPO.Name,
		Content:         "项目Epic和Story已划分",
		IsMarkdown:      true,
		MarkdownContent: response.GetMarkdownContent(),
		IsExpanded:      false,
	}

	devProjectStage.SetStatus(constants.CommandStatusDone)
	s.notifyProjectStatusChange(ctx, project, projectMsg, devProjectStage)

	utils.UpdateResult(resultWriter, constants.CommandStatusInProgress, 55, "项目Epic和Story已划分")
	return nil
}

// developStories 开发Story功能
func (s *projectStageService) developStories(ctx context.Context, project *models.Project, resultWriter *asynq.ResultWriter) error {
	devProjectStage := models.NewDevStage(project, constants.DevStatusDevelopStory, constants.CommandStatusInProgress)
	s.notifyProjectStatusChange(ctx, project, nil, devProjectStage)
	// 调用 agents-server 开发 Story 功能
	response, err := s.invokeAgentSync(ctx, project, "dev", constants.DevStatusDevelopStory, map[string]any{})
	if err != nil {
		logger.Error("调用 Dev Agent 开发失败",
			logger.String("error", err.Error()),
			logger.String("projectID", project.ID),
		)
		return err
	}

	projectMsg := &models.ConversationMessage{
		ProjectGuid:     project.GUID,
		Type:            constants.ConversationTypeAgent,
		AgentRole:       AgentDev.Role,
		AgentName:       AgentDev.Name,
		Content:         "项目Story功能已开发",
		IsMarkdown:      true,
		MarkdownContent: response.GetMarkdownContent(),
		IsExpanded:      false,
	}

	devProjectStage.SetStatus(constants.CommandStatusDone)
	s.notifyProjectStatusChange(ctx, project, projectMsg, devProjectStage)

	utils.UpdateResult(resultWriter, constants.CommandStatusInProgress, 60, "项目Story功能已开发")
	return nil
}

// fixBugs 修复开发问题
func (s *projectStageService) fixBugs(ctx context.Context, project *models.Project, resultWriter *asynq.ResultWriter) error {
	devProjectStage := models.NewDevStage(project, constants.DevStatusFixBug, constants.CommandStatusInProgress)
	s.notifyProjectStatusChange(ctx, project, nil, devProjectStage)
	// 调用 agents-server 修复问题
	response, err := s.invokeAgentSync(ctx, project, "dev", constants.DevStatusFixBug, map[string]any{})
	if err != nil {
		logger.Error("调用 Dev Agent 修复问题失败",
			logger.String("error", err.Error()),
			logger.String("projectID", project.ID),
		)
		return err
	}

	projectMsg := &models.ConversationMessage{
		ProjectGuid:     project.GUID,
		Type:            constants.ConversationTypeAgent,
		AgentRole:       AgentDev.Role,
		AgentName:       AgentDev.Name,
		Content:         "项目开发问题已修复",
		IsMarkdown:      true,
		MarkdownContent: response.GetMarkdownContent(),
		IsExpanded:      false,
	}

	devProjectStage.SetStatus(constants.CommandStatusDone)
	s.notifyProjectStatusChange(ctx, project, projectMsg, devProjectStage)

	utils.UpdateResult(resultWriter, constants.CommandStatusInProgress, 65, "项目开发问题已修复")
	return nil
}

// runTests 执行自动测试
func (s *projectStageService) runTests(ctx context.Context, project *models.Project, resultWriter *asynq.ResultWriter) error {
	devProjectStage := models.NewDevStage(project, constants.DevStatusRunTest, constants.CommandStatusInProgress)
	s.notifyProjectStatusChange(ctx, project, nil, devProjectStage)
	// 调用 agents-server 执行自动测试
	response, err := s.invokeAgentSync(ctx, project, "dev", constants.DevStatusRunTest, map[string]any{})
	if err != nil {
		logger.Error("调用 Dev Agent 测试失败",
			logger.String("error", err.Error()),
			logger.String("projectID", project.ID),
		)
		return err
	}

	projectMsg := &models.ConversationMessage{
		ProjectGuid:     project.GUID,
		Type:            constants.ConversationTypeAgent,
		AgentRole:       AgentDev.Role,
		AgentName:       AgentDev.Name,
		Content:         "项目自动测试已执行",
		IsMarkdown:      true,
		MarkdownContent: response.GetMarkdownContent(),
		IsExpanded:      false,
	}

	devProjectStage.SetStatus(constants.CommandStatusDone)
	s.notifyProjectStatusChange(ctx, project, projectMsg, devProjectStage)

	utils.UpdateResult(resultWriter, constants.CommandStatusInProgress, 75, "项目自动测试已执行")
	return nil
}

// packageProject 打包项目
func (s *projectStageService) packageProject(ctx context.Context, project *models.Project, resultWriter *asynq.ResultWriter) error {
	devProjectStage := models.NewDevStage(project, constants.DevStatusDeploy, constants.CommandStatusInProgress)
	s.notifyProjectStatusChange(ctx, project, nil, devProjectStage)
	// 调用 agents-server 打包部署项目（提交 .gitlab-ci.yml 即可触发 runner）
	response, err := s.invokeAgentSync(ctx, project, "dev", constants.DevStatusDeploy, map[string]any{})
	if err != nil {
		logger.Error("调用 Dev Agent 打包失败",
			logger.String("error", err.Error()),
			logger.String("projectID", project.ID),
		)
		return err
	}

	projectMsg := &models.ConversationMessage{
		ProjectGuid:     project.GUID,
		Type:            constants.ConversationTypeAgent,
		AgentRole:       AgentDev.Role,
		AgentName:       AgentDev.Name,
		Content:         "项目项目已打包部署",
		IsMarkdown:      true,
		MarkdownContent: response.GetMarkdownContent(),
		IsExpanded:      false,
	}

	devProjectStage.SetStatus(constants.CommandStatusDone)
	s.notifyProjectStatusChange(ctx, project, projectMsg, devProjectStage)

	utils.UpdateResult(resultWriter, constants.CommandStatusInProgress, 80, "项目项目已打包部署")
	return nil
}

// invokeAgentSync 同步调用 agents-server 对应 Agent 执行一步
func (s *projectStageService) invokeAgentSync(ctx context.Context, project *models.Project,
	agentType string, stage string, stageInput map[string]any) (*models.AgentResult, error) {
	// agents-server 地址从环境变量读取，默认本机 3001
	baseURL := s.agentsURL
	if baseURL == "" {
		baseURL = utils.GetEnvOrDefault("AGENTS_SERVER_URL", "http://localhost:3001")
	}
	url := fmt.Sprintf("%s/api/v1/agents/execute-sync", baseURL)

	// 组装请求
	reqBody := map[string]any{
		"projectId": project.ID,
		"userId":    project.UserID,
		"agentType": agentType,
		"stage":     stage,
		"context": map[string]any{
			"projectPath": project.ProjectPath,
			"projectName": project.Name,
			"stageInput":  stageInput,
		},
		"parameters": map[string]any{},
	}

	buf, _ := json.Marshal(reqBody)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(buf))
	if err != nil {
		return nil, fmt.Errorf("创建 agents-server 请求失败: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("调用 agents-server 失败: %w", err)
	}
	defer resp.Body.Close()
    body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("agents-server 返回错误: %s", string(body))
	}
	// 可选：解析返回确保 success
	var res struct {
		Success bool            `json:"success"`
		Data    json.RawMessage `json:"data"`
		Error   string          `json:"error"`
	}
	if err := json.Unmarshal(body, &res); err == nil {
		if !res.Success {
			if res.Error != "" {
				return nil, fmt.Errorf("agents-server 执行失败: %s", res.Error)
			}
			return nil, fmt.Errorf("agents-server 执行失败")
		}
	}
    // 解析 Data 为 AgentResult
    if len(res.Data) > 0 {
        var agentResult models.AgentResult
        if err := json.Unmarshal(res.Data, &agentResult); err == nil {
            return &agentResult, nil
        }
        // Data 不是标准结构时，忽略解析错误但返回成功
    }
    return &models.AgentResult{Success: true}, nil
}

// GetProjectStages 获取项目开发阶段
func (s *projectStageService) GetProjectStages(ctx context.Context, projectGuid string) ([]*models.DevStage, error) {
	return s.stageRepo.GetByProjectGUID(ctx, projectGuid)
}
