package repository

import (
	"context"
	"time"

	"go-shop/internal/entity"
)

// OrderQueryOptions 订单查询选项
type OrderQueryOptions struct {
	UserID    *int64 // 用户ID筛选
	Status    *int   // 状态筛选
	StartDate *string // 开始日期筛选 (YYYY-MM-DD)
	EndDate   *string // 结束日期筛选 (YYYY-MM-DD)
	SortBy    string  // 排序字段 (created_at, total_amount, paid_at)
	SortOrder string  // 排序方向 (asc, desc)
	Offset    int     // 偏移量
	Limit     int     // 限制数量
}

// OrderRepository 订单仓库接口
type OrderRepository interface {
	// Create 创建订单（包含订单项）
	Create(ctx context.Context, order *entity.Order) error

	// GetByID 根据ID获取订单
	GetByID(ctx context.Context, id string) (*entity.Order, error)

	// GetByIDWithItems 根据ID获取订单（包含订单项）
	GetByIDWithItems(ctx context.Context, id string) (*entity.Order, error)

	// Update 更新订单信息
	Update(ctx context.Context, order *entity.Order) error

	// UpdateStatus 更新订单状态
	UpdateStatus(ctx context.Context, orderID string, status int) error

	// UpdatePaymentInfo 更新支付信息
	UpdatePaymentInfo(ctx context.Context, orderID string, paidAt *time.Time) error

	// UpdateShippingInfo 更新发货信息
	UpdateShippingInfo(ctx context.Context, orderID string, expressCompany, expressNo string, shippedAt *time.Time) error

	// UpdateCompletionInfo 更新完成信息
	UpdateCompletionInfo(ctx context.Context, orderID string, completedAt *time.Time) error

	// Delete 删除订单（软删除）
	Delete(ctx context.Context, id string) error

	// List 获取订单列表（支持筛选、排序、分页）
	List(ctx context.Context, options OrderQueryOptions) ([]*entity.Order, int64, error)

	// ListByUser 根据用户ID获取订单列表
	ListByUser(ctx context.Context, userID int64, offset, limit int) ([]*entity.Order, int64, error)

	// ListByStatus 根据状态获取订单列表
	ListByStatus(ctx context.Context, status int, offset, limit int) ([]*entity.Order, int64, error)

	// ListByUserAndStatus 根据用户ID和状态获取订单列表
	ListByUserAndStatus(ctx context.Context, userID int64, status int, offset, limit int) ([]*entity.Order, int64, error)

	// GetUserOrderCount 获取用户订单总数
	GetUserOrderCount(ctx context.Context, userID int64) (int64, error)

	// GetUserOrderCountByStatus 获取用户指定状态的订单数量
	GetUserOrderCountByStatus(ctx context.Context, userID int64, status int) (int64, error)

	// CountOrders 统计订单总数
	CountOrders(ctx context.Context) (int64, error)

	// CountOrdersByStatus 统计指定状态的订单数量
	CountOrdersByStatus(ctx context.Context, status int) (int64, error)

	// CountOrdersByDateRange 统计指定日期范围内的订单数量
	CountOrdersByDateRange(ctx context.Context, startDate, endDate string) (int64, error)

	// GetTotalSales 获取总销售额
	GetTotalSales(ctx context.Context) (int64, error)

	// GetTotalSalesByDateRange 获取指定日期范围内的销售额
	GetTotalSalesByDateRange(ctx context.Context, startDate, endDate string) (int64, error)

	// GetTotalSalesByUser 获取用户总消费金额
	GetTotalSalesByUser(ctx context.Context, userID int64) (int64, error)

	// ExistsByID 检查订单是否存在
	ExistsByID(ctx context.Context, id string) (bool, error)

	// CanCancel 检查订单是否可以取消
	CanCancel(ctx context.Context, orderID string) (bool, error)

	// CanPay 检查订单是否可以支付
	CanPay(ctx context.Context, orderID string) (bool, error)

	// CanShip 检查订单是否可以发货
	CanShip(ctx context.Context, orderID string) (bool, error)

	// CanComplete 检查订单是否可以完成
	CanComplete(ctx context.Context, orderID string) (bool, error)

	// GetPendingOrders 获取待处理订单（超时未支付等）
	GetPendingOrders(ctx context.Context, timeoutMinutes int) ([]*entity.Order, error)

	// GetRecentOrders 获取最近订单
	GetRecentOrders(ctx context.Context, limit int) ([]*entity.Order, error)

	// GetTopSellingProducts 获取热销商品统计
	GetTopSellingProducts(ctx context.Context, limit int) ([]ProductSalesStats, error)
}

// OrderItemRepository 订单项仓库接口
type OrderItemRepository interface {
	// Create 创建订单项
	Create(ctx context.Context, item *entity.OrderItem) error

	// BatchCreate 批量创建订单项
	BatchCreate(ctx context.Context, items []*entity.OrderItem) error

	// GetByOrderID 根据订单ID获取订单项列表
	GetByOrderID(ctx context.Context, orderID string) ([]*entity.OrderItem, error)

	// GetByID 根据ID获取订单项
	GetByID(ctx context.Context, id int64) (*entity.OrderItem, error)

	// Update 更新订单项
	Update(ctx context.Context, item *entity.OrderItem) error

	// Delete 删除订单项
	Delete(ctx context.Context, id int64) error

	// DeleteByOrderID 根据订单ID删除所有订单项
	DeleteByOrderID(ctx context.Context, orderID string) error

	// CountByOrderID 统计订单项数量
	CountByOrderID(ctx context.Context, orderID string) (int64, error)

	// GetOrderTotal 计算订单总金额
	GetOrderTotal(ctx context.Context, orderID string) (int64, error)
}

// ProductSalesStats 商品销售统计
type ProductSalesStats struct {
	ProductID    int64  `json:"product_id"`
	ProductName  string `json:"product_name"`
	TotalSold    int64  `json:"total_sold"`    // 总销售数量
	TotalRevenue int64  `json:"total_revenue"` // 总销售额
}