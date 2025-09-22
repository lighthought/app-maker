package utils

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"autocodeweb-backend/pkg/logger"

	deepseek "github.com/cohesion-org/deepseek-go"
	"github.com/cohesion-org/deepseek-go/constants"
	deepseekUtils "github.com/cohesion-org/deepseek-go/utils"
	api "github.com/ollama/ollama/api"
)

// ProjectSummaryResponse 项目总结响应结构
type ProjectSummaryResponse struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

// extractJSONFromResponse 从 AI 响应中提取 JSON 内容
func extractJSONFromResponse(response string) (string, error) {
	// 移除 <think> 标签及其内容
	thinkRegex := regexp.MustCompile(`<think>.*?</think>`)
	response = thinkRegex.ReplaceAllString(response, "")

	// 查找 ```json 代码块
	jsonRegex := regexp.MustCompile("```json\\s*\\n([\\s\\S]*?)\\n```")
	matches := jsonRegex.FindStringSubmatch(response)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1]), nil
	}

	// 如果没有找到代码块，尝试查找纯 JSON 对象
	jsonObjectRegex := regexp.MustCompile(`\{[^{}]*"title"[^{}]*"content"[^{}]*\}`)
	jsonMatch := jsonObjectRegex.FindString(response)
	if jsonMatch != "" {
		return jsonMatch, nil
	}

	// 如果都没有找到，返回原始响应
	return strings.TrimSpace(response), nil
}

// IsOllamaRunning 检查 Ollama 是否运行
func IsOllamaRunning() bool {
	client := http.Client{
		Timeout: 2 * time.Second,
	}
	baseURL := GetEnvOrDefault("OLLAMA_URL", "http://chat.app-maker.localhost:11434")
	resp, err := client.Get(baseURL + "/api/tags")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// ollamaCompletionStream implements the ChatCompletionStream interface for Ollama.
type ollamaCompletionStream struct {
	ctx    context.Context    // Context for stream cancellation
	cancel context.CancelFunc // Function to cancel the context
	resp   *http.Response     // HTTP response from the Ollama API
	reader *bufio.Reader      // Buffered reader for streaming response
}

func convertToOllamaMessages(messages []deepseek.ChatCompletionMessage) []api.Message {
	converted := make([]api.Message, len(messages))
	for i, msg := range messages {
		converted[i] = api.Message{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}
	return converted
}

func (s *ollamaCompletionStream) Recv() (*deepseek.StreamChatCompletionResponse, error) {
	reader := s.reader
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return nil, io.EOF
			}
			return nil, fmt.Errorf("error reading stream: %w", err)
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var ollamaResp deepseek.OllamaStreamResponse
		if err := json.Unmarshal([]byte(line), &ollamaResp); err != nil {
			return nil, fmt.Errorf("unmarshal error: %w, raw data: %s", err, line)
		}

		// Convert Ollama response to StreamChatCompletionResponse format
		response := &deepseek.StreamChatCompletionResponse{
			Model: ollamaResp.Model,
			Choices: []deepseek.StreamChoices{
				{
					Index: 0,
					Delta: deepseek.StreamDelta{
						Content: ollamaResp.Message.Content,
						Role:    ollamaResp.Message.Role,
					},
					FinishReason: ollamaResp.DoneReason,
				},
			},
		}

		if ollamaResp.Done && ollamaResp.Message.Content == "" {
			return nil, io.EOF
		}

		return response, nil
	}
}

// Close terminates the Ollama stream
func (s *ollamaCompletionStream) Close() error {
	s.cancel()
	err := s.resp.Body.Close()
	if err != nil {
		return fmt.Errorf("failed to close response body: %w", err)
	}
	return nil
}

// 重写 CreateOllamaChatCompletionStream 方法，支持非本地的 ollama
func CreateOllamaChatCompletionStream(
	ctx context.Context,
	request *deepseek.StreamChatCompletionRequest,
) (*ollamaCompletionStream, error) {
	if !IsOllamaRunning() {
		return &ollamaCompletionStream{}, fmt.Errorf("Ollama server is not running")
	}
	if request == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	c := deepseek.Client{
		BaseURL: GetEnvOrDefault("OLLAMA_URL", "http://chat.app-maker.localhost:11434"),
	}
	var s bool = true
	// Convert messages to Ollama format
	ollamaRequest := &api.ChatRequest{
		Model:    request.Model,
		Messages: convertToOllamaMessages(request.Messages),
		Stream:   &s,
	}

	req, err := deepseekUtils.NewRequestBuilder(c.AuthToken).
		SetBaseURL(c.BaseURL).
		SetPath("/api/chat/").
		SetBodyFromStruct(ollamaRequest).
		Build(ctx)

	if err != nil {
		return nil, fmt.Errorf("error building request: %w", err)
	}

	resp, err := deepseek.HandleSendChatCompletionRequest(c, req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, deepseek.HandleAPIError(resp)
	}

	ctx, cancel := context.WithCancel(ctx)
	stream := &ollamaCompletionStream{
		ctx:    ctx,
		cancel: cancel,
		resp:   resp,
		reader: bufio.NewReader(resp.Body),
	}
	return stream, nil
}

func convertToDeepseekResponse(response api.ChatResponse) *deepseek.ChatCompletionResponse {
	return &deepseek.ChatCompletionResponse{
		Model:   response.Model,
		Created: response.CreatedAt.Unix(),
		Choices: []deepseek.Choice{
			{
				Message: deepseek.Message{
					Role:    response.Message.Role,
					Content: response.Message.Content,
				},
				FinishReason: response.DoneReason,
			},
		},
		Usage: deepseek.Usage{
			TotalTokens: response.PromptEvalCount + response.EvalCount,
		},
	}
}

func CreateOllamaChatCompletion(req *deepseek.ChatCompletionRequest) (deepseek.ChatCompletionResponse, error) {
	if !IsOllamaRunning() {
		return deepseek.ChatCompletionResponse{}, fmt.Errorf("Ollama server is not running")
	}

	if req == nil {
		return deepseek.ChatCompletionResponse{}, fmt.Errorf("request cannot be nil")
	}

	ollamaUrl := GetEnvOrDefault("OLLAMA_URL", "http://chat.app-maker.localhost:11434")
	baseUrl, err := url.Parse(ollamaUrl)
	if err != nil {
		return deepseek.ChatCompletionResponse{}, fmt.Errorf("failed to parse ollama url: %w", err)
	}
	client := api.NewClient(baseUrl, http.DefaultClient)

	var lastResponse api.ChatResponse
	response := func(response api.ChatResponse) error {
		lastResponse = response
		return nil
	}

	stream := false
	err = client.Chat(context.Background(), &api.ChatRequest{
		Model:    req.Model,
		Messages: convertToOllamaMessages(req.Messages),
		Stream:   &stream,
	}, response)

	if err != nil {
		return deepseek.ChatCompletionResponse{}, fmt.Errorf("error sending request: %w", err)
	}

	convertedResponse := convertToDeepseekResponse(lastResponse)
	return *convertedResponse, nil
}

// GenerateProjectSummary 生成项目总结
func GenerateProjectSummary(requirements string) (*ProjectSummaryResponse, error) {
	if !IsOllamaRunning() {
		return nil, errors.New("Ollama server is not running")
	}

	// 构建系统提示词
	systemPrompt := `你是一个需求总结专家。请将以下用户需求文本总结为50字左右，使其易于阅读和理解。` +
		`总结应简洁明了，并抓住应用或网站需求的主要内容。` +
		`避免使用复杂的句子结构或技术术语。整个对话和指令都应以中文呈现。` +
		`另外，给出符合要点的一到两个单词英文的标题，类似 GirlDress。` +
		`输出json格式的结果，例如: {"title": "GirlDress", "content": "女生装扮应用，分享效果，导入购物链接"}` +
		`限制：不要输出任何图标、表情、特殊符号、emoji等`

	req := &deepseek.ChatCompletionRequest{
		Model: GetEnvOrDefault("OLLAMA_MODEL", "deepseek-r1:14b"),
		Messages: []deepseek.ChatCompletionMessage{
			{Role: constants.ChatMessageRoleSystem, Content: systemPrompt},
			{Role: constants.ChatMessageRoleUser, Content: fmt.Sprintf("用户需求：%s", requirements)}},
	}

	res, err := CreateOllamaChatCompletion(req)
	if err != nil {
		return nil, fmt.Errorf("生成 AI 响应失败: %w", err)
	}

	fullMessage := res.Choices[0].Message.Content
	logger.Info("The reponse is: " + fullMessage)

	// 提取 JSON 内容
	jsonContent, err := extractJSONFromResponse(fullMessage)
	if err != nil {
		logger.Error("提取 JSON 内容失败",
			logger.String("error", err.Error()),
			logger.String("response", fullMessage),
		)
		return nil, fmt.Errorf("提取 JSON 内容失败: %w", err)
	}

	logger.Info("Extracted JSON content: " + jsonContent)

	var summary ProjectSummaryResponse

	if err := json.Unmarshal([]byte(jsonContent), &summary); err != nil {
		logger.Error("解析 AI 响应失败",
			logger.String("error", err.Error()),
			logger.String("jsonContent", jsonContent),
			logger.String("originalResponse", fullMessage),
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
