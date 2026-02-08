package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-shop/config"
	"go-shop/internal/handler"
	"go-shop/internal/repository"
	"go-shop/internal/router"
	"go-shop/internal/service"
	"go-shop/pkg/auth"
	"go-shop/pkg/cache"
	"go-shop/pkg/database"
	"go-shop/pkg/logger"
	"go-shop/pkg/oss"
	"go-shop/pkg/sms"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func main() {
	// 初始化日志
	if err := logger.InitLogger("info", "logs/app.log"); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// 初始化请求日志记录器（按日期分文件）
	if err := logger.InitRequestLogger("logs/requests"); err != nil {
		log.Fatalf("Failed to initialize request logger: %v", err)
	}

	logger.Info("Go-Shop 电商系统启动中...")

	// 加载配置
	cfg, err := loadConfig()
	if err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}
	logger.Info("Configuration loaded successfully")

	// 初始化数据库
	db, err := database.NewMySQLConnection(cfg.Database.MySQL)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	logger.Info("Database connected successfully", zap.String("host", cfg.Database.MySQL.Host))

	// 初始化 Redis
	redisClient, err := cache.NewRedisClient(cfg.Cache.Redis)
	if err != nil {
		logger.Fatal("Failed to connect to Redis", zap.Error(err))
	}
	logger.Info("Redis connected successfully", zap.String("host", cfg.Cache.Redis.Host))

	// 初始化依赖
	deps := initializeDependencies(db, redisClient, cfg)
	logger.Info("Dependencies initialized successfully")

	// 设置 Gin 模式
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 配置路由
	routerConfig := &router.RouterConfig{
		AuthHandler:           deps.AuthHandler,
		UserHandler:           deps.UserHandler,
		ProductHandler:        deps.ProductHandler,
		CartHandler:           deps.CartHandler,
		OrderHandler:          deps.OrderHandler,
		AdminHandler:          deps.AdminHandler,
		SKUHandler:            deps.SKUHandler,
		CouponHandler:         deps.CouponHandler,
		RefundHandler:         deps.RefundHandler,
		ReviewHandler:         deps.ReviewHandler,
		UploadHandler:         deps.UploadHandler,
		FavoriteHandler:       deps.FavoriteHandler,
		NotificationHandler:   deps.NotificationHandler,
		SMSHandler:            deps.SMSHandler,
		StatisticsHandler:     deps.StatisticsHandler,
		OperationLogHandler:   deps.OperationLogHandler,
		RecommendationHandler: deps.RecommendationHandler,
		JWTManager:            deps.JWTManager,
	}
	r := router.SetupRouter(routerConfig)

	// 创建 HTTP 服务器
	srv := &http.Server{
		Addr:           fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:        r,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	// 启动服务器
	go func() {
		logger.Info("Server starting", zap.Int("port", cfg.Server.Port), zap.String("mode", cfg.Server.Mode))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// 优雅关闭
	gracefulShutdown(srv, db, redisClient)
}

// Dependencies 依赖容器
type Dependencies struct {
	// Managers
	JWTManager      *auth.JWTManager
	PasswordManager *auth.PasswordManager

	// Repositories
	UserRepo         repository.UserRepository
	ProductRepo      repository.ProductRepository
	CategoryRepo     repository.CategoryRepository
	OrderRepo        repository.OrderRepository
	CartRepo         repository.CartRepository
	AddressRepo      repository.AddressRepository
	SKURepo          repository.SKURepository
	CouponRepo       repository.CouponRepository
	UserCouponRepo   repository.UserCouponRepository
	RefundRepo       repository.RefundRepository
	ReviewRepo       repository.ReviewRepository
	ImageRepo        repository.ImageRepository
	FavoriteRepo     repository.FavoriteRepository
	NotificationRepo repository.NotificationRepository
	StatisticsRepo   repository.StatisticsRepository
	OperationLogRepo repository.OperationLogRepository
	UserBehaviorRepo repository.UserBehaviorRepository

	// Services
	UserService            *service.UserService
	ProductService         *service.ProductService
	CartService            *service.CartService
	OrderService           *service.OrderService
	PaymentService         *service.PaymentService
	AddressService         service.AddressService
	SKUService             *service.SKUService
	CouponService          *service.CouponService
	InventoryService       service.InventoryService
	RefundService          *service.RefundService
	ReviewService          *service.ReviewService
	UploadService          service.UploadService
	FavoriteService        service.FavoriteService
	NotificationService    service.NotificationService
	SMSService             service.SMSService
	StatisticsService      service.StatisticsService
	RecommendationService  service.RecommendationService

	// Handlers
	AuthHandler           *handler.AuthHandler
	UserHandler           *handler.UserHandler
	ProductHandler        *handler.ProductHandler
	CartHandler           *handler.CartHandler
	OrderHandler          *handler.OrderHandler
	AdminHandler          *handler.AdminHandler
	SKUHandler            *handler.SKUHandler
	CouponHandler         *handler.CouponHandler
	RefundHandler         *handler.RefundHandler
	ReviewHandler         *handler.ReviewHandler
	UploadHandler         *handler.UploadHandler
	FavoriteHandler       *handler.FavoriteHandler
	NotificationHandler   *handler.NotificationHandler
	SMSHandler            *handler.SMSHandler
	StatisticsHandler     *handler.StatisticsHandler
	OperationLogHandler   *handler.OperationLogHandler
	RecommendationHandler *handler.RecommendationHandler
}

// loadConfig 加载配置
func loadConfig() (*config.Config, error) {
	paths := []string{
		"config/config.yaml",
		"./config/config.yaml",
		"../config/config.yaml",
	}

	var cfg *config.Config
	var err error

	for _, path := range paths {
		cfg, err = config.LoadConfig(path)
		if err == nil {
			logger.Info("Configuration loaded", zap.String("path", path))
			return cfg, nil
		}
	}

	return nil, fmt.Errorf("failed to load config from any path: %w", err)
}

// initializeDependencies 初始化所有依赖
func initializeDependencies(db *gorm.DB, redisClient *redis.Client, cfg *config.Config) *Dependencies {
	deps := &Dependencies{}

	// 初始化管理器
	deps.JWTManager = auth.NewJWTManager(cfg.JWT.Secret, cfg.JWT.ExpireHours)
	deps.PasswordManager = auth.NewPasswordManager(nil)

	// 初始化仓库
	deps.UserRepo = repository.NewUserRepository(db)
	deps.ProductRepo = repository.NewProductRepository(db)
	deps.CategoryRepo = repository.NewCategoryRepository(db)
	deps.OrderRepo = repository.NewOrderRepository(db)
	deps.CartRepo = repository.NewCartRepository(redisClient, deps.ProductRepo)
	deps.AddressRepo = repository.NewAddressRepository(db)
	deps.SKURepo = repository.NewSKURepository(db)
	deps.CouponRepo = repository.NewCouponRepository(db)
	deps.UserCouponRepo = repository.NewUserCouponRepository(db)
	deps.RefundRepo = repository.NewRefundRepository(db)
	deps.ReviewRepo = repository.NewReviewRepository(db)
	deps.ImageRepo = repository.NewImageRepository(db)
	deps.FavoriteRepo = repository.NewFavoriteRepository(db)
	deps.NotificationRepo = repository.NewNotificationRepository(db)
	deps.StatisticsRepo = repository.NewStatisticsRepository(db)
	deps.OperationLogRepo = repository.NewOperationLogRepository(db)
	deps.UserBehaviorRepo = repository.NewUserBehaviorRepository(db)

	// 初始化 OSS 客户端（如果配置了）
	var ossClient oss.Client
	if cfg.OSS.Endpoint != "" {
		ossConfig := &oss.Config{
			Endpoint:        cfg.OSS.Endpoint,
			AccessKeyID:     cfg.OSS.AccessKeyID,
			AccessKeySecret: cfg.OSS.AccessKeySecret,
			BucketName:      cfg.OSS.BucketName,
			Domain:          cfg.OSS.Domain,
			UseSSL:          cfg.OSS.UseSSL,
			Region:          cfg.OSS.Region,
		}
		
		var err error
		ossClient, err = oss.NewClient(ossConfig)
		if err != nil {
			logger.Warn("Failed to initialize OSS client, upload service will not be available", zap.Error(err))
		} else {
			logger.Info("OSS client initialized successfully", zap.String("endpoint", cfg.OSS.Endpoint))
		}
	}

	// 初始化 SMS 客户端（如果配置了）
	var smsClient sms.Client
	if cfg.SMS.AccessKeyID != "" {
		smsConfig := &sms.Config{
			AccessKeyID:     cfg.SMS.AccessKeyID,
			AccessKeySecret: cfg.SMS.AccessKeySecret,
			SignName:        cfg.SMS.SignName,
			TemplateCode:    cfg.SMS.TemplateCode,
			Region:          cfg.SMS.Region,
		}
		
		var err error
		smsClient, err = sms.NewClient(smsConfig)
		if err != nil {
			logger.Warn("Failed to initialize SMS client, SMS service will not be available", zap.Error(err))
		} else {
			logger.Info("SMS client initialized successfully")
		}
	}

	// 初始化服务
	deps.UserService = service.NewUserService(deps.UserRepo, deps.PasswordManager, deps.JWTManager)
	deps.ProductService = service.NewProductService(deps.ProductRepo, deps.CategoryRepo)
	deps.CartService = service.NewCartService(deps.CartRepo, deps.ProductRepo, deps.SKURepo)
	deps.SKUService = service.NewSKUService(deps.SKURepo, deps.ProductRepo)
	deps.CouponService = service.NewCouponService(deps.CouponRepo, deps.UserCouponRepo, redisClient)
	deps.InventoryService = service.NewInventoryService(deps.SKURepo, redisClient, db)
	deps.RefundService = service.NewRefundService(deps.RefundRepo, deps.OrderRepo, deps.SKURepo, deps.InventoryService, deps.CouponService)
	deps.ReviewService = service.NewReviewService(deps.ReviewRepo, deps.OrderRepo)
	deps.FavoriteService = service.NewFavoriteService(deps.FavoriteRepo, deps.ProductRepo)
	deps.NotificationService = service.NewNotificationService(deps.NotificationRepo)
	deps.OrderService = service.NewOrderService(deps.OrderRepo, deps.CartRepo, deps.ProductRepo, deps.AddressRepo, deps.SKURepo, deps.CouponService, deps.InventoryService, nil)
	deps.PaymentService = service.NewPaymentService(deps.OrderRepo, deps.ProductRepo)
	deps.AddressService = service.NewAddressService(deps.AddressRepo)
	
	// 初始化上传服务（如果 OSS 客户端已初始化）
	if ossClient != nil {
		deps.UploadService = service.NewUploadService(ossClient, deps.ImageRepo)
		logger.Info("Upload service initialized successfully")
	}

	// 初始化短信服务（如果 SMS 客户端已初始化）
	if smsClient != nil {
		deps.SMSService = service.NewSMSService(smsClient, redisClient)
		logger.Info("SMS service initialized successfully")
	}

	// 初始化统计服务
	deps.StatisticsService = service.NewStatisticsService(deps.StatisticsRepo, deps.UserRepo)
	logger.Info("Statistics service initialized successfully")

	// 初始化推荐服务
	deps.RecommendationService = service.NewRecommendationService(deps.UserBehaviorRepo, deps.ProductRepo)
	logger.Info("Recommendation service initialized successfully")

	// 初始化处理器
	deps.AuthHandler = handler.NewAuthHandler(deps.UserService, deps.JWTManager)
	deps.UserHandler = handler.NewUserHandler(deps.UserService, deps.AddressService)
	deps.ProductHandler = handler.NewProductHandler(deps.ProductService)
	deps.CartHandler = handler.NewCartHandler(deps.CartService)
	deps.OrderHandler = handler.NewOrderHandler(deps.OrderService, deps.PaymentService)
	deps.AdminHandler = handler.NewAdminHandler(deps.UserService, deps.ProductService, deps.OrderService, deps.SKUService)
	deps.SKUHandler = handler.NewSKUHandler(deps.SKUService)
	deps.CouponHandler = handler.NewCouponHandler(deps.CouponService)
	deps.RefundHandler = handler.NewRefundHandler(deps.RefundService)
	deps.ReviewHandler = handler.NewReviewHandler(deps.ReviewService)
	deps.FavoriteHandler = handler.NewFavoriteHandler(deps.FavoriteService)
	deps.NotificationHandler = handler.NewNotificationHandler(deps.NotificationService)
	
	// 初始化上传处理器（如果上传服务已初始化）
	if deps.UploadService != nil {
		deps.UploadHandler = handler.NewUploadHandler(deps.UploadService)
		logger.Info("Upload handler initialized successfully")
	}

	// 初始化短信处理器（如果短信服务已初始化）
	if deps.SMSService != nil {
		deps.SMSHandler = handler.NewSMSHandler(deps.SMSService, deps.UserService)
		logger.Info("SMS handler initialized successfully")
	}

	// 初始化统计处理器
	deps.StatisticsHandler = handler.NewStatisticsHandler(deps.StatisticsService)
	logger.Info("Statistics handler initialized successfully")

	// 初始化操作日志处理器
	deps.OperationLogHandler = handler.NewOperationLogHandler(deps.OperationLogRepo)
	logger.Info("Operation log handler initialized successfully")

	// 初始化推荐处理器
	deps.RecommendationHandler = handler.NewRecommendationHandler(deps.RecommendationService)
	logger.Info("Recommendation handler initialized successfully")

	return deps
}

// gracefulShutdown 优雅关闭
func gracefulShutdown(srv *http.Server, db *gorm.DB, redisClient *redis.Client) {
	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// 设置超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 关闭 HTTP 服务器
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	// 关闭数据库连接
	if db != nil {
		sqlDB, err := db.DB()
		if err == nil {
			if err := sqlDB.Close(); err != nil {
				logger.Error("Failed to close database connection", zap.Error(err))
			} else {
				logger.Info("Database connection closed")
			}
		}
	}

	// 关闭 Redis 连接
	if redisClient != nil {
		if err := redisClient.Close(); err != nil {
			logger.Error("Failed to close Redis connection", zap.Error(err))
		} else {
			logger.Info("Redis connection closed")
		}
	}

	logger.Info("Server exited successfully")
}
