package handler

import (
	"strconv"

	"go-shop/internal/dto"
	"go-shop/internal/middleware"
	"go-shop/internal/service"
	"go-shop/pkg/response"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService    *service.UserService
	addressService service.AddressService
}

func NewUserHandler(userService *service.UserService, addressService service.AddressService) *UserHandler {
	return &UserHandler{
		userService:    userService,
		addressService: addressService,
	}
}

func (h *UserHandler) GetUserInfo(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		response.Unauthorized(c, "未认证")
		return
	}

	user, err := h.userService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, user)
}

func (h *UserHandler) UpdateUserInfo(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		response.Unauthorized(c, "未认证")
		return
	}

	var req dto.UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParam(c, err.Error())
		return
	}

	user, err := h.userService.UpdateUser(c.Request.Context(), userID, &req)
	if err != nil {
		response.InvalidParam(c, err.Error())
		return
	}

	response.Success(c, user)
}

func (h *UserHandler) GetAddressList(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		response.Unauthorized(c, "未认证")
		return
	}

	addresses, err := h.addressService.GetUserAddresses(c.Request.Context(), userID)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, addresses)
}

func (h *UserHandler) GetAddress(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		response.Unauthorized(c, "未认证")
		return
	}

	addressID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParam(c, "无效的地址ID")
		return
	}

	address, err := h.addressService.GetAddress(c.Request.Context(), userID, addressID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, address)
}

func (h *UserHandler) CreateAddress(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		response.Unauthorized(c, "未认证")
		return
	}

	var req dto.AddressCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParam(c, err.Error())
		return
	}

	address, err := h.addressService.CreateAddress(c.Request.Context(), userID, &req)
	if err != nil {
		response.InvalidParam(c, err.Error())
		return
	}

	response.Success(c, address)
}

func (h *UserHandler) UpdateAddress(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		response.Unauthorized(c, "未认证")
		return
	}

	addressID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParam(c, "无效的地址ID")
		return
	}

	var req dto.AddressUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParam(c, err.Error())
		return
	}

	address, err := h.addressService.UpdateAddress(c.Request.Context(), userID, addressID, &req)
	if err != nil {
		response.InvalidParam(c, err.Error())
		return
	}

	response.Success(c, address)
}

func (h *UserHandler) DeleteAddress(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		response.Unauthorized(c, "未认证")
		return
	}

	addressID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParam(c, "无效的地址ID")
		return
	}

	if err := h.addressService.DeleteAddress(c.Request.Context(), userID, addressID); err != nil {
		response.InvalidParam(c, err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *UserHandler) SetDefaultAddress(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		response.Unauthorized(c, "未认证")
		return
	}

	addressID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParam(c, "无效的地址ID")
		return
	}

	if err := h.addressService.SetDefaultAddress(c.Request.Context(), userID, addressID); err != nil {
		response.InvalidParam(c, err.Error())
		return
	}

	response.Success(c, nil)
}
