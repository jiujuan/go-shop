package dto

import "time"

// 用户相关请求DTO

// UserRegisterRequest 用户注册请求
type UserRegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6,max=100"`
	Email    string `json:"email" binding:"required,email,max=100"`
}

// UserLoginRequest 用户登录请求
type UserLoginRequest struct {
	Username string `json:"username" binding:"required,max=50" validate:"required,max=50"`
	Password string `json:"password" binding:"required,max=100" validate:"required,max=100"`
}

// UserUpdateRequest 用户信息更新请求
type UserUpdateRequest struct {
	Email string `json:"email" binding:"omitempty,email,max=100" validate:"omitempty,email,max=100"`
}

// 商品相关请求DTO

// ProductCreateRequest 商品创建请求
type ProductCreateRequest struct {
	CategoryID  int64  `json:"category_id" binding:"required,min=1" validate:"required,min=1"`
	Name        string `json:"name" binding:"required,min=1,max=200" validate:"required,min=1,max=200"`
	Description string `json:"description" binding:"max=5000" validate:"max=5000"`
	Price       int64  `json:"price" binding:"required,min=1" validate:"required,min=1"` // 单位：分
	Stock       int    `json:"stock" binding:"required,min=0" validate:"required,min=0"`
	CoverImage  string `json:"cover_image" binding:"max=500" validate:"max=500"`
}

// ProductUpdateRequest 商品更新请求
type ProductUpdateRequest struct {
	CategoryID  *int64  `json:"category_id" binding:"omitempty,min=1" validate:"omitempty,min=1"`
	Name        *string `json:"name" binding:"omitempty,min=1,max=200" validate:"omitempty,min=1,max=200"`
	Description *string `json:"description" binding:"omitempty,max=5000" validate:"omitempty,max=5000"`
	Price       *int64  `json:"price" binding:"omitempty,min=1" validate:"omitempty,min=1"`
	Stock       *int    `json:"stock" binding:"omitempty,min=0" validate:"omitempty,min=0"`
	CoverImage  *string `json:"cover_image" binding:"omitempty,max=500" validate:"omitempty,max=500"`
	Status      *int    `json:"status" binding:"omitempty,oneof=0 1" validate:"omitempty,oneof=0 1"`
}

// ProductListRequest 商品列表查询请求
type ProductListRequest struct {
	CategoryID *int64  `form:"category_id" binding:"omitempty,min=1" validate:"omitempty,min=1"`
	Keyword    *string `form:"keyword" binding:"omitempty,max=100" validate:"omitempty,max=100"`
	SortBy     *string `form:"sort_by" binding:"omitempty,oneof=price created_at" validate:"omitempty,oneof=price created_at"`
	SortOrder  *string `form:"sort_order" binding:"omitempty,oneof=asc desc" validate:"omitempty,oneof=asc desc"`
	Page       int     `form:"page" binding:"omitempty,min=1" validate:"omitempty,min=1"`
	PageSize   int     `form:"page_size" binding:"omitempty,min=1,max=100" validate:"omitempty,min=1,max=100"`
}

// 购物车相关请求DTO

// CartAddItemRequest 添加商品到购物车请求
type CartAddItemRequest struct {
	ProductID int64  `json:"product_id" binding:"required,min=1" validate:"required,min=1"`
	SKUID     *int64 `json:"sku_id" binding:"omitempty,min=1" validate:"omitempty,min=1"` // SKU ID（可选）
	Quantity  int    `json:"quantity" binding:"required,min=1,max=999" validate:"required,min=1,max=999"`
}

// CartUpdateItemRequest 更新购物车商品数量请求
type CartUpdateItemRequest struct {
	ProductID int64  `json:"product_id" binding:"required,min=1" validate:"required,min=1"`
	SKUID     *int64 `json:"sku_id" binding:"omitempty,min=1" validate:"omitempty,min=1"` // SKU ID（可选）
	Quantity  int    `json:"quantity" binding:"required,min=1,max=999" validate:"required,min=1,max=999"`
}

// CartRemoveItemRequest 从购物车移除商品请求
type CartRemoveItemRequest struct {
	ProductID int64  `json:"product_id" binding:"required,min=1" validate:"required,min=1"`
	SKUID     *int64 `json:"sku_id" binding:"omitempty,min=1" validate:"omitempty,min=1"` // SKU ID（可选）
}

// 订单相关请求DTO

// OrderCreateRequest 订单创建请求
type OrderCreateRequest struct {
	AddressID     int64  `json:"address_id" binding:"required,min=1" validate:"required,min=1"`
	UserCouponID  *int64 `json:"user_coupon_id" binding:"omitempty,min=1" validate:"omitempty,min=1"` // 用户优惠券ID（可选）
}

// OrderUpdateStatusRequest 订单状态更新请求
type OrderUpdateStatusRequest struct {
	Status         int     `json:"status" binding:"required,oneof=0 1 2 3 4" validate:"required,oneof=0 1 2 3 4"`
	ExpressCompany *string `json:"express_company" binding:"omitempty,max=50" validate:"omitempty,max=50"`
	ExpressNo      *string `json:"express_no" binding:"omitempty,max=50" validate:"omitempty,max=50"`
}

// OrderListRequest 订单列表查询请求
type OrderListRequest struct {
	Status   *int       `form:"status" binding:"omitempty,oneof=0 1 2 3 4" validate:"omitempty,oneof=0 1 2 3 4"`
	StartAt  *time.Time `form:"start_at" binding:"omitempty" validate:"omitempty"`
	EndAt    *time.Time `form:"end_at" binding:"omitempty" validate:"omitempty"`
	Page     int        `form:"page" binding:"omitempty,min=1" validate:"omitempty,min=1"`
	PageSize int        `form:"page_size" binding:"omitempty,min=1,max=100" validate:"omitempty,min=1,max=100"`
}

// 地址相关请求DTO

// AddressCreateRequest 地址创建请求
type AddressCreateRequest struct {
	RecipientName string `json:"recipient_name" binding:"required,min=1,max=50"`
	Phone         string `json:"phone" binding:"required,min=11,max=11"`
	Province      string `json:"province" binding:"required,min=1,max=50"`
	City          string `json:"city" binding:"required,min=1,max=50"`
	District      string `json:"district" binding:"required,min=1,max=50"`
	Detail        string `json:"detail" binding:"required,min=1,max=200"`
	IsDefault     *bool  `json:"is_default" binding:"omitempty"`
}

// AddressUpdateRequest 地址更新请求
type AddressUpdateRequest struct {
	RecipientName *string `json:"recipient_name" binding:"omitempty,min=1,max=50"`
	Phone         *string `json:"phone" binding:"omitempty,min=11,max=11"`
	Province      *string `json:"province" binding:"omitempty,min=1,max=50"`
	City          *string `json:"city" binding:"omitempty,min=1,max=50"`
	District      *string `json:"district" binding:"omitempty,min=1,max=50"`
	Detail        *string `json:"detail" binding:"omitempty,min=1,max=200"`
	IsDefault     *bool   `json:"is_default" binding:"omitempty"`
}

// 支付相关请求DTO

// PaymentRequest 支付请求
type PaymentRequest struct {
	OrderID     string `json:"order_id" binding:"required,len=32" validate:"required,len=32"`
	PaymentType string `json:"payment_type" binding:"required,oneof=alipay wechat" validate:"required,oneof=alipay wechat"`
}

// 分页请求基础结构
type PaginationRequest struct {
	Page     int `form:"page" binding:"omitempty,min=1" validate:"omitempty,min=1"`
	PageSize int `form:"page_size" binding:"omitempty,min=1,max=100" validate:"omitempty,min=1,max=100"`
}

// GetDefaultPagination 获取默认分页参数
func (p *PaginationRequest) GetDefaultPagination() (page, pageSize int) {
	page = p.Page
	if page <= 0 {
		page = 1
	}
	
	pageSize = p.PageSize
	if pageSize <= 0 {
		pageSize = 20
	} else if pageSize > 100 {
		pageSize = 100
	}
	
	return page, pageSize
}

// GetOffset 获取数据库查询偏移量
func (p *PaginationRequest) GetOffset() int {
	page, pageSize := p.GetDefaultPagination()
	return (page - 1) * pageSize
}

// GetLimit 获取数据库查询限制数量
func (p *PaginationRequest) GetLimit() int {
	_, pageSize := p.GetDefaultPagination()
	return pageSize
}

// SKU相关请求DTO

// ProductSpecCreateRequest 商品规格创建请求
type ProductSpecCreateRequest struct {
	SpecName  string `json:"spec_name" binding:"required,min=1,max=50" validate:"required,min=1,max=50"`
	SpecValue string `json:"spec_value" binding:"required,min=1,max=100" validate:"required,min=1,max=100"`
	SortOrder int    `json:"sort_order" binding:"omitempty,min=0" validate:"omitempty,min=0"`
}

// ProductSKUCreateRequest SKU创建请求
type ProductSKUCreateRequest struct {
	SKUCode    string            `json:"sku_code" binding:"required,min=1,max=50" validate:"required,min=1,max=50"`
	SpecValues map[string]string `json:"spec_values" binding:"required" validate:"required"`
	Price      int64             `json:"price" binding:"required,min=1" validate:"required,min=1"` // 单位：分
	Stock      int               `json:"stock" binding:"required,min=0" validate:"required,min=0"`
	Image      string            `json:"image" binding:"omitempty,max=255" validate:"omitempty,max=255"`
}

// ProductSKUUpdateRequest SKU更新请求
type ProductSKUUpdateRequest struct {
	SKUCode    *string            `json:"sku_code" binding:"omitempty,min=1,max=50" validate:"omitempty,min=1,max=50"`
	SpecValues *map[string]string `json:"spec_values" binding:"omitempty" validate:"omitempty"`
	Price      *int64             `json:"price" binding:"omitempty,min=1" validate:"omitempty,min=1"`
	Stock      *int               `json:"stock" binding:"omitempty,min=0" validate:"omitempty,min=0"`
	Image      *string            `json:"image" binding:"omitempty,max=255" validate:"omitempty,max=255"`
	IsActive   *bool              `json:"is_active" binding:"omitempty" validate:"omitempty"`
}

// 优惠券相关请求DTO

// CouponCreateRequest 优惠券创建请求
type CouponCreateRequest struct {
	Name          string    `json:"name" binding:"required,min=1,max=100" validate:"required,min=1,max=100"`
	Type          string    `json:"type" binding:"required,oneof=discount full_reduction no_threshold" validate:"required,oneof=discount full_reduction no_threshold"`
	DiscountValue float64   `json:"discount_value" binding:"required,gt=0" validate:"required,gt=0"`
	MinAmount     float64   `json:"min_amount" binding:"omitempty,gte=0" validate:"omitempty,gte=0"`
	MaxDiscount   float64   `json:"max_discount" binding:"omitempty,gte=0" validate:"omitempty,gte=0"`
	TotalCount    int       `json:"total_count" binding:"required,gt=0" validate:"required,gt=0"`
	StartTime     time.Time `json:"start_time" binding:"required" validate:"required"`
	EndTime       time.Time `json:"end_time" binding:"required" validate:"required"`
	IsActive      bool      `json:"is_active" binding:"omitempty" validate:"omitempty"`
}

// CouponListRequest 优惠券列表查询请求
type CouponListRequest struct {
	Type     *string `form:"type" binding:"omitempty,oneof=discount full_reduction no_threshold" validate:"omitempty,oneof=discount full_reduction no_threshold"`
	IsActive *bool   `form:"is_active" binding:"omitempty" validate:"omitempty"`
	Page     int     `form:"page" binding:"omitempty,min=1" validate:"omitempty,min=1"`
	PageSize int     `form:"page_size" binding:"omitempty,min=1,max=100" validate:"omitempty,min=1,max=100"`
}

// UserCouponListRequest 用户优惠券列表查询请求
type UserCouponListRequest struct {
	Status   *string `form:"status" binding:"omitempty,oneof=unused used expired" validate:"omitempty,oneof=unused used expired"`
	Page     int     `form:"page" binding:"omitempty,min=1" validate:"omitempty,min=1"`
	PageSize int     `form:"page_size" binding:"omitempty,min=1,max=100" validate:"omitempty,min=1,max=100"`
}

// AvailableCouponRequest 可用优惠券查询请求
type AvailableCouponRequest struct {
	OrderAmount float64 `form:"order_amount" binding:"required,gt=0" validate:"required,gt=0"`
}

// 退款相关请求DTO

// RefundCreateRequest 退款申请创建请求
type RefundCreateRequest struct {
	RefundType string   `json:"refund_type" binding:"required,oneof=refund_only return_refund" validate:"required,oneof=refund_only return_refund"`
	Reason     string   `json:"reason" binding:"required,min=1,max=500" validate:"required,min=1,max=500"`
	Images     []string `json:"images" binding:"omitempty,max=9,dive,url" validate:"omitempty,max=9,dive,url"` // 最多9张图片
}

// RefundListRequest 退款列表查询请求
type RefundListRequest struct {
	Status   *string `form:"status" binding:"omitempty,oneof=pending approved rejected completed" validate:"omitempty,oneof=pending approved rejected completed"`
	Page     int     `form:"page" binding:"omitempty,min=1" validate:"omitempty,min=1"`
	PageSize int     `form:"page_size" binding:"omitempty,min=1,max=100" validate:"omitempty,min=1,max=100"`
}

// RefundRejectRequest 退款拒绝请求
type RefundRejectRequest struct {
	Reason string `json:"reason" binding:"required,min=1,max=500" validate:"required,min=1,max=500"`
}

// AdminRefundListRequest 管理员退款列表查询请求
type AdminRefundListRequest struct {
	UserID    *int64  `form:"user_id" binding:"omitempty,min=1" validate:"omitempty,min=1"`
	OrderID   *string `form:"order_id" binding:"omitempty" validate:"omitempty"`
	Status    *string `form:"status" binding:"omitempty,oneof=pending approved rejected completed" validate:"omitempty,oneof=pending approved rejected completed"`
	StartDate *string `form:"start_date" binding:"omitempty" validate:"omitempty"` // YYYY-MM-DD
	EndDate   *string `form:"end_date" binding:"omitempty" validate:"omitempty"`   // YYYY-MM-DD
	SortBy    *string `form:"sort_by" binding:"omitempty,oneof=created_at refund_amount" validate:"omitempty,oneof=created_at refund_amount"`
	SortOrder *string `form:"sort_order" binding:"omitempty,oneof=asc desc" validate:"omitempty,oneof=asc desc"`
	Page      int     `form:"page" binding:"omitempty,min=1" validate:"omitempty,min=1"`
	PageSize  int     `form:"page_size" binding:"omitempty,min=1,max=100" validate:"omitempty,min=1,max=100"`
}

// 评价相关请求DTO

// ReviewCreateRequest 评价创建请求
type ReviewCreateRequest struct {
	ProductID int64    `json:"product_id" binding:"required,min=1" validate:"required,min=1"`
	Rating    int      `json:"rating" binding:"required,min=1,max=5" validate:"required,min=1,max=5"`
	Content   string   `json:"content" binding:"omitempty,max=1000" validate:"omitempty,max=1000"`
	Images    []string `json:"images" binding:"omitempty,max=9,dive,url" validate:"omitempty,max=9,dive,url"` // 最多9张图片
}

// ReviewListRequest 评价列表查询请求
type ReviewListRequest struct {
	Filter   *string `form:"filter" binding:"omitempty,oneof=all positive medium negative with_img" validate:"omitempty,oneof=all positive medium negative with_img"`
	Page     int     `form:"page" binding:"omitempty,min=1" validate:"omitempty,min=1"`
	PageSize int     `form:"page_size" binding:"omitempty,min=1,max=100" validate:"omitempty,min=1,max=100"`
}

// 短信验证码相关请求DTO

// SMSSendRequest 发送验证码请求
type SMSSendRequest struct {
	Phone string `json:"phone" binding:"required,len=11" validate:"required,len=11"` // 手机号，11位数字
}

// SMSLoginRequest 短信验证码登录请求
type SMSLoginRequest struct {
	Phone string `json:"phone" binding:"required,len=11" validate:"required,len=11"` // 手机号，11位数字
	Code  string `json:"code" binding:"required,len=6" validate:"required,len=6"`    // 验证码，6位数字
}

// PhoneBindRequest 绑定手机号请求
type PhoneBindRequest struct {
	Phone string `json:"phone" binding:"required,len=11" validate:"required,len=11"` // 手机号，11位数字
	Code  string `json:"code" binding:"required,len=6" validate:"required,len=6"`    // 验证码，6位数字
}