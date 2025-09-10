package middleware

import (
	"net/http"
	"strings"

	"autocodeweb-backend/internal/models"
	"autocodeweb-backend/internal/utils"
	"autocodeweb-backend/pkg/auth"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware 认证中间件
func AuthMiddleware(jwtService *auth.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取Authorization头
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Code:      models.UNAUTHORIZED,
				Message:   "Authorization header required",
				Timestamp: utils.GetCurrentTime(),
			})
			c.Abort()
			return
		}

		// 解析Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Code:      models.UNAUTHORIZED,
				Message:   "Invalid authorization format",
				Timestamp: utils.GetCurrentTime(),
			})
			c.Abort()
			return
		}

		token := parts[1]

		// 验证JWT token
		claims, err := jwtService.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Code:      models.UNAUTHORIZED,
				Message:   "Invalid token",
				Timestamp: utils.GetCurrentTime(),
			})
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
