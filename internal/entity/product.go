package entity

import "time"

// Product 商品实体
type Product struct {
	ID          int64     `json:"id" gorm:"primaryKey"`
	CategoryID  int64     `json:"category_id"`
	Name        string    `json:"name" gorm:"size:200"`
	Description string    `json:"description" gorm:"type:text"`
	Price       int64     `json:"price"` // 单位：分
	Stock       int       `json:"stock"`
	CoverImage  string    `json:"cover_image" gorm:"size:500"`
	Status      int       `json:"status"` // 1-上架，0-下架
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}