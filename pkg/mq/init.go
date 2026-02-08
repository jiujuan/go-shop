package mq

import (
	"fmt"
	"time"
)

// InitProducer 初始化生产者
func InitProducer(nameServers []string, groupName string, retry int, timeout int) (*Producer, error) {
	config := &Config{
		NameServers: nameServers,
		GroupName:   groupName,
		Retry:       retry,
		Timeout:     time.Duration(timeout) * time.Second,
	}

	producer, err := NewProducer(config)
	if err != nil {
		return nil, fmt.Errorf("failed to init producer: %w", err)
	}

	return producer, nil
}

// InitConsumer 初始化消费者
func InitConsumer(nameServers []string, groupName string, retry int, timeout int) (*Consumer, error) {
	config := &Config{
		NameServers: nameServers,
		GroupName:   groupName,
		Retry:       retry,
		Timeout:     time.Duration(timeout) * time.Second,
	}

	consumer, err := NewConsumer(config, groupName)
	if err != nil {
		return nil, fmt.Errorf("failed to init consumer: %w", err)
	}

	return consumer, nil
}
