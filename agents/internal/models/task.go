package models

import "time"

type TaskRequest struct {
	ProjectID string                 `json:"projectId" binding:"required"`
	UserID    string                 `json:"userId" binding:"required"`
	AgentType AgentType              `json:"agentType"`
	Stage     DevStage               `json:"stage" binding:"required"`
	Context   TaskContext            `json:"context" binding:"required"`
	Params    map[string]interface{} `json:"parameters"`
}

type TaskContext struct {
	ProjectID           string                 `json:"projectId" binding:"required"`
	UserID              string                 `json:"userId" binding:"required"`
	ProjectPath         string                 `json:"projectPath" binding:"required"`
	ProjectName         string                 `json:"projectName"`
	CurrentStage        DevStage               `json:"currentStage"`
	Artifacts           []map[string]any       `json:"artifacts"`
	StageInput          map[string]interface{} `json:"stageInput"`
	PreviousStageOutput map[string]interface{} `json:"previousStageOutput"`
}

type AgentTask struct {
	ID         string                 `json:"id"`
	ProjectID  string                 `json:"projectId"`
	UserID     string                 `json:"userId"`
	AgentType  AgentType              `json:"agentType"`
	Stage      DevStage               `json:"stage"`
	Status     TaskStatus             `json:"status"`
	Progress   int                    `json:"progress"`
	Parameters map[string]interface{} `json:"parameters"`
	Context    TaskContext            `json:"context"`
	Result     map[string]interface{} `json:"result,omitempty"`
	Message    string                 `json:"message,omitempty"`
	CreatedAt  time.Time              `json:"createdAt"`
	UpdatedAt  time.Time              `json:"updatedAt"`
}

type TaskResponse struct {
	TaskID  string     `json:"taskId"`
	Status  TaskStatus `json:"status"`
	Message string     `json:"message"`
}

type QueueStats struct {
	AgentType     AgentType `json:"agentType"`
	Queued        int       `json:"queued"`
	InProgress    int       `json:"inProgress"`
	Completed     int       `json:"completed"`
	Failed        int       `json:"failed"`
	ActiveWorkers int       `json:"activeWorkers"`
}

const (
	AgentTypePO        AgentType = "po"
	AgentTypePM        AgentType = "pm"
	AgentTypeDev       AgentType = "dev"
	AgentTypeAnalyse   AgentType = "analyse"
	AgentTypeArchitect AgentType = "architect"
	AgentTypeUXExpert  AgentType = "ux-expert"
)
