package middleware

import (
	"net/http"
	"shared-models/auth"
	"shared-models/common"
	"shared-models/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware 认证中间件
func AuthMiddleware(jwtService *auth.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取Authorization头
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, utils.GetErrorResponse(common.UNAUTHORIZED, "Authorization header required"))
			c.Abort()
			return
		}

		// 解析Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != common.TokenHeaderPrefix {
			c.JSON(http.StatusUnauthorized, utils.GetErrorResponse(common.UNAUTHORIZED, "Invalid authorization format"))
			c.Abort()
			return
		}

		token := parts[1]

		// 验证JWT token
		claims, err := jwtService.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, utils.GetErrorResponse(common.UNAUTHORIZED, common.MESSAGE_INVALID_TOKEN))
			c.Abort()
			return
		}

		// 将用户信息存储到上下文中
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_username", claims.Username)

		c.Next()
	}
}
