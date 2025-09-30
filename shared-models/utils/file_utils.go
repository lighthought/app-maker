package utils

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"shared-models/logger"
	"strings"
)

const (
	baseDir = "/app/data"
)

// GetProjectPath 获取项目路径
func GetProjectPath(userID, projectGuid string) string {
	return filepath.Join(baseDir, "projects", userID, projectGuid)
}

// GetTemplatePath 获取模板路径
func GetTemplatePath() string {
	return filepath.Join(baseDir, "template.zip")
}

// GetCachePath 获取缓存路径
func GetCachePath() string {
	return filepath.Join(baseDir, "projects", "cache")
}

// isPathInFolders 检查路径是否在文件夹列表中
func IsPathInFolders(path string, folders []string) bool {
	for _, folder := range folders {
		if strings.HasPrefix(path, folder) || path == folder {
			return true
		}
	}
	return false
}

// isPathInFiles 检查路径是否在文件列表中
func IsPathInFiles(path string, files []string) bool {
	for _, file := range files {
		if file == path {
			return true
		}
	}
	return false
}

// 检查文件是否存在
func IsFileExists(filePath string) bool {
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// 检查目录是否存在
func IsDirectoryExists(filePath string) bool {
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

// 获取文件信息
func GetFileInfo(filePath string) (os.FileInfo, error) {
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("文件不存在或路径不正确")
	}
	if info.IsDir() {
		return nil, fmt.Errorf("路径是目录，不是文件")
	}
	return info, nil
}

// 确保目标目录存在
func EnsureDirectoryExists(filePath string) bool {
	if err := os.MkdirAll(filePath, 0755); err != nil {
		logger.Error("创建目录失败",
			logger.String("error", err.Error()),
			logger.String("filePath", filePath),
		)
		return false
	}
	return true
}

// 解压zip文件到指定目录
func ExtractZipFile(zipPath, projectPath string) bool {
	// 打开模板zip文件
	zipFile, err := zip.OpenReader(zipPath)
	if err != nil {
		logger.Error("打开zip文件失败",
			logger.String("error", err.Error()),
			logger.String("zipPath", zipPath),
		)
		return false
	}
	defer zipFile.Close()

	logger.Info("模板zip文件打开成功",
		logger.Int("fileCount", len(zipFile.File)),
		logger.String("templatePath", zipPath),
	)

	// 解压文件
	extractedCount := 0
	for _, file := range zipFile.File {
		// 创建文件路径
		filePath := filepath.Join(projectPath, file.Name)

		// 如果是目录，创建目录
		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(filePath, 0755); err != nil {
				logger.Error("创建目录失败",
					logger.String("error", err.Error()),
					logger.String("filePath", filePath),
				)
				return false
			}
			continue
		}

		// 确保父目录存在
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			logger.Error("创建父目录失败",
				logger.String("error", err.Error()),
				logger.String("filePath", filePath),
			)
			return false
		}

		// 创建文件
		destFile, err := os.Create(filePath)
		if err != nil {
			logger.Error("创建文件失败",
				logger.String("error", err.Error()),
				logger.String("filePath", filePath),
			)
			return false
		}
		defer destFile.Close()

		// 打开源文件
		srcFile, err := file.Open()
		if err != nil {
			logger.Error("打开源文件失败",
				logger.String("error", err.Error()),
				logger.String("fileName", file.Name),
			)
			return false
		}
		defer srcFile.Close()

		// 复制内容
		if _, err := io.Copy(destFile, srcFile); err != nil {
			logger.Error("复制文件内容失败",
				logger.String("error", err.Error()),
				logger.String("fileName", file.Name),
			)
			return false
		}

		extractedCount++
	}

	logger.Info("zip文件解压完成",
		logger.String("zipPath", zipPath),
		logger.String("projectPath", projectPath),
		logger.Int("extractedCount", extractedCount),
	)

	return true
}

// 获取文件所有文本字符串内容
func GetAllTextContent(filePath string) []string {
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return []string{}
	}

	// 解析需要替换的文件列表
	fileList := strings.Split(string(fileContent), "\n")
	return fileList
}

// 获取安全文件路径
func GetSafeFilePath(filePath string) (string, error) {
	// 1. 安全清理文件名，防止目录遍历攻击
	safeFilename := strings.ReplaceAll(filePath, "..", "")

	// 2. 拼接完整的文件路径
	full_path := filepath.Join(baseDir, safeFilename)
	logger.Info("获取安全路径",
		logger.String("filePath", filePath),
		logger.String("full_path", full_path),
	)

	// 4. 检查文件是否存在
	if !IsFileExists(full_path) {
		return "", fmt.Errorf("文件不存在")
	}

	return full_path, nil
}
