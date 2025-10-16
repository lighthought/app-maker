package handlers

import (
	"net/http"

	"autocodeweb-backend/internal/services"
	"shared-models/common"
	"shared-models/utils"

	"github.com/gin-gonic/gin"
)

type EpicHandler struct {
	epicService services.EpicService
}

func NewEpicHandler(epicService services.EpicService) *EpicHandler {
	return &EpicHandler{epicService: epicService}
}

// GetProjectEpics 获取项目的所有 Epics 和 Stories
// @Summary 获取项目的 Epics
// @Description 根据项目 GUID 获取所有 Epics 和 Stories
// @Tags Epic
// @Accept json
// @Produce json
// @Param guid path string true "项目 GUID"
// @Success 200 {object} common.Response
// @Failure 400 {object} common.ErrorResponse
// @Failure 500 {object} common.ErrorResponse
// @Router /api/v1/projects/{guid}/epics [get]
func (h *EpicHandler) GetProjectEpics(c *gin.Context) {
	projectGuid := c.Param("guid")
	if projectGuid == "" {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "项目GUID不能为空"))
		return
	}

	epics, err := h.epicService.GetByProjectGuid(c.Request.Context(), projectGuid)
	if err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "获取 Epics 失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.GetSuccessResponse("获取成功", epics))
}

// GetProjectMvpEpics 获取项目的 MVP Epics
// @Summary 获取项目的 MVP Epics
// @Description 根据项目 GUID 获取 MVP 阶段的 Epics 和 Stories (P0 优先级)
// @Tags Epic
// @Accept json
// @Produce json
// @Param guid path string true "项目 GUID"
// @Success 200 {object} common.Response
// @Failure 400 {object} common.ErrorResponse
// @Failure 500 {object} common.ErrorResponse
// @Router /api/v1/projects/{guid}/mvp-epics [get]
func (h *EpicHandler) GetProjectMvpEpics(c *gin.Context) {
	projectGuid := c.Param("guid")
	if projectGuid == "" {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "项目GUID不能为空"))
		return
	}

	epics, err := h.epicService.GetMvpEpicsByProjectGuid(c.Request.Context(), projectGuid)
	if err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "获取 MVP Epics 失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.GetSuccessResponse("获取成功", epics))
}
