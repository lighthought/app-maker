package handlers

import (
	"net/http"

	"github.com/lighthought/app-maker/shared-models/agent"
	"github.com/lighthought/app-maker/shared-models/common"
	"github.com/lighthought/app-maker/shared-models/utils"

	"github.com/lighthought/app-maker/agents/internal/services"

	"github.com/gin-gonic/gin"
)

type PoHandler struct {
	agentTaskService services.AgentTaskService
	promptService    services.PromptService
}

func NewPoHandler(agentTaskService services.AgentTaskService, promptService services.PromptService) *PoHandler {
	return &PoHandler{
		agentTaskService: agentTaskService,
		promptService:    promptService,
	}
}

func (h *PoHandler) getAgentPrompt(cliTool string) string {
	if cliTool == common.CliToolGemini {
		return "@.bmad-core/agents/po.md"
	}
	return "@bmad/po.mdc"
}

// GetEpicsAndStories godoc
// @Summary 获取史诗和用户故事
// @Description 基于PRD和架构设计生成Epics和Stories文档
// @Tags PO
// @Accept json
// @Produce json
// @Param request body agent.GetEpicsAndStoriesReq true "史诗故事请求"
// @Success 200 {object} common.Response "成功响应"
// @Failure 400 {object} common.ErrorResponse "参数错误"
// @Failure 500 {object} common.ErrorResponse "服务器错误"
// @Router /api/v1/agent/po/epicsandstories [get]
func (s *PoHandler) GetEpicsAndStories(c *gin.Context) {
	var req agent.GetEpicsAndStoriesReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "参数校验失败: "+err.Error()))
		return
	}

	agentPrompt := s.getAgentPrompt(req.CliTool)

	data := map[string]interface{}{
		"AgentPrompt": agentPrompt,
		"PrdPath":     req.PrdPath,
		"ArchFolder":  req.ArchFolder,
	}

	message, err := s.promptService.GetPrompt("po", "get_epics_and_stories", data)
	if err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "生成Prompt失败: "+err.Error()))
		return
	}

	taskInfo, err := s.agentTaskService.EnqueueWithCli(req.ProjectGuid, common.AgentTypePO, message,
		req.CliTool, common.DevStatusPlanEpicAndStory)
	if err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "获取史诗和用户故事任务失败: "+err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.GetSuccessResponse("获取史诗和用户故事成功", taskInfo.ID))
}
