package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"go-shop/internal/entity"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockOperationLogRepository 模拟操作日志仓库
type MockOperationLogRepository struct {
	mock.Mock
}

func (m *MockOperationLogRepository) Create(ctx context.Context, log *entity.OperationLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *MockOperationLogRepository) GetByID(ctx context.Context, id int64) (*entity.OperationLog, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.OperationLog), args.Error(1)
}

func (m *MockOperationLogRepository) List(ctx context.Context, offset, limit int) ([]*entity.OperationLog, int64, error) {
	args := m.Called(ctx, offset, limit)
	return args.Get(0).([]*entity.OperationLog), args.Get(1).(int64), args.Error(2)
}

func (m *MockOperationLogRepository) ListByUser(ctx context.Context, userID int64, offset, limit int) ([]*entity.OperationLog, int64, error) {
	args := m.Called(ctx, userID, offset, limit)
	return args.Get(0).([]*entity.OperationLog), args.Get(1).(int64), args.Error(2)
}

func (m *MockOperationLogRepository) ListByModule(ctx context.Context, module string, offset, limit int) ([]*entity.OperationLog, int64, error) {
	args := m.Called(ctx, module, offset, limit)
	return args.Get(0).([]*entity.OperationLog), args.Get(1).(int64), args.Error(2)
}

func (m *MockOperationLogRepository) ListByOperation(ctx context.Context, operation string, offset, limit int) ([]*entity.OperationLog, int64, error) {
	args := m.Called(ctx, operation, offset, limit)
	return args.Get(0).([]*entity.OperationLog), args.Get(1).(int64), args.Error(2)
}

func (m *MockOperationLogRepository) ListByTimeRange(ctx context.Context, startTime, endTime time.Time, offset, limit int) ([]*entity.OperationLog, int64, error) {
	args := m.Called(ctx, startTime, endTime, offset, limit)
	return args.Get(0).([]*entity.OperationLog), args.Get(1).(int64), args.Error(2)
}

func (m *MockOperationLogRepository) ListWithFilters(ctx context.Context, userID *int64, userType *entity.UserType, module *string, operation *string, startTime *time.Time, endTime *time.Time, offset, limit int) ([]*entity.OperationLog, int64, error) {
	args := m.Called(ctx, userID, userType, module, operation, startTime, endTime, offset, limit)
	return args.Get(0).([]*entity.OperationLog), args.Get(1).(int64), args.Error(2)
}

func (m *MockOperationLogRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockOperationLogRepository) DeleteByUser(ctx context.Context, userID int64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockOperationLogRepository) DeleteOldLogs(ctx context.Context, days int) error {
	args := m.Called(ctx, days)
	return args.Error(0)
}

func (m *MockOperationLogRepository) BatchCreate(ctx context.Context, logs []*entity.OperationLog) error {
	args := m.Called(ctx, logs)
	return args.Error(0)
}

func (m *MockOperationLogRepository) CountByModule(ctx context.Context, module string) (int64, error) {
	args := m.Called(ctx, module)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockOperationLogRepository) CountByOperation(ctx context.Context, operation string) (int64, error) {
	args := m.Called(ctx, operation)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockOperationLogRepository) CountErrorLogs(ctx context.Context, startTime, endTime time.Time) (int64, error) {
	args := m.Called(ctx, startTime, endTime)
	return args.Get(0).(int64), args.Error(1)
}

// TestOperationLogMiddleware_BasicRequest 测试基本请求记录
func TestOperationLogMiddleware_BasicRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockOperationLogRepository)
	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(log *entity.OperationLog) bool {
		return log.Method == "GET" && log.Path == "/api/v1/products"
	})).Return(nil)

	router := gin.New()
	router.Use(OperationLogMiddleware(mockRepo))
	router.GET("/api/v1/products", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/api/v1/products", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	// 等待异步日志保存
	time.Sleep(100 * time.Millisecond)

	mockRepo.AssertExpectations(t)
}

// TestOperationLogMiddleware_WithAuthentication 测试带认证的请求记录
func TestOperationLogMiddleware_WithAuthentication(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockOperationLogRepository)
	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(log *entity.OperationLog) bool {
		return log.UserID != nil && *log.UserID == 123 && log.UserType != nil && *log.UserType == entity.UserTypeUser
	})).Return(nil)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		// 模拟认证中间件设置用户信息
		c.Set(ContextKeyUserID, int64(123))
		c.Set(ContextKeyIsAdmin, false)
		c.Next()
	})
	router.Use(OperationLogMiddleware(mockRepo))
	router.GET("/api/v1/orders", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/api/v1/orders", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	// 等待异步日志保存
	time.Sleep(100 * time.Millisecond)

	mockRepo.AssertExpectations(t)
}

// TestOperationLogMiddleware_WithAdminUser 测试管理员用户请求记录
func TestOperationLogMiddleware_WithAdminUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockOperationLogRepository)
	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(log *entity.OperationLog) bool {
		return log.UserID != nil && *log.UserID == 1 && log.UserType != nil && *log.UserType == entity.UserTypeAdmin
	})).Return(nil)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		// 模拟认证中间件设置管理员信息
		c.Set(ContextKeyUserID, int64(1))
		c.Set(ContextKeyIsAdmin, true)
		c.Next()
	})
	router.Use(OperationLogMiddleware(mockRepo))
	router.POST("/api/v1/admin/products", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("POST", "/api/v1/admin/products", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	// 等待异步日志保存
	time.Sleep(100 * time.Millisecond)

	mockRepo.AssertExpectations(t)
}

// TestOperationLogMiddleware_SanitizePassword 测试密码脱敏
func TestOperationLogMiddleware_SanitizePassword(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockOperationLogRepository)
	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(log *entity.OperationLog) bool {
		// 验证密码已被脱敏
		return !bytes.Contains([]byte(log.Request), []byte("secret123"))
	})).Return(nil)

	router := gin.New()
	router.Use(OperationLogMiddleware(mockRepo))
	router.POST("/api/v1/auth/login", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	body := map[string]string{
		"username": "testuser",
		"password": "secret123",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	// 等待异步日志保存
	time.Sleep(100 * time.Millisecond)

	mockRepo.AssertExpectations(t)
}

// TestOperationLogMiddleware_ErrorHandling 测试错误处理
func TestOperationLogMiddleware_ErrorHandling(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockOperationLogRepository)
	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(log *entity.OperationLog) bool {
		return log.Error != "" && *log.Status == 500
	})).Return(nil)

	router := gin.New()
	router.Use(OperationLogMiddleware(mockRepo))
	router.GET("/api/v1/error", func(c *gin.Context) {
		c.Error(assert.AnError)
		c.JSON(500, gin.H{"error": "internal error"})
	})

	req := httptest.NewRequest("GET", "/api/v1/error", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 500, w.Code)

	// 等待异步日志保存
	time.Sleep(100 * time.Millisecond)

	mockRepo.AssertExpectations(t)
}

// TestSanitizeJSONData 测试JSON数据脱敏
func TestSanitizeJSONData(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name: "脱敏密码字段",
			input: map[string]interface{}{
				"username": "test",
				"password": "secret123",
			},
			expected: map[string]interface{}{
				"username": "test",
				"password": "***REDACTED***",
			},
		},
		{
			name: "脱敏token字段",
			input: map[string]interface{}{
				"access_token": "abc123",
				"data":         "normal",
			},
			expected: map[string]interface{}{
				"access_token": "***REDACTED***",
				"data":         "normal",
			},
		},
		{
			name: "脱敏嵌套对象",
			input: map[string]interface{}{
				"user": map[string]interface{}{
					"name":     "test",
					"password": "secret",
				},
			},
			expected: map[string]interface{}{
				"user": map[string]interface{}{
					"name":     "test",
					"password": "***REDACTED***",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sanitizeJSONData(tt.input)
			assert.Equal(t, tt.expected, tt.input)
		})
	}
}

// TestDetermineOperationAndModule 测试操作类型和模块确定
func TestDetermineOperationAndModule(t *testing.T) {
	tests := []struct {
		name              string
		method            string
		path              string
		expectedOperation string
		expectedModule    string
	}{
		{
			name:              "登录操作",
			method:            "POST",
			path:              "/api/v1/auth/login",
			expectedOperation: "login",
			expectedModule:    "auth",
		},
		{
			name:              "创建商品",
			method:            "POST",
			path:              "/api/v1/admin/products",
			expectedOperation: "create_products",
			expectedModule:    "products",
		},
		{
			name:              "更新订单",
			method:            "PUT",
			path:              "/api/v1/orders/123",
			expectedOperation: "update_orders",
			expectedModule:    "orders",
		},
		{
			name:              "删除用户",
			method:            "DELETE",
			path:              "/api/v1/admin/users/456",
			expectedOperation: "delete_users",
			expectedModule:    "users",
		},
		{
			name:              "查看商品列表",
			method:            "GET",
			path:              "/api/v1/products",
			expectedOperation: "view_products",
			expectedModule:    "products",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			operation, module := determineOperationAndModule(tt.method, tt.path)
			assert.Equal(t, tt.expectedOperation, operation)
			assert.Equal(t, tt.expectedModule, module)
		})
	}
}

// TestSanitizeRequestData 测试请求数据脱敏
func TestSanitizeRequestData(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains string
		notContains string
	}{
		{
			name:        "JSON格式密码脱敏",
			input:       `{"username":"test","password":"secret123"}`,
			contains:    "***REDACTED***",
			notContains: "secret123",
		},
		{
			name:        "空字符串",
			input:       "",
			contains:    "",
			notContains: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeRequestData(tt.input)
			if tt.contains != "" {
				assert.Contains(t, result, tt.contains)
			}
			if tt.notContains != "" {
				assert.NotContains(t, result, tt.notContains)
			}
		})
	}
}

// TestOperationLogMiddleware_DoesNotBlockMainFlow 测试日志失败不阻塞主流程
func TestOperationLogMiddleware_DoesNotBlockMainFlow(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockOperationLogRepository)
	// 模拟日志保存失败
	mockRepo.On("Create", mock.Anything, mock.Anything).Return(assert.AnError)

	router := gin.New()
	router.Use(OperationLogMiddleware(mockRepo))
	router.GET("/api/v1/products", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/api/v1/products", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 即使日志保存失败，主流程应该正常返回
	assert.Equal(t, 200, w.Code)

	// 等待异步日志保存尝试
	time.Sleep(100 * time.Millisecond)

	mockRepo.AssertExpectations(t)
}
