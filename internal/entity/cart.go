package entity

// CartItem 购物车商品项实体
type CartItem struct {
	UserID     int64       `json:"user_id"`
	ProductID  int64       `json:"product_id"`
	SKUID      *int64      `json:"sku_id,omitempty"`      // SKU ID（可选）
	SKUCode    *string     `json:"sku_code,omitempty"`    // SKU编码（可选）
	SpecValues *SpecValues `json:"spec_values,omitempty"` // 规格组合（可选）
	Quantity   int         `json:"quantity"`
	Product    *Product    `json:"product,omitempty"`
	SKU        *ProductSKU `json:"sku,omitempty"` // SKU详情（可选）
}

// Cart 购物车实体
type Cart struct {
	UserID     int64      `json:"user_id"`
	Items      []CartItem `json:"items"`
	TotalCount int        `json:"total_count"`
	TotalPrice int64      `json:"total_price"`
}

// AddItem 添加商品到购物车
func (c *Cart) AddItem(item CartItem) {
	// 查找是否已存在相同商品和SKU
	for i, existingItem := range c.Items {
		// 如果有SKU，需要匹配ProductID和SKUID
		if item.SKUID != nil && existingItem.SKUID != nil {
			if existingItem.ProductID == item.ProductID && *existingItem.SKUID == *item.SKUID {
				c.Items[i].Quantity += item.Quantity
				c.calculateTotals()
				return
			}
		} else if item.SKUID == nil && existingItem.SKUID == nil {
			// 如果没有SKU，只匹配ProductID
			if existingItem.ProductID == item.ProductID {
				c.Items[i].Quantity += item.Quantity
				c.calculateTotals()
				return
			}
		}
	}
	
	// 添加新商品
	c.Items = append(c.Items, item)
	c.calculateTotals()
}

// UpdateItemQuantity 更新商品数量
func (c *Cart) UpdateItemQuantity(productID int64, skuID *int64, quantity int) {
	for i, item := range c.Items {
		// 匹配商品ID和SKU ID
		if item.ProductID == productID {
			// 如果指定了SKU ID，需要匹配SKU
			if skuID != nil && item.SKUID != nil && *item.SKUID == *skuID {
				if quantity <= 0 {
					// 删除商品
					c.Items = append(c.Items[:i], c.Items[i+1:]...)
				} else {
					c.Items[i].Quantity = quantity
				}
				c.calculateTotals()
				return
			} else if skuID == nil && item.SKUID == nil {
				// 没有SKU的情况
				if quantity <= 0 {
					// 删除商品
					c.Items = append(c.Items[:i], c.Items[i+1:]...)
				} else {
					c.Items[i].Quantity = quantity
				}
				c.calculateTotals()
				return
			}
		}
	}
}

// RemoveItem 删除商品
func (c *Cart) RemoveItem(productID int64, skuID *int64) {
	for i, item := range c.Items {
		if item.ProductID == productID {
			// 如果指定了SKU ID，需要匹配SKU
			if skuID != nil && item.SKUID != nil && *item.SKUID == *skuID {
				c.Items = append(c.Items[:i], c.Items[i+1:]...)
				c.calculateTotals()
				return
			} else if skuID == nil && item.SKUID == nil {
				// 没有SKU的情况
				c.Items = append(c.Items[:i], c.Items[i+1:]...)
				c.calculateTotals()
				return
			}
		}
	}
}

// Clear 清空购物车
func (c *Cart) Clear() {
	c.Items = []CartItem{}
	c.TotalCount = 0
	c.TotalPrice = 0
}

// calculateTotals 计算总数量和总价格
func (c *Cart) calculateTotals() {
	c.TotalCount = 0
	c.TotalPrice = 0
	
	for _, item := range c.Items {
		c.TotalCount += item.Quantity
		// 优先使用SKU价格，如果没有SKU则使用商品价格
		if item.SKU != nil {
			c.TotalPrice += int64(item.Quantity) * item.SKU.Price
		} else if item.Product != nil {
			c.TotalPrice += int64(item.Quantity) * item.Product.Price
		}
	}
}