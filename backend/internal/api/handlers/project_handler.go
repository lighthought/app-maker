package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"autocodeweb-backend/internal/models"
	"autocodeweb-backend/internal/services"
	"autocodeweb-backend/pkg/logger"

	"github.com/gin-gonic/gin"
)

// ProjectHandler 项目处理器
type ProjectHandler struct {
	projectService services.ProjectService
	tagService     services.TagService
}

// NewProjectHandler 创建项目处理器实例
func NewProjectHandler(projectService services.ProjectService, tagService services.TagService) *ProjectHandler {
	return &ProjectHandler{
		projectService: projectService,
		tagService:     tagService,
	}
}

// CreateProject godoc
// @Summary 创建项目
// @Description 创建新项目
// @Tags 项目管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer 用户令牌"
// @Param project body models.CreateProjectRequest true "项目创建请求"
// @Success 200 {object} models.Response{data=models.ProjectInfo} "项目创建成功"
// @Failure 400 {object} models.ErrorResponse "请求参数错误"
// @Failure 401 {object} models.ErrorResponse "未授权"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /api/v1/projects [post]
func (h *ProjectHandler) CreateProject(c *gin.Context) {
	logger.Info("收到创建项目请求",
		logger.String("userAgent", c.GetHeader("User-Agent")),
		logger.String("remoteAddr", c.ClientIP()),
	)

	var req models.CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("请求参数绑定失败",
			logger.String("error", err.Error()),
			logger.String("requestBody", fmt.Sprintf("%v", c.Request.Body)),
		)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:      http.StatusBadRequest,
			Message:   "请求参数错误",
			Timestamp: time.Now().Format(time.RFC3339),
		})
		return
	}

	logger.Info("请求参数验证通过",
		logger.String("requirements", req.Requirements),
	)

	// 从中间件获取用户ID
	userID := c.GetString("user_id")
	logger.Info("获取用户ID", logger.String("userID", userID))

	project, err := h.projectService.CreateProject(c.Request.Context(), &req, userID)
	if err != nil {
		logger.Error("创建项目失败",
			logger.String("error", err.Error()),
			logger.String("userID", userID),
		)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:      http.StatusInternalServerError,
			Message:   "创建项目失败: " + err.Error(),
			Timestamp: time.Now().Format(time.RFC3339),
		})
		return
	}

	logger.Info("项目创建成功",
		logger.String("projectID", project.ID),
		logger.String("projectName", project.Name),
		logger.String("userID", userID),
	)

	c.JSON(http.StatusOK, models.Response{
		Code:      0,
		Message:   "项目创建成功",
		Data:      project,
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// GetProject godoc
// @Summary 获取项目信息
// @Description 根据项目ID获取项目详细信息
// @Tags 项目管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer 用户令牌"
// @Param id path string true "项目ID"
// @Success 200 {object} models.Response{data=models.ProjectInfo} "获取项目成功"
// @Failure 400 {object} models.ErrorResponse "请求参数错误"
// @Failure 401 {object} models.ErrorResponse "未授权"
// @Failure 404 {object} models.ErrorResponse "项目不存在"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /api/v1/projects/{id} [get]
func (h *ProjectHandler) GetProject(c *gin.Context) {
	projectID := c.Param("id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:      http.StatusBadRequest,
			Message:   "项目ID不能为空",
			Timestamp: time.Now().Format(time.RFC3339),
		})
		return
	}

	// 从中间件获取用户ID
	userID := c.GetString("user_id")

	project, err := h.projectService.GetProject(c.Request.Context(), projectID, userID)
	if err != nil {
		if err.Error() == "access denied" {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Code:      http.StatusForbidden,
				Message:   "访问被拒绝",
				Timestamp: time.Now().Format(time.RFC3339),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:      http.StatusInternalServerError,
			Message:   "获取项目失败: " + err.Error(),
			Timestamp: time.Now().Format(time.RFC3339),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code:      0,
		Message:   "获取项目成功",
		Data:      project,
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// UpdateProject godoc
// @Summary 更新项目
// @Description 更新项目信息
// @Tags 项目管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer 用户令牌"
// @Param id path string true "项目ID"
// @Param project body models.UpdateProjectRequest true "项目更新请求"
// @Success 200 {object} models.Response{data=models.ProjectInfo} "项目更新成功"
// @Failure 400 {object} models.ErrorResponse "请求参数错误"
// @Failure 401 {object} models.ErrorResponse "未授权"
// @Failure 403 {object} models.ErrorResponse "访问被拒绝"
// @Failure 404 {object} models.ErrorResponse "项目不存在"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /api/v1/projects/{id} [put]
func (h *ProjectHandler) UpdateProject(c *gin.Context) {
	projectID := c.Param("id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:      http.StatusBadRequest,
			Message:   "项目ID不能为空",
			Timestamp: time.Now().Format(time.RFC3339),
		})
		return
	}

	var req models.UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:      http.StatusBadRequest,
			Message:   "请求参数错误",
			Timestamp: time.Now().Format(time.RFC3339),
		})
		return
	}

	// 从中间件获取用户ID
	userID := c.GetString("user_id")

	project, err := h.projectService.UpdateProject(c.Request.Context(), projectID, &req, userID)
	if err != nil {
		if err.Error() == "access denied" {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Code:      http.StatusForbidden,
				Message:   "访问被拒绝",
				Timestamp: time.Now().Format(time.RFC3339),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:      http.StatusInternalServerError,
			Message:   "更新项目失败: " + err.Error(),
			Timestamp: time.Now().Format(time.RFC3339),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code:      0,
		Message:   "项目更新成功",
		Data:      project,
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// DeleteProject godoc
// @Summary 删除项目
// @Description 删除指定项目
// @Tags 项目管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer 用户令牌"
// @Param id path string true "项目ID"
// @Success 200 {object} models.Response "项目删除成功"
// @Failure 400 {object} models.ErrorResponse "请求参数错误"
// @Failure 401 {object} models.ErrorResponse "未授权"
// @Failure 403 {object} models.ErrorResponse "访问被拒绝"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /api/v1/projects/{id} [delete]
func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	projectID := c.Param("id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:      http.StatusBadRequest,
			Message:   "项目ID不能为空",
			Timestamp: time.Now().Format(time.RFC3339),
		})
		return
	}

	// 从中间件获取用户ID
	userID := c.GetString("user_id")

	err := h.projectService.DeleteProject(c.Request.Context(), projectID, userID)
	if err != nil {
		if err.Error() == "access denied" {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Code:      http.StatusForbidden,
				Message:   "访问被拒绝",
				Timestamp: time.Now().Format(time.RFC3339),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:      http.StatusInternalServerError,
			Message:   "删除项目失败: " + err.Error(),
			Timestamp: time.Now().Format(time.RFC3339),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code:      0,
		Message:   "项目删除成功",
		Data:      nil,
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// ListProjects godoc
// @Summary 获取项目列表
// @Description 获取项目列表，支持分页和筛选
// @Tags 项目管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer 用户令牌"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param status query string false "项目状态" Enums(draft, in_progress, completed, failed)
// @Param search query string false "搜索关键词"
// @Success 200 {object} models.Response{data=models.PaginationResponse{data=[]models.ProjectInfo}} "获取项目列表成功"
// @Failure 400 {object} models.ErrorResponse "请求参数错误"
// @Failure 401 {object} models.ErrorResponse "未授权"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /api/v1/projects [get]
func (h *ProjectHandler) ListProjects(c *gin.Context) {
	var req models.ProjectListRequest

	// 解析查询参数
	if page, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil {
		req.Page = page
	} else {
		req.Page = 1
	}

	if pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "10")); err == nil {
		req.PageSize = pageSize
	} else {
		req.PageSize = 10
	}

	req.Status = c.Query("status")
	req.Search = c.Query("search")

	// 从中间件获取用户ID
	userID := c.GetString("user_id")

	projects, pagination, err := h.projectService.GetUserProjects(c.Request.Context(), userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:      http.StatusInternalServerError,
			Message:   "获取项目列表失败: " + err.Error(),
			Timestamp: time.Now().Format(time.RFC3339),
		})
		return
	}

	// 使用 projects 变量来构建响应
	_ = projects // 避免未使用变量警告

	c.JSON(http.StatusOK, models.Response{
		Code:      0,
		Message:   "获取项目列表成功",
		Data:      pagination,
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// UpdateProjectStatus godoc
// @Summary 更新项目状态
// @Description 更新项目状态
// @Tags 项目管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer 用户令牌"
// @Param id path string true "项目ID"
// @Param status query string true "项目状态" Enums(draft, in_progress, completed, failed)
// @Success 200 {object} models.Response "状态更新成功"
// @Failure 400 {object} models.ErrorResponse "请求参数错误"
// @Failure 401 {object} models.ErrorResponse "未授权"
// @Failure 403 {object} models.ErrorResponse "访问被拒绝"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /api/v1/projects/{id}/status [put]
func (h *ProjectHandler) UpdateProjectStatus(c *gin.Context) {
	projectID := c.Param("id")
	status := c.Query("status")

	if projectID == "" || status == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:      http.StatusBadRequest,
			Message:   "项目ID和状态不能为空",
			Timestamp: time.Now().Format(time.RFC3339),
		})
		return
	}

	// 从中间件获取用户ID
	userID := c.GetString("user_id")

	err := h.projectService.UpdateProjectStatus(c.Request.Context(), projectID, status, userID)
	if err != nil {
		if err.Error() == "access denied" {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Code:      http.StatusForbidden,
				Message:   "访问被拒绝",
				Timestamp: time.Now().Format(time.RFC3339),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:      http.StatusInternalServerError,
			Message:   "更新项目状态失败: " + err.Error(),
			Timestamp: time.Now().Format(time.RFC3339),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code:      0,
		Message:   "项目状态更新成功",
		Data:      nil,
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// GetProjectTags godoc
// @Summary 获取项目标签
// @Description 获取项目的所有标签
// @Tags 项目管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer 用户令牌"
// @Param id path string true "项目ID"
// @Success 200 {object} models.Response{data=[]models.TagInfo} "获取项目标签成功"
// @Failure 400 {object} models.ErrorResponse "请求参数错误"
// @Failure 401 {object} models.ErrorResponse "未授权"
// @Failure 403 {object} models.ErrorResponse "访问被拒绝"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /api/v1/projects/{id}/tags [get]
func (h *ProjectHandler) GetProjectTags(c *gin.Context) {
	projectID := c.Param("id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:      http.StatusBadRequest,
			Message:   "项目ID不能为空",
			Timestamp: time.Now().Format(time.RFC3339),
		})
		return
	}

	// 从中间件获取用户ID
	userID := c.GetString("user_id")

	tags, err := h.projectService.GetProjectTags(c.Request.Context(), projectID, userID)
	if err != nil {
		if err.Error() == "access denied" {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Code:      http.StatusForbidden,
				Message:   "访问被拒绝",
				Timestamp: time.Now().Format(time.RFC3339),
			})
			return
		}
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

// DownloadProject godoc
// @Summary 下载项目文件
// @Description 将项目文件打包为zip并下载
// @Tags 项目管理
// @Accept json
// @Produce application/zip
// @Param Authorization header string true "Bearer 用户令牌"
// @Param id path string true "项目ID"
// @Success 200 {file} file "项目文件zip包"
// @Failure 400 {object} models.ErrorResponse "请求参数错误"
// @Failure 401 {object} models.ErrorResponse "未授权"
// @Failure 403 {object} models.ErrorResponse "访问被拒绝"
// @Failure 404 {object} models.ErrorResponse "项目不存在"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /api/v1/projects/{id}/download [get]
func (h *ProjectHandler) DownloadProject(c *gin.Context) {
	projectID := c.Param("id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:      http.StatusBadRequest,
			Message:   "项目ID不能为空",
			Timestamp: time.Now().Format(time.RFC3339),
		})
		return
	}

	// 从中间件获取用户ID
	userID := c.GetString("user_id")

	// 获取项目信息
	project, err := h.projectService.GetProject(c.Request.Context(), projectID, userID)
	if err != nil {
		if err.Error() == "access denied" {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Code:      http.StatusForbidden,
				Message:   "访问被拒绝",
				Timestamp: time.Now().Format(time.RFC3339),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:      http.StatusInternalServerError,
			Message:   "获取项目信息失败: " + err.Error(),
			Timestamp: time.Now().Format(time.RFC3339),
		})
		return
	}

	// 生成zip文件
	zipData, err := h.projectService.DownloadProject(c.Request.Context(), projectID, userID)
	if err != nil {
		logger.Error("生成项目zip文件失败",
			logger.String("error", err.Error()),
			logger.String("projectID", projectID),
		)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:      http.StatusInternalServerError,
			Message:   "生成项目文件失败: " + err.Error(),
			Timestamp: time.Now().Format(time.RFC3339),
		})
		return
	}

	// 设置响应头
	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.zip\"", project.Name))
	c.Header("Content-Length", fmt.Sprintf("%d", len(zipData)))

	// 返回zip文件
	c.Data(http.StatusOK, "application/zip", zipData)
}
