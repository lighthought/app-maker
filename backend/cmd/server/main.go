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

	"autocodeweb-backend/internal/api/middleware"
	"autocodeweb-backend/internal/api/routes"
	"autocodeweb-backend/internal/config"
	"autocodeweb-backend/internal/database"
	"autocodeweb-backend/pkg/cache"
	"autocodeweb-backend/pkg/logger"

	_ "autocodeweb-backend/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token
func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志
	if err := logger.Init(cfg.Log.Level, cfg.Log.File); err != nil {
		fmt.Printf("初始化日志失败: %v\n", err)
		os.Exit(1)
	}

	logger.Info("启动AutoCodeWeb后端服务")

	// 连接数据库
	if err := database.Connect(cfg.Database); err != nil {
		logger.Fatal("连接数据库失败", logger.String("error", err.Error()))
	}
	defer database.Close()

	// 连接Redis
	if err := database.ConnectRedis(cfg.Redis); err != nil {
		logger.Warn("连接Redis失败，将使用内存缓存", logger.String("error", err.Error()))
	} else {
		logger.Info("Redis连接成功")
		defer database.CloseRedis()
	}

	// 初始化缓存系统
	var cacheInstance cache.Cache
	var monitor *cache.Monitor

	if database.GetRedis() != nil {
		// 创建缓存配置
		cacheConfig := cache.Config{
			Type:     cache.CacheTypeRedis,
			Host:     cfg.Redis.Host,
			Port:     cfg.Redis.Port,
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
			PoolSize: 10,
			MinIdle:  5,
		}

		// 创建缓存实例
		if cacheInstance, err = cache.NewCache(cacheConfig); err != nil {
			logger.Warn("创建缓存实例失败，将使用内存缓存", logger.String("error", err.Error()))
		} else {
			logger.Info("缓存系统初始化成功")
			// 创建监控实例
			monitor = cache.NewMonitor(database.GetRedis())
		}
	}

	// 如果缓存初始化失败，设置为 nil
	if cacheInstance == nil {
		logger.Warn("缓存系统不可用，相关功能将受限")
	}

	// 设置Gin模式
	if cfg.App.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建Gin引擎
	engine := gin.New()

	// 注册中间件
	engine.Use(gin.Logger())
	engine.Use(gin.Recovery())
	engine.Use(middleware.CORS(cfg.CORS))
	engine.Use(middleware.RequestID())
	engine.Use(gin.Recovery())

	// 添加Swagger文档路由
	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	// 注册路由
	routes.Register(engine, cfg, cacheInstance, monitor, database.GetDB())

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

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("服务器强制关闭", logger.String("error", err.Error()))
	}

	logger.Info("服务器已关闭")
}
