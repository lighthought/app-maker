package services

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"autocodeweb-backend/internal/config"
	"autocodeweb-backend/internal/constants"
	"autocodeweb-backend/internal/models"
	"autocodeweb-backend/internal/repositories"
	"autocodeweb-backend/internal/tasks"
	"autocodeweb-backend/internal/utils"
	"autocodeweb-backend/pkg/logger"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"gorm.io/gorm"
)

// ProjectService 项目服务接口
type ProjectService interface {
	// 基础CRUD操作
	CreateProject(ctx context.Context, req *models.CreateProjectRequest, userID string) (*models.ProjectInfo, error)
	GetProject(ctx context.Context, projectID, userID string) (*models.ProjectInfo, error)
	DeleteProject(ctx context.Context, projectID, userID string) error
	ListProjects(ctx context.Context, req *models.ProjectListRequest, userID string) (*models.PaginationResponse, error)

	// 用户项目管理
	GetUserProjects(ctx context.Context, userID string, req *models.ProjectListRequest) (*models.PaginationResponse, error)

	// 项目开发阶段管理
	GetProjectStages(ctx context.Context, projectID string) ([]*models.DevStage, error)

	// 检查项目访问权限
	CheckProjectAccess(ctx context.Context, projectID, userID string) (*models.Project, error)

	// CreateDownloadProjectTask 创建项目下载任务
	CreateDownloadProjectTask(ctx context.Context, projectID, projectPath string) (string, error)
}

// projectService 项目服务实现
type projectService struct {
	projectRepo         repositories.ProjectRepository
	projectMsgRepo      repositories.MessageRepository
	templateService     ProjectTemplateService
	projectStageService *ProjectStageService
	nameGenerator       ProjectNameGenerator
	asyncClient         *asynq.Client
	gitService          *GitService
}

// NewProjectService 创建项目服务实例
func NewProjectService(
	db *gorm.DB,
	asyncClient *asynq.Client,
	fileService FileService,
	cfg *config.Config,
) ProjectService {
	projectRepo := repositories.NewProjectRepository(db)
	projectMsgRepo := repositories.NewMessageRepository(db)
	stageRepo := repositories.NewStageRepository(db)
	gitService := NewGitService()
	gitService.SetupSSH()
	return &projectService{
		projectRepo:         projectRepo,
		projectMsgRepo:      projectMsgRepo,
		templateService:     NewProjectTemplateService(fileService),
		projectStageService: NewProjectStageService(projectRepo, stageRepo),
		nameGenerator:       NewProjectNameGenerator(cfg),
		asyncClient:         asyncClient,
		gitService:          gitService,
	}
}

// CreateProject 创建项目
func (s *projectService) CreateProject(ctx context.Context, req *models.CreateProjectRequest, userID string) (*models.ProjectInfo, error) {
	logger.Info("开始创建项目",
		logger.String("userID", userID),
		logger.String("requirements", req.Requirements),
	)

	filePath := filepath.Join("/app/data/projects", userID, uuid.New().String()) // 这里是假的路径，需要替换为真实的路径
	newProject := &models.Project{
		Requirements: req.Requirements,
		UserID:       userID,
		Status:       "draft",
		ProjectPath:  filePath,
		BackendPort:  9501,
		FrontendPort: 3501,
		RedisPort:    7501,
		PostgresPort: 5501,
	}

	logger.Info("数据库新建项目")
	if err := s.projectRepo.Create(ctx, newProject); err != nil {
		logger.Error("保存项目到数据库失败",
			logger.String("error", err.Error()),
			logger.String("projectID", newProject.ID),
		)
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	// 替换为最终的项目路径
	newProject.ProjectPath = utils.GetProjectPath(userID, newProject.ID)
	logger.Info("生成项目路径", logger.String("projectPath", newProject.ProjectPath))

	// 自动生成项目配置信息和密码信息
	bGerated := s.nameGenerator.GenerateProjectConfig(req.Requirements, newProject)
	if !bGerated {
		logger.Error("自动生成项目配置信息失败", logger.String("requirements", req.Requirements))
		return nil, fmt.Errorf("failed to generate project config: %w", errors.New("failed to generate project config"))
	}

	logger.Info("自动生成项目名", logger.String("projectName", newProject.Name))

	// 自动获取可用端口
	ports, err := s.projectRepo.GetNextAvailablePorts(ctx)
	if err != nil {
		logger.Error("获取可用端口失败",
			logger.String("error", err.Error()),
		)
		return nil, fmt.Errorf("failed to get available ports: %w", err)
	}

	logger.Info("自动分配端口",
		logger.Int("backendPort", ports.BackendPort),
		logger.Int("frontendPort", ports.FrontendPort),
		logger.Int("redisPort", ports.RedisPort),
		logger.Int("postgresPort", ports.PostgresPort),
	)

	newProject.BackendPort = ports.BackendPort
	newProject.FrontendPort = ports.FrontendPort
	newProject.RedisPort = ports.RedisPort
	newProject.PostgresPort = ports.PostgresPort

	// TODO: 获取下一个可用的子网段
	if newProject.Subnetwork == "" {
		newProject.Subnetwork = "172.20.0.0/16"
	}

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

	newProject.UserID = userID
	go s.afterCreateProject(context.Background(), newProject)

	projectInfo, err := s.GetProject(ctx, newProject.ID, userID)
	if err != nil {
		logger.Error("获取项目信息失败",
			logger.String("error", err.Error()),
			logger.String("projectID", newProject.ID),
		)
		return nil, err
	}
	return projectInfo, nil
}

// generateProjectName 生成项目名
func (s *projectService) afterCreateProject(ctx context.Context, project *models.Project) {
	if project == nil {
		return
	}
	// 1. 获取项目名和描述
	logger.Info("开始生成项目名和描述",
		logger.String("projectID", project.ID),
		logger.String("requirements", project.Requirements),
	)
	summary, err := utils.GenerateProjectSummary(project.Requirements)
	if err != nil {
		logger.Error("生成项目名和描述失败",
			logger.String("error", err.Error()),
			logger.String("projectID", project.ID),
		)
		return
	}

	projectMsg := &models.ConversationMessage{
		ProjectID:       project.ID,
		Type:            constants.ConversationTypeSystem,
		AgentRole:       constants.AgentAnalyst.Role,
		AgentName:       constants.AgentAnalyst.Name,
		Content:         "项目简介已生成",
		IsMarkdown:      true,
		MarkdownContent: summary.Title + "\r\n" + summary.Content,
		IsExpanded:      true,
	}

	if err := s.projectMsgRepo.Create(ctx, projectMsg); err != nil {
		logger.Error("保存项目消息失败",
			logger.String("error", err.Error()),
			logger.String("projectID", project.ID),
		)
		return
	}

	project.Name = summary.Title
	project.Description = summary.Content
	logger.Info("生成项目名和描述成功",
		logger.String("projectID", project.ID),
		logger.String("projectName", project.Name),
		logger.String("projectDescription", project.Description),
	)

	s.projectRepo.Update(ctx, project)

	// 2. 初始化项目模板
	logger.Info("2. 开始初始化项目模板",
		logger.String("projectID", project.ID),
		logger.String("projectPath", project.ProjectPath),
	)

	if err := s.templateService.InitializeProject(ctx, project); err != nil {
		// 模板初始化失败不影响项目创建，但记录错误
		logger.Error("项目模板初始化失败",
			logger.String("error", err.Error()),
			logger.String("projectID", project.ID),
			logger.String("projectPath", project.ProjectPath),
		)
		return
	}

	projectMsg2 := &models.ConversationMessage{
		ProjectID:       project.ID,
		Type:            constants.ConversationTypeSystem,
		AgentRole:       constants.AgentDev.Role,
		AgentName:       constants.AgentDev.Name,
		Content:         "项目模板初始化成功",
		IsMarkdown:      true,
		MarkdownContent: project.ID + ", " + project.Name + "\r\n" + project.ProjectPath,
		IsExpanded:      true,
	}

	if err := s.projectMsgRepo.Create(ctx, projectMsg2); err != nil {
		logger.Error("保存项目消息失败",
			logger.String("error", err.Error()),
			logger.String("projectID", project.ID),
		)
	}

	// 3. 提交代码到GitLab，触发自动编译打包
	err = s.commitProjectToGit(ctx, project)
	if err != nil {
		logger.Error("提交代码到GitLab失败",
			logger.String("error", err.Error()),
			logger.String("projectID", project.ID),
		)
	} else {
		projectMsg3 := &models.ConversationMessage{
			ProjectID:       project.ID,
			Type:            constants.ConversationTypeSystem,
			AgentRole:       constants.AgentDev.Role,
			AgentName:       constants.AgentDev.Name,
			Content:         "项目代码已成功提交到GitLab",
			IsMarkdown:      true,
			MarkdownContent: project.ID + ", " + project.Name + "\r\n" + project.ProjectPath,
			IsExpanded:      true,
		}

		if err := s.projectMsgRepo.Create(ctx, projectMsg3); err != nil {
			logger.Error("保存项目消息失败",
				logger.String("error", err.Error()),
				logger.String("projectID", project.ID),
			)
		}
	}

	// asynq 异步调用 agents-server 的接口
	if project.GitlabRepoURL != "" {
		taskInfo, err := s.asyncClient.Enqueue(tasks.NewProjectDevelopmentTask(project.ID, project.GitlabRepoURL))
		if err != nil {
			logger.Error("创建项目开发任务失败",
				logger.String("error", err.Error()),
				logger.String("projectID", project.ID),
			)
		} else {
			logger.Info("项目创建完成，开发流程已启动",
				logger.String("projectID", project.ID),
				logger.String("projectName", project.Name),
				logger.String("status", project.Status),
			)

			projectMsg4 := &models.ConversationMessage{
				ProjectID:  project.ID,
				Type:       constants.ConversationTypeSystem,
				AgentRole:  constants.AgentPM.Role,
				AgentName:  constants.AgentPM.Name,
				Content:    "项目创建完成，开发流程已启动",
				IsMarkdown: true,
				MarkdownContent: "```json\n{\n\"id\": \"" + project.ID +
					"\",\n\"name\": \"" + project.Name +
					"\",\n\"path\":\"" + project.ProjectPath +
					"\",\n \"taskID\": \"" + taskInfo.ID + "\"\n}```",
				IsExpanded: true,
			}

			if err := s.projectMsgRepo.Create(ctx, projectMsg4); err != nil {
				logger.Error("保存项目消息失败",
					logger.String("error", err.Error()),
					logger.String("projectID", project.ID),
				)
			}
		}
	}

}

// GetProject 获取项目信息
func (s *projectService) GetProject(ctx context.Context, projectID, userID string) (*models.ProjectInfo, error) {
	project, err := s.CheckProjectAccess(ctx, projectID, userID)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, errors.New("access denied")
	}
	return s.convertToProjectInfo(project), nil
}

// DeleteProject 删除项目
func (s *projectService) DeleteProject(ctx context.Context, projectID, userID string) error {
	// 检查权限
	isOwner, err := s.projectRepo.IsOwner(ctx, projectID, userID)
	if err != nil {
		return err
	}
	if !isOwner {
		return errors.New("access denied")
	}

	// 获取项目信息
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return fmt.Errorf("获取项目信息失败: %w", err)
	}

	// 如果项目路径存在，异步打包缓存
	if project.ProjectPath != "" && utils.IsDirectoryExists(project.ProjectPath) == true {
		s.asyncClient.Enqueue(tasks.NewProjectBackupTask(projectID, project.ProjectPath))
	}

	return s.projectRepo.Delete(ctx, projectID)
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
		projectInfos[i] = s.convertToProjectInfo(project)
	}

	// 构建分页响应
	totalPages := (int(total) + req.PageSize - 1) / req.PageSize
	pagination := &models.PaginationResponse{
		Code:        models.SUCCESS_CODE,
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
		projectInfos[i] = s.convertToProjectInfo(project)
	}

	// 构建分页响应
	totalPages := (int(total) + req.PageSize - 1) / req.PageSize
	pagination := &models.PaginationResponse{
		Code:        models.SUCCESS_CODE,
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

// GetProjectStages 获取项目开发阶段
func (s *projectService) GetProjectStages(ctx context.Context, projectID string) ([]*models.DevStage, error) {
	return s.projectStageService.GetProjectStages(ctx, projectID)
}

// 检查项目访问权限
func (s *projectService) CheckProjectAccess(ctx context.Context, projectID, userID string) (*models.Project, error) {
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	// 检查权限（用户只能查看自己的项目，管理员可以查看所有项目）
	// 这里简化处理，实际应该从JWT中获取用户角色
	isOwner, err := s.projectRepo.IsOwner(ctx, projectID, userID)
	if err != nil {
		return nil, err
	}
	if !isOwner {
		return nil, errors.New("access denied")
	}
	return project, nil
}

// convertToProjectInfo 将Project模型转换为ProjectInfo响应格式
func (s *projectService) convertToProjectInfo(project *models.Project) *models.ProjectInfo {
	projectInfo := &models.ProjectInfo{
		ID:           project.ID,
		Name:         project.Name,
		Description:  project.Description,
		Status:       project.Status,
		Requirements: project.Requirements,
		ProjectPath:  project.ProjectPath,
		BackendPort:  project.BackendPort,
		FrontendPort: project.FrontendPort,
		UserID:       project.UserID,
		CreatedAt:    project.CreatedAt,
		UpdatedAt:    project.UpdatedAt,
	}

	// 转换用户信息
	if project.User.ID != "" {
		projectInfo.User = models.UserInfo{
			ID:        project.User.ID,
			Email:     project.User.Email,
			Username:  project.User.Username,
			Role:      project.User.Role,
			Status:    project.User.Status,
			CreatedAt: project.User.CreatedAt,
		}
	}

	return projectInfo
}

// DownloadProject 下载项目文件
func (s *projectService) CreateDownloadProjectTask(ctx context.Context, projectID, projectPath string) (string, error) {
	// 检查项目路径是否存在
	if utils.IsDirectoryExists(projectPath) == false {
		logger.Error("项目路径为空", logger.String("projectPath", projectPath))
		return "", fmt.Errorf("项目路径为空")
	}

	// 异步方法，返回任务 ID
	info, err := s.asyncClient.Enqueue(tasks.NewProjectDownloadTask(projectID, projectPath))
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
		ProjectID:     project.ID,
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

	project.Status = "in_progress"
	project.GitlabRepoURL = gitConfig.ProjectPath
	s.projectRepo.Update(ctx, project)

	logger.Info("项目代码已成功提交到GitLab",
		logger.String("projectID", project.ID),
		logger.String("userID", project.UserID),
	)
	return nil
}
