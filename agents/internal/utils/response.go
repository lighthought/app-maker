package utils

// Success 返回成功响应
func Success(c interface{}, data interface{}) {
	// 这里应该是gin.Context，但为了避免循环依赖，使用interface{}
}

// Error 返回错误响应
func Error(c interface{}, status int, message string) {
	// 这里应该是gin.Context，但为了避免循环依赖，使用interface{}
}
