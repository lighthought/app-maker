package repositories

import (
	"context"
	"errors"
	"time"

	"autocodeweb-backend/internal/models"

	"gorm.io/gorm"
)

type TaskRepository interface {
	// 基础CRUD操作
	Create(ctx context.Context, task *models.Task) error
	GetByID(ctx context.Context, id string) (*models.Task, error)
	Update(ctx context.Context, task *models.Task) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, req *models.TaskListRequest) ([]*models.Task, int64, error)

	// 状态管理
	UpdateStatus(ctx context.Context, id string, status models.TaskStatus, errorMessage, result string) error
	GetByStatus(ctx context.Context, status models.TaskStatus) ([]*models.Task, error)
	GetReadyTasks(ctx context.Context) ([]*models.Task, error) // 获取可执行的任务（依赖已完成）

	// 依赖关系管理
	AddDependency(ctx context.Context, taskID, dependencyID string) error
	RemoveDependency(ctx context.Context, taskID, dependencyID string) error
	GetDependencies(ctx context.Context, taskID string) ([]*models.Task, error)
	GetDependents(ctx context.Context, taskID string) ([]*models.Task, error) // 获取依赖此任务的任务
	CheckDependenciesCompleted(ctx context.Context, taskID string) (bool, error)

	// 重试和回滚
	IncrementRetryCount(ctx context.Context, id string) error
	ResetRetryCount(ctx context.Context, id string) error
	GetFailedTasks(ctx context.Context) ([]*models.Task, error)

	// 项目相关
	GetByProjectID(ctx context.Context, projectID string) ([]*models.Task, error)
	GetByUserID(ctx context.Context, userID string) ([]*models.Task, error)

	// 优先级和排序
	GetByPriority(ctx context.Context, priority models.TaskPriority) ([]*models.Task, error)
	GetOverdueTasks(ctx context.Context) ([]*models.Task, error) // 获取超期任务
}

type taskRepository struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) TaskRepository {
	return &taskRepository{db: db}
}

// Create 创建任务
func (r *taskRepository) Create(ctx context.Context, task *models.Task) error {
	return r.db.WithContext(ctx).Create(task).Error
}

// GetByID 根据ID获取任务
func (r *taskRepository) GetByID(ctx context.Context, id string) (*models.Task, error) {
	var task models.Task
	err := r.db.WithContext(ctx).
		Preload("Project").
		Preload("User").
		Preload("Logs").
		Where("id = ?", id).
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
func (r *taskRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.Task{}, "id = ?", id).Error
}

// List 获取任务列表
func (r *taskRepository) List(ctx context.Context, req *models.TaskListRequest) ([]*models.Task, int64, error) {
	var tasks []*models.Task
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Task{})

	// 应用过滤条件
	if req.ProjectID != "" {
		query = query.Where("project_id = ?", req.ProjectID)
	}
	if req.UserID != "" {
		query = query.Where("user_id = ?", req.UserID)
	}
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}
	if req.Priority > 0 {
		query = query.Where("priority = ?", req.Priority)
	}
	if req.Search != "" {
		query = query.Where("name ILIKE ? OR description ILIKE ?", "%"+req.Search+"%", "%"+req.Search+"%")
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (req.Page - 1) * req.PageSize
	err := query.
		Preload("Project").
		Preload("User").
		Order("priority DESC, created_at DESC").
		Offset(offset).
		Limit(req.PageSize).
		Find(&tasks).Error

	return tasks, total, err
}

// UpdateStatus 更新任务状态
func (r *taskRepository) UpdateStatus(ctx context.Context, id string, status models.TaskStatus, errorMessage, result string) error {
	updates := map[string]interface{}{
		"status":        status,
		"updated_at":    time.Now(),
		"error_message": errorMessage,
		"result":        result,
	}

	// 根据状态设置时间
	switch status {
	case models.TaskStatusRunning:
		now := time.Now()
		updates["started_at"] = &now
	case models.TaskStatusCompleted, models.TaskStatusFailed, models.TaskStatusCancelled, models.TaskStatusRolledBack:
		now := time.Now()
		updates["completed_at"] = &now
	}

	return r.db.WithContext(ctx).Model(&models.Task{}).Where("id = ?", id).Updates(updates).Error
}

// GetByStatus 根据状态获取任务
func (r *taskRepository) GetByStatus(ctx context.Context, status models.TaskStatus) ([]*models.Task, error) {
	var tasks []*models.Task
	err := r.db.WithContext(ctx).
		Preload("Project").
		Preload("User").
		Where("status = ?", status).
		Order("priority DESC, created_at ASC").
		Find(&tasks).Error
	return tasks, err
}

// GetReadyTasks 获取可执行的任务（依赖已完成）
func (r *taskRepository) GetReadyTasks(ctx context.Context) ([]*models.Task, error) {
	var tasks []*models.Task

	// 获取所有待执行的任务
	err := r.db.WithContext(ctx).
		Preload("Project").
		Preload("User").
		Where("status = ?", models.TaskStatusPending).
		Order("priority DESC, created_at ASC").
		Find(&tasks).Error
	if err != nil {
		return nil, err
	}

	// 过滤出依赖已完成的任务
	var readyTasks []*models.Task
	for _, task := range tasks {
		if len(task.Dependencies) == 0 {
			readyTasks = append(readyTasks, task)
			continue
		}

		// 检查所有依赖是否已完成
		allCompleted := true
		for _, depID := range task.Dependencies {
			var depTask models.Task
			if err := r.db.WithContext(ctx).Where("id = ?", depID).First(&depTask).Error; err != nil {
				allCompleted = false
				break
			}
			if depTask.Status != models.TaskStatusCompleted {
				allCompleted = false
				break
			}
		}

		if allCompleted {
			readyTasks = append(readyTasks, task)
		}
	}

	return readyTasks, nil
}

// AddDependency 添加依赖关系
func (r *taskRepository) AddDependency(ctx context.Context, taskID, dependencyID string) error {
	// 检查任务是否存在
	var task, depTask models.Task
	if err := r.db.WithContext(ctx).Where("id = ?", taskID).First(&task).Error; err != nil {
		return err
	}
	if err := r.db.WithContext(ctx).Where("id = ?", dependencyID).First(&depTask).Error; err != nil {
		return err
	}

	// 检查是否已存在依赖关系
	var existingDep models.TaskDependency
	err := r.db.WithContext(ctx).
		Where("task_id = ? AND dependency_id = ?", taskID, dependencyID).
		First(&existingDep).Error
	if err == nil {
		return errors.New("dependency already exists")
	}

	// 创建依赖关系
	dependency := &models.TaskDependency{
		TaskID:       taskID,
		DependencyID: dependencyID,
		CreatedAt:    time.Now(),
	}

	return r.db.WithContext(ctx).Create(dependency).Error
}

// RemoveDependency 移除依赖关系
func (r *taskRepository) RemoveDependency(ctx context.Context, taskID, dependencyID string) error {
	return r.db.WithContext(ctx).
		Where("task_id = ? AND dependency_id = ?", taskID, dependencyID).
		Delete(&models.TaskDependency{}).Error
}

// GetDependencies 获取任务的依赖
func (r *taskRepository) GetDependencies(ctx context.Context, taskID string) ([]*models.Task, error) {
	var tasks []*models.Task
	err := r.db.WithContext(ctx).
		Joins("JOIN task_dependencies ON tasks.id = task_dependencies.dependency_id").
		Where("task_dependencies.task_id = ?", taskID).
		Find(&tasks).Error
	return tasks, err
}

// GetDependents 获取依赖此任务的任务
func (r *taskRepository) GetDependents(ctx context.Context, taskID string) ([]*models.Task, error) {
	var tasks []*models.Task
	err := r.db.WithContext(ctx).
		Joins("JOIN task_dependencies ON tasks.id = task_dependencies.task_id").
		Where("task_dependencies.dependency_id = ?", taskID).
		Find(&tasks).Error
	return tasks, err
}

// CheckDependenciesCompleted 检查依赖是否已完成
func (r *taskRepository) CheckDependenciesCompleted(ctx context.Context, taskID string) (bool, error) {
	var task models.Task
	if err := r.db.WithContext(ctx).Where("id = ?", taskID).First(&task).Error; err != nil {
		return false, err
	}

	if len(task.Dependencies) == 0 {
		return true, nil
	}

	// 检查所有依赖是否已完成
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Task{}).
		Where("id IN ? AND status = ?", task.Dependencies, models.TaskStatusCompleted).
		Count(&count).Error

	return int(count) == len(task.Dependencies), err
}

// IncrementRetryCount 增加重试次数
func (r *taskRepository) IncrementRetryCount(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Model(&models.Task{}).
		Where("id = ?", id).
		UpdateColumn("retry_count", gorm.Expr("retry_count + 1")).Error
}

// ResetRetryCount 重置重试次数
func (r *taskRepository) ResetRetryCount(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Model(&models.Task{}).
		Where("id = ?", id).
		Update("retry_count", 0).Error
}

// GetFailedTasks 获取失败的任务
func (r *taskRepository) GetFailedTasks(ctx context.Context) ([]*models.Task, error) {
	var tasks []*models.Task
	err := r.db.WithContext(ctx).
		Preload("Project").
		Preload("User").
		Where("status = ?", models.TaskStatusFailed).
		Order("created_at DESC").
		Find(&tasks).Error
	return tasks, err
}

// GetByProjectID 根据项目ID获取任务
func (r *taskRepository) GetByProjectID(ctx context.Context, projectID string) ([]*models.Task, error) {
	var tasks []*models.Task
	err := r.db.WithContext(ctx).
		Preload("Project").
		Preload("User").
		Where("project_id = ?", projectID).
		Order("priority DESC, created_at DESC").
		Find(&tasks).Error
	return tasks, err
}

// GetByUserID 根据用户ID获取任务
func (r *taskRepository) GetByUserID(ctx context.Context, userID string) ([]*models.Task, error) {
	var tasks []*models.Task
	err := r.db.WithContext(ctx).
		Preload("Project").
		Preload("User").
		Where("user_id = ?", userID).
		Order("priority DESC, created_at DESC").
		Find(&tasks).Error
	return tasks, err
}

// GetByPriority 根据优先级获取任务
func (r *taskRepository) GetByPriority(ctx context.Context, priority models.TaskPriority) ([]*models.Task, error) {
	var tasks []*models.Task
	err := r.db.WithContext(ctx).
		Preload("Project").
		Preload("User").
		Where("priority = ?", priority).
		Order("created_at ASC").
		Find(&tasks).Error
	return tasks, err
}

// GetOverdueTasks 获取超期任务
func (r *taskRepository) GetOverdueTasks(ctx context.Context) ([]*models.Task, error) {
	var tasks []*models.Task
	now := time.Now()
	err := r.db.WithContext(ctx).
		Preload("Project").
		Preload("User").
		Where("deadline IS NOT NULL AND deadline < ? AND status NOT IN ?",
			now, []models.TaskStatus{models.TaskStatusCompleted, models.TaskStatusCancelled, models.TaskStatusRolledBack}).
		Order("deadline ASC").
		Find(&tasks).Error
	return tasks, err
}
