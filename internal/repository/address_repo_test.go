package repository

import (
	"context"
	"fmt"
	"testing"

	"go-shop/internal/entity"

	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type AddressRepositoryTestSuite struct {
	suite.Suite
	db          *gorm.DB
	addressRepo AddressRepository
	userRepo    UserRepository
}

func (suite *AddressRepositoryTestSuite) SetupSuite() {
	// 设置SQLite数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	suite.Require().NoError(err)

	// 自动迁移
	err = db.AutoMigrate(&entity.User{}, &entity.Address{})
	suite.Require().NoError(err)

	suite.db = db
	suite.userRepo = NewUserRepository(db)
	suite.addressRepo = NewAddressRepository(db)
}

func (suite *AddressRepositoryTestSuite) TearDownTest() {
	// 清理测试数据
	suite.db.Exec("DELETE FROM addresses")
	suite.db.Exec("DELETE FROM users")
	suite.db.Exec("DELETE FROM sqlite_sequence WHERE name IN ('addresses', 'users')")
}

func (suite *AddressRepositoryTestSuite) createTestUser() *entity.User {
	ctx := context.Background()
	user := &entity.User{
		Username: "addresstest",
		Password: "hashedpassword",
		Email:    "addresstest@example.com",
		IsAdmin:  false,
	}
	err := suite.userRepo.Create(ctx, user)
	suite.Require().NoError(err)
	return user
}

func (suite *AddressRepositoryTestSuite) createTestAddress(userID int64, isDefault bool) *entity.Address {
	return &entity.Address{
		UserID:        userID,
		RecipientName: "张三",
		Phone:         "13800138000",
		Province:      "北京市",
		City:          "北京市",
		District:      "朝阳区",
		Detail:        "某某街道123号",
		IsDefault:     isDefault,
	}
}

func (suite *AddressRepositoryTestSuite) TestCreateAddress() {
	ctx := context.Background()
	user := suite.createTestUser()

	// 测试创建第一个地址（应该自动设为默认）
	address := suite.createTestAddress(user.ID, false)
	err := suite.addressRepo.Create(ctx, address)
	suite.NoError(err)
	suite.NotZero(address.ID)
	suite.True(address.IsDefault) // 第一个地址应该自动设为默认

	// 测试创建第二个地址（非默认）
	address2 := suite.createTestAddress(user.ID, false)
	address2.RecipientName = "李四"
	address2.Detail = "另一个街道456号"
	err = suite.addressRepo.Create(ctx, address2)
	suite.NoError(err)
	suite.False(address2.IsDefault) // 第二个地址不应该是默认

	// 测试创建第三个地址（设为默认）
	address3 := suite.createTestAddress(user.ID, true)
	address3.RecipientName = "王五"
	address3.Detail = "第三个街道789号"
	err = suite.addressRepo.Create(ctx, address3)
	suite.NoError(err)
	suite.True(address3.IsDefault)

	// 验证之前的默认地址已被取消
	updatedAddress, err := suite.addressRepo.GetByID(ctx, address.ID)
	suite.NoError(err)
	suite.False(updatedAddress.IsDefault)

	// 测试创建无效地址
	invalidAddress := &entity.Address{
		UserID: user.ID,
		// 缺少必要字段
	}
	err = suite.addressRepo.Create(ctx, invalidAddress)
	suite.Error(err)
	suite.Equal(entity.ErrRecipientNameRequired, err)

	// 测试无效用户ID
	invalidUserAddress := suite.createTestAddress(0, false)
	err = suite.addressRepo.Create(ctx, invalidUserAddress)
	suite.Error(err)
	suite.Equal(entity.ErrInvalidUserID, err)
}

func (suite *AddressRepositoryTestSuite) TestGetByID() {
	ctx := context.Background()
	user := suite.createTestUser()
	address := suite.createTestAddress(user.ID, true)

	// 创建地址
	err := suite.addressRepo.Create(ctx, address)
	suite.NoError(err)

	// 测试获取存在的地址
	foundAddress, err := suite.addressRepo.GetByID(ctx, address.ID)
	suite.NoError(err)
	suite.Equal(address.ID, foundAddress.ID)
	suite.Equal(address.RecipientName, foundAddress.RecipientName)
	suite.Equal(address.Phone, foundAddress.Phone)

	// 测试获取不存在的地址
	_, err = suite.addressRepo.GetByID(ctx, 99999)
	suite.Error(err)
	suite.Equal(entity.ErrAddressNotFound, err)

	// 测试无效ID
	_, err = suite.addressRepo.GetByID(ctx, 0)
	suite.Error(err)
	suite.Equal(entity.ErrInvalidAddressID, err)
}

func (suite *AddressRepositoryTestSuite) TestUpdateAddress() {
	ctx := context.Background()
	user := suite.createTestUser()
	address := suite.createTestAddress(user.ID, true)

	// 创建地址
	err := suite.addressRepo.Create(ctx, address)
	suite.NoError(err)

	// 测试更新地址信息
	address.RecipientName = "更新后的姓名"
	address.Phone = "13900139000"
	address.Detail = "更新后的详细地址"
	err = suite.addressRepo.Update(ctx, address)
	suite.NoError(err)

	// 验证更新结果
	updatedAddress, err := suite.addressRepo.GetByID(ctx, address.ID)
	suite.NoError(err)
	suite.Equal("更新后的姓名", updatedAddress.RecipientName)
	suite.Equal("13900139000", updatedAddress.Phone)
	suite.Equal("更新后的详细地址", updatedAddress.Detail)

	// 测试更新不存在的地址
	nonExistentAddress := suite.createTestAddress(user.ID, false)
	nonExistentAddress.ID = 99999
	err = suite.addressRepo.Update(ctx, nonExistentAddress)
	suite.Error(err)
	suite.Equal(entity.ErrAddressNotFound, err)

	// 测试更新无效地址
	invalidAddress := &entity.Address{
		ID:     address.ID,
		UserID: user.ID,
		// 缺少必要字段
	}
	err = suite.addressRepo.Update(ctx, invalidAddress)
	suite.Error(err)
	suite.Equal(entity.ErrRecipientNameRequired, err)
}

func (suite *AddressRepositoryTestSuite) TestDeleteAddress() {
	ctx := context.Background()
	user := suite.createTestUser()

	// 创建两个地址
	address1 := suite.createTestAddress(user.ID, true)
	err := suite.addressRepo.Create(ctx, address1)
	suite.NoError(err)

	address2 := suite.createTestAddress(user.ID, false)
	address2.RecipientName = "李四"
	err = suite.addressRepo.Create(ctx, address2)
	suite.NoError(err)

	// 测试删除非默认地址
	err = suite.addressRepo.Delete(ctx, address2.ID)
	suite.NoError(err)

	// 验证地址已删除
	_, err = suite.addressRepo.GetByID(ctx, address2.ID)
	suite.Error(err)
	suite.Equal(entity.ErrAddressNotFound, err)

	// 测试删除默认地址（应该设置另一个地址为默认）
	address3 := suite.createTestAddress(user.ID, false)
	address3.RecipientName = "王五"
	err = suite.addressRepo.Create(ctx, address3)
	suite.NoError(err)

	// 删除默认地址
	err = suite.addressRepo.Delete(ctx, address1.ID)
	suite.NoError(err)

	// 验证另一个地址变成了默认地址
	updatedAddress3, err := suite.addressRepo.GetByID(ctx, address3.ID)
	suite.NoError(err)
	suite.True(updatedAddress3.IsDefault)

	// 测试删除不存在的地址
	err = suite.addressRepo.Delete(ctx, 99999)
	suite.Error(err)
	suite.Equal(entity.ErrAddressNotFound, err)
}

func (suite *AddressRepositoryTestSuite) TestGetByUserID() {
	ctx := context.Background()
	user := suite.createTestUser()

	// 创建多个地址
	address1 := suite.createTestAddress(user.ID, true)
	err := suite.addressRepo.Create(ctx, address1)
	suite.NoError(err)

	address2 := suite.createTestAddress(user.ID, false)
	address2.RecipientName = "李四"
	err = suite.addressRepo.Create(ctx, address2)
	suite.NoError(err)

	address3 := suite.createTestAddress(user.ID, false)
	address3.RecipientName = "王五"
	err = suite.addressRepo.Create(ctx, address3)
	suite.NoError(err)

	// 测试获取用户所有地址
	addresses, err := suite.addressRepo.GetByUserID(ctx, user.ID)
	suite.NoError(err)
	suite.Len(addresses, 3)

	// 验证默认地址排在第一位
	suite.True(addresses[0].IsDefault)
	suite.Equal(address1.ID, addresses[0].ID)

	// 测试获取不存在用户的地址
	addresses, err = suite.addressRepo.GetByUserID(ctx, 99999)
	suite.NoError(err)
	suite.Empty(addresses)

	// 测试无效用户ID
	_, err = suite.addressRepo.GetByUserID(ctx, 0)
	suite.Error(err)
	suite.Equal(entity.ErrInvalidUserID, err)
}

func (suite *AddressRepositoryTestSuite) TestGetDefaultByUserID() {
	ctx := context.Background()
	user := suite.createTestUser()

	// 测试用户没有地址的情况
	_, err := suite.addressRepo.GetDefaultByUserID(ctx, user.ID)
	suite.Error(err)
	suite.Equal(entity.ErrAddressNotFound, err)

	// 创建默认地址
	address := suite.createTestAddress(user.ID, true)
	err = suite.addressRepo.Create(ctx, address)
	suite.NoError(err)

	// 测试获取默认地址
	defaultAddress, err := suite.addressRepo.GetDefaultByUserID(ctx, user.ID)
	suite.NoError(err)
	suite.Equal(address.ID, defaultAddress.ID)
	suite.True(defaultAddress.IsDefault)

	// 测试无效用户ID
	_, err = suite.addressRepo.GetDefaultByUserID(ctx, 0)
	suite.Error(err)
	suite.Equal(entity.ErrInvalidUserID, err)
}

func (suite *AddressRepositoryTestSuite) TestSetDefault() {
	ctx := context.Background()
	user := suite.createTestUser()

	// 创建两个地址
	address1 := suite.createTestAddress(user.ID, true)
	err := suite.addressRepo.Create(ctx, address1)
	suite.NoError(err)

	address2 := suite.createTestAddress(user.ID, false)
	address2.RecipientName = "李四"
	err = suite.addressRepo.Create(ctx, address2)
	suite.NoError(err)

	// 测试设置默认地址
	err = suite.addressRepo.SetDefault(ctx, user.ID, address2.ID)
	suite.NoError(err)

	// 验证新的默认地址
	defaultAddress, err := suite.addressRepo.GetDefaultByUserID(ctx, user.ID)
	suite.NoError(err)
	suite.Equal(address2.ID, defaultAddress.ID)

	// 验证之前的默认地址已取消
	updatedAddress1, err := suite.addressRepo.GetByID(ctx, address1.ID)
	suite.NoError(err)
	suite.False(updatedAddress1.IsDefault)

	// 测试设置已经是默认的地址
	err = suite.addressRepo.SetDefault(ctx, user.ID, address2.ID)
	suite.Error(err)
	suite.Equal(entity.ErrAddressAlreadyDefault, err)

	// 测试设置不存在的地址
	err = suite.addressRepo.SetDefault(ctx, user.ID, 99999)
	suite.Error(err)
	suite.Equal(entity.ErrAddressNotFound, err)
}

func (suite *AddressRepositoryTestSuite) TestUnsetDefault() {
	ctx := context.Background()
	user := suite.createTestUser()

	// 创建两个地址
	address1 := suite.createTestAddress(user.ID, true)
	err := suite.addressRepo.Create(ctx, address1)
	suite.NoError(err)

	address2 := suite.createTestAddress(user.ID, false)
	address2.RecipientName = "李四"
	err = suite.addressRepo.Create(ctx, address2)
	suite.NoError(err)

	// 测试取消默认地址
	err = suite.addressRepo.UnsetDefault(ctx, user.ID, address1.ID)
	suite.NoError(err)

	// 验证另一个地址变成了默认地址
	defaultAddress, err := suite.addressRepo.GetDefaultByUserID(ctx, user.ID)
	suite.NoError(err)
	suite.Equal(address2.ID, defaultAddress.ID)

	// 测试只有一个地址时取消默认（应该失败）
	err = suite.addressRepo.Delete(ctx, address1.ID)
	suite.NoError(err)

	err = suite.addressRepo.UnsetDefault(ctx, user.ID, address2.ID)
	suite.Error(err)
	suite.Equal(entity.ErrCannotDeleteDefault, err)
}

func (suite *AddressRepositoryTestSuite) TestCountByUserID() {
	ctx := context.Background()
	user := suite.createTestUser()

	// 测试用户没有地址的情况
	count, err := suite.addressRepo.CountByUserID(ctx, user.ID)
	suite.NoError(err)
	suite.Equal(int64(0), count)

	// 创建地址
	address1 := suite.createTestAddress(user.ID, true)
	err = suite.addressRepo.Create(ctx, address1)
	suite.NoError(err)

	address2 := suite.createTestAddress(user.ID, false)
	address2.RecipientName = "李四"
	err = suite.addressRepo.Create(ctx, address2)
	suite.NoError(err)

	// 测试获取地址数量
	count, err = suite.addressRepo.CountByUserID(ctx, user.ID)
	suite.NoError(err)
	suite.Equal(int64(2), count)

	// 测试无效用户ID
	_, err = suite.addressRepo.CountByUserID(ctx, 0)
	suite.Error(err)
	suite.Equal(entity.ErrInvalidUserID, err)
}

func (suite *AddressRepositoryTestSuite) TestExistsByID() {
	ctx := context.Background()
	user := suite.createTestUser()
	address := suite.createTestAddress(user.ID, true)

	// 测试地址不存在的情况
	exists, err := suite.addressRepo.ExistsByID(ctx, 99999)
	suite.NoError(err)
	suite.False(exists)

	// 创建地址
	err = suite.addressRepo.Create(ctx, address)
	suite.NoError(err)

	// 测试地址存在的情况
	exists, err = suite.addressRepo.ExistsByID(ctx, address.ID)
	suite.NoError(err)
	suite.True(exists)

	// 测试无效ID
	_, err = suite.addressRepo.ExistsByID(ctx, 0)
	suite.Error(err)
	suite.Equal(entity.ErrInvalidAddressID, err)
}

func (suite *AddressRepositoryTestSuite) TestExistsByUserIDAndID() {
	ctx := context.Background()
	user1 := suite.createTestUser()
	user2 := &entity.User{
		Username: "addresstest2",
		Password: "hashedpassword",
		Email:    "addresstest2@example.com",
		IsAdmin:  false,
	}
	err := suite.userRepo.Create(ctx, user2)
	suite.Require().NoError(err)

	address := suite.createTestAddress(user1.ID, true)
	err = suite.addressRepo.Create(ctx, address)
	suite.NoError(err)

	// 测试正确的用户和地址
	exists, err := suite.addressRepo.ExistsByUserIDAndID(ctx, user1.ID, address.ID)
	suite.NoError(err)
	suite.True(exists)

	// 测试错误的用户
	exists, err = suite.addressRepo.ExistsByUserIDAndID(ctx, user2.ID, address.ID)
	suite.NoError(err)
	suite.False(exists)

	// 测试不存在的地址
	exists, err = suite.addressRepo.ExistsByUserIDAndID(ctx, user1.ID, 99999)
	suite.NoError(err)
	suite.False(exists)
}

func (suite *AddressRepositoryTestSuite) TestGetUserAddressWithPagination() {
	ctx := context.Background()
	user := suite.createTestUser()

	// 创建多个地址
	for i := 0; i < 5; i++ {
		address := suite.createTestAddress(user.ID, i == 0)
		address.RecipientName = fmt.Sprintf("用户%d", i+1)
		err := suite.addressRepo.Create(ctx, address)
		suite.NoError(err)
	}

	// 测试分页获取
	addresses, total, err := suite.addressRepo.GetUserAddressWithPagination(ctx, user.ID, 0, 3)
	suite.NoError(err)
	suite.Equal(int64(5), total)
	suite.Len(addresses, 3)

	// 验证默认地址排在第一位
	suite.True(addresses[0].IsDefault)

	// 测试第二页
	addresses, total, err = suite.addressRepo.GetUserAddressWithPagination(ctx, user.ID, 3, 3)
	suite.NoError(err)
	suite.Equal(int64(5), total)
	suite.Len(addresses, 2)
}

func (suite *AddressRepositoryTestSuite) TestDeleteByUserID() {
	ctx := context.Background()
	user := suite.createTestUser()

	// 创建多个地址
	address1 := suite.createTestAddress(user.ID, true)
	err := suite.addressRepo.Create(ctx, address1)
	suite.NoError(err)

	address2 := suite.createTestAddress(user.ID, false)
	address2.RecipientName = "李四"
	err = suite.addressRepo.Create(ctx, address2)
	suite.NoError(err)

	// 测试删除用户所有地址
	err = suite.addressRepo.DeleteByUserID(ctx, user.ID)
	suite.NoError(err)

	// 验证地址已删除
	addresses, err := suite.addressRepo.GetByUserID(ctx, user.ID)
	suite.NoError(err)
	suite.Empty(addresses)
}

func (suite *AddressRepositoryTestSuite) TestHasDefaultAddress() {
	ctx := context.Background()
	user := suite.createTestUser()

	// 测试用户没有地址的情况
	hasDefault, err := suite.addressRepo.HasDefaultAddress(ctx, user.ID)
	suite.NoError(err)
	suite.False(hasDefault)

	// 创建默认地址
	address := suite.createTestAddress(user.ID, true)
	err = suite.addressRepo.Create(ctx, address)
	suite.NoError(err)

	// 测试用户有默认地址的情况
	hasDefault, err = suite.addressRepo.HasDefaultAddress(ctx, user.ID)
	suite.NoError(err)
	suite.True(hasDefault)
}

func (suite *AddressRepositoryTestSuite) TestGetAddressesInRegion() {
	ctx := context.Background()
	user := suite.createTestUser()

	// 创建不同地区的地址
	address1 := suite.createTestAddress(user.ID, true)
	address1.Province = "北京市"
	address1.City = "北京市"
	err := suite.addressRepo.Create(ctx, address1)
	suite.NoError(err)

	address2 := suite.createTestAddress(user.ID, false)
	address2.Province = "上海市"
	address2.City = "上海市"
	address2.RecipientName = "李四"
	err = suite.addressRepo.Create(ctx, address2)
	suite.NoError(err)

	address3 := suite.createTestAddress(user.ID, false)
	address3.Province = "北京市"
	address3.City = "北京市"
	address3.RecipientName = "王五"
	err = suite.addressRepo.Create(ctx, address3)
	suite.NoError(err)

	// 测试按省份查询
	addresses, err := suite.addressRepo.GetAddressesInRegion(ctx, "北京市", "", 10)
	suite.NoError(err)
	suite.Len(addresses, 2)

	// 测试按省份和城市查询
	addresses, err = suite.addressRepo.GetAddressesInRegion(ctx, "上海市", "上海市", 10)
	suite.NoError(err)
	suite.Len(addresses, 1)
	suite.Equal("李四", addresses[0].RecipientName)

	// 测试查询不存在的地区
	addresses, err = suite.addressRepo.GetAddressesInRegion(ctx, "广东省", "", 10)
	suite.NoError(err)
	suite.Empty(addresses)
}

func (suite *AddressRepositoryTestSuite) TestGetAddressStatistics() {
	ctx := context.Background()
	user1 := suite.createTestUser()
	user2 := &entity.User{
		Username: "addresstest2",
		Password: "hashedpassword",
		Email:    "addresstest2@example.com",
		IsAdmin:  false,
	}
	err := suite.userRepo.Create(ctx, user2)
	suite.Require().NoError(err)

	// 为用户1创建地址
	address1 := suite.createTestAddress(user1.ID, true)
	address1.Province = "北京市"
	address1.City = "北京市"
	err = suite.addressRepo.Create(ctx, address1)
	suite.NoError(err)

	address2 := suite.createTestAddress(user1.ID, false)
	address2.Province = "上海市"
	address2.City = "上海市"
	address2.RecipientName = "李四"
	err = suite.addressRepo.Create(ctx, address2)
	suite.NoError(err)

	// 为用户2创建地址
	address3 := suite.createTestAddress(user2.ID, true)
	address3.Province = "北京市"
	address3.City = "北京市"
	address3.RecipientName = "王五"
	err = suite.addressRepo.Create(ctx, address3)
	suite.NoError(err)

	// 测试获取统计信息
	stats, err := suite.addressRepo.GetAddressStatistics(ctx)
	suite.NoError(err)
	suite.Equal(int64(3), stats.TotalAddresses)
	suite.Equal(int64(2), stats.UsersWithAddresses)
	suite.Equal(1.5, stats.AverageAddressPerUser)
	suite.NotEmpty(stats.TopProvinces)
	suite.NotEmpty(stats.TopCities)

	// 验证热门省份
	suite.Equal("北京市", stats.TopProvinces[0].Province)
	suite.Equal(int64(2), stats.TopProvinces[0].Count)
}

func TestAddressRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(AddressRepositoryTestSuite))
}