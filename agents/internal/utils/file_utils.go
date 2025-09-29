package utils

import "os"

// 检查目录是否存在
func IsDirectoryExists(filePath string) bool {
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}
