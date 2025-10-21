// Package main Agents Service API
//
// Agents Service 是一个基于Go + Gin的多Agent协作系统，
// 为各种AI Agent提供统一的API接口，包括分析、产品、架构、开发、UX等Agent。
//
//	Schemes: http, https
//	Host: localhost:9090
//	BasePath: /api/v1
//	Version: 1.0.0
//	Title: Agents Service API
//	Description: Agents Service 是一个基于Go + Gin的多Agent协作系统，为各种AI Agent提供统一的API接口。
//
//	Consumes:
//	- application/json
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

	"github.com/lighthought/app-maker/agents/internal/api/routes"
	"github.com/lighthought/app-maker/agents/internal/config"
	"github.com/lighthought/app-maker/agents/internal/container"

	_ "github.com/lighthought/app-maker/agents/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// 加载配置
func loadConfig() (*config.Config, error) {
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

	if cfg.App.Environment == common.EnvironmentProduction {
		gin.SetMode(gin.ReleaseMode)
	}
	return cfg, nil
}

func setupContainer(cfg *config.Config) (*container.Container, *gin.Engine) {
	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery())

	// 添加Swagger文档路由
	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	container := container.NewContainer(cfg)

	// 注册路由
	routes.Register(engine, container)
	return container, engine
}

func startServer(cfg *config.Config, container *container.Container, engine *gin.Engine) {
	srv := &http.Server{
		Addr:    ":" + cfg.App.Port,
		Handler: engine,
	}

	go func() {
		fmt.Printf("HTTP 服务监听端口: %s\n", cfg.App.Port)
		fmt.Printf("Swagger 文档地址: http://localhost:%s/swagger/index.html\n", cfg.App.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("启动失败: %v\n", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("收到退出信号，开始优雅关闭")
	container.Stop()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		fmt.Printf("强制关闭: %v\n", err)
	}
	fmt.Println("服务已关闭")
}

func main() {
	cfg, err := loadConfig()
	if err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		os.Exit(1)
	}

	container, engine := setupContainer(cfg)
	if container == nil {
		logger.Fatal("依赖注入容器初始化失败，程序退出")
		os.Exit(1)
	}
	startServer(cfg, container, engine)
}
