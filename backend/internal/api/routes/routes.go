package routes

import (
	"github.com/lighthought/app-maker/shared-models/common"
	"github.com/lighthought/app-maker/shared-models/logger"

	"github.com/lighthought/app-maker/backend/internal/api/middleware"
	"github.com/lighthought/app-maker/backend/internal/container"

	"github.com/gin-gonic/gin"
)

// 设置空路由
func setPostEmptyEndpoint(routers *gin.RouterGroup, relativePath string, message string) {
	routers.POST(relativePath, func(c *gin.Context) {
		c.JSON(200, gin.H{"message": message})
	})
}

// 设置空GET路由
func setGetEmptyEndpoint(routers *gin.RouterGroup, relativePath string, message string) {
	routers.GET(relativePath, func(c *gin.Context) {
		c.JSON(200, gin.H{"message": message})
	})
}

// 设置空PUT路由
func setPutEmptyEndpoint(routers *gin.RouterGroup, relativePath string, message string) {
	routers.PUT(relativePath, func(c *gin.Context) {
		c.JSON(200, gin.H{"message": message})
	})
}

// 设置空DELETE路由
func setDeleteEmptyEndpoint(routers *gin.RouterGroup, relativePath string, message string) {
	routers.DELETE(relativePath, func(c *gin.Context) {
		c.JSON(200, gin.H{"message": message})
	})
}

// 注册缓存项API
func registerCacheApiRoutes(routers *gin.RouterGroup, container *container.Container) {
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
			setPostEmptyEndpoint(cache, "/health", "Cache health endpoint - TODO")
			setPostEmptyEndpoint(cache, "/stats", "Cache stats endpoint - TODO")
			setPostEmptyEndpoint(cache, "/memory", "Cache memory endpoint - TODO")
			setPostEmptyEndpoint(cache, "/keyspace", "Cache keyspace endpoint - TODO")
			setPostEmptyEndpoint(cache, "/performance", "Cache performance endpoint - TODO")
		}
	}
}

// 注册认证相关API
func registerAuthApiRoutes(routers *gin.RouterGroup, container *container.Container) {
	var userHandler = container.UserHandler
	// 2.认证相关路由（无需认证）
	auth := routers.Group("/auth")
	{
		if userHandler != nil {
			auth.POST("/register", userHandler.Register)
			auth.POST("/login", userHandler.Login)
			auth.POST("/refresh", userHandler.RefreshToken)
		} else {
			setPostEmptyEndpoint(auth, "/register", "User register endpoint - TODO")
			setPostEmptyEndpoint(auth, "/login", "User login endpoint - TODO")
			setPostEmptyEndpoint(auth, "/refresh", "User refresh token endpoint - TODO")
		}
	}
}

// 注册用户相关API
func registerUserApiRoutes(routers *gin.RouterGroup, authMiddleware gin.HandlerFunc, container *container.Container) {
	users := routers.Group("/users")
	userHandler := container.UserHandler
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
			setGetEmptyEndpoint(users, "/profile", "User profile endpoint - TODO")
			setPutEmptyEndpoint(users, "/profile", "User update profile endpoint - TODO")
			setPostEmptyEndpoint(users, "/change-password", "User change password endpoint - TODO")
			setPostEmptyEndpoint(users, "/logout", "User logout endpoint - TODO")

			setGetEmptyEndpoint(users, "/settings", "User settings endpoint - TODO")
			setPutEmptyEndpoint(users, "/settings", "User update settings endpoint - TODO")

			setGetEmptyEndpoint(users, "/", "User list endpoint - TODO")
			setDeleteEmptyEndpoint(users, "/:user_id", "User delete endpoint - TODO")
		}
	}
}

// 注册项目API路由
func registerProjectApiRoutes(routers *gin.RouterGroup, authMiddleware gin.HandlerFunc, container *container.Container) {
	projects := routers.Group("/projects")
	projectHandler := container.ProjectHandler
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
			setPostEmptyEndpoint(projects, "/", "Project create endpoint - TODO")
			setGetEmptyEndpoint(projects, "/", "Project list endpoint - TODO")
			setGetEmptyEndpoint(projects, "/:guid", "Project get endpoint - TODO")
			setPutEmptyEndpoint(projects, "/:guid", "Project update endpoint - TODO")
			setDeleteEmptyEndpoint(projects, "/:guid", "Project delete endpoint - TODO")
			setGetEmptyEndpoint(projects, "/:guid/stages", "Project stages endpoint - TODO")
			setGetEmptyEndpoint(projects, "/download/:guid", "Project download endpoint - TODO")
			setPostEmptyEndpoint(projects, "/:guid/deploy", "Project deploy endpoint - TODO")
			setPostEmptyEndpoint(projects, "/:guid/preview-link", "Project preview link endpoint - TODO")
		}
	}
}

// 注册文件相关API 路由
func registerFileApiRoutes(routers *gin.RouterGroup, authMiddleware gin.HandlerFunc, container *container.Container) {
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
			setGetEmptyEndpoint(files, "/download", "File download endpoint - TODO")
			setGetEmptyEndpoint(files, "/files/:projectId", "File list endpoint - TODO")
			setGetEmptyEndpoint(files, "/filecontent/:projectId", "File content endpoint - TODO")
		}
	}
}

// 注册对话相关API路由
func registerChatApiRoutes(routers *gin.RouterGroup, authMiddleware gin.HandlerFunc, container *container.Container) {
	// 初始化对话相关依赖
	var chatHandler = container.ChatHandler
	conversations := routers.Group("/chat")
	conversations.Use(authMiddleware) // 应用认证中间件
	{
		if chatHandler != nil {
			conversations.GET("/messages/:guid", chatHandler.GetProjectMessages)       // 获取对话历史
			conversations.POST("/chat/:guid", chatHandler.AddChatMessage)              // 添加对话消息
			conversations.POST("/send-to-agent/:guid", chatHandler.SendMessageToAgent) // 向指定 Agent 发送消息
		} else {
			setGetEmptyEndpoint(conversations, "/messages/:guid", "Chat messages endpoint - TODO")
			setPostEmptyEndpoint(conversations, "/chat/:guid", "Chat add message endpoint - TODO")
			setPostEmptyEndpoint(conversations, "/send-to-agent/:guid", "Send message to agent endpoint - TODO")
		}
	}
}

// 注册任务相关API路由
func registerTaskApiRoutes(routers *gin.RouterGroup, authMiddleware gin.HandlerFunc, container *container.Container) {
	var taskHandler = container.TaskHandler

	tasks := routers.Group("/tasks")
	tasks.Use(authMiddleware) // 应用认证中间件
	{
		if taskHandler != nil {
			tasks.GET("/:id", taskHandler.GetTaskStatus)    // 获取任务状
			tasks.POST("/:id/retry", taskHandler.RetryTask) // 重试任务
		} else {
			setGetEmptyEndpoint(tasks, "/:id", "Task status endpoint - TODO")
			setPostEmptyEndpoint(tasks, "/:id/retry", "Task retry endpoint - TODO")
		}
	}
}

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
		var healthHandler = container.HealthHandler
		if healthHandler != nil {
			routers.GET("/health", healthHandler.HealthCheck)
		} else {
			setPostEmptyEndpoint(routers, "/health", "Health check endpoint - TODO")
		}

		// 预览路由（无需认证）
		var projectHandler = container.ProjectHandler
		if projectHandler != nil {
			routers.GET("/preview/:token", projectHandler.GetPreviewByToken)
		}

		// 1. 缓存相关路由（无需认证）
		registerCacheApiRoutes(routers, container)

		// 2. 认证相关路由（无需认证）
		registerAuthApiRoutes(routers, container)

		// 3.用户相关路由（需要认证）
		registerUserApiRoutes(routers, authMiddleware, container)

		// 4.项目路由
		registerProjectApiRoutes(routers, authMiddleware, container)

		// 5. 文件相关路由
		registerFileApiRoutes(routers, authMiddleware, container)

		// 6. 对话相关路由
		registerChatApiRoutes(routers, authMiddleware, container)

		// 7. 任务相关路由
		registerTaskApiRoutes(routers, authMiddleware, container)

		// 8.调试路由
		debug := routers.Group("/debug")
		debug.Use(authMiddleware) // 应用认证中间件
		{
			if webSocketHandler != nil {
				debug.GET("/websocket", webSocketHandler.GetWebSocketDebugInfo) // WebSocket 调试信息
			} else {
				setGetEmptyEndpoint(debug, "/websocket", "WebSocket debug endpoint - TODO")
			}
		}
	}
}
