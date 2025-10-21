package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/lighthought/app-maker/shared-models/agent"
	"github.com/lighthought/app-maker/shared-models/logger"
	"github.com/lighthought/app-maker/shared-models/utils"

	"github.com/lighthought/app-maker/backend/internal/services"

	"github.com/gin-gonic/gin"
)

// HealthHandler 健康检查处理器
type HealthHandler struct {
	environmentService services.EnvironmentService
	webSocketService   services.WebSocketService
}

// NewHealthHandler 创建健康处理器实例
func NewHealthHandler(environmentService services.EnvironmentService, webSocketService services.WebSocketService) *HealthHandler {
	return &HealthHandler{
		environmentService: environmentService,
		webSocketService:   webSocketService,
	}
}

// HealthCheck 健康检查
// @Summary 健康检查
// @Description 检查服务是否正常运行，包括依赖服务状态
// @Tags 健康检查
// @Accept json
// @Produce json
// @Success 200 {object} agent.BackendHealthResp "成功响应"
// @Failure 500 {object} map[string]string "服务器内部错误"
// @Router /api/v1/health [get]
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	startTime := time.Now()
	logger.Info("开始健康检查")

	var services []agent.ServiceStatus
	// 创建统一的健康检查响应结构
	healthResp := &agent.BackendHealthResp{
		Status:    "healthy",
		Service:   "autocodeweb-backend",
		Version:   "1.0.0",
		Timestamp: utils.GetCurrentTime(),
	}

	// 检查数据库连接状态
	dbStartTime := time.Now()
	dbStatus, err := h.environmentService.CheckDatabaseHealth(c.Request.Context())
	if err != nil {
		logger.Error("数据库健康检查失败", logger.String("error", err.Error()))
		healthResp.Status = "degraded"
	}
	services = append(services, *dbStatus)
	dbDuration := time.Since(dbStartTime)
	logger.Info("数据库健康检查完成", logger.String("duration", dbDuration.String()))

	// 检查 Redis 连接状态
	redisStartTime := time.Now()
	redisStatus, err := h.environmentService.CheckRedisHealth(c.Request.Context())
	if err != nil {
		logger.Error("Redis健康检查失败", logger.String("error", err.Error()))
		healthResp.Status = "degraded"
	}
	services = append(services, *redisStatus)
	redisDuration := time.Since(redisStartTime)
	logger.Info("Redis健康检查完成", logger.String("duration", redisDuration.String()))

	// 检查 WebSocket 服务状态
	wsStartTime := time.Now()
	wsStats := h.webSocketService.GetStats()
	if wsStats != nil {
		services = append(services, agent.ServiceStatus{
			Name:      "websocket",
			Status:    "healthy",
			Message:   fmt.Sprintf("WebSocket服务正常运行，当前连接数: %v", wsStats["total_clients"]),
			Version:   "1.0.0",
			CheckedAt: utils.GetCurrentTime(),
		})
	}
	wsDuration := time.Since(wsStartTime)
	logger.Info("WebSocket健康检查完成", logger.String("duration", wsDuration.String()))

	// 检查 Agent 服务状态
	agentStartTime := time.Now()
	agentHealth, err := h.environmentService.CheckAgentHealth(c.Request.Context())
	if err != nil {
		logger.Error("Agent健康检查失败", logger.String("error", err.Error()))
		healthResp.Status = "degraded"
	}
	agentDuration := time.Since(agentStartTime)
	logger.Info("Agent健康检查完成", logger.String("duration", agentDuration.String()))

	healthResp.Services = services
	healthResp.Agent = agentHealth
	healthResp.Timestamp = utils.GetCurrentTime()

	totalDuration := time.Since(startTime)
	logger.Info("健康检查完成",
		logger.String("total_duration", totalDuration.String()),
		logger.String("db_duration", dbDuration.String()),
		logger.String("redis_duration", redisDuration.String()),
		logger.String("ws_duration", wsDuration.String()),
		logger.String("agent_duration", agentDuration.String()))

	c.JSON(http.StatusOK, utils.GetSuccessResponse("健康检查成功", healthResp))
}
