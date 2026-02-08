package dto

import "time"

// 用户相关响应DTO

// UserResponse 用户信息响应
type UserResponse struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	IsAdmin   bool      `json:"is_admin"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserLoginResponse 用户登录响应
type UserLoginResponse struct {
	User  UserResponse `json:"user"`
	Token string       `json:"token"`
}

// UserListResponse 用户列表响应
type UserListResponse struct {
	Users      []UserResponse     `json:"users"`
	Pagination PaginationResponse `json:"pagination"`
}

// 商品相关响应DTO

// ProductResponse 商品信息响应
type ProductResponse struct {
	ID          int64     `json:"id"`
	CategoryID  int64     `json:"category_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       int64     `json:"price"` // 单位：分
	Stock       int       `json:"stock"`
	CoverImage  string    `json:"cover_image"`
	Status      int       `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ProductListResponse 商品列表响应
type ProductListResponse struct {
	Products   []ProductResponse `json:"products"`
	Pagination PaginationResponse `json:"pagination"`
}

// CategoryResponse 商品分类响应
type CategoryResponse struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	ParentID  int64     `json:"parent_id"`
	SortOrder int       `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
}

// 购物车相关响应DTO

// CartItemResponse 购物车商品项响应
type CartItemResponse struct {
	ProductID      int64            `json:"product_id"`
	SKUID          *int64           `json:"sku_id,omitempty"`      // SKU ID（可选）
	SKUCode        *string          `json:"sku_code,omitempty"`    // SKU编码（可选）
	SpecValues     *map[string]string `json:"spec_values,omitempty"` // 规格组合（可选）
	Quantity       int              `json:"quantity"`
	Product        ProductResponse  `json:"product"`
	SubtotalAmount int64            `json:"subtotal_amount"` // 小计金额，单位：分
}

// CartResponse 购物车响应
type CartResponse struct {
	UserID     int64              `json:"user_id"`
	Items      []CartItemResponse `json:"items"`
	TotalCount int                `json:"total_count"`
	TotalPrice int64              `json:"total_price"` // 总金额，单位：分
}

// 订单相关响应DTO

// OrderItemResponse 订单商品项响应
type OrderItemResponse struct {
	ID             int64              `json:"id"`
	ProductID      int64              `json:"product_id"`
	SKUID          *int64             `json:"sku_id,omitempty"`      // SKU ID（可选）
	SKUCode        *string            `json:"sku_code,omitempty"`    // SKU编码（可选）
	SpecValues     *map[string]string `json:"spec_values,omitempty"` // 规格组合（可选）
	ProductName    string             `json:"product_name"`
	ProductImage   string             `json:"product_image"`
	Price          int64              `json:"price"`           // 单价，单位：分
	Quantity       int                `json:"quantity"`
	SubtotalAmount int64              `json:"subtotal_amount"` // 小计，单位：分
}

// AddressSnapshot 地址快照
type AddressSnapshot struct {
	RecipientName string `json:"recipient_name"`
	Phone         string `json:"phone"`
	Province      string `json:"province"`
	City          string `json:"city"`
	District      string `json:"district"`
	Detail        string `json:"detail"`
}

// OrderResponse 订单信息响应
type OrderResponse struct {
	ID              string              `json:"id"`
	UserID          int64               `json:"user_id"`
	TotalAmount     int64               `json:"total_amount"`    // 商品总金额，单位：分
	CouponID        *int64              `json:"coupon_id"`       // 使用的优惠券ID
	CouponDiscount  int64               `json:"coupon_discount"` // 优惠券折扣金额，单位：分
	FinalAmount     int64               `json:"final_amount"`    // 最终支付金额，单位：分
	Status          int                 `json:"status"`
	StatusText      string              `json:"status_text"`
	AddressSnapshot AddressSnapshot     `json:"address_snapshot"`
	ExpressCompany  string              `json:"express_company"`
	ExpressNo       string              `json:"express_no"`
	Items           []OrderItemResponse `json:"items"`
	CreatedAt       time.Time           `json:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at"`
	PaidAt          *time.Time          `json:"paid_at"`
	ShippedAt       *time.Time          `json:"shipped_at"`
	CompletedAt     *time.Time          `json:"completed_at"`
}

// OrderListResponse 订单列表响应
type OrderListResponse struct {
	Orders     []OrderResponse    `json:"orders"`
	Pagination PaginationResponse `json:"pagination"`
}

// 地址相关响应DTO

// AddressResponse 地址信息响应
type AddressResponse struct {
	ID            int64     `json:"id"`
	UserID        int64     `json:"user_id"`
	RecipientName string    `json:"recipient_name"`
	Phone         string    `json:"phone"`
	Province      string    `json:"province"`
	City          string    `json:"city"`
	District      string    `json:"district"`
	Detail        string    `json:"detail"`
	IsDefault     bool      `json:"is_default"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// AddressListResponse 地址列表响应
type AddressListResponse struct {
	Addresses  []AddressResponse  `json:"addresses"`
	Pagination PaginationResponse `json:"pagination"`
}

// 支付相关响应DTO

// PaymentResponse 支付响应
type PaymentResponse struct {
	OrderID       string `json:"order_id"`
	PaymentURL    string `json:"payment_url"`
	PaymentStatus string `json:"payment_status"`
}

// 分页响应结构
type PaginationResponse struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// NewPaginationResponse 创建分页响应
func NewPaginationResponse(page, pageSize int, total int64) PaginationResponse {
	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
	return PaginationResponse{
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
	}
}

// 统计相关响应DTO

// DashboardStatsResponse 仪表盘统计响应
type DashboardStatsResponse struct {
	TotalUsers    int64 `json:"total_users"`
	TotalProducts int64 `json:"total_products"`
	TotalOrders   int64 `json:"total_orders"`
	TotalRevenue  int64 `json:"total_revenue"` // 单位：分
}

// 订单状态映射
var OrderStatusMap = map[int]string{
	0: "待支付",
	1: "已支付",
	2: "已发货",
	3: "已完成",
	4: "已取消",
}

// GetOrderStatusText 获取订单状态文本
func GetOrderStatusText(status int) string {
	if text, exists := OrderStatusMap[status]; exists {
		return text
	}
	return "未知状态"
}

// 商品状态映射
var ProductStatusMap = map[int]string{
	0: "下架",
	1: "上架",
}

// GetProductStatusText 获取商品状态文本
func GetProductStatusText(status int) string {
	if text, exists := ProductStatusMap[status]; exists {
		return text
	}
	return "未知状态"
}

// SKU相关响应DTO

// ProductSpecResponse 商品规格响应
type ProductSpecResponse struct {
	ID        int64     `json:"id"`
	ProductID int64     `json:"product_id"`
	SpecName  string    `json:"spec_name"`
	SpecValue string    `json:"spec_value"`
	SortOrder int       `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ProductSKUResponse SKU响应
type ProductSKUResponse struct {
	ID         int64             `json:"id"`
	ProductID  int64             `json:"product_id"`
	SKUCode    string            `json:"sku_code"`
	SpecValues map[string]string `json:"spec_values"`
	Price      int64             `json:"price"` // 单位：分
	Stock      int               `json:"stock"`
	Image      string            `json:"image"`
	IsActive   bool              `json:"is_active"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
}

// 优惠券相关响应DTO

// CouponResponse 优惠券响应
type CouponResponse struct {
	ID            int64     `json:"id"`
	Name          string    `json:"name"`
	Type          string    `json:"type"`
	DiscountValue float64   `json:"discount_value"`
	MinAmount     float64   `json:"min_amount"`
	MaxDiscount   float64   `json:"max_discount"`
	TotalCount    int       `json:"total_count"`
	UsedCount     int       `json:"used_count"`
	RemainCount   int       `json:"remain_count"` // 剩余数量
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	IsActive      bool      `json:"is_active"`
	IsExpired     bool      `json:"is_expired"` // 是否已过期
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// CouponListResponse 优惠券列表响应
type CouponListResponse struct {
	Coupons    []CouponResponse   `json:"coupons"`
	Pagination PaginationResponse `json:"pagination"`
}

// UserCouponResponse 用户优惠券响应
type UserCouponResponse struct {
	ID         int64           `json:"id"`
	UserID     int64           `json:"user_id"`
	CouponID   int64           `json:"coupon_id"`
	OrderID    *int64          `json:"order_id,omitempty"`
	Status     string          `json:"status"`
	StatusText string          `json:"status_text"`
	ReceivedAt time.Time       `json:"received_at"`
	UsedAt     *time.Time      `json:"used_at,omitempty"`
	ExpiredAt  time.Time       `json:"expired_at"`
	IsExpired  bool            `json:"is_expired"`
	IsUsable   bool            `json:"is_usable"`
	Coupon     *CouponResponse `json:"coupon,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}

// UserCouponListResponse 用户优惠券列表响应
type UserCouponListResponse struct {
	Coupons    []UserCouponResponse `json:"coupons"`
	Pagination PaginationResponse   `json:"pagination"`
}

// 退款相关响应DTO

// RefundResponse 退款信息响应
type RefundResponse struct {
	ID           uint      `json:"id"`
	OrderID      string    `json:"order_id"`
	UserID       int64     `json:"user_id"`
	RefundType   string    `json:"refund_type"`
	RefundTypeText string  `json:"refund_type_text"`
	RefundAmount int64     `json:"refund_amount"` // 退款金额，单位：分
	Reason       string    `json:"reason"`
	Images       []string  `json:"images"`
	Status       string    `json:"status"`
	StatusText   string    `json:"status_text"`
	RejectReason string    `json:"reject_reason,omitempty"`
	ReviewedBy   *int64    `json:"reviewed_by,omitempty"`
	ReviewedAt   *time.Time `json:"reviewed_at,omitempty"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// RefundListResponse 退款列表响应
type RefundListResponse struct {
	Refunds    []RefundResponse   `json:"refunds"`
	Pagination PaginationResponse `json:"pagination"`
}

// 退款类型映射
var RefundTypeMap = map[string]string{
	"refund_only":   "仅退款",
	"return_refund": "退货退款",
}

// GetRefundTypeText 获取退款类型文本
func GetRefundTypeText(refundType string) string {
	if text, exists := RefundTypeMap[refundType]; exists {
		return text
	}
	return "未知类型"
}

// 退款状态映射
var RefundStatusMap = map[string]string{
	"pending":   "待审核",
	"approved":  "已同意",
	"rejected":  "已拒绝",
	"completed": "已退款",
}

// GetRefundStatusText 获取退款状态文本
func GetRefundStatusText(status string) string {
	if text, exists := RefundStatusMap[status]; exists {
		return text
	}
	return "未知状态"
}

// 评价相关响应DTO

// ReviewResponse 评价信息响应
type ReviewResponse struct {
	ID        int64     `json:"id"`
	ProductID int64     `json:"product_id"`
	OrderID   int64     `json:"order_id"`
	UserID    int64     `json:"user_id"`
	Username  string    `json:"username,omitempty"` // 用户昵称（可选）
	Rating    int       `json:"rating"`
	Content   string    `json:"content"`
	Images    []string  `json:"images"`
	LikeCount int       `json:"like_count"`
	IsHidden  bool      `json:"is_hidden"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ReviewListResponse 评价列表响应
type ReviewListResponse struct {
	Reviews    []ReviewResponse   `json:"reviews"`
	Pagination PaginationResponse `json:"pagination"`
}

// ReviewRatingStatsResponse 评价评分统计响应
type ReviewRatingStatsResponse struct {
	AverageRating float64 `json:"average_rating"` // 平均评分
	TotalCount    int64   `json:"total_count"`    // 总评价数
	Rating5Count  int64   `json:"rating_5_count"` // 5星评价数
	Rating4Count  int64   `json:"rating_4_count"` // 4星评价数
	Rating3Count  int64   `json:"rating_3_count"` // 3星评价数
	Rating2Count  int64   `json:"rating_2_count"` // 2星评价数
	Rating1Count  int64   `json:"rating_1_count"` // 1星评价数
}

// 收藏相关响应DTO

// FavoriteItemResponse 收藏商品项响应
type FavoriteItemResponse struct {
	ID         int64           `json:"id"`
	UserID     int64           `json:"user_id"`
	ProductID  int64           `json:"product_id"`
	Product    ProductResponse `json:"product"`
	IsActive   bool            `json:"is_active"`   // 商品是否上架
	CreatedAt  time.Time       `json:"created_at"`
}

// FavoriteListResponse 收藏列表响应
type FavoriteListResponse struct {
	Favorites  []FavoriteItemResponse `json:"favorites"`
	Pagination PaginationResponse     `json:"pagination"`
}

// 通知相关响应DTO

// NotificationResponse 通知信息响应
type NotificationResponse struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Type      string    `json:"type"`
	TypeText  string    `json:"type_text"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Link      string    `json:"link,omitempty"`
	IsRead    bool      `json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NotificationListResponse 通知列表响应
type NotificationListResponse struct {
	Notifications []NotificationResponse `json:"notifications"`
	Pagination    PaginationResponse     `json:"pagination"`
}

// UnreadCountResponse 未读通知数量响应
type UnreadCountResponse struct {
	UnreadCount int64 `json:"unread_count"`
}

// 通知类型映射
var NotificationTypeMap = map[string]string{
	"order":     "订单通知",
	"promotion": "促销通知",
	"system":    "系统通知",
}

// GetNotificationTypeText 获取通知类型文本
func GetNotificationTypeText(notificationType string) string {
	if text, exists := NotificationTypeMap[notificationType]; exists {
		return text
	}
	return "未知类型"
}