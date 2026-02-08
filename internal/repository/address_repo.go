package repository

import (
	"context"

	"go-shop/internal/entity"
)

// AddressRepository 地址仓库接口
type AddressRepository interface {
	// Create 创建地址
	Create(ctx context.Context, address *entity.Address) error

	// GetByID 根据ID获取地址
	GetByID(ctx context.Context, id int64) (*entity.Address, error)

	// Update 更新地址信息
	Update(ctx context.Context, address *entity.Address) error

	// Delete 删除地址
	Delete(ctx context.Context, id int64) error

	// GetByUserID 获取用户的所有地址
	GetByUserID(ctx context.Context, userID int64) ([]*entity.Address, error)

	// GetDefaultByUserID 获取用户的默认地址
	GetDefaultByUserID(ctx context.Context, userID int64) (*entity.Address, error)

	// SetDefault 设置默认地址
	SetDefault(ctx context.Context, userID int64, addressID int64) error

	// UnsetDefault 取消默认地址
	UnsetDefault(ctx context.Context, userID int64, addressID int64) error

	// CountByUserID 获取用户地址数量
	CountByUserID(ctx context.Context, userID int64) (int64, error)

	// ExistsByID 检查地址是否存在
	ExistsByID(ctx context.Context, id int64) (bool, error)

	// ExistsByUserIDAndID 检查用户是否拥有指定地址
	ExistsByUserIDAndID(ctx context.Context, userID int64, addressID int64) (bool, error)

	// GetUserAddressWithPagination 分页获取用户地址列表
	GetUserAddressWithPagination(ctx context.Context, userID int64, offset, limit int) ([]*entity.Address, int64, error)

	// DeleteByUserID 删除用户的所有地址
	DeleteByUserID(ctx context.Context, userID int64) error

	// HasDefaultAddress 检查用户是否有默认地址
	HasDefaultAddress(ctx context.Context, userID int64) (bool, error)

	// GetAddressesInRegion 获取指定地区的地址列表（管理员功能）
	GetAddressesInRegion(ctx context.Context, province, city string, limit int) ([]*entity.Address, error)

	// GetAddressStatistics 获取地址统计信息（管理员功能）
	GetAddressStatistics(ctx context.Context) (*AddressStatistics, error)
}

// AddressStatistics 地址统计信息
type AddressStatistics struct {
	TotalAddresses       int64                    `json:"total_addresses"`       // 总地址数
	UsersWithAddresses   int64                    `json:"users_with_addresses"`  // 有地址的用户数
	AverageAddressPerUser float64                 `json:"average_address_per_user"` // 平均每用户地址数
	TopProvinces         []ProvinceStatistic      `json:"top_provinces"`         // 热门省份
	TopCities            []CityStatistic          `json:"top_cities"`            // 热门城市
}

// ProvinceStatistic 省份统计
type ProvinceStatistic struct {
	Province string `json:"province"`
	Count    int64  `json:"count"`
}

// CityStatistic 城市统计
type CityStatistic struct {
	Province string `json:"province"`
	City     string `json:"city"`
	Count    int64  `json:"count"`
}