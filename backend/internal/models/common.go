package models

import (
	"time"
)

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

// PaginationRequest 分页请求
type PaginationRequest struct {
	Page     int `json:"page" form:"page" binding:"min=1" example:"1"`
	PageSize int `json:"page_size" form:"page_size" binding:"min=1,max=100" example:"10"`
}

// PaginationResponse 分页响应
type PaginationResponse struct {
	Total       int         `json:"total" example:"100"`
	Page        int         `json:"page" example:"1"`
	PageSize    int         `json:"page_size" example:"10"`
	TotalPages  int         `json:"total_pages" example:"10"`
	Data        interface{} `json:"data"`
	HasNext     bool        `json:"has_next" example:"true"`
	HasPrevious bool        `json:"has_previous" example:"false"`
}

// BaseModel 基础模型
type BaseModel struct {
	ID        string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt time.Time `json:"deleted_at,omitempty"`
}

// UserInfo 用户信息（用于响应）
type UserInfo struct {
	ID        string    `json:"id" example:"uuid"`
	Email     string    `json:"email" example:"user@example.com"`
	Username  string    `json:"username" example:"username"`
	Role      string    `json:"role" example:"user"`
	Status    string    `json:"status" example:"active"`
	CreatedAt time.Time `json:"created_at"`
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

// LoginResponse 登录响应
type LoginResponse struct {
	User         UserInfo `json:"user"`
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	ExpiresIn    int64    `json:"expires_in" example:"3600"`
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

// CreateProjectRequest 创建项目请求
type CreateProjectRequest struct {
	Name         string   `json:"name" binding:"required,min=1,max=100" example:"我的项目"`
	Description  string   `json:"description" binding:"omitempty,max=500" example:"项目描述"`
	Requirements string   `json:"requirements" binding:"required" example:"项目需求描述"`
	BackendPort  int      `json:"backend_port" binding:"omitempty,min=1024,max=65535" example:"8080"`
	FrontendPort int      `json:"frontend_port" binding:"omitempty,min=1024,max=65535" example:"3000"`
	TagIDs       []string `json:"tag_ids,omitempty" example:"['tag1', 'tag2']"`
}

// UpdateProjectRequest 更新项目请求
type UpdateProjectRequest struct {
	Name         string   `json:"name" binding:"omitempty,min=1,max=100" example:"更新后的项目名"`
	Description  string   `json:"description" binding:"omitempty,max=500" example:"更新后的项目描述"`
	Requirements string   `json:"requirements" binding:"omitempty" example:"更新后的项目需求"`
	BackendPort  int      `json:"backend_port" binding:"omitempty,min=1024,max=65535" example:"8080"`
	FrontendPort int      `json:"frontend_port" binding:"omitempty,min=1024,max=65535" example:"3000"`
	Status       string   `json:"status" binding:"omitempty,oneof=draft in_progress completed failed" example:"in_progress"`
	TagIDs       []string `json:"tag_ids,omitempty" example:"['tag1', 'tag2']"`
}

// ProjectListRequest 项目列表请求
type ProjectListRequest struct {
	PaginationRequest
	Status string   `json:"status" form:"status" binding:"omitempty,oneof=draft in_progress completed failed" example:"in_progress"`
	TagIDs []string `json:"tag_ids" form:"tag_ids" example:"['tag1', 'tag2']"`
	UserID string   `json:"user_id" form:"user_id" example:"uuid"`
	Search string   `json:"search" form:"search" example:"项目名称关键词"`
}

// ProjectInfo 项目信息（用于响应）
type ProjectInfo struct {
	ID           string    `json:"id" example:"uuid"`
	Name         string    `json:"name" example:"项目名称"`
	Description  string    `json:"description" example:"项目描述"`
	Status       string    `json:"status" example:"in_progress"`
	Requirements string    `json:"requirements" example:"项目需求"`
	ProjectPath  string    `json:"project_path" example:"/path/to/project"`
	BackendPort  int       `json:"backend_port" example:"8080"`
	FrontendPort int       `json:"frontend_port" example:"3000"`
	UserID       string    `json:"user_id" example:"uuid"`
	User         UserInfo  `json:"user,omitempty"`
	Tags         []TagInfo `json:"tags,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TagInfo 标签信息（用于响应）
type TagInfo struct {
	ID    string `json:"id" example:"uuid"`
	Name  string `json:"name" example:"标签名称"`
	Color string `json:"color" example:"#666666"`
}

// CreateTagRequest 创建标签请求
type CreateTagRequest struct {
	Name  string `json:"name" binding:"required,min=1,max=50" example:"标签名称"`
	Color string `json:"color" binding:"omitempty,hexcolor" example:"#666666"`
}

// UpdateTagRequest 更新标签请求
type UpdateTagRequest struct {
	Name  string `json:"name" binding:"omitempty,min=1,max=50" example:"更新后的标签名称"`
	Color string `json:"color" binding:"omitempty,hexcolor" example:"#FF0000"`
}

// TaskInfo 任务信息（用于响应）
type TaskInfo struct {
	ID           string     `json:"id" example:"uuid"`
	ProjectID    string     `json:"project_id" example:"uuid"`
	UserID       string     `json:"user_id" example:"uuid"`
	Name         string     `json:"name" example:"任务名称"`
	Description  string     `json:"description" example:"任务描述"`
	Status       string     `json:"status" example:"pending"`
	Priority     int        `json:"priority" example:"2"`
	Dependencies []string   `json:"dependencies" example:"['task1', 'task2']"`
	MaxRetries   int        `json:"max_retries" example:"3"`
	RetryCount   int        `json:"retry_count" example:"0"`
	RetryDelay   int        `json:"retry_delay" example:"60"`
	StartedAt    *time.Time `json:"started_at"`
	CompletedAt  *time.Time `json:"completed_at"`
	Deadline     *time.Time `json:"deadline"`
	Result       string     `json:"result" example:"执行结果"`
	ErrorMessage string     `json:"error_message" example:"错误信息"`
	Metadata     string     `json:"metadata" example:"元数据"`
	Tags         []string   `json:"tags" example:"['tag1', 'tag2']"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// CreateTaskRequest 创建任务请求
type CreateTaskRequest struct {
	ProjectID    string     `json:"project_id" binding:"required" example:"uuid"`
	Name         string     `json:"name" binding:"required,min=1,max=100" example:"任务名称"`
	Description  string     `json:"description" binding:"omitempty,max=500" example:"任务描述"`
	Priority     int        `json:"priority" binding:"omitempty,min=1,max=4" example:"2"`
	Dependencies []string   `json:"dependencies" example:"['task1', 'task2']"`
	MaxRetries   int        `json:"max_retries" binding:"omitempty,min=0,max=10" example:"3"`
	RetryDelay   int        `json:"retry_delay" binding:"omitempty,min=0" example:"60"`
	Deadline     *time.Time `json:"deadline"`
	Metadata     string     `json:"metadata" example:"元数据"`
	Tags         []string   `json:"tags" example:"['tag1', 'tag2']"`
}

// UpdateTaskRequest 更新任务请求
type UpdateTaskRequest struct {
	Name         string     `json:"name" binding:"omitempty,min=1,max=100" example:"更新后的任务名"`
	Description  string     `json:"description" binding:"omitempty,max=500" example:"更新后的任务描述"`
	Priority     int        `json:"priority" binding:"omitempty,min=1,max=4" example:"3"`
	Dependencies []string   `json:"dependencies" example:"['task1', 'task2']"`
	MaxRetries   int        `json:"max_retries" binding:"omitempty,min=0,max=10" example:"5"`
	RetryDelay   int        `json:"retry_delay" binding:"omitempty,min=0" example:"120"`
	Deadline     *time.Time `json:"deadline"`
	Metadata     string     `json:"metadata" example:"元数据"`
	Tags         []string   `json:"tags" example:"['tag1', 'tag2']"`
}

// TaskListRequest 任务列表请求
type TaskListRequest struct {
	PaginationRequest
	ProjectID string   `json:"project_id" form:"project_id" example:"uuid"`
	UserID    string   `json:"user_id" form:"user_id" example:"uuid"`
	Status    string   `json:"status" form:"status" example:"pending"`
	Priority  int      `json:"priority" form:"priority" example:"2"`
	Tags      []string `json:"tags" form:"tags" example:"['tag1', 'tag2']"`
	Search    string   `json:"search" form:"search" example:"任务名称关键词"`
}

// TaskLogInfo 任务日志信息（用于响应）
type TaskLogInfo struct {
	ID        string    `json:"id" example:"uuid"`
	TaskID    string    `json:"task_id" example:"uuid"`
	Level     string    `json:"level" example:"info"`
	Message   string    `json:"message" example:"日志消息"`
	Data      string    `json:"data" example:"额外数据"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateTaskLogRequest 创建任务日志请求
type CreateTaskLogRequest struct {
	TaskID  string `json:"task_id" binding:"required" example:"uuid"`
	Level   string `json:"level" binding:"required,oneof=info warn error" example:"info"`
	Message string `json:"message" binding:"required" example:"日志消息"`
	Data    string `json:"data" example:"额外数据"`
}

// TaskStatusUpdateRequest 任务状态更新请求
type TaskStatusUpdateRequest struct {
	Status       string `json:"status" binding:"required,oneof=pending running completed failed cancelled retrying rolled_back" example:"running"`
	ErrorMessage string `json:"error_message" example:"错误信息"`
	Result       string `json:"result" example:"执行结果"`
}

// TaskRetryRequest 任务重试请求
type TaskRetryRequest struct {
	Force bool `json:"force" example:"false"` // 强制重试，忽略依赖检查
}

// TaskRollbackRequest 任务回滚请求
type TaskRollbackRequest struct {
	Reason string `json:"reason" binding:"required" example:"回滚原因"`
}
