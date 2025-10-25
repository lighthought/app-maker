package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/hibiken/asynq"
	"github.com/lighthought/app-maker/backend/internal/models"
	"github.com/lighthought/app-maker/backend/internal/repositories"
	"github.com/lighthought/app-maker/shared-models/agent"
	"github.com/lighthought/app-maker/shared-models/common"
	"github.com/lighthought/app-maker/shared-models/logger"
	"github.com/lighthought/app-maker/shared-models/tasks"
	"github.com/lighthought/app-maker/shared-models/utils"
)

// asynq 异步处理业务接口
type AsyncTaskService interface {
	// 处理异步任务
	ProcessTask(ctx context.Context, task *asynq.Task) error
}

// asynq 异步处理业务实现
type asyncTaskService struct {
	repositories  *repositories.Repository
	commonService ProjectCommonService
	devService    ProjectDevService
	agentService  AgentInteractService
}

// NewAsyncService 创建 asynq 异步处理业务
func NewAsyncTaskService(repositories *repositories.Repository, commonService ProjectCommonService,
	devService ProjectDevService, agentService AgentInteractService) AsyncTaskService {
	return &asyncTaskService{
		repositories:  repositories,
		commonService: commonService,
		devService:    devService,
		agentService:  agentService,
	}
}

// ProcessTask 处理 asynq 任务
func (h *asyncTaskService) ProcessTask(ctx context.Context, task *asynq.Task) error {
	switch task.Type() {
	// 项目开发阶段
	case common.TaskTypeProjectStage:
		return h.handleProjectStageTask(ctx, task)
	// 项目开发阶段状态消息任务
	case common.TaskTypeAgentTaskResponse:
		return h.handleAgentResponseTask(ctx, task)
	// 与 Agent 对话任务
	case common.TaskTypeAgentChat:
		return h.handleAgentChatTask(ctx, task)
	// 项目下载任务
	case common.TaskTypeProjectDownload:
		return h.HandleProjectDownloadTask(ctx, task)
	// 项目备份任务
	case common.TaskTypeProjectBackup:
		return h.HandleProjectBackupTask(ctx, task)
	// 部署项目任务
	case common.TaskTypeProjectDeploy:
		return h.handleProjectDeployTask(ctx, task)
	default:
		return fmt.Errorf("unexpected task type %s", task.Type())
	}
}

// 通用的阶段任务处理方法
func (s *asyncTaskService) handleProjectStageTask(ctx context.Context, t *asynq.Task) error {
	var payload tasks.ProjectStageTaskPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		logger.Error("failed to unmarshal project task payload", logger.String("error", err.Error()))
		return asynq.SkipRetry
	}
	resultWriter := t.ResultWriter()
	logger.Info("处理阶段任务", logger.String("taskID", resultWriter.TaskID()), logger.String("stage", payload.StageName))

	project, err := s.repositories.ProjectRepo.GetByGUID(ctx, payload.ProjectGuid)
	if err != nil {
		logger.Error("failed to get project information", logger.String("error", err.Error()))
		return asynq.SkipRetry
	}

	// 执行阶段
	stageItem := s.devService.GetStageItem(common.DevStatus(payload.StageName))
	if stageItem == nil {
		return fmt.Errorf("编排阶段不存在: %s", payload.StageName)
	}

	if stageItem.SkipInDevMode && utils.IsDevEnvironment() {
		logger.Info("开发模式：跳过阶段", logger.String("stageName", payload.StageName))
		s.devService.ProceedToNextStage(ctx, project, common.DevStatus(payload.StageName)) // 跳过阶段，直接执行下一阶段
		return nil
	}

	// 获取或创建阶段记录
	stage, isDone, err := s.commonService.CreateOrUpdateStage(ctx, project, resultWriter.TaskID(), payload.ProjectGuid, payload.StageName)
	if err != nil {
		logger.Error("failed to create or update stage", logger.String("error", err.Error()))
		return asynq.SkipRetry
	}
	if isDone { // 当前阶段已经完成，直接跳到下一阶段
		return s.devService.ProceedToNextStage(ctx, project, common.DevStatus(payload.StageName))
	}
	tasks.UpdateResult(resultWriter, common.CommonStatusInProgress, 10, "create stage")

	s.commonService.UpdateStageStatus(ctx, stage, common.CommonStatusInProgress, "")
	// 更新项目状态
	if err := s.commonService.UpdateProjectToStage(ctx, project, resultWriter.TaskID(), payload.StageName); err != nil {
		logger.Error("failed to update project to stage", logger.String("error", err.Error()))
		return asynq.SkipRetry
	}
	tasks.UpdateResult(resultWriter, common.CommonStatusInProgress, 30, "set project stage set to "+payload.StageName)

	taskID, err := stageItem.ReqHandler(ctx, project)
	if err != nil {
		s.commonService.UpdateStageStatus(ctx, stage, common.CommonStatusFailed, err.Error()) // 更新阶段状态为失败
		s.commonService.UpdateProjectToStatus(ctx, project, common.CommonStatusFailed)        // 更新项目状态为失败
		logger.Error("阶段任务执行失败", logger.String("error", err.Error()))
		return err
	}

	if taskID == "" {
		logger.Info("阶段任务执行成功，没有请求到 Agent，跳过阶段，直接执行下一阶段", logger.String("stageName", payload.StageName))
		s.devService.ProceedToNextStage(ctx, project, common.DevStatus(payload.StageName)) // 跳过阶段，直接执行下一阶段
		return nil
	}

	tasks.UpdateResult(resultWriter, common.CommonStatusDone, 100, payload.StageName+" has request to agent")
	logger.Info("阶段任务执行成功", logger.String("AgentTaskID", taskID))
	return nil
}

// 处理与 Agent 对话任务
func (s *asyncTaskService) handleAgentChatTask(ctx context.Context, task *asynq.Task) error {
	var req agent.ChatReq
	if err := json.Unmarshal(task.Payload(), &req); err != nil {
		logger.Error("json.Unmarshal failed", logger.String("error", err.Error()))
		return asynq.SkipRetry
	}

	resultWriter := task.ResultWriter()
	tasks.UpdateResult(resultWriter, common.CommonStatusInProgress, 0, "开始处理对话任务")

	// 创建用户消息
	userMessage := models.NewUserChatMessage(req.ProjectGuid, req.Message)
	if err := s.commonService.CreateAndNotifyMessage(ctx, req.ProjectGuid, userMessage); err != nil {
		return fmt.Errorf("保存用户消息失败: %w", err)
	}
	tasks.UpdateResult(resultWriter, common.CommonStatusInProgress, 10, "处理对话数据")

	// 恢复暂停中的任务
	project, currentStage, err := s.commonService.ResumeProjectAndStage(ctx, req.ProjectGuid)
	if err != nil {
		return fmt.Errorf("恢复项目和阶段失败: %w", err)
	}

	req.DevStage = currentStage.Name

	tasks.UpdateResult(resultWriter, common.CommonStatusInProgress, 35, "和 Agent 对话中...")

	_, err = s.agentService.ChatWithAgent(ctx, project, &req)
	if err != nil {
		return fmt.Errorf("failed to chat with agent: %w", err)
	}

	tasks.UpdateResult(resultWriter, common.CommonStatusDone, 100, "和 Agent 对话完成")
	return nil
}

// 执行当前阶段的响应处理，然后跳到下一个阶段
func (s *asyncTaskService) doStageResponseAndGoToNext(ctx context.Context, stageItem *models.DevStageItem,
	message *agent.AgentTaskStatusMessage, response *tasks.TaskResult,
	project *models.Project, stage *models.DevStage, stageName common.DevStatus) error {
	var err error
	if stageItem.RespHandler != nil {
		err = stageItem.RespHandler(ctx, message, response)
	}
	if err == nil {
		s.commonService.UpdateStageStatus(ctx, stage, common.CommonStatusDone, "")
		s.devService.ProceedToNextStage(ctx, project, stageName)
	}
	return err
}

// 处理 agent 任务完成或失败响应
func (s *asyncTaskService) handleAgentResponseTask(ctx context.Context, t *asynq.Task) error {
	var message agent.AgentTaskStatusMessage
	if err := json.Unmarshal(t.Payload(), &message); err != nil {
		logger.Error("json.Unmarshal failed", logger.String("error", err.Error()))
		return asynq.SkipRetry
	}

	resultWriter := t.ResultWriter()
	tasks.UpdateResult(resultWriter, common.CommonStatusInProgress, 10, "获取 Agent 任务结果...")
	response, err := s.agentService.WaitForTaskCompletion(ctx, message.TaskID)
	if err != nil {
		return fmt.Errorf("waiting for task completion failed: %s", err.Error())
	}

	if message.DevStage == string(common.DevStatusUnknown) { // 阵列用 Unknown 表示聊天
		err := s.devService.OnChatResponse(ctx, &message, response)
		tasks.UpdateResult(resultWriter, common.CommonStatusDone, 100, "Agent 响应为聊天，已完成.")
		return err
	}

	project, err := s.repositories.ProjectRepo.GetByGUID(ctx, message.ProjectGuid)
	if err != nil {
		tasks.UpdateResult(resultWriter, common.CommonStatusFailed, 0, "获取项目失败")
		return fmt.Errorf("failed to get project information: %s", err.Error())
	}

	stage, err := s.repositories.ProjectStageRepo.GetByProjectGuidAndName(ctx, message.ProjectGuid, message.DevStage)
	if err != nil {
		tasks.UpdateResult(resultWriter, common.CommonStatusFailed, 0, "获取项目阶段失败")
		return fmt.Errorf("failed to get stage information: %s", err.Error())
	}

	if message.Status == common.CommonStatusFailed {
		s.commonService.UpdateStageStatus(ctx, stage, common.CommonStatusFailed, response.Message)
		s.commonService.UpdateProjectToStatus(ctx, project, common.CommonStatusFailed)
		tasks.UpdateResult(resultWriter, common.CommonStatusDone, 100, "Agent 消息为失败，已同步错误信息")
		return nil
	}

	if message.Status != common.CommonStatusDone {
		logger.Error("任务状态异常", logger.String("taskID", message.TaskID), logger.String("status", message.Status))
		s.commonService.UpdateStageStatus(ctx, stage, common.CommonStatusFailed, response.Message)
		s.commonService.UpdateProjectToStatus(ctx, project, common.CommonStatusFailed)
		tasks.UpdateResult(resultWriter, common.CommonStatusDone, 100, "Agent 消息状态异常，已同步错误信息")
		return nil
	}

	stageName := common.DevStatus(message.DevStage)
	stageItem := s.devService.GetStageItem(stageName)
	if stageItem == nil {
		tasks.UpdateResult(resultWriter, common.CommonStatusFailed, 100, "获取项目阶段失败: "+message.DevStage)
		return fmt.Errorf("编排阶段不存在: %s", stageName)
	}

	err = nil
	if project.AutoGoNext || !stageItem.NeedConfirm /*|| !utils.ContainsQuestion(response.Message)*/ {
		err = s.doStageResponseAndGoToNext(ctx, stageItem, &message, response, project, stage, stageName)
	} else {
		s.commonService.UpdateProjectWaitingForUserConfirm(ctx, project, stageName, response.Message)
	}

	tasks.UpdateResult(resultWriter, common.CommonStatusDone, 100, "Agent 响应为聊天，已完成.")
	return err
}

// HandleProjectBackupTask 处理项目备份任务
func (s *asyncTaskService) HandleProjectBackupTask(ctx context.Context, t *asynq.Task) error {
	var payload tasks.ProjectTaskPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}
	resultWriter := t.ResultWriter()
	logger.Info("handle project backup task", logger.String("taskID", resultWriter.TaskID()))

	resultPath, projectPath, err := s.zipProjectPath(t)
	if err != nil {
		tasks.UpdateResult(resultWriter, common.CommonStatusFailed, 0, "failed to zip project file: "+err.Error())
		return fmt.Errorf("failed to zip project file: %s, projectID: %s", err.Error(), resultWriter.TaskID())
	}
	tasks.UpdateResult(resultWriter, common.CommonStatusInProgress, 60, "project file zipped to cache")

	// 删除项目目录
	tasks.UpdateResult(resultWriter, common.CommonStatusInProgress, 80, "deleting project directory")
	if err := os.RemoveAll(projectPath); err != nil {
		tasks.UpdateResult(resultWriter, common.CommonStatusFailed, 0, "failed to delete project directory: "+err.Error())
		return fmt.Errorf("failed to delete project directory: %s, projectPath: %s", err.Error(), projectPath)
	}
	tasks.UpdateResult(resultWriter, common.CommonStatusDone, 100, resultPath)
	return nil
}

// HandleProjectDownloadTask 处理项目下载任务
func (s *asyncTaskService) HandleProjectDownloadTask(ctx context.Context, t *asynq.Task) error {
	resultWriter := t.ResultWriter()
	logger.Info("handle project download task", logger.String("taskID", resultWriter.TaskID()))

	resultPath, _, err := s.zipProjectPath(t)
	if err != nil {
		tasks.UpdateResult(resultWriter, common.CommonStatusFailed, 0, "failed to zip project file: "+err.Error())
	}
	tasks.UpdateResult(resultWriter, common.CommonStatusDone, 100, resultPath)
	return nil
}

// zipProjectPath 压缩项目路径
func (s *asyncTaskService) zipProjectPath(t *asynq.Task) (string, string, error) {
	// 1. 解析任务负载
	var payload tasks.ProjectTaskPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return "", "", fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}
	projectID := payload.ProjectID
	projectGuid := payload.ProjectGuid
	projectPath := payload.ProjectPath
	resultWriter := t.ResultWriter()

	// 创建缓存目录
	cacheDir := utils.GetCachePath()
	// 生成缓存文件名
	cacheFileName := fmt.Sprintf("%s_%s", projectGuid, time.Now().Format("20060102_150405"))

	tasks.UpdateResult(resultWriter, common.CommonStatusInProgress, 30, "zipping project file...")
	// 使用 utils 压缩到缓存
	resultPath, err := utils.CompressDirectoryToDir(context.Background(), projectPath, cacheDir, cacheFileName)
	if err != nil {
		tasks.UpdateResult(resultWriter, common.CommonStatusFailed, 0, "failed to zip project file: "+err.Error())
		return "", projectPath, fmt.Errorf("failed to zip project file: %s, projectID: %s, projectGuid: %s", err.Error(), projectID, projectGuid)
	}
	return resultPath, projectPath, nil
}

// 处理项目部署任务
func (s *asyncTaskService) handleProjectDeployTask(ctx context.Context, t *asynq.Task) error {
	var req agent.DeployReq
	if err := json.Unmarshal(t.Payload(), &req); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}
	resultWriter := t.ResultWriter()
	logger.Info("处理项目部署任务", logger.String("taskID", resultWriter.TaskID()))

	project, err := s.repositories.ProjectRepo.GetByGUID(ctx, req.ProjectGuid)
	if err != nil {
		return fmt.Errorf("获取项目信息失败: %w", err)
	}

	tasks.UpdateResult(resultWriter, common.CommonStatusInProgress, 10, "use agent to package project")
	// 使用较长的超时时间，因为部署任务可能需要较长时间
	taskID, err := s.agentService.PackageProject(ctx, project)
	if err != nil {
		tasks.UpdateResult(resultWriter, common.CommonStatusFailed, 0, MESSAGE_AGENT_CALL_FAILED+err.Error())
		return err
	}

	tasks.UpdateResult(resultWriter, common.CommonStatusInProgress, 30, "agent task id"+taskID)
	// 部署是独立的任务，这里直接同步等待完成
	response, err := s.agentService.WaitForTaskCompletion(ctx, taskID)
	if err != nil {
		tasks.UpdateResult(resultWriter, common.CommonStatusFailed, 0, MESSAGE_AGENT_CALL_FAILED+err.Error())
		return err
	}

	projectMsg := &models.ConversationMessage{
		ProjectGuid:     project.GUID,
		Type:            common.ConversationTypeAgent,
		AgentRole:       common.AgentDev.Role,
		AgentName:       common.AgentDev.Name,
		Content:         MESSAGE_STAGE_DEPLOYED,
		IsMarkdown:      true,
		MarkdownContent: response.Message,
		IsExpanded:      true,
	}
	s.commonService.CreateAndNotifyMessage(ctx, project.GUID, projectMsg)

	// 设置预览 URL
	if project.PreviewUrl == "" {
		project.PreviewUrl = fmt.Sprintf("http://%s.app-maker.localhost", project.GUID)
		if err := s.repositories.ProjectRepo.Update(ctx, project); err != nil {
			logger.Error("更新项目预览URL失败",
				logger.String("error", err.Error()),
				logger.String("projectID", project.ID),
			)
		} else {
			logger.Info("项目预览URL已设置",
				logger.String("projectID", project.ID),
				logger.String("previewUrl", project.PreviewUrl),
			)

			// 通知前端预览URL已设置
			s.commonService.UpdateAndNotifyProjectInfo(ctx, project)
		}
	}
	tasks.UpdateResult(resultWriter, common.CommonStatusDone, 100, MESSAGE_STAGE_DEPLOYED)
	return nil
}
