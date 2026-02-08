package utils

import (
	"context"
	"log"

	"go-shop/internal/entity"
	"go-shop/internal/repository"

	"gorm.io/gorm"
)

// TestUserRepository 测试用户仓库功能
func TestUserRepository(db *gorm.DB) {
	log.Println("Testing User Repository functionality...")

	// 创建用户仓库实例
	userRepo := repository.NewUserRepository(db)
	ctx := context.Background()

	// 测试创建用户
	testUser := &entity.User{
		Username: "testuser123",
		Password: "hashedpassword123",
		Email:    "testuser123@example.com",
		IsAdmin:  false,
	}

	err := userRepo.Create(ctx, testUser)
	if err != nil {
		log.Printf("✗ User creation failed: %v", err)
		return
	}
	log.Printf("✓ User created successfully with ID: %d", testUser.ID)

	// 测试根据ID获取用户
	foundUser, err := userRepo.GetByID(ctx, testUser.ID)
	if err != nil {
		log.Printf("✗ Get user by ID failed: %v", err)
		return
	}
	log.Printf("✓ User found by ID: %s (%s)", foundUser.Username, foundUser.Email)

	// 测试根据用户名获取用户
	foundByUsername, err := userRepo.GetByUsername(ctx, testUser.Username)
	if err != nil {
		log.Printf("✗ Get user by username failed: %v", err)
		return
	}
	log.Printf("✓ User found by username: %s (ID: %d)", foundByUsername.Username, foundByUsername.ID)

	// 测试根据邮箱获取用户
	foundByEmail, err := userRepo.GetByEmail(ctx, testUser.Email)
	if err != nil {
		log.Printf("✗ Get user by email failed: %v", err)
		return
	}
	log.Printf("✓ User found by email: %s (ID: %d)", foundByEmail.Email, foundByEmail.ID)

	// 测试用户名唯一性检查
	exists, err := userRepo.ExistsByUsername(ctx, testUser.Username)
	if err != nil {
		log.Printf("✗ Check username exists failed: %v", err)
		return
	}
	log.Printf("✓ Username exists check: %v", exists)

	// 测试邮箱唯一性检查
	exists, err = userRepo.ExistsByEmail(ctx, testUser.Email)
	if err != nil {
		log.Printf("✗ Check email exists failed: %v", err)
		return
	}
	log.Printf("✓ Email exists check: %v", exists)

	// 测试更新用户信息
	testUser.Email = "updated_testuser123@example.com"
	err = userRepo.Update(ctx, testUser)
	if err != nil {
		log.Printf("✗ User update failed: %v", err)
		return
	}
	log.Printf("✓ User updated successfully")

	// 验证更新
	updatedUser, err := userRepo.GetByID(ctx, testUser.ID)
	if err != nil {
		log.Printf("✗ Get updated user failed: %v", err)
		return
	}
	log.Printf("✓ User email updated to: %s", updatedUser.Email)

	// 测试更新密码
	newPassword := "newhashedpassword123"
	err = userRepo.UpdatePassword(ctx, testUser.ID, newPassword)
	if err != nil {
		log.Printf("✗ Password update failed: %v", err)
		return
	}
	log.Printf("✓ Password updated successfully")

	// 测试用户统计
	count, err := userRepo.CountUsers(ctx)
	if err != nil {
		log.Printf("✗ Count users failed: %v", err)
		return
	}
	log.Printf("✓ Total users count: %d", count)

	// 测试用户列表
	users, total, err := userRepo.List(ctx, 0, 10)
	if err != nil {
		log.Printf("✗ List users failed: %v", err)
		return
	}
	log.Printf("✓ User list retrieved: %d users (total: %d)", len(users), total)

	// 测试重复用户名创建（应该失败）
	duplicateUser := &entity.User{
		Username: testUser.Username,
		Password: "anotherpassword",
		Email:    "another@example.com",
	}

	err = userRepo.Create(ctx, duplicateUser)
	if err != nil {
		log.Printf("✓ Duplicate username correctly rejected: %v", err)
	} else {
		log.Printf("✗ Duplicate username should have been rejected")
	}

	// 清理测试数据
	err = userRepo.Delete(ctx, testUser.ID)
	if err != nil {
		log.Printf("✗ User deletion failed: %v", err)
		return
	}
	log.Printf("✓ Test user deleted successfully")

	log.Println("All User Repository tests passed!")
}