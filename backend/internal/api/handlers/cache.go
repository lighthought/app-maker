package handlers

import (
	"net/http"

	"autocodeweb-backend/pkg/cache"

	"github.com/gin-gonic/gin"
)

// CacheHandler 缓存处理器
type CacheHandler struct {
	cache   cache.Cache
	monitor *cache.Monitor
}

// NewCacheHandler 创建新的缓存处理器
func NewCacheHandler(cache cache.Cache, monitor *cache.Monitor) *CacheHandler {
	return &CacheHandler{
		cache:   cache,
		monitor: monitor,
	}
}

// HealthCheck 缓存健康检查
func (h *CacheHandler) HealthCheck(c *gin.Context) {
	if err := h.cache.Ping(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "unhealthy",
			"error":   err.Error(),
			"message": "缓存服务不可用",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"message": "缓存服务正常",
	})
}

// GetStats 获取缓存统计信息
func (h *CacheHandler) GetStats(c *gin.Context) {
	stats, err := h.monitor.GetFullStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"message": "获取统计信息失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   stats,
	})
}

// GetMemoryUsage 获取内存使用情况
func (h *CacheHandler) GetMemoryUsage(c *gin.Context) {
	memory, err := h.monitor.GetMemoryUsage()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"message": "获取内存使用情况失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   memory,
	})
}

// GetKeyspaceStats 获取键空间统计
func (h *CacheHandler) GetKeyspaceStats(c *gin.Context) {
	stats, err := h.monitor.GetKeyspaceStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"message": "获取键空间统计失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   stats,
	})
}

// GetPerformanceMetrics 获取性能指标
func (h *CacheHandler) GetPerformanceMetrics(c *gin.Context) {
	metrics, err := h.monitor.GetPerformanceMetrics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"message": "获取性能指标失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   metrics,
	})
}
