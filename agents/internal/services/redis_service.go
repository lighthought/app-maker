package services

import (
	"fmt"

	"github.com/lighthought/app-maker/shared-models/agent"
	"github.com/lighthought/app-maker/shared-models/cache"
	"github.com/lighthought/app-maker/shared-models/common"
	"github.com/lighthought/app-maker/shared-models/logger"
	"github.com/lighthought/app-maker/shared-models/tasks"
	"github.com/lighthought/app-maker/shared-models/utils"
)

// PublishService 发布服务接口
type RedisService interface {
	// 发布任务状态消息到 Redis Pub/Sub
	PublishTaskStatus(taskPayload *tasks.AgentExecuteTaskPayload, taskID, status, message string) error

	// 把会话ID保存到缓存中
	SaveSessionByProjectGuid(projectGuid, agentType, sessionID string)

	// 根据projectGuid从缓存中获取 sessionId
	GetSessionByProjectGuid(projectGuid, agentType string) string
}

// publishService 发布服务实现
type redisService struct {
	cacheInstance cache.Cache
}

// NewPublishService 创建发布服务
func NewRedisService(cacheInstance cache.Cache) RedisService {
	return &redisService{
		cacheInstance: cacheInstance,
	}
}

// publishTaskStatus 发布任务状态消息到 Redis Pub/Sub
func (h *redisService) PublishTaskStatus(taskPayload *tasks.AgentExecuteTaskPayload, taskID, status, message string) error {
	if h.cacheInstance == nil {
		return fmt.Errorf("cache instance is nil")
	}

	statusMsg := &agent.AgentTaskStatusMessage{
		TaskID:      taskID,
		ProjectGuid: taskPayload.ProjectGUID,
		AgentType:   taskPayload.AgentType,
		Status:      status,
		Message:     message,
		DevStage:    string(taskPayload.DevStage),
		Timestamp:   utils.GetCurrentTime(),
	}

	bytes := statusMsg.ToBytes()
	// 发布到 Redis Pub/Sub
	err := h.cacheInstance.Publish(common.RedisPubSubChannelAgentTask, bytes)
	if err != nil {
		return fmt.Errorf("发布任务状态消息失败: %w", err)
	}

	logger.Info("任务状态消息已发布",
		logger.String("taskID", taskID),
		logger.String("projectGuid", taskPayload.ProjectGUID),
		logger.String("agentType", taskPayload.AgentType),
		logger.String("status", status),
		logger.String("message", message),
		logger.String("devStage", string(taskPayload.DevStage)),
	)
	return nil
}

// 把会话ID保存到缓存中
func (h *redisService) SaveSessionByProjectGuid(projectGuid, agentType, sessionID string) {
	if h.cacheInstance == nil {
		logger.Warn("Redis client is nil, cannot save session",
			logger.String("projectGuid", projectGuid),
			logger.String("sessionID", sessionID))
		return
	}

	if projectGuid == "" || sessionID == "" {
		logger.Warn("Invalid parameters for saving session",
			logger.String("projectGuid", projectGuid),
			logger.String("sessionID", sessionID))
		return
	}

	key := cache.GetProjectAgentSessionCacheKey(projectGuid, agentType)
	// 设置过期时间为 24 小时，避免会话数据永久占用内存
	expiration := common.CacheExpirationDay

	err := h.cacheInstance.Set(key, sessionID, expiration)
	if err != nil {
		logger.Error("Failed to save session to cache",
			logger.String("projectGuid", projectGuid),
			logger.String("sessionID", sessionID),
			logger.String("error", err.Error()),
		)
		return
	}

	logger.Info("Saved session for project",
		logger.String("projectGuid", projectGuid),
		logger.String("sessionID", sessionID),
		logger.String("expiration", expiration.String()),
	)
}

// 根据projectGuid从缓存中获取 sessionId
func (h *redisService) GetSessionByProjectGuid(projectGuid, agentType string) string {
	if h.cacheInstance == nil {
		logger.Warn("Cache instance is nil, cannot get session", logger.String("projectGuid", projectGuid))
		return ""
	}

	key := cache.GetProjectAgentSessionCacheKey(projectGuid, agentType)
	var sessionID string
	err := h.cacheInstance.Get(key, &sessionID)
	if err != nil {
		logger.Error("Failed to get session from Redis",
			logger.String("projectGuid", projectGuid),
			logger.String("error", err.Error()))
		return ""
	}

	logger.Info("Retrieved session for project",
		logger.String("projectGuid", projectGuid),
		logger.String("sessionID", sessionID))
	return sessionID
}
