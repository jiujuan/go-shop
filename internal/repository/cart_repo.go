package repository

import (
	"context"
	"time"

	"go-shop/internal/entity"
)

// CartRepository 购物车仓库接口
type CartRepository interface {
	// GetCart 获取用户购物车
	GetCart(ctx context.Context, userID int64) (*entity.Cart, error)

	// AddItem 添加商品到购物车
	AddItem(ctx context.Context, userID int64, item entity.CartItem) error

	// UpdateItemQuantity 更新购物车商品数量
	UpdateItemQuantity(ctx context.Context, userID int64, productID int64, quantity int) error

	// RemoveItem 从购物车删除商品
	RemoveItem(ctx context.Context, userID int64, productID int64) error

	// ClearCart 清空用户购物车
	ClearCart(ctx context.Context, userID int64) error

	// GetItemCount 获取购物车商品总数量
	GetItemCount(ctx context.Context, userID int64) (int, error)

	// GetTotalPrice 获取购物车总价格
	GetTotalPrice(ctx context.Context, userID int64) (int64, error)

	// HasItem 检查购物车是否包含指定商品
	HasItem(ctx context.Context, userID int64, productID int64) (bool, error)

	// GetItem 获取购物车中的指定商品
	GetItem(ctx context.Context, userID int64, productID int64) (*entity.CartItem, error)

	// GetItemQuantity 获取购物车中指定商品的数量
	GetItemQuantity(ctx context.Context, userID int64, productID int64) (int, error)

	// SetExpiration 设置购物车过期时间
	SetExpiration(ctx context.Context, userID int64, expiration time.Duration) error

	// GetExpiration 获取购物车过期时间
	GetExpiration(ctx context.Context, userID int64) (time.Duration, error)

	// RefreshExpiration 刷新购物车过期时间
	RefreshExpiration(ctx context.Context, userID int64) error

	// IsEmpty 检查购物车是否为空
	IsEmpty(ctx context.Context, userID int64) (bool, error)

	// MergeCart 合并购物车（用于用户登录后合并临时购物车）
	MergeCart(ctx context.Context, fromUserID, toUserID int64) error

	// BatchAddItems 批量添加商品到购物车
	BatchAddItems(ctx context.Context, userID int64, items []entity.CartItem) error

	// GetCartWithProducts 获取购物车（包含商品详情）
	GetCartWithProducts(ctx context.Context, userID int64) (*entity.Cart, error)

	// ValidateCart 验证购物车（检查商品是否存在、库存是否充足等）
	ValidateCart(ctx context.Context, userID int64) ([]CartValidationError, error)

	// GetActiveUserCarts 获取活跃用户购物车列表（管理员功能）
	GetActiveUserCarts(ctx context.Context, limit int) ([]int64, error)

	// DeleteExpiredCarts 删除过期的购物车
	DeleteExpiredCarts(ctx context.Context) (int64, error)

	// GetCartStatistics 获取购物车统计信息
	GetCartStatistics(ctx context.Context) (*CartStatistics, error)
}

// CartValidationError 购物车验证错误
type CartValidationError struct {
	ProductID int64  `json:"product_id"`
	Error     string `json:"error"`
	ErrorType string `json:"error_type"` // "not_found", "out_of_stock", "unavailable"
}

// CartStatistics 购物车统计信息
type CartStatistics struct {
	TotalActiveCarts    int64   `json:"total_active_carts"`    // 活跃购物车总数
	TotalItems          int64   `json:"total_items"`           // 购物车商品总数
	AverageItemsPerCart float64 `json:"average_items_per_cart"` // 平均每个购物车的商品数
	TotalValue          int64   `json:"total_value"`           // 购物车总价值
	AverageCartValue    float64 `json:"average_cart_value"`    // 平均购物车价值
}