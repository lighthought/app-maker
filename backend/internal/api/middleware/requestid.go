package middleware

import (
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
)

// RequestID 请求ID中间件
func RequestID() gin.HandlerFunc {
	return requestid.New()
}
