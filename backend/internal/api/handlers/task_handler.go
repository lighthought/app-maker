package handlers

import (
	"net/http"
	"strconv"

	"autocodeweb-backend/internal/services"

	"github.com/gin-gonic/gin"
)

// TaskHandler 任务处理器
// @Summary 任务处理器
// @Description 处理任务相关的API请求
type TaskHandler struct {
	taskService services.TaskService
}

// NewTaskHandler 创建任务处理器
// @Summary 创建任务处理器
// @Description 创建并返回一个新的任务处理器实例
func NewTaskHandler(taskService services.TaskService) *TaskHandler {
	return &TaskHandler{
		taskService: taskService,
	}
}

// GetProjectTasks 获取项目任务列表
// @Summary 获取项目任务列表
// @Description 获取指定项目的所有任务列表，支持分页查询
// @Tags 任务管理
// @Accept json
// @Produce json
// @Param projectId path string true "项目ID"
// @Param page query int false "页码，默认为1"
// @Param pageSize query int false "每页数量，默认为10"
// @Param Authorization header string true "Bearer 用户令牌"
// @Success 200 {object} map[string]interface{} "成功返回任务列表"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/tasks/project/{projectId} [get]
func (h *TaskHandler) GetProjectTasks(c *gin.Context) {
	projectID := c.Param("projectId")
	userID := c.GetString("user_id")

	// 分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	limit := pageSize
	offset := (page - 1) * pageSize

	// 获取任务列表
	tasks, err := h.taskService.GetProjectTasks(c.Request.Context(), projectID, userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    1,
			"message": "获取任务列表失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": tasks,
	})
}

// GetTaskDetails 获取任务详情
// @Summary 获取任务详情
// @Description 获取指定任务的详细信息
// @Tags 任务管理
// @Accept json
// @Produce json
// @Param taskId path string true "任务ID"
// @Param Authorization header string true "Bearer 用户令牌"
// @Success 200 {object} map[string]interface{} "成功返回任务详情"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 404 {object} map[string]interface{} "任务不存在"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/tasks/{taskId} [get]
func (h *TaskHandler) GetTaskDetails(c *gin.Context) {
	taskID := c.Param("taskId")
	userID := c.GetString("user_id")

	// 获取任务详情
	task, err := h.taskService.GetTaskDetails(c.Request.Context(), taskID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    1,
			"message": "获取任务详情失败: " + err.Error(),
		})
		return
	}

	if task == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    1,
			"message": "任务不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": task,
	})
}

// GetTaskLogs 获取任务日志
// @Summary 获取任务日志
// @Description 获取指定任务的执行日志，支持分页查询
// @Tags 任务管理
// @Accept json
// @Produce json
// @Param taskId path string true "任务ID"
// @Param page query int false "页码，默认为1"
// @Param pageSize query int false "每页数量，默认为50"
// @Param Authorization header string true "Bearer 用户令牌"
// @Success 200 {object} map[string]interface{} "成功返回任务日志"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/tasks/{taskId}/logs [get]
func (h *TaskHandler) GetTaskLogs(c *gin.Context) {
	taskID := c.Param("taskId")
	userID := c.GetString("user_id")

	// 分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "50"))
	limit := pageSize
	offset := (page - 1) * pageSize

	// 获取任务日志
	logs, err := h.taskService.GetTaskLogs(c.Request.Context(), taskID, userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    1,
			"message": "获取任务日志失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": logs,
	})
}

// CancelTask 取消任务
// @Summary 取消任务
// @Description 取消正在执行的任务
// @Tags 任务管理
// @Accept json
// @Produce json
// @Param taskId path string true "任务ID"
// @Param Authorization header string true "Bearer 用户令牌"
// @Success 200 {object} map[string]interface{} "成功取消任务"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/tasks/{taskId}/cancel [post]
func (h *TaskHandler) CancelTask(c *gin.Context) {
	taskID := c.Param("taskId")
	userID := c.GetString("user_id")

	// 取消任务
	err := h.taskService.CancelTask(c.Request.Context(), taskID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    1,
			"message": "取消任务失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "任务已取消",
	})
}
