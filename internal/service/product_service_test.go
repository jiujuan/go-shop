package service

import (
	"context"
	"testing"

	"go-shop/internal/dto"
	"go-shop/internal/entity"
	"go-shop/internal/repository"
	"go-shop/pkg/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockProductRepository 模拟商品仓库
type MockProductRepository struct {
	mock.Mock
}

func (m *MockProductRepository) Create(ctx context.Context, product *entity.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *MockProductRepository) GetByID(ctx context.Context, id int64) (*entity.Product, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Product), args.Error(1)
}

func (m *MockProductRepository) Update(ctx context.Context, product *entity.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *MockProductRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockProductRepository) List(ctx context.Context, options repository.ProductQueryOptions) ([]*entity.Product, int64, error) {
	args := m.Called(ctx, options)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entity.Product), args.Get(1).(int64), args.Error(2)
}

func (m *MockProductRepository) ListByCategory(ctx context.Context, categoryID int64, offset, limit int) ([]*entity.Product, int64, error) {
	args := m.Called(ctx, categoryID, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entity.Product), args.Get(1).(int64), args.Error(2)
}

func (m *MockProductRepository) Search(ctx context.Context, keyword string, offset, limit int) ([]*entity.Product, int64, error) {
	args := m.Called(ctx, keyword, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entity.Product), args.Get(1).(int64), args.Error(2)
}

func (m *MockProductRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	args := m.Called(ctx, name)
	return args.Bool(0), args.Error(1)
}

func (m *MockProductRepository) ExistsByNameExcludeID(ctx context.Context, name string, excludeID int64) (bool, error) {
	args := m.Called(ctx, name, excludeID)
	return args.Bool(0), args.Error(1)
}

func (m *MockProductRepository) CheckStock(ctx context.Context, productID int64, quantity int) (bool, error) {
	args := m.Called(ctx, productID, quantity)
	return args.Bool(0), args.Error(1)
}

func (m *MockProductRepository) UpdateStock(ctx context.Context, productID int64, stock int) error {
	args := m.Called(ctx, productID, stock)
	return args.Error(0)
}

func (m *MockProductRepository) UpdateStatus(ctx context.Context, productID int64, status int) error {
	args := m.Called(ctx, productID, status)
	return args.Error(0)
}

func (m *MockProductRepository) DecrementStock(ctx context.Context, productID int64, quantity int) error {
	args := m.Called(ctx, productID, quantity)
	return args.Error(0)
}

func (m *MockProductRepository) IncrementStock(ctx context.Context, productID int64, quantity int) error {
	args := m.Called(ctx, productID, quantity)
	return args.Error(0)
}

func (m *MockProductRepository) GetTopProducts(ctx context.Context, limit int) ([]*entity.Product, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Product), args.Error(1)
}

func (m *MockProductRepository) GetLatestProducts(ctx context.Context, limit int) ([]*entity.Product, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Product), args.Error(1)
}

func (m *MockProductRepository) GetLowStockProducts(ctx context.Context, threshold int) ([]*entity.Product, error) {
	args := m.Called(ctx, threshold)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Product), args.Error(1)
}

// 其他未实现的方法（为了满足接口）
func (m *MockProductRepository) GetByIDWithCategory(ctx context.Context, id int64) (*entity.Product, *entity.Category, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*entity.Product), args.Get(1).(*entity.Category), args.Error(2)
}

func (m *MockProductRepository) ListByStatus(ctx context.Context, status int, offset, limit int) ([]*entity.Product, int64, error) {
	args := m.Called(ctx, status, offset, limit)
	return args.Get(0).([]*entity.Product), args.Get(1).(int64), args.Error(2)
}

func (m *MockProductRepository) BatchUpdateStock(ctx context.Context, updates []repository.StockUpdate) error {
	args := m.Called(ctx, updates)
	return args.Error(0)
}

func (m *MockProductRepository) CountProducts(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockProductRepository) CountProductsByCategory(ctx context.Context, categoryID int64) (int64, error) {
	args := m.Called(ctx, categoryID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockProductRepository) CountProductsByStatus(ctx context.Context, status int) (int64, error) {
	args := m.Called(ctx, status)
	return args.Get(0).(int64), args.Error(1)
}

// MockCategoryRepository 模拟分类仓库
type MockCategoryRepository struct {
	mock.Mock
}

func (m *MockCategoryRepository) GetByID(ctx context.Context, id int64) (*entity.Category, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Category), args.Error(1)
}

// 其他未实现的方法（为了满足接口）
func (m *MockCategoryRepository) Create(ctx context.Context, category *entity.Category) error {
	args := m.Called(ctx, category)
	return args.Error(0)
}

func (m *MockCategoryRepository) Update(ctx context.Context, category *entity.Category) error {
	args := m.Called(ctx, category)
	return args.Error(0)
}

func (m *MockCategoryRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCategoryRepository) List(ctx context.Context) ([]*entity.Category, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*entity.Category), args.Error(1)
}

func (m *MockCategoryRepository) ListByParentID(ctx context.Context, parentID int64) ([]*entity.Category, error) {
	args := m.Called(ctx, parentID)
	return args.Get(0).([]*entity.Category), args.Error(1)
}

func (m *MockCategoryRepository) GetTopLevelCategories(ctx context.Context) ([]*entity.Category, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*entity.Category), args.Error(1)
}

func (m *MockCategoryRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	args := m.Called(ctx, name)
	return args.Bool(0), args.Error(1)
}

func (m *MockCategoryRepository) ExistsByNameExcludeID(ctx context.Context, name string, excludeID int64) (bool, error) {
	args := m.Called(ctx, name, excludeID)
	return args.Bool(0), args.Error(1)
}

func (m *MockCategoryRepository) CountCategories(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

// 测试用例

func TestProductService_CreateProduct(t *testing.T) {
	// 设置测试数据库
	utils.InitTestDB()
	defer utils.CleanupTestData()

	// 创建模拟仓库
	mockProductRepo := new(MockProductRepository)
	mockCategoryRepo := new(MockCategoryRepository)

	// 创建服务实例
	service := NewProductService(mockProductRepo, mockCategoryRepo)

	ctx := context.Background()

	t.Run("成功创建商品", func(t *testing.T) {
		// 准备测试数据
		req := &dto.ProductCreateRequest{
			CategoryID:  1,
			Name:        "测试商品",
			Description: "测试商品描述",
			Price:       10000, // 100元
			Stock:       50,
			CoverImage:  "http://example.com/image.jpg",
		}

		category := &entity.Category{
			ID:   1,
			Name: "测试分类",
		}

		// 设置模拟期望
		mockCategoryRepo.On("GetByID", ctx, int64(1)).Return(category, nil)
		mockProductRepo.On("ExistsByName", ctx, "测试商品").Return(false, nil)
		mockProductRepo.On("Create", ctx, mock.AnythingOfType("*entity.Product")).Return(nil)

		// 执行测试
		result, err := service.CreateProduct(ctx, req)

		// 验证结果
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, req.Name, result.Name)
		assert.Equal(t, req.CategoryID, result.CategoryID)
		assert.Equal(t, req.Price, result.Price)
		assert.Equal(t, req.Stock, result.Stock)
		assert.Equal(t, 1, result.Status) // 默认上架

		// 验证模拟调用
		mockCategoryRepo.AssertExpectations(t)
		mockProductRepo.AssertExpectations(t)
	})

	t.Run("分类不存在", func(t *testing.T) {
		req := &dto.ProductCreateRequest{
			CategoryID:  999,
			Name:        "测试商品",
			Description: "测试商品描述",
			Price:       10000,
			Stock:       50,
		}

		// 设置模拟期望
		mockCategoryRepo.On("GetByID", ctx, int64(999)).Return(nil, entity.ErrCategoryNotFound)

		// 执行测试
		result, err := service.CreateProduct(ctx, req)

		// 验证结果
		assert.Error(t, err)
		assert.Equal(t, entity.ErrInvalidCategoryID, err)
		assert.Nil(t, result)

		// 验证模拟调用
		mockCategoryRepo.AssertExpectations(t)
	})

	t.Run("商品名称已存在", func(t *testing.T) {
		req := &dto.ProductCreateRequest{
			CategoryID:  1,
			Name:        "已存在商品",
			Description: "测试商品描述",
			Price:       10000,
			Stock:       50,
		}

		category := &entity.Category{
			ID:   1,
			Name: "测试分类",
		}

		// 设置模拟期望
		mockCategoryRepo.On("GetByID", ctx, int64(1)).Return(category, nil)
		mockProductRepo.On("ExistsByName", ctx, "已存在商品").Return(true, nil)

		// 执行测试
		result, err := service.CreateProduct(ctx, req)

		// 验证结果
		assert.Error(t, err)
		assert.Equal(t, entity.ErrProductAlreadyExists, err)
		assert.Nil(t, result)

		// 验证模拟调用
		mockCategoryRepo.AssertExpectations(t)
		mockProductRepo.AssertExpectations(t)
	})

	t.Run("无效的请求参数", func(t *testing.T) {
		// 测试空请求
		result, err := service.CreateProduct(ctx, nil)
		assert.Error(t, err)
		assert.Nil(t, result)

		// 测试无效分类ID
		req := &dto.ProductCreateRequest{
			CategoryID: 0,
			Name:       "测试商品",
			Price:      10000,
			Stock:      50,
		}
		result, err = service.CreateProduct(ctx, req)
		assert.Error(t, err)
		assert.Equal(t, entity.ErrInvalidCategoryID, err)
		assert.Nil(t, result)

		// 测试空商品名称
		req = &dto.ProductCreateRequest{
			CategoryID: 1,
			Name:       "",
			Price:      10000,
			Stock:      50,
		}
		result, err = service.CreateProduct(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, result)

		// 测试无效价格
		req = &dto.ProductCreateRequest{
			CategoryID: 1,
			Name:       "测试商品",
			Price:      0,
			Stock:      50,
		}
		result, err = service.CreateProduct(ctx, req)
		assert.Error(t, err)
		assert.Equal(t, entity.ErrInvalidPrice, err)
		assert.Nil(t, result)

		// 测试无效库存
		req = &dto.ProductCreateRequest{
			CategoryID: 1,
			Name:       "测试商品",
			Price:      10000,
			Stock:      -1,
		}
		result, err = service.CreateProduct(ctx, req)
		assert.Error(t, err)
		assert.Equal(t, entity.ErrInvalidStock, err)
		assert.Nil(t, result)
	})
}

func TestProductService_GetProductByID(t *testing.T) {
	// 设置测试数据库
	utils.InitTestDB()
	defer utils.CleanupTestData()

	// 创建模拟仓库
	mockProductRepo := new(MockProductRepository)
	mockCategoryRepo := new(MockCategoryRepository)

	// 创建服务实例
	service := NewProductService(mockProductRepo, mockCategoryRepo)

	ctx := context.Background()

	t.Run("成功获取商品", func(t *testing.T) {
		// 准备测试数据
		product := &entity.Product{
			ID:          1,
			CategoryID:  1,
			Name:        "测试商品",
			Description: "测试商品描述",
			Price:       10000,
			Stock:       50,
			CoverImage:  "http://example.com/image.jpg",
			Status:      1,
		}

		// 设置模拟期望
		mockProductRepo.On("GetByID", ctx, int64(1)).Return(product, nil)

		// 执行测试
		result, err := service.GetProductByID(ctx, 1)

		// 验证结果
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, product.ID, result.ID)
		assert.Equal(t, product.Name, result.Name)
		assert.Equal(t, product.Price, result.Price)

		// 验证模拟调用
		mockProductRepo.AssertExpectations(t)
	})

	t.Run("商品不存在", func(t *testing.T) {
		// 设置模拟期望
		mockProductRepo.On("GetByID", ctx, int64(999)).Return(nil, entity.ErrProductNotFound)

		// 执行测试
		result, err := service.GetProductByID(ctx, 999)

		// 验证结果
		assert.Error(t, err)
		assert.Equal(t, entity.ErrProductNotFound, err)
		assert.Nil(t, result)

		// 验证模拟调用
		mockProductRepo.AssertExpectations(t)
	})

	t.Run("无效的商品ID", func(t *testing.T) {
		// 执行测试
		result, err := service.GetProductByID(ctx, 0)

		// 验证结果
		assert.Error(t, err)
		assert.Equal(t, entity.ErrInvalidProductID, err)
		assert.Nil(t, result)

		result, err = service.GetProductByID(ctx, -1)
		assert.Error(t, err)
		assert.Equal(t, entity.ErrInvalidProductID, err)
		assert.Nil(t, result)
	})
}

func TestProductService_UpdateProductStock(t *testing.T) {
	// 设置测试数据库
	utils.InitTestDB()
	defer utils.CleanupTestData()

	// 创建模拟仓库
	mockProductRepo := new(MockProductRepository)
	mockCategoryRepo := new(MockCategoryRepository)

	// 创建服务实例
	service := NewProductService(mockProductRepo, mockCategoryRepo)

	ctx := context.Background()

	t.Run("成功更新库存", func(t *testing.T) {
		product := &entity.Product{
			ID:    1,
			Name:  "测试商品",
			Stock: 50,
		}

		// 设置模拟期望
		mockProductRepo.On("GetByID", ctx, int64(1)).Return(product, nil)
		mockProductRepo.On("UpdateStock", ctx, int64(1), 100).Return(nil)

		// 执行测试
		err := service.UpdateProductStock(ctx, 1, 100)

		// 验证结果
		assert.NoError(t, err)

		// 验证模拟调用
		mockProductRepo.AssertExpectations(t)
	})

	t.Run("商品不存在", func(t *testing.T) {
		// 设置模拟期望
		mockProductRepo.On("GetByID", ctx, int64(999)).Return(nil, entity.ErrProductNotFound)

		// 执行测试
		err := service.UpdateProductStock(ctx, 999, 100)

		// 验证结果
		assert.Error(t, err)
		assert.Equal(t, entity.ErrProductNotFound, err)

		// 验证模拟调用
		mockProductRepo.AssertExpectations(t)
	})

	t.Run("无效参数", func(t *testing.T) {
		// 无效商品ID
		err := service.UpdateProductStock(ctx, 0, 100)
		assert.Error(t, err)
		assert.Equal(t, entity.ErrInvalidProductID, err)

		// 无效库存
		err = service.UpdateProductStock(ctx, 1, -1)
		assert.Error(t, err)
		assert.Equal(t, entity.ErrInvalidStock, err)
	})
}

func TestProductService_CheckStock(t *testing.T) {
	// 设置测试数据库
	utils.InitTestDB()
	defer utils.CleanupTestData()

	// 创建模拟仓库
	mockProductRepo := new(MockProductRepository)
	mockCategoryRepo := new(MockCategoryRepository)

	// 创建服务实例
	service := NewProductService(mockProductRepo, mockCategoryRepo)

	ctx := context.Background()

	t.Run("库存充足", func(t *testing.T) {
		// 设置模拟期望
		mockProductRepo.On("CheckStock", ctx, int64(1), 10).Return(true, nil)

		// 执行测试
		sufficient, err := service.CheckStock(ctx, 1, 10)

		// 验证结果
		assert.NoError(t, err)
		assert.True(t, sufficient)

		// 验证模拟调用
		mockProductRepo.AssertExpectations(t)
	})

	t.Run("库存不足", func(t *testing.T) {
		// 设置模拟期望
		mockProductRepo.On("CheckStock", ctx, int64(1), 100).Return(false, nil)

		// 执行测试
		sufficient, err := service.CheckStock(ctx, 1, 100)

		// 验证结果
		assert.NoError(t, err)
		assert.False(t, sufficient)

		// 验证模拟调用
		mockProductRepo.AssertExpectations(t)
	})

	t.Run("无效参数", func(t *testing.T) {
		// 无效商品ID
		sufficient, err := service.CheckStock(ctx, 0, 10)
		assert.Error(t, err)
		assert.Equal(t, entity.ErrInvalidProductID, err)
		assert.False(t, sufficient)

		// 无效数量
		sufficient, err = service.CheckStock(ctx, 1, 0)
		assert.Error(t, err)
		assert.False(t, sufficient)
	})
}