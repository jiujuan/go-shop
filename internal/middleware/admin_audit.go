package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"time"

	"go-shop/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AdminAuditLog 管理员操作日志结构
type AdminAuditLog struct {
	Timestamp  time.Time              `json:"timestamp"`
	UserID     int64                  `json:"user_id"`
	Username   string                 `json:"username"`
	Method     string                 `json:"method"`
	Path       string                 `json:"path"`
	Query      string                 `json:"query"`
	Body       map[string]interface{} `json:"body,omitempty"`
	IP         string                 `json:"ip"`
	UserAgent  string                 `json:"user_agent"`
	StatusCode int                    `json:"status_code"`
	Duration   int64                  `json:"duration_ms"`
	Error      string                 `json:"error,omitempty"`
}

// AdminAuditMiddleware 管理员操作审计中间件
func AdminAuditMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录开始时间
		startTime := time.Now()

		// 获取用户信息
		userID, _ := GetUserID(c)
		username, _ := GetUsername(c)

		// 读取请求体（如果有）
		var requestBody map[string]interface{}
		if c.Request.Body != nil && c.Request.Method != "GET" {
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err == nil {
				// 恢复请求体供后续使用
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				
				// 尝试解析JSON
				if len(bodyBytes) > 0 {
					json.Unmarshal(bodyBytes, &requestBody)
					// 移除敏感字段
					sanitizeRequestBody(requestBody)
				}
			}
		}

		// 创建响应写入器包装器
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		// 处理请求
		c.Next()

		// 计算耗时
		duration := time.Since(startTime)

		// 构建审计日志
		auditLog := AdminAuditLog{
			Timestamp:  startTime,
			UserID:     userID,
			Username:   username,
			Method:     c.Request.Method,
			Path:       c.Request.URL.Path,
			Query:      c.Request.URL.RawQuery,
			Body:       requestBody,
			IP:         c.ClientIP(),
			UserAgent:  c.Request.UserAgent(),
			StatusCode: c.Writer.Status(),
			Duration:   duration.Milliseconds(),
		}

		// 如果有错误，记录错误信息
		if len(c.Errors) > 0 {
			auditLog.Error = c.Errors.String()
		}

		// 记录审计日志
		logAdminOperation(auditLog)
	}
}

// bodyLogWriter 响应写入器包装器
type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// logAdminOperation 记录管理员操作
func logAdminOperation(log AdminAuditLog) {
	fields := []zap.Field{
		zap.Time("timestamp", log.Timestamp),
		zap.Int64("user_id", log.UserID),
		zap.String("username", log.Username),
		zap.String("method", log.Method),
		zap.String("path", log.Path),
		zap.String("ip", log.IP),
		zap.Int("status", log.StatusCode),
		zap.Int64("duration_ms", log.Duration),
	}

	if log.Query != "" {
		fields = append(fields, zap.String("query", log.Query))
	}

	if len(log.Body) > 0 {
		fields = append(fields, zap.Any("body", log.Body))
	}

	if log.Error != "" {
		fields = append(fields, zap.String("error", log.Error))
		logger.Error("Admin operation failed", fields...)
	} else {
		logger.Info("Admin operation", fields...)
	}
}

// sanitizeRequestBody 清理请求体中的敏感信息
func sanitizeRequestBody(body map[string]interface{}) {
	sensitiveFields := []string{
		"password",
		"old_password",
		"new_password",
		"token",
		"secret",
		"api_key",
		"access_token",
		"refresh_token",
	}

	for _, field := range sensitiveFields {
		if _, exists := body[field]; exists {
			body[field] = "***REDACTED***"
		}
	}
}

// AdminOperationLogger 管理员操作日志记录器（用于特定操作）
type AdminOperationLogger struct {
	Operation string
	Target    string
	Details   map[string]interface{}
}

// LogAdminAction 记录特定的管理员操作
func LogAdminAction(c *gin.Context, operation, target string, details map[string]interface{}) {
	userID, _ := GetUserID(c)
	username, _ := GetUsername(c)

	fields := []zap.Field{
		zap.String("operation", operation),
		zap.String("target", target),
		zap.Int64("user_id", userID),
		zap.String("username", username),
		zap.String("ip", c.ClientIP()),
	}

	if details != nil {
		fields = append(fields, zap.Any("details", details))
	}

	logger.Info("Admin action", fields...)
}
