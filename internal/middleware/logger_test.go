package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go-shop/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestLoggerMiddleware(t *testing.T) {
	// 初始化日志
	logger.InitLogger("info", "")

	router := setupTestRouter()
	router.Use(LoggerMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test?param=value", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
