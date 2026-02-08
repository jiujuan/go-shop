package repository

import (
	"context"
	"errors"
	"time"

	"go-shop/internal/entity"

	"gorm.io/gorm"
)

// operationLogRepositoryImpl 操作日志仓库GORM实现
type operationLogRepositoryImpl struct {
	db *gorm.DB
}

// NewOperationLogRepository 创建操作日志仓库实例
func NewOperationLogRepository(db *gorm.DB) OperationLogRepository {
	return &operationLogRepositoryImpl{
		db: db,
	}
}

// Create 创建操作日志
func (r *operationLogRepositoryImpl) Create(ctx context.Context, log *entity.OperationLog) error {
	if log == nil {
		return errors.New("operation log cannot be nil")
	}

	// 验证日志数据
	if err := log.Validate(); err != nil {
		return err
	}

	// 创建日志
	if err := r.db.WithContext(ctx).Create(log).Error; err != nil {
		return err
	}

	return nil
}

// GetByID 根据ID获取操作日志
func (r *operationLogRepositoryImpl) GetByID(ctx context.Context, id int64) (*entity.OperationLog, error) {
	if id <= 0 {
		return nil, errors.New("invalid operation log id")
	}

	var log entity.OperationLog
	if err := r.db.WithContext(ctx).First(&log, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("operation log not found")
		}
		return nil, err
	}

	return &log, nil
}

// List 获取操作日志列表（支持分页）
func (r *operationLogRepositoryImpl) List(ctx context.Context, offset, limit int) ([]*entity.OperationLog, int64, error) {
	// 获取总数
	var total int64
	if err := r.db.WithContext(ctx).
		Model(&entity.OperationLog{}).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取日志列表
	var logs []*entity.OperationLog
	if err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// ListByUser 根据用户获取操作日志列表（支持分页）
func (r *operationLogRepositoryImpl) ListByUser(ctx context.Context, userID int64, offset, limit int) ([]*entity.OperationLog, int64, error) {
	if userID <= 0 {
		return nil, 0, errors.New("invalid user id")
	}

	// 获取总数
	var total int64
	if err := r.db.WithContext(ctx).
		Model(&entity.OperationLog{}).
		Where("user_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取日志列表
	var logs []*entity.OperationLog
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// ListByModule 根据模块获取操作日志列表（支持分页）
func (r *operationLogRepositoryImpl) ListByModule(ctx context.Context, module string, offset, limit int) ([]*entity.OperationLog, int64, error) {
	if module == "" {
		return nil, 0, errors.New("module is required")
	}

	// 获取总数
	var total int64
	if err := r.db.WithContext(ctx).
		Model(&entity.OperationLog{}).
		Where("module = ?", module).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取日志列表
	var logs []*entity.OperationLog
	if err := r.db.WithContext(ctx).
		Where("module = ?", module).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// ListByOperation 根据操作类型获取操作日志列表（支持分页）
func (r *operationLogRepositoryImpl) ListByOperation(ctx context.Context, operation string, offset, limit int) ([]*entity.OperationLog, int64, error) {
	if operation == "" {
		return nil, 0, errors.New("operation is required")
	}

	// 获取总数
	var total int64
	if err := r.db.WithContext(ctx).
		Model(&entity.OperationLog{}).
		Where("operation = ?", operation).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取日志列表
	var logs []*entity.OperationLog
	if err := r.db.WithContext(ctx).
		Where("operation = ?", operation).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// ListByTimeRange 根据时间范围获取操作日志列表（支持分页）
func (r *operationLogRepositoryImpl) ListByTimeRange(ctx context.Context, startTime, endTime time.Time, offset, limit int) ([]*entity.OperationLog, int64, error) {
	if startTime.After(endTime) {
		return nil, 0, errors.New("start time must be before end time")
	}

	// 获取总数
	var total int64
	if err := r.db.WithContext(ctx).
		Model(&entity.OperationLog{}).
		Where("created_at >= ? AND created_at <= ?", startTime, endTime).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取日志列表
	var logs []*entity.OperationLog
	if err := r.db.WithContext(ctx).
		Where("created_at >= ? AND created_at <= ?", startTime, endTime).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// ListWithFilters 根据多个条件筛选操作日志列表（支持分页）
func (r *operationLogRepositoryImpl) ListWithFilters(ctx context.Context, userID *int64, userType *entity.UserType, module *string, operation *string, startTime *time.Time, endTime *time.Time, offset, limit int) ([]*entity.OperationLog, int64, error) {
	// 构建查询
	query := r.db.WithContext(ctx).Model(&entity.OperationLog{})

	// 添加用户ID筛选
	if userID != nil {
		if *userID <= 0 {
			return nil, 0, errors.New("invalid user id")
		}
		query = query.Where("user_id = ?", *userID)
	}

	// 添加用户类型筛选
	if userType != nil {
		query = query.Where("user_type = ?", *userType)
	}

	// 添加模块筛选
	if module != nil {
		if *module == "" {
			return nil, 0, errors.New("module cannot be empty")
		}
		query = query.Where("module = ?", *module)
	}

	// 添加操作类型筛选
	if operation != nil {
		if *operation == "" {
			return nil, 0, errors.New("operation cannot be empty")
		}
		query = query.Where("operation = ?", *operation)
	}

	// 添加时间范围筛选
	if startTime != nil && endTime != nil {
		if startTime.After(*endTime) {
			return nil, 0, errors.New("start time must be before end time")
		}
		query = query.Where("created_at >= ? AND created_at <= ?", *startTime, *endTime)
	} else if startTime != nil {
		query = query.Where("created_at >= ?", *startTime)
	} else if endTime != nil {
		query = query.Where("created_at <= ?", *endTime)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取日志列表
	var logs []*entity.OperationLog
	if err := query.
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// Delete 删除操作日志
func (r *operationLogRepositoryImpl) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return errors.New("invalid operation log id")
	}

	// 检查日志是否存在
	_, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// 删除日志
	if err := r.db.WithContext(ctx).Delete(&entity.OperationLog{}, id).Error; err != nil {
		return err
	}

	return nil
}

// DeleteByUser 删除用户的所有操作日志
func (r *operationLogRepositoryImpl) DeleteByUser(ctx context.Context, userID int64) error {
	if userID <= 0 {
		return errors.New("invalid user id")
	}

	// 删除用户的所有日志
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&entity.OperationLog{}).Error; err != nil {
		return err
	}

	return nil
}

// DeleteOldLogs 删除超过指定天数的操作日志
func (r *operationLogRepositoryImpl) DeleteOldLogs(ctx context.Context, days int) error {
	if days <= 0 {
		return errors.New("invalid days parameter")
	}

	// 计算截止时间
	cutoffTime := time.Now().AddDate(0, 0, -days)

	// 删除旧日志
	if err := r.db.WithContext(ctx).
		Where("created_at < ?", cutoffTime).
		Delete(&entity.OperationLog{}).Error; err != nil {
		return err
	}

	return nil
}

// BatchCreate 批量创建操作日志
func (r *operationLogRepositoryImpl) BatchCreate(ctx context.Context, logs []*entity.OperationLog) error {
	if len(logs) == 0 {
		return nil
	}

	// 验证所有日志数据
	for _, log := range logs {
		if log == nil {
			return errors.New("operation log cannot be nil")
		}
		if err := log.Validate(); err != nil {
			return err
		}
	}

	// 批量创建日志
	if err := r.db.WithContext(ctx).Create(logs).Error; err != nil {
		return err
	}

	return nil
}

// CountByModule 统计指定模块的日志数量
func (r *operationLogRepositoryImpl) CountByModule(ctx context.Context, module string) (int64, error) {
	if module == "" {
		return 0, errors.New("module is required")
	}

	var count int64
	if err := r.db.WithContext(ctx).
		Model(&entity.OperationLog{}).
		Where("module = ?", module).
		Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

// CountByOperation 统计指定操作类型的日志数量
func (r *operationLogRepositoryImpl) CountByOperation(ctx context.Context, operation string) (int64, error) {
	if operation == "" {
		return 0, errors.New("operation is required")
	}

	var count int64
	if err := r.db.WithContext(ctx).
		Model(&entity.OperationLog{}).
		Where("operation = ?", operation).
		Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

// CountErrorLogs 统计错误日志数量
func (r *operationLogRepositoryImpl) CountErrorLogs(ctx context.Context, startTime, endTime time.Time) (int64, error) {
	if startTime.After(endTime) {
		return 0, errors.New("start time must be before end time")
	}

	var count int64
	if err := r.db.WithContext(ctx).
		Model(&entity.OperationLog{}).
		Where("created_at >= ? AND created_at <= ? AND error IS NOT NULL AND error != ''", startTime, endTime).
		Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}
