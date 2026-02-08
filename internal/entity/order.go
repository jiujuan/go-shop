package entity

import (
	"encoding/json"
	"time"
)

// Order 订单实体
type Order struct {
	ID              string          `json:"id" gorm:"primaryKey;size:32"`
	UserID          int64           `json:"user_id"`
	TotalAmount     int64           `json:"total_amount"`     // 商品总金额，单位：分
	CouponID        *int64          `json:"coupon_id"`        // 使用的优惠券ID
	UserCouponID    *int64          `json:"user_coupon_id"`   // 用户优惠券ID（用于标记已使用）
	CouponDiscount  int64           `json:"coupon_discount"`  // 优惠券折扣金额，单位：分
	FinalAmount     int64           `json:"final_amount"`     // 最终支付金额，单位：分
	Status          int             `json:"status"`           // 0-待支付，1-已支付，2-已发货，3-已完成，4-已取消
	AddressSnapshot json.RawMessage `json:"address_snapshot" gorm:"type:json"`
	ExpressCompany  string          `json:"express_company" gorm:"size:50"`
	ExpressNo       string          `json:"express_no" gorm:"size:50"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
	PaidAt          *time.Time      `json:"paid_at"`
	ShippedAt       *time.Time      `json:"shipped_at"`
	CompletedAt     *time.Time      `json:"completed_at"`
	Items           []OrderItem     `json:"items" gorm:"foreignKey:OrderID"`
}

// OrderItem 订单商品项实体
type OrderItem struct {
	ID             int64       `json:"id" gorm:"primaryKey"`
	OrderID        string      `json:"order_id" gorm:"size:32"`
	ProductID      int64       `json:"product_id"`
	SKUID          *int64      `json:"sku_id,omitempty"`      // SKU ID（可选）
	SKUCode        *string     `json:"sku_code,omitempty"`    // SKU编码（可选）
	SpecValues     *SpecValues `json:"spec_values,omitempty"` // 规格组合（可选）
	ProductName    string      `json:"product_name" gorm:"size:200"`
	ProductImage   string      `json:"product_image" gorm:"size:500"`
	Price          int64       `json:"price"`           // 商品单价（快照，单位：分）
	Quantity       int         `json:"quantity"`        // 购买数量
	SubtotalAmount int64       `json:"subtotal_amount"` // 小计金额（单位：分）
	CreatedAt      time.Time   `json:"created_at"`
}

// OrderStatus 订单状态常量
const (
	OrderStatusPending   = 0 // 待支付
	OrderStatusPaid      = 1 // 已支付
	OrderStatusShipped   = 2 // 已发货
	OrderStatusCompleted = 3 // 已完成
	OrderStatusCancelled = 4 // 已取消
)