package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	pkgErrors "go-shop/pkg/errors"
	"go-shop/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
	logger.InitLogger("info", "")
}

func TestErrorHandlerMiddleware_BusinessError(t *testing.T) {
	router := gin.New()
	router.Use(ErrorHandlerMiddleware())

	router.GET("/test", func(c *gin.Context) {
		err := pkgErrors.New(pkgErrors.CodeUserNotFound, "")
		c.Error(err)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "用户不存在")
}

func TestErrorHandlerMiddleware_GenericError(t *testing.T) {
	router := gin.New()
	router.Use(ErrorHandlerMiddleware())

	router.GET("/test", func(c *gin.Context) {
		err := errors.New("generic error")
		c.Error(err)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "服务器内部错误")
}

func TestErrorHandlerMiddleware_NoError(t *testing.T) {
	router := gin.New()
	router.Use(ErrorHandlerMiddleware())

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "success")
}

func TestAbortWithError(t *testing.T) {
	router := gin.New()
	router.Use(ErrorHandlerMiddleware())

	router.GET("/test", func(c *gin.Context) {
		err := pkgErrors.New(pkgErrors.CodeUnauthorized, "")
		AbortWithError(c, err)
		// 这行不应该执行
		c.JSON(http.StatusOK, gin.H{"message": "should not reach here"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.NotContains(t, w.Body.String(), "should not reach here")
}

func TestAbortWithBusinessError(t *testing.T) {
	router := gin.New()
	router.Use(ErrorHandlerMiddleware())

	router.GET("/test", func(c *gin.Context) {
		AbortWithBusinessError(c, pkgErrors.CodeForbidden, "自定义禁止访问消息")
		// 这行不应该执行
		c.JSON(http.StatusOK, gin.H{"message": "should not reach here"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "自定义禁止访问消息")
	assert.NotContains(t, w.Body.String(), "should not reach here")
}

func TestErrorHandlerMiddleware_MultipleErrors(t *testing.T) {
	router := gin.New()
	router.Use(ErrorHandlerMiddleware())

	router.GET("/test", func(c *gin.Context) {
		// 添加多个错误，应该只处理最后一个
		c.Error(errors.New("first error"))
		c.Error(pkgErrors.New(pkgErrors.CodeProductNotFound, ""))
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "商品不存在")
}
