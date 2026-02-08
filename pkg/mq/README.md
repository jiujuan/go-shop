# RocketMQ 客户端封装

本包提供了 RocketMQ 生产者和消费者的基础封装，简化了消息队列的使用。

## 功能特性

- 生产者封装：支持同步、异步、延迟消息发送
- 消费者封装：支持 Topic 和 Tag 订阅
- 配置管理：统一的配置结构
- Topic 常量：预定义的 Topic 和 Tag 常量
- 消息工具：消息序列化和反序列化

## Topic 定义

系统定义了以下 Topic：

- `order-events`: 订单事件（订单创建、支付、取消、完成）
- `inventory-events`: 库存事件（库存扣减、恢复、低库存警告）
- `notification-events`: 通知事件（订单通知、促销通知、系统通知）
- `statistics-events`: 统计事件（订单统计、商品统计、用户统计）

## 使用示例

### 初始化生产者

```go
import "go-shop/pkg/mq"

// 从配置初始化
producer, err := mq.InitProducer(
    []string{"localhost:9876"},
    "go-shop-producer-group",
    2,  // 重试次数
    3,  // 超时时间（秒）
)
if err != nil {
    log.Fatal(err)
}
defer producer.Shutdown()
```

### 发送消息

```go
import (
    "context"
    "encoding/json"
    "go-shop/pkg/mq"
)

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
}
```

### 初始化消费者

```go
import "go-shop/pkg/mq"

// 从配置初始化
consumer, err := mq.InitConsumer(
    []string{"localhost:9876"},
    "go-shop-order-consumer",
    2,  // 重试次数
    3,  // 超时时间（秒）
)
if err != nil {
    log.Fatal(err)
}
defer consumer.Shutdown()
```

### 订阅消息

```go
import (
    "context"
    "log"
    "go-shop/pkg/mq"
    "github.com/apache/rocketmq-client-go/v2/consumer"
    "github.com/apache/rocketmq-client-go/v2/primitive"
)

// 定义消息处理函数
handler := func(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
    for _, msg := range msgs {
        log.Printf("Received message: %s", string(msg.Body))
        
        // 处理消息
        // ...
    }
    return mq.ConsumeSuccess(), nil
}

// 订阅指定 Tag 的消息
err = consumer.SubscribeWithTag(mq.TopicOrderEvents, mq.TagOrderPaid, handler)
if err != nil {
    log.Fatal(err)
}

// 启动消费者
err = consumer.Start()
if err != nil {
    log.Fatal(err)
}
```

### 订阅所有消息

```go
// 订阅 Topic 的所有消息
err = consumer.SubscribeAll(mq.TopicOrderEvents, handler)
if err != nil {
    log.Fatal(err)
}
```

### 异步发送消息

```go
import (
    "context"
    "log"
    "go-shop/pkg/mq"
    "github.com/apache/rocketmq-client-go/v2/primitive"
)

// 定义回调函数
callback := func(ctx context.Context, result *primitive.SendResult, err error) {
    if err != nil {
        log.Printf("Failed to send message: %v", err)
        return
    }
    log.Printf("Message sent successfully: %s", result.MsgID)
}

// 异步发送消息
err := producer.SendAsyncMessage(
    context.Background(),
    mq.TopicOrderEvents,
    mq.TagOrderCreated,
    body,
    callback,
)
```

### 发送延迟消息

```go
// 发送延迟消息（延迟级别 3 = 10 秒）
// 延迟级别: 1s 5s 10s 30s 1m 2m 3m 4m 5m 6m 7m 8m 9m 10m 20m 30m 1h 2h
result, err := producer.SendDelayMessage(
    context.Background(),
    mq.TopicOrderEvents,
    mq.TagOrderCancelled,
    body,
    3, // 延迟级别
)
```

## 配置说明

在 `config.yaml` 中配置 RocketMQ：

```yaml
mq:
  name_servers:
    - localhost:9876
  group_name: go-shop-consumer-group
  retry: 2
  timeout: 3 # 秒
```

## 注意事项

1. 生产者和消费者使用完毕后需要调用 `Shutdown()` 方法释放资源
2. 消费者需要先订阅 Topic，然后调用 `Start()` 启动
3. 消息处理函数应该返回 `ConsumeSuccess()` 或 `ConsumeRetryLater()`
4. 建议使用 Tag 来区分不同类型的消息
5. 延迟消息的延迟级别是固定的，不支持自定义延迟时间
