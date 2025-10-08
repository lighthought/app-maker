package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"autocodeweb-backend/internal/models"
	"shared-models/logger"
	"shared-models/utils"

	"github.com/hibiken/asynq"
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

	// GetRelativeFiles 获取相对路径的文件列表
	GetRelativeFiles(projectPath, subFolder string) ([]string, error)
}

// projectFileService 项目文件服务实现
type fileService struct {
	asyncClient *asynq.Client
}

// NewProjectFileService 创建项目文件服务
func NewFileService(asyncClient *asynq.Client) FileService {
	return &fileService{
		asyncClient: asyncClient,
	}
}

// loadPreviewFilesConfig 加载预览文件配置
func (s *fileService) loadPreviewFilesConfig(userID, projectGuid string) (*PreviewFilesConfig, error) {
	projectPath := utils.GetProjectPath(userID, projectGuid)
	if projectPath == "" {
		return nil, fmt.Errorf("项目路径为空")
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
		return nil, fmt.Errorf("读取预览文件配置失败: %w", err)
	}

	var config PreviewFilesConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析预览文件配置失败: %w", err)
	}

	return &config, nil
}

// GetProjectFiles 获取项目文件列表
func (s *fileService) GetProjectFiles(ctx context.Context, userID, projectGuid, path string) ([]models.FileItem, error) {
	// 构建项目路径
	projectPath := utils.GetProjectPath(userID, projectGuid)
	if path != "" {
		projectPath = filepath.Join(projectPath, path)
	}

	// 检查路径是否存在
	if utils.IsDirectoryExists(projectPath) == false {
		logger.Info("项目的子目录路径不存在", logger.String("projectPath", projectPath))
		return []models.FileItem{}, fmt.Errorf("项目的子目录路径不存在: %s", projectPath)
	}

	// 加载预览文件配置
	config, err := s.loadPreviewFilesConfig(userID, projectGuid)
	if err != nil {
		return nil, fmt.Errorf("加载预览文件配置失败: %w", err)
	}

	var files []models.FileItem

	if path == "" {
		// 根目录：只返回满足条件的文件夹和根目录下的文件
		files, err = s.getRootDirectoryFiles(projectPath, config)
	} else {
		// 子目录：返回该目录下满足条件的文件
		files, err = s.getSubDirectoryFiles(projectPath, path, config)
	}

	if err != nil {
		return nil, fmt.Errorf("获取文件列表失败: %w", err)
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

// GetRelativeFiles 获取相对路径的文件列表
func (s *fileService) GetRelativeFiles(projectPath, subFolder string) ([]string, error) {
	var fileNames []string

	entries, err := os.ReadDir(filepath.Join(projectPath, subFolder))
	if err != nil {
		logger.Error("读取目录内容失败", logger.String("projectPath", projectPath), logger.String("subFolder", subFolder))
		return nil, err
	}

	for _, entry := range entries {
		fileNames = append(fileNames, entry.Name())
	}

	return fileNames, nil
}

// getSubDirectoryFiles 获取子目录文件
func (s *fileService) getSubDirectoryFiles(projectPath, currentPath string, config *PreviewFilesConfig) ([]models.FileItem, error) {
	var files []models.FileItem

	// 读取目录内容
	entries, err := os.ReadDir(projectPath)
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
			files = append(files, models.FileItem{
				Name:       entry.Name(),
				Path:       entryPath,
				Type:       fileItemType,
				Size:       0,
				ModifiedAt: info.ModTime().Format(time.RFC3339),
			})
			continue
		}

		if !bIsDir && utils.IsPathInFiles(entryPath, config.Files) {
			files = append(files, models.FileItem{
				Name:       entry.Name(),
				Path:       entryPath,
				Type:       fileItemType,
				Size:       info.Size(),
				ModifiedAt: info.ModTime().Format(time.RFC3339),
			})
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
		return nil, fmt.Errorf("文件路径为空")
	}

	// 构建完整文件路径
	projectPath := utils.GetProjectPath(userID, projectGuid)
	if projectPath == "" {
		return nil, fmt.Errorf("项目文件路径为空")
	}

	fullPath := filepath.Join(projectPath, filePath)
	info, err := utils.GetFileInfo(fullPath)
	if err != nil {
		return nil, fmt.Errorf("获取文件信息失败: %w", err)
	}
	// 读取文件内容
	content, err := utils.GetFileContent(fullPath, encoding)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}

	return &models.FileContent{
		Path:       filePath,
		Size:       info.Size(),
		ModifiedAt: info.ModTime().Format(time.RFC3339),
		Content:    content,
	}, nil
}
