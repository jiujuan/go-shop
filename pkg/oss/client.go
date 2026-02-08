package oss

import (
	"context"
	"fmt"
	"io"
	"mime"
	"path/filepath"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Client OSS 客户端接口
type Client interface {
	// UploadFile 上传文件
	UploadFile(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (string, error)

	// DownloadFile 下载文件
	DownloadFile(ctx context.Context, objectName string) (io.ReadCloser, error)

	// DeleteFile 删除文件
	DeleteFile(ctx context.Context, objectName string) error

	// GetFileURL 获取文件访问 URL
	GetFileURL(objectName string) string

	// ValidateImage 验证图片文件
	ValidateImage(reader io.Reader, filename string, maxSize int64) error
}

// minioClient MinIO 客户端实现
type minioClient struct {
	client     *minio.Client
	config     *Config
	bucketName string
}

// NewClient 创建 OSS 客户端
func NewClient(config *Config) (Client, error) {
	// 验证配置
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// 创建 MinIO 客户端
	client, err := minio.New(config.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.AccessKeyID, config.AccessKeySecret, ""),
		Secure: config.UseSSL,
		Region: config.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	// 检查存储桶是否存在
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	exists, err := client.BucketExists(ctx, config.BucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket existence: %w", err)
	}

	// 如果存储桶不存在，创建它
	if !exists {
		err = client.MakeBucket(ctx, config.BucketName, minio.MakeBucketOptions{
			Region: config.Region,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	return &minioClient{
		client:     client,
		config:     config,
		bucketName: config.BucketName,
	}, nil
}

// UploadFile 上传文件
func (c *minioClient) UploadFile(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (string, error) {
	// 如果未指定 Content-Type，尝试从文件名推断
	if contentType == "" {
		contentType = mime.TypeByExtension(filepath.Ext(objectName))
		if contentType == "" {
			contentType = "application/octet-stream"
		}
	}

	// 上传文件
	_, err := c.client.PutObject(ctx, c.bucketName, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	// 返回文件 URL
	return c.GetFileURL(objectName), nil
}

// DownloadFile 下载文件
func (c *minioClient) DownloadFile(ctx context.Context, objectName string) (io.ReadCloser, error) {
	object, err := c.client.GetObject(ctx, c.bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}

	return object, nil
}

// DeleteFile 删除文件
func (c *minioClient) DeleteFile(ctx context.Context, objectName string) error {
	err := c.client.RemoveObject(ctx, c.bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// GetFileURL 获取文件访问 URL
func (c *minioClient) GetFileURL(objectName string) string {
	// 如果配置了 CDN 域名，使用 CDN 域名
	if c.config.Domain != "" {
		return fmt.Sprintf("%s/%s", strings.TrimRight(c.config.Domain, "/"), objectName)
	}

	// 否则使用 OSS 端点
	protocol := "http"
	if c.config.UseSSL {
		protocol = "https"
	}

	return fmt.Sprintf("%s://%s/%s/%s", protocol, c.config.Endpoint, c.bucketName, objectName)
}

// ValidateImage 验证图片文件
func (c *minioClient) ValidateImage(reader io.Reader, filename string, maxSize int64) error {
	// 验证文件大小
	if maxSize <= 0 {
		maxSize = MaxImageSize
	}

	// 从文件名获取扩展名
	ext := strings.ToLower(filepath.Ext(filename))
	contentType := mime.TypeByExtension(ext)

	// 验证文件格式
	isValidFormat := false
	for _, allowedFormat := range AllowedImageFormats {
		if contentType == allowedFormat {
			isValidFormat = true
			break
		}
	}

	if !isValidFormat {
		return fmt.Errorf("invalid image format: %s, allowed formats: %v", contentType, AllowedImageFormats)
	}

	return nil
}
