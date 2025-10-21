package handlers

import (
	"net/http"

	"github.com/lighthought/app-maker/shared-models/agent"
	"github.com/lighthought/app-maker/shared-models/common"
	"github.com/lighthought/app-maker/shared-models/utils"

	"github.com/lighthought/app-maker/agents/internal/services"

	"github.com/gin-gonic/gin"
)

// ProjectHandler 处理项目级接口
type ProjectHandler struct {
	agentTaskService services.AgentTaskService
	projectService   services.ProjectService
}

// NewProjectHandler 创建 ProjectHandler
func NewProjectHandler(agentTaskService services.AgentTaskService, projectService services.ProjectService) *ProjectHandler {
	return &ProjectHandler{agentTaskService: agentTaskService, projectService: projectService}
}

// SetupProjectEnvironment godoc
// @Summary 项目环境准备
// @Description 为项目设置开发环境，包括安装bmad-method等工具
// @Tags Project
// @Accept json
// @Produce json
// @Param request body agent.SetupProjEnvReq true "项目环境准备请求"
// @Success 200 {object} common.Response "成功响应"
// @Failure 400 {object} common.Response "参数错误"
// @Failure 500 {object} common.Response "服务器错误"
// @Router /api/v1/projects/setup [post]
func (h *ProjectHandler) SetupProjectEnvironment(c *gin.Context) {
	var req = agent.SetupProjEnvReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "参数校验失败: "+err.Error()))
		return
	}

	taskInfo, err := h.agentTaskService.EnqueueSetupReq(&req)
	if err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.GetSuccessResponse("项目环境准备成功", taskInfo.ID))
}
