package utils

import (
	"context"
	"log"

	"go-shop/internal/entity"
	"go-shop/internal/repository"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// TestCartRepository 测试购物车仓库功能
func TestCartRepository(db *gorm.DB, redisClient *redis.Client) {
	log.Println("Testing Cart Repository functionality...")

	if redisClient == nil {
		log.Println("✗ Redis client not available, skipping cart repository tests")
		return
	}

	// 创建依赖的仓库实例
	userRepo := repository.NewUserRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	productRepo := repository.NewProductRepository(db)
	cartRepo := repository.NewCartRepository(redisClient, productRepo)
	
	ctx := context.Background()

	// 创建测试数据
	testUser := &entity.User{
		Username: "carttest",
		Password: "hashedpassword",
		Email:    "carttest@example.com",
		IsAdmin:  false,
	}

	err := userRepo.Create(ctx, testUser)
	if err != nil {
		log.Printf("✗ Failed to create test user: %v", err)
		return
	}
	log.Printf("✓ Test user created with ID: %d", testUser.ID)

	// 创建测试分类
	testCategory := &entity.Category{
		Name:      "购物车测试分类",
		ParentID:  0,
		SortOrder: 1,
	}
	err = categoryRepo.Create(ctx, testCategory)
	if err != nil {
		log.Printf("✗ Failed to create test category: %v", err)
		return
	}
	log.Printf("✓ Test category created with ID: %d", testCategory.ID)

	// 创建测试商品
	testProduct := &entity.Product{
		CategoryID:  testCategory.ID,
		Name:        "购物车测试商品",
		Description: "用于购物车测试的商品",
		Price:       9999, // 99.99元
		Stock:       100,
		CoverImage:  "http://example.com/cart-test-product.jpg",
		Status:      1, // 上架状态
	}
	err = productRepo.Create(ctx, testProduct)
	if err != nil {
		log.Printf("✗ Failed to create test product: %v", err)
		return
	}
	log.Printf("✓ Test product created with ID: %d", testProduct.ID)

	// 测试获取空购物车
	cart, err := cartRepo.GetCart(ctx, testUser.ID)
	if err != nil {
		log.Printf("✗ Get empty cart failed: %v", err)
		return
	}
	log.Printf("✓ Empty cart retrieved: %d items", len(cart.Items))

	// 测试添加商品到购物车
	item := entity.CartItem{
		ProductID: testProduct.ID,
		Quantity:  2,
	}

	err = cartRepo.AddItem(ctx, testUser.ID, item)
	if err != nil {
		log.Printf("✗ Add item to cart failed: %v", err)
		return
	}
	log.Printf("✓ Item added to cart successfully")

	// 验证购物车内容
	cart, err = cartRepo.GetCart(ctx, testUser.ID)
	if err != nil {
		log.Printf("✗ Get cart after adding item failed: %v", err)
		return
	}
	log.Printf("✓ Cart contains %d items, total count: %d, total price: %d", 
		len(cart.Items), cart.TotalCount, cart.TotalPrice)

	// 测试购物车查询功能
	itemCount, err := cartRepo.GetItemCount(ctx, testUser.ID)
	if err != nil {
		log.Printf("✗ Get item count failed: %v", err)
		return
	}
	log.Printf("✓ Item count: %d", itemCount)

	totalPrice, err := cartRepo.GetTotalPrice(ctx, testUser.ID)
	if err != nil {
		log.Printf("✗ Get total price failed: %v", err)
		return
	}
	log.Printf("✓ Total price: %d", totalPrice)

	hasItem, err := cartRepo.HasItem(ctx, testUser.ID, testProduct.ID)
	if err != nil {
		log.Printf("✗ Check has item failed: %v", err)
		return
	}
	log.Printf("✓ Has item check: %v", hasItem)

	// 测试更新商品数量
	err = cartRepo.UpdateItemQuantity(ctx, testUser.ID, testProduct.ID, 5)
	if err != nil {
		log.Printf("✗ Update item quantity failed: %v", err)
		return
	}
	log.Printf("✓ Item quantity updated to 5")

	// 验证更新结果
	quantity, err := cartRepo.GetItemQuantity(ctx, testUser.ID, testProduct.ID)
	if err != nil {
		log.Printf("✗ Get item quantity failed: %v", err)
		return
	}
	log.Printf("✓ Current item quantity: %d", quantity)

	// 测试购物车验证
	validationErrors, err := cartRepo.ValidateCart(ctx, testUser.ID)
	if err != nil {
		log.Printf("✗ Validate cart failed: %v", err)
		return
	}
	log.Printf("✓ Cart validation completed: %d errors", len(validationErrors))

	// 测试获取包含商品详情的购物车
	cartWithProducts, err := cartRepo.GetCartWithProducts(ctx, testUser.ID)
	if err != nil {
		log.Printf("✗ Get cart with products failed: %v", err)
		return
	}
	if len(cartWithProducts.Items) > 0 && cartWithProducts.Items[0].Product != nil {
		log.Printf("✓ Cart with products retrieved: product name = %s", 
			cartWithProducts.Items[0].Product.Name)
	}

	// 测试购物车统计
	stats, err := cartRepo.GetCartStatistics(ctx)
	if err != nil {
		log.Printf("✗ Get cart statistics failed: %v", err)
		return
	}
	log.Printf("✓ Cart statistics: %d active carts, %d total items", 
		stats.TotalActiveCarts, stats.TotalItems)

	// 测试删除商品
	err = cartRepo.RemoveItem(ctx, testUser.ID, testProduct.ID)
	if err != nil {
		log.Printf("✗ Remove item failed: %v", err)
		return
	}
	log.Printf("✓ Item removed from cart")

	// 验证删除结果
	isEmpty, err := cartRepo.IsEmpty(ctx, testUser.ID)
	if err != nil {
		log.Printf("✗ Check is empty failed: %v", err)
		return
	}
	log.Printf("✓ Cart is empty: %v", isEmpty)

	// 测试清空购物车
	// 先重新添加商品
	err = cartRepo.AddItem(ctx, testUser.ID, item)
	if err != nil {
		log.Printf("✗ Re-add item failed: %v", err)
		return
	}

	err = cartRepo.ClearCart(ctx, testUser.ID)
	if err != nil {
		log.Printf("✗ Clear cart failed: %v", err)
		return
	}
	log.Printf("✓ Cart cleared successfully")

	// 清理测试数据
	err = userRepo.Delete(ctx, testUser.ID)
	if err != nil {
		log.Printf("✗ Failed to delete test user: %v", err)
	}

	err = productRepo.Delete(ctx, testProduct.ID)
	if err != nil {
		log.Printf("✗ Failed to delete test product: %v", err)
	}

	err = categoryRepo.Delete(ctx, testCategory.ID)
	if err != nil {
		log.Printf("✗ Failed to delete test category: %v", err)
	}

	log.Println("All Cart Repository tests completed!")
}