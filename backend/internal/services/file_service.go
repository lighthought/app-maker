package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/lighthought/app-maker/shared-models/logger"
	"github.com/lighthought/app-maker/shared-models/utils"

	"github.com/lighthought/app-maker/backend/internal/models"
)

// PreviewFilesConfig 预览项目文件配置
type PreviewFilesConfig struct {
	Folders []string `json:"folders"`
	Files   []string `json:"files"`
}

// ProjectFileService 项目文件服务接口
type FileService interface {
	// GetProjectFiles 获取项目文件列表
	GetProjectFiles(ctx context.Context, userID, projectGuid, path string) ([]models.FileItem, error)

	// GetFileContent 获取文件内容
	GetFileContent(ctx context.Context, userID, projectGuid, filePath, encoding string) (*models.FileContent, error)

	// SyncEpicsToFiles 将数据库中的 Epics 和 Stories 同步到项目文件
	SyncEpicsToFiles(ctx context.Context, projectPath string, epics []*models.Epic) error
}

// projectFileService 项目文件服务实现
type fileService struct {
	gitService GitService
}

// NewProjectFileService 创建项目文件服务
func NewFileService(gitService GitService) FileService {
	return &fileService{
		gitService: gitService,
	}
}

// loadPreviewFilesConfig 加载预览文件配置
func (s *fileService) loadPreviewFilesConfig(userID, projectGuid string) (*PreviewFilesConfig, error) {
	projectPath := utils.GetProjectPath(userID, projectGuid)
	if projectPath == "" {
		return nil, fmt.Errorf("project path is empty")
	}

	configPath := filepath.Join(projectPath, "preview_files.json")

	// 检查配置文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// 如果配置文件不存在，返回默认配置
		return &PreviewFilesConfig{
			Folders: []string{"backend", "frontend"},
			Files:   []string{"README.md", "docker-compose.yml"},
		}, nil
	}

	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read preview files config: %s", err.Error())
	}

	var config PreviewFilesConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse preview files config: %s", err.Error())
	}

	return &config, nil
}

// 通过 git 获取最新的代码
func (s *fileService) refreshProjectFiles(ctx context.Context, userID, projectGuid, projectRootPath, path string) error {
	if path == "" {
		gitConfig := &GitConfig{
			UserID:        userID,
			GUID:          projectGuid,
			ProjectPath:   projectRootPath,
			CommitMessage: "Auto commit by App Maker",
		}
		return s.gitService.Pull(ctx, gitConfig)
	}
	return nil
}

// GetProjectFiles 获取项目文件列表
func (s *fileService) GetProjectFiles(ctx context.Context, userID, projectGuid, path string) ([]models.FileItem, error) {
	// 构建项目路径
	projectRootPath := utils.GetProjectPath(userID, projectGuid)
	var projectPath string
	if path != "" {
		projectPath = filepath.Join(projectRootPath, path)
	} else {
		projectPath = projectRootPath
	}

	// 检查路径是否存在
	if !utils.IsDirectoryExists(projectPath) {
		logger.Info("sub directory path does not exist", logger.String("projectPath", projectPath))
		return []models.FileItem{}, fmt.Errorf("sub directory path does not exist: %s", projectPath)
	}

	// 刷新，重新从 git 上拉取最新的文档和代码
	if err := s.refreshProjectFiles(ctx, userID, projectGuid, projectRootPath, path); err != nil {
		logger.Error("failed to refresh project files", logger.String("error", err.Error()))
	}

	// 加载预览文件配置
	config, err := s.loadPreviewFilesConfig(userID, projectGuid)
	if err != nil {
		return nil, fmt.Errorf("failed to load preview files config: %s", err.Error())
	}

	var files []models.FileItem
	if path == "" { // 根目录：只返回满足条件的文件夹和根目录下的文件
		files, err = s.getRootDirectoryFiles(projectPath, config)
	} else { // 子目录：返回该目录下满足条件的文件
		files, err = s.getSubDirectoryFiles(projectPath, path, config)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get file list: %s", err.Error())
	}
	return files, nil
}

// getRootDirectoryFiles 获取根目录文件
func (s *fileService) getRootDirectoryFiles(projectPath string, config *PreviewFilesConfig) ([]models.FileItem, error) {
	var files []models.FileItem

	// 1. 添加配置中指定的根目录文件
	for _, filePath := range config.Files {
		fullPath := filepath.Join(projectPath, filePath)
		if info, err := os.Stat(fullPath); err == nil && !info.IsDir() {
			files = append(files, models.FileItem{
				Name:       info.Name(),
				Path:       filePath,
				Type:       "file",
				Size:       info.Size(),
				ModifiedAt: info.ModTime().Format(time.RFC3339),
			})
		}
	}

	// 2. 添加配置中指定的文件夹（如果非空）
	for _, folderPath := range config.Folders {
		fullPath := filepath.Join(projectPath, folderPath)
		if info, err := os.Stat(fullPath); err == nil && info.IsDir() {
			// 检查文件夹是否非空
			if s.isDirectoryNotEmpty(fullPath) {
				files = append(files, models.FileItem{
					Name:       info.Name(),
					Path:       folderPath,
					Type:       "folder",
					Size:       0,
					ModifiedAt: info.ModTime().Format(time.RFC3339),
				})
			}
		}
	}

	return files, nil
}

// getSubDirectoryFiles 获取子目录文件
func (s *fileService) getSubDirectoryFiles(projectPath, currentPath string, config *PreviewFilesConfig) ([]models.FileItem, error) {
	var files []models.FileItem
	entries, err := os.ReadDir(projectPath) // 读取目录内容
	if err != nil {
		logger.Error("读取目录内容失败", logger.String("projectPath", projectPath), logger.String("currentPath", currentPath))
		return nil, err
	}

	logger.Info("读取目录内容:", logger.String("projectPath", projectPath), logger.String("currentPath", currentPath))
	for _, entry := range entries {
		// 跳过隐藏文件
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		entryPath := filepath.Join(currentPath, entry.Name())
		fullPath := filepath.Join(projectPath, entry.Name())

		logger.Info("读取目录内容:", logger.String("entryName", entry.Name()), logger.String("entryPath", entryPath), logger.String("fullPath", fullPath))

		info, err := entry.Info()
		if err != nil {
			continue
		}
		bIsDir := entry.IsDir()
		fileItemType := "file"
		if bIsDir {
			fileItemType = "folder"
		}
		if utils.IsPathInFolders(entryPath, config.Folders) {
			files = append(files, *models.NewFileItem(entry.Name(), entryPath, fileItemType, 0, info.ModTime().Format(time.RFC3339)))
			continue
		}

		if !bIsDir && utils.IsPathInFiles(entryPath, config.Files) {
			files = append(files, *models.NewFileItem(entry.Name(), entryPath, fileItemType, info.Size(), info.ModTime().Format(time.RFC3339)))
		}
	}
	return files, nil
}

// isDirectoryNotEmpty 检查目录是否非空
func (s *fileService) isDirectoryNotEmpty(dirPath string) bool {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return false
	}

	// 过滤掉隐藏文件
	for _, entry := range entries {
		if !strings.HasPrefix(entry.Name(), ".") {
			return true
		}
	}

	return false
}

// GetFileContent 获取文件内容
func (s *fileService) GetFileContent(ctx context.Context, userID, projectGuid, filePath, encoding string) (*models.FileContent, error) {
	if filePath == "" {
		return nil, fmt.Errorf("filePath is empty")
	}

	// 构建完整文件路径
	projectPath := utils.GetProjectPath(userID, projectGuid)
	if projectPath == "" {
		return nil, fmt.Errorf("project file path is empty")
	}

	fullPath := filepath.Join(projectPath, filePath)
	info, err := utils.GetFileInfo(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %s", err.Error())
	}
	// 读取文件内容
	content, err := utils.GetFileContent(fullPath, encoding)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %s", err.Error())
	}

	return &models.FileContent{
		Path:       filePath,
		Size:       info.Size(),
		ModifiedAt: info.ModTime().Format(time.RFC3339),
		Content:    content,
	}, nil
}

// SyncEpicsToFiles 将数据库中的 Epics 和 Stories 同步到项目文件
func (s *fileService) SyncEpicsToFiles(ctx context.Context, projectPath string, epics []*models.Epic) error {
	storiesDir := filepath.Join(projectPath, "docs/stories")
	// 确保 stories 目录存在
	if err := os.MkdirAll(storiesDir, 0755); err != nil {
		return fmt.Errorf("failed to create stories directory: %s", err.Error())
	}

	// 1. 获取当前存在的所有 epic 文件
	existingFiles, _ := filepath.Glob(filepath.Join(storiesDir, "epic*.md"))
	existingFileMap := make(map[string]bool)
	for _, f := range existingFiles {
		existingFileMap[filepath.Base(f)] = true
	}

	// 2. 写入数据库中的 epics
	for _, epic := range epics {
		if epic.FilePath == "" { // 如果没有文件路径，生成一个默认的
			epic.FilePath = fmt.Sprintf("epic%d-%s-stories.md", epic.EpicNumber, strings.ToLower(strings.ReplaceAll(epic.Name, " ", "-")))
		}

		filePath := filepath.Join(storiesDir, epic.FilePath)
		content := s.generateEpicMarkdown(epic)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write Epic file: %s", err.Error())
		}
		delete(existingFileMap, epic.FilePath) // 从待删除列表中移除
		logger.Info("Epic file synchronized", logger.String("epicName", epic.Name), logger.String("filePath", filePath))
	}

	// 3. 删除用户已删除的 epic 文件
	for fileName := range existingFileMap {
		filePath := filepath.Join(storiesDir, fileName)
		if err := os.Remove(filePath); err != nil {
			logger.Warn("failed to delete Epic file", logger.String("filePath", filePath), logger.String("error", err.Error()))
		} else {
			logger.Info("Epic file deleted", logger.String("filePath", filePath))
		}
	}
	return nil
}

func (s *fileService) appendStory(content *strings.Builder, story *models.Story) {
	content.WriteString(fmt.Sprintf("### %s: %s\n\n", story.StoryNumber, story.Title))
	if story.Description != "" {
		content.WriteString(fmt.Sprintf("**描述**: %s\n\n", story.Description))
	}
	content.WriteString(fmt.Sprintf("- **优先级**: %s\n", story.Priority))
	content.WriteString(fmt.Sprintf("- **预估天数**: %d 天\n", story.EstimatedDays))
	content.WriteString(fmt.Sprintf("- **状态**: %s\n", story.Status))
	if story.Depends != "" {
		content.WriteString(fmt.Sprintf("- **依赖**: %s\n", story.Depends))
	}
	if story.Techs != "" {
		content.WriteString(fmt.Sprintf("- **技术要点**: %s\n", story.Techs))
	}
	content.WriteString("\n")
	if story.AcceptanceCriteria != "" {
		content.WriteString("**验收标准**:\n")
		content.WriteString(fmt.Sprintf("%s\n\n", story.AcceptanceCriteria))
	}
	if story.Content != "" {
		content.WriteString("**详细内容**:\n")
		content.WriteString(fmt.Sprintf("%s\n\n", story.Content))
	}
	content.WriteString("---\n\n")
}

// generateEpicMarkdown 生成 Epic 的 Markdown 内容
func (s *fileService) generateEpicMarkdown(epic *models.Epic) string {
	var content strings.Builder
	// Epic 标题
	content.WriteString(fmt.Sprintf("# Epic %d: %s\n\n", epic.EpicNumber, epic.Name))
	// Epic 描述
	if epic.Description != "" {
		content.WriteString(fmt.Sprintf("## 描述\n\n%s\n\n", epic.Description))
	}
	// Epic 信息
	content.WriteString("## Epic 信息\n\n")
	content.WriteString(fmt.Sprintf("- **优先级**: %s\n", epic.Priority))
	content.WriteString(fmt.Sprintf("- **预估天数**: %d 天\n", epic.EstimatedDays))
	content.WriteString(fmt.Sprintf("- **状态**: %s\n\n", epic.Status))

	// Stories 列表
	if len(epic.Stories) > 0 {
		content.WriteString("## 用户故事\n\n")
		for _, story := range epic.Stories {
			s.appendStory(&content, &story)
		}
	}

	return content.String()
}
