package agent

import "encoding/json"

// EnvironmentInfo 环境信息结构体
type EnvironmentInfo struct {
	OS           string `json:"os"`           // 操作系统
	Architecture string `json:"architecture"` // 系统架构
	GoVersion    string `json:"go_version"`   // Go版本
	Runtime      string `json:"runtime"`      // 运行时环境
	Memory       string `json:"memory"`       // 内存使用情况
	CPU          string `json:"cpu"`          // CPU信息
}

// AgentHealthResp Agent 健康检查响应
type AgentHealthResp struct {
	Status    string          `json:"status"`
	Version   string          `json:"version"`
	Tools     []AgentToolInfo `json:"tools"`
	CheckedAt string          `json:"checked_at"`
}

type AgentToolInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
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

type ServiceStatus struct {
	Name      string `json:"name"`
	Status    string `json:"status"`
	Message   string `json:"message"`
	Version   string `json:"version"`
	CheckedAt string `json:"checked_at"`
}

// BackendHealthResp Backend 健康检查响应
type BackendHealthResp struct {
	Status    string           `json:"status"`
	Service   string           `json:"service"`
	Version   string           `json:"version"`
	Timestamp string           `json:"timestamp"`
	Services  []ServiceStatus  `json:"services,omitempty"`
	Agent     *AgentHealthResp `json:"agent,omitempty"`
}

// AgentTaskStatusMessage Agent 任务状态消息（用于 Redis Pub/Sub）
type AgentTaskStatusMessage struct {
	TaskID      string `json:"task_id"`      // 任务ID
	ProjectGuid string `json:"project_guid"` // 项目GUID
	AgentType   string `json:"agent_type"`   // Agent类型
	Status      string `json:"status"`       // 任务状态
	DevStage    string `json:"dev_stage"`    // 开发阶段：initializing, preparing, developing, testing, deploying, completed, failed
	Message     string `json:"message"`      // 状态消息
	Timestamp   string `json:"timestamp"`    // 时间戳
}

func (a *AgentTaskStatusMessage) ToBytes() []byte {
	bytes, err := json.Marshal(a)
	if err != nil {
		return nil
	}
	return bytes
}
