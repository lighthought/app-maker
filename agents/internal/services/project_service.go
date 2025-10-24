package services

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/lighthought/app-maker/shared-models/agent"
	"github.com/lighthought/app-maker/shared-models/common"
	"github.com/lighthought/app-maker/shared-models/logger"
	"github.com/lighthought/app-maker/shared-models/tasks"
	"github.com/lighthought/app-maker/shared-models/utils"

	"github.com/hibiken/asynq"
)

// ProjectService 项目服务
type ProjectService interface {
	ProcessTask(ctx context.Context, task *asynq.Task) error
}

// projectService 项目服务实现
type projectService struct {
	commandService   CommandService
	agentTaskService AgentTaskService
	fileService      FileService
	redisService     RedisService
}

// NewProjectService 创建项目服务
func NewProjectService(commandService CommandService,
	agentTaskService AgentTaskService,
	redisService RedisService,
	fileService FileService) ProjectService {
	return &projectService{
		commandService:   commandService,
		agentTaskService: agentTaskService,
		redisService:     redisService,
		fileService:      fileService,
	}
}

// ProcessTask 处理任务
func (s *projectService) ProcessTask(ctx context.Context, task *asynq.Task) error {
	switch task.Type() {
	// 项目环境准备
	case common.TaskTypeAgentSetup:
		return s.agentSetupProject(ctx, task)
	// 部署项目
	case common.TaskTypeProjectDeploy:
		return s.projectDeploy(ctx, task)
	default:
		return fmt.Errorf("unexpected task type %s", task.Type())
	}
}

// checkGitRepository 检查项目的 gitlab 环境
func (s *projectService) checkGitRepository(ctx context.Context, req agent.SetupProjEnvReq, projectPath string) (string, error) {
	var markdownResult string = "项目开发环境初始化：\n"
	if !utils.IsDirectoryExists(projectPath) {
		// git clone 项目
		gitUrl := strings.Replace(req.GitlabRepoUrl, "git@gitlab:app-maker", "http://gitlab.app-maker.localhost/app-maker", 1)
		res := s.commandService.SimpleExecute(ctx, "", "git", "clone", gitUrl, req.ProjectGuid)
		if !res.Success {
			logger.Error("git clone 项目失败", logger.String("error", res.Error))
			return "", fmt.Errorf("git clone 项目失败: %s", res.Error)
		}

		markdownResult += "* git clone 成功：\n"
		logger.Info("git clone 项目成功", logger.String("ProjectGuid", req.ProjectGuid))
	} else {
		logger.Info("project 目录已存在", logger.String("projectPath", projectPath))
		// git.exe pull --progress -v --no-rebase -- "origin"
		res := s.commandService.SimpleExecute(ctx, req.ProjectGuid, "git", "pull", "--progress", "-v", "--no-rebase", "--", "origin")
		if !res.Success {
			logger.Error("git pull 项目失败", logger.String("error", res.Error))
			return "", fmt.Errorf("git pull 项目失败: %s", res.Error)
		}

		markdownResult += "* git pull 成功：\n"
	}

	// 配置不用转换 LF 为 CRLF，避免提交一堆实际没有修改的代码和文档
	s.commandService.SimpleExecute(ctx, req.ProjectGuid, "git", "config", "core.autocrlf", "false")
	return markdownResult, nil
}

// 检查、安装 bmad-method
func (s *projectService) installBmad(ctx context.Context, req agent.SetupProjEnvReq,
	projectPath, markdownResult string) (string, error) {
	// 优先使用请求参数
	installBmad := req.SetupBmadMethod
	bmadCliType := req.BmadCliType
	// 如果请求参数为空，检测本地目录
	if bmadCliType == "" {
		bmadCliType = s.fileService.DetectCliTool(req.ProjectGuid)
	}

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
		} else {
			// 安装 bmad-method 使用指定的 CLI 工具
			res := s.commandService.SimpleExecute(ctx, req.ProjectGuid, "npx", "bmad-method", "install", "-f", "-i", bmadCliType, "-d", ".")
			if !res.Success {
				logger.Error("agent 安装失败", logger.String("error", res.Error))
				return "", fmt.Errorf("bmad-method 安装失败: %s", res.Error)
			}

			markdownResult += fmt.Sprintf("* agent (%s) 安装成功\n", bmadCliType)
			logger.Info("agent 安装成功", logger.String("projectPath", projectPath), logger.String("cliTool", bmadCliType))
		}
	}
	return markdownResult, nil
}

// 安装代码依赖
func (s *projectService) installCodeDependencies(ctx context.Context, req agent.SetupProjEnvReq,
	projectPath, markdownResult string) (string, error) {
	// 安装 frontend 代码依赖
	var frontendModulePath = filepath.Join(projectPath, "frontend", "node_modules")
	if !utils.IsDirectoryExists(frontendModulePath) {
		subPath := req.ProjectGuid + "/frontend"
		res := s.commandService.SimpleExecute(ctx, subPath, "npm", "install")
		if !res.Success {
			logger.Error("frontend 安装失败", logger.String("error", res.Error))
			return "", fmt.Errorf("frontend 安装失败: %s", res.Error)
		}

		logger.Info("frontend 安装成功", logger.String("projectPath", projectPath))
		markdownResult += "* frontend 安装成功\n"
	} else {
		logger.Info("frontend node_modules 已存在", logger.String("projectPath", projectPath))
		markdownResult += "* frontend 已安装过\n"
	}

	// 安装 backend 代码依赖
	if !utils.IsFileExists(filepath.Join(projectPath, "backend", "server")) {
		subPath := req.ProjectGuid + "/backend"
		goMod := s.commandService.SimpleExecute(ctx, subPath, "go", "mod", "download")
		build := s.commandService.SimpleExecute(ctx, subPath, "go", "build", "-o", "server", "./cmd/server")
		if !goMod.Success || !build.Success {
			logger.Error("backend 安装失败", logger.String("error", goMod.Error+build.Error))
			return "", fmt.Errorf("backend 安装失败: %s", goMod.Error+build.Error)
		}

		logger.Info("backend 安装成功", logger.String("projectPath", projectPath))
		markdownResult += "* backend 安装成功\n"
	} else {
		logger.Info("backend 已安装过", logger.String("projectPath", projectPath))
		markdownResult += "* backend 已安装过\n"
	}
	return markdownResult, nil
}

// 初始化项目环境
func (s *projectService) agentSetupProject(ctx context.Context, task *asynq.Task) error {
	var req agent.SetupProjEnvReq
	if err := json.Unmarshal(task.Payload(), &req); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	// 1.检查 workspace 目录下是否有 project 目录，如果没有，则 git clone 项目
	payload := tasks.AgentExecuteTaskPayload{
		ProjectGUID: req.ProjectGuid,
		AgentType:   common.AgentTypePM,
		DevStage:    common.DevStatusSetupAgents,
	}

	var projectPath = s.fileService.GetProjectPath(req.ProjectGuid)
	markdownResult, err := s.checkGitRepository(ctx, req, projectPath)
	if err != nil {
		logger.Error("检查 git 仓库失败", logger.String("error", err.Error()))
		tasks.UpdateResult(task.ResultWriter(), common.CommonStatusFailed, 0, "检查 git 仓库失败: "+err.Error())
		s.redisService.PublishTaskStatus(&payload, task.ResultWriter().TaskID(), common.CommonStatusFailed, "检查 git 仓库失败: "+err.Error())
		return err
	}

	// 2.检查、安装 bmad-method
	markdownResult, err = s.installBmad(ctx, req, projectPath, markdownResult)
	if err != nil {
		logger.Error("安装 bmad-method 失败", logger.String("error", err.Error()))
		tasks.UpdateResult(task.ResultWriter(), common.CommonStatusFailed, 0, "安装 bmad-method 失败: "+err.Error())
		s.redisService.PublishTaskStatus(&payload, task.ResultWriter().TaskID(), common.CommonStatusFailed, "安装 bmad-method 失败: "+err.Error())
		return err
	}

	// 3. 安装代码依赖
	markdownResult, err = s.installCodeDependencies(ctx, req, projectPath, markdownResult)
	if err != nil {
		logger.Error("安装代码依赖失败", logger.String("error", err.Error()))
		tasks.UpdateResult(task.ResultWriter(), common.CommonStatusFailed, 0, "安装代码依赖失败: "+err.Error())
		s.redisService.PublishTaskStatus(&payload, task.ResultWriter().TaskID(), common.CommonStatusFailed, "安装代码依赖失败: "+err.Error())
		return fmt.Errorf("安装代码依赖失败: %s", err.Error())
	}

	tasks.UpdateResult(task.ResultWriter(), common.CommonStatusDone, 100, markdownResult)
	s.redisService.PublishTaskStatus(&payload, task.ResultWriter().TaskID(), common.CommonStatusDone, "项目环境准备完成")
	logger.Info("项目环境准备完成", logger.String("projectGuid", req.ProjectGuid))
	return nil
}

// chatAfterExecuteFailed 聊天后执行失败
func (s *projectService) chatAfterExecuteFailed(ctx context.Context, projectGuid, cmdDesc, process string, cmd ...string) (string, error) {
	logger.Info("执行命令",
		logger.String("projectGuid", projectGuid),
		logger.String("process", process),
		logger.String("cmd", strings.Join(cmd, " ")))

	buildResult := s.commandService.SimpleExecute(ctx, projectGuid, process, cmd...)
	if !buildResult.Success {
		logger.Error(cmdDesc+"失败",
			logger.String("projectGuid", projectGuid),
			logger.String("error", buildResult.Error),
			logger.String("output", buildResult.Output),
		)
		prompt := cmdDesc + "失败了，帮我修复下，最后执行 '" + process + " " + strings.Join(cmd, " ") + "' 命令" + buildResult.Error
		result, err := s.agentTaskService.ChatWithAgent(ctx, projectGuid, common.AgentTypeDev,
			prompt)
		if err != nil {
			return "", fmt.Errorf("%s失败: %s", cmdDesc, err.Error())
		}
		if !result.Success {

			return "", fmt.Errorf("%s失败: %s", cmdDesc, result.Error)
		}
		buildResult = *result
	}
	return buildResult.Output, nil
}

// 部署项目
func (s *projectService) projectDeploy(ctx context.Context, task *asynq.Task) error {
	var req agent.DeployReq
	if err := json.Unmarshal(task.Payload(), &req); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}
	logger.Info("开始执行项目部署", logger.String("projectGuid", req.ProjectGuid))

	payload := tasks.AgentExecuteTaskPayload{
		ProjectGUID: req.ProjectGuid,
		AgentType:   common.AgentTypeDev,
		DevStage:    common.DevStatusDeploy,
	}
	// 1. 执行 make build-dev 构建项目
	buildResult, err2 := s.chatAfterExecuteFailed(ctx, req.ProjectGuid, "构建项目", "make", "build-dev")
	if err2 != nil {
		tasks.UpdateResult(task.ResultWriter(), common.CommonStatusFailed, 0, "构建项目失败: "+err2.Error())
		s.redisService.PublishTaskStatus(&payload, task.ResultWriter().TaskID(), common.CommonStatusFailed, "构建项目失败: "+err2.Error())
		return err2
	}
	tasks.UpdateResult(task.ResultWriter(), common.CommonStatusInProgress, 50, buildResult)
	s.redisService.PublishTaskStatus(&payload, task.ResultWriter().TaskID(), common.CommonStatusInProgress, "构建项目成功")

	// 2. 执行 make run-dev 启动项目
	buildResult, err3 := s.chatAfterExecuteFailed(ctx, req.ProjectGuid, "启动项目", "make", "run-dev")
	if err3 != nil {
		tasks.UpdateResult(task.ResultWriter(), common.CommonStatusFailed, 0, "启动项目失败: "+err3.Error())
		s.redisService.PublishTaskStatus(&payload, task.ResultWriter().TaskID(), common.CommonStatusFailed, "启动项目失败: "+err3.Error())
		return err3
	}
	tasks.UpdateResult(task.ResultWriter(), common.CommonStatusDone, 100, buildResult)
	s.redisService.PublishTaskStatus(&payload, task.ResultWriter().TaskID(), common.CommonStatusDone, "启动项目成功")
	logger.Info("项目部署完成", logger.String("projectGuid", req.ProjectGuid))
	return nil
}
