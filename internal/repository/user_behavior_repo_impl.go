package repository

import (
	"context"
	"errors"
	"time"

	"go-shop/internal/entity"

	"gorm.io/gorm"
)

// userBehaviorRepositoryImpl 用户行为仓库GORM实现
type userBehaviorRepositoryImpl struct {
	db *gorm.DB
}

// NewUserBehaviorRepository 创建用户行为仓库实例
func NewUserBehaviorRepository(db *gorm.DB) UserBehaviorRepository {
	return &userBehaviorRepositoryImpl{
		db: db,
	}
}

// Create 创建用户行为记录
func (r *userBehaviorRepositoryImpl) Create(ctx context.Context, behavior *entity.UserBehavior) error {
	if behavior == nil {
		return errors.New("behavior cannot be nil")
	}

	// 验证行为数据
	if err := behavior.Validate(); err != nil {
		return err
	}

	// 创建行为记录
	if err := r.db.WithContext(ctx).Create(behavior).Error; err != nil {
		return err
	}

	return nil
}

// GetByID 根据ID获取用户行为
func (r *userBehaviorRepositoryImpl) GetByID(ctx context.Context, id int64) (*entity.UserBehavior, error) {
	if id <= 0 {
		return nil, errors.New("invalid behavior id")
	}

	var behavior entity.UserBehavior
	if err := r.db.WithContext(ctx).First(&behavior, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("behavior not found")
		}
		return nil, err
	}

	return &behavior, nil
}

// ListByUser 获取用户的行为列表
func (r *userBehaviorRepositoryImpl) ListByUser(ctx context.Context, userID int64, offset, limit int) ([]*entity.UserBehavior, int64, error) {
	if userID <= 0 {
		return nil, 0, errors.New("invalid user id")
	}

	// 获取总数
	var total int64
	if err := r.db.WithContext(ctx).
		Model(&entity.UserBehavior{}).
		Where("user_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取行为列表
	var behaviors []*entity.UserBehavior
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&behaviors).Error; err != nil {
		return nil, 0, err
	}

	return behaviors, total, nil
}

// ListByUserAndType 获取用户指定类型的行为列表
func (r *userBehaviorRepositoryImpl) ListByUserAndType(ctx context.Context, userID int64, actionType entity.ActionType, offset, limit int) ([]*entity.UserBehavior, int64, error) {
	if userID <= 0 {
		return nil, 0, errors.New("invalid user id")
	}

	// 获取总数
	var total int64
	if err := r.db.WithContext(ctx).
		Model(&entity.UserBehavior{}).
		Where("user_id = ? AND action_type = ?", userID, actionType).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取行为列表
	var behaviors []*entity.UserBehavior
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND action_type = ?", userID, actionType).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&behaviors).Error; err != nil {
		return nil, 0, err
	}

	return behaviors, total, nil
}

// ListByProduct 获取商品的行为列表
func (r *userBehaviorRepositoryImpl) ListByProduct(ctx context.Context, productID int64, offset, limit int) ([]*entity.UserBehavior, int64, error) {
	if productID <= 0 {
		return nil, 0, entity.ErrInvalidProductID
	}

	// 获取总数
	var total int64
	if err := r.db.WithContext(ctx).
		Model(&entity.UserBehavior{}).
		Where("product_id = ?", productID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取行为列表
	var behaviors []*entity.UserBehavior
	if err := r.db.WithContext(ctx).
		Where("product_id = ?", productID).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&behaviors).Error; err != nil {
		return nil, 0, err
	}

	return behaviors, total, nil
}

// ListRecentByUser 获取用户最近的行为记录
func (r *userBehaviorRepositoryImpl) ListRecentByUser(ctx context.Context, userID int64, limit int) ([]*entity.UserBehavior, error) {
	if userID <= 0 {
		return nil, errors.New("invalid user id")
	}

	var behaviors []*entity.UserBehavior
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&behaviors).Error; err != nil {
		return nil, err
	}

	return behaviors, nil
}

// CountByUser 统计用户行为数量
func (r *userBehaviorRepositoryImpl) CountByUser(ctx context.Context, userID int64) (int64, error) {
	if userID <= 0 {
		return 0, errors.New("invalid user id")
	}

	var count int64
	if err := r.db.WithContext(ctx).
		Model(&entity.UserBehavior{}).
		Where("user_id = ?", userID).
		Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

// CountByUserAndType 统计用户指定类型的行为数量
func (r *userBehaviorRepositoryImpl) CountByUserAndType(ctx context.Context, userID int64, actionType entity.ActionType) (int64, error) {
	if userID <= 0 {
		return 0, errors.New("invalid user id")
	}

	var count int64
	if err := r.db.WithContext(ctx).
		Model(&entity.UserBehavior{}).
		Where("user_id = ? AND action_type = ?", userID, actionType).
		Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

// CountByProduct 统计商品的行为数量
func (r *userBehaviorRepositoryImpl) CountByProduct(ctx context.Context, productID int64) (int64, error) {
	if productID <= 0 {
		return 0, entity.ErrInvalidProductID
	}

	var count int64
	if err := r.db.WithContext(ctx).
		Model(&entity.UserBehavior{}).
		Where("product_id = ?", productID).
		Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

// DeleteOldRecords 删除指定时间之前的记录
func (r *userBehaviorRepositoryImpl) DeleteOldRecords(ctx context.Context, before time.Time) error {
	if err := r.db.WithContext(ctx).
		Where("created_at < ?", before).
		Delete(&entity.UserBehavior{}).Error; err != nil {
		return err
	}

	return nil
}

// DeleteExcessRecords 删除用户超出限制的旧记录（保留最近的N条）
func (r *userBehaviorRepositoryImpl) DeleteExcessRecords(ctx context.Context, userID int64, keepCount int) error {
	if userID <= 0 {
		return errors.New("invalid user id")
	}
	if keepCount <= 0 {
		return errors.New("keep count must be positive")
	}

	// 获取用户的行为总数
	count, err := r.CountByUser(ctx, userID)
	if err != nil {
		return err
	}

	// 如果记录数未超过限制，无需删除
	if count <= int64(keepCount) {
		return nil
	}

	// 查询需要保留的最新记录的ID列表
	var keepIDs []int64
	if err := r.db.WithContext(ctx).
		Model(&entity.UserBehavior{}).
		Select("id").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(keepCount).
		Pluck("id", &keepIDs).Error; err != nil {
		return err
	}

	// 删除不在保留列表中的记录
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND id NOT IN ?", userID, keepIDs).
		Delete(&entity.UserBehavior{}).Error; err != nil {
		return err
	}

	return nil
}

// GetUserViewHistory 获取用户浏览历史（仅view类型）
func (r *userBehaviorRepositoryImpl) GetUserViewHistory(ctx context.Context, userID int64, limit int) ([]*entity.UserBehavior, error) {
	if userID <= 0 {
		return nil, errors.New("invalid user id")
	}

	var behaviors []*entity.UserBehavior
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND action_type = ?", userID, entity.ActionTypeView).
		Order("created_at DESC").
		Limit(limit).
		Find(&behaviors).Error; err != nil {
		return nil, err
	}

	return behaviors, nil
}

// GetUserViewedProductIDs 获取用户浏览过的商品ID列表
func (r *userBehaviorRepositoryImpl) GetUserViewedProductIDs(ctx context.Context, userID int64, limit int) ([]int64, error) {
	if userID <= 0 {
		return nil, errors.New("invalid user id")
	}

	var productIDs []int64
	if err := r.db.WithContext(ctx).
		Model(&entity.UserBehavior{}).
		Select("DISTINCT product_id").
		Where("user_id = ? AND action_type = ?", userID, entity.ActionTypeView).
		Order("created_at DESC").
		Limit(limit).
		Pluck("product_id", &productIDs).Error; err != nil {
		return nil, err
	}

	return productIDs, nil
}

// BatchCreate 批量创建用户行为记录
func (r *userBehaviorRepositoryImpl) BatchCreate(ctx context.Context, behaviors []*entity.UserBehavior) error {
	if len(behaviors) == 0 {
		return nil
	}

	// 验证所有行为数据
	for _, behavior := range behaviors {
		if err := behavior.Validate(); err != nil {
			return err
		}
	}

	// 批量创建
	if err := r.db.WithContext(ctx).Create(behaviors).Error; err != nil {
		return err
	}

	return nil
}

// DeleteByUser 删除用户的所有行为记录
func (r *userBehaviorRepositoryImpl) DeleteByUser(ctx context.Context, userID int64) error {
	if userID <= 0 {
		return errors.New("invalid user id")
	}

	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&entity.UserBehavior{}).Error; err != nil {
		return err
	}

	return nil
}

// DeleteByProduct 删除商品的所有行为记录
func (r *userBehaviorRepositoryImpl) DeleteByProduct(ctx context.Context, productID int64) error {
	if productID <= 0 {
		return entity.ErrInvalidProductID
	}

	if err := r.db.WithContext(ctx).
		Where("product_id = ?", productID).
		Delete(&entity.UserBehavior{}).Error; err != nil {
		return err
	}

	return nil
}
