package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"github.com/lighthought/app-maker/shared-models/common"
	"github.com/lighthought/app-maker/shared-models/logger"
	"github.com/lighthought/app-maker/shared-models/tasks"
	"github.com/lighthought/app-maker/shared-models/utils"

	"github.com/lighthought/app-maker/backend/internal/models"
	"github.com/lighthought/app-maker/backend/internal/services"
)

// TaskHandler 任务处理器
type TaskHandler struct {
	inspector  *asynq.Inspector
	devService services.ProjectDevService
}

// NewTaskHandler 创建任务处理器实例
func NewTaskHandler(inspector *asynq.Inspector, devService services.ProjectDevService) *TaskHandler {
	if inspector == nil {
		logger.Error("inspector is nil!")
		return nil
	}
	return &TaskHandler{
		inspector:  inspector,
		devService: devService,
	}
}

// GetTaskStatus godoc
// @Summary 获取任务状态
// @Description 获取任务状态
// @Tags Task
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "任务ID"
// @Success 200 {object} common.Response "成功响应"
// @Failure 404 {object} common.ErrorResponse "任务不存在"
// @Failure 500 {object} common.ErrorResponse "服务器内部错误"
// @Router /api/v1/tasks/{id} [get]
func (s *TaskHandler) GetTaskStatus(c *gin.Context) {
	taskID := c.Param("id")

	// 查询任务信息
	info, err := s.inspector.GetTaskInfo("default", taskID) // "default" 是队列名
	if err != nil {
		completedTasks, err := s.inspector.ListCompletedTasks("default")
		if err != nil {
			c.JSON(http.StatusOK, utils.GetErrorResponse(common.INTERNAL_ERROR, "获取任务状态失败: "+err.Error()))
			return
		}

		for _, task := range completedTasks {
			if task.ID == taskID {
				logger.Info("任务已完成", logger.String("taskID", taskID))
				info = task
				break
			}
		}
	}

	if info == nil {
		c.JSON(http.StatusNotFound, utils.GetErrorResponse(common.NOT_FOUND, "任务不存在, "+err.Error()))
		return
	}

	taskResult := tasks.TaskResult{
		TaskID:   taskID,
		Status:   common.CommonStatusInProgress,
		Progress: 0,
		Message:  "任务执行中",
	}
	if info.Result == nil {
		c.JSON(http.StatusOK, utils.GetSuccessResponse("获取任务状态成功", taskResult))
		return
	}

	if len(info.Result) == 0 {
		c.JSON(http.StatusOK, utils.GetSuccessResponse("获取任务状态成功", taskResult))
		return
	}

	err = json.Unmarshal(info.Result, &taskResult)
	if err != nil {
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.INTERNAL_ERROR, "解析任务结果失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.GetSuccessResponse("获取任务状态成功", taskResult))
}

// 重试任务
// @Summary 重试任务
// @Description 重试任务
// @Tags Task
// @Accept json
// @Produce json
// @Security Bearer
// @Param project body models.RetryTaskRequest true "任务重试请求"
// @Success 200 {object} common.Response "成功响应"
// @Failure 500 {object} common.ErrorResponse "服务器内部错误"
// @Router /api/v1/tasks/retry [post]
func (s *TaskHandler) RetryTask(c *gin.Context) {
	var req models.RetryTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("请求参数绑定失败",
			logger.String("error", err.Error()),
			logger.String("requestBody", fmt.Sprintf("%v", c.Request.Body)),
		)
		c.JSON(http.StatusBadRequest, utils.GetErrorResponse(common.VALIDATION_ERROR, "请求参数错误, "+err.Error()))
		return
	}

	err := s.inspector.RunTask("default", req.TaskID)
	if err != nil {
		if err.Error() == "asynq: task not found" {
			if req.StageID != "" {
				if err2 := s.devService.RenewCurrentStageTask(c.Request.Context(), req.StageID); err2 == nil {
					c.JSON(http.StatusOK, utils.GetSuccessResponse("重新创建阶段任务成功", nil))
					return
				}
			}
		}

		c.JSON(http.StatusInternalServerError, utils.GetErrorResponse(common.INTERNAL_ERROR, "重试任务失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.GetSuccessResponse("重试任务成功", nil))
}
