package mq

import (
	"fmt"
	"time"
)

// Config RocketMQ 配置
type Config struct {
	NameServers []string      `yaml:"name_servers"` // NameServer 地址列表
	GroupName   string        `yaml:"group_name"`   // 消费者组名称
	Retry       int           `yaml:"retry"`        // 重试次数
	Timeout     time.Duration `yaml:"timeout"`      // 超时时间（秒）
}

// Validate 验证配置
func (c *Config) Validate() error {
	if len(c.NameServers) == 0 {
		return fmt.Errorf("name_servers is required")
	}

	if c.GroupName == "" {
		return fmt.Errorf("group_name is required")
	}

	if c.Retry <= 0 {
		c.Retry = 2 // 默认重试 2 次
	}

	if c.Timeout <= 0 {
		c.Timeout = 3 * time.Second // 默认超时 3 秒
	}

	return nil
}

// Topic 常量定义
const (
	// TopicOrderEvents 订单事件 Topic
	TopicOrderEvents = "order-events"

	// TopicInventoryEvents 库存事件 Topic
	TopicInventoryEvents = "inventory-events"

	// TopicNotificationEvents 通知事件 Topic
	TopicNotificationEvents = "notification-events"

	// TopicStatisticsEvents 统计事件 Topic
	TopicStatisticsEvents = "statistics-events"
)

// Tag 常量定义
const (
	// TagOrderCreated 订单创建标签
	TagOrderCreated = "order_created"

	// TagOrderPaid 订单支付标签
	TagOrderPaid = "order_paid"

	// TagOrderCancelled 订单取消标签
	TagOrderCancelled = "order_cancelled"

	// TagOrderCompleted 订单完成标签
	TagOrderCompleted = "order_completed"

	// TagInventoryDeducted 库存扣减标签
	TagInventoryDeducted = "inventory_deducted"

	// TagInventoryRestored 库存恢复标签
	TagInventoryRestored = "inventory_restored"

	// TagInventoryLowStock 低库存警告标签
	TagInventoryLowStock = "inventory_low_stock"

	// TagNotificationOrder 订单通知标签
	TagNotificationOrder = "notification_order"

	// TagNotificationPromotion 促销通知标签
	TagNotificationPromotion = "notification_promotion"

	// TagNotificationSystem 系统通知标签
	TagNotificationSystem = "notification_system"

	// TagStatisticsOrder 订单统计标签
	TagStatisticsOrder = "statistics_order"

	// TagStatisticsProduct 商品统计标签
	TagStatisticsProduct = "statistics_product"

	// TagStatisticsUser 用户统计标签
	TagStatisticsUser = "statistics_user"
)
