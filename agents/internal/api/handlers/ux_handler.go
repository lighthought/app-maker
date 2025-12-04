package handlers

import (
	"net/http"

	"github.com/lighthought/app-maker/shared-models/agent"
	"github.com/lighthought/app-maker/shared-models/common"
	"github.com/lighthought/app-maker/shared-models/utils"

	"github.com/lighthought/app-maker/agents/internal/services"

	"github.com/gin-gonic/gin"
)

type UxHandler struct {
	agentTaskService services.AgentTaskService
	promptService    services.PromptService
}

func NewUxHandler(agentTaskService services.AgentTaskService, promptService services.PromptService) *UxHandler {
	return &UxHandler{
		agentTaskService: agentTaskService,
		promptService:    promptService,
	}
}

func (h *UxHandler) getAgentPrompt(cliTool string) string {
	if cliTool == common.CliToolGemini {
		return "@.bmad-core/agents/ux-expert.md"
	}
	return "@bmad/ux-expert.mdc"
}

// GetUXStandard godoc
// @Summary 获取UX设计标准
// @Description 基于PRD生成UX设计规范和页面提示词
// @Tags UX
// @Accept json
// @Produce json
// @Param request body agent.GetUXStandardReq true "UX标准请求"
// @Success 200 {object} common.Response "成功响应"
// @Failure 400 {object} common.ErrorResponse "参数错误"
// @Failure 500 {object} common.ErrorResponse "服务器错误"
// @Router /api/v1/agent/ux-expert/ux-standard [get]
func (s *UxHandler) GetUXStandard(c *gin.Context) {
	var req agent.GetUXStandardReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "参数校验失败: "+err.Error()))
		return
	}

	agentPrompt := s.getAgentPrompt(req.CliTool)

	data := map[string]interface{}{
		"AgentPrompt":  agentPrompt,
		"PrdPath":      req.PrdPath,
		"Requirements": req.Requirements,
	}

	message, err := s.promptService.GetPrompt("ux", "get_ux_standard", data)
	if err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "生成Prompt失败: "+err.Error()))
		return
	}

	taskInfo, err := s.agentTaskService.EnqueueWithCli(req.ProjectGuid, common.AgentTypeUX, message,
		req.CliTool, common.DevStatusDefineUXStandard)
	if err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "UX标准生成任务失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.GetSuccessResponse("UX标准生成任务成功", taskInfo.ID))
}
