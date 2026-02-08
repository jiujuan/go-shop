package router

import (
	"go-shop/internal/handler"
	"go-shop/internal/middleware"
	"go-shop/pkg/auth"

	"github.com/gin-gonic/gin"
)

type RouterConfig struct {
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
	JWTManager            *auth.JWTManager
}

// SetupRouter 配置路由
func SetupRouter(config *RouterConfig) *gin.Engine {
	router := gin.New()

	// 全局中间件
	router.Use(middleware.LoggerMiddleware())
	router.Use(middleware.RecoveryMiddleware())
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.RateLimitMiddleware(100, 100)) // 每秒100个请求，容量100

	// API v1 路由组
	v1 := router.Group("/api/v1")
	{
		// 公开路由（无需认证）
		setupPublicRoutes(v1, config)

		// 需要认证的路由
		setupAuthenticatedRoutes(v1, config)

		// 管理员路由
		setupAdminRoutes(v1, config)
	}

	// API v2 路由组（新功能）
	v2 := router.Group("/api/v2")
	{
		// 公开路由 V2（无需认证）
		setupPublicV2Routes(v2, config)

		// 管理员路由 V2
		setupAdminV2Routes(v2, config)
	}

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"message": "Go-Shop API is running",
		})
	})

	return router
}

// setupPublicRoutes 配置公开路由
func setupPublicRoutes(rg *gin.RouterGroup, config *RouterConfig) {
	// 认证相关
	auth := rg.Group("/auth")
	{
		auth.POST("/register", config.AuthHandler.Register)
		auth.POST("/login", config.AuthHandler.Login)
	}

	// 商品相关（浏览商品无需登录）
	products := rg.Group("/products")
	{
		products.GET("", config.ProductHandler.GetProductList)
		products.GET("/:id", config.ProductHandler.GetProductDetail)
	}

	// 支付方式查询（无需登录）
	rg.GET("/payment-methods", config.OrderHandler.GetPaymentMethods)
}

// setupAuthenticatedRoutes 配置需要认证的路由
func setupAuthenticatedRoutes(rg *gin.RouterGroup, config *RouterConfig) {
	// 应用认证中间件
	authenticated := rg.Group("")
	authenticated.Use(middleware.AuthMiddleware(config.JWTManager))

	// 认证相关
	auth := authenticated.Group("/auth")
	{
		auth.POST("/logout", config.AuthHandler.Logout)
		auth.POST("/refresh", config.AuthHandler.RefreshToken)
		auth.GET("/profile", config.AuthHandler.GetProfile)
	}

	// 用户相关
	users := authenticated.Group("/users")
	{
		users.GET("/me", config.UserHandler.GetUserInfo)
		users.PUT("/me", config.UserHandler.UpdateUserInfo)

		// 地址管理
		addresses := users.Group("/addresses")
		{
			addresses.GET("", config.UserHandler.GetAddressList)
			addresses.POST("", config.UserHandler.CreateAddress)
			addresses.GET("/:id", config.UserHandler.GetAddress)
			addresses.PUT("/:id", config.UserHandler.UpdateAddress)
			addresses.DELETE("/:id", config.UserHandler.DeleteAddress)
			addresses.PUT("/:id/default", config.UserHandler.SetDefaultAddress)
		}
	}

	// 购物车相关
	cart := authenticated.Group("/cart")
	{
		cart.GET("", config.CartHandler.GetCart)
		cart.POST("/items", config.CartHandler.AddItem)
		cart.PUT("/items", config.CartHandler.UpdateItemQuantity)
		cart.DELETE("/items/:productId", config.CartHandler.RemoveItem)
		cart.DELETE("", config.CartHandler.ClearCart)
	}

	// 订单相关
	orders := authenticated.Group("/orders")
	{
		orders.POST("", config.OrderHandler.CreateOrder)
		orders.GET("", config.OrderHandler.GetOrderList)
		orders.GET("/:id", config.OrderHandler.GetOrderDetail)
		orders.POST("/:id/pay", config.OrderHandler.PayOrder)
		orders.POST("/:id/cancel", config.OrderHandler.CancelOrder)
		orders.POST("/:id/complete", config.OrderHandler.CompleteOrder)

		// 支付相关
		orders.POST("/:id/payment", config.OrderHandler.CreatePayment)
		orders.GET("/:id/payment-status", config.OrderHandler.CheckPaymentStatus)
		orders.POST("/:id/payment-callback", config.OrderHandler.ProcessPaymentCallback)
		orders.POST("/:id/simulate-payment", config.OrderHandler.SimulatePaymentCallback)
	}
}

// setupAdminRoutes 配置管理员路由
func setupAdminRoutes(rg *gin.RouterGroup, config *RouterConfig) {
	// 应用认证、管理员权限和审计日志中间件
	admin := rg.Group("/admin")
	admin.Use(middleware.AuthMiddleware(config.JWTManager))
	admin.Use(middleware.AdminMiddleware())
	admin.Use(middleware.AdminAuditMiddleware())

	// 商品管理
	products := admin.Group("/products")
	{
		products.POST("", config.AdminHandler.CreateProduct)
		products.PUT("/:id", config.AdminHandler.UpdateProduct)
		products.DELETE("/:id", config.AdminHandler.DeleteProduct)
		products.PUT("/:id/status", config.AdminHandler.UpdateProductStatus)
	}

	// 订单管理
	orders := admin.Group("/orders")
	{
		orders.GET("", config.AdminHandler.GetAllOrders)
		orders.PUT("/:id/status", config.AdminHandler.UpdateOrderStatus)
		orders.POST("/:id/ship", config.AdminHandler.ShipOrder)
	}

	// 用户管理
	users := admin.Group("/users")
	{
		users.GET("", config.AdminHandler.GetUserList)
		users.DELETE("/:id", config.AdminHandler.DeleteUser)
	}

	// 统计信息
	admin.GET("/statistics", config.AdminHandler.GetOrderStatistics)
}

// setupAdminV2Routes 配置管理员V2路由（新功能）
func setupAdminV2Routes(rg *gin.RouterGroup, config *RouterConfig) {
	// 应用认证、管理员权限和审计日志中间件
	admin := rg.Group("/admin")
	admin.Use(middleware.AuthMiddleware(config.JWTManager))
	admin.Use(middleware.AdminMiddleware())
	admin.Use(middleware.AdminAuditMiddleware())

	// 商品SKU管理
	products := admin.Group("/products")
	{
		// SKU管理
		products.POST("/:id/specs", config.AdminHandler.CreateProductSpec)
		products.POST("/:id/skus", config.AdminHandler.CreateProductSKU)
		products.PUT("/:id/skus/:sku_id", config.AdminHandler.UpdateProductSKU)
	}

	// 优惠券管理
	coupons := admin.Group("/coupons")
	{
		// 需求：2.1 - 创建优惠券
		coupons.POST("", config.CouponHandler.CreateCoupon)
		// 需求：2.9 - 获取优惠券列表
		coupons.GET("", config.CouponHandler.GetCouponList)
	}

	// 退款管理
	refunds := admin.Group("/refunds")
	{
		// 需求：3.5 - 获取所有退款申请
		refunds.GET("", config.RefundHandler.GetAllRefunds)
		// 需求：3.6 - 审核通过退款申请
		refunds.PUT("/:id/approve", config.RefundHandler.ApproveRefund)
		// 需求：3.7 - 审核拒绝退款申请
		refunds.PUT("/:id/reject", config.RefundHandler.RejectRefund)
	}

	// 评价管理
	reviews := admin.Group("/reviews")
	{
		// 获取所有评价列表
		reviews.GET("", config.ReviewHandler.GetAllReviews)
		// 需求：4.9 - 隐藏评价
		reviews.PUT("/:id/hide", config.ReviewHandler.HideReview)
		// 显示评价（可选）
		reviews.PUT("/:id/show", config.ReviewHandler.ShowReview)
	}

	// 图片管理
	images := admin.Group("/images")
	{
		// 需求：5.7 - 获取图片列表
		images.GET("", config.UploadHandler.ListImages)
		// 需求：5.8 - 删除图片
		images.DELETE("/:id", config.UploadHandler.DeleteImage)
	}

	// 数据统计
	stats := admin.Group("/stats")
	{
		// 需求：10.1 - 数据概览
		stats.GET("/overview", config.StatisticsHandler.GetSalesOverview)
		// 需求：10.2 - 销售趋势
		stats.GET("/sales", config.StatisticsHandler.GetSalesTrend)
		// 需求：10.3 - 商品销售排行
		stats.GET("/products/ranking", config.StatisticsHandler.GetProductRanking)
		// 需求：10.4 - 用户增长趋势
		stats.GET("/users/growth", config.StatisticsHandler.GetUserGrowth)
		// 需求：10.7 - 导出数据
		stats.GET("/export", config.StatisticsHandler.ExportData)
	}

	// 操作日志
	logs := admin.Group("/logs")
	{
		// 需求：11.4, 11.5 - 获取操作日志列表
		logs.GET("", config.OperationLogHandler.GetOperationLogs)
		// 需求：11.4 - 获取日志详情
		logs.GET("/:id", config.OperationLogHandler.GetOperationLogDetail)
	}
}

// setupPublicV2Routes 配置公开V2路由（新功能）
func setupPublicV2Routes(rg *gin.RouterGroup, config *RouterConfig) {
	// 短信验证码相关（无需登录）
	sms := rg.Group("/sms")
	{
		// 需求：9.2, 9.3, 9.4 - 发送验证码
		sms.POST("/send", config.SMSHandler.SendVerificationCode)
	}

	// 认证相关（无需登录）
	auth := rg.Group("/auth")
	{
		// 需求：9.5, 9.6, 9.7 - 短信验证码登录
		auth.POST("/login/sms", config.SMSHandler.LoginWithSMS)
	}

	// 商品SKU相关（浏览商品SKU无需登录）
	products := rg.Group("/products")
	{
		// 获取商品所有SKU
		products.GET("/:id/skus", config.SKUHandler.GetProductSKUs)
		// 获取指定SKU详情
		products.GET("/:id/skus/:sku_id", config.SKUHandler.GetProductSKUDetail)
		// 需求：4.5 - 获取商品评价列表
		products.GET("/:id/reviews", config.ReviewHandler.GetProductReviews)
		// 获取商品评分统计（可选）
		products.GET("/:id/rating-stats", config.ReviewHandler.GetProductRatingStats)
	}

	// 优惠券相关（浏览优惠券无需登录）
	coupons := rg.Group("/coupons")
	{
		// 需求：2.2 - 获取可领取优惠券列表
		coupons.GET("", config.CouponHandler.GetAvailableCoupons)
	}

	// 需要认证的用户路由
	authenticated := rg.Group("")
	authenticated.Use(middleware.AuthMiddleware(config.JWTManager))
	{
		// 图片上传相关（需要登录）
		upload := authenticated.Group("/upload")
		{
			// 需求：5.3, 5.4, 5.5 - 上传单张图片
			upload.POST("/image", config.UploadHandler.UploadImage)
			// 需求：5.3, 5.4, 5.5 - 批量上传图片
			upload.POST("/images", config.UploadHandler.UploadImages)
		}

		// 优惠券相关（需要登录）
		coupons := authenticated.Group("/coupons")
		{
			// 需求：2.3 - 领取优惠券
			coupons.POST("/:id/claim", config.CouponHandler.ClaimCoupon)
		}

		// 用户优惠券相关
		user := authenticated.Group("/user")
		{
			// 需求：2.4 - 获取用户优惠券列表
			user.GET("/coupons", config.CouponHandler.GetUserCoupons)
			// 需求：2.4 - 获取可用优惠券（用于结算页）
			user.GET("/coupons/available", config.CouponHandler.GetAvailableUserCoupons)
			// 需求：3.3 - 获取用户退款列表
			user.GET("/refunds", config.RefundHandler.GetUserRefunds)
			// 需求：7.4 - 获取收藏列表
			user.GET("/favorites", config.FavoriteHandler.GetUserFavorites)
			// 需求：8.5 - 获取通知列表
			user.GET("/notifications", config.NotificationHandler.GetUserNotifications)
			// 需求：8.4 - 获取未读通知数量
			user.GET("/notifications/unread-count", config.NotificationHandler.GetUnreadCount)
			// 需求：8.6 - 标记通知为已读
			user.PUT("/notifications/:id/read", config.NotificationHandler.MarkNotificationAsRead)
			// 需求：9.8 - 绑定手机号
			user.POST("/phone/bind", config.SMSHandler.BindPhone)
		}

		// 订单相关
		orders := authenticated.Group("/orders")
		{
			// 需求：3.2 - 申请退款
			orders.POST("/:id/refunds", config.RefundHandler.CreateRefund)
			// 需求：4.2 - 提交评价
			orders.POST("/:id/reviews", config.ReviewHandler.CreateReview)
		}

		// 退款详情（可选）
		refunds := authenticated.Group("/refunds")
		{
			refunds.GET("/:id", config.RefundHandler.GetRefundDetail)
		}

		// 评价相关
		reviews := authenticated.Group("/reviews")
		{
			// 需求：4.7 - 点赞评价
			reviews.POST("/:id/like", config.ReviewHandler.LikeReview)
			// 获取评价详情（可选）
			reviews.GET("/:id", config.ReviewHandler.GetReviewDetail)
		}

		// 收藏相关
		favorites := authenticated.Group("/favorites")
		{
			// 需求：7.2 - 添加收藏
			favorites.POST("", config.FavoriteHandler.AddFavorite)
			// 需求：7.3 - 取消收藏
			favorites.DELETE("/:product_id", config.FavoriteHandler.RemoveFavorite)
		}
	}

	// 推荐相关（部分需要登录，部分无需登录）
	recommendations := rg.Group("/recommendations")
	{
		// 需求：12.4 - 获取热门商品推荐（无需登录）
		recommendations.GET("/hot", config.RecommendationHandler.GetHotProducts)
		// 需求：12.3 - 获取相似商品推荐（无需登录）
		recommendations.GET("/similar/:product_id", config.RecommendationHandler.GetSimilarProducts)
	}

	// 个性化推荐（需要登录）
	authenticatedRecommendations := rg.Group("/recommendations")
	authenticatedRecommendations.Use(middleware.AuthMiddleware(config.JWTManager))
	{
		// 需求：12.2 - 获取个性化推荐
		authenticatedRecommendations.GET("/personal", config.RecommendationHandler.GetPersonalRecommendations)
	}

	// 用户行为记录（需要登录）
	userBehaviors := rg.Group("/user")
	userBehaviors.Use(middleware.AuthMiddleware(config.JWTManager))
	{
		// 需求：12.1 - 记录用户浏览行为
		userBehaviors.POST("/behaviors", config.RecommendationHandler.RecordBehavior)
	}
}
