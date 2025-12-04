package handlers

import (
	"net/http"

	"github.com/lighthought/app-maker/shared-models/agent"
	"github.com/lighthought/app-maker/shared-models/common"
	"github.com/lighthought/app-maker/shared-models/utils"

	"github.com/lighthought/app-maker/agents/internal/services"

	"github.com/gin-gonic/gin"
)

// AnalyseHandler 负责分析 Agent 的接口
type AnalyseHandler struct {
	agentTaskService services.AgentTaskService
	promptService    services.PromptService
}

// NewAnalyseHandler 创建新的分析 Handler
func NewAnalyseHandler(agentTaskService services.AgentTaskService, promptService services.PromptService) *AnalyseHandler {
	return &AnalyseHandler{
		agentTaskService: agentTaskService,
		promptService:    promptService,
	}
}

func (h *AnalyseHandler) getAgentPrompt(cliTool string) string {
	if cliTool == common.CliToolGemini {
		return "@.bmad-core/agents/analyst.md"
	}
	return "@bmad/analyst.mdc"
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

	agentPrompt := s.getAgentPrompt(req.CliTool)

	data := map[string]interface{}{
		"AgentPrompt":  agentPrompt,
		"Requirements": req.Requirements,
	}

	message, err := s.promptService.GetPrompt("analyst", "project_brief", data)
	if err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "生成Prompt失败: "+err.Error()))
		return
	}

	taskInfo, err := s.agentTaskService.EnqueueWithCli(req.ProjectGuid, common.AgentTypeAnalyse, message,
		req.CliTool, common.DevStatusCheckRequirement)
	if err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "异步任务压入失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.GetSuccessResponse("项目概览任务创建成功", taskInfo.ID))
}
