package repository

import (
	"context"
	"time"

	"go-shop/internal/entity"
)

// OperationLogRepository 操作日志仓库接口
type OperationLogRepository interface {
	// Create 创建操作日志
	Create(ctx context.Context, log *entity.OperationLog) error

	// GetByID 根据ID获取操作日志
	GetByID(ctx context.Context, id int64) (*entity.OperationLog, error)

	// List 获取操作日志列表（支持分页）
	List(ctx context.Context, offset, limit int) ([]*entity.OperationLog, int64, error)

	// ListByUser 根据用户获取操作日志列表（支持分页）
	ListByUser(ctx context.Context, userID int64, offset, limit int) ([]*entity.OperationLog, int64, error)

	// ListByModule 根据模块获取操作日志列表（支持分页）
	ListByModule(ctx context.Context, module string, offset, limit int) ([]*entity.OperationLog, int64, error)

	// ListByOperation 根据操作类型获取操作日志列表（支持分页）
	ListByOperation(ctx context.Context, operation string, offset, limit int) ([]*entity.OperationLog, int64, error)

	// ListByTimeRange 根据时间范围获取操作日志列表（支持分页）
	ListByTimeRange(ctx context.Context, startTime, endTime time.Time, offset, limit int) ([]*entity.OperationLog, int64, error)

	// ListWithFilters 根据多个条件筛选操作日志列表（支持分页）
	// 参数可为nil表示不筛选该条件
	ListWithFilters(ctx context.Context, userID *int64, userType *entity.UserType, module *string, operation *string, startTime *time.Time, endTime *time.Time, offset, limit int) ([]*entity.OperationLog, int64, error)

	// Delete 删除操作日志
	Delete(ctx context.Context, id int64) error

	// DeleteByUser 删除用户的所有操作日志
	DeleteByUser(ctx context.Context, userID int64) error

	// DeleteOldLogs 删除超过指定天数的操作日志
	DeleteOldLogs(ctx context.Context, days int) error

	// BatchCreate 批量创建操作日志
	BatchCreate(ctx context.Context, logs []*entity.OperationLog) error

	// CountByModule 统计指定模块的日志数量
	CountByModule(ctx context.Context, module string) (int64, error)

	// CountByOperation 统计指定操作类型的日志数量
	CountByOperation(ctx context.Context, operation string) (int64, error)

	// CountErrorLogs 统计错误日志数量
	CountErrorLogs(ctx context.Context, startTime, endTime time.Time) (int64, error)
}
