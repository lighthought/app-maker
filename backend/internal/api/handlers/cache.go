package handlers

import (
	"net/http"

	"github.com/lighthought/app-maker/shared-models/cache"
	"github.com/lighthought/app-maker/shared-models/common"
	"github.com/lighthought/app-maker/shared-models/utils"

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
		c.JSON(http.StatusServiceUnavailable, utils.GetErrorResponse(common.SERVICE_UNAVAILABLE, "缓存服务不可用"))
		return
	}

	c.JSON(http.StatusOK, utils.GetSuccessResponse("缓存服务正常", nil))
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
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.INTERNAL_ERROR, "获取统计信息失败, "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.GetSuccessResponse("成功获取统计信息", stats))
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
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.INTERNAL_ERROR, "获取内存使用情况失败, "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.GetSuccessResponse("成功获取内存使用情况", memory))
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
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.INTERNAL_ERROR, "获取键空间统计失败, "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.GetSuccessResponse("成功获取键空间统计", stats))
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
		c.JSON(http.StatusOK, utils.GetErrorResponse(common.INTERNAL_ERROR, "获取性能指标失败, "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.GetSuccessResponse("成功获取性能指标", metrics))
}
