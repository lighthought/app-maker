package services

import (
	"app-maker-agents/internal/utils"
	"context"
	"fmt"
	"path/filepath"
	"shared-models/agent"
	"time"
)

type ProjectService interface {
	SetupProjectEnvironment(ctx context.Context, req *agent.SetupProjEnvReq) (*agent.SetupProjEnvRes, error)
}

type projectService struct {
	commandService *CommandService
	workspacePath  string
}

func NewProjectService(commandService *CommandService, workspacePath string) ProjectService {
	return &projectService{commandService: commandService, workspacePath: workspacePath}
}

func (s *projectService) SetupProjectEnvironment(ctx context.Context, req *agent.SetupProjEnvReq) (*agent.SetupProjEnvRes, error) {
	installBmad := req.SetupBmadMethod
	bmadCliType := req.BmadCliType

	var resp = agent.SetupProjEnvRes{
		BmadMethodStatus: "success",
		FrontendStatus:   "success",
		BackendStatus:    "success",
	}
	// 检查 workspace 目录下是否有 project 目录，如果有，则删除
	if s.workspacePath == "" {
		s.workspacePath = utils.GetEnvOrDefault("WORKSPACE_PATH", "F:/app-maker/app_data")
	}

	var projectPath = filepath.Join(s.workspacePath, req.ProjectGuid)
	if !utils.IsDirectoryExists(projectPath) {
		// git clone 项目
		cmd := `git clone ` + req.GitlabRepoUrl
		res := s.commandService.Execute(ctx, req.ProjectGuid, cmd, 5*time.Minute)
		if !res.Success {
			resp.BackendStatus = "failed"
			return &resp, fmt.Errorf("git clone 项目失败")
		}
	}

	if installBmad {
		cmd := `npx bmad-method install -f -i ` + bmadCliType + ` -d .`
		res := s.commandService.Execute(ctx, req.ProjectGuid, cmd, 5*time.Minute)
		resp.BmadMethodStatus = "done"
		if !res.Success {
			resp.BmadMethodStatus = "failed"
		}
	}

	cmd := `cd frontend && npm install`
	res := s.commandService.Execute(ctx, req.ProjectGuid, cmd, 5*time.Minute)
	resp.FrontendStatus = "done"
	if !res.Success {
		resp.FrontendStatus = "failed"
	}

	goMod := s.commandService.Execute(ctx, req.ProjectGuid, `cd backend && go mod download`, 5*time.Minute)
	build := s.commandService.Execute(ctx, req.ProjectGuid, `cd backend && go build -o server ./cmd/server`, 5*time.Minute)
	resp.BackendStatus = "done"
	if !goMod.Success || !build.Success {
		resp.BackendStatus = "failed"
	}

	return &resp, nil
}
