package service

import (
	"context"
	"errors"

	"go-shop/internal/dto"
	"go-shop/internal/entity"
	"go-shop/internal/repository"
)

// CartService 购物车业务服务
type CartService struct {
	cartRepo    repository.CartRepository
	productRepo repository.ProductRepository
	skuRepo     repository.SKURepository
}

// NewCartService 创建购物车服务实例
func NewCartService(
	cartRepo repository.CartRepository,
	productRepo repository.ProductRepository,
	skuRepo repository.SKURepository,
) *CartService {
	return &CartService{
		cartRepo:    cartRepo,
		productRepo: productRepo,
		skuRepo:     skuRepo,
	}
}

// GetCart 获取用户购物车
func (s *CartService) GetCart(ctx context.Context, userID int64) (*dto.CartResponse, error) {
	if userID <= 0 {
		return nil, entity.ErrInvalidUserID
	}

	// 获取购物车（包含商品详情）
	cart, err := s.cartRepo.GetCartWithProducts(ctx, userID)
	if err != nil {
		return nil, err
	}

	return s.entityToResponse(cart), nil
}

// AddItem 添加商品到购物车
// 需求：1.5 - 当用户将商品加入购物车时，系统应记录选择的SKU信息
func (s *CartService) AddItem(ctx context.Context, userID int64, req *dto.CartAddItemRequest) (*dto.CartResponse, error) {
	if userID <= 0 {
		return nil, entity.ErrInvalidUserID
	}

	// 验证请求参数
	if err := s.validateAddItemRequest(req); err != nil {
		return nil, err
	}

	// 验证商品是否存在
	product, err := s.productRepo.GetByID(ctx, req.ProductID)
	if err != nil {
		return nil, err
	}

	// 检查商品是否上架
	if product.Status != 1 {
		return nil, entity.ErrProductNotAvailable
	}

	// 创建购物车商品项
	item := entity.CartItem{
		UserID:    userID,
		ProductID: req.ProductID,
		Quantity:  req.Quantity,
		Product:   product,
	}

	// 如果指定了SKU，验证SKU并使用SKU的库存和价格
	if req.SKUID != nil {
		sku, err := s.skuRepo.GetSKUByID(ctx, *req.SKUID)
		if err != nil {
			return nil, errors.New("SKU不存在")
		}

		// 验证SKU属于该商品
		if sku.ProductID != req.ProductID {
			return nil, errors.New("SKU不属于该商品")
		}

		// 检查SKU是否启用
		if !sku.IsActive {
			return nil, errors.New("SKU已下架")
		}

		// 检查SKU库存是否充足
		if sku.Stock < req.Quantity {
			return nil, entity.ErrProductOutOfStock
		}

		// 设置SKU信息
		item.SKUID = &sku.ID
		item.SKUCode = &sku.SKUCode
		item.SpecValues = &sku.SpecValues
		item.SKU = sku
	} else {
		// 没有SKU时，检查商品库存
		if product.Stock < req.Quantity {
			return nil, entity.ErrProductOutOfStock
		}
	}

	// 添加到购物车
	if err := s.cartRepo.AddItem(ctx, userID, item); err != nil {
		return nil, err
	}

	// 返回更新后的购物车
	return s.GetCart(ctx, userID)
}

// UpdateItemQuantity 更新购物车商品数量
func (s *CartService) UpdateItemQuantity(ctx context.Context, userID int64, req *dto.CartUpdateItemRequest) (*dto.CartResponse, error) {
	if userID <= 0 {
		return nil, entity.ErrInvalidUserID
	}

	// 验证请求参数
	if err := s.validateUpdateItemRequest(req); err != nil {
		return nil, err
	}

	// 检查商品是否在购物车中
	hasItem, err := s.cartRepo.HasItem(ctx, userID, req.ProductID)
	if err != nil {
		return nil, err
	}
	if !hasItem {
		return nil, entity.ErrCartItemNotFound
	}

	// 如果指定了SKU，验证SKU库存
	if req.SKUID != nil {
		sku, err := s.skuRepo.GetSKUByID(ctx, *req.SKUID)
		if err != nil {
			return nil, errors.New("SKU不存在")
		}

		if sku.Stock < req.Quantity {
			return nil, entity.ErrProductOutOfStock
		}
	} else {
		// 验证商品库存
		product, err := s.productRepo.GetByID(ctx, req.ProductID)
		if err != nil {
			return nil, err
		}

		if product.Stock < req.Quantity {
			return nil, entity.ErrProductOutOfStock
		}
	}

	// 更新商品数量
	if err := s.cartRepo.UpdateItemQuantity(ctx, userID, req.ProductID, req.Quantity); err != nil {
		return nil, err
	}

	// 返回更新后的购物车
	return s.GetCart(ctx, userID)
}

// RemoveItem 从购物车删除商品
func (s *CartService) RemoveItem(ctx context.Context, userID int64, productID int64) (*dto.CartResponse, error) {
	if userID <= 0 {
		return nil, entity.ErrInvalidUserID
	}

	if productID <= 0 {
		return nil, entity.ErrInvalidProductID
	}

	// 检查商品是否在购物车中
	hasItem, err := s.cartRepo.HasItem(ctx, userID, productID)
	if err != nil {
		return nil, err
	}
	if !hasItem {
		return nil, entity.ErrCartItemNotFound
	}

	// 删除商品
	if err := s.cartRepo.RemoveItem(ctx, userID, productID); err != nil {
		return nil, err
	}

	// 返回更新后的购物车
	return s.GetCart(ctx, userID)
}

// ClearCart 清空购物车
func (s *CartService) ClearCart(ctx context.Context, userID int64) error {
	if userID <= 0 {
		return entity.ErrInvalidUserID
	}

	return s.cartRepo.ClearCart(ctx, userID)
}

// GetItemCount 获取购物车商品总数量
func (s *CartService) GetItemCount(ctx context.Context, userID int64) (int, error) {
	if userID <= 0 {
		return 0, entity.ErrInvalidUserID
	}

	return s.cartRepo.GetItemCount(ctx, userID)
}

// GetTotalPrice 获取购物车总价格
func (s *CartService) GetTotalPrice(ctx context.Context, userID int64) (int64, error) {
	if userID <= 0 {
		return 0, entity.ErrInvalidUserID
	}

	return s.cartRepo.GetTotalPrice(ctx, userID)
}

// ValidateCart 验证购物车
func (s *CartService) ValidateCart(ctx context.Context, userID int64) ([]repository.CartValidationError, error) {
	if userID <= 0 {
		return nil, entity.ErrInvalidUserID
	}

	// 检查购物车是否为空
	isEmpty, err := s.cartRepo.IsEmpty(ctx, userID)
	if err != nil {
		return nil, err
	}
	if isEmpty {
		return nil, entity.ErrCartEmpty
	}

	// 验证购物车
	return s.cartRepo.ValidateCart(ctx, userID)
}

// MergeCart 合并购物车（用于用户登录后合并临时购物车）
func (s *CartService) MergeCart(ctx context.Context, fromUserID, toUserID int64) error {
	if fromUserID <= 0 || toUserID <= 0 {
		return entity.ErrInvalidUserID
	}

	if fromUserID == toUserID {
		return nil // 相同用户，无需合并
	}

	return s.cartRepo.MergeCart(ctx, fromUserID, toUserID)
}

// BatchAddItems 批量添加商品到购物车
func (s *CartService) BatchAddItems(ctx context.Context, userID int64, items []dto.CartAddItemRequest) (*dto.CartResponse, error) {
	if userID <= 0 {
		return nil, entity.ErrInvalidUserID
	}

	if len(items) == 0 {
		return s.GetCart(ctx, userID)
	}

	// 验证所有商品
	cartItems := make([]entity.CartItem, 0, len(items))
	for _, req := range items {
		// 验证请求参数
		if err := s.validateAddItemRequest(&req); err != nil {
			return nil, err
		}

		// 验证商品是否存在
		product, err := s.productRepo.GetByID(ctx, req.ProductID)
		if err != nil {
			return nil, err
		}

		// 检查商品是否上架
		if product.Status != 1 {
			return nil, entity.ErrProductNotAvailable
		}

		// 检查库存是否充足
		if product.Stock < req.Quantity {
			return nil, entity.ErrProductOutOfStock
		}

		// 创建购物车商品项
		cartItems = append(cartItems, entity.CartItem{
			UserID:    userID,
			ProductID: req.ProductID,
			Quantity:  req.Quantity,
			Product:   product,
		})
	}

	// 批量添加到购物车
	if err := s.cartRepo.BatchAddItems(ctx, userID, cartItems); err != nil {
		return nil, err
	}

	// 返回更新后的购物车
	return s.GetCart(ctx, userID)
}

// CheckCartStock 检查购物车库存是否充足
func (s *CartService) CheckCartStock(ctx context.Context, userID int64) (bool, error) {
	if userID <= 0 {
		return false, entity.ErrInvalidUserID
	}

	// 验证购物车
	validationErrors, err := s.cartRepo.ValidateCart(ctx, userID)
	if err != nil {
		return false, err
	}

	// 检查是否有库存不足的商品
	for _, validationError := range validationErrors {
		if validationError.ErrorType == "out_of_stock" {
			return false, nil
		}
	}

	return true, nil
}

// GetCartSummary 获取购物车摘要信息
func (s *CartService) GetCartSummary(ctx context.Context, userID int64) (map[string]interface{}, error) {
	if userID <= 0 {
		return nil, entity.ErrInvalidUserID
	}

	// 获取购物车
	cart, err := s.cartRepo.GetCartWithProducts(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 验证购物车
	validationErrors, err := s.cartRepo.ValidateCart(ctx, userID)
	if err != nil {
		return nil, err
	}

	summary := map[string]interface{}{
		"user_id":           userID,
		"total_items":       cart.TotalCount,
		"total_price":       cart.TotalPrice,
		"item_count":        len(cart.Items),
		"is_empty":          len(cart.Items) == 0,
		"has_errors":        len(validationErrors) > 0,
		"validation_errors": validationErrors,
	}

	return summary, nil
}

// RefreshCart 刷新购物车（更新商品价格和库存信息）
func (s *CartService) RefreshCart(ctx context.Context, userID int64) (*dto.CartResponse, error) {
	if userID <= 0 {
		return nil, entity.ErrInvalidUserID
	}

	// 获取购物车（包含最新商品详情）
	cart, err := s.cartRepo.GetCartWithProducts(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 转换为响应格式
	return s.entityToResponse(cart), nil
}

// 验证和辅助方法

// validateAddItemRequest 验证添加商品请求
func (s *CartService) validateAddItemRequest(req *dto.CartAddItemRequest) error {
	if req == nil {
		return errors.New("添加商品请求不能为空")
	}

	if req.ProductID <= 0 {
		return entity.ErrInvalidProductID
	}

	if req.Quantity <= 0 {
		return errors.New("商品数量必须大于0")
	}

	if req.Quantity > 999 {
		return errors.New("商品数量不能超过999")
	}

	return nil
}

// validateUpdateItemRequest 验证更新商品请求
func (s *CartService) validateUpdateItemRequest(req *dto.CartUpdateItemRequest) error {
	if req == nil {
		return errors.New("更新商品请求不能为空")
	}

	if req.ProductID <= 0 {
		return entity.ErrInvalidProductID
	}

	if req.Quantity <= 0 {
		return errors.New("商品数量必须大于0")
	}

	if req.Quantity > 999 {
		return errors.New("商品数量不能超过999")
	}

	return nil
}

// entityToResponse 将购物车实体转换为响应DTO
func (s *CartService) entityToResponse(cart *entity.Cart) *dto.CartResponse {
	items := make([]dto.CartItemResponse, len(cart.Items))
	
	for i, item := range cart.Items {
		var productResp dto.ProductResponse
		if item.Product != nil {
			productResp = dto.ProductResponse{
				ID:          item.Product.ID,
				CategoryID:  item.Product.CategoryID,
				Name:        item.Product.Name,
				Description: item.Product.Description,
				Price:       item.Product.Price,
				Stock:       item.Product.Stock,
				CoverImage:  item.Product.CoverImage,
				Status:      item.Product.Status,
				CreatedAt:   item.Product.CreatedAt,
				UpdatedAt:   item.Product.UpdatedAt,
			}
		}

		// 计算小计金额：优先使用SKU价格，否则使用商品价格
		var subtotal int64
		if item.SKU != nil {
			subtotal = int64(item.Quantity) * item.SKU.Price
		} else {
			subtotal = int64(item.Quantity) * productResp.Price
		}

		// 转换SpecValues
		var specValues *map[string]string
		if item.SpecValues != nil {
			sv := map[string]string(*item.SpecValues)
			specValues = &sv
		}

		items[i] = dto.CartItemResponse{
			ProductID:      item.ProductID,
			SKUID:          item.SKUID,
			SKUCode:        item.SKUCode,
			SpecValues:     specValues,
			Quantity:       item.Quantity,
			Product:        productResp,
			SubtotalAmount: subtotal,
		}
	}

	return &dto.CartResponse{
		UserID:     cart.UserID,
		Items:      items,
		TotalCount: cart.TotalCount,
		TotalPrice: cart.TotalPrice,
	}
}
