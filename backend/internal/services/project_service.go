package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/lighthought/app-maker/shared-models/agent"
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

	// 创建项目下载任务
	CreateDownloadProjectTask(ctx context.Context, projectID, projectGuid, projectPath string) (string, error)

	// 创建部署项目任务
	CreateDeployProjectTask(ctx context.Context, project *models.Project) (string, error)

	// 处理任务
	ProcessTask(ctx context.Context, task *asynq.Task) error
}

// projectService 项目服务实现
type projectService struct {
	projectRepo      repositories.ProjectRepository
	projectMsgRepo   repositories.MessageRepository
	projectStageRepo repositories.StageRepository
	asyncClient      *asynq.Client
	templateService  ProjectTemplateService
	gitService       GitService
	webSocketService WebSocketService
	config           *config.Config
}

// NewProjectService 创建项目服务实例
func NewProjectService(
	projectRepo repositories.ProjectRepository,
	projectMsgRepo repositories.MessageRepository,
	projectStageRepo repositories.StageRepository,
	asyncClient *asynq.Client,
	templateService ProjectTemplateService,
	gitService GitService,
	webSocketService WebSocketService,
	config *config.Config,
) ProjectService {
	if asyncClient == nil {
		logger.Error("asyncClient is nil!")
		return nil
	}
	return &projectService{
		projectRepo:      projectRepo,
		projectMsgRepo:   projectMsgRepo,
		projectStageRepo: projectStageRepo,
		asyncClient:      asyncClient,
		templateService:  templateService,
		gitService:       gitService,
		webSocketService: webSocketService,
		config:           config,
	}
}

// 统一由这个函数更新项目状态
func (s *projectService) notifyProjectStatusChange(ctx context.Context,
	project *models.Project, message *models.ConversationMessage, stageName common.DevStatus) {
	if message != nil {
		// 保存用户消息
		if err := s.projectMsgRepo.Create(ctx, message); err != nil {
			logger.Error("保存项目消息失败",
				logger.String("error", err.Error()),
				logger.String("projectID", project.ID),
			)
		}
		s.webSocketService.NotifyProjectMessage(ctx, project.GUID, message)
	}

	if stageName != "" {
		// 插入项目阶段
		stage := models.NewDevStage(project, stageName, common.CommonStatusInProgress)

		if err := s.projectStageRepo.Create(ctx, stage); err != nil {
			logger.Error("插入项目阶段失败",
				logger.String("error", err.Error()),
				logger.String("projectID", project.ID),
			)
		}
		logger.Info("插入项目阶段成功", logger.String("projectID", project.ID))

		project.SetDevStatus(stageName)
		s.projectRepo.Update(ctx, project)

		s.webSocketService.NotifyProjectStageUpdate(ctx, project.GUID, stage)
	}
}

// 更新项目网络设置
func (s *projectService) updateProjectNetworkSetting(ctx context.Context, project *models.Project) error {
	// 自动获取可用端口
	ports, err := s.projectRepo.GetNextAvailablePorts(ctx)
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

	// TODO: 获取下一个可用的子网段
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
	if err := s.projectRepo.Create(ctx, newProject); err != nil {
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

	// asynq 异步调用初始化项目流程
	taskInfo, err := s.asyncClient.Enqueue(tasks.NewProjectInitTask(newProject.ID, newProject.GUID, newProject.ProjectPath))
	if err != nil {
		stage := models.NewDevStage(newProject, common.DevStatusInitializing, common.CommonStatusFailed)
		if err = s.projectStageRepo.Create(ctx, stage); err != nil {
			logger.Error("插入项目阶段失败",
				logger.String("error", err.Error()),
				logger.String("projectID", newProject.ID),
			)
		}
		s.webSocketService.NotifyProjectStageUpdate(ctx, newProject.GUID, stage)
		return nil, fmt.Errorf("failed to create project init task: %s", err.Error())
	}

	newProject.CurrentTaskID = taskInfo.ID

	// 更新项目
	logger.Info("保存项目到数据库")
	if err := s.projectRepo.Update(ctx, newProject); err != nil {
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
	s.notifyProjectStatusChange(ctx, newProject, userMsg, common.DevStatusInitializing)
	return projectInfo, nil
}

// ProcessTask 处理任务
func (s *projectService) ProcessTask(ctx context.Context, task *asynq.Task) error {
	switch task.Type() {
	case common.TaskTypeProjectInit:
		return s.HandleProjectInitTask(ctx, task)
	default:
		return fmt.Errorf("unexpected task type %s", task.Type())
	}
}

// 统一由这个函数更新项目阶段
func (s *projectService) updateStage(ctx context.Context, stage *models.DevStage) {
	s.projectStageRepo.Update(ctx, stage)
	s.webSocketService.NotifyProjectStageUpdate(ctx, stage.ProjectGuid, stage)
}

func (s *projectService) reportTaskAndStageError(ctx context.Context,
	projectID string, errMsg string, resultWriter *asynq.ResultWriter, devStage *models.DevStage) {
	if resultWriter != nil {
		tasks.UpdateResult(resultWriter, common.CommonStatusFailed, 0, errMsg)
	}

	if devStage != nil {
		devStage.SetStatus(common.CommonStatusFailed)
		devStage.TaskID = resultWriter.TaskID()
		devStage.FailedReason = errMsg
		s.updateStage(ctx, devStage)
	}

	logger.Error("项目初始化任务执行失败：",
		logger.String("error", errMsg),
		logger.String("projectID", projectID),
	)
}

// updateProjectToEnvironmentStage 更新项目阶段
func (s *projectService) updateProjectToEnvironmentStage(ctx context.Context, projectID, taskID string) (*models.Project, *models.DevStage) {
	// 更新 initializing 的 stage 为 done，表示 API 内部这个 async 调用成功，也获取到了合法的数据
	stage, err := s.projectStageRepo.UpdateStageToDone(ctx, projectID, string(common.DevStatusInitializing))
	if err != nil {
		logger.Error("更新项目阶段失败",
			logger.String("error", err.Error()),
			logger.String("projectID", projectID),
		)
	}
	s.webSocketService.NotifyProjectStageUpdate(ctx, stage.ProjectGuid, stage)

	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		logger.Error("获取项目信息失败",
			logger.String("error", err.Error()),
			logger.String("projectID", projectID),
		)
		return nil, nil
	}

	project.Status = common.CommonStatusInProgress
	project.CurrentTaskID = taskID
	project.SetDevStatus(common.DevStatusSetupEnvironment)
	s.projectRepo.Update(ctx, project)

	s.webSocketService.NotifyProjectInfoUpdate(ctx, project.GUID, project)

	// 已经有过环境准备的阶段，取原来的数据
	projectStages, err := s.projectStageRepo.GetByProjectID(ctx, projectID)
	if err != nil {
		logger.Error("获取项目阶段失败",
			logger.String("error", err.Error()),
			logger.String("projectID", projectID),
		)
	}
	for _, stage := range projectStages {
		if stage.Name == string(common.DevStatusSetupEnvironment) {
			stage.SetStatus(common.CommonStatusInProgress)
			stage.TaskID = taskID
			s.updateStage(ctx, stage)
			return project, stage
		}
	}

	// 没有，才插入环境准备的阶段
	projectStage := models.NewDevStage(project, common.DevStatusSetupEnvironment, common.CommonStatusInProgress)
	projectStage.TaskID = taskID
	if err := s.projectStageRepo.Create(ctx, projectStage); err != nil {
		logger.Error("插入项目阶段失败",
			logger.String("error", err.Error()),
			logger.String("projectID", project.ID),
		)
		return project, nil
	}
	s.webSocketService.NotifyProjectStageUpdate(ctx, project.GUID, projectStage)
	return project, projectStage
}

// updateProjectNameAndBrief 更新项目名和描述
func (s *projectService) updateProjectNameAndBrief(ctx context.Context, project *models.Project,
	resultWriter *asynq.ResultWriter, projectStage *models.DevStage) error {
	if project == nil || projectStage == nil {
		logger.Error("invalid project or project stage parameters")
		return fmt.Errorf("invalid project or project stage parameters")
	}

	// 已经生成过，跳过
	if project.Name != "" && project.Name != common.DefaultProjectName && project.Description != "" {
		tasks.UpdateResult(resultWriter, common.CommonStatusInProgress, 40, "project name and description already exist")
		logger.Info("project name and description already exist, skip generation")
		return nil
	}

	logger.Info("1. start generating project name and description",
		logger.String("projectID", project.ID),
		logger.String("requirements", project.Requirements),
	)
	summary, err := utils.GenerateProjectSummary(project.Requirements)
	if err != nil {
		s.reportTaskAndStageError(ctx, project.ID, "failed to generate project name and description", resultWriter, projectStage)
		return fmt.Errorf("failed to generate project name and description: %s", err.Error())
	}

	tasks.UpdateResult(resultWriter, common.CommonStatusInProgress, 20, "生成项目名和描述成功")
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

	if err := s.projectMsgRepo.Create(ctx, projectMsg); err != nil {
		logger.Error("failed to save project message",
			logger.String("error", err.Error()),
			logger.String("projectID", project.ID),
		)
		return fmt.Errorf("failed to save project message: %s", err.Error())
	}
	s.webSocketService.NotifyProjectMessage(ctx, project.GUID, projectMsg)
	tasks.UpdateResult(resultWriter, common.CommonStatusInProgress, 30, "project summary generated")

	project.Name = strings.ToLower(summary.Title)
	project.Description = summary.Content
	logger.Info("project name and description generated successfully",
		logger.String("projectID", project.ID),
		logger.String("projectName", project.Name),
		logger.String("projectDescription", project.Description),
	)

	s.projectRepo.Update(ctx, project)
	s.webSocketService.NotifyProjectInfoUpdate(ctx, project.GUID, project)
	tasks.UpdateResult(resultWriter, common.CommonStatusInProgress, 40, "project information updated successfully")
	return nil
}

func (s *projectService) initProjectTemplate(ctx context.Context, project *models.Project,
	resultWriter *asynq.ResultWriter, projectStage *models.DevStage) {
	logger.Info("2. start initializing project template",
		logger.String("projectID", project.ID),
		logger.String("projectPath", project.ProjectPath),
	)

	// 已经初始化过，跳过
	if project.ProjectPath != "" && utils.IsDirectoryExists(project.ProjectPath) {
		tasks.UpdateResult(resultWriter, common.CommonStatusInProgress, 60, "project template already initialized")
		logger.Info("project template already initialized, skip initialization")
		return
	}

	if err := s.templateService.InitializeProject(ctx, project); err != nil {
		// 模板初始化失败不影响项目创建，但记录错误
		logger.Error("failed to initialize project template",
			logger.String("error", err.Error()),
			logger.String("projectID", project.ID),
			logger.String("projectPath", project.ProjectPath),
		)
		s.reportTaskAndStageError(ctx, project.ID, "模板初始化失败", resultWriter, projectStage)
		return
	}

	tasks.UpdateResult(resultWriter, common.CommonStatusInProgress, 60, "project template initialized successfully")

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

	if err := s.projectMsgRepo.Create(ctx, projectMsg); err != nil {
		logger.Error("failed to save project message",
			logger.String("error", err.Error()),
			logger.String("projectID", project.ID),
		)
	}
}

// commitProject 提交代码到GitLab
func (s *projectService) commitProject(ctx context.Context, project *models.Project,
	resultWriter *asynq.ResultWriter, projectStage *models.DevStage) error {

	// 已经提交过，跳过
	if project.GitlabRepoURL != "" {
		tasks.UpdateResult(resultWriter, common.CommonStatusInProgress, 70, "project code already committed to GitLab")
		logger.Info("project code already committed to GitLab, skip commit")
		return nil
	}

	err := s.commitProjectToGit(ctx, project)
	if err != nil {
		logger.Error("failed to commit project code to GitLab",
			logger.String("error", err.Error()),
			logger.String("projectID", project.ID),
		)
		s.reportTaskAndStageError(ctx, project.ID, "提交代码到GitLab失败", resultWriter, projectStage)
		return fmt.Errorf("failed to commit project code to GitLab: %s", err.Error())
	}

	tasks.UpdateResult(resultWriter, common.CommonStatusInProgress, 70, "project code committed to GitLab successfully")
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

	if err := s.projectMsgRepo.Create(ctx, projectMsg); err != nil {
		logger.Error("failed to save project message",
			logger.String("error", err.Error()),
			logger.String("projectID", project.ID),
			logger.String("projectGUID", project.GUID),
		)
	}

	if project.GitlabRepoURL == "" {
		s.reportTaskAndStageError(ctx, project.ID, "GitLab仓库URL为空", resultWriter, projectStage)
		return fmt.Errorf("GitLab repository URL is empty, projectGUID: %s", project.GUID)
	}
	return nil
}

// startDevelopingStage 调用AgentServer
func (s *projectService) startDevelopingStage(ctx context.Context, project *models.Project,
	resultWriter *asynq.ResultWriter, projectStage *models.DevStage) error {
	tasks.UpdateResult(resultWriter, common.CommonStatusInProgress, 80, "project created successfully, development process started")

	//异步创建项目开发任务
	taskInfo, err := s.asyncClient.Enqueue(tasks.NewProjectDevelopmentTask(project.ID, project.GUID, project.GitlabRepoURL))
	if err != nil {
		logger.Error("failed to create project development task",
			logger.String("error", err.Error()),
			logger.String("projectID", project.ID),
		)
		s.reportTaskAndStageError(ctx, project.ID, "failed to create project development task", resultWriter, projectStage)
		return fmt.Errorf("failed to create project development task: %s", err.Error())
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
			"\",\n \"taskID\": \"" + taskInfo.ID + "\"\n}\n```",
		IsExpanded: true,
	}

	s.notifyProjectStatusChange(ctx, project, projectMsg, "")

	tasks.UpdateResult(resultWriter, common.CommonStatusDone, 100, "project initialization task completed")

	project.Status = common.CommonStatusInProgress
	project.CurrentTaskID = taskInfo.ID
	s.projectRepo.Update(ctx, project)
	s.webSocketService.NotifyProjectInfoUpdate(ctx, project.GUID, project)

	projectStage.SetStatus(common.CommonStatusDone)
	s.updateStage(ctx, projectStage)
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
		logger.Error("project or project stage is empty")
		tasks.UpdateResult(resultWriter, common.CommonStatusFailed, 0, "project or project stage is empty")
		return fmt.Errorf("project or project stage is empty")
	}

	// 1. 更新项目名和描述
	err := s.updateProjectNameAndBrief(ctx, project, resultWriter, projectStage)
	if err != nil {
		tasks.UpdateResult(resultWriter, common.CommonStatusFailed, 0, "failed to update project name and description")
		return fmt.Errorf("failed to update project name and description: %s", err.Error())
	}

	// 2. 初始化项目模板
	s.initProjectTemplate(ctx, project, resultWriter, projectStage)

	// 3. 提交代码到GitLab
	err = s.commitProject(ctx, project, resultWriter, projectStage)
	if err != nil {
		tasks.UpdateResult(resultWriter, common.CommonStatusFailed, 0, "failed to commit project code to GitLab")
		return fmt.Errorf("failed to commit project code to GitLab: %s", err.Error())
	}

	// 4. 开始项目开发过程
	err = s.startDevelopingStage(ctx, project, resultWriter, projectStage)
	if err != nil {
		tasks.UpdateResult(resultWriter, common.CommonStatusFailed, 0, "failed to call AgentServer")
		return fmt.Errorf("failed to call AgentServer: %s", err.Error())
	}

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
	if err := s.projectRepo.Update(ctx, project); err != nil {
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
		s.asyncClient.Enqueue(tasks.NewProjectBackupTask(project.ID, projectGuid, project.ProjectPath))
	}

	return s.projectRepo.Delete(ctx, project.ID)
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
	projects, total, err := s.projectRepo.List(ctx, req)
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
	projects, total, err := s.projectRepo.GetByUserID(ctx, userID, req)
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
	project, err := s.projectRepo.GetByGUID(ctx, projectGuid)
	if err != nil {
		return nil, err
	}

	// 检查权限（用户只能查看自己的项目，管理员可以查看所有项目）
	// 这里简化处理，实际应该从JWT中获取用户角色
	isOwner, err := s.projectRepo.IsOwner(ctx, project.ID, userID)
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
	return s.projectRepo.GetByID(ctx, projectID)
}

// DownloadProject 下载项目文件
func (s *projectService) CreateDownloadProjectTask(ctx context.Context, projectID, projectGuid, projectPath string) (string, error) {
	// 检查项目路径是否存在
	if !utils.IsDirectoryExists(projectPath) {
		logger.Error("project path is empty", logger.String("projectPath", projectPath))
		return "", fmt.Errorf("project path is empty")
	}

	// 异步方法，返回任务 ID
	info, err := s.asyncClient.Enqueue(tasks.NewProjectDownloadTask(projectID, projectGuid, projectPath))
	if err != nil {
		return "", fmt.Errorf("failed to download project file: %s", err.Error())
	}

	return info.ID, nil
}

// 创建部署项目任务
func (s *projectService) CreateDeployProjectTask(ctx context.Context, project *models.Project) (string, error) {
	req := &agent.DeployReq{
		ProjectGuid:   project.GUID,
		Environment:   "dev",
		DeployOptions: map[string]interface{}{},
	}
	// 异步方法，返回任务 ID
	info, err := s.asyncClient.Enqueue(tasks.NewProjectDeployTask(req))
	if err != nil {
		return "", fmt.Errorf("failed to create deploy project task: %s", err.Error())
	}

	return info.ID, nil
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
		logger.Error("failed to commit project code to GitLab",
			logger.String("projectID", project.ID),
			logger.String("error", err.Error()),
		)
		return fmt.Errorf("failed to commit project code to GitLab: %s", err.Error())
	}

	project.Status = common.CommonStatusInProgress
	s.projectRepo.Update(ctx, project)

	logger.Info("project code committed to GitLab successfully",
		logger.String("projectID", project.ID),
		logger.String("userID", project.UserID),
	)
	return nil
}
