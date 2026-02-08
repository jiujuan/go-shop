package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go-shop/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRecoveryMiddleware(t *testing.T) {
	// 初始化日志
	logger.InitLogger("info", "")

	router := setupTestRouter()
	router.Use(RecoveryMiddleware())
	router.GET("/panic", func(c *gin.Context) {
		panic("test panic")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestRecoveryMiddleware_NoPanic(t *testing.T) {
	logger.InitLogger("info", "")

	router := setupTestRouter()
	router.Use(RecoveryMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
