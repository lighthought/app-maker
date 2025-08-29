package services

import (
	"context"
	"errors"

	"autocodeweb-backend/internal/models"
	"autocodeweb-backend/internal/repositories"
)

// TagService 标签服务接口
type TagService interface {
	// 基础CRUD操作
	CreateTag(ctx context.Context, req *models.CreateTagRequest) (*models.TagInfo, error)
	GetTag(ctx context.Context, tagID string) (*models.TagInfo, error)
	UpdateTag(ctx context.Context, tagID string, req *models.UpdateTagRequest) (*models.TagInfo, error)
	DeleteTag(ctx context.Context, tagID string) error
	ListTags(ctx context.Context) ([]*models.TagInfo, error)

	// 标签管理
	GetPopularTags(ctx context.Context, limit int) ([]*models.TagInfo, error)
	GetTagsByProject(ctx context.Context, projectID string) ([]*models.TagInfo, error)
}

// tagService 标签服务实现
type tagService struct {
	tagRepo repositories.TagRepository
}

// NewTagService 创建标签服务实例
func NewTagService(tagRepo repositories.TagRepository) TagService {
	return &tagService{
		tagRepo: tagRepo,
	}
}

// CreateTag 创建标签
func (s *tagService) CreateTag(ctx context.Context, req *models.CreateTagRequest) (*models.TagInfo, error) {
	// 检查标签名称是否已存在
	existingTag, err := s.tagRepo.GetByName(ctx, req.Name)
	if err == nil && existingTag != nil {
		return nil, errors.New("tag name already exists")
	}

	// 设置默认颜色
	if req.Color == "" {
		req.Color = "#666666"
	}

	// 创建标签
	tag := &models.Tag{
		Name:  req.Name,
		Color: req.Color,
	}

	if err := s.tagRepo.Create(ctx, tag); err != nil {
		return nil, err
	}

	return &models.TagInfo{
		ID:    tag.ID,
		Name:  tag.Name,
		Color: tag.Color,
	}, nil
}

// GetTag 获取标签信息
func (s *tagService) GetTag(ctx context.Context, tagID string) (*models.TagInfo, error) {
	tag, err := s.tagRepo.GetByID(ctx, tagID)
	if err != nil {
		return nil, err
	}

	return &models.TagInfo{
		ID:    tag.ID,
		Name:  tag.Name,
		Color: tag.Color,
	}, nil
}

// UpdateTag 更新标签
func (s *tagService) UpdateTag(ctx context.Context, tagID string, req *models.UpdateTagRequest) (*models.TagInfo, error) {
	// 获取现有标签
	tag, err := s.tagRepo.GetByID(ctx, tagID)
	if err != nil {
		return nil, err
	}

	// 如果更新名称，检查是否与其他标签冲突
	if req.Name != "" && req.Name != tag.Name {
		existingTag, err := s.tagRepo.GetByName(ctx, req.Name)
		if err == nil && existingTag != nil {
			return nil, errors.New("tag name already exists")
		}
		tag.Name = req.Name
	}

	// 更新颜色
	if req.Color != "" {
		tag.Color = req.Color
	}

	// 保存更新
	if err := s.tagRepo.Update(ctx, tag); err != nil {
		return nil, err
	}

	return &models.TagInfo{
		ID:    tag.ID,
		Name:  tag.Name,
		Color: tag.Color,
	}, nil
}

// DeleteTag 删除标签
func (s *tagService) DeleteTag(ctx context.Context, tagID string) error {
	return s.tagRepo.Delete(ctx, tagID)
}

// ListTags 获取所有标签
func (s *tagService) ListTags(ctx context.Context) ([]*models.TagInfo, error) {
	tags, err := s.tagRepo.List(ctx)
	if err != nil {
		return nil, err
	}

	tagInfos := make([]*models.TagInfo, len(tags))
	for i, tag := range tags {
		tagInfos[i] = &models.TagInfo{
			ID:    tag.ID,
			Name:  tag.Name,
			Color: tag.Color,
		}
	}

	return tagInfos, nil
}

// GetPopularTags 获取热门标签
func (s *tagService) GetPopularTags(ctx context.Context, limit int) ([]*models.TagInfo, error) {
	if limit <= 0 {
		limit = 10
	}

	tags, err := s.tagRepo.GetPopularTags(ctx, limit)
	if err != nil {
		return nil, err
	}

	tagInfos := make([]*models.TagInfo, len(tags))
	for i, tag := range tags {
		tagInfos[i] = &models.TagInfo{
			ID:    tag.ID,
			Name:  tag.Name,
			Color: tag.Color,
		}
	}

	return tagInfos, nil
}

// GetTagsByProject 获取项目的标签
func (s *tagService) GetTagsByProject(ctx context.Context, projectID string) ([]*models.TagInfo, error) {
	tags, err := s.tagRepo.GetByProjectID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	tagInfos := make([]*models.TagInfo, len(tags))
	for i, tag := range tags {
		tagInfos[i] = &models.TagInfo{
			ID:    tag.ID,
			Name:  tag.Name,
			Color: tag.Color,
		}
	}

	return tagInfos, nil
}
