package services

import (
	"app-maker-agents/internal/models"
	"context"
	"time"
)

type ProjectService interface {
	SetupProjectEnvironment(ctx context.Context, req *models.SetupProjEnvReq) (*models.SetupProjEnvRes, error)
}

type projectService struct {
	commandService *CommandService
}

func NewProjectService(commandService *CommandService) ProjectService {
	return &projectService{commandService: commandService}
}

func (s *projectService) SetupProjectEnvironment(ctx context.Context, req *models.SetupProjEnvReq) (*models.SetupProjEnvRes, error) {
	installBmad := req.SetupBmadMethod
	bmadCliType := req.BmadCliType

	var resp = models.SetupProjEnvRes{
		BmadMethodStatus: "success",
		FrontendStatus:   "success",
		BackendStatus:    "success",
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
