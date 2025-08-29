package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"autocodeweb-backend/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	fmt.Println("测试数据库连接...")

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 测试直接连接
	fmt.Printf("尝试连接到数据库: %s:%d/%s\n", cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Password, cfg.Database.Name, cfg.Database.SSLMode)

	// 设置连接超时
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.Database.ConnectTimeout)*time.Second)
	defer cancel()

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}

	// 测试连接
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("获取sql.DB失败: %v", err)
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		log.Fatalf("数据库ping失败: %v", err)
	}

	fmt.Println("✅ 数据库连接成功!")

	// 测试数据库版本
	var version string
	if err := db.Raw("SELECT version()").Scan(&version).Error; err != nil {
		log.Printf("获取数据库版本失败: %v", err)
	} else {
		fmt.Printf("数据库版本: %s\n", version)
	}

	// 测试UUID扩展
	var uuidExists bool
	if err := db.Raw("SELECT EXISTS(SELECT 1 FROM pg_extension WHERE extname = 'uuid-ossp')").Scan(&uuidExists).Error; err != nil {
		log.Printf("检查UUID扩展失败: %v", err)
	} else {
		if uuidExists {
			fmt.Println("✅ UUID扩展已启用")
		} else {
			fmt.Println("⚠️  UUID扩展未启用")
		}
	}

	// 测试pgcrypto扩展
	var pgcryptoExists bool
	if err := db.Raw("SELECT EXISTS(SELECT 1 FROM pg_extension WHERE extname = 'pgcrypto')").Scan(&pgcryptoExists).Error; err != nil {
		log.Printf("检查pgcrypto扩展失败: %v", err)
	} else {
		if pgcryptoExists {
			fmt.Println("✅ pgcrypto扩展已启用")
		} else {
			fmt.Println("⚠️  pgcrypto扩展未启用")
		}
	}

	// 关闭连接
	if err := sqlDB.Close(); err != nil {
		log.Printf("关闭数据库连接失败: %v", err)
	}

	fmt.Println("数据库连接测试完成")
}
