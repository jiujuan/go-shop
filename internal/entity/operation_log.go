package entity

import (
	"errors"
	"time"
)

// UserType 用户类型
type UserType string

const (
	UserTypeUser  UserType = "user"  // 普通用户
	UserTypeAdmin UserType = "admin" // 管理员
)

// OperationLog 操作日志实体
type OperationLog struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	UserID    *int64    `json:"user_id" gorm:"index"`                    // 用户ID（可为空，如未登录操作）
	UserType  *UserType `json:"user_type" gorm:"size:20;index"`          // 用户类型
	Operation string    `json:"operation" gorm:"size:50;not null;index"` // 操作类型
	Module    string    `json:"module" gorm:"size:50;not null;index"`    // 模块名称
	Method    string    `json:"method" gorm:"size:10"`                   // HTTP方法
	Path      string    `json:"path" gorm:"size:255"`                    // 请求路径
	IP        string    `json:"ip" gorm:"size:50"`                       // IP地址
	UserAgent string    `json:"user_agent" gorm:"size:255"`              // User Agent
	Request   string    `json:"request" gorm:"type:text"`                // 请求参数（敏感信息脱敏）
	Response  string    `json:"response" gorm:"type:text"`               // 响应结果
	Status    *int      `json:"status"`                                  // HTTP状态码
	Duration  *int64    `json:"duration"`                                // 请求耗时（毫秒）
	Error     string    `json:"error" gorm:"type:text"`                  // 错误信息
	CreatedAt time.Time `json:"created_at" gorm:"index"`                 // 创建时间
}

// TableName 指定表名
func (OperationLog) TableName() string {
	return "operation_logs"
}

// Validate 验证操作日志数据
func (o *OperationLog) Validate() error {
	if o.Operation == "" {
		return errors.New("operation is required")
	}
	if len(o.Operation) > 50 {
		return errors.New("operation too long")
	}
	if o.Module == "" {
		return errors.New("module is required")
	}
	if len(o.Module) > 50 {
		return errors.New("module too long")
	}
	if o.UserType != nil {
		if *o.UserType != UserTypeUser && *o.UserType != UserTypeAdmin {
			return errors.New("invalid user type")
		}
	}
	return nil
}

// IsValidUserType 检查用户类型是否有效
func IsValidUserType(t string) bool {
	return t == string(UserTypeUser) || t == string(UserTypeAdmin)
}
