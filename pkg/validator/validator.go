package validator

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// CustomValidator 自定义验证器
type CustomValidator struct {
	validator *validator.Validate
}

// NewCustomValidator 创建自定义验证器实例
func NewCustomValidator() *CustomValidator {
	v := validator.New()
	
	// 注册自定义验证规则
	v.RegisterValidation("phone", validatePhone)
	v.RegisterValidation("username", validateUsername)
	v.RegisterValidation("password", validatePassword)
	
	// 注册字段名称函数
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
	
	return &CustomValidator{validator: v}
}

// ValidateStruct 验证结构体
func (cv *CustomValidator) ValidateStruct(obj interface{}) error {
	if kindOfData(obj) == reflect.Struct {
		if err := cv.validator.Struct(obj); err != nil {
			return cv.formatValidationError(err)
		}
	}
	return nil
}

// Engine 返回验证器引擎
func (cv *CustomValidator) Engine() interface{} {
	return cv.validator
}

// formatValidationError 格式化验证错误
func (cv *CustomValidator) formatValidationError(err error) error {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		var errorMessages []string
		for _, e := range validationErrors {
			errorMessages = append(errorMessages, cv.getErrorMessage(e))
		}
		return fmt.Errorf("%s", strings.Join(errorMessages, "; "))
	}
	return err
}

// getErrorMessage 获取错误消息
func (cv *CustomValidator) getErrorMessage(fe validator.FieldError) string {
	field := fe.Field()
	tag := fe.Tag()
	param := fe.Param()
	
	switch tag {
	case "required":
		return fmt.Sprintf("%s 是必填字段", field)
	case "min":
		return fmt.Sprintf("%s 最小长度为 %s", field, param)
	case "max":
		return fmt.Sprintf("%s 最大长度为 %s", field, param)
	case "len":
		return fmt.Sprintf("%s 长度必须为 %s", field, param)
	case "email":
		return fmt.Sprintf("%s 必须是有效的邮箱地址", field)
	case "oneof":
		return fmt.Sprintf("%s 必须是以下值之一: %s", field, param)
	case "phone":
		return fmt.Sprintf("%s 必须是有效的手机号码", field)
	case "username":
		return fmt.Sprintf("%s 只能包含字母、数字和下划线", field)
	case "password":
		return fmt.Sprintf("%s 必须包含至少一个字母和一个数字", field)
	default:
		return fmt.Sprintf("%s 验证失败", field)
	}
}

// kindOfData 获取数据类型
func kindOfData(data interface{}) reflect.Kind {
	value := reflect.ValueOf(data)
	valueType := value.Kind()
	
	if valueType == reflect.Ptr {
		valueType = value.Elem().Kind()
	}
	return valueType
}

// 自定义验证规则

// validatePhone 验证手机号码
func validatePhone(fl validator.FieldLevel) bool {
	phone := fl.Field().String()
	// 中国手机号码正则表达式
	phoneRegex := regexp.MustCompile(`^1[3-9]\d{9}$`)
	return phoneRegex.MatchString(phone)
}

// validateUsername 验证用户名
func validateUsername(fl validator.FieldLevel) bool {
	username := fl.Field().String()
	// 用户名只能包含字母、数字和下划线
	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	return usernameRegex.MatchString(username)
}

// validatePassword 验证密码强度
func validatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	
	// 至少包含一个字母
	hasLetter := regexp.MustCompile(`[a-zA-Z]`).MatchString(password)
	// 至少包含一个数字
	hasNumber := regexp.MustCompile(`\d`).MatchString(password)
	
	return hasLetter && hasNumber
}

// InitGinValidator 初始化 Gin 验证器
func InitGinValidator() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		// 注册自定义验证规则到 Gin 的验证器
		v.RegisterValidation("phone", validatePhone)
		v.RegisterValidation("username", validateUsername)
		v.RegisterValidation("password", validatePassword)
		
		// 注册字段名称函数
		v.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			if name == "-" {
				return ""
			}
			return name
		})
	}
}

// ValidationErrorResponse 验证错误响应结构
type ValidationErrorResponse struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// FormatGinValidationError 格式化 Gin 验证错误
func FormatGinValidationError(err error) []ValidationErrorResponse {
	var errors []ValidationErrorResponse
	
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			errors = append(errors, ValidationErrorResponse{
				Field:   e.Field(),
				Message: getGinErrorMessage(e),
			})
		}
	}
	
	return errors
}

// getGinErrorMessage 获取 Gin 验证错误消息
func getGinErrorMessage(fe validator.FieldError) string {
	field := fe.Field()
	tag := fe.Tag()
	param := fe.Param()
	
	switch tag {
	case "required":
		return fmt.Sprintf("%s 是必填字段", field)
	case "min":
		return fmt.Sprintf("%s 最小长度为 %s", field, param)
	case "max":
		return fmt.Sprintf("%s 最大长度为 %s", field, param)
	case "len":
		return fmt.Sprintf("%s 长度必须为 %s", field, param)
	case "email":
		return fmt.Sprintf("%s 必须是有效的邮箱地址", field)
	case "oneof":
		return fmt.Sprintf("%s 必须是以下值之一: %s", field, param)
	case "phone":
		return fmt.Sprintf("%s 必须是有效的手机号码", field)
	case "username":
		return fmt.Sprintf("%s 只能包含字母、数字和下划线", field)
	case "password":
		return fmt.Sprintf("%s 必须包含至少一个字母和一个数字", field)
	default:
		return fmt.Sprintf("%s 验证失败", field)
	}
}