package models

import (
	"time"
)

type AgentType string
type DevStage string
type TaskStatus string

// TODO: 公共模型
// GitConfig Git配置
type GitConfig struct {
	UserID        string
	GUID          string
	ProjectPath   string
	CommitMessage string
}

type ExecResult struct {
	Success bool
	Stdout  string
	Stderr  string
	Err     error
}

type ExecRequest struct {
	Command string
	Timeout time.Duration
	Done    chan ExecResult
}
