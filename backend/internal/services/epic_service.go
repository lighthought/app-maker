package services

import (
	"context"

	"autocodeweb-backend/internal/models"
	"autocodeweb-backend/internal/repositories"
)

type EpicService interface {
	// GetByProjectGuid 根据项目 GUID 获取所有 Epics 和 Stories
	GetByProjectGuid(ctx context.Context, projectGuid string) ([]*models.Epic, error)
	// GetMvpEpicsByProject 获取项目的 MVP Epics
	GetMvpEpicsByProjectGuid(ctx context.Context, projectGuid string) ([]*models.Epic, error)
	// UpdateStoryStatus 更新 Story 状态
	UpdateStoryStatus(ctx context.Context, storyID string, status string) error
}

type epicService struct {
	epicRepo    repositories.EpicRepository
	storyRepo   repositories.StoryRepository
	projectRepo repositories.ProjectRepository
}

func NewEpicService(
	epicRepo repositories.EpicRepository,
	storyRepo repositories.StoryRepository,
	projectRepo repositories.ProjectRepository,
) EpicService {
	return &epicService{
		epicRepo:    epicRepo,
		storyRepo:   storyRepo,
		projectRepo: projectRepo,
	}
}

func (s *epicService) GetByProjectGuid(ctx context.Context, projectGuid string) ([]*models.Epic, error) {
	return s.epicRepo.GetByProjectGuid(ctx, projectGuid)
}

func (s *epicService) GetMvpEpicsByProjectGuid(ctx context.Context, projectGuid string) ([]*models.Epic, error) {
	// 先获取项目
	project, err := s.projectRepo.GetByGUID(ctx, projectGuid)
	if err != nil {
		return nil, err
	}

	return s.epicRepo.GetMvpEpicsByProject(ctx, project.ID)
}

func (s *epicService) UpdateStoryStatus(ctx context.Context, storyID string, status string) error {
	return s.storyRepo.UpdateStatus(ctx, storyID, status)
}
