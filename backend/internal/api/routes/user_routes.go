package routes

import (
	"autocodeweb-backend/internal/api/handlers"

	"github.com/gin-gonic/gin"
)

// RegisterUserRoutes 注册用户相关路由
func RegisterUserRoutes(router *gin.RouterGroup, userHandler *handlers.UserHandler, authMiddleware gin.HandlerFunc) {
	// 认证相关路由（无需认证）
	auth := router.Group("/auth")
	{
		auth.POST("/register", userHandler.Register)
		auth.POST("/login", userHandler.Login)
		auth.POST("/refresh", userHandler.RefreshToken)
	}

	// 用户相关路由（需要认证）
	users := router.Group("/users")
	users.Use(authMiddleware) // 应用认证中间件
	{
		// 用户档案管理
		users.GET("/profile", userHandler.GetUserProfile)
		users.PUT("/profile", userHandler.UpdateUserProfile)
		users.POST("/change-password", userHandler.ChangePassword)
		users.POST("/logout", userHandler.Logout)

		// 管理员功能
		users.GET("", userHandler.GetUserList)
		users.DELETE("/:user_id", userHandler.DeleteUser)
	}
}
