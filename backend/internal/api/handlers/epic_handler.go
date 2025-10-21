package handlers

import (
	"net/http"

	"github.com/lighthought/app-maker/shared-models/common"
	"github.com/lighthought/app-maker/shared-models/utils"

	"github.com/lighthought/app-maker/backend/internal/models"
	"github.com/lighthought/app-maker/backend/internal/services"

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

// UpdateEpicOrder 更新 Epic 排序
// @Summary 更新 Epic 排序
// @Description 更新指定 Epic 的显示顺序
// @Tags Epic
// @Accept json
// @Produce json
// @Param guid path string true "项目 GUID"
// @Param epicId path string true "Epic ID"
// @Param request body models.UpdateEpicOrderRequest true "排序请求"
// @Success 200 {object} common.Response
// @Failure 400 {object} common.ErrorResponse
// @Failure 500 {object} common.ErrorResponse
// @Router /api/v1/projects/{guid}/epics/{epicId}/order [put]
func (h *EpicHandler) UpdateEpicOrder(c *gin.Context) {
	projectGuid := c.Param("guid")
	epicID := c.Param("epicId")

	if projectGuid == "" || epicID == "" {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "项目GUID和Epic ID不能为空"))
		return
	}

	var req models.UpdateEpicOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "请求参数错误: "+err.Error()))
		return
	}

	err := h.epicService.UpdateEpicOrder(c.Request.Context(), epicID, req.Order)
	if err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "更新 Epic 排序失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.GetSuccessResponse("更新 Epic 排序成功", nil))
}

// UpdateEpic 更新 Epic 内容
// @Summary 更新 Epic 内容
// @Description 更新指定 Epic 的内容
// @Tags Epic
// @Accept json
// @Produce json
// @Param guid path string true "项目 GUID"
// @Param epicId path string true "Epic ID"
// @Param request body models.UpdateEpicRequest true "更新请求"
// @Success 200 {object} common.Response
// @Failure 400 {object} common.ErrorResponse
// @Failure 500 {object} common.ErrorResponse
// @Router /api/v1/projects/{guid}/epics/{epicId} [put]
func (h *EpicHandler) UpdateEpic(c *gin.Context) {
	projectGuid := c.Param("guid")
	epicID := c.Param("epicId")

	if projectGuid == "" || epicID == "" {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "项目GUID和Epic ID不能为空"))
		return
	}

	var req models.UpdateEpicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "请求参数错误: "+err.Error()))
		return
	}

	err := h.epicService.UpdateEpic(c.Request.Context(), epicID, &req)
	if err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "更新 Epic 失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.GetSuccessResponse("更新 Epic 成功", nil))
}

// DeleteEpic 删除 Epic
// @Summary 删除 Epic
// @Description 删除指定的 Epic
// @Tags Epic
// @Accept json
// @Produce json
// @Param guid path string true "项目 GUID"
// @Param epicId path string true "Epic ID"
// @Success 200 {object} common.Response
// @Failure 400 {object} common.ErrorResponse
// @Failure 500 {object} common.ErrorResponse
// @Router /api/v1/projects/{guid}/epics/{epicId} [delete]
func (h *EpicHandler) DeleteEpic(c *gin.Context) {
	projectGuid := c.Param("guid")
	epicID := c.Param("epicId")

	if projectGuid == "" || epicID == "" {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "项目GUID和Epic ID不能为空"))
		return
	}

	err := h.epicService.DeleteEpic(c.Request.Context(), epicID)
	if err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "删除 Epic 失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.GetSuccessResponse("删除 Epic 成功", nil))
}

// UpdateStoryOrder 更新 Story 排序
// @Summary 更新 Story 排序
// @Description 更新指定 Story 的显示顺序
// @Tags Epic
// @Accept json
// @Produce json
// @Param guid path string true "项目 GUID"
// @Param epicId path string true "Epic ID"
// @Param storyId path string true "Story ID"
// @Param request body models.UpdateStoryOrderRequest true "排序请求"
// @Success 200 {object} common.Response
// @Failure 400 {object} common.ErrorResponse
// @Failure 500 {object} common.ErrorResponse
// @Router /api/v1/projects/{guid}/epics/{epicId}/stories/{storyId}/order [put]
func (h *EpicHandler) UpdateStoryOrder(c *gin.Context) {
	projectGuid := c.Param("guid")
	epicID := c.Param("epicId")
	storyID := c.Param("storyId")

	if projectGuid == "" || epicID == "" || storyID == "" {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "项目GUID、Epic ID和Story ID不能为空"))
		return
	}

	var req models.UpdateStoryOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "请求参数错误: "+err.Error()))
		return
	}

	err := h.epicService.UpdateStoryOrder(c.Request.Context(), storyID, req.Order)
	if err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "更新 Story 排序失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.GetSuccessResponse("更新 Story 排序成功", nil))
}

// UpdateStory 更新 Story 内容
// @Summary 更新 Story 内容
// @Description 更新指定 Story 的内容
// @Tags Epic
// @Accept json
// @Produce json
// @Param guid path string true "项目 GUID"
// @Param epicId path string true "Epic ID"
// @Param storyId path string true "Story ID"
// @Param request body models.UpdateStoryRequest true "更新请求"
// @Success 200 {object} common.Response
// @Failure 400 {object} common.ErrorResponse
// @Failure 500 {object} common.ErrorResponse
// @Router /api/v1/projects/{guid}/epics/{epicId}/stories/{storyId} [put]
func (h *EpicHandler) UpdateStory(c *gin.Context) {
	projectGuid := c.Param("guid")
	epicID := c.Param("epicId")
	storyID := c.Param("storyId")

	if projectGuid == "" || epicID == "" || storyID == "" {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "项目GUID、Epic ID和Story ID不能为空"))
		return
	}

	var req models.UpdateStoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "请求参数错误: "+err.Error()))
		return
	}

	err := h.epicService.UpdateStory(c.Request.Context(), storyID, &req)
	if err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "更新 Story 失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.GetSuccessResponse("更新 Story 成功", nil))
}

// DeleteStory 删除 Story
// @Summary 删除 Story
// @Description 删除指定的 Story
// @Tags Epic
// @Accept json
// @Produce json
// @Param guid path string true "项目 GUID"
// @Param epicId path string true "Epic ID"
// @Param storyId path string true "Story ID"
// @Success 200 {object} common.Response
// @Failure 400 {object} common.ErrorResponse
// @Failure 500 {object} common.ErrorResponse
// @Router /api/v1/projects/{guid}/epics/{epicId}/stories/{storyId} [delete]
func (h *EpicHandler) DeleteStory(c *gin.Context) {
	projectGuid := c.Param("guid")
	epicID := c.Param("epicId")
	storyID := c.Param("storyId")

	if projectGuid == "" || epicID == "" || storyID == "" {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "项目GUID、Epic ID和Story ID不能为空"))
		return
	}

	err := h.epicService.DeleteStory(c.Request.Context(), storyID)
	if err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "删除 Story 失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.GetSuccessResponse("删除 Story 成功", nil))
}

// BatchDeleteStories 批量删除 Stories
// @Summary 批量删除 Stories
// @Description 批量删除指定的 Stories
// @Tags Epic
// @Accept json
// @Produce json
// @Param guid path string true "项目 GUID"
// @Param request body models.BatchDeleteStoriesRequest true "批量删除请求"
// @Success 200 {object} common.Response
// @Failure 400 {object} common.ErrorResponse
// @Failure 500 {object} common.ErrorResponse
// @Router /api/v1/projects/{guid}/epics/stories/batch-delete [delete]
func (h *EpicHandler) BatchDeleteStories(c *gin.Context) {
	projectGuid := c.Param("guid")

	if projectGuid == "" {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "项目GUID不能为空"))
		return
	}

	var req models.BatchDeleteStoriesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "请求参数错误: "+err.Error()))
		return
	}

	err := h.epicService.BatchDeleteStories(c.Request.Context(), req.StoryIDs)
	if err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "批量删除 Stories 失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.GetSuccessResponse("批量删除 Stories 成功", nil))
}

// ConfirmEpicsAndStories 确认 Epics 和 Stories
// @Summary 确认 Epics 和 Stories
// @Description 用户确认 Epics 和 Stories，继续执行流程
// @Tags Epic
// @Accept json
// @Produce json
// @Param guid path string true "项目 GUID"
// @Param request body models.ConfirmEpicsAndStoriesRequest true "确认请求"
// @Success 200 {object} common.Response
// @Failure 400 {object} common.ErrorResponse
// @Failure 500 {object} common.ErrorResponse
// @Router /api/v1/projects/{guid}/epics/confirm [post]
func (h *EpicHandler) ConfirmEpicsAndStories(c *gin.Context) {
	projectGuid := c.Param("guid")

	if projectGuid == "" {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "项目GUID不能为空"))
		return
	}

	var req models.ConfirmEpicsAndStoriesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "请求参数错误: "+err.Error()))
		return
	}

	err := h.epicService.ConfirmEpicsAndStories(c.Request.Context(), projectGuid, req.Action)
	if err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.ERROR_CODE, "确认 Epics 和 Stories 失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.GetSuccessResponse("确认 Epics 和 Stories 成功", nil))
}
