package handlers

import (
	"fmt"
	"net/http"
	"shared-models/common"
	"strconv"

	"autocodeweb-backend/internal/models"
	"autocodeweb-backend/internal/services"
	"autocodeweb-backend/internal/utils"
	"autocodeweb-backend/pkg/logger"

	"github.com/gin-gonic/gin"
)

// ProjectHandler 项目处理器
type ProjectHandler struct {
	projectService      services.ProjectService
	projectStageService services.ProjectStageService
}

// NewProjectHandler 创建项目处理器实例
func NewProjectHandler(projectService services.ProjectService, projectStageService services.ProjectStageService) *ProjectHandler {
	return &ProjectHandler{
		projectService:      projectService,
		projectStageService: projectStageService,
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
		c.JSON(http.StatusBadRequest, common.ErrorResponse{
			Code:      common.VALIDATION_ERROR,
			Message:   "请求参数错误, " + err.Error(),
			Timestamp: utils.GetCurrentTime(),
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
		c.JSON(http.StatusOK, common.ErrorResponse{
			Code:      common.INTERNAL_ERROR,
			Message:   "创建项目失败: " + err.Error(),
			Timestamp: utils.GetCurrentTime(),
		})
		return
	}

	logger.Info("项目创建成功",
		logger.String("projectGUID", project.GUID),
		logger.String("projectName", project.Name),
		logger.String("userID", userID),
	)

	c.JSON(http.StatusOK, common.Response{
		Code:      common.SUCCESS_CODE,
		Message:   "项目创建成功",
		Data:      project,
		Timestamp: utils.GetCurrentTime(),
	})
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
		c.JSON(http.StatusBadRequest, common.ErrorResponse{
			Code:      common.VALIDATION_ERROR,
			Message:   "项目GUID不能为空",
			Timestamp: utils.GetCurrentTime(),
		})
		return
	}

	// 从中间件获取用户ID
	userID := c.GetString("user_id")

	project, err := h.projectService.GetProject(c.Request.Context(), projectGuid, userID)
	if err != nil {
		if err.Error() == "access denied" {
			c.JSON(http.StatusForbidden, common.ErrorResponse{
				Code:      common.FORBIDDEN,
				Message:   "访问被拒绝",
				Timestamp: utils.GetCurrentTime(),
			})
			return
		}
		c.JSON(http.StatusOK, common.ErrorResponse{
			Code:      common.INTERNAL_ERROR,
			Message:   "获取项目失败: " + err.Error(),
			Timestamp: utils.GetCurrentTime(),
		})
		return
	}

	c.JSON(http.StatusOK, common.Response{
		Code:      common.SUCCESS_CODE,
		Message:   "获取项目成功",
		Data:      project,
		Timestamp: utils.GetCurrentTime(),
	})
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
		c.JSON(http.StatusBadRequest, common.ErrorResponse{
			Code:      common.VALIDATION_ERROR,
			Message:   "项目GUID不能为空",
			Timestamp: utils.GetCurrentTime(),
		})
		return
	}

	// 从中间件获取用户ID
	userID := c.GetString("user_id")

	err := h.projectService.DeleteProject(c.Request.Context(), projectGuid, userID)
	if err != nil {
		if err.Error() == "access denied" {
			c.JSON(http.StatusForbidden, common.ErrorResponse{
				Code:      common.FORBIDDEN,
				Message:   "访问被拒绝",
				Timestamp: utils.GetCurrentTime(),
			})
			return
		}
		c.JSON(http.StatusOK, common.ErrorResponse{
			Code:      common.INTERNAL_ERROR,
			Message:   "删除项目失败: " + err.Error(),
			Timestamp: utils.GetCurrentTime(),
		})
		return
	}

	c.JSON(http.StatusOK, common.Response{
		Code:      common.SUCCESS_CODE,
		Message:   "项目删除成功",
		Data:      nil,
		Timestamp: utils.GetCurrentTime(),
	})
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
		c.JSON(http.StatusOK, common.ErrorResponse{
			Code:      common.INTERNAL_ERROR,
			Message:   "获取项目列表失败: " + err.Error(),
			Timestamp: utils.GetCurrentTime(),
		})
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
		c.JSON(http.StatusBadRequest, common.ErrorResponse{
			Code:      common.VALIDATION_ERROR,
			Message:   "项目GUID不能为空",
			Timestamp: utils.GetCurrentTime(),
		})
		return
	}

	stages, err := h.projectStageService.GetProjectStages(c.Request.Context(), projectGuid)
	if err != nil {
		c.JSON(http.StatusOK, common.ErrorResponse{
			Code:      common.INTERNAL_ERROR,
			Message:   "获取开发阶段失败, " + err.Error(),
			Timestamp: utils.GetCurrentTime(),
		})
		return
	}

	c.JSON(http.StatusOK, common.Response{
		Code:      common.SUCCESS_CODE,
		Message:   "获取开发阶段成功",
		Data:      stages,
		Timestamp: utils.GetCurrentTime(),
	})
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
		c.JSON(http.StatusBadRequest, common.ErrorResponse{
			Code:      http.StatusBadRequest,
			Message:   "项目GUID不能为空",
			Timestamp: utils.GetCurrentTime(),
		})
		return
	}

	// 从中间件获取用户ID
	userID := c.GetString("user_id")

	// 获取项目信息
	project, err := h.projectService.CheckProjectAccess(c.Request.Context(), projectGuid, userID)
	if err != nil {
		if err.Error() == "access denied" {
			c.JSON(http.StatusForbidden, common.ErrorResponse{
				Code:      http.StatusForbidden,
				Message:   "访问被拒绝",
				Timestamp: utils.GetCurrentTime(),
			})
			return
		}
		c.JSON(http.StatusOK, common.ErrorResponse{
			Code:      http.StatusInternalServerError,
			Message:   "获取项目信息失败: " + err.Error(),
			Timestamp: utils.GetCurrentTime(),
		})
		return
	}

	// 生成项目压缩任务
	taskID, err := h.projectService.CreateDownloadProjectTask(c.Request.Context(), project.ID, projectGuid, project.ProjectPath)
	if err != nil {
		logger.Error("生成项目压缩任务失败",
			logger.String("error", err.Error()),
			logger.String("projectID", project.ID),
		)
		c.JSON(http.StatusOK, common.ErrorResponse{
			Code:      http.StatusInternalServerError,
			Message:   "生成项目压缩任务失败: " + err.Error(),
			Timestamp: utils.GetCurrentTime(),
		})
		return
	}

	c.JSON(http.StatusOK, common.Response{
		Code:      common.SUCCESS_CODE,
		Message:   "success",
		Data:      taskID,
		Timestamp: utils.GetCurrentTime(),
	})
}
