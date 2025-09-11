package handlers

import (
	"autocodeweb-backend/internal/config"
	"autocodeweb-backend/internal/models"
	"autocodeweb-backend/internal/utils"
	"autocodeweb-backend/pkg/logger"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
)

// TaskHandler 任务处理器
type TaskHandler struct {
	clientOpt *asynq.RedisClientOpt
	inspector *asynq.Inspector
}

// NewTaskHandler 创建任务处理器实例
func NewTaskHandler(cfg *config.Config) *TaskHandler {
	clientOpt := &asynq.RedisClientOpt{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	}
	inspector := asynq.NewInspector(clientOpt)
	return &TaskHandler{
		clientOpt: clientOpt,
		inspector: inspector,
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
// @Success 200 {object} models.Response "成功响应"
// @Failure 404 {object} models.ErrorResponse "任务不存在"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /api/v1/tasks/{id} [get]
func (s *TaskHandler) GetTaskStatus(c *gin.Context) {
	taskID := c.Param("id")

	// 查询任务信息
	info, err := s.inspector.GetTaskInfo("default", taskID) // "default" 是队列名
	if err != nil {
		completedTasks, err := s.inspector.ListCompletedTasks("default")
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Code:      models.INTERNAL_ERROR,
				Message:   "获取任务状态失败: " + err.Error(),
				Timestamp: utils.GetCurrentTime(),
			})
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
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Code:      models.NOT_FOUND,
			Message:   "任务不存在, " + err.Error(),
			Timestamp: utils.GetCurrentTime(),
		})
		return
	}

	taskResult := models.TaskResult{
		TaskID:   taskID,
		Status:   models.TaskStatusInProgress,
		Progress: 0,
		Message:  "任务执行中",
	}
	if info.Result == nil {
		c.JSON(http.StatusOK, models.Response{
			Code:      models.SUCCESS_CODE,
			Message:   "获取任务状态成功",
			Data:      taskResult,
			Timestamp: utils.GetCurrentTime(),
		})
		return
	}

	if len(info.Result) == 0 {
		c.JSON(http.StatusOK, models.Response{
			Code:      models.SUCCESS_CODE,
			Message:   "获取任务状态成功",
			Data:      taskResult,
			Timestamp: utils.GetCurrentTime(),
		})
		return
	}

	err = json.Unmarshal(info.Result, &taskResult)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:      models.INTERNAL_ERROR,
			Message:   "解析任务结果失败: " + err.Error(),
			Timestamp: utils.GetCurrentTime(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code:      models.SUCCESS_CODE,
		Message:   "获取任务状态成功",
		Data:      taskResult,
		Timestamp: utils.GetCurrentTime(),
	})
}
