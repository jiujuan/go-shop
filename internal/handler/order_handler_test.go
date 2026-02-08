package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-shop/internal/dto"
	"go-shop/internal/entity"
	"go-shop/internal/middleware"
	"go-shop/internal/repository"
	"go-shop/internal/service"
	"go-shop/pkg/auth"
	"go-shop/pkg/cache"
	"go-shop/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func setupOrderTest(t *testing.T) (*OrderHandler, *gorm.DB, *auth.JWTManager, func()) {
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
	
	orderRepo := repository.NewOrderRepository(db)
	productRepo := repository.NewProductRepository(db)
	skuRepo := repository.NewSKURepository(db)
	cartRepo := repository.NewCartRepository(redisClient, productRepo)
	addressRepo := repository.NewAddressRepository(db)
	
	// Create nil CouponService for tests that don't use coupons
	inventoryService := service.NewInventoryService(skuRepo, redisClient, db)
	orderService := service.NewOrderService(orderRepo, cartRepo, productRepo, addressRepo, skuRepo, nil, inventoryService, nil)
	paymentService := service.NewPaymentService(orderRepo, productRepo)
	handler := NewOrderHandler(orderService, paymentService)
	
	jwtManager := auth.NewJWTManager("test-secret", 24)
	
	cleanup := func() {
		utils.CleanupTestData()
		redisClient.FlushDB(context.Background())
	}
	
	return handler, db, jwtManager, cleanup
}

func createTestAddress(t *testing.T, db *gorm.DB, userID int64) *entity.Address {
	addressRepo := repository.NewAddressRepository(db)
	address := &entity.Address{
		UserID:        userID,
		RecipientName: "张三",
		Phone:         "13800138000",
		Province:      "北京市",
		City:          "北京市",
		District:      "朝阳区",
		Detail:        "某某街道123号",
		IsDefault:     true,
	}
	addressRepo.Create(context.Background(), address)
	return address
}

func TestOrderHandler_CreateOrder(t *testing.T) {
	handler, db, jwtManager, cleanup := setupOrderTest(t)
	defer cleanup()

	user := createTestUser(t, db)
	category := createTestCategory(t, db)
	product := createTestProduct(t, db, category.ID)
	address := createTestAddress(t, db, user.ID)
	token, _ := jwtManager.GenerateToken(user.ID, user.Username, false)

	redisClient, _ := cache.NewRedisClient(cache.RedisConfig{
		Host:     "localhost",
		Port:     6379,
		Password: "",
		Database: 1,
	})
	productRepo := repository.NewProductRepository(db)
	cartRepo := repository.NewCartRepository(redisClient, productRepo)
	cartRepo.AddItem(context.Background(), user.ID, entity.CartItem{
		UserID:    user.ID,
		ProductID: product.ID,
		Quantity:  1,
		Product:   product,
	})

	router := gin.New()
	router.Use(middleware.AuthMiddleware(jwtManager))
	router.POST("/orders", handler.CreateOrder)

	reqBody := dto.OrderCreateRequest{
		AddressID: address.ID,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestOrderHandler_GetOrderList(t *testing.T) {
	handler, db, jwtManager, cleanup := setupOrderTest(t)
	defer cleanup()

	user := createTestUser(t, db)
	token, _ := jwtManager.GenerateToken(user.ID, user.Username, false)

	router := gin.New()
	router.Use(middleware.AuthMiddleware(jwtManager))
	router.GET("/orders", handler.GetOrderList)

	req := httptest.NewRequest(http.MethodGet, "/orders", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestOrderHandler_CreatePayment(t *testing.T) {
	handler, db, jwtManager, cleanup := setupOrderTest(t)
	defer cleanup()

	user := createTestUser(t, db)
	category := createTestCategory(t, db)
	product := createTestProduct(t, db, category.ID)
	address := createTestAddress(t, db, user.ID)
	token, _ := jwtManager.GenerateToken(user.ID, user.Username, false)

	// 创建订单
	redisClient, _ := cache.NewRedisClient(cache.RedisConfig{
		Host:     "localhost",
		Port:     6379,
		Password: "",
		Database: 1,
	})
	productRepo := repository.NewProductRepository(db)
	cartRepo := repository.NewCartRepository(redisClient, productRepo)
	cartRepo.AddItem(context.Background(), user.ID, entity.CartItem{
		UserID:    user.ID,
		ProductID: product.ID,
		Quantity:  1,
		Product:   product,
	})

	orderRepo := repository.NewOrderRepository(db)
	skuRepo := repository.NewSKURepository(db)
	addressRepo := repository.NewAddressRepository(db)
	inventoryService := service.NewInventoryService(skuRepo, redisClient, db)
	orderService := service.NewOrderService(orderRepo, cartRepo, productRepo, addressRepo, skuRepo, nil, inventoryService, nil)
	orderResp, _ := orderService.CreateOrder(context.Background(), user.ID, &dto.OrderCreateRequest{
		AddressID: address.ID,
	})

	router := gin.New()
	router.Use(middleware.AuthMiddleware(jwtManager))
	router.POST("/payments", handler.CreatePayment)

	reqBody := dto.PaymentRequest{
		OrderID:     orderResp.ID,
		PaymentType: "alipay",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/payments", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	data := response["data"].(map[string]interface{})
	assert.Equal(t, orderResp.ID, data["order_id"])
	assert.NotEmpty(t, data["payment_url"])
}

func TestOrderHandler_CheckPaymentStatus(t *testing.T) {
	handler, db, jwtManager, cleanup := setupOrderTest(t)
	defer cleanup()

	user := createTestUser(t, db)
	category := createTestCategory(t, db)
	product := createTestProduct(t, db, category.ID)
	address := createTestAddress(t, db, user.ID)
	token, _ := jwtManager.GenerateToken(user.ID, user.Username, false)

	// 创建订单
	redisClient, _ := cache.NewRedisClient(cache.RedisConfig{
		Host:     "localhost",
		Port:     6379,
		Password: "",
		Database: 1,
	})
	productRepo := repository.NewProductRepository(db)
	cartRepo := repository.NewCartRepository(redisClient, productRepo)
	cartRepo.AddItem(context.Background(), user.ID, entity.CartItem{
		UserID:    user.ID,
		ProductID: product.ID,
		Quantity:  1,
		Product:   product,
	})

	orderRepo := repository.NewOrderRepository(db)
	skuRepo := repository.NewSKURepository(db)
	addressRepo := repository.NewAddressRepository(db)
	inventoryService := service.NewInventoryService(skuRepo, redisClient, db)
	orderService := service.NewOrderService(orderRepo, cartRepo, productRepo, addressRepo, skuRepo, nil, inventoryService, nil)
	orderResp, _ := orderService.CreateOrder(context.Background(), user.ID, &dto.OrderCreateRequest{
		AddressID: address.ID,
	})

	router := gin.New()
	router.Use(middleware.AuthMiddleware(jwtManager))
	router.GET("/orders/:id/payment-status", handler.CheckPaymentStatus)

	req := httptest.NewRequest(http.MethodGet, "/orders/"+orderResp.ID+"/payment-status", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	data := response["data"].(map[string]interface{})
	assert.Equal(t, "pending", data["payment_status"])
}

func TestOrderHandler_GetPaymentMethods(t *testing.T) {
	handler, _, _, cleanup := setupOrderTest(t)
	defer cleanup()

	router := gin.New()
	router.GET("/payment-methods", handler.GetPaymentMethods)

	req := httptest.NewRequest(http.MethodGet, "/payment-methods", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	data := response["data"].(map[string]interface{})
	methods := data["payment_methods"].([]interface{})
	assert.GreaterOrEqual(t, len(methods), 2)
}

func TestOrderHandler_SimulatePaymentCallback(t *testing.T) {
	handler, db, _, cleanup := setupOrderTest(t)
	defer cleanup()

	user := createTestUser(t, db)
	category := createTestCategory(t, db)
	product := createTestProduct(t, db, category.ID)
	address := createTestAddress(t, db, user.ID)

	// 创建订单
	redisClient, _ := cache.NewRedisClient(cache.RedisConfig{
		Host:     "localhost",
		Port:     6379,
		Password: "",
		Database: 1,
	})
	productRepo := repository.NewProductRepository(db)
	cartRepo := repository.NewCartRepository(redisClient, productRepo)
	cartRepo.AddItem(context.Background(), user.ID, entity.CartItem{
		UserID:    user.ID,
		ProductID: product.ID,
		Quantity:  1,
		Product:   product,
	})

	orderRepo := repository.NewOrderRepository(db)
	skuRepo := repository.NewSKURepository(db)
	addressRepo := repository.NewAddressRepository(db)
	inventoryService := service.NewInventoryService(skuRepo, redisClient, db)
	orderService := service.NewOrderService(orderRepo, cartRepo, productRepo, addressRepo, skuRepo, nil, inventoryService, nil)
	orderResp, _ := orderService.CreateOrder(context.Background(), user.ID, &dto.OrderCreateRequest{
		AddressID: address.ID,
	})

	router := gin.New()
	router.POST("/orders/:id/simulate-payment", handler.SimulatePaymentCallback)

	reqBody := map[string]bool{"success": true}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/orders/"+orderResp.ID+"/simulate-payment", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// 验证订单状态已更新
	order, _ := orderRepo.GetByID(context.Background(), orderResp.ID)
	assert.Equal(t, entity.OrderStatusPaid, order.Status)
}
