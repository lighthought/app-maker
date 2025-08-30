package repositories

import (
	"context"
	"time"

	"autocodeweb-backend/internal/models"

	"gorm.io/gorm"
)

type TaskLogRepository interface {
	// 基础CRUD操作
	Create(ctx context.Context, log *models.TaskLog) error
	GetByID(ctx context.Context, id string) (*models.TaskLog, error)
	Delete(ctx context.Context, id string) error

	// 任务日志查询
	GetByTaskID(ctx context.Context, taskID string) ([]*models.TaskLog, error)
	GetByTaskIDAndLevel(ctx context.Context, taskID, level string) ([]*models.TaskLog, error)
	GetRecentLogs(ctx context.Context, taskID string, limit int) ([]*models.TaskLog, error)

	// 批量操作
	CreateBatch(ctx context.Context, logs []*models.TaskLog) error
	DeleteByTaskID(ctx context.Context, taskID string) error
	DeleteOldLogs(ctx context.Context, before time.Time) error

	// 统计查询
	GetLogCountByTaskID(ctx context.Context, taskID string) (int64, error)
	GetLogCountByLevel(ctx context.Context, taskID, level string) (int64, error)
}

type taskLogRepository struct {
	db *gorm.DB
}

func NewTaskLogRepository(db *gorm.DB) TaskLogRepository {
	return &taskLogRepository{db: db}
}

// Create 创建任务日志
func (r *taskLogRepository) Create(ctx context.Context, log *models.TaskLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// GetByID 根据ID获取任务日志
func (r *taskLogRepository) GetByID(ctx context.Context, id string) (*models.TaskLog, error) {
	var log models.TaskLog
	err := r.db.WithContext(ctx).
		Preload("Task").
		Where("id = ?", id).
		First(&log).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &log, nil
}

// Delete 删除任务日志
func (r *taskLogRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.TaskLog{}, "id = ?", id).Error
}

// GetByTaskID 根据任务ID获取日志
func (r *taskLogRepository) GetByTaskID(ctx context.Context, taskID string) ([]*models.TaskLog, error) {
	var logs []*models.TaskLog
	err := r.db.WithContext(ctx).
		Preload("Task").
		Where("task_id = ?", taskID).
		Order("created_at DESC").
		Find(&logs).Error
	return logs, err
}

// GetByTaskIDAndLevel 根据任务ID和日志级别获取日志
func (r *taskLogRepository) GetByTaskIDAndLevel(ctx context.Context, taskID, level string) ([]*models.TaskLog, error) {
	var logs []*models.TaskLog
	err := r.db.WithContext(ctx).
		Preload("Task").
		Where("task_id = ? AND level = ?", taskID, level).
		Order("created_at DESC").
		Find(&logs).Error
	return logs, err
}

// GetRecentLogs 获取最近的日志
func (r *taskLogRepository) GetRecentLogs(ctx context.Context, taskID string, limit int) ([]*models.TaskLog, error) {
	var logs []*models.TaskLog
	err := r.db.WithContext(ctx).
		Preload("Task").
		Where("task_id = ?", taskID).
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}

// CreateBatch 批量创建日志
func (r *taskLogRepository) CreateBatch(ctx context.Context, logs []*models.TaskLog) error {
	if len(logs) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).CreateInBatches(logs, 100).Error
}

// DeleteByTaskID 根据任务ID删除日志
func (r *taskLogRepository) DeleteByTaskID(ctx context.Context, taskID string) error {
	return r.db.WithContext(ctx).Delete(&models.TaskLog{}, "task_id = ?", taskID).Error
}

// DeleteOldLogs 删除旧日志
func (r *taskLogRepository) DeleteOldLogs(ctx context.Context, before time.Time) error {
	return r.db.WithContext(ctx).Delete(&models.TaskLog{}, "created_at < ?", before).Error
}

// GetLogCountByTaskID 获取任务日志数量
func (r *taskLogRepository) GetLogCountByTaskID(ctx context.Context, taskID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.TaskLog{}).
		Where("task_id = ?", taskID).
		Count(&count).Error
	return count, err
}

// GetLogCountByLevel 根据级别获取任务日志数量
func (r *taskLogRepository) GetLogCountByLevel(ctx context.Context, taskID, level string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.TaskLog{}).
		Where("task_id = ? AND level = ?", taskID, level).
		Count(&count).Error
	return count, err
}
