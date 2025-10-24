package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/lighthought/app-maker/backend/internal/models"
	"github.com/lighthought/app-maker/backend/internal/repositories"
	"github.com/lighthought/app-maker/shared-models/agent"
	"github.com/lighthought/app-maker/shared-models/common"
	"github.com/lighthought/app-maker/shared-models/logger"
	"github.com/lighthought/app-maker/shared-models/tasks"
)

// 项目开发服务
type ProjectDevService interface {
	// 初始化阶段项
	InitStageItems()

	// 获取阶段项
	GetStageItem(stageName common.DevStatus) *models.DevStageItem

	// 处理 Agent 对话响应
	OnChatResponse(ctx context.Context, message *agent.AgentTaskStatusMessage, response *tasks.TaskResult) error

	// 进入下一阶段的通用方法
	ProceedToNextStage(ctx context.Context,
		project *models.Project, currentStage common.DevStatus) error
}

// 项目开发业务实现
type projectDevService struct {
	repositories         *repositories.Repository
	asyncClientService   AsyncClientService
	agentInteractService AgentInteractService
	commonService        ProjectCommonService
	stageItems           []*models.DevStageItem
}

// NewProjectDevService 创建项目开发服务
func NewProjectDevService(
	repositories *repositories.Repository,
	asyncClientService AsyncClientService,
	agentInteractService AgentInteractService,
	commonService ProjectCommonService,
) ProjectDevService {
	return &projectDevService{
		repositories:         repositories,
		asyncClientService:   asyncClientService,
		agentInteractService: agentInteractService,
		commonService:        commonService,
	}
}

// InitStageItems 初始化阶段项
// 开发调试阶段，修改每个阶段的 SkipInDevMode 属性就可以了
func (s *projectDevService) InitStageItems() {
	s.stageItems = []*models.DevStageItem{
		{Name: common.DevStatusSetupAgents, Desc: "准备项目 Agents 环境", NeedConfirm: false,
			ReqHandler: s.agentInteractService.SetupAgentsEnviroment, RespHandler: s.OnPendingAgentResponse},
		{Name: common.DevStatusCheckRequirement, Desc: "检查需求", NeedConfirm: true,
			ReqHandler: s.agentInteractService.CheckRequirement, RespHandler: s.OnCheckRequirementResponse},
		{Name: common.DevStatusGeneratePRD, Desc: "生成PRD", NeedConfirm: true,
			ReqHandler: s.agentInteractService.GeneratePRD, RespHandler: s.OnGeneratePRDResponse},
		{Name: common.DevStatusDefineUXStandard, Desc: "定义UX标准", NeedConfirm: true,
			ReqHandler: s.agentInteractService.DefineUXStandards, RespHandler: s.OnDefineUXStandardsResponse},
		{Name: common.DevStatusDesignArchitecture, Desc: "设计架构", NeedConfirm: true,
			ReqHandler: s.agentInteractService.DesignArchitecture, RespHandler: s.OnDesignArchitectureResponse},
		{Name: common.DevStatusPlanEpicAndStory, Desc: "规划 Epic 和 Story", NeedConfirm: true,
			ReqHandler: s.agentInteractService.PlanEpicsAndStories, RespHandler: s.OnPlanEpicsAndStoriesResponse},
		{Name: common.DevStatusDefineDataModel, Desc: "定义数据模型", NeedConfirm: true,
			ReqHandler: s.agentInteractService.DefineDataModel, RespHandler: s.OnDefineDataModelResponse, SkipInDevMode: true},
		{Name: common.DevStatusDefineAPI, Desc: "定义 API", NeedConfirm: true,
			ReqHandler: s.agentInteractService.DefineAPIs, RespHandler: s.OnDefineAPIsResponse, SkipInDevMode: true},
		{Name: common.DevStatusGeneratePages, Desc: "生成前端页面", NeedConfirm: true,
			ReqHandler: s.agentInteractService.GenerateFrontendPages, RespHandler: s.OnGenerateFrontendPagesResponse, SkipInDevMode: true},
		{Name: common.DevStatusDevelopStory, Desc: "开发 Story", NeedConfirm: true,
			ReqHandler: s.agentInteractService.DevelopStories, RespHandler: s.OnDevelopStoriesResponse, SkipInDevMode: true},
		{Name: common.DevStatusFixBug, Desc: "修复 Bug", NeedConfirm: false,
			ReqHandler: s.agentInteractService.FixBugs, RespHandler: s.OnFixBugsResponse, SkipInDevMode: true},
		{Name: common.DevStatusRunTest, Desc: "运行测试", NeedConfirm: false,
			ReqHandler: s.agentInteractService.RunTests, RespHandler: s.OnRunTestsResponse, SkipInDevMode: true},
		{Name: common.DevStatusDeploy, Desc: "打包部署", NeedConfirm: false,
			ReqHandler: s.agentInteractService.PackageProject, RespHandler: s.OnPackageProjectResponse},
	}
}

// ProceedToNextStage 进入下一阶段的通用方法
func (s *projectDevService) ProceedToNextStage(ctx context.Context,
	project *models.Project, currentStage common.DevStatus) error {
	// 优先使用项目级配置，其次用户级配置
	// autoGoNext := project.AutoGoNext
	// if !autoGoNext {
	// 	autoGoNext = project.User.AutoGoNext
	// }

	// if requireConfirm && !autoGoNext {
	// 	s.commonService.UpdateProjectWaitingForUserConfirm(ctx, project, currentStage)
	// 	return nil
	// }

	// 自动进入下一阶段
	nextStage := s.getNextStage(currentStage)
	if nextStage == nil {
		s.commonService.UpdateProjectToStatus(ctx, project, common.CommonStatusDone) // 没有下一阶段，项目完成
		s.commonService.EnsureProjectPrevieUrl(ctx, project.GUID)
		return nil
	}

	// 创建下一阶段任务
	s.asyncClientService.EnqueueProjectStageTask(nextStage.NeedConfirm, project.GUID, string(nextStage.Name))
	return nil
}

// getNextStage 获取下一阶段
func (s *projectDevService) getNextStage(currentStage common.DevStatus) *models.DevStageItem {
	// 定义需要执行的阶段数组，方便调试过程中跳过耗时较多的故事实现等阶段
	for index, processStage := range s.stageItems {
		if currentStage == processStage.Name {
			// 返回下一个
			if index+1 < len(s.stageItems) {
				return s.stageItems[index+1]
			} else {
				return nil
			}
		}
	}
	return nil
}

// getStageExecutor 获取阶段执行器
func (s *projectDevService) GetStageItem(stageName common.DevStatus) *models.DevStageItem {
	for _, stage := range s.stageItems {
		if stage.Name == stageName {
			return stage
		}
	}
	return nil
}

// OnChatResponse 处理 Agent 对话响应
func (s *projectDevService) OnChatResponse(ctx context.Context, message *agent.AgentTaskStatusMessage, response *tasks.TaskResult) error {
	agentRole := common.GetAgentByAgentType(message.AgentType)
	projectMsg := &models.ConversationMessage{
		ProjectGuid:     message.ProjectGuid,
		Type:            common.ConversationTypeAgent,
		AgentRole:       agentRole.Role,
		AgentName:       agentRole.Name,
		Content:         "Agent 已完成",
		IsMarkdown:      true,
		MarkdownContent: response.Message,
		IsExpanded:      true,
	}

	return s.commonService.CreateAndNotifyMessage(ctx, message.ProjectGuid, projectMsg)
}

// OnPendingAgentResponse 处理 Agent 准备项目环境响应
func (s *projectDevService) OnPendingAgentResponse(ctx context.Context, message *agent.AgentTaskStatusMessage, response *tasks.TaskResult) error {
	projectMsg := &models.ConversationMessage{
		ProjectGuid:     message.ProjectGuid,
		Type:            common.ConversationTypeAgent,
		AgentRole:       common.AgentPM.Role,
		AgentName:       common.AgentPM.Name,
		Content:         "项目开发环境已准备完成",
		IsMarkdown:      true,
		MarkdownContent: response.Message,
		IsExpanded:      true,
	}

	return s.commonService.CreateAndNotifyMessage(ctx, message.ProjectGuid, projectMsg)
}

// OnCheckRequirementResponse 处理检查需求响应
func (s *projectDevService) OnCheckRequirementResponse(ctx context.Context, message *agent.AgentTaskStatusMessage, response *tasks.TaskResult) error {
	projectMsg := &models.ConversationMessage{
		ProjectGuid:     message.ProjectGuid,
		Type:            common.ConversationTypeAgent,
		AgentRole:       common.AgentAnalyst.Role,
		AgentName:       common.AgentAnalyst.Name,
		Content:         "项目需求已检查完成",
		IsMarkdown:      true,
		MarkdownContent: response.Message,
		IsExpanded:      true,
	}
	return s.commonService.CreateAndNotifyMessage(ctx, message.ProjectGuid, projectMsg)
}

// OnGeneratePRDResponse 处理生成 PRD 响应
func (s *projectDevService) OnGeneratePRDResponse(ctx context.Context, message *agent.AgentTaskStatusMessage, response *tasks.TaskResult) error {
	projectMsg := &models.ConversationMessage{
		ProjectGuid:     message.ProjectGuid,
		Type:            common.ConversationTypeAgent,
		AgentRole:       common.AgentPM.Role,
		AgentName:       common.AgentPM.Name,
		Content:         "项目PRD文档已生成",
		IsMarkdown:      true,
		MarkdownContent: response.Message,
		IsExpanded:      true,
	}

	return s.commonService.CreateAndNotifyMessage(ctx, message.ProjectGuid, projectMsg)
}

// OnDefineUXStandardsResponse 处理定义 UX 标准响应
func (s *projectDevService) OnDefineUXStandardsResponse(ctx context.Context, message *agent.AgentTaskStatusMessage, response *tasks.TaskResult) error {
	projectMsg := &models.ConversationMessage{
		ProjectGuid:     message.ProjectGuid,
		Type:            common.ConversationTypeAgent,
		AgentRole:       common.AgentUXExpert.Role,
		AgentName:       common.AgentUXExpert.Name,
		Content:         "项目UX标准已定义",
		IsMarkdown:      true,
		MarkdownContent: response.Message,
		IsExpanded:      true,
	}
	return s.commonService.CreateAndNotifyMessage(ctx, message.ProjectGuid, projectMsg)
}

// OnDesignArchitectureResponse 处理设计系统架构响应
func (s *projectDevService) OnDesignArchitectureResponse(ctx context.Context, message *agent.AgentTaskStatusMessage, response *tasks.TaskResult) error {
	projectMsg := &models.ConversationMessage{
		ProjectGuid:     message.ProjectGuid,
		Type:            common.ConversationTypeAgent,
		AgentRole:       common.AgentArchitect.Role,
		AgentName:       common.AgentArchitect.Name,
		Content:         "项目系统架构已设计",
		IsMarkdown:      true,
		MarkdownContent: response.Message,
		IsExpanded:      true,
	}

	return s.commonService.CreateAndNotifyMessage(ctx, message.ProjectGuid, projectMsg)
}

// OnDefineDataModelResponse 处理定义数据模型响应
func (s *projectDevService) OnDefineDataModelResponse(ctx context.Context, message *agent.AgentTaskStatusMessage, response *tasks.TaskResult) error {
	projectMsg := &models.ConversationMessage{
		ProjectGuid:     message.ProjectGuid,
		Type:            common.ConversationTypeAgent,
		AgentRole:       common.AgentArchitect.Role,
		AgentName:       common.AgentArchitect.Name,
		Content:         "项目数据模型已定义",
		IsMarkdown:      true,
		MarkdownContent: response.Message,
		IsExpanded:      true,
	}

	return s.commonService.CreateAndNotifyMessage(ctx, message.ProjectGuid, projectMsg)
}

// OnDefineAPIsResponse 处理定义 API 接口响应
func (s *projectDevService) OnDefineAPIsResponse(ctx context.Context, message *agent.AgentTaskStatusMessage, response *tasks.TaskResult) error {
	projectMsg := &models.ConversationMessage{
		ProjectGuid:     message.ProjectGuid,
		Type:            common.ConversationTypeAgent,
		AgentRole:       common.AgentArchitect.Role,
		AgentName:       common.AgentArchitect.Name,
		Content:         "项目API接口已定义",
		IsMarkdown:      true,
		MarkdownContent: response.Message,
		IsExpanded:      true,
	}

	return s.commonService.CreateAndNotifyMessage(ctx, message.ProjectGuid, projectMsg)
}

// extractMvpEpicsJSON 从 markdown 内容中提取 MVP Epics JSON
func (s *projectDevService) extractMvpEpicsJSON(content string) (*models.MvpEpicsData, error) {
	// 查找 JSON 代码块
	jsonStart := strings.Index(content, "```json")
	if jsonStart == -1 {
		logger.Warn("未找到 JSON 代码块")
		return nil, fmt.Errorf("未找到 JSON 代码块")
	}

	jsonStart += len("```json")
	jsonEnd := strings.Index(content[jsonStart:], "```")
	if jsonEnd == -1 {
		logger.Warn("JSON 代码块未闭合")
		return nil, fmt.Errorf("JSON 代码块未闭合")
	}

	jsonContent := strings.TrimSpace(content[jsonStart : jsonStart+jsonEnd])

	var mvpData models.MvpEpicsData
	if err := json.Unmarshal([]byte(jsonContent), &mvpData); err != nil {
		logger.Error("解析 MVP Epics JSON 失败", logger.String("error", err.Error()))
		return nil, fmt.Errorf("解析 MVP Epics JSON 失败: %w", err)
	}

	logger.Info("成功解析 MVP Epics JSON", logger.Int("epic_count", len(mvpData.MvpEpics)))
	return &mvpData, nil
}

// SaveMvpEpics 保存 MVP Epics 到数据库
func (s *projectDevService) saveMvpEpics(ctx context.Context, projectGuid string, mvpData *models.MvpEpicsData) error {
	if mvpData == nil || len(mvpData.MvpEpics) == 0 {
		return fmt.Errorf("MVP Epics 数据为空")
	}

	project, err := s.repositories.ProjectRepo.GetByGUID(ctx, projectGuid)
	if err != nil {
		return fmt.Errorf("获取项目失败: %w", err)
	}

	// 遍历每个 Epic
	for _, epicItem := range mvpData.MvpEpics {
		// 创建 Epic
		epic := &models.Epic{
			ProjectID:     project.ID,
			ProjectGuid:   project.GUID,
			EpicNumber:    epicItem.EpicNumber,
			Name:          epicItem.Name,
			Description:   epicItem.Description,
			Priority:      epicItem.Priority,
			EstimatedDays: epicItem.EstimatedDays,
			Status:        common.CommonStatusPending,
			FilePath:      epicItem.FilePath,
		}

		// 保存 Epic
		if err := s.repositories.EpicRepo.Create(ctx, epic); err != nil {
			logger.Error("保存 Epic 失败",
				logger.String("epic_name", epic.Name),
				logger.String("error", err.Error()))
			return fmt.Errorf("保存 Epic 失败: %w", err)
		}

		logger.Info("Epic 已保存",
			logger.String("epic_id", epic.ID),
			logger.String("epic_name", epic.Name))

		// 遍历 Epic 下的每个 Story
		for _, storyItem := range epicItem.Stories {
			story := &models.Story{
				EpicID:        epic.ID,
				StoryNumber:   storyItem.StoryNumber,
				Title:         storyItem.Title,
				Description:   storyItem.Description,
				Priority:      storyItem.Priority,
				EstimatedDays: storyItem.EstimatedDays,
				Status:        common.CommonStatusPending,
				FilePath:      epic.FilePath, // Story 的 FilePath 与 Epic 相同
				Depends:       storyItem.Depends,
				Techs:         storyItem.Techs,
			}

			// 保存 Story
			if err := s.repositories.StoryRepo.Create(ctx, story); err != nil {
				logger.Error("保存 Story 失败",
					logger.String("story_number", story.StoryNumber),
					logger.String("story_title", story.Title),
					logger.String("error", err.Error()))
				return fmt.Errorf("保存 Story 失败: %w", err)
			}

			logger.Info("Story 已保存",
				logger.String("story_id", story.ID),
				logger.String("story_number", story.StoryNumber),
				logger.String("story_title", story.Title))
		}
	}

	logger.Info("所有 MVP Epics 和 Stories 已保存",
		logger.Int("epic_count", len(mvpData.MvpEpics)))
	return nil
}

// OnPlanEpicsAndStoriesResponse 处理划分 Epic 和 Story 响应
func (s *projectDevService) OnPlanEpicsAndStoriesResponse(ctx context.Context, message *agent.AgentTaskStatusMessage, response *tasks.TaskResult) error {
	// 解析返回的 markdown 中的 MVP Epics JSON 信息
	mvpData, err := s.extractMvpEpicsJSON(response.Message)
	if err == nil && mvpData != nil {
		// 保存到数据库
		if err := s.saveMvpEpics(ctx, message.ProjectGuid, mvpData); err != nil {
			logger.Error("保存 MVP Epics 失败", logger.String("error", err.Error()))
		} else {
			logger.Info("MVP Epics 已保存到数据库")
		}
	} else {
		logger.Warn("未能提取 MVP Epics JSON，将依赖文件方式读取", logger.String("error", err.Error()))
	}

	projectMsg := &models.ConversationMessage{
		ProjectGuid:     message.ProjectGuid,
		Type:            common.ConversationTypeAgent,
		AgentRole:       common.AgentPO.Role,
		AgentName:       common.AgentPO.Name,
		Content:         "项目Epic和Story已划分",
		IsMarkdown:      true,
		MarkdownContent: response.Message,
		IsExpanded:      true,
	}

	return s.commonService.CreateAndNotifyMessage(ctx, message.ProjectGuid, projectMsg)
}

// OnGenerateFrontendPagesResponse 处理生成前端页面响应
func (s *projectDevService) OnGenerateFrontendPagesResponse(ctx context.Context, message *agent.AgentTaskStatusMessage, response *tasks.TaskResult) error {
	projectMsg := &models.ConversationMessage{
		ProjectGuid:     message.ProjectGuid,
		Type:            common.ConversationTypeAgent,
		AgentRole:       common.AgentDev.Role,
		AgentName:       common.AgentDev.Name,
		Content:         "前端关键页面已生成",
		IsMarkdown:      true,
		MarkdownContent: response.Message,
		IsExpanded:      true,
	}
	return s.commonService.CreateAndNotifyMessage(ctx, message.ProjectGuid, projectMsg)
}

// OnDevelopStoriesResponse 处理开发 Story 响应
func (s *projectDevService) OnDevelopStoriesResponse(ctx context.Context, message *agent.AgentTaskStatusMessage, response *tasks.TaskResult) error {
	projectMsg := &models.ConversationMessage{
		ProjectGuid:     message.ProjectGuid,
		Type:            common.ConversationTypeAgent,
		AgentRole:       common.AgentDev.Role,
		AgentName:       common.AgentDev.Name,
		Content:         MESSAGE_STORY_DEVELOPED,
		IsMarkdown:      true,
		MarkdownContent: response.Message,
		IsExpanded:      true,
	}

	return s.commonService.CreateAndNotifyMessage(ctx, message.ProjectGuid, projectMsg)
}

// OnFixBugsResponse 处理修复 Bug 响应
func (s *projectDevService) OnFixBugsResponse(ctx context.Context, message *agent.AgentTaskStatusMessage, response *tasks.TaskResult) error {
	projectMsg := &models.ConversationMessage{
		ProjectGuid:     message.ProjectGuid,
		Type:            common.ConversationTypeAgent,
		AgentRole:       common.AgentDev.Role,
		AgentName:       common.AgentDev.Name,
		Content:         "项目开发问题已修复",
		IsMarkdown:      true,
		MarkdownContent: response.Message,
		IsExpanded:      true,
	}

	return s.commonService.CreateAndNotifyMessage(ctx, message.ProjectGuid, projectMsg)
}

// OnRunTestsResponse 处理运行测试响应
func (s *projectDevService) OnRunTestsResponse(ctx context.Context, message *agent.AgentTaskStatusMessage, response *tasks.TaskResult) error {
	projectMsg := &models.ConversationMessage{
		ProjectGuid:     message.ProjectGuid,
		Type:            common.ConversationTypeAgent,
		AgentRole:       common.AgentDev.Role,
		AgentName:       common.AgentDev.Name,
		Content:         "项目自动测试已执行",
		IsMarkdown:      true,
		MarkdownContent: response.Message,
		IsExpanded:      true,
	}

	return s.commonService.CreateAndNotifyMessage(ctx, message.ProjectGuid, projectMsg)
}

// OnPackageProjectResponse 处理打包部署项目响应
func (s *projectDevService) OnPackageProjectResponse(ctx context.Context, message *agent.AgentTaskStatusMessage, response *tasks.TaskResult) error {
	projectMsg := &models.ConversationMessage{
		ProjectGuid:     message.ProjectGuid,
		Type:            common.ConversationTypeAgent,
		AgentRole:       common.AgentDev.Role,
		AgentName:       common.AgentDev.Name,
		Content:         MESSAGE_STAGE_DEPLOYED,
		IsMarkdown:      true,
		MarkdownContent: response.Message,
		IsExpanded:      true,
	}

	if err := s.commonService.EnsureProjectPrevieUrl(ctx, message.ProjectGuid); err != nil {
		logger.Error("确保项目预览URL失败", logger.String("error", err.Error()))
	}

	return s.commonService.CreateAndNotifyMessage(ctx, message.ProjectGuid, projectMsg)
}
