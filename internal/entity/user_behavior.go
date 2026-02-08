package entity

import (
	"errors"
	"time"
)

// ActionType 用户行为类型
type ActionType string

const (
	ActionTypeView     ActionType = "view"      // 浏览
	ActionTypeClick    ActionType = "click"     // 点击
	ActionTypeAddCart  ActionType = "add_cart"  // 加购
	ActionTypePurchase ActionType = "purchase"  // 购买
)

// UserBehavior 用户行为实体
type UserBehavior struct {
	ID         int64      `json:"id" gorm:"primaryKey"`
	UserID     int64      `json:"user_id" gorm:"index;not null"`
	ProductID  int64      `json:"product_id" gorm:"index;not null"`
	ActionType ActionType `json:"action_type" gorm:"size:20;not null;index"`
	CreatedAt  time.Time  `json:"created_at" gorm:"index"`
}

// TableName 指定表名
func (UserBehavior) TableName() string {
	return "user_behaviors"
}

// Validate 验证用户行为数据
func (ub *UserBehavior) Validate() error {
	if ub.UserID <= 0 {
		return errors.New("invalid user id")
	}
	if ub.ProductID <= 0 {
		return ErrInvalidProductID
	}
	if ub.ActionType == "" {
		return errors.New("action type is required")
	}
	// 验证行为类型是否有效
	switch ub.ActionType {
	case ActionTypeView, ActionTypeClick, ActionTypeAddCart, ActionTypePurchase:
		// 有效的行为类型
	default:
		return errors.New("invalid action type")
	}
	return nil
}

// IsValidActionType 检查行为类型是否有效
func IsValidActionType(actionType string) bool {
	switch ActionType(actionType) {
	case ActionTypeView, ActionTypeClick, ActionTypeAddCart, ActionTypePurchase:
		return true
	default:
		return false
	}
}
