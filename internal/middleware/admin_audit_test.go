package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-shop/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
	logger.InitLogger("info", "")
}

func TestAdminAuditMiddleware(t *testing.T) {
	router := gin.New()
	
	// 模拟认证中间件设置用户信息
	router.Use(func(c *gin.Context) {
		c.Set(ContextKeyUserID, int64(1))
		c.Set(ContextKeyUsername, "admin")
		c.Set(ContextKeyIsAdmin, true)
		c.Next()
	})
	
	router.Use(AdminAuditMiddleware())

	router.POST("/admin/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	reqBody := map[string]interface{}{
		"name": "test",
		"password": "secret123",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/admin/test?param=value", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "test-agent")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAdminAuditMiddleware_GET(t *testing.T) {
	router := gin.New()
	
	router.Use(func(c *gin.Context) {
		c.Set(ContextKeyUserID, int64(1))
		c.Set(ContextKeyUsername, "admin")
		c.Next()
	})
	
	router.Use(AdminAuditMiddleware())

	router.GET("/admin/users", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"users": []string{"user1", "user2"}})
	})

	req := httptest.NewRequest(http.MethodGet, "/admin/users?page=1&size=10", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAdminAuditMiddleware_WithError(t *testing.T) {
	router := gin.New()
	
	router.Use(func(c *gin.Context) {
		c.Set(ContextKeyUserID, int64(1))
		c.Set(ContextKeyUsername, "admin")
		c.Next()
	})
	
	router.Use(AdminAuditMiddleware())

	router.DELETE("/admin/user/:id", func(c *gin.Context) {
		c.Error(assert.AnError)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"})
	})

	req := httptest.NewRequest(http.MethodDelete, "/admin/user/123", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestSanitizeRequestBody(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name: "Sanitize password",
			input: map[string]interface{}{
				"username": "admin",
				"password": "secret123",
			},
			expected: map[string]interface{}{
				"username": "admin",
				"password": "***REDACTED***",
			},
		},
		{
			name: "Sanitize multiple sensitive fields",
			input: map[string]interface{}{
				"name":          "test",
				"password":      "secret",
				"token":         "abc123",
				"api_key":       "key123",
			},
			expected: map[string]interface{}{
				"name":          "test",
				"password":      "***REDACTED***",
				"token":         "***REDACTED***",
				"api_key":       "***REDACTED***",
			},
		},
		{
			name: "No sensitive fields",
			input: map[string]interface{}{
				"name":  "test",
				"email": "test@example.com",
			},
			expected: map[string]interface{}{
				"name":  "test",
				"email": "test@example.com",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sanitizeRequestBody(tt.input)
			assert.Equal(t, tt.expected, tt.input)
		})
	}
}

func TestLogAdminAction(t *testing.T) {
	router := gin.New()
	
	router.Use(func(c *gin.Context) {
		c.Set(ContextKeyUserID, int64(1))
		c.Set(ContextKeyUsername, "admin")
		c.Next()
	})

	router.POST("/admin/product", func(c *gin.Context) {
		LogAdminAction(c, "CREATE_PRODUCT", "product:123", map[string]interface{}{
			"name":  "New Product",
			"price": 99.99,
		})
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest(http.MethodPost, "/admin/product", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAdminMiddleware_Enhanced(t *testing.T) {
	tests := []struct {
		name           string
		isAdmin        bool
		expectedStatus int
	}{
		{
			name:           "Admin user allowed",
			isAdmin:        true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Non-admin user forbidden",
			isAdmin:        false,
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			
			// 模拟认证
			router.Use(func(c *gin.Context) {
				c.Set(ContextKeyUserID, int64(1))
				c.Set(ContextKeyUsername, "testuser")
				c.Set(ContextKeyIsAdmin, tt.isAdmin)
				c.Next()
			})
			
			router.Use(AdminMiddleware())

			router.GET("/admin/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			req := httptest.NewRequest(http.MethodGet, "/admin/test", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
