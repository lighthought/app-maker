package models

import "encoding/json"

// 任务类型常量
const (
	TypeEmailDelivery      = "email:deliver"
	TypeProjectDownload    = "project:download"    // 下载项目
	TypeProjectBackup      = "project:backup"      // 备份项目
	TypeProjectInit        = "project:init"        // 初始化项目
	TypeProjectDevelopment = "project:development" // 开发项目
	TypeWebSocketBroadcast = "ws:broadcast"        // WebSocket 消息广播
)

// 任务优先级
const (
	TaskQueueCritical = 6 // 高优先级
	TaskQueueDefault  = 3 // 中优先级
	TaskQueueLow      = 1 // 低优先级
)

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

// 发送邮件任务的负载
type EmailTaskPayload struct {
	UserID  string `json:"user_id"`
	Content string `json:"content"`
}

// 发送邮件任务的负载转换为 []byte
func (p *EmailTaskPayload) ToBytes() []byte {
	bytes, err := json.Marshal(p)
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

type WebSocketTaskPayload struct {
	ProjectGUID string            `json:"project_guid"`
	Message     *WebSocketMessage `json:"message"`
}

func (p *WebSocketTaskPayload) ToBytes() []byte {
	bytes, err := json.Marshal(p)
	if err != nil {
		return nil
	}
	return bytes
}
