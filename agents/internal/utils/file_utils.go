package utils

import (
	"app-maker-agents/pkg/logger"
	"os"
)

// 检查目录是否存在
func IsDirectoryExists(filePath string) bool {
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
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

// 检查文件是否存在
func IsFileExists(filePath string) bool {
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
