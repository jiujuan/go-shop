package service

import (
	"context"
	"testing"

	"go-shop/internal/dto"
	"go-shop/internal/entity"
	"go-shop/internal/repository"

	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type AddressServiceTestSuite struct {
	suite.Suite
	db             *gorm.DB
	addressRepo    repository.AddressRepository
	userRepo       repository.UserRepository
	addressService AddressService
}

func (suite *AddressServiceTestSuite) SetupSuite() {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	suite.Require().NoError(err)

	err = db.AutoMigrate(&entity.User{}, &entity.Address{})
	suite.Require().NoError(err)

	suite.db = db
	suite.userRepo = repository.NewUserRepository(db)
	suite.addressRepo = repository.NewAddressRepository(db)
	suite.addressService = NewAddressService(suite.addressRepo)
}

func (suite *AddressServiceTestSuite) TearDownTest() {
	suite.db.Exec("DELETE FROM addresses")
	suite.db.Exec("DELETE FROM users")
	suite.db.Exec("DELETE FROM sqlite_sequence WHERE name IN ('addresses', 'users')")
}

func (suite *AddressServiceTestSuite) createTestUser() *entity.User {
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

func (suite *AddressServiceTestSuite) TestCreateAddress() {
	ctx := context.Background()
	user := suite.createTestUser()

	// 测试创建第一个地址
	isDefaultFalse := false
	req := &dto.AddressCreateRequest{
		RecipientName: "张三",
		Phone:         "13800138000",
		Province:      "北京市",
		City:          "北京市",
		District:      "朝阳区",
		Detail:        "某某街道123号",
		IsDefault:     &isDefaultFalse,
	}

	resp, err := suite.addressService.CreateAddress(ctx, user.ID, req)
	suite.NoError(err)
	suite.NotNil(resp)
	suite.Equal(req.RecipientName, resp.RecipientName)
	suite.Equal(req.Phone, resp.Phone)
	suite.True(resp.IsDefault) // 第一个地址应该自动设为默认

	// 测试创建第二个地址
	req2 := &dto.AddressCreateRequest{
		RecipientName: "李四",
		Phone:         "13900139000",
		Province:      "上海市",
		City:          "上海市",
		District:      "浦东新区",
		Detail:        "另一个街道456号",
		IsDefault:     &isDefaultFalse,
	}

	resp2, err := suite.addressService.CreateAddress(ctx, user.ID, req2)
	suite.NoError(err)
	suite.NotNil(resp2)
	suite.False(resp2.IsDefault) // 第二个地址不应该是默认

	// 测试无效用户ID
	_, err = suite.addressService.CreateAddress(ctx, 0, req)
	suite.Error(err)
	suite.Equal(entity.ErrInvalidUserID, err)
}

func (suite *AddressServiceTestSuite) TestGetAddress() {
	ctx := context.Background()
	user := suite.createTestUser()

	// 创建地址
	isDefaultTrue := true
	req := &dto.AddressCreateRequest{
		RecipientName: "张三",
		Phone:         "13800138000",
		Province:      "北京市",
		City:          "北京市",
		District:      "朝阳区",
		Detail:        "某某街道123号",
		IsDefault:     &isDefaultTrue,
	}

	created, err := suite.addressService.CreateAddress(ctx, user.ID, req)
	suite.NoError(err)

	// 测试获取地址
	resp, err := suite.addressService.GetAddress(ctx, user.ID, created.ID)
	suite.NoError(err)
	suite.NotNil(resp)
	suite.Equal(created.ID, resp.ID)
	suite.Equal(req.RecipientName, resp.RecipientName)

	// 测试获取不存在的地址
	_, err = suite.addressService.GetAddress(ctx, user.ID, 99999)
	suite.Error(err)
	suite.Equal(entity.ErrAddressNotFound, err)

	// 测试无效用户ID
	_, err = suite.addressService.GetAddress(ctx, 0, created.ID)
	suite.Error(err)
	suite.Equal(entity.ErrInvalidUserID, err)

	// 测试无效地址ID
	_, err = suite.addressService.GetAddress(ctx, user.ID, 0)
	suite.Error(err)
	suite.Equal(entity.ErrInvalidAddressID, err)
}

func (suite *AddressServiceTestSuite) TestUpdateAddress() {
	ctx := context.Background()
	user := suite.createTestUser()

	// 创建地址
	isDefaultTrue := true
	req := &dto.AddressCreateRequest{
		RecipientName: "张三",
		Phone:         "13800138000",
		Province:      "北京市",
		City:          "北京市",
		District:      "朝阳区",
		Detail:        "某某街道123号",
		IsDefault:     &isDefaultTrue,
	}

	created, err := suite.addressService.CreateAddress(ctx, user.ID, req)
	suite.NoError(err)

	// 测试更新地址
	newName := "更新后的张三"
	newPhone := "13700137000"
	updateReq := &dto.AddressUpdateRequest{
		RecipientName: &newName,
		Phone:         &newPhone,
	}

	resp, err := suite.addressService.UpdateAddress(ctx, user.ID, created.ID, updateReq)
	suite.NoError(err)
	suite.NotNil(resp)
	suite.Equal(newName, resp.RecipientName)
	suite.Equal(newPhone, resp.Phone)
	suite.Equal(req.Province, resp.Province) // 未更新的字段应保持不变

	// 测试更新不存在的地址
	_, err = suite.addressService.UpdateAddress(ctx, user.ID, 99999, updateReq)
	suite.Error(err)
	suite.Equal(entity.ErrAddressNotFound, err)

	// 测试无效用户ID
	_, err = suite.addressService.UpdateAddress(ctx, 0, created.ID, updateReq)
	suite.Error(err)
	suite.Equal(entity.ErrInvalidUserID, err)
}

func (suite *AddressServiceTestSuite) TestDeleteAddress() {
	ctx := context.Background()
	user := suite.createTestUser()

	// 创建两个地址
	isDefaultTrue := true
	isDefaultFalse := false
	req1 := &dto.AddressCreateRequest{
		RecipientName: "张三",
		Phone:         "13800138000",
		Province:      "北京市",
		City:          "北京市",
		District:      "朝阳区",
		Detail:        "某某街道123号",
		IsDefault:     &isDefaultTrue,
	}

	created1, err := suite.addressService.CreateAddress(ctx, user.ID, req1)
	suite.NoError(err)

	req2 := &dto.AddressCreateRequest{
		RecipientName: "李四",
		Phone:         "13900139000",
		Province:      "上海市",
		City:          "上海市",
		District:      "浦东新区",
		Detail:        "另一个街道456号",
		IsDefault:     &isDefaultFalse,
	}

	created2, err := suite.addressService.CreateAddress(ctx, user.ID, req2)
	suite.NoError(err)

	// 测试删除非默认地址
	err = suite.addressService.DeleteAddress(ctx, user.ID, created2.ID)
	suite.NoError(err)

	// 验证地址已删除
	_, err = suite.addressService.GetAddress(ctx, user.ID, created2.ID)
	suite.Error(err)
	suite.Equal(entity.ErrAddressNotFound, err)

	// 测试删除不存在的地址
	err = suite.addressService.DeleteAddress(ctx, user.ID, 99999)
	suite.Error(err)
	suite.Equal(entity.ErrAddressNotFound, err)

	// 测试无效用户ID
	err = suite.addressService.DeleteAddress(ctx, 0, created1.ID)
	suite.Error(err)
	suite.Equal(entity.ErrInvalidUserID, err)
}

func (suite *AddressServiceTestSuite) TestGetUserAddresses() {
	ctx := context.Background()
	user := suite.createTestUser()

	// 测试空地址列表
	addresses, err := suite.addressService.GetUserAddresses(ctx, user.ID)
	suite.NoError(err)
	suite.Empty(addresses)

	// 创建多个地址
	isDefaultTrue := true
	isDefaultFalse := false
	req1 := &dto.AddressCreateRequest{
		RecipientName: "张三",
		Phone:         "13800138000",
		Province:      "北京市",
		City:          "北京市",
		District:      "朝阳区",
		Detail:        "某某街道123号",
		IsDefault:     &isDefaultTrue,
	}

	_, err = suite.addressService.CreateAddress(ctx, user.ID, req1)
	suite.NoError(err)

	req2 := &dto.AddressCreateRequest{
		RecipientName: "李四",
		Phone:         "13900139000",
		Province:      "上海市",
		City:          "上海市",
		District:      "浦东新区",
		Detail:        "另一个街道456号",
		IsDefault:     &isDefaultFalse,
	}

	_, err = suite.addressService.CreateAddress(ctx, user.ID, req2)
	suite.NoError(err)

	// 测试获取地址列表
	addresses, err = suite.addressService.GetUserAddresses(ctx, user.ID)
	suite.NoError(err)
	suite.Len(addresses, 2)

	// 验证默认地址排在第一位
	suite.True(addresses[0].IsDefault)

	// 测试无效用户ID
	_, err = suite.addressService.GetUserAddresses(ctx, 0)
	suite.Error(err)
	suite.Equal(entity.ErrInvalidUserID, err)
}

func (suite *AddressServiceTestSuite) TestGetDefaultAddress() {
	ctx := context.Background()
	user := suite.createTestUser()

	// 测试用户没有地址的情况
	_, err := suite.addressService.GetDefaultAddress(ctx, user.ID)
	suite.Error(err)
	suite.Equal(entity.ErrAddressNotFound, err)

	// 创建默认地址
	isDefaultTrue := true
	req := &dto.AddressCreateRequest{
		RecipientName: "张三",
		Phone:         "13800138000",
		Province:      "北京市",
		City:          "北京市",
		District:      "朝阳区",
		Detail:        "某某街道123号",
		IsDefault:     &isDefaultTrue,
	}

	created, err := suite.addressService.CreateAddress(ctx, user.ID, req)
	suite.NoError(err)

	// 测试获取默认地址
	resp, err := suite.addressService.GetDefaultAddress(ctx, user.ID)
	suite.NoError(err)
	suite.NotNil(resp)
	suite.Equal(created.ID, resp.ID)
	suite.True(resp.IsDefault)

	// 测试无效用户ID
	_, err = suite.addressService.GetDefaultAddress(ctx, 0)
	suite.Error(err)
	suite.Equal(entity.ErrInvalidUserID, err)
}

func (suite *AddressServiceTestSuite) TestSetDefaultAddress() {
	ctx := context.Background()
	user := suite.createTestUser()

	// 创建两个地址
	isDefaultTrue := true
	isDefaultFalse := false
	req1 := &dto.AddressCreateRequest{
		RecipientName: "张三",
		Phone:         "13800138000",
		Province:      "北京市",
		City:          "北京市",
		District:      "朝阳区",
		Detail:        "某某街道123号",
		IsDefault:     &isDefaultTrue,
	}

	created1, err := suite.addressService.CreateAddress(ctx, user.ID, req1)
	suite.NoError(err)

	req2 := &dto.AddressCreateRequest{
		RecipientName: "李四",
		Phone:         "13900139000",
		Province:      "上海市",
		City:          "上海市",
		District:      "浦东新区",
		Detail:        "另一个街道456号",
		IsDefault:     &isDefaultFalse,
	}

	created2, err := suite.addressService.CreateAddress(ctx, user.ID, req2)
	suite.NoError(err)

	// 测试设置默认地址
	err = suite.addressService.SetDefaultAddress(ctx, user.ID, created2.ID)
	suite.NoError(err)

	// 验证新的默认地址
	defaultAddr, err := suite.addressService.GetDefaultAddress(ctx, user.ID)
	suite.NoError(err)
	suite.Equal(created2.ID, defaultAddr.ID)

	// 验证之前的默认地址已取消
	addr1, err := suite.addressService.GetAddress(ctx, user.ID, created1.ID)
	suite.NoError(err)
	suite.False(addr1.IsDefault)

	// 测试设置不存在的地址
	err = suite.addressService.SetDefaultAddress(ctx, user.ID, 99999)
	suite.Error(err)
	suite.Equal(entity.ErrAddressNotFound, err)

	// 测试无效用户ID
	err = suite.addressService.SetDefaultAddress(ctx, 0, created2.ID)
	suite.Error(err)
	suite.Equal(entity.ErrInvalidUserID, err)
}

func (suite *AddressServiceTestSuite) TestGetUserAddressesWithPagination() {
	ctx := context.Background()
	user := suite.createTestUser()

	// 创建多个地址
	for i := 0; i < 5; i++ {
		isDefault := i == 0
		req := &dto.AddressCreateRequest{
			RecipientName: "测试用户",
			Phone:         "13800138000",
			Province:      "北京市",
			City:          "北京市",
			District:      "朝阳区",
			Detail:        "某某街道123号",
			IsDefault:     &isDefault,
		}

		_, err := suite.addressService.CreateAddress(ctx, user.ID, req)
		suite.NoError(err)
	}

	// 测试分页获取
	resp, err := suite.addressService.GetUserAddressesWithPagination(ctx, user.ID, 1, 3)
	suite.NoError(err)
	suite.NotNil(resp)
	suite.Equal(1, resp.Pagination.Page)
	suite.Equal(3, resp.Pagination.PageSize)
	suite.Equal(int64(5), resp.Pagination.Total)
	suite.Equal(2, resp.Pagination.TotalPages)
	suite.Len(resp.Addresses, 3)

	// 测试第二页
	resp, err = suite.addressService.GetUserAddressesWithPagination(ctx, user.ID, 2, 3)
	suite.NoError(err)
	suite.Len(resp.Addresses, 2)

	// 测试无效用户ID
	_, err = suite.addressService.GetUserAddressesWithPagination(ctx, 0, 1, 10)
	suite.Error(err)
	suite.Equal(entity.ErrInvalidUserID, err)
}

func (suite *AddressServiceTestSuite) TestUserPermissionCheck() {
	ctx := context.Background()
	user1 := suite.createTestUser()

	// 创建第二个用户
	user2 := &entity.User{
		Username: "addresstest2",
		Password: "hashedpassword",
		Email:    "addresstest2@example.com",
		IsAdmin:  false,
	}
	err := suite.userRepo.Create(ctx, user2)
	suite.Require().NoError(err)

	// 用户1创建地址
	isDefaultTrue := true
	req := &dto.AddressCreateRequest{
		RecipientName: "张三",
		Phone:         "13800138000",
		Province:      "北京市",
		City:          "北京市",
		District:      "朝阳区",
		Detail:        "某某街道123号",
		IsDefault:     &isDefaultTrue,
	}

	created, err := suite.addressService.CreateAddress(ctx, user1.ID, req)
	suite.NoError(err)

	// 用户2尝试访问用户1的地址（应该失败）
	_, err = suite.addressService.GetAddress(ctx, user2.ID, created.ID)
	suite.Error(err)
	suite.Equal(entity.ErrAddressNotFound, err)

	// 用户2尝试更新用户1的地址（应该失败）
	newName := "恶意更新"
	updateReq := &dto.AddressUpdateRequest{
		RecipientName: &newName,
	}
	_, err = suite.addressService.UpdateAddress(ctx, user2.ID, created.ID, updateReq)
	suite.Error(err)
	suite.Equal(entity.ErrAddressNotFound, err)

	// 用户2尝试删除用户1的地址（应该失败）
	err = suite.addressService.DeleteAddress(ctx, user2.ID, created.ID)
	suite.Error(err)
	suite.Equal(entity.ErrAddressNotFound, err)

	// 用户2尝试设置用户1的地址为默认（应该失败）
	err = suite.addressService.SetDefaultAddress(ctx, user2.ID, created.ID)
	suite.Error(err)
	suite.Equal(entity.ErrAddressNotFound, err)
}

func TestAddressServiceTestSuite(t *testing.T) {
	suite.Run(t, new(AddressServiceTestSuite))
}
