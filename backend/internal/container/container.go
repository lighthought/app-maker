package container

import (
	"context"
	"fmt"

	"github.com/lighthought/app-maker/backend/internal/api/handlers"
	"github.com/lighthought/app-maker/backend/internal/config"
	"github.com/lighthought/app-maker/backend/internal/repositories"
	"github.com/lighthought/app-maker/backend/internal/services"
	"github.com/redis/go-redis/v9"

	"log"
	"time"

	"github.com/lighthought/app-maker/shared-models/auth"
	"github.com/lighthought/app-maker/shared-models/cache"
	"github.com/lighthought/app-maker/shared-models/common"
	"github.com/lighthought/app-maker/shared-models/logger"

	"github.com/hibiken/asynq"
	"gorm.io/gorm"
)

// Container 依赖注入容器
type Container struct {
	// background Items
	AsyncClient    *asynq.Client    // asynq异步客户端
	AsyncInspector *asynq.Inspector // asynq异步检查器
	AsyncServer    *asynq.Server    // asynq异步服务器
	CachMonitor    *cache.Monitor   // redis缓存监控器
	JWTService     *auth.JWTService // JWT服务
	CacheInstance  cache.Cache      // redis缓存实例

	// Repositories
	Repositories *repositories.Repository // 仓库集合

	// Services
	UserService            services.UserService            // 用户服务
	ProjectTemplateService services.ProjectTemplateService // 项目模板服务
	ProjectCommonService   services.ProjectCommonService   // 项目通用服务
	ProjectDevService      services.ProjectDevService      // 项目开发服务
	AgentInteractService   services.AgentInteractService   // Agent交互服务
	ProjectService         services.ProjectService         // 项目服务
	MessageService         services.MessageService         // 消息服务
	GitService             services.GitService             // Git服务
	FileService            services.FileService            // 文件服务
	WebSocketService       services.WebSocketService       // WebSocket服务
	PreviewService         services.PreviewService         // 预览服务
	EpicService            services.EpicService            // 史诗服务
	RedisPubSubService     services.RedisPubSubService     // Redis Pub/Sub服务
	EnvironmentService     services.EnvironmentService     // 环境服务
	AsyncClientService     services.AsyncClientService     // 异步客户端服务
	AsyncTaskService       services.AsyncTaskService       // 异步任务处理服务

	// Handlers
	UserHandler      *handlers.UserHandler      // 用户处理器
	TaskHandler      *handlers.TaskHandler      // 任务处理器
	ProjectHandler   *handlers.ProjectHandler   // 项目处理器
	FileHandler      *handlers.FileHandler      // 文件处理器
	ChatHandler      *handlers.ChatHandler      // 聊天处理器
	CacheHandler     *handlers.CacheHandler     // 缓存处理器
	WebSocketHandler *handlers.WebSocketHandler // WebSocket处理器
	EpicHandler      *handlers.EpicHandler      // 史诗处理器
	HealthHandler    *handlers.HealthHandler    // 健康处理器
}

// 初始化外部服务
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
	c.Repositories = repositories.NewRepositories(db)
}

// 初始化 services
func (c *Container) initServices(cfg *config.Config, db *gorm.DB) {
	// 完全不依赖其他服务，也不被其他服务引用
	c.MessageService = services.NewMessageService(c.Repositories.MessageRepo)
	c.PreviewService = services.NewPreviewService(c.Repositories.PreviewTokenRepo)
	c.UserService = services.NewUserService(c.Repositories.UserRepo, c.JWTService, cfg.JWT.Expire)

	// 会被其他服务引用的服务
	gitService := services.NewGitService()
	fileServie := services.NewFileService(gitService)
	asyncClientService := services.NewAsyncClientService(c.AsyncClient)
	agentInteractService := services.NewAgentInteractService(c.Repositories, cfg.Agents.URL)

	environmentService := services.NewEnvironmentService(agentInteractService, db, c.CacheInstance)
	projectTemplateService := services.NewProjectTemplateService(fileServie)

	webSocketService := services.NewWebSocketService(asyncClientService, c.Repositories)
	projectCommonService := services.NewProjectCommonService(c.Repositories,
		webSocketService)
	projectDevService := services.NewProjectDevService(c.Repositories, asyncClientService, agentInteractService, projectCommonService)

	c.GitService = gitService
	c.FileService = fileServie
	c.EnvironmentService = environmentService
	c.ProjectTemplateService = projectTemplateService
	c.AsyncClientService = asyncClientService
	c.AgentInteractService = agentInteractService
	c.WebSocketService = webSocketService
	c.ProjectCommonService = projectCommonService
	c.ProjectDevService = projectDevService

	// 简单引用其他服务
	c.EpicService = services.NewEpicService(c.Repositories, fileServie)
	c.RedisPubSubService = services.NewRedisPubSubService(asyncClientService, cfg)

	// 需要引用多个其他服务的核心业务服务
	c.ProjectService = services.NewProjectService(c.Repositories, projectTemplateService,
		projectCommonService, gitService, asyncClientService, cfg)
	c.AsyncTaskService = services.NewAsyncTaskService(c.Repositories, projectCommonService, projectDevService, agentInteractService)

	// 如果是本地主机运行，则不用执行，只有容器运行才需要初始化 SSH
	if cfg.App.Environment != common.EnvironmentLocalDebug {
		gitService.SetupSSH()
	}

	c.ProjectDevService.InitStageItems()
}

// 初始化 handlers
func (c *Container) initHandlers() {
	c.CacheHandler = handlers.NewCacheHandler(c.CacheInstance, c.CachMonitor)
	c.ChatHandler = handlers.NewChatHandler(c.MessageService, c.FileService, c.ProjectService, c.AsyncClientService)
	c.FileHandler = handlers.NewFileHandler(c.FileService, c.ProjectService)
	c.ProjectHandler = handlers.NewProjectHandler(c.ProjectService, c.AsyncClientService, c.ProjectCommonService, c.PreviewService)
	c.TaskHandler = handlers.NewTaskHandler(c.AsyncInspector)
	c.UserHandler = handlers.NewUserHandler(c.UserService)
	c.WebSocketHandler = handlers.NewWebSocketHandler(c.WebSocketService, c.ProjectService, c.JWTService)
	c.EpicHandler = handlers.NewEpicHandler(c.EpicService)
	c.HealthHandler = handlers.NewHealthHandler(c.EnvironmentService, c.AgentInteractService, c.WebSocketService)
}

// 启动异步服务
func (c *Container) startAsyncServices(cfg *config.Config, asyncOpt *asynq.RedisClientOpt) {
	// 有缓存，才处理异步任务
	if c.CacheInstance != nil {
		c.AsyncServer = initAsynqWorker(asyncOpt, cfg.Asynq.Concurrency,
			c.ProjectService, c.AsyncTaskService, c.WebSocketService)
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
	container.initServices(cfg, db)
	container.initHandlers()
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
func initAsynqWorker(redisClientOpt *asynq.RedisClientOpt, concurrency int, projectService services.ProjectService,
	asyncTaskService services.AsyncTaskService, webSocketService services.WebSocketService) *asynq.Server {
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

	mux := asynq.NewServeMux() // 注册任务处理器

	mux.Handle(common.TaskTypeProjectInit, projectService)

	mux.Handle(common.TaskTypeProjectStage, asyncTaskService)
	mux.Handle(common.TaskTypeAgentTaskResponse, asyncTaskService)
	mux.Handle(common.TaskTypeAgentChat, asyncTaskService)
	mux.Handle(common.TaskTypeProjectDownload, asyncTaskService)
	mux.Handle(common.TaskTypeProjectBackup, asyncTaskService)
	mux.Handle(common.TaskTypeProjectDeploy, asyncTaskService)

	mux.Handle(common.TaskTypeWebSocketBroadcast, webSocketService)
	// ... 注册其他任务处理器

	go func() { // 启动服务器
		logger.Info("异步服务启动中... ")
		if err := server.Run(mux); err != nil { // 启动 Worker
			log.Fatal("Could not start worker: ", err)
		}
	}()
	return server
}
