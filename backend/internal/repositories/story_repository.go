package repositories

import (
	"context"

	"autocodeweb-backend/internal/models"

	"gorm.io/gorm"
)

type StoryRepository interface {
	// Create 创建 Story
	Create(ctx context.Context, story *models.Story) error
	// Update 更新 Story
	Update(ctx context.Context, story *models.Story) error
	// Delete 删除 Story
	Delete(ctx context.Context, id string) error
	// GetByID 根据 ID 获取 Story
	GetByID(ctx context.Context, id string) (*models.Story, error)
	// GetByEpicID 根据 Epic ID 获取所有 Stories
	GetByEpicID(ctx context.Context, epicID string) ([]*models.Story, error)
	// UpdateStatus 更新 Story 状态
	UpdateStatus(ctx context.Context, id string, status string) error
	// BatchCreate 批量创建 Stories
	BatchCreate(ctx context.Context, stories []*models.Story) error
}

type storyRepository struct {
	db *gorm.DB
}

func NewStoryRepository(db *gorm.DB) StoryRepository {
	return &storyRepository{db: db}
}

func (r *storyRepository) Create(ctx context.Context, story *models.Story) error {
	return r.db.WithContext(ctx).Create(story).Error
}

func (r *storyRepository) Update(ctx context.Context, story *models.Story) error {
	return r.db.WithContext(ctx).Save(story).Error
}

func (r *storyRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.Story{}, "id = ?", id).Error
}

func (r *storyRepository) GetByID(ctx context.Context, id string) (*models.Story, error) {
	var story models.Story
	err := r.db.WithContext(ctx).First(&story, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &story, nil
}

func (r *storyRepository) GetByEpicID(ctx context.Context, epicID string) ([]*models.Story, error) {
	var stories []*models.Story
	err := r.db.WithContext(ctx).
		Where("epic_id = ? AND deleted_at IS NULL", epicID).
		Find(&stories).Error
	return stories, err
}

func (r *storyRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	return r.db.WithContext(ctx).
		Model(&models.Story{}).
		Where("id = ?", id).
		Update("status", status).Error
}

func (r *storyRepository) BatchCreate(ctx context.Context, stories []*models.Story) error {
	return r.db.WithContext(ctx).CreateInBatches(stories, 100).Error
}
