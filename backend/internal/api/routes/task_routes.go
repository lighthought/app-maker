package routes

import (
	"autocodeweb-backend/internal/api/handlers"

	"github.com/gin-gonic/gin"
)

// RegisterTaskRoutes 注册任务相关路由
func RegisterTaskRoutes(router *gin.RouterGroup, taskHandler *handlers.TaskHandler, authMiddleware gin.HandlerFunc) {
	// 任务路由组 - 独立的任务管理
	tasks := router.Group("/tasks")
	tasks.Use(authMiddleware) // 应用认证中间件
	{
		tasks.GET("/project/:projectId", taskHandler.GetProjectTasks)
		tasks.GET("/:taskId", taskHandler.GetTaskDetails)
		tasks.GET("/:taskId/logs", taskHandler.GetTaskLogs)
		tasks.POST("/:taskId/cancel", taskHandler.CancelTask)
	}
}
