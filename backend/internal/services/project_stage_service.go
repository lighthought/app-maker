package services

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"autocodeweb-backend/internal/models"
	"autocodeweb-backend/internal/repositories"
	"autocodeweb-backend/internal/utils"
	"autocodeweb-backend/pkg/logger"

	"golang.org/x/sync/semaphore"
)

// ProjectStageService 任务执行服务
type ProjectStageService struct {
	projectRepo repositories.ProjectRepository
	stageRepo   repositories.StageRepository
	// 线程池控制
	semaphore      *semaphore.Weighted
	maxConcurrency int64
	mu             sync.Mutex
}

// NewTaskExecutionService 创建任务执行服务
func NewProjectStageService(
	projectRepo repositories.ProjectRepository,
	stageRepo repositories.StageRepository,
) *ProjectStageService {
	maxConcurrency := int64(3) // 限制同时执行3个项目开发任务

	return &ProjectStageService{
		projectRepo:    projectRepo,
		stageRepo:      stageRepo,
		semaphore:      semaphore.NewWeighted(maxConcurrency),
		maxConcurrency: maxConcurrency,
	}
}

// StartProjectDevelopment 启动项目开发流程
func (s *ProjectStageService) StartProjectDevelopment(ctx context.Context, projectID string) error {
	logger.Info("开始项目开发流程", logger.String("projectID", projectID))

	// 获取项目信息
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return fmt.Errorf("获取项目信息失败: %w", err)
	}

	// 通过 agents-server 的接口，启动项目的开发过程，初始化项目开发环境

	// 更新项目状态为环境处理中
	project.Status = "in_progress"
	project.DevStatus = models.DevStatusEnvironmentProcessing
	project.DevProgress = project.GetDevStageProgress()

	if err := s.projectRepo.Update(ctx, project); err != nil {
		return fmt.Errorf("更新项目状态失败: %w", err)
	}

	// TODO: 需要实现 用户 MCP 工具，让 Agents 能够调用，得到当前阶段的响应是否需要调整

	// 使用线程池异步执行开发流程
	//go s.executeWithSemaphore(context.Background(), project)

	return nil
}

// executeWithSemaphore 使用信号量控制并发执行
func (s *ProjectStageService) executeWithSemaphore(ctx context.Context, project *models.Project) {
	// 获取信号量许可
	if err := s.semaphore.Acquire(ctx, 1); err != nil {
		logger.Error("获取信号量失败",
			logger.String("projectID", project.ID),
			logger.String("error", err.Error()),
		)
		return
	}
	defer s.semaphore.Release(1)

	logger.Info("获得执行许可，开始执行项目开发流程",
		logger.String("projectID", project.ID),
	)

	// 执行开发流程
	s.executeProjectDevelopment(ctx, project)
}

// executeProjectDevelopment 执行项目开发流程
func (s *ProjectStageService) executeProjectDevelopment(ctx context.Context, project *models.Project) {
	logger.Info("开始执行项目开发流程",
		logger.String("projectID", project.ID),
	)

	// // 更新项目状态为环境就绪
	project.DevStatus = models.DevStatusEnvironmentDone
	project.DevProgress = project.GetDevStageProgress()
	s.projectRepo.Update(ctx, project)

	// 2. 执行开发阶段
	stages := []struct {
		status      string
		description string
		executor    func(context.Context, *models.Project) error
	}{
		{models.DevStatusPRDGenerating, "生成PRD文档", s.generatePRD},
		{models.DevStatusUXDefining, "定义UX标准", s.defineUXStandards},
		{models.DevStatusArchDesigning, "设计系统架构", s.designArchitecture},
		{models.DevStatusDataModeling, "定义数据模型", s.defineDataModel},
		{models.DevStatusAPIDefining, "定义API接口", s.defineAPIs},
		{models.DevStatusEpicPlanning, "划分Epic和Story", s.planEpicsAndStories},
		{models.DevStatusStoryDeveloping, "开发Story功能", s.developStories},
		{models.DevStatusBugFixing, "修复开发问题", s.fixBugs},
		{models.DevStatusTesting, "执行自动测试", s.runTests},
		{models.DevStatusPackaging, "打包项目", s.packageProject},
	}

	for _, stage := range stages {
		// 更新项目状态
		project.DevStatus = stage.status
		project.DevProgress = project.GetDevStageProgress()
		s.projectRepo.Update(ctx, project)

		// 执行阶段
		if err := stage.executor(ctx, project); err != nil {
			logger.Error("开发阶段执行失败",
				logger.String("projectID", project.ID),
				logger.String("stage", stage.status),
				logger.String("error", err.Error()),
			)

			// 更新项目状态为失败
			project.DevStatus = models.DevStatusFailed
			project.Status = "failed"
			s.projectRepo.Update(ctx, project)

			//s.addTaskLog(ctx, task.ID, "error", fmt.Sprintf("%s失败: %s", stage.description, err.Error()))
			return
		}

		// 记录阶段完成
		//s.addTaskLog(ctx, task.ID, "success", fmt.Sprintf("%s完成", stage.description))
	}

	// 开发完成
	project.DevStatus = models.DevStatusCompleted
	project.Status = "completed"
	project.DevProgress = 100
	s.projectRepo.Update(ctx, project)

	//s.addTaskLog(ctx, task.ID, "success", "项目开发流程完成")

	logger.Info("项目开发流程执行完成",
		logger.String("projectID", project.ID),
	)
}

// generatePRD 生成PRD文档
func (s *ProjectStageService) generatePRD(ctx context.Context, project *models.Project) error {
	projectDir := utils.GetProjectPath(project.UserID, project.ID)

	//s.addTaskLog(ctx, task.ID, "info", "开始生成产品需求文档...")

	// 使用 cursor-cli 生成PRD
	cmd := exec.Command("cursor", "chat", "--project", projectDir, "--message",
		fmt.Sprintf("请根据以下需求生成详细的产品需求文档(PRD)：%s", project.Requirements))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("生成PRD失败: %w", err)
	}

	// 保存PRD到项目目录
	prdPath := filepath.Join(projectDir, "docs", "PRD.md")
	os.MkdirAll(filepath.Dir(prdPath), 0755)

	if err := os.WriteFile(prdPath, output, 0644); err != nil {
		return fmt.Errorf("保存PRD文件失败: %w", err)
	}

	//s.addTaskLog(ctx, task.ID, "success", "PRD文档生成完成")
	return nil
}

// defineUXStandards 定义UX标准
func (s *ProjectStageService) defineUXStandards(ctx context.Context, project *models.Project) error {
	projectDir := utils.GetProjectPath(project.UserID, project.ID)

	//s.addTaskLog(ctx, task.ID, "info", "开始定义用户体验标准...")

	// 使用 cursor-cli 定义UX标准
	cmd := exec.Command("cursor", "chat", "--project", projectDir, "--message",
		"请根据PRD文档定义用户体验(UX)设计标准，包括设计原则、交互规范、视觉规范等")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("定义UX标准失败: %w", err)
	}

	// 保存UX标准到项目目录
	uxPath := filepath.Join(projectDir, "docs", "UX_Standards.md")
	if err := os.WriteFile(uxPath, output, 0644); err != nil {
		return fmt.Errorf("保存UX标准文件失败: %w", err)
	}

	//s.addTaskLog(ctx, task.ID, "success", "UX标准定义完成")
	return nil
}

// designArchitecture 设计系统架构
func (s *ProjectStageService) designArchitecture(ctx context.Context, project *models.Project) error {
	projectDir := utils.GetProjectPath(project.UserID, project.ID)

	//s.addTaskLog(ctx, task.ID, "info", "开始设计系统架构...")

	// 使用 cursor-cli 设计架构
	cmd := exec.Command("cursor", "chat", "--project", projectDir, "--message",
		"请根据PRD和UX标准设计系统架构，包括技术栈选择、系统架构图、部署架构等")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("设计架构失败: %w", err)
	}

	// 保存架构设计到项目目录
	archPath := filepath.Join(projectDir, "docs", "Architecture.md")
	if err := os.WriteFile(archPath, output, 0644); err != nil {
		return fmt.Errorf("保存架构设计文件失败: %w", err)
	}

	//s.addTaskLog(ctx, task.ID, "success", "系统架构设计完成")
	return nil
}

// defineDataModel 定义数据模型
func (s *ProjectStageService) defineDataModel(ctx context.Context, project *models.Project) error {
	projectDir := utils.GetProjectPath(project.UserID, project.ID)

	//s.addTaskLog(ctx, task.ID, "info", "开始定义数据模型...")

	// 使用 cursor-cli 定义数据模型
	cmd := exec.Command("cursor", "chat", "--project", projectDir, "--message",
		"请根据系统架构设计数据模型，包括数据库表结构、实体关系图、API数据结构等")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("定义数据模型失败: %w", err)
	}

	// 保存数据模型到项目目录
	dataPath := filepath.Join(projectDir, "docs", "Data_Model.md")
	if err := os.WriteFile(dataPath, output, 0644); err != nil {
		return fmt.Errorf("保存数据模型文件失败: %w", err)
	}

	//s.addTaskLog(ctx, task.ID, "success", "数据模型定义完成")
	return nil
}

// defineAPIs 定义API接口
func (s *ProjectStageService) defineAPIs(ctx context.Context, project *models.Project) error {
	projectDir := utils.GetProjectPath(project.UserID, project.ID)

	//s.addTaskLog(ctx, task.ID, "info", "开始定义API接口...")

	// 使用 cursor-cli 定义API接口
	cmd := exec.Command("cursor", "chat", "--project", projectDir, "--message",
		"请根据数据模型定义API接口规范，包括接口路径、请求方法、参数、响应格式等")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("定义API接口失败: %w", err)
	}

	// 保存API接口定义到项目目录
	apiPath := filepath.Join(projectDir, "docs", "API_Specification.md")
	if err := os.WriteFile(apiPath, output, 0644); err != nil {
		return fmt.Errorf("保存API接口定义文件失败: %w", err)
	}

	//s.addTaskLog(ctx, task.ID, "success", "API接口定义完成")
	return nil
}

// planEpicsAndStories 划分Epic和Story
func (s *ProjectStageService) planEpicsAndStories(ctx context.Context, project *models.Project) error {
	projectDir := utils.GetProjectPath(project.UserID, project.ID)

	//s.addTaskLog(ctx, task.ID, "info", "开始划分Epic和Story...")

	// 使用 cursor-cli 划分Epic和Story
	cmd := exec.Command("cursor", "chat", "--project", projectDir, "--message",
		"请根据PRD和API规范划分Epic和Story，包括功能模块、用户故事、验收标准等")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("划分Epic和Story失败: %w", err)
	}

	// 保存Epic和Story规划到项目目录
	epicPath := filepath.Join(projectDir, "docs", "Epics_and_Stories.md")
	if err := os.WriteFile(epicPath, output, 0644); err != nil {
		return fmt.Errorf("保存Epic和Story规划文件失败: %w", err)
	}

	//s.addTaskLog(ctx, task.ID, "success", "Epic和Story划分完成")
	return nil
}

// developStories 开发Story功能
func (s *ProjectStageService) developStories(ctx context.Context, project *models.Project) error {
	projectDir := utils.GetProjectPath(project.UserID, project.ID)

	//s.addTaskLog(ctx, task.ID, "info", "开始开发Story功能...")

	// 使用 cursor-cli 开发Story功能
	cmd := exec.Command("cursor", "chat", "--project", projectDir, "--message",
		"请根据Epic和Story规划开始实际开发，按照优先级逐个实现Story功能")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("开发Story功能失败: %w, %s", err, output)
	}

	//s.addTaskLog(ctx, task.ID, "success", "Story功能开发完成")
	return nil
}

// fixBugs 修复开发问题
func (s *ProjectStageService) fixBugs(ctx context.Context, project *models.Project) error {
	projectDir := utils.GetProjectPath(project.UserID, project.ID)

	//s.addTaskLog(ctx, task.ID, "info", "开始修复开发问题...")

	// 使用 cursor-cli 修复问题
	cmd := exec.Command("cursor", "chat", "--project", projectDir, "--message",
		"请检查并修复开发过程中的问题，包括代码错误、逻辑问题、性能问题等")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("修复开发问题失败: %w, %s", err, output)
	}

	//s.addTaskLog(ctx, task.ID, "success", "开发问题修复完成")
	return nil
}

// runTests 执行自动测试
func (s *ProjectStageService) runTests(ctx context.Context, project *models.Project) error {
	projectDir := utils.GetProjectPath(project.UserID, project.ID)

	//s.addTaskLog(ctx, task.ID, "info", "开始执行自动测试...")

	// 使用 cursor-cli 执行测试
	cmd := exec.Command("cursor", "chat", "--project", projectDir, "--message",
		"请为项目编写并执行自动测试，包括单元测试、集成测试、端到端测试等")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("执行自动测试失败: %w, %s", err, output)
	}

	//s.addTaskLog(ctx, task.ID, "success", "自动测试执行完成")
	return nil
}

// packageProject 打包项目
func (s *ProjectStageService) packageProject(ctx context.Context, project *models.Project) error {
	projectDir := utils.GetProjectPath(project.UserID, project.ID)

	//s.addTaskLog(ctx, task.ID, "info", "开始打包项目...")

	// 使用 cursor-cli 打包项目
	cmd := exec.Command("cursor", "chat", "--project", projectDir, "--message",
		"请为项目创建Docker配置和部署脚本，确保项目可以正常打包和部署")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("打包项目失败: %w, %s", err, output)
	}

	//s.addTaskLog(ctx, task.ID, "success", "项目打包完成")
	return nil
}

// GetProjectStages 获取项目开发阶段
func (s *ProjectStageService) GetProjectStages(ctx context.Context, projectID string) ([]*models.DevStage, error) {
	return s.stageRepo.GetByProjectID(ctx, projectID)
}

// CreateDevStage 创建开发阶段
func (s *ProjectStageService) CreateDevStage(ctx context.Context, stage *models.DevStage) error {
	return s.stageRepo.Create(ctx, stage)
}

// UpdateDevStage 更新开发阶段
func (s *ProjectStageService) UpdateDevStage(ctx context.Context, stage *models.DevStage) error {
	return s.stageRepo.Update(ctx, stage)
}

// UpdateStageStatus 更新阶段状态
func (s *ProjectStageService) UpdateStageStatus(ctx context.Context, stageID string, status string) error {
	return s.stageRepo.UpdateStatus(ctx, stageID, status)
}
