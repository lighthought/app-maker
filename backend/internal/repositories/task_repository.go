package repositories

import (
	"context"
	"errors"
	"autocodeweb-backend/internal/models"
	"gorm.io/gorm"
)

// TaskRepository 任务仓储接口
type TaskRepository interface {
	// 基础CRUD操作
	Create(ctx context.Context, task *models.Task) error
	GetByID(ctx context.Context, taskID string) (*models.Task, error)
	Update(ctx context.Context, task *models.Task) error
	Delete(ctx context.Context, taskID string) error
	List(ctx context.Context, projectID string, limit, offset int) ([]*models.Task, error)

	// 任务状态管理
	UpdateStatus(ctx context.Context, taskID, status string) error
	GetByStatus(ctx context.Context, status string) ([]*models.Task, error)
	GetByProjectID(ctx context.Context, projectID string) ([]*models.Task, error)

	// 任务日志管理
	CreateLog(ctx context.Context, log *models.TaskLog) error
	GetLogs(ctx context.Context, taskID string, limit, offset int) ([]*models.TaskLog, error)
	GetLatestLogs(ctx context.Context, taskID string, limit int) ([]*models.TaskLog, error)
}

// taskRepository 任务仓储实现
type taskRepository struct {
	db *gorm.DB
}

// NewTaskRepository 创建任务仓储实例
func NewTaskRepository(db *gorm.DB) TaskRepository {
	return &taskRepository{db: db}
}

// Create 创建任务
func (r *taskRepository) Create(ctx context.Context, task *models.Task) error {
	return r.db.WithContext(ctx).Create(task).Error
}

// GetByID 根据ID获取任务
func (r *taskRepository) GetByID(ctx context.Context, taskID string) (*models.Task, error) {
	var task models.Task
	err := r.db.WithContext(ctx).
		Preload("Project").
		Preload("Logs").
		Where("id = ?", taskID).
		First(&task).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &task, nil
}

// Update 更新任务
func (r *taskRepository) Update(ctx context.Context, task *models.Task) error {
	return r.db.WithContext(ctx).Save(task).Error
}

// Delete 删除任务
func (r *taskRepository) Delete(ctx context.Context, taskID string) error {
	return r.db.WithContext(ctx).Delete(&models.Task{}, "id = ?", taskID).Error
}

// List 获取任务列表
func (r *taskRepository) List(ctx context.Context, projectID string, limit, offset int) ([]*models.Task, error) {
	var tasks []*models.Task
	query := r.db.WithContext(ctx).Model(&models.Task{})
	
	if projectID != "" {
		query = query.Where("project_id = ?", projectID)
	}
	
	err := query.
		Preload("Project").
		Preload("Logs").
		Order("priority ASC, created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&tasks).Error
	
	return tasks, err
}

// UpdateStatus 更新任务状态
func (r *taskRepository) UpdateStatus(ctx context.Context, taskID, status string) error {
	return r.db.WithContext(ctx).
		Model(&models.Task{}).
		Where("id = ?", taskID).
		Update("status", status).Error
}

// GetByStatus 根据状态获取任务
func (r *taskRepository) GetByStatus(ctx context.Context, status string) ([]*models.Task, error) {
	var tasks []*models.Task
	err := r.db.WithContext(ctx).
		Preload("Project").
		Preload("Logs").
		Where("status = ?", status).
		Order("priority ASC, created_at ASC").
		Find(&tasks).Error
	return tasks, err
}

// GetByProjectID 根据项目ID获取任务
func (r *taskRepository) GetByProjectID(ctx context.Context, projectID string) ([]*models.Task, error) {
	var tasks []*models.Task
	err := r.db.WithContext(ctx).
		Preload("Project").
		Preload("Logs").
		Where("project_id = ?", projectID).
		Order("priority ASC, created_at DESC").
		Find(&tasks).Error
	return tasks, err
}

// CreateLog 创建任务日志
func (r *taskRepository) CreateLog(ctx context.Context, log *models.TaskLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// GetLogs 获取任务日志
func (r *taskRepository) GetLogs(ctx context.Context, taskID string, limit, offset int) ([]*models.TaskLog, error) {
	var logs []*models.TaskLog
	err := r.db.WithContext(ctx).
		Where("task_id = ?", taskID).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&logs).Error
	return logs, err
}

// GetLatestLogs 获取最新的任务日志
func (r *taskRepository) GetLatestLogs(ctx context.Context, taskID string, limit int) ([]*models.TaskLog, error) {
	var logs []*models.TaskLog
	err := r.db.WithContext(ctx).
		Where("task_id = ?", taskID).
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}
