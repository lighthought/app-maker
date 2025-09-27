package utils

// SuccessResponse 成功响应结构
type SuccessResponse struct {
	Success bool        `json:"success" example:"true"`
	Data    interface{} `json:"data"`
	Message string      `json:"message,omitempty" example:"操作成功"`
}

// ErrorResponse 错误响应结构
type ErrorResponse struct {
	Success bool   `json:"success" example:"false"`
	Error   string `json:"error" example:"错误信息"`
	Message string `json:"message,omitempty" example:"操作失败"`
}

// Success 返回成功响应
func Success(c interface{}, data interface{}) {
	// 这里应该是gin.Context，但为了避免循环依赖，使用interface{}
}

// Error 返回错误响应
func Error(c interface{}, status int, message string) {
	// 这里应该是gin.Context，但为了避免循环依赖，使用interface{}
}
