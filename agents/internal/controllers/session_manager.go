package controllers

import (
	"bufio"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
	"time"

	"app-maker-agents/internal/models"
	"app-maker-agents/internal/utils"
	"app-maker-agents/pkg/logger"
)

// SessionManager 维护项目级命令执行会话
type SessionManager struct {
	workspacePath string
	sessions      sync.Map // map[string]*session
}

// NewSessionManager 创建 SessionManager
func NewSessionManager(workspacePath string) *SessionManager {
	return &SessionManager{workspacePath: workspacePath}
}

// Execute 在指定项目会话中执行命令
func (m *SessionManager) Execute(projectPath, command string, timeout time.Duration) models.ExecResult {
	sess, err := m.getOrCreateSession(projectPath)
	if err != nil {
		logger.Error("创建项目会话失败", logger.String("projectPath", projectPath), logger.ErrorField(err))
		return models.ExecResult{Success: false, Err: err}
	}

	req := models.ExecRequest{
		Command: command,
		Timeout: timeout,
		Done:    make(chan models.ExecResult, 1),
	}

	sess.Queue <- req
	return <-req.Done
}

func (m *SessionManager) getOrCreateSession(projectPath string) (*Session, error) {
	if val, ok := m.sessions.Load(projectPath); ok {
		logger.Info("获取项目会话成功, 直接返回", logger.String("projectPath", projectPath))
		return val.(*Session), nil
	}

	// Double-check locking
	var createErr error
	var sess *Session
	m.sessions.LoadOrStore(projectPath, &Session{})
	val, _ := m.sessions.Load(projectPath)
	sess = val.(*Session)

	sess.Mutex.Lock()
	defer sess.Mutex.Unlock()

	if sess.Cmd != nil {
		return sess, nil
	}

	if !utils.IsDirectoryExists(m.workspacePath) {
		utils.EnsureDirectoryExists(m.workspacePath)
	}

	cmd, err := m.createShell(projectPath)
	if err != nil {
		createErr = err
		if createErr != nil {
			logger.Error("创建项目会话失败", logger.String("projectPath", projectPath), logger.ErrorField(createErr))
		}
		return nil, createErr
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		createErr = err
		logger.Error("创建项目会话失败", logger.String("projectPath", projectPath), logger.ErrorField(err))
		return nil, createErr
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		createErr = err
		logger.Error("创建项目会话失败", logger.String("projectPath", projectPath), logger.ErrorField(err))
		return nil, createErr
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		createErr = err
		logger.Error("创建项目会话失败", logger.String("projectPath", projectPath), logger.ErrorField(err))
		return nil, createErr
	}

	bufStdout := bufio.NewReader(stdout)
	bufStderr := bufio.NewReader(stderr)

	if err := cmd.Start(); err != nil {
		createErr = err
		logger.Error("创建项目会话失败", logger.String("projectPath", projectPath), logger.ErrorField(err))
		return nil, createErr
	}

	sess.Cmd = cmd
	sess.Stdin = stdin
	sess.Stdout = bufStdout
	sess.Stderr = bufStderr
	sess.Queue = make(chan models.ExecRequest, 1)
	sess.Closing = make(chan struct{})
	sess.ProjectPath = projectPath

	logger.Info("创建项目会话成功", logger.String("projectPath", projectPath))

	go sess.Loop()

	return sess, nil
}

// createShell 根据系统创建适当的 shell
func (m *SessionManager) createShell(projectPath string) (*exec.Cmd, error) {
	shell := ""
	var args []string

	switch runtime.GOOS {
	case "windows":
		shell = "cmd"
		args = []string{"/C", "cd", "/d", filepath.Join(m.workspacePath, projectPath), "&&", "claude", "--dangerously-skip-permissions"}
	default:
		shell = "bash"
		args = []string{"-i", "cd", filepath.Join(m.workspacePath, projectPath), "&&", "claude", "--dangerously-skip-permissions"}
	}

	cmd := exec.Command(shell, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: false}
	cmd.Dir = filepath.Join(m.workspacePath, projectPath)
	cmd.Env = os.Environ()

	return cmd, nil
}
