package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"autocodeweb-backend/internal/models"
	"shared-models/logger"
	"shared-models/utils"
)

// ProjectTemplateService 项目模板服务接口
type ProjectTemplateService interface {
	// 项目模板管理
	InitializeProject(ctx context.Context, project *models.Project) error
	ExtractTemplate(ctx context.Context, projectID string, projectPath string) error
	ReplacePlaceholders(ctx context.Context, projectPath string, project *models.Project) error
	// 重命名文件
	RenameFiles(ctx context.Context, projectPath string, project *models.Project) error
}

// projectTemplateService 项目模板服务实现
type projectTemplateService struct {
	fileService FileService
}

// NewProjectTemplateService 创建项目模板服务实例
func NewProjectTemplateService(fileService FileService) ProjectTemplateService {
	return &projectTemplateService{
		fileService: fileService,
	}
}

// InitializeProject 初始化项目
func (s *projectTemplateService) InitializeProject(ctx context.Context, project *models.Project) error {
	templatePath := utils.GetTemplatePath()
	logger.Info("==> enter. 开始初始化项目模板",
		logger.String("projectID", project.ID),
		logger.String("projectPath", project.ProjectPath),
		logger.String("templatePath", templatePath),
	)

	// 检查模板文件是否存在
	if !utils.IsFileExists(templatePath) {
		logger.Error("模板文件不存在", logger.String("templatePath", templatePath))
		return fmt.Errorf("template file not found: %s", templatePath)
	}

	// 1. 解压模板
	logger.Info("==> 1. 开始解压模板文件")
	if err := s.ExtractTemplate(ctx, project.ID, project.ProjectPath); err != nil {
		logger.Error("解压模板失败",
			logger.String("error", err.Error()),
			logger.String("projectID", project.ID),
		)
		return fmt.Errorf("failed to extract template: %w", err)
	}
	logger.Info("模板解压完成", logger.String("projectID", project.ID))

	// 2. 替换占位符
	logger.Info("==> 2. 开始替换文件占位符", logger.String("projectID", project.ID))
	if err := s.ReplacePlaceholders(ctx, project.ProjectPath, project); err != nil {
		logger.Error("替换占位符失败",
			logger.String("error", err.Error()),
			logger.String("projectID", project.ID),
		)
		return fmt.Errorf("failed to replace placeholders: %w", err)
	}
	logger.Info("占位符替换完成", logger.String("projectID", project.ID))

	// 3. 重命名文件
	logger.Info("==> 3. 开始重命名文件", logger.String("projectID", project.ID))
	if err := s.RenameFiles(ctx, project.ProjectPath, project); err != nil {
		logger.Error("重命名文件失败",
			logger.String("error", err.Error()),
			logger.String("projectID", project.ID),
		)
		return fmt.Errorf("failed to rename files: %w", err)
	}
	logger.Info("文件重命名完成", logger.String("projectID", project.ID))

	logger.Info("==> exit. 项目模板初始化完成",
		logger.String("projectID", project.ID),
		logger.String("projectPath", project.ProjectPath),
	)

	return nil
}

// ExtractTemplate 解压项目模板
func (s *projectTemplateService) ExtractTemplate(ctx context.Context, projectID string, projectPath string) error {
	templatePath := utils.GetTemplatePath()
	logger.Info("开始解压模板",
		logger.String("projectID", projectID),
		logger.String("projectPath", projectPath),
		logger.String("templatePath", templatePath),
	)

	if utils.EnsureDirectoryExists(projectPath) != nil {
		logger.Error("项目目录创建失败", logger.String("projectPath", projectPath))
		return fmt.Errorf("failed to create project directory: %s", projectPath)
	}

	logger.Info("项目目录创建成功",
		logger.String("projectID", projectID),
		logger.String("projectPath", projectPath),
	)

	if !utils.ExtractZipFile(templatePath, projectPath) {
		logger.Error("模板解压失败", logger.String("projectID", projectID))
		return fmt.Errorf("failed to extract template: %s", templatePath)
	}

	logger.Info("模板解压完成",
		logger.String("projectID", projectID),
		logger.String("projectPath", projectPath),
	)

	return nil
}

// ReplacePlaceholders 替换文件中的占位符
func (s *projectTemplateService) ReplacePlaceholders(ctx context.Context, projectPath string, project *models.Project) error {
	// 读取replace.txt文件，获取需要替换的文件列表
	replaceFilePath := filepath.Join(projectPath, "replace.txt")

	fileList := utils.GetAllTextContent(replaceFilePath)
	if len(fileList) == 0 {
		return s.replaceInAllFiles(projectPath, project)
	}

	// 替换每个文件中的占位符
	for _, filePath := range fileList {
		filePath = strings.TrimSpace(filePath)
		if filePath == "" {
			continue
		}

		// 将 Windows 路径分隔符转换为 Linux 路径分隔符
		filePath = strings.ReplaceAll(filePath, "\\", "/")

		fullPath := filepath.Join(projectPath, filePath)
		if err := s.replaceInFile(fullPath, project); err != nil {
			return fmt.Errorf("failed to replace in file %s: %w", filePath, err)
		}
	}

	// 删除replace.txt文件
	os.Remove(replaceFilePath)

	return nil
}

// replaceInAllFiles 替换所有文本文件中的占位符
func (s *projectTemplateService) replaceInAllFiles(projectPath string, project *models.Project) error {
	return filepath.Walk(projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录
		if info.IsDir() {
			return nil
		}

		// 只处理文本文件
		ext := strings.ToLower(filepath.Ext(path))
		textExtensions := []string{".txt", ".md", ".json", ".yaml", ".yml", ".js", ".ts", ".vue", ".go", ".py", ".java", ".xml", ".html", ".css", ".scss", ".sh", ".bat", ".dockerfile", ".env"}

		isTextFile := false
		for _, textExt := range textExtensions {
			if ext == textExt {
				isTextFile = true
				break
			}
		}

		if isTextFile {
			if err := s.replaceInFile(path, project); err != nil {
				return fmt.Errorf("failed to replace in file %s: %w", path, err)
			}
		}

		return nil
	})
}

// replaceInFile 替换单个文件中的占位符
func (s *projectTemplateService) replaceInFile(filePath string, project *models.Project) error {
	// 读取文件内容
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// 定义替换映射
	replacements := map[string]string{
		"${PRODUCT_NAME}":      project.Name,
		"${PRODUCT_DESC}":      project.Description,
		"${APP_SECRET_KEY}":    project.AppSecretKey,
		"${DATABASE_PASSWORD}": project.DatabasePassword,
		"${REDIS_PASSWORD}":    project.RedisPassword,
		"${JWT_SECRET_KEY}":    project.JwtSecretKey,
		"${SUBNETWORK}":        project.Subnetwork,
		"${API_BASE_URL}":      project.ApiBaseUrl,
		"${BACKEND_PORT}":      fmt.Sprintf("%d", project.BackendPort),
		"${FRONTEND_PORT}":     fmt.Sprintf("%d", project.FrontendPort),
		"${PROJECT_ID}":        project.ID,
		"${REDIS_PORT}":        fmt.Sprintf("%d", project.RedisPort),
		"${DATABASE_PORT}":     fmt.Sprintf("%d", project.PostgresPort),
		"${DATABASE_NAME}":     project.Name,
		"${DATABASE_USER}":     project.Name,
		"${USER_ID}":           project.UserID,
	}

	// 执行替换
	newContent := string(content)
	for placeholder, value := range replacements {
		newContent = strings.ReplaceAll(newContent, placeholder, value)
	}

	// 如果内容有变化，写回文件
	if newContent != string(content) {
		if err := os.WriteFile(filePath, []byte(newContent), 0644); err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}
	}

	return nil
}

// RenameFiles 重命名文件
func (s *projectTemplateService) RenameFiles(ctx context.Context, projectPath string, project *models.Project) error {
	// 读取rename.txt文件，获取需要替换的文件列表
	renameFilePath := filepath.Join(projectPath, "rename.txt")
	renameFile, err := os.ReadFile(renameFilePath)
	if err != nil {
		// 写文件日志
		logger.Error("读取rename.txt文件失败",
			logger.String("error", err.Error()),
			logger.String("renameFilePath", renameFilePath),
		)
		return nil
	}

	// 解析需要重命名的文件列表
	fileList := strings.Split(string(renameFile), "\n")

	// 重命名每个文件
	for _, filePath := range fileList {
		filePath := strings.TrimSpace(filePath)
		if filePath == "" {
			continue
		}

		// 将 Windows 路径分隔符转换为 Linux 路径分隔符
		filePath = strings.ReplaceAll(filePath, "\\", "/")

		from_to_paths := strings.Split(filePath, ",")
		if len(from_to_paths) != 2 {
			logger.Error("重命名文件格式错误",
				logger.String("filePath", filePath),
			)
			continue
		}

		// 重命名文件
		fromPath := filepath.Join(projectPath, strings.TrimSpace(from_to_paths[0]))
		toPath := filepath.Join(projectPath, strings.TrimSpace(from_to_paths[1]))
		if err := os.Rename(fromPath, toPath); err != nil {
			logger.Error("重命名文件失败",
				logger.String("error", err.Error()),
				logger.String("fromPath", fromPath),
				logger.String("toPath", toPath),
			)
			continue
		}
	}

	// 删除rename.txt文件
	os.Remove(renameFilePath)

	return nil
}
