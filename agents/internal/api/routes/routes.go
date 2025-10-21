package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/lighthought/app-maker/agents/internal/container"
	"github.com/lighthought/app-maker/shared-models/common"
)

// 设置空POST路由
func setPostEmptyEndpoint(routers *gin.RouterGroup, relativePath string, message string) {
	routers.POST(relativePath, func(c *gin.Context) {
		c.JSON(200, gin.H{"message": message})
	})
}

// 注册Agent API路由
func registerAgentApiRoutes(routers *gin.RouterGroup, container *container.Container) {
	agent := routers.Group("/agent") // Chat Handler
	{
		chat := agent.Group("/chat")
		{
			chatHandler := container.ChatHandler
			if chatHandler != nil {
				chat.POST("", chatHandler.ChatWithAgent)
			} else {
				setPostEmptyEndpoint(chat, "", "Chat with agent endpoint - TODO")
			}
		}

		analyse := agent.Group("/analyse") // 分析师 Agent
		{
			analyseHandler := container.AnalyseHandler
			if analyseHandler != nil {
				analyse.POST("/project-brief", analyseHandler.ProjectBrief) // 生成项目概览
			} else {
				setPostEmptyEndpoint(analyse, "/project-brief", "Analyse project brief endpoint - TODO")
			}
		}

		po := agent.Group("/po") // 产品经理 Agent
		{
			poHandler := container.PoHandler
			if poHandler != nil {
				po.POST("/epicsandstories", poHandler.GetEpicsAndStories) // 获取史诗和故事
			} else {
				setPostEmptyEndpoint(po, "/epicsandstories", "Po epics and stories endpoint - TODO")
			}
		}

		pm := agent.Group("/pm") // 项目经理 Agent
		{
			pmHandler := container.PmHandler
			if pmHandler != nil {
				pm.POST("/prd", pmHandler.GetPRD) // 获取PRD
			} else {
				setPostEmptyEndpoint(pm, "/prd", "Pm prd endpoint - TODO")
			}
		}

		dev := agent.Group("/dev") // 开发 Agent
		{
			devHandler := container.DevHandler
			if devHandler != nil {
				dev.POST("/fixbug", devHandler.FixBug)            // 修复bug
				dev.POST("/implstory", devHandler.ImplementStory) // 实现故事
				dev.POST("/runtest", devHandler.RunTest)          // 运行测试
				dev.POST("/deploy", devHandler.Deploy)            // 部署
			} else {
				setPostEmptyEndpoint(dev, "/fixbug", "Dev fix bug endpoint - TODO")
				setPostEmptyEndpoint(dev, "/implstory", "Dev implement story endpoint - TODO")
				setPostEmptyEndpoint(dev, "/runtest", "Dev run test endpoint - TODO")
				setPostEmptyEndpoint(dev, "/deploy", "Dev deploy endpoint - TODO")
			}
		}

		architect := agent.Group("/architect") // 架构师 Agent
		{
			architectHandler := container.ArchitectHandler
			if architectHandler != nil {
				architect.POST("/architect", architectHandler.GetArchitecture)      // 获取架构
				architect.POST("/apidefinition", architectHandler.GetAPIDefinition) // 获取API定义
				architect.POST("/database", architectHandler.GetDatabaseDesign)     // 获取数据库设计
			} else {
				setPostEmptyEndpoint(architect, "/architect", "Architect get architecture endpoint - TODO")
				setPostEmptyEndpoint(architect, "/apidefinition", "Architect get api definition endpoint - TODO")
				setPostEmptyEndpoint(architect, "/database", "Architect get database design endpoint - TODO")
			}
		}

		ux := agent.Group("/ux-expert") // 用户体验专家 Agent
		{
			uxHandler := container.UxHandler
			if uxHandler != nil {
				ux.POST("/ux-standard", uxHandler.GetUXStandard) // 获取用户体验标准
			} else {
				setPostEmptyEndpoint(ux, "/ux-standard", "Ux get ux standard endpoint - TODO")
			}
		}
	}
}

// Register 注册路由
func Register(engine *gin.Engine, container *container.Container) {
	routers := engine.Group(common.DefaultApiPrefix)
	{
		routers.GET("/health", container.HealthHandler.HealthCheck)

		projectHandler := container.ProjectHandler
		project := routers.Group("/project") // 项目API路由
		{
			if projectHandler != nil {
				project.POST("/setup", projectHandler.SetupProjectEnvironment) // 准备项目环境
			} else {
				setPostEmptyEndpoint(project, "/setup", "Project setup endpoint - TODO")
			}
		}

		var taskHandler = container.TaskHandler
		tasks := routers.Group("/tasks") // 异步任务路由
		{
			if taskHandler != nil {
				tasks.GET("/:id", taskHandler.GetTaskStatus) // 获取任务状
			} else {
				tasks.GET("/:id", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "Task status endpoint - TODO"})
				})
			}
		}

		// 注册agent相关的 API
		registerAgentApiRoutes(routers, container)
	}
}
