package utils

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/ollama/ollama/api"

	"autocodeweb-backend/pkg/logger"
)

// ProjectSummaryResponse 项目总结响应结构
type ProjectSummaryResponse struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

// InitOllamaClient 初始化 Ollama 客户端（使用 ollama 库）
func InitOllamaClient(baseURL string) (*api.Client, error) {
	if baseURL == "" {
		baseURL = GetEnvOrDefault("OLLAMA_URL", "http://chat.app-maker.localhost:11434")
	}

	ollamaUrl, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("解析 baseURL 失败: %w", err)
	}

	// 使用 ollama 库，但配置为本地 Ollama 服务
	client := api.NewClient(ollamaUrl, http.DefaultClient) // 不需要 API Key
	return client, nil
}

// TestConnection 测试 Ollama 连接
func TestConnection(client *api.Client) error {
	if client == nil {
		return errors.New("Ollama client is nil")
	}
	version, err := client.Version(context.Background())
	logger.Info("Ollama 连接测试成功",
		logger.String("Ollama version", version),
	)
	return err
}

// GenerateProjectSummary 生成项目总结
func GenerateProjectSummary(client *api.Client, requirements string) (*ProjectSummaryResponse, error) {
	if client == nil {
		return nil, errors.New("Ollama client is nil")
	}

	// 构建系统提示词
	systemPrompt := `你是一个需求总结专家。请将以下用户需求文本总结为50字左右，使其易于阅读和理解。总结应简洁明了，并抓住应用或网站需求的主要内容。` +
		`避免使用复杂的句子结构或技术术语。整个对话和指令都应以中文呈现。` +
		`另外，给出符合要点的一到两个单词英文的标题，类似 GirlDress。` +
		`输出json格式的结果，例如:{"title": "GirlDress", "content": "女生装扮应用，分享效果，导入购物链接"}`

	// 构建用户消息
	userMessage := fmt.Sprintf("用户需求：%s", requirements)

	var summary ProjectSummaryResponse
	err := client.Generate(context.Background(), &api.GenerateRequest{
		Model:  GetEnvOrDefault("OLLAMA_MODEL", "deepseek-r1:14b"),
		System: systemPrompt,
		Prompt: userMessage,
	}, func(response api.GenerateResponse) error {

		if err := json.Unmarshal([]byte(response.Response), &summary); err != nil {
			logger.Error("解析 AI 响应失败",
				logger.String("error", err.Error()),
				logger.String("response", response.Response),
			)
			return fmt.Errorf("解析 AI 响应失败: %w", err)
		}

		// 验证响应格式
		if summary.Title == "" || summary.Content == "" {
			logger.Error("AI 响应格式不正确",
				logger.String("title", summary.Title),
				logger.String("content", summary.Content),
			)
			return errors.New("AI 响应格式不正确")
		}

		logger.Info("AI 项目总结生成成功",
			logger.String("title", summary.Title),
			logger.String("content", summary.Content),
		)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("生成 AI 响应失败: %w", err)
	}

	return &summary, nil
}
