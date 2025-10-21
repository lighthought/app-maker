package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisCache Redis 缓存实现
type RedisCache struct {
	client *redis.Client
	ctx    context.Context
}

// NewRedisCache 创建新的 Redis 缓存实例
func NewRedisCache(client *redis.Client) *RedisCache {
	return &RedisCache{
		client: client,
		ctx:    context.Background(),
	}
}

// Set 设置缓存值
func (c *RedisCache) Set(key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to serialize value: %s", err.Error())
	}

	return c.client.Set(c.ctx, key, data, expiration).Err()
}

// Get 获取缓存值
func (c *RedisCache) Get(key string, dest interface{}) error {
	data, err := c.client.Get(c.ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("key does not exist: %s", key)
		}
		return fmt.Errorf("failed to get cache: %s", err.Error())
	}

	return json.Unmarshal(data, dest)
}

// Delete 删除缓存键
func (c *RedisCache) Delete(key string) error {
	return c.client.Del(c.ctx, key).Err()
}

// Exists 检查键是否存在
func (c *RedisCache) Exists(key string) bool {
	result, err := c.client.Exists(c.ctx, key).Result()
	return err == nil && result > 0
}

// SetInt 设置整数值
func (c *RedisCache) SetInt(key string, value int, expiration time.Duration) error {
	return c.client.Set(c.ctx, key, value, expiration).Err()
}

// GetInt 获取整数值
func (c *RedisCache) GetInt(key string) (int, error) {
	result, err := c.client.Get(c.ctx, key).Int()
	if err != nil {
		if err == redis.Nil {
			return 0, fmt.Errorf("key does not exist: %s", key)
		}
		return 0, fmt.Errorf("failed to get int value: %s", err.Error())
	}
	return result, nil
}

// SetMultiple 批量设置缓存值
func (c *RedisCache) SetMultiple(values map[string]interface{}, expiration time.Duration) error {
	pipe := c.client.Pipeline()

	for key, value := range values {
		data, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to serialize value: %s", err.Error())
		}
		pipe.Set(c.ctx, key, data, expiration)
	}

	_, err := pipe.Exec(c.ctx)
	return err
}

// DeleteMultiple 批量删除缓存键
func (c *RedisCache) DeleteMultiple(keys []string) error {
	if len(keys) == 0 {
		return nil
	}

	return c.client.Del(c.ctx, keys...).Err()
}

// 发布消息到 Redis Pub/Sub
func (c *RedisCache) Publish(channel string, message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to serialize message: %s", err.Error())
	}
	return c.client.Publish(c.ctx, channel, data).Err()
}

// 订阅消息
func (c *RedisCache) Subscribe(channel string) *redis.PubSub {
	return c.client.Subscribe(c.ctx, channel)
}

// Keys 获取匹配模式的键
func (c *RedisCache) Keys(pattern string) ([]string, error) {
	return c.client.Keys(c.ctx, pattern).Result()
}

// Expire 设置键的过期时间
func (c *RedisCache) Expire(key string, expiration time.Duration) error {
	return c.client.Expire(c.ctx, key, expiration).Err()
}

// TTL 获取键的剩余生存时间
func (c *RedisCache) TTL(key string) (time.Duration, error) {
	result, err := c.client.TTL(c.ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get TTL: %s", err.Error())
	}
	return result, nil
}

// Ping 健康检查
func (c *RedisCache) Ping() error {
	return c.client.Ping(c.ctx).Err()
}

// Close 关闭连接
func (c *RedisCache) Close() error {
	return c.client.Close()
}
