package container

import (
	"app-maker-agents/internal/api/handlers"
	"app-maker-agents/internal/config"
	"app-maker-agents/internal/services"
	"fmt"
	"log"

	"shared-models/auth"
	"shared-models/cache"
	"shared-models/common"
	"shared-models/logger"

	"github.com/hibiken/asynq"
)

type Container struct {
	// External Services
	AsyncClient    *asynq.Client
	AsyncInspector *asynq.Inspector
	AsyncServer    *asynq.Server
	JWTService     *auth.JWTService
	CacheInstance  cache.Cache

	// Internal Services
	CommandService   services.CommandService
	GitService       services.GitService
	FileService      services.FileService
	AgentTaskService services.AgentTaskService
	ProjectService   services.ProjectService

	// API Handlers
	ProjectHandler   *handlers.ProjectHandler
	ChatHandler      *handlers.ChatHandler
	AnalyseHandler   *handlers.AnalyseHandler
	PmHandler        *handlers.PmHandler
	UxHandler        *handlers.UxHandler
	ArchitectHandler *handlers.ArchitectHandler
	PoHandler        *handlers.PoHandler
	DevHandler       *handlers.DevHandler
	TaskHandler      *handlers.TaskHandler
	HealthHandler    *handlers.HealthHandler
}

func NewContainer(cfg *config.Config) *Container {
	asyncOpt := asynq.RedisClientOpt{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	}

	// 创建独立的 Redis 客户端用于缓存
	cacheInstance, _ := cache.NewCache(cache.Config{
		Host:     cfg.Redis.Host,
		Port:     cfg.Redis.Port,
		Password: cfg.Redis.Password,
		DB:       2,
		PoolSize: 10,
		MinIdle:  5,
	})

	asyncClient := asynq.NewClient(asyncOpt)
	asyncInspector := asynq.NewInspector(asyncOpt)

	commandSvc := services.NewCommandService(cfg.Command, cfg.App.WorkspacePath)
	gitService := services.NewGitService(commandSvc)

	fileSvc := services.NewFileService(commandSvc, cfg.App.WorkspacePath)
	agentTaskService := services.NewAgentTaskService(commandSvc, fileSvc, gitService, asyncClient, cacheInstance)
	projectSvc := services.NewProjectService(commandSvc, agentTaskService, fileSvc)

	asynqServer := initAsynqWorker(&asyncOpt, cfg.Asynq.Concurrency, agentTaskService, projectSvc)

	projectHandler := handlers.NewProjectHandler(agentTaskService, projectSvc)
	chatHandler := handlers.NewChatHandler(agentTaskService)
	analyseHandler := handlers.NewAnalyseHandler(agentTaskService)
	pmHandler := handlers.NewPmHandler(agentTaskService)
	poHandler := handlers.NewPoHandler(agentTaskService)
	devHandler := handlers.NewDevHandler(agentTaskService, commandSvc)
	architectHandler := handlers.NewArchitectHandler(agentTaskService)
	uxHandler := handlers.NewUxHandler(agentTaskService)
	taskHandler := handlers.NewTaskHandler(asyncInspector)
	healthHandler := handlers.NewHealthHandler(cacheInstance)

	return &Container{
		AsyncClient:      asyncClient,
		AsyncInspector:   asyncInspector,
		AgentTaskService: agentTaskService,
		AsyncServer:      asynqServer,
		CommandService:   commandSvc,
		GitService:       gitService,
		CacheInstance:    cacheInstance,
		ProjectHandler:   projectHandler,
		ChatHandler:      chatHandler,
		AnalyseHandler:   analyseHandler,
		PmHandler:        pmHandler,
		PoHandler:        poHandler,
		DevHandler:       devHandler,
		ArchitectHandler: architectHandler,
		UxHandler:        uxHandler,
		TaskHandler:      taskHandler,
		HealthHandler:    healthHandler,
	}
}

func initAsynqWorker(redisClientOpt *asynq.RedisClientOpt, concurrency int,
	agentTaskService services.AgentTaskService,
	projectSvc services.ProjectService) *asynq.Server {
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
	mux.Handle(common.TaskTypeAgentExecute, agentTaskService)
	mux.Handle(common.TaskTypeAgentChat, agentTaskService)
	mux.Handle(common.TaskTypeAgentSetup, projectSvc)
	mux.Handle(common.TaskTypeProjectDeploy, projectSvc)
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

func (c *Container) Stop() {
	logger.Info("Stopping container... ")
	if c.AsyncServer != nil {
		c.AsyncServer.Shutdown()
	}
	if c.AsyncClient != nil {
		c.AsyncClient.Close()
	}
	if c.AsyncInspector != nil {
		c.AsyncInspector.Close()
	}
	if c.CacheInstance != nil {
		c.CacheInstance.Close()
	}
}
