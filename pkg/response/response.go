package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// 响应状态码常量
const (
	CodeSuccess      = 0     // 成功
	CodeError        = -1    // 通用错误
	CodeInvalidParam = 10001 // 参数错误
	CodeUnauthorized = 10002 // 未授权
	CodeForbidden    = 10003 // 禁止访问
	CodeNotFound     = 10004 // 资源不存在
	CodeConflict     = 10005 // 资源冲突
	CodeTooManyRequests = 10006 // 请求过于频繁
	CodeServerError  = 50000 // 服务器内部错误
)

// 响应消息常量
const (
	MsgSuccess      = "操作成功"
	MsgError        = "操作失败"
	MsgInvalidParam = "参数错误"
	MsgUnauthorized = "未授权访问"
	MsgForbidden    = "禁止访问"
	MsgNotFound     = "资源不存在"
	MsgConflict     = "资源冲突"
	MsgServerError  = "服务器内部错误"
)

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: MsgSuccess,
		Data:    data,
	})
}

// SuccessWithMessage 成功响应（自定义消息）
func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: message,
		Data:    data,
	})
}

// Error 错误响应
func Error(c *gin.Context, code int, message string) {
	httpStatus := getHTTPStatus(code)
	c.JSON(httpStatus, Response{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}

// ErrorWithData 错误响应（带数据）
func ErrorWithData(c *gin.Context, code int, message string, data interface{}) {
	httpStatus := getHTTPStatus(code)
	c.JSON(httpStatus, Response{
		Code:    code,
		Message: message,
		Data:    data,
	})
}

// InvalidParam 参数错误响应
func InvalidParam(c *gin.Context, message string) {
	Error(c, CodeInvalidParam, message)
}

// Unauthorized 未授权响应
func Unauthorized(c *gin.Context, message string) {
	if message == "" {
		message = MsgUnauthorized
	}
	Error(c, CodeUnauthorized, message)
}

// Forbidden 禁止访问响应
func Forbidden(c *gin.Context, message string) {
	if message == "" {
		message = MsgForbidden
	}
	Error(c, CodeForbidden, message)
}

// NotFound 资源不存在响应
func NotFound(c *gin.Context, message string) {
	if message == "" {
		message = MsgNotFound
	}
	Error(c, CodeNotFound, message)
}

// Conflict 资源冲突响应
func Conflict(c *gin.Context, message string) {
	if message == "" {
		message = MsgConflict
	}
	Error(c, CodeConflict, message)
}

// ServerError 服务器错误响应
func ServerError(c *gin.Context, message string) {
	if message == "" {
		message = MsgServerError
	}
	Error(c, CodeServerError, message)
}

// getHTTPStatus 根据业务错误码获取 HTTP 状态码
func getHTTPStatus(code int) int {
	switch code {
	case CodeSuccess:
		return http.StatusOK
	case CodeInvalidParam:
		return http.StatusBadRequest
	case CodeUnauthorized:
		return http.StatusUnauthorized
	case CodeForbidden:
		return http.StatusForbidden
	case CodeNotFound:
		return http.StatusNotFound
	case CodeConflict:
		return http.StatusConflict
	case CodeTooManyRequests:
		return http.StatusTooManyRequests
	case CodeServerError:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}