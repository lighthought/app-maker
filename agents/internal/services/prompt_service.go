package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/lighthought/app-maker/shared-models/logger"
)

// PromptTemplate 定义 Prompt 模板结构
type PromptTemplate struct {
	Template string `json:"template"`
}

// PromptConfig 定义 Prompt 配置结构
type PromptConfig map[string]map[string]PromptTemplate

// PromptService Prompt 服务接口
type PromptService interface {
	// GetPrompt 获取并渲染 Prompt
	GetPrompt(agentType, action string, data interface{}) (string, error)
	// Reload 重新加载配置
	Reload() error
}

type promptService struct {
	configPath string
	config     PromptConfig
}

// NewPromptService 创建 Prompt 服务
func NewPromptService(configPath string) PromptService {
	s := &promptService{
		configPath: configPath,
		config:     make(PromptConfig),
	}
	if err := s.Reload(); err != nil {
		logger.Error("Failed to load prompt config", logger.String("error", err.Error()))
	}
	return s
}

// Reload 重新加载配置
func (s *promptService) Reload() error {
	content, err := os.ReadFile(s.configPath)
	if err != nil {
		// 尝试从默认位置加载
		defaultPath := filepath.Join("configs", "prompts.json")
		content, err = os.ReadFile(defaultPath)
		if err != nil {
			return fmt.Errorf("failed to read prompt config file: %w", err)
		}
	}

	var config PromptConfig
	if err := json.Unmarshal(content, &config); err != nil {
		return fmt.Errorf("failed to parse prompt config: %w", err)
	}

	s.config = config
	logger.Info("Prompt config loaded successfully")
	return nil
}

// GetPrompt 获取并渲染 Prompt
func (s *promptService) GetPrompt(agentType, action string, data interface{}) (string, error) {
	agentConfig, ok := s.config[agentType]
	if !ok {
		return "", fmt.Errorf("agent type not found: %s", agentType)
	}

	promptTemplate, ok := agentConfig[action]
	if !ok {
		return "", fmt.Errorf("action not found: %s for agent: %s", action, agentType)
	}

	// 解析模板
	tmpl, err := template.New("prompt").Parse(promptTemplate.Template)
	if err != nil {
		return "", fmt.Errorf("failed to parse prompt template: %w", err)
	}

	// 渲染模板
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute prompt template: %w", err)
	}

	return buf.String(), nil
}
