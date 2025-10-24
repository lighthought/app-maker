package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/lighthought/app-maker/backend/internal/models"
	"github.com/lighthought/app-maker/backend/internal/repositories"
	"github.com/lighthought/app-maker/shared-models/agent"
	"github.com/lighthought/app-maker/shared-models/client"
	"github.com/lighthought/app-maker/shared-models/common"
	"github.com/lighthought/app-maker/shared-models/logger"
	"github.com/lighthought/app-maker/shared-models/tasks"
	"github.com/lighthought/app-maker/shared-models/utils"
)

// Agent 交互服务接口，用于处理和 Agent 交互的请求和响应
type AgentInteractService interface {
	// 与 Agent 对话
	ChatWithAgent(ctx context.Context, project *models.Project, req *agent.ChatReq) (string, error)

	// 检查 Agent 版本信息
	CheckAgentVersion(ctx context.Context) (*agent.AgentHealthResp, error)

	// 等待任务完成
	WaitForTaskCompletion(ctx context.Context, taskID string) (*tasks.TaskResult, error)

	// 准备项目 Agents 环境
	SetupAgentsEnviroment(ctx context.Context, project *models.Project) (string, error)

	// 检查需求
	CheckRequirement(ctx context.Context, project *models.Project) (string, error)

	// 生成PRD
	GeneratePRD(ctx context.Context, project *models.Project) (string, error)

	// 定义UX标准
	DefineUXStandards(ctx context.Context, project *models.Project) (string, error)

	// 设计架构
	DesignArchitecture(ctx context.Context, project *models.Project) (string, error)

	// 定义数据模型
	DefineDataModel(ctx context.Context, project *models.Project) (string, error)

	// 定义API接口
	DefineAPIs(ctx context.Context, project *models.Project) (string, error)

	// 划分Epic和Story
	PlanEpicsAndStories(ctx context.Context, project *models.Project) (string, error)

	// 生成前端页面
	GenerateFrontendPages(ctx context.Context, project *models.Project) (string, error)

	// 开发Story
	DevelopStories(ctx context.Context, project *models.Project) (string, error)

	// 修复Bug
	FixBugs(ctx context.Context, project *models.Project) (string, error)

	// 运行测试
	RunTests(ctx context.Context, project *models.Project) (string, error)

	// 打包部署
	PackageProject(ctx context.Context, project *models.Project) (string, error)
}

// Agent 交互服务实现
type agentInteractService struct {
	repositories   *repositories.Repository
	agentsURL      string
	defaultTimeout time.Duration
}

// NewAgentInteractService 创建 Agent 交互服务
func NewAgentInteractService(repositories *repositories.Repository, agentsURL string) AgentInteractService {
	return &agentInteractService{
		repositories:   repositories,
		agentsURL:      agentsURL,
		defaultTimeout: time.Duration(5 * time.Minute),
	}
}

// getCliTool 获取项目的 CLI 工具类型
func (s *agentInteractService) getCliTool(project *models.Project) string {
	cliTool := project.CliTool
	if cliTool == "" {
		cliTool = project.User.DefaultCliTool
	}
	if cliTool == "" {
		cliTool = common.CliToolClaudeCode
	}
	return cliTool
}

func (s *agentInteractService) getAgentClient(timeout time.Duration) *client.AgentClient {
	if s.agentsURL == "" {
		s.agentsURL = utils.GetEnvOrDefault("AGENTS_SERVER_URL", "http://localhost:8088")
	}
	return client.NewAgentClient(s.agentsURL, timeout)
}

// ChatWithAgent 与 Agent 对话
func (s *agentInteractService) ChatWithAgent(ctx context.Context, project *models.Project, req *agent.ChatReq) (string, error) {
	req.CliTool = s.getCliTool(project)
	agentClient := s.getAgentClient(5 * time.Minute)
	return agentClient.ChatWithAgent(ctx, req)
}

// CheckAgentVersion 检查 Agent 版本信息
func (s *agentInteractService) CheckAgentVersion(ctx context.Context) (*agent.AgentHealthResp, error) {
	agentClient := s.getAgentClient(5 * time.Minute)

	healthResp, err := agentClient.CheckVersion(ctx)
	if err != nil {
		logger.Error("agent health check failed",
			logger.String("agentsURL", s.agentsURL),
			logger.String("error", err.Error()))
		return nil, fmt.Errorf("agent server is not available: %s", err.Error())
	}

	logger.Info("Agent 版本检查成功",
		logger.String("status", healthResp.Status),
		logger.String("version", healthResp.Version))

	return healthResp, nil
}

func (s *agentInteractService) WaitForTaskCompletion(ctx context.Context, taskID string) (*tasks.TaskResult, error) {
	agentClient := s.getAgentClient(s.defaultTimeout)
	return agentClient.WaitForTaskCompletion(ctx, taskID)
}

// checkAgentHealthWithTimeout 带超时的 Agent 健康检查
func (s *agentInteractService) checkAgentHealthWithTimeout(ctx context.Context, timeout time.Duration) error {
	// 创建带超时的上下文
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if s.agentsURL == "" {
		s.agentsURL = utils.GetEnvOrDefault("AGENTS_SERVER_URL", "http://localhost:8088")
	}

	logger.Info("开始检查 Agent 服务健康状态",
		logger.String("agentsURL", s.agentsURL))

	agentClient := client.NewAgentClient(s.agentsURL, 10*time.Second) // 健康检查的空接口，10s足够了
	err := agentClient.CheckHealth(timeoutCtx)
	if err != nil {
		logger.Error("agent health check failed",
			logger.String("agentsURL", s.agentsURL),
			logger.String("error", err.Error()))
		return fmt.Errorf("agent server is not available: %s", err.Error())
	}

	logger.Info("Agent 健康检查成功")

	return nil
}

// IsAgentHealthy 检查 Agent 是否健康（简化版，返回布尔值）
func (s *agentInteractService) IsAgentHealthy(ctx context.Context) bool {
	// 使用较短的超时时间进行快速检查
	err := s.checkAgentHealthWithTimeout(ctx, 5*time.Second)
	if err != nil {
		logger.Warn("Agent 服务不可用", logger.String("error", err.Error()))
		return false
	}
	return true
}

// pendingAgents 准备项目开发环境
func (s *agentInteractService) SetupAgentsEnviroment(ctx context.Context,
	project *models.Project) (string, error) {
	// 设置 CLI 工具和模型配置，如果项目没有设置则使用用户的默认设置
	cliTool := project.CliTool
	aiModel := project.AiModel
	modelProvider := project.ModelProvider
	modelApiUrl := project.ModelApiUrl

	if cliTool == "" {
		cliTool = project.User.DefaultCliTool
	}
	if aiModel == "" {
		aiModel = project.User.DefaultAiModel
	}
	if modelProvider == "" {
		modelProvider = project.User.DefaultModelProvider
	}
	if modelApiUrl == "" {
		modelApiUrl = project.User.DefaultModelApiUrl
	}

	// 如果还是空，使用系统默认值
	if cliTool == "" {
		cliTool = common.CliToolClaudeCode
	}
	if aiModel == "" {
		aiModel = common.DefaultModelByProvider[common.ModelProviderZhipu]
	}
	if modelProvider == "" {
		modelProvider = common.ModelProviderZhipu
	}
	if modelApiUrl == "" {
		modelApiUrl = common.DefaultAPIUrlByProvider[common.ModelProviderZhipu]
	}

	// 获取 API Token
	apiToken := project.ApiToken
	if apiToken == "" {
		apiToken = project.User.DefaultApiToken
	}

	agentClient := s.getAgentClient(5 * time.Minute)
	taskID, err := agentClient.SetupProjectEnvironment(ctx, &agent.SetupProjEnvReq{
		ProjectGuid:     project.GUID,
		GitlabRepoUrl:   project.GitlabRepoURL,
		SetupBmadMethod: true,
		BmadCliType:     cliTool,
		AiModel:         aiModel,
		ModelProvider:   modelProvider,
		ModelApiUrl:     modelApiUrl,
		ApiToken:        apiToken,
	})
	if err != nil {
		return "", err
	}

	return taskID, nil
}

// checkRequirement 检查需求
func (s *agentInteractService) CheckRequirement(ctx context.Context,
	project *models.Project) (string, error) {
	req := &agent.GetProjBriefReq{
		ProjectGuid:  project.GUID,
		Requirements: project.Requirements,
		CliTool:      s.getCliTool(project),
	}

	agentClient := s.getAgentClient(s.defaultTimeout)
	taskID, err := agentClient.AnalyseProjectBrief(ctx, req)
	if err != nil {
		return "", err
	}

	return taskID, nil
}

// generatePRD 生成PRD文档
func (s *agentInteractService) GeneratePRD(ctx context.Context,
	project *models.Project) (string, error) {
	generatePrdReq := &agent.GetPRDReq{
		ProjectGuid:  project.GUID,
		Requirements: project.Requirements,
		CliTool:      s.getCliTool(project),
	}
	// 调用 agents-server 生成 PRD 文档，并提交到 GitLab
	agentClient := s.getAgentClient(s.defaultTimeout)
	taskID, err := agentClient.GetPRD(ctx, generatePrdReq)
	if err != nil {
		return "", err
	}

	return taskID, nil
}

// DefineUXStandards 定义UX标准
func (s *agentInteractService) DefineUXStandards(ctx context.Context,
	project *models.Project) (string, error) {
	req := &agent.GetUXStandardReq{
		ProjectGuid:  project.GUID,
		Requirements: project.Requirements,
		PrdPath:      PATH_PRD,
		CliTool:      s.getCliTool(project),
	}
	// 调用 agents-server 定义 UX 标准
	agentClient := s.getAgentClient(s.defaultTimeout)
	taskID, err := agentClient.GetUXStandard(ctx, req)
	if err != nil {
		return "", err
	}

	return taskID, nil
}

// designArchitecture 设计系统架构
func (s *agentInteractService) DesignArchitecture(ctx context.Context,
	project *models.Project) (string, error) {
	req := &agent.GetArchitectureReq{
		ProjectGuid: project.GUID,
		PrdPath:     PATH_PRD,
		UxSpecPath:  PATH_UX_SPEC,
		// 从模板中读取架构信息
		TemplateArchDescription: "1. 前端：vue.js+ vite ；\n" +
			"2. 后端服务和 API： GO + Gin 框架实现 API、数据库用 PostgreSql、缓存用 Redis。\n" +
			"3. 部署相关的脚本已经有了，用的 docker，前端用一个 nginx ，配置 /api 重定向到 /backend:port ，这样就能在前端项目中访问后端 API 了。" +
			" 引用关系是：前端依赖后端，后端依赖 Redis 和 PostgreSql。",
		CliTool: s.getCliTool(project),
	}
	// 调用 agents-server 设计系统架构
	agentClient := s.getAgentClient(s.defaultTimeout)
	taskID, err := agentClient.GetArchitecture(ctx, req)
	if err != nil {
		return "", err
	}

	return taskID, nil
}

// defineDataModel 定义数据模型
func (s *agentInteractService) DefineDataModel(ctx context.Context,
	project *models.Project) (string, error) {
	req := &agent.GetDatabaseDesignReq{
		ProjectGuid:   project.GUID,
		PrdPath:       PATH_PRD,
		ArchFolder:    "docs/arch",
		StoriesFolder: FOLDER_STORIES,
		CliTool:       s.getCliTool(project),
	}
	// 调用 agents-server 定义数据模型
	agentClient := s.getAgentClient(s.defaultTimeout)
	taskID, err := agentClient.GetDatabaseDesign(ctx, req)
	if err != nil {
		return "", err
	}

	return taskID, nil
}

// defineAPIs 定义API接口
func (s *agentInteractService) DefineAPIs(ctx context.Context,
	project *models.Project) (string, error) {
	req := &agent.GetAPIDefinitionReq{
		ProjectGuid:   project.GUID,
		PrdPath:       PATH_PRD,
		DbFolder:      "docs/db",
		StoriesFolder: FOLDER_STORIES,
		CliTool:       s.getCliTool(project),
	}
	// 调用 agents-server 定义 API 接口
	agentClient := s.getAgentClient(s.defaultTimeout)
	taskID, err := agentClient.GetAPIDefinition(ctx, req)
	if err != nil {
		return "", err
	}

	return taskID, nil
}

// planEpicsAndStories 划分Epic和Story
func (s *agentInteractService) PlanEpicsAndStories(ctx context.Context,
	project *models.Project) (string, error) {
	req := &agent.GetEpicsAndStoriesReq{
		ProjectGuid: project.GUID,
		PrdPath:     PATH_PRD,
		ArchFolder:  "docs/arch",
		CliTool:     s.getCliTool(project),
	}
	// 调用 agents-server 划分 Epics 和 Stories
	agentClient := s.getAgentClient(s.defaultTimeout)
	taskID, err := agentClient.GetEpicsAndStories(ctx, req)
	if err != nil {
		return "", err
	}

	return taskID, nil
}

// generateFrontendPages 生成前端关键页面 (Vibe Coding)
func (s *agentInteractService) GenerateFrontendPages(ctx context.Context,
	project *models.Project) (string, error) {
	// 检查 page-prompt.md 文件是否存在
	pagePromptRelPath := "docs/ux/page-prompt.md"
	pagePromptFiles, _ := utils.GetRelativeFiles(project.ProjectPath, "docs/ux")
	hasPagePrompt := false
	for _, file := range pagePromptFiles {
		if strings.Contains(file, "page-prompt") || strings.Contains(file, "prompt") {
			hasPagePrompt = true
			break
		}
	}

	if !hasPagePrompt {
		logger.Warn("未找到 page-prompt.md 文件，跳过前端页面生成")
		return "", nil
	}

	logger.Info("开始生成前端页面", logger.String("pagePromptPath", pagePromptRelPath))

	// 根据 CLI 类型选择不同的 prompt
	var agentPrompt string
	cliTool := s.getCliTool(project)
	if cliTool == common.CliToolGemini {
		agentPrompt = "@.bmad-core/agents/dev.md"
	} else {
		agentPrompt = "@bmad/dev.mdc"
	}

	// 调用 Dev Agent 生成前端页面
	message := agentPrompt + " 请基于 @docs/ux/page-prompt.md 中的页面设计提示词," +
		"在前端项目 frontend/src/pages/ 目录下生成关键页面组件。" +
		"使用 Vue 3 + TypeScript + Naive UI,遵循现有项目的代码风格和架构。" +
		"只生成 page-prompt.md 中明确定义的页面，不要生成其他页面。" +
		"注意：始终用中文回答我。"

	req := &agent.ChatReq{
		ProjectGuid: project.GUID,
		AgentType:   common.AgentTypeDev,
		Message:     message,
	}

	agentClient := s.getAgentClient(s.defaultTimeout)
	taskID, err := agentClient.ChatWithAgent(ctx, req)
	if err != nil {
		logger.Error("生成前端页面失败", logger.String("error", err.Error()))
		return "", err
	}
	logger.Info("前端页面生成完成")
	return taskID, nil
}

// DevelopStories 开发Story功能 (只实现 MVP Stories)
func (s *agentInteractService) DevelopStories(ctx context.Context,
	project *models.Project) (string, error) {
	// 尝试从数据库获取 MVP 阶段的 Epics (P0 优先级)
	mvpEpics, err := s.repositories.EpicRepo.GetMvpEpicsByProject(ctx, project.ID)

	agentClient := s.getAgentClient(s.defaultTimeout)
	// 如果数据库中没有 MVP Epics，fallback 到文件方式
	if err != nil || len(mvpEpics) == 0 {
		logger.Warn("数据库中未找到 MVP Epics，使用文件方式", logger.String("error", err.Error()))
		return s.DevelopStoriesFromFiles(ctx, project, agentClient)
	}

	logger.Info("从数据库读取到 MVP Epics", logger.Int("count", len(mvpEpics)))

	req := &agent.ImplementStoryReq{
		ProjectGuid: project.GUID,
		PrdPath:     PATH_PRD,
		ArchFolder:  "docs/arch/",
		DbFolder:    "docs/db/",
		ApiFolder:   "docs/api/",
		UxSpecPath:  PATH_UX_SPEC,
		EpicFile:    "docs/stories/",
		StoryFile:   "",
		CliTool:     s.getCliTool(project),
	}

	developStoryCount := 0
	var lastResponse *tasks.TaskResult

	// 按 Epic 和 Story 的顺序实现
	// TODO: 现在 agent 接口改异步了，通过 sub 服务返回，这里需要重新设计，看怎么做到一个个实现 MVP 的 epics 和用户故事逐步开发
	for epicIndex, epic := range mvpEpics {
		_, err = s.DevelopEpicStories(ctx, project, agentClient, req, epic, epicIndex, len(mvpEpics))
		if err != nil {
			return "", err
		}
		if lastResponse != nil {
			developStoryCount += 1
		}
	}

	return "", nil
}

// 开发单个故事
func (s *agentInteractService) developSingleStory(ctx context.Context, agentClient *client.AgentClient,
	req *agent.ImplementStoryReq, story *models.Story) (string, error) {
	// 设置 Story 文件路径
	req.StoryFile = story.FilePath
	req.EpicFile = story.FilePath

	logger.Info("开始实现 Story",
		logger.String("story_number", story.StoryNumber),
		logger.String("story_title", story.Title),
		logger.String("story_file", story.FilePath))

	// 调用 Dev Agent 实现 Story
	taskID, err := agentClient.ImplementStory(ctx, req)
	if err != nil {
		logger.Error("Story 实现失败",
			logger.String("story_number", story.StoryNumber),
			logger.String("error", err.Error()))

		// 更新 Story 状态为失败
		story.Status = common.CommonStatusFailed
		s.repositories.StoryRepo.Update(ctx, story)
		return "", err
	}

	// 更新 Story 状态为完成
	story.Status = common.CommonStatusDone
	if err := s.repositories.StoryRepo.Update(ctx, story); err != nil {
		logger.Error("更新 Story 状态失败", logger.String("error", err.Error()))
	}

	logger.Info("Story 实现成功",
		logger.String("story_number", story.StoryNumber),
		logger.String("story_title", story.Title))
	return taskID, nil
}

// updateEpicStatus 更新 Epic 状态
func (s *agentInteractService) updateEpicStatus(ctx context.Context, epic *models.Epic) error {
	allStoriesDone := true
	for _, story := range epic.Stories {
		if story.Status != common.CommonStatusDone {
			allStoriesDone = false
			break
		}
	}
	if allStoriesDone {
		epic.Status = common.CommonStatusDone
		if err := s.repositories.EpicRepo.Update(ctx, epic); err != nil {
			logger.Error("更新 Epic 状态失败", logger.String("error", err.Error()))
			return err
		}
		logger.Info("Epic 已完成", logger.String("epic_name", epic.Name))
	}
	return nil
}

// 开发单个 epic 下面的用户故事
func (s *agentInteractService) DevelopEpicStories(ctx context.Context,
	project *models.Project, agentClient *client.AgentClient,
	req *agent.ImplementStoryReq, epic *models.Epic, epicIndex, mvpEpicCount int) (string, error) {
	logger.Info("开始实现 Epic",
		logger.String("epic_id", epic.ID),
		logger.String("epic_name", epic.Name),
		logger.Int("story_count", len(epic.Stories)))

	var taskID string

	iStoryCount := 0
	for storyIndex, story := range epic.Stories {
		// 跳过已完成的 Story
		if story.Status == common.CommonStatusDone {
			logger.Info("Story 已完成，跳过",
				logger.String("story_number", story.StoryNumber),
				logger.String("story_title", story.Title))
			continue
		}

		// 开发环境只实现第一个 Story
		if iStoryCount >= 1 && utils.IsDevEnvironment() {
			logger.Info("开发模式：跳过 Story",
				logger.String("story_number", story.StoryNumber),
				logger.String("story_title", story.Title))

			// 模拟完成
			// taskID = &tasks.TaskResult{
			// 	Message: fmt.Sprintf("开发模式：跳过 Story %s - %s", story.StoryNumber, story.Title),
			// }
			continue
		}

		response, err := s.developSingleStory(ctx, agentClient, req, &story)
		if err != nil {
			return "", err
		}

		iStoryCount++
		taskID = response

		// 不是最后一个 Story，发送中间消息
		if !(epicIndex == (mvpEpicCount-1) && storyIndex == len(epic.Stories)-1) {
			// projectMsg := models.NewDevAgentMessage(project.GUID, fmt.Sprintf("Story %s 已完成", story.StoryNumber), "")
			// todo: 异步执行多个用户故事的开发过程
			// s.notifyProjectStatusChange(ctx, project, projectMsg, nil)
		}
	}

	// Epic 完成，更新 Epic 状态
	s.updateEpicStatus(ctx, epic)
	return taskID, nil
}

// DevelopStoriesFromFiles 从文件方式开发 Stories (fallback)
func (s *agentInteractService) DevelopStoriesFromFiles(ctx context.Context,
	project *models.Project, agentClient *client.AgentClient) (string, error) {
	req := &agent.ImplementStoryReq{
		ProjectGuid: project.GUID,
		PrdPath:     PATH_PRD,
		ArchFolder:  "docs/arch/",
		DbFolder:    "docs/db/",
		ApiFolder:   "docs/api/",
		UxSpecPath:  PATH_UX_SPEC,
		EpicFile:    "docs/stories/",
		StoryFile:   "",
		CliTool:     s.getCliTool(project),
	}

	storyFiles, err := utils.GetRelativeFiles(project.ProjectPath, FOLDER_STORIES)
	if err != nil || len(storyFiles) == 0 {
		taskID, err := agentClient.ImplementStory(ctx, req)
		if err != nil {
			return "", err
		}

		return taskID, nil
	}

	var response = &tasks.TaskResult{}
	developStoryCount := 0
	bDev := (utils.GetEnvOrDefault("ENVIRONMENT", common.EnvironmentDevelopment) == common.EnvironmentDevelopment)

	var taskID string
	// 获取 stories 下的文件，循环开发每个 Story
	for index, storyFile := range storyFiles {
		// development 模式，只开发一个
		if developStoryCount < 1 || !bDev {
			req.StoryFile = storyFile
			// 调用 agents-server 开发 Story 功能
			taskID, err = agentClient.ImplementStory(ctx, req)
			if err != nil {
				return "", err
			}

			developStoryCount += 1
		} else {
			response.Message = "开发需求故事" + storyFile + "已完成"
		}

		if index < len(storyFiles)-1 {
			// TODO: 更新过程中的消息
			// projectMsg := models.NewDevAgentMessage(project.GUID, MESSAGE_STORY_DEVELOPED, response.Message)
			//s.notifyProjectStatusChange(ctx, project, projectMsg, nil)
		}
	}

	return taskID, nil
}

// fixBugs 修复开发问题
func (s *agentInteractService) FixBugs(ctx context.Context,
	project *models.Project) (string, error) {
	req := &agent.FixBugReq{
		ProjectGuid:    project.GUID,
		BugDescription: "修复开发问题",
		CliTool:        s.getCliTool(project),
	}
	// 调用 agents-server 修复问题
	agentClient := s.getAgentClient(s.defaultTimeout)
	taskID, err := agentClient.FixBug(ctx, req)
	if err != nil {
		return "", err
	}
	return taskID, nil
}

// runTests 执行自动测试
func (s *agentInteractService) RunTests(ctx context.Context,
	project *models.Project) (string, error) {
	req := &agent.RunTestReq{
		ProjectGuid: project.GUID,
		CliTool:     s.getCliTool(project),
	}
	// 调用 agents-server 执行自动测试
	agentClient := s.getAgentClient(s.defaultTimeout)
	taskID, err := agentClient.RunTest(ctx, req)
	if err != nil {
		return "", err
	}

	return taskID, nil
}

// packageProject 打包项目
func (s *agentInteractService) PackageProject(ctx context.Context,
	project *models.Project) (string, error) {
	req := &agent.DeployReq{
		ProjectGuid:   project.GUID,
		Environment:   "dev",
		DeployOptions: map[string]interface{}{},
		CliTool:       s.getCliTool(project),
	}

	agentClient := s.getAgentClient(s.defaultTimeout)
	taskID, err := agentClient.Deploy(ctx, req)
	if err != nil {
		return "", err
	}

	return taskID, nil
}
