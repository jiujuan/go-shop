package property

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"go-shop/internal/entity"
	"go-shop/internal/repository"
	"go-shop/internal/service"
	"go-shop/pkg/utils"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
)

// InventoryPropertyTestSuite 库存属性测试套件
type InventoryPropertyTestSuite struct {
	suite.Suite
	inventoryService service.InventoryService
	skuService       *service.SKUService
	skuRepo          repository.SKURepository
	productRepo      repository.ProductRepository
	redisClient      *redis.Client
	ctx              context.Context
}

// SetupSuite 设置测试套件
func (suite *InventoryPropertyTestSuite) SetupSuite() {
	// 初始化测试数据库
	db := utils.InitTestDB()

	// 初始化Redis客户端
	suite.redisClient = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   1, // 使用测试数据库
	})

	// 创建仓库实例
	suite.skuRepo = repository.NewSKURepository(db)
	suite.productRepo = repository.NewProductRepository(db)

	// 创建服务实例
	suite.skuService = service.NewSKUService(suite.skuRepo, suite.productRepo)
	suite.inventoryService = service.NewInventoryService(suite.skuRepo, suite.redisClient, db)

	suite.ctx = context.Background()
}

// TearDownSuite 清理测试套件
func (suite *InventoryPropertyTestSuite) TearDownSuite() {
	// 清理Redis
	suite.redisClient.FlushDB(suite.ctx)
	suite.redisClient.Close()
	
	// 清理数据库
	utils.CleanupTestDB()
}

// SetupTest 每个测试前的设置
func (suite *InventoryPropertyTestSuite) SetupTest() {
	// 清理测试数据
	utils.CleanupTestData()
	
	// 清理Redis
	suite.redisClient.FlushDB(suite.ctx)
}

// TestProperty19_InventoryCacheConsistency 属性 19: 库存缓存一致性
// Feature: go-shop-v2-enhancements, Property 19: 库存缓存一致性
// 验证需求：6.1
func (suite *InventoryPropertyTestSuite) TestProperty19_InventoryCacheConsistency() {
	properties := gopter.NewProperties(nil)

	properties.Property("对于任意SKU，当数据库中的库存发生变化后，Redis缓存中的库存数据应该在短时间内（< 1秒）同步更新",
		prop.ForAll(
			func(productID int64, initialStock int, newStock int) bool {
				// 清理测试数据
				utils.CleanupTestData()
				suite.redisClient.FlushDB(suite.ctx)

				// 确保参数有效
				if productID <= 0 {
					productID = 1
				}
				if initialStock < 0 {
					initialStock = 100
				}
				if newStock < 0 {
					newStock = 50
				}

				// 创建测试商品
				product := &entity.Product{
					ID:          productID,
					Name:        "Test Product",
					Description: "Test Description",
					Price:       10000,
					Stock:       initialStock,
					CategoryID:  1,
					Status:      1,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				err := suite.productRepo.Create(suite.ctx, product)
				if err != nil {
					suite.T().Logf("Failed to create product: %v", err)
					return false
				}

				// 创建SKU
				skuCode := fmt.Sprintf("SKU%d", time.Now().UnixNano())
				sku := &entity.ProductSKU{
					ProductID: productID,
					SKUCode:   skuCode,
					SpecValues: entity.SpecValues{
						"颜色": "红色",
						"尺寸": "XL",
					},
					Price:    10000,
					Stock:    initialStock,
					IsActive: true,
				}

				err = suite.skuService.CreateSKU(suite.ctx, sku)
				if err != nil {
					suite.T().Logf("Failed to create SKU: %v", err)
					return false
				}

				// 同步初始库存到Redis
				err = suite.inventoryService.SyncStockToRedis(suite.ctx, sku.ID)
				if err != nil {
					suite.T().Logf("Failed to sync initial stock: %v", err)
					return false
				}

				// 验证初始库存已同步
				inventoryKey := fmt.Sprintf("inventory:sku:%d", sku.ID)
				cachedStock, err := suite.redisClient.Get(suite.ctx, inventoryKey).Int()
				if err != nil {
					suite.T().Logf("Failed to get cached stock: %v", err)
					return false
				}
				if cachedStock != initialStock {
					suite.T().Logf("Initial stock mismatch: expected %d, got %d", initialStock, cachedStock)
					return false
				}

				// 更新数据库库存
				err = suite.skuRepo.UpdateSKUStock(suite.ctx, sku.ID, newStock)
				if err != nil {
					suite.T().Logf("Failed to update stock: %v", err)
					return false
				}

				// 同步库存到Redis
				startTime := time.Now()
				err = suite.inventoryService.SyncStockToRedis(suite.ctx, sku.ID)
				syncDuration := time.Since(startTime)

				if err != nil {
					suite.T().Logf("Failed to sync stock: %v", err)
					return false
				}

				// 验证：同步时间应该小于1秒
				if syncDuration >= time.Second {
					suite.T().Logf("Sync took too long: %v", syncDuration)
					return false
				}

				// 验证：Redis缓存中的库存应该已更新
				updatedCachedStock, err := suite.redisClient.Get(suite.ctx, inventoryKey).Int()
				if err != nil {
					suite.T().Logf("Failed to get updated cached stock: %v", err)
					return false
				}

				if updatedCachedStock != newStock {
					suite.T().Logf("Stock not synced: expected %d, got %d", newStock, updatedCachedStock)
					return false
				}

				return true
			},
			genProductID(),
			genStock(),
			genStock(),
		))

	properties.TestingRun(suite.T(), gopter.ConsoleReporter(false))
}

// TestProperty20_PreDeductStockCorrectness 属性 20: 库存预扣减正确性
// Feature: go-shop-v2-enhancements, Property 20: 库存预扣减正确性
// 验证需求：6.3
func (suite *InventoryPropertyTestSuite) TestProperty20_PreDeductStockCorrectness() {
	properties := gopter.NewProperties(nil)

	properties.Property("对于任意订单创建操作，系统应该在Redis中预扣减库存，记录预扣减状态，预扣减后的可用库存 = 原库存 - 订单数量",
		prop.ForAll(
			func(productID int64, orderID int64, initialStock int, quantity int) bool {
				// 清理测试数据
				utils.CleanupTestData()
				suite.redisClient.FlushDB(suite.ctx)

				// 确保参数有效
				if productID <= 0 {
					productID = 1
				}
				if orderID <= 0 {
					orderID = 1
				}
				if initialStock < 10 {
					initialStock = 100
				}
				if quantity <= 0 || quantity > initialStock {
					quantity = 5
				}

				// 创建测试商品和SKU
				product := &entity.Product{
					ID:          productID,
					Name:        "Test Product",
					Description: "Test Description",
					Price:       10000,
					Stock:       initialStock,
					CategoryID:  1,
					Status:      1,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				err := suite.productRepo.Create(suite.ctx, product)
				if err != nil {
					suite.T().Logf("Failed to create product: %v", err)
					return false
				}

				skuCode := fmt.Sprintf("SKU%d", time.Now().UnixNano())
				sku := &entity.ProductSKU{
					ProductID: productID,
					SKUCode:   skuCode,
					SpecValues: entity.SpecValues{
						"颜色": "红色",
						"尺寸": "XL",
					},
					Price:    10000,
					Stock:    initialStock,
					IsActive: true,
				}

				err = suite.skuService.CreateSKU(suite.ctx, sku)
				if err != nil {
					suite.T().Logf("Failed to create SKU: %v", err)
					return false
				}

				// 同步库存到Redis
				err = suite.inventoryService.SyncStockToRedis(suite.ctx, sku.ID)
				if err != nil {
					suite.T().Logf("Failed to sync stock: %v", err)
					return false
				}

				// 验证初始库存充足
				availableBefore, err := suite.inventoryService.CheckStock(suite.ctx, sku.ID, quantity)
				if err != nil {
					suite.T().Logf("Failed to check stock before: %v", err)
					return false
				}
				if !availableBefore {
					suite.T().Logf("Stock check failed before pre-deduction: need %d, have %d", quantity, initialStock)
					return false
				}

				// 预扣减库存
				err = suite.inventoryService.PreDeductStock(suite.ctx, sku.ID, quantity, orderID)
				if err != nil {
					suite.T().Logf("Failed to pre-deduct stock: %v", err)
					return false
				}

				// 验证：预扣减记录应该存在
				preDeductKey := fmt.Sprintf("inventory:lock:%d:%d", sku.ID, orderID)
				preDeductQuantity, err := suite.redisClient.Get(suite.ctx, preDeductKey).Int()
				if err != nil {
					suite.T().Logf("Pre-deduction record not found: %v", err)
					return false
				}

				if preDeductQuantity != quantity {
					suite.T().Logf("Pre-deduction quantity mismatch: expected %d, got %d", quantity, preDeductQuantity)
					return false
				}

				// 验证：可用库存应该减少
				availableAfter, err := suite.inventoryService.CheckStock(suite.ctx, sku.ID, initialStock-quantity+1)
				if err != nil {
					suite.T().Logf("Failed to check stock after: %v", err)
					return false
				}

				// 可用库存应该不足以满足 (initialStock - quantity + 1) 的需求
				if availableAfter {
					suite.T().Logf("Available stock should be less than %d after pre-deduction", initialStock-quantity+1)
					return false
				}

				// 验证：可用库存应该等于 initialStock - quantity
				canFulfill, err := suite.inventoryService.CheckStock(suite.ctx, sku.ID, initialStock-quantity)
				if err != nil {
					suite.T().Logf("Failed to check exact stock: %v", err)
					return false
				}

				if !canFulfill {
					suite.T().Logf("Should be able to fulfill %d items", initialStock-quantity)
					return false
				}

				return true
			},
			genProductID(),
			genPositiveInt64(),
			genStock(),
			genQuantity(),
		))

	properties.TestingRun(suite.T(), gopter.ConsoleReporter(false))
}

// TestProperty21_ConfirmDeductCorrectness 属性 21: 库存确认扣减正确性
// Feature: go-shop-v2-enhancements, Property 21: 库存确认扣减正确性
// 验证需求：6.4
func (suite *InventoryPropertyTestSuite) TestProperty21_ConfirmDeductCorrectness() {
	properties := gopter.NewProperties(nil)

	properties.Property("对于任意订单支付成功操作，系统应该将预扣减的库存确认扣减到数据库，清除预扣减记录，数据库库存 = 原库存 - 订单数量",
		prop.ForAll(
			func(productID int64, orderID int64, initialStock int, quantity int) bool {
				// 清理测试数据
				utils.CleanupTestData()
				suite.redisClient.FlushDB(suite.ctx)

				// 确保参数有效
				if productID <= 0 {
					productID = 1
				}
				if orderID <= 0 {
					orderID = 1
				}
				if initialStock < 10 {
					initialStock = 100
				}
				if quantity <= 0 || quantity > initialStock {
					quantity = 5
				}

				// 创建测试商品和SKU
				product := &entity.Product{
					ID:          productID,
					Name:        "Test Product",
					Description: "Test Description",
					Price:       10000,
					Stock:       initialStock,
					CategoryID:  1,
					Status:      1,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				err := suite.productRepo.Create(suite.ctx, product)
				if err != nil {
					suite.T().Logf("Failed to create product: %v", err)
					return false
				}

				skuCode := fmt.Sprintf("SKU%d", time.Now().UnixNano())
				sku := &entity.ProductSKU{
					ProductID: productID,
					SKUCode:   skuCode,
					SpecValues: entity.SpecValues{
						"颜色": "红色",
						"尺寸": "XL",
					},
					Price:    10000,
					Stock:    initialStock,
					IsActive: true,
				}

				err = suite.skuService.CreateSKU(suite.ctx, sku)
				if err != nil {
					suite.T().Logf("Failed to create SKU: %v", err)
					return false
				}

				// 同步库存到Redis
				err = suite.inventoryService.SyncStockToRedis(suite.ctx, sku.ID)
				if err != nil {
					suite.T().Logf("Failed to sync stock: %v", err)
					return false
				}

				// 预扣减库存
				err = suite.inventoryService.PreDeductStock(suite.ctx, sku.ID, quantity, orderID)
				if err != nil {
					suite.T().Logf("Failed to pre-deduct stock: %v", err)
					return false
				}

				// 确认扣减
				err = suite.inventoryService.ConfirmDeduct(suite.ctx, sku.ID, orderID)
				if err != nil {
					suite.T().Logf("Failed to confirm deduct: %v", err)
					return false
				}

				// 验证：预扣减记录应该被删除
				preDeductKey := fmt.Sprintf("inventory:lock:%d:%d", sku.ID, orderID)
				_, err = suite.redisClient.Get(suite.ctx, preDeductKey).Result()
				if err != redis.Nil {
					suite.T().Logf("Pre-deduction record should be deleted, but got: %v", err)
					return false
				}

				// 验证：数据库库存应该减少
				updatedSKU, err := suite.skuService.GetSKUByID(suite.ctx, sku.ID)
				if err != nil {
					suite.T().Logf("Failed to get updated SKU: %v", err)
					return false
				}

				expectedStock := initialStock - quantity
				if updatedSKU.Stock != expectedStock {
					suite.T().Logf("Database stock mismatch: expected %d, got %d", expectedStock, updatedSKU.Stock)
					return false
				}

				// 验证：Redis缓存也应该更新
				inventoryKey := fmt.Sprintf("inventory:sku:%d", sku.ID)
				cachedStock, err := suite.redisClient.Get(suite.ctx, inventoryKey).Int()
				if err != nil {
					suite.T().Logf("Failed to get cached stock: %v", err)
					return false
				}

				if cachedStock != expectedStock {
					suite.T().Logf("Cached stock mismatch: expected %d, got %d", expectedStock, cachedStock)
					return false
				}

				return true
			},
			genProductID(),
			genPositiveInt64(),
			genStock(),
			genQuantity(),
		))

	properties.TestingRun(suite.T(), gopter.ConsoleReporter(false))
}

// TestProperty22_ReleaseStockCorrectness 属性 22: 库存释放正确性
// Feature: go-shop-v2-enhancements, Property 22: 库存释放正确性
// 验证需求：6.5
func (suite *InventoryPropertyTestSuite) TestProperty22_ReleaseStockCorrectness() {
	properties := gopter.NewProperties(nil)

	properties.Property("对于任意订单取消操作，系统应该释放Redis中预扣减的库存，清除预扣减记录，释放后的可用库存 = 原库存",
		prop.ForAll(
			func(productID int64, orderID int64, initialStock int, quantity int) bool {
				// 清理测试数据
				utils.CleanupTestData()
				suite.redisClient.FlushDB(suite.ctx)

				// 确保参数有效
				if productID <= 0 {
					productID = 1
				}
				if orderID <= 0 {
					orderID = 1
				}
				if initialStock < 10 {
					initialStock = 100
				}
				if quantity <= 0 || quantity > initialStock {
					quantity = 5
				}

				// 创建测试商品和SKU
				product := &entity.Product{
					ID:          productID,
					Name:        "Test Product",
					Description: "Test Description",
					Price:       10000,
					Stock:       initialStock,
					CategoryID:  1,
					Status:      1,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				err := suite.productRepo.Create(suite.ctx, product)
				if err != nil {
					suite.T().Logf("Failed to create product: %v", err)
					return false
				}

				skuCode := fmt.Sprintf("SKU%d", time.Now().UnixNano())
				sku := &entity.ProductSKU{
					ProductID: productID,
					SKUCode:   skuCode,
					SpecValues: entity.SpecValues{
						"颜色": "红色",
						"尺寸": "XL",
					},
					Price:    10000,
					Stock:    initialStock,
					IsActive: true,
				}

				err = suite.skuService.CreateSKU(suite.ctx, sku)
				if err != nil {
					suite.T().Logf("Failed to create SKU: %v", err)
					return false
				}

				// 同步库存到Redis
				err = suite.inventoryService.SyncStockToRedis(suite.ctx, sku.ID)
				if err != nil {
					suite.T().Logf("Failed to sync stock: %v", err)
					return false
				}

				// 预扣减库存
				err = suite.inventoryService.PreDeductStock(suite.ctx, sku.ID, quantity, orderID)
				if err != nil {
					suite.T().Logf("Failed to pre-deduct stock: %v", err)
					return false
				}

				// 验证预扣减后可用库存减少
				canFulfillAfterPreDeduct, err := suite.inventoryService.CheckStock(suite.ctx, sku.ID, initialStock)
				if err != nil {
					suite.T().Logf("Failed to check stock after pre-deduct: %v", err)
					return false
				}
				if canFulfillAfterPreDeduct {
					suite.T().Logf("Available stock should be less than initial after pre-deduction")
					return false
				}

				// 释放库存
				err = suite.inventoryService.ReleaseStock(suite.ctx, sku.ID, orderID)
				if err != nil {
					suite.T().Logf("Failed to release stock: %v", err)
					return false
				}

				// 验证：预扣减记录应该被删除
				preDeductKey := fmt.Sprintf("inventory:lock:%d:%d", sku.ID, orderID)
				_, err = suite.redisClient.Get(suite.ctx, preDeductKey).Result()
				if err != redis.Nil {
					suite.T().Logf("Pre-deduction record should be deleted, but got: %v", err)
					return false
				}

				// 验证：可用库存应该恢复到原始值
				canFulfillAfterRelease, err := suite.inventoryService.CheckStock(suite.ctx, sku.ID, initialStock)
				if err != nil {
					suite.T().Logf("Failed to check stock after release: %v", err)
					return false
				}

				if !canFulfillAfterRelease {
					suite.T().Logf("Available stock should be restored to initial value after release")
					return false
				}

				return true
			},
			genProductID(),
			genPositiveInt64(),
			genStock(),
			genQuantity(),
		))

	properties.TestingRun(suite.T(), gopter.ConsoleReporter(false))
}

// TestProperty23_InsufficientStockRejection 属性 23: 库存不足拒绝
// Feature: go-shop-v2-enhancements, Property 23: 库存不足拒绝
// 验证需求：6.6
func (suite *InventoryPropertyTestSuite) TestProperty23_InsufficientStockRejection() {
	properties := gopter.NewProperties(nil)

	properties.Property("对于任意加购或下单请求，当SKU的可用库存（实际库存 - 预扣减库存）小于请求数量时，系统应该拒绝该请求",
		prop.ForAll(
			func(productID int64, initialStock int, requestQuantity int) bool {
				// 清理测试数据
				utils.CleanupTestData()
				suite.redisClient.FlushDB(suite.ctx)

				// 确保参数有效
				if productID <= 0 {
					productID = 1
				}
				if initialStock < 1 {
					initialStock = 10
				}
				// 确保请求数量大于库存
				if requestQuantity <= initialStock {
					requestQuantity = initialStock + 10
				}

				// 创建测试商品和SKU
				product := &entity.Product{
					ID:          productID,
					Name:        "Test Product",
					Description: "Test Description",
					Price:       10000,
					Stock:       initialStock,
					CategoryID:  1,
					Status:      1,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				err := suite.productRepo.Create(suite.ctx, product)
				if err != nil {
					suite.T().Logf("Failed to create product: %v", err)
					return false
				}

				skuCode := fmt.Sprintf("SKU%d", time.Now().UnixNano())
				sku := &entity.ProductSKU{
					ProductID: productID,
					SKUCode:   skuCode,
					SpecValues: entity.SpecValues{
						"颜色": "红色",
						"尺寸": "XL",
					},
					Price:    10000,
					Stock:    initialStock,
					IsActive: true,
				}

				err = suite.skuService.CreateSKU(suite.ctx, sku)
				if err != nil {
					suite.T().Logf("Failed to create SKU: %v", err)
					return false
				}

				// 同步库存到Redis
				err = suite.inventoryService.SyncStockToRedis(suite.ctx, sku.ID)
				if err != nil {
					suite.T().Logf("Failed to sync stock: %v", err)
					return false
				}

				// 验证：库存检查应该返回false（库存不足）
				available, err := suite.inventoryService.CheckStock(suite.ctx, sku.ID, requestQuantity)
				if err != nil {
					suite.T().Logf("Failed to check stock: %v", err)
					return false
				}

				if available {
					suite.T().Logf("Stock check should return false when request quantity (%d) > available stock (%d)", requestQuantity, initialStock)
					return false
				}

				// 验证：预扣减应该失败
				orderID := int64(1)
				err = suite.inventoryService.PreDeductStock(suite.ctx, sku.ID, requestQuantity, orderID)
				if err == nil {
					suite.T().Logf("Pre-deduct should fail when stock is insufficient")
					return false
				}

				// 验证错误类型
				if err != entity.ErrProductOutOfStock {
					suite.T().Logf("Expected ErrProductOutOfStock, got: %v", err)
					return false
				}

				return true
			},
			genProductID(),
			genStock(),
			gen.IntRange(20, 200), // 请求数量大于库存
		))

	properties.TestingRun(suite.T(), gopter.ConsoleReporter(false))
}

// TestProperty24_ConcurrentInventoryDeductionSafety 属性 24: 并发库存扣减安全性（关键属性）
// Feature: go-shop-v2-enhancements, Property 24: 并发库存扣减安全性
// 验证需求：6.8
func (suite *InventoryPropertyTestSuite) TestProperty24_ConcurrentInventoryDeductionSafety() {
	properties := gopter.NewProperties(nil)

	properties.Property("对于任意N个并发的订单创建请求（购买同一SKU），当初始库存为M时，最终库存永远不应该为负数（防超卖）",
		prop.ForAll(
			func(productID int64, initialStock int, orderQuantities []int) bool {
				// 清理测试数据
				utils.CleanupTestData()
				suite.redisClient.FlushDB(suite.ctx)

				// 确保参数有效
				if productID <= 0 {
					productID = 1
				}
				if initialStock < 10 {
					initialStock = 50
				}
				if len(orderQuantities) == 0 {
					orderQuantities = []int{5, 5, 5}
				}
				// 限制并发数量，避免测试时间过长
				if len(orderQuantities) > 10 {
					orderQuantities = orderQuantities[:10]
				}
				// 确保每个订单数量有效
				for i := range orderQuantities {
					if orderQuantities[i] <= 0 {
						orderQuantities[i] = 5
					}
					if orderQuantities[i] > initialStock {
						orderQuantities[i] = initialStock / 2
					}
				}

				// 创建测试商品和SKU
				product := &entity.Product{
					ID:          productID,
					Name:        "Test Product",
					Description: "Test Description",
					Price:       10000,
					Stock:       initialStock,
					CategoryID:  1,
					Status:      1,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				err := suite.productRepo.Create(suite.ctx, product)
				if err != nil {
					suite.T().Logf("Failed to create product: %v", err)
					return false
				}

				skuCode := fmt.Sprintf("SKU%d", time.Now().UnixNano())
				sku := &entity.ProductSKU{
					ProductID: productID,
					SKUCode:   skuCode,
					SpecValues: entity.SpecValues{
						"颜色": "红色",
						"尺寸": "XL",
					},
					Price:    10000,
					Stock:    initialStock,
					IsActive: true,
				}

				err = suite.skuService.CreateSKU(suite.ctx, sku)
				if err != nil {
					suite.T().Logf("Failed to create SKU: %v", err)
					return false
				}

				// 同步库存到Redis
				err = suite.inventoryService.SyncStockToRedis(suite.ctx, sku.ID)
				if err != nil {
					suite.T().Logf("Failed to sync stock: %v", err)
					return false
				}

				// 并发创建订单（预扣减库存）
				var wg sync.WaitGroup
				results := make([]error, len(orderQuantities))
				successCount := 0
				totalRequested := 0

				for i, quantity := range orderQuantities {
					totalRequested += quantity
					wg.Add(1)
					go func(index int, qty int) {
						defer wg.Done()
						orderID := int64(index + 1)
						err := suite.inventoryService.PreDeductStock(suite.ctx, sku.ID, qty, orderID)
						results[index] = err
					}(i, quantity)
				}

				wg.Wait()

				// 统计成功的订单数量和总数量
				successTotal := 0
				for i, err := range results {
					if err == nil {
						successCount++
						successTotal += orderQuantities[i]
					}
				}

				// 验证：成功订单的总数量应该 <= 初始库存
				if successTotal > initialStock {
					suite.T().Logf("Success total (%d) exceeds initial stock (%d)", successTotal, initialStock)
					return false
				}

				// 获取最终的可用库存
				availableStock := initialStock - successTotal

				// 验证：最终可用库存应该 >= 0（防超卖）
				if availableStock < 0 {
					suite.T().Logf("Available stock is negative: %d (initial: %d, success total: %d)", availableStock, initialStock, successTotal)
					return false
				}

				// 验证：如果所有订单的总数量 <= 初始库存，则成功订单的总数量应该 <= 总请求数量
				// 注意：由于并发和分布式锁的特性，不是所有订单都一定能成功获取锁
				// 但成功的订单总数量不应该超过初始库存
				if totalRequested <= initialStock {
					if successTotal > totalRequested {
						suite.T().Logf("Success total (%d) should not exceed total requested (%d)", successTotal, totalRequested)
						return false
					}
					// 至少应该有一些订单成功（如果总请求量合理）
					if successCount == 0 && totalRequested > 0 {
						suite.T().Logf("At least some orders should succeed when total requested (%d) <= initial stock (%d)", totalRequested, initialStock)
						return false
					}
				}

				// 验证：如果所有订单的总数量 > 初始库存，则只有部分订单成功
				if totalRequested > initialStock {
					if successTotal > initialStock {
						suite.T().Logf("Success total (%d) should not exceed initial stock (%d)", successTotal, initialStock)
						return false
					}
				}

				// 验证：通过CheckStock验证可用库存
				canFulfillMore, err := suite.inventoryService.CheckStock(suite.ctx, sku.ID, availableStock+1)
				if err != nil {
					suite.T().Logf("Failed to check stock: %v", err)
					return false
				}
				if canFulfillMore {
					suite.T().Logf("Should not be able to fulfill more than available stock")
					return false
				}

				return true
			},
			genProductID(),
			gen.IntRange(20, 100), // 初始库存
			gen.SliceOfN(5, gen.IntRange(5, 20)), // 5个并发订单，每个5-20件
		))

	properties.TestingRun(suite.T(), gopter.ConsoleReporter(false))
}

// TestInventoryPropertyTestSuite 运行测试套件
func TestInventoryPropertyTestSuite(t *testing.T) {
	suite.Run(t, new(InventoryPropertyTestSuite))
}
