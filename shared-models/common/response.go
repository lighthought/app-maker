package common

import "time"

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

// BaseResponse 基础响应结构
type BaseResponse struct {
	Success   bool        `json:"success"`
	Code      int         `json:"code,omitempty"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
	Timestamp string      `json:"timestamp,omitempty"`
}

// SuccessResponse 成功响应
func SuccessResponse(data interface{}, message ...string) *BaseResponse {
	msg := "操作成功"
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}

	return &BaseResponse{
		Success:   true,
		Code:      SUCCESS_CODE,
		Message:   msg,
		Data:      data,
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
	}
}

// AgentResult Agent 执行结果
type AgentResult struct {
	Success         bool   `json:"success"`
	Output          string `json:"output,omitempty"`
	Error           string `json:"error,omitempty"`
	MarkdownContent string `json:"markdown_content,omitempty"`
}

// GetMarkdownContent 获取 Markdown 内容
func (ar *AgentResult) GetMarkdownContent() string {
	if ar.MarkdownContent != "" {
		return ar.MarkdownContent
	}
	return ar.Output
}
