package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"autocodeweb-backend/internal/models"
	"autocodeweb-backend/internal/repositories"
	"autocodeweb-backend/pkg/logger"

	"github.com/google/uuid"
	"golang.org/x/sync/semaphore"
)

// TaskExecutionService 任务执行服务
type TaskExecutionService struct {
	projectService    ProjectService
	projectRepo       repositories.ProjectRepository
	taskRepo          repositories.TaskRepository
	projectDevService *ProjectDevService
	baseProjectsDir   string
	jenkinsConfig     *JenkinsConfig

	// 线程池控制
	semaphore      *semaphore.Weighted
	maxConcurrency int64
	mu             sync.Mutex
}

// JenkinsConfig Jenkins 配置
type JenkinsConfig struct {
	BaseURL     string
	Username    string
	APIToken    string
	JobName     string
	RemoteToken string
	Timeout     time.Duration
}

// JenkinsBuildRequest Jenkins 构建请求
type JenkinsBuildRequest struct {
	UserID      string `json:"user_id"`
	ProjectID   string `json:"project_id"`
	ProjectPath string `json:"project_path"`
	BuildType   string `json:"build_type"` // dev 或 prod
}

// NewTaskExecutionService 创建任务执行服务
func NewTaskExecutionService(
	projectService ProjectService,
	projectRepo repositories.ProjectRepository,
	taskRepo repositories.TaskRepository,
	projectDevService *ProjectDevService,
	baseProjectsDir string,
) *TaskExecutionService {
	maxConcurrency := int64(3) // 限制同时执行3个项目开发任务

	// Jenkins 配置
	jenkinsConfig := &JenkinsConfig{
		BaseURL:     getEnvOrDefault("JENKINS_URL", "http://10.0.0.6:5016"),
		Username:    getEnvOrDefault("JENKINS_USERNAME", "admin"),
		APIToken:    getEnvOrDefault("JENKINS_API_TOKEN", "119ffe6f373f1cb4b4b4e9a27ca5b1890f"),
		JobName:     getEnvOrDefault("JENKINS_JOB_NAME", "app-maker-flow"),
		RemoteToken: getEnvOrDefault("JENKINS_REMOTE_TOKEN", ""),
		Timeout:     30 * time.Minute,
	}

	return &TaskExecutionService{
		projectService:    projectService,
		projectRepo:       projectRepo,
		taskRepo:          taskRepo,
		projectDevService: projectDevService,
		baseProjectsDir:   baseProjectsDir,
		jenkinsConfig:     jenkinsConfig,
		semaphore:         semaphore.NewWeighted(maxConcurrency),
		maxConcurrency:    maxConcurrency,
	}
}

// getEnvOrDefault 获取环境变量或返回默认值
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// StartProjectDevelopment 启动项目开发流程
func (s *TaskExecutionService) StartProjectDevelopment(ctx context.Context, projectID string) error {
	logger.Info("开始项目开发流程", logger.String("projectID", projectID))

	// 获取项目信息
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return fmt.Errorf("获取项目信息失败: %w", err)
	}

	// 更新项目状态为环境处理中
	project.Status = "in_progress"
	project.DevStatus = models.DevStatusEnvironmentProcessing
	project.DevProgress = project.GetDevStageProgress()

	if err := s.projectRepo.Update(ctx, project); err != nil {
		return fmt.Errorf("更新项目状态失败: %w", err)
	}

	// 创建开发任务
	taskID := uuid.New().String()
	task := &models.Task{
		ID:          taskID,
		ProjectID:   projectID,
		Type:        "project_development",
		Status:      "pending",
		Priority:    1,
		Description: "项目开发流程",
		CreatedAt:   time.Now(),
	}

	if err := s.taskRepo.Create(ctx, task); err != nil {
		return fmt.Errorf("创建任务失败: %w", err)
	}

	// 更新项目的当前任务ID
	project.CurrentTaskID = taskID
	if err := s.projectRepo.Update(ctx, project); err != nil {
		return fmt.Errorf("更新项目任务ID失败: %w", err)
	}

	// 使用线程池异步执行开发流程
	go s.executeWithSemaphore(context.Background(), project, task)

	return nil
}

// executeWithSemaphore 使用信号量控制并发执行
func (s *TaskExecutionService) executeWithSemaphore(ctx context.Context, project *models.Project, task *models.Task) {
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
		logger.String("taskID", task.ID),
	)

	// 执行开发流程
	s.executeProjectDevelopment(ctx, project, task)
}

// executeProjectDevelopment 执行项目开发流程
func (s *TaskExecutionService) executeProjectDevelopment(ctx context.Context, project *models.Project, task *models.Task) {
	logger.Info("开始执行项目开发流程",
		logger.String("projectID", project.ID),
		logger.String("taskID", task.ID),
	)

	// 更新任务状态为执行中
	task.Status = "in_progress"
	task.StartedAt = &time.Time{}
	*task.StartedAt = time.Now()
	s.taskRepo.Update(ctx, task)

	// 1. 首先初始化开发环境
	s.addTaskLog(ctx, task.ID, "info", "开始初始化项目开发环境...")

	projectDir := project.ProjectPath
	if err := s.projectDevService.SetupProjectDevEnvironment(project); err != nil {
		logger.Error("初始化开发环境失败",
			logger.String("projectID", project.ID),
			logger.String("projectDir", projectDir),
			logger.String("error", err.Error()),
		)

		// 更新项目状态为失败
		project.DevStatus = models.DevStatusFailed
		project.Status = "failed"
		s.projectRepo.Update(ctx, project)

		// 更新任务状态
		task.Status = "failed"
		task.CompletedAt = &time.Time{}
		*task.CompletedAt = time.Now()
		s.taskRepo.Update(ctx, task)

		s.addTaskLog(ctx, task.ID, "error", fmt.Sprintf("初始化开发环境失败: %s", err.Error()))
		return
	}

	// 更新项目状态为环境就绪
	project.DevStatus = models.DevStatusEnvironmentDone
	project.DevProgress = project.GetDevStageProgress()
	s.projectRepo.Update(ctx, project)
	s.addTaskLog(ctx, task.ID, "success", "项目开发环境初始化完成")

	// 2. 执行开发阶段
	stages := []struct {
		status      string
		description string
		executor    func(context.Context, *models.Project, *models.Task) error
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

		// 记录阶段开始
		s.addTaskLog(ctx, task.ID, "info", fmt.Sprintf("开始%s", stage.description))

		// 执行阶段
		if err := stage.executor(ctx, project, task); err != nil {
			logger.Error("开发阶段执行失败",
				logger.String("projectID", project.ID),
				logger.String("stage", stage.status),
				logger.String("error", err.Error()),
			)

			// 更新项目状态为失败
			project.DevStatus = models.DevStatusFailed
			project.Status = "failed"
			s.projectRepo.Update(ctx, project)

			// 更新任务状态
			task.Status = "failed"
			task.CompletedAt = &time.Time{}
			*task.CompletedAt = time.Now()
			s.taskRepo.Update(ctx, task)

			s.addTaskLog(ctx, task.ID, "error", fmt.Sprintf("%s失败: %s", stage.description, err.Error()))
			return
		}

		// 记录阶段完成
		s.addTaskLog(ctx, task.ID, "success", fmt.Sprintf("%s完成", stage.description))
	}

	// 开发完成
	project.DevStatus = models.DevStatusCompleted
	project.Status = "completed"
	project.DevProgress = 100
	s.projectRepo.Update(ctx, project)

	// 更新任务状态
	task.Status = "completed"
	task.CompletedAt = &time.Time{}
	*task.CompletedAt = time.Now()
	s.taskRepo.Update(ctx, task)

	s.addTaskLog(ctx, task.ID, "success", "项目开发流程完成")

	// 触发 Jenkins 构建和部署
	go s.triggerJenkinsBuild(context.Background(), project, task)

	logger.Info("项目开发流程执行完成",
		logger.String("projectID", project.ID),
		logger.String("taskID", task.ID),
	)
}

// generatePRD 生成PRD文档
func (s *TaskExecutionService) generatePRD(ctx context.Context, project *models.Project, task *models.Task) error {
	projectDir := filepath.Join(s.baseProjectsDir, project.UserID, project.ID)

	s.addTaskLog(ctx, task.ID, "info", "开始生成产品需求文档...")

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

	s.addTaskLog(ctx, task.ID, "success", "PRD文档生成完成")
	return nil
}

// defineUXStandards 定义UX标准
func (s *TaskExecutionService) defineUXStandards(ctx context.Context, project *models.Project, task *models.Task) error {
	projectDir := filepath.Join(s.baseProjectsDir, project.UserID, project.ID)

	s.addTaskLog(ctx, task.ID, "info", "开始定义用户体验标准...")

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

	s.addTaskLog(ctx, task.ID, "success", "UX标准定义完成")
	return nil
}

// designArchitecture 设计系统架构
func (s *TaskExecutionService) designArchitecture(ctx context.Context, project *models.Project, task *models.Task) error {
	projectDir := filepath.Join(s.baseProjectsDir, project.UserID, project.ID)

	s.addTaskLog(ctx, task.ID, "info", "开始设计系统架构...")

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

	s.addTaskLog(ctx, task.ID, "success", "系统架构设计完成")
	return nil
}

// defineDataModel 定义数据模型
func (s *TaskExecutionService) defineDataModel(ctx context.Context, project *models.Project, task *models.Task) error {
	projectDir := filepath.Join(s.baseProjectsDir, project.UserID, project.ID)

	s.addTaskLog(ctx, task.ID, "info", "开始定义数据模型...")

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

	s.addTaskLog(ctx, task.ID, "success", "数据模型定义完成")
	return nil
}

// defineAPIs 定义API接口
func (s *TaskExecutionService) defineAPIs(ctx context.Context, project *models.Project, task *models.Task) error {
	projectDir := filepath.Join(s.baseProjectsDir, project.UserID, project.ID)

	s.addTaskLog(ctx, task.ID, "info", "开始定义API接口...")

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

	s.addTaskLog(ctx, task.ID, "success", "API接口定义完成")
	return nil
}

// planEpicsAndStories 划分Epic和Story
func (s *TaskExecutionService) planEpicsAndStories(ctx context.Context, project *models.Project, task *models.Task) error {
	projectDir := filepath.Join(s.baseProjectsDir, project.UserID, project.ID)

	s.addTaskLog(ctx, task.ID, "info", "开始划分Epic和Story...")

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

	s.addTaskLog(ctx, task.ID, "success", "Epic和Story划分完成")
	return nil
}

// developStories 开发Story功能
func (s *TaskExecutionService) developStories(ctx context.Context, project *models.Project, task *models.Task) error {
	projectDir := filepath.Join(s.baseProjectsDir, project.UserID, project.ID)

	s.addTaskLog(ctx, task.ID, "info", "开始开发Story功能...")

	// 使用 cursor-cli 开发Story功能
	cmd := exec.Command("cursor", "chat", "--project", projectDir, "--message",
		"请根据Epic和Story规划开始实际开发，按照优先级逐个实现Story功能")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("开发Story功能失败: %w, %s", err, output)
	}

	s.addTaskLog(ctx, task.ID, "success", "Story功能开发完成")
	return nil
}

// fixBugs 修复开发问题
func (s *TaskExecutionService) fixBugs(ctx context.Context, project *models.Project, task *models.Task) error {
	projectDir := filepath.Join(s.baseProjectsDir, project.UserID, project.ID)

	s.addTaskLog(ctx, task.ID, "info", "开始修复开发问题...")

	// 使用 cursor-cli 修复问题
	cmd := exec.Command("cursor", "chat", "--project", projectDir, "--message",
		"请检查并修复开发过程中的问题，包括代码错误、逻辑问题、性能问题等")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("修复开发问题失败: %w, %s", err, output)
	}

	s.addTaskLog(ctx, task.ID, "success", "开发问题修复完成")
	return nil
}

// runTests 执行自动测试
func (s *TaskExecutionService) runTests(ctx context.Context, project *models.Project, task *models.Task) error {
	projectDir := filepath.Join(s.baseProjectsDir, project.UserID, project.ID)

	s.addTaskLog(ctx, task.ID, "info", "开始执行自动测试...")

	// 使用 cursor-cli 执行测试
	cmd := exec.Command("cursor", "chat", "--project", projectDir, "--message",
		"请为项目编写并执行自动测试，包括单元测试、集成测试、端到端测试等")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("执行自动测试失败: %w, %s", err, output)
	}

	s.addTaskLog(ctx, task.ID, "success", "自动测试执行完成")
	return nil
}

// packageProject 打包项目
func (s *TaskExecutionService) packageProject(ctx context.Context, project *models.Project, task *models.Task) error {
	projectDir := filepath.Join(s.baseProjectsDir, project.UserID, project.ID)

	s.addTaskLog(ctx, task.ID, "info", "开始打包项目...")

	// 使用 cursor-cli 打包项目
	cmd := exec.Command("cursor", "chat", "--project", projectDir, "--message",
		"请为项目创建Docker配置和部署脚本，确保项目可以正常打包和部署")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("打包项目失败: %w, %s", err, output)
	}

	s.addTaskLog(ctx, task.ID, "success", "项目打包完成")
	return nil
}

// addTaskLog 添加任务日志
func (s *TaskExecutionService) addTaskLog(ctx context.Context, taskID, level, message string) {
	log := &models.TaskLog{
		ID:        uuid.New().String(),
		TaskID:    taskID,
		Level:     level,
		Message:   message,
		CreatedAt: time.Now(),
	}

	if err := s.taskRepo.CreateLog(ctx, log); err != nil {
		logger.Error("添加任务日志失败",
			logger.String("taskID", taskID),
			logger.String("error", err.Error()),
		)
	}
}

// triggerJenkinsBuild 触发 Jenkins 构建
func (s *TaskExecutionService) triggerJenkinsBuild(ctx context.Context, project *models.Project, task *models.Task) {
	logger.Info("开始触发 Jenkins 构建",
		logger.String("projectID", project.ID),
		logger.String("taskID", task.ID),
	)

	s.addTaskLog(ctx, task.ID, "info", "开始触发 Jenkins 构建和部署...")

	// 构建 Jenkins 请求
	// 将容器内路径转换为主机路径
	hostProjectPath := s.convertToHostPath(project.ProjectPath)

	buildRequest := &JenkinsBuildRequest{
		UserID:      project.UserID,
		ProjectID:   project.ID,
		ProjectPath: hostProjectPath,
		BuildType:   "dev", // 默认使用开发环境
	}

	// 序列化请求
	jsonData, err := json.Marshal(buildRequest)
	if err != nil {
		logger.Error("序列化 Jenkins 请求失败",
			logger.String("projectID", project.ID),
			logger.String("error", err.Error()),
		)
		s.addTaskLog(ctx, task.ID, "error", fmt.Sprintf("序列化 Jenkins 请求失败: %s", err.Error()))
		return
	}

	// 构建 Jenkins API URL
	var jenkinsURL string
	var req *http.Request

	if s.jenkinsConfig.RemoteToken != "" {
		// 使用远程令牌触发
		jenkinsURL = fmt.Sprintf("%s/job/%s/buildWithParameters", s.jenkinsConfig.BaseURL, s.jenkinsConfig.JobName)
		params := fmt.Sprintf("USER_ID=%s&PROJECT_ID=%s&PROJECT_PATH=%s&BUILD_TYPE=%s&token=%s",
			buildRequest.UserID, buildRequest.ProjectID, buildRequest.ProjectPath, buildRequest.BuildType, s.jenkinsConfig.RemoteToken)
		jenkinsURL = jenkinsURL + "?" + params

		req, err = http.NewRequestWithContext(ctx, "POST", jenkinsURL, nil)
	} else {
		// 使用 API Token 触发
		jenkinsURL = fmt.Sprintf("%s/job/%s/buildWithParameters", s.jenkinsConfig.BaseURL, s.jenkinsConfig.JobName)
		req, err = http.NewRequestWithContext(ctx, "POST", jenkinsURL, bytes.NewBuffer(jsonData))
	}

	if err != nil {
		logger.Error("创建 Jenkins 请求失败",
			logger.String("projectID", project.ID),
			logger.String("error", err.Error()),
		)
		s.addTaskLog(ctx, task.ID, "error", fmt.Sprintf("创建 Jenkins 请求失败: %s", err.Error()))
		return
	}

	// 设置请求头
	if s.jenkinsConfig.RemoteToken == "" {
		req.Header.Set("Content-Type", "application/json")
		if s.jenkinsConfig.APIToken != "" {
			req.SetBasicAuth(s.jenkinsConfig.Username, s.jenkinsConfig.APIToken)
		}
	}

	// 发送请求
	client := &http.Client{Timeout: s.jenkinsConfig.Timeout}
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("发送 Jenkins 请求失败",
			logger.String("projectID", project.ID),
			logger.String("error", err.Error()),
		)
		s.addTaskLog(ctx, task.ID, "error", fmt.Sprintf("发送 Jenkins 请求失败: %s", err.Error()))
		return
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		logger.Info("Jenkins 构建触发成功",
			logger.String("projectID", project.ID),
			logger.String("statusCode", fmt.Sprintf("%d", resp.StatusCode)),
		)
		s.addTaskLog(ctx, task.ID, "success", "Jenkins 构建触发成功，开始构建和部署项目")

		// 更新项目状态为部署中
		project.Status = "deploying"
		s.projectRepo.Update(ctx, project)
	} else {
		logger.Error("Jenkins 构建触发失败",
			logger.String("projectID", project.ID),
			logger.String("statusCode", fmt.Sprintf("%d", resp.StatusCode)),
		)
		s.addTaskLog(ctx, task.ID, "error", fmt.Sprintf("Jenkins 构建触发失败，状态码: %d", resp.StatusCode))
	}
}

// triggerJenkinsBuildWithScript 使用脚本触发 Jenkins 构建（备用方案）
func (s *TaskExecutionService) triggerJenkinsBuildWithScript(ctx context.Context, project *models.Project, task *models.Task) {
	logger.Info("使用脚本触发 Jenkins 构建",
		logger.String("projectID", project.ID),
		logger.String("taskID", task.ID),
	)

	s.addTaskLog(ctx, task.ID, "info", "使用脚本触发 Jenkins 构建...")

	// 构建脚本命令
	scriptPath := "/scripts/jenkins-trigger.sh"
	cmd := exec.CommandContext(ctx, "bash", scriptPath,
		"--user-id", project.UserID,
		"--project-id", project.ID,
		"--project-path", project.ProjectPath,
		"--build-type", "dev",
		"--jenkins-url", s.jenkinsConfig.BaseURL,
		"--job-name", s.jenkinsConfig.JobName,
	)

	// 执行脚本
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Error("Jenkins 构建脚本执行失败",
			logger.String("projectID", project.ID),
			logger.String("error", err.Error()),
			logger.String("output", string(output)),
		)
		s.addTaskLog(ctx, task.ID, "error", fmt.Sprintf("Jenkins 构建脚本执行失败: %s", err.Error()))
		return
	}

	logger.Info("Jenkins 构建脚本执行成功",
		logger.String("projectID", project.ID),
		logger.String("output", string(output)),
	)
	s.addTaskLog(ctx, task.ID, "success", "Jenkins 构建脚本执行成功")
}

// convertToHostPath 将容器内路径转换为主机路径
func (s *TaskExecutionService) convertToHostPath(containerPath string) string {
	// 容器内路径: /app/data/projects/USER00000000002/PROJ00000000006
	// 主机路径: F:/app-maker/app_data/projects/USER00000000002/PROJ00000000006

	// 获取主机数据目录
	hostDataDir := getEnvOrDefault("APP_DATA_HOME", "F:/app-maker/app_data")

	// 替换路径前缀
	if strings.HasPrefix(containerPath, "/app/data/") {
		relativePath := strings.TrimPrefix(containerPath, "/app/data/")
		return filepath.Join(hostDataDir, relativePath)
	}

	// 如果不是预期的路径格式，返回原路径
	return containerPath
}
