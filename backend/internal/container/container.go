package container

import (
	"github.com/lighthought/app-maker/backend/internal/api/handlers"
	"github.com/lighthought/app-maker/backend/internal/config"
	"github.com/lighthought/app-maker/backend/internal/repositories"
	"github.com/lighthought/app-maker/backend/internal/services"
	"github.com/lighthought/app-maker/backend/internal/worker"

	"context"
	"fmt"
	"log"
	"time"

	"github.com/lighthought/app-maker/shared-models/auth"
	"github.com/lighthought/app-maker/shared-models/cache"
	"github.com/lighthought/app-maker/shared-models/common"
	"github.com/lighthought/app-maker/shared-models/logger"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Container 依赖注入容器
type Container struct {
	// background Items
	AsyncClient    *asynq.Client
	AsyncInspector *asynq.Inspector
	AsyncServer    *asynq.Server
	CachMonitor    *cache.Monitor
	JWTService     *auth.JWTService
	CacheInstance  cache.Cache

	// Repositories
	UserRepository         repositories.UserRepository
	StageRepository        repositories.StageRepository
	ProjectRepository      repositories.ProjectRepository
	MessageRepository      repositories.MessageRepository
	PreviewTokenRepository repositories.PreviewTokenRepository
	EpicRepository         repositories.EpicRepository
	StoryRepository        repositories.StoryRepository

	// Services
	UserService            services.UserService
	ProjectTemplateService services.ProjectTemplateService
	ProjectStageService    services.ProjectStageService
	ProjectService         services.ProjectService
	MessageService         services.MessageService
	GitService             services.GitService
	FileService            services.FileService
	WebSocketService       services.WebSocketService
	PreviewService         services.PreviewService
	EpicService            services.EpicService
	RedisPubSubService     services.RedisPubSubService
	EnvironmentService     services.EnvironmentService

	// Handlers
	UserHandler      *handlers.UserHandler
	TaskHandler      *handlers.TaskHandler
	ProjectHandler   *handlers.ProjectHandler
	FileHandler      *handlers.FileHandler
	ChatHandler      *handlers.ChatHandler
	CacheHandler     *handlers.CacheHandler
	WebSocketHandler *handlers.WebSocketHandler
	EpicHandler      *handlers.EpicHandler
	HealthHandler    *handlers.HealthHandler
}

func NewContainer(cfg *config.Config, db *gorm.DB, redis *redis.Client) *Container {
	// 初始化缓存系统
	var cacheInstance cache.Cache
	var err error

	redisClientOpt := asynq.RedisClientOpt{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       common.CacheDbBackendAsynq,
	}

	if redis != nil {
		// 创建缓存配置
		cacheConfig := cache.Config{
			Host:     cfg.Redis.Host,
			Port:     cfg.Redis.Port,
			Password: cfg.Redis.Password,
			DB:       common.CacheDbDatabase,
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
		DB:       common.CacheDbBackendAsynq,
	}
	asyncInspector := asynq.NewInspector(asyncRedisClientOpt)

	jwtService := auth.NewJWTService(cfg.JWT.SecretKey, time.Duration(cfg.JWT.Expire)*time.Hour)
	cachMonitor := cache.NewMonitor(redis)

	// repositories
	userRepository := repositories.NewUserRepository(db)
	stageRepository := repositories.NewStageRepository(db)
	projectRepository := repositories.NewProjectRepository(db)
	messageRepository := repositories.NewMessageRepository(db)
	previewTokenRepository := repositories.NewPreviewTokenRepository(db)
	epicRepository := repositories.NewEpicRepository(db)
	storyRepository := repositories.NewStoryRepository(db)

	// services
	webSocketService := services.NewWebSocketService(asyncClient,
		stageRepository, messageRepository, projectRepository)
	messageService := services.NewMessageService(messageRepository)

	userService := services.NewUserService(userRepository, jwtService, cfg.JWT.Expire)
	gitService := services.NewGitService()
	// 如果是本地主机运行，则不用执行，只有容器运行才需要初始化 SSH
	if cfg.App.Environment != common.EnvironmentLocalDebug {
		gitService.SetupSSH()
	}

	fileService := services.NewFileService(asyncClient, gitService)
	environmentService := services.NewEnvironmentService(cfg.Agents.URL, db, cacheInstance)
	projectTemplateService := services.NewProjectTemplateService(fileService)

	projectStageService := services.NewProjectStageService(projectRepository,
		stageRepository, messageRepository, webSocketService, gitService, fileService, asyncClient,
		epicRepository, storyRepository, environmentService, cfg.Agents.URL)

	projectService := services.NewProjectService(projectRepository, messageRepository, stageRepository,
		asyncClient, projectTemplateService, gitService, webSocketService, cfg)

	previewService := services.NewPreviewService(previewTokenRepository)
	epicService := services.NewEpicService(epicRepository, storyRepository, projectRepository, fileService)
	redisPubSubService := services.NewRedisPubSubService(redis, projectStageService)

	var asynqServer *asynq.Server
	// 有缓存，才处理异步任务
	if cacheInstance != nil {
		projectTaskHandler := worker.NewProjectTaskWorker()
		asynqServer = initAsynqWorker(&redisClientOpt, cfg.Asynq.Concurrency, projectTaskHandler, projectService, projectStageService, webSocketService)
	}

	// 启动 WebSocket 服务
	go func() {
		logger.Info("WebSocket 服务启动中...")
		if err := webSocketService.Start(context.Background()); err != nil {
			logger.Error("WebSocket 服务启动失败", logger.String("error", err.Error()))
		}
	}()

	// 启动 Redis Pub/Sub 服务
	go func() {
		logger.Info("Redis Pub/Sub 服务启动中...")
		if err := redisPubSubService.Start(context.Background()); err != nil {
			logger.Error("Redis Pub/Sub 服务启动失败", logger.String("error", err.Error()))
		}
	}()

	// handlers
	cacheHandler := handlers.NewCacheHandler(cacheInstance, cachMonitor)
	chatHandler := handlers.NewChatHandler(messageService, fileService, projectService, projectStageService)
	fileHandler := handlers.NewFileHandler(fileService, projectService)
	projectHandler := handlers.NewProjectHandler(projectService, projectStageService, previewService)
	taskHandler := handlers.NewTaskHandler(asyncInspector)
	userHandler := handlers.NewUserHandler(userService)
	webSocketHandler := handlers.NewWebSocketHandler(webSocketService, projectService, jwtService)
	epicHandler := handlers.NewEpicHandler(epicService)
	healthHandler := handlers.NewHealthHandler(environmentService, webSocketService)

	return &Container{
		AsyncClient:            asyncClient,
		AsyncServer:            asynqServer,
		JWTService:             jwtService,
		CachMonitor:            cachMonitor,
		AsyncInspector:         asyncInspector,
		CacheInstance:          cacheInstance,
		UserRepository:         userRepository,
		StageRepository:        stageRepository,
		ProjectRepository:      projectRepository,
		MessageRepository:      messageRepository,
		PreviewTokenRepository: previewTokenRepository,
		EpicRepository:         epicRepository,
		StoryRepository:        storyRepository,
		UserService:            userService,
		FileService:            fileService,
		ProjectTemplateService: projectTemplateService,
		ProjectStageService:    projectStageService,
		ProjectService:         projectService,
		MessageService:         messageService,
		GitService:             gitService,
		WebSocketService:       webSocketService,
		PreviewService:         previewService,
		EpicService:            epicService,
		RedisPubSubService:     redisPubSubService,
		EnvironmentService:     environmentService,
		CacheHandler:           cacheHandler,
		ChatHandler:            chatHandler,
		FileHandler:            fileHandler,
		ProjectHandler:         projectHandler,
		TaskHandler:            taskHandler,
		UserHandler:            userHandler,
		WebSocketHandler:       webSocketHandler,
		EpicHandler:            epicHandler,
		HealthHandler:          healthHandler,
	}
}

// 停止
func (c *Container) Stop() {
	c.WebSocketService.Stop()
	logger.Info("WebSocket 服务已停止")
	c.RedisPubSubService.Stop()
	logger.Info("Redis Pub/Sub 服务已停止")
	c.AsyncInspector.Close()
	logger.Info("AsyncInspector 已关闭")
	c.AsyncClient.Close()
	logger.Info("AsyncClient 已关闭")
	c.AsyncServer.Shutdown()
	logger.Info("AsyncServer 已关闭")
	c.CacheInstance.Close()
	logger.Info("CacheInstance 已关闭")
}

// 初始化异步服务
func initAsynqWorker(redisClientOpt *asynq.RedisClientOpt, concurrency int,
	projectTaskHandler *worker.ProjectTaskHandler,
	projectService services.ProjectService,
	projectStageService services.ProjectStageService,
	webSocketService services.WebSocketService) *asynq.Server {
	// 配置 Worker
	server := asynq.NewServer(
		redisClientOpt,
		asynq.Config{
			Concurrency: concurrency, // 并发 worker 数量
			// 可以按权重指定优先处理哪些队列
			Queues: map[string]int{
				"critical": common.TaskQueueCritical,
				"default":  common.TaskQueueDefault,
				"low":      common.TaskQueueLow,
			},
		},
	)

	// 注册任务处理器
	mux := asynq.NewServeMux()
	mux.Handle(common.TaskTypeProjectDownload, projectTaskHandler)
	mux.Handle(common.TaskTypeProjectBackup, projectTaskHandler)
	mux.Handle(common.TaskTypeProjectInit, projectService)
	mux.Handle(common.TaskTypeProjectDevelopment, projectStageService)
	mux.Handle(common.TaskTypeProjectDeploy, projectStageService)
	mux.Handle(common.TaskTypeAgentChat, projectStageService)
	mux.Handle(common.TaskTypeWebSocketBroadcast, webSocketService)
	// ... 注册其他任务处理器

	// 启动服务器
	go func() {
		logger.Info("异步服务启动中... ")
		// 启动 Worker
		if err := server.Run(mux); err != nil {
			log.Fatal("Could not start worker: ", err)
		}
	}()

	return server
}
