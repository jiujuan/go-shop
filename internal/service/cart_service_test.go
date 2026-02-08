package service

import (
	"context"
	"testing"
	"time"

	"go-shop/internal/dto"
	"go-shop/internal/entity"
	"go-shop/internal/repository"
	"go-shop/pkg/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockCartRepository 模拟购物车仓库
type MockCartRepository struct {
	mock.Mock
}

func (m *MockCartRepository) GetCart(ctx context.Context, userID int64) (*entity.Cart, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Cart), args.Error(1)
}

func (m *MockCartRepository) AddItem(ctx context.Context, userID int64, item entity.CartItem) error {
	args := m.Called(ctx, userID, item)
	return args.Error(0)
}

func (m *MockCartRepository) UpdateItemQuantity(ctx context.Context, userID int64, productID int64, quantity int) error {
	args := m.Called(ctx, userID, productID, quantity)
	return args.Error(0)
}

func (m *MockCartRepository) RemoveItem(ctx context.Context, userID int64, productID int64) error {
	args := m.Called(ctx, userID, productID)
	return args.Error(0)
}

func (m *MockCartRepository) ClearCart(ctx context.Context, userID int64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockCartRepository) GetItemCount(ctx context.Context, userID int64) (int, error) {
	args := m.Called(ctx, userID)
	return args.Int(0), args.Error(1)
}

func (m *MockCartRepository) GetTotalPrice(ctx context.Context, userID int64) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCartRepository) HasItem(ctx context.Context, userID int64, productID int64) (bool, error) {
	args := m.Called(ctx, userID, productID)
	return args.Bool(0), args.Error(1)
}

func (m *MockCartRepository) GetItem(ctx context.Context, userID int64, productID int64) (*entity.CartItem, error) {
	args := m.Called(ctx, userID, productID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.CartItem), args.Error(1)
}

func (m *MockCartRepository) GetItemQuantity(ctx context.Context, userID int64, productID int64) (int, error) {
	args := m.Called(ctx, userID, productID)
	return args.Int(0), args.Error(1)
}

func (m *MockCartRepository) IsEmpty(ctx context.Context, userID int64) (bool, error) {
	args := m.Called(ctx, userID)
	return args.Bool(0), args.Error(1)
}

func (m *MockCartRepository) MergeCart(ctx context.Context, fromUserID, toUserID int64) error {
	args := m.Called(ctx, fromUserID, toUserID)
	return args.Error(0)
}

func (m *MockCartRepository) BatchAddItems(ctx context.Context, userID int64, items []entity.CartItem) error {
	args := m.Called(ctx, userID, items)
	return args.Error(0)
}

func (m *MockCartRepository) GetCartWithProducts(ctx context.Context, userID int64) (*entity.Cart, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Cart), args.Error(1)
}

func (m *MockCartRepository) ValidateCart(ctx context.Context, userID int64) ([]repository.CartValidationError, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]repository.CartValidationError), args.Error(1)
}

// 其他未实现的方法（为了满足接口）
func (m *MockCartRepository) SetExpiration(ctx context.Context, userID int64, expiration time.Duration) error {
	args := m.Called(ctx, userID, expiration)
	return args.Error(0)
}

func (m *MockCartRepository) GetExpiration(ctx context.Context, userID int64) (time.Duration, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(time.Duration), args.Error(1)
}

func (m *MockCartRepository) RefreshExpiration(ctx context.Context, userID int64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockCartRepository) GetActiveUserCarts(ctx context.Context, limit int) ([]int64, error) {
	args := m.Called(ctx, limit)
	return args.Get(0).([]int64), args.Error(1)
}

func (m *MockCartRepository) DeleteExpiredCarts(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCartRepository) GetCartStatistics(ctx context.Context) (*repository.CartStatistics, error) {
	args := m.Called(ctx)
	return args.Get(0).(*repository.CartStatistics), args.Error(1)
}

// 测试用例

func TestCartService_GetCart(t *testing.T) {
	// 设置测试数据库
	utils.InitTestDB()
	defer utils.CleanupTestData()

	// 创建模拟仓库
	mockCartRepo := new(MockCartRepository)
	mockProductRepo := new(MockProductRepository)

	// 创建服务实例
	mockSKURepo := new(MockSKURepository); service := NewCartService(mockCartRepo, mockProductRepo, mockSKURepo)

	ctx := context.Background()

	t.Run("成功获取购物车", func(t *testing.T) {
		// 准备测试数据
		product := &entity.Product{
			ID:    1,
			Name:  "测试商品",
			Price: 10000,
			Stock: 50,
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

		// 设置模拟期望
		mockCartRepo.On("GetCartWithProducts", ctx, int64(1)).Return(cart, nil)

		// 执行测试
		result, err := service.GetCart(ctx, 1)

		// 验证结果
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, int64(1), result.UserID)
		assert.Equal(t, 1, len(result.Items))
		assert.Equal(t, 2, result.TotalCount)
		assert.Equal(t, int64(20000), result.TotalPrice)

		// 验证模拟调用
		mockCartRepo.AssertExpectations(t)
	})

	t.Run("无效的用户ID", func(t *testing.T) {
		// 执行测试
		result, err := service.GetCart(ctx, 0)

		// 验证结果
		assert.Error(t, err)
		assert.Equal(t, entity.ErrInvalidUserID, err)
		assert.Nil(t, result)
	})
}

func TestCartService_AddItem(t *testing.T) {
	// 设置测试数据库
	utils.InitTestDB()
	defer utils.CleanupTestData()

	// 创建模拟仓库
	mockCartRepo := new(MockCartRepository)
	mockProductRepo := new(MockProductRepository)

	// 创建服务实例
	mockSKURepo := new(MockSKURepository); service := NewCartService(mockCartRepo, mockProductRepo, mockSKURepo)

	ctx := context.Background()

	t.Run("成功添加商品", func(t *testing.T) {
		// 准备测试数据
		req := &dto.CartAddItemRequest{
			ProductID: 1,
			Quantity:  2,
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

		// 设置模拟期望
		mockProductRepo.On("GetByID", ctx, int64(1)).Return(product, nil)
		mockCartRepo.On("AddItem", ctx, int64(1), mock.AnythingOfType("entity.CartItem")).Return(nil)
		mockCartRepo.On("GetCartWithProducts", ctx, int64(1)).Return(cart, nil)

		// 执行测试
		result, err := service.AddItem(ctx, 1, req)

		// 验证结果
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 1, len(result.Items))
		assert.Equal(t, 2, result.TotalCount)

		// 验证模拟调用
		mockProductRepo.AssertExpectations(t)
		mockCartRepo.AssertExpectations(t)
	})

	t.Run("商品不存在", func(t *testing.T) {
		req := &dto.CartAddItemRequest{
			ProductID: 999,
			Quantity:  2,
		}

		// 设置模拟期望
		mockProductRepo.On("GetByID", ctx, int64(999)).Return(nil, entity.ErrProductNotFound)

		// 执行测试
		result, err := service.AddItem(ctx, 1, req)

		// 验证结果
		assert.Error(t, err)
		assert.Equal(t, entity.ErrProductNotFound, err)
		assert.Nil(t, result)

		// 验证模拟调用
		mockProductRepo.AssertExpectations(t)
	})

	t.Run("商品已下架", func(t *testing.T) {
		// 创建新的模拟仓库
		mockCartRepo := new(MockCartRepository)
		mockProductRepo := new(MockProductRepository)
		mockSKURepo := new(MockSKURepository); service := NewCartService(mockCartRepo, mockProductRepo, mockSKURepo)

		req := &dto.CartAddItemRequest{
			ProductID: 1,
			Quantity:  2,
		}

		product := &entity.Product{
			ID:     1,
			Name:   "测试商品",
			Price:  10000,
			Stock:  50,
			Status: 0, // 已下架
		}

		// 设置模拟期望
		mockProductRepo.On("GetByID", ctx, int64(1)).Return(product, nil)

		// 执行测试
		result, err := service.AddItem(ctx, 1, req)

		// 验证结果
		assert.Error(t, err)
		assert.Equal(t, entity.ErrProductNotAvailable, err)
		assert.Nil(t, result)

		// 验证模拟调用
		mockProductRepo.AssertExpectations(t)
	})

	t.Run("库存不足", func(t *testing.T) {
		// 创建新的模拟仓库
		mockCartRepo := new(MockCartRepository)
		mockProductRepo := new(MockProductRepository)
		mockSKURepo := new(MockSKURepository); service := NewCartService(mockCartRepo, mockProductRepo, mockSKURepo)

		req := &dto.CartAddItemRequest{
			ProductID: 1,
			Quantity:  100,
		}

		product := &entity.Product{
			ID:     1,
			Name:   "测试商品",
			Price:  10000,
			Stock:  50, // 库存不足
			Status: 1,
		}

		// 设置模拟期望
		mockProductRepo.On("GetByID", ctx, int64(1)).Return(product, nil)

		// 执行测试
		result, err := service.AddItem(ctx, 1, req)

		// 验证结果
		assert.Error(t, err)
		assert.Equal(t, entity.ErrProductOutOfStock, err)
		assert.Nil(t, result)

		// 验证模拟调用
		mockProductRepo.AssertExpectations(t)
	})

	t.Run("无效参数", func(t *testing.T) {
		// 测试空请求
		result, err := service.AddItem(ctx, 1, nil)
		assert.Error(t, err)
		assert.Nil(t, result)

		// 测试无效用户ID
		req := &dto.CartAddItemRequest{
			ProductID: 1,
			Quantity:  2,
		}
		result, err = service.AddItem(ctx, 0, req)
		assert.Error(t, err)
		assert.Equal(t, entity.ErrInvalidUserID, err)
		assert.Nil(t, result)

		// 测试无效商品ID
		req = &dto.CartAddItemRequest{
			ProductID: 0,
			Quantity:  2,
		}
		result, err = service.AddItem(ctx, 1, req)
		assert.Error(t, err)
		assert.Equal(t, entity.ErrInvalidProductID, err)
		assert.Nil(t, result)

		// 测试无效数量
		req = &dto.CartAddItemRequest{
			ProductID: 1,
			Quantity:  0,
		}
		result, err = service.AddItem(ctx, 1, req)
		assert.Error(t, err)
		assert.Nil(t, result)

		// 测试数量超过限制
		req = &dto.CartAddItemRequest{
			ProductID: 1,
			Quantity:  1000,
		}
		result, err = service.AddItem(ctx, 1, req)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestCartService_UpdateItemQuantity(t *testing.T) {
	// 设置测试数据库
	utils.InitTestDB()
	defer utils.CleanupTestData()

	// 创建模拟仓库
	mockCartRepo := new(MockCartRepository)
	mockProductRepo := new(MockProductRepository)

	// 创建服务实例
	mockSKURepo := new(MockSKURepository); service := NewCartService(mockCartRepo, mockProductRepo, mockSKURepo)

	ctx := context.Background()

	t.Run("成功更新数量", func(t *testing.T) {
		req := &dto.CartUpdateItemRequest{
			ProductID: 1,
			Quantity:  5,
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
					Quantity:  5,
					Product:   product,
				},
			},
			TotalCount: 5,
			TotalPrice: 50000,
		}

		// 设置模拟期望
		mockCartRepo.On("HasItem", ctx, int64(1), int64(1)).Return(true, nil)
		mockProductRepo.On("GetByID", ctx, int64(1)).Return(product, nil)
		mockCartRepo.On("UpdateItemQuantity", ctx, int64(1), int64(1), 5).Return(nil)
		mockCartRepo.On("GetCartWithProducts", ctx, int64(1)).Return(cart, nil)

		// 执行测试
		result, err := service.UpdateItemQuantity(ctx, 1, req)

		// 验证结果
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 5, result.TotalCount)

		// 验证模拟调用
		mockCartRepo.AssertExpectations(t)
		mockProductRepo.AssertExpectations(t)
	})

	t.Run("商品不在购物车中", func(t *testing.T) {
		req := &dto.CartUpdateItemRequest{
			ProductID: 999,
			Quantity:  5,
		}

		// 设置模拟期望
		mockCartRepo.On("HasItem", ctx, int64(1), int64(999)).Return(false, nil)

		// 执行测试
		result, err := service.UpdateItemQuantity(ctx, 1, req)

		// 验证结果
		assert.Error(t, err)
		assert.Equal(t, entity.ErrCartItemNotFound, err)
		assert.Nil(t, result)

		// 验证模拟调用
		mockCartRepo.AssertExpectations(t)
	})

	t.Run("库存不足", func(t *testing.T) {
		req := &dto.CartUpdateItemRequest{
			ProductID: 1,
			Quantity:  100,
		}

		product := &entity.Product{
			ID:     1,
			Name:   "测试商品",
			Price:  10000,
			Stock:  50,
			Status: 1,
		}

		// 设置模拟期望
		mockCartRepo.On("HasItem", ctx, int64(1), int64(1)).Return(true, nil)
		mockProductRepo.On("GetByID", ctx, int64(1)).Return(product, nil)

		// 执行测试
		result, err := service.UpdateItemQuantity(ctx, 1, req)

		// 验证结果
		assert.Error(t, err)
		assert.Equal(t, entity.ErrProductOutOfStock, err)
		assert.Nil(t, result)

		// 验证模拟调用
		mockCartRepo.AssertExpectations(t)
		mockProductRepo.AssertExpectations(t)
	})
}

func TestCartService_RemoveItem(t *testing.T) {
	// 设置测试数据库
	utils.InitTestDB()
	defer utils.CleanupTestData()

	// 创建模拟仓库
	mockCartRepo := new(MockCartRepository)
	mockProductRepo := new(MockProductRepository)

	// 创建服务实例
	mockSKURepo := new(MockSKURepository); service := NewCartService(mockCartRepo, mockProductRepo, mockSKURepo)

	ctx := context.Background()

	t.Run("成功删除商品", func(t *testing.T) {
		cart := &entity.Cart{
			UserID:     1,
			Items:      []entity.CartItem{},
			TotalCount: 0,
			TotalPrice: 0,
		}

		// 设置模拟期望
		mockCartRepo.On("HasItem", ctx, int64(1), int64(1)).Return(true, nil)
		mockCartRepo.On("RemoveItem", ctx, int64(1), int64(1)).Return(nil)
		mockCartRepo.On("GetCartWithProducts", ctx, int64(1)).Return(cart, nil)

		// 执行测试
		result, err := service.RemoveItem(ctx, 1, 1)

		// 验证结果
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 0, len(result.Items))

		// 验证模拟调用
		mockCartRepo.AssertExpectations(t)
	})

	t.Run("商品不在购物车中", func(t *testing.T) {
		// 设置模拟期望
		mockCartRepo.On("HasItem", ctx, int64(1), int64(999)).Return(false, nil)

		// 执行测试
		result, err := service.RemoveItem(ctx, 1, 999)

		// 验证结果
		assert.Error(t, err)
		assert.Equal(t, entity.ErrCartItemNotFound, err)
		assert.Nil(t, result)

		// 验证模拟调用
		mockCartRepo.AssertExpectations(t)
	})

	t.Run("无效参数", func(t *testing.T) {
		// 无效用户ID
		result, err := service.RemoveItem(ctx, 0, 1)
		assert.Error(t, err)
		assert.Equal(t, entity.ErrInvalidUserID, err)
		assert.Nil(t, result)

		// 无效商品ID
		result, err = service.RemoveItem(ctx, 1, 0)
		assert.Error(t, err)
		assert.Equal(t, entity.ErrInvalidProductID, err)
		assert.Nil(t, result)
	})
}

func TestCartService_ClearCart(t *testing.T) {
	// 设置测试数据库
	utils.InitTestDB()
	defer utils.CleanupTestData()

	// 创建模拟仓库
	mockCartRepo := new(MockCartRepository)
	mockProductRepo := new(MockProductRepository)

	// 创建服务实例
	mockSKURepo := new(MockSKURepository); service := NewCartService(mockCartRepo, mockProductRepo, mockSKURepo)

	ctx := context.Background()

	t.Run("成功清空购物车", func(t *testing.T) {
		// 设置模拟期望
		mockCartRepo.On("ClearCart", ctx, int64(1)).Return(nil)

		// 执行测试
		err := service.ClearCart(ctx, 1)

		// 验证结果
		assert.NoError(t, err)

		// 验证模拟调用
		mockCartRepo.AssertExpectations(t)
	})

	t.Run("无效的用户ID", func(t *testing.T) {
		// 执行测试
		err := service.ClearCart(ctx, 0)

		// 验证结果
		assert.Error(t, err)
		assert.Equal(t, entity.ErrInvalidUserID, err)
	})
}

func TestCartService_ValidateCart(t *testing.T) {
	// 设置测试数据库
	utils.InitTestDB()
	defer utils.CleanupTestData()

	// 创建模拟仓库
	mockCartRepo := new(MockCartRepository)
	mockProductRepo := new(MockProductRepository)

	// 创建服务实例
	mockSKURepo := new(MockSKURepository); service := NewCartService(mockCartRepo, mockProductRepo, mockSKURepo)

	ctx := context.Background()

	t.Run("购物车验证通过", func(t *testing.T) {
		// 设置模拟期望
		mockCartRepo.On("IsEmpty", ctx, int64(1)).Return(false, nil)
		mockCartRepo.On("ValidateCart", ctx, int64(1)).Return([]repository.CartValidationError{}, nil)

		// 执行测试
		errors, err := service.ValidateCart(ctx, 1)

		// 验证结果
		assert.NoError(t, err)
		assert.NotNil(t, errors)
		assert.Equal(t, 0, len(errors))

		// 验证模拟调用
		mockCartRepo.AssertExpectations(t)
	})

	t.Run("购物车为空", func(t *testing.T) {
		// 创建新的模拟仓库
		mockCartRepo := new(MockCartRepository)
		mockProductRepo := new(MockProductRepository)
		mockSKURepo := new(MockSKURepository); service := NewCartService(mockCartRepo, mockProductRepo, mockSKURepo)

		// 设置模拟期望
		mockCartRepo.On("IsEmpty", ctx, int64(1)).Return(true, nil)

		// 执行测试
		errors, err := service.ValidateCart(ctx, 1)

		// 验证结果
		assert.Error(t, err)
		assert.Equal(t, entity.ErrCartEmpty, err)
		assert.Nil(t, errors)

		// 验证模拟调用
		mockCartRepo.AssertExpectations(t)
	})

	t.Run("购物车有验证错误", func(t *testing.T) {
		// 创建新的模拟仓库
		mockCartRepo := new(MockCartRepository)
		mockProductRepo := new(MockProductRepository)
		mockSKURepo := new(MockSKURepository); service := NewCartService(mockCartRepo, mockProductRepo, mockSKURepo)

		validationErrors := []repository.CartValidationError{
			{
				ProductID: 1,
				Error:     "库存不足",
				ErrorType: "out_of_stock",
			},
		}

		// 设置模拟期望
		mockCartRepo.On("IsEmpty", ctx, int64(1)).Return(false, nil)
		mockCartRepo.On("ValidateCart", ctx, int64(1)).Return(validationErrors, nil)

		// 执行测试
		errors, err := service.ValidateCart(ctx, 1)

		// 验证结果
		assert.NoError(t, err)
		assert.NotNil(t, errors)
		assert.Equal(t, 1, len(errors))
		assert.Equal(t, "out_of_stock", errors[0].ErrorType)

		// 验证模拟调用
		mockCartRepo.AssertExpectations(t)
	})
}

