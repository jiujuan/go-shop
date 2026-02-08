package mq_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"go-shop/pkg/mq"

	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
)

// Example_producer 演示生产者使用
func Example_producer() {
	// 初始化生产者
	producer, err := mq.InitProducer(
		[]string{"localhost:9876"},
		"go-shop-producer-group",
		2,
		3,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer producer.Shutdown()

	// 准备消息数据
	data := map[string]interface{}{
		"order_id": 12345,
		"user_id":  67890,
		"status":   "paid",
	}
	body, _ := json.Marshal(data)

	// 发送消息
	result, err := producer.SendMessage(
		context.Background(),
		mq.TopicOrderEvents,
		mq.TagOrderPaid,
		body,
	)
	if err != nil {
		log.Printf("Failed to send message: %v", err)
		return
	}

	fmt.Printf("Message sent successfully: %s\n", result.MsgID)
}

// Example_consumer 演示消费者使用
func Example_consumer() {
	// 初始化消费者
	mqConsumer, err := mq.InitConsumer(
		[]string{"localhost:9876"},
		"go-shop-order-consumer",
		2,
		3,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer mqConsumer.Shutdown()

	// 定义消息处理函数
	handler := func(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
		for _, msg := range msgs {
			fmt.Printf("Received message: %s\n", string(msg.Body))

			// 解析消息
			var data map[string]interface{}
			if err := json.Unmarshal(msg.Body, &data); err != nil {
				log.Printf("Failed to unmarshal message: %v", err)
				return mq.ConsumeRetryLater(), nil
			}

			// 处理消息
			fmt.Printf("Processing order: %v\n", data["order_id"])
		}
		return mq.ConsumeSuccess(), nil
	}

	// 订阅消息
	err = mqConsumer.SubscribeWithTag(mq.TopicOrderEvents, mq.TagOrderPaid, handler)
	if err != nil {
		log.Fatal(err)
	}

	// 启动消费者
	err = mqConsumer.Start()
	if err != nil {
		log.Fatal(err)
	}

	// 保持运行
	time.Sleep(time.Hour)
}

// Example_asyncProducer 演示异步生产者使用
func Example_asyncProducer() {
	// 初始化生产者
	producer, err := mq.InitProducer(
		[]string{"localhost:9876"},
		"go-shop-producer-group",
		2,
		3,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer producer.Shutdown()

	// 准备消息数据
	data := map[string]interface{}{
		"order_id": 12345,
		"user_id":  67890,
		"status":   "created",
	}
	body, _ := json.Marshal(data)

	// 定义回调函数
	callback := func(ctx context.Context, result *primitive.SendResult, err error) {
		if err != nil {
			log.Printf("Failed to send message: %v", err)
			return
		}
		fmt.Printf("Message sent successfully: %s\n", result.MsgID)
	}

	// 异步发送消息
	err = producer.SendAsyncMessage(
		context.Background(),
		mq.TopicOrderEvents,
		mq.TagOrderCreated,
		body,
		callback,
	)
	if err != nil {
		log.Printf("Failed to send async message: %v", err)
	}

	// 等待异步发送完成
	time.Sleep(time.Second)
}
