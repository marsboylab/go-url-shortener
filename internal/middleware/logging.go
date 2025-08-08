package middleware

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func RequestLogger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("[%s] \"%s %s %s\" %d %s \"%s\" \"%s\" %s\n",
			param.TimeStamp.Format("2006/01/02 - 15:04:05"),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.Request.Referer(),
			param.ClientIP,
		)
	})
}

func AccessLogger() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		
		c.Next()
		
		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		
		if raw != "" {
			path = path + "?" + raw
		}
		
		// 에러가 있는 경우 별도 로깅
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				log.Printf("Error: %v", err.Error())
			}
		}
		
		// API 키 정보 (마스킹)
		apiKey := c.GetHeader("X-API-Key")
		maskedAPIKey := ""
		if apiKey != "" {
			if len(apiKey) > 8 {
				maskedAPIKey = apiKey[:4] + "****" + apiKey[len(apiKey)-4:]
			} else {
				maskedAPIKey = "****"
			}
		}
		
		log.Printf("[ACCESS] %s %s %d %v %s %s",
			method,
			path,
			statusCode,
			latency,
			clientIP,
			maskedAPIKey,
		)
	})
}

func JSONBinding() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Content-Type 확인
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
			contentType := c.GetHeader("Content-Type")
			if contentType != "" && !strings.Contains(contentType, "application/json") {
				c.JSON(400, gin.H{
					"error":   "invalid_content_type",
					"message": "Content-Type must be application/json",
				})
				c.Abort()
				return
			}
		}
		
		c.Next()
	})
}