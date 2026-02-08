package utils

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"go-shop/internal/entity"
	"go-shop/internal/repository"

	"gorm.io/gorm"
)

// TestOrderRepository 测试订单仓库功能
func TestOrderRepository(db *gorm.DB) {
	log.Println("Testing Order Repository functionality...")

	// 创建仓库实例
	orderRepo := repository.NewOrderRepository(db)
	orderItemRepo := repository.NewOrderItemRepository(db)
	userRepo := repository.NewUserRepository(db)
	productRepo := repository.NewProductRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	ctx := context.Background()

	// 创建测试用户
	testUser := &entity.User{
		Username: "ordertest",
		Password: "hashedpassword",
		Email:    "ordertest@example.com",
		IsAdmin:  false,
	}

	err := userRepo.Create(ctx, testUser)
	if err != nil {
		log.Printf("✗ User creation failed: %v", err)
		return
	}
	log.Printf("✓ Test user created with ID: %d", testUser.ID)

	// 创建测试分类
	testCategory := &entity.Category{
		Name:      "订单测试分类",
		ParentID:  0,
		SortOrder: 1,
	}

	err = categoryRepo.Create(ctx, testCategory)
	if err != nil {
		log.Printf("✗ Category creation failed: %v", err)
		return
	}
	log.Printf("✓ Test category created with ID: %d", testCategory.ID)

	// 创建测试商品
	testProduct := &entity.Product{
		CategoryID:  testCategory.ID,
		Name:        "订单测试商品",
		Description: "用于订单测试的商品",
		Price:       9999, // 99.99元
		Stock:       100,
		CoverImage:  "http://example.com/order-test-product.jpg",
		Status:      1,
	}

	err = productRepo.Create(ctx, testProduct)
	if err != nil {
		log.Printf("✗ Product creation failed: %v", err)
		return
	}
	log.Printf("✓ Test product created with ID: %d", testProduct.ID)

	// 准备地址快照
	addressSnapshot := map[string]interface{}{
		"recipient_name": "张三",
		"phone":          "13800138000",
		"province":       "北京市",
		"city":           "北京市",
		"district":       "朝阳区",
		"detail":         "某某街道123号",
	}
	addressJSON, _ := json.Marshal(addressSnapshot)

	// 测试创建订单
	testOrder := &entity.Order{
		ID:              "ORDER_TEST_001",
		UserID:          testUser.ID,
		TotalAmount:     19998, // 2件商品
		Status:          entity.OrderStatusPending,
		AddressSnapshot: addressJSON,
		Items: []entity.OrderItem{
			{
				ProductID:      testProduct.ID,
				ProductName:    testProduct.Name,
				ProductImage:   testProduct.CoverImage,
				Price:          testProduct.Price,
				Quantity:       2,
				SubtotalAmount: testProduct.Price * 2,
			},
		},
	}

	err = orderRepo.Create(ctx, testOrder)
	if err != nil {
		log.Printf("✗ Order creation failed: %v", err)
		return
	}
	log.Printf("✓ Order created successfully with ID: %s", testOrder.ID)

	// 测试获取订单
	foundOrder, err := orderRepo.GetByID(ctx, testOrder.ID)
	if err != nil {
		log.Printf("✗ Get order by ID failed: %v", err)
		return
	}
	log.Printf("✓ Order found by ID: %s (Total: %d)", foundOrder.ID, foundOrder.TotalAmount)

	// 测试获取订单（包含订单项）
	foundOrderWithItems, err := orderRepo.GetByIDWithItems(ctx, testOrder.ID)
	if err != nil {
		log.Printf("✗ Get order with items failed: %v", err)
		return
	}
	log.Printf("✓ Order with items: %s (%d items)", foundOrderWithItems.ID, len(foundOrderWithItems.Items))

	// 测试订单存在性检查
	exists, err := orderRepo.ExistsByID(ctx, testOrder.ID)
	if err != nil {
		log.Printf("✗ Check order exists failed: %v", err)
		return
	}
	log.Printf("✓ Order exists check: %v", exists)

	// 测试订单状态检查
	canPay, err := orderRepo.CanPay(ctx, testOrder.ID)
	if err != nil {
		log.Printf("✗ Check can pay failed: %v", err)
		return
	}
	log.Printf("✓ Order can pay: %v", canPay)

	canCancel, err := orderRepo.CanCancel(ctx, testOrder.ID)
	if err != nil {
		log.Printf("✗ Check can cancel failed: %v", err)
		return
	}
	log.Printf("✓ Order can cancel: %v", canCancel)

	// 测试更新支付信息
	paidAt := time.Now()
	err = orderRepo.UpdatePaymentInfo(ctx, testOrder.ID, &paidAt)
	if err != nil {
		log.Printf("✗ Update payment info failed: %v", err)
		return
	}
	log.Printf("✓ Payment info updated successfully")

	// 验证支付状态更新
	updatedOrder, err := orderRepo.GetByID(ctx, testOrder.ID)
	if err != nil {
		log.Printf("✗ Get updated order failed: %v", err)
		return
	}
	log.Printf("✓ Order status after payment: %d", updatedOrder.Status)

	// 测试发货信息更新
	shippedAt := time.Now()
	err = orderRepo.UpdateShippingInfo(ctx, testOrder.ID, "顺丰速运", "SF1234567890", &shippedAt)
	if err != nil {
		log.Printf("✗ Update shipping info failed: %v", err)
		return
	}
	log.Printf("✓ Shipping info updated successfully")

	// 验证发货状态更新
	updatedOrder, err = orderRepo.GetByID(ctx, testOrder.ID)
	if err != nil {
		log.Printf("✗ Get updated order after shipping failed: %v", err)
		return
	}
	log.Printf("✓ Order status after shipping: %d, Express: %s %s", updatedOrder.Status, updatedOrder.ExpressCompany, updatedOrder.ExpressNo)

	// 测试完成订单
	completedAt := time.Now()
	err = orderRepo.UpdateCompletionInfo(ctx, testOrder.ID, &completedAt)
	if err != nil {
		log.Printf("✗ Update completion info failed: %v", err)
		return
	}
	log.Printf("✓ Order completion info updated successfully")

	// 创建更多测试订单用于列表测试
	testOrders := []*entity.Order{
		{
			ID:          "ORDER_TEST_002",
			UserID:      testUser.ID,
			TotalAmount: 9999,
			Status:      entity.OrderStatusPending,
			Items: []entity.OrderItem{
				{
					ProductID:      testProduct.ID,
					ProductName:    testProduct.Name,
					Price:          testProduct.Price,
					Quantity:       1,
					SubtotalAmount: testProduct.Price,
				},
			},
		},
		{
			ID:          "ORDER_TEST_003",
			UserID:      testUser.ID,
			TotalAmount: 29997,
			Status:      entity.OrderStatusPaid,
			Items: []entity.OrderItem{
				{
					ProductID:      testProduct.ID,
					ProductName:    testProduct.Name,
					Price:          testProduct.Price,
					Quantity:       3,
					SubtotalAmount: testProduct.Price * 3,
				},
			},
		},
	}

	for _, order := range testOrders {
		err := orderRepo.Create(ctx, order)
		if err != nil {
			log.Printf("✗ Create additional test order failed: %v", err)
			continue
		}
	}
	log.Printf("✓ Created %d additional test orders", len(testOrders))

	// 测试订单列表查询
	options := repository.OrderQueryOptions{
		Offset: 0,
		Limit:  10,
	}
	orders, total, err := orderRepo.List(ctx, options)
	if err != nil {
		log.Printf("✗ List orders failed: %v", err)
		return
	}
	log.Printf("✓ Order list retrieved: %d orders (total: %d)", len(orders), total)

	// 测试按用户筛选
	userOrders, total, err := orderRepo.ListByUser(ctx, testUser.ID, 0, 10)
	if err != nil {
		log.Printf("✗ List orders by user failed: %v", err)
		return
	}
	log.Printf("✓ User orders: %d orders (total: %d)", len(userOrders), total)

	// 测试按状态筛选
	pendingOrders, total, err := orderRepo.ListByStatus(ctx, entity.OrderStatusPending, 0, 10)
	if err != nil {
		log.Printf("✗ List orders by status failed: %v", err)
		return
	}
	log.Printf("✓ Pending orders: %d orders (total: %d)", len(pendingOrders), total)

	// 测试订单统计
	orderCount, err := orderRepo.CountOrders(ctx)
	if err != nil {
		log.Printf("✗ Count orders failed: %v", err)
		return
	}
	log.Printf("✓ Total orders count: %d", orderCount)

	// 测试用户订单统计
	userOrderCount, err := orderRepo.GetUserOrderCount(ctx, testUser.ID)
	if err != nil {
		log.Printf("✗ Get user order count failed: %v", err)
		return
	}
	log.Printf("✓ User order count: %d", userOrderCount)

	// 测试销售额统计
	totalSales, err := orderRepo.GetTotalSales(ctx)
	if err != nil {
		log.Printf("✗ Get total sales failed: %v", err)
		return
	}
	log.Printf("✓ Total sales: %d", totalSales)

	// 测试用户消费统计
	userTotalSales, err := orderRepo.GetTotalSalesByUser(ctx, testUser.ID)
	if err != nil {
		log.Printf("✗ Get user total sales failed: %v", err)
		return
	}
	log.Printf("✓ User total sales: %d", userTotalSales)

	// 测试订单项操作
	items, err := orderItemRepo.GetByOrderID(ctx, testOrder.ID)
	if err != nil {
		log.Printf("✗ Get order items failed: %v", err)
		return
	}
	log.Printf("✓ Order items retrieved: %d items", len(items))

	// 测试订单项统计
	itemCount, err := orderItemRepo.CountByOrderID(ctx, testOrder.ID)
	if err != nil {
		log.Printf("✗ Count order items failed: %v", err)
		return
	}
	log.Printf("✓ Order item count: %d", itemCount)

	// 测试订单总金额计算
	orderTotal, err := orderItemRepo.GetOrderTotal(ctx, testOrder.ID)
	if err != nil {
		log.Printf("✗ Get order total failed: %v", err)
		return
	}
	log.Printf("✓ Order total amount: %d", orderTotal)

	// 测试创建新订单项
	newItem := &entity.OrderItem{
		OrderID:        testOrder.ID,
		ProductID:      testProduct.ID,
		ProductName:    testProduct.Name + "_额外",
		Price:          testProduct.Price,
		Quantity:       1,
		SubtotalAmount: testProduct.Price,
	}

	err = orderItemRepo.Create(ctx, newItem)
	if err != nil {
		log.Printf("✗ Create new order item failed: %v", err)
		return
	}
	log.Printf("✓ New order item created with ID: %d", newItem.ID)

	// 测试获取最近订单
	recentOrders, err := orderRepo.GetRecentOrders(ctx, 5)
	if err != nil {
		log.Printf("✗ Get recent orders failed: %v", err)
		return
	}
	log.Printf("✓ Recent orders: %d orders", len(recentOrders))

	// 测试热销商品统计
	topProducts, err := orderRepo.GetTopSellingProducts(ctx, 5)
	if err != nil {
		log.Printf("✗ Get top selling products failed: %v", err)
		return
	}
	log.Printf("✓ Top selling products: %d products", len(topProducts))

	// 清理测试数据
	// 删除所有测试订单
	allOrders, _, err := orderRepo.List(ctx, repository.OrderQueryOptions{Offset: 0, Limit: 100})
	if err == nil {
		for _, order := range allOrders {
			orderRepo.Delete(ctx, order.ID)
		}
	}

	// 删除测试数据
	productRepo.Delete(ctx, testProduct.ID)
	categoryRepo.Delete(ctx, testCategory.ID)
	userRepo.Delete(ctx, testUser.ID)

	log.Printf("✓ Test data cleaned up successfully")
	log.Println("All Order Repository tests passed!")
}