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
)

// ProjectService 项目服务接口
type ProjectService interface {
	// 基础CRUD操作
	CreateProject(ctx context.Context, req *models.CreateProjectRequest, userID string) (*models.ProjectInfo, error)
	GetProject(ctx context.Context, projectID, userID string) (*models.ProjectInfo, error)
	UpdateProject(ctx context.Context, projectID string, req *models.UpdateProjectRequest, userID string) (*models.ProjectInfo, error)
	DeleteProject(ctx context.Context, projectID, userID string) error
	ListProjects(ctx context.Context, req *models.ProjectListRequest, userID string) ([]*models.ProjectInfo, *models.PaginationResponse, error)

	// 项目状态管理
	UpdateProjectStatus(ctx context.Context, projectID, status, userID string) error
	GetProjectsByStatus(ctx context.Context, status, userID string) ([]*models.ProjectInfo, error)

	// 项目标签管理
	AddProjectTags(ctx context.Context, projectID string, tagIDs []string, userID string) error
	RemoveProjectTags(ctx context.Context, projectID string, tagIDs []string, userID string) error
	GetProjectTags(ctx context.Context, projectID, userID string) ([]*models.TagInfo, error)

	// 项目路径管理
	UpdateProjectPath(ctx context.Context, projectID, projectPath, userID string) error
	GetProjectByPath(ctx context.Context, projectPath, userID string) (*models.ProjectInfo, error)

	// 用户项目管理
	GetUserProjects(ctx context.Context, userID string, req *models.ProjectListRequest) ([]*models.ProjectInfo, *models.PaginationResponse, error)

	// 项目下载
	DownloadProject(ctx context.Context, projectID, userID string) ([]byte, error)
}

// projectService 项目服务实现
type projectService struct {
	projectRepo         repositories.ProjectRepository
	tagRepo             repositories.TagRepository
	templateService     ProjectTemplateService
	taskExecutionService *TaskExecutionService
	nameGenerator       ProjectNameGenerator
	zipUtils            *utils.ZipUtils
}

// NewProjectService 创建项目服务实例
func NewProjectService(
	projectRepo repositories.ProjectRepository, 
	tagRepo repositories.TagRepository, 
	templateService ProjectTemplateService,
	taskExecutionService *TaskExecutionService,
) ProjectService {
	return &projectService{
		projectRepo:         projectRepo,
		tagRepo:            tagRepo,
		templateService:    templateService,
		taskExecutionService: taskExecutionService,
		nameGenerator:      NewProjectNameGenerator(),
		zipUtils:           utils.NewZipUtils(),
	}
}

// CreateProject 创建项目
func (s *projectService) CreateProject(ctx context.Context, req *models.CreateProjectRequest, userID string) (*models.ProjectInfo, error) {
	logger.Info("开始创建项目",
		logger.String("userID", userID),
		logger.String("requirements", req.Requirements),
	)

	projectConfig := &models.Project{}
	// 自动生成项目配置信息和密码信息
	bGerated := s.nameGenerator.GenerateProjectConfig(req.Requirements, projectConfig)
	if !bGerated {
		logger.Error("自动生成项目配置信息失败", logger.String("requirements", req.Requirements))
		return nil, fmt.Errorf("failed to generate project config: %w", errors.New("failed to generate project config"))
	}

	logger.Info("自动生成项目名", logger.String("projectName", projectConfig.Name))

	// 生成项目路径 - 使用 /app/projects 而不是 /projects
	projectPath := filepath.Join("/app/data/projects", userID, uuid.New().String())
	logger.Info("生成项目路径", logger.String("projectPath", projectPath))

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

	projectConfig.BackendPort = availableBackendPort
	projectConfig.FrontendPort = availableFrontendPort

	// TODO: 获取下一个可用的子网段
	if projectConfig.Subnetwork == "" {
		projectConfig.Subnetwork = "172.20.0.0/16"
	}

	// 创建项目
	project := &models.Project{
		Name:             projectConfig.Name,
		Description:      projectConfig.Description,
		Requirements:     req.Requirements,
		BackendPort:      projectConfig.BackendPort,
		FrontendPort:     projectConfig.FrontendPort,
		ApiBaseUrl:       projectConfig.ApiBaseUrl,
		AppSecretKey:     projectConfig.AppSecretKey,
		DatabasePassword: projectConfig.DatabasePassword,
		RedisPassword:    projectConfig.RedisPassword,
		JwtSecretKey:     projectConfig.JwtSecretKey,
		Subnetwork:       projectConfig.Subnetwork,
		UserID:           userID,
		Status:           "draft",
		ProjectPath:      projectPath,
	}

	logger.Info("保存项目到数据库")
	if err := s.projectRepo.Create(ctx, project); err != nil {
		logger.Error("保存项目到数据库失败",
			logger.String("error", err.Error()),
			logger.String("projectID", project.ID),
		)
		return nil, fmt.Errorf("failed to create project: %w", err)
	}
	logger.Info("项目保存成功", logger.String("projectID", project.ID))

	// 初始化项目模板
	logger.Info("开始初始化项目模板",
		logger.String("projectID", project.ID),
		logger.String("projectPath", projectPath),
	)
	if err := s.templateService.InitializeProject(ctx, project); err != nil {
		// 模板初始化失败不影响项目创建，但记录错误
		logger.Error("项目模板初始化失败",
			logger.String("error", err.Error()),
			logger.String("projectID", project.ID),
			logger.String("projectPath", projectPath),
		)
	} else {
		logger.Info("项目模板初始化成功",
			logger.String("projectID", project.ID),
			logger.String("projectPath", projectPath),
		)
	}

	// 添加标签
	// if len(tagIDs) > 0 {
	// 	logger.Info("开始添加项目标签",
	// 		logger.String("projectID", project.ID),
	// 		logger.String("tagIDs", fmt.Sprintf("%v", req.TagIDs)),
	// 	)
	// 	if err := s.projectRepo.AddTags(ctx, project.ID, req.TagIDs); err != nil {
	// 		// 标签添加失败不影响项目创建，只记录日志
	// 		logger.Error("添加项目标签失败",
	// 			logger.String("error", err.Error()),
	// 			logger.String("projectID", project.ID),
	// 			logger.String("tagIDs", fmt.Sprintf("%v", req.TagIDs)),
	// 		)
	// 	} else {
	// 		logger.Info("项目标签添加成功",
	// 			logger.String("projectID", project.ID),
	// 			logger.String("tagIDs", fmt.Sprintf("%v", req.TagIDs)),
	// 		)
	// 	}
	// }

	// 获取创建后的项目信息
	logger.Info("获取创建后的项目信息", logger.String("projectID", project.ID))
	projectInfo, err := s.GetProject(ctx, project.ID, userID)
	if err != nil {
		logger.Error("获取项目信息失败",
			logger.String("error", err.Error()),
			logger.String("projectID", project.ID),
		)
		return nil, err
	}

	// 启动项目开发流程（异步）
	go func() {
		if err := s.taskExecutionService.StartProjectDevelopment(context.Background(), project.ID); err != nil {
			logger.Error("启动项目开发流程失败",
				logger.String("error", err.Error()),
				logger.String("projectID", project.ID),
			)
		}
	}()

	logger.Info("项目创建完成，开发流程已启动",
		logger.String("projectID", project.ID),
		logger.String("projectName", project.Name),
		logger.String("status", project.Status),
	)

	return projectInfo, nil
}

// GetProject 获取项目信息
func (s *projectService) GetProject(ctx context.Context, projectID, userID string) (*models.ProjectInfo, error) {
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

	return s.convertToProjectInfo(project), nil
}

// UpdateProject 更新项目
func (s *projectService) UpdateProject(ctx context.Context, projectID string, req *models.UpdateProjectRequest, userID string) (*models.ProjectInfo, error) {
	// 检查权限
	isOwner, err := s.projectRepo.IsOwner(ctx, projectID, userID)
	if err != nil {
		return nil, err
	}
	if !isOwner {
		return nil, errors.New("access denied")
	}

	// 获取现有项目
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	// 更新字段
	if req.Name != "" {
		project.Name = req.Name
	}
	if req.Description != "" {
		project.Description = req.Description
	}
	if req.Requirements != "" {
		project.Requirements = req.Requirements
	}
	if req.BackendPort > 0 {
		project.BackendPort = req.BackendPort
	}
	if req.FrontendPort > 0 {
		project.FrontendPort = req.FrontendPort
	}
	if req.Status != "" {
		project.Status = req.Status
	}

	project.UpdatedAt = time.Now()

	// 保存更新
	if err := s.projectRepo.Update(ctx, project); err != nil {
		return nil, fmt.Errorf("failed to update project: %w", err)
	}

	// 更新标签
	if req.TagIDs != nil {
		// 先移除所有现有标签
		currentTags, err := s.projectRepo.GetTags(ctx, projectID)
		if err == nil {
			var currentTagIDs []string
			for _, tag := range currentTags {
				currentTagIDs = append(currentTagIDs, tag.ID)
			}
			if len(currentTagIDs) > 0 {
				s.projectRepo.RemoveTags(ctx, projectID, currentTagIDs)
			}
		}

		// 添加新标签
		if len(req.TagIDs) > 0 {
			if err := s.projectRepo.AddTags(ctx, projectID, req.TagIDs); err != nil {
				// 标签更新失败不影响项目更新
			}
		}
	}

	// 获取更新后的项目信息
	return s.GetProject(ctx, projectID, userID)
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
		if _, err := os.Stat(project.ProjectPath); err == nil {
			// 创建缓存目录
			cacheDir := "/app/data/projects/cache"

			// 生成缓存文件名
			cacheFileName := fmt.Sprintf("%s_%s_%s", project.ID, project.Name, time.Now().Format("20060102_150405"))

			// 使用 zipUtils 压缩到缓存
			_, err := s.zipUtils.CompressDirectoryToCache(ctx, project.ProjectPath, cacheDir, cacheFileName)
			if err != nil {
				logger.Error("打包项目到缓存失败",
					logger.String("projectID", projectID),
					logger.ErrorField(err))
			} else {
				logger.Info("项目已打包到缓存",
					logger.String("projectID", projectID))
			}
		}
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

// UpdateProjectStatus 更新项目状态
func (s *projectService) UpdateProjectStatus(ctx context.Context, projectID, status, userID string) error {
	// 检查权限
	isOwner, err := s.projectRepo.IsOwner(ctx, projectID, userID)
	if err != nil {
		return err
	}
	if !isOwner {
		return errors.New("access denied")
	}

	// 验证状态值
	validStatuses := map[string]bool{
		"draft":       true,
		"in_progress": true,
		"completed":   true,
		"failed":      true,
	}

	if !validStatuses[status] {
		return fmt.Errorf("invalid status: %s", status)
	}

	return s.projectRepo.UpdateStatus(ctx, projectID, status)
}

// GetProjectsByStatus 根据状态获取项目
func (s *projectService) GetProjectsByStatus(ctx context.Context, status, userID string) ([]*models.ProjectInfo, error) {
	projects, err := s.projectRepo.GetByStatus(ctx, status, userID)
	if err != nil {
		return nil, err
	}

	projectInfos := make([]*models.ProjectInfo, len(projects))
	for i, project := range projects {
		projectInfos[i] = s.convertToProjectInfo(project)
	}

	return projectInfos, nil
}

// AddProjectTags 为项目添加标签
func (s *projectService) AddProjectTags(ctx context.Context, projectID string, tagIDs []string, userID string) error {
	// 检查权限
	isOwner, err := s.projectRepo.IsOwner(ctx, projectID, userID)
	if err != nil {
		return err
	}
	if !isOwner {
		return errors.New("access denied")
	}

	return s.projectRepo.AddTags(ctx, projectID, tagIDs)
}

// RemoveProjectTags 从项目移除标签
func (s *projectService) RemoveProjectTags(ctx context.Context, projectID string, tagIDs []string, userID string) error {
	// 检查权限
	isOwner, err := s.projectRepo.IsOwner(ctx, projectID, userID)
	if err != nil {
		return err
	}
	if !isOwner {
		return errors.New("access denied")
	}

	return s.projectRepo.RemoveTags(ctx, projectID, tagIDs)
}

// GetProjectTags 获取项目标签
func (s *projectService) GetProjectTags(ctx context.Context, projectID, userID string) ([]*models.TagInfo, error) {
	// 检查权限
	isOwner, err := s.projectRepo.IsOwner(ctx, projectID, userID)
	if err != nil {
		return nil, err
	}
	if !isOwner {
		return nil, errors.New("access denied")
	}

	tags, err := s.projectRepo.GetTags(ctx, projectID)
	if err != nil {
		return nil, err
	}

	tagInfos := make([]*models.TagInfo, len(tags))
	for i, tag := range tags {
		tagInfos[i] = &models.TagInfo{
			ID:    tag.ID,
			Name:  tag.Name,
			Color: tag.Color,
		}
	}

	return tagInfos, nil
}

// UpdateProjectPath 更新项目路径
func (s *projectService) UpdateProjectPath(ctx context.Context, projectID, projectPath, userID string) error {
	// 检查权限
	isOwner, err := s.projectRepo.IsOwner(ctx, projectID, userID)
	if err != nil {
		return err
	}
	if !isOwner {
		return errors.New("access denied")
	}

	return s.projectRepo.UpdateProjectPath(ctx, projectID, projectPath)
}

// GetProjectByPath 根据路径获取项目
func (s *projectService) GetProjectByPath(ctx context.Context, projectPath, userID string) (*models.ProjectInfo, error) {
	project, err := s.projectRepo.GetByProjectPath(ctx, projectPath)
	if err != nil {
		return nil, err
	}

	// 检查权限
	isOwner, err := s.projectRepo.IsOwner(ctx, project.ID, userID)
	if err != nil {
		return nil, err
	}
	if !isOwner {
		return nil, errors.New("access denied")
	}

	return s.convertToProjectInfo(project), nil
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

	// 转换标签信息
	if len(project.Tags) > 0 {
		projectInfo.Tags = make([]models.TagInfo, len(project.Tags))
		for i, tag := range project.Tags {
			projectInfo.Tags[i] = models.TagInfo{
				ID:    tag.ID,
				Name:  tag.Name,
				Color: tag.Color,
			}
		}
	}

	return projectInfo
}

// DownloadProject 下载项目文件
func (s *projectService) DownloadProject(ctx context.Context, projectID, userID string) ([]byte, error) {
	// 检查权限
	isOwner, err := s.projectRepo.IsOwner(ctx, projectID, userID)
	if err != nil {
		return nil, err
	}
	if !isOwner {
		return nil, errors.New("access denied")
	}

	// 获取项目信息
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	// 检查项目路径是否存在
	if project.ProjectPath == "" {
		return nil, fmt.Errorf("项目路径为空")
	}

	// 使用 zipUtils 压缩项目文件
	zipData, err := s.zipUtils.CompressDirectoryToBytes(ctx, project.ProjectPath)
	if err != nil {
		return nil, fmt.Errorf("打包项目文件失败: %w", err)
	}

	return zipData, nil
}
