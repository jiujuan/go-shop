package entity

import "time"

// Address 收货地址实体
type Address struct {
	ID            int64     `json:"id" gorm:"primaryKey"`
	UserID        int64     `json:"user_id"`
	RecipientName string    `json:"recipient_name" gorm:"size:50"`
	Phone         string    `json:"phone" gorm:"size:20"`
	Province      string    `json:"province" gorm:"size:50"`
	City          string    `json:"city" gorm:"size:50"`
	District      string    `json:"district" gorm:"size:50"`
	Detail        string    `json:"detail" gorm:"size:200"`
	IsDefault     bool      `json:"is_default" gorm:"default:false"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// GetFullAddress 获取完整地址字符串
func (a *Address) GetFullAddress() string {
	return a.Province + a.City + a.District + a.Detail
}

// Validate 验证地址信息完整性
func (a *Address) Validate() error {
	if a.RecipientName == "" {
		return ErrRecipientNameRequired
	}
	if a.Phone == "" {
		return ErrPhoneRequired
	}
	if a.Province == "" {
		return ErrProvinceRequired
	}
	if a.City == "" {
		return ErrCityRequired
	}
	if a.District == "" {
		return ErrDistrictRequired
	}
	if a.Detail == "" {
		return ErrDetailRequired
	}
	return nil
}