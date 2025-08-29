package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetTaskStatus 获取任务状态
func GetTaskStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "获取任务状态功能待实现",
	})
}
