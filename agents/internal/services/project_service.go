package services

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"shared-models/agent"
	"shared-models/common"
	"shared-models/logger"
	"shared-models/tasks"
	"shared-models/utils"
	"strings"

	"github.com/hibiken/asynq"
)

type ProjectService interface {
	ProcessTask(ctx context.Context, task *asynq.Task) error
	GetWorkspacePath() string
}

type projectService struct {
	commandService CommandService
	workspacePath  string
}

func NewProjectService(commandService CommandService, workspacePath string) ProjectService {
	return &projectService{commandService: commandService, workspacePath: workspacePath}
}

// ProcessTask 处理任务
func (s *projectService) ProcessTask(ctx context.Context, task *asynq.Task) error {
	switch task.Type() {
	case common.TaskTypeAgentSetup:
		return s.agentSetupProject(ctx, task)
	default:
		return fmt.Errorf("unexpected task type %s", task.Type())
	}
}

// 获取工作空间路径
func (s *projectService) GetWorkspacePath() string {
	return s.workspacePath
}

// 初始化项目环境
func (s *projectService) agentSetupProject(ctx context.Context, task *asynq.Task) error {
	var req agent.SetupProjEnvReq
	if err := json.Unmarshal(task.Payload(), &req); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	installBmad := req.SetupBmadMethod
	bmadCliType := req.BmadCliType

	// 检查 workspace 目录下是否有 project 目录，如果有，则删除
	if s.workspacePath == "" {
		s.workspacePath = utils.GetEnvOrDefault("WORKSPACE_PATH", "F:/app-maker/app_data")
	}

	var markdownResult string = "项目开发环境初始化：\n"
	// 检查 workspace 目录下是否有 project 目录，如果没有，则 git clone 项目
	var projectPath = filepath.Join(s.workspacePath, req.ProjectGuid)
	if !utils.IsDirectoryExists(projectPath) {
		// git clone 项目
		gitUrl := strings.Replace(req.GitlabRepoUrl, "git@gitlab:app-maker", "http://gitlab.app-maker.localhost/app-maker", 1)
		res := s.commandService.SimpleExecute(ctx, "", "git", "clone", gitUrl, req.ProjectGuid)
		if !res.Success {
			tasks.UpdateResult(task.ResultWriter(), common.CommonStatusFailed, 0, "git clone 项目失败: "+res.Error)
			return fmt.Errorf("git clone 项目失败: %s", res.Error)
		}

		markdownResult += "* git clone 成功：\n"
		logger.Info("git clone 项目成功", logger.String("ProjectGuid", req.ProjectGuid))
		tasks.UpdateResult(task.ResultWriter(), common.CommonStatusInProgress, 30, "git clone 项目成功")
	} else {
		logger.Info("project 目录已存在", logger.String("projectPath", projectPath))
		// git.exe pull --progress -v --no-rebase -- "origin"
		res := s.commandService.SimpleExecute(ctx, req.ProjectGuid, "git", "pull", "--progress", "-v", "--no-rebase", "--", "origin")
		if !res.Success {
			tasks.UpdateResult(task.ResultWriter(), common.CommonStatusFailed, 0, "git pull 项目失败: "+res.Error)
			return fmt.Errorf("git pull 项目失败: %s", res.Error)
		}

		markdownResult += "* git pull 成功：\n"
		tasks.UpdateResult(task.ResultWriter(), common.CommonStatusInProgress, 30, "project 目录已存在, git pull 更新代码成功")
	}

	if installBmad {
		var bmadCorePath = filepath.Join(s.GetWorkspacePath(), req.ProjectGuid, ".bmad-core")
		if !utils.IsDirectoryExists(bmadCorePath) {
			res := s.commandService.SimpleExecute(ctx, req.ProjectGuid, "npx", "bmad-method", "install", "-f", "-i", bmadCliType, "-d", ".")
			if !res.Success {
				logger.Error("bmad-method 安装失败", logger.String("projectPath", projectPath), logger.String("error", res.Error))
				tasks.UpdateResult(task.ResultWriter(), common.CommonStatusFailed, 0, "bmad-method 安装失败: "+res.Error)
				return fmt.Errorf("bmad-method 安装失败: %s", res.Error)
			}

			logger.Info("bmad-method 安装成功", logger.String("projectPath", projectPath))
			tasks.UpdateResult(task.ResultWriter(), common.CommonStatusInProgress, 60, "bmad-method 安装成功, "+res.Output)
			markdownResult += "* agents 安装成功：\n"
		} else {
			logger.Info("bmad-method 已安装过", logger.String("projectPath", projectPath))
			tasks.UpdateResult(task.ResultWriter(), common.CommonStatusInProgress, 60, "bmad-method 已安装过")
			markdownResult += "* agents 已安装过：\n"
		}
	}

	var frontendModulePath = filepath.Join(s.GetWorkspacePath(), req.ProjectGuid, "frontend", "node_modules")
	if !utils.IsDirectoryExists(frontendModulePath) {
		subPath := req.ProjectGuid + "/frontend"
		res := s.commandService.SimpleExecute(ctx, subPath, "npm", "install")
		if !res.Success {
			tasks.UpdateResult(task.ResultWriter(), common.CommonStatusFailed, 0, "frontend 安装失败: "+res.Error)
			return fmt.Errorf("frontend 安装失败: %s", res.Error)
		}
		logger.Info("frontend 安装成功", logger.String("projectPath", projectPath))
		tasks.UpdateResult(task.ResultWriter(), common.CommonStatusInProgress, 80, "frontend 安装成功")
		markdownResult += "* frontend 安装成功\n"
	} else {
		logger.Info("frontend node_modules 已存在", logger.String("projectPath", projectPath))
		tasks.UpdateResult(task.ResultWriter(), common.CommonStatusInProgress, 80, "frontend node_modules 已存在")
		markdownResult += "* frontend 已安装过\n"
	}

	if !utils.IsFileExists(filepath.Join(s.GetWorkspacePath(), req.ProjectGuid, "backend", "server")) {
		subPath := req.ProjectGuid + "/backend"
		goMod := s.commandService.SimpleExecute(ctx, subPath, "go", "mod", "download")
		build := s.commandService.SimpleExecute(ctx, subPath, "go", "build", "-o", "server", "./cmd/server")
		if !goMod.Success || !build.Success {
			tasks.UpdateResult(task.ResultWriter(), common.CommonStatusFailed, 0, "backend 安装失败: "+goMod.Error+build.Error)
			return fmt.Errorf("backend 安装失败: %s", goMod.Error+build.Error)
		}

		logger.Info("backend 安装成功", logger.String("projectPath", projectPath))
		tasks.UpdateResult(task.ResultWriter(), common.CommonStatusDone, 95, "backend 安装成功")
		markdownResult += "* backend 安装成功\n"
	} else {
		logger.Info("backend server 已存在", logger.String("projectPath", projectPath))
		tasks.UpdateResult(task.ResultWriter(), common.CommonStatusDone, 95, "backend 安装成功")
		markdownResult += "* backend 已安装过\n"
	}

	tasks.UpdateResult(task.ResultWriter(), common.CommonStatusDone, 100, markdownResult)
	return nil
}
