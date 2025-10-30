package services

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
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

// configurePOAgent 配置 PO Agent 的 customization 字段
func (s *projectService) configurePOAgent(ctx context.Context, req agent.SetupProjEnvReq, projectPath, markdownResult string) (string, error) {
	// 检测 CLI 工具类型
	bmadCliType := req.BmadCliType
	if bmadCliType == "" {
		bmadCliType = s.fileService.DetectCliTool(req.ProjectGuid)
	}

	// 确定配置文件路径
	var poConfigPath string
	switch bmadCliType {
	case common.CliToolGemini:
		poConfigPath = filepath.Join(projectPath, ".bmad-core", "agents", "po.md")
	case common.CliToolQwenCode:
		poConfigPath = filepath.Join(projectPath, ".qwen", "bmad-method", "QWEN.md")
	case common.CliToolClaudeCode:
		poConfigPath = filepath.Join(projectPath, ".claude", "commands", "BMad", "agents", "po.md")
	default:
		// 默认使用 .bmad-core 路径
		poConfigPath = filepath.Join(projectPath, ".bmad-core", "agents", "po.md")
	}

	// 检查文件是否存在
	if !utils.IsFileExists(poConfigPath) {
		logger.Info("PO Agent 配置文件不存在，跳过配置", logger.String("path", poConfigPath))
		markdownResult += "* PO Agent 配置文件不存在，跳过配置\n"
		return markdownResult, nil
	}

	// 读取原文件内容
	file, err := os.Open(poConfigPath)
	if err != nil {
		logger.Error("打开 PO Agent 配置文件失败", logger.String("error", err.Error()))
		return markdownResult, fmt.Errorf("打开 PO Agent 配置文件失败: %s", err.Error())
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	foundCustomization := false
	inCustomization := false
	customizationStarted := false

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		// 查找 customization 字段
		if strings.HasPrefix(line, "customization:") {
			foundCustomization = true
			// 如果是 null，替换为新的配置
			if strings.Contains(line, "null") {
				// 开始多行配置
				lines = append(lines, "customization: |")
				lines = append(lines, "    # 强制输出 MVP JSON 格式的 Epics 和 Stories")
				lines = append(lines, "    CRITICAL MVP JSON OUTPUT REQUIREMENT:")
				lines = append(lines, "    When creating Epics and Stories, you MUST ALWAYS output a JSON block at the end of your response containing MVP phase information.")
				lines = append(lines, "    ")
				lines = append(lines, "    The JSON format must be exactly:")
				lines = append(lines, "    ```json")
				lines = append(lines, "    {")
				lines = append(lines, "      \"mvp_epics\": [")
				lines = append(lines, "        {")
				lines = append(lines, "          \"epic_number\": 1,")
				lines = append(lines, "          \"name\": \"Epic名称\",")
				lines = append(lines, "          \"description\": \"Epic描述\",")
				lines = append(lines, "          \"priority\": \"P0\",")
				lines = append(lines, "          \"estimated_days\": 20,")
				lines = append(lines, "          \"file_path\": \"docs/stories/epic1-xxx-stories.md\",")
				lines = append(lines, "          \"stories\": [")
				lines = append(lines, "            {")
				lines = append(lines, "              \"story_number\": \"US-001\",")
				lines = append(lines, "              \"title\": \"Story标题\",")
				lines = append(lines, "              \"description\": \"Story描述\",")
				lines = append(lines, "              \"priority\": \"P0\",")
				lines = append(lines, "              \"estimated_days\": 3,")
				lines = append(lines, "              \"depends\": \"依赖的其他Story\",")
				lines = append(lines, "              \"techs\": \"技术要点\"")
				lines = append(lines, "            }")
				lines = append(lines, "          ]")
				lines = append(lines, "        }")
				lines = append(lines, "      ]")
				lines = append(lines, "    }")
				lines = append(lines, "    ```")
				lines = append(lines, "    ")
				lines = append(lines, "    MVP EPICS DEFINITION:")
				lines = append(lines, "    - Only include P0 priority epics (core functionality)")
				lines = append(lines, "    - Include all stories within those P0 epics")
				lines = append(lines, "    - Ensure accurate work estimates")
				lines = append(lines, "    - Match the actual file paths created")
				lines = append(lines, "    ")
				lines = append(lines, "    This JSON output is MANDATORY and non-negotiable. Do not skip or modify this requirement.")
				customizationStarted = true
				inCustomization = true
				continue
			} else if strings.Contains(line, "|") {
				// 已经有配置，保持原样
				inCustomization = true
				customizationStarted = true
			}
		} else if inCustomization && customizationStarted {
			// 检查是否还在多行配置中
			if strings.HasPrefix(line, "persona:") || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "commands:") {
				inCustomization = false
			}
		}

		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		logger.Error("读取 PO Agent 配置文件失败", logger.String("error", err.Error()))
		return markdownResult, fmt.Errorf("读取 PO Agent 配置文件失败: %s", err.Error())
	}

	// 如果没有找到 customization 字段，添加它
	if !foundCustomization {
		// 重新读取文件内容并插入 customization
		file, err := os.Open(poConfigPath)
		if err != nil {
			return markdownResult, fmt.Errorf("重新打开 PO Agent 配置文件失败: %s", err.Error())
		}
		defer file.Close()

		lines = nil
		scanner = bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			lines = append(lines, line)

			// 在 persona: 字段之前插入 customization
			if strings.TrimSpace(line) == "persona:" {
				// 插入 customization 配置
				lines = append(lines, "  customization: |")
				lines = append(lines, "    # 强制输出 MVP JSON 格式的 Epics 和 Stories")
				lines = append(lines, "    CRITICAL MVP JSON OUTPUT REQUIREMENT:")
				lines = append(lines, "    When creating Epics and Stories, you MUST ALWAYS output a JSON block at the end of your response containing MVP phase information.")
				lines = append(lines, "    ")
				lines = append(lines, "    The JSON format must be exactly:")
				lines = append(lines, "    ```json")
				lines = append(lines, "    {")
				lines = append(lines, "      \"mvp_epics\": [")
				lines = append(lines, "        {")
				lines = append(lines, "          \"epic_number\": 1,")
				lines = append(lines, "          \"name\": \"Epic名称\",")
				lines = append(lines, "          \"description\": \"Epic描述\",")
				lines = append(lines, "          \"priority\": \"P0\",")
				lines = append(lines, "          \"estimated_days\": 20,")
				lines = append(lines, "          \"file_path\": \"docs/stories/epic1-xxx-stories.md\",")
				lines = append(lines, "          \"stories\": [")
				lines = append(lines, "            {")
				lines = append(lines, "              \"story_number\": \"US-001\",")
				lines = append(lines, "              \"title\": \"Story标题\",")
				lines = append(lines, "              \"description\": \"Story描述\",")
				lines = append(lines, "              \"priority\": \"P0\",")
				lines = append(lines, "              \"estimated_days\": 3,")
				lines = append(lines, "              \"depends\": \"依赖的其他Story\",")
				lines = append(lines, "              \"techs\": \"技术要点\"")
				lines = append(lines, "            }")
				lines = append(lines, "          ]")
				lines = append(lines, "        }")
				lines = append(lines, "      ]")
				lines = append(lines, "    }")
				lines = append(lines, "    ```")
				lines = append(lines, "    ")
				lines = append(lines, "    MVP EPICS DEFINITION:")
				lines = append(lines, "    - Only include P0 priority epics (core functionality)")
				lines = append(lines, "    - Include all stories within those P0 epics")
				lines = append(lines, "    - Ensure accurate work estimates")
				lines = append(lines, "    - Match the actual file paths created")
				lines = append(lines, "    ")
				lines = append(lines, "    This JSON output is MANDATORY and non-negotiable. Do not skip or modify this requirement.")
				break
			}
		}

		if err := scanner.Err(); err != nil {
			return markdownResult, fmt.Errorf("重新读取 PO Agent 配置文件失败: %s", err.Error())
		}
	}

	// 写入更新后的内容
	outputFile, err := os.Create(poConfigPath)
	if err != nil {
		logger.Error("创建 PO Agent 配置文件失败", logger.String("error", err.Error()))
		return markdownResult, fmt.Errorf("创建 PO Agent 配置文件失败: %s", err.Error())
	}
	defer outputFile.Close()

	writer := bufio.NewWriter(outputFile)
	for _, line := range lines {
		if _, err := writer.WriteString(line + "\n"); err != nil {
			logger.Error("写入 PO Agent 配置文件失败", logger.String("error", err.Error()))
			return markdownResult, fmt.Errorf("写入 PO Agent 配置文件失败: %s", err.Error())
		}
	}
	if err := writer.Flush(); err != nil {
		logger.Error("刷新 PO Agent 配置文件失败", logger.String("error", err.Error()))
		return markdownResult, fmt.Errorf("刷新 PO Agent 配置文件失败: %s", err.Error())
	}

	logger.Info("PO Agent 配置更新成功", logger.String("path", poConfigPath))
	markdownResult += "* PO Agent 配置更新成功，添加了 MVP JSON 输出要求\n"
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

	// 3. 配置 PO Agent 的 customization 字段
	markdownResult, err = s.configurePOAgent(ctx, req, projectPath, markdownResult)
	if err != nil {
		logger.Error("配置 PO Agent 失败", logger.String("error", err.Error()))
		// PO Agent 配置失败不是致命错误，继续执行
		logger.Warn("PO Agent 配置失败，但继续执行后续步骤", logger.String("error", err.Error()))
		markdownResult += "* PO Agent 配置失败，但不影响后续步骤\n"
	}

	// 4. 安装代码依赖
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
