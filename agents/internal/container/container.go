package container

import (
	"app-maker-agents/internal/api/handlers"
	"app-maker-agents/internal/config"
	"app-maker-agents/internal/services"
	"shared-models/auth"
)

type Container struct {
	CommandService *services.CommandService
	GitService     services.GitService

	ProjectHandler   *handlers.ProjectHandler
	AnalyseHandler   *handlers.AnalyseHandler
	PmHandler        *handlers.PmHandler
	PoHandler        *handlers.PoHandler
	DevHandler       *handlers.DevHandler
	ArchitectHandler *handlers.ArchitectHandler
	UxHandler        *handlers.UxHandler

	JWTService *auth.JWTService
}

func NewContainer(cfg *config.Config) *Container {
	commandSvc := services.NewCommandService(cfg.Command, cfg.App.WorkspacePath)
	gitSvc := services.NewGitService()
	projectSvc := services.NewProjectService(commandSvc, cfg.App.WorkspacePath)

	projectHandler := handlers.NewProjectHandler(projectSvc)
	analyseHandler := handlers.NewAnalyseHandler(commandSvc)
	pmHandler := handlers.NewPmHandler(commandSvc)
	poHandler := handlers.NewPoHandler(commandSvc)
	devHandler := handlers.NewDevHandler(commandSvc)
	architectHandler := handlers.NewArchitectHandler(commandSvc)
	uxHandler := handlers.NewUxHandler(commandSvc)

	return &Container{
		CommandService:   commandSvc,
		GitService:       gitSvc,
		ProjectHandler:   projectHandler,
		AnalyseHandler:   analyseHandler,
		PmHandler:        pmHandler,
		PoHandler:        poHandler,
		DevHandler:       devHandler,
		ArchitectHandler: architectHandler,
		UxHandler:        uxHandler,
	}
}
