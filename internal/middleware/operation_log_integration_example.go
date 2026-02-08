package middleware

import (
	"go-shop/internal/repository"

	"github.com/gin-gonic/gin"
)

// Example: 如何在路由中使用操作日志中间件
//
// 在 router.go 中的 SetupRouter 函数中添加：
//
// func SetupRouter(config *RouterConfig, operationLogRepo repository.OperationLogRepository) *gin.Engine {
//     router := gin.New()
//
//     // 全局中间件
//     router.Use(middleware.LoggerMiddleware())
//     router.Use(middleware.RecoveryMiddleware())
//     router.Use(middleware.CORSMiddleware())
//     router.Use(middleware.RateLimitMiddleware(100, 100))
//     router.Use(middleware.OperationLogMiddleware(operationLogRepo)) // 添加操作日志中间件
//
//     // ... 其他路由配置
//
//     return router
// }
//
// 或者只在特定路由组中使用：
//
// func setupAdminRoutes(rg *gin.RouterGroup, config *RouterConfig, operationLogRepo repository.OperationLogRepository) {
//     admin := rg.Group("/admin")
//     admin.Use(middleware.AuthMiddleware(config.JWTManager))
//     admin.Use(middleware.AdminMiddleware())
//     admin.Use(middleware.OperationLogMiddleware(operationLogRepo)) // 只记录管理员操作
//
//     // ... 管理员路由配置
// }

// ExampleOperationLogMiddleware 示例：如何使用操作日志中间件
func ExampleOperationLogMiddleware(operationLogRepo repository.OperationLogRepository) *gin.Engine {
	router := gin.New()

	// 全局使用操作日志中间件
	router.Use(OperationLogMiddleware(operationLogRepo))

	// 示例路由
	router.GET("/api/v1/products", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	router.POST("/api/v1/orders", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "order created"})
	})

	return router
}

// 中间件特性说明：
//
// 1. 异步日志保存：日志保存操作在goroutine中异步执行，不阻塞主请求流程
//
// 2. 敏感信息脱敏：自动脱敏以下字段：
//    - password, old_password, new_password, confirm_password
//    - token, access_token, refresh_token
//    - secret, api_key, private_key
//    - credit_card, card_number, cvv
//    - ssn, id_card, bank_account
//
// 3. 自动识别操作类型和模块：
//    - 根据HTTP方法和路径自动确定操作类型（login, create_xxx, update_xxx等）
//    - 自动提取模块名称（products, orders, users等）
//
// 4. 用户信息记录：
//    - 自动从上下文中提取用户ID和用户类型（user/admin）
//    - 记录IP地址、User Agent等请求信息
//
// 5. 性能优化：
//    - 请求体和响应体大小限制为10KB
//    - 异步保存，不影响主流程性能
//
// 6. 错误处理：
//    - 使用defer捕获panic，确保日志失败不影响主流程
//    - 日志保存失败时记录错误日志，但不中断请求处理
//
// 7. 请求耗时统计：
//    - 自动记录每个请求的处理时间（毫秒）
//
// 8. 错误信息记录：
//    - 自动记录请求处理过程中的错误信息
//    - 记录HTTP状态码
