package utils

import (
	"log"

	"go-shop/internal/dto"
	"go-shop/pkg/validator"
)

// TestDTOAndValidation 测试DTO和验证功能
func TestDTOAndValidation() {
	log.Println("Testing DTO and Validation utilities...")

	// 初始化验证器
	v := validator.NewCustomValidator()

	// 测试用户注册请求DTO
	userReq := dto.UserRegisterRequest{
		Username: "testuser",
		Password: "password123",
		Email:    "test@example.com",
	}

	if err := v.ValidateStruct(userReq); err != nil {
		log.Printf("User registration validation failed: %v", err)
	} else {
		log.Println("✓ User registration DTO validation passed")
	}

	// 测试无效的用户注册请求
	invalidUserReq := dto.UserRegisterRequest{
		Username: "ab", // 太短
		Password: "password", // 没有数字
		Email:    "invalid-email", // 无效邮箱
	}

	if err := v.ValidateStruct(invalidUserReq); err != nil {
		log.Printf("✓ Invalid user registration correctly rejected: %v", err)
	} else {
		log.Println("✗ Invalid user registration should have been rejected")
	}

	// 测试地址创建请求DTO
	addressReq := dto.AddressCreateRequest{
		RecipientName: "张三",
		Phone:         "13800138000",
		Province:      "北京市",
		City:          "北京市",
		District:      "朝阳区",
		Detail:        "某某街道某某号",
	}

	if err := v.ValidateStruct(addressReq); err != nil {
		log.Printf("Address creation validation failed: %v", err)
	} else {
		log.Println("✓ Address creation DTO validation passed")
	}

	// 测试分页功能
	pagination := dto.PaginationRequest{Page: 2, PageSize: 10}
	page, pageSize := pagination.GetDefaultPagination()
	offset := pagination.GetOffset()
	limit := pagination.GetLimit()

	log.Printf("✓ Pagination: Page=%d, PageSize=%d, Offset=%d, Limit=%d", 
		page, pageSize, offset, limit)

	// 测试分页响应
	paginationResp := dto.NewPaginationResponse(1, 20, 105)
	log.Printf("✓ Pagination Response: Page=%d, PageSize=%d, Total=%d, TotalPages=%d",
		paginationResp.Page, paginationResp.PageSize, paginationResp.Total, paginationResp.TotalPages)

	// 测试订单状态映射
	statusText := dto.GetOrderStatusText(1)
	log.Printf("✓ Order status mapping: Status 1 = %s", statusText)

	log.Println("All DTO and validation tests passed!")
}