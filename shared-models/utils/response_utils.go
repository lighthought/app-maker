package utils

import (
	"shared-models/common"
)

// GetSuccessResponse 获取成功响应
func GetSuccessResponse(message string, data interface{}) common.Response {
	return common.Response{
		Code:      common.SUCCESS_CODE,
		Message:   message,
		Data:      data,
		Timestamp: GetCurrentTime(),
	}
}

// GetErrorResponse 获取错误响应
func GetErrorResponse(code int, message string) common.ErrorResponse {
	return common.ErrorResponse{
		Code:      common.ERROR_CODE,
		Message:   message,
		Timestamp: GetCurrentTime(),
	}
}
