package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"autocodeweb-backend/internal/models"

	"gorm.io/gorm"
)

// ProjectRepository 项目仓库接口
type ProjectRepository interface {
	// 基础CRUD操作
	Create(ctx context.Context, project *models.Project) error
	GetByID(ctx context.Context, id string) (*models.Project, error)
	Update(ctx context.Context, project *models.Project) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, req *models.ProjectListRequest) ([]*models.Project, int64, error)

	// 项目状态管理
	UpdateStatus(ctx context.Context, id string, status string) error
	GetByStatus(ctx context.Context, status string, userID string) ([]*models.Project, error)

	// 项目标签管理
	AddTags(ctx context.Context, projectID string, tagIDs []string) error
	RemoveTags(ctx context.Context, projectID string, tagIDs []string) error
	GetTags(ctx context.Context, projectID string) ([]*models.Tag, error)

	// 项目路径管理
	GetByProjectPath(ctx context.Context, projectPath string) (*models.Project, error)
	UpdateProjectPath(ctx context.Context, id string, projectPath string) error

	// 用户权限检查
	IsOwner(ctx context.Context, projectID, userID string) (bool, error)
	GetByUserID(ctx context.Context, userID string, req *models.ProjectListRequest) ([]*models.Project, int64, error)
}

// projectRepository 项目仓库实现
type projectRepository struct {
	db *gorm.DB
}

// NewProjectRepository 创建项目仓库实例
func NewProjectRepository(db *gorm.DB) ProjectRepository {
	return &projectRepository{db: db}
}

// Create 创建项目
func (r *projectRepository) Create(ctx context.Context, project *models.Project) error {
	return r.db.WithContext(ctx).Create(project).Error
}

// GetByID 根据ID获取项目
func (r *projectRepository) GetByID(ctx context.Context, id string) (*models.Project, error) {
	var project models.Project
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Tags").
		Preload("Tasks").
		Where("id = ?", id).
		First(&project).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("project not found: %s", id)
		}
		return nil, err
	}
	return &project, nil
}

// Update 更新项目
func (r *projectRepository) Update(ctx context.Context, project *models.Project) error {
	return r.db.WithContext(ctx).Save(project).Error
}

// Delete 删除项目
func (r *projectRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.Project{}, "id = ?", id).Error
}

// List 获取项目列表
func (r *projectRepository) List(ctx context.Context, req *models.ProjectListRequest) ([]*models.Project, int64, error) {
	var projects []*models.Project
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Project{})

	// 应用过滤条件
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}
	if req.UserID != "" {
		query = query.Where("user_id = ?", req.UserID)
	}
	if req.Search != "" {
		searchTerm := "%" + strings.ToLower(req.Search) + "%"
		query = query.Where("LOWER(name) LIKE ? OR LOWER(description) LIKE ?", searchTerm, searchTerm)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 应用分页
	offset := (req.Page - 1) * req.PageSize
	query = query.Offset(offset).Limit(req.PageSize)

	// 预加载关联数据
	query = query.Preload("User").Preload("Tags").Preload("Tasks")

	// 排序
	query = query.Order("created_at DESC")

	// 执行查询
	if err := query.Find(&projects).Error; err != nil {
		return nil, 0, err
	}

	return projects, total, nil
}

// UpdateStatus 更新项目状态
func (r *projectRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	return r.db.WithContext(ctx).Model(&models.Project{}).
		Where("id = ?", id).
		Update("status", status).Error
}

// GetByStatus 根据状态获取项目
func (r *projectRepository) GetByStatus(ctx context.Context, status string, userID string) ([]*models.Project, error) {
	var projects []*models.Project
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Tags").
		Where("status = ? AND user_id = ?", status, userID).
		Order("created_at DESC").
		Find(&projects).Error
	return projects, err
}

// AddTags 为项目添加标签
func (r *projectRepository) AddTags(ctx context.Context, projectID string, tagIDs []string) error {
	if len(tagIDs) == 0 {
		return nil
	}

	// 检查标签是否存在
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Tag{}).
		Where("id IN ?", tagIDs).
		Count(&count).Error; err != nil {
		return err
	}
	if int(count) != len(tagIDs) {
		return errors.New("some tags not found")
	}

	// 添加项目标签关联
	var projectTags []models.ProjectTag
	for _, tagID := range tagIDs {
		projectTags = append(projectTags, models.ProjectTag{
			ProjectID: projectID,
			TagID:     tagID,
		})
	}

	return r.db.WithContext(ctx).Create(&projectTags).Error
}

// RemoveTags 从项目移除标签
func (r *projectRepository) RemoveTags(ctx context.Context, projectID string, tagIDs []string) error {
	if len(tagIDs) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).
		Where("project_id = ? AND tag_id IN ?", projectID, tagIDs).
		Delete(&models.ProjectTag{}).Error
}

// GetTags 获取项目标签
func (r *projectRepository) GetTags(ctx context.Context, projectID string) ([]*models.Tag, error) {
	var tags []*models.Tag
	err := r.db.WithContext(ctx).
		Joins("JOIN project_tags ON tags.id = project_tags.tag_id").
		Where("project_tags.project_id = ?", projectID).
		Find(&tags).Error
	return tags, err
}

// GetByProjectPath 根据项目路径获取项目
func (r *projectRepository) GetByProjectPath(ctx context.Context, projectPath string) (*models.Project, error) {
	var project models.Project
	err := r.db.WithContext(ctx).
		Where("project_path = ?", projectPath).
		First(&project).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("project not found with path: %s", projectPath)
		}
		return nil, err
	}
	return &project, nil
}

// UpdateProjectPath 更新项目路径
func (r *projectRepository) UpdateProjectPath(ctx context.Context, id string, projectPath string) error {
	return r.db.WithContext(ctx).Model(&models.Project{}).
		Where("id = ?", id).
		Update("project_path", projectPath).Error
}

// IsOwner 检查用户是否为项目所有者
func (r *projectRepository) IsOwner(ctx context.Context, projectID, userID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Project{}).
		Where("id = ? AND user_id = ?", projectID, userID).
		Count(&count).Error
	return count > 0, err
}

// GetByUserID 获取用户的项目列表
func (r *projectRepository) GetByUserID(ctx context.Context, userID string, req *models.ProjectListRequest) ([]*models.Project, int64, error) {
	req.UserID = userID
	return r.List(ctx, req)
}
