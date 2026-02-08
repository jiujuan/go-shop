package oss

import (
	"fmt"
	"time"
)

// Config OSS 配置
type Config struct {
	Endpoint        string `yaml:"endpoint"`         // OSS 端点地址
	AccessKeyID     string `yaml:"access_key_id"`    // Access Key ID
	AccessKeySecret string `yaml:"access_key_secret"` // Access Key Secret
	BucketName      string `yaml:"bucket_name"`      // 存储桶名称
	Domain          string `yaml:"domain"`           // CDN 域名（可选）
	UseSSL          bool   `yaml:"use_ssl"`          // 是否使用 SSL
	Region          string `yaml:"region"`           // 区域（可选）
}

// Validate 验证配置
func (c *Config) Validate() error {
	if c.Endpoint == "" {
		return fmt.Errorf("endpoint is required")
	}

	if c.AccessKeyID == "" {
		return fmt.Errorf("access_key_id is required")
	}

	if c.AccessKeySecret == "" {
		return fmt.Errorf("access_key_secret is required")
	}

	if c.BucketName == "" {
		return fmt.Errorf("bucket_name is required")
	}

	return nil
}

// UploadOptions 上传选项
type UploadOptions struct {
	ContentType string        // 文件 MIME 类型
	MaxSize     int64         // 最大文件大小（字节）
	Timeout     time.Duration // 上传超时时间
}

// DefaultUploadOptions 默认上传选项
var DefaultUploadOptions = UploadOptions{
	ContentType: "application/octet-stream",
	MaxSize:     5 * 1024 * 1024, // 5MB
	Timeout:     30 * time.Second,
}

// AllowedImageFormats 允许的图片格式
var AllowedImageFormats = []string{
	"image/jpeg",
	"image/jpg",
	"image/png",
	"image/gif",
}

// MaxImageSize 最大图片大小（5MB）
const MaxImageSize = 5 * 1024 * 1024
