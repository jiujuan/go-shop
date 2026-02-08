package service

import (
	"context"
	"testing"

	"go-shop/internal/dto"
	"go-shop/internal/entity"
	"go-shop/internal/repository"
	"go-shop/pkg/auth"
	"go-shop/pkg/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// UserServiceTestSuite 用户服务测试套件
type UserServiceTestSuite struct {
	suite.Suite
	userService *UserService
	userRepo    repository.UserRepository
	ctx         context.Context
}

// SetupSuite 设置测试套件
func (suite *UserServiceTestSuite) SetupSuite() {
	// 初始化测试数据库
	db := utils.InitTestDB()
	
	// 创建仓库实例
	suite.userRepo = repository.NewUserRepository(db)
	
	// 创建认证管理器
	passwordManager := auth.NewPasswordManager(nil)
	jwtManager := auth.NewJWTManager("test-secret", 24)
	
	// 创建服务实例
	suite.userService = NewUserService(suite.userRepo, passwordManager, jwtManager)
	
	suite.ctx = context.Background()
}

// TearDownSuite 清理测试套件
func (suite *UserServiceTestSuite) TearDownSuite() {
	utils.CleanupTestDB()
}

// SetupTest 每个测试前的设置
func (suite *UserServiceTestSuite) SetupTest() {
	utils.CleanupTestData()
}

// TestRegister 测试用户注册
func (suite *UserServiceTestSuite) TestRegister() {
	// 测试成功注册
	req := &dto.UserRegisterRequest{
		Username: "testuser",
		Password: "password123",
		Email:    "test@example.com",
	}
	
	user, err := suite.userService.Register(suite.ctx, req)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), user)
	assert.Equal(suite.T(), req.Username, user.Username)
	assert.Equal(suite.T(), req.Email, user.Email)
	assert.False(suite.T(), user.IsAdmin)
	assert.NotZero(suite.T(), user.ID)
	
	// 测试用户名重复
	_, err = suite.userService.Register(suite.ctx, req)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), entity.ErrUserAlreadyExists, err)
	
	// 测试邮箱重复
	req2 := &dto.UserRegisterRequest{
		Username: "testuser2",
		Password: "password123",
		Email:    "test@example.com", // 相同邮箱
	}
	_, err = suite.userService.Register(suite.ctx, req2)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), entity.ErrEmailAlreadyExists, err)
}

// TestLogin 测试用户登录
func (suite *UserServiceTestSuite) TestLogin() {
	// 先注册一个用户
	registerReq := &dto.UserRegisterRequest{
		Username: "loginuser",
		Password: "password123",
		Email:    "login@example.com",
	}
	
	_, err := suite.userService.Register(suite.ctx, registerReq)
	assert.NoError(suite.T(), err)
	
	// 测试成功登录
	loginReq := &dto.UserLoginRequest{
		Username: "loginuser",
		Password: "password123",
	}
	
	loginResp, err := suite.userService.Login(suite.ctx, loginReq)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), loginResp)
	assert.Equal(suite.T(), registerReq.Username, loginResp.User.Username)
	assert.NotEmpty(suite.T(), loginResp.Token)
	
	// 测试错误密码
	loginReq.Password = "wrongpassword"
	_, err = suite.userService.Login(suite.ctx, loginReq)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), entity.ErrInvalidPassword, err)
	
	// 测试不存在的用户
	loginReq.Username = "nonexistent"
	loginReq.Password = "password123"
	_, err = suite.userService.Login(suite.ctx, loginReq)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), entity.ErrInvalidPassword, err) // 不暴露用户是否存在
}

// TestGetUserByID 测试根据ID获取用户
func (suite *UserServiceTestSuite) TestGetUserByID() {
	// 先注册一个用户
	registerReq := &dto.UserRegisterRequest{
		Username: "getuser",
		Password: "password123",
		Email:    "get@example.com",
	}
	
	user, err := suite.userService.Register(suite.ctx, registerReq)
	assert.NoError(suite.T(), err)
	
	// 测试获取用户
	foundUser, err := suite.userService.GetUserByID(suite.ctx, user.ID)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), foundUser)
	assert.Equal(suite.T(), user.ID, foundUser.ID)
	assert.Equal(suite.T(), user.Username, foundUser.Username)
	
	// 测试获取不存在的用户
	_, err = suite.userService.GetUserByID(suite.ctx, 99999)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), entity.ErrUserNotFound, err)
}

// TestUpdateUser 测试更新用户信息
func (suite *UserServiceTestSuite) TestUpdateUser() {
	// 先注册一个用户
	registerReq := &dto.UserRegisterRequest{
		Username: "updateuser",
		Password: "password123",
		Email:    "update@example.com",
	}
	
	user, err := suite.userService.Register(suite.ctx, registerReq)
	assert.NoError(suite.T(), err)
	
	// 测试更新邮箱
	updateReq := &dto.UserUpdateRequest{
		Email: "newemail@example.com",
	}
	
	updatedUser, err := suite.userService.UpdateUser(suite.ctx, user.ID, updateReq)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), updatedUser)
	assert.Equal(suite.T(), updateReq.Email, updatedUser.Email)
	assert.Equal(suite.T(), user.Username, updatedUser.Username) // 用户名不变
}

// TestUpdatePassword 测试更新密码
func (suite *UserServiceTestSuite) TestUpdatePassword() {
	// 先注册一个用户
	registerReq := &dto.UserRegisterRequest{
		Username: "passworduser",
		Password: "oldpassword",
		Email:    "password@example.com",
	}
	
	user, err := suite.userService.Register(suite.ctx, registerReq)
	assert.NoError(suite.T(), err)
	
	// 测试更新密码
	err = suite.userService.UpdatePassword(suite.ctx, user.ID, "oldpassword", "newpassword")
	assert.NoError(suite.T(), err)
	
	// 验证新密码可以登录
	loginReq := &dto.UserLoginRequest{
		Username: registerReq.Username,
		Password: "newpassword",
	}
	
	_, err = suite.userService.Login(suite.ctx, loginReq)
	assert.NoError(suite.T(), err)
	
	// 验证旧密码不能登录
	loginReq.Password = "oldpassword"
	_, err = suite.userService.Login(suite.ctx, loginReq)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), entity.ErrInvalidPassword, err)
	
	// 测试错误的旧密码
	err = suite.userService.UpdatePassword(suite.ctx, user.ID, "wrongoldpassword", "anotherpassword")
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), entity.ErrInvalidPassword, err)
}

// TestValidation 测试输入验证
func (suite *UserServiceTestSuite) TestValidation() {
	// 测试空用户名
	req := &dto.UserRegisterRequest{
		Username: "",
		Password: "password123",
		Email:    "test@example.com",
	}
	_, err := suite.userService.Register(suite.ctx, req)
	assert.Error(suite.T(), err)
	
	// 测试短用户名
	req.Username = "ab"
	_, err = suite.userService.Register(suite.ctx, req)
	assert.Error(suite.T(), err)
	
	// 测试短密码
	req.Username = "validuser"
	req.Password = "123"
	_, err = suite.userService.Register(suite.ctx, req)
	assert.Error(suite.T(), err)
	
	// 测试无效邮箱
	req.Password = "password123"
	req.Email = "invalid-email"
	_, err = suite.userService.Register(suite.ctx, req)
	assert.Error(suite.T(), err)
}

// TestUserServiceTestSuite 运行测试套件
func TestUserServiceTestSuite(t *testing.T) {
	suite.Run(t, new(UserServiceTestSuite))
}