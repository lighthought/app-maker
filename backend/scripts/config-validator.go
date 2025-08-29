package main

import (
	"fmt"
	"log"
	"os"

	"autocodeweb-backend/internal/config"
)

func main() {
	fmt.Println("🔍 配置验证工具")
	fmt.Println("==================")

	// 设置环境变量
	if len(os.Args) > 1 {
		env := os.Args[1]
		os.Setenv("APP_ENVIRONMENT", env)
		fmt.Printf("设置环境: %s\n", env)
	}

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ 配置加载失败: %v", err)
	}

	fmt.Println("✅ 配置加载成功")
	fmt.Printf("环境: %s\n", cfg.App.Environment)
	fmt.Printf("端口: %s\n", cfg.App.Port)
	fmt.Printf("数据库主机: %s\n", cfg.Database.Host)
	fmt.Printf("数据库名称: %s\n", cfg.Database.Name)
	fmt.Printf("Redis主机: %s\n", cfg.Redis.Host)
	fmt.Printf("日志级别: %s\n", cfg.Log.Level)

	fmt.Println("\n🎉 所有配置验证通过！")
}
