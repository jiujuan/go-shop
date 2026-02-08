package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-shop/internal/entity"
	"go-shop/internal/repository"
	"go-shop/internal/service"
	"go-shop/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func setupProductTest(t *testing.T) (*ProductHandler, *gorm.DB, func()) {
	gin.SetMode(gin.TestMode)
	
	db := utils.InitTestDB()
	
	productRepo := repository.NewProductRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	productService := service.NewProductService(productRepo, categoryRepo)
	handler := NewProductHandler(productService)
	
	cleanup := func() {
		utils.CleanupTestData()
	}
	
	return handler, db, cleanup
}

func createTestCategory(t *testing.T, db *gorm.DB) *entity.Category {
	categoryRepo := repository.NewCategoryRepository(db)
	category := &entity.Category{
		Name:      "测试分类",
		ParentID:  0,
		SortOrder: 1,
	}
	categoryRepo.Create(context.Background(), category)
	return category
}

func createTestProduct(t *testing.T, db *gorm.DB, categoryID int64) *entity.Product {
	productRepo := repository.NewProductRepository(db)
	product := &entity.Product{
		CategoryID:  categoryID,
		Name:        "测试商品",
		Description: "测试商品描述",
		Price:       10000,
		Stock:       100,
		CoverImage:  "test.jpg",
		Status:      1,
	}
	productRepo.Create(context.Background(), product)
	return product
}

func TestProductHandler_GetProductList(t *testing.T) {
	handler, db, cleanup := setupProductTest(t)
	defer cleanup()

	category := createTestCategory(t, db)
	createTestProduct(t, db, category.ID)

	router := gin.New()
	router.GET("/products", handler.GetProductList)

	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestProductHandler_GetProductDetail(t *testing.T) {
	handler, db, cleanup := setupProductTest(t)
	defer cleanup()

	category := createTestCategory(t, db)
	product := createTestProduct(t, db, category.ID)

	router := gin.New()
	router.GET("/products/:id", handler.GetProductDetail)

	req := httptest.NewRequest(http.MethodGet, "/products/"+string(rune(product.ID+'0')), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestProductHandler_GetProductsByCategory(t *testing.T) {
	handler, db, cleanup := setupProductTest(t)
	defer cleanup()

	category := createTestCategory(t, db)
	createTestProduct(t, db, category.ID)

	router := gin.New()
	router.GET("/categories/:id/products", handler.GetProductsByCategory)

	req := httptest.NewRequest(http.MethodGet, "/categories/"+string(rune(category.ID+'0'))+"/products", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestProductHandler_SearchProducts(t *testing.T) {
	handler, db, cleanup := setupProductTest(t)
	defer cleanup()

	category := createTestCategory(t, db)
	createTestProduct(t, db, category.ID)

	router := gin.New()
	router.GET("/products/search", handler.SearchProducts)

	req := httptest.NewRequest(http.MethodGet, "/products/search?keyword=测试", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
