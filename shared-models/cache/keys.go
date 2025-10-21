package cache

import (
	"fmt"
	"strings"
)

// User 用户相关缓存键
func GetUserCacheKey(userID string, suffix string) string {
	return fmt.Sprintf("user:%s:%s", userID, suffix)
}

// UserProfile 用户资料缓存键
func GetUserProfileCacheKey(userID string) string {
	return GetUserCacheKey(userID, "profile")
}

// UserProjects 用户项目列表缓存键
func GetUserProjectsCacheKey(userID string) string {
	return GetUserCacheKey(userID, "projects")
}

// Project 项目相关缓存键
func GetProjectCacheKey(projectGuid string, suffix string) string {
	return fmt.Sprintf("project:%s:%s", projectGuid, suffix)
}

// ProjectInfo 项目信息缓存键
func GetProjectInfoCacheKey(projectGuid string) string {
	return GetProjectCacheKey(projectGuid, "info")
}

// ProjectAgentSession 项目Agent会话缓存键
func GetProjectAgentSessionCacheKey(projectGuid, agentType string) string {
	return GetProjectCacheKey(projectGuid, "sessions:"+agentType)
}

// RateLimit 限流相关缓存键
func GetRateLimitCacheKey(clientIP string) string {
	return fmt.Sprintf("rate_limit:%s", clientIP)
}

// Cache 通用缓存键
func GetCacheCacheKey(category string, identifier string) string {
	return fmt.Sprintf("cache:%s:%s", category, identifier)
}

// BuildCacheKey 构建自定义缓存键
func BuildCacheKey(parts ...string) string {
	return strings.Join(parts, ":")
}
