package handlers

import (
	"autocodeweb-backend/internal/models"
	"autocodeweb-backend/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthCheck 健康检查
// @Summary 健康检查
// @Description 检查服务是否正常运行
// @Tags 健康检查
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "成功响应"
// @Failure 500 {object} map[string]string "服务器内部错误"
// @Router /api/v1/health [get]
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, models.Response{
		Code:      models.SUCCESS_CODE,
		Message:   "AutoCodeWeb Backend is running",
		Data:      "1.0.0",
		Timestamp: utils.GetCurrentTime(),
	})
}
