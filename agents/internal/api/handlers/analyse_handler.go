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
}

// NewAnalyseHandler 创建新的分析 Handler
func NewAnalyseHandler(agentTaskService services.AgentTaskService) *AnalyseHandler {
	return &AnalyseHandler{agentTaskService: agentTaskService}
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

	// 根据 CLI 类型选择不同的 prompt
	var agentPrompt string
	if req.CliTool == common.CliToolGemini {
		agentPrompt = "@.bmad-core/agents/analyst.md"
	} else {
		agentPrompt = "@bmad/analyst.mdc"
	}

	message := agentPrompt + " 请你为我生成项目简介，再执行市场研究。输出对应的文档到 docs/analyse/ 目录下。\n" +
		"我的需求是：\n" + req.Requirements +
		"\n\n注意：1.始终用中文回答我，文件内容也使用中文（专有名词、代码片段和一些简单的英文除外）。\n" +
		"2. 如果 docs/analyse/ 目录下已经有完善的项目简介和市场研究文档，直接返回概要信息，不用再尝试各种研究和调查过程，原来的文档保持不变。\n" +
		"3. 不需要你关心技术方向，这个我后续会和架构师深入讨论。\n" +
		"4. 不要问我任何问题，请基于我的需求判断我想要开发的应用或网站类型。\n" +
		"5. 市场研究文档的内容包括：竞争对手分析、目标市场规模、用户需求分析、商业模式可行性。"

	taskInfo, err := s.agentTaskService.EnqueueWithCli(req.ProjectGuid, common.AgentTypeAnalyse, message,
		req.CliTool, common.DevStatusCheckRequirement)
	if err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "异步任务压入失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.GetSuccessResponse("项目概览任务创建成功", taskInfo.ID))
}
