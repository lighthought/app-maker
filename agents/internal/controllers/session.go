package controllers

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
	"time"

	"app-maker-agents/internal/models"
	"app-maker-agents/internal/utils"
	"app-maker-agents/pkg/logger"
)

type Session struct {
	Cmd         *exec.Cmd
	Stdin       io.WriteCloser
	Stdout      *bufio.Reader
	Stderr      *bufio.Reader
	Queue       chan models.ExecRequest
	Closing     chan struct{}
	Closed      bool
	ClosedMutex sync.Mutex
	ProjectPath string
	CliType     string
	Mutex       sync.Mutex
}

func (s *Session) Loop() {
	logger.Info("Loop 开始", logger.String("projectPath", s.ProjectPath), logger.String("cliType", s.CliType))
	for {
		select {
		case req := <-s.Queue:
			s.execute(req)
		case <-s.Closing:
			s.cleanup()
			logger.Info("Loop 会话已退出", logger.String("projectPath", s.ProjectPath), logger.String("cliType", s.CliType))
			return
		default:
			logger.Info("Loop 等待命令", logger.String("projectPath", s.ProjectPath), logger.String("cliType", s.CliType))
			time.Sleep(3 * time.Second)
		}
	}
}

func (s *Session) execute(req models.ExecRequest) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	logger.Info("Loop 执行命令", logger.String("command", req.Command))

	if s.Cmd.ProcessState != nil && s.Cmd.ProcessState.Exited() {
		logger.Info("Loop 会话已退出")
		req.Done <- models.ExecResult{Success: false, Err: errors.New("会话已退出")}
		return
	}

	token := fmt.Sprintf("__CMD_DONE_%d_%s__", time.Now().UnixNano(), utils.RandomString(6))
	command := fmt.Sprintf("%s && echo %s:$?\n", req.Command, token)

	if _, err := s.Stdin.Write([]byte(command)); err != nil {
		logger.Error("Loop 执行命令失败", logger.String("command", req.Command), logger.ErrorField(err))
		req.Done <- models.ExecResult{Success: false, Err: err}
		return
	}

	stdoutBuf := strings.Builder{}
	stderrBuf := strings.Builder{}
	deadline := time.Now().Add(req.Timeout)

	for {
		if req.Timeout > 0 && time.Now().After(deadline) {
			logger.Error("Loop 命令执行超时", logger.String("command", req.Command))
			req.Done <- models.ExecResult{Success: false, Err: errors.New("命令执行超时")}
			return
		}

		stdoutLine, stdoutErr := s.Stdout.ReadString('\n')
		if stdoutErr != nil {
			logger.Error("Loop 读取 stdout 失败", logger.String("command", req.Command), logger.ErrorField(stdoutErr))
			req.Done <- models.ExecResult{Success: false, Err: stdoutErr}
			return
		}

		if strings.Contains(stdoutLine, token) {
			parts := strings.Split(strings.TrimSpace(stdoutLine), ":")
			if len(parts) == 2 {
				exitCode := parts[1]
				success := exitCode == "0"
				logger.Info("Loop 命令执行成功", logger.String("command", req.Command), logger.String("stdout", strings.TrimSpace(stdoutBuf.String())), logger.String("stderr", strings.TrimSpace(stderrBuf.String())))
				req.Done <- models.ExecResult{Success: success, Stdout: strings.TrimSpace(stdoutBuf.String()), Stderr: strings.TrimSpace(stderrBuf.String())}
				return
			}
		}

		stdoutBuf.WriteString(stdoutLine)

		for {
			if s.stderrBuffered() == 0 {
				break
			}
			stderrLine, err := s.Stderr.ReadString('\n')
			if err != nil {
				logger.Error("Loop 读取 stderr 失败", logger.String("command", req.Command), logger.ErrorField(err))
				req.Done <- models.ExecResult{Success: false, Err: err}
				return
			}
			stderrBuf.WriteString(stderrLine)
		}
	}
}

func (s *Session) stderrBuffered() int {
	if s.Stderr == nil {
		return 0
	}
	return s.Stderr.Buffered()
}

func (s *Session) cleanup() {
	s.ClosedMutex.Lock()
	defer s.ClosedMutex.Unlock()
	if s.Closed {
		return
	}
	s.Closed = true
	if s.Stdin != nil {
		s.Stdin.Close()
	}
	if s.Cmd != nil && s.Cmd.Process != nil {
		s.Cmd.Process.Kill()
	}
}
