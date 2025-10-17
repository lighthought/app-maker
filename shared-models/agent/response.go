package agent

// AgentHealthResp Agent 健康检查响应
type AgentHealthResp struct {
	Status      string            `json:"status"`
	Version     string            `json:"version"`
	Environment map[string]string `json:"environment"`
}

// 项目环境准备响应
type SetupProjEnvResp struct {
	BmadMethodStatus string `json:"bmad_method_status" example:"success"`
	FrontendStatus   string `json:"frontend_status" example:"success"`
	BackendStatus    string `json:"backend_status" example:"success"`
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

// AgentTaskStatusMessage Agent 任务状态消息（用于 Redis Pub/Sub）
type AgentTaskStatusMessage struct {
	TaskID      string `json:"task_id"`      // 任务ID
	ProjectGuid string `json:"project_guid"` // 项目GUID
	AgentType   string `json:"agent_type"`   // Agent类型
	Status      string `json:"status"`       // 任务状态：running, done, failed
	Message     string `json:"message"`      // 状态消息
	Timestamp   string `json:"timestamp"`    // 时间戳
}
