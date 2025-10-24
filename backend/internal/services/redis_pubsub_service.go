package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/lighthought/app-maker/backend/internal/config"
	"github.com/lighthought/app-maker/shared-models/agent"
	"github.com/lighthought/app-maker/shared-models/common"
	"github.com/lighthought/app-maker/shared-models/logger"

	"github.com/redis/go-redis/v9"
)

// RedisPubSubService Redis Pub/Sub 服务接口
type RedisPubSubService interface {
	// 启动订阅服务
	Start(ctx context.Context) error
	// 停止订阅服务
	Stop() error
	// 处理 Agent 任务状态消息
	HandleAgentTaskStatus(ctx context.Context, message *agent.AgentTaskStatusMessage) error
}

// redisPubSubService Redis Pub/Sub 服务实现
type redisPubSubService struct {
	redisClient  *redis.Client
	asyncService AsyncClientService
	pubsub       *redis.PubSub
	stopChan     chan struct{}
}

// NewRedisPubSubService 创建 Redis Pub/Sub 服务
func NewRedisPubSubService(asyncService AsyncClientService, cfg *config.Config) RedisPubSubService {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       common.CacheDbDatabase,
	})
	return &redisPubSubService{
		redisClient:  redisClient,
		asyncService: asyncService,
		stopChan:     make(chan struct{}),
	}
}

// Start 启动订阅服务
func (s *redisPubSubService) Start(ctx context.Context) error {
	if s.redisClient == nil {
		return fmt.Errorf("redis client is nil")
	}

	// 订阅 Agent 任务状态频道
	s.pubsub = s.redisClient.Subscribe(ctx, common.RedisPubSubChannelAgentTask)

	// 启动消息处理协程
	go s.messageHandler(ctx)

	logger.Info("Redis Pub/Sub 服务已启动",
		logger.String("channel", common.RedisPubSubChannelAgentTask))

	return nil
}

// Stop 停止订阅服务
func (s *redisPubSubService) Stop() error {
	if s.pubsub != nil {
		if err := s.pubsub.Close(); err != nil {
			return fmt.Errorf("failed to close Pub/Sub connection: %s", err.Error())
		}
	}

	close(s.stopChan)
	logger.Info("Redis Pub/Sub service stopped")

	return nil
}

// messageHandler 消息处理协程
func (s *redisPubSubService) messageHandler(ctx context.Context) {
	for {
		select {
		case <-s.stopChan:
			logger.Info("Redis Pub/Sub 消息处理协程已停止")
			return
		case <-ctx.Done():
			logger.Info("Redis Pub/Sub 消息处理协程因上下文取消而停止")
			return
		default:
			// 接收消息
			msg, err := s.pubsub.ReceiveMessage(ctx)
			if err != nil {
				if err == redis.ErrClosed {
					logger.Info("Redis Pub/Sub 连接已关闭")
					return
				}
				logger.Error("接收 Redis Pub/Sub 消息失败",
					logger.String("error", err.Error()))
				time.Sleep(time.Second) // 等待一秒后重试
				continue
			}

			// 处理消息
			if err := s.processMessage(ctx, msg.Payload); err != nil {
				logger.Error("处理 Redis Pub/Sub 消息失败",
					logger.String("payload", msg.Payload),
					logger.String("error", err.Error()))
			}
		}
	}
}

// processMessage 处理接收到的消息
func (s *redisPubSubService) processMessage(ctx context.Context, payload string) error {
	var statusMsg agent.AgentTaskStatusMessage
	if err := json.Unmarshal([]byte(payload), &statusMsg); err != nil {
		return fmt.Errorf("failed to deserialize task status message: %s", err.Error())
	}

	logger.Info("received Agent task status message",
		logger.String("taskID", statusMsg.TaskID),
		logger.String("projectGuid", statusMsg.ProjectGuid),
		logger.String("agentType", statusMsg.AgentType),
		logger.String("status", statusMsg.Status),
		logger.String("message", statusMsg.Message))

	// 处理消息
	return s.HandleAgentTaskStatus(ctx, &statusMsg)
}

// HandleAgentTaskStatus 处理 Agent 任务状态消息
func (s *redisPubSubService) HandleAgentTaskStatus(ctx context.Context, message *agent.AgentTaskStatusMessage) error {
	if message == nil {
		return fmt.Errorf("message is nil")
	}

	// 根据任务状态处理
	switch message.Status {
	case common.CommonStatusInProgress:
		logger.Info("Agent task started executing",
			logger.String("taskID", message.TaskID),
			logger.String("projectGuid", message.ProjectGuid),
			logger.String("agentType", message.AgentType))
		return nil

	case common.CommonStatusDone:
		logger.Info("Agent task executed successfully",
			logger.String("taskID", message.TaskID),
			logger.String("projectGuid", message.ProjectGuid),
			logger.String("agentType", message.AgentType))

		_, err := s.asyncService.EnqueueAgentTaskResponseTask(message)
		if err != nil {
			return fmt.Errorf("failed to enqueue agent task response task: %s", err.Error())
		}
		return nil

	case common.CommonStatusFailed:
		logger.Error("Agent 任务执行失败",
			logger.String("taskID", message.TaskID),
			logger.String("projectGuid", message.ProjectGuid),
			logger.String("agentType", message.AgentType),
			logger.String("error", message.Message))

		_, err := s.asyncService.EnqueueAgentTaskResponseTask(message)
		if err != nil {
			return fmt.Errorf("failed to enqueue agent task response task: %s", err.Error())
		}
		return nil

	default:
		logger.Warn("未知的 Agent 任务状态",
			logger.String("taskID", message.TaskID),
			logger.String("status", message.Status))
	}

	return nil
}
