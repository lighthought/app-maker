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
}

type projectService struct {
	commandService   CommandService
	agentTaskService AgentTaskService
	fileService      FileService
}

func NewProjectService(commandService CommandService,
	agentTaskService AgentTaskService,
	fileService FileService) ProjectService {
	return &projectService{
		commandService:   commandService,
		agentTaskService: agentTaskService,
		fileService:      fileService,
	}
}

// ProcessTask 处理任务
func (s *projectService) ProcessTask(ctx context.Context, task *asynq.Task) error {
	switch task.Type() {
	case common.TaskTypeAgentSetup:
		return s.agentSetupProject(ctx, task)
	case common.TaskTypeProjectDeploy:
		return s.projectDeploy(ctx, task)
	default:
		return fmt.Errorf("unexpected task type %s", task.Type())
	}
}

// 初始化项目环境
func (s *projectService) agentSetupProject(ctx context.Context, task *asynq.Task) error {
	var req agent.SetupProjEnvReq
	if err := json.Unmarshal(task.Payload(), &req); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	installBmad := req.SetupBmadMethod

	var markdownResult string = "项目开发环境初始化：\n"
	// 检查 workspace 目录下是否有 project 目录，如果没有，则 git clone 项目
	var projectPath = s.fileService.GetProjectPath(req.ProjectGuid)

	// 1. 优先使用请求参数
	bmadCliType := req.BmadCliType
	// 2. 如果请求参数为空，检测本地目录
	if bmadCliType == "" {
		bmadCliType = s.fileService.DetectCliTool(req.ProjectGuid)
	}
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

	// 配置不用转换 LF 为 CRLF，避免提交一堆实际没有修改的代码和文档
	s.commandService.SimpleExecute(ctx, req.ProjectGuid, "git", "config", "core.autocrlf", "false")

	// 3. 验证 CLI 工具对应的目录是否存在
	cliDirMap := map[string]string{
		common.CliToolClaudeCode: ".claude",
		common.CliToolQwenCode:   ".qwen",
		common.CliToolGemini:     ".gemini",
	}

	cliDir := cliDirMap[bmadCliType]
	needInstall := installBmad || !utils.IsDirectoryExists(filepath.Join(projectPath, cliDir))

	if needInstall {
		if utils.IsDirectoryExists(filepath.Join(projectPath, cliDir)) {
			logger.Info("agent 已安装", logger.String("projectPath", projectPath), logger.String("cliTool", bmadCliType))
			markdownResult += fmt.Sprintf("* agent (%s) 已安装\n", bmadCliType)
			tasks.UpdateResult(task.ResultWriter(), common.CommonStatusInProgress, 60, markdownResult)
		} else {
			// 安装 bmad-method 使用指定的 CLI 工具
			res := s.commandService.SimpleExecute(ctx, req.ProjectGuid, "npx", "bmad-method", "install", "-f", "-i", bmadCliType, "-d", ".")
			if !res.Success {
				tasks.UpdateResult(task.ResultWriter(), common.CommonStatusFailed, 0, "agent 安装失败: "+res.Error)
				return fmt.Errorf("bmad-method 安装失败: %s", res.Error)
			}

			markdownResult += fmt.Sprintf("* agent (%s) 安装成功\n", bmadCliType)
			tasks.UpdateResult(task.ResultWriter(), common.CommonStatusInProgress, 60, markdownResult)
		}
	}

	var frontendModulePath = filepath.Join(projectPath, "frontend", "node_modules")
	if !utils.IsDirectoryExists(frontendModulePath) {
		subPath := req.ProjectGuid + "/frontend"
		res := s.commandService.SimpleExecute(ctx, subPath, "npm", "install")
		if !res.Success {
			tasks.UpdateResult(task.ResultWriter(), common.CommonStatusFailed, 0, "frontend 安装失败: "+res.Error)
			return fmt.Errorf("frontend 安装失败: %s", res.Error)
		}

		tasks.UpdateResult(task.ResultWriter(), common.CommonStatusInProgress, 80, "frontend 安装成功")
		markdownResult += "* frontend 安装成功\n"
	} else {
		tasks.UpdateResult(task.ResultWriter(), common.CommonStatusInProgress, 80, "frontend node_modules 已存在")
		markdownResult += "* frontend 已安装过\n"
	}

	if !utils.IsFileExists(filepath.Join(projectPath, "backend", "server")) {
		subPath := req.ProjectGuid + "/backend"
		goMod := s.commandService.SimpleExecute(ctx, subPath, "go", "mod", "download")
		build := s.commandService.SimpleExecute(ctx, subPath, "go", "build", "-o", "server", "./cmd/server")
		if !goMod.Success || !build.Success {
			tasks.UpdateResult(task.ResultWriter(), common.CommonStatusFailed, 0, "backend 安装失败: "+goMod.Error+build.Error)
			return fmt.Errorf("backend 安装失败: %s", goMod.Error+build.Error)
		}

		tasks.UpdateResult(task.ResultWriter(), common.CommonStatusDone, 95, "backend 安装成功")
		markdownResult += "* backend 安装成功\n"
	} else {
		tasks.UpdateResult(task.ResultWriter(), common.CommonStatusDone, 95, "backend 安装成功")
		markdownResult += "* backend 已安装过\n"
	}

	tasks.UpdateResult(task.ResultWriter(), common.CommonStatusDone, 100, markdownResult)
	return nil
}

// 部署项目
func (s *projectService) projectDeploy(ctx context.Context, task *asynq.Task) error {
	var req agent.DeployReq
	if err := json.Unmarshal(task.Payload(), &req); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}
	logger.Info("开始执行项目部署", logger.String("projectGuid", req.ProjectGuid))

	// 1. 执行 make build-dev 构建项目
	logger.Info("执行 make build-dev", logger.String("projectGuid", req.ProjectGuid))
	buildResult := s.commandService.SimpleExecute(ctx, req.ProjectGuid, "make", "build-dev")
	if !buildResult.Success {
		logger.Error("项目构建失败",
			logger.String("projectGuid", req.ProjectGuid),
			logger.String("error", buildResult.Error),
			logger.String("output", buildResult.Output),
		)
		prompt := "项目构建失败了，帮我修复下，最后执行 'make buid-dev' 命令" + buildResult.Error
		result, err := s.agentTaskService.ChatWithAgent(ctx, req.ProjectGuid, common.AgentTypeDev,
			prompt)
		if err != nil {
			tasks.UpdateResult(task.ResultWriter(), common.CommonStatusFailed, 0, "项目构建失败: "+err.Error())
			return fmt.Errorf("项目构建失败: %s", err.Error())
		}
		if !result.Success {
			tasks.UpdateResult(task.ResultWriter(), common.CommonStatusFailed, 0, "项目构建失败: "+result.Error)
			return fmt.Errorf("项目构建失败: %s", result.Error)
		}
		buildResult = *result
	}

	tasks.UpdateResult(task.ResultWriter(), common.CommonStatusInProgress, 50, buildResult.Output)
	logger.Info("项目构建成功", logger.String("projectGuid", req.ProjectGuid))

	// 2. 执行 make run-dev 启动项目
	logger.Info("执行 make run-dev", logger.String("projectGuid", req.ProjectGuid))
	runResult := s.commandService.SimpleExecute(ctx, req.ProjectGuid, "make", "run-dev")
	if !runResult.Success {
		logger.Error("项目启动失败",
			logger.String("projectGuid", req.ProjectGuid),
			logger.String("error", runResult.Error),
			logger.String("output", runResult.Output),
		)
		prompt := "项目启动失败了，帮我修复下，最后执行 'make run-dev' 命令" + runResult.Error
		result, err := s.agentTaskService.ChatWithAgent(ctx, req.ProjectGuid, common.AgentTypeDev,
			prompt)
		buildResult = *result
		if err != nil {
			tasks.UpdateResult(task.ResultWriter(), common.CommonStatusFailed, 0, "项目启动失败: "+err.Error())
			return fmt.Errorf("项目启动失败: %s", err.Error())
		}
		if !result.Success {
			tasks.UpdateResult(task.ResultWriter(), common.CommonStatusFailed, 0, "项目启动失败: "+result.Error)
			return fmt.Errorf("项目启动失败: %s", result.Error)
		}
	}

	logger.Info("项目部署完成", logger.String("projectGuid", req.ProjectGuid))
	tasks.UpdateResult(task.ResultWriter(), common.CommonStatusDone, 100, buildResult.Output)
	return nil
}
