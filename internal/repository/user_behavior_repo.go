package repository

import (
	"context"
	"time"

	"go-shop/internal/entity"
)

// UserBehaviorRepository 用户行为仓库接口
type UserBehaviorRepository interface {
	// Create 创建用户行为记录
	Create(ctx context.Context, behavior *entity.UserBehavior) error

	// GetByID 根据ID获取用户行为
	GetByID(ctx context.Context, id int64) (*entity.UserBehavior, error)

	// ListByUser 获取用户的行为列表
	ListByUser(ctx context.Context, userID int64, offset, limit int) ([]*entity.UserBehavior, int64, error)

	// ListByUserAndType 获取用户指定类型的行为列表
	ListByUserAndType(ctx context.Context, userID int64, actionType entity.ActionType, offset, limit int) ([]*entity.UserBehavior, int64, error)

	// ListByProduct 获取商品的行为列表
	ListByProduct(ctx context.Context, productID int64, offset, limit int) ([]*entity.UserBehavior, int64, error)

	// ListRecentByUser 获取用户最近的行为记录
	ListRecentByUser(ctx context.Context, userID int64, limit int) ([]*entity.UserBehavior, error)

	// CountByUser 统计用户行为数量
	CountByUser(ctx context.Context, userID int64) (int64, error)

	// CountByUserAndType 统计用户指定类型的行为数量
	CountByUserAndType(ctx context.Context, userID int64, actionType entity.ActionType) (int64, error)

	// CountByProduct 统计商品的行为数量
	CountByProduct(ctx context.Context, productID int64) (int64, error)

	// DeleteOldRecords 删除指定时间之前的记录
	DeleteOldRecords(ctx context.Context, before time.Time) error

	// DeleteExcessRecords 删除用户超出限制的旧记录（保留最近的N条）
	DeleteExcessRecords(ctx context.Context, userID int64, keepCount int) error

	// GetUserViewHistory 获取用户浏览历史（仅view类型）
	GetUserViewHistory(ctx context.Context, userID int64, limit int) ([]*entity.UserBehavior, error)

	// GetUserViewedProductIDs 获取用户浏览过的商品ID列表
	GetUserViewedProductIDs(ctx context.Context, userID int64, limit int) ([]int64, error)

	// BatchCreate 批量创建用户行为记录
	BatchCreate(ctx context.Context, behaviors []*entity.UserBehavior) error

	// DeleteByUser 删除用户的所有行为记录
	DeleteByUser(ctx context.Context, userID int64) error

	// DeleteByProduct 删除商品的所有行为记录
	DeleteByProduct(ctx context.Context, productID int64) error
}
