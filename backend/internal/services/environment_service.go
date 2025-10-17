package services

import (
	"context"
	"fmt"
	"time"

	"shared-models/agent"
	"shared-models/client"
	"shared-models/logger"
	"shared-models/utils"
)

// EnvironmentService 环境检查服务接口
type EnvironmentService interface {
	// CheckAgentHealth 检查 Agent 服务健康状态
	CheckAgentHealth(ctx context.Context) (*agent.AgentHealthResp, error)

	// IsAgentHealthy 检查 Agent 是否健康（简化版，返回布尔值）
	IsAgentHealthy(ctx context.Context) bool
}

// environmentService 环境检查服务实现
type environmentService struct {
	agentsURL string
	timeout   time.Duration
}

// NewEnvironmentService 创建环境检查服务
func NewEnvironmentService(agentsURL string) EnvironmentService {
	return &environmentService{
		agentsURL: agentsURL,
		timeout:   10 * time.Second, // 默认10秒超时
	}
}

// CheckAgentHealth 检查 Agent 服务健康状态
func (s *environmentService) CheckAgentHealth(ctx context.Context) (*agent.AgentHealthResp, error) {
	if s.agentsURL == "" {
		s.agentsURL = utils.GetEnvOrDefault("AGENTS_SERVER_URL", "http://host.docker.internal:8088")
	}

	logger.Info("开始检查 Agent 服务健康状态",
		logger.String("agentsURL", s.agentsURL))

	agentClient := client.NewAgentClient(s.agentsURL, s.timeout)

	healthResp, err := agentClient.HealthCheck(ctx)
	if err != nil {
		logger.Error("Agent 健康检查失败",
			logger.String("agentsURL", s.agentsURL),
			logger.String("error", err.Error()))
		return nil, fmt.Errorf("Agent 服务不可用: %w", err.Error())
	}

	logger.Info("Agent 健康检查成功",
		logger.String("status", healthResp.Status),
		logger.String("version", healthResp.Version))

	return healthResp, nil
}

// checkAgentHealthWithTimeout 带超时的 Agent 健康检查
func (s *environmentService) checkAgentHealthWithTimeout(ctx context.Context, timeout time.Duration) (*agent.AgentHealthResp, error) {
	// 创建带超时的上下文
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return s.CheckAgentHealth(timeoutCtx)
}

// IsAgentHealthy 检查 Agent 是否健康（简化版，返回布尔值）
func (s *environmentService) IsAgentHealthy(ctx context.Context) bool {
	// 使用较短的超时时间进行快速检查
	healthResp, err := s.checkAgentHealthWithTimeout(ctx, 5*time.Second)
	if err != nil {
		logger.Warn("Agent 服务健康检查失败",
			logger.String("error", err.Error()))
		return false
	}

	// 检查状态是否为 "healthy" 或 "ok"
	return healthResp.Status == "healthy" || healthResp.Status == "ok" || healthResp.Status == "running"
}
