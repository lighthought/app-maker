package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/lighthought/app-maker/shared-models/common"
	"github.com/lighthought/app-maker/shared-models/logger"
	"github.com/lighthought/app-maker/shared-models/tasks"
	"github.com/lighthought/app-maker/shared-models/utils"

	"github.com/lighthought/app-maker/backend/internal/config"
	"github.com/lighthought/app-maker/backend/internal/models"
	"github.com/lighthought/app-maker/backend/internal/repositories"

	"github.com/hibiken/asynq"
)

// ProjectService 项目服务接口
type ProjectService interface {
	// 基础CRUD操作
	CreateProject(ctx context.Context, req *models.CreateProjectRequest, userID string) (*models.ProjectInfo, error)
	GetProject(ctx context.Context, projectGuid, userID string) (*models.ProjectInfo, error)
	UpdateProject(ctx context.Context, projectGuid string, req *models.UpdateProjectRequest, userID string) (*models.ProjectInfo, error)
	DeleteProject(ctx context.Context, projectGuid, userID string) error
	ListProjects(ctx context.Context, req *models.ProjectListRequest, userID string) (*common.PaginationResponse, error)

	// 用户项目管理
	GetUserProjects(ctx context.Context, userID string, req *models.ProjectListRequest) (*common.PaginationResponse, error)

	// 检查项目访问权限
	CheckProjectAccess(ctx context.Context, projectGuid, userID string) (*models.Project, error)

	// 通过项目ID获取项目（内部使用）
	GetProjectByID(ctx context.Context, projectID string) (*models.Project, error)

	// 处理任务
	ProcessTask(ctx context.Context, task *asynq.Task) error
}

const (
	MESSAGE_FAILED_INSERT_PROJECT_STAGE = "插入项目阶段失败"
	MESSAGE_FAILED_SAVE_MESSAGE         = "failed to save project message"
	MESSAGE_FAILED_COMMIT_TO_GITLAB     = "failed to commit project code to GitLab"
)

// projectService 项目服务实现
type projectService struct {
	repositories *repositories.Repository

	templateService    ProjectTemplateService
	commonService      ProjectCommonService
	gitService         GitService
	asyncClientService AsyncClientService

	config *config.Config
}

// NewProjectService 创建项目服务实例
func NewProjectService(
	repositories *repositories.Repository,
	templateService ProjectTemplateService,
	commonService ProjectCommonService,
	gitService GitService,
	asyncClientService AsyncClientService,
	config *config.Config,
) ProjectService {
	return &projectService{
		repositories:       repositories,
		templateService:    templateService,
		commonService:      commonService,
		gitService:         gitService,
		asyncClientService: asyncClientService,
		config:             config,
	}
}

// 更新项目网络设置
func (s *projectService) updateProjectNetworkSetting(ctx context.Context, project *models.Project) error {
	// 自动获取可用端口
	ports, err := s.repositories.ProjectRepo.GetNextAvailablePorts(ctx)
	if err != nil {
		logger.Error("获取可用端口失败",
			logger.String("error", err.Error()),
		)
		return fmt.Errorf("failed to get available ports: %w", err)
	}

	logger.Info("自动分配端口",
		logger.Int("backendPort", ports.BackendPort),
		logger.Int("frontendPort", ports.FrontendPort),
		logger.Int("redisPort", ports.RedisPort),
		logger.Int("postgresPort", ports.PostgresPort),
	)

	project.BackendPort = ports.BackendPort
	project.FrontendPort = ports.FrontendPort
	project.RedisPort = ports.RedisPort
	project.PostgresPort = ports.PostgresPort

	if project.Subnetwork == "" {
		project.Subnetwork = "172.20.0.0/16"
	}
	return nil
}

// 生成项目配置
func GenerateProjectConfig(requirements string, projectConfig *models.Project) {
	// 设置项目配置
	projectConfig.Name = common.DefaultProjectName
	projectConfig.Description = "这是一个新的项目"
	projectConfig.Requirements = requirements
	projectConfig.ApiBaseUrl = common.DefaultApiPrefix

	// 生成密码
	passwordUtils := utils.NewPasswordUtils()
	projectConfig.AppSecretKey = passwordUtils.GenerateRandomPassword("app")
	projectConfig.RedisPassword = passwordUtils.GenerateRandomPassword("redis")
	projectConfig.JwtSecretKey = passwordUtils.GenerateRandomPassword("jwt")
	projectConfig.DatabasePassword = passwordUtils.GenerateRandomPassword("database")
	projectConfig.Subnetwork = "172.20.0.0/16"

	logger.Info("项目配置生成成功",
		logger.String("projectName", projectConfig.Name),
		logger.String("projectDescription", projectConfig.Description),
	)
}

// CreateProject 创建项目
func (s *projectService) CreateProject(ctx context.Context, req *models.CreateProjectRequest, userID string) (*models.ProjectInfo, error) {
	logger.Info("开始创建项目",
		logger.String("userID", userID),
		logger.String("requirements", req.Requirements),
	)

	newProject := models.GetDefaultProject(userID, req.Requirements)
	if err := s.repositories.ProjectRepo.Create(ctx, newProject); err != nil {
		logger.Error("保存项目到数据库失败",
			logger.String("error", err.Error()),
			logger.String("projectID", newProject.ID),
		)
		return nil, fmt.Errorf("failed to create project: %w", err)
	}
	logger.Info("数据库新建项目成功", logger.String("projectID", newProject.ID))

	// 替换为最终的项目路径
	newProject.ProjectPath = utils.GetProjectPath(userID, newProject.GUID)
	logger.Info("生成项目路径", logger.String("projectPath", newProject.ProjectPath))

	// 自动生成项目配置信息和密码信息
	GenerateProjectConfig(req.Requirements, newProject)

	// 更新项目网络设置
	if err := s.updateProjectNetworkSetting(ctx, newProject); err != nil {
		logger.Error("更新项目网络设置失败",
			logger.String("error", err.Error()),
			logger.String("projectID", newProject.ID),
		)
		return nil, fmt.Errorf("failed to update project network setting: %w", err)
	}

	stage, err := s.commonService.CreateAndNotifyProjectStage(ctx, newProject, common.DevStatusInitializing)
	if err != nil {
		return nil, fmt.Errorf("failed to create project init stage: %w", err)
	}

	// asynq 异步调用初始化项目流程
	taskID, err := s.asyncClientService.EnqueueProjectInitTask(newProject.ID, newProject.GUID, newProject.ProjectPath)
	if err != nil {
		s.commonService.UpdateStageStatus(ctx, stage, common.CommonStatusFailed, "failed to create project init task")
		return nil, fmt.Errorf("failed to create project init task: %s", err.Error())
	}

	newProject.CurrentTaskID = taskID

	// 更新项目
	logger.Info("保存项目到数据库")
	if err := s.repositories.ProjectRepo.Update(ctx, newProject); err != nil {
		logger.Error("保存项目到数据库失败",
			logger.String("error", err.Error()),
			logger.String("projectID", newProject.ID),
		)
		return nil, fmt.Errorf("failed to create project: %w", err)
	}
	logger.Info("项目保存成功", logger.String("projectID", newProject.ID))

	projectInfo, err := s.GetProject(ctx, newProject.GUID, userID)
	if err != nil {
		logger.Error("获取项目信息失败",
			logger.String("error", err.Error()),
			logger.String("projectID", newProject.ID),
		)
		return nil, err
	}

	userMsg := models.NewUserMessage(newProject)
	s.commonService.CreateAndNotifyMessage(ctx, newProject.GUID, userMsg)
	return projectInfo, nil
}

// ProcessTask 处理任务
func (s *projectService) ProcessTask(ctx context.Context, task *asynq.Task) error {
	switch task.Type() {
	// 项目初始化任务
	case common.TaskTypeProjectInit:
		return s.HandleProjectInitTask(ctx, task)
	default:
		return fmt.Errorf("unexpected task type %s", task.Type())
	}
}

// reportTaskAndStageError 报告任务和阶段错误
func (s *projectService) reportTaskAndStageError(ctx context.Context,
	resultWriter *asynq.ResultWriter, devStage *models.DevStage, taskID, projectGuid, errMsg string) {
	if resultWriter != nil {
		tasks.UpdateResult(resultWriter, common.CommonStatusFailed, 0, errMsg)
	}

	if devStage != nil {
		s.commonService.UpdateStageStatus(ctx, devStage, common.CommonStatusFailed, errMsg)
	}

	logger.Error("报告任务和阶段错误：",
		logger.String("error", errMsg),
		logger.String("taskID", taskID),
		logger.String("projectGUID", projectGuid),
	)
}

// updateProjectToEnvironmentStage 更新项目阶段
func (s *projectService) updateProjectToEnvironmentStage(ctx context.Context, projectID, taskID string) (*models.Project, *models.DevStage) {
	project, err := s.repositories.ProjectRepo.GetByID(ctx, projectID)
	if err != nil {
		logger.Error("获取项目信息失败",
			logger.String("error", err.Error()),
			logger.String("projectID", projectID),
		)
		return nil, nil
	}

	// 更新 initializing 的 stage 为 done，表示 API 内部这个 async 调用成功，也获取到了合法的数据
	stage, isDone, err := s.commonService.CreateOrUpdateStage(ctx, project, taskID, project.GUID, string(common.DevStatusInitializing))
	if err != nil {
		logger.Error("更新项目阶段失败",
			logger.String("error", err.Error()),
			logger.String("projectID", projectID),
		)
	}
	if !isDone { // 更新初始化阶段为已完成
		s.commonService.UpdateStageStatus(ctx, stage, common.CommonStatusDone, "")
	}

	stageEnvironment, _, err := s.commonService.CreateOrUpdateStage(ctx, project, taskID, project.GUID, string(common.DevStatusSetupEnvironment))
	if err != nil {
		logger.Error("创建或更新项目阶段失败",
			logger.String("error", err.Error()),
			logger.String("projectID", projectID),
		)
		return nil, nil
	}

	s.commonService.UpdateProjectToStage(ctx, project, taskID, string(common.DevStatusSetupEnvironment))

	return project, stageEnvironment
}

// updateProjectNameAndBrief 更新项目名和描述
func (s *projectService) updateProjectNameAndBrief(ctx context.Context, project *models.Project) error {
	if project == nil {
		logger.Error("invalid project parameters")
		return fmt.Errorf("invalid project parameters")
	}

	// 已经生成过，跳过
	if project.Name != "" && project.Name != common.DefaultProjectName && project.Description != "" {
		logger.Info("project name and description already exist, skip generation")
		return nil
	}

	logger.Info("1. start generating project name and description",
		logger.String("projectID", project.ID),
		logger.String("requirements", project.Requirements),
	)

	// 通过 utils 调用 ollama 生成项目名和描述
	summary, err := utils.GenerateProjectSummary(project.Requirements)
	if err != nil {
		return fmt.Errorf("failed to generate project name and description: %s", err.Error())
	}

	projectMsg := &models.ConversationMessage{
		ProjectGuid:     project.GUID,
		Type:            common.ConversationTypeSystem,
		AgentRole:       common.AgentAnalyst.Role,
		AgentName:       common.AgentAnalyst.Name,
		Content:         "项目简介已生成",
		IsMarkdown:      true,
		MarkdownContent: "* 项目名称：" + summary.Title + ",\n* 项目简介：" + summary.Content,
		IsExpanded:      true,
	}

	s.commonService.CreateAndNotifyMessage(ctx, project.GUID, projectMsg)

	project.Name = strings.ToLower(summary.Title)
	project.Description = summary.Content
	logger.Info("project name and description generated successfully",
		logger.String("projectID", project.ID),
		logger.String("projectName", project.Name),
		logger.String("projectDescription", project.Description),
	)

	s.commonService.UpdateAndNotifyProjectInfo(ctx, project)
	return nil
}

// initProjectTemplate 初始化项目模板
func (s *projectService) initProjectTemplate(ctx context.Context, project *models.Project) error {
	logger.Info("2. start initializing project template", logger.String("projectID", project.ID), logger.String("projectPath", project.ProjectPath))

	// 已经初始化过，跳过
	if project.ProjectPath != "" && utils.IsDirectoryExists(project.ProjectPath) {
		logger.Info("project template already initialized, skip initialization")
		return nil
	}

	if err := s.templateService.InitializeProject(ctx, project); err != nil {
		// 模板初始化失败不影响项目创建，但记录错误
		logger.Error("failed to initialize project template", logger.String("error", err.Error()),
			logger.String("projectID", project.ID), logger.String("projectPath", project.ProjectPath),
		)
		return err
	}

	projectMsg := &models.ConversationMessage{
		ProjectGuid:     project.GUID,
		Type:            common.ConversationTypeSystem,
		AgentRole:       common.AgentDev.Role,
		AgentName:       common.AgentDev.Name,
		Content:         "项目模板初始化成功",
		IsMarkdown:      true,
		MarkdownContent: "* 项目GUID：" + project.GUID + ",\n* 项目名称：" + project.Name + ",\n* 项目路径：" + project.ProjectPath,
		IsExpanded:      true,
	}

	s.commonService.CreateAndNotifyMessage(ctx, project.GUID, projectMsg)
	logger.Info("project template initialized successfully", logger.String("projectID", project.ID))
	return nil
}

// commitProject 提交代码到GitLab
func (s *projectService) commitProject(ctx context.Context, project *models.Project) error {
	if project.GitlabRepoURL != "" { // 已经提交过，跳过
		logger.Info("project code already committed to GitLab, skip commit")
		return nil
	}

	if err := s.commitProjectToGit(ctx, project); err != nil {
		logger.Error(MESSAGE_FAILED_COMMIT_TO_GITLAB, logger.String("error", err.Error()), logger.String("projectID", project.ID))
		return fmt.Errorf("%s:%s", MESSAGE_FAILED_COMMIT_TO_GITLAB, err.Error())
	}
	if project.GitlabRepoURL == "" {
		return fmt.Errorf("GitLab repository URL is empty, projectGUID: %s", project.GUID)
	}

	projectMsg := &models.ConversationMessage{
		ProjectGuid:     project.GUID,
		Type:            common.ConversationTypeSystem,
		AgentRole:       common.AgentDev.Role,
		AgentName:       common.AgentDev.Name,
		Content:         "项目代码已成功提交到GitLab",
		IsMarkdown:      true,
		MarkdownContent: "* 项目GUID：" + project.GUID + ", \n* 项目名称：" + project.Name + ",\n* GitLab仓库路径：" + project.GitlabRepoURL,
		IsExpanded:      true,
	}

	s.commonService.CreateAndNotifyMessage(ctx, project.GUID, projectMsg)
	logger.Info("project code committed to GitLab successfully", logger.String("projectID", project.ID))
	return nil
}

// startDevelopingStage 调用AgentServer
func (s *projectService) startDevelopingStage(ctx context.Context, project *models.Project) error {
	//异步创建项目开发任务
	taskID, err := s.asyncClientService.EnqueueProjectStageTask(false, project.GUID, string(common.DevStatusSetupAgents))
	if err != nil {
		logger.Error("failed to create waiting agents task",
			logger.String("error", err.Error()),
			logger.String("projectID", project.ID),
		)
		return fmt.Errorf("failed to create developing stage task: %s", err.Error())
	}
	projectMsg := &models.ConversationMessage{
		ProjectGuid: project.GUID,
		Type:        common.ConversationTypeSystem,
		AgentRole:   common.AgentPM.Role,
		AgentName:   common.AgentPM.Name,
		Content:     "项目创建完成，开发流程已启动",
		IsMarkdown:  true,
		MarkdownContent: "```json\n{\nguid\": \"" + project.GUID +
			"\",\n\"name\": \"" + project.Name +
			"\",\n\"path\":\"" + project.ProjectPath +
			"\",\n \"taskID\": \"" + taskID + "\"\n}\n```",
		IsExpanded: true,
	}

	if err := s.commonService.CreateAndNotifyMessage(ctx, project.GUID, projectMsg); err != nil {
		logger.Error("创建并通知消息失败", logger.String("error", err.Error()), logger.String("projectID", project.ID))
	}

	project.Status = common.CommonStatusInProgress
	project.CurrentTaskID = taskID

	s.commonService.UpdateAndNotifyProjectInfo(ctx, project)
	return nil
}

// HandleProjectInitTask 处理项目初始化任务
func (s *projectService) HandleProjectInitTask(ctx context.Context, t *asynq.Task) error {
	resultWriter := t.ResultWriter()
	tasks.UpdateResult(resultWriter, common.CommonStatusInProgress, 10, "project initialization task executing")

	var payload tasks.ProjectTaskPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}
	logger.Info("handle project initialization task",
		logger.String("taskID", resultWriter.TaskID()),
		logger.String("projectID", payload.ProjectID),
	)

	// 0. 更新项目阶段
	project, projectStage := s.updateProjectToEnvironmentStage(ctx, payload.ProjectID, resultWriter.TaskID())
	if project == nil || projectStage == nil {
		s.reportTaskAndStageError(ctx, resultWriter, projectStage, resultWriter.TaskID(), payload.ProjectGuid, "project or project stage is empty")
		return asynq.SkipRetry
	}

	// 1. 更新项目名和描述
	err := s.updateProjectNameAndBrief(ctx, project)
	if err != nil {
		s.reportTaskAndStageError(ctx, resultWriter, projectStage, resultWriter.TaskID(), project.GUID, err.Error())
		return err
	}

	// 2. 初始化项目模板
	if err = s.initProjectTemplate(ctx, project); err != nil {
		s.reportTaskAndStageError(ctx, resultWriter, projectStage, resultWriter.TaskID(), project.GUID, err.Error())
		return err
	}

	// 3. 提交代码到GitLab
	if err = s.commitProject(ctx, project); err != nil {
		s.reportTaskAndStageError(ctx, resultWriter, projectStage, resultWriter.TaskID(), project.GUID, err.Error())
		return err
	}

	// 4. 开始项目开发过程
	err = s.startDevelopingStage(ctx, project)
	if err != nil {
		s.reportTaskAndStageError(ctx, resultWriter, projectStage, resultWriter.TaskID(), project.GUID, err.Error())
		return err
	}

	s.commonService.UpdateStageStatus(ctx, projectStage, common.CommonStatusDone, "")
	return nil
}

// GetProject 获取项目信息
func (s *projectService) GetProject(ctx context.Context, projectGuid, userID string) (*models.ProjectInfo, error) {
	project, err := s.CheckProjectAccess(ctx, projectGuid, userID)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, errors.New(common.MESSAGE_ACCESS_DENIED)
	}
	return models.ConvertToProjectInfo(project), nil
}

// UpdateProject 更新项目
func (s *projectService) UpdateProject(ctx context.Context, projectGuid string, req *models.UpdateProjectRequest, userID string) (*models.ProjectInfo, error) {
	// 检查项目访问权限
	project, err := s.CheckProjectAccess(ctx, projectGuid, userID)
	if err != nil {
		return nil, err
	}

	// 更新项目字段（只更新非 nil 的字段）
	if req.Name != nil {
		project.Name = *req.Name
	}
	if req.Description != nil {
		project.Description = *req.Description
	}
	if req.CliTool != nil {
		project.CliTool = *req.CliTool
	}
	if req.AiModel != nil {
		project.AiModel = *req.AiModel
	}
	if req.ModelProvider != nil {
		project.ModelProvider = *req.ModelProvider
	}
	if req.ModelApiUrl != nil {
		project.ModelApiUrl = *req.ModelApiUrl
	}

	// 保存更新
	if err := s.repositories.ProjectRepo.Update(ctx, project); err != nil {
		logger.Error("failed to update project",
			logger.String("projectGuid", projectGuid),
			logger.String("error", err.Error()),
		)
		return nil, fmt.Errorf("failed to update project: %s", err.Error())
	}

	// 返回更新后的项目信息
	return s.GetProject(ctx, projectGuid, userID)
}

// DeleteProject 删除项目
func (s *projectService) DeleteProject(ctx context.Context, projectGuid, userID string) error {
	// 检查权限
	project, err := s.CheckProjectAccess(ctx, projectGuid, userID)
	if err != nil {
		return err
	}
	if project == nil {
		return errors.New(common.MESSAGE_ACCESS_DENIED)
	}

	// 如果项目路径存在，异步打包缓存
	if project.ProjectPath != "" && utils.IsDirectoryExists(project.ProjectPath) {
		s.asyncClientService.EnqueueProjectBackupTask(project.ID, projectGuid, project.ProjectPath)
	}

	return s.repositories.ProjectRepo.Delete(ctx, project.ID)
}

// ListProjects 获取项目列表
func (s *projectService) ListProjects(ctx context.Context, req *models.ProjectListRequest, userID string) (*common.PaginationResponse, error) {
	// 设置默认分页参数
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	// 获取项目列表
	projects, total, err := s.repositories.ProjectRepo.List(ctx, req)
	if err != nil {
		return nil, err
	}

	// 转换为响应格式
	projectInfos := make([]*models.ProjectInfo, len(projects))
	for i, project := range projects {
		projectInfos[i] = models.ConvertToProjectInfo(project)
	}

	// 构建分页响应
	pagination := utils.GetPaginationResponse(int(total), req.Page, req.PageSize, projectInfos)
	return pagination, nil
}

// GetUserProjects 获取用户的项目列表
func (s *projectService) GetUserProjects(ctx context.Context, userID string, req *models.ProjectListRequest) (*common.PaginationResponse, error) {
	// 设置默认分页参数
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	// 获取用户项目列表
	projects, total, err := s.repositories.ProjectRepo.GetByUserID(ctx, userID, req)
	if err != nil {
		return nil, err
	}

	// 转换为响应格式
	projectInfos := make([]*models.ProjectInfo, len(projects))
	for i, project := range projects {
		projectInfos[i] = models.ConvertToProjectInfo(project)
	}

	// 构建分页响应
	pagination := utils.GetPaginationResponse(int(total), req.Page, req.PageSize, projectInfos)
	return pagination, nil
}

// 检查项目访问权限
func (s *projectService) CheckProjectAccess(ctx context.Context, projectGuid, userID string) (*models.Project, error) {
	project, err := s.repositories.ProjectRepo.GetByGUID(ctx, projectGuid)
	if err != nil {
		return nil, err
	}

	// 检查权限（用户只能查看自己的项目，管理员可以查看所有项目）
	// 这里简化处理，实际应该从JWT中获取用户角色
	isOwner, err := s.repositories.ProjectRepo.IsOwner(ctx, project.ID, userID)
	if err != nil {
		return nil, err
	}
	if !isOwner {
		return nil, errors.New(common.MESSAGE_ACCESS_DENIED)
	}
	return project, nil
}

// GetProjectByID 通过项目ID获取项目（内部使用，不做权限检查）
func (s *projectService) GetProjectByID(ctx context.Context, projectID string) (*models.Project, error) {
	return s.repositories.ProjectRepo.GetByID(ctx, projectID)
}

// commitProjectToGit 提交项目代码到GitLab
func (s *projectService) commitProjectToGit(ctx context.Context, project *models.Project) error {
	logger.Info("start committing project code to GitLab",
		logger.String("projectID", project.ID),
		logger.String("userID", project.UserID),
	)

	// 构建Git配置
	gitConfig := &GitConfig{
		UserID:        project.UserID,
		GUID:          project.GUID,
		ProjectPath:   project.ProjectPath,
		CommitMessage: fmt.Sprintf("Auto commit by App Maker - %s", project.Name),
		Environment:   s.config.App.Environment,
	}

	logger.Info("initialize project Git repository",
		logger.String("projectID", project.ID),
		logger.String("userID", project.UserID),
		logger.String("projectPath", project.ProjectPath),
	)

	// 初始化Git仓库
	giturl, err := s.gitService.InitializeGit(ctx, gitConfig)
	if err != nil {
		logger.Error("failed to initialize Git repository",
			logger.String("projectID", project.ID),
			logger.String("error", err.Error()),
		)
		return fmt.Errorf("failed to initialize Git repository: %s", err.Error())
	}

	project.GitlabRepoURL = giturl
	// 提交并推送代码
	if err := s.gitService.CommitAndPush(ctx, gitConfig); err != nil {
		logger.Error(MESSAGE_FAILED_COMMIT_TO_GITLAB,
			logger.String("projectID", project.ID),
			logger.String("error", err.Error()),
		)
		return fmt.Errorf("%s: %s", MESSAGE_FAILED_COMMIT_TO_GITLAB, err.Error())
	}

	project.Status = common.CommonStatusInProgress
	s.repositories.ProjectRepo.Update(ctx, project)

	logger.Info("project code committed to GitLab successfully",
		logger.String("projectID", project.ID),
		logger.String("userID", project.UserID),
	)
	return nil
}
