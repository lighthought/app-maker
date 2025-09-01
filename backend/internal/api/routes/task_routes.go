package routes

import (
	"autocodeweb-backend/internal/api/handlers"

	"github.com/gin-gonic/gin"
)

// RegisterTaskRoutes 注册任务相关路由
func RegisterTaskRoutes(router *gin.RouterGroup, taskHandler *handlers.TaskHandler, authMiddleware gin.HandlerFunc) {
	tasks := router.Group("/tasks")
	tasks.Use(authMiddleware) // 应用认证中间件
	{
		// 基础CRUD操作
		tasks.POST("", taskHandler.CreateTask)
		tasks.GET("", taskHandler.ListTasks)
		tasks.GET("/:id", taskHandler.GetTask)
		tasks.PUT("/:id", taskHandler.UpdateTask)
		tasks.DELETE("/:id", taskHandler.DeleteTask)

		// 状态管理
		tasks.PUT("/:id/status", taskHandler.UpdateTaskStatus)
		tasks.POST("/:id/start", taskHandler.StartTask)
		tasks.POST("/:id/complete", taskHandler.CompleteTask)

		// 重试和回滚
		tasks.POST("/:id/retry", taskHandler.RetryTask)
		tasks.POST("/:id/rollback", taskHandler.RollbackTask)

		// 依赖关系
		tasks.GET("/:id/dependencies", taskHandler.GetTaskDependencies)

		// 任务日志
		tasks.GET("/:id/logs", taskHandler.GetTaskLogs)

		// 按状态查询
		tasks.GET("/status/:status", taskHandler.GetTasksByStatus)

		// 按优先级查询
		tasks.GET("/priority/:priority", taskHandler.GetTasksByPriority)

		// 特殊查询
		tasks.GET("/overdue", taskHandler.GetOverdueTasks)
		tasks.GET("/failed", taskHandler.GetFailedTasks)
	}

	// 项目任务路由 - 使用与项目路由相同的参数名
	projects := router.Group("/projects")
	projects.Use(authMiddleware) // 应用认证中间件
	{
		projects.GET("/:id/tasks", taskHandler.GetTasksByProject)
	}
}
