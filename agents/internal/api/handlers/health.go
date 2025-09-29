package handlers

import (
	"app-maker-agents/internal/utils"
	"net/http"
	"shared-models/agent"
	"shared-models/common"

	"github.com/gin-gonic/gin"
)

// HealthCheck 健康检查
// @Summary 健康检查
// @Description 检查服务是否正常运行
// @Tags 健康检查
// @Accept json
// @Produce json
// @Success 200 {object} common.Response "成功响应"
// @Failure 500 {object} common.ErrorResponse "服务器内部错误"
// @Router /api/v1/health [get]
func HealthCheck(c *gin.Context) {
	// TODO: 检查 git 环境、本地workspace是否存在、检查 git 命令 npm 命令、npx 命令 node 命令是否能够执行
	c.JSON(http.StatusOK, common.Response{
		Code:    common.SUCCESS_CODE,
		Message: "App Maker Agents is running",
		Data: agent.AgentHealthResp{
			Status:  "running",
			Version: "1.0.0",
		},
		Timestamp: utils.GetCurrentTime(),
	})
}
