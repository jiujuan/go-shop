package sms

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
		errMsg  string
	}{
		{
			name: "valid config",
			config: &Config{
				AccessKeyID:     "test-key-id",
				AccessKeySecret: "test-key-secret",
				SignName:        "Go-Shop",
				TemplateCode:    "SMS_123456789",
				Region:          "cn-hangzhou",
			},
			wantErr: false,
		},
		{
			name: "valid config with default region",
			config: &Config{
				AccessKeyID:     "test-key-id",
				AccessKeySecret: "test-key-secret",
				SignName:        "Go-Shop",
				TemplateCode:    "SMS_123456789",
			},
			wantErr: false,
		},
		{
			name: "missing access_key_id",
			config: &Config{
				AccessKeySecret: "test-key-secret",
				SignName:        "Go-Shop",
				TemplateCode:    "SMS_123456789",
			},
			wantErr: true,
			errMsg:  "access_key_id is required",
		},
		{
			name: "missing access_key_secret",
			config: &Config{
				AccessKeyID:  "test-key-id",
				SignName:     "Go-Shop",
				TemplateCode: "SMS_123456789",
			},
			wantErr: true,
			errMsg:  "access_key_secret is required",
		},
		{
			name: "missing sign_name",
			config: &Config{
				AccessKeyID:     "test-key-id",
				AccessKeySecret: "test-key-secret",
				TemplateCode:    "SMS_123456789",
			},
			wantErr: true,
			errMsg:  "sign_name is required",
		},
		{
			name: "missing template_code",
			config: &Config{
				AccessKeyID:     "test-key-id",
				AccessKeySecret: "test-key-secret",
				SignName:        "Go-Shop",
			},
			wantErr: true,
			errMsg:  "template_code is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				// 验证默认区域是否设置
				if tt.config.Region == "" {
					assert.Equal(t, "cn-hangzhou", tt.config.Region)
				}
			}
		})
	}
}

func TestConstants(t *testing.T) {
	// 测试常量值
	assert.Equal(t, "sms:code:", RedisKeyCodePrefix)
	assert.Equal(t, "sms:sent:", RedisKeySentPrefix)
	assert.Equal(t, "sms:count:", RedisKeyCountPrefix)
	assert.Equal(t, 5*time.Minute, CodeExpiration)
	assert.Equal(t, 60*time.Second, SentExpiration)
	assert.Equal(t, 1*time.Hour, CountExpiration)
	assert.Equal(t, 5, MaxSendCount)
	assert.Equal(t, 6, CodeLength)
}

func TestDefaultSendOptions(t *testing.T) {
	// 测试默认发送选项
	assert.Equal(t, 5*time.Second, DefaultSendOptions.Timeout)
}
