package dto

import (
	"testing"

	"go-shop/pkg/validator"
)

func TestUserRegisterRequest_Validation(t *testing.T) {
	v := validator.NewCustomValidator()

	tests := []struct {
		name    string
		request UserRegisterRequest
		wantErr bool
	}{
		{
			name: "valid request",
			request: UserRegisterRequest{
				Username: "testuser",
				Password: "password123",
				Email:    "test@example.com",
			},
			wantErr: false,
		},
		{
			name: "invalid username - too short",
			request: UserRegisterRequest{
				Username: "ab",
				Password: "password123",
				Email:    "test@example.com",
			},
			wantErr: true,
		},
		{
			name: "invalid password - no number",
			request: UserRegisterRequest{
				Username: "testuser",
				Password: "password",
				Email:    "test@example.com",
			},
			wantErr: true,
		},
		{
			name: "invalid email",
			request: UserRegisterRequest{
				Username: "testuser",
				Password: "password123",
				Email:    "invalid-email",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateStruct(tt.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateStruct() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAddressCreateRequest_Validation(t *testing.T) {
	v := validator.NewCustomValidator()

	tests := []struct {
		name    string
		request AddressCreateRequest
		wantErr bool
	}{
		{
			name: "valid request",
			request: AddressCreateRequest{
				RecipientName: "张三",
				Phone:         "13800138000",
				Province:      "北京市",
				City:          "北京市",
				District:      "朝阳区",
				Detail:        "某某街道某某号",
			},
			wantErr: false,
		},
		{
			name: "invalid phone",
			request: AddressCreateRequest{
				RecipientName: "张三",
				Phone:         "12345",
				Province:      "北京市",
				City:          "北京市",
				District:      "朝阳区",
				Detail:        "某某街道某某号",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateStruct(tt.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateStruct() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPaginationRequest_GetDefaultPagination(t *testing.T) {
	tests := []struct {
		name         string
		request      PaginationRequest
		expectedPage int
		expectedSize int
	}{
		{
			name:         "default values",
			request:      PaginationRequest{},
			expectedPage: 1,
			expectedSize: 20,
		},
		{
			name:         "custom values",
			request:      PaginationRequest{Page: 2, PageSize: 10},
			expectedPage: 2,
			expectedSize: 10,
		},
		{
			name:         "invalid page",
			request:      PaginationRequest{Page: 0, PageSize: 10},
			expectedPage: 1,
			expectedSize: 10,
		},
		{
			name:         "invalid page size",
			request:      PaginationRequest{Page: 1, PageSize: 0},
			expectedPage: 1,
			expectedSize: 20,
		},
		{
			name:         "page size too large",
			request:      PaginationRequest{Page: 1, PageSize: 200},
			expectedPage: 1,
			expectedSize: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page, size := tt.request.GetDefaultPagination()
			if page != tt.expectedPage {
				t.Errorf("GetDefaultPagination() page = %v, want %v", page, tt.expectedPage)
			}
			if size != tt.expectedSize {
				t.Errorf("GetDefaultPagination() size = %v, want %v", size, tt.expectedSize)
			}
		})
	}
}

func TestNewPaginationResponse(t *testing.T) {
	tests := []struct {
		name           string
		page           int
		pageSize       int
		total          int64
		expectedPages  int
	}{
		{
			name:          "exact division",
			page:          1,
			pageSize:      10,
			total:         100,
			expectedPages: 10,
		},
		{
			name:          "with remainder",
			page:          1,
			pageSize:      10,
			total:         105,
			expectedPages: 11,
		},
		{
			name:          "zero total",
			page:          1,
			pageSize:      10,
			total:         0,
			expectedPages: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := NewPaginationResponse(tt.page, tt.pageSize, tt.total)
			if resp.TotalPages != tt.expectedPages {
				t.Errorf("NewPaginationResponse() TotalPages = %v, want %v", resp.TotalPages, tt.expectedPages)
			}
			if resp.Page != tt.page {
				t.Errorf("NewPaginationResponse() Page = %v, want %v", resp.Page, tt.page)
			}
			if resp.PageSize != tt.pageSize {
				t.Errorf("NewPaginationResponse() PageSize = %v, want %v", resp.PageSize, tt.pageSize)
			}
			if resp.Total != tt.total {
				t.Errorf("NewPaginationResponse() Total = %v, want %v", resp.Total, tt.total)
			}
		})
	}
}