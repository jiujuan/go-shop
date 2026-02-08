package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"go-shop/internal/entity"

	"github.com/redis/go-redis/v9"
)

// cartRepositoryImpl 购物车仓库Redis实现
type cartRepositoryImpl struct {
	redis       *redis.Client
	productRepo ProductRepository
	keyPrefix   string
	defaultTTL  time.Duration
}

// NewCartRepository 创建购物车仓库实例
func NewCartRepository(redisClient *redis.Client, productRepo ProductRepository) CartRepository {
	return &cartRepositoryImpl{
		redis:       redisClient,
		productRepo: productRepo,
		keyPrefix:   "cart:",
		defaultTTL:  24 * time.Hour, // 默认24小时过期
	}
}

// getCartKey 获取购物车Redis键
func (r *cartRepositoryImpl) getCartKey(userID int64) string {
	return fmt.Sprintf("%s%d", r.keyPrefix, userID)
}

// getCartMetaKey 获取购物车元数据Redis键
func (r *cartRepositoryImpl) getCartMetaKey(userID int64) string {
	return fmt.Sprintf("%smeta:%d", r.keyPrefix, userID)
}

// GetCart 获取用户购物车
func (r *cartRepositoryImpl) GetCart(ctx context.Context, userID int64) (*entity.Cart, error) {
	if userID <= 0 {
		return nil, entity.ErrInvalidUserID
	}

	cartKey := r.getCartKey(userID)
	
	// 获取购物车数据
	cartData, err := r.redis.Get(ctx, cartKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			// 购物车不存在，返回空购物车
			return &entity.Cart{
				UserID:     userID,
				Items:      []entity.CartItem{},
				TotalCount: 0,
				TotalPrice: 0,
			}, nil
		}
		return nil, err
	}

	// 解析购物车数据
	var cart entity.Cart
	if err := json.Unmarshal([]byte(cartData), &cart); err != nil {
		return nil, err
	}

	// 刷新过期时间
	r.RefreshExpiration(ctx, userID)

	return &cart, nil
}

// saveCart 保存购物车到Redis
func (r *cartRepositoryImpl) saveCart(ctx context.Context, cart *entity.Cart) error {
	cartKey := r.getCartKey(cart.UserID)
	
	// 序列化购物车数据
	cartData, err := json.Marshal(cart)
	if err != nil {
		return err
	}

	// 保存到Redis并设置过期时间
	if err := r.redis.Set(ctx, cartKey, cartData, r.defaultTTL).Err(); err != nil {
		return err
	}

	// 更新元数据
	metaKey := r.getCartMetaKey(cart.UserID)
	metaData := map[string]interface{}{
		"last_updated": time.Now().Unix(),
		"item_count":   cart.TotalCount,
		"total_price":  cart.TotalPrice,
	}
	
	metaJSON, _ := json.Marshal(metaData)
	r.redis.Set(ctx, metaKey, metaJSON, r.defaultTTL)

	return nil
}

// AddItem 添加商品到购物车
func (r *cartRepositoryImpl) AddItem(ctx context.Context, userID int64, item entity.CartItem) error {
	if userID <= 0 {
		return entity.ErrInvalidUserID
	}
	if item.ProductID <= 0 {
		return entity.ErrInvalidProductID
	}
	if item.Quantity <= 0 {
		return errors.New("quantity must be greater than 0")
	}

	// 验证商品是否存在
	product, err := r.productRepo.GetByID(ctx, item.ProductID)
	if err != nil {
		return err
	}

	// 检查商品是否上架
	if product.Status != 1 {
		return entity.ErrProductNotAvailable
	}

	// 检查库存
	if product.Stock < item.Quantity {
		return entity.ErrProductOutOfStock
	}

	// 获取当前购物车
	cart, err := r.GetCart(ctx, userID)
	if err != nil {
		return err
	}

	// 设置商品信息
	item.UserID = userID
	item.Product = product

	// 添加商品到购物车
	cart.AddItem(item)

	// 保存购物车
	return r.saveCart(ctx, cart)
}

// UpdateItemQuantity 更新购物车商品数量
func (r *cartRepositoryImpl) UpdateItemQuantity(ctx context.Context, userID int64, productID int64, quantity int) error {
	if userID <= 0 {
		return entity.ErrInvalidUserID
	}
	if productID <= 0 {
		return entity.ErrInvalidProductID
	}

	// 获取当前购物车
	cart, err := r.GetCart(ctx, userID)
	if err != nil {
		return err
	}

	// 检查商品是否在购物车中
	found := false
	for _, item := range cart.Items {
		if item.ProductID == productID {
			found = true
			break
		}
	}
	if !found {
		return entity.ErrCartItemNotFound
	}

	// 如果数量大于0，检查库存
	if quantity > 0 {
		product, err := r.productRepo.GetByID(ctx, productID)
		if err != nil {
			return err
		}
		if product.Stock < quantity {
			return entity.ErrProductOutOfStock
		}
	}

	// 更新商品数量（不指定SKU ID，因为这个方法不支持SKU）
	cart.UpdateItemQuantity(productID, nil, quantity)

	// 保存购物车
	return r.saveCart(ctx, cart)
}

// RemoveItem 从购物车删除商品
func (r *cartRepositoryImpl) RemoveItem(ctx context.Context, userID int64, productID int64) error {
	if userID <= 0 {
		return entity.ErrInvalidUserID
	}
	if productID <= 0 {
		return entity.ErrInvalidProductID
	}

	// 获取当前购物车
	cart, err := r.GetCart(ctx, userID)
	if err != nil {
		return err
	}

	// 检查商品是否在购物车中
	found := false
	for _, item := range cart.Items {
		if item.ProductID == productID {
			found = true
			break
		}
	}
	if !found {
		return entity.ErrCartItemNotFound
	}

	// 删除商品（不指定SKU ID，因为这个方法不支持SKU）
	cart.RemoveItem(productID, nil)

	// 保存购物车
	return r.saveCart(ctx, cart)
}

// ClearCart 清空用户购物车
func (r *cartRepositoryImpl) ClearCart(ctx context.Context, userID int64) error {
	if userID <= 0 {
		return entity.ErrInvalidUserID
	}

	cartKey := r.getCartKey(userID)
	metaKey := r.getCartMetaKey(userID)

	// 删除购物车数据和元数据
	pipe := r.redis.Pipeline()
	pipe.Del(ctx, cartKey)
	pipe.Del(ctx, metaKey)
	_, err := pipe.Exec(ctx)

	return err
}

// GetItemCount 获取购物车商品总数量
func (r *cartRepositoryImpl) GetItemCount(ctx context.Context, userID int64) (int, error) {
	cart, err := r.GetCart(ctx, userID)
	if err != nil {
		return 0, err
	}
	return cart.TotalCount, nil
}

// GetTotalPrice 获取购物车总价格
func (r *cartRepositoryImpl) GetTotalPrice(ctx context.Context, userID int64) (int64, error) {
	cart, err := r.GetCart(ctx, userID)
	if err != nil {
		return 0, err
	}
	return cart.TotalPrice, nil
}

// HasItem 检查购物车是否包含指定商品
func (r *cartRepositoryImpl) HasItem(ctx context.Context, userID int64, productID int64) (bool, error) {
	cart, err := r.GetCart(ctx, userID)
	if err != nil {
		return false, err
	}

	for _, item := range cart.Items {
		if item.ProductID == productID {
			return true, nil
		}
	}
	return false, nil
}

// GetItem 获取购物车中的指定商品
func (r *cartRepositoryImpl) GetItem(ctx context.Context, userID int64, productID int64) (*entity.CartItem, error) {
	cart, err := r.GetCart(ctx, userID)
	if err != nil {
		return nil, err
	}

	for _, item := range cart.Items {
		if item.ProductID == productID {
			return &item, nil
		}
	}
	return nil, entity.ErrCartItemNotFound
}

// GetItemQuantity 获取购物车中指定商品的数量
func (r *cartRepositoryImpl) GetItemQuantity(ctx context.Context, userID int64, productID int64) (int, error) {
	item, err := r.GetItem(ctx, userID, productID)
	if err != nil {
		if errors.Is(err, entity.ErrCartItemNotFound) {
			return 0, nil
		}
		return 0, err
	}
	return item.Quantity, nil
}

// SetExpiration 设置购物车过期时间
func (r *cartRepositoryImpl) SetExpiration(ctx context.Context, userID int64, expiration time.Duration) error {
	if userID <= 0 {
		return entity.ErrInvalidUserID
	}

	cartKey := r.getCartKey(userID)
	metaKey := r.getCartMetaKey(userID)

	// 设置过期时间
	pipe := r.redis.Pipeline()
	pipe.Expire(ctx, cartKey, expiration)
	pipe.Expire(ctx, metaKey, expiration)
	_, err := pipe.Exec(ctx)

	return err
}

// GetExpiration 获取购物车过期时间
func (r *cartRepositoryImpl) GetExpiration(ctx context.Context, userID int64) (time.Duration, error) {
	if userID <= 0 {
		return 0, entity.ErrInvalidUserID
	}

	cartKey := r.getCartKey(userID)
	return r.redis.TTL(ctx, cartKey).Result()
}

// RefreshExpiration 刷新购物车过期时间
func (r *cartRepositoryImpl) RefreshExpiration(ctx context.Context, userID int64) error {
	return r.SetExpiration(ctx, userID, r.defaultTTL)
}

// IsEmpty 检查购物车是否为空
func (r *cartRepositoryImpl) IsEmpty(ctx context.Context, userID int64) (bool, error) {
	cart, err := r.GetCart(ctx, userID)
	if err != nil {
		return true, err
	}
	return len(cart.Items) == 0, nil
}

// MergeCart 合并购物车（用于用户登录后合并临时购物车）
func (r *cartRepositoryImpl) MergeCart(ctx context.Context, fromUserID, toUserID int64) error {
	if fromUserID <= 0 || toUserID <= 0 {
		return entity.ErrInvalidUserID
	}
	if fromUserID == toUserID {
		return nil // 相同用户，无需合并
	}

	// 获取源购物车
	fromCart, err := r.GetCart(ctx, fromUserID)
	if err != nil {
		return err
	}

	// 如果源购物车为空，直接返回
	if len(fromCart.Items) == 0 {
		return nil
	}

	// 获取目标购物车
	toCart, err := r.GetCart(ctx, toUserID)
	if err != nil {
		return err
	}

	// 合并商品
	for _, item := range fromCart.Items {
		// 检查目标购物车是否已有该商品
		found := false
		for i, existingItem := range toCart.Items {
			if existingItem.ProductID == item.ProductID {
				// 累加数量
				toCart.Items[i].Quantity += item.Quantity
				found = true
				break
			}
		}
		
		if !found {
			// 添加新商品
			item.UserID = toUserID
			toCart.Items = append(toCart.Items, item)
		}
	}

	// 重新计算总计
	toCart.UserID = toUserID
	for i := range toCart.Items {
		toCart.Items[i].UserID = toUserID
	}

	// 保存合并后的购物车
	if err := r.saveCart(ctx, toCart); err != nil {
		return err
	}

	// 清空源购物车
	return r.ClearCart(ctx, fromUserID)
}

// BatchAddItems 批量添加商品到购物车
func (r *cartRepositoryImpl) BatchAddItems(ctx context.Context, userID int64, items []entity.CartItem) error {
	if userID <= 0 {
		return entity.ErrInvalidUserID
	}
	if len(items) == 0 {
		return nil
	}

	// 获取当前购物车
	cart, err := r.GetCart(ctx, userID)
	if err != nil {
		return err
	}

	// 批量验证商品
	for _, item := range items {
		if item.ProductID <= 0 {
			return entity.ErrInvalidProductID
		}
		if item.Quantity <= 0 {
			return errors.New("quantity must be greater than 0")
		}

		// 验证商品是否存在
		product, err := r.productRepo.GetByID(ctx, item.ProductID)
		if err != nil {
			return err
		}

		// 检查商品是否上架
		if product.Status != 1 {
			return entity.ErrProductNotAvailable
		}

		// 检查库存
		if product.Stock < item.Quantity {
			return entity.ErrProductOutOfStock
		}
	}

	// 批量添加商品
	for _, item := range items {
		// 获取商品信息
		product, _ := r.productRepo.GetByID(ctx, item.ProductID)
		item.UserID = userID
		item.Product = product
		
		// 添加到购物车
		cart.AddItem(item)
	}

	// 保存购物车
	return r.saveCart(ctx, cart)
}

// GetCartWithProducts 获取购物车（包含商品详情）
func (r *cartRepositoryImpl) GetCartWithProducts(ctx context.Context, userID int64) (*entity.Cart, error) {
	cart, err := r.GetCart(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 为每个商品项加载商品详情
	for i, item := range cart.Items {
		if item.Product == nil {
			product, err := r.productRepo.GetByID(ctx, item.ProductID)
			if err != nil {
				// 如果商品不存在，跳过该项
				continue
			}
			cart.Items[i].Product = product
		}
	}

	// 重新计算总计（基于最新的商品价格）
	cart.TotalCount = 0
	cart.TotalPrice = 0
	
	validItems := []entity.CartItem{}
	for _, item := range cart.Items {
		if item.Product != nil {
			cart.TotalCount += item.Quantity
			cart.TotalPrice += int64(item.Quantity) * item.Product.Price
			validItems = append(validItems, item)
		}
	}
	
	cart.Items = validItems

	// 如果购物车有变化，保存更新
	if len(validItems) != len(cart.Items) {
		r.saveCart(ctx, cart)
	}

	return cart, nil
}

// ValidateCart 验证购物车（检查商品是否存在、库存是否充足等）
func (r *cartRepositoryImpl) ValidateCart(ctx context.Context, userID int64) ([]CartValidationError, error) {
	cart, err := r.GetCart(ctx, userID)
	if err != nil {
		return nil, err
	}

	var validationErrors []CartValidationError

	for _, item := range cart.Items {
		// 检查商品是否存在
		product, err := r.productRepo.GetByID(ctx, item.ProductID)
		if err != nil {
			if errors.Is(err, entity.ErrProductNotFound) {
				validationErrors = append(validationErrors, CartValidationError{
					ProductID: item.ProductID,
					Error:     "商品不存在",
					ErrorType: "not_found",
				})
			}
			continue
		}

		// 检查商品是否上架
		if product.Status != 1 {
			validationErrors = append(validationErrors, CartValidationError{
				ProductID: item.ProductID,
				Error:     "商品已下架",
				ErrorType: "unavailable",
			})
			continue
		}

		// 检查库存
		if product.Stock < item.Quantity {
			validationErrors = append(validationErrors, CartValidationError{
				ProductID: item.ProductID,
				Error:     fmt.Sprintf("库存不足，当前库存：%d，需要：%d", product.Stock, item.Quantity),
				ErrorType: "out_of_stock",
			})
		}
	}

	return validationErrors, nil
}

// GetActiveUserCarts 获取活跃用户购物车列表（管理员功能）
func (r *cartRepositoryImpl) GetActiveUserCarts(ctx context.Context, limit int) ([]int64, error) {
	if limit <= 0 {
		limit = 100
	}

	// 扫描所有购物车键
	pattern := r.keyPrefix + "*"
	keys, err := r.redis.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, err
	}

	var userIDs []int64
	for _, key := range keys {
		// 提取用户ID
		if len(key) > len(r.keyPrefix) {
			userIDStr := key[len(r.keyPrefix):]
			if userID, err := strconv.ParseInt(userIDStr, 10, 64); err == nil {
				userIDs = append(userIDs, userID)
				if len(userIDs) >= limit {
					break
				}
			}
		}
	}

	return userIDs, nil
}

// DeleteExpiredCarts 删除过期的购物车
func (r *cartRepositoryImpl) DeleteExpiredCarts(ctx context.Context) (int64, error) {
	// Redis会自动删除过期的键，这里主要是清理可能的孤立元数据
	pattern := r.keyPrefix + "meta:*"
	keys, err := r.redis.Keys(ctx, pattern).Result()
	if err != nil {
		return 0, err
	}

	var deletedCount int64
	for _, metaKey := range keys {
		// 检查对应的购物车是否存在
		userIDStr := metaKey[len(r.keyPrefix+"meta:"):]
		cartKey := r.keyPrefix + userIDStr
		
		exists, err := r.redis.Exists(ctx, cartKey).Result()
		if err != nil {
			continue
		}
		
		if exists == 0 {
			// 购物车不存在，删除元数据
			r.redis.Del(ctx, metaKey)
			deletedCount++
		}
	}

	return deletedCount, nil
}

// GetCartStatistics 获取购物车统计信息
func (r *cartRepositoryImpl) GetCartStatistics(ctx context.Context) (*CartStatistics, error) {
	// 获取所有活跃购物车
	userIDs, err := r.GetActiveUserCarts(ctx, 1000) // 限制1000个
	if err != nil {
		return nil, err
	}

	stats := &CartStatistics{
		TotalActiveCarts: int64(len(userIDs)),
	}

	if len(userIDs) == 0 {
		return stats, nil
	}

	var totalItems int64
	var totalValue int64

	for _, userID := range userIDs {
		cart, err := r.GetCart(ctx, userID)
		if err != nil {
			continue
		}

		totalItems += int64(cart.TotalCount)
		totalValue += cart.TotalPrice
	}

	stats.TotalItems = totalItems
	stats.TotalValue = totalValue

	if stats.TotalActiveCarts > 0 {
		stats.AverageItemsPerCart = float64(totalItems) / float64(stats.TotalActiveCarts)
		stats.AverageCartValue = float64(totalValue) / float64(stats.TotalActiveCarts)
	}

	return stats, nil
}