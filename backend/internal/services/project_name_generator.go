package services

import (
	"autocodeweb-backend/internal/config"
	"autocodeweb-backend/internal/models"
	"autocodeweb-backend/internal/utils"
	"fmt"

	"autocodeweb-backend/pkg/logger"
)

// ProjectNameGenerator 项目名生成器接口
type ProjectNameGenerator interface {
	GenerateProjectConfig(requirements string, projectConfig *models.Project) bool
}

// projectNameGenerator 项目名生成器实现
type projectNameGenerator struct {
	config *config.Config
}

// NewProjectNameGenerator 创建项目名生成器实例
func NewProjectNameGenerator(cfg *config.Config) ProjectNameGenerator {
	return &projectNameGenerator{
		config: cfg,
	}
}

// GenerateProjectConfig 根据需求生成项目配置
func (g *projectNameGenerator) GenerateProjectConfig(requirements string, projectConfig *models.Project) bool {
	logger.Info("开始使用 AI 生成项目配置",
		logger.String("requirements", requirements),
	)

	// 初始化 Ollama 客户端
	client := utils.InitOllamaClient(g.config.AI.Ollama.BaseURL)
	if client == nil {
		logger.Error("无法初始化 Ollama 客户端")
		return g.fallbackToDefaultConfig(requirements, projectConfig)
	}

	// 测试连接
	if err := utils.TestConnection(client); err != nil {
		logger.Error("Ollama 连接测试失败",
			logger.String("error", err.Error()),
		)
		return g.fallbackToDefaultConfig(requirements, projectConfig)
	}

	// 使用 AI 生成项目总结
	summary, err := utils.GenerateProjectSummary(client, requirements)
	if err != nil {
		logger.Error("AI 生成项目总结失败",
			logger.String("error", err.Error()),
		)
		return g.fallbackToDefaultConfig(requirements, projectConfig)
	}

	// 设置项目配置
	projectConfig.Name = summary.Title
	projectConfig.Description = summary.Content
	projectConfig.ApiBaseUrl = "/api/v1"

	// 生成密码
	passwordUtils := utils.NewPasswordUtils()
	projectConfig.AppSecretKey = passwordUtils.GenerateRandomPassword("app")
	projectConfig.RedisPassword = passwordUtils.GenerateRandomPassword("redis")
	projectConfig.JwtSecretKey = passwordUtils.GenerateRandomPassword("jwt")
	projectConfig.DatabasePassword = passwordUtils.GenerateRandomPassword("database")
	projectConfig.Subnetwork = "172.20.0.0/16"

	logger.Info("AI 项目配置生成成功",
		logger.String("projectName", projectConfig.Name),
		logger.String("projectDescription", projectConfig.Description),
	)

	return true
}

// fallbackToDefaultConfig 回退到默认配置
func (g *projectNameGenerator) fallbackToDefaultConfig(requirements string, projectConfig *models.Project) bool {
	logger.Info("使用默认配置生成项目信息")

	// 使用简单的规则生成项目名
	projectName := g.generateSimpleProjectName(requirements)
	projectDescription := requirements

	projectConfig.Name = projectName
	projectConfig.Description = projectDescription
	projectConfig.ApiBaseUrl = "/api/v1"

	// 生成密码
	passwordUtils := utils.NewPasswordUtils()
	projectConfig.AppSecretKey = passwordUtils.GenerateRandomPassword("app")
	projectConfig.RedisPassword = passwordUtils.GenerateRandomPassword("redis")
	projectConfig.JwtSecretKey = passwordUtils.GenerateRandomPassword("jwt")
	projectConfig.DatabasePassword = passwordUtils.GenerateRandomPassword("database")
	projectConfig.Subnetwork = "172.20.0.0/16"

	logger.Info("默认项目配置生成成功",
		logger.String("projectName", projectConfig.Name),
		logger.String("projectDescription", projectConfig.Description),
	)

	return true
}

// generateSimpleProjectName 生成简单的项目名
func (g *projectNameGenerator) generateSimpleProjectName(requirements string) string {
	// 简单的关键词提取和项目名生成
	keywords := []string{"app", "web", "mobile", "platform", "tool", "system"}

	for _, keyword := range keywords {
		if contains(requirements, keyword) {
			return fmt.Sprintf("My%sApp", capitalize(keyword))
		}
	}

	return "MyProject"
}

// contains 检查字符串是否包含子字符串（不区分大小写）
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					findSubstring(s, substr)))
}

// findSubstring 查找子字符串
func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// capitalize 首字母大写
func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return string(s[0]-32) + s[1:]
}
