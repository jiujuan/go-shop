package mq

import (
	"encoding/json"
	"fmt"

	"github.com/apache/rocketmq-client-go/v2/primitive"
)

// Message 通用消息结构
type Message struct {
	EventType string                 `json:"event_type"` // 事件类型
	Data      map[string]interface{} `json:"data"`       // 消息数据
	Timestamp int64                  `json:"timestamp"`  // 时间戳
}

// MarshalMessage 序列化消息
func MarshalMessage(msg *Message) ([]byte, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}
	return data, nil
}

// UnmarshalMessage 反序列化消息
func UnmarshalMessage(data []byte) (*Message, error) {
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %w", err)
	}
	return &msg, nil
}

// ExtractMessageBody 从 MessageExt 中提取消息体
func ExtractMessageBody(msg *primitive.MessageExt) ([]byte, error) {
	if msg == nil {
		return nil, fmt.Errorf("message is nil")
	}
	return msg.Body, nil
}

// GetMessageID 获取消息 ID
func GetMessageID(msg *primitive.MessageExt) string {
	if msg == nil {
		return ""
	}
	return msg.MsgId
}

// GetMessageTopic 获取消息 Topic
func GetMessageTopic(msg *primitive.MessageExt) string {
	if msg == nil {
		return ""
	}
	return msg.Topic
}

// GetMessageTag 获取消息 Tag
func GetMessageTag(msg *primitive.MessageExt) string {
	if msg == nil {
		return ""
	}
	return msg.GetTags()
}

// GetMessageKeys 获取消息 Keys
func GetMessageKeys(msg *primitive.MessageExt) string {
	if msg == nil {
		return ""
	}
	return msg.GetKeys()
}
