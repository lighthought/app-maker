package services

import (
	"context"
	"fmt"
	"shared-models/logger"
)

type GitService interface {
	CommitAndPush(ctx context.Context, projectGuid string, commitMsg string) error
}

type gitService struct {
	commandService CommandService
}

// NewGitService 创建Git服务
func NewGitService(commandService CommandService) GitService {
	return &gitService{
		commandService: commandService,
	}
}

// CommitAndPush 提交并推送代码
func (s *gitService) CommitAndPush(ctx context.Context, projectGuid string, commitMsg string) error {
	logger.Info("开始提交并推送代码", logger.String("GUID", projectGuid))

	// 添加所有文件
	if result := s.commandService.SimpleExecute(ctx, projectGuid, "git", "add", "."); !result.Success {
		return fmt.Errorf("添加文件到Git失败: %s", result.Error)
	}

	// 检查是否有变更， 暂时跳过，在大量文件提交的时候，存在直接退出的情况
	if !s.hasChanges(ctx, projectGuid) {
		logger.Info("没有文件变更，跳过提交", logger.String("GUID", projectGuid))
		return nil
	}

	// 提交变更
	if commitMsg == "" {
		commitMsg = fmt.Sprintf("Auto commit by App Maker - %s", projectGuid)
	}

	if result := s.commandService.SimpleExecute(ctx, projectGuid, "git", "commit", "-m", commitMsg); !result.Success {
		return fmt.Errorf("提交代码失败: %s", result.Error)
	}

	// 推送到远程仓库
	if result := s.commandService.SimpleExecute(ctx, projectGuid, "git", "push", "-u", "origin", "master"); !result.Success {
		// 如果master分支不存在，尝试main分支
		if result := s.commandService.SimpleExecute(ctx, projectGuid, "git", "push", "-u", "origin", "main"); !result.Success {
			return fmt.Errorf("推送代码失败: %s", result.Error)
		}
	}

	logger.Info("项目文档、代码提交并推送完成", logger.String("GUID", projectGuid), logger.String("commitMessage", commitMsg))
	return nil
}

// hasChanges 检查是否有文件变更
func (s *gitService) hasChanges(ctx context.Context, projectDir string) bool {
	result := s.commandService.SimpleExecute(ctx, projectDir, "git", "diff", "--cached", "--quiet")
	return !result.Success
}
