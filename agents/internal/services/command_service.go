package services

import (
	"context"
	"time"

	"app-maker-agents/internal/config"
	"app-maker-agents/internal/controllers"
)

// CommandResult 命令执行结果
type CommandResult struct {
	Success bool   `json:"success"`
	Output  string `json:"output"`
	Error   string `json:"error,omitempty"`
}

// CommandService 命令执行服务，负责按项目维护会话执行命令
type CommandService struct {
	sessions *controllers.SessionManager
	timeout  time.Duration
}

// NewCommandService 创建命令执行服务
func NewCommandService(cfg config.CommandConfig) *CommandService {
	return &CommandService{
		sessions: controllers.NewSessionManager(),
		timeout:  cfg.Timeout,
	}
}

// Execute 执行命令，使用项目级持久会话
func (s *CommandService) Execute(ctx context.Context, projectPath, command string, timeout time.Duration) CommandResult {
	if timeout == 0 {
		timeout = s.timeout
	}

	if projectPath == "" {
		return CommandResult{Success: false, Error: "projectPath 不能为空"}
	}

	res := s.sessions.Execute(projectPath, command, timeout)
	return CommandResult{Success: res.Success, Output: res.Stdout, Error: res.Err.Error()}
}
