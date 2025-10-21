package utils

import (
	"github.com/lighthought/app-maker/shared-models/common"
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

func GetPaginationResponse(total int, page int, pageSize int, data interface{}) *common.PaginationResponse {
	totalPages := (total + pageSize - 1) / pageSize
	hasNext := page < totalPages
	hasPrevious := page > 1

	return &common.PaginationResponse{
		Code:        common.SUCCESS_CODE,
		Message:     "success",
		Total:       total,
		Page:        page,
		PageSize:    pageSize,
		TotalPages:  totalPages,
		Data:        data,
		HasNext:     hasNext,
		HasPrevious: hasPrevious,
		Timestamp:   GetCurrentTime(),
	}
}
