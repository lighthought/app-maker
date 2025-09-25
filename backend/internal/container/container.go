package container

import (
	"autocodeweb-backend/internal/api/handlers"
	"autocodeweb-backend/internal/config"
	"autocodeweb-backend/internal/models"
	"autocodeweb-backend/internal/repositories"
	"autocodeweb-backend/internal/services"
	"autocodeweb-backend/internal/worker"
	"autocodeweb-backend/pkg/auth"
	"autocodeweb-backend/pkg/cache"
	"autocodeweb-backend/pkg/logger"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Container 依赖注入容器
type Container struct {
	// background Items
	AsyncClient    *asynq.Client
	AsyncInspector *asynq.Inspector
	CachMonitor    *cache.Monitor
	JWTService     *auth.JWTService
	CacheInstance  cache.Cache

	// Repositories
	UserRepository    repositories.UserRepository
	StageRepository   repositories.StageRepository
	ProjectRepository repositories.ProjectRepository
	MessageRepository repositories.MessageRepository

	// Services
	UserService            services.UserService
	ProjectTemplateService services.ProjectTemplateService
	ProjectStageService    services.ProjectStageService
	ProjectService         services.ProjectService
	ProjectNameGenerator   services.ProjectNameGenerator
	MessageService         services.MessageService
	GitService             services.GitService
	FileService            services.FileService
	WebSocketService       services.WebSocketService

	// Handlers
	UserHandler      *handlers.UserHandler
	TaskHandler      *handlers.TaskHandler
	ProjectHandler   *handlers.ProjectHandler
	FileHandler      *handlers.FileHandler
	ChatHandler      *handlers.ChatHandler
	CacheHandler     *handlers.CacheHandler
	WebSocketHandler *handlers.WebSocketHandler
}

func NewContainer(cfg *config.Config, db *gorm.DB, redis *redis.Client) *Container {
	// 初始化缓存系统
	var cacheInstance cache.Cache
	var err error

	redisClientOpt := asynq.RedisClientOpt{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	}

	if redis != nil {
		// 创建缓存配置
		cacheConfig := cache.Config{
			Type:     cache.CacheTypeRedis,
			Host:     cfg.Redis.Host,
			Port:     cfg.Redis.Port,
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
			PoolSize: 10,
			MinIdle:  5,
		}

		// 创建缓存实例
		if cacheInstance, err = cache.NewCache(cacheConfig); err != nil {
			logger.Warn("创建缓存实例失败，将使用内存缓存", logger.String("error", err.Error()))
		} else {
			logger.Info("缓存系统初始化成功")
		}
	}

	// asynq items
	asyncClient := asynq.NewClient(redisClientOpt)
	asyncRedisClientOpt := &asynq.RedisClientOpt{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	}
	asyncInspector := asynq.NewInspector(asyncRedisClientOpt)

	jwtService := auth.NewJWTService(cfg.JWT.SecretKey, time.Duration(cfg.JWT.Expire)*time.Hour)
	cachMonitor := cache.NewMonitor(redis)

	// repositories
	userRepository := repositories.NewUserRepository(db)
	stageRepository := repositories.NewStageRepository(db)
	projectRepository := repositories.NewProjectRepository(db)
	messageRepository := repositories.NewMessageRepository(db)

	// services
	webSocketService := services.NewWebSocketService(asyncClient, stageRepository, messageRepository, projectRepository)
	messageService := services.NewMessageService(messageRepository)

	userService := services.NewUserService(userRepository, jwtService, cfg.JWT.Expire)
	fileService := services.NewFileService(asyncClient)
	projectTemplateService := services.NewProjectTemplateService(fileService)
	projectStageService := services.NewProjectStageService(projectRepository, stageRepository, messageRepository, webSocketService)
	projectNameGenerator := services.NewProjectNameGenerator()
	gitService := services.NewGitService()
	gitService.SetupSSH()

	projectService := services.NewProjectService(projectRepository, messageRepository, stageRepository,
		asyncClient, projectTemplateService, projectNameGenerator, gitService, webSocketService)

	// 有缓存，才处理异步任务
	if cacheInstance != nil {
		projectTaskHandler := worker.NewProjectTaskWorker()
		initAsynqWorker(&redisClientOpt, cfg.Asynq.Concurrency, projectTaskHandler, projectService, projectStageService, webSocketService)
	}

	// 启动 WebSocket 服务
	go func() {
		logger.Info("WebSocket 服务启动中...")
		if err := webSocketService.Start(context.Background()); err != nil {
			logger.Error("WebSocket 服务启动失败", logger.String("error", err.Error()))
		}
	}()

	// handlers
	cacheHandler := handlers.NewCacheHandler(cacheInstance, cachMonitor)
	chatHandler := handlers.NewChatHandler(messageService, fileService)
	fileHandler := handlers.NewFileHandler(fileService, projectService)
	projectHandler := handlers.NewProjectHandler(projectService, projectStageService)
	taskHandler := handlers.NewTaskHandler(asyncInspector)
	userHandler := handlers.NewUserHandler(userService)
	webSocketHandler := handlers.NewWebSocketHandler(webSocketService, projectService, jwtService)
	return &Container{
		AsyncClient:            asyncClient,
		JWTService:             jwtService,
		CachMonitor:            cachMonitor,
		AsyncInspector:         asyncInspector,
		CacheInstance:          cacheInstance,
		UserRepository:         userRepository,
		StageRepository:        stageRepository,
		ProjectRepository:      projectRepository,
		MessageRepository:      messageRepository,
		UserService:            userService,
		FileService:            fileService,
		ProjectTemplateService: projectTemplateService,
		ProjectStageService:    projectStageService,
		ProjectService:         projectService,
		ProjectNameGenerator:   projectNameGenerator,
		MessageService:         messageService,
		GitService:             gitService,
		WebSocketService:       webSocketService,
		CacheHandler:           cacheHandler,
		ChatHandler:            chatHandler,
		FileHandler:            fileHandler,
		ProjectHandler:         projectHandler,
		TaskHandler:            taskHandler,
		UserHandler:            userHandler,
		WebSocketHandler:       webSocketHandler,
	}
}

// 初始化异步服务
func initAsynqWorker(redisClientOpt *asynq.RedisClientOpt, concurrency int,
	projectTaskHandler *worker.ProjectTaskHandler,
	projectService services.ProjectService,
	projectStageService services.ProjectStageService,
	webSocketService services.WebSocketService) {
	// 配置 Worker
	server := asynq.NewServer(
		redisClientOpt,
		asynq.Config{
			Concurrency: concurrency, // 并发 worker 数量
			// 可以按权重指定优先处理哪些队列
			Queues: map[string]int{
				"critical": models.TaskQueueCritical,
				"default":  models.TaskQueueDefault,
				"low":      models.TaskQueueLow,
			},
		},
	)

	// 注册任务处理器
	mux := asynq.NewServeMux()
	mux.Handle(models.TypeProjectDownload, projectTaskHandler)
	mux.Handle(models.TypeProjectBackup, projectTaskHandler)
	mux.Handle(models.TypeProjectInit, projectService)
	mux.Handle(models.TypeProjectDevelopment, projectStageService)
	mux.Handle(models.TypeWebSocketBroadcast, webSocketService)
	// ... 注册其他任务处理器

	// 启动服务器
	go func() {
		logger.Info("异步服务启动中... ")
		// 启动 Worker
		if err := server.Run(mux); err != nil {
			log.Fatal("Could not start worker: ", err)
		}
	}()
}
