// Package main AutoCodeWeb Backend API
//
// AutoCodeWeb Backend 是一个基于Go + Gin + GORM + PostgreSQL + Redis的多Agent协作系统后端，
// 为前端提供高性能的API服务，包括项目管理、BMad-Method集成、任务执行等功能。
//
//	Schemes: http, https
//	Host: localhost:8080
//	BasePath: /api/v1
//	Version: 1.0.0
//	Title: AutoCodeWeb Backend API
//	Description: AutoCodeWeb Backend 是一个基于Go + Gin + GORM + PostgreSQL + Redis的多Agent协作系统后端，为前端提供高性能的API服务，包括项目管理、BMad-Method集成、任务执行等功能。
//
//	Consumes:
//	- application/json
//	- multipart/form-data
//
//	Produces:
//	- application/json
//
//	Security:
//	- bearer
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lighthought/app-maker/shared-models/common"
	"github.com/lighthought/app-maker/shared-models/logger"

	"github.com/lighthought/app-maker/backend/internal/api/middleware"
	"github.com/lighthought/app-maker/backend/internal/api/routes"
	"github.com/lighthought/app-maker/backend/internal/config"
	"github.com/lighthought/app-maker/backend/internal/container"
	"github.com/lighthought/app-maker/backend/internal/database"

	_ "github.com/lighthought/app-maker/backend/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// loadConfig 加载配置
func loadConfigAndService() (*config.Config, error) {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		return nil, fmt.Errorf("加载配置失败: %v", err)
	}

	// 初始化日志
	if err := logger.Init(cfg.Log.Level, cfg.Log.File); err != nil {
		fmt.Printf("初始化日志失败: %v\n", err)
		return nil, fmt.Errorf("初始化日志失败: %v", err)
	}

	logger.Info("启动 AppMaker 后端服务, 配置和日志初始化成功")

	// 连接数据库
	if err := database.Connect(cfg.Database); err != nil {
		logger.Fatal("连接数据库失败", logger.String("error", err.Error()))
		return nil, fmt.Errorf("连接数据库失败: %v", err)
	}
	logger.Info("数据库连接成功")

	// 连接Redis
	if err := database.ConnectRedis(cfg.Redis); err != nil {
		logger.Warn("连接Redis失败，将使用内存缓存", logger.String("error", err.Error()))
		return nil, fmt.Errorf("连接Redis失败: %v", err)
	}

	logger.Info("Redis连接成功")
	return cfg, nil
}

// setupContainer 初始化依赖注入容器
func setupContainer(cfg *config.Config) (*container.Container, *gin.Engine) {
	var theContainer *container.Container
	if database.GetDB() != nil && database.GetRedis() != nil {
		theContainer = container.NewContainer(cfg, database.GetDB(), database.GetRedis())
		logger.Info("依赖注入容器初始化成功")
	} else {
		logger.Warn("数据库未连接，部分功能不可用")
	}

	// 设置Gin模式
	if cfg.App.Environment == common.EnvironmentProduction {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建Gin引擎
	engine := gin.New()

	// 注册中间件
	engine.Use(gin.Logger())
	engine.Use(middleware.CORS(cfg.CORS))
	engine.Use(middleware.RequestID())
	engine.Use(gin.Recovery())

	// 添加Swagger文档路由
	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	// 注册路由
	routes.Register(engine, theContainer)
	return theContainer, engine
}

func startServer(cfg *config.Config, engine *gin.Engine, theContainer *container.Container) {
	// 创建HTTP服务器
	srv := &http.Server{
		Addr:    ":" + cfg.App.Port,
		Handler: engine,
	}

	// 启动服务器
	go func() {
		logger.Info("服务器启动在端口 " + cfg.App.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("启动服务器失败", logger.String("error", err.Error()))
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("正在关闭服务器...")
	theContainer.Stop()

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("服务器强制关闭", logger.String("error", err.Error()))
	}

	defer database.Close()
	defer database.CloseRedis()

	logger.Info("服务器已关闭")
}

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token
func main() {
	cfg, err := loadConfigAndService()
	if err != nil {
		logger.Fatal("加载配置、连接缓存或数据库失败，程序退出", logger.String("error", err.Error()))
		os.Exit(1)
	}

	container, engine := setupContainer(cfg)
	if container == nil {
		logger.Fatal("依赖注入容器初始化失败，程序退出")
		os.Exit(1)
	}

	startServer(cfg, engine, container)
}
