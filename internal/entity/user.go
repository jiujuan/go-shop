package entity

import "time"

// User 用户实体
type User struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	Username  string    `json:"username" gorm:"uniqueIndex;size:50"`
	Password  string    `json:"-" gorm:"size:100"`
	Email     string    `json:"email" gorm:"uniqueIndex;size:100"`
	Phone     string    `json:"phone" gorm:"uniqueIndex;size:20"` // 手机号
	IsAdmin   bool      `json:"is_admin" gorm:"default:false"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}