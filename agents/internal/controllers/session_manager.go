package controllers

import (
	"bufio"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"app-maker-agents/internal/models"
	"app-maker-agents/pkg/logger"
)

// SessionManager 维护项目级命令执行会话
type SessionManager struct {
	sessions sync.Map // map[string]*session
}

// NewSessionManager 创建 SessionManager
func NewSessionManager() *SessionManager {
	return &SessionManager{}
}

// Execute 在指定项目会话中执行命令
func (m *SessionManager) Execute(projectPath, command string, timeout time.Duration) models.ExecResult {
	sess, err := m.getOrCreateSession(projectPath)
	if err != nil {
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

	cmd, err := createShell(projectPath)
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

	go sess.Loop()

	return sess, nil
}

// createShell 根据系统创建适当的 shell
func createShell(projectPath string) (*exec.Cmd, error) {
	shell := ""
	args := []string{}

	switch runtime.GOOS {
	case "windows":
		shell = "cmd"
		args = []string{"/C", "cd /d " + filepath.Clean(projectPath) + " && cmd"}
	default:
		shell = "bash"
		args = []string{"-lc", "cd " + filepath.Clean(projectPath) + " && bash"}
	}

	cmd := exec.Command(shell, args...)
	cmd.Dir = projectPath
	cmd.Env = append([]string{}, exec.Command(shell).Env...)

	return cmd, nil
}
