package project

import "time"

// Project 项目模型
type Project struct {
	ID           string    `json:"id"`
	GUID         string    `json:"guid"`
	UserID       string    `json:"user_id"`
	Name         string    `json:"name"`
	Requirements string    `json:"requirements"`
	ProjectPath  string    `json:"project_path"`
	Status       string    `json:"status"`
	DevStatus    string    `json:"dev_status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// SetupProjectEnvReq 项目环境准备请求
type SetupProjectEnvReq struct {
	ProjectID       string `json:"project_id" validate:"required" example:"1234567890"`
	ProjectGuid     string `json:"project_guid" validate:"required" example:"1234567890"`
	GitlabRepoUrl   string `json:"gitlab_repo_url" validate:"required" example:"https://gitlab.example.com/project.git"`
	SetupBmadMethod bool   `json:"setup_bmad_method" example:"true"`
	BmadCliType     string `json:"bmad_cli_type" example:"claude"`
}

// DevStage 开发阶段
type DevStage struct {
	ID          string    `json:"id"`
	ProjectID   string    `json:"project_id"`
	ProjectGuid string    `json:"project_guid"`
	Name        string    `json:"name"`
	Status      string    `json:"status"`
	Progress    int       `json:"progress"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ConversationMessage 会话消息
type ConversationMessage struct {
	ID              string    `json:"id"`
	ProjectGuid     string    `json:"project_guid"`
	Type            string    `json:"type"` // user, agent
	AgentRole       string    `json:"agent_role,omitempty"`
	AgentName       string    `json:"agent_name,omitempty"`
	Content         string    `json:"content"`
	IsMarkdown      bool      `json:"is_markdown"`
	MarkdownContent string    `json:"markdown_content,omitempty"`
	IsExpanded      bool      `json:"is_expanded"`
	CreatedAt       time.Time `json:"created_at"`
}

// ProjectTaskPayload 项目任务载荷
type ProjectTaskPayload struct {
	ProjectID   string `json:"project_id"`
	ProjectGuid string `json:"project_guid"`
	UserID      string `json:"user_id"`
}
