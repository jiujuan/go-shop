package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	err := New(CodeUserNotFound, "")
	assert.NotNil(t, err)
	assert.Equal(t, CodeUserNotFound, err.Code)
	assert.Equal(t, "用户不存在", err.Message)
}

func TestNewWithCustomMessage(t *testing.T) {
	customMsg := "自定义错误消息"
	err := New(CodeUserNotFound, customMsg)
	assert.NotNil(t, err)
	assert.Equal(t, CodeUserNotFound, err.Code)
	assert.Equal(t, customMsg, err.Message)
}

func TestWrap(t *testing.T) {
	originalErr := errors.New("original error")
	err := Wrap(CodeInternalError, originalErr)
	
	assert.NotNil(t, err)
	assert.Equal(t, CodeInternalError, err.Code)
	assert.Equal(t, "服务器内部错误", err.Message)
	assert.Equal(t, originalErr, err.Err)
}

func TestWrapf(t *testing.T) {
	originalErr := errors.New("original error")
	err := Wrapf(CodeInternalError, originalErr, "failed to process: %s", "test")
	
	assert.NotNil(t, err)
	assert.Equal(t, CodeInternalError, err.Code)
	assert.Equal(t, "failed to process: test", err.Message)
	assert.Equal(t, originalErr, err.Err)
}

func TestBusinessError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *BusinessError
		expected string
	}{
		{
			name: "Error without wrapped error",
			err: &BusinessError{
				Code:    CodeUserNotFound,
				Message: "用户不存在",
			},
			expected: "[1001] 用户不存在",
		},
		{
			name: "Error with wrapped error",
			err: &BusinessError{
				Code:    CodeInternalError,
				Message: "服务器内部错误",
				Err:     errors.New("database connection failed"),
			},
			expected: "[500] 服务器内部错误: database connection failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestGetMessage(t *testing.T) {
	tests := []struct {
		code     ErrorCode
		expected string
	}{
		{CodeSuccess, "操作成功"},
		{CodeUserNotFound, "用户不存在"},
		{CodeProductOutOfStock, "商品库存不足"},
		{ErrorCode(99999), "未知错误"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, GetMessage(tt.code))
		})
	}
}

func TestGetHTTPStatus(t *testing.T) {
	tests := []struct {
		code     ErrorCode
		expected int
	}{
		{CodeSuccess, 200},
		{CodeInvalidParam, 400},
		{CodeUnauthorized, 401},
		{CodeForbidden, 403},
		{CodeNotFound, 404},
		{CodeConflict, 409},
		{CodeTooManyRequests, 429},
		{CodeInternalError, 500},
		{CodeUserNotFound, 404},
		{CodeUserAlreadyExists, 409},
		{CodeTokenInvalid, 401},
	}

	for _, tt := range tests {
		t.Run(GetMessage(tt.code), func(t *testing.T) {
			assert.Equal(t, tt.expected, GetHTTPStatus(tt.code))
		})
	}
}

func TestBusinessError_Unwrap(t *testing.T) {
	originalErr := errors.New("original error")
	err := Wrap(CodeInternalError, originalErr)
	
	unwrapped := errors.Unwrap(err)
	assert.Equal(t, originalErr, unwrapped)
}

func TestPredefinedErrors(t *testing.T) {
	tests := []struct {
		name string
		err  *BusinessError
		code ErrorCode
	}{
		{"ErrUserNotFound", ErrUserNotFound, CodeUserNotFound},
		{"ErrInvalidPassword", ErrInvalidPassword, CodeInvalidPassword},
		{"ErrProductNotFound", ErrProductNotFound, CodeProductNotFound},
		{"ErrOrderNotFound", ErrOrderNotFound, CodeOrderNotFound},
		{"ErrCartEmpty", ErrCartEmpty, CodeCartEmpty},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.code, tt.err.Code)
			assert.NotEmpty(t, tt.err.Message)
		})
	}
}
