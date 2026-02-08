package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"go-shop/pkg/logger"
	"go-shop/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RecoveryMiddleware 错误恢复中间件
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 记录panic信息
				logger.Error("Panic recovered",
					zap.Any("error", err),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
					zap.String("stack", string(debug.Stack())),
				)

				// 返回错误响应
				response.Error(c, http.StatusInternalServerError, fmt.Sprintf("服务器内部错误: %v", err))
				c.Abort()
			}
		}()

		c.Next()
	}
}
