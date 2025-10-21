package handlers

import (
	"net/http"
	"os/exec"
	"strings"

	"shared-models/agent"
	"shared-models/utils"

	"github.com/gin-gonic/gin"
)

// checkCommandVersion 检查命令版本
func checkCommandVersion(command string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// HealthCheck 健康检查
// @Summary 健康检查
// @Description 检查服务是否正常运行，包括依赖服务状态
// @Tags 健康检查
// @Accept json
// @Produce json
// @Success 200 {object} common.Response "成功响应"
// @Failure 500 {object} common.ErrorResponse "服务器内部错误"
// @Router /api/v1/health [get]
func HealthCheck(c *gin.Context) {
	result := agent.AgentHealthResp{
		Status:    "running",
		Version:   "1.0.0",
		CheckedAt: utils.GetCurrentTime(),
	}

	var tools []agent.AgentToolInfo

	// 检查 Node.js
	if version, err := checkCommandVersion("node", "--version"); err == nil {
		tools = append(tools, agent.AgentToolInfo{
			Name:    "node",
			Version: version,
		})
	}

	// 检查 npm
	if version, err := checkCommandVersion("npm", "--version"); err == nil {
		tools = append(tools, agent.AgentToolInfo{
			Name:    "npm",
			Version: version,
		})
	}

	// 检查 npx
	if version, err := checkCommandVersion("npx", "--version"); err == nil {
		tools = append(tools, agent.AgentToolInfo{
			Name:    "npx",
			Version: version,
		})
	}

	// 检查 git
	if version, err := checkCommandVersion("git", "--version"); err == nil {
		tools = append(tools, agent.AgentToolInfo{
			Name:    "git",
			Version: strings.ReplaceAll(version, "git version ", ""),
		})
	}

	// 检查 claude-code
	if version, err := checkCommandVersion("claude", "--version"); err == nil {
		tools = append(tools, agent.AgentToolInfo{
			Name:    "claude-code",
			Version: strings.ReplaceAll(version, " (Claude Code)", ""),
		})
	}

	// 检查 qwen-code
	if version, err := checkCommandVersion("qwen", "--version"); err == nil {
		tools = append(tools, agent.AgentToolInfo{
			Name:    "qwen-code",
			Version: version,
		})
	}

	// 检查 gemini
	if version, err := checkCommandVersion("gemini", "--version"); err == nil {
		tools = append(tools, agent.AgentToolInfo{
			Name:    "gemini",
			Version: version,
		})
	}

	result.Tools = tools
	result.CheckedAt = utils.GetCurrentTime()

	c.JSON(http.StatusOK, utils.GetSuccessResponse("App Maker Agents is running", result))
}
