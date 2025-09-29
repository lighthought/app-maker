package services

import (
	"app-maker-agents/internal/utils"
	"app-maker-agents/pkg/logger"
	"context"
	"fmt"
	"path/filepath"
	"shared-models/agent"
	"shared-models/common"
)

type ProjectService interface {
	SetupProjectEnvironment(ctx context.Context, req *agent.SetupProjEnvReq) (*agent.SetupProjEnvResp, error)
}

type projectService struct {
	commandService *CommandService
	workspacePath  string
}

func NewProjectService(commandService *CommandService, workspacePath string) ProjectService {
	return &projectService{commandService: commandService, workspacePath: workspacePath}
}

func (s *projectService) SetupProjectEnvironment(ctx context.Context, req *agent.SetupProjEnvReq) (*agent.SetupProjEnvResp, error) {
	installBmad := req.SetupBmadMethod
	bmadCliType := req.BmadCliType

	var resp = agent.SetupProjEnvResp{
		BmadMethodStatus: "success",
		FrontendStatus:   "success",
		BackendStatus:    "success",
	}
	// 检查 workspace 目录下是否有 project 目录，如果有，则删除
	if s.workspacePath == "" {
		s.workspacePath = utils.GetEnvOrDefault("WORKSPACE_PATH", "F:/app-maker/app_data")
	}

	// 检查 workspace 目录下是否有 project 目录，如果没有，则 git clone 项目
	var projectPath = filepath.Join(s.workspacePath, req.ProjectGuid)
	if !utils.IsDirectoryExists(projectPath) {
		// git clone 项目
		res := s.commandService.SimpleExecute(ctx, "", "git", "clone", req.GitlabRepoUrl, req.ProjectGuid)
		if !res.Success {
			resp.BackendStatus = common.CommonStatusFailed
			return &resp, fmt.Errorf("git clone 项目失败: %s", res.Error)
		}

		logger.Info("git clone 项目成功", logger.String("ProjectGuid", req.ProjectGuid))
	} else {
		logger.Info("project 目录已存在", logger.String("projectPath", projectPath))
	}

	if installBmad {
		var bmadCorePath = filepath.Join(s.commandService.WorkspacePath, req.ProjectGuid, ".bmad-core")
		if !utils.IsDirectoryExists(bmadCorePath) {
			res := s.commandService.SimpleExecute(ctx, req.ProjectGuid, "npx", "bmad-method", "install", "-f", "-i", bmadCliType, "-d", ".")
			resp.BmadMethodStatus = common.CommonStatusDone
			if !res.Success {
				resp.BmadMethodStatus = common.CommonStatusFailed
				logger.Error("bmad-method 安装失败", logger.String("projectPath", projectPath), logger.String("error", res.Error))
			} else {
				logger.Info("bmad-method 安装成功", logger.String("projectPath", projectPath))
			}
		} else {
			logger.Info("bmad-method 已安装过", logger.String("projectPath", projectPath))
			resp.BmadMethodStatus = common.CommonStatusDone
		}
	}

	var frontendModulePath = filepath.Join(s.commandService.WorkspacePath, req.ProjectGuid, "frontend", "node_modules")
	if !utils.IsDirectoryExists(frontendModulePath) {
		res := s.commandService.SimpleExecute(ctx, req.ProjectGuid+"/frontend", "npm", "install")
		resp.FrontendStatus = common.CommonStatusDone
		if !res.Success {
			resp.FrontendStatus = common.CommonStatusFailed
			logger.Error("frontend 安装失败", logger.String("projectPath", projectPath), logger.String("error", res.Error))
		} else {
			logger.Info("frontend 安装成功", logger.String("projectPath", projectPath))
		}
	} else {
		logger.Info("frontend node_modules 已存在", logger.String("projectPath", projectPath))
		resp.FrontendStatus = common.CommonStatusDone
	}

	if !utils.IsFileExists(filepath.Join(s.commandService.WorkspacePath, req.ProjectGuid, "backend", "server")) {
		goMod := s.commandService.SimpleExecute(ctx, req.ProjectGuid+"/backend", "go", "mod", "download")
		build := s.commandService.SimpleExecute(ctx, req.ProjectGuid+"/backend", "go", "build", "-o", "server", "./cmd/server")
		resp.BackendStatus = common.CommonStatusDone
		if !goMod.Success || !build.Success {
			resp.BackendStatus = common.CommonStatusFailed
			logger.Error("backend 安装失败", logger.String("projectPath", projectPath), logger.String("error", goMod.Error+build.Error))
		} else {
			logger.Info("backend 安装成功", logger.String("projectPath", projectPath))
		}
	} else {
		logger.Info("backend server 已存在", logger.String("projectPath", projectPath))
		resp.BackendStatus = common.CommonStatusDone
	}

	return &resp, nil
}
