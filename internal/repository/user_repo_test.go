package repository

import (
	"context"
	"testing"

	"go-shop/internal/entity"

	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type UserRepositoryTestSuite struct {
	suite.Suite
	db   *gorm.DB
	repo UserRepository
}

func (suite *UserRepositoryTestSuite) SetupSuite() {
	// 使用内存SQLite数据库进行测试
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	suite.Require().NoError(err)

	// 自动迁移
	err = db.AutoMigrate(&entity.User{})
	suite.Require().NoError(err)

	suite.db = db
	suite.repo = NewUserRepository(db)
}

func (suite *UserRepositoryTestSuite) TearDownTest() {
	// 清理测试数据
	suite.db.Exec("DELETE FROM users")
}

func (suite *UserRepositoryTestSuite) TestCreate() {
	ctx := context.Background()

	// 测试创建用户
	user := &entity.User{
		Username: "testuser",
		Password: "hashedpassword",
		Email:    "test@example.com",
		IsAdmin:  false,
	}

	err := suite.repo.Create(ctx, user)
	suite.NoError(err)
	suite.NotZero(user.ID)

	// 测试用户名重复
	duplicateUser := &entity.User{
		Username: "testuser",
		Password: "hashedpassword2",
		Email:    "test2@example.com",
	}

	err = suite.repo.Create(ctx, duplicateUser)
	suite.Error(err)
	suite.Equal(entity.ErrUserAlreadyExists, err)

	// 测试邮箱重复
	duplicateEmailUser := &entity.User{
		Username: "testuser2",
		Password: "hashedpassword3",
		Email:    "test@example.com",
	}

	err = suite.repo.Create(ctx, duplicateEmailUser)
	suite.Error(err)
	suite.Equal(entity.ErrEmailAlreadyExists, err)

	// 测试nil用户
	err = suite.repo.Create(ctx, nil)
	suite.Error(err)
}

func (suite *UserRepositoryTestSuite) TestGetByID() {
	ctx := context.Background()

	// 创建测试用户
	user := &entity.User{
		Username: "testuser",
		Password: "hashedpassword",
		Email:    "test@example.com",
	}
	err := suite.repo.Create(ctx, user)
	suite.NoError(err)

	// 测试获取存在的用户
	foundUser, err := suite.repo.GetByID(ctx, user.ID)
	suite.NoError(err)
	suite.Equal(user.Username, foundUser.Username)
	suite.Equal(user.Email, foundUser.Email)

	// 测试获取不存在的用户
	_, err = suite.repo.GetByID(ctx, 99999)
	suite.Error(err)
	suite.Equal(entity.ErrUserNotFound, err)

	// 测试无效ID
	_, err = suite.repo.GetByID(ctx, 0)
	suite.Error(err)
	suite.Equal(entity.ErrInvalidUserID, err)
}

func (suite *UserRepositoryTestSuite) TestGetByUsername() {
	ctx := context.Background()

	// 创建测试用户
	user := &entity.User{
		Username: "testuser",
		Password: "hashedpassword",
		Email:    "test@example.com",
	}
	err := suite.repo.Create(ctx, user)
	suite.NoError(err)

	// 测试获取存在的用户
	foundUser, err := suite.repo.GetByUsername(ctx, "testuser")
	suite.NoError(err)
	suite.Equal(user.ID, foundUser.ID)
	suite.Equal(user.Email, foundUser.Email)

	// 测试获取不存在的用户
	_, err = suite.repo.GetByUsername(ctx, "nonexistent")
	suite.Error(err)
	suite.Equal(entity.ErrUserNotFound, err)

	// 测试空用户名
	_, err = suite.repo.GetByUsername(ctx, "")
	suite.Error(err)
	suite.Equal(entity.ErrInvalidUsername, err)
}

func (suite *UserRepositoryTestSuite) TestGetByEmail() {
	ctx := context.Background()

	// 创建测试用户
	user := &entity.User{
		Username: "testuser",
		Password: "hashedpassword",
		Email:    "test@example.com",
	}
	err := suite.repo.Create(ctx, user)
	suite.NoError(err)

	// 测试获取存在的用户
	foundUser, err := suite.repo.GetByEmail(ctx, "test@example.com")
	suite.NoError(err)
	suite.Equal(user.ID, foundUser.ID)
	suite.Equal(user.Username, foundUser.Username)

	// 测试获取不存在的用户
	_, err = suite.repo.GetByEmail(ctx, "nonexistent@example.com")
	suite.Error(err)
	suite.Equal(entity.ErrUserNotFound, err)

	// 测试空邮箱
	_, err = suite.repo.GetByEmail(ctx, "")
	suite.Error(err)
	suite.Equal(entity.ErrInvalidEmail, err)
}

func (suite *UserRepositoryTestSuite) TestUpdate() {
	ctx := context.Background()

	// 创建测试用户
	user := &entity.User{
		Username: "testuser",
		Password: "hashedpassword",
		Email:    "test@example.com",
	}
	err := suite.repo.Create(ctx, user)
	suite.NoError(err)

	// 测试更新用户信息
	user.Email = "updated@example.com"
	err = suite.repo.Update(ctx, user)
	suite.NoError(err)

	// 验证更新
	updatedUser, err := suite.repo.GetByID(ctx, user.ID)
	suite.NoError(err)
	suite.Equal("updated@example.com", updatedUser.Email)

	// 测试更新不存在的用户
	nonExistentUser := &entity.User{
		ID:       99999,
		Username: "nonexistent",
		Email:    "nonexistent@example.com",
	}
	err = suite.repo.Update(ctx, nonExistentUser)
	suite.Error(err)
	suite.Equal(entity.ErrUserNotFound, err)
}

func (suite *UserRepositoryTestSuite) TestUpdatePassword() {
	ctx := context.Background()

	// 创建测试用户
	user := &entity.User{
		Username: "testuser",
		Password: "hashedpassword",
		Email:    "test@example.com",
	}
	err := suite.repo.Create(ctx, user)
	suite.NoError(err)

	// 测试更新密码
	newHashedPassword := "newhashedpassword"
	err = suite.repo.UpdatePassword(ctx, user.ID, newHashedPassword)
	suite.NoError(err)

	// 验证密码更新
	updatedUser, err := suite.repo.GetByID(ctx, user.ID)
	suite.NoError(err)
	suite.Equal(newHashedPassword, updatedUser.Password)

	// 测试更新不存在用户的密码
	err = suite.repo.UpdatePassword(ctx, 99999, "somepassword")
	suite.Error(err)
	suite.Equal(entity.ErrUserNotFound, err)

	// 测试无效ID
	err = suite.repo.UpdatePassword(ctx, 0, "somepassword")
	suite.Error(err)
	suite.Equal(entity.ErrInvalidUserID, err)
}

func (suite *UserRepositoryTestSuite) TestExistsByUsername() {
	ctx := context.Background()

	// 创建测试用户
	user := &entity.User{
		Username: "testuser",
		Password: "hashedpassword",
		Email:    "test@example.com",
	}
	err := suite.repo.Create(ctx, user)
	suite.NoError(err)

	// 测试存在的用户名
	exists, err := suite.repo.ExistsByUsername(ctx, "testuser")
	suite.NoError(err)
	suite.True(exists)

	// 测试不存在的用户名
	exists, err = suite.repo.ExistsByUsername(ctx, "nonexistent")
	suite.NoError(err)
	suite.False(exists)

	// 测试空用户名
	exists, err = suite.repo.ExistsByUsername(ctx, "")
	suite.NoError(err)
	suite.False(exists)
}

func (suite *UserRepositoryTestSuite) TestExistsByEmail() {
	ctx := context.Background()

	// 创建测试用户
	user := &entity.User{
		Username: "testuser",
		Password: "hashedpassword",
		Email:    "test@example.com",
	}
	err := suite.repo.Create(ctx, user)
	suite.NoError(err)

	// 测试存在的邮箱
	exists, err := suite.repo.ExistsByEmail(ctx, "test@example.com")
	suite.NoError(err)
	suite.True(exists)

	// 测试不存在的邮箱
	exists, err = suite.repo.ExistsByEmail(ctx, "nonexistent@example.com")
	suite.NoError(err)
	suite.False(exists)

	// 测试空邮箱
	exists, err = suite.repo.ExistsByEmail(ctx, "")
	suite.NoError(err)
	suite.False(exists)
}

func (suite *UserRepositoryTestSuite) TestList() {
	ctx := context.Background()

	// 创建多个测试用户
	users := []*entity.User{
		{Username: "user1", Password: "pass1", Email: "user1@example.com"},
		{Username: "user2", Password: "pass2", Email: "user2@example.com"},
		{Username: "user3", Password: "pass3", Email: "user3@example.com"},
	}

	for _, user := range users {
		err := suite.repo.Create(ctx, user)
		suite.NoError(err)
	}

	// 测试分页查询
	userList, total, err := suite.repo.List(ctx, 0, 2)
	suite.NoError(err)
	suite.Equal(int64(3), total)
	suite.Len(userList, 2)

	// 测试第二页
	userList, total, err = suite.repo.List(ctx, 2, 2)
	suite.NoError(err)
	suite.Equal(int64(3), total)
	suite.Len(userList, 1)
}

func (suite *UserRepositoryTestSuite) TestCountUsers() {
	ctx := context.Background()

	// 初始计数应该为0
	count, err := suite.repo.CountUsers(ctx)
	suite.NoError(err)
	suite.Equal(int64(0), count)

	// 创建用户
	user := &entity.User{
		Username: "testuser",
		Password: "hashedpassword",
		Email:    "test@example.com",
	}
	err = suite.repo.Create(ctx, user)
	suite.NoError(err)

	// 计数应该为1
	count, err = suite.repo.CountUsers(ctx)
	suite.NoError(err)
	suite.Equal(int64(1), count)
}

func TestUserRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(UserRepositoryTestSuite))
}