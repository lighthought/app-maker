package handlers

import (
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/lighthought/app-maker/shared-models/agent"
	"github.com/lighthought/app-maker/shared-models/cache"
	"github.com/lighthought/app-maker/shared-models/logger"
	"github.com/lighthought/app-maker/shared-models/utils"

	"github.com/gin-gonic/gin"
)

// HealthHandler 健康检查处理器
type HealthHandler struct {
	cacheInstance cache.Cache
}

// NewHealthHandler 创建健康检查处理器
func NewHealthHandler(cacheInstance cache.Cache) *HealthHandler {
	return &HealthHandler{
		cacheInstance: cacheInstance,
	}
}

const (
	VERSION_PARAMETER = "--version"
)

// CheckVersion 检查版本
// @Summary 检查版本
// @Description 检查服务是否正常运行，包括依赖服务状态
// @Tags 检查版本
// @Accept json
// @Produce json
// @Success 200 {object} common.Response "成功响应"
// @Failure 500 {object} common.ErrorResponse "服务器内部错误"
// @Router /api/v1/version [get]
func (h *HealthHandler) CheckVersion(c *gin.Context) {
	startTime := time.Now()
	logger.Info("开始 Agent 健康检查")

	// 缓存键
	cacheKey := "agent:health:check"
	cacheExpiration := 5 * time.Minute // 缓存5分钟

	// 尝试从缓存获取
	var resp agent.AgentHealthResp
	if h.cacheInstance != nil {
		err := h.cacheInstance.Get(cacheKey, &resp)
		if err == nil {
			logger.Info("从缓存返回健康检查结果", logger.String("duration", time.Since(startTime).String()))
			c.JSON(http.StatusOK, utils.GetSuccessResponse("App Maker Agents is running (cached)", resp))
			return
		}
	}

	// 缓存未命中，执行实际检查
	result := agent.AgentHealthResp{
		Status:    "running",
		Version:   "1.0.0",
		CheckedAt: utils.GetCurrentTime(),
	}

	var tools []agent.AgentToolInfo
	toolsStartTime := time.Now()

	// 检查 Node.js
	if version, err := checkCommandVersion("node", "--version"); err == nil {
		tools = append(tools, agent.AgentToolInfo{
			Name:    "node",
			Version: version,
		})
	}

	// 检查 npm
	if version, err := checkCommandVersion("npm", VERSION_PARAMETER); err == nil {
		tools = append(tools, agent.AgentToolInfo{
			Name:    "npm",
			Version: version,
		})
	}

	// 检查 npx
	if version, err := checkCommandVersion("npx", VERSION_PARAMETER); err == nil {
		tools = append(tools, agent.AgentToolInfo{
			Name:    "npx",
			Version: version,
		})
	}

	// 检查 git
	if version, err := checkCommandVersion("git", VERSION_PARAMETER); err == nil {
		tools = append(tools, agent.AgentToolInfo{
			Name:    "git",
			Version: strings.ReplaceAll(version, "git version ", ""),
		})
	}

	// 检查 claude-code
	if version, err := checkCommandVersion("claude", VERSION_PARAMETER); err == nil {
		tools = append(tools, agent.AgentToolInfo{
			Name:    "claude-code",
			Version: strings.ReplaceAll(version, " (Claude Code)", ""),
		})
	}

	// 检查 qwen-code
	if version, err := checkCommandVersion("qwen", VERSION_PARAMETER); err == nil {
		tools = append(tools, agent.AgentToolInfo{
			Name:    "qwen-code",
			Version: version,
		})
	}

	// 检查 gemini
	if version, err := checkCommandVersion("gemini", VERSION_PARAMETER); err == nil {
		tools = append(tools, agent.AgentToolInfo{
			Name:    "gemini",
			Version: version,
		})
	}

	toolsDuration := time.Since(toolsStartTime)
	logger.Info("工具版本检查完成", logger.String("duration", toolsDuration.String()))

	result.Tools = tools
	result.CheckedAt = utils.GetCurrentTime()

	// 将结果存入缓存
	if h.cacheInstance != nil {
		err := h.cacheInstance.Set(cacheKey, result, cacheExpiration)
		if err != nil {
			logger.Error("Failed to cache health check result", logger.String("error", err.Error()))
		}
		logger.Info("健康检查结果已缓存", logger.String("expiration", cacheExpiration.String()))
	}

	totalDuration := time.Since(startTime)
	logger.Info("Agent 健康检查完成",
		logger.String("total_duration", totalDuration.String()),
		logger.String("tools_duration", toolsDuration.String()))

	c.JSON(http.StatusOK, utils.GetSuccessResponse("App Maker Agents is running", result))
}

// CheckHealth 检查健康
// @Summary 检查健康
// @Description 检查服务是否正常运行，包括依赖服务状态
// @Tags 检查健康
// @Accept json
// @Produce json
// @Success 200 {object} common.Response "成功响应"
// @Failure 500 {object} common.ErrorResponse "服务器内部错误"
// @Router /api/v1/health [get]
func (h *HealthHandler) CheckHealth(c *gin.Context) {
	c.JSON(http.StatusOK, utils.GetSuccessResponse("App Maker Agents is running", nil))
}

// checkCommandVersion 检查命令版本
func checkCommandVersion(command string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}
