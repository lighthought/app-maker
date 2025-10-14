package handlers

import (
	"os/exec"
	"strings"
	"net/http"
	"shared-models/agent"
	"shared-models/utils"

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
	result := agent.AgentHealthResp{
		Status:      "running",
		Version:     "1.0.0",
		Environment: make(map[string]string),
	}

	// 检查 Node.js
	if version, err := exec.Command("node", "--version").Output(); err == nil {
		result.Environment["node"] = strings.TrimSpace(string(version))
	} else {
		result.Environment["node"] = "not found"
	}

	// 检查 npm
	if version, err := exec.Command("npm", "--version").Output(); err == nil {
		result.Environment["npm"] = strings.TrimSpace(string(version))
	} else {
		result.Environment["npm"] = "not found"
	}

	// 检查 npx
	if version, err := exec.Command("npx", "--version").Output(); err == nil {
		result.Environment["npx"] = strings.TrimSpace(string(version))
	} else {
		result.Environment["npx"] = "not found"
	}

	// 检查 git
	if version, err := exec.Command("git", "--version").Output(); err == nil {
		result.Environment["git"] = strings.TrimSpace(string(version))
	} else {
		result.Environment["git"] = "not found"
	}

	// 检查 claude-code
	if version, err := exec.Command("claude", "--version").Output(); err == nil {
		result.Environment["claude-code"] = strings.TrimSpace(string(version))
	} else {
		result.Environment["claude-code"] = "not installed"
	}

	// 检查 qwen-code
	if version, err := exec.Command("qwen", "--version").Output(); err == nil {
		result.Environment["qwen-code"] = strings.TrimSpace(string(version))
	} else {
		result.Environment["qwen-code"] = "not installed"
	}

	// 检查 gemini
	if version, err := exec.Command("gemini", "--version").Output(); err == nil {
		result.Environment["gemini"] = strings.TrimSpace(string(version))
	} else {
		result.Environment["gemini"] = "not installed"
	}

	c.JSON(http.StatusOK, utils.GetSuccessResponse("App Maker Agents is running", result))
}
