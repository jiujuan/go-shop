package mq

import (
	"context"
	"fmt"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
)

// MessageHandler 消息处理函数类型
type MessageHandler func(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error)

// Consumer RocketMQ 消费者封装
type Consumer struct {
	consumer rocketmq.PushConsumer
	config   *Config
}

// NewConsumer 创建消费者实例
func NewConsumer(config *Config, groupName string) (*Consumer, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// 如果指定了 groupName，使用指定的，否则使用配置中的
	if groupName == "" {
		groupName = config.GroupName
	}

	// 创建消费者
	c, err := rocketmq.NewPushConsumer(
		consumer.WithNameServer(config.NameServers),
		consumer.WithGroupName(groupName),
		consumer.WithConsumerModel(consumer.Clustering), // 集群模式
		consumer.WithConsumeFromWhere(consumer.ConsumeFromFirstOffset),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer: %w", err)
	}

	return &Consumer{
		consumer: c,
		config:   config,
	}, nil
}

// Subscribe 订阅 Topic
func (c *Consumer) Subscribe(topic string, selector consumer.MessageSelector, handler MessageHandler) error {
	// 订阅消息
	err := c.consumer.Subscribe(topic, selector, func(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
		return handler(ctx, msgs...)
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe topic %s: %w", topic, err)
	}

	return nil
}

// SubscribeWithTag 订阅指定 Tag 的消息
func (c *Consumer) SubscribeWithTag(topic, tag string, handler MessageHandler) error {
	selector := consumer.MessageSelector{
		Type:       consumer.TAG,
		Expression: tag,
	}

	return c.Subscribe(topic, selector, handler)
}

// SubscribeAll 订阅 Topic 的所有消息
func (c *Consumer) SubscribeAll(topic string, handler MessageHandler) error {
	selector := consumer.MessageSelector{
		Type:       consumer.TAG,
		Expression: "*", // 订阅所有 Tag
	}

	return c.Subscribe(topic, selector, handler)
}

// Start 启动消费者
func (c *Consumer) Start() error {
	if err := c.consumer.Start(); err != nil {
		return fmt.Errorf("failed to start consumer: %w", err)
	}
	return nil
}

// Shutdown 关闭消费者
func (c *Consumer) Shutdown() error {
	if err := c.consumer.Shutdown(); err != nil {
		return fmt.Errorf("failed to shutdown consumer: %w", err)
	}
	return nil
}

// ConsumeSuccess 消费成功
func ConsumeSuccess() consumer.ConsumeResult {
	return consumer.ConsumeSuccess
}

// ConsumeRetryLater 稍后重试
func ConsumeRetryLater() consumer.ConsumeResult {
	return consumer.ConsumeRetryLater
}
