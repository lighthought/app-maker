package cache

import "time"

// Cache 缓存接口定义
type Cache interface {
	// 基础缓存操作
	Set(key string, value interface{}, expiration time.Duration) error
	Get(key string, dest interface{}) error
	Delete(key string) error
	Exists(key string) bool

	// 数值类型操作
	SetInt(key string, value int, expiration time.Duration) error
	GetInt(key string) (int, error)

	// 批量操作
	SetMultiple(values map[string]interface{}, expiration time.Duration) error
	DeleteMultiple(keys []string) error

	// 键管理
	Keys(pattern string) ([]string, error)
	Expire(key string, expiration time.Duration) error
	TTL(key string) (time.Duration, error)

	// 健康检查
	Ping() error
	Close() error
}
