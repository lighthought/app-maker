package repositories

import (
	"context"
	"errors"
	"fmt"

	"autocodeweb-backend/internal/models"

	"gorm.io/gorm"
)

// TagRepository 标签仓库接口
type TagRepository interface {
	// 基础CRUD操作
	Create(ctx context.Context, tag *models.Tag) error
	GetByID(ctx context.Context, id string) (*models.Tag, error)
	GetByName(ctx context.Context, name string) (*models.Tag, error)
	Update(ctx context.Context, tag *models.Tag) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]*models.Tag, error)

	// 项目标签管理
	GetByProjectID(ctx context.Context, projectID string) ([]*models.Tag, error)
	GetPopularTags(ctx context.Context, limit int) ([]*models.Tag, error)
}

// tagRepository 标签仓库实现
type tagRepository struct {
	db *gorm.DB
}

// NewTagRepository 创建标签仓库实例
func NewTagRepository(db *gorm.DB) TagRepository {
	return &tagRepository{db: db}
}

// Create 创建标签
func (r *tagRepository) Create(ctx context.Context, tag *models.Tag) error {
	return r.db.WithContext(ctx).Create(tag).Error
}

// GetByID 根据ID获取标签
func (r *tagRepository) GetByID(ctx context.Context, id string) (*models.Tag, error) {
	var tag models.Tag
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&tag).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("tag not found: %s", id)
		}
		return nil, err
	}
	return &tag, nil
}

// GetByName 根据名称获取标签
func (r *tagRepository) GetByName(ctx context.Context, name string) (*models.Tag, error) {
	var tag models.Tag
	err := r.db.WithContext(ctx).
		Where("name = ?", name).
		First(&tag).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("tag not found with name: %s", name)
		}
		return nil, err
	}
	return &tag, nil
}

// Update 更新标签
func (r *tagRepository) Update(ctx context.Context, tag *models.Tag) error {
	return r.db.WithContext(ctx).Save(tag).Error
}

// Delete 删除标签
func (r *tagRepository) Delete(ctx context.Context, id string) error {
	// 检查标签是否被项目使用
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.ProjectTag{}).
		Where("tag_id = ?", id).
		Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		return errors.New("cannot delete tag that is used by projects")
	}

	return r.db.WithContext(ctx).Delete(&models.Tag{}, "id = ?", id).Error
}

// List 获取所有标签
func (r *tagRepository) List(ctx context.Context) ([]*models.Tag, error) {
	var tags []*models.Tag
	err := r.db.WithContext(ctx).
		Order("name ASC").
		Find(&tags).Error
	return tags, err
}

// GetByProjectID 获取项目的标签
func (r *tagRepository) GetByProjectID(ctx context.Context, projectID string) ([]*models.Tag, error) {
	var tags []*models.Tag
	err := r.db.WithContext(ctx).
		Joins("JOIN project_tags ON tags.id = project_tags.tag_id").
		Where("project_tags.project_id = ?", projectID).
		Order("tags.name ASC").
		Find(&tags).Error
	return tags, err
}

// GetPopularTags 获取热门标签
func (r *tagRepository) GetPopularTags(ctx context.Context, limit int) ([]*models.Tag, error) {
	var tags []*models.Tag
	err := r.db.WithContext(ctx).
		Select("tags.*, COUNT(project_tags.project_id) as usage_count").
		Joins("LEFT JOIN project_tags ON tags.id = project_tags.tag_id").
		Group("tags.id").
		Order("usage_count DESC, tags.name ASC").
		Limit(limit).
		Find(&tags).Error
	return tags, err
}
