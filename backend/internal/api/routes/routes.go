package routes

import (
	"autocodeweb-backend/internal/api/handlers"
	"autocodeweb-backend/internal/api/middleware"
	"autocodeweb-backend/internal/container"
	"shared-models/common"
	"shared-models/logger"

	"github.com/gin-gonic/gin"
)

// Register 注册所有路由
func Register(engine *gin.Engine, container *container.Container) {
	if container == nil {
		logger.Fatal("container is nil, unable to register routes")
		return
	}

	// 创建 JWT 服务
	var jwtService = container.JWTService
	if jwtService == nil {
		logger.Fatal("jwtService is nil, unable to register routes")
		return
	}
	// 创建认证中间件
	authMiddleware := middleware.AuthMiddleware(jwtService)

	// WebSocket 路由 - 直接注册到引擎上，不使用 API 路由组
	var webSocketHandler = container.WebSocketHandler
	if webSocketHandler != nil {
		// WebSocket 连接路由 - Token 通过查询参数传递，在 Handler 内部验证
		engine.GET("/ws/project/:guid", webSocketHandler.WebSocketUpgrade)
	}

	// API v1 路由组
	routers := engine.Group(common.DefaultApiPrefix)
	{
		// 0.健康检查
		routers.GET("/health", handlers.HealthCheck)

		// 预览路由（无需认证）
		var projectHandler = container.ProjectHandler
		if projectHandler != nil {
			routers.GET("/preview/:token", projectHandler.GetPreviewByToken)
		}

		// 1.缓存相关路由
		var cacheHandler = container.CacheHandler
		cache := routers.Group("/cache")
		{
			if cacheHandler != nil {
				cache.GET("/health", cacheHandler.HealthCheck)
				cache.GET("/stats", cacheHandler.GetStats)
				cache.GET("/memory", cacheHandler.GetMemoryUsage)
				cache.GET("/keyspace", cacheHandler.GetKeyspaceStats)
				cache.GET("/performance", cacheHandler.GetPerformanceMetrics)
			} else {
				cache.GET("/health", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "Cache health endpoint - TODO"})
				})
				cache.GET("/stats", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "Cache stats endpoint - TODO"})
				})
				cache.GET("/memory", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "Cache memory endpoint - TODO"})
				})
				cache.GET("/keyspace", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "Cache keyspace endpoint - TODO"})
				})
				cache.GET("/performance", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "Cache performance endpoint - TODO"})
				})
			}
		}

		// 初始化用户相关依赖
		var userHandler = container.UserHandler
		// 2.认证相关路由（无需认证）
		auth := routers.Group("/auth")
		{
			if userHandler != nil {
				auth.POST("/register", userHandler.Register)
				auth.POST("/login", userHandler.Login)
				auth.POST("/refresh", userHandler.RefreshToken)
			} else {
				auth.POST("/register", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "User register endpoint - TODO"})
				})
				auth.POST("/login", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "User login endpoint - TODO"})
				})
				auth.POST("/refresh", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "User refresh token endpoint - TODO"})
				})
			}
		}

		// 3.用户相关路由（需要认证）
		users := routers.Group("/users")
		users.Use(authMiddleware) // 应用认证中间件
		{
			if userHandler != nil {
				// 用户档案管理
				users.GET("/profile", userHandler.GetUserProfile)
				users.PUT("/profile", userHandler.UpdateUserProfile)
				users.POST("/change-password", userHandler.ChangePassword)
				users.POST("/logout", userHandler.Logout)

				// 用户设置管理
				users.GET("/settings", userHandler.GetUserSettings)
				users.PUT("/settings", userHandler.UpdateUserSettings)

				// 管理员功能
				users.GET("", userHandler.GetUserList)
				users.DELETE("/:user_id", userHandler.DeleteUser)
			} else {
				users.GET("/profile", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "User profile endpoint - TODO"})
				})
				users.PUT("/profile", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "User update profile endpoint - TODO"})
				})
				users.POST("/change-password", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "User change password endpoint - TODO"})
				})
				users.POST("/logout", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "User logout endpoint - TODO"})
				})
				users.GET("", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "User list endpoint - TODO"})
				})
				users.DELETE("/:user_id", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "User delete endpoint - TODO"})
				})
			}
		}

		// 4.项目路由
		projects := routers.Group("/projects")
		projects.Use(authMiddleware) // 应用认证中间件
		{
			var epicHandler = container.EpicHandler
			if projectHandler != nil {
				projects.POST("/", projectHandler.CreateProject)                         // 创建项目
				projects.GET("/", projectHandler.ListProjects)                           // 获取项目列表
				projects.GET("/:guid", projectHandler.GetProject)                        // 获取项目详情
				projects.PUT("/:guid", projectHandler.UpdateProject)                     // 更新项目
				projects.DELETE("/:guid", projectHandler.DeleteProject)                  // 删除项目
				projects.GET("/:guid/stages", projectHandler.GetProjectStages)           // 获取项目开发阶段
				projects.GET("/download/:guid", projectHandler.DownloadProject)          // 下载项目文件
				projects.POST("/:guid/deploy", projectHandler.DeployProject)             // 部署项目
				projects.POST("/:guid/preview-link", projectHandler.GeneratePreviewLink) // 生成预览分享链接

				// Epic 相关路由
				if epicHandler != nil {
					projects.GET("/:guid/epics", epicHandler.GetProjectEpics)        // 获取项目 Epics
					projects.GET("/:guid/mvp-epics", epicHandler.GetProjectMvpEpics) // 获取项目 MVP Epics

					// Epic 编辑相关接口
					projects.PUT("/:guid/epics/:epicId/order", epicHandler.UpdateEpicOrder) // 更新 Epic 排序
					projects.PUT("/:guid/epics/:epicId", epicHandler.UpdateEpic)            // 更新 Epic 内容
					projects.DELETE("/:guid/epics/:epicId", epicHandler.DeleteEpic)         // 删除 Epic

					// Story 编辑相关接口
					projects.PUT("/:guid/epics/:epicId/stories/:storyId/order", epicHandler.UpdateStoryOrder) // 更新 Story 排序
					projects.PUT("/:guid/epics/:epicId/stories/:storyId", epicHandler.UpdateStory)            // 更新 Story 内容
					projects.DELETE("/:guid/epics/:epicId/stories/:storyId", epicHandler.DeleteStory)         // 删除 Story
					projects.DELETE("/:guid/epics/stories/batch-delete", epicHandler.BatchDeleteStories)      // 批量删除 Stories

					// 确认接口
					projects.POST("/:guid/epics/confirm", epicHandler.ConfirmEpicsAndStories) // 确认 Epics 和 Stories
				} else {
					projects.GET("/:guid/epics", func(c *gin.Context) {
						c.JSON(200, gin.H{"message": "Project epics endpoint - TODO"})
					})
					projects.GET("/:guid/mvp-epics", func(c *gin.Context) {
						c.JSON(200, gin.H{"message": "Project mvp epics endpoint - TODO"})
					})
				}
			} else {
				projects.POST("/", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "Project create endpoint - TODO"})
				})
				projects.GET("/", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "Project list endpoint - TODO"})
				})
				projects.GET("/:guid", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "Project get endpoint - TODO"})
				})
				projects.PUT("/:guid", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "Project update endpoint - TODO"})
				})
				projects.DELETE("/:guid", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "Project delete endpoint - TODO"})
				})
				projects.GET("/:guid/stages", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "Project stages endpoint - TODO"})
				})
				projects.GET("/download/:guid", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "Project download endpoint - TODO"})
				})
				projects.POST("/:guid/deploy", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "Project deploy endpoint - TODO"})
				})
				projects.POST("/:guid/preview-link", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "Project preview link endpoint - TODO"})
				})
			}
		}

		var fileHandler = container.FileHandler
		// 5.文件路由
		files := routers.Group("/files")
		files.Use(authMiddleware) // 应用认证中间件
		{
			if fileHandler != nil {
				files.GET("/download", fileHandler.DownloadFile)            // 下载项目文件
				files.GET("/files/:guid", fileHandler.GetProjectFiles)      // 获取文件列表
				files.GET("/filecontent/:guid", fileHandler.GetFileContent) // 获取文件内容
			} else {
				files.GET("/download", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "File download endpoint - TODO"})
				})
				files.GET("/files/:projectId", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "File list endpoint - TODO"})
				})
				files.GET("/filecontent/:projectId", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "File content endpoint - TODO"})
				})
			}
		}

		// 初始化对话相关依赖
		var chatHandler = container.ChatHandler

		// 6.对话路由
		conversations := routers.Group("/chat")
		conversations.Use(authMiddleware) // 应用认证中间件
		{
			if chatHandler != nil {
				conversations.GET("/messages/:guid", chatHandler.GetProjectMessages)       // 获取对话历史
				conversations.POST("/chat/:guid", chatHandler.AddChatMessage)              // 添加对话消息
				conversations.POST("/send-to-agent/:guid", chatHandler.SendMessageToAgent) // 向指定 Agent 发送消息
			} else {
				conversations.GET("/messages/:guid", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "Chat messages endpoint - TODO"})
				})
				conversations.POST("/chat/:guid", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "Chat add message endpoint - TODO"})
				})
				conversations.POST("/send-to-agent/:guid", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "Send message to agent endpoint - TODO"})
				})
			}
		}

		var taskHandler = container.TaskHandler
		// 7.任务路由
		tasks := routers.Group("/tasks")
		tasks.Use(authMiddleware) // 应用认证中间件
		{
			if taskHandler != nil {
				tasks.GET("/:id", taskHandler.GetTaskStatus)    // 获取任务状
				tasks.POST("/:id/retry", taskHandler.RetryTask) // 重试任务
			} else {
				tasks.GET("/:id", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "Task status endpoint - TODO"})
				})

				tasks.POST("/:id/retry", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "Task retry endpoint - TODO"})
				})
			}
		}

		// 8.调试路由
		debug := routers.Group("/debug")
		debug.Use(authMiddleware) // 应用认证中间件
		{
			if webSocketHandler != nil {
				debug.GET("/websocket", webSocketHandler.GetWebSocketDebugInfo) // WebSocket 调试信息
			} else {
				debug.GET("/websocket", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "WebSocket debug endpoint - TODO"})
				})
			}
		}
	}
}
