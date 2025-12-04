package handlers

import (
	"net/http"

	"github.com/lighthought/app-maker/shared-models/agent"
	"github.com/lighthought/app-maker/shared-models/common"
	"github.com/lighthought/app-maker/shared-models/utils"

	"github.com/lighthought/app-maker/agents/internal/services"

	"github.com/gin-gonic/gin"
)

type DevHandler struct {
	agentTaskService services.AgentTaskService
	commandService   services.CommandService
	promptService    services.PromptService
}

func NewDevHandler(agentTaskService services.AgentTaskService, commandService services.CommandService, promptService services.PromptService) *DevHandler {
	return &DevHandler{
		agentTaskService: agentTaskService,
		commandService:   commandService,
		promptService:    promptService,
	}
}

func (h *DevHandler) getAgentPrompt(cliTool string) string {
	if cliTool == common.CliToolGemini {
		return "@.bmad-core/agents/dev.md"
	}
	return "@bmad/dev.mdc"
}

// ImplementStory godoc
// @Summary 实现用户故事
// @Description 基于PRD、架构设计和UX标准实现用户故事
// @Tags Dev
// @Accept json
// @Produce json
// @Param request body agent.ImplementStoryReq true "实现故事请求"
// @Success 200 {object} common.Response "成功响应"
// @Failure 400 {object} common.ErrorResponse "参数错误"
// @Failure 500 {object} common.ErrorResponse "服务器错误"
// @Router /api/v1/agent/dev/implstory [post]
func (h *DevHandler) ImplementStory(c *gin.Context) {
	var req agent.ImplementStoryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "参数校验失败: "+err.Error()))
		return
	}

	agentPrompt := h.getAgentPrompt(req.CliTool)

	data := map[string]interface{}{
		"AgentPrompt": agentPrompt,
		"PrdPath":     req.PrdPath,
		"ArchFolder":  req.ArchFolder,
		"UxSpecPath":  req.UxSpecPath,
		"EpicFile":    req.EpicFile,
		"StoryFile":   req.StoryFile,
		"DbFolder":    req.DbFolder,
		"ApiFolder":   req.ApiFolder,
	}

	message, err := h.promptService.GetPrompt("dev", "implement_story", data)
	if err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "生成Prompt失败: "+err.Error()))
		return
	}

	taskInfo, err := h.agentTaskService.EnqueueWithCli(req.ProjectGuid, common.AgentTypeDev, message,
		req.CliTool, common.DevStatusDevelopStory)
	if err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "实现用户故事任务失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.GetSuccessResponse("实现用户故事任务创建成功", taskInfo.ID))
}

// FixBug godoc
// @Summary 修复Bug
// @Description 根据Bug描述修复项目中的问题
// @Tags Dev
// @Accept json
// @Produce json
// @Param request body agent.FixBugReq true "修复Bug请求"
// @Success 200 {object} common.Response "成功响应"
// @Failure 400 {object} common.ErrorResponse "参数错误"
// @Failure 500 {object} common.ErrorResponse "服务器错误"
// @Router /api/v1/agent/dev/fixbug [post]
func (h *DevHandler) FixBug(c *gin.Context) {
	var req agent.FixBugReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "参数校验失败: "+err.Error()))
		return
	}

	agentPrompt := h.getAgentPrompt(req.CliTool)

	data := map[string]interface{}{
		"AgentPrompt":    agentPrompt,
		"BugDescription": req.BugDescription,
	}

	message, err := h.promptService.GetPrompt("dev", "fix_bug", data)
	if err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "生成Prompt失败: "+err.Error()))
		return
	}

	taskInfo, err := h.agentTaskService.EnqueueWithCli(req.ProjectGuid, common.AgentTypeDev, message,
		req.CliTool, common.DevStatusFixBug)
	if err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "修复Bug任务失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.GetSuccessResponse("修复Bug任务创建成功", taskInfo.ID))
}

// RunTest godoc
// @Summary 运行测试
// @Description 执行项目的自动测试流程
// @Tags Dev
// @Accept json
// @Produce json
// @Param request body agent.FixBugReq true "运行测试请求"
// @Success 200 {object} common.Response "成功响应"
// @Failure 400 {object} common.ErrorResponse "参数错误"
// @Failure 500 {object} common.ErrorResponse "服务器错误"
// @Router /api/v1/agent/dev/runtest [post]
func (h *DevHandler) RunTest(c *gin.Context) {
	var req agent.RunTestReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "参数校验失败: "+err.Error()))
		return
	}

	agentPrompt := h.getAgentPrompt(req.CliTool)

	data := map[string]interface{}{
		"AgentPrompt": agentPrompt,
	}

	message, err := h.promptService.GetPrompt("dev", "run_test", data)
	if err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "生成Prompt失败: "+err.Error()))
		return
	}

	taskInfo, err := h.agentTaskService.EnqueueWithCli(req.ProjectGuid, common.AgentTypeDev, message,
		req.CliTool, common.DevStatusRunTest)
	if err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "运行测试任务失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.GetSuccessResponse("测试任务创建成功", taskInfo.ID))
}

// ImplementFrontend godoc
// @Summary 实现前端
// @Description 生成前端关键页面
// @Tags Dev
// @Accept json
// @Produce json
// @Param request body agent.ImplementFrontendReq true "实现前端请求"
// @Success 200 {object} common.Response "成功响应"
// @Failure 400 {object} common.ErrorResponse "参数错误"
// @Failure 500 {object} common.ErrorResponse "服务器错误"
// @Router /api/v1/agent/dev/frontend [post]
func (h *DevHandler) ImplementFrontend(c *gin.Context) {
	var req agent.ImplementFrontendReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "参数校验失败: "+err.Error()))
		return
	}

	agentPrompt := h.getAgentPrompt(req.CliTool)

	data := map[string]interface{}{
		"AgentPrompt": agentPrompt,
	}

	message, err := h.promptService.GetPrompt("dev", "implement_frontend", data)
	if err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "生成Prompt失败: "+err.Error()))
		return
	}

	taskInfo, err := h.agentTaskService.EnqueueWithCli(req.ProjectGuid, common.AgentTypeDev, message,
		req.CliTool, common.DevStatusGeneratePages)
	if err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "前端生成任务失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.GetSuccessResponse("前端生成任务创建成功", taskInfo.ID))
}

// Deploy godoc
// @Summary 部署项目
// @Description 执行项目的打包部署流程
// @Tags Dev
// @Accept json
// @Produce json
// @Param request body agent.FixBugReq true "部署请求"
// @Success 200 {object} common.Response "成功响应"
// @Failure 400 {object} common.ErrorResponse "参数错误"
// @Failure 500 {object} common.ErrorResponse "服务器错误"
// @Router /api/v1/agent/dev/deploy [post]
func (h *DevHandler) Deploy(c *gin.Context) {
	var req agent.DeployReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "参数校验失败: "+err.Error()))
		return
	}

	// 异步执行部署任务
	taskInfo, err := h.agentTaskService.EnqueueDeployReq(&req)
	if err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "部署任务失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.GetSuccessResponse("部署任务已启动", taskInfo.ID))

}
