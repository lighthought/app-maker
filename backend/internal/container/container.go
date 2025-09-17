package container

import (
	"autocodeweb-backend/internal/api/handlers"
	"autocodeweb-backend/internal/config"
	"autocodeweb-backend/internal/repositories"
	"autocodeweb-backend/internal/services"
	"autocodeweb-backend/pkg/auth"
	"autocodeweb-backend/pkg/cache"
	"fmt"
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

	// Handlers
	UserHandler    *handlers.UserHandler
	TaskHandler    *handlers.TaskHandler
	ProjectHandler *handlers.ProjectHandler
	FileHandler    *handlers.FileHandler
	ChatHandler    *handlers.ChatHandler
	CacheHandler   *handlers.CacheHandler
}

func NewContainer(cfg *config.Config, db *gorm.DB, redis *redis.Client, cacheInstance cache.Cache) *Container {
	// asynq items
	asyncClient := asynq.NewClient(asynq.RedisClientOpt{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
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
	userService := services.NewUserService(userRepository, jwtService, cfg.JWT.Expire)
	fileService := services.NewFileService(asyncClient)
	projectTemplateService := services.NewProjectTemplateService(fileService)
	projectStageService := services.NewProjectStageService(projectRepository, stageRepository)
	projectNameGenerator := services.NewProjectNameGenerator()
	gitService := services.NewGitService()
	gitService.SetupSSH()
	projectService := services.NewProjectService(projectRepository, messageRepository,
		asyncClient, projectTemplateService, projectNameGenerator, gitService)
	messageService := services.NewMessageService(messageRepository)

	// handlers
	cacheHandler := handlers.NewCacheHandler(cacheInstance, cachMonitor)
	chatHandler := handlers.NewChatHandler(messageService, fileService)
	fileHandler := handlers.NewFileHandler(fileService, projectService)
	projectHandler := handlers.NewProjectHandler(projectService, projectStageService)
	taskHandler := handlers.NewTaskHandler(asyncInspector)
	userHandler := handlers.NewUserHandler(userService)
	return &Container{
		AsyncClient:    asyncClient,
		JWTService:     jwtService,
		CachMonitor:    cachMonitor,
		AsyncInspector: asyncInspector,

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
		CacheHandler:           cacheHandler,
		ChatHandler:            chatHandler,
		FileHandler:            fileHandler,
		ProjectHandler:         projectHandler,
		TaskHandler:            taskHandler,
		UserHandler:            userHandler,
	}
}
