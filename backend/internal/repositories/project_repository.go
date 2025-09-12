package repositories

import (
	"context"
	"errors"
	"fmt"
	"slices"
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
	IsPortAvailable(ctx context.Context, port int, portType string) (bool, error)
	GetNextAvailablePorts(ctx context.Context) (*models.Ports, error)

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
	query = query.Preload("User")

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
// TODO: 改成通过当前表中最大的，获取到下一个可用的端口
func (r *projectRepository) IsPortAvailable(ctx context.Context, port int, portType string) (bool, error) {
	var count int64
	var query *gorm.DB

	// 常用的端口
	commonPorts := []int{80, 443, 3000, 5432, 6379, 8080, 8081, 8082, 8083, 8888, 8098}

	// 检查端口是否在常用端口中
	if slices.Contains(commonPorts, port) {
		return false, nil
	}

	switch portType {
	case "backend":
		query = r.db.WithContext(ctx).Model(&models.Project{}).Where("backend_port = ?", port)
	case "frontend":
		query = r.db.WithContext(ctx).Model(&models.Project{}).Where("frontend_port = ?", port)
	case "redis":
		query = r.db.WithContext(ctx).Model(&models.Project{}).Where("redis_port = ?", port)
	case "postgres":
		query = r.db.WithContext(ctx).Model(&models.Project{}).Where("postgres_port = ?", port)
	default:
		return false, fmt.Errorf("invalid port type: %s", portType)
	}

	err := query.Count(&count).Error
	if err != nil {
		return false, err
	}

	return count == 0, nil
}

// GetNextAvailablePorts 获取下一个可用的端口
func (r *projectRepository) GetNextAvailablePorts(ctx context.Context) (*models.Ports, error) {
	// 从默认端口开始查找
	ports := models.Ports{
		BackendPort:  9501,
		FrontendPort: 3501,
		RedisPort:    7501,
		PostgresPort: 5501,
	}

	// 查找可用的后端端口
	for {
		available, err := r.IsPortAvailable(ctx, ports.BackendPort, "backend")
		if err != nil {
			return nil, fmt.Errorf("failed to check backend port %d: %w", ports.BackendPort, err)
		}
		if available {
			break
		}
		logger.Info("后端端口被占用，尝试下一个端口", logger.Int("port", ports.BackendPort))
		ports.BackendPort++
	}

	// 查找可用的前端端口
	for {
		available, err := r.IsPortAvailable(ctx, ports.FrontendPort, "frontend")
		if err != nil {
			return nil, fmt.Errorf("failed to check frontend port %d: %w", ports.FrontendPort, err)
		}
		if available {
			break
		}
		logger.Info("前端端口被占用，尝试下一个端口", logger.Int("port", ports.FrontendPort))
		ports.FrontendPort++
	}

	// 查找可用的Redis端口
	for {
		available, err := r.IsPortAvailable(ctx, ports.RedisPort, "redis")
		if err != nil {
			return nil, fmt.Errorf("failed to check redis port %d: %w", ports.RedisPort, err)
		}
		if available {
			break
		}
		logger.Info("Redis端口被占用，尝试下一个端口", logger.Int("port", ports.RedisPort))
		ports.RedisPort++
	}

	// 查找可用的Postgres端口
	for {
		available, err := r.IsPortAvailable(ctx, ports.PostgresPort, "postgres")
		if err != nil {
			return nil, fmt.Errorf("failed to check postgres port %d: %w", ports.PostgresPort, err)
		}
		if available {
			break
		}
		logger.Info("Postgres端口被占用，尝试下一个端口", logger.Int("port", ports.PostgresPort))
		ports.PostgresPort++
	}

	return &ports, nil
}
