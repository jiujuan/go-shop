package mq

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				NameServers: []string{"localhost:9876"},
				GroupName:   "test-group",
				Retry:       2,
				Timeout:     3 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "missing name servers",
			config: &Config{
				GroupName: "test-group",
				Retry:     2,
				Timeout:   3 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "missing group name",
			config: &Config{
				NameServers: []string{"localhost:9876"},
				Retry:       2,
				Timeout:     3 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "default retry and timeout",
			config: &Config{
				NameServers: []string{"localhost:9876"},
				GroupName:   "test-group",
				Retry:       0,
				Timeout:     0,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Check defaults are set
				if tt.config.Retry == 0 {
					assert.Equal(t, 2, tt.config.Retry)
				}
				if tt.config.Timeout == 0 {
					assert.Equal(t, 3*time.Second, tt.config.Timeout)
				}
			}
		})
	}
}

func TestTopicConstants(t *testing.T) {
	// Test that topic constants are defined
	assert.Equal(t, "order-events", TopicOrderEvents)
	assert.Equal(t, "inventory-events", TopicInventoryEvents)
	assert.Equal(t, "notification-events", TopicNotificationEvents)
	assert.Equal(t, "statistics-events", TopicStatisticsEvents)
}

func TestTagConstants(t *testing.T) {
	// Test that tag constants are defined
	assert.Equal(t, "order_created", TagOrderCreated)
	assert.Equal(t, "order_paid", TagOrderPaid)
	assert.Equal(t, "order_cancelled", TagOrderCancelled)
	assert.Equal(t, "order_completed", TagOrderCompleted)

	assert.Equal(t, "inventory_deducted", TagInventoryDeducted)
	assert.Equal(t, "inventory_restored", TagInventoryRestored)
	assert.Equal(t, "inventory_low_stock", TagInventoryLowStock)

	assert.Equal(t, "notification_order", TagNotificationOrder)
	assert.Equal(t, "notification_promotion", TagNotificationPromotion)
	assert.Equal(t, "notification_system", TagNotificationSystem)

	assert.Equal(t, "statistics_order", TagStatisticsOrder)
	assert.Equal(t, "statistics_product", TagStatisticsProduct)
	assert.Equal(t, "statistics_user", TagStatisticsUser)
}

func TestMarshalUnmarshalMessage(t *testing.T) {
	// Create a test message
	msg := &Message{
		EventType: "test_event",
		Data: map[string]interface{}{
			"key1": "value1",
			"key2": 123,
		},
		Timestamp: time.Now().Unix(),
	}

	// Marshal
	data, err := MarshalMessage(msg)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	// Unmarshal
	unmarshaledMsg, err := UnmarshalMessage(data)
	assert.NoError(t, err)
	assert.Equal(t, msg.EventType, unmarshaledMsg.EventType)
	assert.Equal(t, msg.Timestamp, unmarshaledMsg.Timestamp)
	assert.Equal(t, msg.Data["key1"], unmarshaledMsg.Data["key1"])
}

func TestConsumeResults(t *testing.T) {
	// Test consume result helpers
	successResult := ConsumeSuccess()
	assert.NotNil(t, successResult)

	retryResult := ConsumeRetryLater()
	assert.NotNil(t, retryResult)
}
