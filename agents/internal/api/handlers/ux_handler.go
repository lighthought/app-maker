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
}

func NewUxHandler(agentTaskService services.AgentTaskService) *UxHandler {
	return &UxHandler{agentTaskService: agentTaskService}
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

	// 根据 CLI 类型选择不同的 prompt
	var agentPrompt string
	if req.CliTool == common.CliToolGemini {
		agentPrompt = "@.bmad-core/agents/ux-expert.md"
	} else {
		agentPrompt = "@bmad/ux-expert.mdc"
	}

	message := agentPrompt + " 帮我基于PRD文档 @" + req.PrdPath +
		" 和参考页面设计(如果需求有提及的话)，输出前端的 UX Spec 到 docs/ux/ux-spec.md。" +
		"关键web页面的文生网站提示词到 docs/ux/page-prompt.md。\n我的需求是：\n" + req.Requirements +
		"\n\n注意：\n1. 始终用中文回答我，文件内容也使用中文（专有名词、代码片段和一些简单的英文除外）。\n" +
		"2. 重要: 所有生成的文件名必须使用英文命名，不要使用中文文件名。例如: 'page-prompt.md' 而不是'页面提示词.md'。\n" +
		"3. 如果 docs/ux/ 目录下已经有完善的 UX Spec 和页面提示词，直接返回概要信息，不用再尝试生成，原来的文档保持不变。"

	taskInfo, err := s.agentTaskService.EnqueueWithCli(req.ProjectGuid, common.AgentTypeUX, message, req.CliTool)
	if err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "UX标准生成任务失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.GetSuccessResponse("UX标准生成任务成功", taskInfo.ID))
}
