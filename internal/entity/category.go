package entity

import "time"

// Category 商品分类实体
type Category struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"size:50"`
	ParentID  int64     `json:"parent_id" gorm:"default:0"`
	SortOrder int       `json:"sort_order" gorm:"default:0"`
	CreatedAt time.Time `json:"created_at"`
}

// IsTopLevel 判断是否为顶级分类
func (c *Category) IsTopLevel() bool {
	return c.ParentID == 0
}