package container

import (
	"context"
	"fmt"

	"github.com/lighthought/app-maker/backend/internal/api/handlers"
	"github.com/lighthought/app-maker/backend/internal/config"
	"github.com/lighthought/app-maker/backend/internal/repositories"
	"github.com/lighthought/app-maker/backend/internal/services"
	"github.com/lighthought/app-maker/backend/internal/worker"

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

func (c *Container) initExternalService(cfg *config.Config, redis *redis.Client, asyncOpt *asynq.RedisClientOpt) {
	// 初始化缓存系统
	var cacheInstance cache.Cache
	var err error
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
	c.CacheInstance = cacheInstance

	c.AsyncClient = asynq.NewClient(asyncOpt)
	c.AsyncInspector = asynq.NewInspector(asyncOpt)
	c.JWTService = auth.NewJWTService(cfg.JWT.SecretKey, time.Duration(cfg.JWT.Expire)*time.Hour)
	c.CachMonitor = cache.NewMonitor(redis)
}

// 初始化 repositories
func (c *Container) initRepositories(db *gorm.DB) {
	c.UserRepository = repositories.NewUserRepository(db)
	c.StageRepository = repositories.NewStageRepository(db)
	c.ProjectRepository = repositories.NewProjectRepository(db)
	c.MessageRepository = repositories.NewMessageRepository(db)
	c.PreviewTokenRepository = repositories.NewPreviewTokenRepository(db)
	c.EpicRepository = repositories.NewEpicRepository(db)
	c.StoryRepository = repositories.NewStoryRepository(db)
}

func (c *Container) initServices(cfg *config.Config, db *gorm.DB, redis *redis.Client) {
	c.WebSocketService = services.NewWebSocketService(c.AsyncClient,
		c.StageRepository, c.MessageRepository, c.ProjectRepository)
	c.MessageService = services.NewMessageService(c.MessageRepository)

	c.UserService = services.NewUserService(c.UserRepository, c.JWTService, cfg.JWT.Expire)
	c.GitService = services.NewGitService()
	// 如果是本地主机运行，则不用执行，只有容器运行才需要初始化 SSH
	if cfg.App.Environment != common.EnvironmentLocalDebug {
		c.GitService.SetupSSH()
	}

	c.FileService = services.NewFileService(c.AsyncClient, c.GitService)
	c.EnvironmentService = services.NewEnvironmentService(cfg.Agents.URL, db, c.CacheInstance)
	c.ProjectTemplateService = services.NewProjectTemplateService(c.FileService)

	c.ProjectStageService = services.NewProjectStageService(c.ProjectRepository,
		c.StageRepository, c.MessageRepository, c.WebSocketService, c.GitService, c.FileService, c.AsyncClient,
		c.EpicRepository, c.StoryRepository, c.EnvironmentService, cfg.Agents.URL)

	c.ProjectService = services.NewProjectService(c.ProjectRepository, c.MessageRepository, c.StageRepository,
		c.AsyncClient, c.ProjectTemplateService, c.GitService, c.WebSocketService, cfg)

	c.PreviewService = services.NewPreviewService(c.PreviewTokenRepository)
	c.EpicService = services.NewEpicService(c.EpicRepository, c.StoryRepository, c.ProjectRepository, c.FileService)
	c.RedisPubSubService = services.NewRedisPubSubService(redis, c.ProjectStageService)
}

// 初始化 handlers
func (c *Container) initHandlers(cfg *config.Config) {
	c.CacheHandler = handlers.NewCacheHandler(c.CacheInstance, c.CachMonitor)
	c.ChatHandler = handlers.NewChatHandler(c.MessageService, c.FileService, c.ProjectService, c.ProjectStageService)
	c.FileHandler = handlers.NewFileHandler(c.FileService, c.ProjectService)
	c.ProjectHandler = handlers.NewProjectHandler(c.ProjectService, c.ProjectStageService, c.PreviewService)
	c.TaskHandler = handlers.NewTaskHandler(c.AsyncInspector)
	c.UserHandler = handlers.NewUserHandler(c.UserService)
	c.WebSocketHandler = handlers.NewWebSocketHandler(c.WebSocketService, c.ProjectService, c.JWTService)
	c.EpicHandler = handlers.NewEpicHandler(c.EpicService)
	c.HealthHandler = handlers.NewHealthHandler(c.EnvironmentService, c.WebSocketService)
}

// 启动异步服务
func (c *Container) startAsyncServices(cfg *config.Config, asyncOpt *asynq.RedisClientOpt) {
	// 有缓存，才处理异步任务
	if c.CacheInstance != nil {
		projectTaskHandler := worker.NewProjectTaskWorker()
		c.AsyncServer = initAsynqWorker(asyncOpt, cfg.Asynq.Concurrency, projectTaskHandler,
			c.ProjectService, c.ProjectStageService, c.WebSocketService)
	}

	// 启动 WebSocket 服务
	go func() {
		logger.Info("WebSocket 服务启动中...")
		if err := c.WebSocketService.Start(context.Background()); err != nil {
			logger.Error("WebSocket 服务启动失败", logger.String("error", err.Error()))
		}
	}()

	// 启动 Redis Pub/Sub 服务
	go func() {
		logger.Info("Redis Pub/Sub 服务启动中...")
		if err := c.RedisPubSubService.Start(context.Background()); err != nil {
			logger.Error("Redis Pub/Sub 服务启动失败", logger.String("error", err.Error()))
		}
	}()
}

func NewContainer(cfg *config.Config, db *gorm.DB, redis *redis.Client) *Container {
	var container Container
	asyncOpt := asynq.RedisClientOpt{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       common.CacheDbBackendAsynq,
	}

	container.initExternalService(cfg, redis, &asyncOpt)
	container.initRepositories(db)
	container.initServices(cfg, db, redis)
	container.initHandlers(cfg)
	container.startAsyncServices(cfg, &asyncOpt)
	return &container
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
