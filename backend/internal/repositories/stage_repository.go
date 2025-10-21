package repositories

import (
	"context"

	"github.com/lighthought/app-maker/backend/internal/models"

	"github.com/lighthought/app-maker/shared-models/common"

	"gorm.io/gorm"
)

// StageRepository 开发阶段仓库接口
type StageRepository interface {
	// Create 创建开发阶段
	Create(ctx context.Context, stage *models.DevStage) error

	// GetByProjectID 获取项目的开发阶段列表
	GetByProjectID(ctx context.Context, projectID string) ([]*models.DevStage, error)

	// GetByProjectGUID 根据项目GUID获取开发阶段列表
	GetByProjectGUID(ctx context.Context, projectGuid string) ([]*models.DevStage, error)

	// GetByProjectGuidAndName 根据项目GUID和阶段名称获取开发阶段
	GetByProjectGuidAndName(ctx context.Context, projectGuid, name string) (*models.DevStage, error)

	// 更新 stage 的状态为 done
	UpdateStageToDone(ctx context.Context, projectID, name string) (*models.DevStage, error)

	// GetByID 根据ID获取开发阶段
	GetByID(ctx context.Context, id string) (*models.DevStage, error)

	// Update 更新开发阶段
	Update(ctx context.Context, stage *models.DevStage) error

	// Delete 删除开发阶段
	Delete(ctx context.Context, id string) error

	// UpdateStatus 更新开发阶段状态
	UpdateStatus(ctx context.Context, id string, status string) error
}

// stageRepository 开发阶段仓库实现
type stageRepository struct {
	db *gorm.DB
}

// NewStageRepository 创建开发阶段仓库
func NewStageRepository(db *gorm.DB) StageRepository {
	return &stageRepository{db: db}
}

func (r *stageRepository) Create(ctx context.Context, stage *models.DevStage) error {
	return r.db.WithContext(ctx).Create(stage).Error
}

// GetByProjectID 根据项目ID获取开发阶段列表
func (r *stageRepository) GetByProjectID(ctx context.Context, projectID string) ([]*models.DevStage, error) {
	var stages []*models.DevStage
	err := r.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Order("created_at ASC").
		Find(&stages).Error
	return stages, err
}

// GetByProjectGUID 根据项目GUID获取开发阶段列表
func (r *stageRepository) GetByProjectGUID(ctx context.Context, projectGuid string) ([]*models.DevStage, error) {
	var stages []*models.DevStage
	err := r.db.WithContext(ctx).
		Where("project_guid = ?", projectGuid).
		Order("created_at ASC").
		Find(&stages).Error
	return stages, err
}

// GetByProjectGuidAndName 根据项目GUID和阶段名称获取开发阶段
func (r *stageRepository) GetByProjectGuidAndName(ctx context.Context, projectGuid, name string) (*models.DevStage, error) {
	var stage models.DevStage
	err := r.db.WithContext(ctx).
		Where("project_guid = ?", projectGuid).
		Where("name = ?", name).
		First(&stage).Error
	return &stage, err
}

// 更新 stage 的状态为 done
func (r *stageRepository) UpdateStageToDone(ctx context.Context, projectID, name string) (*models.DevStage, error) {
	var stage models.DevStage
	err := r.db.WithContext(ctx).
		Model(&models.DevStage{}).
		Where("project_id = ?", projectID).
		Where("name = ?", name).
		Update("status", common.CommonStatusDone).
		Update("progress", 100). // 同时更新进度
		First(&stage).Error
	return &stage, err
}

func (r *stageRepository) GetByID(ctx context.Context, id string) (*models.DevStage, error) {
	var stage models.DevStage
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&stage).Error
	if err != nil {
		return nil, err
	}
	return &stage, nil
}

func (r *stageRepository) Update(ctx context.Context, stage *models.DevStage) error {
	return r.db.WithContext(ctx).Save(stage).Error
}

func (r *stageRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.DevStage{}, "id = ?", id).Error
}

func (r *stageRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	return r.db.WithContext(ctx).
		Model(&models.DevStage{}).
		Where("id = ?", id).
		Update("status", status).Error
}
