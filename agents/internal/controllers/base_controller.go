package controllers

import (
	"context"

	"app-maker-agents/internal/models"
)

// BaseController 为所有 Agent 控制器提供公共接口
type BaseController interface {
	Execute(ctx context.Context, input models.TaskContext, params map[string]interface{}) (map[string]interface{}, error)
}
