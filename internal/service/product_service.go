package service

import (
	"context"
	"errors"
	"strings"

	"go-shop/internal/dto"
	"go-shop/internal/entity"
	"go-shop/internal/repository"
)

// ProductService 商品业务服务
type ProductService struct {
	productRepo  repository.ProductRepository
	categoryRepo repository.CategoryRepository
}

// NewProductService 创建商品服务实例
func NewProductService(
	productRepo repository.ProductRepository,
	categoryRepo repository.CategoryRepository,
) *ProductService {
	return &ProductService{
		productRepo:  productRepo,
		categoryRepo: categoryRepo,
	}
}

// CreateProduct 创建商品
func (s *ProductService) CreateProduct(ctx context.Context, req *dto.ProductCreateRequest) (*dto.ProductResponse, error) {
	// 验证输入参数
	if err := s.validateCreateRequest(req); err != nil {
		return nil, err
	}

	// 验证分类是否存在
	_, err := s.categoryRepo.GetByID(ctx, req.CategoryID)
	if err != nil {
		if errors.Is(err, entity.ErrCategoryNotFound) {
			return nil, entity.ErrInvalidCategoryID
		}
		return nil, err
	}

	// 检查商品名称是否已存在
	exists, err := s.productRepo.ExistsByName(ctx, req.Name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, entity.ErrProductAlreadyExists
	}

	// 创建商品实体
	product := &entity.Product{
		CategoryID:  req.CategoryID,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		CoverImage:  req.CoverImage,
		Status:      1, // 默认上架
	}

	if err := s.productRepo.Create(ctx, product); err != nil {
		return nil, err
	}

	// 返回商品信息
	return s.entityToResponse(product), nil
}

// UpdateProduct 更新商品信息
func (s *ProductService) UpdateProduct(ctx context.Context, productID int64, req *dto.ProductUpdateRequest) (*dto.ProductResponse, error) {
	if productID <= 0 {
		return nil, entity.ErrInvalidProductID
	}

	if err := s.validateUpdateRequest(req); err != nil {
		return nil, err
	}

	// 获取当前商品信息
	product, err := s.productRepo.GetByID(ctx, productID)
	if err != nil {
		return nil, err
	}

	// 验证分类是否存在（如果提供）
	if req.CategoryID != nil {
		_, err := s.categoryRepo.GetByID(ctx, *req.CategoryID)
		if err != nil {
			if errors.Is(err, entity.ErrCategoryNotFound) {
				return nil, entity.ErrInvalidCategoryID
			}
			return nil, err
		}
		product.CategoryID = *req.CategoryID
	}

	// 检查商品名称是否已存在（如果提供）
	if req.Name != nil {
		exists, err := s.productRepo.ExistsByNameExcludeID(ctx, *req.Name, productID)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, entity.ErrProductAlreadyExists
		}
		product.Name = *req.Name
	}

	// 更新其他字段
	if req.Description != nil {
		product.Description = *req.Description
	}
	if req.Price != nil {
		product.Price = *req.Price
	}
	if req.Stock != nil {
		product.Stock = *req.Stock
	}
	if req.CoverImage != nil {
		product.CoverImage = *req.CoverImage
	}
	if req.Status != nil {
		product.Status = *req.Status
	}

	// 保存更新
	if err := s.productRepo.Update(ctx, product); err != nil {
		return nil, err
	}

	// 返回更新后的商品信息
	return s.entityToResponse(product), nil
}

// GetProductByID 根据ID获取商品信息
func (s *ProductService) GetProductByID(ctx context.Context, productID int64) (*dto.ProductResponse, error) {
	if productID <= 0 {
		return nil, entity.ErrInvalidProductID
	}

	product, err := s.productRepo.GetByID(ctx, productID)
	if err != nil {
		return nil, err
	}

	return s.entityToResponse(product), nil
}

// GetProductList 获取商品列表
func (s *ProductService) GetProductList(ctx context.Context, req *dto.ProductListRequest) (*dto.ProductListResponse, error) {
	// 设置默认分页参数
	page, pageSize := s.getDefaultPagination(req.Page, req.PageSize)
	offset := (page - 1) * pageSize

	// 构建查询选项
	options := repository.ProductQueryOptions{
		CategoryID: req.CategoryID,
		Keyword:    req.Keyword,
		Status:     getIntPtr(1), // 默认只查询上架商品
		Offset:     offset,
		Limit:      pageSize,
	}

	// 设置排序
	if req.SortBy != nil {
		options.SortBy = *req.SortBy
	} else {
		options.SortBy = "created_at" // 默认按创建时间排序
	}

	if req.SortOrder != nil {
		options.SortOrder = *req.SortOrder
	} else {
		options.SortOrder = "desc" // 默认降序
	}

	// 获取商品列表
	products, total, err := s.productRepo.List(ctx, options)
	if err != nil {
		return nil, err
	}

	// 转换为响应格式
	productResponses := make([]dto.ProductResponse, len(products))
	for i, product := range products {
		productResponses[i] = *s.entityToResponse(product)
	}

	// 创建分页响应
	pagination := dto.NewPaginationResponse(page, pageSize, total)

	return &dto.ProductListResponse{
		Products:   productResponses,
		Pagination: pagination,
	}, nil
}

// SearchProducts 搜索商品
func (s *ProductService) SearchProducts(ctx context.Context, keyword string, page, pageSize int) (*dto.ProductListResponse, error) {
	// 设置默认分页参数
	page, pageSize = s.getDefaultPagination(page, pageSize)
	offset := (page - 1) * pageSize

	// 搜索商品
	products, total, err := s.productRepo.Search(ctx, keyword, offset, pageSize)
	if err != nil {
		return nil, err
	}

	// 转换为响应格式
	productResponses := make([]dto.ProductResponse, len(products))
	for i, product := range products {
		productResponses[i] = *s.entityToResponse(product)
	}

	// 创建分页响应
	pagination := dto.NewPaginationResponse(page, pageSize, total)

	return &dto.ProductListResponse{
		Products:   productResponses,
		Pagination: pagination,
	}, nil
}

// GetProductsByCategory 根据分类获取商品列表
func (s *ProductService) GetProductsByCategory(ctx context.Context, categoryID int64, page, pageSize int) (*dto.ProductListResponse, error) {
	if categoryID <= 0 {
		return nil, entity.ErrInvalidCategoryID
	}

	// 验证分类是否存在
	_, err := s.categoryRepo.GetByID(ctx, categoryID)
	if err != nil {
		return nil, err
	}

	// 设置默认分页参数
	page, pageSize = s.getDefaultPagination(page, pageSize)
	offset := (page - 1) * pageSize

	// 获取分类下的商品
	products, total, err := s.productRepo.ListByCategory(ctx, categoryID, offset, pageSize)
	if err != nil {
		return nil, err
	}

	// 转换为响应格式
	productResponses := make([]dto.ProductResponse, len(products))
	for i, product := range products {
		productResponses[i] = *s.entityToResponse(product)
	}

	// 创建分页响应
	pagination := dto.NewPaginationResponse(page, pageSize, total)

	return &dto.ProductListResponse{
		Products:   productResponses,
		Pagination: pagination,
	}, nil
}

// UpdateProductStatus 更新商品状态
func (s *ProductService) UpdateProductStatus(ctx context.Context, productID int64, status int) error {
	if productID <= 0 {
		return entity.ErrInvalidProductID
	}

	// 验证状态值
	if status != 0 && status != 1 {
		return errors.New("商品状态只能是0（下架）或1（上架）")
	}

	// 检查商品是否存在
	_, err := s.productRepo.GetByID(ctx, productID)
	if err != nil {
		return err
	}

	// 更新状态
	return s.productRepo.UpdateStatus(ctx, productID, status)
}

// UpdateProductStock 更新商品库存
func (s *ProductService) UpdateProductStock(ctx context.Context, productID int64, stock int) error {
	if productID <= 0 {
		return entity.ErrInvalidProductID
	}

	if stock < 0 {
		return entity.ErrInvalidStock
	}

	// 检查商品是否存在
	_, err := s.productRepo.GetByID(ctx, productID)
	if err != nil {
		return err
	}

	// 更新库存
	return s.productRepo.UpdateStock(ctx, productID, stock)
}

// CheckStock 检查商品库存
func (s *ProductService) CheckStock(ctx context.Context, productID int64, quantity int) (bool, error) {
	if productID <= 0 {
		return false, entity.ErrInvalidProductID
	}

	if quantity <= 0 {
		return false, errors.New("数量必须大于0")
	}

	return s.productRepo.CheckStock(ctx, productID, quantity)
}

// DecrementStock 减少商品库存
func (s *ProductService) DecrementStock(ctx context.Context, productID int64, quantity int) error {
	if productID <= 0 {
		return entity.ErrInvalidProductID
	}

	if quantity <= 0 {
		return errors.New("数量必须大于0")
	}

	// 检查库存是否充足
	sufficient, err := s.productRepo.CheckStock(ctx, productID, quantity)
	if err != nil {
		return err
	}
	if !sufficient {
		return entity.ErrProductOutOfStock
	}

	// 减少库存
	return s.productRepo.DecrementStock(ctx, productID, quantity)
}

// IncrementStock 增加商品库存
func (s *ProductService) IncrementStock(ctx context.Context, productID int64, quantity int) error {
	if productID <= 0 {
		return entity.ErrInvalidProductID
	}

	if quantity <= 0 {
		return errors.New("数量必须大于0")
	}

	return s.productRepo.IncrementStock(ctx, productID, quantity)
}

// DeleteProduct 删除商品
func (s *ProductService) DeleteProduct(ctx context.Context, productID int64) error {
	if productID <= 0 {
		return entity.ErrInvalidProductID
	}

	// 检查商品是否存在
	_, err := s.productRepo.GetByID(ctx, productID)
	if err != nil {
		return err
	}

	return s.productRepo.Delete(ctx, productID)
}

// GetTopProducts 获取热门商品
func (s *ProductService) GetTopProducts(ctx context.Context, limit int) ([]dto.ProductResponse, error) {
	if limit <= 0 {
		limit = 10
	}

	products, err := s.productRepo.GetTopProducts(ctx, limit)
	if err != nil {
		return nil, err
	}

	// 转换为响应格式
	productResponses := make([]dto.ProductResponse, len(products))
	for i, product := range products {
		productResponses[i] = *s.entityToResponse(product)
	}

	return productResponses, nil
}

// GetLatestProducts 获取最新商品
func (s *ProductService) GetLatestProducts(ctx context.Context, limit int) ([]dto.ProductResponse, error) {
	if limit <= 0 {
		limit = 10
	}

	products, err := s.productRepo.GetLatestProducts(ctx, limit)
	if err != nil {
		return nil, err
	}

	// 转换为响应格式
	productResponses := make([]dto.ProductResponse, len(products))
	for i, product := range products {
		productResponses[i] = *s.entityToResponse(product)
	}

	return productResponses, nil
}

// GetLowStockProducts 获取低库存商品
func (s *ProductService) GetLowStockProducts(ctx context.Context, threshold int) ([]dto.ProductResponse, error) {
	if threshold <= 0 {
		threshold = 10 // 默认阈值
	}

	products, err := s.productRepo.GetLowStockProducts(ctx, threshold)
	if err != nil {
		return nil, err
	}

	// 转换为响应格式
	productResponses := make([]dto.ProductResponse, len(products))
	for i, product := range products {
		productResponses[i] = *s.entityToResponse(product)
	}

	return productResponses, nil
}

// 验证和辅助方法

// validateCreateRequest 验证创建请求
func (s *ProductService) validateCreateRequest(req *dto.ProductCreateRequest) error {
	if req == nil {
		return errors.New("创建请求不能为空")
	}

	if req.CategoryID <= 0 {
		return entity.ErrInvalidCategoryID
	}

	if strings.TrimSpace(req.Name) == "" {
		return errors.New("商品名称不能为空")
	}

	if len(req.Name) > 200 {
		return errors.New("商品名称长度不能超过200个字符")
	}

	if req.Price <= 0 {
		return entity.ErrInvalidPrice
	}

	if req.Stock < 0 {
		return entity.ErrInvalidStock
	}

	return nil
}

// validateUpdateRequest 验证更新请求
func (s *ProductService) validateUpdateRequest(req *dto.ProductUpdateRequest) error {
	if req == nil {
		return errors.New("更新请求不能为空")
	}

	if req.CategoryID != nil && *req.CategoryID <= 0 {
		return entity.ErrInvalidCategoryID
	}

	if req.Name != nil {
		if strings.TrimSpace(*req.Name) == "" {
			return errors.New("商品名称不能为空")
		}
		if len(*req.Name) > 200 {
			return errors.New("商品名称长度不能超过200个字符")
		}
	}

	if req.Price != nil && *req.Price <= 0 {
		return entity.ErrInvalidPrice
	}

	if req.Stock != nil && *req.Stock < 0 {
		return entity.ErrInvalidStock
	}

	if req.Status != nil && (*req.Status != 0 && *req.Status != 1) {
		return errors.New("商品状态只能是0（下架）或1（上架）")
	}

	return nil
}

// entityToResponse 将商品实体转换为响应DTO
func (s *ProductService) entityToResponse(product *entity.Product) *dto.ProductResponse {
	return &dto.ProductResponse{
		ID:          product.ID,
		CategoryID:  product.CategoryID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Stock:       product.Stock,
		CoverImage:  product.CoverImage,
		Status:      product.Status,
		CreatedAt:   product.CreatedAt,
		UpdatedAt:   product.UpdatedAt,
	}
}

// getDefaultPagination 获取默认分页参数
func (s *ProductService) getDefaultPagination(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	} else if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

// getIntPtr 获取int指针
func getIntPtr(value int) *int {
	return &value
}