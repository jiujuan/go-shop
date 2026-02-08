package handler

import (
	"strconv"

	"go-shop/internal/dto"
	"go-shop/internal/middleware"
	"go-shop/internal/service"
	"go-shop/pkg/response"

	"github.com/gin-gonic/gin"
)

type CartHandler struct {
	cartService *service.CartService
}

func NewCartHandler(cartService *service.CartService) *CartHandler {
	return &CartHandler{
		cartService: cartService,
	}
}

func (h *CartHandler) GetCart(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		response.Unauthorized(c, "未认证")
		return
	}

	cart, err := h.cartService.GetCart(c.Request.Context(), userID)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, cart)
}

func (h *CartHandler) AddItem(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		response.Unauthorized(c, "未认证")
		return
	}

	var req dto.CartAddItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParam(c, err.Error())
		return
	}

	cart, err := h.cartService.AddItem(c.Request.Context(), userID, &req)
	if err != nil {
		response.InvalidParam(c, err.Error())
		return
	}

	response.Success(c, cart)
}

func (h *CartHandler) UpdateItemQuantity(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		response.Unauthorized(c, "未认证")
		return
	}

	var req dto.CartUpdateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParam(c, err.Error())
		return
	}

	cart, err := h.cartService.UpdateItemQuantity(c.Request.Context(), userID, &req)
	if err != nil {
		response.InvalidParam(c, err.Error())
		return
	}

	response.Success(c, cart)
}

func (h *CartHandler) RemoveItem(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		response.Unauthorized(c, "未认证")
		return
	}

	productID, err := strconv.ParseInt(c.Param("productId"), 10, 64)
	if err != nil {
		response.InvalidParam(c, "无效的商品ID")
		return
	}

	cart, err := h.cartService.RemoveItem(c.Request.Context(), userID, productID)
	if err != nil {
		response.InvalidParam(c, err.Error())
		return
	}

	response.Success(c, cart)
}

func (h *CartHandler) ClearCart(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		response.Unauthorized(c, "未认证")
		return
	}

	if err := h.cartService.ClearCart(c.Request.Context(), userID); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, nil)
}
