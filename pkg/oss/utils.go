package oss

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// GenerateObjectName 生成对象名称
// 格式: {prefix}/{date}/{uuid}.{ext}
func GenerateObjectName(filename, prefix string) string {
	// 获取文件扩展名
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		ext = ".bin"
	}

	// 生成 UUID
	id := uuid.New().String()

	// 生成日期路径
	date := time.Now().Format("2006/01/02")

	// 组合对象名称
	if prefix != "" {
		return fmt.Sprintf("%s/%s/%s%s", prefix, date, id, ext)
	}

	return fmt.Sprintf("%s/%s%s", date, id, ext)
}

// UploadImage 上传图片（带验证）
func UploadImage(ctx context.Context, client Client, reader io.Reader, filename string, prefix string) (string, error) {
	// 验证图片
	if err := client.ValidateImage(reader, filename, MaxImageSize); err != nil {
		return "", err
	}

	// 生成对象名称
	objectName := GenerateObjectName(filename, prefix)

	// 读取文件内容到内存（用于获取大小）
	// 注意：这里为了简化实现，将文件读入内存
	// 在生产环境中，应该使用更高效的方式处理大文件
	data, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	// 验证文件大小
	if int64(len(data)) > MaxImageSize {
		return "", fmt.Errorf("image size exceeds maximum allowed size: %d bytes", MaxImageSize)
	}

	// 上传文件
	contentType := GetContentType(filename)
	url, err := client.UploadFile(ctx, objectName, strings.NewReader(string(data)), int64(len(data)), contentType)
	if err != nil {
		return "", err
	}

	return url, nil
}

// UploadImages 批量上传图片
func UploadImages(ctx context.Context, client Client, files []io.Reader, filenames []string, prefix string, maxCount int) ([]string, error) {
	if len(files) != len(filenames) {
		return nil, fmt.Errorf("files and filenames length mismatch")
	}

	if maxCount > 0 && len(files) > maxCount {
		return nil, fmt.Errorf("too many files: maximum %d files allowed", maxCount)
	}

	urls := make([]string, 0, len(files))

	for i, file := range files {
		url, err := UploadImage(ctx, client, file, filenames[i], prefix)
		if err != nil {
			return urls, fmt.Errorf("failed to upload file %s: %w", filenames[i], err)
		}
		urls = append(urls, url)
	}

	return urls, nil
}

// GetContentType 根据文件名获取 Content-Type
func GetContentType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".bmp":
		return "image/bmp"
	default:
		return "application/octet-stream"
	}
}

// ExtractObjectName 从 URL 中提取对象名称
func ExtractObjectName(url string) string {
	// 移除协议和域名部分
	parts := strings.Split(url, "/")
	if len(parts) < 2 {
		return url
	}

	// 找到存储桶名称后的部分
	for i := range parts {
		if i >= 3 { // 跳过 protocol://domain/bucket
			return strings.Join(parts[i:], "/")
		}
	}

	return url
}
