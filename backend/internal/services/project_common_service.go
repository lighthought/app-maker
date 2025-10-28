package services

import (
	"context"
	"fmt"

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
		stage common.DevStatus, message string)

	// 创建并通知用户消息
	CreateAndNotifyMessage(ctx context.Context, projectGuid string,
		message *models.ConversationMessage) error

	// 创建或更新阶段
	CreateOrUpdateStage(ctx context.Context, project *models.Project,
		taskID, projectGuid, stageName string, needConfirm bool) (*models.DevStage, bool, error)

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

	// 更新阶段用户确认状态
	UpdateStageUserConfirmed(ctx context.Context, projectGuid, stageName string) error

	// 恢复项目和阶段
	ResumeProjectAndStage(ctx context.Context, projectGuid string) (*models.Project, *models.DevStage, error)
}

// ProjectStageService 任务执行服务
type projectCommonService struct {
	repositories     *repositories.Repository
	webSocketService WebSocketService
	enviroment       string
}

// NewTaskExecutionService 创建任务执行服务
func NewProjectCommonService(
	repositories *repositories.Repository,
	webSocketService WebSocketService,
	enviroment string,
) ProjectCommonService {
	return &projectCommonService{
		repositories:     repositories,
		webSocketService: webSocketService,
		enviroment:       enviroment,
	}
}

// GetProjectStages 获取项目开发阶段
func (s *projectCommonService) GetProjectStages(ctx context.Context, projectGuid string) ([]*models.DevStage, error) {
	return s.repositories.ProjectStageRepo.GetByProjectGUID(ctx, projectGuid)
}

// UpdateProjectWaitingForUserConfirm 更新项目状态为等待用户确认
func (s *projectCommonService) UpdateProjectWaitingForUserConfirm(ctx context.Context, project *models.Project,
	stage common.DevStatus, message string) {
	// 设置项目状态为等待用户确认
	project.WaitingForUserConfirm = true
	project.Status = common.CommonStatusPaused
	project.ConfirmStage = string(stage)
	s.repositories.ProjectRepo.Update(ctx, project)
	s.webSocketService.NotifyProjectInfoUpdate(ctx, project.GUID, project)

	// 通过 WebSocket 通知前端
	s.webSocketService.NotifyUserConfirmRequired(ctx, project.GUID, stage, message)
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
	taskID, projectGuid, stageName string, needConfirm bool) (*models.DevStage, bool, error) {
	// 查找已有的阶段信息
	devProjectStage, err := s.repositories.ProjectStageRepo.GetByProjectGuidAndName(ctx, projectGuid, stageName)
	if err != nil {
		devProjectStage = models.NewDevStage(project, common.DevStatus(stageName), common.CommonStatusInProgress)
		devProjectStage.TaskID = taskID
		devProjectStage.NeedConfirm = needConfirm
		if err := s.repositories.ProjectStageRepo.Create(ctx, devProjectStage); err != nil {
			return nil, false, fmt.Errorf("创建阶段记录失败: %w", err)
		}
		devProjectStage, _ = s.repositories.ProjectStageRepo.GetByProjectGuidAndName(ctx, projectGuid, stageName)
	} else if devProjectStage.Status == common.CommonStatusDone {
		return devProjectStage, true, nil
	} else {
		devProjectStage.TaskID = taskID
		devProjectStage.SetStatus(common.CommonStatusInProgress)
		s.repositories.ProjectStageRepo.Update(ctx, devProjectStage)
	}

	s.webSocketService.NotifyProjectStageUpdate(ctx, project.GUID, devProjectStage)
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

	switch status {
	case common.CommonStatusDone:
		project.SetDevStatus(common.DevStatusDone)
		project.Status = common.CommonStatusDone
	case common.CommonStatusFailed:
		project.SetDevStatus(common.DevStatusFailed)
		project.Status = common.CommonStatusFailed
	case common.CommonStatusPaused:
		project.SetDevStatus(common.DevStatusPaused)
		project.Status = common.CommonStatusPaused
	case common.CommonStatusInProgress:
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

	if project.PreviewUrl != "" {
		return nil
	}
	// 设置预览 URL
	switch s.enviroment {
	case common.EnvironmentLocalDebug:
		project.PreviewUrl = fmt.Sprintf("http://localhost:%d", project.FrontendPort)
	case common.EnvironmentDevelopment:
		project.PreviewUrl = fmt.Sprintf("http://%s.app-maker.localhost", projectGuid)
	case common.EnvironmentProduction:
		project.PreviewUrl = fmt.Sprintf("http://%s.app-maker.lighthought.com", projectGuid)
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

	switch status {
	case common.CommonStatusDone:
		now := utils.GetTimeNow()
		stage.SetStatus(common.CommonStatusDone)
		stage.CompletedAt = &now
	case common.CommonStatusFailed:
		stage.SetStatus(common.CommonStatusFailed)
		stage.FailedReason = failedReason
	case common.CommonStatusInProgress:
		stage.SetStatus(common.CommonStatusInProgress)
	case common.CommonStatusPaused:
		stage.SetStatus(common.CommonStatusPaused)
	}
	if err := s.repositories.ProjectStageRepo.Update(ctx, stage); err != nil {
		return fmt.Errorf("failed to update stage: %s", err.Error())
	}

	s.webSocketService.NotifyProjectStageUpdate(ctx, stage.ProjectGuid, stage)
	logger.Info("更新阶段状态为完成成功", logger.String("stageID", stage.ID), logger.String("stageName", stage.Name))
	return nil
}

// 更新阶段用户确认状态
func (s *projectCommonService) UpdateStageUserConfirmed(ctx context.Context, projectGuid, stageName string) error {
	stage, err := s.repositories.ProjectStageRepo.GetByProjectGuidAndName(ctx, projectGuid, stageName)
	if err != nil {
		return fmt.Errorf("获取阶段信息失败: %w", err)
	}
	stage.UserConfirmed = true
	now := utils.GetTimeNow()
	stage.SetStatus(common.CommonStatusDone)
	stage.CompletedAt = &now
	if err := s.repositories.ProjectStageRepo.Update(ctx, stage); err != nil {
		return fmt.Errorf("更新阶段用户确认状态失败: %w", err)
	}

	s.webSocketService.NotifyProjectStageUpdate(ctx, stage.ProjectGuid, stage)
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
