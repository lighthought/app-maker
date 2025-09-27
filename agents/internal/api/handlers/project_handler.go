package handlers

import (
	"net/http"

	"app-maker-agents/internal/models"
	"app-maker-agents/internal/services"
	"app-maker-agents/internal/utils"

	"github.com/gin-gonic/gin"
)

// ProjectHandler 处理项目级接口
type ProjectHandler struct {
	projectService services.ProjectService
}

// NewProjectHandler 创建 ProjectHandler
func NewProjectHandler(projectService services.ProjectService) *ProjectHandler {
	return &ProjectHandler{projectService: projectService}
}

// SetupProjectEnvironment godoc
// @Summary 项目环境准备
// @Description 为项目设置开发环境，包括安装bmad-method等工具
// @Tags Project
// @Accept json
// @Produce json
// @Param request body models.SetupProjEnvReq true "项目环境准备请求"
// @Success 200 {object} models.Response "成功响应"
// @Failure 400 {object} models.Response "参数错误"
// @Failure 500 {object} models.Response "服务器错误"
// @Router /api/v1/projects/setup [post]
func (h *ProjectHandler) SetupProjectEnvironment(c *gin.Context) {
	var req = models.SetupProjEnvReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "参数校验失败: "+err.Error())
		return
	}

	response, err := h.projectService.SetupProjectEnvironment(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusOK, models.Response{
			Code:      models.ERROR_CODE,
			Message:   err.Error(),
			Timestamp: utils.GetCurrentTime(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code:      models.SUCCESS_CODE,
		Message:   "项目环境准备成功",
		Data:      response,
		Timestamp: utils.GetCurrentTime(),
	})
}
