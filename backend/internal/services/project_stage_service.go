package services

import (
	"context"
	"encoding/json"
	"fmt"

	"autocodeweb-backend/internal/constants"
	"autocodeweb-backend/internal/models"
	"autocodeweb-backend/internal/repositories"
	"autocodeweb-backend/internal/utils"
	"autocodeweb-backend/pkg/logger"

	"github.com/hibiken/asynq"
)

type ProjectStageService interface {
	GetProjectStages(ctx context.Context, projectID string) ([]*models.DevStage, error)
	ProcessTask(ctx context.Context, task *asynq.Task) error
}

// ProjectStageService 任务执行服务
type projectStageService struct {
	projectRepo repositories.ProjectRepository
	stageRepo   repositories.StageRepository
}

// NewTaskExecutionService 创建任务执行服务
func NewProjectStageService(
	projectRepo repositories.ProjectRepository,
	stageRepo repositories.StageRepository,
) ProjectStageService {
	return &projectStageService{
		projectRepo: projectRepo,
		stageRepo:   stageRepo,
	}
}

// ProcessTask 处理项目任务
func (h *projectStageService) ProcessTask(ctx context.Context, task *asynq.Task) error {
	switch task.Type() {
	case models.TypeProjectDevelopment:
		return h.HandleProjectDevelopmentTask(ctx, task)
	default:
		return fmt.Errorf("unexpected task type %s", task.Type())
	}
}

// HandleProjectDevelopmentTask 处理项目开发任务
func (s *projectStageService) HandleProjectDevelopmentTask(ctx context.Context, t *asynq.Task) error {

	var payload models.ProjectTaskPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}
	resultWriter := t.ResultWriter()
	logger.Info("处理项目开发任务", logger.String("taskID", resultWriter.TaskID()))

	// 更新 pending agents 的 stage 为 done，表示 API 内部这个 async 调用成功，也获取到了合法的数据
	if err := s.stageRepo.UpdateStageToDone(ctx, payload.ProjectID, constants.DevStatusPendingAgents); err != nil {
		logger.Error("更新项目阶段失败",
			logger.String("error", err.Error()),
			logger.String("projectID", payload.ProjectID),
		)
	}

	project, err := s.projectRepo.GetByID(ctx, payload.ProjectID)
	if err != nil {
		return fmt.Errorf("获取项目信息失败: %w", err)
	}

	s.executeProjectDevelopment(ctx, project)
	utils.UpdateResult(resultWriter, constants.CommandStatusDone, 100, "项目开发任务完成")
	return nil
}

// executeProjectDevelopment 执行项目开发流程
func (s *projectStageService) executeProjectDevelopment(ctx context.Context, project *models.Project) {
	logger.Info("开始执行项目开发流程",
		logger.String("projectID", project.ID),
	)

	// // 更新项目状态为环境就绪
	project.DevStatus = constants.DevStatusCheckRequirement
	project.DevProgress = constants.GetDevStageProgress(project.DevStatus)
	s.projectRepo.Update(ctx, project)

	// 2. 执行开发阶段
	stages := []struct {
		status      string
		description string
		executor    func(context.Context, *models.Project) error
	}{
		{constants.DevStatusGeneratePRD, "生成PRD文档", s.generatePRD},
		{constants.DevStatusDefineUXStandard, "定义UX标准", s.defineUXStandards},
		{constants.DevStatusDesignArchitecture, "设计系统架构", s.designArchitecture},
		{constants.DevStatusDefineDataModel, "定义数据模型", s.defineDataModel},
		{constants.DevStatusDefineAPI, "定义API接口", s.defineAPIs},
		{constants.DevStatusPlanEpicAndStory, "划分Epic和Story", s.planEpicsAndStories},
		{constants.DevStatusDevelopStory, "开发Story功能", s.developStories},
		{constants.DevStatusFixBug, "修复开发问题", s.fixBugs},
		{constants.DevStatusRunTest, "执行自动测试", s.runTests},
		{constants.DevStatusDeploy, "打包项目", s.packageProject},
	}

	for _, stage := range stages {
		// 更新项目状态
		project.SetDevStatus(stage.status)
		s.projectRepo.Update(ctx, project)

		// 执行阶段
		if err := stage.executor(ctx, project); err != nil {
			logger.Error("开发阶段执行失败",
				logger.String("projectID", project.ID),
				logger.String("stage", stage.status),
				logger.String("error", err.Error()),
			)

			// 更新项目状态为失败
			project.SetDevStatus(constants.DevStatusFailed)
			s.projectRepo.Update(ctx, project)

			return
		}

	}

	// 开发完成
	project.SetDevStatus(constants.DevStatusDone)
	project.Status = constants.CommandStatusDone
	s.projectRepo.Update(ctx, project)

	logger.Info("项目开发流程执行完成",
		logger.String("projectID", project.ID),
	)
}

// generatePRD 生成PRD文档
func (s *projectStageService) generatePRD(ctx context.Context, project *models.Project) error {
	//projectDir := utils.GetProjectPath(project.UserID, project.ID)

	//s.addTaskLog(ctx, task.ID, "info", "开始生成产品需求文档...")

	//s.addTaskLog(ctx, task.ID, "success", "PRD文档生成完成")
	return nil
}

// defineUXStandards 定义UX标准
func (s *projectStageService) defineUXStandards(ctx context.Context, project *models.Project) error {
	// := utils.GetProjectPath(project.UserID, project.ID)

	//s.addTaskLog(ctx, task.ID, "info", "开始定义用户体验标准...")

	// 使用 cursor-cli 定义UX标准

	//s.addTaskLog(ctx, task.ID, "success", "UX标准定义完成")
	return nil
}

// designArchitecture 设计系统架构
func (s *projectStageService) designArchitecture(ctx context.Context, project *models.Project) error {
	//projectDir := utils.GetProjectPath(project.UserID, project.ID)

	//s.addTaskLog(ctx, task.ID, "info", "开始设计系统架构...")

	// 使用 cursor-cli 设计架构

	//s.addTaskLog(ctx, task.ID, "success", "系统架构设计完成")
	return nil
}

// defineDataModel 定义数据模型
func (s *projectStageService) defineDataModel(ctx context.Context, project *models.Project) error {
	//projectDir := utils.GetProjectPath(project.UserID, project.ID)

	//s.addTaskLog(ctx, task.ID, "info", "开始定义数据模型...")

	// 使用 cursor-cli 定义数据模型

	//s.addTaskLog(ctx, task.ID, "success", "数据模型定义完成")
	return nil
}

// defineAPIs 定义API接口
func (s *projectStageService) defineAPIs(ctx context.Context, project *models.Project) error {
	//projectDir := utils.GetProjectPath(project.UserID, project.ID)

	//s.addTaskLog(ctx, task.ID, "info", "开始定义API接口...")

	// 使用 cursor-cli 定义API接口

	//s.addTaskLog(ctx, task.ID, "success", "API接口定义完成")
	return nil
}

// planEpicsAndStories 划分Epic和Story
func (s *projectStageService) planEpicsAndStories(ctx context.Context, project *models.Project) error {
	//projectDir := utils.GetProjectPath(project.UserID, project.ID)

	//s.addTaskLog(ctx, task.ID, "info", "开始划分Epic和Story...")

	// 使用 cursor-cli 划分Epic和Story

	//s.addTaskLog(ctx, task.ID, "success", "Epic和Story划分完成")
	return nil
}

// developStories 开发Story功能
func (s *projectStageService) developStories(ctx context.Context, project *models.Project) error {
	//projectDir := utils.GetProjectPath(project.UserID, project.ID)

	//s.addTaskLog(ctx, task.ID, "info", "开始开发Story功能...")

	// 使用 cursor-cli 开发Story功能

	//s.addTaskLog(ctx, task.ID, "success", "Story功能开发完成")
	return nil
}

// fixBugs 修复开发问题
func (s *projectStageService) fixBugs(ctx context.Context, project *models.Project) error {
	//projectDir := utils.GetProjectPath(project.UserID, project.ID)

	//s.addTaskLog(ctx, task.ID, "info", "开始修复开发问题...")

	// 使用 cursor-cli 修复问题

	//s.addTaskLog(ctx, task.ID, "success", "开发问题修复完成")
	return nil
}

// runTests 执行自动测试
func (s *projectStageService) runTests(ctx context.Context, project *models.Project) error {
	//projectDir := utils.GetProjectPath(project.UserID, project.ID)

	//s.addTaskLog(ctx, task.ID, "info", "开始执行自动测试...")

	// 使用 cursor-cli 执行测试

	//s.addTaskLog(ctx, task.ID, "success", "自动测试执行完成")
	return nil
}

// packageProject 打包项目
func (s *projectStageService) packageProject(ctx context.Context, project *models.Project) error {

	//s.addTaskLog(ctx, task.ID, "info", "开始打包项目...")

	// 使用 cursor-cli 打包项目

	//s.addTaskLog(ctx, task.ID, "success", "项目打包完成")
	return nil
}

// GetProjectStages 获取项目开发阶段
func (s *projectStageService) GetProjectStages(ctx context.Context, projectID string) ([]*models.DevStage, error) {
	return s.stageRepo.GetByProjectID(ctx, projectID)
}
