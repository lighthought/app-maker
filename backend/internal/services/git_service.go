package services

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"shared-models/logger"
	"shared-models/utils"
)

type GitService interface {
	SetupSSH() error
	GetPublicKey() (string, error)
	InitializeGit(ctx context.Context, config *GitConfig) (string, error)
	CommitAndPush(ctx context.Context, config *GitConfig) error
	Pull(ctx context.Context, config *GitConfig) error
}

// GitService Git操作服务
type gitService struct {
	gitlabURL      string
	gitlabUsername string
	gitlabEmail    string
	sshKeyPath     string
	sshKnownHosts  string
}

// NewGitService 创建Git服务
func NewGitService() GitService {
	return &gitService{
		gitlabURL:      utils.GetEnvOrDefault("GITLAB_URL", "git@gitlab.lighthought.com"),
		gitlabUsername: utils.GetEnvOrDefault("GITLAB_USERNAME", "John"),
		gitlabEmail:    utils.GetEnvOrDefault("GITLAB_EMAIL", "qqjack2012@gmail.com"),
		sshKeyPath:     utils.GetEnvOrDefault("SSH_KEY_PATH", "/home/appuser/.ssh/id_rsa"),
		sshKnownHosts:  utils.GetEnvOrDefault("SSH_KNOWN_HOSTS", "/home/appuser/.ssh/known_hosts"),
	}
}

// GitConfig Git配置
type GitConfig struct {
	UserID        string
	GUID          string
	ProjectPath   string
	CommitMessage string
}

// SetupSSH 配置SSH密钥和known_hosts
func (s *gitService) SetupSSH() error {
	logger.Info("配置SSH密钥")

	// 检查SSH密钥是否存在
	if _, err := os.Stat(s.sshKeyPath); os.IsNotExist(err) {
		logger.Info("SSH密钥不存在，生成新的密钥对")
		if err := s.generateSSHKey(); err != nil {
			return fmt.Errorf("生成SSH密钥失败: %w", err)
		}
	}

	// 配置SSH known_hosts
	if err := s.setupKnownHosts(); err != nil {
		return fmt.Errorf("配置known_hosts失败: %w", err)
	}

	logger.Info("SSH配置完成")
	return nil
}

// generateSSHKey 生成SSH密钥对
func (s *gitService) generateSSHKey() error {
	// 确保 .ssh 目录存在且有正确权限
	sshDir := filepath.Dir(s.sshKeyPath)
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		return fmt.Errorf("创建SSH目录失败: %w", err)
	}

	// 检查目录权限
	if stat, err := os.Stat(sshDir); err == nil {
		logger.Info("SSH目录权限检查",
			logger.String("path", sshDir),
			logger.String("mode", stat.Mode().String()),
		)
	}

	cmd := exec.CommandContext(context.Background(),
		"ssh-keygen", "-t", "rsa", "-b", "4096", "-f", s.sshKeyPath,
		"-C", s.gitlabEmail, "-N", "")
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Error("生成SSH密钥失败",
			logger.String("error", err.Error()),
			logger.String("output", string(output)),
			logger.String("sshDir", sshDir),
		)
		return fmt.Errorf("ssh-keygen: %w", err)
	}

	// 设置正确的文件权限
	if err := os.Chmod(s.sshKeyPath, 0600); err != nil {
		logger.Warn("设置私钥权限失败", logger.String("error", err.Error()))
	}

	publicKeyPath := s.sshKeyPath + ".pub"
	if err := os.Chmod(publicKeyPath, 0644); err != nil {
		logger.Warn("设置公钥权限失败", logger.String("error", err.Error()))
	}

	logger.Info("SSH密钥生成成功",
		logger.String("privateKey", s.sshKeyPath),
		logger.String("publicKey", publicKeyPath),
	)
	return nil
}

// setupKnownHosts 配置SSH known_hosts
func (s *gitService) setupKnownHosts() error {
	// 从GITLAB_URL提取主机名
	hostname := strings.TrimPrefix(s.gitlabURL, "git@")
	hostname = strings.TrimPrefix(hostname, "https://")
	hostname = strings.TrimPrefix(hostname, "http://")
	hostname = strings.TrimSuffix(hostname, ":22")
	hostname = strings.Split(hostname, ":")[0]

	// 使用 ssh-keyscan 获取远程主机密钥
	cmd := exec.CommandContext(context.Background(), "ssh-keyscan", "-H", hostname)
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Error("获取主机密钥失败",
			logger.String("hostname", hostname),
			logger.String("error", err.Error()),
			logger.String("output", string(output)),
		)
		return fmt.Errorf("ssh-keyscan: %w", err)
	}

	// 写入known_hosts文件
	if err := os.WriteFile(s.sshKnownHosts, output, 0644); err != nil {
		return fmt.Errorf("写入known_hosts失败: %w", err)
	}

	logger.Info("known_hosts配置成功",
		logger.String("hostname", hostname),
	)
	return nil
}

// GetPublicKey 获取SSH公钥内容
func (s *gitService) GetPublicKey() (string, error) {
	publicKeyPath := s.sshKeyPath + ".pub"
	content, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return "", fmt.Errorf("读取公钥失败: %w", err)
	}
	return string(content), nil
}

// InitializeGit 初始化Git仓库
func (s *gitService) InitializeGit(ctx context.Context, config *GitConfig) (string, error) {
	projectDir := config.ProjectPath

	logger.Info("初始化Git仓库",
		logger.String("GUID", config.GUID),
		logger.String("projectPath", projectDir),
	)

	// 首先配置SSH
	if err := s.SetupSSH(); err != nil {
		return "", fmt.Errorf("SSH配置失败: %w", err)
	}

	// 检查是否已经是Git仓库
	if s.isGitRepository(projectDir) {
		logger.Info("项目已经是Git仓库，跳过初始化",
			logger.String("GUID", config.GUID),
		)

		// 添加远程仓库
		remoteURL := s.buildRemoteURL(config.GUID)
		if err := s.runGitCommand(ctx, projectDir, "remote", "add", "origin", remoteURL); err != nil {
			return "", fmt.Errorf("添加远程仓库失败: %w", err)
		}
		return remoteURL, nil
	}

	// 初始化Git仓库
	if err := s.runGitCommand(ctx, projectDir, "init"); err != nil {
		return "", fmt.Errorf("初始化Git仓库失败: %w", err)
	}

	// 配置Git用户信息
	if err := s.runGitCommand(ctx, projectDir, "config", "user.name", s.gitlabUsername); err != nil {
		return "", fmt.Errorf("配置Git用户名失败: %w", err)
	}

	if err := s.runGitCommand(ctx, projectDir, "config", "user.email", s.gitlabEmail); err != nil {
		return "", fmt.Errorf("配置Git邮箱失败: %w", err)
	}

	// 添加远程仓库
	remoteURL := s.buildRemoteURL(config.GUID)
	if err := s.runGitCommand(ctx, projectDir, "remote", "add", "origin", remoteURL); err != nil {
		return "", fmt.Errorf("添加远程仓库失败: %w", err)
	}

	// 创建master分支
	if err := s.runGitCommand(ctx, projectDir, "branch", "-M", "master"); err != nil {
		return "", fmt.Errorf("创建master分支失败: %w", err)
	}

	logger.Info("Git仓库初始化完成",
		logger.String("GUID", config.GUID),
		logger.String("remoteURL", remoteURL),
	)

	return remoteURL, nil
}

// CommitAndPush 提交并推送代码
func (s *gitService) CommitAndPush(ctx context.Context, config *GitConfig) error {
	projectDir := config.ProjectPath

	logger.Info("开始提交并推送代码",
		logger.String("GUID", config.GUID),
		logger.String("projectPath", projectDir),
	)

	// 添加所有文件
	if err := s.runGitCommand(ctx, projectDir, "add", "."); err != nil {
		return fmt.Errorf("添加文件到Git失败: %w", err)
	}

	// 检查是否有变更
	if !s.hasChanges(projectDir) {
		logger.Info("没有文件变更，跳过提交",
			logger.String("GUID", config.GUID),
		)
		return nil
	}

	// 提交变更
	commitMsg := config.CommitMessage
	if commitMsg == "" {
		commitMsg = fmt.Sprintf("Auto commit by App Maker - %s", config.GUID)
	}

	if err := s.runGitCommand(ctx, projectDir, "commit", "-m", commitMsg); err != nil {
		return fmt.Errorf("提交代码失败: %w", err)
	}

	// 推送到远程仓库
	if err := s.pushToRemote(ctx, projectDir, config.GUID); err != nil {
		return fmt.Errorf("推送代码失败: %w", err)
	}

	logger.Info("代码提交并推送完成",
		logger.String("GUID", config.GUID),
		logger.String("commitMessage", commitMsg),
	)

	return nil
}

// pushToRemote 推送到远程仓库
func (s *gitService) pushToRemote(ctx context.Context, projectDir, guid string) error {
	// 构建SSH格式的远程URL
	remoteURL := s.buildRemoteURL(guid)

	// 设置远程URL
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

// buildRemoteURL 构建远程仓库URL（SSH格式）
func (s *gitService) buildRemoteURL(guid string) string {
	// 从GITLAB_URL提取主机名
	hostname := strings.TrimPrefix(s.gitlabURL, "git@")
	hostname = strings.TrimPrefix(hostname, "https://")
	hostname = strings.TrimPrefix(hostname, "http://")
	hostname = strings.TrimSuffix(hostname, ":22")
	hostname = strings.Split(hostname, ":")[0]

	return fmt.Sprintf("git@%s:app-maker/%s.git", hostname, guid)
}

// runGitCommand 执行Git命令
func (s *gitService) runGitCommand(ctx context.Context, workDir string, args ...string) error {
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
func (s *gitService) isGitRepository(projectDir string) bool {
	gitDir := filepath.Join(projectDir, ".git")
	_, err := os.Stat(gitDir)
	return err == nil
}

// Pull 拉取远程仓库的最新代码
func (s *gitService) Pull(ctx context.Context, config *GitConfig) error {
	projectDir := config.ProjectPath

	logger.Info("开始拉取远程仓库代码",
		logger.String("GUID", config.GUID),
		logger.String("projectPath", projectDir),
	)

	// 检查是否是Git仓库
	if !s.isGitRepository(projectDir) {
		return fmt.Errorf("项目目录不是Git仓库: %s", projectDir)
	}

	// 拉取远程代码
	if err := s.runGitCommand(ctx, projectDir, "pull", "origin", "master"); err != nil {
		// 如果master分支不存在，尝试main分支
		if err := s.runGitCommand(ctx, projectDir, "pull", "origin", "main"); err != nil {
			return fmt.Errorf("拉取远程代码失败: %w", err)
		}
	}

	logger.Info("远程代码拉取完成",
		logger.String("GUID", config.GUID),
	)

	return nil
}

// hasChanges 检查是否有文件变更
func (s *gitService) hasChanges(projectDir string) bool {
	cmd := exec.Command("git", "diff", "--cached", "--quiet")
	cmd.Dir = projectDir
	err := cmd.Run()
	return err != nil // 如果有错误说明有变更
}
