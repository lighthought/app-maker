package cache

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// CacheType 缓存类型
type CacheType string

const (
	CacheTypeRedis  CacheType = "redis"
	CacheTypeMemory CacheType = "memory" // 预留内存缓存类型
)

// Config 缓存配置
type Config struct {
	Type     CacheType `json:"type"`
	Host     string    `json:"host"`
	Port     int       `json:"port"`
	Password string    `json:"password"`
	DB       int       `json:"db"`
	PoolSize int       `json:"pool_size"`
	MinIdle  int       `json:"min_idle"`
}

// NewCache 创建新的缓存实例
func NewCache(config Config) (Cache, error) {
	switch config.Type {
	case CacheTypeRedis:
		return NewRedisCacheWithConfig(config)
	case CacheTypeMemory:
		return nil, fmt.Errorf("内存缓存暂未实现")
	default:
		return nil, fmt.Errorf("不支持的缓存类型: %s", config.Type)
	}
}

// NewRedisCacheWithConfig 使用配置创建 Redis 缓存实例
func NewRedisCacheWithConfig(config Config) (*RedisCache, error) {
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
		return nil, fmt.Errorf("Redis 连接测试失败: %w", err)
	}

	return NewRedisCache(client), nil
}
