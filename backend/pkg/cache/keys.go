package cache

import (
	"fmt"
	"strings"
	"time"
)

// KeyBuilder 缓存键构建器
type KeyBuilder struct{}

// NewKeyBuilder 创建新的键构建器
func NewKeyBuilder() *KeyBuilder {
	return &KeyBuilder{}
}

// User 用户相关缓存键
func (kb *KeyBuilder) User(userID string, suffix string) string {
	return fmt.Sprintf("user:%s:%s", userID, suffix)
}

// UserProfile 用户资料缓存键
func (kb *KeyBuilder) UserProfile(userID string) string {
	return kb.User(userID, "profile")
}

// UserProjects 用户项目列表缓存键
func (kb *KeyBuilder) UserProjects(userID string) string {
	return kb.User(userID, "projects")
}

// Project 项目相关缓存键
func (kb *KeyBuilder) Project(projectID string, suffix string) string {
	return fmt.Sprintf("project:%s:%s", projectID, suffix)
}

// ProjectInfo 项目信息缓存键
func (kb *KeyBuilder) ProjectInfo(projectID string) string {
	return kb.Project(projectID, "info")
}

// Session 会话相关缓存键
func (kb *KeyBuilder) Session(sessionID string) string {
	return fmt.Sprintf("session:%s", sessionID)
}

// RateLimit 限流相关缓存键
func (kb *KeyBuilder) RateLimit(clientIP string) string {
	return fmt.Sprintf("rate_limit:%s", clientIP)
}

// Cache 通用缓存键
func (kb *KeyBuilder) Cache(category string, identifier string) string {
	return fmt.Sprintf("cache:%s:%s", category, identifier)
}

// BuildKey 构建自定义缓存键
func BuildKey(parts ...string) string {
	return strings.Join(parts, ":")
}

// GetExpiration 获取缓存过期时间
func GetExpiration(expirationType string) time.Duration {
	switch expirationType {
	case "short":
		return 5 * time.Minute
	case "medium":
		return 30 * time.Minute
	case "long":
		return 2 * time.Hour
	case "day":
		return 24 * time.Hour
	case "week":
		return 7 * 24 * time.Hour
	default:
		return 30 * time.Minute
	}
}
