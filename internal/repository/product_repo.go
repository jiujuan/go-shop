package repository

import (
	"context"

	"go-shop/internal/entity"
)

// ProductQueryOptions 商品查询选项
type ProductQueryOptions struct {
	CategoryID *int64  // 分类ID筛选
	Keyword    *string // 关键词搜索
	Status     *int    // 状态筛选 (0-下架, 1-上架)
	MinPrice   *int64  // 最低价格筛选
	MaxPrice   *int64  // 最高价格筛选
	SortBy     string  // 排序字段 (price, created_at, stock)
	SortOrder  string  // 排序方向 (asc, desc)
	Offset     int     // 偏移量
	Limit      int     // 限制数量
}

// ProductRepository 商品仓库接口
type ProductRepository interface {
	// Create 创建商品
	Create(ctx context.Context, product *entity.Product) error

	// GetByID 根据ID获取商品
	GetByID(ctx context.Context, id int64) (*entity.Product, error)

	// GetByIDWithCategory 根据ID获取商品（包含分类信息）
	GetByIDWithCategory(ctx context.Context, id int64) (*entity.Product, *entity.Category, error)

	// Update 更新商品信息
	Update(ctx context.Context, product *entity.Product) error

	// UpdateStock 更新商品库存
	UpdateStock(ctx context.Context, productID int64, stock int) error

	// UpdateStatus 更新商品状态
	UpdateStatus(ctx context.Context, productID int64, status int) error

	// Delete 删除商品（软删除）
	Delete(ctx context.Context, id int64) error

	// List 获取商品列表（支持筛选、排序、分页）
	List(ctx context.Context, options ProductQueryOptions) ([]*entity.Product, int64, error)

	// ListByCategory 根据分类获取商品列表
	ListByCategory(ctx context.Context, categoryID int64, offset, limit int) ([]*entity.Product, int64, error)

	// ListByStatus 根据状态获取商品列表
	ListByStatus(ctx context.Context, status int, offset, limit int) ([]*entity.Product, int64, error)

	// Search 搜索商品（按名称和描述）
	Search(ctx context.Context, keyword string, offset, limit int) ([]*entity.Product, int64, error)

	// GetTopProducts 获取热门商品（按销量或其他指标）
	GetTopProducts(ctx context.Context, limit int) ([]*entity.Product, error)

	// GetLatestProducts 获取最新商品
	GetLatestProducts(ctx context.Context, limit int) ([]*entity.Product, error)

	// CheckStock 检查商品库存是否充足
	CheckStock(ctx context.Context, productID int64, quantity int) (bool, error)

	// DecrementStock 减少商品库存（原子操作）
	DecrementStock(ctx context.Context, productID int64, quantity int) error

	// IncrementStock 增加商品库存（原子操作）
	IncrementStock(ctx context.Context, productID int64, quantity int) error

	// BatchUpdateStock 批量更新商品库存
	BatchUpdateStock(ctx context.Context, updates []StockUpdate) error

	// CountProducts 统计商品总数
	CountProducts(ctx context.Context) (int64, error)

	// CountProductsByCategory 统计分类下的商品数量
	CountProductsByCategory(ctx context.Context, categoryID int64) (int64, error)

	// CountProductsByStatus 统计指定状态的商品数量
	CountProductsByStatus(ctx context.Context, status int) (int64, error)

	// GetLowStockProducts 获取低库存商品
	GetLowStockProducts(ctx context.Context, threshold int) ([]*entity.Product, error)

	// ExistsByName 检查商品名称是否存在
	ExistsByName(ctx context.Context, name string) (bool, error)

	// ExistsByNameExcludeID 检查商品名称是否存在（排除指定ID）
	ExistsByNameExcludeID(ctx context.Context, name string, excludeID int64) (bool, error)
}

// StockUpdate 库存更新结构
type StockUpdate struct {
	ProductID int64 // 商品ID
	Quantity  int   // 更新数量（正数增加，负数减少）
}

// CategoryRepository 分类仓库接口
type CategoryRepository interface {
	// Create 创建分类
	Create(ctx context.Context, category *entity.Category) error

	// GetByID 根据ID获取分类
	GetByID(ctx context.Context, id int64) (*entity.Category, error)

	// Update 更新分类信息
	Update(ctx context.Context, category *entity.Category) error

	// Delete 删除分类
	Delete(ctx context.Context, id int64) error

	// List 获取分类列表
	List(ctx context.Context) ([]*entity.Category, error)

	// ListByParentID 根据父分类ID获取子分类列表
	ListByParentID(ctx context.Context, parentID int64) ([]*entity.Category, error)

	// GetTopLevelCategories 获取顶级分类列表
	GetTopLevelCategories(ctx context.Context) ([]*entity.Category, error)

	// ExistsByName 检查分类名称是否存在
	ExistsByName(ctx context.Context, name string) (bool, error)

	// ExistsByNameExcludeID 检查分类名称是否存在（排除指定ID）
	ExistsByNameExcludeID(ctx context.Context, name string, excludeID int64) (bool, error)

	// CountCategories 统计分类总数
	CountCategories(ctx context.Context) (int64, error)
}