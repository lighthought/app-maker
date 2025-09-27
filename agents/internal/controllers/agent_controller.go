package controllers

import (
	"context"

	"app-maker-agents/internal/models"
)

// PlaceholderController 提供默认实现，后续可替换为真实的 Agent 逻辑
type PlaceholderController struct {
	agentType models.AgentType
}

// NewPlaceholderController 创建占位控制器
func NewPlaceholderController(agentType models.AgentType) *PlaceholderController {
	return &PlaceholderController{agentType: agentType}
}

// Execute 暂时返回固定信息，待接入真实 Agent
func (c *PlaceholderController) Execute(ctx context.Context, input models.TaskContext, params map[string]interface{}) (map[string]interface{}, error) {
	return map[string]interface{}{
		"agentType": c.agentType,
		"message":   "该 agent 控制器尚未实现具体逻辑",
		"context":   input,
		"params":    params,
	}, nil
}
