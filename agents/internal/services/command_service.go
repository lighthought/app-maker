package services

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"app-maker-agents/internal/config"
)

// CommandResult 命令执行结果
type CommandResult struct {
	Success bool   `json:"success"`
	Output  string `json:"output"`
	Error   string `json:"error,omitempty"`
}

// CommandService 命令执行服务，负责按项目维护会话执行命令
type CommandService struct {
	timeout       time.Duration
	WorkspacePath string
}

// NewCommandService 创建命令执行服务
func NewCommandService(cfg config.CommandConfig, workspacePath string) *CommandService {
	return &CommandService{
		timeout:       cfg.Timeout,
		WorkspacePath: workspacePath,
	}
}

// SimpleExecute 直接执行命令，不使用 session 管理
func (s *CommandService) SimpleExecute(ctx context.Context, subfolder, process string, arg ...string) CommandResult {

	fmt.Printf("🔧 直接执行命令: %s (工作目录: %s, 超时: %v)\n", process, s.WorkspacePath, s.timeout)

	// 根据操作系统选择 shell 和参数
	cmd := exec.Command(process, arg...)

	// 设置工作目录
	if subfolder != "" {
		cmd.Dir = filepath.Join(s.WorkspacePath, subfolder)
	} else {
		cmd.Dir = s.WorkspacePath
	}

	// 设置环境变量 - 继承当前进程的环境变量
	cmd.Env = os.Environ()

	// 执行命令并获取输出
	output, err := cmd.CombinedOutput()
	outputStr := strings.TrimSpace(string(output))

	// 判断执行结果
	success := err == nil
	var errorMsg string
	if err != nil {
		errorMsg = err.Error()
	}

	if success {
		fmt.Printf("✅ 命令执行成功: %s %v\n", process, arg)
		if outputStr != "" {
			fmt.Printf("   输出: %s\n", outputStr)
		}
	} else {
		fmt.Printf("❌ 命令执行失败: %s %v\n", process, arg)
		fmt.Printf("   错误: %s\n", errorMsg)
		if outputStr != "" {
			fmt.Printf("   输出: %s\n", outputStr)
		}
	}

	return CommandResult{
		Success: success,
		Output:  outputStr,
		Error:   errorMsg,
	}
}
