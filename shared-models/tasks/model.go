package tasks

import "encoding/json"

// 用于传递任务结果的结构
type TaskResult struct {
	TaskID    string `json:"task_id"`
	Status    string `json:"status"`   // e.g., "pending", "in_progress", "done", "failed"
	Progress  int    `json:"progress"` // 百分比
	Message   string `json:"message"`
	UpdatedAt string `json:"updated_at"`
}

func (t *TaskResult) ToBytes() []byte {
	bytes, err := json.Marshal(t)
	if err != nil {
		return nil
	}
	return bytes
}

// 只有项目ID的负载
type ProjectTaskPayload struct {
	ProjectID   string `json:"project_id"`
	ProjectGuid string `json:"project_guid"`
	ProjectPath string `json:"project_path"`
}

// 只有项目ID的负载转换为 []byte
func (p *ProjectTaskPayload) ToBytes() []byte {
	bytes, err := json.Marshal(p)
	if err != nil {
		return nil
	}
	return bytes
}

// WebSocket消息广播任务负载
type WebSocketTaskPayload struct {
	ProjectGUID string `json:"project_guid"` // 项目GUID
	MessageType string `json:"message_type"` // 消息类型，在常量TypeWebSocketBroadcast中定义
	MessageID   string `json:"message_id"`   // 消息ID，当消息类型为project_message时，必填
	StageID     string `json:"stage_id"`     // 阶段ID，当消息类型为project_stage_update时，必填
	ProjectID   string `json:"project_id"`   // 项目ID，当消息类型为project_info_update时，必填
}

func (p *WebSocketTaskPayload) ToBytes() []byte {
	bytes, err := json.Marshal(p)
	if err != nil {
		return nil
	}
	return bytes
}

// 代理执行任务负载
type AgentExecuteTaskPayload struct {
	ProjectGUID string `json:"project_guid"`
	AgentType   string `json:"agent_type"`
	Message     string `json:"message"`
}

func (a *AgentExecuteTaskPayload) ToBytes() []byte {
	bytes, err := json.Marshal(a)
	if err != nil {
		return nil
	}
	return bytes
}
