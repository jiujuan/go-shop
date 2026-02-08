package handler

import (
	"go-shop/internal/dto"
	"go-shop/internal/middleware"
	"go-shop/internal/service"
	"go-shop/pkg/response"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	orderService   *service.OrderService
	paymentService *service.PaymentService
}

func NewOrderHandler(orderService *service.OrderService, paymentService *service.PaymentService) *OrderHandler {
	return &OrderHandler{
		orderService:   orderService,
		paymentService: paymentService,
	}
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		response.Unauthorized(c, "未认证")
		return
	}

	var req dto.OrderCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParam(c, err.Error())
		return
	}

	order, err := h.orderService.CreateOrder(c.Request.Context(), userID, &req)
	if err != nil {
		response.InvalidParam(c, err.Error())
		return
	}

	response.Success(c, order)
}

func (h *OrderHandler) GetOrderList(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		response.Unauthorized(c, "未认证")
		return
	}

	var req dto.OrderListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.InvalidParam(c, err.Error())
		return
	}

	orders, err := h.orderService.GetUserOrders(c.Request.Context(), userID, &req)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, orders)
}

func (h *OrderHandler) GetOrderDetail(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		response.Unauthorized(c, "未认证")
		return
	}

	orderID := c.Param("id")
	if orderID == "" {
		response.InvalidParam(c, "订单ID不能为空")
		return
	}

	order, err := h.orderService.GetOrderByID(c.Request.Context(), userID, orderID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, order)
}

func (h *OrderHandler) PayOrder(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		response.Unauthorized(c, "未认证")
		return
	}

	orderID := c.Param("id")
	if orderID == "" {
		response.InvalidParam(c, "订单ID不能为空")
		return
	}

	order, err := h.orderService.PayOrder(c.Request.Context(), userID, orderID)
	if err != nil {
		response.InvalidParam(c, err.Error())
		return
	}

	response.Success(c, order)
}

func (h *OrderHandler) CancelOrder(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		response.Unauthorized(c, "未认证")
		return
	}

	orderID := c.Param("id")
	if orderID == "" {
		response.InvalidParam(c, "订单ID不能为空")
		return
	}

	if err := h.orderService.CancelOrder(c.Request.Context(), userID, orderID); err != nil {
		response.InvalidParam(c, err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *OrderHandler) CompleteOrder(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		response.Unauthorized(c, "未认证")
		return
	}

	orderID := c.Param("id")
	if orderID == "" {
		response.InvalidParam(c, "订单ID不能为空")
		return
	}

	order, err := h.orderService.CompleteOrder(c.Request.Context(), userID, orderID)
	if err != nil {
		response.InvalidParam(c, err.Error())
		return
	}

	response.Success(c, order)
}

// CreatePayment 创建支付
func (h *OrderHandler) CreatePayment(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		response.Unauthorized(c, "未认证")
		return
	}

	var req dto.PaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParam(c, err.Error())
		return
	}

	payment, err := h.paymentService.CreatePayment(c.Request.Context(), userID, &req)
	if err != nil {
		response.InvalidParam(c, err.Error())
		return
	}

	response.Success(c, payment)
}

// ProcessPaymentCallback 处理支付回调
func (h *OrderHandler) ProcessPaymentCallback(c *gin.Context) {
	orderID := c.Param("id")
	if orderID == "" {
		response.InvalidParam(c, "订单ID不能为空")
		return
	}

	// 模拟支付回调处理
	err := h.paymentService.ProcessPayment(c.Request.Context(), orderID)
	if err != nil {
		response.InvalidParam(c, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "支付成功"})
}

// CheckPaymentStatus 查询支付状态
func (h *OrderHandler) CheckPaymentStatus(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		response.Unauthorized(c, "未认证")
		return
	}

	orderID := c.Param("id")
	if orderID == "" {
		response.InvalidParam(c, "订单ID不能为空")
		return
	}

	// 验证订单属于当前用户
	order, err := h.orderService.GetOrderByID(c.Request.Context(), userID, orderID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	status, err := h.paymentService.CheckPaymentStatus(c.Request.Context(), order.ID)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"order_id":       orderID,
		"payment_status": status,
	})
}

// GetPaymentMethods 获取支付方式列表
func (h *OrderHandler) GetPaymentMethods(c *gin.Context) {
	methods := h.paymentService.GetPaymentMethods()
	response.Success(c, gin.H{
		"payment_methods": methods,
	})
}

// SimulatePaymentCallback 模拟支付回调（测试用）
func (h *OrderHandler) SimulatePaymentCallback(c *gin.Context) {
	orderID := c.Param("id")
	if orderID == "" {
		response.InvalidParam(c, "订单ID不能为空")
		return
	}

	var req struct {
		Success bool `json:"success" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParam(c, err.Error())
		return
	}

	err := h.paymentService.SimulatePaymentCallback(c.Request.Context(), orderID, req.Success)
	if err != nil {
		response.InvalidParam(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"message": "支付回调处理成功",
		"success": req.Success,
	})
}
