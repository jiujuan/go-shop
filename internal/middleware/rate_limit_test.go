package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRateLimitMiddleware(t *testing.T) {
	router := setupTestRouter()
	router.Use(RateLimitMiddleware(2, 2)) // 每秒2个请求，容量2
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	// 前两个请求应该成功
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	// 第三个请求应该被限流
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusTooManyRequests, w.Code)
}

func TestRateLimiter_Allow(t *testing.T) {
	limiter := NewRateLimiter(10, 10)

	// 消耗所有令牌
	for i := 0; i < 10; i++ {
		assert.True(t, limiter.Allow())
	}

	// 令牌耗尽
	assert.False(t, limiter.Allow())

	// 等待令牌恢复
	time.Sleep(200 * time.Millisecond)
	assert.True(t, limiter.Allow())
}
