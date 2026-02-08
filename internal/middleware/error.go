package middleware

import (
	"errors"
	"net/http"

	"go-shop/pkg/logger"
	pkgErrors "go-shop/pkg/errors"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ErrorResponse 错误响应结构
type ErrorResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ErrorHandlerMiddleware 错误处理中间件
func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 检查是否有错误
		if len(c.Errors) == 0 {
			return
		}

		// 获取最后一个错误
		err := c.Errors.Last().Err

		// 处理业务错误
		var bizErr *pkgErrors.BusinessError
		if errors.As(err, &bizErr) {
			handleBusinessError(c, bizErr)
			return
		}

		// 处理其他错误
		handleGenericError(c, err)
	}
}

// handleBusinessError 处理业务错误
func handleBusinessError(c *gin.Context, err *pkgErrors.BusinessError) {
	httpStatus := pkgErrors.GetHTTPStatus(err.Code)

	// 记录错误日志
	if httpStatus >= 500 {
		logger.Error("Business error",
			zap.Int("code", int(err.Code)),
			zap.String("message", err.Message),
			zap.Error(err.Err),
			zap.String("path", c.Request.URL.Path),
			zap.String("method", c.Request.Method),
		)
	} else {
		logger.Warn("Business error",
			zap.Int("code", int(err.Code)),
			zap.String("message", err.Message),
			zap.String("path", c.Request.URL.Path),
			zap.String("method", c.Request.Method),
		)
	}

	// 返回错误响应
	c.JSON(httpStatus, ErrorResponse{
		Code:    int(err.Code),
		Message: err.Message,
	})
}

// handleGenericError 处理通用错误
func handleGenericError(c *gin.Context, err error) {
	// 记录错误日志
	logger.Error("Unhandled error",
		zap.Error(err),
		zap.String("path", c.Request.URL.Path),
		zap.String("method", c.Request.Method),
	)

	// 返回通用错误响应
	c.JSON(http.StatusInternalServerError, ErrorResponse{
		Code:    int(pkgErrors.CodeInternalError),
		Message: "服务器内部错误",
	})
}

// AbortWithError 中止请求并返回错误
func AbortWithError(c *gin.Context, err error) {
	c.Error(err)
	c.Abort()
}

// AbortWithBusinessError 中止请求并返回业务错误
func AbortWithBusinessError(c *gin.Context, code pkgErrors.ErrorCode, message string) {
	err := pkgErrors.New(code, message)
	c.Error(err)
	c.Abort()
}
