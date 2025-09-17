package services

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"autocodeweb-backend/internal/utils"
	"autocodeweb-backend/pkg/logger"
)

// GitService Git操作服务
type GitService struct {
	gitlabURL      string
	gitlabUsername string
	gitlabEmail    string
	gitlabToken    string
}

// NewGitService 创建Git服务
func NewGitService() *GitService {
	return &GitService{
		gitlabURL:      utils.GetEnvOrDefault("GITLAB_URL", "http://gitlab.app-maker.localhost"),
		gitlabUsername: utils.GetEnvOrDefault("GITLAB_USERNAME", "John"),
		gitlabEmail:    utils.GetEnvOrDefault("GITLAB_EMAIL", "qqjack2012@gmail.com"),
		gitlabToken:    utils.GetEnvOrDefault("GITLAB_TOKEN", "glpat-_S4kLnmj3UJNvjFMqvG_b286MQp1OjMH.01.0w18kp84i"),
	}
}

// GitConfig Git配置
type GitConfig struct {
	UserID        string
	ProjectID     string
	ProjectPath   string
	CommitMessage string
}

// InitializeGit 初始化Git仓库
func (s *GitService) InitializeGit(ctx context.Context, config *GitConfig) error {
	projectDir := config.ProjectPath

	logger.Info("初始化Git仓库",
		logger.String("projectID", config.ProjectID),
		logger.String("projectPath", projectDir),
	)

	// 检查是否已经是Git仓库
	if s.isGitRepository(projectDir) {
		logger.Info("项目已经是Git仓库，跳过初始化",
			logger.String("projectID", config.ProjectID),
		)
		return nil
	}

	// 初始化Git仓库
	if err := s.runGitCommand(ctx, projectDir, "init"); err != nil {
		return fmt.Errorf("初始化Git仓库失败: %w", err)
	}

	// 配置Git用户信息
	if err := s.runGitCommand(ctx, projectDir, "config", "user.name", s.gitlabUsername); err != nil {
		return fmt.Errorf("配置Git用户名失败: %w", err)
	}

	if err := s.runGitCommand(ctx, projectDir, "config", "user.email", s.gitlabEmail); err != nil {
		return fmt.Errorf("配置Git邮箱失败: %w", err)
	}

	// 添加远程仓库
	remoteURL := s.buildRemoteURL(config.ProjectID)
	if err := s.runGitCommand(ctx, projectDir, "remote", "add", "origin", remoteURL); err != nil {
		return fmt.Errorf("添加远程仓库失败: %w", err)
	}

	// 创建master分支
	if err := s.runGitCommand(ctx, projectDir, "branch", "-M", "master"); err != nil {
		return fmt.Errorf("创建master分支失败: %w", err)
	}

	logger.Info("Git仓库初始化完成",
		logger.String("projectID", config.ProjectID),
		logger.String("remoteURL", remoteURL),
	)

	return nil
}

// CommitAndPush 提交并推送代码
func (s *GitService) CommitAndPush(ctx context.Context, config *GitConfig) error {
	projectDir := config.ProjectPath

	logger.Info("开始提交并推送代码",
		logger.String("projectID", config.ProjectID),
		logger.String("projectPath", projectDir),
	)

	// 添加所有文件
	if err := s.runGitCommand(ctx, projectDir, "add", "."); err != nil {
		return fmt.Errorf("添加文件到Git失败: %w", err)
	}

	// 检查是否有变更
	if !s.hasChanges(projectDir) {
		logger.Info("没有文件变更，跳过提交",
			logger.String("projectID", config.ProjectID),
		)
		return nil
	}

	// 提交变更
	commitMsg := config.CommitMessage
	if commitMsg == "" {
		commitMsg = fmt.Sprintf("Auto commit by App Maker - %s", config.ProjectID)
	}

	if err := s.runGitCommand(ctx, projectDir, "commit", "-m", commitMsg); err != nil {
		return fmt.Errorf("提交代码失败: %w", err)
	}

	// 推送到远程仓库
	if err := s.pushToRemote(ctx, projectDir, config.ProjectID); err != nil {
		return fmt.Errorf("推送代码失败: %w", err)
	}

	logger.Info("代码提交并推送完成",
		logger.String("projectID", config.ProjectID),
		logger.String("commitMessage", commitMsg),
	)

	return nil
}

// pushToRemote 推送到远程仓库
func (s *GitService) pushToRemote(ctx context.Context, projectDir, projectID string) error {
	// 构建带认证的远程URL
	remoteURL := s.buildAuthenticatedRemoteURL(projectID)

	// 设置远程URL（包含认证信息）
	if err := s.runGitCommand(ctx, projectDir, "remote", "set-url", "origin", remoteURL); err != nil {
		return fmt.Errorf("设置远程URL失败: %w", err)
	}

	// 推送到远程仓库
	if err := s.runGitCommand(ctx, projectDir, "push", "-u", "origin", "master"); err != nil {
		// 如果master分支不存在，尝试main分支
		if err := s.runGitCommand(ctx, projectDir, "push", "-u", "origin", "main"); err != nil {
			return fmt.Errorf("推送代码失败: %w", err)
		}
	}

	return nil
}

// buildRemoteURL 构建远程仓库URL
func (s *GitService) buildRemoteURL(projectID string) string {
	return fmt.Sprintf("%s/app-maker/%s.git", s.gitlabURL, projectID)
}

// buildAuthenticatedRemoteURL 构建带认证的远程仓库URL
func (s *GitService) buildAuthenticatedRemoteURL(projectID string) string {
	if s.gitlabToken != "" {
		// 使用令牌认证
		baseURL := strings.TrimPrefix(s.gitlabURL, "http://")
		return fmt.Sprintf("http://%s:%s@%s/app-maker/%s.git",
			s.gitlabUsername, s.gitlabToken, baseURL, projectID)
	}

	// 没有令牌时使用普通URL
	return s.buildRemoteURL(projectID)
}

// runGitCommand 执行Git命令
func (s *GitService) runGitCommand(ctx context.Context, workDir string, args ...string) error {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = workDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Error("Git命令执行失败",
			logger.String("command", fmt.Sprintf("git %s", strings.Join(args, " "))),
			logger.String("workDir", workDir),
			logger.String("error", err.Error()),
			logger.String("output", string(output)),
		)
		return fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}

	logger.Debug("Git命令执行成功",
		logger.String("command", fmt.Sprintf("git %s", strings.Join(args, " "))),
		logger.String("workDir", workDir),
		logger.String("output", string(output)),
	)

	return nil
}

// isGitRepository 检查是否是Git仓库
func (s *GitService) isGitRepository(projectDir string) bool {
	gitDir := filepath.Join(projectDir, ".git")
	_, err := os.Stat(gitDir)
	return err == nil
}

// hasChanges 检查是否有文件变更
func (s *GitService) hasChanges(projectDir string) bool {
	cmd := exec.Command("git", "diff", "--cached", "--quiet")
	cmd.Dir = projectDir
	err := cmd.Run()
	return err != nil // 如果有错误说明有变更
}
