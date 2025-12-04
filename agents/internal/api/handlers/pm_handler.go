package handlers

import (
	"net/http"

	"github.com/lighthought/app-maker/shared-models/agent"
	"github.com/lighthought/app-maker/shared-models/common"
	"github.com/lighthought/app-maker/shared-models/utils"

	"github.com/lighthought/app-maker/agents/internal/services"

	"github.com/gin-gonic/gin"
)

// PmHandler 负责产品经理 Agent 的接口
type PmHandler struct {
	agentTaskService services.AgentTaskService
	promptService    services.PromptService
}

// NewPmHandler 创建新的 PM Handler
func NewPmHandler(agentTaskService services.AgentTaskService, promptService services.PromptService) *PmHandler {
	return &PmHandler{
		agentTaskService: agentTaskService,
		promptService:    promptService,
	}
}

func (h *PmHandler) getAgentPrompt(cliTool string) string {
	if cliTool == common.CliToolGemini {
		return "@.bmad-core/agents/pm.md"
	}
	return "@bmad/pm.mdc"
}

// GetPRD godoc
// @Summary 获取产品需求文档
// @Description 根据需求生成PRD文档
// @Tags PM
// @Accept json
// @Produce json
// @Param request body agent.GetPRDReq true "PRD请求"
// @Success 200 {object} common.Response "成功响应"
// @Failure 400 {object} common.ErrorResponse "参数错误"
// @Failure 500 {object} common.ErrorResponse "服务器错误"
// @Router /api/v1/agent/pm/prd [get]
func (s *PmHandler) GetPRD(c *gin.Context) {
	var req agent.GetPRDReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "参数校验失败: "+err.Error()))
		return
	}

	agentPrompt := s.getAgentPrompt(req.CliTool)

	data := map[string]interface{}{
		"AgentPrompt":  agentPrompt,
		"Requirements": req.Requirements,
	}

	message, err := s.promptService.GetPrompt("pm", "get_prd", data)
	if err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "生成Prompt失败: "+err.Error()))
		return
	}

	taskInfo, err := s.agentTaskService.EnqueueWithCli(req.ProjectGuid, common.AgentTypePM, message,
		req.CliTool, common.DevStatusGeneratePRD)
	if err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "PRD 生成失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.GetSuccessResponse("PRD 生成成功", taskInfo.ID))
}
