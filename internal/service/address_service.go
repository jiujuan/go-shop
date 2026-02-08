package service

import (
	"context"

	"go-shop/internal/dto"
	"go-shop/internal/entity"
	"go-shop/internal/repository"
)

// AddressService 地址业务服务接口
type AddressService interface {
	// CreateAddress 创建地址
	CreateAddress(ctx context.Context, userID int64, req *dto.AddressCreateRequest) (*dto.AddressResponse, error)

	// GetAddress 获取地址详情
	GetAddress(ctx context.Context, userID int64, addressID int64) (*dto.AddressResponse, error)

	// UpdateAddress 更新地址
	UpdateAddress(ctx context.Context, userID int64, addressID int64, req *dto.AddressUpdateRequest) (*dto.AddressResponse, error)

	// DeleteAddress 删除地址
	DeleteAddress(ctx context.Context, userID int64, addressID int64) error

	// GetUserAddresses 获取用户所有地址
	GetUserAddresses(ctx context.Context, userID int64) ([]*dto.AddressResponse, error)

	// GetDefaultAddress 获取用户默认地址
	GetDefaultAddress(ctx context.Context, userID int64) (*dto.AddressResponse, error)

	// SetDefaultAddress 设置默认地址
	SetDefaultAddress(ctx context.Context, userID int64, addressID int64) error

	// GetUserAddressesWithPagination 分页获取用户地址
	GetUserAddressesWithPagination(ctx context.Context, userID int64, page, pageSize int) (*dto.AddressListResponse, error)
}

type addressServiceImpl struct {
	addressRepo repository.AddressRepository
}

// NewAddressService 创建地址服务实例
func NewAddressService(addressRepo repository.AddressRepository) AddressService {
	return &addressServiceImpl{
		addressRepo: addressRepo,
	}
}

// CreateAddress 创建地址
func (s *addressServiceImpl) CreateAddress(ctx context.Context, userID int64, req *dto.AddressCreateRequest) (*dto.AddressResponse, error) {
	if userID <= 0 {
		return nil, entity.ErrInvalidUserID
	}

	// 转换DTO为实体
	isDefault := false
	if req.IsDefault != nil {
		isDefault = *req.IsDefault
	}

	address := &entity.Address{
		UserID:        userID,
		RecipientName: req.RecipientName,
		Phone:         req.Phone,
		Province:      req.Province,
		City:          req.City,
		District:      req.District,
		Detail:        req.Detail,
		IsDefault:     isDefault,
	}

	// 创建地址
	if err := s.addressRepo.Create(ctx, address); err != nil {
		return nil, err
	}

	return s.toAddressResponse(address), nil
}

// GetAddress 获取地址详情
func (s *addressServiceImpl) GetAddress(ctx context.Context, userID int64, addressID int64) (*dto.AddressResponse, error) {
	if userID <= 0 {
		return nil, entity.ErrInvalidUserID
	}
	if addressID <= 0 {
		return nil, entity.ErrInvalidAddressID
	}

	// 检查地址是否属于该用户
	exists, err := s.addressRepo.ExistsByUserIDAndID(ctx, userID, addressID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, entity.ErrAddressNotFound
	}

	// 获取地址
	address, err := s.addressRepo.GetByID(ctx, addressID)
	if err != nil {
		return nil, err
	}

	return s.toAddressResponse(address), nil
}

// UpdateAddress 更新地址
func (s *addressServiceImpl) UpdateAddress(ctx context.Context, userID int64, addressID int64, req *dto.AddressUpdateRequest) (*dto.AddressResponse, error) {
	if userID <= 0 {
		return nil, entity.ErrInvalidUserID
	}
	if addressID <= 0 {
		return nil, entity.ErrInvalidAddressID
	}

	// 检查地址是否属于该用户
	exists, err := s.addressRepo.ExistsByUserIDAndID(ctx, userID, addressID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, entity.ErrAddressNotFound
	}

	// 获取现有地址
	address, err := s.addressRepo.GetByID(ctx, addressID)
	if err != nil {
		return nil, err
	}

	// 更新字段
	if req.RecipientName != nil {
		address.RecipientName = *req.RecipientName
	}
	if req.Phone != nil {
		address.Phone = *req.Phone
	}
	if req.Province != nil {
		address.Province = *req.Province
	}
	if req.City != nil {
		address.City = *req.City
	}
	if req.District != nil {
		address.District = *req.District
	}
	if req.Detail != nil {
		address.Detail = *req.Detail
	}
	if req.IsDefault != nil {
		address.IsDefault = *req.IsDefault
	}

	// 更新地址
	if err := s.addressRepo.Update(ctx, address); err != nil {
		return nil, err
	}

	return s.toAddressResponse(address), nil
}

// DeleteAddress 删除地址
func (s *addressServiceImpl) DeleteAddress(ctx context.Context, userID int64, addressID int64) error {
	if userID <= 0 {
		return entity.ErrInvalidUserID
	}
	if addressID <= 0 {
		return entity.ErrInvalidAddressID
	}

	// 检查地址是否属于该用户
	exists, err := s.addressRepo.ExistsByUserIDAndID(ctx, userID, addressID)
	if err != nil {
		return err
	}
	if !exists {
		return entity.ErrAddressNotFound
	}

	// 删除地址
	return s.addressRepo.Delete(ctx, addressID)
}

// GetUserAddresses 获取用户所有地址
func (s *addressServiceImpl) GetUserAddresses(ctx context.Context, userID int64) ([]*dto.AddressResponse, error) {
	if userID <= 0 {
		return nil, entity.ErrInvalidUserID
	}

	// 获取地址列表
	addresses, err := s.addressRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 转换为响应DTO
	responses := make([]*dto.AddressResponse, 0, len(addresses))
	for _, addr := range addresses {
		responses = append(responses, s.toAddressResponse(addr))
	}

	return responses, nil
}

// GetDefaultAddress 获取用户默认地址
func (s *addressServiceImpl) GetDefaultAddress(ctx context.Context, userID int64) (*dto.AddressResponse, error) {
	if userID <= 0 {
		return nil, entity.ErrInvalidUserID
	}

	// 获取默认地址
	address, err := s.addressRepo.GetDefaultByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return s.toAddressResponse(address), nil
}

// SetDefaultAddress 设置默认地址
func (s *addressServiceImpl) SetDefaultAddress(ctx context.Context, userID int64, addressID int64) error {
	if userID <= 0 {
		return entity.ErrInvalidUserID
	}
	if addressID <= 0 {
		return entity.ErrInvalidAddressID
	}

	// 检查地址是否属于该用户
	exists, err := s.addressRepo.ExistsByUserIDAndID(ctx, userID, addressID)
	if err != nil {
		return err
	}
	if !exists {
		return entity.ErrAddressNotFound
	}

	// 设置默认地址
	return s.addressRepo.SetDefault(ctx, userID, addressID)
}

// GetUserAddressesWithPagination 分页获取用户地址
func (s *addressServiceImpl) GetUserAddressesWithPagination(ctx context.Context, userID int64, page, pageSize int) (*dto.AddressListResponse, error) {
	if userID <= 0 {
		return nil, entity.ErrInvalidUserID
	}

	// 计算偏移量
	offset, limit := s.calculatePagination(page, pageSize)

	// 获取分页数据
	addresses, total, err := s.addressRepo.GetUserAddressWithPagination(ctx, userID, offset, limit)
	if err != nil {
		return nil, err
	}

	// 转换为响应DTO
	addressResponses := make([]dto.AddressResponse, 0, len(addresses))
	for _, addr := range addresses {
		addressResponses = append(addressResponses, *s.toAddressResponse(addr))
	}

	pagination := dto.NewPaginationResponse(page, pageSize, total)
	return &dto.AddressListResponse{
		Addresses:  addressResponses,
		Pagination: pagination,
	}, nil
}

// calculatePagination 计算分页参数
func (s *addressServiceImpl) calculatePagination(page, pageSize int) (offset, limit int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	offset = (page - 1) * pageSize
	limit = pageSize
	return
}

// toAddressResponse 转换实体为响应DTO
func (s *addressServiceImpl) toAddressResponse(address *entity.Address) *dto.AddressResponse {
	if address == nil {
		return nil
	}

	return &dto.AddressResponse{
		ID:            address.ID,
		UserID:        address.UserID,
		RecipientName: address.RecipientName,
		Phone:         address.Phone,
		Province:      address.Province,
		City:          address.City,
		District:      address.District,
		Detail:        address.Detail,
		IsDefault:     address.IsDefault,
		CreatedAt:     address.CreatedAt,
		UpdatedAt:     address.UpdatedAt,
	}
}
