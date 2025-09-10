package models

import "time"

// Response 通用响应结构
type Response struct {
	Code      int         `json:"code" example:"0"`
	Message   string      `json:"message" example:"success"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp string      `json:"timestamp" example:"2025-08-29T10:00:00Z"`
}

const (
	SUCCESS_CODE          = 0    // 成功
	ERROR_CODE            = 1    // 错误
	VALIDATION_ERROR      = 400  // 请求参数验证失败
	UNAUTHORIZED          = 401  // 未认证或认证失败
	FORBIDDEN             = 403  // 权限不足
	NOT_FOUND             = 404  // 资源不存在
	CONFLICT              = 409  // 资源冲突
	RATE_LIMIT_EXCEEDED   = 429  // 请求频率超限
	INTERNAL_ERROR        = 500  // 服务器内部错误
	SERVICE_UNAVAILABLE   = 503  // 服务不可用
	PROJECT_NOT_FOUND     = 2404 // 项目不存在
	PROJECT_ACCESS_DENIED = 2403 // 项目访问权限不足
	AGENT_SESSION_EXPIRED = 2410 // Agent会话已过期
	TASK_INTERNAL_ERROR   = 2500 // 任务内部错误
	DEPLOYMENT_ERROR      = 2501 // 部署错误
	INSUFFICIENT_QUOTA    = 2429 // 配额不足
)

// ErrorResponse 错误响应结构
type ErrorResponse struct {
	Code      int    `json:"code" example:"400"`
	Message   string `json:"message" example:"请求参数错误"`
	Timestamp string `json:"timestamp" example:"2025-08-29T10:00:00Z"`
}

// PaginationResponse 分页响应
type PaginationResponse struct {
	Code        int         `json:"code" example:"0"`
	Message     string      `json:"message" example:"success"`
	Total       int         `json:"total" example:"100"`
	Page        int         `json:"page" example:"1"`
	PageSize    int         `json:"page_size" example:"10"`
	TotalPages  int         `json:"total_pages" example:"10"`
	Data        interface{} `json:"data"`
	HasNext     bool        `json:"has_next" example:"true"`
	HasPrevious bool        `json:"has_previous" example:"false"`
	Timestamp   string      `json:"timestamp" example:"2025-08-29T10:00:00Z"`
}

// UserInfo 用户信息（用于响应）
type UserInfo struct {
	ID        string    `json:"id" example:"varchar(50)"`
	Email     string    `json:"email" example:"user@example.com"`
	Username  string    `json:"username" example:"username"`
	Role      string    `json:"role" example:"user"`
	Status    string    `json:"status" example:"active"`
	CreatedAt time.Time `json:"created_at"`
}

// ProjectInfo 项目信息（用于响应）
type ProjectInfo struct {
	ID           string    `json:"id" example:"PROJ_00000000001"`
	Name         string    `json:"name" example:"项目名称"`
	Description  string    `json:"description" example:"项目描述"`
	Status       string    `json:"status" example:"in_progress"`
	Requirements string    `json:"requirements" example:"项目需求"`
	ProjectPath  string    `json:"project_path" example:"/path/to/project"`
	BackendPort  int       `json:"backend_port" example:"8080"`
	FrontendPort int       `json:"frontend_port" example:"3000"`
	UserID       string    `json:"user_id" example:"USER_00000000001"`
	User         UserInfo  `json:"user,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	User         UserInfo `json:"user"`
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	ExpiresIn    int64    `json:"expires_in" example:"3600"`
}
