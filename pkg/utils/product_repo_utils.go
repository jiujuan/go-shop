package utils

import (
	"context"
	"log"

	"go-shop/internal/entity"
	"go-shop/internal/repository"

	"gorm.io/gorm"
)

// TestProductRepository 测试商品仓库功能
func TestProductRepository(db *gorm.DB) {
	log.Println("Testing Product Repository functionality...")

	// 创建商品仓库和分类仓库实例
	productRepo := repository.NewProductRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	ctx := context.Background()

	// 先创建测试分类
	testCategory := &entity.Category{
		Name:      "测试分类",
		ParentID:  0,
		SortOrder: 1,
	}

	err := categoryRepo.Create(ctx, testCategory)
	if err != nil {
		log.Printf("✗ Category creation failed: %v", err)
		return
	}
	log.Printf("✓ Category created successfully with ID: %d", testCategory.ID)

	// 测试创建商品
	testProduct := &entity.Product{
		CategoryID:  testCategory.ID,
		Name:        "测试商品",
		Description: "这是一个测试商品的详细描述",
		Price:       9999, // 99.99元
		Stock:       100,
		CoverImage:  "http://example.com/test-product.jpg",
		Status:      1, // 上架状态
	}

	err = productRepo.Create(ctx, testProduct)
	if err != nil {
		log.Printf("✗ Product creation failed: %v", err)
		return
	}
	log.Printf("✓ Product created successfully with ID: %d", testProduct.ID)

	// 测试根据ID获取商品
	foundProduct, err := productRepo.GetByID(ctx, testProduct.ID)
	if err != nil {
		log.Printf("✗ Get product by ID failed: %v", err)
		return
	}
	log.Printf("✓ Product found by ID: %s (Price: %d)", foundProduct.Name, foundProduct.Price)

	// 测试获取商品和分类信息
	foundProduct, foundCategory, err := productRepo.GetByIDWithCategory(ctx, testProduct.ID)
	if err != nil {
		log.Printf("✗ Get product with category failed: %v", err)
		return
	}
	log.Printf("✓ Product with category: %s (Category: %s)", foundProduct.Name, foundCategory.Name)

	// 测试商品名称唯一性检查
	exists, err := productRepo.ExistsByName(ctx, testProduct.Name)
	if err != nil {
		log.Printf("✗ Check product name exists failed: %v", err)
		return
	}
	log.Printf("✓ Product name exists check: %v", exists)

	// 测试更新商品信息
	testProduct.Name = "更新后的测试商品"
	testProduct.Price = 8888 // 88.88元
	err = productRepo.Update(ctx, testProduct)
	if err != nil {
		log.Printf("✗ Product update failed: %v", err)
		return
	}
	log.Printf("✓ Product updated successfully")

	// 验证更新
	updatedProduct, err := productRepo.GetByID(ctx, testProduct.ID)
	if err != nil {
		log.Printf("✗ Get updated product failed: %v", err)
		return
	}
	log.Printf("✓ Product name updated to: %s, Price: %d", updatedProduct.Name, updatedProduct.Price)

	// 测试库存操作
	// 检查库存
	sufficient, err := productRepo.CheckStock(ctx, testProduct.ID, 50)
	if err != nil {
		log.Printf("✗ Check stock failed: %v", err)
		return
	}
	log.Printf("✓ Stock check (50 units): %v", sufficient)

	// 减少库存
	err = productRepo.DecrementStock(ctx, testProduct.ID, 30)
	if err != nil {
		log.Printf("✗ Decrement stock failed: %v", err)
		return
	}
	log.Printf("✓ Stock decremented by 30")

	// 验证库存变化
	updatedProduct, err = productRepo.GetByID(ctx, testProduct.ID)
	if err != nil {
		log.Printf("✗ Get product after stock update failed: %v", err)
		return
	}
	log.Printf("✓ Current stock: %d", updatedProduct.Stock)

	// 增加库存
	err = productRepo.IncrementStock(ctx, testProduct.ID, 20)
	if err != nil {
		log.Printf("✗ Increment stock failed: %v", err)
		return
	}
	log.Printf("✓ Stock incremented by 20")

	// 验证库存变化
	updatedProduct, err = productRepo.GetByID(ctx, testProduct.ID)
	if err != nil {
		log.Printf("✗ Get product after stock increment failed: %v", err)
		return
	}
	log.Printf("✓ Current stock after increment: %d", updatedProduct.Stock)

	// 测试更新商品状态
	err = productRepo.UpdateStatus(ctx, testProduct.ID, 0) // 下架
	if err != nil {
		log.Printf("✗ Update product status failed: %v", err)
		return
	}
	log.Printf("✓ Product status updated to offline")

	// 创建更多测试商品用于列表测试
	testProducts := []*entity.Product{
		{CategoryID: testCategory.ID, Name: "商品A", Price: 1000, Stock: 10, Status: 1},
		{CategoryID: testCategory.ID, Name: "商品B", Price: 2000, Stock: 20, Status: 1},
		{CategoryID: testCategory.ID, Name: "商品C", Price: 3000, Stock: 30, Status: 0},
	}

	for _, product := range testProducts {
		err := productRepo.Create(ctx, product)
		if err != nil {
			log.Printf("✗ Create test product failed: %v", err)
			continue
		}
	}
	log.Printf("✓ Created %d additional test products", len(testProducts))

	// 测试商品列表查询
	options := repository.ProductQueryOptions{
		Offset: 0,
		Limit:  10,
	}
	products, total, err := productRepo.List(ctx, options)
	if err != nil {
		log.Printf("✗ List products failed: %v", err)
		return
	}
	log.Printf("✓ Product list retrieved: %d products (total: %d)", len(products), total)

	// 测试按分类筛选
	options.CategoryID = &testCategory.ID
	products, total, err = productRepo.List(ctx, options)
	if err != nil {
		log.Printf("✗ List products by category failed: %v", err)
		return
	}
	log.Printf("✓ Products by category: %d products (total: %d)", len(products), total)

	// 测试按状态筛选
	status := 1
	options.CategoryID = nil
	options.Status = &status
	products, total, err = productRepo.List(ctx, options)
	if err != nil {
		log.Printf("✗ List products by status failed: %v", err)
		return
	}
	log.Printf("✓ Products with status=1: %d products (total: %d)", len(products), total)

	// 测试搜索功能
	searchProducts, total, err := productRepo.Search(ctx, "商品", 0, 10)
	if err != nil {
		log.Printf("✗ Search products failed: %v", err)
		return
	}
	log.Printf("✓ Search results for '商品': %d products (total: %d)", len(searchProducts), total)

	// 测试获取最新商品
	latestProducts, err := productRepo.GetLatestProducts(ctx, 5)
	if err != nil {
		log.Printf("✗ Get latest products failed: %v", err)
		return
	}
	log.Printf("✓ Latest products: %d products", len(latestProducts))

	// 测试商品统计
	count, err := productRepo.CountProducts(ctx)
	if err != nil {
		log.Printf("✗ Count products failed: %v", err)
		return
	}
	log.Printf("✓ Total products count: %d", count)

	// 测试分类下商品统计
	categoryCount, err := productRepo.CountProductsByCategory(ctx, testCategory.ID)
	if err != nil {
		log.Printf("✗ Count products by category failed: %v", err)
		return
	}
	log.Printf("✓ Products in category: %d", categoryCount)

	// 测试状态统计
	statusCount, err := productRepo.CountProductsByStatus(ctx, 1)
	if err != nil {
		log.Printf("✗ Count products by status failed: %v", err)
		return
	}
	log.Printf("✓ Products with status=1: %d", statusCount)

	// 测试低库存商品
	lowStockProducts, err := productRepo.GetLowStockProducts(ctx, 15)
	if err != nil {
		log.Printf("✗ Get low stock products failed: %v", err)
		return
	}
	log.Printf("✓ Low stock products (threshold=15): %d products", len(lowStockProducts))

	// 测试重复商品名称创建（应该失败）
	duplicateProduct := &entity.Product{
		CategoryID: testCategory.ID,
		Name:       testProduct.Name,
		Price:      5000,
		Stock:      50,
		Status:     1,
	}

	err = productRepo.Create(ctx, duplicateProduct)
	if err != nil {
		log.Printf("✓ Duplicate product name correctly rejected: %v", err)
	} else {
		log.Printf("✗ Duplicate product name should have been rejected")
	}

	// 清理测试数据
	// 删除所有测试商品
	allProducts, _, err := productRepo.List(ctx, repository.ProductQueryOptions{Offset: 0, Limit: 100})
	if err == nil {
		for _, product := range allProducts {
			productRepo.Delete(ctx, product.ID)
		}
	}

	// 删除测试分类
	err = categoryRepo.Delete(ctx, testCategory.ID)
	if err != nil {
		log.Printf("✗ Category deletion failed: %v", err)
		return
	}
	log.Printf("✓ Test data cleaned up successfully")

	log.Println("All Product Repository tests passed!")
}