package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func APIKeyAuth(validAPIKey string) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		
		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "API key is required",
			})
			c.Abort()
			return
		}
		
		// API 키 검증 (실제 환경에서는 데이터베이스나 더 복잡한 검증 로직 사용)
		if !isValidAPIKey(apiKey, validAPIKey) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Invalid API key",
			})
			c.Abort()
			return
		}
		
		c.Set("api_key", apiKey)
		c.Next()
	})
}

func isValidAPIKey(provided, valid string) bool {
	return strings.TrimSpace(provided) == strings.TrimSpace(valid)
}

func GetAPIKeyFromContext(c *gin.Context) string {
	if apiKey, exists := c.Get("api_key"); exists {
		if key, ok := apiKey.(string); ok {
			return key
		}
	}
	return ""
}