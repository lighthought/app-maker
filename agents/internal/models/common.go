package models

import (
	"time"
)

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
