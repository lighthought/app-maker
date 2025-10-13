package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"autocodeweb-backend/internal/models"
	"autocodeweb-backend/internal/services"
	"shared-models/common"
	"shared-models/logger"
	"shared-models/utils"

	"github.com/gin-gonic/gin"
)

// ProjectHandler 项目处理器
type ProjectHandler struct {
	projectService      services.ProjectService
	projectStageService services.ProjectStageService
	previewService      services.PreviewService
}

// NewProjectHandler 创建项目处理器实例
func NewProjectHandler(projectService services.ProjectService, projectStageService services.ProjectStageService, previewService services.PreviewService) *ProjectHandler {
	return &ProjectHandler{
		projectService:      projectService,
		projectStageService: projectStageService,
		previewService:      previewService,
	}
}

// CreateProject godoc
// @Summary 创建项目
// @Description 创建新项目
// @Tags 项目管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param project body models.CreateProjectRequest true "项目创建请求"
// @Success 200 {object} common.Response{data=models.ProjectInfo} "项目创建成功"
// @Failure 400 {object} common.ErrorResponse "请求参数错误"
// @Failure 401 {object} common.ErrorResponse "未授权"
// @Failure 500 {object} common.ErrorResponse "服务器内部错误"
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
		c.JSON(http.StatusBadRequest, utils.GetErrorResponse(common.VALIDATION_ERROR, "请求参数错误, "+err.Error()))
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
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.INTERNAL_ERROR, "创建项目失败: "+err.Error()))
		return
	}

	logger.Info("项目创建成功",
		logger.String("projectGUID", project.GUID),
		logger.String("projectName", project.Name),
		logger.String("userID", userID),
	)

	c.JSON(http.StatusOK, utils.GetSuccessResponse("项目创建成功", project))
}

// GetProject godoc
// @Summary 获取项目信息
// @Description 根据项目ID获取项目详细信息
// @Tags 项目管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param guid path string true "项目GUID"
// @Success 200 {object} common.Response{data=models.ProjectInfo} "获取项目成功"
// @Failure 400 {object} common.ErrorResponse "请求参数错误"
// @Failure 401 {object} common.ErrorResponse "未授权"
// @Failure 404 {object} common.ErrorResponse "项目不存在"
// @Failure 500 {object} common.ErrorResponse "服务器内部错误"
// @Router /api/v1/projects/{guid} [get]
func (h *ProjectHandler) GetProject(c *gin.Context) {
	projectGuid := c.Param("guid")
	if projectGuid == "" {
		c.JSON(http.StatusBadRequest, utils.GetErrorResponse(common.VALIDATION_ERROR, "项目GUID不能为空"))
		return
	}

	// 从中间件获取用户ID
	userID := c.GetString("user_id")

	project, err := h.projectService.GetProject(c.Request.Context(), projectGuid, userID)
	if err != nil {
		if err.Error() == common.MESSAGE_ACCESS_DENIED {
			c.JSON(http.StatusForbidden, utils.GetErrorResponse(common.FORBIDDEN, "访问被拒绝"))
			return
		}
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.INTERNAL_ERROR, "获取项目失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.GetSuccessResponse("获取项目成功", project))
}

// DeleteProject godoc
// @Summary 删除项目
// @Description 删除指定项目
// @Tags 项目管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param guid path string true "项目GUID"
// @Success 200 {object} common.Response "项目删除成功"
// @Failure 400 {object} common.ErrorResponse "请求参数错误"
// @Failure 401 {object} common.ErrorResponse "未授权"
// @Failure 403 {object} common.ErrorResponse "访问被拒绝"
// @Failure 500 {object} common.ErrorResponse "服务器内部错误"
// @Router /api/v1/projects/{guid} [delete]
func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	projectGuid := c.Param("guid")
	if projectGuid == "" {
		c.JSON(http.StatusBadRequest, utils.GetErrorResponse(common.VALIDATION_ERROR, "项目GUID不能为空"))
		return
	}

	// 从中间件获取用户ID
	userID := c.GetString("user_id")

	err := h.projectService.DeleteProject(c.Request.Context(), projectGuid, userID)
	if err != nil {
		if err.Error() == common.MESSAGE_ACCESS_DENIED {
			c.JSON(http.StatusForbidden, utils.GetErrorResponse(common.FORBIDDEN, "访问被拒绝"))
			return
		}
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.INTERNAL_ERROR, "删除项目失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.GetSuccessResponse("项目删除成功", nil))
}

// ListProjects godoc
// @Summary 获取项目列表
// @Description 获取项目列表，支持分页和筛选
// @Tags 项目管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param status query string false "项目状态" Enums(draft, in_progress, completed, failed)
// @Param search query string false "搜索关键词"
// @Success 200 {object} common.Response{data=models.PaginationResponse{data=[]models.ProjectInfo}} "获取项目列表成功"
// @Failure 400 {object} common.ErrorResponse "请求参数错误"
// @Failure 401 {object} common.ErrorResponse "未授权"
// @Failure 500 {object} common.ErrorResponse "服务器内部错误"
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

	pagination, err := h.projectService.GetUserProjects(c.Request.Context(), userID, &req)
	if err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.INTERNAL_ERROR, "获取项目列表失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, pagination)
}

// GetProjectStages 获取项目开发阶段
// @Summary 获取项目开发阶段
// @Description 获取指定项目的开发阶段信息
// @Tags 开发阶段
// @Accept json
// @Produce json
// @Security Bearer
// @Param guid path string true "项目GUID"
// @Success 200 {object} map[string]interface{} "成功响应"
// @Failure 400 {object} map[string]string "请求参数错误"
// @Failure 500 {object} map[string]string "服务器内部错误"
// @Router /api/v1/projects/{guid}/stages [get]
func (h *ProjectHandler) GetProjectStages(c *gin.Context) {
	projectGuid := c.Param("guid")
	if projectGuid == "" {
		c.JSON(http.StatusBadRequest, utils.GetErrorResponse(common.VALIDATION_ERROR, "项目GUID不能为空"))
		return
	}

	stages, err := h.projectStageService.GetProjectStages(c.Request.Context(), projectGuid)
	if err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.INTERNAL_ERROR, "获取开发阶段失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.GetSuccessResponse("获取开发阶段成功", stages))
}

// DownloadProject godoc
// @Summary 下载项目文件
// @Description 将项目文件打包为zip并下载
// @Tags 项目管理
// @Accept json
// @Produce application/zip
// @Security Bearer
// @Param guid path string true "项目GUID"
// @Success 200 {file} file "项目文件zip包"
// @Failure 400 {object} common.ErrorResponse "请求参数错误"
// @Failure 401 {object} common.ErrorResponse "未授权"
// @Failure 403 {object} common.ErrorResponse "访问被拒绝"
// @Failure 404 {object} common.ErrorResponse "项目不存在"
// @Failure 500 {object} common.ErrorResponse "服务器内部错误"
// @Router /api/v1/projects/download/{guid} [get]
func (h *ProjectHandler) DownloadProject(c *gin.Context) {
	projectGuid := c.Param("guid")
	if projectGuid == "" {
		c.JSON(http.StatusBadRequest, utils.GetErrorResponse(common.VALIDATION_ERROR, "项目GUID不能为空"))
		return
	}

	// 从中间件获取用户ID
	userID := c.GetString("user_id")

	// 获取项目信息
	project, err := h.projectService.CheckProjectAccess(c.Request.Context(), projectGuid, userID)
	if err != nil {
		if err.Error() == common.MESSAGE_ACCESS_DENIED {
			c.JSON(http.StatusForbidden, utils.GetErrorResponse(common.FORBIDDEN, "访问被拒绝"))
			return
		}
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.INTERNAL_ERROR, "获取项目信息失败: "+err.Error()))
		return
	}

	// 生成项目压缩任务
	taskID, err := h.projectService.CreateDownloadProjectTask(c.Request.Context(), project.ID, projectGuid, project.ProjectPath)
	if err != nil {
		logger.Error("生成项目压缩任务失败",
			logger.String("error", err.Error()),
			logger.String("projectID", project.ID),
		)
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.INTERNAL_ERROR, "生成项目压缩任务失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.GetSuccessResponse("生成项目压缩任务成功", taskID))
}

// DeployProject godoc
// @Summary 部署项目
// @Description 部署指定项目
// @Tags 项目管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param guid path string true "项目GUID"
// @Success 200 {object} common.Response "项目部署成功"
// @Failure 400 {object} common.ErrorResponse "请求参数错误"
// @Failure 401 {object} common.ErrorResponse "未授权"
// @Failure 403 {object} common.ErrorResponse "访问被拒绝"
// @Failure 500 {object} common.ErrorResponse "服务器内部错误"
// @Router /api/v1/projects/{guid}/deploy [post]
func (h *ProjectHandler) DeployProject(c *gin.Context) {
	projectGuid := c.Param("guid")
	if projectGuid == "" {
		c.JSON(http.StatusBadRequest, utils.GetErrorResponse(common.VALIDATION_ERROR, "项目GUID不能为空"))
		return
	}

	// 从中间件获取用户ID
	userID := c.GetString("user_id")

	// 验证用户权限
	project, err := h.projectService.CheckProjectAccess(c.Request.Context(), projectGuid, userID)
	if err != nil {
		if err.Error() == common.MESSAGE_ACCESS_DENIED {
			c.JSON(http.StatusForbidden, utils.GetErrorResponse(common.FORBIDDEN, "访问被拒绝"))
			return
		}
		c.JSON(http.StatusInternalServerError, utils.GetErrorResponse(common.INTERNAL_ERROR, "获取项目信息失败: "+err.Error()))
		return
	}

	taskID, err := h.projectService.CreateDeployProjectTask(c.Request.Context(), project)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.GetErrorResponse(common.INTERNAL_ERROR, "创建部署项目任务失败: "+err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.GetSuccessResponse("创建部署项目任务成功", taskID))
}

// GeneratePreviewLink godoc
// @Summary 生成预览分享链接
// @Description 为项目生成可分享的预览链接
// @Tags 项目管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param guid path string true "项目GUID"
// @Param days query int false "过期天数" default(7)
// @Success 200 {object} common.Response{data=map[string]interface{}} "成功生成分享链接"
// @Failure 400 {object} common.ErrorResponse "请求参数错误"
// @Failure 401 {object} common.ErrorResponse "未授权"
// @Failure 403 {object} common.ErrorResponse "访问被拒绝"
// @Failure 500 {object} common.ErrorResponse "服务器内部错误"
// @Router /api/v1/projects/{guid}/preview-link [post]
func (h *ProjectHandler) GeneratePreviewLink(c *gin.Context) {
	projectGuid := c.Param("guid")
	if projectGuid == "" {
		c.JSON(http.StatusBadRequest, utils.GetErrorResponse(common.VALIDATION_ERROR, "项目GUID不能为空"))
		return
	}

	// 从中间件获取用户ID
	userID := c.GetString("user_id")

	// 验证用户权限
	project, err := h.projectService.CheckProjectAccess(c.Request.Context(), projectGuid, userID)
	if err != nil {
		if err.Error() == common.MESSAGE_ACCESS_DENIED {
			c.JSON(http.StatusForbidden, utils.GetErrorResponse(common.FORBIDDEN, "访问被拒绝"))
			return
		}
		c.JSON(http.StatusInternalServerError, utils.GetErrorResponse(common.INTERNAL_ERROR, "获取项目信息失败: "+err.Error()))
		return
	}

	// 获取过期天数参数（默认7天）
	daysStr := c.DefaultQuery("days", "7")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days <= 0 {
		days = 7
	}

	// 生成预览令牌
	token, err := h.previewService.GeneratePreviewToken(c.Request.Context(), project.ID, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.GetErrorResponse(common.INTERNAL_ERROR, "生成预览令牌失败: "+err.Error()))
		return
	}

	// 构建完整的分享链接
	baseURL := c.Request.Host
	shareLink := fmt.Sprintf("http://%s/api/v1/preview/%s", baseURL, token.Token)

	shareInfo := models.ProjectShareInfo{
		Token:     token.Token,
		ExpiresAt: token.ExpiresAt,
		ShareLink: shareLink,
	}
	c.JSON(http.StatusOK, utils.GetSuccessResponse("生成分享链接成功", shareInfo))
}

// GetPreviewByToken godoc
// @Summary 通过令牌访问预览
// @Description 通过分享令牌访问项目预览（无需认证）
// @Tags 项目管理
// @Accept json
// @Produce json
// @Param token path string true "预览令牌"
// @Success 302 {string} string "重定向到预览页面"
// @Failure 400 {object} common.ErrorResponse "请求参数错误"
// @Failure 404 {object} common.ErrorResponse "令牌不存在或已过期"
// @Router /api/v1/preview/{token} [get]
func (h *ProjectHandler) GetPreviewByToken(c *gin.Context) {
	tokenStr := c.Param("token")
	if tokenStr == "" {
		c.JSON(http.StatusBadRequest, utils.GetErrorResponse(common.VALIDATION_ERROR, "预览令牌不能为空"))
		return
	}

	// 获取预览令牌和项目信息
	previewToken, err := h.previewService.GetPreviewByToken(c.Request.Context(), tokenStr)
	if err != nil {
		c.JSON(http.StatusNotFound, utils.GetErrorResponse(common.NOT_FOUND, "预览令牌无效或已过期"))
		return
	}

	// 获取项目信息
	project, err := h.projectService.GetProjectByID(c.Request.Context(), previewToken.ProjectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.GetErrorResponse(common.INTERNAL_ERROR, "获取项目信息失败: "+err.Error()))
		return
	}

	// 如果项目有预览URL，重定向到预览页面
	if project.PreviewUrl != "" {
		c.Redirect(http.StatusFound, project.PreviewUrl)
		return
	}

	// 否则返回项目信息
	c.JSON(http.StatusOK, utils.GetSuccessResponse("获取预览信息成功", map[string]interface{}{
		"project_guid": project.GUID,
		"project_name": project.Name,
		"message":      "项目预览暂未部署",
	}))
}
