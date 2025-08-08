package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type RateLimiter struct {
	requests map[string][]time.Time
	mutex    sync.RWMutex
	limit    int
	window   time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
	
	// 주기적으로 오래된 요청 기록 정리
	go rl.cleanup()
	
	return rl
}

func (rl *RateLimiter) Allow(key string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	
	now := time.Now()
	cutoff := now.Add(-rl.window)
	
	// 해당 키의 요청 기록 가져오기
	if rl.requests[key] == nil {
		rl.requests[key] = make([]time.Time, 0)
	}
	
	// 윈도우 밖의 오래된 요청 제거
	requests := rl.requests[key]
	validRequests := make([]time.Time, 0, len(requests))
	
	for _, requestTime := range requests {
		if requestTime.After(cutoff) {
			validRequests = append(validRequests, requestTime)
		}
	}
	
	// 현재 요청이 제한을 초과하는지 확인
	if len(validRequests) >= rl.limit {
		rl.requests[key] = validRequests
		return false
	}
	
	// 현재 요청 추가
	validRequests = append(validRequests, now)
	rl.requests[key] = validRequests
	
	return true
}

// cleanup은 주기적으로 오래된 요청 기록을 정리합니다
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		rl.mutex.Lock()
		now := time.Now()
		cutoff := now.Add(-rl.window * 2) // 윈도우의 2배 시간 이전 기록 삭제
		
		for key, requests := range rl.requests {
			validRequests := make([]time.Time, 0, len(requests))
			for _, requestTime := range requests {
				if requestTime.After(cutoff) {
					validRequests = append(validRequests, requestTime)
				}
			}
			
			if len(validRequests) == 0 {
				delete(rl.requests, key)
			} else {
				rl.requests[key] = validRequests
			}
		}
		rl.mutex.Unlock()
	}
}

// 전역 속도 제한기 인스턴스
var globalRateLimiter = NewRateLimiter(60, time.Minute) // 분당 60회

// RateLimit는 속도 제한 미들웨어를 제공합니다
func RateLimit() gin.HandlerFunc {
	return RateLimitWithLimiter(globalRateLimiter)
}

// RateLimitWithLimiter는 커스텀 속도 제한기를 사용하는 미들웨어를 제공합니다
func RateLimitWithLimiter(limiter *RateLimiter) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// 클라이언트 식별자 생성 (IP + User-Agent 조합)
		clientID := getClientID(c)
		
		if !limiter.Allow(clientID) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "rate_limit_exceeded",
				"message": fmt.Sprintf("Rate limit exceeded: %d requests per %v", limiter.limit, limiter.window),
				"details": gin.H{
					"limit":  limiter.limit,
					"window": limiter.window.String(),
				},
			})
			c.Abort()
			return
		}
		
		c.Next()
	})
}

func getClientID(c *gin.Context) string {
	// X-Forwarded-For 헤더에서 실제 IP 추출
	clientIP := c.ClientIP()
	
	// API 키가 있으면 API 키 기반으로 식별
	if apiKey := c.GetHeader("X-API-Key"); apiKey != "" {
		return fmt.Sprintf("api:%s", apiKey)
	}
	
	// 그렇지 않으면 IP 기반으로 식별
	return fmt.Sprintf("ip:%s", clientIP)
}

// CustomRateLimit는 커스텀 제한으로 속도 제한 미들웨어를 생성합니다
func CustomRateLimit(limit int, window time.Duration) gin.HandlerFunc {
	limiter := NewRateLimiter(limit, window)
	return RateLimitWithLimiter(limiter)
}