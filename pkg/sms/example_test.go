package sms_test

import (
	"context"
	"fmt"
	"log"

	"go-shop/pkg/sms"
)

// ExampleInit 演示如何初始化短信服务客户端
func ExampleInit() {
	// 创建配置
	config := &sms.Config{
		AccessKeyID:     "your-access-key-id",
		AccessKeySecret: "your-access-key-secret",
		SignName:        "Go-Shop",
		TemplateCode:    "SMS_123456789",
		Region:          "cn-hangzhou",
	}

	// 初始化全局客户端
	err := sms.Init(config)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("SMS client initialized successfully")
}

// ExampleClient_SendVerificationCode 演示如何发送验证码
func ExampleClient_SendVerificationCode() {
	// 注意：这个示例需要有效的阿里云配置才能运行
	config := &sms.Config{
		AccessKeyID:     "your-access-key-id",
		AccessKeySecret: "your-access-key-secret",
		SignName:        "Go-Shop",
		TemplateCode:    "SMS_123456789",
		Region:          "cn-hangzhou",
	}

	// 初始化客户端
	err := sms.Init(config)
	if err != nil {
		log.Fatal(err)
	}

	// 获取客户端
	client, err := sms.GetClient()
	if err != nil {
		log.Fatal(err)
	}

	// 发送验证码
	code, err := client.SendVerificationCode(context.Background(), "13800138000")
	if err != nil {
		log.Printf("发送验证码失败: %v", err)
		return
	}

	fmt.Printf("验证码已发送: %s\n", code)
}

// ExampleConfig_Validate 演示配置验证
func ExampleConfig_Validate() {
	// 有效配置
	validConfig := &sms.Config{
		AccessKeyID:     "test-key-id",
		AccessKeySecret: "test-key-secret",
		SignName:        "Go-Shop",
		TemplateCode:    "SMS_123456789",
		Region:          "cn-hangzhou",
	}

	err := validConfig.Validate()
	if err != nil {
		fmt.Printf("配置验证失败: %v\n", err)
	} else {
		fmt.Println("配置验证成功")
	}

	// 无效配置（缺少必填字段）
	invalidConfig := &sms.Config{
		AccessKeyID: "test-key-id",
		// 缺少其他必填字段
	}

	err = invalidConfig.Validate()
	if err != nil {
		fmt.Printf("配置验证失败: %v\n", err)
	}

	// Output:
	// 配置验证成功
	// 配置验证失败: access_key_secret is required
}
