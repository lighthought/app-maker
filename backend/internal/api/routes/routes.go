package routes

import (
	"autocodeweb-backend/internal/api/handlers"
	"autocodeweb-backend/internal/api/middleware"
	"autocodeweb-backend/internal/config"
	"autocodeweb-backend/pkg/cache"

	"github.com/gin-gonic/gin"
)

// Register 注册所有路由
func Register(engine *gin.Engine, cfg *config.Config, cacheInstance cache.Cache, monitor *cache.Monitor) {
	// API v1 路由组
	v1 := engine.Group("/api/v1")
	{
		// 健康检查
		v1.GET("/health", handlers.HealthCheck)

		// 缓存相关路由
		cacheHandler := handlers.NewCacheHandler(cacheInstance, monitor)
		cache := v1.Group("/cache")
		{
			cache.GET("/health", cacheHandler.HealthCheck)
			cache.GET("/stats", cacheHandler.GetStats)
			cache.GET("/memory", cacheHandler.GetMemoryUsage)
			cache.GET("/keyspace", cacheHandler.GetKeyspaceStats)
			cache.GET("/performance", cacheHandler.GetPerformanceMetrics)
		}

		// 认证相关路由
		auth := v1.Group("/auth")
		{
			auth.POST("/register", handlers.Register)
			auth.POST("/login", handlers.Login)
		}

		// 需要认证的路由
		protected := v1.Group("/")
		protected.Use(middleware.AuthMiddleware(cfg.JWT.SecretKey))
		{
			// 用户相关
			users := protected.Group("/users")
			{
				users.GET("/profile", handlers.GetUserProfile)
				users.PUT("/profile", handlers.UpdateUserProfile)
			}

			// 项目相关
			projects := protected.Group("/projects")
			{
				projects.POST("/", handlers.CreateProject)
				projects.GET("/", handlers.GetProjects)
				projects.GET("/:id", handlers.GetProject)
				projects.PUT("/:id", handlers.UpdateProject)
				projects.DELETE("/:id", handlers.DeleteProject)
			}

			// 任务相关
			tasks := protected.Group("/tasks")
			{
				tasks.GET("/:id/status", handlers.GetTaskStatus)
			}
		}
	}
}
