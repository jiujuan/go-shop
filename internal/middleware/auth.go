package middleware

import (
	"strings"

	"go-shop/pkg/auth"
	"go-shop/pkg/logger"
	"go-shop/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	// ContextKeyUserID 用户ID上下文键
	ContextKeyUserID = "user_id"
	// ContextKeyUsername 用户名上下文键
	ContextKeyUsername = "username"
	// ContextKeyIsAdmin 管理员标识上下文键
	ContextKeyIsAdmin = "is_admin"
)

// AuthMiddleware JWT认证中间件
func AuthMiddleware(jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取Token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "未提供认证令牌")
			c.Abort()
			return
		}

		// 验证Token格式 (Bearer <token>)
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Unauthorized(c, "认证令牌格式错误")
			c.Abort()
			return
		}

		tokenString := parts[1]

		// 解析Token
		claims, err := jwtManager.ParseToken(tokenString)
		if err != nil {
			response.Unauthorized(c, "认证令牌无效或已过期")
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set(ContextKeyUserID, claims.UserID)
		c.Set(ContextKeyUsername, claims.Username)
		c.Set(ContextKeyIsAdmin, claims.IsAdmin)

		c.Next()
	}
}

// AdminMiddleware 管理员权限中间件
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户ID
		userID, exists := GetUserID(c)
		if !exists {
			response.Unauthorized(c, "未认证")
			c.Abort()
			return
		}

		// 获取用户管理员标识
		isAdmin, exists := c.Get(ContextKeyIsAdmin)
		if !exists {
			response.Unauthorized(c, "未认证")
			c.Abort()
			return
		}

		// 验证是否为管理员
		if !isAdmin.(bool) {
			// 记录未授权访问尝试
			username, _ := GetUsername(c)
			logger.Warn("Unauthorized admin access attempt",
				zap.Int64("user_id", userID),
				zap.String("username", username),
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method),
				zap.String("ip", c.ClientIP()),
			)
			
			response.Forbidden(c, "需要管理员权限")
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetUserID 从上下文获取用户ID
func GetUserID(c *gin.Context) (int64, bool) {
	userID, exists := c.Get(ContextKeyUserID)
	if !exists {
		return 0, false
	}
	return userID.(int64), true
}

// GetUsername 从上下文获取用户名
func GetUsername(c *gin.Context) (string, bool) {
	username, exists := c.Get(ContextKeyUsername)
	if !exists {
		return "", false
	}
	return username.(string), true
}

// GetIsAdmin 从上下文获取管理员标识
func GetIsAdmin(c *gin.Context) (bool, bool) {
	isAdmin, exists := c.Get(ContextKeyIsAdmin)
	if !exists {
		return false, false
	}
	return isAdmin.(bool), true
}
