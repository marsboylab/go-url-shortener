package middleware

import (
	"github.com/gin-gonic/gin"
)

func CORS() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		
		allowedOrigins := map[string]bool{
			"http://localhost:3000":     true,
			"http://localhost:8080":     true,
			"https://marsboy.dev":       true,
			"https://admin.marsboy.dev": true,
		}
		
		// 개발 환경에서는 모든 도메인 허용
		if gin.Mode() == gin.DebugMode {
			c.Header("Access-Control-Allow-Origin", "*")
		} else if allowedOrigins[origin] {
			c.Header("Access-Control-Allow-Origin", origin)
		}
		
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-API-Key")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400") // 24시간
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	})
}