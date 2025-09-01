package services

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"autocodeweb-backend/internal/models"
)

// ProjectTemplateService 项目模板服务接口
type ProjectTemplateService interface {
	// 项目模板管理
	InitializeProject(ctx context.Context, project *models.Project) error
	ExtractTemplate(ctx context.Context, projectID string, projectPath string) error
	ReplacePlaceholders(ctx context.Context, projectPath string, project *models.Project) error
}

// projectTemplateService 项目模板服务实现
type projectTemplateService struct {
	templatePath string
}

// NewProjectTemplateService 创建项目模板服务实例
func NewProjectTemplateService(templatePath string) ProjectTemplateService {
	return &projectTemplateService{
		templatePath: templatePath,
	}
}

// InitializeProject 初始化项目
func (s *projectTemplateService) InitializeProject(ctx context.Context, project *models.Project) error {
	// 1. 解压模板
	if err := s.ExtractTemplate(ctx, project.ID, project.ProjectPath); err != nil {
		return fmt.Errorf("failed to extract template: %w", err)
	}

	// 2. 替换占位符
	if err := s.ReplacePlaceholders(ctx, project.ProjectPath, project); err != nil {
		return fmt.Errorf("failed to replace placeholders: %w", err)
	}

	return nil
}

// ExtractTemplate 解压项目模板
func (s *projectTemplateService) ExtractTemplate(ctx context.Context, projectID string, projectPath string) error {
	// 确保目标目录存在
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	// 打开模板zip文件
	zipFile, err := zip.OpenReader(s.templatePath)
	if err != nil {
		return fmt.Errorf("failed to open template zip: %w", err)
	}
	defer zipFile.Close()

	// 解压文件
	for _, file := range zipFile.File {
		// 创建文件路径
		filePath := filepath.Join(projectPath, file.Name)

		// 如果是目录，创建目录
		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(filePath, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", filePath, err)
			}
			continue
		}

		// 确保父目录存在
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			return fmt.Errorf("failed to create parent directory for %s: %w", filePath, err)
		}

		// 创建文件
		destFile, err := os.Create(filePath)
		if err != nil {
			return fmt.Errorf("failed to create file %s: %w", filePath, err)
		}
		defer destFile.Close()

		// 打开源文件
		srcFile, err := file.Open()
		if err != nil {
			return fmt.Errorf("failed to open source file %s: %w", file.Name, err)
		}
		defer srcFile.Close()

		// 复制内容
		if _, err := io.Copy(destFile, srcFile); err != nil {
			return fmt.Errorf("failed to copy file %s: %w", file.Name, err)
		}
	}

	return nil
}

// ReplacePlaceholders 替换文件中的占位符
func (s *projectTemplateService) ReplacePlaceholders(ctx context.Context, projectPath string, project *models.Project) error {
	// 读取replace.txt文件，获取需要替换的文件列表
	replaceFilePath := filepath.Join(projectPath, "replace.txt")
	replaceFile, err := os.ReadFile(replaceFilePath)
	if err != nil {
		// 如果replace.txt不存在，则替换所有文本文件
		return s.replaceInAllFiles(projectPath, project)
	}

	// 解析需要替换的文件列表
	fileList := strings.Split(string(replaceFile), "\n")

	// 替换每个文件中的占位符
	for _, filePath := range fileList {
		filePath = strings.TrimSpace(filePath)
		if filePath == "" {
			continue
		}

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
		"${PRODUCT_NAME}":  project.Name,
		"${PRODUCT_DESC}":  project.Description,
		"${BACKEND_PORT}":  fmt.Sprintf("%d", project.BackendPort),
		"${FRONTEND_PORT}": fmt.Sprintf("%d", project.FrontendPort),
		"${PROJECT_ID}":    project.ID,
		"${USER_ID}":       project.UserID,
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
