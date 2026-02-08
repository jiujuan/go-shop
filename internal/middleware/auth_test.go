package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go-shop/pkg/auth"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestAuthMiddleware_Success(t *testing.T) {
	// 创建JWT管理器
	jwtManager := auth.NewJWTManager("test-secret-key", 24)

	// 生成有效Token
	token, err := jwtManager.GenerateToken(1, "testuser", false)
	assert.NoError(t, err)

	// 设置路由
	router := setupTestRouter()
	router.GET("/test", AuthMiddleware(jwtManager), func(c *gin.Context) {
		userID, _ := GetUserID(c)
		username, _ := GetUsername(c)
		isAdmin, _ := GetIsAdmin(c)

		c.JSON(http.StatusOK, gin.H{
			"user_id":  userID,
			"username": username,
			"is_admin": isAdmin,
		})
	})

	// 创建请求
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	// 执行请求
	router.ServeHTTP(w, req)

	// 验证响应
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthMiddleware_NoToken(t *testing.T) {
	jwtManager := auth.NewJWTManager("test-secret-key", 24)

	router := setupTestRouter()
	router.GET("/test", AuthMiddleware(jwtManager), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_InvalidTokenFormat(t *testing.T) {
	jwtManager := auth.NewJWTManager("test-secret-key", 24)

	router := setupTestRouter()
	router.GET("/test", AuthMiddleware(jwtManager), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// 测试无效格式
	testCases := []string{
		"InvalidToken",
		"Bearer",
		"Basic token123",
	}

	for _, tc := range testCases {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", tc)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	jwtManager := auth.NewJWTManager("test-secret-key", 24)

	router := setupTestRouter()
	router.GET("/test", AuthMiddleware(jwtManager), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAdminMiddleware_Success(t *testing.T) {
	jwtManager := auth.NewJWTManager("test-secret-key", 24)

	// 生成管理员Token
	token, err := jwtManager.GenerateToken(1, "admin", true)
	assert.NoError(t, err)

	router := setupTestRouter()
	router.GET("/admin", AuthMiddleware(jwtManager), AdminMiddleware(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "admin access granted"})
	})

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAdminMiddleware_NonAdmin(t *testing.T) {
	jwtManager := auth.NewJWTManager("test-secret-key", 24)

	// 生成普通用户Token
	token, err := jwtManager.GenerateToken(1, "user", false)
	assert.NoError(t, err)

	router := setupTestRouter()
	router.GET("/admin", AuthMiddleware(jwtManager), AdminMiddleware(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "admin access granted"})
	})

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestAdminMiddleware_NoAuth(t *testing.T) {
	router := setupTestRouter()
	router.GET("/admin", AdminMiddleware(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "admin access granted"})
	})

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetUserID(t *testing.T) {
	router := setupTestRouter()
	router.GET("/test", func(c *gin.Context) {
		c.Set(ContextKeyUserID, int64(123))
		userID, exists := GetUserID(c)
		assert.True(t, exists)
		assert.Equal(t, int64(123), userID)
		c.JSON(http.StatusOK, gin.H{"user_id": userID})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetUsername(t *testing.T) {
	router := setupTestRouter()
	router.GET("/test", func(c *gin.Context) {
		c.Set(ContextKeyUsername, "testuser")
		username, exists := GetUsername(c)
		assert.True(t, exists)
		assert.Equal(t, "testuser", username)
		c.JSON(http.StatusOK, gin.H{"username": username})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetIsAdmin(t *testing.T) {
	router := setupTestRouter()
	router.GET("/test", func(c *gin.Context) {
		c.Set(ContextKeyIsAdmin, true)
		isAdmin, exists := GetIsAdmin(c)
		assert.True(t, exists)
		assert.True(t, isAdmin)
		c.JSON(http.StatusOK, gin.H{"is_admin": isAdmin})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestContextHelpers_NotExists(t *testing.T) {
	router := setupTestRouter()
	router.GET("/test", func(c *gin.Context) {
		_, exists := GetUserID(c)
		assert.False(t, exists)

		_, exists = GetUsername(c)
		assert.False(t, exists)

		_, exists = GetIsAdmin(c)
		assert.False(t, exists)

		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
