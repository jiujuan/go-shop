package repository

import (
	"context"
	"testing"

	"go-shop/internal/entity"

	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type ProductRepositoryTestSuite struct {
	suite.Suite
	db          *gorm.DB
	productRepo ProductRepository
	categoryRepo CategoryRepository
}

func (suite *ProductRepositoryTestSuite) SetupSuite() {
	// 使用内存SQLite数据库进行测试
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	suite.Require().NoError(err)

	// 自动迁移
	err = db.AutoMigrate(&entity.Product{}, &entity.Category{})
	suite.Require().NoError(err)

	suite.db = db
	suite.productRepo = NewProductRepository(db)
	suite.categoryRepo = NewCategoryRepository(db)
}

func (suite *ProductRepositoryTestSuite) TearDownTest() {
	// 清理测试数据
	suite.db.Exec("DELETE FROM products")
	suite.db.Exec("DELETE FROM categories")
	// 重置自增ID
	suite.db.Exec("DELETE FROM sqlite_sequence WHERE name IN ('products', 'categories')")
}

func (suite *ProductRepositoryTestSuite) TestCreateProduct() {
	ctx := context.Background()

	// 先创建一个分类
	category := &entity.Category{
		Name:      "测试分类",
		ParentID:  0,
		SortOrder: 1,
	}
	err := suite.categoryRepo.Create(ctx, category)
	suite.NoError(err)

	// 测试创建商品
	product := &entity.Product{
		CategoryID:  category.ID,
		Name:        "测试商品",
		Description: "这是一个测试商品",
		Price:       9999, // 99.99元
		Stock:       100,
		CoverImage:  "http://example.com/image.jpg",
		Status:      1,
	}

	err = suite.productRepo.Create(ctx, product)
	suite.NoError(err)
	suite.NotZero(product.ID)

	// 测试商品名称重复
	duplicateProduct := &entity.Product{
		CategoryID:  category.ID,
		Name:        "测试商品",
		Description: "另一个测试商品",
		Price:       8888,
		Stock:       50,
		Status:      1,
	}

	err = suite.productRepo.Create(ctx, duplicateProduct)
	suite.Error(err)
	suite.Equal(entity.ErrProductAlreadyExists, err)

	// 测试nil商品
	err = suite.productRepo.Create(ctx, nil)
	suite.Error(err)

	// 测试无效价格
	invalidPriceProduct := &entity.Product{
		CategoryID: category.ID,
		Name:       "无效价格商品",
		Price:      0,
		Stock:      10,
		Status:     1,
	}
	err = suite.productRepo.Create(ctx, invalidPriceProduct)
	suite.Error(err)
	suite.Equal(entity.ErrInvalidPrice, err)
}

func (suite *ProductRepositoryTestSuite) TestGetByID() {
	ctx := context.Background()

	// 创建测试分类和商品
	category := &entity.Category{Name: "测试分类", ParentID: 0}
	err := suite.categoryRepo.Create(ctx, category)
	suite.NoError(err)

	product := &entity.Product{
		CategoryID:  category.ID,
		Name:        "测试商品",
		Description: "测试描述",
		Price:       9999,
		Stock:       100,
		Status:      1,
	}
	err = suite.productRepo.Create(ctx, product)
	suite.NoError(err)

	// 测试获取存在的商品
	foundProduct, err := suite.productRepo.GetByID(ctx, product.ID)
	suite.NoError(err)
	suite.Equal(product.Name, foundProduct.Name)
	suite.Equal(product.Price, foundProduct.Price)

	// 测试获取不存在的商品
	_, err = suite.productRepo.GetByID(ctx, 99999)
	suite.Error(err)
	suite.Equal(entity.ErrProductNotFound, err)

	// 测试无效ID
	_, err = suite.productRepo.GetByID(ctx, 0)
	suite.Error(err)
	suite.Equal(entity.ErrInvalidProductID, err)
}

func (suite *ProductRepositoryTestSuite) TestGetByIDWithCategory() {
	ctx := context.Background()

	// 创建测试分类和商品
	category := &entity.Category{Name: "测试分类", ParentID: 0}
	err := suite.categoryRepo.Create(ctx, category)
	suite.NoError(err)

	product := &entity.Product{
		CategoryID: category.ID,
		Name:       "测试商品",
		Price:      9999,
		Stock:      100,
		Status:     1,
	}
	err = suite.productRepo.Create(ctx, product)
	suite.NoError(err)

	// 测试获取商品和分类信息
	foundProduct, foundCategory, err := suite.productRepo.GetByIDWithCategory(ctx, product.ID)
	suite.NoError(err)
	suite.Equal(product.Name, foundProduct.Name)
	suite.Equal(category.Name, foundCategory.Name)
}

func (suite *ProductRepositoryTestSuite) TestUpdateProduct() {
	ctx := context.Background()

	// 创建测试分类和商品
	category := &entity.Category{Name: "测试分类", ParentID: 0}
	err := suite.categoryRepo.Create(ctx, category)
	suite.NoError(err)

	product := &entity.Product{
		CategoryID: category.ID,
		Name:       "测试商品",
		Price:      9999,
		Stock:      100,
		Status:     1,
	}
	err = suite.productRepo.Create(ctx, product)
	suite.NoError(err)

	// 测试更新商品信息
	product.Name = "更新后的商品名称"
	product.Price = 8888
	err = suite.productRepo.Update(ctx, product)
	suite.NoError(err)

	// 验证更新
	updatedProduct, err := suite.productRepo.GetByID(ctx, product.ID)
	suite.NoError(err)
	suite.Equal("更新后的商品名称", updatedProduct.Name)
	suite.Equal(int64(8888), updatedProduct.Price)

	// 测试更新不存在的商品
	nonExistentProduct := &entity.Product{
		ID:   99999,
		Name: "不存在的商品",
	}
	err = suite.productRepo.Update(ctx, nonExistentProduct)
	suite.Error(err)
	suite.Equal(entity.ErrProductNotFound, err)
}

func (suite *ProductRepositoryTestSuite) TestUpdateStock() {
	ctx := context.Background()

	// 创建测试分类和商品
	category := &entity.Category{Name: "测试分类", ParentID: 0}
	err := suite.categoryRepo.Create(ctx, category)
	suite.NoError(err)

	product := &entity.Product{
		CategoryID: category.ID,
		Name:       "测试商品",
		Price:      9999,
		Stock:      100,
		Status:     1,
	}
	err = suite.productRepo.Create(ctx, product)
	suite.NoError(err)

	// 测试更新库存
	err = suite.productRepo.UpdateStock(ctx, product.ID, 50)
	suite.NoError(err)

	// 验证库存更新
	updatedProduct, err := suite.productRepo.GetByID(ctx, product.ID)
	suite.NoError(err)
	suite.Equal(50, updatedProduct.Stock)

	// 测试无效库存
	err = suite.productRepo.UpdateStock(ctx, product.ID, -1)
	suite.Error(err)
	suite.Equal(entity.ErrInvalidStock, err)
}

func (suite *ProductRepositoryTestSuite) TestStockOperations() {
	ctx := context.Background()

	// 创建测试分类和商品
	category := &entity.Category{Name: "测试分类", ParentID: 0}
	err := suite.categoryRepo.Create(ctx, category)
	suite.NoError(err)

	product := &entity.Product{
		CategoryID: category.ID,
		Name:       "测试商品",
		Price:      9999,
		Stock:      100,
		Status:     1,
	}
	err = suite.productRepo.Create(ctx, product)
	suite.NoError(err)

	// 测试检查库存
	sufficient, err := suite.productRepo.CheckStock(ctx, product.ID, 50)
	suite.NoError(err)
	suite.True(sufficient)

	insufficient, err := suite.productRepo.CheckStock(ctx, product.ID, 150)
	suite.NoError(err)
	suite.False(insufficient)

	// 测试减少库存
	err = suite.productRepo.DecrementStock(ctx, product.ID, 30)
	suite.NoError(err)

	// 验证库存减少
	updatedProduct, err := suite.productRepo.GetByID(ctx, product.ID)
	suite.NoError(err)
	suite.Equal(70, updatedProduct.Stock)

	// 测试库存不足时减少库存
	err = suite.productRepo.DecrementStock(ctx, product.ID, 100)
	suite.Error(err)
	suite.Equal(entity.ErrProductOutOfStock, err)

	// 测试增加库存
	err = suite.productRepo.IncrementStock(ctx, product.ID, 20)
	suite.NoError(err)

	// 验证库存增加
	updatedProduct, err = suite.productRepo.GetByID(ctx, product.ID)
	suite.NoError(err)
	suite.Equal(90, updatedProduct.Stock)
}

func (suite *ProductRepositoryTestSuite) TestListProducts() {
	ctx := context.Background()

	// 创建测试分类
	category1 := &entity.Category{Name: "分类1", ParentID: 0}
	category2 := &entity.Category{Name: "分类2", ParentID: 0}
	err := suite.categoryRepo.Create(ctx, category1)
	suite.NoError(err)
	err = suite.categoryRepo.Create(ctx, category2)
	suite.NoError(err)

	// 创建多个测试商品
	products := []*entity.Product{
		{CategoryID: category1.ID, Name: "商品1", Price: 1000, Stock: 10, Status: 1},
		{CategoryID: category1.ID, Name: "商品2", Price: 2000, Stock: 20, Status: 1},
		{CategoryID: category2.ID, Name: "商品3", Price: 3000, Stock: 30, Status: 0},
	}

	for _, product := range products {
		err := suite.productRepo.Create(ctx, product)
		suite.NoError(err)
	}

	// 测试获取所有商品
	options := ProductQueryOptions{
		Offset: 0,
		Limit:  10,
	}
	productList, total, err := suite.productRepo.List(ctx, options)
	suite.NoError(err)
	suite.Equal(int64(3), total)
	suite.Len(productList, 3)

	// 测试按分类筛选
	options.CategoryID = &category1.ID
	productList, total, err = suite.productRepo.List(ctx, options)
	suite.NoError(err)
	suite.Equal(int64(2), total)
	suite.Len(productList, 2)

	// 测试按状态筛选
	status := 1
	options.CategoryID = nil
	options.Status = &status
	productList, total, err = suite.productRepo.List(ctx, options)
	suite.NoError(err)
	suite.Equal(int64(2), total)
	suite.Len(productList, 2)

	// 测试价格范围筛选
	minPrice := int64(1500)
	maxPrice := int64(2500)
	options.Status = nil
	options.MinPrice = &minPrice
	options.MaxPrice = &maxPrice
	productList, total, err = suite.productRepo.List(ctx, options)
	suite.NoError(err)
	suite.Equal(int64(1), total)
	suite.Len(productList, 1)
	suite.Equal("商品2", productList[0].Name)
}

func (suite *ProductRepositoryTestSuite) TestSearchProducts() {
	ctx := context.Background()

	// 创建测试分类和商品
	category := &entity.Category{Name: "测试分类", ParentID: 0}
	err := suite.categoryRepo.Create(ctx, category)
	suite.NoError(err)

	products := []*entity.Product{
		{CategoryID: category.ID, Name: "苹果手机", Description: "最新款苹果手机", Price: 5000, Stock: 10, Status: 1},
		{CategoryID: category.ID, Name: "华为手机", Description: "华为旗舰手机", Price: 4000, Stock: 20, Status: 1},
		{CategoryID: category.ID, Name: "小米电视", Description: "小米智能电视", Price: 3000, Stock: 15, Status: 1},
	}

	for _, product := range products {
		err := suite.productRepo.Create(ctx, product)
		suite.NoError(err)
	}

	// 测试搜索手机
	productList, total, err := suite.productRepo.Search(ctx, "手机", 0, 10)
	suite.NoError(err)
	suite.Equal(int64(2), total)
	suite.Len(productList, 2)

	// 测试搜索苹果
	productList, total, err = suite.productRepo.Search(ctx, "苹果", 0, 10)
	suite.NoError(err)
	suite.Equal(int64(1), total)
	suite.Len(productList, 1)
	suite.Equal("苹果手机", productList[0].Name)
}

func (suite *ProductRepositoryTestSuite) TestExistsByName() {
	ctx := context.Background()

	// 创建测试分类和商品
	category := &entity.Category{Name: "测试分类", ParentID: 0}
	err := suite.categoryRepo.Create(ctx, category)
	suite.NoError(err)

	product := &entity.Product{
		CategoryID: category.ID,
		Name:       "测试商品",
		Price:      9999,
		Stock:      100,
		Status:     1,
	}
	err = suite.productRepo.Create(ctx, product)
	suite.NoError(err)

	// 测试存在的商品名称
	exists, err := suite.productRepo.ExistsByName(ctx, "测试商品")
	suite.NoError(err)
	suite.True(exists)

	// 测试不存在的商品名称
	exists, err = suite.productRepo.ExistsByName(ctx, "不存在的商品")
	suite.NoError(err)
	suite.False(exists)

	// 测试空商品名称
	exists, err = suite.productRepo.ExistsByName(ctx, "")
	suite.NoError(err)
	suite.False(exists)
}

func (suite *ProductRepositoryTestSuite) TestCountProducts() {
	ctx := context.Background()

	// 初始计数应该为0
	count, err := suite.productRepo.CountProducts(ctx)
	suite.NoError(err)
	suite.Equal(int64(0), count)

	// 创建测试分类和商品
	category := &entity.Category{Name: "测试分类", ParentID: 0}
	err = suite.categoryRepo.Create(ctx, category)
	suite.NoError(err)

	product := &entity.Product{
		CategoryID: category.ID,
		Name:       "测试商品",
		Price:      9999,
		Stock:      100,
		Status:     1,
	}
	err = suite.productRepo.Create(ctx, product)
	suite.NoError(err)

	// 计数应该为1
	count, err = suite.productRepo.CountProducts(ctx)
	suite.NoError(err)
	suite.Equal(int64(1), count)
}

func TestProductRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(ProductRepositoryTestSuite))
}