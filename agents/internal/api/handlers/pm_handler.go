package handlers

import (
	"net/http"
	"shared-models/agent"
	"time"

	"app-maker-agents/internal/services"
	"app-maker-agents/internal/utils"

	"github.com/gin-gonic/gin"
)

// PmHandler 负责产品经理 Agent 的接口
type PmHandler struct {
	commandService *services.CommandService
}

// NewPmHandler 创建新的 PM Handler
func NewPmHandler(commandService *services.CommandService) *PmHandler {
	return &PmHandler{commandService: commandService}
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
		utils.Error(c, http.StatusBadRequest, "参数校验失败: "+err.Error())
		return
	}

	message := "@bmad/pm.mdc 我希望你根据我的需求帮我输出 PRD 文档到 docs/PRD.md。\n" +
		"简化部署和运维、商业模式、成功指标、风险评估中的市场和运营风险。\n" +
		"技术选型我后续再和架构师深入讨论，主题颜色我后续再和 ux 专家讨论。\n" +
		"我的需求是：" + req.Requirements

	result := s.commandService.Execute(c.Request.Context(), req.ProjectGuid, message, 5*time.Minute)
	if !result.Success {
		utils.Error(c, http.StatusInternalServerError, "分析任务失败: "+result.Error)
		return
	}
	// TODO: 检查实际输出的文档，组装成结果，返回给 backend

	utils.Success(c, result.Output)
}
