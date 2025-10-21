package cache

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// Config 缓存配置
type Config struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password"`
	DB       int    `json:"db"`
	PoolSize int    `json:"pool_size"`
	MinIdle  int    `json:"min_idle"`
}

// NewRedisCache 使用配置创建 Redis 缓存实例
func NewCache(config Config) (Cache, error) {
	// 设置默认值
	if config.PoolSize == 0 {
		config.PoolSize = 10
	}
	if config.MinIdle == 0 {
		config.MinIdle = 5
	}

	// 创建 Redis 客户端
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password:     config.Password,
		DB:           config.DB,
		PoolSize:     config.PoolSize,
		MinIdleConns: config.MinIdle,
	})

	// 测试连接
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to test Redis connection: %s", err.Error())
	}

	return NewRedisCache(client), nil
}
