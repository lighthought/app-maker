package handlers

import (
	"autocodeweb-backend/internal/models"
	"autocodeweb-backend/internal/services"
	"autocodeweb-backend/pkg/logger"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// FileHandler 文件处理器
type FileHandler struct {
	projectService services.ProjectService
	fileService    services.FileService
}

// NewProjectHandler 创建项目处理器实例
func NewFileHandler(fileService services.FileService, projectService services.ProjectService) *FileHandler {
	return &FileHandler{
		fileService:    fileService,
		projectService: projectService,
	}
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
// @Router /api/v1/files/download/{projectId} [get]
func (h *FileHandler) DownloadProject(c *gin.Context) {
	projectID := c.Param("projectId")
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
	project, err := h.projectService.CheckProjectAccess(c.Request.Context(), projectID, userID)
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
	zipData, err := h.fileService.DownloadProject(c.Request.Context(), project.ProjectPath)
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

// GetProjectFiles 获取项目文件列表
// @Summary 获取项目文件列表
// @Description 获取指定项目的文件树结构
// @Tags 项目文件
// @Accept json
// @Produce json
// @Param projectId path string true "项目ID"
// @Param path query string false "目录路径"
// @Success 200 {object} map[string]interface{} "成功响应"
// @Failure 400 {object} map[string]string "请求参数错误"
// @Failure 500 {object} map[string]string "服务器内部错误"
// @Router /api/v1/files/files/{projectId} [get]
func (h *FileHandler) GetProjectFiles(c *gin.Context) {
	projectID := c.Param("projectId")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "项目ID不能为空"})
		return
	}

	path := c.Query("path")
	userID := c.GetString("user_id")

	files, err := h.fileService.GetProjectFiles(c.Request.Context(), userID, projectID, path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取文件列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    files,
	})
}

// GetFileContent 获取文件内容
// @Summary 获取文件内容
// @Description 获取指定文件的内容
// @Tags 项目文件
// @Accept json
// @Produce json
// @Param projectId path string true "项目ID"
// @Param filePath query string true "文件路径"
// @Success 200 {object} map[string]interface{} "成功响应"
// @Failure 400 {object} map[string]string "请求参数错误"
// @Failure 500 {object} map[string]string "服务器内部错误"
// @Router /api/v1/files/filecontent/{projectId} [get]
func (h *FileHandler) GetFileContent(c *gin.Context) {
	projectID := c.Param("projectId")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "项目ID不能为空"})
		return
	}

	filePath := c.Query("filePath")
	if filePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件路径不能为空"})
		return
	}

	userID := c.GetString("user_id")

	content, err := h.fileService.GetFileContent(c.Request.Context(), userID, projectID, filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取文件内容失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    content,
	})
}
