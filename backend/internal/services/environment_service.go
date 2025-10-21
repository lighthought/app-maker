package services

import (
	"context"
	"fmt"
	"time"

	"shared-models/agent"
	"shared-models/cache"
	"shared-models/client"
	"shared-models/logger"
	"shared-models/utils"

	"gorm.io/gorm"
)

// EnvironmentService 环境检查服务接口
type EnvironmentService interface {
	// CheckAgentHealth 检查 Agent 服务健康状态
	CheckAgentHealth(ctx context.Context) (*agent.AgentHealthResp, error)

	// IsAgentHealthy 检查 Agent 是否健康（简化版，返回布尔值）
	IsAgentHealthy(ctx context.Context) bool

	// CheckDatabaseHealth 检查数据库连接状态
	CheckDatabaseHealth(ctx context.Context) (*agent.ServiceStatus, error)

	// CheckRedisHealth 检查 Redis 连接状态
	CheckRedisHealth(ctx context.Context) (*agent.ServiceStatus, error)
}

// environmentService 环境检查服务实现
type environmentService struct {
	agentsURL     string
	db            *gorm.DB
	cacheInstance cache.Cache
	timeout       time.Duration
}

// NewEnvironmentService 创建环境检查服务
func NewEnvironmentService(agentsURL string, db *gorm.DB, cacheInstance cache.Cache) EnvironmentService {
	return &environmentService{
		agentsURL:     agentsURL,
		db:            db,
		cacheInstance: cacheInstance,
		timeout:       10 * time.Second, // 默认10秒超时
	}
}

// CheckAgentHealth 检查 Agent 服务健康状态
func (s *environmentService) CheckAgentHealth(ctx context.Context) (*agent.AgentHealthResp, error) {
	if s.agentsURL == "" {
		s.agentsURL = utils.GetEnvOrDefault("AGENTS_SERVER_URL", "http://localhost:8088")
	}

	logger.Info("开始检查 Agent 服务健康状态",
		logger.String("agentsURL", s.agentsURL))

	agentClient := client.NewAgentClient(s.agentsURL, s.timeout)

	healthResp, err := agentClient.HealthCheck(ctx)
	if err != nil {
		logger.Error("agent health check failed",
			logger.String("agentsURL", s.agentsURL),
			logger.String("error", err.Error()))
		return nil, fmt.Errorf("agent server is not available: %s", err.Error())
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

// CheckDatabaseHealth 检查数据库连接状态
func (s *environmentService) CheckDatabaseHealth(ctx context.Context) (*agent.ServiceStatus, error) {
	if s.db == nil {
		return &agent.ServiceStatus{
			Name:      "database",
			Status:    "unhealthy",
			Message:   "数据库连接未配置",
			Version:   "unknown",
			CheckedAt: utils.GetCurrentTime(),
		}, fmt.Errorf("database connection not configured")
	}

	// 创建带超时的上下文
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// 执行简单的数据库查询来检查连接
	var result int
	err := s.db.WithContext(timeoutCtx).Raw("SELECT 1").Scan(&result).Error

	if err != nil {
		logger.Error("数据库健康检查失败", logger.String("error", err.Error()))
		return &agent.ServiceStatus{
			Name:      "database",
			Status:    "unhealthy",
			Message:   fmt.Sprintf("数据库连接失败: %s", err.Error()),
			Version:   "unknown",
			CheckedAt: utils.GetCurrentTime(),
		}, err
	}

	logger.Info("数据库健康检查成功")
	return &agent.ServiceStatus{
		Name:      "database",
		Status:    "healthy",
		Message:   "数据库连接正常",
		Version:   "1.0.0",
		CheckedAt: utils.GetCurrentTime(),
	}, nil
}

// CheckRedisHealth 检查 Redis 连接状态
func (s *environmentService) CheckRedisHealth(ctx context.Context) (*agent.ServiceStatus, error) {
	if s.cacheInstance == nil {
		return &agent.ServiceStatus{
			Name:      "redis",
			Status:    "unhealthy",
			Message:   "Redis连接未配置",
			Version:   "unknown",
			CheckedAt: utils.GetCurrentTime(),
		}, fmt.Errorf("redis connection not configured")
	}

	// 执行简单的 Redis 操作来检查连接
	if err := s.cacheInstance.Ping(); err != nil {
		logger.Error("Redis健康检查失败", logger.String("error", err.Error()))
		return &agent.ServiceStatus{
			Name:      "redis",
			Status:    "unhealthy",
			Message:   fmt.Sprintf("Redis连接失败: %s", err.Error()),
			Version:   "unknown",
			CheckedAt: utils.GetCurrentTime(),
		}, err
	}

	logger.Info("Redis健康检查成功")
	return &agent.ServiceStatus{
		Name:      "redis",
		Status:    "healthy",
		Message:   "Redis连接正常",
		Version:   "1.0.0",
		CheckedAt: utils.GetCurrentTime(),
	}, nil
}
