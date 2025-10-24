package services

import (
	"context"
	"fmt"

	"github.com/lighthought/app-maker/backend/internal/models"
	"github.com/lighthought/app-maker/backend/internal/repositories"
)

type EpicService interface {
	// GetByProjectGuid 根据项目 GUID 获取所有 Epics 和 Stories
	GetByProjectGuid(ctx context.Context, projectGuid string) ([]*models.Epic, error)
	// GetMvpEpicsByProject 获取项目的 MVP Epics
	GetMvpEpicsByProjectGuid(ctx context.Context, projectGuid string) ([]*models.Epic, error)
	// UpdateStoryStatus 更新 Story 状态
	UpdateStoryStatus(ctx context.Context, storyID string, status string) error

	// Epic 编辑相关方法
	UpdateEpicOrder(ctx context.Context, epicID string, order int) error
	UpdateEpic(ctx context.Context, epicID string, req *models.UpdateEpicRequest) error
	DeleteEpic(ctx context.Context, epicID string) error

	// Story 编辑相关方法
	UpdateStoryOrder(ctx context.Context, storyID string, order int) error
	UpdateStory(ctx context.Context, storyID string, req *models.UpdateStoryRequest) error
	DeleteStory(ctx context.Context, storyID string) error
	BatchDeleteStories(ctx context.Context, storyIDs []string) error

	// 确认相关方法
	ConfirmEpicsAndStories(ctx context.Context, projectGuid string, action string) error
}

type epicService struct {
	repositories *repositories.Repository
	fileService  FileService
}

func NewEpicService(
	repositories *repositories.Repository,
	fileService FileService,
) EpicService {
	return &epicService{
		repositories: repositories,
		fileService:  fileService,
	}
}

func (s *epicService) GetByProjectGuid(ctx context.Context, projectGuid string) ([]*models.Epic, error) {
	return s.repositories.EpicRepo.GetByProjectGuid(ctx, projectGuid)
}

func (s *epicService) GetMvpEpicsByProjectGuid(ctx context.Context, projectGuid string) ([]*models.Epic, error) {
	// 先获取项目
	project, err := s.repositories.ProjectRepo.GetByGUID(ctx, projectGuid)
	if err != nil {
		return nil, err
	}

	return s.repositories.EpicRepo.GetMvpEpicsByProject(ctx, project.ID)
}

func (s *epicService) UpdateStoryStatus(ctx context.Context, storyID string, status string) error {
	return s.repositories.StoryRepo.UpdateStatus(ctx, storyID, status)
}

// UpdateEpicOrder 更新 Epic 排序
func (s *epicService) UpdateEpicOrder(ctx context.Context, epicID string, order int) error {
	return s.repositories.EpicRepo.UpdateDisplayOrder(ctx, epicID, order)
}

// UpdateEpic 更新 Epic 内容
func (s *epicService) UpdateEpic(ctx context.Context, epicID string, req *models.UpdateEpicRequest) error {
	epic, err := s.repositories.EpicRepo.GetByID(ctx, epicID)
	if err != nil {
		return err
	}

	// 更新字段
	if req.Name != nil {
		epic.Name = *req.Name
	}
	if req.Description != nil {
		epic.Description = *req.Description
	}
	if req.Priority != nil {
		epic.Priority = *req.Priority
	}
	if req.EstimatedDays != nil {
		epic.EstimatedDays = *req.EstimatedDays
	}

	return s.repositories.EpicRepo.Update(ctx, epic)
}

// DeleteEpic 删除 Epic
func (s *epicService) DeleteEpic(ctx context.Context, epicID string) error {
	return s.repositories.EpicRepo.Delete(ctx, epicID)
}

// UpdateStoryOrder 更新 Story 排序
func (s *epicService) UpdateStoryOrder(ctx context.Context, storyID string, order int) error {
	return s.repositories.StoryRepo.UpdateDisplayOrder(ctx, storyID, order)
}

// UpdateStory 更新 Story 内容
func (s *epicService) UpdateStory(ctx context.Context, storyID string, req *models.UpdateStoryRequest) error {
	story, err := s.repositories.StoryRepo.GetByID(ctx, storyID)
	if err != nil {
		return err
	}

	// 更新字段
	if req.Title != nil {
		story.Title = *req.Title
	}
	if req.Description != nil {
		story.Description = *req.Description
	}
	if req.Priority != nil {
		story.Priority = *req.Priority
	}
	if req.EstimatedDays != nil {
		story.EstimatedDays = *req.EstimatedDays
	}
	if req.Depends != nil {
		story.Depends = *req.Depends
	}
	if req.Techs != nil {
		story.Techs = *req.Techs
	}
	if req.Content != nil {
		story.Content = *req.Content
	}
	if req.AcceptanceCriteria != nil {
		story.AcceptanceCriteria = *req.AcceptanceCriteria
	}

	return s.repositories.StoryRepo.Update(ctx, story)
}

// DeleteStory 删除 Story
func (s *epicService) DeleteStory(ctx context.Context, storyID string) error {
	return s.repositories.StoryRepo.Delete(ctx, storyID)
}

// BatchDeleteStories 批量删除 Stories
func (s *epicService) BatchDeleteStories(ctx context.Context, storyIDs []string) error {
	return s.repositories.StoryRepo.BatchDelete(ctx, storyIDs)
}

// ConfirmEpicsAndStories 确认 Epics 和 Stories
func (s *epicService) ConfirmEpicsAndStories(ctx context.Context, projectGuid string, action string) error {
	// 获取项目信息
	project, err := s.repositories.ProjectRepo.GetByGUID(ctx, projectGuid)
	if err != nil {
		return fmt.Errorf("failed to get project info: %s", err.Error())
	}

	// 获取当前的 Epics 和 Stories
	epics, err := s.repositories.EpicRepo.GetByProjectGuid(ctx, projectGuid)
	if err != nil {
		return fmt.Errorf("failed to get Epics: %s", err.Error())
	}

	// 根据 action 处理不同的确认逻辑
	switch action {
	case "confirm":
		// 确认并继续：同步到文件，然后触发下一阶段的执行
		if err := s.fileService.SyncEpicsToFiles(ctx, project.ProjectPath, epics); err != nil {
			return fmt.Errorf("failed to sync Epics to files: %s", err.Error())
		}

		// 清除等待确认状态
		project.WaitingForUserConfirm = false
		project.ConfirmStage = ""
		if err := s.repositories.ProjectRepo.Update(ctx, project); err != nil {
			return fmt.Errorf("failed to update project status: %s", err.Error())
		}

		// TODO: 触发下一阶段的执行
		// 这里需要调用 ProjectStageService 继续下一阶段

	case "skip":
		// 跳过确认：直接继续下一阶段
		project.WaitingForUserConfirm = false
		project.ConfirmStage = ""
		if err := s.repositories.ProjectRepo.Update(ctx, project); err != nil {
			return fmt.Errorf("failed to update project status: %s", err.Error())
		}

		// TODO: 触发下一阶段的执行

	case "regenerate":
		// 重新生成：调用 PO Agent 重新生成 Epics 和 Stories
		// TODO: 实现重新生成逻辑

	default:
		return fmt.Errorf("unknown confirm action: %s", action)
	}

	return nil
}
