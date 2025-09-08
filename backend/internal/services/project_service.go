package services

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"autocodeweb-backend/internal/models"
	"autocodeweb-backend/internal/repositories"
	"autocodeweb-backend/internal/utils"
	"autocodeweb-backend/pkg/logger"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ProjectService 项目服务接口
type ProjectService interface {
	// 基础CRUD操作
	CreateProject(ctx context.Context, req *models.CreateProjectRequest, userID string) (*models.ProjectInfo, error)
	GetProject(ctx context.Context, projectID, userID string) (*models.ProjectInfo, error)
	DeleteProject(ctx context.Context, projectID, userID string) error
	ListProjects(ctx context.Context, req *models.ProjectListRequest, userID string) ([]*models.ProjectInfo, *models.PaginationResponse, error)

	// 用户项目管理
	GetUserProjects(ctx context.Context, userID string, req *models.ProjectListRequest) ([]*models.ProjectInfo, *models.PaginationResponse, error)

	// 项目开发阶段管理
	GetProjectStages(ctx context.Context, projectID string) ([]*models.DevStage, error)

	// 检查项目访问权限
	CheckProjectAccess(ctx context.Context, projectID, userID string) (*models.Project, error)
}

// projectService 项目服务实现
type projectService struct {
	projectRepo         repositories.ProjectRepository
	templateService     ProjectTemplateService
	projectStageService *ProjectStageService
	nameGenerator       ProjectNameGenerator
	zipUtils            *utils.ZipUtils
	fileUtils           *utils.FileUtils
}

// NewProjectService 创建项目服务实例
func NewProjectService(
	db *gorm.DB,
	fileService FileService,
) ProjectService {
	projectRepo := repositories.NewProjectRepository(db)
	stageRepo := repositories.NewStageRepository(db)
	fileUtils := fileService.GetFileUtils()
	return &projectService{
		projectRepo:         projectRepo,
		templateService:     NewProjectTemplateService(fileService),
		projectStageService: NewProjectStageService(projectRepo, stageRepo, fileUtils),
		nameGenerator:       NewProjectNameGenerator(),
		zipUtils:            utils.NewZipUtils(fileUtils),
		fileUtils:           fileUtils,
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
	newProject.ProjectPath = s.fileUtils.GetProjectPath(userID, newProject.ID)
	logger.Info("生成项目路径", logger.String("projectPath", newProject.ProjectPath))

	// 自动生成项目配置信息和密码信息
	bGerated := s.nameGenerator.GenerateProjectConfig(req.Requirements, newProject)
	if !bGerated {
		logger.Error("自动生成项目配置信息失败", logger.String("requirements", req.Requirements))
		return nil, fmt.Errorf("failed to generate project config: %w", errors.New("failed to generate project config"))
	}

	logger.Info("自动生成项目名", logger.String("projectName", newProject.Name))

	// 自动获取可用端口
	availableBackendPort, availableFrontendPort, err := s.projectRepo.GetNextAvailablePorts(ctx)
	if err != nil {
		logger.Error("获取可用端口失败",
			logger.String("error", err.Error()),
		)
		return nil, fmt.Errorf("failed to get available ports: %w", err)
	}

	logger.Info("自动分配端口",
		logger.Int("backendPort", availableBackendPort),
		logger.Int("frontendPort", availableFrontendPort),
	)

	newProject.BackendPort = availableBackendPort
	newProject.FrontendPort = availableFrontendPort

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

	// 初始化项目模板
	logger.Info("开始初始化项目模板",
		logger.String("projectID", newProject.ID),
		logger.String("projectPath", newProject.ProjectPath),
	)
	if err := s.templateService.InitializeProject(ctx, newProject); err != nil {
		// 模板初始化失败不影响项目创建，但记录错误
		logger.Error("项目模板初始化失败",
			logger.String("error", err.Error()),
			logger.String("projectID", newProject.ID),
			logger.String("projectPath", newProject.ProjectPath),
		)
	} else {
		logger.Info("项目模板初始化成功",
			logger.String("projectID", newProject.ID),
			logger.String("projectPath", newProject.ProjectPath),
		)
	}

	// 获取创建后的项目信息
	logger.Info("获取创建后的项目信息", logger.String("projectID", newProject.ID))
	projectInfo, err := s.GetProject(ctx, newProject.ID, userID)
	if err != nil {
		logger.Error("获取项目信息失败",
			logger.String("error", err.Error()),
			logger.String("projectID", newProject.ID),
		)
		return nil, err
	}

	// 启动项目开发流程（异步）
	go func() {
		if err := s.projectStageService.StartProjectDevelopment(context.Background(), newProject.ID); err != nil {
			logger.Error("启动项目开发流程失败",
				logger.String("error", err.Error()),
				logger.String("projectID", newProject.ID),
			)
		}
	}()

	logger.Info("项目创建完成，开发流程已启动",
		logger.String("projectID", newProject.ID),
		logger.String("projectName", newProject.Name),
		logger.String("status", newProject.Status),
	)

	return projectInfo, nil
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

	// 如果项目路径存在，先打包缓存
	if project.ProjectPath != "" {
		// 异步打包项目到缓存
		go func(project *models.Project) {
			if _, err := os.Stat(project.ProjectPath); err == nil {
				// 创建缓存目录
				cacheDir := "/app/data/projects/cache"

				// 生成缓存文件名
				cacheFileName := fmt.Sprintf("%s_%s", project.ID, time.Now().Format("20060102_150405"))

				// 使用 zipUtils 压缩到缓存
				_, err := s.zipUtils.CompressDirectoryToCache(context.Background(), project.ProjectPath, cacheDir, cacheFileName)
				if err != nil {
					logger.Error("异步打包项目到缓存失败",
						logger.String("projectID", project.ID),
						logger.ErrorField(err))
				} else {
					logger.Info("项目已异步打包到缓存",
						logger.String("projectID", project.ID))
				}
			}
		}(project)
	}

	// 删除项目目录
	if project.ProjectPath != "" {
		if err := os.RemoveAll(project.ProjectPath); err != nil {
			logger.Error("删除项目目录失败",
				logger.String("projectID", projectID),
				logger.String("projectPath", project.ProjectPath),
				logger.ErrorField(err))
		} else {
			logger.Info("项目目录已删除",
				logger.String("projectID", projectID),
				logger.String("projectPath", project.ProjectPath))
		}
	}

	return s.projectRepo.Delete(ctx, projectID)
}

// ListProjects 获取项目列表
func (s *projectService) ListProjects(ctx context.Context, req *models.ProjectListRequest, userID string) ([]*models.ProjectInfo, *models.PaginationResponse, error) {
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
		return nil, nil, err
	}

	// 转换为响应格式
	projectInfos := make([]*models.ProjectInfo, len(projects))
	for i, project := range projects {
		projectInfos[i] = s.convertToProjectInfo(project)
	}

	// 构建分页响应
	totalPages := (int(total) + req.PageSize - 1) / req.PageSize
	pagination := &models.PaginationResponse{
		Total:       int(total),
		Page:        req.Page,
		PageSize:    req.PageSize,
		TotalPages:  totalPages,
		Data:        projectInfos,
		HasNext:     req.Page < totalPages,
		HasPrevious: req.Page > 1,
	}

	return projectInfos, pagination, nil
}

// GetUserProjects 获取用户的项目列表
func (s *projectService) GetUserProjects(ctx context.Context, userID string, req *models.ProjectListRequest) ([]*models.ProjectInfo, *models.PaginationResponse, error) {
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
		return nil, nil, err
	}

	// 转换为响应格式
	projectInfos := make([]*models.ProjectInfo, len(projects))
	for i, project := range projects {
		projectInfos[i] = s.convertToProjectInfo(project)
	}

	// 构建分页响应
	totalPages := (int(total) + req.PageSize - 1) / req.PageSize
	pagination := &models.PaginationResponse{
		Total:       int(total),
		Page:        req.Page,
		PageSize:    req.PageSize,
		TotalPages:  totalPages,
		Data:        projectInfos,
		HasNext:     req.Page < totalPages,
		HasPrevious: req.Page > 1,
	}

	return projectInfos, pagination, nil
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
