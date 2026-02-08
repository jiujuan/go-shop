package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"regexp"
	"strings"
	"time"

	"go-shop/internal/entity"
	"go-shop/internal/repository"
	"go-shop/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// OperationLogMiddleware 操作日志中间件
// 记录所有HTTP请求，异步保存日志，不阻塞主流程
func OperationLogMiddleware(repo repository.OperationLogRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录开始时间
		startTime := time.Now()

		// 获取用户信息（如果已认证）
		var userID *int64
		var userType *entity.UserType
		if uid, exists := GetUserID(c); exists {
			userID = &uid
			// 判断用户类型
			if isAdmin, exists := GetIsAdmin(c); exists && isAdmin {
				adminType := entity.UserTypeAdmin
				userType = &adminType
			} else {
				normalType := entity.UserTypeUser
				userType = &normalType
			}
		}

		// 读取请求体（如果有）
		var requestBody string
		if c.Request.Body != nil && shouldLogRequestBody(c.Request.Method) {
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err == nil {
				// 恢复请求体供后续使用
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				
				// 脱敏处理
				requestBody = sanitizeRequestData(string(bodyBytes))
			}
		}

		// 创建响应写入器包装器以捕获响应
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		// 处理请求
		c.Next()

		// 计算耗时
		duration := time.Since(startTime).Milliseconds()

		// 获取响应状态码
		status := c.Writer.Status()

		// 获取错误信息（如果有）
		var errorMsg string
		if len(c.Errors) > 0 {
			errorMsg = c.Errors.String()
		}

		// 获取响应体（限制大小）
		responseBody := sanitizeResponseData(blw.body.String())

		// 确定操作类型和模块
		operation, module := determineOperationAndModule(c.Request.Method, c.Request.URL.Path)

		// 构建操作日志
		log := &entity.OperationLog{
			UserID:    userID,
			UserType:  userType,
			Operation: operation,
			Module:    module,
			Method:    c.Request.Method,
			Path:      c.Request.URL.Path,
			IP:        c.ClientIP(),
			UserAgent: c.Request.UserAgent(),
			Request:   requestBody,
			Response:  responseBody,
			Status:    &status,
			Duration:  &duration,
			Error:     errorMsg,
			CreatedAt: startTime,
		}

		// 异步保存日志（不阻塞主流程）
		go saveOperationLogAsync(repo, log)
	}
}

// saveOperationLogAsync 异步保存操作日志
func saveOperationLogAsync(repo repository.OperationLogRepository, log *entity.OperationLog) {
	// 使用defer捕获panic，确保日志失败不影响主流程
	defer func() {
		if r := recover(); r != nil {
			logger.Error("Failed to save operation log (panic recovered)",
				zap.Any("panic", r),
				zap.String("path", log.Path),
				zap.String("method", log.Method),
			)
		}
	}()

	// 验证日志数据
	if err := log.Validate(); err != nil {
		logger.Error("Invalid operation log data",
			zap.Error(err),
			zap.String("path", log.Path),
			zap.String("method", log.Method),
		)
		return
	}

	// 保存日志到数据库（使用新的context，因为原请求context可能已关闭）
	ctx := context.Background()
	if err := repo.Create(ctx, log); err != nil {
		// 日志保存失败，记录错误但不影响主流程
		logger.Error("Failed to save operation log to database",
			zap.Error(err),
			zap.String("path", log.Path),
			zap.String("method", log.Method),
			zap.Int64p("user_id", log.UserID),
		)
	}
}

// shouldLogRequestBody 判断是否应该记录请求体
func shouldLogRequestBody(method string) bool {
	// 只记录POST、PUT、PATCH、DELETE请求的请求体
	return method == "POST" || method == "PUT" || method == "PATCH" || method == "DELETE"
}

// sanitizeRequestData 脱敏请求数据
func sanitizeRequestData(data string) string {
	if data == "" {
		return ""
	}

	// 限制长度（最多10KB）
	if len(data) > 10240 {
		data = data[:10240] + "...[truncated]"
	}

	// 尝试解析为JSON并脱敏
	var jsonData map[string]interface{}
	if err := json.Unmarshal([]byte(data), &jsonData); err == nil {
		sanitizeJSONData(jsonData)
		if sanitized, err := json.Marshal(jsonData); err == nil {
			return string(sanitized)
		}
	}

	// 如果不是JSON，使用正则表达式脱敏
	return sanitizeWithRegex(data)
}

// sanitizeResponseData 脱敏响应数据
func sanitizeResponseData(data string) string {
	if data == "" {
		return ""
	}

	// 限制长度（最多10KB）
	if len(data) > 10240 {
		data = data[:10240] + "...[truncated]"
	}

	// 尝试解析为JSON并脱敏
	var jsonData map[string]interface{}
	if err := json.Unmarshal([]byte(data), &jsonData); err == nil {
		sanitizeJSONData(jsonData)
		if sanitized, err := json.Marshal(jsonData); err == nil {
			return string(sanitized)
		}
	}

	return data
}

// sanitizeJSONData 脱敏JSON数据
func sanitizeJSONData(data map[string]interface{}) {
	// 敏感字段列表
	sensitiveFields := []string{
		"password",
		"old_password",
		"new_password",
		"confirm_password",
		"token",
		"access_token",
		"refresh_token",
		"secret",
		"api_key",
		"private_key",
		"credit_card",
		"card_number",
		"cvv",
		"ssn",
		"id_card",
		"bank_account",
	}

	for key, value := range data {
		lowerKey := strings.ToLower(key)
		
		// 检查是否为敏感字段
		for _, sensitive := range sensitiveFields {
			if strings.Contains(lowerKey, sensitive) {
				data[key] = "***REDACTED***"
				break
			}
		}

		// 递归处理嵌套对象
		if nestedMap, ok := value.(map[string]interface{}); ok {
			sanitizeJSONData(nestedMap)
		}

		// 处理数组
		if arr, ok := value.([]interface{}); ok {
			for _, item := range arr {
				if nestedMap, ok := item.(map[string]interface{}); ok {
					sanitizeJSONData(nestedMap)
				}
			}
		}
	}
}

// sanitizeWithRegex 使用正则表达式脱敏
func sanitizeWithRegex(data string) string {
	// 脱敏密码字段
	passwordRegex := regexp.MustCompile(`("password"\s*:\s*)"[^"]*"`)
	data = passwordRegex.ReplaceAllString(data, `$1"***REDACTED***"`)

	// 脱敏token字段
	tokenRegex := regexp.MustCompile(`("token"\s*:\s*)"[^"]*"`)
	data = tokenRegex.ReplaceAllString(data, `$1"***REDACTED***"`)

	// 脱敏API key字段
	apiKeyRegex := regexp.MustCompile(`("api_key"\s*:\s*)"[^"]*"`)
	data = apiKeyRegex.ReplaceAllString(data, `$1"***REDACTED***"`)

	return data
}

// determineOperationAndModule 根据请求路径和方法确定操作类型和模块
func determineOperationAndModule(method, path string) (operation, module string) {
	// 默认值
	operation = method
	module = "unknown"

	// 解析路径
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 2 {
		return
	}

	// 跳过 "api" 和版本号
	startIdx := 0
	for i, part := range parts {
		if part == "api" || strings.HasPrefix(part, "v") {
			startIdx = i + 1
		} else {
			break
		}
	}

	if startIdx >= len(parts) {
		return
	}

	// 确定模块
	module = parts[startIdx]

	// 特殊处理admin路径
	if module == "admin" && startIdx+1 < len(parts) {
		module = parts[startIdx+1]
	}

	// 确定操作类型
	switch {
	case strings.Contains(path, "/login"):
		operation = "login"
	case strings.Contains(path, "/logout"):
		operation = "logout"
	case strings.Contains(path, "/register"):
		operation = "register"
	case method == "POST" && !strings.Contains(path, "/login") && !strings.Contains(path, "/register"):
		operation = "create_" + module
	case method == "PUT" || method == "PATCH":
		operation = "update_" + module
	case method == "DELETE":
		operation = "delete_" + module
	case method == "GET":
		operation = "view_" + module
	default:
		operation = strings.ToLower(method) + "_" + module
	}

	return
}
