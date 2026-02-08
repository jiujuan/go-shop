package repository

import (
	"context"
	"errors"
	"time"

	"go-shop/internal/entity"

	"gorm.io/gorm"
)

// notificationRepositoryImpl 通知仓库GORM实现
type notificationRepositoryImpl struct {
	db *gorm.DB
}

// NewNotificationRepository 创建通知仓库实例
func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepositoryImpl{
		db: db,
	}
}

// Create 创建通知
func (r *notificationRepositoryImpl) Create(ctx context.Context, notification *entity.Notification) error {
	if notification == nil {
		return errors.New("notification cannot be nil")
	}

	// 验证通知数据
	if err := notification.Validate(); err != nil {
		return err
	}

	// 创建通知
	if err := r.db.WithContext(ctx).Create(notification).Error; err != nil {
		return err
	}

	return nil
}

// GetByID 根据ID获取通知
func (r *notificationRepositoryImpl) GetByID(ctx context.Context, id int64) (*entity.Notification, error) {
	if id <= 0 {
		return nil, errors.New("invalid notification id")
	}

	var notification entity.Notification
	if err := r.db.WithContext(ctx).First(&notification, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("notification not found")
		}
		return nil, err
	}

	return &notification, nil
}

// ListByUser 获取用户的通知列表（支持分页）
func (r *notificationRepositoryImpl) ListByUser(ctx context.Context, userID int64, offset, limit int) ([]*entity.Notification, int64, error) {
	if userID <= 0 {
		return nil, 0, errors.New("invalid user id")
	}

	// 获取总数
	var total int64
	if err := r.db.WithContext(ctx).
		Model(&entity.Notification{}).
		Where("user_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取通知列表
	var notifications []*entity.Notification
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&notifications).Error; err != nil {
		return nil, 0, err
	}

	return notifications, total, nil
}

// ListByUserAndType 根据用户和类型获取通知列表（支持分页）
func (r *notificationRepositoryImpl) ListByUserAndType(ctx context.Context, userID int64, notificationType entity.NotificationType, offset, limit int) ([]*entity.Notification, int64, error) {
	if userID <= 0 {
		return nil, 0, errors.New("invalid user id")
	}

	// 获取总数
	var total int64
	if err := r.db.WithContext(ctx).
		Model(&entity.Notification{}).
		Where("user_id = ? AND type = ?", userID, notificationType).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取通知列表
	var notifications []*entity.Notification
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND type = ?", userID, notificationType).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&notifications).Error; err != nil {
		return nil, 0, err
	}

	return notifications, total, nil
}

// ListByUserAndReadStatus 根据用户和已读状态获取通知列表（支持分页）
func (r *notificationRepositoryImpl) ListByUserAndReadStatus(ctx context.Context, userID int64, isRead bool, offset, limit int) ([]*entity.Notification, int64, error) {
	if userID <= 0 {
		return nil, 0, errors.New("invalid user id")
	}

	// 获取总数
	var total int64
	if err := r.db.WithContext(ctx).
		Model(&entity.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, isRead).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取通知列表
	var notifications []*entity.Notification
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_read = ?", userID, isRead).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&notifications).Error; err != nil {
		return nil, 0, err
	}

	return notifications, total, nil
}

// ListByUserWithFilters 根据用户获取通知列表（支持类型和已读状态筛选）
func (r *notificationRepositoryImpl) ListByUserWithFilters(ctx context.Context, userID int64, notificationType *entity.NotificationType, isRead *bool, offset, limit int) ([]*entity.Notification, int64, error) {
	if userID <= 0 {
		return nil, 0, errors.New("invalid user id")
	}

	// 构建查询
	query := r.db.WithContext(ctx).Model(&entity.Notification{}).Where("user_id = ?", userID)

	// 添加类型筛选
	if notificationType != nil {
		query = query.Where("type = ?", *notificationType)
	}

	// 添加已读状态筛选
	if isRead != nil {
		query = query.Where("is_read = ?", *isRead)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取通知列表
	var notifications []*entity.Notification
	if err := query.
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&notifications).Error; err != nil {
		return nil, 0, err
	}

	return notifications, total, nil
}

// MarkAsRead 标记通知为已读
func (r *notificationRepositoryImpl) MarkAsRead(ctx context.Context, id int64) error {
	if id <= 0 {
		return errors.New("invalid notification id")
	}

	// 检查通知是否存在
	_, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// 更新已读状态
	if err := r.db.WithContext(ctx).
		Model(&entity.Notification{}).
		Where("id = ?", id).
		Update("is_read", true).Error; err != nil {
		return err
	}

	return nil
}

// MarkAllAsRead 标记用户所有通知为已读
func (r *notificationRepositoryImpl) MarkAllAsRead(ctx context.Context, userID int64) error {
	if userID <= 0 {
		return errors.New("invalid user id")
	}

	// 更新所有未读通知为已读
	if err := r.db.WithContext(ctx).
		Model(&entity.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Update("is_read", true).Error; err != nil {
		return err
	}

	return nil
}

// GetUnreadCount 获取用户未读通知数量
func (r *notificationRepositoryImpl) GetUnreadCount(ctx context.Context, userID int64) (int64, error) {
	if userID <= 0 {
		return 0, errors.New("invalid user id")
	}

	var count int64
	if err := r.db.WithContext(ctx).
		Model(&entity.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

// GetUnreadCountByType 获取用户指定类型的未读通知数量
func (r *notificationRepositoryImpl) GetUnreadCountByType(ctx context.Context, userID int64, notificationType entity.NotificationType) (int64, error) {
	if userID <= 0 {
		return 0, errors.New("invalid user id")
	}

	var count int64
	if err := r.db.WithContext(ctx).
		Model(&entity.Notification{}).
		Where("user_id = ? AND type = ? AND is_read = ?", userID, notificationType, false).
		Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

// Delete 删除通知
func (r *notificationRepositoryImpl) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return errors.New("invalid notification id")
	}

	// 检查通知是否存在
	_, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// 删除通知
	if err := r.db.WithContext(ctx).Delete(&entity.Notification{}, id).Error; err != nil {
		return err
	}

	return nil
}

// DeleteByUser 删除用户的所有通知
func (r *notificationRepositoryImpl) DeleteByUser(ctx context.Context, userID int64) error {
	if userID <= 0 {
		return errors.New("invalid user id")
	}

	// 删除用户的所有通知
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&entity.Notification{}).Error; err != nil {
		return err
	}

	return nil
}

// DeleteOldNotifications 删除超过指定天数的通知
func (r *notificationRepositoryImpl) DeleteOldNotifications(ctx context.Context, days int) error {
	if days <= 0 {
		return errors.New("invalid days parameter")
	}

	// 计算截止时间
	cutoffTime := time.Now().AddDate(0, 0, -days)

	// 删除旧通知
	if err := r.db.WithContext(ctx).
		Where("created_at < ?", cutoffTime).
		Delete(&entity.Notification{}).Error; err != nil {
		return err
	}

	return nil
}

// BatchCreate 批量创建通知
func (r *notificationRepositoryImpl) BatchCreate(ctx context.Context, notifications []*entity.Notification) error {
	if len(notifications) == 0 {
		return nil
	}

	// 验证所有通知数据
	for _, notification := range notifications {
		if notification == nil {
			return errors.New("notification cannot be nil")
		}
		if err := notification.Validate(); err != nil {
			return err
		}
	}

	// 批量创建通知
	if err := r.db.WithContext(ctx).Create(notifications).Error; err != nil {
		return err
	}

	return nil
}
