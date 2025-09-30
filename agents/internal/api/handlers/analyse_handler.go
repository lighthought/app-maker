package handlers

import (
	"app-maker-agents/internal/services"
	"net/http"
	"shared-models/agent"
	"shared-models/common"
	"shared-models/utils"

	"github.com/gin-gonic/gin"
)

// AnalyseHandler 负责分析 Agent 的接口
type AnalyseHandler struct {
	commandService *services.CommandService
}

// NewAnalyseHandler 创建新的分析 Handler
func NewAnalyseHandler(commandService *services.CommandService) *AnalyseHandler {
	return &AnalyseHandler{commandService: commandService}
}

// ProjectBrief godoc
// @Summary 生成项目概览
// @Description 根据需求生成项目简介和市场研究文档
// @Tags Analyse
// @Accept json
// @Produce json
// @Param request body agent.GetProjBriefReq true "项目概览请求"
// @Success 200 {object} common.Response "成功响应"
// @Failure 400 {object} common.ErrorResponse "参数错误"
// @Failure 500 {object} common.ErrorResponse "服务器错误"
// @Router /api/v1/agent/analyse/project-brief [post]
func (s *AnalyseHandler) ProjectBrief(c *gin.Context) {
	var req agent.GetProjBriefReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "参数校验失败: "+err.Error()))
		return
	}

	message := "@bmad/analyst.mdc 请你为我生成项目简介，再执行市场研究。输出对应的文档到 docs/analyse/ 目录下。我的需求是：\n" + req.Requirements
	result := s.commandService.SimpleExecute(c.Request.Context(), req.ProjectGuid, message)
	if !result.Success {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "分析任务失败: "+result.Error))
		return
	}
	// TODO: 检查实际输出的文档，组装成结果，返回给 backend
	agentResult := agent.AgentResult{
		Output: result.Output,
		Error:  result.Error,
	}

	c.JSON(http.StatusOK, utils.GetSuccessResponse("分析任务成功", agentResult))
}
