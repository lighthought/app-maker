package utils

import (
	"os"

	"github.com/lighthought/app-maker/shared-models/common"
)

// GetEnvOrDefault 获取环境变量或返回默认值
func GetEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// IsDevEnvironment 检查是否为开发环境
func IsDevEnvironment() bool {
	environment := GetEnvOrDefault("APP_ENVIRONMENT", "")
	switch environment {
	case common.EnvironmentLocalDebug, common.EnvironmentDevelopment:
		return true
	default:
		return false
	}
}
