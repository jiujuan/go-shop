package repository

import (
	"context"

	"go-shop/internal/entity"
)

// UserRepository 用户仓库接口
type UserRepository interface {
	// Create 创建用户
	Create(ctx context.Context, user *entity.User) error

	// GetByID 根据ID获取用户
	GetByID(ctx context.Context, id int64) (*entity.User, error)

	// GetByUsername 根据用户名获取用户
	GetByUsername(ctx context.Context, username string) (*entity.User, error)

	// GetByEmail 根据邮箱获取用户
	GetByEmail(ctx context.Context, email string) (*entity.User, error)

	// GetByPhone 根据手机号获取用户
	GetByPhone(ctx context.Context, phone string) (*entity.User, error)

	// Update 更新用户信息
	Update(ctx context.Context, user *entity.User) error

	// UpdatePassword 更新用户密码
	UpdatePassword(ctx context.Context, userID int64, hashedPassword string) error

	// Delete 删除用户（软删除）
	Delete(ctx context.Context, id int64) error

	// List 获取用户列表（分页）
	List(ctx context.Context, offset, limit int) ([]*entity.User, int64, error)

	// ExistsByUsername 检查用户名是否存在
	ExistsByUsername(ctx context.Context, username string) (bool, error)

	// ExistsByEmail 检查邮箱是否存在
	ExistsByEmail(ctx context.Context, email string) (bool, error)

	// ExistsByPhone 检查手机号是否存在
	ExistsByPhone(ctx context.Context, phone string) (bool, error)

	// ExistsByUsernameExcludeID 检查用户名是否存在（排除指定ID）
	ExistsByUsernameExcludeID(ctx context.Context, username string, excludeID int64) (bool, error)

	// ExistsByEmailExcludeID 检查邮箱是否存在（排除指定ID）
	ExistsByEmailExcludeID(ctx context.Context, email string, excludeID int64) (bool, error)

	// GetAdminUsers 获取管理员用户列表
	GetAdminUsers(ctx context.Context) ([]*entity.User, error)

	// CountUsers 统计用户总数
	CountUsers(ctx context.Context) (int64, error)

	// CountActiveUsers 统计活跃用户数（最近30天有活动）
	CountActiveUsers(ctx context.Context) (int64, error)
}