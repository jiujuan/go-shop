package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/apache/rocketmq-client-go/v2/primitive"

	"go-shop/internal/dto"
	"go-shop/internal/entity"
	internalMQ "go-shop/internal/mq"
	"go-shop/internal/repository"
	"go-shop/pkg/mq"
)

type OrderService struct {
	orderRepo            repository.OrderRepository
	cartRepo             repository.CartRepository
	productRepo          repository.ProductRepository
	addressRepo          repository.AddressRepository
	skuRepo              repository.SKURepository
	couponService        *CouponService
	inventoryService     InventoryService
	mqProducer           *mq.Producer
	notificationProducer *internalMQ.NotificationProducer
}

func NewOrderService(
	orderRepo repository.OrderRepository,
	cartRepo repository.CartRepository,
	productRepo repository.ProductRepository,
	addressRepo repository.AddressRepository,
	skuRepo repository.SKURepository,
	couponService *CouponService,
	inventoryService InventoryService,
	mqProducer *mq.Producer,
) *OrderService {
	var notificationProducer *internalMQ.NotificationProducer
	if mqProducer != nil {
		notificationProducer = internalMQ.NewNotificationProducer(mqProducer)
	}

	return &OrderService{
		orderRepo:            orderRepo,
		cartRepo:             cartRepo,
		productRepo:          productRepo,
		addressRepo:          addressRepo,
		skuRepo:              skuRepo,
		couponService:        couponService,
		inventoryService:     inventoryService,
		mqProducer:           mqProducer,
		notificationProducer: notificationProducer,
	}
}

// CreateOrder 创建订单（从购物车）
// 需求：1.6 - 当用户创建订单时，系统应基于SKU计算价格和扣减库存
// 需求：2.5 - 当用户选择优惠券时，系统应自动计算折扣后的订单金额
// 需求：2.6 - 当用户提交订单时，系统应验证优惠券的有效性
func (s *OrderService) CreateOrder(ctx context.Context, userID int64, req *dto.OrderCreateRequest) (*dto.OrderResponse, error) {
	if userID <= 0 {
		return nil, entity.ErrInvalidUserID
	}

	if req.AddressID <= 0 {
		return nil, entity.ErrInvalidAddressID
	}

	// 验证购物车
	cart, err := s.cartRepo.GetCartWithProducts(ctx, userID)
	if err != nil {
		return nil, err
	}

	if len(cart.Items) == 0 {
		return nil, entity.ErrCartEmpty
	}

	// 验证购物车商品
	validationErrors, err := s.cartRepo.ValidateCart(ctx, userID)
	if err != nil {
		return nil, err
	}

	if len(validationErrors) > 0 {
		return nil, errors.New("购物车商品验证失败")
	}

	// 获取地址信息
	address, err := s.addressRepo.GetByID(ctx, req.AddressID)
	if err != nil {
		return nil, err
	}

	// 验证地址属于当前用户
	if address.UserID != userID {
		return nil, errors.New("地址不属于当前用户")
	}

	// 创建地址快照
	addressSnapshot := dto.AddressSnapshot{
		RecipientName: address.RecipientName,
		Phone:         address.Phone,
		Province:      address.Province,
		City:          address.City,
		District:      address.District,
		Detail:        address.Detail,
	}

	addressJSON, err := json.Marshal(addressSnapshot)
	if err != nil {
		return nil, err
	}

	// 计算订单总金额（商品金额）
	totalAmount := cart.TotalPrice

	// 处理优惠券（如果提供）
	var couponID *int64
	var couponDiscount int64 = 0
	var userCouponIDToStore *int64

	if req.UserCouponID != nil {
		userCouponIDToStore = req.UserCouponID

		// 获取用户优惠券信息（包含优惠券详情）
		userCoupon, err := s.couponService.userCouponRepo.GetByIDWithCoupon(ctx, *req.UserCouponID)
		if err != nil {
			return nil, errors.New("优惠券不存在")
		}

		// 验证优惠券属于当前用户
		if userCoupon.UserID != userID {
			return nil, errors.New("优惠券不属于当前用户")
		}

		// 验证优惠券是否可用（需求 2.6）
		if !userCoupon.IsUsable() {
			if userCoupon.Status == entity.UserCouponStatusUsed {
				return nil, errors.New("优惠券已使用")
			}
			if userCoupon.Status == entity.UserCouponStatusExpired || userCoupon.IsExpired() {
				return nil, errors.New("优惠券已过期")
			}
			return nil, errors.New("优惠券不可用")
		}

		// 计算折扣金额（需求 2.5）
		// 将分转换为元进行计算
		orderAmountYuan := float64(totalAmount) / 100.0
		discountYuan, err := s.couponService.CalculateDiscount(ctx, userCoupon.Coupon, orderAmountYuan)
		if err != nil {
			return nil, fmt.Errorf("计算优惠券折扣失败: %w", err)
		}

		// 将元转换回分
		couponDiscount = int64(discountYuan * 100)
		couponID = &userCoupon.CouponID
	}

	// 计算最终支付金额
	finalAmount := totalAmount - couponDiscount
	if finalAmount < 0 {
		finalAmount = 0
	}

	// 创建订单
	order := &entity.Order{
		ID:              s.generateOrderID(),
		UserID:          userID,
		TotalAmount:     totalAmount,
		CouponID:        couponID,
		UserCouponID:    userCouponIDToStore,
		CouponDiscount:  couponDiscount,
		FinalAmount:     finalAmount,
		Status:          entity.OrderStatusPending,
		AddressSnapshot: addressJSON,
		Items:           make([]entity.OrderItem, 0, len(cart.Items)),
	}

	// 创建订单项并预扣减库存
	// 需求：6.2, 6.3 - 使用库存服务预扣减库存
	for _, cartItem := range cart.Items {
		if cartItem.Product == nil {
			continue
		}

		var price int64
		var productImage string
		var skuID int64

		// 如果有SKU，使用SKU的价格和库存
		if cartItem.SKUID != nil {
			sku, err := s.skuRepo.GetSKUByID(ctx, *cartItem.SKUID)
			if err != nil {
				return nil, errors.New("SKU不存在")
			}

			skuID = *cartItem.SKUID
			price = sku.Price
			if sku.Image != "" {
				productImage = sku.Image
			} else {
				productImage = cartItem.Product.CoverImage
			}

			// 预扣减SKU库存（使用库存服务）
			// 需求：6.3 - 预扣减库存并记录状态
			orderIDInt := s.orderIDToInt64(order.ID)
			err = s.inventoryService.PreDeductStock(ctx, skuID, cartItem.Quantity, orderIDInt)
			if err != nil {
				// 回滚已预扣减的库存
				s.rollbackPreDeductedStock(ctx, order.Items, order.ID)
				return nil, fmt.Errorf("预扣减库存失败: %w", err)
			}
		} else {
			// 没有SKU，扣减商品库存（原子操作）
			err := s.productRepo.DecrementStock(ctx, cartItem.ProductID, cartItem.Quantity)
			if err != nil {
				// 回滚已扣减的库存
				s.rollbackDeductedStock(ctx, order.Items)
				return nil, err
			}

			price = cartItem.Product.Price
			productImage = cartItem.Product.CoverImage
		}

		// 转换SpecValues
		var specValues *entity.SpecValues
		if cartItem.SpecValues != nil {
			specValues = cartItem.SpecValues
		}

		orderItem := entity.OrderItem{
			OrderID:        order.ID,
			ProductID:      cartItem.ProductID,
			SKUID:          cartItem.SKUID,
			SKUCode:        cartItem.SKUCode,
			SpecValues:     specValues,
			ProductName:    cartItem.Product.Name,
			ProductImage:   productImage,
			Price:          price,
			Quantity:       cartItem.Quantity,
			SubtotalAmount: int64(cartItem.Quantity) * price,
		}

		order.Items = append(order.Items, orderItem)
	}

	// 保存订单
	if err := s.orderRepo.Create(ctx, order); err != nil {
		// 回滚库存
		s.rollbackPreDeductedStock(ctx, order.Items, order.ID)
		s.rollbackDeductedStock(ctx, order.Items)
		return nil, err
	}

	// 发送订单创建通知（需求 8.1）
	if s.notificationProducer != nil {
		err := s.notificationProducer.SendOrderNotification(ctx, userID, order.ID, entity.OrderStatusPending)
		if err != nil {
			// 记录错误但不影响订单创建
			fmt.Printf("warning: failed to send order notification: %v\n", err)
		}
	}

	// 清空购物车
	s.cartRepo.ClearCart(ctx, userID)

	return s.entityToResponse(order), nil
}

// GetOrderByID 获取订单详情
func (s *OrderService) GetOrderByID(ctx context.Context, userID int64, orderID string) (*dto.OrderResponse, error) {
	if userID <= 0 {
		return nil, entity.ErrInvalidUserID
	}

	if orderID == "" {
		return nil, errors.New("订单ID不能为空")
	}

	order, err := s.orderRepo.GetByIDWithItems(ctx, orderID)
	if err != nil {
		return nil, err
	}

	// 验证订单属于当前用户
	if order.UserID != userID {
		return nil, errors.New("订单不属于当前用户")
	}

	return s.entityToResponse(order), nil
}

// GetUserOrders 获取用户订单列表
func (s *OrderService) GetUserOrders(ctx context.Context, userID int64, req *dto.OrderListRequest) (*dto.OrderListResponse, error) {
	if userID <= 0 {
		return nil, entity.ErrInvalidUserID
	}

	page, pageSize := s.getDefaultPagination(req.Page, req.PageSize)
	offset := (page - 1) * pageSize

	var orders []*entity.Order
	var total int64
	var err error

	if req.Status != nil {
		orders, total, err = s.orderRepo.ListByUserAndStatus(ctx, userID, *req.Status, offset, pageSize)
	} else {
		orders, total, err = s.orderRepo.ListByUser(ctx, userID, offset, pageSize)
	}

	if err != nil {
		return nil, err
	}

	orderResponses := make([]dto.OrderResponse, len(orders))
	for i, order := range orders {
		orderResponses[i] = *s.entityToResponse(order)
	}

	pagination := dto.NewPaginationResponse(page, pageSize, total)

	return &dto.OrderListResponse{
		Orders:     orderResponses,
		Pagination: pagination,
	}, nil
}

// PayOrder 支付订单
// 需求：2.7 - 当订单支付成功后，系统应将优惠券标记为已使用
// 需求：6.4 - 当订单支付成功后，确认库存扣减并更新数据库
func (s *OrderService) PayOrder(ctx context.Context, userID int64, orderID string) (*dto.OrderResponse, error) {
	if userID <= 0 {
		return nil, entity.ErrInvalidUserID
	}

	order, err := s.orderRepo.GetByIDWithItems(ctx, orderID)
	if err != nil {
		return nil, err
	}

	if order.UserID != userID {
		return nil, errors.New("订单不属于当前用户")
	}

	if order.Status != entity.OrderStatusPending {
		return nil, errors.New("订单状态不允许支付")
	}

	// 确认库存扣减（需求 6.4）
	orderIDInt := s.orderIDToInt64(orderID)
	for _, item := range order.Items {
		if item.SKUID != nil {
			// 确认SKU库存扣减
			err := s.inventoryService.ConfirmDeduct(ctx, *item.SKUID, orderIDInt)
			if err != nil {
				return nil, fmt.Errorf("确认库存扣减失败: %w", err)
			}

			// 发送库存扣减事件到RocketMQ
			s.sendInventoryEvent(ctx, mq.TagInventoryDeducted, *item.SKUID, item.Quantity, orderID)
		}
	}

	// 更新订单状态为已支付
	now := time.Now()
	if err := s.orderRepo.UpdateStatus(ctx, orderID, entity.OrderStatusPaid); err != nil {
		return nil, err
	}

	if err := s.orderRepo.UpdatePaymentInfo(ctx, orderID, &now); err != nil {
		return nil, err
	}

	// 如果订单使用了优惠券，标记优惠券为已使用（需求 2.7）
	if order.UserCouponID != nil {
		// 将 orderID string 转换为可用的格式
		// 注意：UseCoupon 方法需要 orderID 为 int64，但我们的 orderID 是 string
		// 这里我们传递 0，因为 UseCoupon 主要是标记状态，orderID 用于记录
		if err := s.couponService.UseCoupon(ctx, *order.UserCouponID, 0); err != nil {
			// 记录错误但不影响支付流程
			// 在生产环境中应该记录日志
			// 可以考虑使用消息队列异步处理
		}
	}

	// 发送订单支付成功通知（需求 8.1）
	if s.notificationProducer != nil {
		err := s.notificationProducer.SendOrderNotification(ctx, userID, orderID, entity.OrderStatusPaid)
		if err != nil {
			// 记录错误但不影响支付流程
			fmt.Printf("warning: failed to send order paid notification: %v\n", err)
		}
	}

	// 重新获取订单
	order, err = s.orderRepo.GetByIDWithItems(ctx, orderID)
	if err != nil {
		return nil, err
	}

	return s.entityToResponse(order), nil
}

// CancelOrder 取消订单
// 需求：6.5 - 当订单取消时，释放预扣减的库存
func (s *OrderService) CancelOrder(ctx context.Context, userID int64, orderID string) error {
	if userID <= 0 {
		return entity.ErrInvalidUserID
	}

	order, err := s.orderRepo.GetByIDWithItems(ctx, orderID)
	if err != nil {
		return err
	}

	if order.UserID != userID {
		return errors.New("订单不属于当前用户")
	}

	// 只有待支付状态的订单可以取消
	if order.Status != entity.OrderStatusPending {
		return entity.ErrOrderCannotCancel
	}

	// 更新订单状态
	if err := s.orderRepo.UpdateStatus(ctx, orderID, entity.OrderStatusCancelled); err != nil {
		return err
	}

	// 释放库存（需求 6.5）
	orderIDInt := s.orderIDToInt64(orderID)
	for _, item := range order.Items {
		if item.SKUID != nil {
			// 释放SKU预扣减库存
			err := s.inventoryService.ReleaseStock(ctx, *item.SKUID, orderIDInt)
			if err != nil {
				// 记录错误但继续处理
				fmt.Printf("warning: failed to release stock for sku %d: %v\n", *item.SKUID, err)
			}

			// 发送库存恢复事件到RocketMQ
			s.sendInventoryEvent(ctx, mq.TagInventoryRestored, *item.SKUID, item.Quantity, orderID)
		} else {
			// 恢复商品库存（没有使用库存服务的情况）
			if err := s.productRepo.IncrementStock(ctx, item.ProductID, item.Quantity); err != nil {
				return err
			}
		}
	}

	// 发送订单取消通知（需求 8.1）
	if s.notificationProducer != nil {
		err := s.notificationProducer.SendOrderNotification(ctx, userID, orderID, entity.OrderStatusCancelled)
		if err != nil {
			// 记录错误但不影响取消流程
			fmt.Printf("warning: failed to send order cancelled notification: %v\n", err)
		}
	}

	return nil
}

// ShipOrder 发货（管理员）
func (s *OrderService) ShipOrder(ctx context.Context, orderID string, expressCompany, expressNo string) (*dto.OrderResponse, error) {
	if orderID == "" {
		return nil, errors.New("订单ID不能为空")
	}

	if expressCompany == "" || expressNo == "" {
		return nil, errors.New("快递公司和快递单号不能为空")
	}

	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	if order.Status != entity.OrderStatusPaid {
		return nil, errors.New("只有已支付的订单可以发货")
	}

	// 更新订单状态和物流信息
	now := time.Now()
	if err := s.orderRepo.UpdateStatus(ctx, orderID, entity.OrderStatusShipped); err != nil {
		return nil, err
	}

	if err := s.orderRepo.UpdateShippingInfo(ctx, orderID, expressCompany, expressNo, &now); err != nil {
		return nil, err
	}

	// 发送订单发货通知（需求 8.1）
	if s.notificationProducer != nil {
		err := s.notificationProducer.SendOrderNotification(ctx, order.UserID, orderID, entity.OrderStatusShipped)
		if err != nil {
			// 记录错误但不影响发货流程
			fmt.Printf("warning: failed to send order shipped notification: %v\n", err)
		}
	}

	order, err = s.orderRepo.GetByIDWithItems(ctx, orderID)
	if err != nil {
		return nil, err
	}

	return s.entityToResponse(order), nil
}

// CompleteOrder 完成订单
func (s *OrderService) CompleteOrder(ctx context.Context, userID int64, orderID string) (*dto.OrderResponse, error) {
	if userID <= 0 {
		return nil, entity.ErrInvalidUserID
	}

	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	if order.UserID != userID {
		return nil, errors.New("订单不属于当前用户")
	}

	if order.Status != entity.OrderStatusShipped {
		return nil, errors.New("只有已发货的订单可以完成")
	}

	// 更新订单状态
	now := time.Now()
	if err := s.orderRepo.UpdateStatus(ctx, orderID, entity.OrderStatusCompleted); err != nil {
		return nil, err
	}

	if err := s.orderRepo.UpdateCompletionInfo(ctx, orderID, &now); err != nil {
		return nil, err
	}

	// 发送订单完成通知（需求 8.1）
	if s.notificationProducer != nil {
		err := s.notificationProducer.SendOrderNotification(ctx, userID, orderID, entity.OrderStatusCompleted)
		if err != nil {
			// 记录错误但不影响完成流程
			fmt.Printf("warning: failed to send order completed notification: %v\n", err)
		}
	}

	order, err = s.orderRepo.GetByIDWithItems(ctx, orderID)
	if err != nil {
		return nil, err
	}

	return s.entityToResponse(order), nil
}

// GetAllOrders 获取所有订单（管理员）
func (s *OrderService) GetAllOrders(ctx context.Context, req *dto.OrderListRequest) (*dto.OrderListResponse, error) {
	page, pageSize := s.getDefaultPagination(req.Page, req.PageSize)
	offset := (page - 1) * pageSize

	options := repository.OrderQueryOptions{
		Status:    req.Status,
		SortBy:    "created_at",
		SortOrder: "desc",
		Offset:    offset,
		Limit:     pageSize,
	}

	orders, total, err := s.orderRepo.List(ctx, options)
	if err != nil {
		return nil, err
	}

	orderResponses := make([]dto.OrderResponse, len(orders))
	for i, order := range orders {
		orderResponses[i] = *s.entityToResponse(order)
	}

	pagination := dto.NewPaginationResponse(page, pageSize, total)

	return &dto.OrderListResponse{
		Orders:     orderResponses,
		Pagination: pagination,
	}, nil
}

// UpdateOrderStatus 更新订单状态（管理员）
func (s *OrderService) UpdateOrderStatus(ctx context.Context, orderID string, req *dto.OrderUpdateStatusRequest) (*dto.OrderResponse, error) {
	if orderID == "" {
		return nil, errors.New("订单ID不能为空")
	}

	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	// 验证状态转换是否合法
	if !s.isValidStatusTransition(order.Status, req.Status) {
		return nil, entity.ErrOrderStatusInvalid
	}

	// 更新订单状态
	if err := s.orderRepo.UpdateStatus(ctx, orderID, req.Status); err != nil {
		return nil, err
	}

	// 如果是发货状态，更新物流信息
	if req.Status == entity.OrderStatusShipped && req.ExpressCompany != nil && req.ExpressNo != nil {
		now := time.Now()
		s.orderRepo.UpdateShippingInfo(ctx, orderID, *req.ExpressCompany, *req.ExpressNo, &now)
	}

	order, err = s.orderRepo.GetByIDWithItems(ctx, orderID)
	if err != nil {
		return nil, err
	}

	return s.entityToResponse(order), nil
}

// GetOrderStatistics 获取订单统计信息（管理员）
func (s *OrderService) GetOrderStatistics(ctx context.Context) (map[string]interface{}, error) {
	totalOrders, err := s.orderRepo.CountOrders(ctx)
	if err != nil {
		return nil, err
	}

	totalSales, err := s.orderRepo.GetTotalSales(ctx)
	if err != nil {
		return nil, err
	}

	pendingCount, _ := s.orderRepo.CountOrdersByStatus(ctx, entity.OrderStatusPending)
	paidCount, _ := s.orderRepo.CountOrdersByStatus(ctx, entity.OrderStatusPaid)
	shippedCount, _ := s.orderRepo.CountOrdersByStatus(ctx, entity.OrderStatusShipped)
	completedCount, _ := s.orderRepo.CountOrdersByStatus(ctx, entity.OrderStatusCompleted)
	cancelledCount, _ := s.orderRepo.CountOrdersByStatus(ctx, entity.OrderStatusCancelled)

	stats := map[string]interface{}{
		"total_orders":    totalOrders,
		"total_sales":     totalSales,
		"pending_count":   pendingCount,
		"paid_count":      paidCount,
		"shipped_count":   shippedCount,
		"completed_count": completedCount,
		"cancelled_count": cancelledCount,
	}

	return stats, nil
}

// 辅助方法

func (s *OrderService) generateOrderID() string {
	timestamp := time.Now().UnixNano()
	randomBytes := make([]byte, 8)
	rand.Read(randomBytes)
	
	orderID := hex.EncodeToString([]byte{
		byte(timestamp >> 56),
		byte(timestamp >> 48),
		byte(timestamp >> 40),
		byte(timestamp >> 32),
		byte(timestamp >> 24),
		byte(timestamp >> 16),
		byte(timestamp >> 8),
		byte(timestamp),
	})
	orderID += hex.EncodeToString(randomBytes)
	
	return orderID
}

func (s *OrderService) entityToResponse(order *entity.Order) *dto.OrderResponse {
	var addressSnapshot dto.AddressSnapshot
	json.Unmarshal(order.AddressSnapshot, &addressSnapshot)

	items := make([]dto.OrderItemResponse, len(order.Items))
	for i, item := range order.Items {
		// 转换SpecValues
		var specValues *map[string]string
		if item.SpecValues != nil {
			sv := map[string]string(*item.SpecValues)
			specValues = &sv
		}

		items[i] = dto.OrderItemResponse{
			ID:             item.ID,
			ProductID:      item.ProductID,
			SKUID:          item.SKUID,
			SKUCode:        item.SKUCode,
			SpecValues:     specValues,
			ProductName:    item.ProductName,
			ProductImage:   item.ProductImage,
			Price:          item.Price,
			Quantity:       item.Quantity,
			SubtotalAmount: item.SubtotalAmount,
		}
	}

	return &dto.OrderResponse{
		ID:              order.ID,
		UserID:          order.UserID,
		TotalAmount:     order.TotalAmount,
		CouponID:        order.CouponID,
		CouponDiscount:  order.CouponDiscount,
		FinalAmount:     order.FinalAmount,
		Status:          order.Status,
		StatusText:      dto.GetOrderStatusText(order.Status),
		AddressSnapshot: addressSnapshot,
		ExpressCompany:  order.ExpressCompany,
		ExpressNo:       order.ExpressNo,
		Items:           items,
		CreatedAt:       order.CreatedAt,
		UpdatedAt:       order.UpdatedAt,
		PaidAt:          order.PaidAt,
		ShippedAt:       order.ShippedAt,
		CompletedAt:     order.CompletedAt,
	}
}

func (s *OrderService) getDefaultPagination(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	} else if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

func (s *OrderService) isValidStatusTransition(currentStatus, newStatus int) bool {
	// 定义合法的状态转换
	validTransitions := map[int][]int{
		entity.OrderStatusPending:   {entity.OrderStatusPaid, entity.OrderStatusCancelled},
		entity.OrderStatusPaid:      {entity.OrderStatusShipped, entity.OrderStatusCancelled},
		entity.OrderStatusShipped:   {entity.OrderStatusCompleted},
		entity.OrderStatusCompleted: {},
		entity.OrderStatusCancelled: {},
	}

	allowedStatuses, exists := validTransitions[currentStatus]
	if !exists {
		return false
	}

	for _, status := range allowedStatuses {
		if status == newStatus {
			return true
		}
	}

	return false
}

// orderIDToInt64 将订单ID字符串转换为int64（用于库存服务）
func (s *OrderService) orderIDToInt64(orderID string) int64 {
	// 简单的哈希转换，将字符串转换为int64
	// 在生产环境中可能需要更复杂的转换逻辑
	var hash int64
	for i, c := range orderID {
		hash = hash*31 + int64(c)
		if i >= 8 { // 只使用前8个字符避免溢出
			break
		}
	}
	if hash < 0 {
		hash = -hash
	}
	return hash
}

// rollbackPreDeductedStock 回滚预扣减的库存
func (s *OrderService) rollbackPreDeductedStock(ctx context.Context, items []entity.OrderItem, orderID string) {
	orderIDInt := s.orderIDToInt64(orderID)
	for _, item := range items {
		if item.SKUID != nil {
			err := s.inventoryService.ReleaseStock(ctx, *item.SKUID, orderIDInt)
			if err != nil {
				fmt.Printf("warning: failed to rollback pre-deducted stock for sku %d: %v\n", *item.SKUID, err)
			}
		}
	}
}

// rollbackDeductedStock 回滚已扣减的库存（非SKU商品）
func (s *OrderService) rollbackDeductedStock(ctx context.Context, items []entity.OrderItem) {
	for _, item := range items {
		if item.SKUID == nil {
			err := s.productRepo.IncrementStock(ctx, item.ProductID, item.Quantity)
			if err != nil {
				fmt.Printf("warning: failed to rollback deducted stock for product %d: %v\n", item.ProductID, err)
			}
		}
	}
}

// sendInventoryEvent 发送库存变化事件到RocketMQ
func (s *OrderService) sendInventoryEvent(ctx context.Context, tag string, skuID int64, quantity int, orderID string) {
	if s.mqProducer == nil {
		return
	}

	// 构造消息
	msg := &mq.Message{
		EventType: tag,
		Data: map[string]interface{}{
			"sku_id":   skuID,
			"quantity": quantity,
			"order_id": orderID,
		},
		Timestamp: time.Now().Unix(),
	}

	// 序列化消息
	body, err := mq.MarshalMessage(msg)
	if err != nil {
		fmt.Printf("warning: failed to marshal inventory event: %v\n", err)
		return
	}

	// 异步发送消息
	err = s.mqProducer.SendAsyncMessage(ctx, mq.TopicInventoryEvents, tag, body, func(ctx context.Context, result *primitive.SendResult, err error) {
		if err != nil {
			fmt.Printf("warning: failed to send inventory event: %v\n", err)
		}
	})

	if err != nil {
		fmt.Printf("warning: failed to send async inventory event: %v\n", err)
	}
}
