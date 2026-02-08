package service

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"go-shop/internal/dto"
	"go-shop/internal/entity"
	"go-shop/internal/repository"
	"go-shop/pkg/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockOrderRepository struct {
	mock.Mock
}

func (m *MockOrderRepository) Create(ctx context.Context, order *entity.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockOrderRepository) GetByID(ctx context.Context, id string) (*entity.Order, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Order), args.Error(1)
}

func (m *MockOrderRepository) GetByIDWithItems(ctx context.Context, id string) (*entity.Order, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Order), args.Error(1)
}

func (m *MockOrderRepository) UpdateStatus(ctx context.Context, orderID string, status int) error {
	args := m.Called(ctx, orderID, status)
	return args.Error(0)
}

func (m *MockOrderRepository) UpdatePaymentInfo(ctx context.Context, orderID string, paidAt *time.Time) error {
	args := m.Called(ctx, orderID, paidAt)
	return args.Error(0)
}

func (m *MockOrderRepository) UpdateShippingInfo(ctx context.Context, orderID string, expressCompany, expressNo string, shippedAt *time.Time) error {
	args := m.Called(ctx, orderID, expressCompany, expressNo, shippedAt)
	return args.Error(0)
}

func (m *MockOrderRepository) UpdateCompletionInfo(ctx context.Context, orderID string, completedAt *time.Time) error {
	args := m.Called(ctx, orderID, completedAt)
	return args.Error(0)
}

func (m *MockOrderRepository) ListByUser(ctx context.Context, userID int64, offset, limit int) ([]*entity.Order, int64, error) {
	args := m.Called(ctx, userID, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entity.Order), args.Get(1).(int64), args.Error(2)
}

func (m *MockOrderRepository) ListByUserAndStatus(ctx context.Context, userID int64, status int, offset, limit int) ([]*entity.Order, int64, error) {
	args := m.Called(ctx, userID, status, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entity.Order), args.Get(1).(int64), args.Error(2)
}

func (m *MockOrderRepository) List(ctx context.Context, options repository.OrderQueryOptions) ([]*entity.Order, int64, error) {
	args := m.Called(ctx, options)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entity.Order), args.Get(1).(int64), args.Error(2)
}

func (m *MockOrderRepository) CountOrders(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockOrderRepository) CountOrdersByStatus(ctx context.Context, status int) (int64, error) {
	args := m.Called(ctx, status)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockOrderRepository) GetTotalSales(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

// 其他未实现的方法
func (m *MockOrderRepository) Update(ctx context.Context, order *entity.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockOrderRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockOrderRepository) ListByStatus(ctx context.Context, status int, offset, limit int) ([]*entity.Order, int64, error) {
	args := m.Called(ctx, status, offset, limit)
	return args.Get(0).([]*entity.Order), args.Get(1).(int64), args.Error(2)
}

func (m *MockOrderRepository) GetUserOrderCount(ctx context.Context, userID int64) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockOrderRepository) GetUserOrderCountByStatus(ctx context.Context, userID int64, status int) (int64, error) {
	args := m.Called(ctx, userID, status)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockOrderRepository) CountOrdersByDateRange(ctx context.Context, startDate, endDate string) (int64, error) {
	args := m.Called(ctx, startDate, endDate)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockOrderRepository) GetTotalSalesByDateRange(ctx context.Context, startDate, endDate string) (int64, error) {
	args := m.Called(ctx, startDate, endDate)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockOrderRepository) GetTotalSalesByUser(ctx context.Context, userID int64) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockOrderRepository) ExistsByID(ctx context.Context, id string) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

func (m *MockOrderRepository) CanCancel(ctx context.Context, orderID string) (bool, error) {
	args := m.Called(ctx, orderID)
	return args.Bool(0), args.Error(1)
}

func (m *MockOrderRepository) CanPay(ctx context.Context, orderID string) (bool, error) {
	args := m.Called(ctx, orderID)
	return args.Bool(0), args.Error(1)
}

func (m *MockOrderRepository) CanShip(ctx context.Context, orderID string) (bool, error) {
	args := m.Called(ctx, orderID)
	return args.Bool(0), args.Error(1)
}

func (m *MockOrderRepository) CanComplete(ctx context.Context, orderID string) (bool, error) {
	args := m.Called(ctx, orderID)
	return args.Bool(0), args.Error(1)
}

func (m *MockOrderRepository) GetPendingOrders(ctx context.Context, timeoutMinutes int) ([]*entity.Order, error) {
	args := m.Called(ctx, timeoutMinutes)
	return args.Get(0).([]*entity.Order), args.Error(1)
}

func (m *MockOrderRepository) GetRecentOrders(ctx context.Context, limit int) ([]*entity.Order, error) {
	args := m.Called(ctx, limit)
	return args.Get(0).([]*entity.Order), args.Error(1)
}

func (m *MockOrderRepository) GetTopSellingProducts(ctx context.Context, limit int) ([]repository.ProductSalesStats, error) {
	args := m.Called(ctx, limit)
	return args.Get(0).([]repository.ProductSalesStats), args.Error(1)
}

type MockAddressRepository struct {
	mock.Mock
}

func (m *MockAddressRepository) GetByID(ctx context.Context, id int64) (*entity.Address, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Address), args.Error(1)
}

// 其他未实现的方法
func (m *MockAddressRepository) Create(ctx context.Context, address *entity.Address) error {
	return nil
}

func (m *MockAddressRepository) Update(ctx context.Context, address *entity.Address) error {
	return nil
}

func (m *MockAddressRepository) Delete(ctx context.Context, id int64) error {
	return nil
}

func (m *MockAddressRepository) GetByUserID(ctx context.Context, userID int64) ([]*entity.Address, error) {
	return nil, nil
}

func (m *MockAddressRepository) GetDefaultByUserID(ctx context.Context, userID int64) (*entity.Address, error) {
	return nil, nil
}

func (m *MockAddressRepository) SetDefault(ctx context.Context, userID int64, addressID int64) error {
	return nil
}

func (m *MockAddressRepository) UnsetDefault(ctx context.Context, userID int64, addressID int64) error {
	return nil
}

func (m *MockAddressRepository) CountByUserID(ctx context.Context, userID int64) (int64, error) {
	return 0, nil
}

func (m *MockAddressRepository) ExistsByID(ctx context.Context, id int64) (bool, error) {
	return false, nil
}

func (m *MockAddressRepository) ExistsByUserIDAndID(ctx context.Context, userID int64, addressID int64) (bool, error) {
	return false, nil
}

func (m *MockAddressRepository) GetUserAddressWithPagination(ctx context.Context, userID int64, offset, limit int) ([]*entity.Address, int64, error) {
	return nil, 0, nil
}

func (m *MockAddressRepository) DeleteByUserID(ctx context.Context, userID int64) error {
	return nil
}

func (m *MockAddressRepository) HasDefaultAddress(ctx context.Context, userID int64) (bool, error) {
	return false, nil
}

func (m *MockAddressRepository) GetAddressesInRegion(ctx context.Context, province, city string, limit int) ([]*entity.Address, error) {
	return nil, nil
}

func (m *MockAddressRepository) GetAddressStatistics(ctx context.Context) (*repository.AddressStatistics, error) {
	return nil, nil
}

// MockSKURepository SKU仓库模拟
type MockSKURepository struct {
	mock.Mock
}

func (m *MockSKURepository) Create(ctx context.Context, sku *entity.ProductSKU) error {
	args := m.Called(ctx, sku)
	return args.Error(0)
}

func (m *MockSKURepository) CreateSKU(ctx context.Context, sku *entity.ProductSKU) error {
	args := m.Called(ctx, sku)
	return args.Error(0)
}

func (m *MockSKURepository) GetByID(ctx context.Context, id int64) (*entity.ProductSKU, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.ProductSKU), args.Error(1)
}

func (m *MockSKURepository) GetSKUByID(ctx context.Context, skuID int64) (*entity.ProductSKU, error) {
	args := m.Called(ctx, skuID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.ProductSKU), args.Error(1)
}

func (m *MockSKURepository) GetByProductID(ctx context.Context, productID int64) ([]*entity.ProductSKU, error) {
	args := m.Called(ctx, productID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.ProductSKU), args.Error(1)
}

func (m *MockSKURepository) GetBySKUCode(ctx context.Context, skuCode string) (*entity.ProductSKU, error) {
	args := m.Called(ctx, skuCode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.ProductSKU), args.Error(1)
}

func (m *MockSKURepository) Update(ctx context.Context, sku *entity.ProductSKU) error {
	args := m.Called(ctx, sku)
	return args.Error(0)
}

func (m *MockSKURepository) UpdateStock(ctx context.Context, skuID int64, quantity int) error {
	args := m.Called(ctx, skuID, quantity)
	return args.Error(0)
}

func (m *MockSKURepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockSKURepository) CreateSpec(ctx context.Context, spec *entity.ProductSpec) error {
	args := m.Called(ctx, spec)
	return args.Error(0)
}

func (m *MockSKURepository) GetSpecsByProductID(ctx context.Context, productID int64) ([]*entity.ProductSpec, error) {
	args := m.Called(ctx, productID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.ProductSpec), args.Error(1)
}

// MockInventoryService 库存服务模拟
type MockInventoryService struct {
	mock.Mock
}

func (m *MockInventoryService) PreDeductStock(ctx context.Context, skuID int64, quantity int, orderID int64) error {
	args := m.Called(ctx, skuID, quantity, orderID)
	return args.Error(0)
}

func (m *MockInventoryService) ConfirmDeduct(ctx context.Context, skuID int64, orderID int64) error {
	args := m.Called(ctx, skuID, orderID)
	return args.Error(0)
}

func (m *MockInventoryService) ReleaseStock(ctx context.Context, skuID int64, orderID int64) error {
	args := m.Called(ctx, skuID, orderID)
	return args.Error(0)
}

func (m *MockInventoryService) SyncStockToRedis(ctx context.Context, skuID int64) error {
	args := m.Called(ctx, skuID)
	return args.Error(0)
}

func (m *MockInventoryService) CheckStock(ctx context.Context, skuID int64, quantity int) (bool, error) {
	args := m.Called(ctx, skuID, quantity)
	return args.Bool(0), args.Error(1)
}

func TestOrderService_CreateOrder(t *testing.T) {
	utils.InitTestDB()
	defer utils.CleanupTestData()

	mockOrderRepo := new(MockOrderRepository)
	mockCartRepo := new(MockCartRepository)
	mockProductRepo := new(MockProductRepository)
	mockAddressRepo := new(MockAddressRepository)
	mockSKURepo := new(MockSKURepository)
	mockInventoryService := new(MockInventoryService)
	
	// Create a nil CouponService for tests that don't use coupons
	var mockCouponService *CouponService = nil

	service := NewOrderService(mockOrderRepo, mockCartRepo, mockProductRepo, mockAddressRepo, mockSKURepo, mockCouponService, mockInventoryService, nil)

	ctx := context.Background()

	t.Run("成功创建订单", func(t *testing.T) {
		req := &dto.OrderCreateRequest{
			AddressID: 1,
		}

		product := &entity.Product{
			ID:     1,
			Name:   "测试商品",
			Price:  10000,
			Stock:  50,
			Status: 1,
		}

		cart := &entity.Cart{
			UserID: 1,
			Items: []entity.CartItem{
				{
					UserID:    1,
					ProductID: 1,
					Quantity:  2,
					Product:   product,
				},
			},
			TotalCount: 2,
			TotalPrice: 20000,
		}

		address := &entity.Address{
			ID:            1,
			UserID:        1,
			RecipientName: "张三",
			Phone:         "13800138000",
			Province:      "北京市",
			City:          "北京市",
			District:      "朝阳区",
			Detail:        "某某街道123号",
		}

		mockCartRepo.On("GetCartWithProducts", ctx, int64(1)).Return(cart, nil)
		mockCartRepo.On("ValidateCart", ctx, int64(1)).Return([]repository.CartValidationError{}, nil)
		mockAddressRepo.On("GetByID", ctx, int64(1)).Return(address, nil)
		mockProductRepo.On("DecrementStock", ctx, int64(1), 2).Return(nil)
		mockOrderRepo.On("Create", ctx, mock.AnythingOfType("*entity.Order")).Return(nil)
		mockCartRepo.On("ClearCart", ctx, int64(1)).Return(nil)

		result, err := service.CreateOrder(ctx, 1, req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, int64(1), result.UserID)
		assert.Equal(t, int64(20000), result.TotalAmount)
		assert.Equal(t, entity.OrderStatusPending, result.Status)

		mockCartRepo.AssertExpectations(t)
		mockAddressRepo.AssertExpectations(t)
		mockProductRepo.AssertExpectations(t)
		mockOrderRepo.AssertExpectations(t)
	})

	t.Run("购物车为空", func(t *testing.T) {
		mockCartRepo := new(MockCartRepository)
		mockOrderRepo := new(MockOrderRepository)
		mockProductRepo := new(MockProductRepository)
		mockAddressRepo := new(MockAddressRepository)
		mockSKURepo := new(MockSKURepository)
		mockInventoryService := new(MockInventoryService)
		service := NewOrderService(mockOrderRepo, mockCartRepo, mockProductRepo, mockAddressRepo, mockSKURepo, nil, mockInventoryService, nil)

		req := &dto.OrderCreateRequest{
			AddressID: 1,
		}

		emptyCart := &entity.Cart{
			UserID:     1,
			Items:      []entity.CartItem{},
			TotalCount: 0,
			TotalPrice: 0,
		}

		mockCartRepo.On("GetCartWithProducts", ctx, int64(1)).Return(emptyCart, nil)

		result, err := service.CreateOrder(ctx, 1, req)

		assert.Error(t, err)
		assert.Equal(t, entity.ErrCartEmpty, err)
		assert.Nil(t, result)

		mockCartRepo.AssertExpectations(t)
	})
}

func TestOrderService_PayOrder(t *testing.T) {
	utils.InitTestDB()
	defer utils.CleanupTestData()

	ctx := context.Background()

	t.Run("成功支付订单", func(t *testing.T) {
		mockOrderRepo := new(MockOrderRepository)
		mockCartRepo := new(MockCartRepository)
		mockProductRepo := new(MockProductRepository)
		mockAddressRepo := new(MockAddressRepository)
		mockSKURepo := new(MockSKURepository)
		mockInventoryService := new(MockInventoryService)
		service := NewOrderService(mockOrderRepo, mockCartRepo, mockProductRepo, mockAddressRepo, mockSKURepo, nil, mockInventoryService, nil)

		addressSnapshot, _ := json.Marshal(dto.AddressSnapshot{
			RecipientName: "张三",
			Phone:         "13800138000",
			Province:      "北京市",
			City:          "北京市",
			District:      "朝阳区",
			Detail:        "某某街道123号",
		})

		order := &entity.Order{
			ID:              "test-order-id",
			UserID:          1,
			TotalAmount:     20000,
			Status:          entity.OrderStatusPending,
			AddressSnapshot: addressSnapshot,
			Items:           []entity.OrderItem{},
		}

		orderWithItems := &entity.Order{
			ID:              "test-order-id",
			UserID:          1,
			TotalAmount:     20000,
			Status:          entity.OrderStatusPaid,
			AddressSnapshot: addressSnapshot,
			Items:           []entity.OrderItem{},
		}

		mockOrderRepo.On("GetByIDWithItems", ctx, "test-order-id").Return(order, nil).Once()
		mockOrderRepo.On("UpdateStatus", ctx, "test-order-id", entity.OrderStatusPaid).Return(nil)
		mockOrderRepo.On("UpdatePaymentInfo", ctx, "test-order-id", mock.AnythingOfType("*time.Time")).Return(nil)
		mockOrderRepo.On("GetByIDWithItems", ctx, "test-order-id").Return(orderWithItems, nil).Once()

		result, err := service.PayOrder(ctx, 1, "test-order-id")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, entity.OrderStatusPaid, result.Status)

		mockOrderRepo.AssertExpectations(t)
	})

	t.Run("订单状态不允许支付", func(t *testing.T) {
		mockOrderRepo := new(MockOrderRepository)
		mockCartRepo := new(MockCartRepository)
		mockProductRepo := new(MockProductRepository)
		mockAddressRepo := new(MockAddressRepository)
		mockSKURepo := new(MockSKURepository)
		mockInventoryService := new(MockInventoryService)
		service := NewOrderService(mockOrderRepo, mockCartRepo, mockProductRepo, mockAddressRepo, mockSKURepo, nil, mockInventoryService, nil)

		addressSnapshot, _ := json.Marshal(dto.AddressSnapshot{})

		order := &entity.Order{
			ID:              "test-order-id",
			UserID:          1,
			TotalAmount:     20000,
			Status:          entity.OrderStatusPaid,
			AddressSnapshot: addressSnapshot,
			Items:           []entity.OrderItem{},
		}

		mockOrderRepo.On("GetByIDWithItems", ctx, "test-order-id").Return(order, nil)

		result, err := service.PayOrder(ctx, 1, "test-order-id")

		assert.Error(t, err)
		assert.Nil(t, result)

		mockOrderRepo.AssertExpectations(t)
	})
}

func TestOrderService_CancelOrder(t *testing.T) {
	utils.InitTestDB()
	defer utils.CleanupTestData()

	ctx := context.Background()

	t.Run("成功取消订单", func(t *testing.T) {
		mockOrderRepo := new(MockOrderRepository)
		mockCartRepo := new(MockCartRepository)
		mockProductRepo := new(MockProductRepository)
		mockAddressRepo := new(MockAddressRepository)
		mockSKURepo := new(MockSKURepository)
		mockInventoryService := new(MockInventoryService)
		service := NewOrderService(mockOrderRepo, mockCartRepo, mockProductRepo, mockAddressRepo, mockSKURepo, nil, mockInventoryService, nil)

		addressSnapshot, _ := json.Marshal(dto.AddressSnapshot{})

		order := &entity.Order{
			ID:              "test-order-id",
			UserID:          1,
			TotalAmount:     20000,
			Status:          entity.OrderStatusPending,
			AddressSnapshot: addressSnapshot,
			Items: []entity.OrderItem{
				{
					ProductID: 1,
					Quantity:  2,
				},
			},
		}

		mockOrderRepo.On("GetByIDWithItems", ctx, "test-order-id").Return(order, nil)
		mockOrderRepo.On("UpdateStatus", ctx, "test-order-id", entity.OrderStatusCancelled).Return(nil)
		mockProductRepo.On("IncrementStock", ctx, int64(1), 2).Return(nil)

		err := service.CancelOrder(ctx, 1, "test-order-id")

		assert.NoError(t, err)

		mockOrderRepo.AssertExpectations(t)
		mockProductRepo.AssertExpectations(t)
	})

	t.Run("订单状态不允许取消", func(t *testing.T) {
		mockOrderRepo := new(MockOrderRepository)
		mockCartRepo := new(MockCartRepository)
		mockProductRepo := new(MockProductRepository)
		mockAddressRepo := new(MockAddressRepository)
		mockSKURepo := new(MockSKURepository)
		mockInventoryService := new(MockInventoryService)
		service := NewOrderService(mockOrderRepo, mockCartRepo, mockProductRepo, mockAddressRepo, mockSKURepo, nil, mockInventoryService, nil)

		addressSnapshot, _ := json.Marshal(dto.AddressSnapshot{})

		order := &entity.Order{
			ID:              "test-order-id",
			UserID:          1,
			TotalAmount:     20000,
			Status:          entity.OrderStatusPaid,
			AddressSnapshot: addressSnapshot,
			Items:           []entity.OrderItem{},
		}

		mockOrderRepo.On("GetByIDWithItems", ctx, "test-order-id").Return(order, nil)

		err := service.CancelOrder(ctx, 1, "test-order-id")

		assert.Error(t, err)
		assert.Equal(t, entity.ErrOrderCannotCancel, err)

		mockOrderRepo.AssertExpectations(t)
	})
}

func TestOrderService_GetUserOrders(t *testing.T) {
	utils.InitTestDB()
	defer utils.CleanupTestData()

	mockOrderRepo := new(MockOrderRepository)
	mockCartRepo := new(MockCartRepository)
	mockProductRepo := new(MockProductRepository)
	mockAddressRepo := new(MockAddressRepository)
	mockSKURepo := new(MockSKURepository)
	mockInventoryService := new(MockInventoryService)
	service := NewOrderService(mockOrderRepo, mockCartRepo, mockProductRepo, mockAddressRepo, mockSKURepo, nil, mockInventoryService, nil)

	ctx := context.Background()

	t.Run("成功获取订单列表", func(t *testing.T) {
		req := &dto.OrderListRequest{
			Page:     1,
			PageSize: 20,
		}

		addressSnapshot, _ := json.Marshal(dto.AddressSnapshot{})

		orders := []*entity.Order{
			{
				ID:              "order-1",
				UserID:          1,
				TotalAmount:     20000,
				Status:          entity.OrderStatusPending,
				AddressSnapshot: addressSnapshot,
			},
		}

		mockOrderRepo.On("ListByUser", ctx, int64(1), 0, 20).Return(orders, int64(1), nil)

		result, err := service.GetUserOrders(ctx, 1, req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 1, len(result.Orders))

		mockOrderRepo.AssertExpectations(t)
	})
}
