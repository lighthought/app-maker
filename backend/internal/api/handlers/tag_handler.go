package handlers

import (
	"net/http"
	"strconv"
	"time"

	"autocodeweb-backend/internal/models"
	"autocodeweb-backend/internal/services"

	"github.com/gin-gonic/gin"
)

// TagHandler 标签处理器
type TagHandler struct {
	tagService services.TagService
}

// NewTagHandler 创建标签处理器实例
func NewTagHandler(tagService services.TagService) *TagHandler {
	return &TagHandler{
		tagService: tagService,
	}
}

// CreateTag godoc
// @Summary 创建标签
// @Description 创建新标签
// @Tags 标签管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer 用户令牌"
// @Param tag body models.CreateTagRequest true "标签创建请求"
// @Success 200 {object} models.Response{data=models.TagInfo} "标签创建成功"
// @Failure 400 {object} models.ErrorResponse "请求参数错误"
// @Failure 401 {object} models.ErrorResponse "未授权"
// @Failure 409 {object} models.ErrorResponse "标签名称已存在"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /api/v1/tags [post]
func (h *TagHandler) CreateTag(c *gin.Context) {
	var req models.CreateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:      http.StatusBadRequest,
			Message:   "请求参数错误",
			Timestamp: time.Now().Format(time.RFC3339),
		})
		return
	}

	tag, err := h.tagService.CreateTag(c.Request.Context(), &req)
	if err != nil {
		if err.Error() == "tag name already exists" {
			c.JSON(http.StatusConflict, models.ErrorResponse{
				Code:      http.StatusConflict,
				Message:   "标签名称已存在",
				Timestamp: time.Now().Format(time.RFC3339),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:      http.StatusInternalServerError,
			Message:   "创建标签失败: " + err.Error(),
			Timestamp: time.Now().Format(time.RFC3339),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code:      0,
		Message:   "标签创建成功",
		Data:      tag,
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// GetTag godoc
// @Summary 获取标签信息
// @Description 根据标签ID获取标签详细信息
// @Tags 标签管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer 用户令牌"
// @Param id path string true "标签ID"
// @Success 200 {object} models.Response{data=models.TagInfo} "获取标签成功"
// @Failure 400 {object} models.ErrorResponse "请求参数错误"
// @Failure 401 {object} models.ErrorResponse "未授权"
// @Failure 404 {object} models.ErrorResponse "标签不存在"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /api/v1/tags/{id} [get]
func (h *TagHandler) GetTag(c *gin.Context) {
	tagID := c.Param("id")
	if tagID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:      http.StatusBadRequest,
			Message:   "标签ID不能为空",
			Timestamp: time.Now().Format(time.RFC3339),
		})
		return
	}

	tag, err := h.tagService.GetTag(c.Request.Context(), tagID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:      http.StatusInternalServerError,
			Message:   "获取标签失败: " + err.Error(),
			Timestamp: time.Now().Format(time.RFC3339),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code:      0,
		Message:   "获取标签成功",
		Data:      tag,
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// UpdateTag godoc
// @Summary 更新标签
// @Description 更新标签信息
// @Tags 标签管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer 用户令牌"
// @Param id path string true "标签ID"
// @Param tag body models.UpdateTagRequest true "标签更新请求"
// @Success 200 {object} models.Response{data=models.TagInfo} "标签更新成功"
// @Failure 400 {object} models.ErrorResponse "请求参数错误"
// @Failure 401 {object} models.ErrorResponse "未授权"
// @Failure 404 {object} models.ErrorResponse "标签不存在"
// @Failure 409 {object} models.ErrorResponse "标签名称已存在"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /api/v1/tags/{id} [put]
func (h *TagHandler) UpdateTag(c *gin.Context) {
	tagID := c.Param("id")
	if tagID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:      http.StatusBadRequest,
			Message:   "标签ID不能为空",
			Timestamp: time.Now().Format(time.RFC3339),
		})
		return
	}

	var req models.UpdateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:      http.StatusBadRequest,
			Message:   "请求参数错误",
			Timestamp: time.Now().Format(time.RFC3339),
		})
		return
	}

	tag, err := h.tagService.UpdateTag(c.Request.Context(), tagID, &req)
	if err != nil {
		if err.Error() == "tag name already exists" {
			c.JSON(http.StatusConflict, models.ErrorResponse{
				Code:      http.StatusConflict,
				Message:   "标签名称已存在",
				Timestamp: time.Now().Format(time.RFC3339),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:      http.StatusInternalServerError,
			Message:   "更新标签失败: " + err.Error(),
			Timestamp: time.Now().Format(time.RFC3339),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code:      0,
		Message:   "标签更新成功",
		Data:      tag,
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// DeleteTag godoc
// @Summary 删除标签
// @Description 删除指定标签
// @Tags 标签管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer 用户令牌"
// @Param id path string true "标签ID"
// @Success 200 {object} models.Response "标签删除成功"
// @Failure 400 {object} models.ErrorResponse "请求参数错误"
// @Failure 401 {object} models.ErrorResponse "未授权"
// @Failure 409 {object} models.ErrorResponse "标签正在被使用"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /api/v1/tags/{id} [delete]
func (h *TagHandler) DeleteTag(c *gin.Context) {
	tagID := c.Param("id")
	if tagID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:      http.StatusBadRequest,
			Message:   "标签ID不能为空",
			Timestamp: time.Now().Format(time.RFC3339),
		})
		return
	}

	err := h.tagService.DeleteTag(c.Request.Context(), tagID)
	if err != nil {
		if err.Error() == "cannot delete tag that is used by projects" {
			c.JSON(http.StatusConflict, models.ErrorResponse{
				Code:      http.StatusConflict,
				Message:   "标签正在被项目使用，无法删除",
				Timestamp: time.Now().Format(time.RFC3339),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:      http.StatusInternalServerError,
			Message:   "删除标签失败: " + err.Error(),
			Timestamp: time.Now().Format(time.RFC3339),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code:      0,
		Message:   "标签删除成功",
		Data:      nil,
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// ListTags godoc
// @Summary 获取标签列表
// @Description 获取所有标签列表
// @Tags 标签管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer 用户令牌"
// @Success 200 {object} models.Response{data=[]models.TagInfo} "获取标签列表成功"
// @Failure 401 {object} models.ErrorResponse "未授权"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /api/v1/tags [get]
func (h *TagHandler) ListTags(c *gin.Context) {
	tags, err := h.tagService.ListTags(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:      http.StatusInternalServerError,
			Message:   "获取标签列表失败: " + err.Error(),
			Timestamp: time.Now().Format(time.RFC3339),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code:      0,
		Message:   "获取标签列表成功",
		Data:      tags,
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// GetPopularTags godoc
// @Summary 获取热门标签
// @Description 获取使用频率最高的标签
// @Tags 标签管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer 用户令牌"
// @Param limit query int false "标签数量限制" default(10)
// @Success 200 {object} models.Response{data=[]models.TagInfo} "获取热门标签成功"
// @Failure 401 {object} models.ErrorResponse "未授权"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /api/v1/tags/popular [get]
func (h *TagHandler) GetPopularTags(c *gin.Context) {
	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	tags, err := h.tagService.GetPopularTags(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:      http.StatusInternalServerError,
			Message:   "获取热门标签失败: " + err.Error(),
			Timestamp: time.Now().Format(time.RFC3339),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code:      0,
		Message:   "获取热门标签成功",
		Data:      tags,
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// GetTagsByProject godoc
// @Summary 获取项目标签
// @Description 获取指定项目的所有标签
// @Tags 标签管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer 用户令牌"
// @Param project_id query string true "项目ID"
// @Success 200 {object} models.Response{data=[]models.TagInfo} "获取项目标签成功"
// @Failure 400 {object} models.ErrorResponse "请求参数错误"
// @Failure 401 {object} models.ErrorResponse "未授权"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /api/v1/tags/project [get]
func (h *TagHandler) GetTagsByProject(c *gin.Context) {
	projectID := c.Query("project_id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:      http.StatusBadRequest,
			Message:   "项目ID不能为空",
			Timestamp: time.Now().Format(time.RFC3339),
		})
		return
	}

	tags, err := h.tagService.GetTagsByProject(c.Request.Context(), projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:      http.StatusInternalServerError,
			Message:   "获取项目标签失败: " + err.Error(),
			Timestamp: time.Now().Format(time.RFC3339),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code:      0,
		Message:   "获取项目标签成功",
		Data:      tags,
		Timestamp: time.Now().Format(time.RFC3339),
	})
}
