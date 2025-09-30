package handlers

import (
	"app-maker-agents/internal/services"
	"net/http"
	"shared-models/agent"
	"shared-models/common"
	"shared-models/utils"

	"github.com/gin-gonic/gin"
)

type UxHandler struct {
	commandService *services.CommandService
}

func NewUxHandler(commandService *services.CommandService) *UxHandler {
	return &UxHandler{commandService: commandService}
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

	message := "@bmad/ux-expert.mdc 帮我基于PRD文档 @" + req.PrdPath + " 和参考页面设计(如果需求有提及的话)，输出前端的 UX Spec 到 docs/ux/ux-spec.md。" +
		"关键web页面的文生网站提示词到 docs/ux/page-prompt.md。我的需求是：" + req.Requirements

	result := s.commandService.SimpleExecute(c.Request.Context(), req.ProjectGuid, message)
	if !result.Success {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "UX标准生成任务失败: "+result.Error))
		return
	}
	// TODO: 检查实际输出的文档，组装成结果，返回给 backend

	agentResult := agent.AgentResult{
		Output: result.Output,
		Error:  result.Error,
	}

	c.JSON(http.StatusOK, utils.GetSuccessResponse("UX标准生成任务成功", agentResult))
}
