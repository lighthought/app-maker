package common

// Response 通用响应结构
type Response struct {
	Code      int         `json:"code" example:"0"`
	Message   string      `json:"message" example:"success"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp string      `json:"timestamp" example:"2025-08-29T10:00:00Z"`
}

// ErrorResponse 错误响应结构
type ErrorResponse struct {
	Code      int    `json:"code" example:"400"`
	Message   string `json:"message" example:"请求参数错误"`
	Timestamp string `json:"timestamp" example:"2025-08-29T10:00:00Z"`
}

// AgentResult Agent 执行结果
type AgentResult struct {
	Output          string `json:"output,omitempty"`
	Error           string `json:"error,omitempty"`
	MarkdownContent string `json:"markdown_content,omitempty"`
}

// GetMarkdownContent 获取 Markdown 内容
func (ar *AgentResult) GetMarkdownContent() string {
	if ar.MarkdownContent != "" {
		return ar.MarkdownContent
	}
	if ar.Error != "" {
		return ar.Error
	}
	return ar.Output
}
