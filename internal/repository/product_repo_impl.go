package repository

import (
	"context"
	"errors"
	"fmt"

	"go-shop/internal/entity"

	"gorm.io/gorm"
)

// productRepositoryImpl 商品仓库GORM实现
type productRepositoryImpl struct {
	db *gorm.DB
}

// NewProductRepository 创建商品仓库实例
func NewProductRepository(db *gorm.DB) ProductRepository {
	return &productRepositoryImpl{
		db: db,
	}
}

// Create 创建商品
func (r *productRepositoryImpl) Create(ctx context.Context, product *entity.Product) error {
	if product == nil {
		return errors.New("product cannot be nil")
	}

	// 验证必填字段
	if product.Name == "" {
		return errors.New("product name is required")
	}
	if product.CategoryID <= 0 {
		return entity.ErrInvalidCategoryID
	}
	if product.Price <= 0 {
		return entity.ErrInvalidPrice
	}
	if product.Stock < 0 {
		return entity.ErrInvalidStock
	}

	// 检查商品名称唯一性
	exists, err := r.ExistsByName(ctx, product.Name)
	if err != nil {
		return err
	}
	if exists {
		return entity.ErrProductAlreadyExists
	}

	// 创建商品
	if err := r.db.WithContext(ctx).Create(product).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return entity.ErrProductAlreadyExists
		}
		return err
	}

	return nil
}

// GetByID 根据ID获取商品
func (r *productRepositoryImpl) GetByID(ctx context.Context, id int64) (*entity.Product, error) {
	if id <= 0 {
		return nil, entity.ErrInvalidProductID
	}

	var product entity.Product
	if err := r.db.WithContext(ctx).First(&product, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entity.ErrProductNotFound
		}
		return nil, err
	}

	return &product, nil
}

// GetByIDWithCategory 根据ID获取商品（包含分类信息）
func (r *productRepositoryImpl) GetByIDWithCategory(ctx context.Context, id int64) (*entity.Product, *entity.Category, error) {
	if id <= 0 {
		return nil, nil, entity.ErrInvalidProductID
	}

	var product entity.Product
	var category entity.Category

	// 获取商品信息
	if err := r.db.WithContext(ctx).First(&product, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, entity.ErrProductNotFound
		}
		return nil, nil, err
	}

	// 获取分类信息
	if err := r.db.WithContext(ctx).First(&category, product.CategoryID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 商品存在但分类不存在，返回商品但分类为nil
			return &product, nil, nil
		}
		return nil, nil, err
	}

	return &product, &category, nil
}

// Update 更新商品信息
func (r *productRepositoryImpl) Update(ctx context.Context, product *entity.Product) error {
	if product == nil {
		return errors.New("product cannot be nil")
	}
	if product.ID <= 0 {
		return entity.ErrInvalidProductID
	}

	// 检查商品是否存在
	_, err := r.GetByID(ctx, product.ID)
	if err != nil {
		return err
	}

	// 检查商品名称唯一性（排除当前商品）
	if product.Name != "" {
		exists, err := r.ExistsByNameExcludeID(ctx, product.Name, product.ID)
		if err != nil {
			return err
		}
		if exists {
			return entity.ErrProductAlreadyExists
		}
	}

	// 验证字段
	if product.CategoryID > 0 {
		// 这里可以添加分类存在性检查
	}
	if product.Price > 0 && product.Price <= 0 {
		return entity.ErrInvalidPrice
	}
	if product.Stock < 0 {
		return entity.ErrInvalidStock
	}

	// 更新商品信息
	updates := map[string]interface{}{}
	
	if product.Name != "" {
		updates["name"] = product.Name
	}
	if product.CategoryID > 0 {
		updates["category_id"] = product.CategoryID
	}
	if product.Description != "" {
		updates["description"] = product.Description
	}
	if product.Price > 0 {
		updates["price"] = product.Price
	}
	if product.Stock >= 0 {
		updates["stock"] = product.Stock
	}
	if product.CoverImage != "" {
		updates["cover_image"] = product.CoverImage
	}
	if product.Status >= 0 {
		updates["status"] = product.Status
	}

	if len(updates) == 0 {
		return nil // 没有需要更新的字段
	}

	if err := r.db.WithContext(ctx).Model(&entity.Product{}).Where("id = ?", product.ID).Updates(updates).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return entity.ErrProductAlreadyExists
		}
		return err
	}

	return nil
}

// UpdateStock 更新商品库存
func (r *productRepositoryImpl) UpdateStock(ctx context.Context, productID int64, stock int) error {
	if productID <= 0 {
		return entity.ErrInvalidProductID
	}
	if stock < 0 {
		return entity.ErrInvalidStock
	}

	// 检查商品是否存在
	_, err := r.GetByID(ctx, productID)
	if err != nil {
		return err
	}

	// 更新库存
	if err := r.db.WithContext(ctx).Model(&entity.Product{}).Where("id = ?", productID).Update("stock", stock).Error; err != nil {
		return err
	}

	return nil
}

// UpdateStatus 更新商品状态
func (r *productRepositoryImpl) UpdateStatus(ctx context.Context, productID int64, status int) error {
	if productID <= 0 {
		return entity.ErrInvalidProductID
	}
	if status < 0 || status > 1 {
		return errors.New("invalid product status")
	}

	// 检查商品是否存在
	_, err := r.GetByID(ctx, productID)
	if err != nil {
		return err
	}

	// 更新状态
	if err := r.db.WithContext(ctx).Model(&entity.Product{}).Where("id = ?", productID).Update("status", status).Error; err != nil {
		return err
	}

	return nil
}

// Delete 删除商品（软删除）
func (r *productRepositoryImpl) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return entity.ErrInvalidProductID
	}

	// 检查商品是否存在
	_, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// 软删除商品
	if err := r.db.WithContext(ctx).Delete(&entity.Product{}, id).Error; err != nil {
		return err
	}

	return nil
}

// List 获取商品列表（支持筛选、排序、分页）
func (r *productRepositoryImpl) List(ctx context.Context, options ProductQueryOptions) ([]*entity.Product, int64, error) {
	query := r.db.WithContext(ctx).Model(&entity.Product{})

	// 应用筛选条件
	if options.CategoryID != nil {
		query = query.Where("category_id = ?", *options.CategoryID)
	}
	if options.Status != nil {
		query = query.Where("status = ?", *options.Status)
	}
	if options.MinPrice != nil {
		query = query.Where("price >= ?", *options.MinPrice)
	}
	if options.MaxPrice != nil {
		query = query.Where("price <= ?", *options.MaxPrice)
	}
	if options.Keyword != nil && *options.Keyword != "" {
		keyword := "%" + *options.Keyword + "%"
		query = query.Where("name LIKE ? OR description LIKE ?", keyword, keyword)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 应用排序
	orderClause := "created_at DESC" // 默认排序
	if options.SortBy != "" {
		validSortFields := map[string]bool{
			"price":      true,
			"created_at": true,
			"stock":      true,
			"name":       true,
		}
		if validSortFields[options.SortBy] {
			order := "DESC"
			if options.SortOrder == "asc" {
				order = "ASC"
			}
			orderClause = fmt.Sprintf("%s %s", options.SortBy, order)
		}
	}

	// 应用分页和排序
	var products []*entity.Product
	if err := query.Order(orderClause).Offset(options.Offset).Limit(options.Limit).Find(&products).Error; err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

// ListByCategory 根据分类获取商品列表
func (r *productRepositoryImpl) ListByCategory(ctx context.Context, categoryID int64, offset, limit int) ([]*entity.Product, int64, error) {
	if categoryID <= 0 {
		return nil, 0, entity.ErrInvalidCategoryID
	}

	options := ProductQueryOptions{
		CategoryID: &categoryID,
		Status:     intPtr(1), // 只返回上架商品
		Offset:     offset,
		Limit:      limit,
	}

	return r.List(ctx, options)
}

// ListByStatus 根据状态获取商品列表
func (r *productRepositoryImpl) ListByStatus(ctx context.Context, status int, offset, limit int) ([]*entity.Product, int64, error) {
	options := ProductQueryOptions{
		Status: &status,
		Offset: offset,
		Limit:  limit,
	}

	return r.List(ctx, options)
}

// Search 搜索商品（按名称和描述）
func (r *productRepositoryImpl) Search(ctx context.Context, keyword string, offset, limit int) ([]*entity.Product, int64, error) {
	if keyword == "" {
		return nil, 0, errors.New("search keyword cannot be empty")
	}

	options := ProductQueryOptions{
		Keyword: &keyword,
		Status:  intPtr(1), // 只搜索上架商品
		Offset:  offset,
		Limit:   limit,
	}

	return r.List(ctx, options)
}

// GetTopProducts 获取热门商品（按销量或其他指标）
func (r *productRepositoryImpl) GetTopProducts(ctx context.Context, limit int) ([]*entity.Product, error) {
	if limit <= 0 {
		limit = 10
	}

	var products []*entity.Product
	// 这里简单按创建时间排序，实际应该按销量等指标排序
	if err := r.db.WithContext(ctx).
		Where("status = ?", 1).
		Order("created_at DESC").
		Limit(limit).
		Find(&products).Error; err != nil {
		return nil, err
	}

	return products, nil
}

// GetLatestProducts 获取最新商品
func (r *productRepositoryImpl) GetLatestProducts(ctx context.Context, limit int) ([]*entity.Product, error) {
	if limit <= 0 {
		limit = 10
	}

	var products []*entity.Product
	if err := r.db.WithContext(ctx).
		Where("status = ?", 1).
		Order("created_at DESC").
		Limit(limit).
		Find(&products).Error; err != nil {
		return nil, err
	}

	return products, nil
}

// CheckStock 检查商品库存是否充足
func (r *productRepositoryImpl) CheckStock(ctx context.Context, productID int64, quantity int) (bool, error) {
	if productID <= 0 {
		return false, entity.ErrInvalidProductID
	}
	if quantity <= 0 {
		return false, entity.ErrInvalidStock
	}

	product, err := r.GetByID(ctx, productID)
	if err != nil {
		return false, err
	}

	return product.Stock >= quantity, nil
}

// DecrementStock 减少商品库存（原子操作）
func (r *productRepositoryImpl) DecrementStock(ctx context.Context, productID int64, quantity int) error {
	if productID <= 0 {
		return entity.ErrInvalidProductID
	}
	if quantity <= 0 {
		return entity.ErrInvalidStock
	}

	// 使用原子操作减少库存，确保库存不会变为负数
	result := r.db.WithContext(ctx).Model(&entity.Product{}).
		Where("id = ? AND stock >= ?", productID, quantity).
		Update("stock", gorm.Expr("stock - ?", quantity))

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		// 检查商品是否存在
		_, err := r.GetByID(ctx, productID)
		if err != nil {
			return err
		}
		// 商品存在但库存不足
		return entity.ErrProductOutOfStock
	}

	return nil
}

// IncrementStock 增加商品库存（原子操作）
func (r *productRepositoryImpl) IncrementStock(ctx context.Context, productID int64, quantity int) error {
	if productID <= 0 {
		return entity.ErrInvalidProductID
	}
	if quantity <= 0 {
		return entity.ErrInvalidStock
	}

	// 检查商品是否存在
	_, err := r.GetByID(ctx, productID)
	if err != nil {
		return err
	}

	// 增加库存
	if err := r.db.WithContext(ctx).Model(&entity.Product{}).
		Where("id = ?", productID).
		Update("stock", gorm.Expr("stock + ?", quantity)).Error; err != nil {
		return err
	}

	return nil
}

// BatchUpdateStock 批量更新商品库存
func (r *productRepositoryImpl) BatchUpdateStock(ctx context.Context, updates []StockUpdate) error {
	if len(updates) == 0 {
		return nil
	}

	// 使用事务处理批量更新
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, update := range updates {
			if update.ProductID <= 0 {
				return entity.ErrInvalidProductID
			}

			if update.Quantity > 0 {
				// 增加库存
				if err := tx.Model(&entity.Product{}).
					Where("id = ?", update.ProductID).
					Update("stock", gorm.Expr("stock + ?", update.Quantity)).Error; err != nil {
					return err
				}
			} else if update.Quantity < 0 {
				// 减少库存
				quantity := -update.Quantity
				result := tx.Model(&entity.Product{}).
					Where("id = ? AND stock >= ?", update.ProductID, quantity).
					Update("stock", gorm.Expr("stock - ?", quantity))

				if result.Error != nil {
					return result.Error
				}
				if result.RowsAffected == 0 {
					return entity.ErrProductOutOfStock
				}
			}
		}
		return nil
	})
}

// CountProducts 统计商品总数
func (r *productRepositoryImpl) CountProducts(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entity.Product{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// CountProductsByCategory 统计分类下的商品数量
func (r *productRepositoryImpl) CountProductsByCategory(ctx context.Context, categoryID int64) (int64, error) {
	if categoryID <= 0 {
		return 0, entity.ErrInvalidCategoryID
	}

	var count int64
	if err := r.db.WithContext(ctx).Model(&entity.Product{}).
		Where("category_id = ?", categoryID).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// CountProductsByStatus 统计指定状态的商品数量
func (r *productRepositoryImpl) CountProductsByStatus(ctx context.Context, status int) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entity.Product{}).
		Where("status = ?", status).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// GetLowStockProducts 获取低库存商品
func (r *productRepositoryImpl) GetLowStockProducts(ctx context.Context, threshold int) ([]*entity.Product, error) {
	if threshold < 0 {
		threshold = 10 // 默认阈值
	}

	var products []*entity.Product
	if err := r.db.WithContext(ctx).
		Where("stock <= ? AND status = ?", threshold, 1).
		Order("stock ASC").
		Find(&products).Error; err != nil {
		return nil, err
	}

	return products, nil
}

// ExistsByName 检查商品名称是否存在
func (r *productRepositoryImpl) ExistsByName(ctx context.Context, name string) (bool, error) {
	if name == "" {
		return false, nil
	}

	var count int64
	if err := r.db.WithContext(ctx).Model(&entity.Product{}).
		Where("name = ?", name).
		Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

// ExistsByNameExcludeID 检查商品名称是否存在（排除指定ID）
func (r *productRepositoryImpl) ExistsByNameExcludeID(ctx context.Context, name string, excludeID int64) (bool, error) {
	if name == "" {
		return false, nil
	}

	var count int64
	if err := r.db.WithContext(ctx).Model(&entity.Product{}).
		Where("name = ? AND id != ?", name, excludeID).
		Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

// 辅助函数
func intPtr(i int) *int {
	return &i
}

// categoryRepositoryImpl 分类仓库GORM实现
type categoryRepositoryImpl struct {
	db *gorm.DB
}

// NewCategoryRepository 创建分类仓库实例
func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &categoryRepositoryImpl{
		db: db,
	}
}

// Create 创建分类
func (r *categoryRepositoryImpl) Create(ctx context.Context, category *entity.Category) error {
	if category == nil {
		return errors.New("category cannot be nil")
	}
	if category.Name == "" {
		return errors.New("category name is required")
	}

	// 检查分类名称唯一性
	exists, err := r.ExistsByName(ctx, category.Name)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("category already exists")
	}

	// 创建分类
	if err := r.db.WithContext(ctx).Create(category).Error; err != nil {
		return err
	}

	return nil
}

// GetByID 根据ID获取分类
func (r *categoryRepositoryImpl) GetByID(ctx context.Context, id int64) (*entity.Category, error) {
	if id <= 0 {
		return nil, entity.ErrInvalidCategoryID
	}

	var category entity.Category
	if err := r.db.WithContext(ctx).First(&category, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("category not found")
		}
		return nil, err
	}

	return &category, nil
}

// Update 更新分类信息
func (r *categoryRepositoryImpl) Update(ctx context.Context, category *entity.Category) error {
	if category == nil {
		return errors.New("category cannot be nil")
	}
	if category.ID <= 0 {
		return entity.ErrInvalidCategoryID
	}

	// 检查分类是否存在
	_, err := r.GetByID(ctx, category.ID)
	if err != nil {
		return err
	}

	// 检查分类名称唯一性（排除当前分类）
	if category.Name != "" {
		exists, err := r.ExistsByNameExcludeID(ctx, category.Name, category.ID)
		if err != nil {
			return err
		}
		if exists {
			return errors.New("category name already exists")
		}
	}

	// 更新分类信息
	if err := r.db.WithContext(ctx).Save(category).Error; err != nil {
		return err
	}

	return nil
}

// Delete 删除分类
func (r *categoryRepositoryImpl) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return entity.ErrInvalidCategoryID
	}

	// 检查分类是否存在
	_, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// 检查是否有商品使用此分类
	var count int64
	err = r.db.WithContext(ctx).Model(&entity.Product{}).Where("category_id = ?", id).Count(&count).Error
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("cannot delete category with existing products")
	}

	// 删除分类
	if err := r.db.WithContext(ctx).Delete(&entity.Category{}, id).Error; err != nil {
		return err
	}

	return nil
}

// List 获取分类列表
func (r *categoryRepositoryImpl) List(ctx context.Context) ([]*entity.Category, error) {
	var categories []*entity.Category
	if err := r.db.WithContext(ctx).Order("sort_order ASC, created_at ASC").Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

// ListByParentID 根据父分类ID获取子分类列表
func (r *categoryRepositoryImpl) ListByParentID(ctx context.Context, parentID int64) ([]*entity.Category, error) {
	var categories []*entity.Category
	if err := r.db.WithContext(ctx).
		Where("parent_id = ?", parentID).
		Order("sort_order ASC, created_at ASC").
		Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

// GetTopLevelCategories 获取顶级分类列表
func (r *categoryRepositoryImpl) GetTopLevelCategories(ctx context.Context) ([]*entity.Category, error) {
	return r.ListByParentID(ctx, 0)
}

// ExistsByName 检查分类名称是否存在
func (r *categoryRepositoryImpl) ExistsByName(ctx context.Context, name string) (bool, error) {
	if name == "" {
		return false, nil
	}

	var count int64
	if err := r.db.WithContext(ctx).Model(&entity.Category{}).
		Where("name = ?", name).
		Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

// ExistsByNameExcludeID 检查分类名称是否存在（排除指定ID）
func (r *categoryRepositoryImpl) ExistsByNameExcludeID(ctx context.Context, name string, excludeID int64) (bool, error) {
	if name == "" {
		return false, nil
	}

	var count int64
	if err := r.db.WithContext(ctx).Model(&entity.Category{}).
		Where("name = ? AND id != ?", name, excludeID).
		Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

// CountCategories 统计分类总数
func (r *categoryRepositoryImpl) CountCategories(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entity.Category{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}