package services

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"app-maker-agents/internal/api/models"
	"app-maker-agents/internal/config"
)

// CommandService 命令执行服务，负责按项目维护会话执行命令
type commandService struct {
	timeout       time.Duration
	WorkspacePath string
}

type CommandService interface {
	SimpleExecute(ctx context.Context, subfolder, process string, arg ...string) models.CommandResult
	//GetWorkspacePath() string
}

// NewCommandService 创建命令执行服务
func NewCommandService(cfg config.CommandConfig, workspacePath string) CommandService {
	return &commandService{
		timeout:       cfg.Timeout,
		WorkspacePath: workspacePath,
	}
}

// 获取工作空间路径
// func (s *commandService) GetWorkspacePath() string {
// 	return s.WorkspacePath
// }

// SimpleExecute 直接执行命令，不使用 session 管理
func (s *commandService) SimpleExecute(ctx context.Context, subfolder, process string, arg ...string) models.CommandResult {
	// 根据操作系统选择 shell 和参数
	cmd := exec.Command(process, arg...)

	// 设置工作目录
	if subfolder != "" {
		cmd.Dir = filepath.Join(s.WorkspacePath, subfolder)
	} else {
		cmd.Dir = s.WorkspacePath
	}

	fmt.Printf("🔧 直接执行命令: %s (工作目录: %s, 超时: %v)\n", process, cmd.Dir, s.timeout)

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

	return models.CommandResult{
		Success: success,
		Output:  outputStr,
		Error:   errorMsg,
	}
}
