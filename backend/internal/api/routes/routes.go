package routes

import (
	"autocodeweb-backend/internal/api/handlers"
	"autocodeweb-backend/internal/config"
	"autocodeweb-backend/internal/repositories"
	"autocodeweb-backend/internal/services"
	"autocodeweb-backend/pkg/cache"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Register 注册所有路由
func Register(engine *gin.Engine, cfg *config.Config, cacheInstance cache.Cache, monitor *cache.Monitor, db *gorm.DB) {
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

		// 初始化用户相关依赖
		userRepo := repositories.NewUserRepository(db)
		userSessionRepo := repositories.NewUserSessionRepository(db)
		userService := services.NewUserService(userRepo, userSessionRepo, cfg.JWT.SecretKey, cfg.JWT.Expire)
		userHandler := handlers.NewUserHandler(userService)

		// 注册用户路由
		RegisterUserRoutes(v1, userHandler)

		// 初始化项目和标签相关依赖
		projectRepo := repositories.NewProjectRepository(db)
		tagRepo := repositories.NewTagRepository(db)

		// 初始化项目模板服务
		templateService := services.NewProjectTemplateService("./data/template.zip")
		projectService := services.NewProjectService(projectRepo, tagRepo, templateService)

		tagService := services.NewTagService(tagRepo)
		projectHandler := handlers.NewProjectHandler(projectService, tagService)
		tagHandler := handlers.NewTagHandler(tagService)

		// 注册项目和标签路由
		RegisterProjectRoutes(v1, projectHandler, tagHandler)

		// 初始化任务相关依赖
		taskRepo := repositories.NewTaskRepository(db)
		taskLogRepo := repositories.NewTaskLogRepository(db)
		taskService := services.NewTaskService(taskRepo, taskLogRepo, projectRepo)
		taskHandler := handlers.NewTaskHandler(taskService)

		// 注册任务路由
		RegisterTaskRoutes(v1, taskHandler, nil) // TODO: 添加认证中间件
	}
}
