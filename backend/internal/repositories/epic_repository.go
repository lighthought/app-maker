package repositories

import (
	"context"

	"autocodeweb-backend/internal/models"

	"gorm.io/gorm"
)

type EpicRepository interface {
	// Create 创建 Epic
	Create(ctx context.Context, epic *models.Epic) error
	// Update 更新 Epic
	Update(ctx context.Context, epic *models.Epic) error
	// Delete 删除 Epic
	Delete(ctx context.Context, id string) error
	// GetByID 根据 ID 获取 Epic
	GetByID(ctx context.Context, id string) (*models.Epic, error)
	// GetByProjectID 根据项目 ID 获取所有 Epics
	GetByProjectID(ctx context.Context, projectID string) ([]*models.Epic, error)
	// GetByProjectGuid 根据项目 GUID 获取所有 Epics
	GetByProjectGuid(ctx context.Context, projectGuid string) ([]*models.Epic, error)
	// GetMvpEpicsByProject 获取项目的 MVP 阶段 Epics (P0 优先级)
	GetMvpEpicsByProject(ctx context.Context, projectID string) ([]*models.Epic, error)
	// BatchCreate 批量创建 Epics
	BatchCreate(ctx context.Context, epics []*models.Epic) error
}

type epicRepository struct {
	db *gorm.DB
}

func NewEpicRepository(db *gorm.DB) EpicRepository {
	return &epicRepository{db: db}
}

func (r *epicRepository) Create(ctx context.Context, epic *models.Epic) error {
	return r.db.WithContext(ctx).Create(epic).Error
}

func (r *epicRepository) Update(ctx context.Context, epic *models.Epic) error {
	return r.db.WithContext(ctx).Save(epic).Error
}

func (r *epicRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.Epic{}, "id = ?", id).Error
}

func (r *epicRepository) GetByID(ctx context.Context, id string) (*models.Epic, error) {
	var epic models.Epic
	err := r.db.WithContext(ctx).
		Preload("Stories").
		First(&epic, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &epic, nil
}

func (r *epicRepository) GetByProjectID(ctx context.Context, projectID string) ([]*models.Epic, error) {
	var epics []*models.Epic
	err := r.db.WithContext(ctx).
		Preload("Stories", "deleted_at IS NULL").
		Where("project_id = ? AND deleted_at IS NULL", projectID).
		Order("epic_number ASC").
		Find(&epics).Error
	return epics, err
}

func (r *epicRepository) GetByProjectGuid(ctx context.Context, projectGuid string) ([]*models.Epic, error) {
	var epics []*models.Epic
	err := r.db.WithContext(ctx).
		Preload("Stories", "deleted_at IS NULL").
		Where("project_guid = ? AND deleted_at IS NULL", projectGuid).
		Order("epic_number ASC").
		Find(&epics).Error
	return epics, err
}

func (r *epicRepository) GetMvpEpicsByProject(ctx context.Context, projectID string) ([]*models.Epic, error) {
	var epics []*models.Epic
	err := r.db.WithContext(ctx).
		Preload("Stories", "deleted_at IS NULL AND priority = ?", "P0").
		Where("project_id = ? AND deleted_at IS NULL AND priority = ?", projectID, "P0").
		Order("epic_number ASC").
		Find(&epics).Error
	return epics, err
}

func (r *epicRepository) BatchCreate(ctx context.Context, epics []*models.Epic) error {
	return r.db.WithContext(ctx).CreateInBatches(epics, 100).Error
}
