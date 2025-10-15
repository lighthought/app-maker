package routes

import (
	"app-maker-agents/internal/api/handlers"
	"app-maker-agents/internal/container"
	"shared-models/common"

	"github.com/gin-gonic/gin"
)

// Register 注册路由
func Register(
	engine *gin.Engine,
	container *container.Container,
) {
	routers := engine.Group(common.DefaultApiPrefix)
	{
		routers.GET("/health", handlers.HealthCheck)
		// 项目环境准备
		projectHandler := container.ProjectHandler
		project := routers.Group("/project")
		{
			if projectHandler != nil {
				project.POST("/setup", projectHandler.SetupProjectEnvironment)
			} else {
				project.POST("/setup", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "Project setup endpoint - TODO"})
				})
			}
		}

		// 异步任务路由
		var taskHandler = container.TaskHandler
		tasks := routers.Group("/tasks")
		{
			if taskHandler != nil {
				tasks.GET("/:id", taskHandler.GetTaskStatus) // 获取任务状
			} else {
				tasks.GET("/:id", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "Task status endpoint - TODO"})
				})
			}
		}

		agent := routers.Group("/agent")
		{
			// Chat Handler
			chat := agent.Group("/chat")
			{
				chatHandler := container.ChatHandler
				if chatHandler != nil {
					chat.POST("", chatHandler.ChatWithAgent)
				} else {
					chat.POST("", func(c *gin.Context) {
						c.JSON(200, gin.H{"message": "Chat with agent endpoint - TODO"})
					})
				}
			}

			// 各 Agent Handler
			analyse := agent.Group("/analyse")
			{
				analyseHandler := container.AnalyseHandler
				if analyseHandler != nil {
					analyse.POST("/project-brief", analyseHandler.ProjectBrief)
				} else {
					analyse.POST("/project-brief", func(c *gin.Context) {
						c.JSON(200, gin.H{"message": "Analyse project brief endpoint - TODO"})
					})
				}
			}

			po := agent.Group("/po")
			{
				poHandler := container.PoHandler
				if poHandler != nil {
					po.POST("/epicsandstories", poHandler.GetEpicsAndStories)
				} else {
					po.POST("/epicsandstories", func(c *gin.Context) {
						c.JSON(200, gin.H{"message": "Po epics and stories endpoint - TODO"})
					})
				}
			}

			pm := agent.Group("/pm")
			{
				pmHandler := container.PmHandler
				if pmHandler != nil {
					pm.POST("/prd", pmHandler.GetPRD)
				} else {
					pm.POST("/prd", func(c *gin.Context) {
						c.JSON(200, gin.H{"message": "Pm prd endpoint - TODO"})
					})
				}
			}

			dev := agent.Group("/dev")
			{
				devHandler := container.DevHandler
				if devHandler != nil {
					dev.POST("/fixbug", devHandler.FixBug)
					dev.POST("/implstory", devHandler.ImplementStory)
					dev.POST("/runtest", devHandler.RunTest)
					dev.POST("/deploy", devHandler.Deploy)
				} else {
					dev.POST("/fixbug", func(c *gin.Context) {
						c.JSON(200, gin.H{"message": "Dev fix bug endpoint - TODO"})
					})
					dev.POST("/implstory", func(c *gin.Context) {
						c.JSON(200, gin.H{"message": "Dev implement story endpoint - TODO"})
					})
					dev.POST("/runtest", func(c *gin.Context) {
						c.JSON(200, gin.H{"message": "Dev run test endpoint - TODO"})
					})
					dev.POST("/deploy", func(c *gin.Context) {
						c.JSON(200, gin.H{"message": "Dev deploy endpoint - TODO"})
					})
				}
			}

			architect := agent.Group("/architect")
			{
				architectHandler := container.ArchitectHandler
				if architectHandler != nil {
					architect.POST("/architect", architectHandler.GetArchitecture)
					architect.POST("/apidefinition", architectHandler.GetAPIDefinition)
					architect.POST("/database", architectHandler.GetDatabaseDesign)
				} else {
					architect.POST("/architect", func(c *gin.Context) {
						c.JSON(200, gin.H{"message": "Architect get architecture endpoint - TODO"})
					})
					architect.POST("/apidefinition", func(c *gin.Context) {
						c.JSON(200, gin.H{"message": "Architect get api definition endpoint - TODO"})
					})
					architect.POST("/database", func(c *gin.Context) {
						c.JSON(200, gin.H{"message": "Architect get database design endpoint - TODO"})
					})
				}
			}

			ux := agent.Group("/ux-expert")
			{
				uxHandler := container.UxHandler
				if uxHandler != nil {
					ux.POST("/ux-standard", uxHandler.GetUXStandard)
				} else {
					ux.POST("/ux-standard", func(c *gin.Context) {
						c.JSON(200, gin.H{"message": "Ux get ux standard endpoint - TODO"})
					})
					ux.POST("/page-prompt", func(c *gin.Context) {
						c.JSON(200, gin.H{"message": "Ux get page prompt endpoint - TODO"})
					})
				}
			}
		}

	}
}
