package repository

import (
	"context"
	"errors"
	"time"

	"go-shop/internal/entity"

	"gorm.io/gorm"
)

// userRepositoryImpl 用户仓库GORM实现
type userRepositoryImpl struct {
	db *gorm.DB
}

// NewUserRepository 创建用户仓库实例
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepositoryImpl{
		db: db,
	}
}

// Create 创建用户
func (r *userRepositoryImpl) Create(ctx context.Context, user *entity.User) error {
	if user == nil {
		return errors.New("user cannot be nil")
	}

	// 检查用户名唯一性
	exists, err := r.ExistsByUsername(ctx, user.Username)
	if err != nil {
		return err
	}
	if exists {
		return entity.ErrUserAlreadyExists
	}

	// 检查邮箱唯一性
	if user.Email != "" {
		exists, err = r.ExistsByEmail(ctx, user.Email)
		if err != nil {
			return err
		}
		if exists {
			return entity.ErrEmailAlreadyExists
		}
	}

	// 创建用户
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return entity.ErrUserAlreadyExists
		}
		return err
	}

	return nil
}

// GetByID 根据ID获取用户
func (r *userRepositoryImpl) GetByID(ctx context.Context, id int64) (*entity.User, error) {
	if id <= 0 {
		return nil, entity.ErrInvalidUserID
	}

	var user entity.User
	if err := r.db.WithContext(ctx).First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entity.ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

// GetByUsername 根据用户名获取用户
func (r *userRepositoryImpl) GetByUsername(ctx context.Context, username string) (*entity.User, error) {
	if username == "" {
		return nil, entity.ErrInvalidUsername
	}

	var user entity.User
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entity.ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

// GetByEmail 根据邮箱获取用户
func (r *userRepositoryImpl) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	if email == "" {
		return nil, entity.ErrInvalidEmail
	}

	var user entity.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entity.ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

// GetByPhone 根据手机号获取用户
func (r *userRepositoryImpl) GetByPhone(ctx context.Context, phone string) (*entity.User, error) {
	if phone == "" {
		return nil, errors.New("phone cannot be empty")
	}

	var user entity.User
	if err := r.db.WithContext(ctx).Where("phone = ?", phone).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entity.ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

// Update 更新用户信息
func (r *userRepositoryImpl) Update(ctx context.Context, user *entity.User) error {
	if user == nil {
		return errors.New("user cannot be nil")
	}
	if user.ID <= 0 {
		return entity.ErrInvalidUserID
	}

	// 检查用户是否存在
	_, err := r.GetByID(ctx, user.ID)
	if err != nil {
		return err
	}

	// 检查用户名唯一性（排除当前用户）
	if user.Username != "" {
		exists, err := r.ExistsByUsernameExcludeID(ctx, user.Username, user.ID)
		if err != nil {
			return err
		}
		if exists {
			return entity.ErrUserAlreadyExists
		}
	}

	// 检查邮箱唯一性（排除当前用户）
	if user.Email != "" {
		exists, err := r.ExistsByEmailExcludeID(ctx, user.Email, user.ID)
		if err != nil {
			return err
		}
		if exists {
			return entity.ErrEmailAlreadyExists
		}
	}

	// 更新用户信息（排除密码字段）
	updates := map[string]interface{}{
		"updated_at": time.Now(),
	}

	if user.Username != "" {
		updates["username"] = user.Username
	}
	if user.Email != "" {
		updates["email"] = user.Email
	}
	if user.Phone != "" {
		updates["phone"] = user.Phone
	}
	// 注意：不在这里更新密码，密码更新使用专门的方法

	if err := r.db.WithContext(ctx).Model(&entity.User{}).Where("id = ?", user.ID).Updates(updates).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return entity.ErrUserAlreadyExists
		}
		return err
	}

	return nil
}

// UpdatePassword 更新用户密码
func (r *userRepositoryImpl) UpdatePassword(ctx context.Context, userID int64, hashedPassword string) error {
	if userID <= 0 {
		return entity.ErrInvalidUserID
	}
	if hashedPassword == "" {
		return errors.New("hashed password cannot be empty")
	}

	// 检查用户是否存在
	_, err := r.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	// 更新密码
	if err := r.db.WithContext(ctx).Model(&entity.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"password":   hashedPassword,
		"updated_at": time.Now(),
	}).Error; err != nil {
		return err
	}

	return nil
}

// Delete 删除用户（软删除）
func (r *userRepositoryImpl) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return entity.ErrInvalidUserID
	}

	// 检查用户是否存在
	_, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// 软删除用户
	if err := r.db.WithContext(ctx).Delete(&entity.User{}, id).Error; err != nil {
		return err
	}

	return nil
}

// List 获取用户列表（分页）
func (r *userRepositoryImpl) List(ctx context.Context, offset, limit int) ([]*entity.User, int64, error) {
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = 20
	}

	var users []*entity.User
	var total int64

	// 获取总数
	if err := r.db.WithContext(ctx).Model(&entity.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取用户列表
	if err := r.db.WithContext(ctx).
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// ExistsByUsername 检查用户名是否存在
func (r *userRepositoryImpl) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	if username == "" {
		return false, nil
	}

	var count int64
	if err := r.db.WithContext(ctx).Model(&entity.User{}).Where("username = ?", username).Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

// ExistsByEmail 检查邮箱是否存在
func (r *userRepositoryImpl) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	if email == "" {
		return false, nil
	}

	var count int64
	if err := r.db.WithContext(ctx).Model(&entity.User{}).Where("email = ?", email).Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

// ExistsByPhone 检查手机号是否存在
func (r *userRepositoryImpl) ExistsByPhone(ctx context.Context, phone string) (bool, error) {
	if phone == "" {
		return false, nil
	}

	var count int64
	if err := r.db.WithContext(ctx).Model(&entity.User{}).Where("phone = ?", phone).Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

// ExistsByUsernameExcludeID 检查用户名是否存在（排除指定ID）
func (r *userRepositoryImpl) ExistsByUsernameExcludeID(ctx context.Context, username string, excludeID int64) (bool, error) {
	if username == "" {
		return false, nil
	}

	var count int64
	if err := r.db.WithContext(ctx).Model(&entity.User{}).
		Where("username = ? AND id != ?", username, excludeID).
		Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

// ExistsByEmailExcludeID 检查邮箱是否存在（排除指定ID）
func (r *userRepositoryImpl) ExistsByEmailExcludeID(ctx context.Context, email string, excludeID int64) (bool, error) {
	if email == "" {
		return false, nil
	}

	var count int64
	if err := r.db.WithContext(ctx).Model(&entity.User{}).
		Where("email = ? AND id != ?", email, excludeID).
		Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

// GetAdminUsers 获取管理员用户列表
func (r *userRepositoryImpl) GetAdminUsers(ctx context.Context) ([]*entity.User, error) {
	var users []*entity.User
	if err := r.db.WithContext(ctx).
		Where("is_admin = ?", true).
		Order("created_at DESC").
		Find(&users).Error; err != nil {
		return nil, err
	}

	return users, nil
}

// CountUsers 统计用户总数
func (r *userRepositoryImpl) CountUsers(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entity.User{}).Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

// CountActiveUsers 统计活跃用户数（最近30天有活动）
func (r *userRepositoryImpl) CountActiveUsers(ctx context.Context) (int64, error) {
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	
	var count int64
	if err := r.db.WithContext(ctx).Model(&entity.User{}).
		Where("updated_at >= ?", thirtyDaysAgo).
		Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}