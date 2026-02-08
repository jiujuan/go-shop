package errors

import (
	"fmt"
)

// ErrorCode 错误码类型
type ErrorCode int

// 错误码定义
const (
	// 通用错误码 (0-999)
	CodeSuccess         ErrorCode = 0
	CodeError           ErrorCode = 1
	CodeInternalError   ErrorCode = 500
	CodeInvalidParam    ErrorCode = 400
	CodeUnauthorized    ErrorCode = 401
	CodeForbidden       ErrorCode = 403
	CodeNotFound        ErrorCode = 404
	CodeConflict        ErrorCode = 409
	CodeTooManyRequests ErrorCode = 429

	// 用户相关错误码 (1000-1999)
	CodeUserNotFound        ErrorCode = 1001
	CodeUserAlreadyExists   ErrorCode = 1002
	CodeInvalidPassword     ErrorCode = 1003
	CodeInvalidUsername     ErrorCode = 1004
	CodeInvalidEmail        ErrorCode = 1005
	CodeEmailAlreadyExists  ErrorCode = 1006
	CodeUserDisabled        ErrorCode = 1007

	// 认证相关错误码 (2000-2999)
	CodeTokenInvalid   ErrorCode = 2001
	CodeTokenExpired   ErrorCode = 2002
	CodeTokenMissing   ErrorCode = 2003
	CodePermissionDeny ErrorCode = 2004

	// 商品相关错误码 (3000-3999)
	CodeProductNotFound     ErrorCode = 3001
	CodeProductOutOfStock   ErrorCode = 3002
	CodeProductNotAvailable ErrorCode = 3003
	CodeProductAlreadyExists ErrorCode = 3004
	CodeInvalidProductID    ErrorCode = 3005
	CodeInvalidCategoryID   ErrorCode = 3006
	CodeInvalidStock        ErrorCode = 3007
	CodeInvalidPrice        ErrorCode = 3008

	// 分类相关错误码 (3500-3599)
	CodeCategoryNotFound      ErrorCode = 3501
	CodeCategoryAlreadyExists ErrorCode = 3502

	// 购物车相关错误码 (4000-4999)
	CodeCartItemNotFound ErrorCode = 4001
	CodeCartEmpty        ErrorCode = 4002
	CodeCartItemInvalid  ErrorCode = 4003

	// 订单相关错误码 (5000-5999)
	CodeOrderNotFound      ErrorCode = 5001
	CodeOrderStatusInvalid ErrorCode = 5002
	CodeOrderCannotCancel  ErrorCode = 5003
	CodeOrderAlreadyPaid   ErrorCode = 5004
	CodeOrderCannotPay     ErrorCode = 5005
	CodeOrderCannotShip    ErrorCode = 5006

	// 地址相关错误码 (6000-6999)
	CodeAddressNotFound       ErrorCode = 6001
	CodeAddressAlreadyDefault ErrorCode = 6002
	CodeCannotDeleteDefault   ErrorCode = 6003
	CodeInvalidAddressID      ErrorCode = 6004

	// 支付相关错误码 (7000-7999)
	CodePaymentFailed    ErrorCode = 7001
	CodePaymentTimeout   ErrorCode = 7002
	CodePaymentCancelled ErrorCode = 7003
	CodeRefundFailed     ErrorCode = 7004
)

// 错误消息映射
var errorMessages = map[ErrorCode]string{
	// 通用错误
	CodeSuccess:         "操作成功",
	CodeError:           "操作失败",
	CodeInternalError:   "服务器内部错误",
	CodeInvalidParam:    "参数错误",
	CodeUnauthorized:    "未授权访问",
	CodeForbidden:       "禁止访问",
	CodeNotFound:        "资源不存在",
	CodeConflict:        "资源冲突",
	CodeTooManyRequests: "请求过于频繁",

	// 用户相关
	CodeUserNotFound:       "用户不存在",
	CodeUserAlreadyExists:  "用户已存在",
	CodeInvalidPassword:    "密码错误",
	CodeInvalidUsername:    "用户名格式错误",
	CodeInvalidEmail:       "邮箱格式错误",
	CodeEmailAlreadyExists: "邮箱已被使用",
	CodeUserDisabled:       "用户已被禁用",

	// 认证相关
	CodeTokenInvalid:   "令牌无效",
	CodeTokenExpired:   "令牌已过期",
	CodeTokenMissing:   "缺少认证令牌",
	CodePermissionDeny: "权限不足",

	// 商品相关
	CodeProductNotFound:      "商品不存在",
	CodeProductOutOfStock:    "商品库存不足",
	CodeProductNotAvailable:  "商品已下架",
	CodeProductAlreadyExists: "商品已存在",
	CodeInvalidProductID:     "无效的商品ID",
	CodeInvalidCategoryID:    "无效的分类ID",
	CodeInvalidStock:         "无效的库存数量",
	CodeInvalidPrice:         "无效的商品价格",

	// 分类相关
	CodeCategoryNotFound:      "分类不存在",
	CodeCategoryAlreadyExists: "分类已存在",

	// 购物车相关
	CodeCartItemNotFound: "购物车商品不存在",
	CodeCartEmpty:        "购物车为空",
	CodeCartItemInvalid:  "购物车商品无效",

	// 订单相关
	CodeOrderNotFound:      "订单不存在",
	CodeOrderStatusInvalid: "订单状态无效",
	CodeOrderCannotCancel:  "订单无法取消",
	CodeOrderAlreadyPaid:   "订单已支付",
	CodeOrderCannotPay:     "订单无法支付",
	CodeOrderCannotShip:    "订单无法发货",

	// 地址相关
	CodeAddressNotFound:       "地址不存在",
	CodeAddressAlreadyDefault: "地址已是默认地址",
	CodeCannotDeleteDefault:   "不能删除默认地址",
	CodeInvalidAddressID:      "无效的地址ID",

	// 支付相关
	CodePaymentFailed:    "支付失败",
	CodePaymentTimeout:   "支付超时",
	CodePaymentCancelled: "支付已取消",
	CodeRefundFailed:     "退款失败",
}

// BusinessError 业务错误
type BusinessError struct {
	Code    ErrorCode
	Message string
	Err     error
}

// Error 实现 error 接口
func (e *BusinessError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// Unwrap 实现 errors.Unwrap 接口
func (e *BusinessError) Unwrap() error {
	return e.Err
}

// New 创建业务错误
func New(code ErrorCode, message string) *BusinessError {
	if message == "" {
		message = GetMessage(code)
	}
	return &BusinessError{
		Code:    code,
		Message: message,
	}
}

// Wrap 包装错误
func Wrap(code ErrorCode, err error) *BusinessError {
	return &BusinessError{
		Code:    code,
		Message: GetMessage(code),
		Err:     err,
	}
}

// Wrapf 包装错误（带格式化消息）
func Wrapf(code ErrorCode, err error, format string, args ...interface{}) *BusinessError {
	return &BusinessError{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
		Err:     err,
	}
}

// GetMessage 获取错误消息
func GetMessage(code ErrorCode) string {
	if msg, ok := errorMessages[code]; ok {
		return msg
	}
	return "未知错误"
}

// GetHTTPStatus 获取 HTTP 状态码
func GetHTTPStatus(code ErrorCode) int {
	switch {
	case code == CodeSuccess:
		return 200
	case code >= 400 && code < 500:
		return int(code)
	case code == CodeUnauthorized || code >= 2000 && code < 3000:
		return 401
	case code == CodeForbidden:
		return 403
	case code == CodeNotFound || code%1000 == 1:
		return 404
	case code == CodeConflict || code%1000 == 2:
		return 409
	case code == CodeTooManyRequests:
		return 429
	default:
		return 500
	}
}

// 预定义错误
var (
	// 通用错误
	ErrInternalError   = New(CodeInternalError, "")
	ErrInvalidParam    = New(CodeInvalidParam, "")
	ErrUnauthorized    = New(CodeUnauthorized, "")
	ErrForbidden       = New(CodeForbidden, "")
	ErrNotFound        = New(CodeNotFound, "")
	ErrConflict        = New(CodeConflict, "")
	ErrTooManyRequests = New(CodeTooManyRequests, "")

	// 用户相关
	ErrUserNotFound       = New(CodeUserNotFound, "")
	ErrUserAlreadyExists  = New(CodeUserAlreadyExists, "")
	ErrInvalidPassword    = New(CodeInvalidPassword, "")
	ErrInvalidUsername    = New(CodeInvalidUsername, "")
	ErrInvalidEmail       = New(CodeInvalidEmail, "")
	ErrEmailAlreadyExists = New(CodeEmailAlreadyExists, "")

	// 认证相关
	ErrTokenInvalid   = New(CodeTokenInvalid, "")
	ErrTokenExpired   = New(CodeTokenExpired, "")
	ErrTokenMissing   = New(CodeTokenMissing, "")
	ErrPermissionDeny = New(CodePermissionDeny, "")

	// 商品相关
	ErrProductNotFound     = New(CodeProductNotFound, "")
	ErrProductOutOfStock   = New(CodeProductOutOfStock, "")
	ErrProductNotAvailable = New(CodeProductNotAvailable, "")

	// 购物车相关
	ErrCartItemNotFound = New(CodeCartItemNotFound, "")
	ErrCartEmpty        = New(CodeCartEmpty, "")

	// 订单相关
	ErrOrderNotFound      = New(CodeOrderNotFound, "")
	ErrOrderStatusInvalid = New(CodeOrderStatusInvalid, "")
	ErrOrderCannotCancel  = New(CodeOrderCannotCancel, "")
	ErrOrderAlreadyPaid   = New(CodeOrderAlreadyPaid, "")

	// 地址相关
	ErrAddressNotFound      = New(CodeAddressNotFound, "")
	ErrCannotDeleteDefault  = New(CodeCannotDeleteDefault, "")
	ErrInvalidAddressID     = New(CodeInvalidAddressID, "")
)
