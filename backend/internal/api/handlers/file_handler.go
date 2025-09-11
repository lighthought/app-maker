package handlers

import (
	"autocodeweb-backend/internal/models"
	"autocodeweb-backend/internal/services"
	"autocodeweb-backend/internal/utils"
	"autocodeweb-backend/pkg/logger"
	"net/http"
	"os"

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

// DownloadFile 下载文件
// @Summary 下载文件
// @Description 下载指定文件
// @Tags 项目文件
// @Accept json
// @Produce application/zip
// @Security Bearer
// @Param filePath query string true "文件路径"
// @Success 200 {file} file "文件"
// @Failure 400 {object} models.ErrorResponse "请求参数错误"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /api/v1/files/download [get]
func (h *FileHandler) DownloadFile(c *gin.Context) {
	filePath := c.Query("filePath")
	if filePath == "" {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Code:      models.VALIDATION_ERROR,
			Message:   "文件路径不能为空",
			Timestamp: utils.GetCurrentTime(),
		})
		return
	}

	fullPath, err := utils.GetSafeFilePath(filePath)
	logger.Info("获取安全路径",
		logger.String("filePath", filePath),
		logger.String("fullPath", fullPath),
	)

	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Code:      models.VALIDATION_ERROR,
			Message:   "文件路径不合法: " + err.Error(),
			Timestamp: utils.GetCurrentTime(),
		})
		return
	}

	file, err := os.Open(fullPath)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Code:      models.NOT_FOUND,
			Message:   "文件不存在: " + err.Error(),
			Timestamp: utils.GetCurrentTime(),
		})
		return
	}
	defer file.Close()

	fileInfo, _ := file.Stat()
	// 该方法会自动处理 Range 请求头（断点续传）、If-Modified-Since 等
	http.ServeContent(c.Writer, c.Request, fileInfo.Name(), fileInfo.ModTime(), file)
}

// GetProjectFiles 获取项目文件列表
// @Summary 获取项目文件列表
// @Description 获取指定项目的文件树结构
// @Tags 项目文件
// @Accept json
// @Produce json
// @Security Bearer
// @Param projectId path string true "项目ID"
// @Param path query string false "目录路径"
// @Success 200 {object} map[string]interface{} "成功响应"
// @Failure 400 {object} map[string]string "请求参数错误"
// @Failure 500 {object} map[string]string "服务器内部错误"
// @Router /api/v1/files/files/{projectId} [get]
func (h *FileHandler) GetProjectFiles(c *gin.Context) {
	projectID := c.Param("projectId")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:      models.VALIDATION_ERROR,
			Message:   "项目ID不能为空",
			Timestamp: utils.GetCurrentTime(),
		})
		return
	}

	path := c.Query("path")
	userID := c.GetString("user_id")

	files, err := h.fileService.GetProjectFiles(c.Request.Context(), userID, projectID, path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:      models.INTERNAL_ERROR,
			Message:   "获取文件列表失败",
			Timestamp: utils.GetCurrentTime(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code:      models.SUCCESS_CODE,
		Message:   "success",
		Data:      files,
		Timestamp: utils.GetCurrentTime(),
	})
}

// GetFileContent 获取文件内容
// @Summary 获取文件内容
// @Description 获取指定文件的内容
// @Tags 项目文件
// @Accept json
// @Produce json
// @Security Bearer
// @Param projectId path string true "项目ID"
// @Param filePath query string true "文件路径"
// @Success 200 {object} map[string]interface{} "成功响应"
// @Failure 400 {object} map[string]string "请求参数错误"
// @Failure 500 {object} map[string]string "服务器内部错误"
// @Router /api/v1/files/filecontent/{projectId} [get]
func (h *FileHandler) GetFileContent(c *gin.Context) {
	projectID := c.Param("projectId")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:      models.VALIDATION_ERROR,
			Message:   "项目ID不能为空",
			Timestamp: utils.GetCurrentTime(),
		})
		return
	}

	filePath := c.Query("filePath")
	if filePath == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:      models.VALIDATION_ERROR,
			Message:   "文件路径不能为空",
			Timestamp: utils.GetCurrentTime(),
		})
		return
	}

	userID := c.GetString("user_id")

	content, err := h.fileService.GetFileContent(c.Request.Context(), userID, projectID, filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:      models.INTERNAL_ERROR,
			Message:   "获取文件内容失败",
			Timestamp: utils.GetCurrentTime(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code:      models.SUCCESS_CODE,
		Message:   "success",
		Data:      content,
		Timestamp: utils.GetCurrentTime(),
	})
}
