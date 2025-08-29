package cache

import (
	"testing"
	"time"
)

// TestKeyBuilder 测试键构建器
func TestKeyBuilder(t *testing.T) {
	builder := NewKeyBuilder()

	// 测试用户相关键
	userKey := builder.User("123", "profile")
	if userKey != "user:123:profile" {
		t.Errorf("期望 user:123:profile，得到 %s", userKey)
	}

	// 测试项目相关键
	projectKey := builder.Project("456", "info")
	if projectKey != "project:456:info" {
		t.Errorf("期望 project:456:info，得到 %s", projectKey)
	}

	// 测试任务相关键
	taskKey := builder.Task("789", "status")
	if taskKey != "task:789:status" {
		t.Errorf("期望 task:789:status，得到 %s", taskKey)
	}

	// 测试会话键
	sessionKey := builder.Session("session123")
	if sessionKey != "session:session123" {
		t.Errorf("期望 session:session123，得到 %s", sessionKey)
	}

	// 测试限流键
	rateLimitKey := builder.RateLimit("192.168.1.1")
	if rateLimitKey != "rate_limit:192.168.1.1" {
		t.Errorf("期望 rate_limit:192.168.1.1，得到 %s", rateLimitKey)
	}

	// 测试通用缓存键
	cacheKey := builder.Cache("test", "value")
	if cacheKey != "cache:test:value" {
		t.Errorf("期望 cache:test:value，得到 %s", cacheKey)
	}

	// 测试自定义键构建
	customKey := BuildKey("custom", "key", "parts")
	if customKey != "custom:key:parts" {
		t.Errorf("期望 custom:key:parts，得到 %s", customKey)
	}
}

// TestGetExpiration 测试过期时间获取
func TestGetExpiration(t *testing.T) {
	// 测试短期过期
	shortExp := GetExpiration("short")
	if shortExp != 5*time.Minute {
		t.Errorf("期望 5分钟，得到 %v", shortExp)
	}

	// 测试中期过期
	mediumExp := GetExpiration("medium")
	if mediumExp != 30*time.Minute {
		t.Errorf("期望 30分钟，得到 %v", mediumExp)
	}

	// 测试长期过期
	longExp := GetExpiration("long")
	if longExp != 2*time.Hour {
		t.Errorf("期望 2小时，得到 %v", longExp)
	}

	// 测试默认过期
	defaultExp := GetExpiration("unknown")
	if defaultExp != 30*time.Minute {
		t.Errorf("期望 30分钟默认值，得到 %v", defaultExp)
	}
}

// TestConfig 测试配置结构
func TestConfig(t *testing.T) {
	config := Config{
		Type:     CacheTypeRedis,
		Host:     "localhost",
		Port:     6379,
		Password: "",
		DB:       0,
		PoolSize: 10,
		MinIdle:  5,
	}

	if config.Type != CacheTypeRedis {
		t.Errorf("期望 CacheTypeRedis，得到 %s", config.Type)
	}

	if config.Host != "localhost" {
		t.Errorf("期望 localhost，得到 %s", config.Host)
	}

	if config.Port != 6379 {
		t.Errorf("期望 6379，得到 %d", config.Port)
	}
}
