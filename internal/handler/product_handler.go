package handler

import (
	"strconv"

	"go-shop/internal/dto"
	"go-shop/internal/service"
	"go-shop/pkg/response"

	"github.com/gin-gonic/gin"
)

type ProductHandler struct {
	productService *service.ProductService
}

func NewProductHandler(productService *service.ProductService) *ProductHandler {
	return &ProductHandler{
		productService: productService,
	}
}

func (h *ProductHandler) GetProductList(c *gin.Context) {
	var req dto.ProductListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.InvalidParam(c, err.Error())
		return
	}

	products, err := h.productService.GetProductList(c.Request.Context(), &req)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, products)
}

func (h *ProductHandler) GetProductDetail(c *gin.Context) {
	productID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParam(c, "无效的商品ID")
		return
	}

	product, err := h.productService.GetProductByID(c.Request.Context(), productID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, product)
}

func (h *ProductHandler) GetProductsByCategory(c *gin.Context) {
	categoryID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParam(c, "无效的分类ID")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	products, err := h.productService.GetProductsByCategory(c.Request.Context(), categoryID, page, pageSize)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, products)
}

func (h *ProductHandler) SearchProducts(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		response.InvalidParam(c, "搜索关键词不能为空")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	products, err := h.productService.SearchProducts(c.Request.Context(), keyword, page, pageSize)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, products)
}
