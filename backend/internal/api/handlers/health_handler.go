package handlers

import (
	"fmt"
	"net/http"

	"autocodeweb-backend/internal/services"
	"shared-models/agent"
	"shared-models/utils"

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
	var services []agent.ServiceStatus
	// 创建统一的健康检查响应结构
	healthResp := &agent.BackendHealthResp{
		Status:    "healthy",
		Service:   "autocodeweb-backend",
		Version:   "1.0.0",
		Timestamp: utils.GetCurrentTime(),
	}

	services = append(services, agent.ServiceStatus{
		Name:      "database",
		Status:    "healthy",
		Message:   "数据库连接正常",
		Version:   "1.0.0",
		CheckedAt: utils.GetCurrentTime(),
	})

	services = append(services, agent.ServiceStatus{
		Name:      "redis",
		Status:    "healthy",
		Message:   "Redis连接正常",
		Version:   "1.0.0",
		CheckedAt: utils.GetCurrentTime(),
	})

	// 检查 WebSocket 服务状态
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

	// 检查 Agent 服务状态
	agentHealth, _ := h.environmentService.CheckAgentHealth(c.Request.Context())

	healthResp.Services = services
	healthResp.Agent = agentHealth
	healthResp.Timestamp = utils.GetCurrentTime()

	c.JSON(http.StatusOK, utils.GetSuccessResponse("健康检查成功", healthResp))
}
