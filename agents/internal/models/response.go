package models

// TODO: 响应模型
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

// 项目环境准备响应
type SetupProjEnvRes struct {
	BmadMethodStatus string `json:"bmad_method_status" example:"success"`
	FrontendStatus   string `json:"frontend_status" example:"success"`
	BackendStatus    string `json:"backend_status" example:"success"`
}
