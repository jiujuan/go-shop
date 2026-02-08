package utils

import (
	"context"
	"log"

	"go-shop/internal/entity"
	"go-shop/internal/repository"

	"gorm.io/gorm"
)

// TestAddressRepository 测试地址仓库功能
func TestAddressRepository(db *gorm.DB) {
	log.Println("Testing Address Repository functionality...")

	// 创建仓库实例
	userRepo := repository.NewUserRepository(db)
	addressRepo := repository.NewAddressRepository(db)
	ctx := context.Background()

	// 创建测试用户
	testUser := &entity.User{
		Username: "addresstest",
		Password: "hashedpassword",
		Email:    "addresstest@example.com",
		IsAdmin:  false,
	}

	err := userRepo.Create(ctx, testUser)
	if err != nil {
		log.Printf("✗ Failed to create test user: %v", err)
		return
	}
	log.Printf("✓ Test user created with ID: %d", testUser.ID)

	// 测试创建地址
	testAddress := &entity.Address{
		UserID:        testUser.ID,
		RecipientName: "张三",
		Phone:         "13800138000",
		Province:      "北京市",
		City:          "北京市",
		District:      "朝阳区",
		Detail:        "某某街道123号",
		IsDefault:     false,
	}

	err = addressRepo.Create(ctx, testAddress)
	if err != nil {
		log.Printf("✗ Address creation failed: %v", err)
		return
	}
	log.Printf("✓ Address created successfully with ID: %d", testAddress.ID)

	// 验证第一个地址自动设为默认
	if !testAddress.IsDefault {
		log.Printf("✗ First address should be set as default")
		return
	}
	log.Printf("✓ First address automatically set as default")

	// 测试根据ID获取地址
	foundAddress, err := addressRepo.GetByID(ctx, testAddress.ID)
	if err != nil {
		log.Printf("✗ Get address by ID failed: %v", err)
		return
	}
	log.Printf("✓ Address found by ID: %s (%s)", foundAddress.RecipientName, foundAddress.Phone)

	// 测试获取用户默认地址
	defaultAddress, err := addressRepo.GetDefaultByUserID(ctx, testUser.ID)
	if err != nil {
		log.Printf("✗ Get default address failed: %v", err)
		return
	}
	log.Printf("✓ Default address retrieved: %s", defaultAddress.RecipientName)

	// 测试创建第二个地址
	secondAddress := &entity.Address{
		UserID:        testUser.ID,
		RecipientName: "李四",
		Phone:         "13900139000",
		Province:      "上海市",
		City:          "上海市",
		District:      "浦东新区",
		Detail:        "另一个街道456号",
		IsDefault:     false,
	}

	err = addressRepo.Create(ctx, secondAddress)
	if err != nil {
		log.Printf("✗ Second address creation failed: %v", err)
		return
	}
	log.Printf("✓ Second address created with ID: %d", secondAddress.ID)

	// 测试获取用户所有地址
	addresses, err := addressRepo.GetByUserID(ctx, testUser.ID)
	if err != nil {
		log.Printf("✗ Get user addresses failed: %v", err)
		return
	}
	log.Printf("✓ User addresses retrieved: %d addresses", len(addresses))

	// 测试地址数量统计
	count, err := addressRepo.CountByUserID(ctx, testUser.ID)
	if err != nil {
		log.Printf("✗ Count user addresses failed: %v", err)
		return
	}
	log.Printf("✓ User address count: %d", count)

	// 测试设置默认地址
	err = addressRepo.SetDefault(ctx, testUser.ID, secondAddress.ID)
	if err != nil {
		log.Printf("✗ Set default address failed: %v", err)
		return
	}
	log.Printf("✓ Default address changed successfully")

	// 验证默认地址变更
	newDefaultAddress, err := addressRepo.GetDefaultByUserID(ctx, testUser.ID)
	if err != nil {
		log.Printf("✗ Get new default address failed: %v", err)
		return
	}
	if newDefaultAddress.ID != secondAddress.ID {
		log.Printf("✗ Default address not changed correctly")
		return
	}
	log.Printf("✓ New default address: %s", newDefaultAddress.RecipientName)

	// 测试更新地址信息
	secondAddress.RecipientName = "更新后的李四"
	secondAddress.Phone = "13700137000"
	err = addressRepo.Update(ctx, secondAddress)
	if err != nil {
		log.Printf("✗ Address update failed: %v", err)
		return
	}
	log.Printf("✓ Address updated successfully")

	// 验证更新结果
	updatedAddress, err := addressRepo.GetByID(ctx, secondAddress.ID)
	if err != nil {
		log.Printf("✗ Get updated address failed: %v", err)
		return
	}
	log.Printf("✓ Address updated to: %s (%s)", updatedAddress.RecipientName, updatedAddress.Phone)

	// 测试地址存在性检查
	exists, err := addressRepo.ExistsByID(ctx, testAddress.ID)
	if err != nil {
		log.Printf("✗ Check address exists failed: %v", err)
		return
	}
	log.Printf("✓ Address exists check: %v", exists)

	// 测试用户地址关联检查
	userOwnsAddress, err := addressRepo.ExistsByUserIDAndID(ctx, testUser.ID, testAddress.ID)
	if err != nil {
		log.Printf("✗ Check user owns address failed: %v", err)
		return
	}
	log.Printf("✓ User owns address check: %v", userOwnsAddress)

	// 测试分页获取地址
	pagedAddresses, total, err := addressRepo.GetUserAddressWithPagination(ctx, testUser.ID, 0, 10)
	if err != nil {
		log.Printf("✗ Get paged addresses failed: %v", err)
		return
	}
	log.Printf("✓ Paged addresses retrieved: %d addresses (total: %d)", len(pagedAddresses), total)

	// 测试检查是否有默认地址
	hasDefault, err := addressRepo.HasDefaultAddress(ctx, testUser.ID)
	if err != nil {
		log.Printf("✗ Check has default address failed: %v", err)
		return
	}
	log.Printf("✓ Has default address: %v", hasDefault)

	// 测试地区查询
	regionAddresses, err := addressRepo.GetAddressesInRegion(ctx, "上海市", "", 10)
	if err != nil {
		log.Printf("✗ Get addresses in region failed: %v", err)
		return
	}
	log.Printf("✓ Addresses in Shanghai: %d addresses", len(regionAddresses))

	// 测试地址统计
	stats, err := addressRepo.GetAddressStatistics(ctx)
	if err != nil {
		log.Printf("✗ Get address statistics failed: %v", err)
		return
	}
	log.Printf("✓ Address statistics: %d total addresses, %d users with addresses", 
		stats.TotalAddresses, stats.UsersWithAddresses)

	// 测试删除地址
	err = addressRepo.Delete(ctx, testAddress.ID)
	if err != nil {
		log.Printf("✗ Address deletion failed: %v", err)
		return
	}
	log.Printf("✓ Address deleted successfully")

	// 验证删除结果
	_, err = addressRepo.GetByID(ctx, testAddress.ID)
	if err == nil {
		log.Printf("✗ Address should have been deleted")
		return
	}
	log.Printf("✓ Address deletion verified")

	// 清理测试数据
	err = addressRepo.DeleteByUserID(ctx, testUser.ID)
	if err != nil {
		log.Printf("✗ Failed to delete user addresses: %v", err)
	}

	err = userRepo.Delete(ctx, testUser.ID)
	if err != nil {
		log.Printf("✗ Failed to delete test user: %v", err)
	}

	log.Println("All Address Repository tests completed!")
}