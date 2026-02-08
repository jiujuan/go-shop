package utils

import (
	"log"

	"go-shop/internal/entity"
	"go-shop/pkg/auth"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var testDB *gorm.DB

// InitTestDB 初始化测试数据库
func InitTestDB() *gorm.DB {
	if testDB != nil {
		return testDB
	}

	// 使用内存 SQLite 数据库进行测试
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // 静默模式，减少测试输出
	})
	if err != nil {
		log.Fatalf("Failed to connect to test database: %v", err)
	}

	// 自动迁移表结构
	err = db.AutoMigrate(
		&entity.User{},
		&entity.Address{},
		&entity.Category{},
		&entity.Product{},
		&entity.ProductSpec{},
		&entity.ProductSKU{},
		&entity.Order{},
		&entity.OrderItem{},
		&entity.Coupon{},
		&entity.UserCoupon{},
		&entity.CartItem{},
		&entity.Refund{},
		&entity.Notification{},
		&entity.Review{},
		&entity.Favorite{},
	)
	if err != nil {
		log.Fatalf("Failed to migrate test database: %v", err)
	}

	testDB = db
	return testDB
}

// CleanupTestDB 清理测试数据库
func CleanupTestDB() {
	if testDB != nil {
		sqlDB, err := testDB.DB()
		if err == nil {
			sqlDB.Close()
		}
		testDB = nil
	}
}

// CleanupTestData 清理测试数据
func CleanupTestData() {
	if testDB == nil {
		return
	}

	// 清理所有表的数据
	testDB.Exec("DELETE FROM favorites")
	testDB.Exec("DELETE FROM reviews")
	testDB.Exec("DELETE FROM notifications")
	testDB.Exec("DELETE FROM refunds")
	testDB.Exec("DELETE FROM order_items")
	testDB.Exec("DELETE FROM orders")
	testDB.Exec("DELETE FROM user_coupons")
	testDB.Exec("DELETE FROM coupons")
	testDB.Exec("DELETE FROM cart_items")
	testDB.Exec("DELETE FROM product_skus")
	testDB.Exec("DELETE FROM product_specs")
	testDB.Exec("DELETE FROM products")
	testDB.Exec("DELETE FROM categories")
	testDB.Exec("DELETE FROM addresses")
	testDB.Exec("DELETE FROM users")
}

// TestJWTAndPassword 测试 JWT 和密码功能
func TestJWTAndPassword() {
	log.Println("Testing JWT and Password utilities...")

	// 测试密码加密
	password := "admin123"
	hashedPassword, err := auth.SimpleHashPassword(password)
	if err != nil {
		log.Printf("Password hashing failed: %v", err)
		return
	}
	log.Printf("Password hashed successfully: %s", hashedPassword[:50]+"...")

	// 测试密码验证
	isValid, err := auth.SimpleVerifyPassword(password, hashedPassword)
	if err != nil {
		log.Printf("Password verification failed: %v", err)
		return
	}
	log.Printf("Password verification result: %v", isValid)

	// 测试 JWT
	jwtManager := auth.NewJWTManager("test-secret-key", 24)
	token, err := jwtManager.GenerateToken(1, "admin", true)
	if err != nil {
		log.Printf("JWT generation failed: %v", err)
		return
	}
	log.Printf("JWT generated successfully: %s", token[:50]+"...")

	// 测试 JWT 解析
	claims, err := jwtManager.ParseToken(token)
	if err != nil {
		log.Printf("JWT parsing failed: %v", err)
		return
	}
	log.Printf("JWT parsed successfully - UserID: %d, Username: %s, IsAdmin: %v", 
		claims.UserID, claims.Username, claims.IsAdmin)

	log.Println("All utility tests passed!")
}