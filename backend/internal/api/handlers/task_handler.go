package handlers

import (
	"net/http"
	"strconv"
	"time"

	"autocodeweb-backend/internal/models"
	"autocodeweb-backend/internal/services"

	"github.com/gin-gonic/gin"
)

type TaskHandler struct {
	taskService services.TaskService
}

func NewTaskHandler(taskService services.TaskService) *TaskHandler {
	return &TaskHandler{
		taskService: taskService,
	}
}

// CreateTask godoc
// @Summary 创建任务
// @Description 创建新的任务
// @Tags 任务管理
// @Accept json
// @Produce json
// @Param task body models.CreateTaskRequest true "任务信息"
// @Success 200 {object} models.Response{data=models.TaskInfo}
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/tasks [post]
// @Security BearerAuth
func (h *TaskHandler) CreateTask(c *gin.Context) {
	var req models.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:      http.StatusBadRequest,
			Message:   "请求参数错误",
			Timestamp: getCurrentTimestamp(),
		})
		return
	}

	userID := getUserIDFromContext(c)
	task, err := h.taskService.CreateTask(c.Request.Context(), &req, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:      http.StatusInternalServerError,
			Message:   err.Error(),
			Timestamp: getCurrentTimestamp(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code:      0,
		Message:   "任务创建成功",
		Data:      convertTaskToInfo(task),
		Timestamp: getCurrentTimestamp(),
	})
}

// GetTask godoc
// @Summary 获取任务详情
// @Description 根据任务ID获取任务详细信息
// @Tags 任务管理
// @Accept json
// @Produce json
// @Param id path string true "任务ID"
// @Success 200 {object} models.Response{data=models.TaskInfo}
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/tasks/{id} [get]
// @Security BearerAuth
func (h *TaskHandler) GetTask(c *gin.Context) {
	taskID := c.Param("id")
	userID := getUserIDFromContext(c)

	task, err := h.taskService.GetTask(c.Request.Context(), taskID, userID)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "task not found" {
			status = http.StatusNotFound
		} else if err.Error() == "access denied" {
			status = http.StatusForbidden
		}
		c.JSON(status, models.ErrorResponse{
			Code:      status,
			Message:   err.Error(),
			Timestamp: getCurrentTimestamp(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code:      0,
		Message:   "获取任务成功",
		Data:      convertTaskToInfo(task),
		Timestamp: getCurrentTimestamp(),
	})
}

// UpdateTask godoc
// @Summary 更新任务
// @Description 更新任务信息
// @Tags 任务管理
// @Accept json
// @Produce json
// @Param id path string true "任务ID"
// @Param task body models.UpdateTaskRequest true "任务更新信息"
// @Success 200 {object} models.Response{data=models.TaskInfo}
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/tasks/{id} [put]
// @Security BearerAuth
func (h *TaskHandler) UpdateTask(c *gin.Context) {
	taskID := c.Param("id")
	var req models.UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:      http.StatusBadRequest,
			Message:   "请求参数错误",
			Timestamp: getCurrentTimestamp(),
		})
		return
	}

	userID := getUserIDFromContext(c)
	task, err := h.taskService.UpdateTask(c.Request.Context(), taskID, &req, userID)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "task not found" {
			status = http.StatusNotFound
		} else if err.Error() == "access denied" {
			status = http.StatusForbidden
		}
		c.JSON(status, models.ErrorResponse{
			Code:      status,
			Message:   err.Error(),
			Timestamp: getCurrentTimestamp(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code:      0,
		Message:   "任务更新成功",
		Data:      convertTaskToInfo(task),
		Timestamp: getCurrentTimestamp(),
	})
}

// DeleteTask godoc
// @Summary 删除任务
// @Description 删除指定任务
// @Tags 任务管理
// @Accept json
// @Produce json
// @Param id path string true "任务ID"
// @Success 200 {object} models.Response
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/tasks/{id} [delete]
// @Security BearerAuth
func (h *TaskHandler) DeleteTask(c *gin.Context) {
	taskID := c.Param("id")
	userID := getUserIDFromContext(c)

	err := h.taskService.DeleteTask(c.Request.Context(), taskID, userID)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "task not found" {
			status = http.StatusNotFound
		} else if err.Error() == "access denied" {
			status = http.StatusForbidden
		}
		c.JSON(status, models.ErrorResponse{
			Code:      status,
			Message:   err.Error(),
			Timestamp: getCurrentTimestamp(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code:      0,
		Message:   "任务删除成功",
		Timestamp: getCurrentTimestamp(),
	})
}

// ListTasks godoc
// @Summary 获取任务列表
// @Description 分页获取任务列表，支持过滤和搜索
// @Tags 任务管理
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param project_id query string false "项目ID"
// @Param user_id query string false "用户ID"
// @Param status query string false "任务状态"
// @Param priority query int false "优先级"
// @Param tags query []string false "标签"
// @Param search query string false "搜索关键词"
// @Success 200 {object} models.Response{data=models.PaginationResponse{data=[]models.TaskInfo}}
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/tasks [get]
// @Security BearerAuth
func (h *TaskHandler) ListTasks(c *gin.Context) {
	var req models.TaskListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:      http.StatusBadRequest,
			Message:   "请求参数错误",
			Timestamp: getCurrentTimestamp(),
		})
		return
	}

	// 设置默认值
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}

	userID := getUserIDFromContext(c)
	tasks, total, err := h.taskService.ListTasks(c.Request.Context(), &req, userID)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "access denied" {
			status = http.StatusForbidden
		}
		c.JSON(status, models.ErrorResponse{
			Code:      status,
			Message:   err.Error(),
			Timestamp: getCurrentTimestamp(),
		})
		return
	}

	// 转换为TaskInfo
	taskInfos := make([]models.TaskInfo, len(tasks))
	for i, task := range tasks {
		taskInfos[i] = *convertTaskToInfo(task)
	}

	totalPages := (int(total) + req.PageSize - 1) / req.PageSize
	hasNext := req.Page < totalPages
	hasPrevious := req.Page > 1

	c.JSON(http.StatusOK, models.Response{
		Code:    0,
		Message: "获取任务列表成功",
		Data: models.PaginationResponse{
			Total:       int(total),
			Page:        req.Page,
			PageSize:    req.PageSize,
			TotalPages:  totalPages,
			Data:        taskInfos,
			HasNext:     hasNext,
			HasPrevious: hasPrevious,
		},
		Timestamp: getCurrentTimestamp(),
	})
}

// UpdateTaskStatus godoc
// @Summary 更新任务状态
// @Description 更新任务的状态
// @Tags 任务管理
// @Accept json
// @Produce json
// @Param id path string true "任务ID"
// @Param status body models.TaskStatusUpdateRequest true "状态更新信息"
// @Success 200 {object} models.Response
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/tasks/{id}/status [put]
// @Security BearerAuth
func (h *TaskHandler) UpdateTaskStatus(c *gin.Context) {
	taskID := c.Param("id")
	var req models.TaskStatusUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:      http.StatusBadRequest,
			Message:   "请求参数错误",
			Timestamp: getCurrentTimestamp(),
		})
		return
	}

	userID := getUserIDFromContext(c)
	err := h.taskService.UpdateTaskStatus(c.Request.Context(), taskID, &req, userID)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "task not found" {
			status = http.StatusNotFound
		} else if err.Error() == "access denied" {
			status = http.StatusForbidden
		} else if err.Error() == "invalid status transition" {
			status = http.StatusBadRequest
		}
		c.JSON(status, models.ErrorResponse{
			Code:      status,
			Message:   err.Error(),
			Timestamp: getCurrentTimestamp(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code:      0,
		Message:   "任务状态更新成功",
		Timestamp: getCurrentTimestamp(),
	})
}

// StartTask godoc
// @Summary 启动任务
// @Description 启动指定任务
// @Tags 任务管理
// @Accept json
// @Produce json
// @Param id path string true "任务ID"
// @Success 200 {object} models.Response
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/tasks/{id}/start [post]
// @Security BearerAuth
func (h *TaskHandler) StartTask(c *gin.Context) {
	taskID := c.Param("id")
	userID := getUserIDFromContext(c)

	err := h.taskService.StartTask(c.Request.Context(), taskID, userID)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "task not found" {
			status = http.StatusNotFound
		} else if err.Error() == "access denied" {
			status = http.StatusForbidden
		} else if err.Error() == "task is not in pending status" || err.Error() == "task dependencies not completed" {
			status = http.StatusBadRequest
		}
		c.JSON(status, models.ErrorResponse{
			Code:      status,
			Message:   err.Error(),
			Timestamp: getCurrentTimestamp(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code:      0,
		Message:   "任务启动成功",
		Timestamp: getCurrentTimestamp(),
	})
}

// CompleteTask godoc
// @Summary 完成任务
// @Description 标记任务为完成状态
// @Tags 任务管理
// @Accept json
// @Produce json
// @Param id path string true "任务ID"
// @Param result query string false "执行结果"
// @Success 200 {object} models.Response
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/tasks/{id}/complete [post]
// @Security BearerAuth
func (h *TaskHandler) CompleteTask(c *gin.Context) {
	taskID := c.Param("id")
	result := c.Query("result")
	userID := getUserIDFromContext(c)

	err := h.taskService.CompleteTask(c.Request.Context(), taskID, userID, result)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "task not found" {
			status = http.StatusNotFound
		} else if err.Error() == "access denied" {
			status = http.StatusForbidden
		} else if err.Error() == "invalid status transition" {
			status = http.StatusBadRequest
		}
		c.JSON(status, models.ErrorResponse{
			Code:      status,
			Message:   err.Error(),
			Timestamp: getCurrentTimestamp(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code:      0,
		Message:   "任务完成成功",
		Timestamp: getCurrentTimestamp(),
	})
}

// RetryTask godoc
// @Summary 重试任务
// @Description 重试失败的任务
// @Tags 任务管理
// @Accept json
// @Produce json
// @Param id path string true "任务ID"
// @Param retry body models.TaskRetryRequest true "重试参数"
// @Success 200 {object} models.Response
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/tasks/{id}/retry [post]
// @Security BearerAuth
func (h *TaskHandler) RetryTask(c *gin.Context) {
	taskID := c.Param("id")
	var req models.TaskRetryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:      http.StatusBadRequest,
			Message:   "请求参数错误",
			Timestamp: getCurrentTimestamp(),
		})
		return
	}

	userID := getUserIDFromContext(c)
	err := h.taskService.RetryTask(c.Request.Context(), taskID, &req, userID)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "task not found" {
			status = http.StatusNotFound
		} else if err.Error() == "access denied" {
			status = http.StatusForbidden
		} else if err.Error() == "task is not in failed status" || err.Error() == "max retry count exceeded" || err.Error() == "task dependencies not completed" {
			status = http.StatusBadRequest
		}
		c.JSON(status, models.ErrorResponse{
			Code:      status,
			Message:   err.Error(),
			Timestamp: getCurrentTimestamp(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code:      0,
		Message:   "任务重试成功",
		Timestamp: getCurrentTimestamp(),
	})
}

// RollbackTask godoc
// @Summary 回滚任务
// @Description 回滚已完成的任务
// @Tags 任务管理
// @Accept json
// @Produce json
// @Param id path string true "任务ID"
// @Param rollback body models.TaskRollbackRequest true "回滚参数"
// @Success 200 {object} models.Response
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/tasks/{id}/rollback [post]
// @Security BearerAuth
func (h *TaskHandler) RollbackTask(c *gin.Context) {
	taskID := c.Param("id")
	var req models.TaskRollbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:      http.StatusBadRequest,
			Message:   "请求参数错误",
			Timestamp: getCurrentTimestamp(),
		})
		return
	}

	userID := getUserIDFromContext(c)
	err := h.taskService.RollbackTask(c.Request.Context(), taskID, &req, userID)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "task not found" {
			status = http.StatusNotFound
		} else if err.Error() == "access denied" {
			status = http.StatusForbidden
		} else if err.Error() == "task is not in completed status" || err.Error() == "invalid status transition" {
			status = http.StatusBadRequest
		}
		c.JSON(status, models.ErrorResponse{
			Code:      status,
			Message:   err.Error(),
			Timestamp: getCurrentTimestamp(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code:      0,
		Message:   "任务回滚成功",
		Timestamp: getCurrentTimestamp(),
	})
}

// GetTaskDependencies godoc
// @Summary 获取任务依赖
// @Description 获取指定任务的依赖任务列表
// @Tags 任务管理
// @Accept json
// @Produce json
// @Param id path string true "任务ID"
// @Success 200 {object} models.Response{data=[]models.TaskInfo}
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/tasks/{id}/dependencies [get]
// @Security BearerAuth
func (h *TaskHandler) GetTaskDependencies(c *gin.Context) {
	taskID := c.Param("id")
	userID := getUserIDFromContext(c)

	dependencies, err := h.taskService.GetTaskDependencies(c.Request.Context(), taskID, userID)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "task not found" {
			status = http.StatusNotFound
		} else if err.Error() == "access denied" {
			status = http.StatusForbidden
		}
		c.JSON(status, models.ErrorResponse{
			Code:      status,
			Message:   err.Error(),
			Timestamp: getCurrentTimestamp(),
		})
		return
	}

	// 转换为TaskInfo
	dependencyInfos := make([]models.TaskInfo, len(dependencies))
	for i, dep := range dependencies {
		dependencyInfos[i] = *convertTaskToInfo(dep)
	}

	c.JSON(http.StatusOK, models.Response{
		Code:      0,
		Message:   "获取任务依赖成功",
		Data:      dependencyInfos,
		Timestamp: getCurrentTimestamp(),
	})
}

// GetTaskLogs godoc
// @Summary 获取任务日志
// @Description 获取指定任务的执行日志
// @Tags 任务管理
// @Accept json
// @Produce json
// @Param id path string true "任务ID"
// @Param level query string false "日志级别" Enums(info, warn, error)
// @Success 200 {object} models.Response{data=[]models.TaskLogInfo}
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/tasks/{id}/logs [get]
// @Security BearerAuth
func (h *TaskHandler) GetTaskLogs(c *gin.Context) {
	taskID := c.Param("id")
	level := c.Query("level")
	userID := getUserIDFromContext(c)

	var logs []*models.TaskLog
	var err error

	if level != "" {
		logs, err = h.taskService.GetTaskLogsByLevel(c.Request.Context(), taskID, level, userID)
	} else {
		logs, err = h.taskService.GetTaskLogs(c.Request.Context(), taskID, userID)
	}

	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "task not found" {
			status = http.StatusNotFound
		} else if err.Error() == "access denied" {
			status = http.StatusForbidden
		}
		c.JSON(status, models.ErrorResponse{
			Code:      status,
			Message:   err.Error(),
			Timestamp: getCurrentTimestamp(),
		})
		return
	}

	// 转换为TaskLogInfo
	logInfos := make([]models.TaskLogInfo, len(logs))
	for i, log := range logs {
		logInfos[i] = models.TaskLogInfo{
			ID:        log.ID,
			TaskID:    log.TaskID,
			Level:     log.Level,
			Message:   log.Message,
			Data:      log.Data,
			CreatedAt: log.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, models.Response{
		Code:      0,
		Message:   "获取任务日志成功",
		Data:      logInfos,
		Timestamp: getCurrentTimestamp(),
	})
}

// GetTasksByProject godoc
// @Summary 获取项目任务
// @Description 获取指定项目的所有任务
// @Tags 任务管理
// @Accept json
// @Produce json
// @Param id path string true "项目ID"
// @Success 200 {object} models.Response{data=[]models.TaskInfo}
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/projects/{id}/tasks [get]
// @Security BearerAuth
func (h *TaskHandler) GetTasksByProject(c *gin.Context) {
	projectID := c.Param("id")
	userID := getUserIDFromContext(c)

	tasks, err := h.taskService.GetTasksByProject(c.Request.Context(), projectID, userID)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "project not found" {
			status = http.StatusNotFound
		} else if err.Error() == "access denied" {
			status = http.StatusForbidden
		}
		c.JSON(status, models.ErrorResponse{
			Code:      status,
			Message:   err.Error(),
			Timestamp: getCurrentTimestamp(),
		})
		return
	}

	// 转换为TaskInfo
	taskInfos := make([]models.TaskInfo, len(tasks))
	for i, task := range tasks {
		taskInfos[i] = *convertTaskToInfo(task)
	}

	c.JSON(http.StatusOK, models.Response{
		Code:      0,
		Message:   "获取项目任务成功",
		Data:      taskInfos,
		Timestamp: getCurrentTimestamp(),
	})
}

// GetTasksByStatus godoc
// @Summary 根据状态获取任务
// @Description 获取指定状态的所有任务
// @Tags 任务管理
// @Accept json
// @Produce json
// @Param status path string true "任务状态" Enums(pending, running, completed, failed, cancelled, retrying, rolled_back)
// @Success 200 {object} models.Response{data=[]models.TaskInfo}
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/tasks/status/{status} [get]
// @Security BearerAuth
func (h *TaskHandler) GetTasksByStatus(c *gin.Context) {
	status := c.Param("status")
	userID := getUserIDFromContext(c)

	tasks, err := h.taskService.GetTasksByStatus(c.Request.Context(), status, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:      http.StatusInternalServerError,
			Message:   err.Error(),
			Timestamp: getCurrentTimestamp(),
		})
		return
	}

	// 转换为TaskInfo
	taskInfos := make([]models.TaskInfo, len(tasks))
	for i, task := range tasks {
		taskInfos[i] = *convertTaskToInfo(task)
	}

	c.JSON(http.StatusOK, models.Response{
		Code:      0,
		Message:   "获取任务成功",
		Data:      taskInfos,
		Timestamp: getCurrentTimestamp(),
	})
}

// GetTasksByPriority godoc
// @Summary 根据优先级获取任务
// @Description 获取指定优先级的所有任务
// @Tags 任务管理
// @Accept json
// @Produce json
// @Param priority path int true "任务优先级" minimum(1) maximum(4)
// @Success 200 {object} models.Response{data=[]models.TaskInfo}
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/tasks/priority/{priority} [get]
// @Security BearerAuth
func (h *TaskHandler) GetTasksByPriority(c *gin.Context) {
	priorityStr := c.Param("priority")
	priority, err := strconv.Atoi(priorityStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:      http.StatusBadRequest,
			Message:   "优先级参数错误",
			Timestamp: getCurrentTimestamp(),
		})
		return
	}

	userID := getUserIDFromContext(c)
	tasks, err := h.taskService.GetTasksByPriority(c.Request.Context(), priority, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:      http.StatusInternalServerError,
			Message:   err.Error(),
			Timestamp: getCurrentTimestamp(),
		})
		return
	}

	// 转换为TaskInfo
	taskInfos := make([]models.TaskInfo, len(tasks))
	for i, task := range tasks {
		taskInfos[i] = *convertTaskToInfo(task)
	}

	c.JSON(http.StatusOK, models.Response{
		Code:      0,
		Message:   "获取任务成功",
		Data:      taskInfos,
		Timestamp: getCurrentTimestamp(),
	})
}

// GetOverdueTasks godoc
// @Summary 获取超期任务
// @Description 获取所有超期的任务
// @Tags 任务管理
// @Accept json
// @Produce json
// @Success 200 {object} models.Response{data=[]models.TaskInfo}
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/tasks/overdue [get]
// @Security BearerAuth
func (h *TaskHandler) GetOverdueTasks(c *gin.Context) {
	userID := getUserIDFromContext(c)

	tasks, err := h.taskService.GetOverdueTasks(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:      http.StatusInternalServerError,
			Message:   err.Error(),
			Timestamp: getCurrentTimestamp(),
		})
		return
	}

	// 转换为TaskInfo
	taskInfos := make([]models.TaskInfo, len(tasks))
	for i, task := range tasks {
		taskInfos[i] = *convertTaskToInfo(task)
	}

	c.JSON(http.StatusOK, models.Response{
		Code:      0,
		Message:   "获取超期任务成功",
		Data:      taskInfos,
		Timestamp: getCurrentTimestamp(),
	})
}

// GetFailedTasks godoc
// @Summary 获取失败任务
// @Description 获取所有失败的任务
// @Tags 任务管理
// @Accept json
// @Produce json
// @Success 200 {object} models.Response{data=[]models.TaskInfo}
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/tasks/failed [get]
// @Security BearerAuth
func (h *TaskHandler) GetFailedTasks(c *gin.Context) {
	userID := getUserIDFromContext(c)

	tasks, err := h.taskService.GetFailedTasks(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:      http.StatusInternalServerError,
			Message:   err.Error(),
			Timestamp: getCurrentTimestamp(),
		})
		return
	}

	// 转换为TaskInfo
	taskInfos := make([]models.TaskInfo, len(tasks))
	for i, task := range tasks {
		taskInfos[i] = *convertTaskToInfo(task)
	}

	c.JSON(http.StatusOK, models.Response{
		Code:      0,
		Message:   "获取失败任务成功",
		Data:      taskInfos,
		Timestamp: getCurrentTimestamp(),
	})
}

// 辅助函数
func convertTaskToInfo(task *models.Task) *models.TaskInfo {
	if task == nil {
		return nil
	}

	return &models.TaskInfo{
		ID:           task.ID,
		ProjectID:    task.ProjectID,
		UserID:       task.UserID,
		Name:         task.Name,
		Description:  task.Description,
		Status:       string(task.Status),
		Priority:     int(task.Priority),
		Dependencies: task.Dependencies,
		MaxRetries:   task.MaxRetries,
		RetryCount:   task.RetryCount,
		RetryDelay:   task.RetryDelay,
		StartedAt:    task.StartedAt,
		CompletedAt:  task.CompletedAt,
		Deadline:     task.Deadline,
		Result:       task.Result,
		ErrorMessage: task.ErrorMessage,
		Metadata:     task.Metadata,
		Tags:         task.Tags,
		CreatedAt:    task.CreatedAt,
		UpdatedAt:    task.UpdatedAt,
	}
}

// getCurrentTimestamp 获取当前时间戳
func getCurrentTimestamp() string {
	return time.Now().Format(time.RFC3339)
}

// getUserIDFromContext 从上下文中获取用户ID
func getUserIDFromContext(c *gin.Context) string {
	// TODO: 从JWT token中获取用户ID
	// 暂时返回一个默认值，实际应该从认证中间件中获取
	return "default-user-id"
}
