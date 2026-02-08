package repository

import (
	"context"

	"go-shop/internal/entity"

	"gorm.io/gorm"
)

// addressRepositoryImpl 地址仓库GORM实现
type addressRepositoryImpl struct {
	db *gorm.DB
}

// NewAddressRepository 创建地址仓库实例
func NewAddressRepository(db *gorm.DB) AddressRepository {
	return &addressRepositoryImpl{
		db: db,
	}
}

// Create 创建地址
func (r *addressRepositoryImpl) Create(ctx context.Context, address *entity.Address) error {
	if address == nil {
		return entity.ErrInvalidAddressID
	}

	// 验证地址信息
	if err := address.Validate(); err != nil {
		return err
	}

	// 验证用户ID
	if address.UserID <= 0 {
		return entity.ErrInvalidUserID
	}

	// 开始事务
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 如果是第一个地址或者设置为默认地址，需要处理默认地址逻辑
		if address.IsDefault {
			// 取消该用户的其他默认地址
			if err := tx.Model(&entity.Address{}).
				Where("user_id = ? AND is_default = ?", address.UserID, true).
				Update("is_default", false).Error; err != nil {
				return err
			}
		} else {
			// 检查用户是否已有地址，如果没有则设为默认
			var count int64
			if err := tx.Model(&entity.Address{}).
				Where("user_id = ?", address.UserID).
				Count(&count).Error; err != nil {
				return err
			}
			if count == 0 {
				address.IsDefault = true
			}
		}

		// 创建地址
		return tx.Create(address).Error
	})
}

// GetByID 根据ID获取地址
func (r *addressRepositoryImpl) GetByID(ctx context.Context, id int64) (*entity.Address, error) {
	if id <= 0 {
		return nil, entity.ErrInvalidAddressID
	}

	var address entity.Address
	err := r.db.WithContext(ctx).First(&address, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, entity.ErrAddressNotFound
		}
		return nil, err
	}

	return &address, nil
}

// Update 更新地址信息
func (r *addressRepositoryImpl) Update(ctx context.Context, address *entity.Address) error {
	if address == nil || address.ID <= 0 {
		return entity.ErrInvalidAddressID
	}

	// 验证地址信息
	if err := address.Validate(); err != nil {
		return err
	}

	// 检查地址是否存在
	var existingAddress entity.Address
	if err := r.db.WithContext(ctx).First(&existingAddress, address.ID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return entity.ErrAddressNotFound
		}
		return err
	}

	// 验证用户权限
	if existingAddress.UserID != address.UserID {
		return entity.ErrAddressNotFound
	}

	// 开始事务
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 如果设置为默认地址，需要取消其他默认地址
		if address.IsDefault && !existingAddress.IsDefault {
			if err := tx.Model(&entity.Address{}).
				Where("user_id = ? AND id != ? AND is_default = ?", address.UserID, address.ID, true).
				Update("is_default", false).Error; err != nil {
				return err
			}
		}

		// 更新地址信息
		return tx.Model(&existingAddress).Updates(map[string]interface{}{
			"recipient_name": address.RecipientName,
			"phone":          address.Phone,
			"province":       address.Province,
			"city":           address.City,
			"district":       address.District,
			"detail":         address.Detail,
			"is_default":     address.IsDefault,
		}).Error
	})
}

// Delete 删除地址
func (r *addressRepositoryImpl) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return entity.ErrInvalidAddressID
	}

	// 检查地址是否存在
	var address entity.Address
	if err := r.db.WithContext(ctx).First(&address, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return entity.ErrAddressNotFound
		}
		return err
	}

	// 开始事务
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 如果删除的是默认地址，需要设置另一个地址为默认
		if address.IsDefault {
			var otherAddress entity.Address
			if err := tx.Where("user_id = ? AND id != ?", address.UserID, id).
				First(&otherAddress).Error; err == nil {
				// 有其他地址，设置第一个为默认
				if err := tx.Model(&otherAddress).Update("is_default", true).Error; err != nil {
					return err
				}
			}
		}

		// 删除地址
		return tx.Delete(&address).Error
	})
}

// GetByUserID 获取用户的所有地址
func (r *addressRepositoryImpl) GetByUserID(ctx context.Context, userID int64) ([]*entity.Address, error) {
	if userID <= 0 {
		return nil, entity.ErrInvalidUserID
	}

	var addresses []*entity.Address
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("is_default DESC, created_at DESC").
		Find(&addresses).Error

	return addresses, err
}

// GetDefaultByUserID 获取用户的默认地址
func (r *addressRepositoryImpl) GetDefaultByUserID(ctx context.Context, userID int64) (*entity.Address, error) {
	if userID <= 0 {
		return nil, entity.ErrInvalidUserID
	}

	var address entity.Address
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_default = ?", userID, true).
		First(&address).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, entity.ErrAddressNotFound
		}
		return nil, err
	}

	return &address, nil
}

// SetDefault 设置默认地址
func (r *addressRepositoryImpl) SetDefault(ctx context.Context, userID int64, addressID int64) error {
	if userID <= 0 {
		return entity.ErrInvalidUserID
	}
	if addressID <= 0 {
		return entity.ErrInvalidAddressID
	}

	// 检查地址是否存在且属于该用户
	var address entity.Address
	if err := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", addressID, userID).
		First(&address).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return entity.ErrAddressNotFound
		}
		return err
	}

	// 如果已经是默认地址，直接返回
	if address.IsDefault {
		return entity.ErrAddressAlreadyDefault
	}

	// 开始事务
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 取消该用户的其他默认地址
		if err := tx.Model(&entity.Address{}).
			Where("user_id = ? AND is_default = ?", userID, true).
			Update("is_default", false).Error; err != nil {
			return err
		}

		// 设置新的默认地址
		return tx.Model(&address).Update("is_default", true).Error
	})
}

// UnsetDefault 取消默认地址
func (r *addressRepositoryImpl) UnsetDefault(ctx context.Context, userID int64, addressID int64) error {
	if userID <= 0 {
		return entity.ErrInvalidUserID
	}
	if addressID <= 0 {
		return entity.ErrInvalidAddressID
	}

	// 检查地址是否存在且属于该用户
	var address entity.Address
	if err := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", addressID, userID).
		First(&address).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return entity.ErrAddressNotFound
		}
		return err
	}

	// 如果不是默认地址，直接返回
	if !address.IsDefault {
		return nil
	}

	// 检查是否还有其他地址
	var count int64
	if err := r.db.WithContext(ctx).Model(&entity.Address{}).
		Where("user_id = ? AND id != ?", userID, addressID).
		Count(&count).Error; err != nil {
		return err
	}

	// 如果只有一个地址，不能取消默认
	if count == 0 {
		return entity.ErrCannotDeleteDefault
	}

	// 开始事务
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 取消当前默认地址
		if err := tx.Model(&address).Update("is_default", false).Error; err != nil {
			return err
		}

		// 设置另一个地址为默认
		var otherAddress entity.Address
		if err := tx.Where("user_id = ? AND id != ?", userID, addressID).
			First(&otherAddress).Error; err != nil {
			return err
		}

		return tx.Model(&otherAddress).Update("is_default", true).Error
	})
}

// CountByUserID 获取用户地址数量
func (r *addressRepositoryImpl) CountByUserID(ctx context.Context, userID int64) (int64, error) {
	if userID <= 0 {
		return 0, entity.ErrInvalidUserID
	}

	var count int64
	err := r.db.WithContext(ctx).Model(&entity.Address{}).
		Where("user_id = ?", userID).
		Count(&count).Error

	return count, err
}

// ExistsByID 检查地址是否存在
func (r *addressRepositoryImpl) ExistsByID(ctx context.Context, id int64) (bool, error) {
	if id <= 0 {
		return false, entity.ErrInvalidAddressID
	}

	var count int64
	err := r.db.WithContext(ctx).Model(&entity.Address{}).
		Where("id = ?", id).
		Count(&count).Error

	return count > 0, err
}

// ExistsByUserIDAndID 检查用户是否拥有指定地址
func (r *addressRepositoryImpl) ExistsByUserIDAndID(ctx context.Context, userID int64, addressID int64) (bool, error) {
	if userID <= 0 {
		return false, entity.ErrInvalidUserID
	}
	if addressID <= 0 {
		return false, entity.ErrInvalidAddressID
	}

	var count int64
	err := r.db.WithContext(ctx).Model(&entity.Address{}).
		Where("id = ? AND user_id = ?", addressID, userID).
		Count(&count).Error

	return count > 0, err
}

// GetUserAddressWithPagination 分页获取用户地址列表
func (r *addressRepositoryImpl) GetUserAddressWithPagination(ctx context.Context, userID int64, offset, limit int) ([]*entity.Address, int64, error) {
	if userID <= 0 {
		return nil, 0, entity.ErrInvalidUserID
	}

	var addresses []*entity.Address
	var total int64

	// 获取总数
	if err := r.db.WithContext(ctx).Model(&entity.Address{}).
		Where("user_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("is_default DESC, created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&addresses).Error

	return addresses, total, err
}

// DeleteByUserID 删除用户的所有地址
func (r *addressRepositoryImpl) DeleteByUserID(ctx context.Context, userID int64) error {
	if userID <= 0 {
		return entity.ErrInvalidUserID
	}

	return r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&entity.Address{}).Error
}

// HasDefaultAddress 检查用户是否有默认地址
func (r *addressRepositoryImpl) HasDefaultAddress(ctx context.Context, userID int64) (bool, error) {
	if userID <= 0 {
		return false, entity.ErrInvalidUserID
	}

	var count int64
	err := r.db.WithContext(ctx).Model(&entity.Address{}).
		Where("user_id = ? AND is_default = ?", userID, true).
		Count(&count).Error

	return count > 0, err
}

// GetAddressesInRegion 获取指定地区的地址列表（管理员功能）
func (r *addressRepositoryImpl) GetAddressesInRegion(ctx context.Context, province, city string, limit int) ([]*entity.Address, error) {
	if limit <= 0 {
		limit = 100
	}

	query := r.db.WithContext(ctx).Model(&entity.Address{})

	if province != "" {
		query = query.Where("province = ?", province)
	}

	if city != "" {
		query = query.Where("city = ?", city)
	}

	var addresses []*entity.Address
	err := query.Order("created_at DESC").
		Limit(limit).
		Find(&addresses).Error

	return addresses, err
}

// GetAddressStatistics 获取地址统计信息（管理员功能）
func (r *addressRepositoryImpl) GetAddressStatistics(ctx context.Context) (*AddressStatistics, error) {
	stats := &AddressStatistics{}

	// 获取总地址数
	if err := r.db.WithContext(ctx).Model(&entity.Address{}).
		Count(&stats.TotalAddresses).Error; err != nil {
		return nil, err
	}

	// 获取有地址的用户数
	if err := r.db.WithContext(ctx).Model(&entity.Address{}).
		Distinct("user_id").
		Count(&stats.UsersWithAddresses).Error; err != nil {
		return nil, err
	}

	// 计算平均每用户地址数
	if stats.UsersWithAddresses > 0 {
		stats.AverageAddressPerUser = float64(stats.TotalAddresses) / float64(stats.UsersWithAddresses)
	}

	// 获取热门省份（前10）
	var provinces []ProvinceStatistic
	if err := r.db.WithContext(ctx).Model(&entity.Address{}).
		Select("province, COUNT(*) as count").
		Group("province").
		Order("count DESC").
		Limit(10).
		Find(&provinces).Error; err != nil {
		return nil, err
	}
	stats.TopProvinces = provinces

	// 获取热门城市（前10）
	var cities []CityStatistic
	if err := r.db.WithContext(ctx).Model(&entity.Address{}).
		Select("province, city, COUNT(*) as count").
		Group("province, city").
		Order("count DESC").
		Limit(10).
		Find(&cities).Error; err != nil {
		return nil, err
	}
	stats.TopCities = cities

	return stats, nil
}