package service

import (
	"context"
	"errors"
	"strings"

	"go-shop/internal/dto"
	"go-shop/internal/entity"
	"go-shop/internal/repository"
	"go-shop/pkg/auth"
)

// UserService 用户业务服务
type UserService struct {
	userRepo        repository.UserRepository
	passwordManager *auth.PasswordManager
	jwtManager      *auth.JWTManager
}

// NewUserService 创建用户服务实例
func NewUserService(
	userRepo repository.UserRepository,
	passwordManager *auth.PasswordManager,
	jwtManager *auth.JWTManager,
) *UserService {
	return &UserService{
		userRepo:        userRepo,
		passwordManager: passwordManager,
		jwtManager:      jwtManager,
	}
}

// Register 用户注册
func (s *UserService) Register(ctx context.Context, req *dto.UserRegisterRequest) (*dto.UserResponse, error) {
	if err := s.validateRegisterRequest(req); err != nil {
		return nil, err
	}

	// 检查用户名是否已存在
	exists, err := s.userRepo.ExistsByUsername(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, entity.ErrUserAlreadyExists
	}

	// 检查邮箱是否已存在
	exists, err = s.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, entity.ErrEmailAlreadyExists
	}

	// 加密密码
	hashedPassword, err := s.passwordManager.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// 创建用户实体
	user := &entity.User{
		Username: req.Username,
		Password: hashedPassword,
		Email:    req.Email,
		IsAdmin:  false, // 默认为普通用户
	}

	// 保存用户
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// 返回用户信息（不包含密码）
	return s.entityToResponse(user), nil
}

// Login 用户登录
func (s *UserService) Login(ctx context.Context, req *dto.UserLoginRequest) (*dto.UserLoginResponse, error) {
	if err := s.validateLoginRequest(req); err != nil {
		return nil, err
	}

	// 根据用户名获取用户
	user, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		if errors.Is(err, entity.ErrUserNotFound) {
			return nil, entity.ErrInvalidPassword // 不暴露用户是否存在的信息
		}
		return nil, err
	}

	// 验证密码
	valid, err := s.passwordManager.VerifyPassword(req.Password, user.Password)
	if err != nil {
		return nil, err
	}
	if !valid {
		return nil, entity.ErrInvalidPassword
	}

	// 生成JWT Token
	token, err := s.jwtManager.GenerateToken(user.ID, user.Username, user.IsAdmin)
	if err != nil {
		return nil, err
	}

	// 返回登录响应
	return &dto.UserLoginResponse{
		User:  *s.entityToResponse(user),
		Token: token,
	}, nil
}

// GetUserByID 根据ID获取用户信息
func (s *UserService) GetUserByID(ctx context.Context, userID int64) (*dto.UserResponse, error) {
	if userID <= 0 {
		return nil, entity.ErrInvalidUserID
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return s.entityToResponse(user), nil
}

// UpdateUser 更新用户信息
func (s *UserService) UpdateUser(ctx context.Context, userID int64, req *dto.UserUpdateRequest) (*dto.UserResponse, error) {
	if userID <= 0 {
		return nil, entity.ErrInvalidUserID
	}

	// 验证输入参数
	if err := s.validateUpdateRequest(req); err != nil {
		return nil, err
	}

	// 获取当前用户信息
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 更新用户信息
	if req.Email != "" {
		user.Email = req.Email
	}

	// 保存更新
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	// 返回更新后的用户信息
	return s.entityToResponse(user), nil
}

// UpdatePassword 更新用户密码
func (s *UserService) UpdatePassword(ctx context.Context, userID int64, oldPassword, newPassword string) error {
	if userID <= 0 {
		return entity.ErrInvalidUserID
	}

	// 验证密码参数
	if err := s.validatePassword(oldPassword); err != nil {
		return err
	}
	if err := s.validatePassword(newPassword); err != nil {
		return err
	}

	// 获取用户信息
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	// 验证旧密码
	valid, err := s.passwordManager.VerifyPassword(oldPassword, user.Password)
	if err != nil {
		return err
	}
	if !valid {
		return entity.ErrInvalidPassword
	}

	// 加密新密码
	hashedPassword, err := s.passwordManager.HashPassword(newPassword)
	if err != nil {
		return err
	}

	// 更新密码
	return s.userRepo.UpdatePassword(ctx, userID, hashedPassword)
}

// GetUserList 获取用户列表（管理员功能）
func (s *UserService) GetUserList(ctx context.Context, page, pageSize int) (*dto.UserListResponse, error) {
	// 设置默认分页参数
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	} else if pageSize > 100 {
		pageSize = 100
	}

	// 计算偏移量
	offset := (page - 1) * pageSize

	// 获取用户列表
	users, total, err := s.userRepo.List(ctx, offset, pageSize)
	if err != nil {
		return nil, err
	}

	// 转换为响应格式
	userResponses := make([]dto.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = *s.entityToResponse(user)
	}

	// 创建分页响应
	pagination := dto.NewPaginationResponse(page, pageSize, total)

	return &dto.UserListResponse{
		Users:      userResponses,
		Pagination: pagination,
	}, nil
}

// DeleteUser 删除用户（管理员功能）
func (s *UserService) DeleteUser(ctx context.Context, userID int64) error {
	if userID <= 0 {
		return entity.ErrInvalidUserID
	}

	return s.userRepo.Delete(ctx, userID)
}

// ResetUserPassword 重置用户密码（管理员功能）
func (s *UserService) ResetUserPassword(ctx context.Context, userID int64, newPassword string) error {
	if userID <= 0 {
		return entity.ErrInvalidUserID
	}

	// 验证新密码
	if err := s.validatePassword(newPassword); err != nil {
		return err
	}

	// 加密新密码
	hashedPassword, err := s.passwordManager.HashPassword(newPassword)
	if err != nil {
		return err
	}

	// 更新密码
	return s.userRepo.UpdatePassword(ctx, userID, hashedPassword)
}

// validateRegisterRequest 验证注册请求
func (s *UserService) validateRegisterRequest(req *dto.UserRegisterRequest) error {
	if req == nil {
		return errors.New("注册请求不能为空")
	}

	// 验证用户名
	if err := s.validateUsername(req.Username); err != nil {
		return err
	}

	// 验证密码
	if err := s.validatePassword(req.Password); err != nil {
		return err
	}

	// 验证邮箱
	if err := s.validateEmail(req.Email); err != nil {
		return err
	}

	return nil
}

// validateLoginRequest 验证登录请求
func (s *UserService) validateLoginRequest(req *dto.UserLoginRequest) error {
	if req == nil {
		return errors.New("登录请求不能为空")
	}

	if strings.TrimSpace(req.Username) == "" {
		return errors.New("用户名不能为空")
	}

	if strings.TrimSpace(req.Password) == "" {
		return errors.New("密码不能为空")
	}

	return nil
}

// validateUpdateRequest 验证更新请求
func (s *UserService) validateUpdateRequest(req *dto.UserUpdateRequest) error {
	if req == nil {
		return errors.New("更新请求不能为空")
	}

	// 验证邮箱（如果提供）
	if req.Email != "" {
		if err := s.validateEmail(req.Email); err != nil {
			return err
		}
	}

	return nil
}

// validateUsername 验证用户名
func (s *UserService) validateUsername(username string) error {
	username = strings.TrimSpace(username)
	if username == "" {
		return errors.New("用户名不能为空")
	}
	if len(username) < 3 {
		return errors.New("用户名长度不能少于3个字符")
	}
	if len(username) > 50 {
		return errors.New("用户名长度不能超过50个字符")
	}
	// 可以添加更多用户名格式验证规则
	return nil
}

// validatePassword 验证密码
func (s *UserService) validatePassword(password string) error {
	if strings.TrimSpace(password) == "" {
		return errors.New("密码不能为空")
	}
	if len(password) < 6 {
		return errors.New("密码长度不能少于6个字符")
	}
	if len(password) > 100 {
		return errors.New("密码长度不能超过100个字符")
	}
	// 可以添加更多密码强度验证规则
	return nil
}

// validateEmail 验证邮箱
func (s *UserService) validateEmail(email string) error {
	email = strings.TrimSpace(email)
	if email == "" {
		return errors.New("邮箱不能为空")
	}
	if len(email) > 100 {
		return errors.New("邮箱长度不能超过100个字符")
	}
	// 简单的邮箱格式验证
	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		return entity.ErrInvalidEmail
	}
	return nil
}

// entityToResponse 将用户实体转换为响应DTO
func (s *UserService) entityToResponse(user *entity.User) *dto.UserResponse {
	return &dto.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Phone:     user.Phone,
		IsAdmin:   user.IsAdmin,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

// LoginOrRegisterByPhone 通过手机号登录或注册
// 需求：9.6, 9.7
func (s *UserService) LoginOrRegisterByPhone(ctx context.Context, phone string) (*dto.UserLoginResponse, error) {
	if phone == "" {
		return nil, errors.New("手机号不能为空")
	}

	// 尝试根据手机号获取用户
	user, err := s.userRepo.GetByPhone(ctx, phone)
	if err != nil {
		// 如果用户不存在，则自动注册
		if errors.Is(err, entity.ErrUserNotFound) {
			// 创建新用户
			newUser := &entity.User{
				Username: phone,                    // 使用手机号作为用户名
				Phone:    phone,
				Password: "",                       // 短信登录不需要密码
				Email:    "",                       // 邮箱可选
				IsAdmin:  false,
			}

			// 保存用户
			if err := s.userRepo.Create(ctx, newUser); err != nil {
				return nil, err
			}

			user = newUser
		} else {
			return nil, err
		}
	}

	// 生成JWT Token
	token, err := s.jwtManager.GenerateToken(user.ID, user.Username, user.IsAdmin)
	if err != nil {
		return nil, err
	}

	// 返回登录响应
	return &dto.UserLoginResponse{
		User:  *s.entityToResponse(user),
		Token: token,
	}, nil
}

// BindPhone 绑定手机号
// 需求：9.8
func (s *UserService) BindPhone(ctx context.Context, userID int64, phone string) error {
	if userID <= 0 {
		return entity.ErrInvalidUserID
	}

	if phone == "" {
		return errors.New("手机号不能为空")
	}

	// 检查手机号是否已被其他用户使用
	exists, err := s.userRepo.ExistsByPhone(ctx, phone)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("该手机号已被其他用户绑定")
	}

	// 获取用户信息
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	// 更新手机号
	user.Phone = phone
	if err := s.userRepo.Update(ctx, user); err != nil {
		return err
	}

	return nil
}
