package routes

import (
	"fmt"
	"time"

	"autocodeweb-backend/internal/api/handlers"
	"autocodeweb-backend/internal/api/middleware"
	"autocodeweb-backend/internal/config"
	"autocodeweb-backend/internal/repositories"
	"autocodeweb-backend/internal/services"
	"autocodeweb-backend/pkg/auth"
	"autocodeweb-backend/pkg/cache"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"gorm.io/gorm"
)

// Register 注册所有路由
func Register(engine *gin.Engine, cfg *config.Config, cacheInstance cache.Cache, monitor *cache.Monitor, db *gorm.DB) {
	// 创建 JWT 服务
	jwtService := auth.NewJWTService(cfg.JWT.SecretKey, time.Duration(cfg.JWT.Expire)*time.Hour)

	// 创建认证中间件
	authMiddleware := middleware.AuthMiddleware(jwtService)

	// API v1 路由组
	routers := engine.Group("/api/v1")
	{
		// 0.健康检查
		routers.GET("/health", handlers.HealthCheck)

		// 1.缓存相关路由
		cacheHandler := handlers.NewCacheHandler(cacheInstance, monitor)
		cache := routers.Group("/cache")
		{
			cache.GET("/health", cacheHandler.HealthCheck)
			cache.GET("/stats", cacheHandler.GetStats)
			cache.GET("/memory", cacheHandler.GetMemoryUsage)
			cache.GET("/keyspace", cacheHandler.GetKeyspaceStats)
			cache.GET("/performance", cacheHandler.GetPerformanceMetrics)
		}

		// 初始化用户相关依赖
		userRepo := repositories.NewUserRepository(db)
		userService := services.NewUserService(userRepo, cfg.JWT.SecretKey, cfg.JWT.Expire)
		userHandler := handlers.NewUserHandler(userService)

		// 2.认证相关路由（无需认证）
		auth := routers.Group("/auth")
		{
			auth.POST("/register", userHandler.Register)
			auth.POST("/login", userHandler.Login)
			auth.POST("/refresh", userHandler.RefreshToken)
		}

		// 3.用户相关路由（需要认证）
		users := routers.Group("/users")
		users.Use(authMiddleware) // 应用认证中间件
		{
			// 用户档案管理
			users.GET("/profile", userHandler.GetUserProfile)
			users.PUT("/profile", userHandler.UpdateUserProfile)
			users.POST("/change-password", userHandler.ChangePassword)
			users.POST("/logout", userHandler.Logout)

			// 管理员功能
			users.GET("", userHandler.GetUserList)
			users.DELETE("/:user_id", userHandler.DeleteUser)
		}

		// 重新创建 projectService
		asyncClient := asynq.NewClient(asynq.RedisClientOpt{
			Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
		})
		fileService := services.NewFileService(asyncClient)
		projectService := services.NewProjectService(db, asyncClient, fileService, cfg)
		projectHandler := handlers.NewProjectHandler(projectService)

		// 4.项目路由
		projects := routers.Group("/projects")
		projects.Use(authMiddleware) // 应用认证中间件
		{
			projects.POST("/", projectHandler.CreateProject)                     // 创建项目
			projects.GET("/", projectHandler.ListProjects)                       // 获取项目列表
			projects.GET("/:id", projectHandler.GetProject)                      // 获取项目详情
			projects.DELETE("/:id", projectHandler.DeleteProject)                // 删除项目
			projects.GET("/:id/stages", projectHandler.GetProjectStages)         // 获取项目开发阶段
			projects.GET("/download/:projectId", projectHandler.DownloadProject) // 下载项目文件
		}

		fileHandler := handlers.NewFileHandler(fileService, projectService)
		// 5.文件路由
		files := routers.Group("/files")
		files.Use(authMiddleware) // 应用认证中间件
		{
			files.GET("/download", fileHandler.DownloadFile)                 // 下载项目文件
			files.GET("/files/:projectId", fileHandler.GetProjectFiles)      // 获取文件列表
			files.GET("/filecontent/:projectId", fileHandler.GetFileContent) // 获取文件内容
		}

		// 初始化对话相关依赖
		messageService := services.NewMessageService(db)
		chatHandler := handlers.NewChatHandler(messageService, fileService)

		// 6.对话路由
		conversations := routers.Group("/chat")
		conversations.Use(authMiddleware) // 应用认证中间件
		{
			conversations.GET("/messages/:projectId", chatHandler.GetProjectMessages) // 获取对话历史
			conversations.POST("/chat/:projectId", chatHandler.AddChatMessage)        // 添加对话消息
		}

		taskHandler := handlers.NewTaskHandler(cfg)
		// 7.任务路由
		tasks := routers.Group("/tasks")
		tasks.Use(authMiddleware) // 应用认证中间件
		{
			tasks.GET("/:id", taskHandler.GetTaskStatus) // 获取任务状
		}
	}
}
