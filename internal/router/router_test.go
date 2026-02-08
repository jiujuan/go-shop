package router

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-shop/internal/handler"
	"go-shop/internal/repository"
	"go-shop/internal/service"
	"go-shop/pkg/auth"
	"go-shop/pkg/cache"
	"go-shop/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupRouterTest(t *testing.T) (*gin.Engine, func()) {
	gin.SetMode(gin.TestMode)

	db := utils.InitTestDB()

	redisClient, err := cache.NewRedisClient(cache.RedisConfig{
		Host:     "localhost",
		Port:     6379,
		Password: "",
		Database: 1,
	})
	if err != nil {
		t.Skipf("Redis not available: %v", err)
	}

	// 初始化仓库
	userRepo := repository.NewUserRepository(db)
	productRepo := repository.NewProductRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	cartRepo := repository.NewCartRepository(redisClient, productRepo)
	addressRepo := repository.NewAddressRepository(db)
	skuRepo := repository.NewSKURepository(db)

	// 初始化服务
	passwordManager := auth.NewPasswordManager(nil)
	jwtManager := auth.NewJWTManager("test-secret", 24)

	userService := service.NewUserService(userRepo, passwordManager, jwtManager)
	productService := service.NewProductService(productRepo, categoryRepo)
	cartService := service.NewCartService(cartRepo, productRepo)
	inventoryService := service.NewInventoryService(skuRepo, redisClient, db)
	orderService := service.NewOrderService(orderRepo, cartRepo, productRepo, addressRepo, skuRepo, nil, inventoryService, nil)
	paymentService := service.NewPaymentService(orderRepo, productRepo)
	addressService := service.NewAddressService(addressRepo)
	skuService := service.NewSKUService(skuRepo, productRepo)

	// 初始化处理器
	authHandler := handler.NewAuthHandler(userService, jwtManager)
	userHandler := handler.NewUserHandler(userService, addressService)
	productHandler := handler.NewProductHandler(productService)
	cartHandler := handler.NewCartHandler(cartService)
	orderHandler := handler.NewOrderHandler(orderService, paymentService)
	adminHandler := handler.NewAdminHandler(userService, productService, orderService, skuService)
	skuHandler := handler.NewSKUHandler(skuService)

	// 配置路由
	config := &RouterConfig{
		AuthHandler:    authHandler,
		UserHandler:    userHandler,
		ProductHandler: productHandler,
		CartHandler:    cartHandler,
		OrderHandler:   orderHandler,
		AdminHandler:   adminHandler,
		SKUHandler:     skuHandler,
		JWTManager:     jwtManager,
	}

	router := SetupRouter(config)

	cleanup := func() {
		utils.CleanupTestData()
		redisClient.FlushDB(context.Background())
	}

	return router, cleanup
}

func TestRouter_HealthCheck(t *testing.T) {
	router, cleanup := setupRouterTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "ok")
}

func TestRouter_PublicRoutes(t *testing.T) {
	router, cleanup := setupRouterTest(t)
	defer cleanup()

	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
	}{
		{
			name:       "Get product list",
			method:     http.MethodGet,
			path:       "/api/v1/products",
			wantStatus: http.StatusOK,
		},
		{
			name:       "Get payment methods",
			method:     http.MethodGet,
			path:       "/api/v1/payment-methods",
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestRouter_AuthenticatedRoutes_Unauthorized(t *testing.T) {
	router, cleanup := setupRouterTest(t)
	defer cleanup()

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{
			name:   "Get user info without auth",
			method: http.MethodGet,
			path:   "/api/v1/users/me",
		},
		{
			name:   "Get cart without auth",
			method: http.MethodGet,
			path:   "/api/v1/cart",
		},
		{
			name:   "Get orders without auth",
			method: http.MethodGet,
			path:   "/api/v1/orders",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Code)
		})
	}
}

func TestRouter_AdminRoutes_Unauthorized(t *testing.T) {
	router, cleanup := setupRouterTest(t)
	defer cleanup()

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{
			name:   "Get admin orders without auth",
			method: http.MethodGet,
			path:   "/api/v1/admin/orders",
		},
		{
			name:   "Get admin users without auth",
			method: http.MethodGet,
			path:   "/api/v1/admin/users",
		},
		{
			name:   "Get statistics without auth",
			method: http.MethodGet,
			path:   "/api/v1/admin/statistics",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Code)
		})
	}
}

func TestRouter_APIVersioning(t *testing.T) {
	router, cleanup := setupRouterTest(t)
	defer cleanup()

	// 测试 v1 API
	req := httptest.NewRequest(http.MethodGet, "/api/v1/products", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// 测试不存在的版本
	req = httptest.NewRequest(http.MethodGet, "/api/v2/products", nil)
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
