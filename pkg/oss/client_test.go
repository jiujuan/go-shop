package oss

import (
	"testing"

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
				Endpoint:        "localhost:9000",
				AccessKeyID:     "minioadmin",
				AccessKeySecret: "minioadmin",
				BucketName:      "test-bucket",
			},
			wantErr: false,
		},
		{
			name: "missing endpoint",
			config: &Config{
				AccessKeyID:     "minioadmin",
				AccessKeySecret: "minioadmin",
				BucketName:      "test-bucket",
			},
			wantErr: true,
			errMsg:  "endpoint is required",
		},
		{
			name: "missing access key id",
			config: &Config{
				Endpoint:        "localhost:9000",
				AccessKeySecret: "minioadmin",
				BucketName:      "test-bucket",
			},
			wantErr: true,
			errMsg:  "access_key_id is required",
		},
		{
			name: "missing access key secret",
			config: &Config{
				Endpoint:    "localhost:9000",
				AccessKeyID: "minioadmin",
				BucketName:  "test-bucket",
			},
			wantErr: true,
			errMsg:  "access_key_secret is required",
		},
		{
			name: "missing bucket name",
			config: &Config{
				Endpoint:        "localhost:9000",
				AccessKeyID:     "minioadmin",
				AccessKeySecret: "minioadmin",
			},
			wantErr: true,
			errMsg:  "bucket_name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetFileURL(t *testing.T) {
	tests := []struct {
		name       string
		config     *Config
		objectName string
		want       string
	}{
		{
			name: "with CDN domain",
			config: &Config{
				Endpoint:   "localhost:9000",
				BucketName: "test-bucket",
				Domain:     "https://cdn.example.com",
			},
			objectName: "products/image.jpg",
			want:       "https://cdn.example.com/products/image.jpg",
		},
		{
			name: "without CDN domain - HTTP",
			config: &Config{
				Endpoint:   "localhost:9000",
				BucketName: "test-bucket",
				UseSSL:     false,
			},
			objectName: "products/image.jpg",
			want:       "http://localhost:9000/test-bucket/products/image.jpg",
		},
		{
			name: "without CDN domain - HTTPS",
			config: &Config{
				Endpoint:   "s3.amazonaws.com",
				BucketName: "test-bucket",
				UseSSL:     true,
			},
			objectName: "products/image.jpg",
			want:       "https://s3.amazonaws.com/test-bucket/products/image.jpg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &minioClient{
				config:     tt.config,
				bucketName: tt.config.BucketName,
			}
			got := client.GetFileURL(tt.objectName)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetContentType(t *testing.T) {
	tests := []struct {
		filename string
		want     string
	}{
		{"image.jpg", "image/jpeg"},
		{"image.jpeg", "image/jpeg"},
		{"image.png", "image/png"},
		{"image.gif", "image/gif"},
		{"image.webp", "image/webp"},
		{"image.bmp", "image/bmp"},
		{"file.txt", "application/octet-stream"},
		{"unknown", "application/octet-stream"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			got := GetContentType(tt.filename)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGenerateObjectName(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		prefix   string
	}{
		{
			name:     "with prefix",
			filename: "image.jpg",
			prefix:   "products",
		},
		{
			name:     "without prefix",
			filename: "image.png",
			prefix:   "",
		},
		{
			name:     "no extension",
			filename: "file",
			prefix:   "uploads",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			objectName := GenerateObjectName(tt.filename, tt.prefix)
			
			// 验证对象名称不为空
			assert.NotEmpty(t, objectName)
			
			// 验证包含日期路径
			assert.Contains(t, objectName, "/")
			
			// 验证包含前缀（如果提供）
			if tt.prefix != "" {
				assert.Contains(t, objectName, tt.prefix)
			}
		})
	}
}
