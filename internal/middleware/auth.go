package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// APIKeyAuth는 API 키 기반 인증 미들웨어입니다
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
		
		// 컨텍스트에 API 키 저장
		c.Set("api_key", apiKey)
		c.Next()
	})
}

// isValidAPIKey는 API 키의 유효성을 검사합니다
func isValidAPIKey(provided, valid string) bool {
	// 기본적인 문자열 비교
	// 실제 환경에서는 해시 비교나 데이터베이스 조회 등을 사용
	return strings.TrimSpace(provided) == strings.TrimSpace(valid)
}

// GetAPIKeyFromContext는 컨텍스트에서 API 키를 추출합니다
func GetAPIKeyFromContext(c *gin.Context) string {
	if apiKey, exists := c.Get("api_key"); exists {
		if key, ok := apiKey.(string); ok {
			return key
		}
	}
	return ""
}