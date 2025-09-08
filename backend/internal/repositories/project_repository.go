package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"autocodeweb-backend/internal/models"
	"autocodeweb-backend/pkg/logger"

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

	// 端口管理
	GetAvailablePorts(ctx context.Context, backendPort, frontendPort int) (int, int, error)
	IsPortAvailable(ctx context.Context, port int, portType string) (bool, error)
	GetNextAvailablePorts(ctx context.Context) (int, int, error)

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

// IsPortAvailable 检查端口是否可用
func (r *projectRepository) IsPortAvailable(ctx context.Context, port int, portType string) (bool, error) {
	var count int64
	var query *gorm.DB

	switch portType {
	case "backend":
		query = r.db.WithContext(ctx).Model(&models.Project{}).Where("backend_port = ?", port)
	case "frontend":
		query = r.db.WithContext(ctx).Model(&models.Project{}).Where("frontend_port = ?", port)
	default:
		return false, fmt.Errorf("invalid port type: %s", portType)
	}

	err := query.Count(&count).Error
	if err != nil {
		return false, err
	}

	return count == 0, nil
}

// GetAvailablePorts 获取可用的端口
func (r *projectRepository) GetAvailablePorts(ctx context.Context, backendPort, frontendPort int) (int, int, error) {
	// 检查后端端口
	for {
		available, err := r.IsPortAvailable(ctx, backendPort, "backend")
		if err != nil {
			return 0, 0, fmt.Errorf("failed to check backend port %d: %w", backendPort, err)
		}
		if available {
			break
		}
		logger.Info("后端端口被占用，尝试下一个端口", logger.Int("port", backendPort))
		backendPort++
	}

	// 检查前端端口
	for {
		available, err := r.IsPortAvailable(ctx, frontendPort, "frontend")
		if err != nil {
			return 0, 0, fmt.Errorf("failed to check frontend port %d: %w", frontendPort, err)
		}
		if available {
			break
		}
		logger.Info("前端端口被占用，尝试下一个端口", logger.Int("port", frontendPort))
		frontendPort++
	}

	// 确保前后端端口不冲突
	if backendPort == frontendPort {
		frontendPort++
		// 再次检查前端端口是否可用
		for {
			available, err := r.IsPortAvailable(ctx, frontendPort, "frontend")
			if err != nil {
				return 0, 0, fmt.Errorf("failed to check frontend port %d: %w", frontendPort, err)
			}
			if available {
				break
			}
			logger.Info("前端端口被占用，尝试下一个端口", logger.Int("port", frontendPort))
			frontendPort++
		}
	}

	return backendPort, frontendPort, nil
}

// GetNextAvailablePorts 获取下一个可用的端口
func (r *projectRepository) GetNextAvailablePorts(ctx context.Context) (int, int, error) {
	// 从默认端口开始查找
	backendPort := 8081
	frontendPort := 3001

	// 查找可用的后端端口
	for {
		available, err := r.IsPortAvailable(ctx, backendPort, "backend")
		if err != nil {
			return 0, 0, fmt.Errorf("failed to check backend port %d: %w", backendPort, err)
		}
		if available {
			break
		}
		logger.Info("后端端口被占用，尝试下一个端口", logger.Int("port", backendPort))
		backendPort++
	}

	// 查找可用的前端端口
	for {
		available, err := r.IsPortAvailable(ctx, frontendPort, "frontend")
		if err != nil {
			return 0, 0, fmt.Errorf("failed to check frontend port %d: %w", frontendPort, err)
		}
		if available {
			break
		}
		logger.Info("前端端口被占用，尝试下一个端口", logger.Int("port", frontendPort))
		frontendPort++
	}

	// 确保前后端端口不冲突
	if backendPort == frontendPort {
		frontendPort++
		// 再次检查前端端口是否可用
		for {
			available, err := r.IsPortAvailable(ctx, frontendPort, "frontend")
			if err != nil {
				return 0, 0, fmt.Errorf("failed to check frontend port %d: %w", frontendPort, err)
			}
			if available {
				break
			}
			logger.Info("前端端口被占用，尝试下一个端口", logger.Int("port", frontendPort))
			frontendPort++
		}
	}

	return backendPort, frontendPort, nil
}
