package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"autocodeweb-backend/internal/models"
	"autocodeweb-backend/internal/repositories"
	"shared-models/common"
	"shared-models/logger"
	"shared-models/tasks"
	"shared-models/utils"

	"github.com/hibiken/asynq"
)

// ProjectService 项目服务接口
type ProjectService interface {
	// 基础CRUD操作
	CreateProject(ctx context.Context, req *models.CreateProjectRequest, userID string) (*models.ProjectInfo, error)
	GetProject(ctx context.Context, projectGuid, userID string) (*models.ProjectInfo, error)
	DeleteProject(ctx context.Context, projectGuid, userID string) error
	ListProjects(ctx context.Context, req *models.ProjectListRequest, userID string) (*models.PaginationResponse, error)

	// 用户项目管理
	GetUserProjects(ctx context.Context, userID string, req *models.ProjectListRequest) (*models.PaginationResponse, error)

	// 检查项目访问权限
	CheckProjectAccess(ctx context.Context, projectGuid, userID string) (*models.Project, error)

	// CreateDownloadProjectTask 创建项目下载任务
	CreateDownloadProjectTask(ctx context.Context, projectID, projectGuid, projectPath string) (string, error)

	// ProcessTask 处理任务
	ProcessTask(ctx context.Context, task *asynq.Task) error
}

// projectService 项目服务实现
type projectService struct {
	projectRepo      repositories.ProjectRepository
	projectMsgRepo   repositories.MessageRepository
	projectStageRepo repositories.StageRepository
	asyncClient      *asynq.Client
	templateService  ProjectTemplateService
	nameGenerator    ProjectNameGenerator
	gitService       GitService
	webSocketService WebSocketService
}

// NewProjectService 创建项目服务实例
func NewProjectService(
	projectRepo repositories.ProjectRepository,
	projectMsgRepo repositories.MessageRepository,
	projectStageRepo repositories.StageRepository,
	asyncClient *asynq.Client,
	templateService ProjectTemplateService,
	nameGenerator ProjectNameGenerator,
	gitService GitService,
	webSocketService WebSocketService,
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
		nameGenerator:    nameGenerator,
		gitService:       gitService,
		webSocketService: webSocketService,
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
	bGerated := s.nameGenerator.GenerateProjectConfig(req.Requirements, newProject)
	if !bGerated {
		logger.Error("自动生成项目配置信息失败", logger.String("requirements", req.Requirements))
		return nil, fmt.Errorf("failed to generate project config: %w", errors.New("failed to generate project config"))
	}

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
		return nil, fmt.Errorf("创建项目初始化任务失败: %w", err)
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
	case common.TypeProjectInit:
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
		tasks.UpdateResult(resultWriter, common.CommonStatusFailed, 100, errMsg)
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
	resultWriter *asynq.ResultWriter, projectStage *models.DevStage) {
	if project == nil || projectStage == nil {
		logger.Error("项目或项目阶段为空")
		return
	}

	// 已经生成过，跳过
	if project.Name != "" && project.Name != "newproj" && project.Description != "" {
		tasks.UpdateResult(resultWriter, common.CommonStatusInProgress, 40, "项目名和描述已存在")
		logger.Info("项目名和描述已存在，跳过生成")
		return
	}

	logger.Info("1. 开始生成项目名和描述",
		logger.String("projectID", project.ID),
		logger.String("requirements", project.Requirements),
	)
	summary, err := utils.GenerateProjectSummary(project.Requirements)
	if err != nil {
		s.reportTaskAndStageError(ctx, project.ID, "生成项目名和描述失败", resultWriter, projectStage)
		return
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
		logger.Error("保存项目消息失败",
			logger.String("error", err.Error()),
			logger.String("projectID", project.ID),
		)
	}
	s.webSocketService.NotifyProjectMessage(ctx, project.GUID, projectMsg)
	tasks.UpdateResult(resultWriter, common.CommonStatusInProgress, 30, "项目简介已生成")

	project.Name = strings.ToLower(summary.Title)
	project.Description = summary.Content
	logger.Info("生成项目名和描述成功",
		logger.String("projectID", project.ID),
		logger.String("projectName", project.Name),
		logger.String("projectDescription", project.Description),
	)

	s.projectRepo.Update(ctx, project)
	s.webSocketService.NotifyProjectInfoUpdate(ctx, project.GUID, project)
	tasks.UpdateResult(resultWriter, common.CommonStatusInProgress, 40, "更新项目信息成功")
}

func (s *projectService) initProjectTemplate(ctx context.Context, project *models.Project,
	resultWriter *asynq.ResultWriter, projectStage *models.DevStage) {
	logger.Info("2. 开始初始化项目模板",
		logger.String("projectID", project.ID),
		logger.String("projectPath", project.ProjectPath),
	)

	// 已经初始化过，跳过
	if project.ProjectPath != "" && utils.IsDirectoryExists(project.ProjectPath) {
		tasks.UpdateResult(resultWriter, common.CommonStatusInProgress, 60, "项目模板已初始化")
		logger.Info("项目模板已初始化，跳过初始化")
		return
	}

	if err := s.templateService.InitializeProject(ctx, project); err != nil {
		// 模板初始化失败不影响项目创建，但记录错误
		logger.Error("项目模板初始化失败",
			logger.String("error", err.Error()),
			logger.String("projectID", project.ID),
			logger.String("projectPath", project.ProjectPath),
		)
		s.reportTaskAndStageError(ctx, project.ID, "项目模板初始化失败", resultWriter, projectStage)
		return
	}

	tasks.UpdateResult(resultWriter, common.CommonStatusInProgress, 60, "项目模板初始化成功")

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
		logger.Error("保存项目消息失败",
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
		tasks.UpdateResult(resultWriter, common.CommonStatusInProgress, 70, "项目代码已提交到GitLab")
		logger.Info("项目代码已提交到GitLab，跳过提交")
		return nil
	}

	err := s.commitProjectToGit(ctx, project)
	if err != nil {
		logger.Error("提交代码到GitLab失败",
			logger.String("error", err.Error()),
			logger.String("projectID", project.ID),
		)
		s.reportTaskAndStageError(ctx, project.ID, "提交代码到GitLab失败", resultWriter, projectStage)
		return fmt.Errorf("提交代码到GitLab失败: %w", err)
	}

	tasks.UpdateResult(resultWriter, common.CommonStatusInProgress, 70, "提交代码到GitLab成功")
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
		logger.Error("保存项目消息失败",
			logger.String("error", err.Error()),
			logger.String("projectID", project.ID),
			logger.String("projectGUID", project.GUID),
		)
	}

	if project.GitlabRepoURL == "" {
		s.reportTaskAndStageError(ctx, project.ID, "GitLab仓库URL为空", resultWriter, projectStage)
		return fmt.Errorf("GitLab仓库URL为空, projectGUID: %s", project.GUID)
	}
	return nil
}

// callAgentServer 调用AgentServer
func (s *projectService) callAgentServer(ctx context.Context, project *models.Project,
	resultWriter *asynq.ResultWriter, projectStage *models.DevStage) error {
	tasks.UpdateResult(resultWriter, common.CommonStatusInProgress, 80, "项目创建完成，开发流程已启动")

	taskInfo, err := s.asyncClient.Enqueue(tasks.NewProjectDevelopmentTask(project.ID, project.GUID, project.GitlabRepoURL))
	if err != nil {
		logger.Error("创建项目开发任务失败",
			logger.String("error", err.Error()),
			logger.String("projectID", project.ID),
		)
		s.reportTaskAndStageError(ctx, project.ID, "创建项目开发任务失败", resultWriter, projectStage)
		return fmt.Errorf("创建项目开发任务失败: %w", err)
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

	tasks.UpdateResult(resultWriter, common.CommonStatusDone, 100, "项目初始化任务完成")

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
	tasks.UpdateResult(resultWriter, common.CommonStatusInProgress, 10, "项目初始化任务执行中")

	var payload tasks.ProjectTaskPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}
	logger.Info("处理项目初始化任务",
		logger.String("taskID", resultWriter.TaskID()),
		logger.String("projectID", payload.ProjectID),
	)

	// 0. 更新项目阶段
	project, projectStage := s.updateProjectToEnvironmentStage(ctx, payload.ProjectID, resultWriter.TaskID())
	if project == nil || projectStage == nil {
		logger.Error("项目或项目阶段为空")
		return fmt.Errorf("项目或项目阶段为空")
	}

	// 1. 更新项目名和描述
	s.updateProjectNameAndBrief(ctx, project, resultWriter, projectStage)

	// 2. 初始化项目模板
	s.initProjectTemplate(ctx, project, resultWriter, projectStage)

	// 3. 提交代码到GitLab，触发自动编译打包
	err := s.commitProject(ctx, project, resultWriter, projectStage)
	if err != nil {
		return fmt.Errorf("提交代码到GitLab失败: %w", err)
	}

	// 4. asynq 异步调用 agents-server 的接口
	err = s.callAgentServer(ctx, project, resultWriter, projectStage)
	if err != nil {
		return fmt.Errorf("调用AgentServer失败: %w", err)
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
		return nil, errors.New("access denied")
	}
	return models.ConvertToProjectInfo(project), nil
}

// DeleteProject 删除项目
func (s *projectService) DeleteProject(ctx context.Context, projectGuid, userID string) error {
	// 检查权限
	project, err := s.CheckProjectAccess(ctx, projectGuid, userID)
	if err != nil {
		return err
	}
	if project == nil {
		return errors.New("access denied")
	}

	// 如果项目路径存在，异步打包缓存
	if project.ProjectPath != "" && utils.IsDirectoryExists(project.ProjectPath) {
		s.asyncClient.Enqueue(tasks.NewProjectBackupTask(project.ID, projectGuid, project.ProjectPath))
	}

	return s.projectRepo.Delete(ctx, project.ID)
}

// ListProjects 获取项目列表
func (s *projectService) ListProjects(ctx context.Context, req *models.ProjectListRequest, userID string) (*models.PaginationResponse, error) {
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
	totalPages := (int(total) + req.PageSize - 1) / req.PageSize
	pagination := &models.PaginationResponse{
		Code:        common.SUCCESS_CODE,
		Message:     "success",
		Total:       int(total),
		Page:        req.Page,
		PageSize:    req.PageSize,
		TotalPages:  totalPages,
		Data:        projectInfos,
		HasNext:     req.Page < totalPages,
		HasPrevious: req.Page > 1,
		Timestamp:   utils.GetCurrentTime(),
	}

	return pagination, nil
}

// GetUserProjects 获取用户的项目列表
func (s *projectService) GetUserProjects(ctx context.Context, userID string, req *models.ProjectListRequest) (*models.PaginationResponse, error) {
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
	totalPages := (int(total) + req.PageSize - 1) / req.PageSize
	pagination := &models.PaginationResponse{
		Code:        common.SUCCESS_CODE,
		Message:     "success",
		Total:       int(total),
		Page:        req.Page,
		PageSize:    req.PageSize,
		TotalPages:  totalPages,
		Data:        projectInfos,
		HasNext:     req.Page < totalPages,
		HasPrevious: req.Page > 1,
		Timestamp:   utils.GetCurrentTime(),
	}

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
		return nil, errors.New("access denied")
	}
	return project, nil
}

// DownloadProject 下载项目文件
func (s *projectService) CreateDownloadProjectTask(ctx context.Context, projectID, projectGuid, projectPath string) (string, error) {
	// 检查项目路径是否存在
	if !utils.IsDirectoryExists(projectPath) {
		logger.Error("项目路径为空", logger.String("projectPath", projectPath))
		return "", fmt.Errorf("项目路径为空")
	}

	// 异步方法，返回任务 ID
	info, err := s.asyncClient.Enqueue(tasks.NewProjectDownloadTask(projectID, projectGuid, projectPath))
	if err != nil {
		return "", fmt.Errorf("下载项目文件失败: %w", err)
	}

	return info.ID, nil
}

// commitProjectToGit 提交项目代码到GitLab
func (s *projectService) commitProjectToGit(ctx context.Context, project *models.Project) error {
	logger.Info("开始提交项目代码到GitLab",
		logger.String("projectID", project.ID),
		logger.String("userID", project.UserID),
	)

	// 构建Git配置
	gitConfig := &GitConfig{
		UserID:        project.UserID,
		GUID:          project.GUID,
		ProjectPath:   project.ProjectPath,
		CommitMessage: fmt.Sprintf("Auto commit by App Maker - %s", project.Name),
	}

	logger.Info("初始化项目 Git 仓库",
		logger.String("projectID", project.ID),
		logger.String("userID", project.UserID),
		logger.String("projectPath", project.ProjectPath),
	)

	// 初始化Git仓库
	giturl, err := s.gitService.InitializeGit(ctx, gitConfig)
	if err != nil {
		logger.Error("初始化Git仓库失败",
			logger.String("projectID", project.ID),
			logger.String("error", err.Error()),
		)
		return fmt.Errorf("初始化Git仓库失败: %w", err)
	}

	project.GitlabRepoURL = giturl
	// 提交并推送代码
	if err := s.gitService.CommitAndPush(ctx, gitConfig); err != nil {
		logger.Error("提交代码到GitLab失败",
			logger.String("projectID", project.ID),
			logger.String("error", err.Error()),
		)
		return fmt.Errorf("提交代码到GitLab失败: %w", err)
	}

	project.Status = common.CommonStatusInProgress
	s.projectRepo.Update(ctx, project)

	logger.Info("项目代码已成功提交到GitLab",
		logger.String("projectID", project.ID),
		logger.String("userID", project.UserID),
	)
	return nil
}
