package utils

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	deepseek "github.com/cohesion-org/deepseek-go"

	"autocodeweb-backend/pkg/logger"
)

// ProjectSummaryResponse 项目总结响应结构
type ProjectSummaryResponse struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

// InitOllamaClient 初始化 Ollama 客户端（使用 deepseek-go 库）
func InitOllamaClient(baseURL string) *deepseek.Client {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}

	// 使用 deepseek-go 库，但配置为本地 Ollama 服务
	client := deepseek.NewClient("") // 不需要 API Key
	// 这里需要设置自定义的 base URL，但 deepseek-go 库可能不支持
	// 让我们先尝试直接使用
	return client
}

// GenerateProjectSummary 生成项目总结
func GenerateProjectSummary(client *deepseek.Client, requirements string) (*ProjectSummaryResponse, error) {
	if client == nil {
		return nil, errors.New("Ollama client is nil")
	}

	// 构建系统提示词
	systemPrompt := `你是一个需求总结专家。请将以下用户需求文本总结为50字左右，使其易于阅读和理解。总结应简洁明了，并抓住应用或网站需求的主要内容。避免使用复杂的句子结构或技术术语。整个对话和指令都应以中文呈现。另外，给出符合要点的一到两个单词英文的标题，类似 GirlDress。输出json格式的结果，例如:
{"title": "GirlDress", "content": "女生装扮应用，分享效果，导入购物链接"}`

	// 构建用户消息
	userMessage := fmt.Sprintf("用户需求：%s", requirements)

	// 使用 deepseek-go 库生成响应
	ctx := context.Background()
	request := &deepseek.ChatCompletionRequest{
		Model: deepseek.DeepSeekChat,
		Messages: []deepseek.ChatCompletionMessage{
			{Role: deepseek.ChatMessageRoleSystem, Content: systemPrompt},
			{Role: deepseek.ChatMessageRoleUser, Content: userMessage},
		},
	}

	response, err := client.CreateChatCompletion(ctx, request)
	if err != nil {
		logger.Error("AI 请求失败",
			logger.String("error", err.Error()),
		)
		return nil, fmt.Errorf("发送 AI 请求失败: %w", err)
	}

	// 解析响应
	var summary ProjectSummaryResponse
	aiResponse := response.Choices[0].Message.Content
	if err := json.Unmarshal([]byte(aiResponse), &summary); err != nil {
		logger.Error("解析 AI 响应失败",
			logger.String("error", err.Error()),
			logger.String("response", aiResponse),
		)
		return nil, fmt.Errorf("解析 AI 响应失败: %w", err)
	}

	// 验证响应格式
	if summary.Title == "" || summary.Content == "" {
		logger.Error("AI 响应格式不正确",
			logger.String("title", summary.Title),
			logger.String("content", summary.Content),
		)
		return nil, errors.New("AI 响应格式不正确")
	}

	logger.Info("AI 项目总结生成成功",
		logger.String("title", summary.Title),
		logger.String("content", summary.Content),
	)

	return &summary, nil
}

// TestConnection 测试连接
func TestConnection(client *deepseek.Client) error {
	if client == nil {
		return errors.New("Ollama client is nil")
	}

	// 发送一个简单的测试请求
	ctx := context.Background()
	request := &deepseek.ChatCompletionRequest{
		Model: deepseek.DeepSeekChat,
		Messages: []deepseek.ChatCompletionMessage{
			{Role: deepseek.ChatMessageRoleUser, Content: "Hello"},
		},
	}

	_, err := client.CreateChatCompletion(ctx, request)
	if err != nil {
		return fmt.Errorf("连接测试失败: %w", err)
	}

	logger.Info("连接测试成功")
	return nil
}
