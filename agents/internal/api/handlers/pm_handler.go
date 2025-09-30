package handlers

import (
	"net/http"
	"shared-models/agent"
	"shared-models/common"
	"shared-models/utils"

	"app-maker-agents/internal/services"

	"github.com/gin-gonic/gin"
)

// PmHandler 负责产品经理 Agent 的接口
type PmHandler struct {
	agentTaskService services.AgentTaskService
}

// NewPmHandler 创建新的 PM Handler
func NewPmHandler(agentTaskService services.AgentTaskService) *PmHandler {
	return &PmHandler{agentTaskService: agentTaskService}
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

	message := "@bmad/pm.mdc 我希望你根据我的需求帮我输出 PRD 文档到 docs/PRD.md。\n" +
		"简化部署和运维、商业模式、成功指标、风险评估中的市场和运营风险。\n" +
		"技术选型我后续再和架构师深入讨论，主题颜色我后续再和 ux 专家讨论。\n" +
		"我的需求是：" + req.Requirements

	taskInfo, err := s.agentTaskService.Enqueue(req.ProjectGuid, common.AgentTypePM, message)
	if err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "PRD 生成失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.GetSuccessResponse("PRD 生成成功", taskInfo.ID))
}
