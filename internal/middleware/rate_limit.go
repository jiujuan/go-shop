package middleware

import (
	"sync"
	"time"

	"go-shop/pkg/response"

	"github.com/gin-gonic/gin"
)

// RateLimiter 限流器
type RateLimiter struct {
	rate     int           // 每秒允许的请求数
	capacity int           // 令牌桶容量
	tokens   int           // 当前令牌数
	lastTime time.Time     // 上次更新时间
	mu       sync.Mutex    // 互斥锁
}

// NewRateLimiter 创建限流器
func NewRateLimiter(rate, capacity int) *RateLimiter {
	return &RateLimiter{
		rate:     rate,
		capacity: capacity,
		tokens:   capacity,
		lastTime: time.Now(),
	}
}

// Allow 判断是否允许请求
func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.lastTime).Seconds()

	// 添加新令牌
	rl.tokens += int(elapsed * float64(rl.rate))
	if rl.tokens > rl.capacity {
		rl.tokens = rl.capacity
	}

	rl.lastTime = now

	// 消耗令牌
	if rl.tokens > 0 {
		rl.tokens--
		return true
	}

	return false
}

// RateLimitMiddleware 限流中间件
func RateLimitMiddleware(rate, capacity int) gin.HandlerFunc {
	limiter := NewRateLimiter(rate, capacity)

	return func(c *gin.Context) {
		if !limiter.Allow() {
			response.Error(c, response.CodeTooManyRequests, "请求过于频繁，请稍后再试")
			c.Abort()
			return
		}

		c.Next()
	}
}
