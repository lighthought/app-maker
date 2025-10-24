package services

import (
	"context"
	"fmt"
	"time"

	"github.com/lighthought/app-maker/shared-models/agent"
	"github.com/lighthought/app-maker/shared-models/cache"
	"github.com/lighthought/app-maker/shared-models/logger"
	"github.com/lighthought/app-maker/shared-models/utils"

	"gorm.io/gorm"
)

// EnvironmentService 环境检查服务接口
type EnvironmentService interface {
	// CheckDatabaseHealth 检查数据库连接状态
	CheckDatabaseHealth(ctx context.Context) (*agent.ServiceStatus, error)

	// CheckRedisHealth 检查 Redis 连接状态
	CheckRedisHealth(ctx context.Context) (*agent.ServiceStatus, error)
}

// environmentService 环境检查服务实现
type environmentService struct {
	agentInteractService AgentInteractService
	db                   *gorm.DB
	cacheInstance        cache.Cache
}

// NewEnvironmentService 创建环境检查服务
func NewEnvironmentService(agentInteractService AgentInteractService, db *gorm.DB, cacheInstance cache.Cache) EnvironmentService {
	return &environmentService{
		agentInteractService: agentInteractService,
		db:                   db,
		cacheInstance:        cacheInstance,
	}
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
