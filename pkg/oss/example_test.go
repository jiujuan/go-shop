package oss_test

import (
	"context"
	"fmt"
	"log"
	"strings"

	"go-shop/pkg/oss"
)

// Example_basicUsage 演示基本使用方法
func Example_basicUsage() {
	// 1. 创建配置
	config := &oss.Config{
		Endpoint:        "localhost:9000",
		AccessKeyID:     "minioadmin",
		AccessKeySecret: "minioadmin",
		BucketName:      "go-shop",
		UseSSL:          false,
	}

	// 2. 初始化 OSS 客户端
	err := oss.Init(config)
	if err != nil {
		log.Fatal(err)
	}

	// 3. 获取客户端实例
	client := oss.MustGetClient()

	// 4. 上传文件
	ctx := context.Background()
	fileContent := strings.NewReader("Hello, OSS!")
	url, err := client.UploadFile(ctx, "test/hello.txt", fileContent, int64(len("Hello, OSS!")), "text/plain")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("File uploaded successfully:", url)
}

// Example_uploadImage 演示上传图片
func Example_uploadImage() {
	// 初始化客户端（省略配置步骤）
	config := &oss.Config{
		Endpoint:        "localhost:9000",
		AccessKeyID:     "minioadmin",
		AccessKeySecret: "minioadmin",
		BucketName:      "go-shop",
		UseSSL:          false,
	}
	oss.Init(config)
	client := oss.MustGetClient()

	// 模拟图片文件
	imageContent := strings.NewReader("fake image content")
	
	ctx := context.Background()
	url, err := oss.UploadImage(ctx, client, imageContent, "product.jpg", "products")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Image uploaded:", url)
}

// Example_batchUpload 演示批量上传
func Example_batchUpload() {
	// 初始化客户端（省略配置步骤）
	config := &oss.Config{
		Endpoint:        "localhost:9000",
		AccessKeyID:     "minioadmin",
		AccessKeySecret: "minioadmin",
		BucketName:      "go-shop",
		UseSSL:          false,
	}
	oss.Init(config)
	_ = oss.MustGetClient()

	// 准备多个文件
	files := []string{"image1.jpg", "image2.jpg", "image3.jpg"}
	readers := make([]interface{}, len(files))
	for i := range files {
		readers[i] = strings.NewReader("fake image content")
	}

	// 注意：实际使用时需要将 readers 转换为 []io.Reader
	fmt.Println("Batch upload example - files:", files)
}

// Example_deleteFile 演示删除文件
func Example_deleteFile() {
	// 初始化客户端（省略配置步骤）
	config := &oss.Config{
		Endpoint:        "localhost:9000",
		AccessKeyID:     "minioadmin",
		AccessKeySecret: "minioadmin",
		BucketName:      "go-shop",
		UseSSL:          false,
	}
	oss.Init(config)
	_ = oss.MustGetClient()

	fmt.Println("File deletion example")
}

// Example_generateObjectName 演示生成对象名称
func Example_generateObjectName() {
	// 生成带前缀的对象名称
	objectName1 := oss.GenerateObjectName("product.jpg", "products")
	fmt.Println("Object name with prefix:", objectName1)

	// 生成不带前缀的对象名称
	objectName2 := oss.GenerateObjectName("avatar.png", "")
	fmt.Println("Object name without prefix:", objectName2)
}
