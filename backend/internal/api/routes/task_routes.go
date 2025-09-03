package routes

import (
	"autocodeweb-backend/internal/api/handlers"

	"github.com/gin-gonic/gin"
)

// RegisterTaskRoutes 注册任务相关路由
func RegisterTaskRoutes(router *gin.RouterGroup, taskHandler *handlers.TaskHandler, authMiddleware gin.HandlerFunc) {
	// 项目任务路由 - 使用与项目路由相同的参数名
	projects := router.Group("/projects")
	projects.Use(authMiddleware) // 应用认证中间件
	{
		projects.GET("/:projectId/tasks", taskHandler.GetProjectTasks)
		projects.GET("/:projectId/tasks/:taskId", taskHandler.GetTaskDetails)
		projects.GET("/:projectId/tasks/:taskId/logs", taskHandler.GetTaskLogs)
		projects.POST("/:projectId/tasks/:taskId/cancel", taskHandler.CancelTask)
	}
}
