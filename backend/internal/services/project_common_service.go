package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/lighthought/app-maker/shared-models/common"
	"github.com/lighthought/app-maker/shared-models/logger"
	"github.com/lighthought/app-maker/shared-models/utils"

	"github.com/lighthought/app-maker/backend/internal/models"
	"github.com/lighthought/app-maker/backend/internal/repositories"
)

const (
	MESSAGE_STAGE_DEPLOYED                = "项目项目已打包部署"
	MESSAGE_STORY_DEVELOPED               = "项目Story功能已开发"
	MESSAGE_AGENT_UNAVAILABLE             = "Agent 服务不可用"
	MESSAGE_AGENT_CALL_FAILED             = "Agent 调用失败"
	MESSAGE_CREATE_OR_UPDATE_STAGE_FAILED = "创建或更新阶段失败"
	MESSAGE_PROJECT_IS_NIL                = "project is nil"

	PATH_PRD       = "docs/PRD.md"
	PATH_UX_SPEC   = "docs/ux/ux-spec.md"
	FOLDER_STORIES = "docs/stories"
)

// 项目阶段基础服务
type ProjectCommonService interface {
	// 获取项目开发阶段
	GetProjectStages(ctx context.Context, projectGuid string) ([]*models.DevStage, error)

	// 更新项目状态为等待用户确认
	UpdateProjectWaitingForUserConfirm(ctx context.Context, project *models.Project,
		stage common.DevStatus)

	// 创建并通知用户消息
	CreateAndNotifyMessage(ctx context.Context, projectGuid string,
		message *models.ConversationMessage) error

	// 创建或更新阶段
	CreateOrUpdateStage(ctx context.Context, project *models.Project,
		taskID, projectGuid, stageName string) (*models.DevStage, bool, error)

	// 创建并通知项目阶段
	CreateAndNotifyProjectStage(ctx context.Context, project *models.Project,
		stageName common.DevStatus) (*models.DevStage, error)

	// 确保项目预览URL
	EnsureProjectPrevieUrl(ctx context.Context, projectGuid string) error

	// 更新并通知项目信息
	UpdateAndNotifyProjectInfo(ctx context.Context, project *models.Project) error

	// 更新项目到指定阶段
	UpdateProjectToStage(ctx context.Context, project *models.Project, taskID, stageName string) error

	// 更新项目状态
	UpdateProjectToStatus(ctx context.Context, project *models.Project, status string) error

	// 更新阶段状态
	UpdateStageStatus(ctx context.Context, stage *models.DevStage, status, failedReason string) error

	// 通知项目状态变化
	NotifyProjectStatusChange(ctx context.Context,
		project *models.Project, message *models.ConversationMessage, stage *models.DevStage)

	// 恢复项目和阶段
	ResumeProjectAndStage(ctx context.Context, projectGuid string) (*models.Project, *models.DevStage, error)
}

// ProjectStageService 任务执行服务
type projectCommonService struct {
	repositories     *repositories.Repository
	webSocketService WebSocketService
}

// NewTaskExecutionService 创建任务执行服务
func NewProjectCommonService(
	repositories *repositories.Repository,
	webSocketService WebSocketService,
) ProjectCommonService {
	return &projectCommonService{
		repositories:     repositories,
		webSocketService: webSocketService,
	}
}

// GetProjectStages 获取项目开发阶段
func (s *projectCommonService) GetProjectStages(ctx context.Context, projectGuid string) ([]*models.DevStage, error) {
	return s.repositories.ProjectStageRepo.GetByProjectGUID(ctx, projectGuid)
}

// UpdateProjectWaitingForUserConfirm 更新项目状态为等待用户确认
func (s *projectCommonService) UpdateProjectWaitingForUserConfirm(ctx context.Context, project *models.Project,
	stage common.DevStatus) {
	// 设置项目状态为等待用户确认
	project.WaitingForUserConfirm = true
	project.ConfirmStage = string(stage)
	s.repositories.ProjectRepo.Update(ctx, project)
	s.webSocketService.NotifyProjectInfoUpdate(ctx, project.GUID, project)

	// 通过 WebSocket 通知前端
	s.webSocketService.NotifyUserConfirmRequired(ctx, project.GUID, stage)
}

// CreateAndNotifyMessage 创建并通知用户消息
func (s *projectCommonService) CreateAndNotifyMessage(ctx context.Context, projectGuid string,
	message *models.ConversationMessage) error {
	if message != nil {
		// 保存用户消息
		if err := s.repositories.MessageRepo.Create(ctx, message); err != nil {
			logger.Error("保存项目消息失败",
				logger.String("error", err.Error()),
				logger.String("projectGuid", projectGuid),
			)
		}
		s.webSocketService.NotifyProjectMessage(ctx, projectGuid, message)
	}
	return nil
}

// CreateOrUpdateStage 创建或更新阶段
func (s *projectCommonService) CreateOrUpdateStage(ctx context.Context, project *models.Project,
	taskID, projectGuid, stageName string) (*models.DevStage, bool, error) {
	// 查找已有的阶段信息
	devProjectStage, err := s.repositories.ProjectStageRepo.GetByProjectGuidAndName(ctx, projectGuid, stageName)
	if err != nil {
		devProjectStage = &models.DevStage{
			ProjectGuid: projectGuid,
			Name:        stageName,
			Status:      common.CommonStatusInProgress,
			TaskID:      taskID,
		}
		if err := s.repositories.ProjectStageRepo.Create(ctx, devProjectStage); err != nil {
			return nil, false, fmt.Errorf("创建阶段记录失败: %w", err)
		}
	} else if devProjectStage.Status == common.CommonStatusDone {
		return devProjectStage, true, nil
	} else {
		devProjectStage.TaskID = taskID
		devProjectStage.SetStatus(common.CommonStatusInProgress)
		s.repositories.ProjectStageRepo.Update(ctx, devProjectStage)
		s.webSocketService.NotifyProjectStageUpdate(ctx, project.GUID, devProjectStage)

	}
	return devProjectStage, false, nil
}

func (s *projectCommonService) CreateAndNotifyProjectStage(ctx context.Context, project *models.Project,
	stageName common.DevStatus) (*models.DevStage, error) {
	if stageName == "" {
		return nil, fmt.Errorf("stageName is empty")
	}

	// 插入项目阶段
	stage := models.NewDevStage(project, stageName, common.CommonStatusInProgress)

	if err := s.repositories.ProjectStageRepo.Create(ctx, stage); err != nil {
		logger.Error(MESSAGE_FAILED_INSERT_PROJECT_STAGE,
			logger.String("error", err.Error()),
			logger.String("projectID", project.ID),
		)
	}
	logger.Info("插入项目阶段成功", logger.String("projectID", project.ID))
	s.webSocketService.NotifyProjectStageUpdate(ctx, project.GUID, stage)

	return stage, nil
}

// 更新项目状态
func (s *projectCommonService) UpdateProjectToStatus(ctx context.Context, project *models.Project, status string) error {
	if project == nil {
		return fmt.Errorf("%s", MESSAGE_PROJECT_IS_NIL)
	}

	if status == common.CommonStatusDone {
		project.SetDevStatus(common.DevStatusDone)
		project.Status = common.CommonStatusDone
	} else if status == common.CommonStatusFailed {
		project.SetDevStatus(common.DevStatusFailed)
		project.Status = common.CommonStatusFailed
	} else if status == common.CommonStatusPaused {
		project.SetDevStatus(common.DevStatusPaused)
		project.Status = common.CommonStatusPaused
	} else if status == common.CommonStatusInProgress {
		project.Status = common.CommonStatusInProgress
	}
	s.repositories.ProjectRepo.Update(ctx, project)
	s.webSocketService.NotifyProjectInfoUpdate(ctx, project.GUID, project)
	return nil
}

// EnsureProjectPrevieUrl 确保项目预览URL
func (s *projectCommonService) EnsureProjectPrevieUrl(ctx context.Context, projectGuid string) error {
	project, err := s.repositories.ProjectRepo.GetByGUID(ctx, projectGuid)
	if err != nil {
		return fmt.Errorf("获取项目信息失败: %w", err)
	}

	// 设置预览 URL
	if project.PreviewUrl == "" {
		project.PreviewUrl = fmt.Sprintf("http://%s.app-maker.localhost", projectGuid)
		if err := s.repositories.ProjectRepo.Update(ctx, project); err != nil {
			return fmt.Errorf("更新项目预览URL失败: %w", err)
		}
	}
	// 通知前端预览URL已设置
	s.webSocketService.NotifyProjectInfoUpdate(ctx, project.GUID, project)
	return nil
}

// 更新并通知项目信息
func (s *projectCommonService) UpdateAndNotifyProjectInfo(ctx context.Context, project *models.Project) error {
	if project == nil {
		return fmt.Errorf("%s", MESSAGE_PROJECT_IS_NIL)
	}
	if err := s.repositories.ProjectRepo.Update(ctx, project); err != nil {
		return fmt.Errorf("failed to update project: %s", err.Error())
	}
	s.webSocketService.NotifyProjectInfoUpdate(ctx, project.GUID, project)
	return nil
}

// UpdateProjectToStage 更新项目到指定阶段
func (s *projectCommonService) UpdateProjectToStage(ctx context.Context, project *models.Project, taskID, stageName string) error {
	if project == nil {
		return fmt.Errorf("%s", MESSAGE_PROJECT_IS_NIL)
	}
	project.CurrentTaskID = taskID
	project.Status = common.CommonStatusInProgress
	project.SetDevStatus(common.DevStatus(stageName))
	if err := s.repositories.ProjectRepo.Update(ctx, project); err != nil {
		return fmt.Errorf("failed to update project: %s", err.Error())
	}
	s.webSocketService.NotifyProjectInfoUpdate(ctx, project.GUID, project)
	return nil
}

// UpdateStageStatus 更新阶段状态
func (s *projectCommonService) UpdateStageStatus(ctx context.Context, stage *models.DevStage, status, failedReason string) error {
	if stage == nil {
		return fmt.Errorf("stage is nil")
	}
	if status == common.CommonStatusDone {
		now := utils.GetTimeNow()
		stage.SetStatus(common.CommonStatusDone)
		stage.CompletedAt = &now
	} else if status == common.CommonStatusFailed {
		stage.SetStatus(common.CommonStatusFailed)
		stage.FailedReason = failedReason
	} else if status == common.CommonStatusInProgress {
		stage.SetStatus(common.CommonStatusInProgress)
	} else if status == common.CommonStatusPaused {
		stage.SetStatus(common.CommonStatusPaused)
	}
	if err := s.repositories.ProjectStageRepo.Update(ctx, stage); err != nil {
		return fmt.Errorf("failed to update stage: %s", err.Error())
	}

	s.webSocketService.NotifyProjectStageUpdate(ctx, stage.ProjectGuid, stage)
	logger.Info("更新阶段状态为完成成功", logger.String("stageID", stage.ID), logger.String("stageName", stage.Name))
	return nil
}

// ResumeProjectAndStage 恢复项目和阶段
func (s *projectCommonService) ResumeProjectAndStage(ctx context.Context, projectGuid string) (*models.Project, *models.DevStage, error) {
	// 获取项目信息
	project, err := s.repositories.ProjectRepo.GetByGUID(ctx, projectGuid)
	if err != nil {
		return nil, nil, fmt.Errorf("获取项目信息失败: %w", err)
	}
	if project.Status == common.CommonStatusPaused {
		logger.Info("🔵 [AgentChat] 项目处于暂停状态，恢复为进行中",
			logger.String("projectID", project.ID),
		)
		project.Status = common.CommonStatusInProgress
		s.repositories.ProjectRepo.Update(ctx, project)
		s.webSocketService.NotifyProjectInfoUpdate(ctx, project.GUID, project)
	}

	// 恢复当前暂停的阶段
	currentStage, err := s.repositories.ProjectStageRepo.GetByProjectGuidAndName(ctx, projectGuid, project.DevStatus)
	if err == nil && currentStage != nil && currentStage.Status == common.CommonStatusPaused {
		logger.Info("🔵 [AgentChat] 阶段处于暂停状态，恢复为进行中",
			logger.String("stageID", currentStage.ID),
			logger.String("stageName", currentStage.Name),
		)
		currentStage.Status = common.CommonStatusInProgress
		if err := s.repositories.ProjectStageRepo.Update(ctx, currentStage); err != nil {
			logger.Error("恢复阶段状态失败",
				logger.String("error", err.Error()),
				logger.String("projectID", project.ID),
				logger.String("stageID", currentStage.ID),
			)
		} else {
			s.webSocketService.NotifyProjectStageUpdate(ctx, project.GUID, currentStage)
		}
	}

	return project, currentStage, nil
}

// CheckAgentQuestion 检查 agent 响应是否需要反馈
func (s *projectCommonService) CheckAgentQuestion(project *models.Project, stage *models.DevStage, message *models.ConversationMessage) bool {
	if message.Type == common.ConversationTypeAgent {
		hasQuestion := utils.ContainsQuestion(message.Content) || utils.ContainsQuestion(message.MarkdownContent)
		if hasQuestion {
			message.HasQuestion = true
			message.WaitingUserResponse = true
			message.Content = strings.Replace(message.Content, "已完成", "需要反馈", 1)

			// 暂停项目和当前阶段
			project.Status = common.CommonStatusPaused
			if stage != nil {
				stage.Status = common.CommonStatusPaused
			}

			logger.Info("检测到 Agent 问题，暂停项目执行",
				logger.String("projectID", project.ID),
				logger.String("agentRole", message.AgentRole),
				logger.String("agentName", message.AgentName),
			)
			return true
		}
	}
	return false
}

// NotifyProjectStatusChange 统一由这个函数更新项目状态
func (s *projectCommonService) NotifyProjectStatusChange(ctx context.Context,
	project *models.Project, message *models.ConversationMessage, stage *models.DevStage) {
	if message != nil {
		s.CheckAgentQuestion(project, stage, message)                           // 检查是否需要暂停（Agent 消息包含问题）
		if err := s.repositories.MessageRepo.Create(ctx, message); err != nil { // 保存用户消息
			logger.Error("保存项目消息失败", logger.String("error", err.Error()), logger.String("projectID", project.ID))
		}
		s.webSocketService.NotifyProjectMessage(ctx, project.GUID, message)
	}

	if stage == nil {
		return
	}

	if stage.ID == "" { // 插入项目阶段
		if err := s.repositories.ProjectStageRepo.Create(ctx, stage); err != nil {
			logger.Error("插入项目阶段失败", logger.String("error", err.Error()), logger.String("projectID", project.ID))
		}
		s.UpdateProjectToStage(ctx, project, stage.TaskID, stage.Name)
		logger.Info("插入项目阶段成功", logger.String("projectID", project.ID), logger.String("stageID", stage.ID))
	} else {
		stage.ProjectID = project.ID
		stage.ProjectGuid = project.GUID
		if err := s.repositories.ProjectStageRepo.Update(ctx, stage); err != nil {
			logger.Error("更新项目阶段失败",
				logger.String("error", err.Error()),
				logger.String("projectID", project.ID),
				logger.String("stageID", stage.ID),
				logger.String("stageName", stage.Name),
				logger.String("status", stage.Status),
			)
		}
		s.webSocketService.NotifyProjectStageUpdate(ctx, project.GUID, stage)
		logger.Info("更新项目阶段成功", logger.String("projectID", project.ID), logger.String("stageID", stage.ID))
	}
}
