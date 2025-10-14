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

	// 根据 CLI 类型选择不同的 prompt
	var agentPrompt string
	if req.CliTool == common.CliToolGemini {
		agentPrompt = "@.bmad-core/agents/pm.md"
	} else {
		agentPrompt = "@bmad/pm.mdc"
	}

	message := agentPrompt + " 我希望你根据 @docs/analyse目录下的项目简介和市场研究，以及我的需求帮我输出 PRD.md 文档到 docs 目录下，用 UTF-8 格式编码。\n" +
		"我的需求是：" + req.Requirements +
		"注意：1. 始终用中文回答我，文件内容也使用中文（专有名词、代码片段和一些简单的英文除外）。" +
		"2. 简化部署和运维、商业模式、成功指标、风险评估中的市场和运营风险。\n" +
		"3. 技术选型我后续再和架构师深入讨论，主题颜色我后续再和 ux 专家讨论，不需要你在 PRD 中体现。\n" +
		"4. 不需要你做额外的调查，也不要问我要不要创建文件，直接输出PRD到 docs/PRD.md 文件中。\n" +
		"5. 如果 docs/ 目录下已经有完善的 PRD.md 文件，直接返回概要信息，不用再尝试生成 PRD.md，原来的文档保持不变。"

	taskInfo, err := s.agentTaskService.EnqueueWithCli(req.ProjectGuid, common.AgentTypePM, message, req.CliTool)
	if err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "PRD 生成失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.GetSuccessResponse("PRD 生成成功", taskInfo.ID))
}
