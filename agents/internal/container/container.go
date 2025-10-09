package container

import (
	"app-maker-agents/internal/api/handlers"
	"app-maker-agents/internal/config"
	"app-maker-agents/internal/services"
	"fmt"
	"log"
	"shared-models/auth"
	"shared-models/common"
	"shared-models/logger"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

type Container struct {
	// External Services
	AsyncClient    *asynq.Client
	AsyncInspector *asynq.Inspector
	AsyncServer    *asynq.Server
	JWTService     *auth.JWTService
	RedisClient    *redis.Client

	// Internal Services
	CommandService   services.CommandService
	GitService       services.GitService
	AgentTaskService services.AgentTaskService
	ProjectService   services.ProjectService

	// API Handlers
	ProjectHandler   *handlers.ProjectHandler
	AnalyseHandler   *handlers.AnalyseHandler
	PmHandler        *handlers.PmHandler
	UxHandler        *handlers.UxHandler
	ArchitectHandler *handlers.ArchitectHandler
	PoHandler        *handlers.PoHandler
	DevHandler       *handlers.DevHandler
	TaskHandler      *handlers.TaskHandler
}

func NewContainer(cfg *config.Config) *Container {

	redisClientOpt := asynq.RedisClientOpt{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	}

	// 创建独立的 Redis 客户端用于缓存
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       2,
	})

	asyncClient := asynq.NewClient(redisClientOpt)
	asyncInspector := asynq.NewInspector(redisClientOpt)

	commandSvc := services.NewCommandService(cfg.Command, cfg.App.WorkspacePath)
	gitService := services.NewGitService(commandSvc)

	projectSvc := services.NewProjectService(commandSvc, cfg.App.WorkspacePath)
	agentTaskService := services.NewAgentTaskService(commandSvc, gitService, asyncClient, redisClient)
	asynqServer := initAsynqWorker(&redisClientOpt, cfg.Asynq.Concurrency, agentTaskService, projectSvc)

	projectHandler := handlers.NewProjectHandler(agentTaskService, projectSvc)
	analyseHandler := handlers.NewAnalyseHandler(agentTaskService)
	pmHandler := handlers.NewPmHandler(agentTaskService)
	poHandler := handlers.NewPoHandler(agentTaskService)
	devHandler := handlers.NewDevHandler(agentTaskService)
	architectHandler := handlers.NewArchitectHandler(agentTaskService)
	uxHandler := handlers.NewUxHandler(agentTaskService)
	taskHandler := handlers.NewTaskHandler(asyncInspector)

	return &Container{
		AsyncClient:      asyncClient,
		AsyncInspector:   asyncInspector,
		AgentTaskService: agentTaskService,
		AsyncServer:      asynqServer,
		CommandService:   commandSvc,
		GitService:       gitService,
		RedisClient:      redisClient,
		ProjectHandler:   projectHandler,
		AnalyseHandler:   analyseHandler,
		PmHandler:        pmHandler,
		PoHandler:        poHandler,
		DevHandler:       devHandler,
		ArchitectHandler: architectHandler,
		UxHandler:        uxHandler,
		TaskHandler:      taskHandler,
	}
}

func initAsynqWorker(redisClientOpt *asynq.RedisClientOpt, concurrency int, agentTaskService services.AgentTaskService, projectSvc services.ProjectService) *asynq.Server {
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
	mux.Handle(common.TaskTypeAgentSetup, projectSvc)
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
	c.AsyncServer.Shutdown()
	c.AsyncClient.Close()
	c.AsyncInspector.Close()
	if c.RedisClient != nil {
		c.RedisClient.Close()
	}
}
