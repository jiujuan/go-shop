package config

import (
	"fmt"
	"os"

	"go-shop/pkg/cache"
	"go-shop/pkg/database"

	"gopkg.in/yaml.v3"
)

// Config 应用配置结构
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Cache    CacheConfig    `yaml:"cache"`
	JWT      JWTConfig      `yaml:"jwt"`
	Log      LogConfig      `yaml:"log"`
	MQ       MQConfig       `yaml:"mq"`
	OSS      OSSConfig      `yaml:"oss"`
	SMS      SMSConfig      `yaml:"sms"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port int    `yaml:"port"`
	Mode string `yaml:"mode"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	MySQL database.Config `yaml:"mysql"`
}

// CacheConfig 缓存配置
type CacheConfig struct {
	Redis cache.RedisConfig `yaml:"redis"`
}

// JWTConfig JWT 配置
type JWTConfig struct {
	Secret      string `yaml:"secret"`
	ExpireHours int    `yaml:"expire_hours"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level    string `yaml:"level"`
	FilePath string `yaml:"file_path"`
}

// MQConfig 消息队列配置
type MQConfig struct {
	NameServers []string `yaml:"name_servers"` // NameServer 地址列表
	GroupName   string   `yaml:"group_name"`   // 消费者组名称
	Retry       int      `yaml:"retry"`        // 重试次数
	Timeout     int      `yaml:"timeout"`      // 超时时间（秒）
}

// OSSConfig OSS 配置
type OSSConfig struct {
	Endpoint        string `yaml:"endpoint"`         // OSS 端点地址
	AccessKeyID     string `yaml:"access_key_id"`    // Access Key ID
	AccessKeySecret string `yaml:"access_key_secret"` // Access Key Secret
	BucketName      string `yaml:"bucket_name"`      // 存储桶名称
	Domain          string `yaml:"domain"`           // CDN 域名（可选）
	UseSSL          bool   `yaml:"use_ssl"`          // 是否使用 SSL
	Region          string `yaml:"region"`           // 区域（可选）
}

// SMSConfig 短信服务配置
type SMSConfig struct {
	AccessKeyID     string `yaml:"access_key_id"`     // Access Key ID
	AccessKeySecret string `yaml:"access_key_secret"` // Access Key Secret
	SignName        string `yaml:"sign_name"`         // 短信签名
	TemplateCode    string `yaml:"template_code"`     // 短信模板代码
	Region          string `yaml:"region"`            // 区域（默认 cn-hangzhou）
}

// LoadConfig 加载配置文件
func LoadConfig(configPath string) (*Config, error) {
	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// 解析 YAML
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// 验证配置
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &config, nil
}

// validateConfig 验证配置
func validateConfig(config *Config) error {
	if config.Server.Port <= 0 || config.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", config.Server.Port)
	}

	if config.Database.MySQL.Host == "" {
		return fmt.Errorf("database host is required")
	}

	if config.Database.MySQL.Database == "" {
		return fmt.Errorf("database name is required")
	}

	if config.Cache.Redis.Host == "" {
		return fmt.Errorf("redis host is required")
	}

	if config.JWT.Secret == "" {
		return fmt.Errorf("jwt secret is required")
	}

	// 验证 MQ 配置（可选）
	if len(config.MQ.NameServers) > 0 && config.MQ.GroupName == "" {
		return fmt.Errorf("mq group_name is required when name_servers is configured")
	}

	// 验证 OSS 配置（可选）
	if config.OSS.Endpoint != "" {
		if config.OSS.AccessKeyID == "" {
			return fmt.Errorf("oss access_key_id is required when endpoint is configured")
		}
		if config.OSS.AccessKeySecret == "" {
			return fmt.Errorf("oss access_key_secret is required when endpoint is configured")
		}
		if config.OSS.BucketName == "" {
			return fmt.Errorf("oss bucket_name is required when endpoint is configured")
		}
	}

	// 验证 SMS 配置（可选）
	if config.SMS.AccessKeyID != "" || config.SMS.AccessKeySecret != "" {
		if config.SMS.AccessKeyID == "" {
			return fmt.Errorf("sms access_key_id is required when sms is configured")
		}
		if config.SMS.AccessKeySecret == "" {
			return fmt.Errorf("sms access_key_secret is required when sms is configured")
		}
		if config.SMS.SignName == "" {
			return fmt.Errorf("sms sign_name is required when sms is configured")
		}
		if config.SMS.TemplateCode == "" {
			return fmt.Errorf("sms template_code is required when sms is configured")
		}
	}

	return nil
}

// GetServerAddr 获取服务器地址
func (c *Config) GetServerAddr() string {
	return fmt.Sprintf(":%d", c.Server.Port)
}