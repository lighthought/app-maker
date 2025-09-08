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

// HealthCheck godoc
// @Summary 缓存健康检查
// @Description 检查缓存服务是否正常运行
// @Tags 缓存检查
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "成功响应"
// @Failure 500 {object} map[string]string "服务器内部错误"
// @Router /api/v1/cache/health [get]
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

// GetStats godoc
// @Summary 获取缓存统计信息
// @Description 获取缓存统计信息
// @Tags 缓存检查
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "成功响应"
// @Failure 500 {object} map[string]string "服务器内部错误"
// @Router /api/v1/cache/stats [get]
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

// GetMemoryUsage godoc
// @Summary 获取内存使用情况
// @Description 获取内存使用情况
// @Tags 缓存检查
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "成功响应"
// @Failure 500 {object} map[string]string "服务器内部错误"
// @Router /api/v1/cache/memory [get]
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

// GetKeyspaceStats godoc
// @Summary 获取键空间统计
// @Description 获取键空间统计
// @Tags 缓存检查
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "成功响应"
// @Failure 500 {object} map[string]string "服务器内部错误"
// @Router /api/v1/cache/keyspace [get]
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

// GetPerformanceMetrics godoc
// @Summary 获取性能指标
// @Description 获取性能指标
// @Tags 缓存检查
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "成功响应"
// @Failure 500 {object} map[string]string "服务器内部错误"
// @Router /api/v1/cache/performance [get]
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
