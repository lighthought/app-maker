package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"autocodeweb-backend/internal/config"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	db          *gorm.DB
	redisClient *redis.Client
)

// Connect 连接 PostgreSQL 数据库
func Connect(cfg config.DatabaseConfig) error {
	// 打印数据库配置信息（调试用）
	log.Printf("数据库配置信息:")
	log.Printf("  Host: %s", cfg.Host)
	log.Printf("  Port: %d", cfg.Port)
	log.Printf("  User: %s", cfg.User)
	log.Printf("  Password: %s", cfg.Password)
	log.Printf("  Name: %s", cfg.Name)
	log.Printf("  SSLMode: %s", cfg.SSLMode)
	log.Printf("  ConnectTimeout: %d", cfg.ConnectTimeout)
	log.Printf("  AutoMigrate: %t", cfg.AutoMigrate)
	log.Printf("  SeedData: %t", cfg.SeedData)

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode)

	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		return fmt.Errorf("连接数据库失败: %w", err)
	}

	// 获取底层sql.DB对象
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("获取sql.DB失败: %w", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(3600 * time.Second) // 1小时

	log.Println("数据库连接成功")

	// 暂时跳过迁移和种子数据，先确保基本连接正常
	log.Println("跳过数据库迁移和种子数据，专注于基本连接")

	return nil
}

// ConnectRedis 连接 Redis
func ConnectRedis(cfg config.RedisConfig) error {
	rdb := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     10,
		MinIdleConns: 5,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("连接Redis失败: %w", err)
	}

	redisClient = rdb
	log.Println("Redis连接成功")
	return nil
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return db
}

// GetRedis 获取Redis实例
func GetRedis() *redis.Client {
	return redisClient
}

// Close 关闭数据库连接
func Close() error {
	if db != nil {
		sqlDB, err := db.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

// CloseRedis 关闭Redis连接
func CloseRedis() error {
	if redisClient != nil {
		return redisClient.Close()
	}
	return nil
}
