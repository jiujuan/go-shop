package sms

import (
	"fmt"
	"time"
)

// Config 短信服务配置
type Config struct {
	AccessKeyID     string `yaml:"access_key_id"`     // Access Key ID
	AccessKeySecret string `yaml:"access_key_secret"` // Access Key Secret
	SignName        string `yaml:"sign_name"`         // 短信签名
	TemplateCode    string `yaml:"template_code"`     // 短信模板代码
	Region          string `yaml:"region"`            // 区域（默认 cn-hangzhou）
}

// Validate 验证配置
func (c *Config) Validate() error {
	if c.AccessKeyID == "" {
		return fmt.Errorf("access_key_id is required")
	}

	if c.AccessKeySecret == "" {
		return fmt.Errorf("access_key_secret is required")
	}

	if c.SignName == "" {
		return fmt.Errorf("sign_name is required")
	}

	if c.TemplateCode == "" {
		return fmt.Errorf("template_code is required")
	}

	if c.Region == "" {
		c.Region = "cn-hangzhou" // 默认区域
	}

	return nil
}

// SendOptions 发送选项
type SendOptions struct {
	PhoneNumber string        // 手机号
	Code        string        // 验证码
	Timeout     time.Duration // 发送超时时间
}

// DefaultSendOptions 默认发送选项
var DefaultSendOptions = SendOptions{
	Timeout: 5 * time.Second,
}

// Redis 键前缀常量
const (
	// RedisKeyCodePrefix 验证码存储键前缀
	RedisKeyCodePrefix = "sms:code:"

	// RedisKeySentPrefix 发送记录键前缀（防止重复发送）
	RedisKeySentPrefix = "sms:sent:"

	// RedisKeyCountPrefix 发送次数统计键前缀
	RedisKeyCountPrefix = "sms:count:"

	// CodeExpiration 验证码过期时间（5 分钟）
	CodeExpiration = 5 * time.Minute

	// SentExpiration 发送记录过期时间（60 秒）
	SentExpiration = 60 * time.Second

	// CountExpiration 发送次数统计过期时间（1 小时）
	CountExpiration = 1 * time.Hour

	// MaxSendCount 每小时最大发送次数
	MaxSendCount = 5

	// CodeLength 验证码长度
	CodeLength = 6
)
