package routes

import (
	"autocodeweb-backend/internal/api/handlers"

	"github.com/gin-gonic/gin"
)

// RegisterProjectRoutes 注册项目相关路由
func RegisterProjectRoutes(router *gin.RouterGroup, projectHandler *handlers.ProjectHandler, tagHandler *handlers.TagHandler, authMiddleware gin.HandlerFunc) {
	// 项目路由组
	projects := router.Group("/projects")
	projects.Use(authMiddleware) // 应用认证中间件
	{
		// 项目CRUD操作
		projects.POST("/", projectHandler.CreateProject)      // 创建项目
		projects.GET("/", projectHandler.ListProjects)        // 获取项目列表
		projects.GET("/:id", projectHandler.GetProject)       // 获取项目详情
		projects.PUT("/:id", projectHandler.UpdateProject)    // 更新项目
		projects.DELETE("/:id", projectHandler.DeleteProject) // 删除项目

		// 项目状态管理
		projects.PUT("/:id/status", projectHandler.UpdateProjectStatus) // 更新项目状态

		// 项目标签管理
		projects.GET("/:id/tags", projectHandler.GetProjectTags) // 获取项目标签

		// 端口管理
		projects.GET("/ports/next", projectHandler.GetNextAvailablePorts) // 获取下一个可用端口
	}

	// 标签路由组
	tags := router.Group("/tags")
	tags.Use(authMiddleware) // 应用认证中间件
	{
		// 标签CRUD操作
		tags.POST("/", tagHandler.CreateTag)      // 创建标签
		tags.GET("/", tagHandler.ListTags)        // 获取标签列表
		tags.GET("/:id", tagHandler.GetTag)       // 获取标签详情
		tags.PUT("/:id", tagHandler.UpdateTag)    // 更新标签
		tags.DELETE("/:id", tagHandler.DeleteTag) // 删除标签

		// 标签查询
		tags.GET("/popular", tagHandler.GetPopularTags)   // 获取热门标签
		tags.GET("/project", tagHandler.GetTagsByProject) // 获取项目标签
	}
}
