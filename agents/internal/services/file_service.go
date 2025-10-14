package services

import (
	"path/filepath"
	"shared-models/common"
	"shared-models/utils"
)

type FileService interface {
	// 检测项目的 CLI 工具
	DetectCliTool(projectGuid string) string

	// 获取项目路径
	GetProjectPath(projectGuid string) string

	// 获取工作空间路径
	GetWorkspacePath() string
}

type fileService struct {
	commandService CommandService
	workspacePath  string
}

func NewFileService(commandService CommandService, workspacePath string) FileService {
	return &fileService{
		commandService: commandService,
		workspacePath:  workspacePath,
	}
}

// getProjectPath 获取项目路径
func (s *fileService) GetProjectPath(projectGuid string) string {
	return filepath.Join(s.workspacePath, projectGuid)
}

// 获取工作空间路径
func (s *fileService) GetWorkspacePath() string {
	if s.workspacePath == "" {
		s.workspacePath = utils.GetEnvOrDefault("WORKSPACE_PATH", "F:/app-maker/app_data")
	}
	return s.workspacePath
}

// DetectCliTool 检测项目使用的 CLI 工具类型
func (s *fileService) DetectCliTool(projectGuid string) string {
	projectPath := s.GetProjectPath(projectGuid)

	if utils.IsDirectoryExists(filepath.Join(projectPath, ".claude")) {
		return common.CliToolClaudeCode
	}
	if utils.IsDirectoryExists(filepath.Join(projectPath, ".qwen")) {
		return common.CliToolQwenCode
	}
	if utils.IsDirectoryExists(filepath.Join(projectPath, ".gemini")) {
		return common.CliToolGemini
	}

	return common.CliToolClaudeCode // 默认
}
