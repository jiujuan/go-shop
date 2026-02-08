package mq

import (
	"context"
	"fmt"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
)

// Producer RocketMQ 生产者封装
type Producer struct {
	producer rocketmq.Producer
	config   *Config
}

// NewProducer 创建生产者实例
func NewProducer(config *Config) (*Producer, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// 创建生产者
	p, err := rocketmq.NewProducer(
		producer.WithNameServer(config.NameServers),
		producer.WithRetry(config.Retry),
		producer.WithGroupName(config.GroupName),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}

	// 启动生产者
	if err := p.Start(); err != nil {
		return nil, fmt.Errorf("failed to start producer: %w", err)
	}

	return &Producer{
		producer: p,
		config:   config,
	}, nil
}

// SendMessage 发送消息
func (p *Producer) SendMessage(ctx context.Context, topic, tag string, body []byte) (*primitive.SendResult, error) {
	msg := &primitive.Message{
		Topic: topic,
		Body:  body,
	}

	// 设置标签
	if tag != "" {
		msg.WithTag(tag)
	}

	// 发送消息
	result, err := p.producer.SendSync(ctx, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	return result, nil
}

// SendMessageWithKeys 发送带 Keys 的消息（用于消息追踪）
func (p *Producer) SendMessageWithKeys(ctx context.Context, topic, tag string, keys []string, body []byte) (*primitive.SendResult, error) {
	msg := &primitive.Message{
		Topic: topic,
		Body:  body,
	}

	// 设置标签
	if tag != "" {
		msg.WithTag(tag)
	}

	// 设置 Keys
	if len(keys) > 0 {
		msg.WithKeys(keys)
	}

	// 发送消息
	result, err := p.producer.SendSync(ctx, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	return result, nil
}

// SendAsyncMessage 异步发送消息
func (p *Producer) SendAsyncMessage(ctx context.Context, topic, tag string, body []byte, callback func(context.Context, *primitive.SendResult, error)) error {
	msg := &primitive.Message{
		Topic: topic,
		Body:  body,
	}

	// 设置标签
	if tag != "" {
		msg.WithTag(tag)
	}

	// 异步发送消息
	err := p.producer.SendAsync(ctx, callback, msg)
	if err != nil {
		return fmt.Errorf("failed to send async message: %w", err)
	}

	return nil
}

// SendDelayMessage 发送延迟消息
func (p *Producer) SendDelayMessage(ctx context.Context, topic, tag string, body []byte, delayLevel int) (*primitive.SendResult, error) {
	msg := &primitive.Message{
		Topic: topic,
		Body:  body,
	}

	// 设置标签
	if tag != "" {
		msg.WithTag(tag)
	}

	// 设置延迟级别
	// 1s 5s 10s 30s 1m 2m 3m 4m 5m 6m 7m 8m 9m 10m 20m 30m 1h 2h
	msg.WithDelayTimeLevel(delayLevel)

	// 发送消息
	result, err := p.producer.SendSync(ctx, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to send delay message: %w", err)
	}

	return result, nil
}

// Shutdown 关闭生产者
func (p *Producer) Shutdown() error {
	if err := p.producer.Shutdown(); err != nil {
		return fmt.Errorf("failed to shutdown producer: %w", err)
	}
	return nil
}
