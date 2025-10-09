package models

import "encoding/json"

// CommandResult 命令执行结果
type CommandResult struct {
	Success bool   `json:"success"`
	Output  string `json:"output"`
	Error   string `json:"error,omitempty"`
}

// claude 命令的 json 输出结果
type ClaudeResponse struct {
	Type          string `json:"type"`
	Result        string `json:"result"`
	Subtype       string `json:"subtype"`
	IsError       bool   `json:"is_error"`
	DurationMs    int    `json:"duration_ms"`
	DurationApiMs int    `json:"duration_api_ms"`
	SessionID     string `json:"session_id"`
	Usage         struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

func (c *ClaudeResponse) ToJsonString() string {
	json, err := json.Marshal(c)
	if err != nil {
		return ""
	}
	return string(json)
}
