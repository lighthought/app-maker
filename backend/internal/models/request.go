package models

// PaginationRequest 分页请求
type PaginationRequest struct {
	Page     int `json:"page" form:"page" binding:"min=1" example:"1"`
	PageSize int `json:"page_size" form:"page_size" binding:"min=1,max=100" example:"10"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required,min=6" example:"password123"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Username string `json:"username" binding:"required,min=3,max=20" example:"username"`
	Password string `json:"password" binding:"required,min=6" example:"password123"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required" example:"oldpassword123"`
	NewPassword string `json:"new_password" binding:"required,min=6" example:"newpassword123"`
}

// UpdateProfileRequest 更新用户档案请求
type UpdateProfileRequest struct {
	Username string `json:"username" binding:"omitempty,min=3,max=20" example:"newusername"`
	Email    string `json:"email" binding:"omitempty,email" example:"newemail@example.com"`
}

// UpdateUserSettingsRequest 更新用户设置请求
type UpdateUserSettingsRequest struct {
	DefaultCliTool       string `json:"default_cli_tool" binding:"omitempty,oneof=claude-code qwen-code gemini" example:"claude-code"`
	DefaultAiModel       string `json:"default_ai_model" binding:"omitempty" example:"glm-4.6"`
	DefaultModelProvider string `json:"default_model_provider" binding:"omitempty,oneof=ollama zhipu anthropic openai vllm" example:"zhipu"`
	DefaultModelApiUrl   string `json:"default_model_api_url" binding:"omitempty,url" example:"https://open.bigmodel.cn/api/anthropic"`
	DefaultApiToken      string `json:"default_api_token" binding:"omitempty" example:"sk-..."`
	AutoGoNext           *bool  `json:"auto_go_next" binding:"omitempty" example:"true"` // 自动进入下一阶段配置
}

// CreateProjectRequest 创建项目请求
type CreateProjectRequest struct {
	Requirements string `json:"requirements" binding:"required" example:"项目需求描述"`
}

// UpdateProjectRequest 更新项目请求
type UpdateProjectRequest struct {
	Name          *string `json:"name" binding:"omitempty,min=1,max=200" example:"项目名称"`
	Description   *string `json:"description" binding:"omitempty" example:"项目描述"`
	CliTool       *string `json:"cli_tool" binding:"omitempty,oneof=claude-code qwen-code gemini" example:"claude-code"`
	AiModel       *string `json:"ai_model" binding:"omitempty" example:"glm-4.6"`
	ModelProvider *string `json:"model_provider" binding:"omitempty,oneof=ollama zhipu anthropic openai vllm" example:"zhipu"`
	ModelApiUrl   *string `json:"model_api_url" binding:"omitempty" example:"https://open.bigmodel.cn/api/anthropic"`
}

// ProjectListRequest 项目列表请求
type ProjectListRequest struct {
	PaginationRequest
	Status string   `json:"status" form:"status" binding:"omitempty,oneof=draft in_progress completed failed" example:"in_progress"`
	TagIDs []string `json:"tag_ids" form:"tag_ids" example:"['tag1', 'tag2']"`
	UserID string   `json:"user_id" form:"user_id" example:"USER_00000000001"`
	Search string   `json:"search" form:"search" example:"项目名称关键词"`
}

// JenkinsBuildRequest Jenkins 构建请求
type JenkinsBuildRequest struct {
	UserID      string `json:"user_id"`
	ProjectID   string `json:"project_id"`
	ProjectPath string `json:"project_path"`
	BuildType   string `json:"build_type"` // dev 或 prod
}

// ChatWithAgentRequest 与 Agent 对话请求
type ChatWithAgentRequest struct {
	AgentType string `json:"agent_type" binding:"required"`
	Content   string `json:"content" binding:"required"`
}

// UpdateEpicOrderRequest 更新 Epic 排序请求
type UpdateEpicOrderRequest struct {
	Order int `json:"order" binding:"required,min=0"`
}

// UpdateEpicRequest 更新 Epic 请求
type UpdateEpicRequest struct {
	Name          *string `json:"name,omitempty" binding:"omitempty,min=1,max=200"`
	Description   *string `json:"description,omitempty"`
	Priority      *string `json:"priority,omitempty" binding:"omitempty,oneof=P0 P1 P2 P3"`
	EstimatedDays *int    `json:"estimated_days,omitempty" binding:"omitempty,min=0"`
}

// UpdateStoryOrderRequest 更新 Story 排序请求
type UpdateStoryOrderRequest struct {
	Order int `json:"order" binding:"required,min=0"`
}

// UpdateStoryRequest 更新 Story 请求
type UpdateStoryRequest struct {
	Title              *string `json:"title,omitempty" binding:"omitempty,min=1,max=200"`
	Description        *string `json:"description,omitempty"`
	Priority           *string `json:"priority,omitempty" binding:"omitempty,oneof=P0 P1 P2 P3"`
	EstimatedDays      *int    `json:"estimated_days,omitempty" binding:"omitempty,min=0"`
	Depends            *string `json:"depends,omitempty"`
	Techs              *string `json:"techs,omitempty"`
	Content            *string `json:"content,omitempty"`
	AcceptanceCriteria *string `json:"acceptance_criteria,omitempty"`
}

// BatchDeleteStoriesRequest 批量删除 Stories 请求
type BatchDeleteStoriesRequest struct {
	StoryIDs []string `json:"story_ids" binding:"required,min=1"`
}

// ConfirmEpicsAndStoriesRequest 确认 Epics 和 Stories 请求
type ConfirmEpicsAndStoriesRequest struct {
	Action string `json:"action" binding:"required,oneof=confirm skip regenerate"` // confirm: 确认并继续, skip: 跳过确认, regenerate: 重新生成
}
