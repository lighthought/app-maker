package middleware

import (
	"autocodeweb-backend/internal/config"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORS CORS中间件
func CORS(cfg config.CORSConfig) gin.HandlerFunc {
	corsConfig := cors.Config{
		AllowOrigins:     cfg.AllowedOrigins,
		AllowMethods:     cfg.AllowedMethods,
		AllowHeaders:     cfg.AllowedHeaders,
		AllowCredentials: cfg.AllowCredentials,
		MaxAge:           time.Duration(cfg.MaxAge) * time.Hour,
	}

	// 返回组合的中间件
	return func(c *gin.Context) {
		// 应用 CORS 配置
		corsHandler := cors.New(corsConfig)
		corsHandler(c)

		// 允许在 iframe 中嵌入（移除 X-Frame-Options 限制）
		// 允许所有同源和子域的 iframe 嵌入
		c.Header("X-Frame-Options", "SAMEORIGIN")
		// 或者完全允许所有来源嵌入（更宽松）
		// c.Writer.Header().Del("X-Frame-Options")

		// 设置 Content-Security-Policy 允许 iframe
		// 注意：这里允许所有来源的 iframe，生产环境应该限制特定域名
		c.Header("Content-Security-Policy", "frame-ancestors 'self' *.app-maker.localhost;")

		// 处理预检请求
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
